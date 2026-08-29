# G1V2 generation semantics V1

This document explains the normative JSON of the same name. Explicit arrays
define their own order; mappings use NFC UTF-8 key order; fitted outcomes are
ordinary glyphs in NFC UTF-8 order followed by UNK and EOS. State constraints
filter that order and renormalize once with Neumaier summation.

Generator A uses direct inverse-CDF. Generator B remains independently
constructed: M0/M2/M3/M5 use a fully ordered exponential race, M1 uses the
specified deterministic Walker table, and M4 materializes cumulative matrix
rows. B never calls A. Every U53 increments one corpus-global draw index.

Length counts emitted Unicode scalar glyphs. Glyph 64 is legal and forces
structural completion without another draw. M5 attempts are exactly 0..1023,
retain consumed draws, and map exhaustion to `GENERATION_FAILURE`.

Canonical corpus bytes are NFC UTF-8, one token per line, LF after every token
including the last, and no other whitespace. The empty logical corpus maps to
zero bytes but is not a valid requested synthetic-control corpus.
