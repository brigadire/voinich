# Scorer-consistent graph semantics

Each row in `PATH_GRAPH.tsv` is retained only when the executable scorer reproduces the frozen EVA token exactly under `DROP_UNMAPPED` and `NONE` abbreviation.

Repeated source characters are mapped consistently wherever they occur. A subsequence alignment that silently drops a repeated mapped character is rejected. No compatibility edge receives an unrecorded deletion or an implicit abbreviation.
