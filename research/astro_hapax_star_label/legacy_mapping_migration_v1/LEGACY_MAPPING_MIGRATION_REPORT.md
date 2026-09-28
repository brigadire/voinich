# Legacy astronomical mapping migration report

## Outcome

The legacy study reproduced exactly, and 21 of 92 canonical LABEL received a
valid migrated transcription mapping. All 21 are grouped f68r1 LABEL with a
direct frozen anonymous-label provenance path. No token was assigned from a
Stolfi ordinal alone, to an unmatched coordinate, to a sensitivity-only bound,
or to a human-added LABEL.

```text
CANONICAL_LABELS=92
MAPPED_LABELS=21
RESIDUAL_LABELS=71
GROUPED_LABELS=64
GROUPED_MAPPED=21
UNGROUPED_LABELS=28
UNGROUPED_MAPPED=0
HUMAN_ADDED_LABELS=10
HUMAN_ADDED_MAPPED=0
HAPAX_ENRICHMENT_RUN_AUTHORIZED=NO
LEXICON_MATCH_RUN_AUTHORIZED=NO
```

## Crosswalk interpretation

The useful chain is `legacy matched row -> exact f68r1 ZL3b @Ls locus -> frozen
anonymous spatial LABEL -> direct A source ID -> canonical LABEL -> 3G1`.
Twenty-one rows complete that chain. Five directly linked f68r1 visual loci have
no confirmed legacy occurrence. Three f68r1 canonical grouped LABEL lack the
direct A-label link and remain many-to-many. The 22 non-human grouped f68r2
LABEL and three non-human grouped f68r3 LABEL have page-level legacy candidates
but no independent legacy pixel/object crosswalk. Ten human-added LABEL are
explicitly left unmapped. Ungrouped LABEL have no identified legacy star-series
candidate.

## Coverage and gate

Overall coverage is 21/92 (22.826%); grouped coverage is
21/64 (32.812%); ungrouped
coverage is 0/28. Page coverage is f68r1 21/37, f68r2 0/33,
and f68r3 0/22. Mapping selection is therefore perfectly concentrated in the
grouped arm and in one page. The amendment requires both comparison arms in at
least two panels and sign-invariant missingness bounds. The contrast is not
identified, so the gate stops before any corpus-frequency or hapax join.

No enrichment files were created. The prior Stolfi hapax result remains a
different legacy analysis and is not relabeled as the requested 3G1 comparison.
The minimum residual queue has 71 LABEL and is stratified by absent
legacy coverage, human additions, geometry ambiguity, and missing exact legacy
occurrences.
