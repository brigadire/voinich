# Task83r handoff

Task83r must not open a target yet. First:

1. determinize all map-derived statistical iteration (including hierarchy shuffles and CS7 bins), add same-seed/different-process reproducibility tests, and review all seeded paths;
2. recompute affected ZL, IT, and control F2 outputs from raw data under a new scientific version;
3. recompute transcription stability, PF4/hierarchy, distances/Pareto, CORE statuses, and verdicts;
4. issue a new authoritative scientific freeze and pass `go run ./cmd/fingerprint-v2-verify -manifest <new-manifest>`;
5. only then verify Task81/82 and Task82b portfolios, create a fresh Task83r sentinel, and open its target.

The authoritative IT prepared checksum to use as provenance input is `10286ee7…`, derived from raw `7f27a8…`. This is a provenance fact, not a model-comparison result. Task83 remains invalid and none of its quarantined comparisons may be reused.
