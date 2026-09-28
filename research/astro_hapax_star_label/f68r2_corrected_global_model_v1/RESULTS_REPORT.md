# Corrected global model results

The corrected path model adds `required_mappings` and `forbidden_source_units` to every path. The latter prevents a selected global mapping from emitting extra characters from a source unit that was dropped by that path.

The historical five-assignment S043 result is a required negative fixture and fails global replay; corrected CP-SAT returns no selected assignment on that fixture. Small exhaustive tests agree with the corrected model.

The real baseline on 25,517 frozen paths is currently `SEARCH_INCONCLUSIVE_TIMEOUT`: k=1..5 each reached no incumbent within the 20-second per-k budget, and k=6..12 were not run. This does not establish a zero maximum. E3 profiles and null remain closed until a corrected baseline is completed.
