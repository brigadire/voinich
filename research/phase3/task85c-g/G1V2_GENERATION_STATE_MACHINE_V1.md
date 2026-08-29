# G1V2 generation state machine V1

`BEFORE_TOKEN` forbids BOS and EOS and conditionally samples ordinary glyphs
and UNK. An ordinary glyph is emitted verbatim; UNK emits U+FFFD. Both enter
`INSIDE_TOKEN` with length one. In `INSIDE_TOKEN`, EOS terminates without
emission. Emitting glyph 64 immediately enters `TOKEN_COMPLETE`: length 64 is
legal, termination is structural, and no 65th-glyph/EOS draw is consumed.

`TOKEN_COMPLETE` appends the nonempty logical token in occurrence order.
After the requested positive token count it enters `CORPUS_COMPLETE`, whose
only action is canonical serialization. M5 proposes a whole token per attempt;
valid NFC scalar length 1..64 enters `TOKEN_COMPLETE`, otherwise it advances to
the next attempt without RNG rollback. Zero admissible mass is the existing
`GENERATION_FAILURE` and consumes no draw.
