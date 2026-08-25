# Grammar complexity contract

**Design version:** task85-v1. **Authority:** task85 sections 28-29, 33-34. **Frozen before grammar fitting:** yes. **Target-blind:** yes; no Voynich data, F2 vector, or fit statistic is read to choose the coding scheme below.

## Normative contract

Complexity(G) is a single two-part-code (MDL-style) quantity, always reported alongside (never replaced by) the raw secondary measures in section 3:

```
Complexity(G) = StructureCost(G) + LexiconCost(G) + ExceptionCost(G)
```

All three terms are entropy-coded in bits. No model is "compact" by having a small StructureCost while hiding size in an unaccounted lexicon (task85 section 33) or in unaccounted hand-added exceptions (task85 section 34).

### 1. StructureCost(G)

Every discrete choice point in G's own representation (an automaton transition, a rule/slot alternative, a stratum split) costs `log2(branching factor)` bits, summed over all realized choice points (states x outgoing-symbol alternatives for M3/M4; rule/slot count x per-slot alternative count for M5; n-gram context count x alphabet size for M1/M2/M6; active-STRUCTURAL_STATE-variable count x realized stratum count for M7).

Every free real-valued parameter (a transition probability, a mixture weight, a back-off constant) costs a single frozen constant:

```
bits_per_real_parameter = ceil(log2(N_dev)) / 2
```

(a BIC-style penalty), where `N_dev` is the DEVELOPMENT partition's TOKEN count from `GRAMMAR_CORPUS_SPLIT_MANIFEST.json`'s `corpus_totals`. This constant is shared by every model class and every transcription's DEVELOPMENT count is used for that transcription's own fits; it is never re-derived per model to favor one class.

### 2. LexiconCost(G)

Any TOKEN, COMPONENT, or rule-support table explicitly stored by G (M0's frequency table; M5's per-slot COMPONENT inventories; M6/M7's realized n-gram/stratum tables) is charged a Shannon code over its own DEVELOPMENT frequency distribution:

```
LexiconCost(G) = sum over stored entries e of  -log2( freq(e) / N_dev )
```

An entry that exists only because the model's productive rules already generate it at zero marginal storage cost (e.g. a GLYPH implied by M1's transition table) is not charged twice.

### 3. ExceptionCost(G)

Any hand-added special case (a rule or lexicon entry added outside the frozen training/induction algorithm) is charged its LexiconCost entry plus a fixed exception-flag overhead:

```
ExceptionCost(entry) = -log2(freq(entry) / N_dev) + 1 bit
```

The 1-bit overhead is an arbitrary but fixed, pre-registered constant: its only role is to make many small ad hoc exceptions never strictly cheaper than one generalizing rule of equal frequency-cost, not to make exceptions unusable. Regular rules and lexical/structural exceptions share one representation (a table of entries with a `kind in {RULE, LEXICAL_EXCEPTION, STRUCTURAL_EXCEPTION}` column); nothing is excluded from Complexity(G) for being "just an exception."

### 4. Unseen-symbol handling

A GLYPH/TOKEN/COMPONENT observed in VALIDATION or HELDOUT but absent from DEVELOPMENT's lexicon is scored under the model's own smoothing/back-off distribution (PM1/PM4 in `GRAMMAR_METRIC_REGISTRY.tsv`); it is never charged an ad hoc "infinite" or zero complexity cost, and it is never added to the lexicon after the fact.

### 5. Failure disposition

A model recorded as `COMPLEXITY_UNBOUNDED` or `MEMORIZATION_DOMINATED` in `GRAMMAR_FAILURE_REGISTRY.tsv` has Complexity(G) recorded as `UNBOUNDED`, not a numeric value, and is excluded from the `G_min` argmin (task85 section 30) rather than silently defaulting to zero or infinity in downstream arithmetic.

## Secondary complexity measures

Reported alongside Complexity(G) for interpretability, never used alone to break a G_min tie: raw free-parameter count; raw state/rule/stratum count.

## Cross-model fairness

Two models are complexity-comparable only when both figures are computed under this identical coding scheme on the identical DEVELOPMENT partition and transcription. Complexity(G) is never compared across transcriptions (ZL3b vs IT2a) as if it were a single shared number; cross-transcription robustness is a separate, qualitative check (`GRAMMAR_VALIDATION_CONTRACT.md`).
