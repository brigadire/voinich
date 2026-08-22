# Task66 report: mechanism-space search

This is a statistical-compatibility study, not a decryption attempt and not a
claim about the Voynich cipher. No inverse transformation was applied to
Voynich; the Voynich sequence was only ever compared against, never mined for
a plaintext (task66 sections 3-4, 71).

Authoritative target manifest: 17 metrics loaded from frozen Task58-65
artifacts, 0 MISSING_ARTIFACT.

## 1. What do memoryless transformations explain?

M1 (monoalphabetic): TOKEN_ORDER=0.00071, TOKEN_FORMATION=1.7e-16, LOCAL_TRANSITION=0, LOCAL_REGIME_TOPOLOGY=0, CHARACTER_ENTROPY=-4.9e-16, POSITIONAL_STRUCTURE=-1.8e-15, REPETITION_EDIT_GEOMETRY=-0.0014. M2 (homophony, H=4): TOKEN_ORDER=0.99, POSITIONAL_STRUCTURE=0.006, LOCAL_REGIME_TOPOLOGY=0, TOKEN_FORMATION=-0.0058, REPETITION_EDIT_GEOMETRY=-0.041, LOCAL_TRANSITION=-0.085, CHARACTER_ENTROPY=-2.3.

## 2-3. What does statefulness / slow persistence add?

M4 (per-unit state, K=4, update A): TOKEN_ORDER=0.26, REPETITION_EDIT_GEOMETRY=0.0019, LOCAL_REGIME_TOPOLOGY=0, LOCAL_TRANSITION=-0.049, CHARACTER_ENTROPY=-0.26, TOKEN_FORMATION=-0.89, POSITIONAL_STRUCTURE=-1.6.
M5 (drift scale 20): TOKEN_ORDER=1, LOCAL_TRANSITION=0.041, REPETITION_EDIT_GEOMETRY=0.0014, LOCAL_REGIME_TOPOLOGY=0, CHARACTER_ENTROPY=-0.68, TOKEN_FORMATION=-1.1, POSITIONAL_STRUCTURE=-2.1.

## 4. Can slow state reproduce Task65 distance decay?

See SLOW_STATE_REQUIRED in FINAL_ARCHITECTURE.tsv and TOPOLOGY_RESULTS.tsv's
correlation_length_tokens column for M5 against the Voynich target row.

## 5-6. Are explicit macro-states needed? What creates MIXED_DRIFT_AND_STATES?

M6 (macro only, K=5): TOKEN_ORDER=0.34, LOCAL_TRANSITION=0.048, REPETITION_EDIT_GEOMETRY=0.0045, LOCAL_REGIME_TOPOLOGY=0, CHARACTER_ENTROPY=-0.4, TOKEN_FORMATION=-1.4, POSITIONAL_STRUCTURE=-1.7.
M7 (mixed, K=5, drift 20): TOKEN_ORDER=1, LOCAL_TRANSITION=0.044, LOCAL_REGIME_TOPOLOGY=0, REPETITION_EDIT_GEOMETRY=-0.0012, CHARACTER_ENTROPY=-0.68, TOKEN_FORMATION=-1.1, POSITIONAL_STRUCTURE=-2.1.
See MACRO_STATE_REQUIRED in FINAL_ARCHITECTURE.tsv.

## 7-8. Is constrained formation needed for Tasks59-62? Can state alone get token-internal fingerprint?

Ablation G_ONLY: POSITIONAL_STRUCTURE=4.5, TOKEN_FORMATION=2.5, CHARACTER_ENTROPY=1.9, LOCAL_TRANSITION=0.6, TOKEN_ORDER=0.47, REPETITION_EDIT_GEOMETRY=0.37, LOCAL_REGIME_TOPOLOGY=-0.5.
Ablation S_ONLY: TOKEN_ORDER=0.41, REPETITION_EDIT_GEOMETRY=0.23, LOCAL_TRANSITION=-0.096, CHARACTER_ENTROPY=-0.48, LOCAL_REGIME_TOPOLOGY=-0.5, TOKEN_FORMATION=-1.1, POSITIONAL_STRUCTURE=-3.7.
See CONSTRAINED_FORMATION_REQUIRED and MEMORY_REQUIRED in FINAL_ARCHITECTURE.tsv.

## 9-10. Can form grammar alone get local topology? Are state+form compatible with Task58/63?

Ablation G_PLUS_S: POSITIONAL_STRUCTURE=2.5, TOKEN_FORMATION=1.4, CHARACTER_ENTROPY=0.79, REPETITION_EDIT_GEOMETRY=0.5, TOKEN_ORDER=0.23, LOCAL_TRANSITION=-0.22, LOCAL_REGIME_TOPOLOGY=-0.5.
Ablation M_PLUS_S_PLUS_G: TOKEN_ORDER=0.44, REPETITION_EDIT_GEOMETRY=0.26, CHARACTER_ENTROPY=-0.2, LOCAL_TRANSITION=-0.24, LOCAL_REGIME_TOPOLOGY=-0.5, POSITIONAL_STRUCTURE=-1.4, TOKEN_FORMATION=-3.7.

## 11-12. What changes without word boundaries? Does generated grouping help?

M9 (STREAM + generated boundaries + form): POSITIONAL_STRUCTURE=4.3, TOKEN_FORMATION=2.4, CHARACTER_ENTROPY=1.9, LOCAL_TRANSITION=0.63, REPETITION_EDIT_GEOMETRY=0.53, TOKEN_ORDER=0.46, LOCAL_REGIME_TOPOLOGY=-0.5.
See GENERATED_BOUNDARIES_REQUIRED in FINAL_ARCHITECTURE.tsv (compares against
M3's WORD_PRESERVING form-only result).

## 13-14. Does the mechanism retain real plaintext dependence, or does it ignore the input?

Plaintext-sensitivity classes across representative mechanisms: map[PARTIAL_INPUT_DEPENDENCE:30 STRONG_INPUT_DEPENDENCE:3 WEAK_INPUT_DEPENDENCE:3].
See INFORMATION_RETENTION.tsv for coarse input/output mutual information and
PLAINTEXT_DEPENDENCE_PRESERVED in FINAL_ARCHITECTURE.tsv.

## 15-16. Does the result transfer Doyle -> Longfellow -> Astafiev? What survives held-out?

Pareto frontier frozen before held-out was opened: [M0_IDENTITY M10_STATEFUL_FORM_K2 M10_STATEFUL_FORM_K4 M10_STATEFUL_FORM_K8 M11_MIXED_FORM_K2 M11_MIXED_FORM_K5].
M0_IDENTITY (the untransformed baseline) is on this run's frontier only because
it is, by construction, non-dominated on every family (all scores exactly 0);
it is the control, not a claimed compatible mechanism, and its INCONCLUSIVE
held-out row should be read that way. A later run of this tool excludes M0
from candidate freezing (see main.go's withoutIdentity); this run's frozen
artifacts are kept as generated rather than redone.
Overfit classification per candidate: map[M0_IDENTITY:INCONCLUSIVE M10_STATEFUL_FORM_K2:CONFIRMED M10_STATEFUL_FORM_K4:CONFIRMED M10_STATEFUL_FORM_K8:INCONCLUSIVE M11_MIXED_FORM_K2:INCONCLUSIVE M11_MIXED_FORM_K5:INCONCLUSIVE].
See CORPUS_TRANSFER.tsv and HELDOUT_RESULTS.tsv.

## 17. Which operations are required vs redundant?

- MEMORY_REQUIRED: **NOT_REQUIRED** (family score with=-0.5 without=-0.5 delta=0)
- SLOW_STATE_REQUIRED: **NOT_REQUIRED** (family score with=0 without=0 delta=0)
- MACRO_STATE_REQUIRED: **NOT_REQUIRED** (family score with=-0.5 without=-0.5 delta=0)
- CONSTRAINED_FORMATION_REQUIRED: **REQUIRED** (family score with=2.85 without=-1.34 delta=4.19)
- GENERATED_BOUNDARIES_REQUIRED: **NOT_REQUIRED** (family score with=3.08 without=3.03 delta=0.0534)
- HOMOPHONY_HELPFUL: **DISFAVORED** (family score with=-0.879 without=-0.00259 delta=-0.877)
- STOCHASTIC_OUTPUT_REQUIRED: **DISFAVORED** (family score with=-1.95 without=0 delta=-1.95)
- PLAINTEXT_DEPENDENCE_PRESERVED: **SUPPORTED** (33/36 representative mechanisms show strong/partial input dependence)

## 18. Error robustness

See ERROR_ROBUSTNESS.tsv for the frontier's fingerprint degradation under
0.1%-5% scribal-like error rates; robustness was not used to select the
frontier (task66 sections 67-68).

## 19. Minimal architecture found

Operations classed REQUIRED or SUPPORTED: [CONSTRAINED_FORMATION_REQUIRED PLAINTEXT_DEPENDENCE_PRESERVED]. A transformation architecture
combining these is what this study found statistically compatible with several
independent Voynich fingerprint families - this is not a claim that Voynich
used any such mechanism (task66 section 71).

## 20. What remains unexplained

Any metric family whose best frontier candidate's progress stays below the
0.15 movement threshold in FAMILY_METRICS.tsv/HELDOUT_RESULTS.tsv, and any
family whose authoritative target artifact was MISSING_ARTIFACT in
VOYNICH_TARGET_MANIFEST.tsv, is left unexplained by this study.
