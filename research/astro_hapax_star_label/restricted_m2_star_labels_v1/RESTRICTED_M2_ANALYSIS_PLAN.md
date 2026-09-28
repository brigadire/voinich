# Restricted M2 STAR LABEL analysis plan (frozen)

Plan frozen before extraction of real target scores (2026-09-16). The primary scope is the non-circular `L` text on f68r1 lines 8–36 and f68r2 lines 7–30. Introductory `P/Pb` text and all `C` circular text are excluded; numeric/technical markers are excluded. f68r3 is excluded from the primary run.

The complete frozen occurrence registry is used for corpus-wide normalized-token frequencies. Primary sets are ALL STAR LABELS, HAPAX (frequency exactly 1), and NON-HAPAX; sensitivity sets are frequency ≤2/≤3, exact readable EVA, composite-normalized, and an uncertainty-excluded subset. A functional unit is a line, not a spatial bbox or a star identity.

M2 search space is inherited unchanged from the frozen audit: recurrent units, role-preserving composition, minimum independent support, singleton prohibition (accepted singleton rules = 0), and compression scoring. M1's retained configuration and deterministic seed 20260901 are used; no post-hoc rules, abbreviations, deletions, insertions, or budget increases are allowed.

Directional designs are A (f68r1 train → f68r2 held-out) and B (f68r2 train → f68r1 held-out), followed only by a combined descriptive run. Main modes are all, hapax, non-hapax, transfer, leave-one-label-out, leave-one-term-family-out, and removal of strongest 1/2 tokens. Null families are token-preserving matched shuffles and lexicon controls with identical scoring and budgets; target and null replicate counts are fixed in the config. Plus-one p-values and multiplicity correction are required.

Acceptance requires Gate R0 (reproducible source/config/scoring, separated calibration/validation, reproducible nulls, passing frozen checksums), scope integrity, no singleton rules, directional held-out improvement over matched null, and historical-lexicon specificity. Any failed gate is reported as BLOCKED; no real production assignments or decipherment claim may be emitted.

The existing M2 implementation is an AUDIT_ONLY synthetic generator and contains no real dictionary search/scoring entry point. Therefore this plan is frozen, target scope is audited, and the restricted real-data run is intentionally not authorized.
