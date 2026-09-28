# Generator algorithm

The exact DFS state is `(source_position, target_position, canonical_mapping, used_target_units, globally_unmapped_units)`. Memoization uses the canonical sorted mapping and unmapped sets. It applies only equivalence-safe pruning: fixed-repeat mismatch, injectivity, target-prefix/remaining-length feasibility, and the frozen maximum five mappings. Results are cached by `(transformed_source_stream, EVA_token)` and joined to lexicon/LABEL provenance afterward. Paths are streamed and deduplicated by the full semantic signature.
