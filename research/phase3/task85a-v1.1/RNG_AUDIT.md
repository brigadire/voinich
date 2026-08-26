# RNG audit

The freeze determines seed-field encoding, SHA-256/U64LE conversion, two
SplitMix64 outputs, the PCG-XSL-RR-128/64 output family, and sampling mapping.
It does not determine an LCG stream selector/increment or a warm-up transition.
PCG-XSL-RR-128/64 names an output permutation and state size, not Task86R's
particular fixed stream. The source itself acknowledges this omission.

Task86R chooses multiplier `2360ed051fc65da4:4385df649fccf645`, increment
`5851f42d4c957f2d:14057b7ef767814f`, initializes the state from the two
SplitMix outputs, and advances once before the first `Uint64`, which advances
again. A contract-consistent initialization with no warm-up changes the first
draw for seed 42. Therefore generated samples can change and no output-level
equivalence follows from determinism alone.

`R1_RNG = SCIENTIFIC_COMPLETION`.

The frozen Task86R implementation is now pinned by `RNG_TEST_VECTORS.tsv`;
these vectors document history and do not amend the preregistration.

