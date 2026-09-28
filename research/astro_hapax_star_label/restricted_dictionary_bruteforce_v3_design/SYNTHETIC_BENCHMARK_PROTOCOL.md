# Synthetic Benchmark Protocol (v3 Design)

## 1. Experimental Protocol and Benchmarking Design

This protocol defines the formal benchmarking procedure for evaluating search algorithms on synthetic star label recovery. The evaluation tests whether an algorithm can recover independently sampled substitution systems without brute-force enumeration and without reliance on arbitrary pre-trimmed candidate beams.

## 2. Benchmark Matrix

The benchmark systematically explores the multidimensional configuration space:

| Parameter | Levels Evaluated | Key Focus Levels |
|---|---|---|
| **Table Size ($k$)** | 4, 6, 8, 10, 12 | **8** (Authoritative baseline) |
| **Noise Rate** | 0.0, 0.10, 0.25 | **0.0, 0.10** (0.25 as stress test) |
| **Unmatched Fraction** | 0.0, 0.25, 0.50 | **0.25** |
| **Mapping Mode** | `INJECTIVE`, `MERGE_1` | `INJECTIVE` |
| **Deletion Mode** | `DROP_UNMAPPED`, `SELECTIVE_VOWEL_DROP` | `DROP_UNMAPPED` |
| **Abbreviation** | `NONE`, `SUSPENSION_1`, `PREFIX_4` | `NONE`, `SUSPENSION_1` |
| **Capacity Policy** | `PER_PAGE_CAPACITY_1`, `GLOBAL_CAPACITY_1` | `PER_PAGE_CAPACITY_1` |
| **Independent Seeds** | $\ge 20$ seeds per primary configuration | Seeds 101–120 |

## 3. Metrics Defined

For each benchmark trial, the following metrics are recorded:

1. **Exact Table Recovery (`exact_table_recovery`)**:
   Binary indicator ($1$ if found table $T_{\text{found}} = T_{\text{planted}}$, $0$ otherwise).
2. **Equivalence-Aware Recovery (`equiv_table_recovery`)**:
   Binary indicator ($1$ if found table produces identical encodings on the active lexicon vocabulary as $T_{\text{planted}}$).
3. **Mapping Precision and Recall (`mapping_precision`, `mapping_recall`)**:
   $$\text{Precision} = \frac{|T_{\text{found}} \cap T_{\text{planted}}|}{|T_{\text{found}}|}, \quad \text{Recall} = \frac{|T_{\text{found}} \cap T_{\text{planted}}|}{|T_{\text{planted}}|}$$
4. **Assignment Accuracy (`assignment_accuracy`)**:
   Fraction of labels correctly assigned to their true underlying star identity:
   $$\text{Accuracy} = \frac{\sum_{l} \mathbb{I}[M_{\text{found}}(l) = \text{TrueIdentity}(l)]}{|\text{Labels}|}$$
5. **Cross-Page Held-Out Transfer (`held_out_accuracy`)**:
   Models are selected on Page 1 (`f68r1`) and evaluated out-of-sample on Page 2 (`f68r2`), and vice versa, using the **identical concrete mapping table $T$**.
6. **Cross-Page Exact Mapping Consistency**:
   Must verify that the table used for cross-page validation is the exact same mapping table, not just a member of the same family.
7. **Order Invariance**:
   Shuffling the order of input labels must yield the exact same optimum.
8. **Checkpoint Identity**:
   Serializing state to disk and resuming execution must yield bit-identical results.

## 4. Gate S Minimum Passage Criteria

Passage of Gate S requires meeting or exceeding all pre-established thresholds:

```text
ZERO_NOISE_TABLE_RECOVERY >= 0.80
TEN_PERCENT_NOISE_TABLE_RECOVERY >= 0.70
TEN_PERCENT_HELDOUT_ASSIGNMENT_ACCURACY >= 0.60
TEN_PERCENT_MAPPING_PRECISION >= 0.75
TEN_PERCENT_MAPPING_RECALL >= 0.75
EXACT_SMALL_INSTANCE_OPTIMUM_RATE >= 0.95
ORDER_INVARIANCE = PASS
CHECKPOINT_IDENTITY = PASS
RESOURCE_FEASIBILITY = PASS
OUT_OF_CANDIDATE_SYNTHETIC_RECOVERY = PASS
```
