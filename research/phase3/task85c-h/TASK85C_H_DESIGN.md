# Task85c-h V1.2.1 implementation design

This is a clean implementation run authorized by Task85c-j. The authority chain is V1.2.1 → I2 → E3 → V1.2.1 evidence schemas, with frozen status V2 and generation semantics V1 selected transitively by I2. The historical V1.2 H-SC03 failure remains archived and is not resumed.

Production scientific dispatch lives in `internal/g1v2science`; `internal/g1v2` supplies deterministic execution, publication, retry, restart and conflict quarantine. All validation inputs are OPEN disposable fixtures. No production control, threshold, JobID, DAG or escrow material was created.

The implementation root is SHA-256 over sorted lines `<file-sha256><two spaces><workspace-relative-path><LF>`. The validation and task roots use the same algorithm; the task root excludes the results manifest to avoid recursion.
