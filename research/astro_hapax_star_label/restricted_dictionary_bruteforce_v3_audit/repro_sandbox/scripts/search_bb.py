#!/usr/bin/env python3
"""
Branch-and-Bound search algorithm with provable upper bounds.
Guarantees global optimality on solvable instances by pruning provably suboptimal branches.
"""
from typing import Dict, List, Set, Tuple, Optional
import time
from collections import Counter
from engine import (
    evaluate_table,
    maximum_bipartite_matching,
    compute_complexity,
    encode_word,
    SOURCE_ALPHABET,
    EVA_ALPHABET
)

class BranchAndBoundSolver:
    def __init__(
        self,
        source_alphabet: Tuple[str, ...],
        target_alphabet: Tuple[str, ...],
        table_size: int = 4,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        capacity_policy: str = "PER_PAGE_CAPACITY_1",
        time_limit_sec: float = 60.0
    ):
        self.source_alphabet = sorted(source_alphabet)
        self.target_alphabet = sorted(target_alphabet)
        self.table_size = table_size
        self.mapping_mode = mapping_mode
        self.deletion_mode = deletion_mode
        self.abbreviation = abbreviation
        self.capacity_policy = capacity_policy
        self.time_limit_sec = time_limit_sec
        
        self.nodes_explored = 0
        self.branches_pruned = 0
        self.start_time = 0.0
        self.timed_out = False
        
        self.best_solution = None
        self.best_fitness = -float("inf")

    def _compute_upper_bound(
        self,
        partial_table: Dict[str, str],
        lexicon: Dict[str, Set[str]],
        labels: List[dict]
    ) -> int:
        """
        Admissible upper bound on maximum bipartite matching for any extension
        of partial_table to full table_size.
        """
        # A word W can potentially match label L if all characters of W already in partial_table
        # map to characters present in L, and the number of already mapped characters does not exceed |L|.
        relaxed_idx = {}
        for ident, forms in lexicon.items():
            for w in forms:
                # Check compatibility against labels
                clean_w = w.replace(" ", "")
                # Characters in W already fixed
                fixed_targets = [partial_table[c] for c in clean_w if c in partial_table]
                if len(fixed_targets) > 0:
                    # Potential key: fixed characters multiset
                    fixed_counts = Counter(fixed_targets)
                    for l in labels:
                        lbl_token = l["token"]
                        lbl_counts = Counter(lbl_token)
                        # Check if fixed targets are a subset of label characters
                        if all(lbl_counts[c] >= count for c, count in fixed_counts.items()):
                            relaxed_idx.setdefault(lbl_token, set()).add(ident)
                else:
                    # Completely unconstrained word can potentially match length-compatible label
                    for l in labels:
                        if len(l["token"]) <= len(clean_w):
                            relaxed_idx.setdefault(l["token"], set()).add(ident)
                            
        # Max matching on relaxed graph gives rigorous upper bound on matches
        relaxed_matches = len(maximum_bipartite_matching(labels, relaxed_idx, self.capacity_policy))
        return relaxed_matches

    def solve(
        self,
        lexicon: Dict[str, Set[str]],
        labels: List[dict]
    ) -> dict:
        self.start_time = time.time()
        self.nodes_explored = 0
        self.branches_pruned = 0
        self.timed_out = False
        self.best_fitness = -float("inf")
        self.best_solution = None
        
        # Initial incumbent: empty or simple heuristic seed
        eval_empty = evaluate_table(
            {}, lexicon, labels, self.mapping_mode, self.deletion_mode,
            self.abbreviation, self.capacity_policy
        )
        self.best_fitness = eval_empty["fitness"]
        self.best_solution = eval_empty
        
        # Branching order: frequent source letters first
        all_words = "".join(w for forms in lexicon.values() for w in forms)
        src_freq = Counter(all_words)
        ordered_sources = sorted(
            [s for s in self.source_alphabet if s in src_freq],
            key=lambda s: src_freq[s],
            reverse=True
        )
        # Append remaining
        for s in self.source_alphabet:
            if s not in ordered_sources:
                ordered_sources.append(s)

        def branch(src_idx: int, current_table: Dict[str, str], used_targets: List[str]):
            if self.timed_out:
                return
            if time.time() - self.start_time > self.time_limit_sec:
                self.timed_out = True
                return
                
            self.nodes_explored += 1
            
            # If current table has reached target size, evaluate as full solution
            if len(current_table) == self.table_size:
                res = evaluate_table(
                    current_table, lexicon, labels, self.mapping_mode,
                    self.deletion_mode, self.abbreviation, self.capacity_policy
                )
                if res["fitness"] > self.best_fitness:
                    self.best_fitness = res["fitness"]
                    self.best_solution = res
                return

            # Check if there are enough source characters remaining
            if src_idx >= len(ordered_sources):
                return
            remaining_needed = self.table_size - len(current_table)
            if (len(ordered_sources) - src_idx) < remaining_needed:
                return
                
            # Admissible upper bound check every step
            if len(current_table) >= 2:
                ub_matches = self._compute_upper_bound(current_table, lexicon, labels)
                min_complexity = compute_complexity(
                    dict(list(current_table.items()) + [("?", "?")] * remaining_needed),
                    self.mapping_mode, self.deletion_mode, self.abbreviation
                )
                max_potential_fitness = 100 * ub_matches - min_complexity
                if max_potential_fitness <= self.best_fitness:
                    self.branches_pruned += 1
                    return

            s_char = ordered_sources[src_idx]
            
            # Branch 1: map s_char to target character
            for t_char in self.target_alphabet:
                if self.mapping_mode == "INJECTIVE" and t_char in used_targets:
                    continue
                elif self.mapping_mode == "MERGE_1":
                    # At most one duplicate target allowed
                    if used_targets.count(t_char) >= 2:
                        continue
                    if used_targets.count(t_char) == 1 and any(used_targets.count(x) > 1 for x in set(used_targets)):
                        continue
                        
                current_table[s_char] = t_char
                used_targets.append(t_char)
                branch(src_idx + 1, current_table, used_targets)
                used_targets.pop()
                del current_table[s_char]
                
                if self.timed_out:
                    return

            # Branch 2: skip s_char (leave unmapped)
            branch(src_idx + 1, current_table, used_targets)

        branch(0, {}, [])
        
        elapsed = time.time() - self.start_time
        return {
            "algorithm": "BRANCH_AND_BOUND",
            "best_solution": self.best_solution,
            "best_fitness": self.best_fitness,
            "nodes_explored": self.nodes_explored,
            "branches_pruned": self.branches_pruned,
            "runtime_sec": round(elapsed, 4),
            "timed_out": self.timed_out,
            "global_optimum_guaranteed": not self.timed_out
        }
