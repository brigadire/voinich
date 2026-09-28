# Blindness audit v2

PASS: the builder reads only v1 curation ledgers. It does not read EVA tokens, path graphs, generator/solver inputs, reachability, astronomy witnesses, coverage, or future search results. It does not run an EVA search.

Selection is source-ledger-only. The astronomy baseline is recorded only as the historical scalar `3/27`; astronomy forms are not loaded. No panel is selected by coverage.
