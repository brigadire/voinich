# Decision-path regeneration contract

`validate.py verify <bundle.json>` is the reference design verifier. Its production implementation may be in Go, but it must implement the same versioned registries and consume only evidence records. It must not import or invoke fitting, induction, corpus generation, or model packages.

Regeneration proceeds bottom-up: verify hashes and schemas; evaluate per-baseline PM gates; apply the predictive truth table; evaluate F2 metrics, families, replicates, and scales; derive candidate adequacy; build complexity/equivalence relations; then compare regenerated roots with stored verdicts. Any mismatch refuses verification rather than preferring the stored value.

Mandatory negative tests mutate one dimension at a time: remove a PM value; change a threshold; change an artifact byte/hash; remove an F2 record; encode unavailable as FAIL; replace induction failure with model inadequacy; omit a dependency; change a seed/config/code hash; and introduce duplicate conflicting result bytes. Every mutation must fail closed. Positive fixtures prove unavailable is not FAIL, induction failure is not class rejection, and a complete path regenerates without model execution.

The acceptance record stores verifier code hash, registry hashes, fixture hashes, and every test outcome. A future result freeze is invalid unless the regenerated verdict file is byte-identical to the originally aggregated decision file.
