# Task86C-v2 handoff

Implement the manifest adapter and evidence-only verifier described here, then execute Stages A–F of `TASK86C_V2_CONTROL_DESIGN.md`. Freeze threshold derivation outputs after open development and before any confirmatory input is accessible. Use new independently authored generators, parameters, and seeds; escrow ground truth; publish the analysis freeze before unblinding.

Required inputs are every `G1V2_*.tsv` registry, the PM6 and decision contracts, the Phase-I executor revision/hash closure, natural-corpus provenance, opaque synthetic manifest, escrow manifest, and firewall/access log. Required outputs include per-job transitive evidence manifests, reconstruction output, negative-test ledger, distributed execution ledger, cross-node duplicates, scalability telemetry, blinded analysis freeze, unblinding boundary, recovery table, and the four validation verdicts.

Task86C-v2 is ready to be implemented, not yet passed. It must not access a path containing a Voynich corpus or use Task86C values to set a threshold. Its terminal decision gates Task86V.
