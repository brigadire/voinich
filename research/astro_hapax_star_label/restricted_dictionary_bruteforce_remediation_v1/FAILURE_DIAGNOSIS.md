# Failure Diagnosis

This file was rewritten after an external methodological review found the first version's
conclusion overclaimed. That review, the resulting fixes, and this version's corrected numbers
are all part of the record — see the amendment log in `SEALED_BENCHMARK_PROTOCOL.md` and
`VALIDATION_REPORT.md`. **This is now the current, correct diagnosis** — see "History of this
diagnosis" at the bottom for what changed and why.

## Current gate status (fresh sealed hidden data, 60s CP-SAT budget, table_size=8 reference point)

```text
FRESH_ZERO_NOISE_RECOVERY                      = 1.000  (gate >=0.80)   PASS
FRESH_TEN_PERCENT_NOISE_RECOVERY                = 0.933  (gate >=0.70)   PASS
FRESH_TEN_PERCENT_MAPPING_PRECISION             = 0.969  (gate >=0.75)   PASS
FRESH_TEN_PERCENT_MAPPING_RECALL                = 0.814  (gate >=0.75)   PASS
FRESH_TEN_PERCENT_HELDOUT_ASSIGNMENT_ACCURACY   = 0.589  (gate >=0.60)   FAIL (narrow miss)
```

4 of 5 pre-registered gates pass on fresh sealed data. `WORST_REQUIRED_MODE=FAIL` for the **full**
tested matrix regardless (see below — this is not close to passing for every mode), but the
picture is now sharply bimodal along a structural line, not a diffuse failure and not a resource
question for the whole model class.

## The two populations, and why they must be reported separately

`HIDDEN_RESULTS.tsv`'s 180 rows split cleanly along a line fixed by `solver_cpsat.py`'s disclosed
scope, not by anything discovered post hoc:

| Population | n | Modes | Equivalence-aware recovery rate | Can ever reach `GLOBAL_OPTIMUM_CERTIFIED`? |
|---|---:|---|---:|---|
| **Exact-CP-SAT-scope** | 155 | `INJECTIVE`/`MERGE_1` x `DROP_UNMAPPED` x no abbreviation (the `MAIN_*` grid + `VARIANT_MERGE1`/`LOW_DISTRACTOR_DENSITY`/`HIGH_DISTRACTOR_DENSITY`/`GLOBAL_CAPACITY`) | **85.2%** | Yes |
| **Neighborhood-only** | 25 | `SELECTIVE_VOWEL_DROP`, `SUSPENSION_1` abbreviation, and their compositions with `MERGE_1` (5 `VARIANT_*` cells) | **0.0%** | **No — structurally excluded** |

The neighborhood-only population's 0% is not a soft or borderline number and does not move with
budget (confirmed directly by `BUDGET_SCALING_AUDIT.md`: identical result at 5s through 1200s on
the same instances). `HybridSolver` routes anything outside `solver_cpsat.py`'s scope straight to
an exhaustive-but-radius-1 local search that can only ever report `LOCAL_NEIGHBORHOOD_OPTIMUM` or
`NO_IMPROVEMENT_FOUND` — never a global certificate. No claim about the exact-scope population's
recovery may be extended to this one; they are answered by entirely different code paths.

## What changed since the first version of this diagnosis, and why (full account)

The first hidden run (4s/8s/12s CP-SAT budgets scaled by `table_size`, `n_identities=8`) found
every gate failing and attributed it to `RESOURCE_LIMIT`, reasoning that runs which reached
`GLOBAL_OPTIMUM_CERTIFIED` (27.8% of runs) recovered the true table with 100% precision every
time. An external review of that reasoning found it unsound on three independent grounds, all
confirmed against the actual data before anything was changed:

1. **Selection bias.** Certification correlated strongly with instance ease (near-100% at
   `unmatched=0%`, near-0% at `unmatched=50%`, and literally 0/40 at `table_size=4` regardless of
   noise). "Perfect precision conditional on certification" is not evidence that the other 72% of
   runs would also be perfect given more time — the certified subset is not a random sample of
   difficulty.
2. **Scope exclusion.** 5 of 9 tested `VARIANT_*` modes can never reach `GLOBAL_OPTIMUM_CERTIFIED`
   at all (see table above) — the original "100% precision when certified" claim said literally
   nothing about them, since none of their 25 runs could ever land in that classification.
3. **A genuine, distinct generator bug**, found while investigating why the certified subset's own
   mean recall was capped near 0.42-0.52 despite `exact_table_recovery=True` (byte-identical
   predicted vs. true table) on 44/50 certified rows: `generator.py`'s matched-label loop silently
   violated its own declared capacity policy whenever `n_identities` was small relative to a
   page's occurrence count (`if chosen is None: chosen = rng.choice(identity_ids)`, ignoring
   capacity), fabricating "true" correspondences the scorer's own capacity-1 matcher could never
   jointly satisfy even given the exact true table. At `n_identities=12` over a 30-slot page the
   feasible ceiling was ~12/30 ≈ 0.40 — matching the observed recall almost exactly. This explained
   why `exact_table_recovery=True` (the table really was found) could coexist with `mapping_recall
   ≈0.5` (the recall *metric's own achievable ceiling* was ≈0.5, independent of solver quality) —
   the first version's prose ("recovered the true table exactly, with 100% precision") was
   technically correct about the table but did not explain this, and read as a stronger claim than
   the recall number could support. Separately, `generate_hidden.py` had drifted to its own
   `GeneratorConfig` call using different `n_identities`/`n_occurrences` than the dev grid,
   breaking the "dev and hidden differ only in seed" invariant.

Fixed: `generator.py` now downgrades a capacity-exhausted "intended match" to a genuine distractor
rather than fabricating an infeasible true correspondence (reports `realized_unmatched_rate` for
transparency); `n_identities` raised to 16 (matches the larger synthetic page size, making
`unmatched_rate=0` actually achievable under `PER_PAGE_CAPACITY_1`); `generate_hidden.py` now calls
the same `sealed_harness.build_generator_config()` dev uses, so they cannot drift again.

With the bug fixed, the *original* 4s/8s/12s budgets were then shown (on dev data) to still be
inadequate for the now-larger, capacity-consistent instances. Rather than assert `RESOURCE_LIMIT`
again on selection-biased evidence, `scripts/budget_scaling_audit.py` (`BUDGET_SCALING_AUDIT.md`)
tested the actual causal claim directly: same instances, swept budgets (5s/15s/60s/300s/1200s),
**unconditional** (not certified-only) recovery measured at every budget. Result: for the
exact-CP-SAT-scope population, certification rate and unconditional recall rise together,
monotonically, plateauing at 60s (0%→62%→100% certified; 0.079→0.521→0.746 unconditional recall);
for the neighborhood-only population, nothing moves at any budget. This is the same causal claim
the first diagnosis made, now actually tested rather than inferred from a biased subgroup — and it
held, for the population it can be tested on. The main sealed benchmark was then re-run at a
uniform, dev-validated 60s budget, producing this file's current numbers.

## Diagnosis, corrected

**Exact-CP-SAT-scope population: `RESOURCE_LIMIT` was real and is now resolved.** At a properly
provisioned budget, this population clears 4 of 5 pre-registered gates on fresh sealed data
(equivalence-aware recovery 1.0/0.93, precision 0.97, recall 0.81) with certification reaching
98.7% (153/155). Held-out accuracy (0.59) narrowly misses its 0.60 gate — a small residual gap,
plausibly held-out-split variance at n=5 seeds/cell rather than a resource question (the
`BUDGET_SCALING_AUDIT.md` plateau at 60s showed no further gain from 300s/1200s on the same
instances, so more CP-SAT time specifically would not close this).

**Neighborhood-only population (5/9 tested modes): not a resource question at all.**
`VERIFICATION_COVERAGE_GAP` — no exact verifier exists for `SELECTIVE_VOWEL_DROP`, abbreviation, or
their compositions in this package (`solver_cpsat.py`'s docstring discloses this scope restriction
explicitly). Budget scaling cannot address a code path that does not consult the budget. This
population remains entirely unqualified (0% recovery, budget-invariant, confirmed on both dev and
hidden data) and would need an actual exact solver built for that scope — a materially larger
engineering effort than tuning a time limit — before it could be assessed at all.

**`FULL_MODEL_CLASS_QUALIFIED=NO`** — because the neighborhood-only population is not qualified,
and the tested matrix explicitly includes it. Any future production design must restrict its
claimed scope to the exact-CP-SAT-scope subset (`DROP_UNMAPPED`, no abbreviation, single-character
source alphabet, `>=60s` CP-SAT budget) and must not claim coverage of deletion/abbreviation/
composition modes without first building genuine exact verification for them.

## Null feasibility, corrected framing

The original 1000-replica null pilot (4s budget, `table_size=4`) had `retry_rate=1.0` — every
replica hit `VERIFICATION_TIMEOUT`. This proved the TIME-BOUNDED (uncertified) null pipeline
resource-feasible only; it said nothing about a full EXACT/certified null pipeline, since no
replica in that pilot ever actually certified anything. A supplementary 30-replica pilot at the
now-validated 60s budget and `table_size=8` reference point resolved this cleanly in the positive
direction: **30/30 replicas reached `GLOBAL_OPTIMUM_CERTIFIED`** (`retry_rate=0.0`), mean 13.74s
per replica, projecting to 229 CPU-hours (14.3h/16 cores) for 60,000 replicas of a genuinely
exact/certified null pipeline (`NULL_FEASIBILITY_60S_RESULTS.tsv`). `NULL_RESOURCE_FEASIBILITY=PASS`
in `REMEDIATION_REPORT.md`'s final status block is reported on this pilot, not the original
uncertified one.
