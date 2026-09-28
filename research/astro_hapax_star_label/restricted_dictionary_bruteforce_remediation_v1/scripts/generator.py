#!/usr/bin/env python3
"""
Synthetic instance generator. Physically separate from every solver: this module never imports
oracle_bb/solver_cpsat/solver_heuristic/hybrid, and no solver module imports this one. The only
connection point is the file system (SEALED_BENCHMARK_PROTOCOL.md's dev/hidden pipeline writes
lexicon+labels to one file and truth to a separate file).

Distractor/label lengths are sampled from the REAL length histogram in
REAL_SCOPE_STRUCTURAL_MANIFEST.json (fixing v3_design Finding F013, which used a hand-tuned
discrete approximation instead) -- this file contains only aggregate structural statistics of
the frozen real-label scope (counts/means/histograms), never actual EVA label strings, per
REAL_SCOPE_STRUCTURAL_MANIFEST.json's own restrictions_enforced block.
"""
import json
import random
from pathlib import Path
from typing import Dict, List, Set, Tuple, Optional

from scorer import encode_word, is_valid_table

MANIFEST_PATH = (
    Path(__file__).parent.parent.parent
    / "restricted_dictionary_bruteforce_v3_design"
    / "REAL_SCOPE_STRUCTURAL_MANIFEST.json"
)

SOURCE_POOL = list("abcdefghijklmnop")  # 16-letter synthetic source alphabet, disjoint concept
                                         # from the target EVA alphabet even where letters overlap


def load_manifest() -> dict:
    with open(MANIFEST_PATH) as f:
        return json.load(f)


def _weighted_choice(rng: random.Random, histogram: Dict[str, int]) -> int:
    items = sorted(((int(k), v) for k, v in histogram.items()))
    total = sum(v for _, v in items)
    r = rng.uniform(0, total)
    acc = 0.0
    for length, count in items:
        acc += count
        if r <= acc:
            return length
    return items[-1][0]


class GeneratorConfig:
    def __init__(
        self,
        seed: int,
        n_identities: int = 8,
        table_size: int = 4,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        capacity_policy: str = "PER_PAGE_CAPACITY_1",
        noise_rate: float = 0.0,
        unmatched_rate: float = 0.0,
        n_occurrences: int = 57,
        page_split: Tuple[int, int] = (30, 27),
        forms_per_identity: Tuple[int, int] = (1, 2),
    ):
        self.seed = seed
        self.n_identities = n_identities
        self.table_size = table_size
        self.mapping_mode = mapping_mode
        self.deletion_mode = deletion_mode
        self.abbreviation = abbreviation
        self.capacity_policy = capacity_policy
        self.noise_rate = noise_rate
        self.unmatched_rate = unmatched_rate
        self.n_occurrences = n_occurrences
        self.page_split = page_split
        self.forms_per_identity = forms_per_identity


def _random_valid_table(rng: random.Random, source_alphabet, target_alphabet, table_size, mapping_mode) -> Dict[str, str]:
    srcs = rng.sample(source_alphabet, table_size)
    table = {}
    for s in srcs:
        candidates = list(target_alphabet)
        rng.shuffle(candidates)
        for t in candidates:
            trial = dict(table)
            trial[s] = t
            if is_valid_table(trial, mapping_mode):
                table = trial
                break
    return table


def generate(cfg: GeneratorConfig) -> dict:
    manifest = load_manifest()
    target_alphabet = manifest["eva_alphabet"]["symbols"]
    length_hist = manifest["length_distribution"]["histogram"]

    rng = random.Random(cfg.seed)
    true_table = _random_valid_table(rng, SOURCE_POOL, target_alphabet, cfg.table_size, cfg.mapping_mode)

    lexicon: Dict[str, Set[str]] = {}
    identity_true_encoding: Dict[str, str] = {}
    for i in range(cfg.n_identities):
        n_forms = rng.randint(*cfg.forms_per_identity)
        forms = set()
        primary_len = _weighted_choice(rng, length_hist)
        primary_len = max(2, min(primary_len, len(SOURCE_POOL)))
        primary = "".join(rng.choice(SOURCE_POOL) for _ in range(primary_len))
        forms.add(primary)
        for _ in range(n_forms - 1):
            alt_len = max(2, min(_weighted_choice(rng, length_hist), len(SOURCE_POOL)))
            forms.add("".join(rng.choice(SOURCE_POOL) for _ in range(alt_len)))
        ident_id = f"SYN_ID_{i:03d}"
        lexicon[ident_id] = forms
        enc = encode_word(primary, true_table, cfg.deletion_mode, cfg.abbreviation)
        identity_true_encoding[ident_id] = enc

    n_occ = cfg.n_occurrences
    n_unmatched = round(cfg.unmatched_rate * n_occ)
    n_matched = n_occ - n_unmatched

    page_names = ["f68r1", "f68r2"]
    page_counts = list(cfg.page_split)

    def page_for_index(idx: int) -> str:
        if idx < page_counts[0]:
            return page_names[0]
        return page_names[1]

    identity_ids = sorted(lexicon.keys())
    labels: List[dict] = []

    # matched labels: respect capacity_policy (an identity used at most once per page for
    # PER_PAGE_CAPACITY_1, at most once globally for GLOBAL_CAPACITY_1).
    #
    # IMPORTANT: when no capacity-legal identity remains for a slot that was INTENDED to be a
    # true match (this happens whenever n_identities is small relative to the page's requested
    # matched-label count), that slot is downgraded to a genuine distractor (true_identity=None,
    # random token) instead of being assigned an identity that violates the very capacity policy
    # this generator is supposed to model. An earlier version used `rng.choice(identity_ids)` as a
    # capacity-blind fallback here, which silently manufactured ground truth that the scorer's own
    # capacity-1 bipartite matcher could never fully realize even given the TRUE table -- e.g. at
    # n_identities=12 over a 30-slot page, at most 12/30 "true" matches were ever jointly
    # satisfiable, producing an artificial ~40% recall ceiling that had nothing to do with noise,
    # search quality, or resource budget. See REMEDIATION_REPORT.md for the full account. Callers
    # that need a specific unmatched_rate to actually hold should size n_identities >= the largest
    # page's matched-slot count; `realized_unmatched_rate` in the returned instance reports what
    # was actually achieved when that isn't the case, rather than hiding the shortfall.
    used_global: Set[str] = set()
    used_per_page: Dict[str, Set[str]] = {p: set() for p in page_names}
    matched_slots = list(range(n_matched))
    rng.shuffle(matched_slots)
    id_cycle = identity_ids[:]
    rng.shuffle(id_cycle)
    id_ptr = 0

    occ_page_assignment = [page_for_index(i) for i in range(n_occ)]
    rng.shuffle(occ_page_assignment)

    n_downgraded_to_distractor = 0
    for slot_i, occ_idx in enumerate(matched_slots):
        page = occ_page_assignment[occ_idx]
        chosen = None
        attempts = 0
        while attempts < len(id_cycle) * 2:
            cand = id_cycle[id_ptr % len(id_cycle)]
            id_ptr += 1
            attempts += 1
            if cfg.capacity_policy == "GLOBAL_CAPACITY_1" and cand in used_global:
                continue
            if cfg.capacity_policy == "PER_PAGE_CAPACITY_1" and cand in used_per_page[page]:
                continue
            chosen = cand
            break

        if chosen is None:
            # No capacity-legal identity remains: this slot becomes a real distractor rather than
            # a capacity-violating "true" match.
            n_downgraded_to_distractor += 1
            dlen = max(1, min(_weighted_choice(rng, length_hist), 10))
            labels.append({
                "occurrence_id": f"OCC{occ_idx:04d}",
                "page_id": page,
                "token": "".join(rng.choice(target_alphabet) for _ in range(dlen)),
                "true_identity": None,
            })
            continue

        used_global.add(chosen)
        used_per_page[page].add(chosen)

        token = identity_true_encoding[chosen]
        if rng.random() < cfg.noise_rate and token:
            pos = rng.randrange(len(token))
            new_sym = rng.choice(target_alphabet)
            token = token[:pos] + new_sym + token[pos + 1:]

        labels.append({
            "occurrence_id": f"OCC{occ_idx:04d}",
            "page_id": page,
            "token": token,
            "true_identity": chosen,
        })

    remaining_occ = [i for i in range(n_occ) if i not in matched_slots]
    for occ_idx in remaining_occ:
        page = occ_page_assignment[occ_idx]
        dlen = max(1, min(_weighted_choice(rng, length_hist), 10))
        token = "".join(rng.choice(target_alphabet) for _ in range(dlen))
        labels.append({
            "occurrence_id": f"OCC{occ_idx:04d}",
            "page_id": page,
            "token": token,
            "true_identity": None,
        })

    labels.sort(key=lambda l: l["occurrence_id"])

    n_true_matched = sum(1 for l in labels if l["true_identity"] is not None)
    return {
        "config": vars(cfg),
        "lexicon": {k: sorted(v) for k, v in lexicon.items()},
        "labels": labels,
        "true_table": true_table,
        "target_alphabet": target_alphabet,
        "n_downgraded_to_distractor": n_downgraded_to_distractor,
        "realized_unmatched_rate": round(1 - n_true_matched / n_occ, 4) if n_occ else 0.0,
        "capacity_feasible_as_requested": n_downgraded_to_distractor == 0,
    }


def split_public_private(instance: dict) -> Tuple[dict, dict]:
    """Returns (public, private). public is everything a solver may see: lexicon + labels
    WITHOUT true_identity. private holds true_table + the true_identity map, used only after
    the solver has produced predictions."""
    public_labels = [{"occurrence_id": l["occurrence_id"], "page_id": l["page_id"], "token": l["token"]} for l in instance["labels"]]
    public = {
        "lexicon": instance["lexicon"],
        "labels": public_labels,
        "target_alphabet": instance["target_alphabet"],
        "config_non_truth_fields": {
            k: v for k, v in instance["config"].items() if k not in ("seed",)
        },
    }
    private = {
        "true_table": instance["true_table"],
        "true_identity_by_occurrence": {l["occurrence_id"]: l["true_identity"] for l in instance["labels"]},
        "seed": instance["config"]["seed"],
    }
    return public, private
