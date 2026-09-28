# Sealed Synthetic Generator Specification (v3 Design)

## 1. Objectives and Sealing Guarantees

The Sealed Synthetic Generator (`scripts/synthetic_generator.py`) produces synthetic star label corpora designed to benchmark search algorithms under realistic yet strictly controlled conditions.

### Strict Sealing Policy
1. **Separation of Generator and Solver**: The generator operates as an independent module. Solvers receive only the public dataset (label strings, occurrence IDs, page IDs) and the 94-identity historical lexicon.
2. **Unconstrained Ground-Truth Sampling**: Planted substitution tables are sampled **uniformly at random** from the complete 29-source by 16-target combinatorial universe ($\binom{29}{k} \times P(16, k)$). Hidden truth is **never** sampled from a pre-trimmed beam or candidate catalogue.
3. **Parameter Lock**: Algorithm hyperparameters are frozen prior to running the benchmark; no tuning is permitted post-unblinding.

## 2. Structural Fidelity to Frozen Target Scope

Every generated dataset faithfully reproduces the structural characteristics established in `REAL_SCOPE_STRUCTURAL_MANIFEST.json`:

| Property | Target Scope Property | Synthetic Generation Implementation |
|---|---|---|
| **Total Occurrences** | 57 | Exactly 57 labels per dataset |
| **Page Partition** | 30 on f68r1, 27 on f68r2 | Exact 30/27 split |
| **Alphabet** | 16-symbol EVA alphabet | Strict alphabet containment |
| **Length Distribution** | Mean 6.18, range 1–10 | Preserved through historical lexicon forms and distractor sampling |
| **Symbol Distribution** | Mean 5.14 distinct symbols | Naturally emerges from word lengths and table mapping |
| **Lexicon Source** | 94 canonical identities | Samples from `HISTORICAL_STAR_LEXICON.tsv` (307 attestations) |
| **Capacity Constraints** | `PER_PAGE_CAPACITY_1` | Shared prominent star identities across folios |

## 3. Parameter Grid and Perturbations

1. **Table Sizes**: $k \in \{4, 6, 8, 10, 12\}$.
2. **Mapping Modes**:
   - `INJECTIVE`: Strict 1-to-1 mappings.
   - `MERGE_1`: Exactly one pair of source graphemes maps to an identical target symbol.
3. **Deletion Modes**:
   - `DROP_UNMAPPED`: Unmapped source graphemes omitted.
   - `SELECTIVE_VOWEL_DROP`: Unmapped vowels dropped, unmapped consonants retained.
   - `NONE` (KEEP): Unmapped letters retained.
4. **Abbreviation Modes**:
   - `NONE`: Identity.
   - `SUSPENSION_1`: Final character dropped if length > 3.
   - `PREFIX_4`: Truncated to first 4 characters.
5. **Noise Corruptions**:
   - $0\%$: Clean transmission.
   - $10\%$: 10% random character substitutions (standard Gate S benchmark).
   - $25\%$: 25% random character corruptions (stress test).
6. **Unmatched Distractor Fractions**:
   - $0\%$, $25\%$, $50\%$ of occurrences replaced with unmatched distractors.
