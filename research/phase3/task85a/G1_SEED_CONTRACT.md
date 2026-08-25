# G1 seed and deterministic-ordering contract

The seed is a pure function of the eight fields listed in
`G1_EXECUTABLE_CONTRACT.json`. Normalize strings to Unicode NFC, encode UTF-8,
join fields by the single separator byte `0x1f`, hash with SHA-256, and interpret
digest bytes 0–7 as unsigned little-endian uint64. Decimal integers have no sign
or leading zero. Expand that value with SplitMix64 twice into the 128-bit state
of PCG-XSL-RR-128/64.

Sampling sorts outcomes by Unicode code point; it maps the high 53 bits of one
uint64 to `[0,1)` and uses half-open cumulative intervals. Iteration over map or
set values is forbidden before floating-point accumulation or serialization:
sort keys first. Candidate rows use M0 through M5 order and lexicographic
parameter tuples. TSV and JSON use UTF-8/LF and deterministic field/key order.
Identical inputs therefore yield byte-identical outputs across restart, worker
count, filesystem order, map order, and `GOMAXPROCS`.
