#!/usr/bin/env python3
"""
Exact branch-and-bound oracle. Reused/revalidated from
restricted_dictionary_bruteforce_v3_design/scripts/search_bb.py per
COMPONENT_REUSE_REGISTRY.tsv (BRANCH_AND_BOUND, REUSABLE_AFTER_REVALIDATION).

Ported to call scorer.py instead of the old engine.py, and to expose exactness
guarantees explicitly:
- admissible upper bound (relaxed bipartite matching ignoring unfixed source graphemes)
- explicit time_limit_sec and timed_out flag
- global_optimum_certified is True ONLY when the full search tree was exhausted
  (never when timed_out) -- this is the specific property v3_design's search_cp.py
  violated (F003) and this oracle must not repeat.
"""
from typing import Dict, List, Set, Tuple
import time
from collections import Counter
from scorer import Scorer, maximum_bipartite_matching, compute_complexity, is_valid_table


class BranchAndBoundOracle:
    def __init__(
        self,
        source_alphabet: Tuple[str, ...],
        target_alphabet: Tuple[str, ...],
        table_size: int,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        capacity_policy: str = "PER_PAGE_CAPACITY_1",
        time_limit_sec: float = 30.0,
    ):
        self.source_alphabet = sorted(source_alphabet)
        self.target_alphabet = sorted(target_alphabet)
        self.table_size = table_size
        self.mapping_mode = mapping_mode
        self.deletion_mode = deletion_mode
        self.abbreviation = abbreviation
        self.capacity_policy = capacity_policy
        self.time_limit_sec = time_limit_sec
        self.scorer = Scorer(mapping_mode, deletion_mode, abbreviation, capacity_policy)

    def _upper_bound(self, partial_table: Dict[str, str], lexicon: Dict[str, Set[str]], labels: List[dict]) -> int:
        relaxed_idx: Dict[str, Set[str]] = {}
        for ident, forms in lexicon.items():
            for w in forms:
                clean_w = w.replace(" ", "")
                fixed_targets = [partial_table[c] for c in clean_w if c in partial_table]
                if fixed_targets:
                    fixed_counts = Counter(fixed_targets)
                    for l in labels:
                        lbl_counts = Counter(l["token"])
                        if all(lbl_counts[c] >= cnt for c, cnt in fixed_counts.items()):
                            relaxed_idx.setdefault(l["token"], set()).add(ident)
                else:
                    for l in labels:
                        if len(l["token"]) <= len(clean_w):
                            relaxed_idx.setdefault(l["token"], set()).add(ident)
        return len(maximum_bipartite_matching(labels, relaxed_idx, self.capacity_policy))

    def solve(self, lexicon: Dict[str, Set[str]], labels: List[dict]) -> dict:
        start = time.time()
        nodes_explored = 0
        branches_pruned = 0
        timed_out = False

        # NOTE: the incumbent must NOT be seeded from the empty table. table_size is an exact
        # constraint (every solver in this package -- CP-SAT, heuristic, brute-force validator --
        # searches only tables with exactly table_size entries); seeding from the empty table's
        # score let this oracle silently "win" by comparing against a table of a different size
        # than the one actually requested, which produced 10/100 disagreements against CP-SAT on
        # instances where every full-size table scored worse than the empty one (see
        # CPSAT_ORACLE_PARITY.tsv trials 1,3,10,30,41,61,67,83 before this fix). This is the same
        # seeding pattern v3_design's search_bb.py used; carrying it forward here would have been
        # exactly the kind of un-revalidated reuse COMPONENT_REUSE_REGISTRY.tsv exists to prevent.
        best_fitness = -float("inf")
        best_solution = None

        all_words = "".join(w for forms in lexicon.values() for w in forms)
        src_freq = Counter(all_words)
        ordered_sources = sorted(
            [s for s in self.source_alphabet if s in src_freq], key=lambda s: src_freq[s], reverse=True
        )
        for s in self.source_alphabet:
            if s not in ordered_sources:
                ordered_sources.append(s)

        state = {"nodes": 0, "pruned": 0, "timed_out": False, "best_fitness": best_fitness, "best_solution": best_solution}

        def branch(src_idx: int, current_table: Dict[str, str], used_targets: List[str]):
            if state["timed_out"]:
                return
            if time.time() - start > self.time_limit_sec:
                state["timed_out"] = True
                return
            state["nodes"] += 1

            if len(current_table) == self.table_size:
                res = self.scorer.evaluate(current_table, lexicon, labels)
                if res["fitness"] > state["best_fitness"]:
                    state["best_fitness"] = res["fitness"]
                    state["best_solution"] = res
                return

            if src_idx >= len(ordered_sources):
                return
            remaining_needed = self.table_size - len(current_table)
            if (len(ordered_sources) - src_idx) < remaining_needed:
                return

            if len(current_table) >= 2:
                ub_matches = self._upper_bound(current_table, lexicon, labels)
                min_complexity = compute_complexity(
                    dict(list(current_table.items()) + [(f"?{k}", "?") for k in range(remaining_needed)]),
                    self.mapping_mode, self.deletion_mode, self.abbreviation,
                )
                if 100 * ub_matches - min_complexity <= state["best_fitness"]:
                    state["pruned"] += 1
                    return

            s_char = ordered_sources[src_idx]
            for t_char in self.target_alphabet:
                current_table[s_char] = t_char
                used_targets.append(t_char)
                if is_valid_table(current_table, self.mapping_mode) or len(current_table) < self.table_size:
                    # allow partial invalidity only if it can still resolve by table completion;
                    # for MERGE_1 bound this needs a relaxed partial check:
                    if self._partial_ok(current_table):
                        branch(src_idx + 1, current_table, used_targets)
                used_targets.pop()
                del current_table[s_char]
                if state["timed_out"]:
                    return
            branch(src_idx + 1, current_table, used_targets)

        branch(0, {}, [])

        elapsed = time.time() - start
        return {
            "algorithm": "BRANCH_AND_BOUND_ORACLE",
            "best_solution": state["best_solution"],
            "best_fitness": state["best_fitness"],
            "nodes_explored": state["nodes"],
            "branches_pruned": state["pruned"],
            "runtime_sec": round(elapsed, 4),
            "timed_out": state["timed_out"],
            "global_optimum_certified": not state["timed_out"],
        }

    def _partial_ok(self, table: Dict[str, str]) -> bool:
        from collections import defaultdict
        counts = defaultdict(int)
        for t in table.values():
            counts[t] += 1
        if self.mapping_mode == "INJECTIVE":
            return all(c <= 1 for c in counts.values())
        if self.mapping_mode == "MERGE_1":
            return all(c <= 2 for c in counts.values()) and sum(1 for c in counts.values() if c == 2) <= 1
        return True
