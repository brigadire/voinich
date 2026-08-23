# Voynich Fingerprint v2 — Gap Priority

Status: **DESIGN, NOT FROZEN**. Covers parent task `tasks_ph2/task73.txt`
section 28. Priorities below are a recommendation for the order of future
implementation tasks (Task75 and beyond); Task73 does not implement any
of them (parent task section 27) and does not itself set
`FINGERPRINT_V2_FROZEN` (see the freeze gate at the end of this
document, parent section 29).

## 1. Prioritization table

Each row is a gap *group* (a cluster of related metric families from
[FINGERPRINT_V2_SPEC.md](FINGERPRINT_V2_SPEC.md)/
[FINGERPRINT_V2_IMPLEMENTATION.tsv](FINGERPRINT_V2_IMPLEMENTATION.tsv)),
not an individual metric, since implementation is naturally scoped at
that grain. Value/cost/risk use High/Medium/Low.

| gap group | members | scientific value | discrimination value | implementation cost | data availability | statistical risk | notes |
|---|---|---|---|---|---|---|---|
| shared enabling infrastructure | C-GRAMMAR null generator; $I field surfacing on TokenMetadata | H | H (blocks LP2, EF4, CS1) | L | AVAILABLE | L | small, high-leverage prerequisites; should be built first, not treated as part of any one metric family |
| lexical paradigms | LP1, LP2, LP3, LP4 | H | H | H | AVAILABLE | M | the single largest Phase I gap by name (parent task section 7); LP2's productivity test is the load-bearing statistical result the rest depends on |
| edit-family geometry | EF1, EF2, EF3, EF4, EF5 | H | H | M | AVAILABLE | L-M | directly tests whether the giant edit family is "just" a consequence of bounded grammar; EF1-EF3 are cheap wins, EF4 depends on the shared C-GRAMMAR generator |
| cross-scale (new estimators) | CS1, CS5, CS7 | H | H | M | AVAILABLE, but CS1/CS7 blocked on LP1-LP2 | M | named the primary v2 target together with lexical paradigms (parent task section 14); CS5 is unblocked today, CS1/CS7 are not |
| page/2D stratification | PG1, PG2, PG3, PG4 | H | H | L-M | CONFIRMED AVAILABLE | L | named a "major Phase I gap" (parent report section 14); mostly re-stratifies already-existing estimators over already-parsed metadata, so cost is low relative to value |
| local/line decomposition | LL1, LL2, LL3 | M-H | M-H | M | AVAILABLE | L-M | required by the parent task to avoid collapsing physical-line/local-regime/page effects into one number; reuses Task64/65 data directly |
| token formation extension | TF1, TF2 (non-family strata now), TF3 | M-H | M | M | AVAILABLE | L-M | answers the explicit "how global vs. context-dependent is token formation" question; TF2's family stratum is blocked on lexical paradigms |
| sequence extension | SQ1, SQ2, SQ3, SQ5 | M | M | L-M | AVAILABLE | L, except SQ5 at n=4,5 (M-H, low power expected and to be reported, not hidden) | extends already-known small-but-real effects; lower marginal discrimination value than lexical/cross-scale/page because the qualitative finding (weak order, real local structure) is already established |
| glyph refinement | G1, G2, G3, G4 | M | M | L | AVAILABLE | L | Phase I already grades this level WELL_COVERED; refinements mainly sharpen existing findings and feed G3/G4 into edit-family and cross-scale work |
| hierarchy | HR1 | H | H | M-H | AVAILABLE except lexical-family level | M | directly answers "at what level does structure arise"; full value only reached once lexical-paradigm labels exist, but a coarser version (excluding the family level) is buildable immediately |
| sequence/cross-scale family-dependent items | SQ4, TF2 family stratum, CS1, CS7 | H (contingent) | H (contingent) | M, contingent on LP timing | blocked on LP1-LP2 | M | not a new implementation cost beyond their base estimators; listed separately to make the LP1-LP2 dependency chain explicit |
| compression/predictability | CP1, CP2, CP3 | L-M | L-M | L (CP1, CP3) / M (CP2) | AVAILABLE | L | explicitly a cross-check against G1/G4/TF/SQ findings, not a primary discriminator (parent task section 15); lowest priority of the implementable groups |
| declined (page/2D, image content) | IMG2D | N/A | N/A | N/A | NOT_AVAILABLE | N/A | correctly out of scope; listed for completeness, not for future implementation, per parent task section 12 |
| declined (transcription cross-check) | TRANSCRIPTION_XCHECK | N/A | N/A | N/A | NOT_AVAILABLE | N/A | listed so its absence is tracked, not implemented; would require acquiring a second transcription under `DATA.md` |

## 2. Recommended implementation order

This is a recommendation for scoping future tasks, not a commitment made
by Task73:

1. **Shared enabling infrastructure first.** Build the C-GRAMMAR null
   generator and surface `$I` onto `TokenMetadata` before anything that
   depends on them (LP2, EF4, CS1, PG4). Both are small and unblock
   multiple higher-value groups.
2. **Lexical paradigms next**, because it is both the highest-named gap
   and a dependency for four other groups (TF2's family stratum, HR1's
   full hierarchy, SQ4/CS1/CS7). Sequencing it early shortens the
   dependency chain for everything downstream.
3. **Page/2D stratification and edit-family geometry in parallel** with
   lexical paradigms, since neither blocks nor is blocked by it, and
   both are graded low-to-medium implementation cost for high value.
4. **Local/line decomposition and cross-scale (CS5, then CS1/CS7 once
   unblocked)** follow, consolidating what the first three groups
   produce into the joint/cross-scale statements the parent task
   prioritizes over marginal counts.
5. **Token formation extension, hierarchy, and sequence extension** are
   scoped after the above, since their highest-value components (TF2's
   family stratum, HR1's full hierarchy, SQ4) are gated on lexical
   paradigms already being available, and their non-gated components
   (TF1, TF3, SQ1-SQ3, SQ5, HR1's coarse version) can be picked up
   opportunistically alongside steps 1-4 if capacity allows.
6. **Compression/predictability last**, consistent with its explicitly
   supplementary role (parent task section 15) and its lower
   discrimination-value grade above.

## 3. Freeze gate (parent section 29)

Task73 defines, but does not itself satisfy, the condition for setting
`FINGERPRINT_V2_FROZEN`. Fingerprint v2 may be frozen only when **all**
of the following hold:

1. **Required metrics implemented.** At minimum, every metric family
   graded `NEW_IMPLEMENTATION` or `NEEDS_MINOR_EXTENSION` in
   [FINGERPRINT_V2_IMPLEMENTATION.tsv](FINGERPRINT_V2_IMPLEMENTATION.tsv)
   within the "lexical paradigms," "edit-family geometry," "cross-scale,"
   and "page/2D stratification" groups above is implemented and reported
   (these four are the groups the parent task names as primary gaps;
   the remaining groups may freeze in a documented partial state if
   explicitly justified, but these four may not).
2. **Controls validated.** Every implemented metric has passed the
   control-leakage check in
   [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 3
   for its assigned null(s).
3. **Stability checked.** Every implemented metric has run the minimum
   stability battery in
   [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 4 and
   is either classified stable or explicitly narrowed/deferred per that
   section's instability criterion — no metric enters the frozen
   fingerprint with an unreported stability status.
4. **Redundancy checked.** The pairwise correlation analysis in
   [FINGERPRINT_V2_CONTROLS.md](FINGERPRINT_V2_CONTROLS.md) section 5 has
   been run once across all implemented metrics, and its resolution rule
   has been applied (duplicates dropped/merged, genuine co-variation
   recorded, family-level redundancy reported to the distance/weighting
   document).
5. **Distance defined and exercised.** At least the family-balanced
   distance and the Pareto comparison from
   [FINGERPRINT_V2_DISTANCE.md](FINGERPRINT_V2_DISTANCE.md) section 2
   have been computed at least once, across Voynich and the existing
   C-NAT/C-PHASE1 comparison corpora, to confirm the toolkit runs
   end-to-end before it is relied on for a future Fontana/mnemonic-model
   comparison.
6. **Missing-data semantics defined.** Every implemented metric follows
   its spec's declared missing-data behavior (explicit
   `INSUFFICIENT_SUPPORT`/`NOT_YET_AVAILABLE`/`UNCODED` markers, never
   silent imputation or silent pooling), verified by a
   review pass across implemented metrics at freeze time.
7. **Tests pass.** Standard repository test/verification discipline
   (as used for every prior Phase I task) passes for all implementing
   code.

If any of these seven conditions is unmet, fingerprint v2 remains
`DRAFT`. Partial freezing of specific families (e.g. freezing glyph and
page/2D metrics while lexical-paradigm or cross-scale work is still in
progress) is explicitly **not** a substitute for the full gate above,
because the parent task's own priority ordering (lexical paradigms and
cross-scale as the two primary targets) means a fingerprint frozen
without them would not represent what Task73 set out to fix.
