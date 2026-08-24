# Voinich: Phase I structural analysis

This repository contains reproducible software, frozen experiments, and
research documentation for a conservative structural analysis of the Voynich
Manuscript transcription. Phase I studies observable properties of the
transcription; it does not attempt a translation or assume that transcription
glyphs are letters, spaces are word boundaries, or tokens are words.

## Status and scope

Phase I is complete as an auditable research snapshot. Its supported
conclusion is that the analyzed transcription is a strongly structured
symbolic system and that several simple mechanisms are inadequate. It does
**not** claim that the manuscript has been deciphered, is meaningless, is
unrecoverable, is a cipher, or is a language. The current evidence does not
conclusively select among complex encoded meaningful input, an artificial or
formal symbolic system, and structured message-free generation.

Start with the [Phase I summary](docs/phase1/PHASE1_SUMMARY.md), the
[research report](docs/phase1/PHASE1_RESEARCH_REPORT.md), the
[claim registry](docs/phase1/PHASE1_CLAIMS.tsv), and the
[result index](docs/phase1/RESULT_INDEX.tsv).

## Repository map

- `cmd/` and `internal/`: Go analysis tools and their tests.
- `pipelines/`: deterministic experiment orchestration.
- `research/phase1/`: Phase I analysis implementations.
- `experiments/`: frozen experiment artifacts and reports.
- `docs/`: method, literature, audit, and Phase I documentation.
- `corpora/`: corpus provenance notes; corpus bytes are deliberately not
  generally distributed.

## Quick start

Go 1.25.5 or later is required:

```bash
go build ./...
go vet ./...
go test ./...
go test -race ./...
```

For a documented generic-corpus workflow, prepare a corpus and create an
experiment manifest:

```bash
go run ./cmd/codex_prepare -input input.txt -output prepared.txt
go run ./pipelines/pipeline-orchestrate manifest \
  -experiment-dir experiments/example-v1 \
  -corpus prepared.txt \
  -generic-corpus
```

The Phase II lexical-paradigm Fingerprint v2 block accepts an explicit YAML
configuration and writes reproducible JSON artifacts:

```bash
go run ./cmd/fingerprint-v2-analyze -config fingerprint-v2.yaml
```

See
[`research/phase2/fingerprint/FINGERPRINT_V2_SCHEMA.md`](research/phase2/fingerprint/FINGERPRINT_V2_SCHEMA.md)
for the schema, null models, and output contract. This command does not
bundle or download a canonical corpus.

The page/hierarchy/stability extension and its freeze decision are documented
in
[`research/phase2/fingerprint/TASK79_REPORT.md`](research/phase2/fingerprint/TASK79_REPORT.md).

Canonical Voynich runs additionally require a locally obtained IVTFF source
and an external IVTT installation. Datasets and IVTT are not bundled because
their redistribution rights were not established. Acquisition, checksum,
preparation, command, and placement instructions are in [DATA.md](DATA.md)
and [REPRODUCIBILITY.md](REPRODUCIBILITY.md).

## Data, reproducibility, and citation

The canonical Voynich corpus and the Astafiev historical control have
unresolved redistribution terms and are external-only inputs. See
[DATA.md](DATA.md) and
[the data-license audit](docs/release/DATA_LICENSE_AUDIT.tsv) before obtaining
or sharing them. Frozen artifacts are retained as research evidence; the
[size audit](docs/release/SIZE_AUDIT.md) explains their impact.

Please cite this repository as described in [CITATION.cff](CITATION.cff).
Contribution and scientific-correction policy are in
[CONTRIBUTING.md](CONTRIBUTING.md).

Original project source code is licensed under the
[Apache License 2.0](LICENSE). That grant does not cover third-party datasets,
IVTT, or other external material. No license is granted here for research
documentation or generated/frozen non-code artifacts unless a file explicitly
says otherwise; their scope remains for the owner to determine.
