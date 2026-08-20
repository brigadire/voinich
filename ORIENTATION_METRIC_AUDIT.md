# Orientation metric audit

This audit was made before examining transformed-corpus comparative results.
It classifies existing pipeline outputs under `TOKEN_REVERSE`, which preserves
the complete token multiset while reversing each logical line independently.

| Classification | Existing metrics / stages | Expected relation |
|---|---|---|
| Invariant | N, vocabulary V, frequency spectrum, hapax/dis-legomena, token lengths, character composition; dictionary totals | Exactly identical for `TOKEN_REVERSE`. |
| Predictably swapped within a line | `begin-end-analyze`; left/right context components in structural analysis, structural reliability, distance context, structural projection, and property trajectories | A purely line-bounded predecessor profile maps to the reversed corpus's successor profile, and conversely. Implementations using matched supports or filtering can still change rankings. |
| Orientation-sensitive | transition network; sequence n-grams; higher-order sequences; positional continuation; local regimes; begin/end relations; M0→M1/M1→M2 and replicated profiles | Sequence direction is changed; values need comparison rather than an assumed equality. |
| Uncertain / mixed | normalization comparisons, global regimes, residual diagnostics, and any metric derived from both directional profiles and frequency controls | Frequency inputs are invariant but directional components may change. |

## Cross-line audit

The following stages explicitly treat each input line as an independent
sequence: `sequence-analyze` and the validation sequence metrics. Their
within-line directional statistics have the simple reversal relation above.

Other stages flatten non-empty lines into a continuous stream for at least
their primary calculation: `distance-context-analyze`, `local-regime-analyze`,
`global-regime-analyze`, and higher-order/replicated analyses' continuous
sequence paths. Their boundary adjacency is `last(line N) → first(line N+1)`.
After per-line token reversal it becomes
`first(original line N) → last(original line N+1)`, not the reverse of the
original adjacency. Those results are valid for the stated per-line
right-to-left intervention but are not a mathematical global-stream
reversal oracle.

No pipeline semantics are changed by this audit. Comparisons should report
line-bounded and continuous results separately where both are already emitted.
