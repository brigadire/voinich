# Witness-only null protocol

This is a post-hoc exploratory lower-bound test. It keeps the global-character-shuffle null, 27 token lengths, the frozen 64 operation profiles, exact `DROP_UNMAPPED` scorer, injective mapping, `GLOBAL_CAPACITY_1`, and distinct-EVA support gate.

Each replicate has one deterministic seed and one fixed budget. Profiles are processed in the frozen order: the nine real maximum-5 profiles first, then descending relaxed upper bound and stable profile ID. Paths are generated on demand for one profile and discarded after that profile.

Only an independently replayed support-valid five-LABEL/five-identity witness counts as `WITNESS_FOUND`. A timeout or exhausted search budget is `NO_WITNESS_WITHIN_BUDGET`, never an infeasibility certificate. Implementation errors are `IMPLEMENTATION_FAILURE`.

The conservative success rate is `K/N`, where every unresolved replicate remains a failure. The one-sided 95% Clopper–Pearson lower bound is the decision statistic. This test can establish frequent random attainability; it cannot establish rarity from zero witnesses.
