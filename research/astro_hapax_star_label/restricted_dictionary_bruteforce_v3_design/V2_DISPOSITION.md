# Disposition of Restricted Dictionary Brute-Force v2

## Formal Disposition Status

```text
V2_PREFLIGHT=FAILED
V2_PRODUCTION_RUN=NOT_AUTHORIZED
V2_SCIENTIFIC_RESULT=NONE
```

## Context and Audit Findings

Restricted dictionary brute-force v2 was designed to evaluate global substitution systems across the frozen target scope of 57 astronomical label occurrences (30 on f68r1 and 27 on f68r2). However, during preflight validation (`preflight.py`), v2 was blocked before executing any production search or null runs.

The quantitative preflight audit revealed critical theoretical and architectural blockers:

| Metric | v2 Value | Requirement / Interpretation |
|---|---:|---|
| `LEXICON_SCOPE_JUSTIFIED` | `NO` | Lexicon restricted to 31 canonical identities; ceiling 31/57 |
| `FULL_STAR_LEXICON_REQUIRED` | `YES` | Must cover full medieval astronomical tradition |
| `TABLE_UNIVERSE_SIZE` | `25,453,965,841` | Total bounded table universe (sizes 1–4) |
| `BEAM_SIZE` | `256` | Arbitrary lexicographic slice |
| `BEAM_COVERAGE_FRACTION` | `1.0057e-08` | Extremely small coverage of theoretical space |
| `SYNTHETIC_TABLES_SAMPLED_OUTSIDE_BEAM` | `YES` | 18 of 20 planted tables sampled outside beam |
| `OUT_OF_BEAM_SYNTHETIC_RECOVERY` | `FAIL` | Inability to recover tables outside the arbitrary slice |
| `REAL_LABEL_REACHABILITY_KEEP` | `0.1429` | Only 8 of 56 unique labels reachable under KEEP |
| `REAL_LABEL_REACHABILITY_DROP` | `0.2500` | Only 14 of 56 unique labels reachable under DROP |
| `KEEP_MODE_CAN_PRODUCE_EVA_ONLY` | `NO` | Non-EVA Latin letters (e.g., m, b, u, w, z) remain in output |
| `DROP_MODE_COLLISION_RATE` | `0.0848` | Deletion collapses distinct star names into identical tokens |
| `CROSS_PAGE_EXACT_MAPPING_CONSISTENCY_DEFINED` | `NO` | Consistency defined only at family level, not table level |
| `CANONICAL_IDENTITY_CAPACITY_JUSTIFIED` | `NO` | Global capacity=1 across both pages lacks historical foundation |
| `PRODUCTION_RUN_AUTHORIZED` | `NO` | Blocked prior to scoring |

## Detailed Breakdown of Blockers

### 1. Lexicon Scope and Ceiling
The v2 lexicon contained only 31 canonical star identities (from the D1 expansion), creating a hard mathematical cap of 31/57 (54.4%) maximum achievable coverage on the 57 target occurrences. A valid astronomical experiment requires an independently curated, historically attested star catalogue containing at least 57 canonical identities, ensuring that capacity constraints do not trivially enforce false negative results.

### 2. Beam Coverage and Synthetic Recovery Failure
The bounded table space of mappings from 29 source graphemes to 16 EVA symbols for sizes up to 4 contains 25,453,965,841 tables (24,972,508,305 injective and 481,457,536 single-merge tables). Rather than employing directional optimization, constraint satisfaction, or branch-and-bound pruning, v2 simply selected the first 256 tables lexicographically. When synthetic planted tables were sampled uniformly from the valid parameter space, 90% fell outside this arbitrary slice, leading to an immediate synthetic recovery failure.

### 3. Reachability and Mode Viability
- **KEEP Mode**: Leaving unmapped source characters intact produces hybrid tokens containing Latin graphemes that do not exist in the 16-symbol EVA alphabet (such as `m`, `b`, `g`, `u`, `w`, `z`). These tokens cannot match any real Voynich star label, rendering KEEP mode structurally incapable of producing valid candidate matches.
- **DROP Mode**: Unconditional deletion of unmapped graphemes caused an 8.48% collision rate across dictionary forms, collapsing distinct star identities into identical short tokens.
- **Table Size**: Table size 4 was structurally insufficient to bridge the character distance between historical star names and target labels, yielding a maximum reachability ceiling of only 25%.

### 4. Cross-Page Mapping Consistency and Capacity
In v2, cross-page consistency was defined only at the "family" level (e.g. both pages using some injective system), rather than requiring identical concrete character substitutions across pages f68r1 and f68r2. Furthermore, enforcing global capacity=1 across both pages was philologically and astronomically unjustified, as distinct folios or concentric rings in medieval astronomical manuscripts frequently depict shared prominent navigational and zodiacal stars.

## Conclusion and Package Status

Because production runs and null runs were blocked at preflight:
1. No production scores were calculated against the real star labels.
2. No dictionary matching assignments were made.
3. No astronomical signal or decipherment was evaluated or published.
4. The v2 codebase is preserved unchanged as an audit witness to these methodological findings.
5. All architectural revisions are segregated into the v3 design package.
