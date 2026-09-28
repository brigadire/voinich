# Witness-only null result

The lazy positive-only search attempted 100 deterministic null replicates with a fixed 5-second budget per replicate. One candidate failed full global replay and is classified as an implementation failure; no candidate passed as a witness.

- `WITNESS_FOUND`: 0/100
- `NO_WITNESS_WITHIN_BUDGET`: 99/100
- `IMPLEMENTATION_FAILURE`: 1/100
- observed witness rate over 99 correctly run replicates: 0
- one-sided 95% Clopper–Pearson lower bound: 0
- supplementary two-sided 95% interval over 99 correctly run replicates: [0, approximately 0.0366]

The lower bound is below 0.10. Under the frozen decision rule, `RESULT_COMPATIBLE_WITH_RANDOM=NOT_ESTABLISHED`; the test does not establish rarity either. This witness-only design can confirm frequent random attainability, but it cannot prove that the null rarely reaches 5/27.

No null witness passed independent replay. The real-data calibration witness passed exact output parity, injective global mapping, five distinct LABEL occurrences, five distinct identities, and support ≥2 distinct EVA types.
