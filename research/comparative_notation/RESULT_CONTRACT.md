# Per-corpus future result contract

Every corpus/representation writes only inside its own experiment result
directory: `SOURCE_PROVENANCE.json`, `NORMALIZATION_REPORT.md`,
`REPRESENTATION.md`, `RUN_MANIFEST.json`, `STRUCTURAL_SUMMARY.tsv`,
`STRUCTURAL_CATALOG.md`, `STRUCTURAL_RULES.tsv`, `SYMBOL_GRAMMAR.tsv`,
`TOKEN_GRAMMAR.tsv`, `SEQUENCE_GRAMMAR.tsv`, `LINE_GRAMMAR.tsv`,
`DOCUMENT_GRAMMAR.tsv`, `RAREFACTION.tsv`, `ACCUMULATION_CURVES.tsv`,
`VM_COMPARISON.tsv`, and `VM_COMPARISON.md`.

Classes with multiple corpora may additionally write `CLASS_SUMMARY.tsv`,
`CLASS_SUMMARY.md`, and `WITHIN_CLASS_DISTANCES.tsv`. No `GLOBAL_RANKING`,
`BEST_MATCH`, or `WINNER` artifact is permitted.
