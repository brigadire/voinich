# Task87 handoff

**TASK87_READY: NOT_SUPPORTED.**

G1_MINIMAL_CLASS: NONE. TOKEN_FORMATION_DEPTH: NOT_IDENTIFIABLE.
EXPLICIT_RULE_GRAMMAR_REQUIRED: INCONCLUSIVE. G1_GRAMMAR_SUFFICIENT:
NOT_SUPPORTED. No candidate M0-M5 passed both PredictiveAdequacy and
StructuralAdequacy on either transcription, so there is no frozen G1
model, and no viable equivalence set, for Task87 to build G2 on top of.
Per Task85's nested-model principle, Task87 must not proceed to fit G2 on
an invented or convenience G1 winner.

Per-transcription minimal candidates: ZL3b=NONE, IT2a=NONE (see
G1_SELECTED_MODELS.json).

Root causes (see TASK86R_REPORT.md "Why G1_MINIMAL_CLASS = NONE" for full
detail, all evidenced in the frozen result tables, none a code defect or a
Task86R scope decision):

1. The frozen NEGATIVE_TOKEN_PROTOCOL cannot construct a matched negative
   for any length-1 HELDOUT TOKEN on this corpus scale (the same-length
   alternative space is the glyph alphabet itself, and every glyph already
   occurs as its own standalone token somewhere in the corpus) — PM6, and
   therefore the full PredictiveAdequacy gate, is unavailable for every
   candidate on both transcriptions.
2. M3/M4 finite-state induction (greedy JS-divergence state merging) never
   converges under any frozen `(merge_js_threshold, max_states)` combination
   within the frozen 100,000-operation cap, on either transcription's full
   DEVELOPMENT partition — the same failure mode as on the calibration
   populations. The finite-state class is closed at G1 under this frozen
   grid.
3. StructuralAdequacy (the edit-family F2 gate) also fails independently for
   every candidate on at least one transcription, and M5 assigns exactly
   zero probability to at least one observed VALIDATION/HELDOUT form on
   both transcriptions (a real productive-coverage gap in its frozen
   frequent-substring slot grammar at this hyperparameter grid).

If Task87 (or a future Task86-lineage task) wants a usable G1 result, the
prerequisite is a new, explicitly versioned amendment analogous to
Task85a — e.g. a widened/alternate negative-token construction rule for
length-1 tokens, or a wider M3/M4 induction grid/algorithm — never an
in-place edit of this frozen v1.1 contract.

Full result tables: research/phase3/task86r/*.tsv, *.json. Failure ledger:
G1_FAILURE_LEDGER.tsv. Complexity/predictive/structural/stability results
in their respectively named tables.

Task85's known G2 coverage gap remains: Fingerprint V2 has almost no
G2-specific coverage. This is now moot for an immediate G2 fit given
TASK87_READY=NOT_SUPPORTED, but remains relevant to any future G1 retry
that does produce a frozen model: any additional G2 metrics require
preregistration/freeze before G2 fitting, never introduced mid-fit.
