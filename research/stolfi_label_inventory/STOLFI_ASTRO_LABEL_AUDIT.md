# Stolfi Astronomical label inventory audit

## Verdict

`PARTIALLY_SUFFICIENT`. The inventory yields a reproducible positive `LABEL` set for 112 frozen token occurrences at 89 distinct Stolfi coordinates, but it does not support a reliable full `LABEL / NON_LABEL` classification. Coverage is uneven by panel and label family: the formally mapped coordinates cover 89/143 (62.2378%) physical Stolfi coordinates and 130/191 (68.0628%) source records, while whole f68v2 sector series remain unmapped. Treating every other token as `NON_LABEL` would therefore create known false negatives and would assume completeness that the source does not claim.

## Source and frozen inputs

- Primary inventory: Jorge Stolfi, [“A large list of labels and titles”](https://www.ic.unicamp.br/~stolfi/EXPORT/voynich/98-02-01-lotsa-labels/), machine-readable [`labtit-98-07-20.idx`](https://www.ic.unicamp.br/~stolfi/EXPORT/voynich/98-02-01-lotsa-labels/labtit-98-07-20.idx). The author page says it merges John Grove's list with Stolfi's list and was last edited `1998-07-20 20:54:07`; the file name fixes the release as `98-07-20`. Retrieved `2026-09-01`; 91,861 bytes; 1,485 records; SHA-256 `cb210aaa75dfd2e9d86e63fd4cff1684acdfc2669bd6a6f9969f4e6bfe10071c`.
- Frozen target transcription: `data/ZL3b-n.txt`, header `IVTFF Eva- 2.0 M 5`, `ZL version 3b of 13/05/2025`, SHA-256 `bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc`.
- Frozen occurrence index: `experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl`, SHA-256 `ba0342e15d8c468ec4e9f741e97cdb4a11938fe1f0ae3ac4338b73aaf1bd773a`.
- Canonical token corpus: `data_work/ZL3b-x7.canonical.txt`, SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`.
- Scope is exactly `f67r1, f67r2, f67v1, f68r1, f68r2, f68r3, f68v2, f68v1`; no other panel contributes a candidate.

The Stolfi extraction contains 191 source records at 143 physical coordinates. Multiple records at the same `panel.group.number` are independent transcriber readings of one physical label, so record coverage and physical-coordinate coverage are both reported and are not conflated.

## Matching protocol

1. Parse all 11 pipe-delimited fields and retain every source record on the eight panels, including records whose Stolfi class is `?`; panel membership, not the interpretive class field, defines scope.
2. Normalize representation only: IVTFF comments/control markers are removed, the frozen first branch of bracketed alternatives is selected, braces retain their contents, and EVA/IVTFF dot, comma, hyphen, apostrophe, question-mark and physical-gap separators become token boundaries. Case and glyph strings are preserved. Stolfi `*` is allowed only as an explicit one-or-more-glyph wildcard.
3. Search only within one frozen ZL3b locus and the same panel. No edit distance, phonetic equivalence, semantic similarity, image inspection, or cross-panel vocabulary match is allowed.
4. A Stolfi panel/group is bridged to an IVTFF locus type only when at least two distinct coordinates have unique exact lexical anchors and every such anchor agrees on that locus type. This prevents a unique same-page prose word from being mistaken for the physical label; e.g. f68v2.X.1 `okeody` has a paragraph hit but the X series has no validated ZL3b coordinate bridge.
5. Alternate transcriber records at one Stolfi coordinate share a physical referent. A unique peer anchor may therefore carry a disagreeing reading to the same frozen span. Candidate spans are then solved one-to-one across distinct Stolfi coordinates. Unresolved multiple spans are `AMBIGUOUS`; no choice is made manually.

This protocol found no residual ambiguous record: 0/191. That zero is not evidence of completeness; 61 records have no admissible mapping.

## Panel results

| panel | tokens | records | matched | unmatched | ambiguous | record coverage | physical coordinates matched |
|---|---:|---:|---:|---:|---:|---:|---:|
| f67r1 | 167 | 24 | 20 | 4 | 0 | 0.833333 | 10/12 |
| f67r2 | 191 | 26 | 20 | 6 | 0 | 0.769231 | 13/19 |
| f67v1 | 73 | 10 | 4 | 6 | 0 | 0.400000 | 4/10 |
| f68r1 | 69 | 29 | 24 | 5 | 0 | 0.827586 | 24/29 |
| f68r2 | 89 | 50 | 42 | 8 | 0 | 0.840000 | 21/25 |
| f68r3 | 110 | 12 | 7 | 5 | 0 | 0.583333 | 7/12 |
| f68v2 | 104 | 32 | 9 | 23 | 0 | 0.281250 | 6/28 |
| f68v1 | 98 | 8 | 4 | 4 | 0 | 0.500000 | 4/8 |
| **TOTAL** | **901** | **191** | **130** | **61** | **0** | **0.680628** | **89/143** |

`tokens` comes from the frozen occurrence index. `Stolfi labels` in the TSV means source records, as required by the record-level coverage definition.

## Representation of major astronomical label families

| family (operationally from Stolfi fields/coordinates) | physical coordinates | mapped | assessment |
|---|---:|---:|---|
| circle sectors (f67r1) | 12 | 10 | represented, high but incomplete |
| planet/moon labels | 7 | 7 | represented, complete in this inventory slice |
| star labels | 67 | 53 | represented, high but incomplete |
| radial text (f68v2.R) | 12 | 6 | represented, half mapped |
| sector/inside-star series (f68v2.X/Y) | 15 | 0 | present in Stolfi, unusable against frozen ZL3b |
| outer title (f68v2.Z) | 1 | 0 | present, unmapped |
| other diagram labels | 29 | 13 | selective/incomplete |

Thus the inventory has records on all eight panels and does include the principal astronomical families, not merely isolated examples. The *usable mapping*, however, is strongly concentrated in planet and star series and fails an entire sector-label family. Per-panel usable data is 8/8 only in the minimal sense of at least one unambiguous match; f67v1, f68v2, and f68v1 do not exceed 50% physical-coordinate coverage.

## Classification consequence

The 112 mapped frozen token positions can be labeled `LABEL` without image interpretation under the stated protocol. The complement cannot safely be labeled `NON_LABEL`: 54 known Stolfi physical coordinates (61 source records) remain unmatched, and Stolfi describes the resource as a “large list,” not as an exhaustive token-level annotation of these panels. The inventory is therefore useful as partial independent positive supervision, not as a standalone binary gold standard.

No hapax statistic or hapax test was computed.

## Final status

```text
STOLFI_ASTRO_LABEL_INVENTORY=PARTIALLY_SUFFICIENT
MATCHED_LABEL_COVERAGE=130/191
PANELS_WITH_USABLE_LABEL_DATA=8/8
AMBIGUOUS_MATCH_RATE=0/191
```
