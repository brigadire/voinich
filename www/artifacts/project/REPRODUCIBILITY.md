# Reproducibility

## Requirements

- Go 1.25.5 or newer;
- a POSIX shell;
- sufficient disk for frozen artifacts;
- for canonical Voynich preparation, externally obtained IVTFF and IVTT;
- for the Astafiev control, an externally obtained source file.

External data and IVTT are not bundled. Follow [DATA.md](DATA.md) from a fresh
clone; no knowledge of the private repository is required.

## Clean checkout smoke test

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
git diff --check
```

These checks require no external corpora. A generic public smoke workflow is:

```bash
go run ./cmd/codex_prepare -input input.txt -output prepared.txt
go run ./pipelines/pipeline-orchestrate manifest \
  -experiment-dir experiments/example-v1 \
  -corpus prepared.txt \
  -generic-corpus
```

Use a new experiment directory; never substitute a corpus into a frozen run.

## External-input bootstrap

1. Download the exact `ZL3b-n.txt` and IVTT source listed in
   [DATA.md](DATA.md); build IVTT outside the checkout.
2. Place the IVTFF source at `data/ZL3b-n.txt`.
3. Run
   `IVTT_BIN=/absolute/path/to/ivtt scripts/prepare-external-data.sh ivtff`.
   The script validates both the source and `data_work/ZL3b-x7.txt`.
4. If reproducing the Astafiev control, obtain the documented source, place it
   at `data_test/astafiev-1000-culinar-receipts.txt`, and run
   `scripts/prepare-external-data.sh astafiev`.
5. Confirm `git status --ignored --short` marks all external inputs,
   derivatives, sidecars, and local IVTT installations as ignored.

The legacy canonical pipeline then runs as:

```bash
IVTT_BIN=/absolute/path/to/ivtt scripts/maintenance/run-full-analysis.sh
```

The script retains the historical IVTT `-x7` behavior. Analysis commands
also accept explicit corpus/IVTFF paths where their command-line interfaces
provide those flags; repository-local IVTT is never assumed.

## Frozen experiments

The versioned evidence is under `experiments/`. For a new isolated run:

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

Manifests record source hashes, Go/OS/architecture, arguments, and deterministic
seeds. Scientific flags and frozen artifacts must not be changed when
reproducing an existing result.

Task66-class mechanism-space and full Stage 1–28 runs are expensive. Runtime
depends on CPU, memory, controls, and executor. Use the recorded reports rather
than a fixed laptop benchmark.
