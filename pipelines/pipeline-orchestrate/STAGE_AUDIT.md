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
| metadata-validate | no | yes | yes | no (Class A) |
| cluster-metadata-global | no | no | yes | no (Class A) |
| conditional-regime-analyze | no | no | yes | no (Class A) |
| residual-diagnostic-analyze | no | no | yes | no (Class A) |
| token-relation-validate | no | no | yes | **yes** (Class B/C core; `-generic-corpus`) |
| replicated-local-structure-audit | no | no | yes | **yes** (Class B/C core; `-generic-corpus`) |
| higher-order-sequence-validate | no | no | yes | **yes** (Class B; `-generic-corpus`, chain-gated on stage above) |
| positional-continuation-validate | no | no | yes | **yes** (Class B; `-generic-corpus`, chain-gated on stage above) |
| transition-network-validate | no | no | yes | **yes** (Class B/C core; `-generic-corpus`) |

The corpus-only chain is stages 1–18 in pipeline order. `metadata-validate`
joins that discovery output with the original IVTFF and emits the per-token
metadata map. Stages 19–22 (`metadata-validate`, `cluster-metadata-global`,
`conditional-regime-analyze`, `residual-diagnostic-analyze`) test real
Currier/hand/quire identity itself, not merely use it as a segmentation
device — the dependency there is scientific rather than incidental, so
generic mode does not fabricate a map for them and they stay
`NOT_APPLICABLE` with a stage-specific reason.

Stages 23–27 were re-audited for task43 (see
`../GENERIC_STAGE_APPLICABILITY_AUDIT.md` for the full per-field evidence,
per-stage classification, and hypothesis-equivalence write-ups). Their core
statistics operate on a physical-block partition derived from Currier×hand
identity purely as an opaque grouping device — the block *label*, not its
manuscript meaning, is all any formula reads. Each of these five stages now
accepts `-generic-corpus`, which derives that block partition from
`internal/genericsegmentation` (a deterministic, corpus-size-scaled,
language-agnostic segmentation) instead of a real IVTFF-sourced
`token_metadata_map.tsv`. Any sub-computation that specifically requires
real cross-Currier/cross-hand identity (token-relation-validate's
classification labels, replicated-local-structure-audit's replication
status, transition-network-validate's metadata-transfer diagnostic) is
either recomputed with generic-only vocabulary or skipped outright in
generic mode — never run against fabricated metadata. The IVTFF/Voynich
code path in every one of these five packages is reached only when
`-generic-corpus` is absent (the default), and is unchanged.
