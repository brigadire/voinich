# Repository structure (post-Task70)

Task70 is a repository/refactoring task, not new research: it moves files and
fixes path references so the reorganized tree builds, tests, and runs
identically to before. No scientific algorithm, metric, threshold,
classification, or result changed. See
[docs/audit/TASK70_REPORT.md](audit/TASK70_REPORT.md) for the full migration
record, [docs/audit/REPOSITORY_INVENTORY.tsv](audit/REPOSITORY_INVENTORY.tsv)
for a per-path classification, and
[docs/audit/TASK70_ISSUES.md](audit/TASK70_ISSUES.md) for everything found
along the way that this task deliberately did not fix.

## Layout

```
cmd/                     Reusable Go CLI tools: the Stage1-28 pipeline
                          binaries (dict-gen through vocabulary-growth-
                          analyze) plus generic infra tools
                          (codex_orientation, codex_prepare,
                          experiment-compare, conditional-regime-pki).
                          Each stage keeps its pre-Task70 directory name,
                          now nested one level under cmd/.

pipelines/
  pipeline-orchestrate/   The Task36 orchestration CLI (code + tests +
                          STAGE_AUDIT.md), moved here rather than into
                          cmd/ because it *is* "reproducible analysis
                          orchestration" (Task70 §6), not a stage tool it
                          builds and runs.

internal/                UNCHANGED. All 44 packages keep their import
                          path; this is already the reusable-library layer
                          Task70 §5 asks for, and none of it needed to move.

research/
  phase1/
    README.md            Phase I boundary (through Task69) and a mapping
                          of scientific directions to experiments (§12);
                          not the full Phase I synthesis (that is Task71).
    <experiment>/         Task46/54/54b/57's experiment-specific CLIs
                          (corpus-transform, inverse-transposition-search,
                          inverse-homophony, voynich-validation) and all
                          10 Task58-67 independent/* hypothesis-test CLIs,
                          unwrapped one level (no more independent/
                          prefix). Verified zero import-graph coupling
                          with Stage1-28 code in either direction before
                          the move (see DEPENDENCY_MAP.md §4).

experiments/              UNCHANGED location — canonical, immutable
                          experiment artifacts. Task70 §7 prefers keeping
                          canonical experiment paths when a move isn't
                          required, and none of the 39 experiment
                          directories needed one.

corpora/
  README.md              Documents data/, data_test/, data_work/, ivtt/
                          IN PLACE (§14) — not moved, since two shell
                          scripts read them via hardcoded relative paths
                          and moving them would be pure churn.

docs/
  methods/                Cross-cutting/pipeline-era audit docs that don't
                          belong to exactly one tool (performance,
                          scalability, distributed-execution, generic-
                          stage-applicability, output contract, etc).
  literature/             Task56/68/69 literature-review outputs
                          (LITERATURE_CATALOG.tsv, CLAIMS.tsv,
                          bibliography.bib, and friends).
  roadmap/                RESEARCH_GAPS.md.
  audit/                  ONLY the 6 new Task70 deliverables (§18) — kept
                          separate from docs/methods/ so this meta-audit
                          isn't buried among ~30 historical ones.
  REPOSITORY_STRUCTURE.md This file.

scripts/
  maintenance/            run-full-analysis.sh, run-normalization-
                          analysis.sh, make_analysis_archive.sh — run from
                          the repository root, unchanged in behavior.
```

**Left untouched** (documented in the inventory, not moved or deleted):
`pki/`, `bin/`, `data/`, `data_test/`, `data_work/`, `ivtt/`, `workdir/`,
`workdir.v0/`, `workdir-v2/`, `dataset/`, `task54-audit/`, `.agents/`,
`.codex/`, `.openmono/`, `profiles/`, every root `.tgz`/`.tar.gz` archive,
`ansible/` (apart from the one path fix below), `tasks/*.txt` (historical
task specifications — never edited, including the ones that predate the
"experiment" concept).

## Deliberate deviations from Task70 §4's example structure

Task70 §4 explicitly permits this: *"Don't follow this structure
mechanically if the existing architecture requires a more suitable
variant."*

- **`pipelines/pipeline-orchestrate/` holds real code, not just docs.**
  The orchestrator *is* the "reproducible analysis orchestration" §4 asks
  for; putting only documentation there and leaving the code in `cmd/`
  would have separated a tool from its own description for no reason.
- **No `docs/phase1/`.** `research/phase1/README.md` already satisfies §12;
  a second doc home next to it would just create ambiguity about which one
  is authoritative.
- **`docs/audit/` holds only the 6 new Task70 files.** Many pre-existing
  docs have "AUDIT" in their filename (`GENERIC_STAGE_APPLICABILITY_AUDIT.md`,
  `PERFORMANCE_AUDIT.md`, etc.) but describe one specific tool or a
  cross-cutting methodology from Task1-54 — none of them are Task70's own
  meta-audit deliverables, so none of them live in `docs/audit/`.
- **Per-file doc placement rule.** A doc that documents exactly one
  tool/experiment moved into that tool's own new directory (e.g.
  `EXPERIMENT_COMPARISON.md` → `cmd/experiment-compare/`;
  `INVERSE_HOMOPHONY_DESIGN.md` → `research/phase1/inverse-homophony/`;
  the three `VOCABULARY_GROWTH_*.md` files → `cmd/vocabulary-growth-analyze/`).
  A doc describing cross-cutting methodology or infrastructure
  (`GENERIC_STAGE_APPLICABILITY_AUDIT.md`, `PIPELINE_OUTPUT_CONTRACT.md`,
  `DISTRIBUTED_EXECUTION_*.md`, `PERFORMANCE_*.md`) went to `docs/methods/`
  instead.
- **`experiments/`, `data/`, `data_test/`, `data_work/`, `ivtt/` were not
  moved** — see the corpora/README.md and experiment-immutability notes
  above. §4's `corpora/` and §7's frozen `experiments/` requirements are
  satisfied by documentation, not relocation, because relocation had a real
  cost (breaking two shell scripts' hardcoded paths, and the point of a
  frozen artifact is that its path doesn't move) and no organizational
  benefit here.

## What changed under the hood

Every stage's build source moved from a flat repo-root directory into
`cmd/<name>/`. The pipeline invokes stages by building them as subprocesses
(`pipelines/pipeline-orchestrate/exec.go`: `go build -o <dest> ./<SourceDir>`)
keyed off exactly one string per stage
(`pipelines/pipeline-orchestrate/stages.go`'s `Stage.SourceDir` field), so
the move required updating those 28 strings, plus:

- `README.md`'s build/run snippets and repository-structure section,
- `scripts/maintenance/run-full-analysis.sh` and `run-normalization-
  analysis.sh`'s `go run`/`go build` invocations,
- `ansible/roles/voynich_worker/tasks/build.yml`'s two
  `./conditional-regime-analyze` build-source references (the one file move
  that touches real mTLS worker deployment infrastructure — a path-string
  edit, not a deploy action),
- three test files with relative-path depth assumptions that broke when
  `pipelines/pipeline-orchestrate/` gained one extra directory level
  (`internal/workdir/contract_test.go`,
  `pipelines/pipeline-orchestrate/stages_test.go`,
  `begin_end_manifest_test.go`, `generic_test.go`, `orientation_manifest_test.go`).

No scientific algorithm, statistical estimator, null model, randomization
procedure, corpus normalization, fingerprint, threshold, classification, or
experiment result changed. `go build ./...`, `go vet ./...`, and
`go test ./...` all pass after the move; `go test -race ./...` reproduces
exactly one pre-existing failure that is unrelated to this reorganization
(see TASK70_ISSUES.md).
