# E2 exact expansion protocol

E2 uses only the frozen 57-label scope and 299-row remediation lexicon. It extends E1 exact global injective `DROP_UNMAPPED` substitutions to table sizes 5 and 6. UNMATCHED labels remain allowed. No local errors, historical spelling variants, digraphs, abbreviations, compositions, E3, or E4 are used.

Reachability is computed before any E2 solver call with the exhaustive injective exact compatibility predicate. E1 `d→y, s→h` incumbents are extended to complete warm-start tables of size 5 and 6. The three scoring profiles are recorded separately; because E2 contains only exact and unmatched terms, their objective ordering is equivalent and profile rows are deterministic replays.

The budget is 60 seconds per capacity/table-size configuration, one CP-SAT worker. FEASIBLE and UNKNOWN results retain their bounds/status and are never treated as certified optimum. Top-K and co-optimal enumeration are not claimed when the solver returns without an incumbent.
