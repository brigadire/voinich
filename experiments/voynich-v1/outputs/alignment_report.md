# IVTFF alignment report

Result: **PASS**

- IVTFF file: `data/ZL3b-n.txt`
- frozen corpus file: `data_work/ZL3b-x7.txt`
- frozen corpus SHA256 (recorded now, not historical): `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`
- total pages: 227
- total loci: 5385
- skipped/comment-only loci: 0
- total frozen tokens: 39026
- aligned tokens: 39026
- mismatches: 0
- discovery token count: 39026
- indexing: zero-based token positions, boundaries are positions between token `p-1` and token `p`

The concatenation of all aligned frozen ranges is token-identical to the complete frozen corpus. The parser only creates an alignment representation and does not replace IVTT.

## Metadata coverage

| Metadata | Known tokens | Unknown tokens | Coverage |
|---|---:|---:|---:|
| Currier ($C) | 24573 | 14453 | 62.97% |
| Hand ($H) | 39026 | 0 | 100.00% |
| Folio | 39026 | 0 | 100.00% |
| Paragraph | 39026 | 0 | 100.00% |
| Quire ($Q; no folio heuristic) | 39026 | 0 | 100.00% |

Frozen/aligned first tokens: `fachys ykal ar ataiin shol shory cthres y kor sholdy`; last tokens: `sodal ch al chcthy chckhy qol ain ary oror sheey`. Their identity follows from the exact global invariant. Historical discovery metadata contains token count but no stored edge-token sample.

## Alignment-only normalization

IVTFF comments and `<%>`, `<$>` controls are omitted; `<->`, dots and commas become boundaries; the first explicit `[first:alternative]` reading is selected; braces retain their contents; `@NNN;`, apostrophes and `?` are preserved. Every rule is used solely to compare loci with the canonical frozen tokens.
