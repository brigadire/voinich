# Legacy mapping reproduction

Status: `PASS`.

| Metric | Reproduced |
|---|---:|
| PLANET/MOON | 7/7 |
| CIRCLE-SECTOR | 10/12 |
| STAR LABELS | 53/67 |
| UNMATCHED_COORDINATES | 54 |
| SENSITIVITY_ONLY_POSSIBLE_OCCURRENCES | 77 |

The unit is a distinct Stolfi `panel.group.number` physical coordinate, not a
transcriber variant. The 191 source rows collapse to 143 coordinates; 89 are
matched and 54 unmatched. For each unmatched coordinate the sensitivity count
uses the maximum token count among its variants, totaling 77. These 77 are
accounting bounds only and are not LABEL or recovered occurrences.

The legacy correspondences are deterministic rule-derived matches, not human
visual verification: same panel, a validated Stolfi-series to IVTFF locus-type
bridge, exact/wildcard lexical anchors, coordinate-peer consensus for variants,
and one-to-one span resolution. `MATCHED` means the frozen procedure assigned
an admissible ZL3b span; it does not mean a modern human confirmed the image
location. One physical coordinate may have transcriber variants and may map to
more than one token position when the physical run is multi-token.

The raw `labtit-98-07-20.idx` byte file is not present in this repository. Its
release, size (91,861 bytes), 1,485-row count, and SHA-256
`cb210aaa75dfd2e9d86e63fd4cff1684acdfc2669bd6a6f9969f4e6bfe10071c`
are documented by the frozen inventory audit; no local substitute was invented.
No builder for `STOLFI_ASTRO_LABEL_MATCHES.tsv` is retained locally. The frozen
row-level table and audit are therefore the reproducible migration inputs, while
the bias and legacy-enrichment builders referenced in the manifest remain
available and checksum-verified.
