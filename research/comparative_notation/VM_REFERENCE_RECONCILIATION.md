# VM reference v2 reconciliation (B02 section 37)

The generic analyzer (`internal/notation`, applied to the VM canonical
source through `notation-corpus vm-adapter` → `notation-corpus
vm-reference`) reproduces every anchor value from the frozen VM Structural
Catalog (`research/structure_catalog/VM_STRUCTURAL_CATALOG.md`) exactly:

| anchor | frozen catalog value | generic analyzer (VM_REFERENCE_V2.tsv) | metric_id / regime |
|---|---|---|---|
| alphabet size | 36 | 36 | `G01_ALPHABET_SIZE` |
| initial restriction | 9/36 = 0.25 | 0.25 | `G02_INITIAL_RESTRICTION_DENSITY` |
| final restriction | 0/36 = 0 | 0 | `G03_FINAL_RESTRICTION_DENSITY` |
| bigram occupancy | 379/1296 = 0.292438271605 | 0.292438271605 | `G04_BIGRAM_OCCUPANCY` |
| trigram occupancy | 1569/46656 = 0.0336291152263 | 0.0336291152263 | `G05_TRIGRAM_OCCUPANCY` |
| frequent transition zero density (threshold 10) | 0.9604589793 | 0.960458979298 | `S02_TRANSITION_ZERO_DENSITY` / `FREQ_GE_10` |
| same-line zero density | 0.7693935582 | 0.769393558194 | `L07_SAME_LINE_NONCOOCCURRENCE_DENSITY` / `FREQ_GE_10` |

All seven anchors match within floating-point display precision (the catalog
prints 10 significant digits; the generic analyzer's extra digits are
consistent rounding of the identical exact fraction, e.g.
`379/1296 = 0.29243827160...`).

## One formal defect discovered, and its resolution

Reproducing the same-line anchor required a genuine correction to the
generic analyzer, not a reinterpretation of VM science:

`L06_SAME_LINE_COOCCURRENCE_DENSITY` / `L07_SAME_LINE_NONCOOCCURRENCE_DENSITY`
were originally hard-coded to a `TOP_100` vocabulary support
(`L06_SAME_LINE_COOCCURRENCE_DENSITY_TOP100`). The frozen catalog's
"same-line zero density" anchor is defined on the *frequency ≥ 10* support
(553 token types for the VM corpus — the same set S01/S02 already use),
which is a different, larger vocabulary than an arbitrary top-100 cutoff.
Comparing the two supports directly showed a mismatch that was not a
scientific disagreement but a support-selection bug: L06/L07 were the only
metrics in the registry not yet stratified over the frozen sequence-support
regimes (`FREQ_GE_5`, `FREQ_GE_10`, `TOP_100`, `TOP_250`, `MATCHED_VOCAB`)
that every other support-dependent metric family (S) already uses.

**Resolution**: `lineMetrics` now emits `L06`/`L07` once per frozen support
regime, exactly like `S01`-`S03`/`S06`-`S08`, using the same
`selectVocabulary` call. `TOP_100` remains available as one of the five
regimes; nothing about the top-100 estimate was removed, and the fix is
purely additive stratification. `VM_REFERENCE_V2.tsv`'s `FREQ_GE_10` row now
reproduces the catalog anchor exactly. This is documented here rather than
silently folded into "no discrepancy" because it changed the shape of the
metric's output (added a `regime` dimension); `METRIC_REGISTRY.md` and
`internal/notation`'s `MetricRegistry()` were updated accordingly, and no
existing test asserted the old single-regime shape (verified before
changing it).

## Conclusion

No unresolved discrepancy remains. `B02`'s backward-consistency requirement
(section 37) is satisfied without invoking the STOP clause.
