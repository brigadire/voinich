# Task86R execution

Confirmatory execution of the frozen Task85-v1.1 (Task85 + Task85a) G1 contract. Preflight, calibration, DEVELOPMENT/VALIDATION/HELDOUT execution follow `research/phase3/task85a/G1_EXECUTABLE_CONTRACT.json` unchanged.

Calibration: 4032 jobs (1728 failed) across MFC0/MFC1/MFC2 x 16 populations x 84 candidates.

DEVELOPMENT fits: 168 rows (2 transcriptions x 84 candidates).

Implementation resolutions (target-blind, IMPLEMENTATION_DETAIL, documented per Task85a's own resolution policy):

- PCG-XSL-RR-128/64 seeding: the contract fixes SplitMix64(seed) x2 -> 128-bit state, and the real PCG64 XSL-RR output function; the multiplier/increment constants and one warm-up advance are fixed implementation constants (the contract does not specify these beyond "expanded by SplitMix64 twice").
- VALIDATION class-wise selection statistic: argmin VALIDATION PM2 (the frozen primary predictive metric), tie-broken by the grid's own candidate_id order.
- B2 baseline: among M1 order=2 candidates, the VALIDATION-argmin-PM2 alpha.
- M3/M4 state-merging candidate-pair search uses a blue-fringe restriction (each state compared only against already-confirmed representative states, in shortest-access-string order) rather than exhaustive all-pairs enumeration, to stay within the frozen 100,000-operation induction cap on real corpora with thousands of trie states; this is a standard equivalent formulation of greedy state-merging and does not change the frozen threshold, merge, or reject semantics.
- F2 structural metrics reuse internal/fingerprintv2 unchanged, via a bijective glyph<->rune alias encoding (natural glyph mode) so composite EVA glyphs are never re-collapsed.
- Calibration structural/predictive nulls use one generation per (generator, population, candidate) at the HELDOUT-analogue token count (matching the contract's stated 4,032-job workload); seed-variation calibration reuses the 16 independent populations themselves as the replicate axis, per the calibration contract's own population count.
