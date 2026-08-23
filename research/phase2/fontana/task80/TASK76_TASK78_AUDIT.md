# TASK76 / TASK78 audit

This audit reads the task76/task78 source references, E/I/H/U tables,
machine packages, state specifications, experimental reports, and tests. It
does not treat a runnable profile as historical evidence and uses no Voynich
or Fingerprint material.

| package | status | freeze decision | audit finding |
|---|---|---|---|
| F01 Speculum | ACCEPTED | core + task76 profile | Direct source procedure, explicit state-only serialization, 24/24 literal round trips, and documented ablation/corruption tests meet all eligibility criteria. Capacity, alphabet, and mounting remain profile assumptions. |
| F08 Serpens | ACCEPTED_WITH_LIMITATIONS | core | Ordered positional state and centre-to-edge traversal are reproducible (6/6). Capacity, stop rule, insert mobility, and semantic association are excluded from the frozen core. |
| F11 Arismetricum | ACCEPTED_WITH_LIMITATIONS | core | The frozen claim is only convention-governed index-to-opaque-cue lookup (6/6), not literal text or an arithmetic mapping. |
| F12 Horalogius | ACCEPTED_WITH_LIMITATIONS | core | The frozen claim is temporal cue emission plus an explicit human-memory boundary (12/12 transition checks), not a measured retention effect. |
| F07 Rota | REQUIRES_REPAIR | REFERENCE_ONLY | The cyclic selector is useful sensitivity evidence, but zero, layout, step, and cardinality are profile choices. |
| F10 Cylindrus | REQUIRES_REPAIR | REFERENCE_ONLY | R0/R1 demonstrate that band coupling changes propagation; no invariant route/coupling semantics can be frozen. |
| F02–F06, F09 | EXCLUDED_FROM_FREEZE | EXCLUDED | Task74/task78 leave a transition or traversal rule materially underdetermined. |

## Eligibility checks

The four frozen cores each have an identified source, separated E/I/H/U
claims, components/state schema, an implemented cycle or cue relation,
reproducible transitions, explicit assumptions, a harness outside historical
semantics, limitations, and tests. F01 additionally freezes the tested
Latin23/R12 profile. No frozen entry needs knowledge of the Voynich
manuscript.

## Corrections applied by task80

`task78` described F11/F12 as not ready for a complete physical freeze. This
task therefore freezes only their invariant **operational cores**, not their
R0 numeric capacities, mappings, drive, or human-performance claims. F07 and
F10 remain executable references, not frozen historical models.
