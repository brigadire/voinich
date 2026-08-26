# Task85a-v1.1 confirmatory-integrity audit design

This audit treats Task85, Task85a, the blocked Task86 record, and Task86R at
commit `0e60737` as immutable evidence. It does not refit a candidate, use
HELDOUT for selection, change a threshold, or replace a historical marker.

The static audit traces every scientifically relevant contract key into the
Task86R source. Controlled tests are target-blind except for the explicitly
allowed M3/M4 diagnostic on frozen DEVELOPMENT. The tests live in
`research/phase3/task86r-analyze/audit_task85a_v1_1_test.go` and cover:

1. fixed Task86R RNG vectors and a no-warm-up counterfactual;
2. a normalized exhaustive all-pairs implementation of frozen M3;
3. bounded synthetic corpus enumeration against actual blue-fringe M3;
4. exhaustive M3 induction on DEVELOPMENT only;
5. full-alphabet glyph/alias round trips and F2 label-invariance/direct-mode
   regression;
6. a counting proof for length-one PM6 negative exhaustion.

Existing VALIDATION tables are used without new fitting. They retain only the
winner per class, so PM1/complexity-aware counterfactual selection cannot be
reconstructed; this evidential limitation is reported rather than filled by a
post-hoc rerun.

Classification follows the five-category enum in the task. A resolution is
called equivalent only with a mathematical invariant or controlled regression.
Primary-verdict robustness is evaluated separately from integrity of subsidiary
claims.

