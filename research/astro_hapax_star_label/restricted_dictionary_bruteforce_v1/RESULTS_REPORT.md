# Restricted dictionary brute-force v1 results

Primary conclusion: **ENGINE_INCONCLUSIVE**. Synthetic sensitivity: **ENGINE_SENSITIVITY_INSUFFICIENT**.
No specific star identification or decipherment claim is authorized.

| Endpoint | Selected system | Matched | Worst-family empirical p | FWER p | Best-null coverage advantage |
|---|---|---:|---:|---:|---:|
| JOINT | S000 | 0/57 | 1 | 1 | -0.000552632 |
| f68r1_TO_f68r2 | S000 | 0/27 | 1 | 1 | 0 |
| f68r2_TO_f68r1 | S000 | 0/30 | 1 | 1 | 0 |

The reported system is the deterministic optimum of the finite search. A tied
zero-score optimum is not a discovered correspondence. Selection evaluated all
64 systems, 57 LOO exclusions and 200 page-stratified subsamples. Level A and B
system rows are separated in SYSTEM_RESULTS.tsv; level C was excluded in advance.
No parameter was selected using hapax status or held-out page scores.

There are 90,000 completed informative null datasets (10,000 per family), plus
30,000 conservative historical independent-capacity searches. Every one repeats
joint selection and both train/held-out directions over the complete grid.
INFERENCE_RESULTS.tsv contains all 27 primary comparisons, their Bonferroni and BH
corrections, and nine conservative sensitivity comparisons. Null scores and selected
models are in NULL_RESULTS.tsv; checkpoints allow deterministic seed replay.
The exact order/identity-renaming controls have p=1 by construction and are listed
separately as invariance diagnostics, not informative randomizations.

Answers to the six task questions:

1. Astronomy superiority: NOT ESTABLISHED; all required control families and corrections were evaluated.
2. Cross-page transfer meeting protocol: NO.
3. Model-selection-aware null test: COMPLETE; worst-family joint corrected p=1.
4. Hapax minus non-hapax coverage=0.0; descriptive exact upper-tail p=1.0. This was not used in selection.
5. Published stable formal candidates=0; empty tables are intentional when the criteria fail.
6. Engine recovery=FAIL; results below. A real-data negative is not interpretable as absence of signal because sensitivity failed.

| Corruption fraction | Endpoint | Passing seeds |
|---|---|---:|
| 0 | JOINT | 20/20 |
| 0 | f68r1_TO_f68r2 | 20/20 |
| 0 | f68r2_TO_f68r1 | 19/20 |
| 0.1 | JOINT | 20/20 |
| 0.1 | f68r1_TO_f68r2 | 18/20 |
| 0.1 | f68r2_TO_f68r1 | 19/20 |
| 0.25 | JOINT | 20/20 |
| 0.25 | f68r1_TO_f68r2 | 18/20 |
| 0.25 | f68r2_TO_f68r1 | 17/20 |

Synthetic controls use 57 slots and 31 or fewer planted identities, at least 26
unmatched items, 30/27 split, 20 seeds per noise level, and 99 fully searched
synthetic nulls per dataset. PASS requires at least 18/20 per cell.

## Evidence boundaries

The 64 systems transform historical strings without learning a Latin-to-EVA cipher.
Thus a negative finding rejects only these transformations and this 31-identity
lexicon; it does not reject astronomical naming in the manuscript. Matching has
a ceiling of 31/57, imposed by historical identities, not 63 attestations.
Historical capacity blocks match spelling opportunities but are not historical
synonym sets; independent-capacity sensitivity gives controls more freedom.
Latin/Arabic language and date distributions are not fully balanced. Alchemical
forms come from an edited early-print witness of a medieval tradition. Circular
controls have only 29 independent observed occurrences and are bootstrapped to 57.
Bigram-preserving randomizations can retain most original words; their unchanged
fraction is reported, not concealed. They may offer little discriminatory power.
All comparisons concern anonymous matching, not a labeled star map.

The old enrichment script and COHORT_MEMBERS.tsv already differed from their
registered SHA256 on entry. They were not used as authoritative inputs. The target
and corpus metadata passed frozen checksums, and 26/31 hapax strata were recounted
independently. UPSTREAM_SNAPSHOT.json verifies that this task changed no upstream
files. See VALIDATION_REPORT.md and REPRODUCIBILITY.md.

## Mandatory final status

```text
RESTRICTED_DICTIONARY_BRUTEFORCE=COMPLETE
TARGET_SCOPE_VALID=YES
TARGET_LABELS=57
F68R1_LABELS=30
F68R2_LABELS=27
ASTRONOMICAL_LEXICON_FROZEN=YES
TRANSFORMATION_GRID_FROZEN=YES
SYNTHETIC_RECOVERY=FAIL
NULL_CONTROLS_COMPLETE=YES
BEST_SYSTEM_ID=S000
BEST_SYSTEM_COMPLEXITY=0
JOINT_MATCHED_COVERAGE=0.0
F68R1_TO_F68R2_HELDOUT_COVERAGE=0.0
F68R2_TO_F68R1_HELDOUT_COVERAGE=0.0
BEST_NULL_ADVANTAGE=-0.0005526315789473684
EMPIRICAL_P=1.0
CORRECTED_P=1
CROSS_PAGE_SIGNAL=NO
CORPUS_LEVEL_SIGNAL=NO
STABLE_PAIR_CANDIDATES=0
HAPAX_SIGNAL_ADVANTAGE=0.0
PRIMARY_CONCLUSION=ENGINE_INCONCLUSIVE
SPECIFIC_STAR_IDENTIFICATIONS_AUTHORIZED=NO
DECIPHERMENT_CLAIM_AUTHORIZED=NO
```
