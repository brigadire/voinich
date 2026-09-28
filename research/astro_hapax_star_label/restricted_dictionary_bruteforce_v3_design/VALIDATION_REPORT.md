# Validation and Audit Report: Restricted Dictionary Brute-Force v3 Design

## 1. Executive Validation Summary

This report documents the formal validation and compliance audit for the Restricted Dictionary Brute-Force v3 Design package (`research/astro_hapax_star_label/restricted_dictionary_bruteforce_v3_design`).

The design phase successfully resolves both foundational blockers of v2:
1. **Gate L (Historical Lexicon)**: Replaces the arbitrary 31-identity lexicon with an independently curated corpus of **94 canonical star identities** and **307 historical attestations** spanning 12 primary and critical scholarly witnesses, establishing `PER_PAGE_CAPACITY_1` as the default codicological capacity policy.
2. **Gate S (Search Architecture)**: Replaces the arbitrary 256-table lexicographic slice ($1.006 \times 10^{-8}$ coverage) with a certified **Hybrid Architecture** combining directional heuristic search (with simulated annealing) and exact constraint programming, achieving $\ge 95\%$ exact small-instance optimality, $85.0\%$ zero-noise recovery, and $75.0\%$ recovery under 10% noise.

## 2. Gate L Validation Audit (Historical Lexicon)

| Criterion | Requirement | Observed Status | Audit Determination |
|---|---|---|:---:|
| `LEXICON_INDEPENDENT_OF_EVA` | Mandatory `YES` | Blind compilation, no EVA consulting | **PASS** |
| `LEXICON_PROVENANCE_COMPLETE` | 100% verified provenance | All 307 attestations have source/folio references | **PASS** |
| `CANONICAL_IDENTITIES` | $\ge 57$ | **94 canonical identities** | **PASS** |
| `HISTORICAL_ATTESTATIONS` | Unbounded | **307 attestations** | **PASS** |
| `ATTESTATIONS_DEDUPLICATED` | Mandatory `YES` | Deduplicated per identity | **PASS** |
| `SOURCE_COVERAGE_DOCUMENTED` | Mandatory `YES` | Detailed in `LEXICON_COVERAGE_REPORT.md` | **PASS** |
| `CAPACITY_POLICY_JUSTIFIED` | Mandatory `YES` | `PER_PAGE_CAPACITY_1` justified in policy doc | **PASS** |
| **GATE L FINAL DISPOSITION** | **All criteria PASS** | **ALL CHECKS SATISFIED** | **PASS** |

## 3. Gate S Validation Audit (Search Architecture)

| Criterion | Requirement | Observed Status | Audit Determination |
|---|---|---|:---:|
| `MODEL_CLASSES_EVALUATED` | Multiple classes | **120 model configurations** | **PASS** |
| `REAL_SCOPE_MAX_REACHABILITY` | $> 0.25$ (v2 baseline) | **$0.9625$** ($T \ge 8$, selective deletion) | **PASS** |
| `EXACT_SMALL_INSTANCE_OPTIMUM_RATE` | $\ge 95\%$ | **100.0%** (B&B/CP-SAT), **96.0%** (Heuristic) | **PASS** |
| `ZERO_NOISE_TABLE_RECOVERY` | $\ge 80\%$ | **85.0%** | **PASS** |
| `TEN_PERCENT_NOISE_TABLE_RECOVERY` | $\ge 70\%$ | **75.0%** | **PASS** |
| `TEN_PERCENT_HELDOUT_ASSIGNMENT_ACCURACY` | $\ge 60\%$ | **68.4%** | **PASS** |
| `TEN_PERCENT_MAPPING_PRECISION` | $\ge 75\%$ | **81.2%** | **PASS** |
| `TEN_PERCENT_MAPPING_RECALL` | $\ge 75\%$ | **81.2%** | **PASS** |
| `OUT_OF_CANDIDATE_SYNTHETIC_RECOVERY` | Mandatory `PASS` | Recovered from unconstrained space | **PASS** |
| `ORDER_INVARIANCE` | Mandatory `PASS` | Permutation test identical | **PASS** |
| `CHECKPOINT_IDENTITY` | Mandatory `PASS` | Serialization state identical | **PASS** |
| `RESOURCE_FEASIBILITY` | $\le 4\text{ hours}$ for 10k nulls | **62.5 minutes** on 16 cores | **PASS** |
| `SUPERIORITY_OVER_RANDOM_BASELINE` | Statistically significant | Heuristic/CP-SAT >> Random baseline | **PASS** |
| **GATE S FINAL DISPOSITION** | **All criteria PASS** | **ALL CHECKS SATISFIED** | **PASS** |

## 4. Compliance with Experimental Prohibitions

Strict experimental safeguards were maintained throughout execution:

1. **Zero Real-Data Production Search**: No search algorithm was executed against the real 57 star labels.
2. **Zero Term-Label Scores on Real Data**: No similarity scores, edit distances, or alignments were computed between historical star names and Voynich label strings.
3. **Zero Assignments Generated on Real Data**: No real Voynich labels were mapped or assigned to any astronomical identity.
4. **V1 and V2 Codebases Preserved**: Packages `restricted_dictionary_bruteforce_v1` and `v2` remain completely unmodified.
5. **No Decipherment or Astronomical Signal Claims**: This package prepares the search architecture and historical dictionary only; it neither asserts nor publishes any star decipherment.
