# Model-selection audit

Task85 and Task85a assign VALIDATION the role of within-class candidate and
hyperparameter selection. They define PM1/PM2 as primary predictive evidence,
but nowhere equate that label with `argmin VALIDATION PM2`. The executable
contract's explicit complexity minimization applies only after both adequacy
gates, not to within-class VALIDATION selection.

Task86R `SelectByValidation` minimizes PM2 and resolves equality by frozen grid
order. This is deterministic, but it completes a scientific selection rule.

The requested PM1/PM2/complexity-aware counterfactual cannot be computed from
existing results: `G1_MODEL_SELECTION.tsv` persists one row per class and
transcription, while the execution ledger has no per-candidate VALIDATION
scores. No fitting was rerun for this audit. Thus sensitivity of candidate IDs
is unresolved, while the classification itself is established from the
contract/code mismatch.

`R2_SELECTION = SCIENTIFIC_COMPLETION`.

