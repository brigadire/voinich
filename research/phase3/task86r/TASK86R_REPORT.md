# Task86R report

1. V1.1 preflight: SUPPORTED (validate_contract.py passes; all authoritative hashes verified).
2. Authoritative inputs matched frozen hashes: yes.
3. MFC calibration executed before any Voynich fit: yes.
4. Calibration thresholds frozen before Voynich fitting: yes (`G1_CALIBRATION_FROZEN`).
5. Calibration jobs executed: 4032/4032 (failed: 1728).
6. MFC failures: 1728 jobs (see G1_CALIBRATION_RESULTS.tsv).
7. All 84 candidates attempted per transcription: yes (168 DEVELOPMENT fit rows).
8. DEVELOPMENT-stage failures: 72 rows (see G1_MODEL_FITS.tsv).
9. VALIDATION-selected candidates: see G1_MODEL_SELECTION_REPORT.md.
10. Selection freeze created before HELDOUT opening: yes.
11. PredictiveAdequacy candidates: []
12. StructuralAdequacy candidates: []
13. Both gates: []
14. MEMORIZATION_DOMINATED models: see G1_COMPLEXITY_RESULTS.tsv.
15. COMPLEXITY_UNBOUNDED models: see G1_COMPLEXITY_GROWTH.tsv.
16. Non-converged generation: see G1_GENERATION_RESULTS.tsv.
17. Cross-transcription stability: see G1_TRANSCRIPTION_STABILITY.tsv.
18. Minimal class matches across transcriptions: true (ZL3b=NONE, IT2a=NONE).
19. G1_MINIMAL_CLASS = NONE
20. Model ladder: [{M0 M1 NOT_SUPPORTED} {M1 M2 NOT_SUPPORTED} {M2 M3 NOT_SUPPORTED} {M2 M4 NOT_SUPPORTED} {M2 M5 NOT_SUPPORTED}]
21. See ladder table for the edge where supported representational gain stops.
22. TOKEN_FORMATION_DEPTH = NOT_IDENTIFIABLE
23. EXPLICIT_RULE_GRAMMAR_REQUIRED = INCONCLUSIVE
24. Productive-vs-memorized evidence: see memorization_dominated column, G1_COMPLEXITY_RESULTS.tsv.
25. G1_UNEXPLAINED_STRUCTURE = INCONCLUSIVE
26. PRIMARY: HELDOUT PM1-PM6, complexity-adjusted adequacy, family structural gates, cross-transcription stability. SECONDARY: DEVELOPMENT diagnostics, complexity growth, generation-stability diagnostics.
27. Frozen G1 available for Task87: false.
28. Final Task86R marker: TOKEN_GRAMMAR_FROZEN (issued regardless of whether the identified verdict is a positive, negative, or inconclusive/NONE scientific finding, per task86r.txt section 59).

## Why G1_MINIMAL_CLASS = NONE: three independent, disclosed causes

The confirmatory result is not a single unexplained null; it traces to three
distinct, correctly-recorded findings, each visible in the frozen result
tables:

**1. PM6 (negative-token discrimination) is unavailable for every retained
candidate on both transcriptions** (`G1_NEGATIVE_TOKEN_RESULTS.tsv`:
`negative_exhausted=TRUE` for M0-003, M1-009, M2-003, M5-001, both ZL3b and
IT2a). `NEGATIVE_TOKEN_PROTOCOL.md` requires a class-matched, length-matched
replacement glyph. A length-1 HELDOUT TOKEN's entire same-length alternative
space is the glyph alphabet itself, and in a corpus this size (39K/38K
tokens over 32-45 glyphs) essentially every glyph already appears somewhere
in DEVELOPMENT/VALIDATION/HELDOUT as its own standalone length-1 token —
making a length-1 negative unconstructible by the frozen protocol's own
rules, with no fallback permitted. Because PredictiveAdequacy requires all
five PM1/PM2/PM4/PM5/PM6 gates, this alone is sufficient to fail every
candidate's PredictiveAdequacy. This is a property of the frozen protocol
interacting with Voynichese's small glyph inventory, discovered by running
the protocol exactly as specified, not a code defect or a scope decision
made in Task86R.

**2. M3/M4 (finite-state) training failed for all 9/27 hyperparameter rows
on both real transcriptions' full DEVELOPMENT partitions**
(`G1_MODEL_SELECTION.tsv`: `TRAINING_FAILED: induction operation cap
exceeded`), the same failure mode observed for every M3/M4 candidate on the
high-type-diversity MFC0 calibration populations. The frozen greedy
JS-divergence state-merging induction cannot reduce the DEVELOPMENT prefix
trie's state count to any of the frozen `max_states` values (64/128/256)
within the frozen 100,000-operation cap, for any of the frozen
`merge_js_threshold` values. This closes the finite-state class entirely
at G1 scope under this contract's frozen grid, independent of the PM6
finding above.

**3. Even where PM6 is set aside, `StructuralAdequacy` also fails for
every candidate on at least one transcription** (`G1_FAMILY_VALIDATION.tsv`:
`structural_adequate=FALSE` for all 8 rows) — the edit family (giant
component / isolate share / clustering / degree-frequency Spearman) misses
its 3-of-4 pass threshold for M1, M2, and (on ZL3b) M0 and M5. M5
additionally assigns exactly zero probability to at least one VALIDATION
and one HELDOUT occurrence on both transcriptions (`G1_MODEL_SELECTION.tsv`
`validation_pm1=+Inf`; `G1_PREDICTIVE_RESULTS.tsv` `pm1=+Inf`), the frozen
`unseen_scoring` rule's direct consequence when a form is producible by
neither a retained rule nor the frozen exception table — a real productive-
coverage gap in the frozen frequent-substring slot grammar at this
hyperparameter grid, not a scoring bug (it makes M5's own PredictiveAdequacy
fail even before considering PM6).

None of these three findings is a redefinition of the frozen contract; each
is the literal, disclosed consequence of running Task85+Task85a's frozen
G1 procedure, unchanged, against the real ZL3b/IT2a corpora. `G1_MINIMAL_CLASS
= NONE` is therefore a fully evidenced negative result, not an artifact of
an incomplete or blocked run.

## Required verdicts (task86r.txt section 56)

- `V1_1_PREFLIGHT` = SUPPORTED
- `CALIBRATION_VALID` = SUPPORTED (948 (quantity,metric,candidate) thresholds materialized from 4,032 attempted jobs; 1,728 failures are retained TRAINING_FAILED rows, predominantly M3/M4 induction-cap exhaustion on the high-diversity MFC0/1/2 populations — consistent with finding 2 above, not a calibration defect)
- `CALIBRATION_BEFORE_VOYNICH` = SUPPORTED
- `CALIBRATION_FROZEN` = SUPPORTED (`G1_CALIBRATION_FROZEN`)
- `G1_MODEL_SPACE_EXECUTED` = PARTIAL (M0/M1/M2/M5 executed on DEVELOPMENT/VALIDATION/HELDOUT for both transcriptions; M3/M4 never produced a trained candidate on either real transcription — see finding 2)
- `MODEL_SELECTION_BEFORE_HELDOUT` = SUPPORTED (`GRAMMAR_MODEL_SELECTION_FROZEN`, created before any HELDOUT read)
- `HELDOUT_FIREWALL_PRESERVED` = SUPPORTED
- `G1_PREDICTIVE_STRUCTURE` = NOT_SUPPORTED (finding 1; zero of eight class/transcription rows pass PredictiveAdequacy)
- `G1_STRUCTURAL_REPRODUCTION` = NOT_SUPPORTED (finding 3; zero of eight rows pass StructuralAdequacy)
- `G1_PRODUCTIVE_FORMATION` = NOT_SUPPORTED (M5, the only candidate with an explicit productive/memorized distinction at this scope, assigns zero probability to observed forms outside its retained rule set — see finding 3)
- `G1_CROSS_TRANSCRIPTION_STABILITY` = DIRECTION_STABLE (the negative result itself is stable: `G1_TRANSCRIPTION_STABILITY.tsv` shows every evaluable M0/M1/M2 metric TRANSCRIPTION_STABLE or DIRECTION_STABLE; M5's PM1/PM2 are TRANSCRIPTION_SENSITIVE only because both are `+Inf` on both transcriptions, which the frozen stability rule classifies as sensitive by construction, not because the two transcriptions disagree in substance)
- `G1_GRAMMAR_SUFFICIENT` = NOT_SUPPORTED
- `G1_MINIMAL_CLASS` = NONE
- `TOKEN_FORMATION_DEPTH` = NOT_IDENTIFIABLE
- `EXPLICIT_RULE_GRAMMAR_REQUIRED` = INCONCLUSIVE (no adequate M0-M4 candidate exists to compare against, and M5 itself is not adequate)
- `G1_UNEXPLAINED_STRUCTURE` = INCONCLUSIVE (no minimal class was identified to characterize)
- `TASK87_READY` = NOT_SUPPORTED
