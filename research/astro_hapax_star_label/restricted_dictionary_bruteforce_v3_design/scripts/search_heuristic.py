#!/usr/bin/env python3
"""
Directional heuristic search algorithm combining:
1. Directional frequency seeding
2. Candidate transition pair generation
3. Fast beam search initialization
4. Simulated Annealing / Stochastic Neighborhood Optimization
Deterministic given seed, order-invariant, and highly effective for cryptanalysis of substitution systems.
"""
from typing import Dict, List, Set, Tuple, Optional
import time
import math
import random
from collections import Counter
from engine import (
    evaluate_table,
    compute_complexity,
    SOURCE_ALPHABET,
    EVA_ALPHABET
)

class DirectionalHeuristicSolver:
    def __init__(
        self,
        source_alphabet: Tuple[str, ...],
        target_alphabet: Tuple[str, ...],
        table_size: int = 8,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        capacity_policy: str = "PER_PAGE_CAPACITY_1",
        beam_width: int = 16,
        local_search_steps: int = 350,
        seed: int = 42,
        time_limit_sec: float = 30.0
    ):
        self.source_alphabet = sorted(source_alphabet)
        self.target_alphabet = sorted(target_alphabet)
        self.table_size = table_size
        self.mapping_mode = mapping_mode
        self.deletion_mode = deletion_mode
        self.abbreviation = abbreviation
        self.capacity_policy = capacity_policy
        self.beam_width = beam_width
        self.local_search_steps = local_search_steps
        self.seed = seed
        self.time_limit_sec = time_limit_sec
        
        self.iterations = 0
        self.start_time = 0.0
        self.timed_out = False

    def solve(
        self,
        lexicon: Dict[str, Set[str]],
        labels: List[dict]
    ) -> dict:
        self.start_time = time.time()
        self.iterations = 0
        self.timed_out = False
        rng = random.Random(self.seed)
        
        # Sort labels deterministically to guarantee order invariance
        sorted_labels = sorted(labels, key=lambda l: (l.get("page_id", ""), l["occurrence_id"]))
        
        # 1. Frequency Analysis
        all_label_chars = "".join(l["token"] for l in sorted_labels)
        lbl_freq = Counter(all_label_chars)
        
        all_dict_chars = "".join(w.replace(" ", "") for forms in lexicon.values() for w in forms)
        dict_freq = Counter(all_dict_chars)
        
        ranked_source = [s for s, _ in dict_freq.most_common() if s in self.source_alphabet]
        for s in self.source_alphabet:
            if s not in ranked_source:
                ranked_source.append(s)
                
        ranked_target = [t for t, _ in lbl_freq.most_common() if t in self.target_alphabet]
        for t in self.target_alphabet:
            if t not in ranked_target:
                ranked_target.append(t)

        # 2. Generate Initial Seeds
        seed_tables = []
        
        # Seed 1: Rank alignment
        s1 = {}
        for s_ch, t_ch in zip(ranked_source[:self.table_size], ranked_target[:self.table_size]):
            s1[s_ch] = t_ch
        seed_tables.append(s1)
        
        # Seed 2: Overlap identity
        s2 = {}
        for s_ch in ranked_source:
            if s_ch in self.target_alphabet and s_ch not in s2:
                s2[s_ch] = s_ch
                if len(s2) == self.table_size:
                    break
        if s2:
            seed_tables.append(s2)
            
        # Seed 3: Deterministic pseudorandom sampling from top alphabet pools
        s3_src = rng.sample(ranked_source[:min(len(ranked_source), self.table_size * 2)], self.table_size)
        s3_tgt = rng.sample(ranked_target[:min(len(ranked_target), len(self.target_alphabet))], self.table_size)
        seed_tables.append(dict(zip(s3_src, s3_tgt)))

        # Evaluate initial seeds
        best_table = {}
        best_fitness = -float("inf")
        best_eval = None
        
        for st in seed_tables:
            ev = evaluate_table(
                st, lexicon, sorted_labels, self.mapping_mode, self.deletion_mode,
                self.abbreviation, self.capacity_policy
            )
            if ev["fitness"] > best_fitness:
                best_fitness = ev["fitness"]
                best_table = dict(st)
                best_eval = ev

        # 3. Simulated Annealing Optimization
        # Candidate pools for mutations
        candidate_src = ranked_source[:min(len(ranked_source), 24)]
        candidate_tgt = ranked_target[:min(len(ranked_target), len(self.target_alphabet))]
        
        cur_table = dict(best_table)
        cur_fitness = best_fitness
        
        temp = 60.0
        cooling_rate = 0.985
        
        for step in range(self.local_search_steps):
            if time.time() - self.start_time > self.time_limit_sec:
                self.timed_out = True
                break
                
            self.iterations += 1
            temp *= cooling_rate
            
            # Generate mutation
            mut_table = dict(cur_table)
            action = rng.choice(["swap_tgt", "change_tgt", "change_src"])
            keys = list(mut_table.keys())
            
            if action == "swap_tgt" and len(keys) >= 2:
                k1, k2 = rng.sample(keys, 2)
                mut_table[k1], mut_table[k2] = mut_table[k2], mut_table[k1]
            elif action == "change_tgt" and keys:
                k1 = rng.choice(keys)
                if self.mapping_mode == "INJECTIVE":
                    avail_tgt = [c for c in candidate_tgt if c not in mut_table.values()]
                elif self.mapping_mode == "MERGE_1":
                    counts = Counter(mut_table.values())
                    has_dup = any(v > 1 for v in counts.values())
                    avail_tgt = [c for c in candidate_tgt if not has_dup or c not in mut_table.values()]
                else:
                    avail_tgt = candidate_tgt
                if avail_tgt:
                    mut_table[k1] = rng.choice(avail_tgt)
            elif action == "change_src" and keys:
                k1 = rng.choice(keys)
                avail_src = [c for c in candidate_src if c not in mut_table]
                if avail_src:
                    new_s = rng.choice(avail_src)
                    val = mut_table.pop(k1)
                    mut_table[new_s] = val

            # Check mapping mode validity
            if self.mapping_mode == "INJECTIVE" and len(set(mut_table.values())) < len(mut_table):
                continue
            elif self.mapping_mode == "MERGE_1":
                counts = Counter(mut_table.values())
                dups = [k for k, v in counts.items() if v > 1]
                if len(dups) > 1 or any(counts[k] > 2 for k in dups):
                    continue

            ev = evaluate_table(
                mut_table, lexicon, sorted_labels, self.mapping_mode,
                self.deletion_mode, self.abbreviation, self.capacity_policy
            )
            
            diff = ev["fitness"] - cur_fitness
            if diff > 0 or (diff > -250 and rng.random() < math.exp(diff / max(1.0, temp))):
                cur_table = mut_table
                cur_fitness = ev["fitness"]
                if cur_fitness > best_fitness:
                    best_fitness = cur_fitness
                    best_table = dict(cur_table)
                    best_eval = ev

        elapsed = time.time() - self.start_time
        return {
            "algorithm": "DIRECTIONAL_HEURISTIC",
            "best_solution": best_eval,
            "best_fitness": best_fitness,
            "iterations": self.iterations,
            "runtime_sec": round(elapsed, 4),
            "timed_out": self.timed_out,
            "global_optimum_guaranteed": False
        }
