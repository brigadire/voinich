# VM structural catalog specification

Schema: vm-structural-catalog-v1. Primary input is literal Unicode transcription symbols from the canonical corpus; a transcription symbol is not necessarily a physical atomic glyph. No visual decomposition is attempted. Multi-character transcription conventions are consequently represented as sequences of literal symbols.

Physical lines are hard boundaries for transitions and co-occurrence. The ZL3b IVTFF source is strictly aligned token-for-token to the canonical corpus to attach folio, locus type and section metadata. IT2a is used only for independent stability checks.

Every tested relation separates **observed_status** (the corpus fact) from **inferred_status** (the marginal-preserving analytic null assessment). Every zero remains UNOBSERVED with a concrete corpus_rule; FDR never removes it. P-values use directional Poisson tails with frozen lower-level marginals: within-token adjacency slots for glyphs, predecessor/successor marginals and physical-line boundaries for transitions, fixed token frequencies plus the exact physical-line slot lengths for co-occurrence, and document-category marginals for metadata. BH-FDR is applied per output family. These analytic approximations are explicitly not universal-mechanism claims.

The frequent-token threshold is fixed before inspection at frequency >= 10. The full token-transition complement is stored as a deterministic gzip JSON adjacency-complement representation; the explicit unobserved TSV covers the fixed frequent subset. Edit families are connected components of literal Levenshtein-distance-one insertions, deletions and substitutions; transformations are descriptive relations, not morphemes.

All TSV files are UTF-8, tab-delimited, have one header row, use NA for inapplicable numeric values, and are deterministically ordered.

Generate with go run ./cmd/vm-structure generate. Query without recomputation with go run ./cmd/vm-structure query glyph q, query follows daiin, query absent-with daiin, query position daiin, or query section daiin. Use --catalog DIR before the query type for a non-default frozen catalog.
