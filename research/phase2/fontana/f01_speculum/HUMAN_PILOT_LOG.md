# F01 Speculum — Human Operational Validation Pilot (Task76 Block 6)

**This is an N=1 self-experiment pilot, not a general human-performance
result.** No multi-participant study was run; this section exists because
task76 explicitly allows a pilot self-experiment in that case, on
condition that it is not presented as a general result. Read every number
below as "what happened once," not as an estimate of population
performance.

## Roles

A single operator (this assistant, in this session) filled all three
roles the block describes, which is itself a limitation worth stating
plainly:

- **Encoder**: the CLI (`f01-speculum-analyze -pilot-gen`), which drew
  words from a fixed pool using `math/rand` seeded from `time.Now()` —
  real wall-clock entropy, not a value visible in or derivable from the
  source code, so the operator could not have known the draw in advance
  from having written the program.
- **Decoder**: the operator, reading only the rendered ASCII state
  (`pilot/TRIAL_A_STATE.txt`, `pilot/TRIAL_B_STATE.txt`) and the declared
  parameters (`pilot/TRIAL_*_KNOWN_PARAMETERS.txt`) — never the ground
  truth file before recording a guess.
- **Controller**: the operator, running `-pilot-check` only after
  recording both guesses in the conversation transcript.

**Blinding caveat**: this relies entirely on the operator's self-discipline
not to open `GROUND_TRUTH_DO_NOT_OPEN_BEFORE_GUESSING.txt` before
guessing, not on any structural separation between roles. A genuine
multi-participant study (recommended for `task78`, see
`TASK76_REPORT.md`) would not have this weakness.

## Trial A — intact state, full K

- Declared parameters: `alphabet=latin23 read_radius=5
  order=inner_to_outer message_length=7`, state intact.
- Procedure actually followed: for word position `i = 0..6`, read ring
  `i`'s bracketed letter at sector 5 directly off the ASCII table
  (`ring 0` through `ring 6`, inside out, per the declared order).
- Decoded by hand: `M E M E N T O` → **MEMENTO**.
- Ground truth: **MEMENTO**.
- Result: **exact match.**
- Process notes: two file reads (state + parameters), zero re-reads of
  the state table, zero consultations of any instruction beyond the one
  parameters line (the decoding *procedure* itself did not need to be
  looked up mid-trial — it had already been internalized while writing
  `DecodeFull`, which is itself a limitation: an operator who had not
  just implemented the decoder would need to consult
  `F01_RECONSTRUCTION_DOSSIER.md` §7 first).

## Trial B — corrupted state, full K, error unannounced

- Declared parameters: `message_length=9`, state **may** be corrupted,
  operator not told whether or where.
- Decoded by hand, same procedure, rings 0–8: `S I L E P T I V M` →
  **SILEPTIVM**.
- Ground truth: `SILENTIVM`, with a single substitution injected at ring
  4 (`trial_b_damaged_ring=4`, matching word position 4 = `L/2` for
  `L=9`, i.e. exactly `Config.RingPos` applied to the corruption's own
  `mid` convention from `runCorruption`).
- Result: decoded string **matches the model's predicted corrupted
  output exactly** (`SILEPTIVM`), confirming the digital model's
  `DecodeWithGap`/substitution mechanics agree with a manual reading of
  the rendered state — this is the pilot's main methodological payoff:
  independent confirmation that the ASCII rendering and the underlying
  `State` it renders are faithful to each other, not just internally
  self-consistent code.
- Detection: `SILEPTIVM` does not parse as a recognizable Latin word to
  the operator (an unusual consonant cluster, "LEPT," where none of the
  12 pre-registered natural messages or the 36-word reference lexicon has
  anything similar) — the operator flagged it as likely corrupted *before*
  checking ground truth. This matches the model's `Detectable=true`
  prediction for single-substitution corruption of natural-language
  messages (85/95 detectable rows in `CORRUPTION_RESULTS.tsv`).
- Correction: the operator could not, from `SILEPTIVM` alone, identify
  which position was wrong or recover `SILENTIVM` specifically (nothing
  about the string points at position 4 over any other) — matching the
  model's general prediction that this device offers no redundancy for
  self-correction (`CorrectableWithoutM` is lexicon-dependent, not
  structural; see `TASK76_REPORT.md`'s discussion of why the 48/95
  "correctable" rate should not be read as a property of the device).

## Distinction: reading the device vs. recalling the message

Both trials involved reading a state the operator had never seen encoded
(the words were drawn after the pilot program was already written, using
wall-clock entropy) — there was no possibility of the operator "recalling"
the message from having chosen it. This is the cleanest test the pilot
format allows for task76's requirement to separate device-reading from
recall: **in both trials, 100% of the operator's answer is attributable
to reading the rendered state per the documented procedure, 0% to prior
memory of the message**, because the message was never available to
memorize in the first place.

## Limitations (stated explicitly, per task76's ban on overclaiming)

- N=1 operator, N=2 trials. No timing instrumentation beyond a coarse
  "number of file reads" proxy — no claim about human decoding speed is
  made or should be inferred.
- The operator is the same agent that designed and implemented the
  encoding/decoding model, which likely makes this an *easier* trial than
  a naive human subject would face (the operator already knew the
  procedure cold). `task78` should budget for a real naive-subject pilot
  if a stronger human-performance claim is wanted.
- Only two corruption/knowledge conditions were exercised by hand (full-K
  intact, full-K single-substitution). The combinatorial ablation
  conditions (K2–K7) were validated computationally
  (`ABLATION_RESULTS.tsv`), not by a human decoder attempting them by
  hand; that gap is intentional given the N=1 pilot's scope, not an
  oversight.
