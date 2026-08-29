# Task85c-i design

Historical V1.1 authorities remain immutable. A separate 15-schema Draft 2020-12 family binds evidence to `G1_V2_EXECUTABLE_CONTRACT_V1_2` and is selected deterministically by I1; unknown and mixed versions fail closed. Each V1.2 schema is the corresponding V1.1 schema with only contract/schema identity strings versioned, proven by reversible normalization.

E2 separates its specification version from its target scientific-contract version and otherwise preserves E1's blind/escrow/RNG boundary and JobID function. I1 binds the parent V1.2 contract, E2, V1.2 evidence root, status/reachability V2, and unchanged generation authority without changing scientific semantics. The hash graph is acyclic because E2 and the schema registry do not depend on I1, and I1 does not depend on the graph.
