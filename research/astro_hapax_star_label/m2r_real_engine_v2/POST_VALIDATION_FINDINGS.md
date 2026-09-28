# Post-validation findings — reporting only, no implementation changes

The one-shot hidden evaluation failed. The candidate source, tests, contract and
specifications remain exactly as recorded in CANDIDATE_SEAL.json. No fixes or second
hidden validation were performed. A follow-up implementation must be M2R-v3.

## Recovery and search depth

Across six completed hidden positive base runs, exact rule precision is 0.297222,
recall 0.100000, train assignment accuracy 0.022059 and heldout full-token coverage
0.000000. Positive heldout normalized compression (mean 0.115313) is insufficient:
partial local savings do not imply recovery of the generating system.

A structural limitation is visible in the sealed configuration: starting from an
empty model, each add proposal increases cardinality by only one, and eight beam
iterations permit at most eight rules. Each positive generator's truth contains
15 rules, so exact recall is bounded above by 8/15, already below the required 0.75.
The 96-state cap and 48-rule candidate pool further restrict search. This should
have been detected before hidden validation. No configuration adjustment is allowed
now; the failed gate is retained rather than tuned away.

## Resource accounting and failed reproducibility

All eight null base runs and the out-of-family base run returned INCOMPLETE with
reason `memory`. Some later permutation/checkpoint runs also returned INCOMPLETE,
causing overall order-invariance and checkpoint gates to fail. Early completed
continuous/resume comparisons were identical, but that is not a full-suite PASS.

A separate OS-only diagnostic (no engine, truth or dataset imports) confirmed that
in this Linux environment a freshly exec'd child's `ru_maxrss` can reflect the
parent's previous resident memory. Before a 100 MiB parent allocation the new child
reported 11776 KiB; afterwards an otherwise identical new child reported 114176 KiB,
matching the parent peak. The engine compares this lifetime maximum to 512 MiB.
The orchestration process also materializes large all-pair JSON outputs, increasing
its memory footprint. This provides a concrete explanation for later workers being
rejected immediately despite starting fresh. The implementation remains unchanged.

The observed hash differences therefore do not alone prove an input-order-dependent
Hungarian/beam tie-break. They DO prove failure of the required full execution and
reproduction protocol. CR-07 remains operationally open despite the small regression
fixtures passing. Regression PASS is not a complete audit closure or readiness claim.

## Unavailable error-control metrics

NULL_COMPLETED=0/8
NULL_P95=NA
NULL_P99=NA
NULL_MAX=NA
NULL_ACCEPTED_PROFILE_FALSE_POSITIVE_RATE=NA

Raw GATE_RESULTS.json contains `null_accepted_rate: 0.0`, computed mechanically from
incomplete outputs' `accepted=false` eligibility flags. It is NOT a measured null
false-positive rate; no completed null sample exists. Preserve that raw output for
traceability and use the NA interpretation above. Null separation is failed/unavailable,
not demonstrated. No INCOMPLETE result is evidence of NO_MODEL.

All ten hard-negative base runs completed; predictive acceptances were 0/10. This
cannot establish usefulness when positive recovery fails. Negative latent-rule metrics
are not positive-recovery estimates; several minimum hard-negative generators share
a random pair-specific construction, as explicitly documented in the generator spec.

## Final disposition

The integrity ledger is complete and v1 is unchanged. This is a rejected, integrity-
recorded development candidate, not a frozen engine or production audit package.
M2R_V2_FREEZE_MANIFEST.json is intentionally absent. Real-data contents were not
accessed and real-data search remains unauthorized.
