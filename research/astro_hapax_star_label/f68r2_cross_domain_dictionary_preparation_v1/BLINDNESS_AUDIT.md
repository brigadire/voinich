# Target-blindness audit

Preparation code `prepare.py` does not load the 27 LABEL occurrences, EVA transcription, path graphs, generator, solver, coverage results, or witness assignments. `INPUT_FREEZE.json` records an empty target-file list.

The code only reads the frozen profile registry to record its hash. Botanical rows are selected from a predeclared source list and are not filtered by spelling similarity, reachability, or any search result. Pseudo-lexicons are deterministic Caesar-shift controls generated from botanical forms alone.

The astronomical dictionary is represented only by the preregistered reference count of 299 forms and 94 identities in the matching-statistics file. No astronomical form is used to select or alter a botanical row.

The current package is deliberately not marked search-ready. The seed corpus has 77 rows, while the astronomical reference has 299 forms, and each botanical source row still requires line-level source verification. The correct action is curation before any cross-domain search, not padding the corpus or inspecting EVA matches.
