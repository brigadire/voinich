# v2 preflight audit

Production is blocked. This audit performs no null or production scoring.

| Field | Value |
|---|---:|
| `LEXICON_SCOPE_JUSTIFIED` | `NO` |
| `FULL_STAR_LEXICON_REQUIRED` | `YES` |
| `CANONICAL_IDENTITIES_CURRENT` | `31` |
| `TABLE_SOURCE_GRAPHEMES` | `29` |
| `TABLE_TARGET_EVA_SYMBOLS` | `16` |
| `TABLE_UNIVERSE_INJECTIVE` | `24972508305` |
| `TABLE_UNIVERSE_ONE_MERGE` | `481457536` |
| `TABLE_UNIVERSE_SIZE` | `25453965841` |
| `BEAM_SIZE` | `256` |
| `BEAM_COVERAGE_FRACTION` | `1.0057371868852269e-08` |
| `SYNTHETIC_TABLES_SAMPLED_OUTSIDE_BEAM` | `YES` |
| `OUT_OF_BEAM_SYNTHETIC_CASES` | `18` |
| `OUT_OF_BEAM_SYNTHETIC_RECOVERY` | `FAIL` |
| `REAL_LABEL_REACHABILITY_KEEP` | `0.14285714285714285` |
| `REAL_LABEL_REACHABILITY_DROP` | `0.25` |
| `REAL_LABEL_REACHABILITY` | `0.25` |
| `KEEP_MODE_CAN_PRODUCE_EVA_ONLY` | `NO` |
| `DROP_MODE_COLLISION_RATE` | `0.08475825471698113` |
| `CROSS_PAGE_EXACT_MAPPING_CONSISTENCY_DEFINED` | `NO` |
| `CANONICAL_IDENTITY_CAPACITY_JUSTIFIED` | `NO` |
| `PRODUCTION_RUN_AUTHORIZED` | `NO` |
| `DECISION` | `BLOCK_PRODUCTION_UNTIL_LEXICON_BEAM_REACHABILITY_MAPPING_AND_CAPACITY_REVISIONS` |

## Interpretation

The current 31-identity lexicon leaves the stated coverage ceiling at 31/57. The exact bounded table universe is 25,453,965,841 tables (24,972,508,305 injective and 481,457,536 one-merge), while the current score-blind beam contains 256; its fraction is 1.006e-08.
Independent planted tables were sampled outside the beam (18/20). They cannot be recovered by a beam-only production search, so out-of-beam synthetic recovery is FAIL by construction.
Existential pair reachability before matching is KEEP=8/56 and DROP=14/56. DROP collision diagnostics are intentionally reported separately because deletion can create short accidental forms.
The protocol currently defines consistency only at the family level and has no historical basis for one global canonical identity capacity across both pages. Both must be revised before production.
