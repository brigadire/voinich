# D0/D1 astronomical dictionary comparison

## Frozen design

D0 and D1 were rerun with the audited M1 implementation, the same 20/6 split,
640 pipelines, source/target segmentation, substitution constraints, beam width
64, regularized score, thresholds, and search-level controls. Only the term
corpus differs. Every D1 null replicate receives the complete D1 concept/form
inventory, so dictionary size, variants, lengths, language mix, and class mix
are inherited by the randomized corpus. STAR and PLANET_MOON results are also
reported separately as required by the audit.

PLANET_MOON remains advisory: the five f67r2 labels are anonymous diagram
positions and may not denote five unique objects. Both object-disjoint and
form/rule-disjoint HELD_OUT semantics were equivalent in the prerequisite audit.

## Result

D0 matched 10/20 TRAIN and 0/6
HELD_OUT; D1 matched 12/20 TRAIN and
0/6 HELD_OUT. The real TRAIN gain is 0.100000;
the conservative null-mean gain is 0.085500; the change in real-null
advantage is 0.014500. “Newly explained” means membership in the
top D1 model's canonical maximum matching but not the top D0 model's canonical
matching; anonymous assignments are not object identifications.

## Final status

```text
ASTRO_DICTIONARY_EXPANSION=NULL_COMPATIBLE_GAIN

D0_TRAIN_COVERAGE=0.500000
D1_TRAIN_COVERAGE=0.600000

D0_HELDOUT_COVERAGE=0.000000
D1_HELDOUT_COVERAGE=0.000000

D0_NULL_ADVANTAGE=0.038500
D1_NULL_ADVANTAGE=0.053000

D0_EMPIRICAL_P=0.821782
D1_EMPIRICAL_P=0.326733

D1_MODELS_WITH_COVERAGE_GE_70=0
NEWLY_EXPLAINED_LABELS=2
```

The result tests lexical coverage only. It neither translates labels nor
identifies astronomical objects without independent object-level evidence.
