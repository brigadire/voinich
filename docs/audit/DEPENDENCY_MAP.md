# Dependency map

Built from a direct grep/import-graph audit (not guessed) before any files
moved, then re-verified after the move via `go build ./...` / `go vet ./...`
/ `go test ./...` (all pass — see TASK70_REPORT.md).

## 1. executable → package

Every `cmd/` and `research/phase1/` main package imports only `internal/*`
packages by full module path (`zcore.dev/voinich/internal/...`), never a
relative path — so none of these edges changed when the directories moved.
Full per-tool import lists are in `docs/audit/REPOSITORY_INVENTORY.tsv`
(`scientific_dependency` / `used_by` columns). Highlights:

- `cmd/voinich` (dict-gen, Stage1) → `internal/workdir`
- `cmd/structural-analyze` → `internal/workdir`, `internal/profilestability`
- `cmd/conditional-regime-analyze` → `internal/conditionalregime`,
  `internal/profiling`, `internal/workdir` (also the mTLS distributed
  executor used in production — see §5)
- `cmd/token-relation-validate`, `cmd/replicated-local-structure-audit`,
  `cmd/higher-order-sequence-validate`, `cmd/positional-continuation-validate`,
  `cmd/transition-network-validate` → their own `internal/*` package +
  `internal/conditionalregime` + `internal/genericsegmentation` (Task43/44
  shared generic block-partition layer)
- `research/phase1/mechanism-space-analyze`, `research/phase1/recoverability-analyze`
  → `internal/mechanismspace` (+ `internal/tokenrepetition` for the latter)
- 8 of 10 `research/phase1/*` (Task58/59/60/62/63/64/65 + rozanova-temerev)
  → `internal/evaglyph`, the shared EVA composite-glyph parser introduced
  at Task58/59

No `internal/*` package has zero importers — all 44 are used by at least
one executable.

## 2. pipeline → executable

`pipelines/pipeline-orchestrate` invokes stages by **building each one as a
subprocess**, not calling into it in-process:

- `pipelines/pipeline-orchestrate/exec.go`'s `buildBinary()`:
  `exec.Command("go", "build", "-buildvcs=false", "-o", dest, "./"+sourceDir)`
  with `cmd.Dir = repoPath` (the repo root, passed in by the caller — not
  derived from where pipeline-orchestrate's own source lives, so moving
  pipeline-orchestrate itself required no change here).
- The **only** structural reference to each stage's location is the
  `Stage.SourceDir` string in `pipelines/pipeline-orchestrate/stages.go`
  (28 entries, one per stage) — Task70 updated every one of them from a
  bare directory name (e.g. `"structural-analyze"`) to a `cmd/`-relative
  path (e.g. `"cmd/structural-analyze"`).
- `pipelines/pipeline-orchestrate/run.go`'s `runLogged()` then execs the
  already-built binary directly (`exec.Command(name, args...)`) — no
  further path dependency.

Stage order (dependency chain, matches `stages.go`'s comment "topologically
sorted"): dict-gen → dict-analyze → structural-analyze → sequence-analyze →
begin-end-analyze → structural-normalize → normalization-compare →
structural-validate → structural-profile-stability → structural-reliability
→ soft-structural-space → structural-graphemic → structural-pair-decompose →
distance-context-analyze → local-regime-analyze → property-trajectory-analyze
→ structural-projection-analyze → global-regime-analyze → metadata-validate
→ cluster-metadata-global → conditional-regime-analyze →
residual-diagnostic-analyze → token-relation-validate →
replicated-local-structure-audit → higher-order-sequence-validate →
positional-continuation-validate → transition-network-validate →
vocabulary-growth-analyze (28 stages total).

## 3. experiment → executable

Every `research/phase1/<name>` directory is a standalone `package main`
built and run directly by hand (`go run ./research/phase1/<name> ...`), not
invoked by `pipelines/pipeline-orchestrate` or any other tool. Verified by
grep before the move: `grep -rl "zcore.dev/voinich/independent"` and
`grep -rl "zcore.dev/voinich/pipeline-orchestrate"` across all `.go` files
both returned empty — zero import-graph coupling between Stage1-28 code and
the Task58-67 independent experiments, and zero coupling between
pipeline-orchestrate and either the independent experiments or the other
task-specific research CLIs (`corpus-transform`, `inverse-transposition-search`,
`inverse-homophony`, `voynich-validation`). This isolation is what let
Task70 relocate all of `research/phase1/` without touching any Go import
path anywhere else.

## 4. experiment → corpus

- Task55 (`corpus-transform`) and Task58-67 (`research/phase1/*`) each read
  a specific corpus file (usually a `data_test/*.txt` control corpus, or a
  frozen upstream Voynich derivative) named explicitly on the command line
  or in the experiment's own config — see `docs/audit/PROVENANCE_AUDIT.tsv`
  for per-experiment corpus identifiers and checksum status.
- Task46/54/57's standalone tools (`corpus-transform`,
  `inverse-transposition-search`, `inverse-homophony`, `voynich-validation`)
  generate or consume corpora explicitly via `-corpus`/`-input`/`-out-dir`
  flags — never the Voynich default silently, per
  `internal/workdir/contract_test.go`'s exclusion list (these four are
  "experiment-input generators," not pipeline stages, and are explicitly
  exempted from the shared workdir-output contract).
- Stage1-28 defaults to `data_work/ZL3b-x7.txt` (the IVTT-derived canonical
  Voynich corpus) unless overridden by `-corpus`/`-input`.

## 5. experiment → previous experiment artifact

- `research/phase1/line-regime-analyze` (Task64) explicitly reproduces and
  extends Task58-63's artifacts (its own `REPORT.md` embeds a Task63
  reproduction check).
- `research/phase1/local-regime-topology-analyze` (Task65) explicitly
  reproduces and extends Task64's artifacts.
- `research/phase1/recoverability-analyze` (Task67) chains off Task66's
  frozen `FINAL_ARCHITECTURE.tsv`/`MECHANISM_PARETO.tsv` via content-hash
  references in its own freeze markers (`CANDIDATES_FROZEN`/
  `DECODER_FROZEN`), not a git commit reference.
- Task68/69 (literature audits) reference and cross-index claims made by
  every Task58-67 experiment report (`docs/literature/TASK58_67_LITERATURE_CROSSWALK.tsv`).

Full task→experiment→report→commit detail is in
`docs/audit/TASK_EXPERIMENT_MAP.tsv`.

## 6. Production-relevant edge

`cmd/conditional-regime-analyze` is the one stage binary also deployed to
real remote mTLS worker hosts by `ansible/roles/voynich_worker/`
(Task33-35). `ansible/roles/voynich_worker/tasks/build.yml` builds it from
source (`go build ... ./cmd/conditional-regime-analyze`, updated by Task70
from `./conditional-regime-analyze`); every other ansible file in that role
references the *installed binary name* (`bin/conditional-regime-analyze`),
which Task70 did not need to touch.
