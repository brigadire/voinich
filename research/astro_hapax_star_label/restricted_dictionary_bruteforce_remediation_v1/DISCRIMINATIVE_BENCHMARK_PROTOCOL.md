# Discriminative Benchmark Protocol

Replaces v3_design's `EXACT_SMALL_INSTANCE_RESULTS.tsv` (`COMPONENT_REUSE_REGISTRY.tsv`:
`EXACT_SMALL_INSTANCE_BENCHMARK`, `REJECTED` — Finding F004: every one of its instances had 2-10
tied co-optimal tables, so a random baseline reached the certified optimum on 25/25 instances and
the benchmark discriminated nothing).

## Design rule

Every case in `scripts/discriminative_benchmark.py` is **hand-designed, not randomly sampled**, and
is small enough (`table_size<=6`, source alphabet `<=8`, target alphabet `<=7`) that the entire
table space is brute-force enumerable. This is what makes `certified_optimum_score`,
`equivalence_class_size`, and `truth_rank` in `DISCRIMINATIVE_BENCHMARK_RESULTS.tsv` genuinely
computed values, not estimates. Each case either has a **unique optimum** (equivalence class size
1) or an **explicitly documented equivalence class** (Case 10) — never an accidental multi-way tie
mistaken for a unique optimum, which is exactly what went undetected in v3.

## Cases and the properties each targets

| case_id | targets |
|---|---|
| `C01_UNIQUE_OPTIMUM_BASELINE` | positive control: unique optimum, no adversarial structure |
| `C02_FALSE_FREQUENCY_TRAP` | false frequency correspondence; truth is deliberately **not** the score optimum (decoy identities whose repeated single character coincidentally covers more distractor labels than the true generative table) |
| `C03_MERGE_AMBIGUITY` | bounded `MERGE_1` merge ambiguity |
| `C04_SELECTIVE_DELETION` | `SELECTIVE_VOWEL_DROP` |
| `C05_ABBREVIATION_SUSPENSION` | `SUSPENSION_1` truncation |
| `C06_FULL_COMPOSITION` | merge + selective deletion + abbreviation together |
| `C07_HIGH_UNMATCHED_DISTRACTOR_DENSE` | high unmatched rate, dense distractors |
| `C08_LOW_DISTRACTOR_DENSITY` | contrast case: sparse distractors |
| `C09_NEAR_TIE_NOT_EQUAL` | several close-but-not-equal-scoring tables (tests whether a solver can actually distinguish near-ties, not just get close) |
| `C10_EXPLICIT_EQUIVALENCE_CLASS` | genuine, disclosed multi-member equivalence class (unused source/target slots are interchangeable by construction) |

## Metrics (per case, per algorithm)

`certified_optimum_score` (from brute-force enumeration), `equivalence_class_size`, `truth_score`,
`truth_rank` (1-indexed, ties share rank 1), `truth_is_optimum`, `algo_fitness`, `algo_matched`,
`reached_optimum_score`, `certified` (did *this run* produce a certificate), `timed_out`,
`optimality_gap`.

## Algorithms compared

`BRANCH_AND_BOUND` (exact oracle, any scope), `CP_SAT` (exact, only within `solver_cpsat.py`'s
disclosed scope — reported as `OUT_OF_SCOPE_*` rather than silently skipped when a case falls
outside it), `DIRECTED_HEURISTIC` (multi-start simulated annealing, no certificate),
`HYBRID` (heuristic + verifier, classified per `HYBRID_SEMANTICS.md`), and
`RANDOM_BASELINE_SINGLE_DRAW_MEAN_OF_30`.

**The random baseline is the mean over 30 independent single random draws, not the best of 30.**
An earlier draft of this benchmark used best-of-30 and found it reached the certified optimum on
8/10 cases (`0.800`) purely from having 30 independent tries at small search spaces (24-58,800
tables) — a "best of N" random search is a weak search procedure, not a random baseline, and
reusing it here would have repeated exactly the kind of nearest-of-N-vs-single-draw bias this
project has previously caught and fixed in unrelated null-model work (task65). Switched to the
mean (and a disclosed hit-rate) over 30 independent single draws before any result was reported.

## Result

`DISCRIMINATIVE_BENCHMARK_RESULTS.tsv`: `BRANCH_AND_BOUND`, `CP_SAT` (where in scope), the
`DIRECTED_HEURISTIC`, and `HYBRID` reach the certified optimum on 10/10 cases (`CP_SAT` reports
`OUT_OF_SCOPE` rather than a false result on the 3 cases using `SELECTIVE_VOWEL_DROP` or
abbreviation, its disclosed scope restriction). `RANDOM_BASELINE_SINGLE_DRAW_MEAN_OF_30` never
reaches the certified optimum on any case (hit rate 0/30 draws on every case).
`DISCRIMINATIVE_BENCHMARK=PASS`, `RANDOM_BASELINE_SIGNIFICANTLY_WORSE=YES`.

Reproduce with: `python3 scripts/discriminative_benchmark.py DISCRIMINATIVE_BENCHMARK_RESULTS.tsv`
