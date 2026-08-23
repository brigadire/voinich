# Reproducibility

## Requirements

- Go 1.25.5 or newer (the version declared in `go.mod`);
- a POSIX shell for `scripts/maintenance/run-full-analysis.sh`;
- sufficient disk for frozen artifacts and a multi-core machine for the full
  pipeline; `go test ./...` alone includes tests that take several minutes;
- for the canonical Voynich workflow, a lawfully obtained IVTFF source and
  IVTT binary. Neither is currently cleared for public redistribution.

Build and test a clean checkout with:

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
```

## Corpus preparation

For a generic text corpus, use the deterministic preparer:

```bash
go run ./cmd/codex_prepare -input input.txt -output prepared.txt
```

It writes `prepared.txt` and a `prepared.txt.prepare.json` sidecar with the
input SHA-256 and preparation metadata. Do not substitute a corpus into an
existing frozen experiment: create a new manifest instead.

The canonical workflow requires `data/ZL3b-n.txt`, locally supplied from the
IVTFF source documented in [DATA.md](DATA.md). Generate its analysis input:

```bash
./ivtt/ivtt -x7 data/ZL3b-n.txt data_work/ZL3b-x7.txt
```

## Canonical analysis

The legacy canonical pipeline is:

```bash
scripts/maintenance/run-full-analysis.sh
```

It writes ignored mutable files below `workdir/`; the frozen, versioned
evidence is under `experiments/`. A new isolated experiment should instead
use:

```bash
go run ./pipelines/pipeline-orchestrate manifest \
  -experiment-dir experiments/name-v1
go run ./pipelines/pipeline-orchestrate run \
  -experiment-dir experiments/name-v1
go run ./pipelines/pipeline-orchestrate freeze \
  -experiment-dir experiments/name-v1
go run ./pipelines/pipeline-orchestrate verify \
  -experiment-dir experiments/name-v1
```

The manifest records source hashes, Go/OS/architecture, command arguments and
deterministic seeds. Scientific flags are frozen in each tool's defaults;
do not change them when reproducing an existing result.

## Outputs and runtime

Expected frozen outputs are `manifest.json`, `run-state.json`,
`artifacts.json`, `checksums.sha256`, `REPORT.md`, and the registered contents
of `outputs/`. `pipeline-orchestrate verify` checks their hashes.

Task66-class mechanism-space and full Stage 1–28 runs are expensive
multi-stage experiments. Their runtime depends strongly on CPU count, memory,
selected controls, and worker executor; use the recorded experiment report
rather than treating a laptop run as a fixed benchmark. The quick build/test
commands above are the appropriate smoke test. Seeded analyses record their
seed in manifests or command arguments; unseeded reruns must not be described
as reproductions of a frozen result.
