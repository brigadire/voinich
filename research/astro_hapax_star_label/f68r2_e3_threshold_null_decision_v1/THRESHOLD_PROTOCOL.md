# Threshold protocol for the post-hoc null

Observed statistic: maximum occurrence coverage selected over all 64 frozen E3 operation profiles, with the existing injective mapping and distinct-EVA support gate.

Decision thresholds are frozen before interpreting the null:

| p-value | Interpretation |
|---:|---|
| `p >= 0.10` | 5/27 is fully compatible with random selection |
| `0.05 <= p < 0.10` | weak exploratory signal |
| `p < 0.05` | interesting exploratory signal; post-hoc and not a decipherment |

Identity stability is reported separately from p-value. Low stability cannot be rescued by a small p-value.

The null must repeat model selection across all 64 profiles, mapping complexity, operation complexity, and the same scorer/support semantics. An incomplete null cannot produce a p-value. No E4 search or alternative dictionary search is authorized by this package until the null is complete and the threshold decision is recorded.
