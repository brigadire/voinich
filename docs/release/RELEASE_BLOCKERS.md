# Release blockers

## BLOCKING

None.

## RESOLVED IN TASK72B

- IVTFF source and ordinary derivatives are external-only, ignored, and fully
  documented with source/output checksums.
- Astafiev source, normalized plaintext, and sidecar were removed; external
  acquisition and deterministic preparation are documented.
- Vendored IVTT source/binary were removed; the pipeline resolves an external
  executable using `IVTT_BIN` or `PATH`.
- Original project source code is licensed under Apache-2.0 with an explicit
  scope that excludes third-party and non-code material.
- Exact dependency licenses were verified.
- A clean-root public-history strategy excludes private reachable history and
  the removed `tasks/` tree. The original private history was not rewritten.
- Dedicated gitleaks scans passed for the candidate tree/history.
- At the owner's explicit direction, the two exact ZL3b-x7 payloads formerly
  stored as `normalized_085.txt` and `normalized_090.txt` were excluded while
  their paths, SHA-256, provenance, and local backup were retained. The clean
  public history never contains those blobs.

## MANUAL

- Human owner review of the final candidate and license scope for
  documentation/generated research artifacts.
- Repository visibility and public push remain owner-only actions.
