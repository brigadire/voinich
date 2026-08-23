# Task78 experimental protocol

## Freeze point and scope

The candidate tiers, profiles in `profiles.json`, outcomes, test messages and
seed `780823` are fixed before interpreting `EXPERIMENT_RESULTS.tsv`. Trial
order is fixed rather than randomized because every formal transition is
deterministic and order-independent; the seed is reserved in the manifest so
future stochastic corruption cannot silently choose a new sequence. The
program reads no Voynich corpus, fingerprint or Phase-I result. A successful
digital cycle validates the declared profile only; it does not promote an `H`
parameter to historical evidence.

The primary outcome for literal/lookup profiles is exact recovery. Positional
symbol accuracy, compatible interpretation count, state-space/cycle size and
error class are secondary outcomes. For F12 the primary formal outcome is
correct cue emission plus correct learned lookup; this is not treated as human
recall data. No trials are excluded. Profiles are never pooled.

## Conditions

- F08 R0: six fixed Latin strings; baseline, start/direction/boundary and
  association ablations; substitution, missing insertion, adjacent swap,
  frame collapse, duplication, start shift, reversal and topology-preserving
  geometric deformation. R1 exists only to expose the hypothetical empty-hole
  stop rule and is not selected by performance.
- F07 R0: all 23 selector positions, one full unit-step cycle and one shift.
- F10 R0/R1: all seven-symbol messages in the fixed set; full alignment,
  route ablation, independent-band shift, and suffix-coupled sensitivity.
- F11 R0: six occupied indices; lookup, missing slot, swap and unknown index
  convention.
- F12 R0: a twelve-tick cycle with three scheduled cues; trained, untrained
  and wrong-convention formal decoders. The twelve-tick period and mapping are
  `H` test parameters.

For binomial baseline proportions, the summary reports two-sided Wilson 95%
intervals. Small `n` is descriptive. There are no human participants, delays
or latency measurements in task78; therefore recall-rate generalization and a
human-versus-formal comparison remain `INCONCLUSIVE`.

Run from repository root:

```sh
GOCACHE=/tmp/task78-gocache go run ./research/phase2/fontana/task78-analyze
```
