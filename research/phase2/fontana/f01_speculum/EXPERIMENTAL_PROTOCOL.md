# F01 Speculum — Pre-registered Experimental Protocol (Task76 Block 3/statistical protocol)

Registered (as source: `internal/speculumf01/messages.go`,
`research/phase2/fontana/f01-speculum-analyze/main.go`) before any
baseline/ablation/corruption result was inspected. The message set, the
reference lexicon, and the corruption scenarios below were fixed at
authoring time and were not edited after seeing results; the two
edit-distance-classifier findings recorded in `FORMAL_MODEL.md` and
`TASK76_REPORT.md` are reported, not used to retroactively change the
message set.

## Message set (`MESSAGE_SET.tsv`)

12 natural-language classical Latin messages (`NaturalMessages`,
spelled with V for U, no J/U/W, matching the Latin23 alphabet), chosen
before running anything to cover:

| axis | coverage |
|---|---|
| length | 3, 4, 4, 5, 6, 6, 7, 7, 8, 8, 11, 11 letters |
| repetition | none (PAX, DEVS, GLORIA, NATVRA-low), moderate (MEMORIA, FONTANA, KALENDAE), high (ANNA palindrome, EXPERIMENTA 3xE, CONSTANTINA) |
| rare symbols | EXTRA (X), KALENDAE (K), EXPERIMENTA (X) |
| thematic control | SPECVLVM — the device's own name, included once, not weighted specially in any statistic |
| structure | one deliberate palindrome (ANNA) to test the K6/direction edge case |

Plus 12 **random controls**, one per natural message's exact length,
generated from the same Latin23 alphabet with a fixed seed
(`controlSeed = 74760823`), rejecting (and redrawing) any draw that
collides with the reference lexicon — see `GenerateRandomControls`.
Matched length + matched alphabet, zero language predictability by
construction: this is the Block 7 natural-vs-random comparison pair.

## Reference lexicon (`BaseLexicon`, `internal/speculumf01/messages.go`)

36 classical Latin words fixed before any experiment ran, explicitly
**small and illustrative, not an exhaustive dictionary**. It models the
"known language" contribution `C` from Block 7: how many members of a
combinatorial candidate set are recognizable words versus how many are
not. Because it is small, every "language narrows the candidate set to
size 1" and "correctable without knowing M" result below is an
upper-bound / best-case estimate, not a claim about how a fluent
Neo-Latin reader with an unabridged dictionary would perform (stated
explicitly in `TASK76_REPORT.md`).

## Experimental conditions

- **Baseline** (Block 3): `K_full`, intact `S`, all 24 messages, one
  trial each (the model is deterministic — repeating a trial cannot
  change its outcome, so `n_repeats = 1` per message; determinism itself
  is separately verified by `TestEncodeDeterministic`).
- **Ablation** (Block 4): the 7 combinatorial conditions K1–K7
  (`CondFullKnowledge` … `CondConventionUnknown`) applied to all 24
  messages once each = 168 rows in `ABLATION_RESULTS.tsv`. K8 (no
  instruction at all) is qualitative, not combinatorial (see
  `TASK76_REPORT.md`). K9 (instruction known, state metadata missing) is
  realized as a Block-5 corruption scenario under full `K`, not a
  separate combinatorial run (it would otherwise duplicate the
  `deletion_*` rows).
- **Corruption** (Block 5): 8 named scenarios
  (`single_position_substitution`, `random_ring_shift`,
  `deletion_ring_identity_marked`, `deletion_physical_collapse`,
  `swap_two_elements`, `orientation_mark_loss`,
  `outer_contour_loss_2rings`, `multiple_independent_damages_3`) applied
  to all 24 messages where the scenario's precondition is met (e.g. swap
  needs length >= 2, multi-damage needs length >= 4) = 190 rows in
  `CORRUPTION_RESULTS.tsv`.
- **Alphabet-profile sensitivity**: the full baseline + ablation battery
  repeated once under Modern26 in place of Latin23
  (`ALPHABET_SENSITIVITY.tsv`), per task76's requirement to build more
  than one reconstruction profile when the source admits more than one
  operational interpretation (here: alphabet size).
- **Human pilot** (Block 6): a single self-experiment operator (this
  session) performing all three roles (controller/encoder via the CLI,
  decoder by hand), 2 trials, explicitly reported as `N=1` pilots, not a
  general human-performance result — see `HUMAN_PILOT_LOG.md`.

## Primary outcome

`P(M-hat = M | K_i, S_j)`, reported per condition/scenario as an exact
fraction over the fixed message set (24 baseline/ablation trials, up to
24 corruption trials per scenario). Given the small, fixed *n*, results
are reported as raw fractions with a Wilson 95% interval where a
proportion is quoted in prose in `TASK76_REPORT.md`; no strength-of-effect
claim is made from these small-N proportions beyond what the interval
supports.

## Secondary outcomes

- Symbol/character accuracy and edit distance (`CharacterErrorRate`,
  Levenshtein), per corruption scenario.
- Number of compatible decodings (`NCandidatesRaw`/`NCandidatesLex`) and
  `log2` entropy estimate, per ablation condition — exact where fully
  enumerated, analytic-only (flagged `EnumerationCapped=true`) above the
  tractability cap (permutations of more than 8 elements, or the K7
  compound above length 5).
- Error-propagation locality (`FractionAfterFirstError`) and qualitative
  class (`local`/`synchronization`/`global`/`cascading`).
- Detectability (`Detectable`, via the reference lexicon) and
  correctability without knowing `M` (`CorrectableWithoutM`, nearest
  unique lexicon neighbor at edit distance <= 1) — natural-language
  messages only; undefined (`false`, not evaluated) for random controls
  by design, since a "known language" filter has nothing to bite on for a
  message drawn to look like nothing.

## Rule for ambiguous/degenerate cases

Every ablation condition reports the *full compatible candidate set*
(`NCandidatesRaw`), not a single decoder's lucky guess; `ExactBlindP =
1/NCandidatesRaw` is reported as the best any blind decoder could do, not
as an observed frequency over repeated guesses (the model is
deterministic, so repeated guessing would not sample this distribution
empirically). Where the combinatorial space is intractable to fully
enumerate as strings, only the analytic candidate count is reported and
`EnumerationCapped=true` is set; no attempt is made to approximate or
sample the string set in that regime.
