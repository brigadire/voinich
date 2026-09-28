# Validation report

The compact model uses only enumerated compatibility paths, exact `encode_word` parity, injective source-to-EVA rules, one label per assignment, one canonical identity per assignment, and support checks on distinct EVA token types.

The prior witness fails the first validation gate:

| pair | expected token | scorer output | result |
|---|---|---|---|
| Achernar / `acarnar` | `olchc` | `lolchlc` | FAIL |
| Diphda / `rana secunda` | `chol` | `clhlohl` | FAIL |

Thus the package does not treat coverage 2 as a valid real-data result. The path count and all output hashes are recorded in `RUN_STATUS.json` and `SHA256SUMS`.
