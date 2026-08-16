# pipeline-orchestrate

Task36's single orchestration CLI: runs every current Voynich pipeline
stage (27 commands, `stages.go`) in dependency order against production
parameters, using `conditional-regime-analyze`'s distributed executor
(Task31-34) where that stage supports it, and freezes the result as an
immutable, checksummed experiment directory under `experiments/<name>/`.

## Why every stage's defaults are the production parameters

Every pipeline command's own CLI flag defaults already are the frozen
production parameters established by the task that introduced it (window
sizes, K ranges, permutation counts, seeds - see each package's own doc
comments). This tool therefore never sets a scientific flag: it adds only
operational flags (`-quiet`, `-executor`/`-workers`, mTLS transport for
`-executor remote`). `stageArgs`'s own test
(`TestStageArgsNeverSetsAScientificFlag`) enforces this with an explicit
allow-list of every flag this tool is permitted to pass.

## Commands

```
pipeline-orchestrate manifest -experiment-dir experiments/voynich-v1 [-executor process|goroutine|remote] [-workers N] [-force]
pipeline-orchestrate run      -experiment-dir experiments/voynich-v1 [-only STAGE]
pipeline-orchestrate freeze   -experiment-dir experiments/voynich-v1 [-force]
pipeline-orchestrate verify   -experiment-dir experiments/voynich-v1
pipeline-orchestrate status   -experiment-dir experiments/voynich-v1
```

Run from the repository root (paths are resolved the same way every other
pipeline command resolves them: relative to the current working directory).

### manifest

Writes `experiments/<name>/manifest.json` **once**, before any stage runs:
git commit (and whether the working tree was dirty), SHA256 of the frozen
IVTFF source and the frozen corpus, the exact argument list every one of
the 27 stages will be invoked with, Go/OS/architecture, and an
`ExperimentID` (a SHA256 over all of the above - two runs with the same ID
were given byte-identical instructions). It also writes a companion
`worker` list: honestly "local, N slots on <host>" for the
goroutine/process executor, or the actual authenticated remote worker
identities for `-executor remote` - never a fabricated fleet.

Refuses to overwrite an existing manifest without `-force`, and refuses
entirely once the experiment is `freeze`d.

### run

Builds every stage's binary once (`go build`, no shell involved - args are
passed directly to `exec.Command`, so no manifest value can inject a shell
command) and executes stages in order, streaming each one's combined
stdout/stderr to `experiments/<name>/logs/NN-<stage>.log` and recording
status/timing/CPU/RSS in `experiments/<name>/run-state.json` after every
single stage (atomic write-then-rename). Re-running `run` after a crash -
of the orchestrator itself, not just one stage's own internal checkpoint -
skips every stage already marked `completed` and resumes at exactly the
one that was interrupted (Task36 "checkpoint/resume": a coordinator/worker
crash never requires restarting the whole experiment). It stops at the
first failing stage rather than skipping ahead, since every later stage's
default input path assumes the earlier one actually succeeded.

`-only STAGE` re-runs one named stage regardless of its recorded status
(for investigating a single stage without touching the others' state).

### freeze

Requires every stage to already be `completed`. Copies every file under
`workdir/` (excluding `workdir/bin/`, the build scratch area) into
`experiments/<name>/outputs/`, computes SHA256 for each into
`checksums.sha256`, writes `REPORT.md` (per-stage wall-time/status/CPU/RSS,
total wall-time, manifest/checksum pointers), and writes a read-only
`FROZEN` marker. Every subsequent `run`/`freeze`/`manifest` against that
directory refuses outright unless `-force` is passed - Task36's "no
subsequent pipeline change may silently overwrite this baseline."

### verify

Recomputes every file's SHA256 against `checksums.sha256` and reports any
mismatch or missing file - the drift-detection half of "no silent
overwrite": run this at any point in the future to confirm the frozen
baseline is still exactly what it was when frozen.

### status

Prints each stage's recorded status/duration from `run-state.json` without
running anything.

## Distributed execution (Task33-35)

`-executor remote` requires a coordinator certificate/key and client-CA
bundle (`internal/pki`/`conditional-regime-pki`, Task34) and a reachable
`-remote-listen` address; workers are deployed independently (by hand, or
via `ansible/roles/voynich_worker`, Task35) and dial in. `manifest`'s
`-remote-worker` flag (repeatable) records the authenticated worker
identities actually configured for the run, so the manifest's worker list
reflects reality rather than an assumption. `-executor process`/`goroutine`
use only local CPU cores and record `local, N slots on <hostname>` instead
- there is no requirement to have any remote machine at all; Task28-29's
production estimate for `conditional-regime-analyze` (the only stage this
choice affects) is on the order of an hour with `-executor process` and a
handful of local worker slots on modern hardware.
