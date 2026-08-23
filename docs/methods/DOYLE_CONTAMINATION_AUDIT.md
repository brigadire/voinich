# Doyle contamination audit (Task41)

Audited target: `experiments/doyle-sign-of-four-v1` and the shared repository
`workdir/` it used. The experiment is **scientifically invalid** and must not be
continued or used as a baseline.

The v1 manifest names `data_test/pg2097-2.txt`, but its recorded corpus SHA256
is `d2c67935...`, while the currently present file hashes to `0b260c8a...`.
There is no artifact registry and no frozen `outputs/` snapshot. Its 17 stage
logs contain only analyzer messages and no corpus/provenance header. Therefore:

| files | classification | evidence |
|---|---|---|
| `manifest.json` | UNKNOWN / identity mismatch | recorded corpus hash no longer matches the named file |
| `run-state.json` | UNKNOWN | status-only resume data; no output hashes or ownership |
| `logs/01-*.log` through `logs/17-*.log` | UNKNOWN | no corpus identity, cwd, input list, or output provenance |
| `structural_validation.yaml` | STALE_VOYNICH | embeds `data_work/ZL3b-x7.txt`, 39,026 tokens, and Voynich SHA256 `360d9958...` |
| `structural_reliability.yaml` | STALE_VOYNICH | same path, token count, and SHA256 |
| `global_distributional_regimes.yaml` | STALE_VOYNICH | embeds Voynich corpus path and 39,026-token count |
| `alignment_report.md` | STALE_VOYNICH | names IVTFF `data/ZL3b-n.txt`, Voynich corpus, and its SHA256 |
| `sequence_analysis.yaml` | STALE_VOYNICH | 39,026 occurrences / 5,385 lines; it was read from the shared output path |
| `transition_network_summary.yaml` (when present in the shared tree) | STALE_VOYNICH | metadata-dependent Voynich stage is NOT_APPLICABLE to Doyle |
| Doyle pair artifacts containing `answered_said`, `case_door`, `house_room`, `treasure_window` | VALID_DOYLE content, UNKNOWN ownership | lexical evidence is Doyle, but shared-path execution supplies no experiment provenance |
| every other unregistered shared `workdir/` file | UNKNOWN | content alone cannot prove producing invocation/experiment |

The files in the second part of the table are not inside an immutable v1
snapshot; they are the mutable shared dependencies/outputs that v1 read and
wrote. Their mixed identities prove the root cause but cannot be repaired by
selective deletion or copying. A new isolated directory is required.

