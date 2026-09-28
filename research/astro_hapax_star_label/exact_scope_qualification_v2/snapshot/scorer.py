#!/usr/bin/env python3
"""
Clean-room scorer for restricted_dictionary_bruteforce_remediation_v1.

Single canonical implementation of: normalization, global substitution
mappings (injective and bounded MERGE_1), selective/unmapped deletion,
abbreviation, complexity penalty, per-page identity capacity, unmatched
state, bipartite assignment, and train/held-out split.

The core encode/matching primitives (encode_word, maximum_bipartite_matching,
_match_single_capacity, compute_complexity) are reused verbatim from
restricted_dictionary_bruteforce_v3_design/scripts/engine.py per
COMPONENT_REUSE_REGISTRY.tsv (BIPARTITE_MATCHER, CAPACITY_POLICY) --
those two primitives were independently cross-validated by
restricted_dictionary_bruteforce_v3_audit at 400/400 randomized trials.
Everything else here (held-out split, unmatched-state accounting, the
single evaluate() entrypoint used by every caller) is new.

Every future caller (synthetic dev/hidden benchmarks, the discriminative
benchmark, the null pipeline) MUST call Scorer.evaluate() -- there is no
second scoring code path anywhere in this package.
"""
import unicodedata
import re
from collections import defaultdict
from typing import Dict, List, Set, Tuple, Optional

SOURCE_ALPHABET = tuple("abcdefghijklmnopqrstuvwxyz") + ("kh", "gh", "sh", "th", "dh")
EVA_ALPHABET = tuple("acdefhiklnoprsty")  # from REAL_SCOPE_STRUCTURAL_MANIFEST.json, aggregate only


def normalize_text(s: str) -> str:
    s = unicodedata.normalize("NFKD", s).encode("ascii", "ignore").decode().lower()
    return re.sub(r"[^a-z]+", " ", s).strip()


def encode_word(
    word: str,
    table: Dict[str, str],
    deletion_mode: str = "DROP_UNMAPPED",
    abbreviation: str = "NONE",
) -> str:
    """Reused verbatim from v3_design engine.py (COMPONENT_REUSE_REGISTRY.tsv: reused as an
    input to the independently-validated matcher, not itself separately re-audited since it
    is pure deterministic string transformation with no matching/capacity logic)."""
    clean_w = word.replace(" ", "")
    keys = sorted(table.keys(), key=len, reverse=True)
    out = []
    i = 0
    n = len(clean_w)
    while i < n:
        matched_key = None
        for k in keys:
            if clean_w.startswith(k, i):
                matched_key = k
                break
        if matched_key:
            out.append(table[matched_key])
            i += len(matched_key)
        else:
            ch = clean_w[i]
            if deletion_mode == "DROP_UNMAPPED":
                i += 1
            elif deletion_mode == "SELECTIVE_VOWEL_DROP":
                if ch in "aeiouy":
                    i += 1
                else:
                    out.append(ch)
                    i += 1
            else:
                out.append(ch)
                i += 1
    encoded = "".join(out)
    if abbreviation == "SUSPENSION_1" and len(encoded) > 3:
        encoded = encoded[:-1]
    elif abbreviation == "SUSPENSION_2" and len(encoded) > 4:
        encoded = encoded[:-2]
    elif abbreviation == "PREFIX_4":
        encoded = encoded[:4]
    return encoded


def compute_complexity(
    table: Dict[str, str],
    mapping_mode: str = "INJECTIVE",
    deletion_mode: str = "DROP_UNMAPPED",
    abbreviation: str = "NONE",
) -> int:
    """Reused verbatim from v3_design engine.py."""
    cost = len(table)
    if mapping_mode == "MERGE_1":
        cost += 1
    if deletion_mode == "DROP_UNMAPPED":
        cost += 2
    elif deletion_mode == "SELECTIVE_VOWEL_DROP":
        cost += 1
    if abbreviation == "SUSPENSION_1":
        cost += 1
    elif abbreviation in ("SUSPENSION_2", "PREFIX_4"):
        cost += 2
    return cost


def is_valid_table(table: Dict[str, str], mapping_mode: str) -> bool:
    """Bounded-merge validity check: INJECTIVE forbids any repeated target character;
    MERGE_1 allows at most one target character to be used by exactly two source
    graphemes (a single bounded merge), never three or more."""
    targets = list(table.values())
    counts = defaultdict(int)
    for t in targets:
        counts[t] += 1
    if mapping_mode == "INJECTIVE":
        return all(c == 1 for c in counts.values())
    if mapping_mode == "MERGE_1":
        over = [c for c in counts.values() if c > 2]
        merges = sum(1 for c in counts.values() if c == 2)
        return len(over) == 0 and merges <= 1
    raise ValueError(f"unknown mapping_mode {mapping_mode}")


def maximum_bipartite_matching(
    labels: List[dict],
    lexicon_index: Dict[str, Set[str]],
    capacity_policy: str = "PER_PAGE_CAPACITY_1",
) -> Dict[str, str]:
    """Reused verbatim from v3_design engine.py. Independently cross-validated by
    restricted_dictionary_bruteforce_v3_audit's from-scratch implementation on 400/400
    randomized trials (INDEPENDENT_MATCHING_CHECK.tsv); re-validated again in this package
    at >=500 trials, see SCORING_PARITY_RESULTS.tsv."""
    sorted_labels = sorted(labels, key=lambda l: (l.get("page_id", ""), l["occurrence_id"]))
    if capacity_policy == "PER_PAGE_CAPACITY_1":
        full_assignments = {}
        pages = sorted(set(l.get("page_id", "default") for l in sorted_labels))
        for p in pages:
            p_labels = [l for l in sorted_labels if l.get("page_id", "default") == p]
            full_assignments.update(_match_single_capacity(p_labels, lexicon_index))
        return full_assignments
    else:
        return _match_single_capacity(sorted_labels, lexicon_index)


def _match_single_capacity(
    labels: List[dict],
    lexicon_index: Dict[str, Set[str]],
) -> Dict[str, str]:
    adj = {}
    for l in labels:
        token = l["token"]
        identities = sorted(lexicon_index.get(token, set()))
        adj[l["occurrence_id"]] = identities
    match_star: Dict[str, str] = {}

    def dfs(lbl_id: str, visited: Set[str]) -> bool:
        for star in adj[lbl_id]:
            if star in visited:
                continue
            visited.add(star)
            if star not in match_star or dfs(match_star[star], visited):
                match_star[star] = lbl_id
                return True
        return False

    for l in labels:
        dfs(l["occurrence_id"], set())
    return {lbl_id: star for star, lbl_id in match_star.items()}


def build_index(
    lexicon: Dict[str, Set[str]],
    table: Dict[str, str],
    deletion_mode: str,
    abbreviation: str,
) -> Dict[str, Set[str]]:
    idx: Dict[str, Set[str]] = defaultdict(set)
    for ident, forms in lexicon.items():
        for form in forms:
            enc = encode_word(form, table, deletion_mode, abbreviation)
            if enc:
                idx[enc].add(ident)
    return idx


class Scorer:
    """The one scorer. Synthetic dev, synthetic hidden, discriminative benchmark, and the
    null pipeline all call this class and nothing else."""

    def __init__(
        self,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        capacity_policy: str = "PER_PAGE_CAPACITY_1",
    ):
        self.mapping_mode = mapping_mode
        self.deletion_mode = deletion_mode
        self.abbreviation = abbreviation
        self.capacity_policy = capacity_policy

    def evaluate(
        self,
        table: Dict[str, str],
        lexicon: Dict[str, Set[str]],
        labels: List[dict],
        held_out_labels: Optional[List[dict]] = None,
    ) -> dict:
        """Evaluate `table` on `labels` (the train/optimization set). If `held_out_labels`
        is given, also compute held-out assignment accuracy against ground truth stored in
        each held-out label's 'true_identity' field (only meaningful on synthetic data where
        ground truth exists; None on any hypothetical real-data call, which is why real-data
        calls are refused elsewhere -- see SCIENTIFIC_SAFETY_TESTS.md)."""
        valid = is_valid_table(table, self.mapping_mode)
        idx = build_index(lexicon, table, self.deletion_mode, self.abbreviation)
        assignments = maximum_bipartite_matching(labels, idx, self.capacity_policy)
        matched = len(assignments)
        total = len(labels)
        coverage = matched / total if total else 0.0
        complexity = compute_complexity(table, self.mapping_mode, self.deletion_mode, self.abbreviation)
        fitness = 100 * matched - complexity if valid else -10**9

        unmatched_occurrence_ids = sorted(set(l["occurrence_id"] for l in labels) - set(assignments.keys()))

        result = {
            "valid_table": valid,
            "matched": matched,
            "total": total,
            "unmatched": len(unmatched_occurrence_ids),
            "unmatched_occurrence_ids": unmatched_occurrence_ids,
            "coverage": coverage,
            "complexity": complexity,
            "fitness": fitness,
            "assignments": assignments,
            "table": dict(table),
            "mapping_mode": self.mapping_mode,
            "deletion_mode": self.deletion_mode,
            "abbreviation": self.abbreviation,
            "capacity_policy": self.capacity_policy,
        }

        if held_out_labels is not None:
            ho_idx = idx  # same table/index applied to held-out set (train-only optimization)
            ho_assign = maximum_bipartite_matching(held_out_labels, ho_idx, self.capacity_policy)
            correct = 0
            for l in held_out_labels:
                pred = ho_assign.get(l["occurrence_id"])
                truth = l.get("true_identity")
                if truth is not None and pred == truth:
                    correct += 1
            ho_total = len(held_out_labels)
            result["held_out_assignments"] = ho_assign
            result["held_out_total"] = ho_total
            result["held_out_correct"] = correct
            result["held_out_assignment_accuracy"] = correct / ho_total if ho_total else 0.0

        return result
