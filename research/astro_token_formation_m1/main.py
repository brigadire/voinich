#!/usr/bin/env python3
"""M1 bounded search for a global Latin/Arabic-to-EVA substitution system."""

from __future__ import annotations

import argparse
import csv
import hashlib
import importlib.util
import itertools
import json
import math
import os
import random
import re
import statistics
import sys
import unicodedata
from collections import Counter, defaultdict
from dataclasses import dataclass
from datetime import datetime, timezone
from functools import lru_cache
from multiprocessing import get_context
from pathlib import Path
from typing import Iterable

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/astro_token_formation_m1"
M0_OUT = ROOT / "research/astro_token_formation"
M0_PATH = ROOT / "research/astro_token_formation/main.py"
_spec = importlib.util.spec_from_file_location("astro_m0_for_m1", M0_PATH)
assert _spec and _spec.loader
m0 = importlib.util.module_from_spec(_spec)
sys.modules[_spec.name] = m0
_spec.loader.exec_module(m0)

SEED = 20260901
NULL_REPLICATES = 100
BEAM_WIDTH = 64
FINALISTS_PER_PIPELINE = 4
RETAINED = 25
CAP = 1_000_000_000
SOURCE_DIGRAPHS = ("kh", "gh", "sh", "th", "dh", "ch", "ph", "qu")
TARGET_COMPOSITES = ("cth", "ckh", "cph", "cfh", "iin", "ain", "ch", "sh", "ee", "in")
EXPANSIONS = {"C": "cth", "K": "ckh", "P": "cph", "F": "cfh", "N": "iin", "A": "ain", "H": "ch", "S": "sh", "E": "ee", "I": "in"}


@dataclass(frozen=True)
class Pipeline:
    pipeline_id: str
    lexical: str
    ending: str
    orthography: str
    vowel: str
    abbreviation: str
    complexity: int

    @property
    def rule_string(self) -> str:
        return ";".join((self.lexical, self.ending, self.orthography, self.vowel, self.abbreviation))


@dataclass(frozen=True)
class Edge:
    label: str
    object_id: str
    constraint: tuple[tuple[str, tuple[str, ...]], ...]
    source_form: str
    source_units: tuple[str, ...]


@dataclass(frozen=True)
class State:
    mapping: tuple[tuple[str, tuple[str, ...]], ...]
    used_objects: frozenset[str]
    assignments: tuple[tuple[str, str, str, tuple[str, ...]], ...]


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as fh:
        return list(csv.DictReader(fh, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as fh:
        w = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        w.writeheader(); w.writerows(rows)


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def fmt(x: float) -> str:
    return f"{x:.6f}"


def segment(value: str, priority: tuple[str, ...]) -> tuple[str, ...]:
    result: list[str] = []
    i = 0
    ordered = sorted(priority, key=lambda x: (-len(x), priority.index(x)))
    while i < len(value):
        match = next((unit for unit in ordered if value.startswith(unit, i)), None)
        if match:
            result.append(match); i += len(match)
        else:
            result.append(value[i]); i += 1
    return tuple(result)


def abbreviate(units: tuple[str, ...], mode: str) -> tuple[str, ...]:
    if mode == "SUSPEND_1" and len(units) > 3:
        return units[:-1]
    if mode == "SUSPEND_2" and len(units) > 4:
        return units[:-2]
    if mode == "PREFIX_3" and len(units) > 3:
        return units[:3]
    if mode == "PREFIX_4" and len(units) > 4:
        return units[:4]
    return units


def load_terms() -> list[dict[str, object]]:
    return m0.load_terms()


def load_labels(split: str) -> list[dict[str, str]]:
    return [r for r in read_tsv(M0_OUT / "ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv") if r["selected"] == "1" and r["split"] == split]


def all_pipelines() -> list[Pipeline]:
    result = []
    i = 0
    for lexical in ("FULL_JOIN", "DROP_AL", "HEAD", "TAIL"):
        for ending in ("KEEP", "STRIP_LATIN"):
            for orth in ("IDENTITY", "IJ_UV", "ARABIC_LATIN", "VELAR_COLLAPSE"):
                for vowel in ("KEEP", "DELETE_FINAL", "COLLAPSE_A", "CONTRACT_INTERNAL"):
                    for abbr in ("NONE", "SUSPEND_1", "SUSPEND_2", "PREFIX_3", "PREFIX_4"):
                        i += 1
                        rule = m0.Rule(lexical, ending, orth, vowel, abbr)
                        result.append(Pipeline(f"P{i:04d}", lexical, ending, orth, vowel, abbr, rule.complexity))
    assert len(result) == 640
    return result


def preprocess(terms: list[dict[str, object]], pipe: Pipeline) -> dict[str, tuple[tuple[str, tuple[str, ...]], ...]]:
    rule = m0.Rule(pipe.lexical, pipe.ending, pipe.orthography, pipe.vowel, "NONE")
    result = {}
    for term in terms:
        forms = set()
        for form in term["forms"]:  # type: ignore[index]
            value = m0.apply_rule(str(form), rule)
            units = abbreviate(segment(value, SOURCE_DIGRAPHS), pipe.abbreviation)
            if units:
                forms.add((str(form), units))
        result[str(term["object_id"])] = tuple(sorted(forms))
    return result


def representation_key(pre: dict[str, tuple[tuple[str, tuple[str, ...]], ...]]) -> tuple:
    return tuple((oid, tuple(sorted(units for _form, units in forms))) for oid, forms in sorted(pre.items()))


@lru_cache(maxsize=500_000)
def align_constraints(source: tuple[str, ...], target: tuple[str, ...]) -> tuple[tuple[tuple[str, tuple[str, ...]], ...], ...]:
    n, m = len(source), len(target)
    twos = m - n
    if n == 0 or twos < 0 or twos > n:
        return ()
    result = set()
    for double_positions in itertools.combinations(range(n), twos):
        doubles = set(double_positions); pos = 0; mapping: dict[str, tuple[str, ...]] = {}; valid = True
        for i, unit in enumerate(source):
            width = 2 if i in doubles else 1
            output = target[pos:pos + width]; pos += width
            if unit in mapping and mapping[unit] != output:
                valid = False; break
            mapping[unit] = output
        if valid:
            result.add(tuple(sorted(mapping.items())))
    return tuple(sorted(result))


def build_edges(labels: list[dict[str, str]], terms: list[dict[str, object]], pre: dict[str, tuple[tuple[str, tuple[str, ...]], ...]]) -> dict[str, list[Edge]]:
    term_class = {str(t["object_id"]): str(t["object_class"]) for t in terms}
    result: dict[str, list[Edge]] = {}
    for label in labels:
        target = segment(label["voynich_token"], TARGET_COMPOSITES)
        edges: dict[tuple, Edge] = {}
        for oid, forms in pre.items():
            if term_class[oid] != label["object_class"]:
                continue
            for form, source in forms:
                for constraint in align_constraints(source, target):
                    key = (oid, constraint)
                    candidate = Edge(label["stolfi_coordinate"], oid, constraint, form, source)
                    if key not in edges or (form, source) < (edges[key].source_form, edges[key].source_units):
                        edges[key] = candidate
        result[label["stolfi_coordinate"]] = sorted(edges.values(), key=lambda e: (e.object_id, e.constraint, e.source_form))
    return result


def compatible_merge(mapping_sig: tuple[tuple[str, tuple[str, ...]], ...], constraint: tuple[tuple[str, tuple[str, ...]], ...]) -> tuple[tuple[str, tuple[str, ...]], ...] | None:
    mapping = dict(mapping_sig)
    for source, target in constraint:
        if source in mapping and mapping[source] != target:
            return None
        mapping[source] = target
    return tuple(sorted(mapping.items()))


def mapping_metrics(mapping_sig: tuple[tuple[str, tuple[str, ...]], ...], pipe: Pipeline) -> dict[str, int]:
    outputs = Counter(target for _source, target in mapping_sig)
    one_two = sum(len(target) == 2 for _source, target in mapping_sig)
    digraphs = sum(source in SOURCE_DIGRAPHS for source, _target in mapping_sig)
    collapse = sum(count - 1 for count in outputs.values() if count > 1)
    total = len(mapping_sig) + pipe.complexity + one_two + digraphs + collapse
    return {"mapping_size": len(mapping_sig), "one_to_two": one_two, "digraph_mappings": digraphs, "many_to_one_excess": collapse, "total_complexity": total}


def base_score(matched: int, total: int, mapping_sig: tuple[tuple[str, tuple[str, ...]], ...], pipe: Pipeline, ambiguity_excess: float = 0.0) -> float:
    mm = mapping_metrics(mapping_sig, pipe)
    return (matched / total - .010 * mm["mapping_size"] - .005 * mm["one_to_two"]
            - .005 * mm["digraph_mappings"] - .005 * mm["many_to_one_excess"]
            - .010 * pipe.complexity - .005 * ambiguity_excess)


def state_rank(state: State, total: int, pipe: Pipeline) -> tuple:
    return (-len(state.assignments), -base_score(len(state.assignments), total, state.mapping, pipe), len(state.mapping), state.mapping, tuple(sorted(state.used_objects)))


def search_pipeline(labels: list[dict[str, str]], terms: list[dict[str, object]], pipe: Pipeline, pre: dict[str, tuple[tuple[str, tuple[str, ...]], ...]]) -> list[State]:
    edges = build_edges(labels, terms, pre)
    order = sorted((r["stolfi_coordinate"] for r in labels), key=lambda x: (len(edges[x]), x))
    beam = [State((), frozenset(), ())]
    for label_id in order:
        candidates: dict[tuple, State] = {}
        for state in beam:
            key = (state.mapping, state.used_objects)
            candidates.setdefault(key, state)
            for edge in edges[label_id]:
                if edge.object_id in state.used_objects:
                    continue
                merged = compatible_merge(state.mapping, edge.constraint)
                if merged is None:
                    continue
                assignment = (edge.label, edge.object_id, edge.source_form, edge.source_units)
                new = State(merged, state.used_objects | {edge.object_id}, state.assignments + (assignment,))
                key = (new.mapping, new.used_objects)
                old = candidates.get(key)
                if old is None or new.assignments < old.assignments:
                    candidates[key] = new
        beam = sorted(candidates.values(), key=lambda s: state_rank(s, len(labels), pipe))[:BEAM_WIDTH]
    return sorted(beam, key=lambda s: state_rank(s, len(labels), pipe))[:FINALISTS_PER_PIPELINE]


def render(mapping_sig: tuple[tuple[str, tuple[str, ...]], ...], units: tuple[str, ...]) -> tuple[str, ...] | None:
    mapping = dict(mapping_sig); output: list[str] = []
    for unit in units:
        if unit not in mapping:
            return None
        output.extend(mapping[unit])
    return tuple(output)


def adjacency(labels: list[dict[str, str]], terms: list[dict[str, object]], pre: dict[str, tuple[tuple[str, tuple[str, ...]], ...]], mapping_sig: tuple[tuple[str, tuple[str, ...]], ...], unavailable: set[str] | None = None) -> tuple[dict[str, set[str]], dict[tuple[str, str], list[tuple[str, tuple[str, ...]]]]]:
    unavailable = unavailable or set(); term_class = {str(t["object_id"]): str(t["object_class"]) for t in terms}
    adj: dict[str, set[str]] = defaultdict(set); evidence: dict[tuple[str, str], list[tuple[str, tuple[str, ...]]]] = defaultdict(list)
    for label in labels:
        target = segment(label["voynich_token"], TARGET_COMPOSITES)
        for oid, forms in pre.items():
            if oid in unavailable or term_class[oid] != label["object_class"]:
                continue
            for form, units in forms:
                if render(mapping_sig, units) == target:
                    adj[label["stolfi_coordinate"]].add(oid); evidence[(label["stolfi_coordinate"], oid)].append((form, units))
    return dict(adj), dict(evidence)


def maximum_matching(adj: dict[str, set[str]]) -> dict[str, str]:
    obj_to_label: dict[str, str] = {}
    def visit(label: str, seen: set[str]) -> bool:
        for oid in sorted(adj.get(label, ())):
            if oid in seen: continue
            seen.add(oid)
            if oid not in obj_to_label or visit(obj_to_label[oid], seen):
                obj_to_label[oid] = label; return True
        return False
    for label in sorted(adj): visit(label, set())
    return {label: oid for oid, label in obj_to_label.items()}


def forced_edge_in_max(adj: dict[str, set[str]], label: str, oid: str, maximum: int) -> bool:
    reduced = {lab: {x for x in objs if x != oid} for lab, objs in adj.items() if lab != label}
    return 1 + len(maximum_matching(reduced)) == maximum


def matching_multiplicity(adj: dict[str, set[str]], cap: int = CAP) -> tuple[int, bool]:
    labels = sorted(adj, key=lambda x: (len(adj[x]), x)); objects = sorted({o for v in adj.values() for o in v}); bits = {o: 1 << i for i, o in enumerate(objects)}
    @lru_cache(maxsize=None)
    def rec(i: int, used: int) -> tuple[int, int, bool]:
        if i == len(labels): return (0, 1, False)
        best, count, capped = rec(i + 1, used)
        for oid in sorted(adj[labels[i]]):
            bit = bits[oid]
            if used & bit: continue
            sub_best, sub_count, sub_cap = rec(i + 1, used | bit); value = 1 + sub_best
            if value > best: best, count, capped = value, sub_count, sub_cap
            elif value == best:
                count += sub_count; capped = capped or sub_cap or count >= cap
                if count >= cap: count = cap
        return best, count, capped
    _best, count, capped = rec(0, 0); return count, capped


def evaluate_state(labels: list[dict[str, str]], terms: list[dict[str, object]], pipe: Pipeline, pre: dict[str, tuple[tuple[str, tuple[str, ...]], ...]], state: State) -> dict[str, object]:
    adj, evidence = adjacency(labels, terms, pre, state.mapping)
    match = maximum_matching(adj); ambiguity_excess = sum(max(0, len(adj.get(r["stolfi_coordinate"], set())) - 1) for r in labels) / len(labels)
    metrics = mapping_metrics(state.mapping, pipe)
    return {"pipeline": pipe, "pre": pre, "mapping": state.mapping, "adjacency": adj, "evidence": evidence, "train_match": match,
            "train_coverage": len(match) / len(labels), "ambiguity_excess": ambiguity_excess,
            "score": base_score(len(match), len(labels), state.mapping, pipe, ambiguity_excess), **metrics}


def run_search(labels: list[dict[str, str]], terms: list[dict[str, object]], retain_details: bool = True) -> list[dict[str, object]]:
    representatives: dict[tuple, tuple[Pipeline, dict]] = {}
    for pipe in all_pipelines():
        pre = preprocess(terms, pipe); key = representation_key(pre)
        old = representatives.get(key)
        if old is None or (pipe.complexity, pipe.rule_string) < (old[0].complexity, old[0].rule_string): representatives[key] = (pipe, pre)
    raw = []
    for pipe, pre in sorted(representatives.values(), key=lambda x: x[0].pipeline_id):
        for state in search_pipeline(labels, terms, pipe, pre):
            raw.append(evaluate_state(labels, terms, pipe, pre, state))
    unique = {}
    for model in raw:
        key = (model["pipeline"].rule_string, model["mapping"])
        old = unique.get(key)
        rank = (-model["train_coverage"], -model["score"], model["mapping_size"], model["mapping"])
        if old is None or rank < old[0]: unique[key] = (rank, model)
    models = [x[1] for x in unique.values()]
    models.sort(key=lambda x: (-x["train_coverage"], -x["score"], x["mapping_size"], x["pipeline"].rule_string, x["mapping"]))
    if retain_details:
        return models[:RETAINED]
    best_score = dict(max(models, key=lambda x: (x["score"], x["train_coverage"], -x["mapping_size"])))
    best_score["search_max_train_coverage"] = max(x["train_coverage"] for x in models)
    return [best_score]


def seed_for(control: str, replicate: int) -> int:
    return int(hashlib.sha256(f"{SEED}|{control}|{replicate}".encode()).hexdigest()[:16], 16)


def letter_alphabets(terms: list[dict[str, object]]) -> dict[str, tuple[list[str], list[int]]]:
    result = {}
    for cls in ("STAR", "PLANET_MOON"):
        counts = Counter(ch for t in terms if t["object_class"] == cls for f in t["forms"] for ch in re.sub("[^a-z]", "", unicodedata.normalize("NFKD", str(f)).encode("ascii", "ignore").decode().lower()))
        result[cls] = (sorted(counts), [counts[x] for x in sorted(counts)])
    return result


def alter_terms(terms: list[dict[str, object]], rng: random.Random, pseudo: bool) -> list[dict[str, object]]:
    alphabets = letter_alphabets(terms); result = []
    for term in terms:
        forms = []; alphabet, weights = alphabets[str(term["object_class"])]
        for original in term["forms"]:  # type: ignore[index]
            ws = re.findall(r"[a-z]+", unicodedata.normalize("NFKD", str(original)).encode("ascii", "ignore").decode().lower())
            changed = []
            for word in ws:
                if pseudo: changed.append("".join(rng.choices(alphabet, weights=weights, k=len(word))))
                else:
                    chars = list(word); rng.shuffle(chars); changed.append("".join(chars))
            forms.append(" ".join(changed))
        result.append({"object_id": term["object_id"], "object_class": term["object_class"], "forms": forms})
    return result


def random_pool() -> dict[int, list[str]]:
    pool: dict[int, list[str]] = defaultdict(list)
    path = ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"
    for line in path.read_text(encoding="utf-8").splitlines():
        row = json.loads(line)
        if row["section"] != "A": continue
        value = "".join(EXPANSIONS.get(x, x) for x in row["token"].split("\x1f"))
        if re.fullmatch(r"[a-z]+", value): pool[len(segment(value, TARGET_COMPOSITES))].append(value)
    return dict(pool)


def randomize_labels(labels: list[dict[str, str]], rng: random.Random) -> list[dict[str, str]]:
    available = {k: v.copy() for k, v in random_pool().items()}; result = []
    for i, row in enumerate(labels):
        length = len(segment(row["voynich_token"], TARGET_COMPOSITES)); values = available[length]
        value = values.pop(rng.randrange(len(values)))
        result.append({**row, "stolfi_coordinate": f"NULL_{i:02d}", "voynich_token": value})
    return result


def null_worker(task: tuple[str, int]) -> dict[str, object]:
    control, replicate = task; rng = random.Random(seed_for(control, replicate)); labels = load_labels("TRAIN"); terms = load_terms()
    if control == "RANDOM_VOYNICH_SET": labels = randomize_labels(labels, rng)
    elif control == "SHUFFLED_TERMS": terms = alter_terms(terms, rng, False)
    elif control == "PSEUDODICTIONARY": terms = alter_terms(terms, rng, True)
    best = run_search(labels, terms, retain_details=False)[0]
    return {"control": control, "replicate": replicate, "seed": seed_for(control, replicate), "max_score": best["score"], "max_train_coverage": best["search_max_train_coverage"], "mapping_size": best["mapping_size"], "rule": best["pipeline"].rule_string}


def compute_heldout(model: dict[str, object], held: list[dict[str, str]], terms: list[dict[str, object]]) -> dict[str, object]:
    used = set(model["train_match"].values()); adj, evidence = adjacency(held, terms, model["pre"], model["mapping"], used); match = maximum_matching(adj)
    return {"held_adjacency": adj, "held_evidence": evidence, "held_match": match, "heldout_coverage": len(match) / len(held)}


def model_band(model: dict[str, object]) -> str:
    t, h, a, p = model["train_coverage"], model["heldout_coverage"], model["null_advantage"], model["empirical_p"]
    if t >= .70:
        if h >= .50 and a >= .10 and p <= .05 and model["mapping_size"] <= 18 and model["total_complexity"] <= 25: return "STRONG_CANDIDATE"
        return "OVERFIT"
    if t >= .40 and h >= .30 and a >= .05 and p <= .10 and model["mapping_size"] <= 16 and model["total_complexity"] <= 22: return "PARTIAL_CANDIDATE"
    return "NULL_COMPATIBLE"


def main() -> None:
    parser = argparse.ArgumentParser(); parser.add_argument("--workers", type=int, default=min(4, os.cpu_count() or 1)); parser.add_argument("--observed-only", action="store_true"); args = parser.parse_args()
    train, held, terms = load_labels("TRAIN"), load_labels("HELD_OUT"), load_terms()
    observed = run_search(train, terms, retain_details=True)
    if args.observed_only:
        for i, x in enumerate(observed[:10], 1): print(i, x["train_coverage"], x["score"], x["mapping_size"], x["pipeline"].rule_string)
        return
    tasks = [(control, rep) for control in ("RANDOM_VOYNICH_SET", "SHUFFLED_TERMS", "PSEUDODICTIONARY") for rep in range(NULL_REPLICATES)]
    if args.workers == 1: null_rows = [null_worker(t) for t in tasks]
    else:
        with get_context("fork").Pool(args.workers) as pool: null_rows = list(pool.imap(null_worker, tasks, chunksize=1))
    by_control = defaultdict(list)
    for row in null_rows: by_control[row["control"]].append(row)
    conservative_mean_cov = max(statistics.fmean(x["max_train_coverage"] for x in rows) for rows in by_control.values())
    for model in observed:
        model.update(compute_heldout(model, held, terms)); model["null_advantage"] = model["train_coverage"] - conservative_mean_cov
        ps = [(1 + sum(x["max_score"] >= model["score"] for x in rows)) / (len(rows) + 1) for rows in by_control.values()]
        model["empirical_p"] = max(ps); model["candidate_band"] = model_band(model)
    observed.sort(key=lambda x: (-x["train_coverage"], -x["score"], x["empirical_p"], -x["heldout_coverage"], x["mapping_size"], x["pipeline"].rule_string))
    retained = observed[:RETAINED]
    passing = [m for m in observed if m["candidate_band"] in {"STRONG_CANDIDATE", "PARTIAL_CANDIDATE", "OVERFIT"}]
    if passing:
        retained = []
        for m in passing + observed:
            if m not in retained: retained.append(m)
            if len(retained) == RETAINED: break
    for i, model in enumerate(retained, 1): model["model_id"] = f"M1_{i:03d}"

    model_rows = []
    assignment_rows = []; held_rows = []; unexplained_rows = []
    details = OUT / "M1_MODEL_DETAILS"; details.mkdir(exist_ok=True)
    for model in retained:
        adj = model["adjacency"]; maximum = len(model["train_match"]); multiplicity, capped = matching_multiplicity(adj)
        model["assignment_multiplicity"] = multiplicity; model["multiplicity_capped"] = capped
        both_classes = {r["object_class"] for r in held if r["stolfi_coordinate"] in model["held_match"]} == {"STAR", "PLANET_MOON"}
        mapping_text = ";".join(f"{s}->{'+'.join(t)}" for s, t in model["mapping"])
        model_rows.append({
            "model_id": model["model_id"], "candidate_band": model["candidate_band"], "train_coverage": fmt(model["train_coverage"]), "train_matched": len(model["train_match"]), "train_total": len(train),
            "heldout_coverage": fmt(model["heldout_coverage"]), "heldout_matched": len(model["held_match"]), "heldout_total": len(held), "pipeline_id": model["pipeline"].pipeline_id,
            "preprocessing_rules": model["pipeline"].rule_string, "substitution_table": mapping_text, "mapping_size": model["mapping_size"], "one_to_two_mappings": model["one_to_two"],
            "source_digraph_mappings": model["digraph_mappings"], "many_to_one_excess": model["many_to_one_excess"], "rule_complexity": model["pipeline"].complexity,
            "total_complexity": model["total_complexity"], "ambiguity_excess": fmt(model["ambiguity_excess"]), "assignment_multiplicity": multiplicity,
            "multiplicity_capped": "1" if capped else "0", "regularized_score": fmt(model["score"]), "null_advantage": fmt(model["null_advantage"]), "empirical_p": fmt(model["empirical_p"]),
            "heldout_both_classes": "1" if both_classes else "0", "unexplained_labels": len(train) + len(held) - len(model["train_match"]) - len(model["held_match"]),
        })
        for split_name, labels, a, match, evidence in (("TRAIN", train, adj, model["train_match"], model["evidence"]), ("HELD_OUT", held, model["held_adjacency"], model["held_match"], model["held_evidence"])):
            max_size = len(match)
            for label in sorted(labels, key=lambda x: x["stolfi_coordinate"]):
                lid = label["stolfi_coordinate"]
                if lid not in match:
                    unexplained_rows.append({"model_id": model["model_id"], "split": split_name, "label": lid, "family": label["stolfi_group"], "page": label["panel"], "token": label["voynich_token"], "reason_unmatched": "NO_COMPLETE_MAPPED_TERM" if not a.get(lid) else "ONE_TO_ONE_ASSIGNMENT_CONFLICT"})
                if split_name == "HELD_OUT":
                    held_rows.append({"model_id": model["model_id"], "label": lid, "object_class": label["object_class"], "page": label["panel"], "token": label["voynich_token"], "prediction": "MATCHED" if lid in match else "UNEXPLAINED", "canonical_object_id": match.get(lid, ""), "candidate_object_count": len(a.get(lid, set()))})
                for oid in sorted(a.get(lid, set())):
                    in_any = forced_edge_in_max(a, lid, oid, max_size)
                    ev = evidence[(lid, oid)]
                    assignment_rows.append({"model_id": model["model_id"], "split": split_name, "label": lid, "object_class": label["object_class"], "page": label["panel"], "token": label["voynich_token"], "object_id": oid,
                                            "source_forms": ";".join(sorted({x[0] for x in ev})), "source_unit_sequences": ";".join("+".join(x[1]) for x in ev), "canonical_assignment": "1" if match.get(lid) == oid else "0", "in_any_maximum_assignment": "1" if in_any else "0"})
        detail_lines = [f"# {model['model_id']}", "", f"Band: `{model['candidate_band']}`", "", f"TRAIN: {len(model['train_match'])}/{len(train)} ({model['train_coverage']:.6f})", "", f"HELD_OUT: {len(model['held_match'])}/{len(held)} ({model['heldout_coverage']:.6f})", "", f"Mapping size: {model['mapping_size']}", "", f"Total complexity: {model['total_complexity']}", "", f"Null advantage: {model['null_advantage']:.6f}", "", f"Empirical p: {model['empirical_p']:.6f}", "", f"Assignment multiplicity: {multiplicity}{' (capped)' if capped else ''}", "", "## Substitution table", "", "| Source | EVA output |", "|---|---|"]
        detail_lines += [f"| `{s}` | `{' + '.join(t)}` |" for s, t in model["mapping"]]
        detail_lines += ["", "## Canonical TRAIN assignment", ""] + [f"- `{lid}` ← `{oid}`" for lid, oid in sorted(model["train_match"].items())]
        detail_lines += ["", "## Canonical HELD_OUT assignment", ""] + ([f"- `{lid}` ← `{oid}`" for lid, oid in sorted(model["held_match"].items())] or ["No matches."])
        (details / f"{model['model_id']}.md").write_text("\n".join(detail_lines) + "\n", encoding="utf-8")

    write_tsv(OUT / "M1_TOKEN_FORMATION_MODELS.tsv", list(model_rows[0]), model_rows)
    assignment_fields = ["model_id", "split", "label", "object_class", "page", "token", "object_id", "source_forms", "source_unit_sequences", "canonical_assignment", "in_any_maximum_assignment"]
    write_tsv(OUT / "M1_LABEL_TERM_ASSIGNMENTS.tsv", assignment_fields, assignment_rows)
    write_tsv(OUT / "M1_HELDOUT_RESULTS.tsv", ["model_id", "label", "object_class", "page", "token", "prediction", "canonical_object_id", "candidate_object_count"], held_rows)
    write_tsv(OUT / "M1_UNEXPLAINED_LABELS.tsv", ["model_id", "split", "label", "family", "page", "token", "reason_unmatched"], unexplained_rows)
    null_out = [{**r, "max_score": fmt(r["max_score"]), "max_train_coverage": fmt(r["max_train_coverage"])} for r in sorted(null_rows, key=lambda x: (x["control"], x["replicate"]))]
    write_tsv(OUT / "M1_NULL_RESULTS.tsv", ["control", "replicate", "seed", "max_score", "max_train_coverage", "mapping_size", "rule"], null_out)

    best = retained[0]; strong = [m for m in observed if m["candidate_band"] == "STRONG_CANDIDATE"]; partial = [m for m in observed if m["candidate_band"] == "PARTIAL_CANDIDATE"]
    predictive = [m for m in strong if m["heldout_coverage"] >= 2/3 and {r["object_class"] for r in held if r["stolfi_coordinate"] in m["held_match"]} == {"STAR", "PLANET_MOON"}]
    status = "PREDICTIVE_MODEL_FOUND" if predictive else "STRONG_CANDIDATES_FOUND" if strong else "PARTIAL_CANDIDATES_FOUND" if partial else "NO_MODEL"
    comparison = f"""# M1 comparison with M0

M0's frozen maximum was 0/20 TRAIN and 0/6 HELD_OUT. M1 adds only a global
one-or-two-EVA-unit substitution table and obtains {len(best['train_match'])}/20
TRAIN ({best['train_coverage']:.6f}) and {len(best['held_match'])}/6 HELD_OUT
({best['heldout_coverage']:.6f}) for its top observed model. The raw TRAIN gain
over M0 is {best['train_coverage']:.6f}.

This gain cannot be interpreted without the search-level null. The most
conservative mean of the three null maximum-coverages is
{conservative_mean_cov:.6f}; the best model's advantage is
{best['null_advantage']:.6f}, and its conservative empirical p is
{best['empirical_p']:.6f}. Each null replicate reran the same bounded M1 search.
M0 artifacts were neither rewritten nor replaced.
"""
    (OUT / "M1_COMPARISON_WITH_M0.md").write_text(comparison, encoding="utf-8")
    summaries = []
    for control, rows in sorted(by_control.items()): summaries.append(f"- `{control}`: mean max coverage {statistics.fmean(x['max_train_coverage'] for x in rows):.6f}; p95 {sorted(x['max_train_coverage'] for x in rows)[94]:.6f}; max {max(x['max_train_coverage'] for x in rows):.6f}")
    report = f"""# M1 global grapheme-substitution search

## Result

`{status}`. The best bounded-search model matched {len(best['train_match'])}/20
TRAIN labels ({best['train_coverage']:.6f}) and {len(best['held_match'])}/6
HELD_OUT labels ({best['heldout_coverage']:.6f}). Its table has
{best['mapping_size']} source mappings and total complexity
{best['total_complexity']}.

## Search and controls

M1 retained the frozen M0 sample, corpus, 640 preprocessing pipelines, and
anonymous within-class one-to-one matching. The only new capacity is a single
global source-unit→one/two-EVA-unit table. The deterministic beam width is 64;
therefore this is a bounded optimisation result, not a proof of the global
optimum.

Each control used 100 independently seeded replicates and reran the complete
bounded optimiser:

{chr(10).join(summaries)}

The conservative null advantage is {best['null_advantage']:.6f}; empirical p is
{best['empirical_p']:.6f}. Assignment edges and multiplicity are reported
separately because maximum matching does not identify astronomical objects.

## Interpretation

A positive band means only that a compact global grapheme table is more
compatible with this anonymous vocabulary than the frozen controls. It is not
a translation, language identification, or star identification. Conversely a
negative band rejects only this inventory and bounded beam search. No mapping,
spelling, threshold, or HELD_OUT choice was added after TRAIN inspection.

## Final status

```text
ASTRO_TOKEN_FORMATION_M1={status}
MODELS_FOUND={len(strong) + len(partial)}
MODELS_WITH_TRAIN_COVERAGE_GE_70={sum(m['train_coverage'] >= .70 for m in observed)}
BEST_TRAIN_COVERAGE={best['train_coverage']:.6f}
BEST_HELDOUT_COVERAGE={best['heldout_coverage']:.6f}
BEST_MAPPING_SIZE={best['mapping_size']}
BEST_MODEL_COMPLEXITY={best['total_complexity']}
BEST_NULL_ADVANTAGE={best['null_advantage']:.6f}
BEST_EMPIRICAL_P={best['empirical_p']:.6f}
UNEXPLAINED_LABELS={len(train) + len(held) - len(best['train_match']) - len(best['held_match'])}
```
"""
    (OUT / "M1_TOKEN_FORMATION_REPORT.md").write_text(report, encoding="utf-8")

    artifacts = ["research/astro_token_formation_m1/M1_GRAPHEME_INVENTORY.tsv", "research/astro_token_formation_m1/M1_SUBSTITUTION_RULE_SPACE.md", "research/astro_token_formation_m1/M1_SEARCH_CONFIG.yaml", "research/astro_token_formation_m1/M1_TOKEN_FORMATION_MODELS.tsv", "research/astro_token_formation_m1/M1_LABEL_TERM_ASSIGNMENTS.tsv", "research/astro_token_formation_m1/M1_UNEXPLAINED_LABELS.tsv", "research/astro_token_formation_m1/M1_HELDOUT_RESULTS.tsv", "research/astro_token_formation_m1/M1_NULL_RESULTS.tsv", "research/astro_token_formation_m1/M1_COMPARISON_WITH_M0.md", "research/astro_token_formation_m1/M1_TOKEN_FORMATION_REPORT.md"] + [str(p.relative_to(ROOT)) for p in sorted(details.glob("*.md"))]
    inputs = ["research/astro_token_formation/ASTRO_TERM_CORPUS.tsv", "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv", "research/astro_token_formation/TOKEN_FORMATION_RULE_SPACE.md", "research/astro_token_formation/TOKEN_FORMATION_MODELS.tsv", "research/astro_token_formation_m1/M1_GRAPHEME_INVENTORY.tsv", "research/astro_token_formation_m1/M1_SUBSTITUTION_RULE_SPACE.md", "research/astro_token_formation_m1/M1_SEARCH_CONFIG.yaml", "research/astro_token_formation_m1/main.py", "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"]
    manifest = {"experiment": "astro-token-formation-m1-v1", "generated_utc": datetime.now(timezone.utc).replace(microsecond=0).isoformat(), "seed": SEED, "null_replicates_per_control": NULL_REPLICATES, "beam_width": BEAM_WIDTH, "status": status, "input_sha256": {p: sha(ROOT / p) for p in inputs}, "artifact_sha256": {p: sha(ROOT / p) for p in artifacts}}
    (OUT / "M1_MANIFEST.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    checksum_files = artifacts + ["M1_MANIFEST.json"]
    (OUT / "M1_SHA256SUMS").write_text("".join(f"{sha(ROOT / p)}  {p}\n" for p in sorted(checksum_files)), encoding="utf-8")


if __name__ == "__main__": main()
