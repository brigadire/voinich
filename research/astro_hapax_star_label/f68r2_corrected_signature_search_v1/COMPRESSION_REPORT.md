# Compression report

The exact semantic key is `(LABEL occurrence, EVA token type, canonical identity, required_mappings, forbidden_source_units)`. Transformation semantics are frozen by the source graph and the full scorer contract. No rows differing in any key component are merged.

The graph contains 25,517 paths and 25,469 unique signatures, giving a compression ratio of 1.0019. Multiplicity is therefore low; signature compression removes duplicates but does not materially reduce the semantic search space. Exact-key member provenance is preserved in `PATH_SIGNATURE_MEMBERS.tsv`.\n
Small fixtures and deterministic subset builds agree with the uncompressed semantics.
