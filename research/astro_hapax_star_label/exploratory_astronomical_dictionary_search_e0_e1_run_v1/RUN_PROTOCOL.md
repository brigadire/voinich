# E0–E1 exploratory run protocol

This run is authorized for frozen real STAR LABEL data only at stages E0 and E1. E2, E3, and E4 are out of scope and were not executed.

E0 constructs all 57 × 299 compatibility edges and applies only the predeclared injective exact compatibility predicate. The synthetic false-pruning gate must pass before E1. Real labels that are unreachable under the E1 table-size bound remain in the output as unreachable; this is not an E0 failure.

E1 uses one-character global injective substitutions, `DROP_UNMAPPED`, no abbreviation, table sizes 1–4, unmatched labels allowed, and capacity modes `PER_PAGE_CAPACITY_1` and `GLOBAL_CAPACITY_1`. The minimum independent support is two. Scoring profiles are the frozen CONSERVATIVE, BALANCED, and RECALL_ORIENTED profiles. In E1 the near and weak terms are zero by definition, so the ranking objective is equivalent across profiles; profile rows are deterministic replays of the balanced solver incumbent and are labeled as such.

The per-configuration wall budget is 60 seconds, one CP-SAT worker, with deterministic configuration identity and atomic per-configuration checkpointing. `TOP_K=50` and `COOPTIMAL_LIMIT=500` are target limits; this run records one incumbent per configuration and explicitly marks top-K/co-optimal enumeration as incomplete. A FEASIBLE result retains its bound and gap and is never reported as certified optimum. No post-result threshold or scoring change is permitted.
