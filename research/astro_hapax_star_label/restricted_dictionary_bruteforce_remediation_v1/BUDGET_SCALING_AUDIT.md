# Budget-Scaling Audit

Run in direct response to a methodological review of this package's first `FAILURE_DIAGNOSIS.md`,
which had claimed `RESOURCE_LIMIT` as the primary cause of the sealed hidden benchmark's gate
failures on the strength of a single observation: runs that reached `GLOBAL_OPTIMUM_CERTIFIED`
recovered the true table with 100% precision. The review identified that this was **selection
bias, not evidence for the causal claim**: certification is not randomly distributed across
instances, so "recovery is perfect conditional on certification" does not imply "recovery would
be perfect for the other 72% given more time." It also identified that 5 of 9 `VARIANT_*` modes
(`SELECTIVE_VOWEL_DROP`, `SUSPENSION_1` abbreviation, and their compositions with `MERGE_1`) are
**structurally excluded** from ever reaching `GLOBAL_OPTIMUM_CERTIFIED` — `HybridSolver` routes
them straight to the neighborhood-only verifier, which never produces that classification — so the
"100% precision when certified" evidence says nothing about them at all.

## What this audit actually tests

Reusing the same generator/scorer/solver code, on **fresh sealed seeds** (`33330001+`, disjoint
from both the dev `20260001+` and hidden `27170001+` ranges), the **same generated instance is
solved 5 times at 5 fixed CP-SAT verification budgets** (5s, 15s, 60s, 300s, 1200s) — isolating the
effect of budget from instance-to-instance variance. 5 cells: two "easy" (no noise/unmatched) and
two "hard" (25% noise, 50% unmatched) exact-CP-SAT-scope cells at `table_size` 4/8/8/12, plus one
`SELECTIVE_VOWEL_DROP` cell that is permanently neighborhood-only, as an explicit control for "does
budget matter here at all." 2 seeds/cell (a disclosed reduction — this is a diagnostic audit, not a
fresh qualification claim, so it does not need the sealed-gate seed count). See
`scripts/budget_scaling_audit.py` for the exact cell definitions.

## Result

```text
budget=    5s   certification_rate=0.00   unconditional_mean_recall=0.079   unconditional_eq_recovery_rate=0.00
budget=   15s   certification_rate=0.62   unconditional_mean_recall=0.521   unconditional_eq_recovery_rate=0.62
budget=   60s   certification_rate=1.00   unconditional_mean_recall=0.746   unconditional_eq_recovery_rate=0.88
budget=  300s   certification_rate=1.00   unconditional_mean_recall=0.746   unconditional_eq_recovery_rate=0.88
budget= 1200s   certification_rate=1.00   unconditional_mean_recall=0.746   unconditional_eq_recovery_rate=0.88
```
(exact-CP-SAT-scope cells only, n=8 instance-budget combinations per budget row — 4 cells x 2 seeds)

```text
budget=    5s / 15s / 60s / 300s / 1200s  (SELECTIVE_VOWEL_DROP, neighborhood-only cell)
unconditional_mean_recall = 0.318 at EVERY budget, unconditional_eq_recovery_rate = 0.00 at EVERY budget
```

**Both halves of the user's continuation criterion are met for the exact-CP-SAT-scope portion of
the model class, and cleanly failed for the neighborhood-only portion:**

- For instances within CP-SAT's scope: increasing the budget from 5s to 60s raises certification
  rate 0%→100% **and** unconditional (not certified-only) recall 0.079→0.746 **and** unconditional
  equivalence-aware recovery rate 0%→88%, together, monotonically, with a clear plateau at 60s (no
  further gain at 300s/1200s — CP-SAT's own early termination on proof of optimality means it does
  not "use" the extra allotted time once it has actually finished, wall-times at 60/300/1200s are
  statistically indistinguishable). This is genuine, not selection-biased, evidence that
  `RESOURCE_LIMIT` (specifically: the 4-14s budgets used in the main sealed run) was a real,
  correctable bottleneck for this part of the model class.
- For the neighborhood-only cell: **budget has exactly zero effect at any of the 5 levels tested**,
  because that code path never consults `verification_time_limit_sec` — confirming the review's
  point mechanically, not just in principle. `SELECTIVE_VOWEL_DROP`/abbreviation/composition modes
  are not "resource-limited" in any sense this package's `CP_SAT` solver can address; they are
  **unqualified by omission** — no exact verifier for that scope was ever built, so no amount of
  time changes anything.
- **One instance (`AUDIT_HARD_TS12` seed `33330321`) reached `GLOBAL_OPTIMUM_CERTIFIED` at 60s+ with
  `equivalence_aware_table_recovery=False`.** This is not a contradiction: certification is w.r.t.
  the TRAIN split only (per the task's train-only-optimization requirement); the certified optimal
  table for train does not always also dominate the true table's fitness on the FULL label set
  (train+held-out) — a genuine, informative train/full generalization gap that exists even among
  exactly-solved instances, separate from the resource question entirely.

## What this does and does not establish

This shows `RESOURCE_LIMIT` is a real, well-evidenced (not selection-biased) explanation **for the
CP-SAT-in-scope subset of the model class** (`DROP_UNMAPPED`, no abbreviation, single-character
source alphabet) — a production design that used a properly-sized budget (this audit suggests
≥60s at these instance sizes, not the 4-14s the main sealed run used) would very likely clear the
sealed gates for that subset. It does **not** extend to the 5/9 neighborhood-only modes, which
remain entirely unqualified and are not a resource question — they need a different exact
verification method to ever be assessed, not more time. Recall even at the plateau (0.746
unconditional, on train-optimized/certified full-set evaluation) is still meaningfully below 1.0,
consistent with genuine noise/unmatched effects and the train/full generalization gap noted above,
not a resource artifact.

Sample size is small (n=8 per budget for the exact-scope trend) — sufficient to establish the
qualitative monotonic pattern and the plateau, not sufficient to certify a production threshold. A
follow-up qualification pass (fresh sealed seeds, ≥20/cell, budget fixed at the plateau value found
here) would be the next concrete, falsifiable step for the CP-SAT-in-scope subset specifically.
