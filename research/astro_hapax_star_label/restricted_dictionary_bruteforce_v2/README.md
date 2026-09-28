# Restricted dictionary brute-force v2

This package is a preregistered successor to `restricted_dictionary_bruteforce_v1`.
It tests a bounded global writing system: historical forms are first normalized by
the frozen orthographic grid, then one substitution table maps source graphemes to
EVA symbols for the whole dictionary. The table is selected on one page and frozen
before testing on the other page. No spatial star assignment, per-term rule, or
per-label exception is allowed.

v2 is prepared, tested, and intentionally not claimed as a completed production
experiment. `RUN_STATUS.json` records `PREPARED_NOT_RUN`; production requires the
explicit command in `REPRODUCIBILITY.md` and cannot modify v1.
