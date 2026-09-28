# Residual AI mapping report

The calibrated multi-pass test was executed with four independent blind passes, a blind adjudicator, candidate-order perturbation, and 21 negative controls. The hidden gate failed decisively: adjudicated exact occurrence accuracy was 1/7 and candidate-order stability was 0/7, versus mandatory thresholds of 6/7 and 6/7. Adjudication returned 2 PROBABLE and 19 AMBIGUOUS cases across all 21, with no consensus-exact result and no forced answer.

Consequently the procedure was not applied to the 71 residual LABELs. `RESIDUAL_AI_MAPPING_RESULTS.tsv` and `AI_CONSENSUS_LABEL_TOKEN_MAPPING.tsv` are intentionally absent rather than empty or fabricated. `UNRESOLVED_LABELS.tsv` records all 71 as `NOT_RUN_GATE_FAILED`. Existing primary coverage therefore remains 21/92, entirely legacy-confirmed, with grouped coverage 21/64 and ungrouped coverage 0/28. This remains unusable for an identified grouped-versus-ungrouped hapax comparison.

The negative controls abstained in 21/21, but this cannot offset the positive-case failure. No threshold, prompt, candidate rule, or answer was changed after disclosure. No enrichment, frequency calculation, dictionary search, naming hypothesis, semantic interpretation, spatial edit, or expert-verification claim was made.

The available agent sessions used multiple model families. Because the execution environment shares a repository filesystem, clean-room isolation is procedural and content-audited rather than OS-enforced; this limitation is explicit and prevents overstating independence.

```text
RESIDUAL_AI_MAPPING_STATUS=BLOCKED
MAPPING_METHOD=INDEPENDENT_MULTIPASS_AI_CONSENSUS
LEGACY_MAPPINGS=21
CALIBRATION_VISIBLE=14
EVALUATION_HIDDEN=7
HIDDEN_EXACT_ACCURACY=1/7
HIDDEN_ORDER_STABLE=0/7
RESIDUAL_AI_MAPPING_RUN_AUTHORIZED=NO
RESIDUAL_LABELS=71
RESIDUAL_PRIMARY_MAPPED=0
RESIDUAL_PROBABLE=0
RESIDUAL_AMBIGUOUS=0
RESIDUAL_NO_MATCH=0
RESIDUAL_UNREADABLE=0
RESIDUAL_NOT_RUN_GATE_FAILED=71
TOTAL_PRIMARY_MAPPED=21/92
GROUPED_PRIMARY_MAPPED=21/64
UNGROUPED_PRIMARY_MAPPED=0/28
HAPAX_ENRICHMENT_RUN_AUTHORIZED=NO
HAPAX_ENRICHMENT_PERFORMED=NO
LEXICON_MATCH_PERFORMED=NO
HUMAN_EXPERT_VERIFICATION_CLAIMED=NO
FROZEN_INPUTS_UNCHANGED=YES
RESULTS_REPRODUCIBLE=NO
```
