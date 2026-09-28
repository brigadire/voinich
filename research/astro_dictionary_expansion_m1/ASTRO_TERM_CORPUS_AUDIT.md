# Astronomical term corpus audit

The expanded evidence table stores one attestation per row. `term_id` is the
concept key used by one-to-one matching; multiple spellings under a key do not
create additional dictionary capacity.

## Concepts and forms

| Class | D0 concepts | D1 concepts | Evidence rows |
|---|---:|---:|---:|
| PLANET_MOON | 7 | 7 | 19 |
| STAR | 21 | 31 | 63 |
| ZODIAC | 0 | 12 | 24 |

## Language layers

| Layer | Forms |
|---|---:|
| ARABIC | 50 |
| LATIN | 34 |
| LATINIZED_ARABIC | 22 |

## Date distribution

| Date | Forms |
|---|---:|
| 13th c. | 21 |
| 8th c. | 12 |
| 8th c.; ca. 1025 | 7 |
| ca. 1246 | 1 |
| ca. 1400 | 36 |
| late 9th c. | 29 |

## Source distribution

| Source | Forms |
|---|---:|
| KUNITZSCH_1987 | 1 |
| NMS_T1959_62_RETE | 21 |
| ORDO_PLANETARUM | 12 |
| ORDO_PLANETARUM;WALTERS_W73 | 7 |
| OXFORD_MHS_41468 | 36 |
| OXFORD_MHS_47632 | 29 |

## Variant statistics

Concepts with more than one attestation: **37**.
Attestation rows beyond one-per-concept: **56**.
Exact duplicate normalized forms across sources remain separate evidence rows
but collapse to one form before M1 search.

## Matching capacity

```text
STAR_LABELS=21
STAR_CONCEPTS=31
PLANET_LABELS=5
PLANET_CONCEPTS=7
THEORETICAL_MAX_COVERAGE=1.000000
```

ZODIAC evidence is audited but excluded from M1 because no frozen label has
that class. The capacity calculation therefore uses only STAR and PLANET_MOON.
