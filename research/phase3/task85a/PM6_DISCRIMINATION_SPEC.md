# PM6 negative-discrimination functional

Positives and negatives come only from `NEGATIVE_TOKEN_PROTOCOL.md`. Score each
whole TOKEN by natural log probability under the fitted model; zero probability
is negative infinity. There is one positive and one negative per matched pair,
but ROC AUC is computed occurrence-weighted over the pooled two multisets:

`AUC = (wins + 0.5*ties)/(N_positive*N_negative)`,

where a win has positive score greater than negative score. Exact IEEE float64
equality is a tie, including two negative infinities. Compute equivalently by
midranks after sorting `(score,label,occurrence_index)` ascending; aggregate no
TOKEN types and discard no duplicates. PM6 is computed independently for each
candidate and transcription, then seed-aggregated by the median. Exhaustion or
an empty class makes PM6 unavailable and PredictiveAdequacy fails. Uncertainty
is the deterministic 2.5% and 97.5% percentile of 2,000 matched-pair bootstrap
resamples using the seed namespace `PM6_BOOTSTRAP`; it is reported but does not
replace the MFC threshold gate.
