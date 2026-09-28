# Frozen E3 profile protocol

The matched search used the same 64 frozen E3 operation profiles, `S000` through `S063`, for PANEL_A, PANEL_B, PANEL_C and every pseudo-control. Profile order, operation definitions, corrected global propagation, `GLOBAL_CAPACITY_1`, support by distinct EVA token types, equal search budgets, stop conditions and UNKNOWN handling were identical across panels.

For each panel/profile, the graph was solved, incumbent/bound/status and assignments were stored, and an independent full scorer replay was performed. A profile was not treated as certified from an incomplete UNKNOWN response; decision queries were used until an exact maximum or the next-cover infeasibility was established. `REAL_PANEL_MAXIMA.tsv`, `PROFILE_RESULTS.tsv` and `SCORER_REPLAY_AUDIT.tsv` are the corresponding public tables.
