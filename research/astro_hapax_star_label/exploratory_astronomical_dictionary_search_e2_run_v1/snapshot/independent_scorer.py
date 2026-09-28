#!/usr/bin/env python3
"""
Independent from-scratch reimplementation of the clean-room scorer, used ONLY to check
parity against scorer.Scorer on randomized small instances (SCORING_PARITY_RESULTS.tsv).

Deliberately uses different mechanisms than scorer.py wherever practical, so the two
implementations cannot share a bug:
- encode_word_v2: regex-compiled alternation instead of scorer.py's manual longest-prefix
  scan loop.
- matching_v2: networkx.algorithms.bipartite Hopcroft-Karp instead of scorer.py's manual
  Kuhn/DFS augmenting-path implementation.
This module is NOT used by any benchmark, generator, or production code path -- it exists
only inside run_scoring_parity.py.
"""
import re
from collections import defaultdict
from typing import Dict, List, Set

import networkx as nx


def encode_word_v2(word: str, table: Dict[str, str], deletion_mode: str, abbreviation: str) -> str:
    clean_w = word.replace(" ", "")
    if table:
        pattern = re.compile("|".join(re.escape(k) for k in sorted(table.keys(), key=len, reverse=True)))
    else:
        pattern = re.compile(r"(?!x)x")  # matches nothing

    out = []
    i = 0
    n = len(clean_w)
    while i < n:
        m = pattern.match(clean_w, i)
        if m and m.end() > i:
            out.append(table[m.group(0)])
            i = m.end()
            continue
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


def compute_complexity_v2(table: Dict[str, str], mapping_mode: str, deletion_mode: str, abbreviation: str) -> int:
    cost = len(table)
    cost += {"MERGE_1": 1}.get(mapping_mode, 0)
    cost += {"DROP_UNMAPPED": 2, "SELECTIVE_VOWEL_DROP": 1}.get(deletion_mode, 0)
    cost += {"SUSPENSION_1": 1, "SUSPENSION_2": 2, "PREFIX_4": 2}.get(abbreviation, 0)
    return cost


def is_valid_table_v2(table: Dict[str, str], mapping_mode: str) -> bool:
    counts = defaultdict(int)
    for t in table.values():
        counts[t] += 1
    if mapping_mode == "INJECTIVE":
        return max(counts.values(), default=1) <= 1
    if mapping_mode == "MERGE_1":
        return sum(1 for c in counts.values() if c == 2) <= 1 and all(c <= 2 for c in counts.values())
    raise ValueError(mapping_mode)


def matching_v2(
    labels: List[dict],
    lexicon_index: Dict[str, Set[str]],
    capacity_policy: str,
) -> Dict[str, str]:
    def match_page(page_labels: List[dict]) -> Dict[str, str]:
        g = nx.Graph()
        left = [f"L::{l['occurrence_id']}" for l in page_labels]
        g.add_nodes_from(left, bipartite=0)
        right_seen = set()
        for l in page_labels:
            ln = f"L::{l['occurrence_id']}"
            for star in sorted(lexicon_index.get(l["token"], set())):
                rn = f"R::{star}"
                right_seen.add(rn)
                g.add_edge(ln, rn)
        g.add_nodes_from(right_seen, bipartite=1)
        if g.number_of_edges() == 0:
            return {}
        matching = nx.algorithms.bipartite.matching.hopcroft_karp_matching(g, top_nodes=left)
        out = {}
        for a, b in matching.items():
            if a.startswith("L::"):
                out[a[3:]] = b[3:]
        return out

    sorted_labels = sorted(labels, key=lambda l: (l.get("page_id", ""), l["occurrence_id"]))
    if capacity_policy == "PER_PAGE_CAPACITY_1":
        result: Dict[str, str] = {}
        pages = sorted(set(l.get("page_id", "default") for l in sorted_labels))
        for p in pages:
            page_labels = [l for l in sorted_labels if l.get("page_id", "default") == p]
            result.update(match_page(page_labels))
        return result
    else:
        return match_page(sorted_labels)


def evaluate_v2(
    table: Dict[str, str],
    lexicon: Dict[str, Set[str]],
    labels: List[dict],
    mapping_mode: str,
    deletion_mode: str,
    abbreviation: str,
    capacity_policy: str,
) -> dict:
    valid = is_valid_table_v2(table, mapping_mode)
    idx: Dict[str, Set[str]] = defaultdict(set)
    for ident, forms in lexicon.items():
        for form in forms:
            enc = encode_word_v2(form, table, deletion_mode, abbreviation)
            if enc:
                idx[enc].add(ident)
    assignments = matching_v2(labels, idx, capacity_policy)
    matched = len(assignments)
    complexity = compute_complexity_v2(table, mapping_mode, deletion_mode, abbreviation)
    fitness = 100 * matched - complexity if valid else -10**9
    return {
        "valid_table": valid,
        "matched": matched,
        "total": len(labels),
        "complexity": complexity,
        "fitness": fitness,
        "assignments": assignments,
    }
