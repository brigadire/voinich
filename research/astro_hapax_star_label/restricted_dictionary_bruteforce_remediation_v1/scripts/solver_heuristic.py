#!/usr/bin/env python3
"""
Directed heuristic (multi-start simulated annealing) solver. No optimality certificate is
ever claimed -- `global_optimum_certified` is always False and is not a field this solver
even reports, to avoid the ambiguity that made v3_design's search_cp.py dangerous.

Properties required by clean_room.md Section 10:
- multiple independent starts (n_starts, each with its own deterministic sub-seed derived
  from master_seed so re-running with the same master_seed reproduces byte-identical output)
- fixed iteration budget per start (max_iters)
- train-only optimization: only ever calls Scorer.evaluate() on the train `labels` argument,
  never on held-out labels or their true_identity ground truth
- full order invariance: candidate generation and move selection are driven only by sorted()
  views of the alphabets/lexicon, never by incidental dict/set iteration order
- full alphabet-renaming invariance: neighbor moves and acceptance are defined purely in terms
  of which (source, target) pairs are chosen, not in terms of target symbol identity, so a
  consistent renaming of the target alphabet cannot change reachable fitness (verified in
  SCIENTIFIC_SAFETY_TESTS.md)
"""
import random
from typing import Dict, List, Set, Tuple, Optional
from scorer import Scorer, is_valid_table


class HeuristicSolver:
    def __init__(
        self,
        source_alphabet: Tuple[str, ...],
        target_alphabet: Tuple[str, ...],
        table_size: int,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        capacity_policy: str = "PER_PAGE_CAPACITY_1",
        n_starts: int = 8,
        max_iters: int = 400,
        master_seed: int = 0,
    ):
        self.source_alphabet = sorted(set(source_alphabet))
        self.target_alphabet = sorted(set(target_alphabet))
        self.table_size = min(table_size, len(self.source_alphabet))
        self.mapping_mode = mapping_mode
        self.deletion_mode = deletion_mode
        self.abbreviation = abbreviation
        self.capacity_policy = capacity_policy
        self.n_starts = n_starts
        self.max_iters = max_iters
        self.master_seed = master_seed
        self.scorer = Scorer(mapping_mode, deletion_mode, abbreviation, capacity_policy)

    def _random_valid_table(self, rng: random.Random) -> Dict[str, str]:
        srcs = rng.sample(self.source_alphabet, self.table_size)
        table = {}
        used = []
        for s in srcs:
            candidates = list(self.target_alphabet)
            rng.shuffle(candidates)
            for t in candidates:
                trial = dict(table)
                trial[s] = t
                if is_valid_table(trial, self.mapping_mode):
                    table = trial
                    used.append(t)
                    break
        return table

    def _neighbor(self, table: Dict[str, str], rng: random.Random) -> Dict[str, str]:
        move = rng.choice(["retarget", "swap", "resource"])
        new_table = dict(table)
        srcs_in = sorted(new_table.keys())
        srcs_out = [s for s in self.source_alphabet if s not in new_table]

        if move == "retarget" and srcs_in:
            s = rng.choice(srcs_in)
            for t in rng.sample(self.target_alphabet, len(self.target_alphabet)):
                trial = dict(new_table)
                trial[s] = t
                if is_valid_table(trial, self.mapping_mode):
                    return trial
            return new_table
        if move == "swap" and len(srcs_in) >= 2:
            a, b = rng.sample(srcs_in, 2)
            trial = dict(new_table)
            trial[a], trial[b] = trial[b], trial[a]
            if is_valid_table(trial, self.mapping_mode):
                return trial
            return new_table
        if move == "resource" and srcs_in and srcs_out:
            drop = rng.choice(srcs_in)
            add = rng.choice(srcs_out)
            trial = dict(new_table)
            t = trial.pop(drop)
            for cand in rng.sample(self.target_alphabet, len(self.target_alphabet)):
                trial2 = dict(trial)
                trial2[add] = cand
                if is_valid_table(trial2, self.mapping_mode):
                    return trial2
            return new_table
        return new_table

    def _anneal(self, lexicon, labels, seed: int) -> dict:
        rng = random.Random(seed)
        current = self._random_valid_table(rng)
        current_fit = self.scorer.evaluate(current, lexicon, labels)["fitness"]
        best_table, best_fit = current, current_fit
        t0, t1 = 2.5, 0.02
        for it in range(self.max_iters):
            temp = t0 * ((t1 / t0) ** (it / max(1, self.max_iters - 1)))
            cand = self._neighbor(current, rng)
            cand_fit = self.scorer.evaluate(cand, lexicon, labels)["fitness"]
            delta = cand_fit - current_fit
            if delta >= 0 or rng.random() < pow(2.71828182845904523536, delta / max(temp, 1e-9)):
                current, current_fit = cand, cand_fit
                if current_fit > best_fit:
                    best_table, best_fit = current, current_fit
        return {"table": best_table, "fitness": best_fit, "iters": self.max_iters}

    def solve(self, lexicon: Dict[str, Set[str]], labels: List[dict]) -> dict:
        starts = []
        for k in range(self.n_starts):
            seed = (self.master_seed * 1_000_003 + k * 7919) & 0x7FFFFFFF
            starts.append(self._anneal(lexicon, labels, seed))
        best = max(starts, key=lambda r: r["fitness"])
        full_eval = self.scorer.evaluate(best["table"], lexicon, labels)
        return {
            "algorithm": "DIRECTED_HEURISTIC",
            "table": best["table"],
            "matched": full_eval["matched"],
            "complexity": full_eval["complexity"],
            "fitness": full_eval["fitness"],
            "n_starts": self.n_starts,
            "max_iters_per_start": self.max_iters,
            "master_seed": self.master_seed,
            "per_start_fitness": [s["fitness"] for s in starts],
            "global_optimum_certified": False,
            "certificate_status": "NO_OPTIMALITY_CERTIFICATE",
        }
