# Pipeline stage input audit

This table records the actual command inputs and dependency chain. “Metadata”
means metadata parsed from IVTFF (folio, hand, Currier, and related blocks), not
ordinary artifacts derived solely from token sequences.

| stage | corpus-only | IVTFF-required | metadata-required | generic-applicable |
|---|---:|---:|---:|---:|
| dict-gen | yes | no | no | yes |
| dict-analyze | yes | no | no | yes |
| structural-analyze | yes | no | no | yes |
| sequence-analyze | yes | no | no | yes |
| begin-end-analyze | yes | no | no | yes |
| structural-normalize | yes | no | no | yes |
| normalization-compare | yes | no | no | yes |
| structural-validate | yes | no | no | yes |
| structural-profile-stability | yes | no | no | yes |
| structural-reliability | yes | no | no | yes |
| soft-structural-space | yes | no | no | yes |
| structural-graphemic | yes | no | no | yes |
| structural-pair-decompose | yes | no | no | yes |
| distance-context-analyze | yes | no | no | yes |
| local-regime-analyze | yes | no | no | yes |
| property-trajectory-analyze | yes | no | no | yes |
| structural-projection-analyze | yes | no | no | yes |
| global-regime-analyze | yes | no | no | yes |
| metadata-validate | no | yes | yes | no |
| cluster-metadata-global | no | no | yes | no |
| conditional-regime-analyze | no | no | yes | no |
| residual-diagnostic-analyze | no | no | yes | no |
| token-relation-validate | no | no | yes | no |
| replicated-local-structure-audit | no | no | yes | no |
| higher-order-sequence-validate | no | no | yes | no |
| positional-continuation-validate | no | no | yes | no |
| transition-network-validate | no | no | yes | no |

The corpus-only chain is stages 1–18 in pipeline order. `metadata-validate`
joins that discovery output with the original IVTFF and emits the per-token
metadata map. Every stage after it consumes that map directly, or consumes an
artifact produced by another metadata-dependent stage. The dependency is
scientific rather than incidental, so generic mode does not fabricate a map.
