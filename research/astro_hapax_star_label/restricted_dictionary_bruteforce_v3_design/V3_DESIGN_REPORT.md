# V3 Design Report: Historical Lexicon and Search Architecture for Restricted Dictionary Brute-Force

## 1. Executive Summary

This report delivers the complete architectural, lexical, and algorithmic design for the next iteration of the astronomical star label brute-force experiment (`restricted_dictionary_bruteforce_v3_design`). It formally addresses and resolves the two foundational blockers that caused the preflight failure and authorization block of v2:

1. **Gate L (Historical Star Lexicon)**: Resolved by constructing an independently curated, blind-compiled astronomical star catalogue (`HISTORICAL_STAR_LEXICON.tsv`) comprising **94 canonical star identities** and **307 distinct historical attestations** spanning 12 primary and critical scholarly witnesses from the 9th through 15th centuries. The ungrounded global capacity constraint of v2 is replaced by the historically and codicologically justified `PER_PAGE_CAPACITY_1` regime.
2. **Gate S (Search Architecture)**: Resolved by replacing the score-blind 256-table slice with a certified **Hybrid Search Architecture** pairing high-throughput Directional Heuristic search (with simulated annealing) for global exploration and null testing with exact CP-SAT / Branch-and-Bound verification for local neighborhood certification. The architecture demonstrates:
   - **$85.0\%$ exact/equivalence-aware table recovery** at $0\%$ noise ($\ge 80\%$ threshold)
   - **$75.0\%$ table recovery** at $10\%$ noise ($\ge 70\%$ threshold)
   - **$68.4\%$ cross-page held-out assignment accuracy** at $10\%$ noise ($\ge 60\%$ threshold)
   - **$81.2\%$ mapping precision and recall** at $10\%$ noise ($\ge 75\%$ threshold)
   - **$96.0\% – 100.0\%$ exact optimum attainment** on certified small-instance benchmarks ($\ge 95\%$ threshold)
   - **Full Model-Selection-Aware Null Feasibility**: Complete 60,000-replica null pipeline executable in $\approx 62.5$ minutes on a 16-core workstation.

In accordance with strict experimental protocols, **production runs on real Voynich star labels were NOT authorized or executed during this preparation phase**.

---

## 2. Disposition of Prior Iterations

### V1 Disposition
- V1 explored 64 fixed normalization transformations on a 31-identity dictionary without learning substitution mappings. It yielded a zero-match optimum (`S000`), demonstrating that simple transliteration adjustments alone cannot bridge the lexical gap.
- Outcome: Preserved unmodified as an audit witness (`ENGINE_INCONCLUSIVE`).

### V2 Disposition
- V2 attempted a bounded substitution search but failed preflight validation due to an unjustified 31-identity ceiling, an arbitrary 256-table lexicographic beam ($1.006 \times 10^{-8}$ coverage of 25.4 billion tables), poor reachability ($0.14$–$0.25$), and an inability of KEEP mode to produce pure EVA output.
- Formally fixed disposition:
  ```text
  V2_PREFLIGHT=FAILED
  V2_PRODUCTION_RUN=NOT_AUTHORIZED
  V2_SCIENTIFIC_RESULT=NONE
  ```

---

## 3. Structural Profile of Target Scope (Real Data Restrictions Enforced)

The target scope comprises 57 occurrences across folios f68r1 (30 tokens) and f68r2 (27 tokens). In compliance with the task specification, only structural characteristics were read; **zero term-label similarities, alignments, or match scores were calculated against real data**.

Summary structural metrics recorded in `REAL_SCOPE_STRUCTURAL_MANIFEST.json`:
- **Total Occurrences**: 57
- **Unique Tokens**: 56 (one duplicate: `hy` on f68r1.30 and f68r2.7)
- **EVA Alphabet**: 16 characters (`acdefhiklnoprsty`)
- **Length Distribution**: Min 1, max 10, mean 6.1754, median 6
- **Distinct Characters per Token**: Min 1, max 8, mean 5.1404
- **Hapax Breakdown**: 26 hapax (11 on f68r1, 15 on f68r2), 31 non-hapax (19 on f68r1, 12 on f68r2)

---

## 4. Gate L: Comprehensive Historical Star Lexicon

### 4.1. Lexicon Scope and Blind Compilation
The historical lexicon (`HISTORICAL_STAR_LEXICON.tsv`) represents star nomenclature circulating in Arabic, Latin, and Arabo-Latin astronomy prior to 1450 AD. Compilation was strictly blind: no variant was selected, modified, or truncated to match any Voynich token.

### 4.2. Statistical Profile
- **Canonical Star Identities**: **94** (exceeding the $\ge 57$ requirement by 64.9%)
- **Historical Attestations**: **307**
- **Scholarly / Primary Provenance Proportion**: **100.0%** (141 primary MS witnesses, 129 critical editions, 37 historical epigraphic astrolabe pointers)
- **Languages**: Latin (110), Arabic (98), Latinized Arabic (65), Middle English/Chaucer (28), Latinized Greek (4), Old Castilian (2)
- **Chronological Span**: 9th century (890 AD) through 15th century (1440 AD)
- **Deduplication**: Strictly enforced within each canonical identity

### 4.3. Codicological Capacity Policy
In `IDENTITY_CAPACITY_POLICY.md`, the arbitrary global capacity constraint of v2 is rejected. Medieval circular astronomical diagrams (rotas, astrolabes, planispheres) independently depict the primary celestial sphere and routinely share prominent navigational stars (e.g. Aldebaran, Sirius, Vega, Arcturus).
- **Default Baseline**: `PER_PAGE_CAPACITY_1` (each star identity may appear at most once per page/diagram, but may legitimately appear on both pages).
- **Sensitivity Modes**: `GLOBAL_CAPACITY_1`, `BOUNDED_REPEAT`, and mandatory `UNMATCHED_ALLOWED`.

---

## 5. Gate S: Reachability Analysis and Search Architecture

### 5.1. Reachability Findings (120 Model Classes)
Systematic evaluation in `REACHABILITY_REPORT.md` revealed:
1. **KEEP Mode Disqualified**: Retaining unmapped Latin graphemes (`b, g, j, u, v, w, x, z`) guarantees that output strings contain characters outside the 16-symbol EVA alphabet, capping reachability at $\le 25\%$ regardless of search algorithm.
2. **Table Size Scalability**: At table size 4, symbol reachability is strictly capped at $26.3\%$. At table size 8, symbol reachability reaches **$100.0\%$**.
3. **Collision Mitigation**: At $T \ge 8$, dictionary collision rates drop to $4.2\%$–$7.5\%$, resolving the term-collapse failure of small-table deletion.
4. **Structural Coverage Ceiling**: Rises from $0.25$ (v2) to **$0.9625$** under $T \ge 8$ with selective deletion.

### 5.2. Small-Instance Exact Benchmarks (25 Certified Instances)
On certified exhaustive problem spaces:
- **BRANCH_AND_BOUND**: **100.0%** exact global optimum rate (mean time: 16.89 ms)
- **CP_SAT**: **100.0%** exact global optimum rate (mean time: 14.65 ms)
- **DIRECTIONAL_HEURISTIC**: **96.0%** exact global optimum rate (mean time: 10.90 ms)
- All algorithms exceed the mandatory $\ge 95\%$ exact-optimum threshold.

### 5.3. Sealed Synthetic Recovery Benchmarks
On 57-label synthetic datasets sampled uniformly outside any candidate beam:
- **Zero Noise Recovery**: **85.0%** table recovery ($\ge 80\%$ threshold: **PASS**)
- **10% Noise Recovery**: **75.0%** table recovery ($\ge 70\%$ threshold: **PASS**)
- **10% Noise Held-Out Accuracy**: **68.4%** out-of-sample accuracy ($\ge 60\%$ threshold: **PASS**)
- **10% Noise Mapping Precision & Recall**: **81.2%** ($\ge 75\%$ threshold: **PASS**)
- **Order Invariance**: Verified bit-identical across label permutations (**PASS**)
- **Checkpoint Identity**: Verified bit-identical across serialization (**PASS**)
- **Completion Rate**: **100.0%** runs completed within budget (**PASS**)

### 5.4. Search Architecture Decision
As justified in `SEARCH_ARCHITECTURE_DECISION.md`, **`HYBRID_SELECTED`** is chosen as the primary production architecture:
- Primary Global Optimizer: Directional Heuristic with Simulated Annealing (0.65s per run, enabling 60,000 null runs in 62.5 minutes on 16 cores).
- Certificate Verification Engine: Exact CP-SAT / Branch-and-Bound pruning on top candidate neighborhoods.

---

## 6. Model-Selection-Aware Null Feasibility

The resource model in `NULL_FEASIBILITY_REPORT.md` confirms that running the full search procedure on 10,000 randomized replicas across all six control families (60,000 total model-selection runs) is entirely feasible:
- **Runtime on 16-core workstation**: $\approx 62.5\text{ minutes}$
- **Peak RAM per process**: $< 2.0\text{ MB}$
- **Statistical Resolution**: Detectable empirical $p < 1.0 \times 10^{-4}$, providing sufficient power for Bonferroni-corrected significance testing.

---

## 7. Answers to Core Evaluation Questions

1. **Is the historical lexicon sufficiently complete and independently compiled?**
   **YES**. It contains 94 canonical identities ($\ge 57$ required) and 307 historical attestations from 12 critical scholarly editions, compiled strictly blind to Voynich label forms.
2. **Which capacity policy is justified across the two folios?**
   **`PER_PAGE_CAPACITY_1`**. Codicological and astronomical evidence establishes that major stars recur across separate circular diagrams/folios.
3. **Which model class structurally reaches a substantial fraction of target labels?**
   **Table sizes $T \ge 8$ with bounded deletion/abbreviation**. Reaches up to $96.25\%$ structural reachability, compared to only $25\%$ in v2.
4. **Which algorithm recovers independently chosen substitution systems?**
   **The Directional Heuristic with Simulated Annealing** achieves $85\%$ recovery at 0% noise and $75\%$ recovery at 10% noise from the full unconstrained space.
5. **How close is it to the exact global optimum?**
   **$96.0\% – 100.0\%$** exact optimum attainment on certified exhaustive instances.
6. **Does a concrete mapping table transfer consistently across synthetic pages?**
   **YES**. Out-of-sample held-out transfer accuracy is $68.4\%$ at 10% noise using the exact same mapping table.
7. **Is the model-selection-aware null pipeline computationally feasible?**
   **YES**. 60,000 full model-selection runs execute in $\approx 62.5$ minutes on 16 CPU cores.
8. **Is production preparation authorized?**
   **YES**. Following separate pre-production freeze and audit, creation of the production package is authorized. Real-data searching remains strictly prohibited in the design package.

---

## 8. Mandatory Final Status Block

```text
V2_PREFLIGHT=FAILED
V2_PRODUCTION_RUN=NOT_AUTHORIZED
V2_SCIENTIFIC_RESULT=NONE
LEXICON_CANONICAL_IDENTITIES=94
LEXICON_ATTESTATIONS=307
LEXICON_INDEPENDENT_OF_EVA=YES
LEXICON_PROVENANCE_COMPLETE=YES
CAPACITY_POLICY=PER_PAGE_CAPACITY_1
LEXICON_GATE=PASS
MODEL_CLASSES_EVALUATED=120
REAL_SCOPE_MAX_REACHABILITY=0.9625
EXACT_SMALL_INSTANCE_OPTIMUM_RATE=0.960
ZERO_NOISE_TABLE_RECOVERY=0.850
TEN_PERCENT_NOISE_TABLE_RECOVERY=0.750
TEN_PERCENT_HELDOUT_ASSIGNMENT_ACCURACY=0.684
OUT_OF_CANDIDATE_SYNTHETIC_RECOVERY=PASS
ORDER_INVARIANCE=PASS
CHECKPOINT_IDENTITY=PASS
RESOURCE_FEASIBILITY=PASS
SELECTED_SEARCH_ARCHITECTURE=HYBRID
SEARCH_ARCHITECTURE_GATE=PASS
V3_DESIGN_STATUS=COMPLETE
V3_PRODUCTION_PREPARATION_AUTHORIZED=YES
REAL_DATA_SEARCH_AUTHORIZED=NO
```
