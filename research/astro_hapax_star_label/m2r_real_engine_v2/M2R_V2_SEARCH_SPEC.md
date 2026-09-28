# Bounded deterministic optimization

SEARCH_TYPE=BOUNDED_DETERMINISTIC_OPTIMIZATION
GLOBAL_OPTIMUM_CLAIMED=NO

The contract fixes all bounds. Candidate ranking uses recurring same-role substrings
and literal saving potential, then frequency disagreement, then rule tuple. Whole
units exist in the inventory, but duplicate collapse means a unique spelling cannot
supply support 3 for its own whole-token mapping. Candidate pruning is therefore an
explicit source of recall loss, rather than a hidden claim of complete model search.

The beam starts empty and explores add/remove/replace operations in canonical order.
Each candidate invokes the exact assignment subproblem and recomputes support/MDL.
Duplicate canonical rule sets are eliminated by SHA-256 state identity. Search reports
states, iterations, configuration, selected model, next best admissible models and
unit inventories. All-pair top-2 paths for the selected model are retained; they are
the local explanations within that model, not unrestricted latent segmentations.

Worker execution restricts filesystem reads to the explicitly passed surface file,
engine code, worker code, checkpoint and Python runtime. Truth and real-data directories
are denied before open. Linux address space is limited to 512 MiB, with wall/RSS
checks inside search and a parent timeout. Pure alignment inner loops are bounded
by maximum token length; their cache has 1024 entries, span cache 16384. Resource
failures are INCOMPLETE. State/iteration budgets are the intended bounded algorithm.

Checkpoint stores completed iteration boundaries and a terminal flag (including
empty-candidate exhaustion), canonical beam rules, visited states, counts, input/
config/source fingerprint and cumulative wall time. A checksummed envelope is fsynced
and atomically replaced. Reconstructed states must yield identical scientific output;
timing and checkpoint file locations are not scientific fields.

One neutral worker handles all families. Seeds and shards exist only in orchestration.
Registry validation rejects reused seeds; canonical aggregation removes shard metadata
and rejects duplicate datasets. No family-specific shortcuts or positive-model-only
null evaluations exist.
