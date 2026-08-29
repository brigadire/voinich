# Candidate generation state machine (not frozen)

The coherent PF-SC01 repair would use `BEFORE_TOKEN`, `INSIDE_TOKEN`, and
`TOKEN_COMPLETE`. In `BEFORE_TOKEN`, BOS and EOS are inadmissible; ordinary
glyphs and UNK are sampled from their conditional distribution using exactly
one U53. A sampled UNK emits U+FFFD and counts as a glyph. In `INSIDE_TOKEN`,
EOS terminates without being emitted; BOS remains impossible. Zero admissible
binary64 mass would produce the existing `GENERATION_FAILURE` before any draw.

This state machine closes the empty-token boundary itself, but is deliberately
marked **not frozen**. The parent artifacts do not uniquely define Generator
B, cap timing, retry-counter allocation, outcome ordering, or final corpus-byte
serialization. Filling those gaps would add scientific choices outside the
narrow PF-SC01 repair.

The candidate floating-point convention used by the diagnostic vectors is:
binary64 round-to-nearest/ties-to-even; retain frozen row order while filtering;
Neumaier sum admissible weights; divide each retained weight once by that mass;
build prefix sums with Neumaier; select the first positive-mass outcome for
which `u53 < cumulative`; if rounding leaves no such outcome, select the last
positive-mass admissible outcome. This is testable, but not normative V1.2.
