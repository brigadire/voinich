# Preproduction audit report

## Decision

`CROSS_DOMAIN_SEARCH_AUTHORIZED=NO`. No dictionary search was run.

## Findings

- All 134 botanical rows were independently rechecked against the frozen Isidore snapshot. Exact source presence and locator/context checks were performed. 45 rows are excluded as sort/epithet, fruit, or product/material terms.
- The source extraction assigns one canonical identity per form by string derivation. This is not an independent identity audit; all 134 rows are `IDENTITY_UNRESOLVED`. The reported 134=134 equality is therefore not accepted.
- All 31 control forms have line-level source rechecks and the same extraction mechanism. The 31st row is explicitly retained but excluded from the 30-row panel by deterministic capacity capping.
- All three matched panels have 30 rows and 30 string-distinct identities, but the botanical panel fails identity readiness. The former astronomy 3/27 result is not reused.
- All 99 pseudo-controls have 30 rows, deterministic replay, preserved lengths and character multisets, zero real-lexicon overlap, no internal duplicates, and no global alphabet isomorphism.
- The 64 E3 profile registry is present and configuration parity is frozen; no profile was executed.

## Required remediation

Resolve botanical canonical identities from source evidence or an explicitly frozen external identity authority, rebuild the 30-identity panel by metadata-only deterministic matching, and rerun this audit. Only then may a fresh matched astronomy run and cross-domain search be authorized.
