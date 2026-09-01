# Cross-section reuse of confirmed Astronomical labels

## Result

`ASTRO_LABEL_CROSS_SECTION_PATTERN=SUBSTANTIAL_CROSS_SECTION_REUSE`.

The 112 confirmed Astronomical label occurrences collapse to 108 token types. Of those, 65/108 (0.601851852) occur somewhere outside Astronomical in the frozen corpus; 53/108 (0.490740741) occur in Herbal and 37/108 (0.342592593) in Pharmaceutical. There are 42 global-hapax types. Cross-section reuse therefore reaches a majority of confirmed types, while globally unique forms remain a large minority. Astronomical labels use both shared manuscript vocabulary and locally rare forms rather than behaving mainly as unique identifiers.

The status rule is frozen as follows: `MAINLY_GLOBAL_UNIQUE` if at least half of types are global hapax; otherwise `SUBSTANTIAL_CROSS_SECTION_REUSE` if at least half occur outside Astronomical; otherwise `MAINLY_SECTION_LOCAL` if at least half are confined to Astronomical; otherwise `MIXED`.

## Inputs and normalization

- Confirmed positives only: rows marked `MATCHED` in `STOLFI_ASTRO_LABEL_MATCHES.tsv`, deduplicated first by absolute occurrence and then by frozen token type. The complement and unmatched Stolfi records are never treated as `NON_LABEL`.
- Frozen corpus: all 39,380 occurrence records. Section codes are cross-checked against the broad visual taxonomy before counting.
- Type identity is the repository's frozen EVA composite normalization (`cth/ckh/cph/cfh/iin/ain/ch/sh/ee/in` collapsed to atomic symbols). The output `token` column expands those atoms back to canonical basic EVA for readability; `canonical_token_key` preserves the atomic key.
- Cross-label input: all 1485 records in Stolfi `labtit-98-07-20.idx`. Outside-Astronomical records whose object is `title?` are excluded, leaving 1222 eligible label records. Each exact non-wildcard component of a multi-token label is checked. Absence from this inventory is not interpreted as `NON_LABEL`.

## Main counts

| measure | count | fraction of 108 types |
|---|---:|---:|
| global hapax | 42 | 0.388888889 |
| occurs outside Astronomical | 65 | 0.601851852 |
| occurs in Herbal | 53 | 0.490740741 |
| occurs in Pharmaceutical | 37 | 0.342592593 |
| Astronomical-section hapax but repeated outside | 38 | 0.351851852 |
| independently listed by Stolfi as a label outside Astronomical | 32 | 0.296296296 |

Classification counts: `GLOBAL_HAPAX=42`, `ASTRO_LABEL_ONLY=1`, `LABEL_REUSED_ACROSS_SECTIONS=32`, `ASTRO_LABEL_BUT_RUNNING_TEXT_ELSEWHERE=33`.

Outside-Astronomical Stolfi label reuse by listed section: `Biological=11`, `Cosmological=7`, `Pharmaceutical=15`, `Unknown=2`, `Zodiac=15`. No Herbal cross-label identification is asserted: the Stolfi Herbal slice consists of `title?` records, which are excluded from independent visual-label status.

## By family

| family | types | global hapax | outside Astro | Herbal | Pharmaceutical | outside Stolfi label |
|---|---:|---:|---:|---:|---:|---:|
| CIRCLE_SECTOR | 14 | 2 | 12 | 9 | 10 | 9 |
| OTHER | 22 | 7 | 15 | 12 | 9 | 8 |
| PLANET_MOON | 9 | 5 | 4 | 4 | 3 | 3 |
| RADIAL_TEXT | 13 | 5 | 8 | 8 | 4 | 5 |
| STAR | 54 | 23 | 30 | 24 | 15 | 11 |


Family membership is multi-label at the type level: if one normalized type occurs among confirmed labels from more than one family, it contributes once to each relevant family row. Family totals therefore need not sum to the all-type total. The complete STAR-only table is provided separately.

## Most frequent cross-section forms

| token | families | Astro | Herbal | Pharmaceutical | other sections | global |
|---|---|---:|---:|---:|---:|---:|
| `daiin` | OTHER | 11 | 460 | 98 | 282 | 851 |
| `aiin` | RADIAL_TEXT | 12 | 111 | 35 | 353 | 511 |
| `ar` | RADIAL_TEXT | 13 | 84 | 6 | 308 | 411 |
| `chol` | STAR | 6 | 228 | 44 | 120 | 398 |
| `dar` | CIRCLE_SECTOR;OTHER | 10 | 117 | 25 | 172 | 324 |
| `o` | STAR | 10 | 73 | 18 | 122 | 223 |
| `okeey` | CIRCLE_SECTOR;OTHER | 5 | 15 | 23 | 141 | 184 |
| `okal` | OTHER;PLANET_MOON | 2 | 39 | 5 | 110 | 156 |
| `oteey` | CIRCLE_SECTOR | 7 | 12 | 5 | 121 | 145 |
| `okain` | PLANET_MOON | 1 | 19 | 1 | 120 | 141 |
| `shy` | RADIAL_TEXT | 1 | 69 | 5 | 24 | 99 |
| `am` | PLANET_MOON | 4 | 21 | 3 | 60 | 88 |
| `otol` | STAR | 3 | 31 | 7 | 41 | 82 |
| `sar` | CIRCLE_SECTOR | 4 | 14 | 6 | 57 | 81 |
| `odaiin` | STAR | 1 | 26 | 5 | 36 | 68 |


## Interpretation boundary

The literal answer is substantial cross-section reuse: a majority of confirmed Astronomical label types occur elsewhere, including many Herbal and Pharmaceutical occurrences. At the same time, 42/108 types are global hapax, so the repertoire is heterogeneous rather than uniformly shared. A corpus occurrence outside Astronomical is running-text-or-label reuse unless the independent Stolfi cross-label check also identifies it as a label. Conversely, absence from Stolfi cannot prove running-text-only status because the inventory is incomplete and transcription versions differ.

No semantic identity is inferred from identical token form, and no claim is made that a repeated form names the same object across sections.

## Final status

```text
ASTRO_LABEL_CROSS_SECTION_PATTERN=SUBSTANTIAL_CROSS_SECTION_REUSE
ASTRO_LABEL_TYPES=108
GLOBAL_HAPAX_TYPES=42
OUTSIDE_ASTRO_TYPES=65
HERBAL_REUSED_TYPES=53
PHARMACEUTICAL_REUSED_TYPES=37
CROSS_SECTION_STOLFI_LABEL_TYPES=32
```
