# Profiling protocol

The monolithic remediated runner is probed for labels 10, 20, 30, and 57 at k=3, 4, and 5, with deterministic prefixes of the frozen scope, `GLOBAL_CAPACITY_1`, and `BALANCED`. Probes have a fixed three-second build/return envelope and write a heartbeat through the subprocess outcome. Candidate and support-link counts are recorded before model construction.

The dominant support-link estimate is `|source| × |target| × |distinct label tokens| × |identities|`; it is recorded alongside rule-edge links and candidate edges. The resulting profile distinguishes construction timeout from a solver search result.
