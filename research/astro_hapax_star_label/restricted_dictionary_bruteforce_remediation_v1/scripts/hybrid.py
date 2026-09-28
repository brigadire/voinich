#!/usr/bin/env python3
"""
Real HYBRID orchestration: HeuristicSolver produces an incumbent, then a verifier explores a
formally-defined space and the outcome is classified. This is the orchestration v3_design's
SEARCH_ARCHITECTURE_DECISION.md claimed but never implemented (F009) -- the term HYBRID is only
used in this package because this file actually wires the two solvers together and only
reports a global certificate when a genuinely exhaustive/bounded verifier says so.

Verifier modes (see HYBRID_SEMANTICS below for the exact decision rule):
  EXACT_FULL_SPACE   -- CP-SAT over the entire table space (only within solver_cpsat's scope:
                        DROP_UNMAPPED, abbreviation NONE, single-char source alphabet).
  NEIGHBORHOOD       -- brute-force exhaustive enumeration of every valid table within
                        Hamming distance `radius` of the incumbent (any scope; tractable
                        because radius is small).
  SECOND_OPINION     -- an independent HeuristicSolver run (different seed family, different
                        move mix) used only when neither exact mode applies to this instance's
                        scope; this is NOT exhaustive and can only ever report NO_IMPROVEMENT_FOUND
                        or an improved incumbent, never a certificate.

HYBRID_SEMANTICS decision rule:
  1. EXACT_FULL_SPACE, solver status OPTIMAL                     -> GLOBAL_OPTIMUM_CERTIFIED
  2. EXACT_FULL_SPACE, status FEASIBLE with bound == incumbent    -> GLOBAL_OPTIMUM_CERTIFIED
  3. EXACT_FULL_SPACE, status FEASIBLE with bound > incumbent     -> BOUNDED_GAP_CERTIFIED
  4. EXACT_FULL_SPACE, status neither OPTIMAL nor FEASIBLE        -> VERIFICATION_TIMEOUT
  5. NEIGHBORHOOD, exhaustive, no table in radius beats incumbent -> LOCAL_NEIGHBORHOOD_OPTIMUM
  6. NEIGHBORHOOD, exhaustive, a table in radius beats incumbent  -> incumbent replaced by it;
                                                                      re-classified under rule 5/6
                                                                      recursively up to max_rounds,
                                                                      final round with no further
                                                                      improvement reports
                                                                      LOCAL_NEIGHBORHOOD_OPTIMUM
  7. SECOND_OPINION, independent heuristic finds nothing better   -> NO_IMPROVEMENT_FOUND
     (explicitly NOT a certificate of any kind -- disclosed as the weakest outcome)
  8. SECOND_OPINION finds something better                        -> incumbent replaced, and the
                                                                      *new* result is itself
                                                                      unverified (still NO_IMPROVEMENT_FOUND
                                                                      relative to whatever verification
                                                                      budget remains)
"""
from typing import Dict, List, Set, Tuple, Optional
import itertools
from scorer import Scorer, is_valid_table
from solver_heuristic import HeuristicSolver

try:
    from solver_cpsat import CPSATSolver, SUPPORTED_DELETION_MODES, SUPPORTED_ABBREVIATION
    _CPSAT_AVAILABLE = True
except Exception:
    _CPSAT_AVAILABLE = False


def _cpsat_scope_ok(source_alphabet, deletion_mode, abbreviation) -> bool:
    return (
        _CPSAT_AVAILABLE
        and all(len(c) == 1 for c in source_alphabet)
        and deletion_mode in SUPPORTED_DELETION_MODES
        and abbreviation in SUPPORTED_ABBREVIATION
    )


def _neighborhood_tables(incumbent: Dict[str, str], source_alphabet, target_alphabet, table_size, mapping_mode, radius: int):
    """Exhaustively yields every valid table reachable from incumbent by RETARGETING at most
    `radius` of its existing table_size source characters (the source-character set stays fixed;
    only which target symbol each already-chosen source maps to may change).

    This does NOT enumerate the full table space and filter by distance -- an earlier version did
    exactly that (itertools.combinations(all_sources, table_size) x itertools.product(target_
    alphabet, repeat=table_size), THEN checking `changed <= radius`), which is combinatorially
    catastrophic: at table_size=8 over a 16-symbol target alphabet that is 16**8 ~ 4.3 billion
    candidates generated before any filtering, regardless of how small `radius` is. That bug
    caused a real multi-hour hang during this package's own sealed-benchmark dev-grid run (see
    REMEDIATION_REPORT.md) before it was caught and fixed. This version instead constructs
    candidates directly: choose which <=radius of the table_size existing keys to retarget, and
    enumerate only their new target assignments -- C(table_size, r) * |target_alphabet|**r
    candidates per r, e.g. table_size=8, |target_alphabet|=16, radius=2: 1 + 128 + 7168 ~= 7300
    candidates total, not billions. Swapping which SOURCE characters are used is out of scope for
    this neighborhood definition (a disclosed restriction -- see HYBRID_SEMANTICS.md); it is not
    needed at the radius=1 operating point this package actually uses.
    """
    keys = sorted(incumbent.keys())
    for r in range(0, min(radius, len(keys)) + 1):
        for combo_keys in itertools.combinations(keys, r):
            for assignment in itertools.product(target_alphabet, repeat=r):
                table = dict(incumbent)
                for k, v in zip(combo_keys, assignment):
                    table[k] = v
                if is_valid_table(table, mapping_mode):
                    yield table


class HybridSolver:
    def __init__(
        self,
        source_alphabet: Tuple[str, ...],
        target_alphabet: Tuple[str, ...],
        table_size: int,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        capacity_policy: str = "PER_PAGE_CAPACITY_1",
        master_seed: int = 0,
        neighborhood_radius: int = 2,
        neighborhood_max_rounds: int = 3,
        verification_time_limit_sec: float = 20.0,
        heuristic_n_starts: int = 8,
        heuristic_max_iters: int = 400,
    ):
        self.source_alphabet = sorted(set(source_alphabet))
        self.target_alphabet = sorted(set(target_alphabet))
        self.table_size = table_size
        self.mapping_mode = mapping_mode
        self.deletion_mode = deletion_mode
        self.abbreviation = abbreviation
        self.capacity_policy = capacity_policy
        self.master_seed = master_seed
        self.neighborhood_radius = neighborhood_radius
        self.neighborhood_max_rounds = neighborhood_max_rounds
        self.verification_time_limit_sec = verification_time_limit_sec
        self.heuristic_n_starts = heuristic_n_starts
        self.heuristic_max_iters = heuristic_max_iters
        self.scorer = Scorer(mapping_mode, deletion_mode, abbreviation, capacity_policy)

    def solve(self, lexicon: Dict[str, Set[str]], labels: List[dict]) -> dict:
        heur = HeuristicSolver(
            self.source_alphabet, self.target_alphabet, self.table_size, self.mapping_mode,
            self.deletion_mode, self.abbreviation, self.capacity_policy,
            n_starts=self.heuristic_n_starts, max_iters=self.heuristic_max_iters, master_seed=self.master_seed,
        )
        incumbent_result = heur.solve(lexicon, labels)
        incumbent_table = incumbent_result["table"]
        incumbent_fitness = incumbent_result["fitness"]

        if _cpsat_scope_ok(self.source_alphabet, self.deletion_mode, self.abbreviation):
            verifier_mode = "EXACT_FULL_SPACE"
            cps = CPSATSolver(
                self.source_alphabet, self.target_alphabet, self.table_size, self.mapping_mode,
                self.capacity_policy, time_limit_sec=self.verification_time_limit_sec,
            )
            vr = cps.solve(lexicon, labels, self.deletion_mode, self.abbreviation)
            if vr["is_optimal"]:
                final_fitness = max(incumbent_fitness, vr["fitness"])
                final_table = vr["table"] if vr["fitness"] >= incumbent_fitness else incumbent_table
                classification = "GLOBAL_OPTIMUM_CERTIFIED"
                gap = 0
            elif vr["is_feasible_only"]:
                bound = vr["best_objective_bound"]
                best_seen = max(incumbent_fitness, vr["fitness"])
                gap = (bound - best_seen) if bound is not None else None
                final_fitness = best_seen
                final_table = vr["table"] if vr["fitness"] >= incumbent_fitness else incumbent_table
                classification = "GLOBAL_OPTIMUM_CERTIFIED" if gap == 0 else "BOUNDED_GAP_CERTIFIED"
            else:
                classification = "VERIFICATION_TIMEOUT"
                final_fitness = incumbent_fitness
                final_table = incumbent_table
                gap = None
            return {
                "algorithm": "HYBRID",
                "verifier_mode": verifier_mode,
                "incumbent_fitness": incumbent_fitness,
                "final_fitness": final_fitness,
                "final_table": final_table,
                "classification": classification,
                "optimality_gap": gap,
                "verifier_status": vr["status"],
            }

        # Fall back: exhaustive local-neighborhood verification (works for any scope)
        current_table, current_fitness = incumbent_table, incumbent_fitness
        rounds_used = 0
        for round_idx in range(self.neighborhood_max_rounds):
            rounds_used += 1
            best_in_nbhd_table, best_in_nbhd_fitness = current_table, current_fitness
            for cand in _neighborhood_tables(
                current_table, self.source_alphabet, self.target_alphabet, self.table_size,
                self.mapping_mode, self.neighborhood_radius,
            ):
                fit = self.scorer.evaluate(cand, lexicon, labels)["fitness"]
                if fit > best_in_nbhd_fitness:
                    best_in_nbhd_table, best_in_nbhd_fitness = cand, fit
            if best_in_nbhd_fitness <= current_fitness:
                return {
                    "algorithm": "HYBRID",
                    "verifier_mode": f"NEIGHBORHOOD(radius={self.neighborhood_radius})",
                    "incumbent_fitness": incumbent_fitness,
                    "final_fitness": current_fitness,
                    "final_table": current_table,
                    "classification": "LOCAL_NEIGHBORHOOD_OPTIMUM",
                    "optimality_gap": None,
                    "rounds_used": rounds_used,
                }
            current_table, current_fitness = best_in_nbhd_table, best_in_nbhd_fitness

        return {
            "algorithm": "HYBRID",
            "verifier_mode": f"NEIGHBORHOOD(radius={self.neighborhood_radius})",
            "incumbent_fitness": incumbent_fitness,
            "final_fitness": current_fitness,
            "final_table": current_table,
            "classification": "NO_IMPROVEMENT_FOUND",
            "optimality_gap": None,
            "rounds_used": rounds_used,
            "notes": "neighborhood search kept improving through neighborhood_max_rounds without "
                     "converging to a local optimum or reaching an exact verifier; reported as "
                     "NO_IMPROVEMENT_FOUND (weakest outcome) rather than fabricating a stronger claim.",
        }
