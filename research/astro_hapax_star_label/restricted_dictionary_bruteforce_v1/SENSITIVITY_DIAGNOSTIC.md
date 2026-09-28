# Synthetic recovery diagnostic (post-run description)

The preregistered synthetic gate fails. At corruption 25%, f68r2→f68r1 passes
17/20 seeds rather than the required 18/20. No threshold, seed, form or system was
changed after observing this result. The final scientific conclusion is therefore
ENGINE_INCONCLUSIVE / ENGINE_SENSITIVITY_INSUFFICIENT, regardless of real scores.

All 60 joint synthetic searches pass; every synthetic endpoint has an empirical
p=0.01 against its 99 full-selection length/endpoint-preserving null dictionaries.
All held-out true surviving-pair recovery fractions exceed the 60% minimum.
The failing component is transformation equivalence: some train splits select
S000 (identity) instead of planted S008 (j→i, v→u). On their held-out forms those
systems agree on less than the preregistered 90%. This shows a limitation of
selecting the global rule from a small page under a complexity penalty, rather
than failure of maximum matching to recover exact supplied strings.

For each noise level the passing-seed counts (joint, f68r1→f68r2, f68r2→f68r1) are:

- 0%: 20, 20, 19.
- 10%: 20, 18, 19.
- 25%: 20, 18, 17.

These diagnostics explain the frozen decision; they are not a new experiment or
permission to relax the gate. Power against unknown encodings or richer
abbreviation rules has not been demonstrated. Any alternative selection policy
or transformation family requires a separately preregistered v2.
