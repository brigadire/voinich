# Final witness-only null result

The first run produced one apparent witness, but independent replay with the full global mapping rejected it because path-local mappings generated extra output characters. That record was classified as an implementation failure and is excluded.

The corrected runner attempted 100 deterministic null replicates with a fixed 5-second budget. It found zero valid witnesses after full global-scorer replay. One attempted replicate produced a path-local candidate that failed global replay and is recorded as `IMPLEMENTATION_FAILURE`; the remaining 99 are correctly run non-witnesses:

- `WITNESS_FOUND`: 0/100;
- `NO_WITNESS_WITHIN_BUDGET`: 99/100;
- `IMPLEMENTATION_FAILURE`: 1/100;
- observed rate: 0;
- one-sided 95% Clopper–Pearson lower bound: 0;
- two-sided 95% interval over the 99 correctly run replicates: [0, approximately 0.0366].

Under the requested conservative rule, `RESULT_COMPATIBLE_WITH_RANDOM=NOT_ESTABLISHED`. The witness-only test cannot establish rarity from zero witnesses. It also cannot support an exploratory signal. `NULL_RARITY_ESTABLISHED=NO` and `SCIENTIFIC_CLAIM=NONE` remain mandatory.
