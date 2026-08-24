# Task83a design: Fingerprint V2 provenance reconciliation

## Scope and isolation

This is an integrity experiment, not a model-comparison experiment. The only Task83 evidence admitted is its invalidation/checksum diagnostic. Task83 rankings, distances, compatibility statements, evidence levels, trajectory results, and class conclusions are quarantined and were not read or used. Task83 remains `TASK83_EXPERIMENT_INVALID`.

The decision rule was fixed before regeneration: select neither candidate hash by scientific convenience. Reconstruct the pre-Task83 documented pipeline and Task79c-era code, regenerate twice from frozen raw bytes, inspect independently embedded input hashes, regenerate feasible downstream computations, and audit every checksum-bearing input in the old freeze. No preprocessing-option search for either target hash is permitted.

## Runs

1. Verify raw/current/checksum evidence and reconstruct Git history.
2. In two independent clean temporary directories run `ivtff-x7-extract` from raw IT2a, then `codex_prepare prepare -encoding utf-8 -case preserve -line-policy preserve`.
3. Repeat with the Task79c source snapshot (`6f185579…`); compare parser/preparer code with recorded HEAD `d568e54…` and current code.
4. Repeat the equivalent ZL3b preparation twice.
5. Run the historical Task79c analyzer from regenerated IT2a and compare canonicalized scientific payloads. Re-run PF4/hierarchy and full-portfolio distance/Pareto into temporary outputs.
6. Audit the complete old manifest input set and normative artifacts. Bind JSON artifacts to their embedded corpus hashes.
7. Issue a refreeze only if every manifest-only gate passes. Verify the refreeze recursively and run two negative regression tests.

Temporary regeneration products are deliberately outside the authoritative tree until classified. No frozen scientific artifact is silently replaced.

## Classification rule

`MANIFEST_ONLY_ERROR` requires raw provenance, unambiguous preprocessing, deterministic `10286e…` regeneration, independent Task79c binding, reproduced scientific results, no evidence that `3fb953…` was a scientific input, and clean remaining provenance. Otherwise the appropriate stronger error class is required.
