# Task70 report

Repository Reorganization and Provenance Audit — final report.

## Old architecture

~36 flat Go `main.go` directories at repo root (Stage1-28 pipeline CLIs,
Task34/45/46/48/51/57 standalone tools, and the Task58-67 `independent/`
hypothesis-test CLIs each one level deeper), mixed at the same level with
~50 loose audit `.md`/`.tsv` files spanning Task1 through Task69, ~23
duplicate/orphaned `.tgz` archives, three `workdir*` scratch trees, and the
already-good-but-invisible separation between reusable pipeline code
(`internal/*`) and independent research code (`independent/*`) — invisible
because both lived at the same directory depth as everything else. `cmd/`
existed but was empty. See `docs/audit/REPOSITORY_INVENTORY.tsv` for the
full pre-move classification of every path.

## New architecture

```
cmd/            27 Stage1-28 CLI wrappers + dict-gen (as cmd/voinich) +
                4 generic infra tools (codex_orientation, codex_prepare,
                experiment-compare, conditional-regime-pki)
pipelines/      pipeline-orchestrate (the Task36 orchestrator itself)
internal/       UNCHANGED — 44 packages, already the reusable-library layer
research/phase1/  14 task-specific/independent experiment CLIs (Task46/54/
                57/58-67), plus README.md mapping scientific directions
experiments/    UNCHANGED location — canonical, immutable artifacts
corpora/        README.md documenting data/, data_test/, data_work/, ivtt/
                in place (not moved)
docs/           methods/ (cross-cutting historical audits), literature/
                (Task56/68/69), roadmap/ (RESEARCH_GAPS.md), audit/ (this
                task's 6 deliverables only), REPOSITORY_STRUCTURE.md
scripts/maintenance/  run-full-analysis.sh, run-normalization-analysis.sh,
                make_analysis_archive.sh
```

Full structure, and the rationale for every deliberate deviation from
Task70 §4's example layout, is in `docs/REPOSITORY_STRUCTURE.md`.

## Migration summary

- **171 files renamed** via `git mv` (history-preserving): 32 into `cmd/`
  as whole directories, 14 into `research/phase1/`, ~45 documentation
  files into `docs/methods/`, `docs/literature/`, `docs/roadmap/`, or
  co-located with the single tool/experiment they document, 3 shell
  scripts into `scripts/maintenance/`, plus `task69.tgz` co-located with
  the `experiments/task69-literature/` directory it mirrors.
- **3 files modified in place**: `README.md` (build/run snippets +
  repository-structure section updated for the new layout),
  `ansible/roles/voynich_worker/tasks/build.yml` (the one production
  build-path fix — `./conditional-regime-analyze` →
  `./cmd/conditional-regime-analyze`, a path-string edit only, no deploy
  action taken), `internal/workdir/contract_test.go` (test logic adapted to
  scan `cmd/` instead of the repo root, preserving its exact original
  checked-file set).
- **1 mechanical config change**: all 28 `Stage.SourceDir` entries in
  `pipelines/pipeline-orchestrate/stages.go` prefixed with `cmd/`.
- **8 new documentation files created**: `docs/REPOSITORY_STRUCTURE.md`,
  `docs/audit/REPOSITORY_INVENTORY.tsv` (239 rows), `docs/audit/DEPENDENCY_MAP.md`,
  `docs/audit/TASK_EXPERIMENT_MAP.tsv` (76 rows, all 70 tasks + variants),
  `docs/audit/PROVENANCE_AUDIT.tsv` (13 Phase I experiments), this file,
  `corpora/README.md`, `research/phase1/README.md`.
- **4 test files fixed for a relative-path depth change**: moving
  `pipelines/pipeline-orchestrate/` one directory deeper than its old
  root-level location broke 4 tests' `filepath.Abs("..")`-style repo-root
  assumptions (`stages_test.go`, `begin_end_manifest_test.go`,
  `generic_test.go`, `orientation_manifest_test.go`) — each fixed to the
  correct `../..` offset; `stages_test.go`'s "every command directory is
  accounted for" scan was additionally updated to target `cmd/` (where
  `SourceDir` values now point) instead of the repo root directly.
- **No scientific algorithm, statistical estimator, null model,
  randomization procedure, corpus normalization, fingerprint, threshold,
  classification, or experiment result was changed.** Every edit was a
  path, string, or test-infrastructure fix required by the move itself.

## Provenance state

Full detail in `docs/audit/PROVENANCE_AUDIT.tsv` and
`docs/audit/TASK_EXPERIMENT_MAP.tsv`. Summary:

- **COMPLETE**: `experiments/voynich-v1` (Task36 baseline), Task59, 60, 64,
  65 — verified git commit, corpus checksum, seed, and report all present.
- **PARTIAL**: Task55 (rich manifest, but directory untracked in git),
  Task57 (Phase A only tracked-in-substance-but-untracked-in-git; Phase B
  missing), Task58, 61, 62, 63 (no git_commit field), Task67 (no
  manifest-level git_commit or corpus sha256, but real per-row seeds and a
  content-hash chain of custody back to Task66).
- **PARTIAL/LEGACY_MISSING**: Task66 (no git_commit, no centralized corpus
  sha256, and a manifest-claimed per-mechanism seed/config_hash tracking
  that this pass could not independently locate).
- Tasks 1-54 predate the "experiment" concept entirely — they built the
  Stage1-28 pipeline and its supporting infrastructure. Their entries in
  `docs/audit/TASK_EXPERIMENT_MAP.tsv` name the code/feature each task
  introduced (drawn directly from each task file's own text, or from
  grep-confirmed first appearance of the resulting CLI name), but git
  commits for that era were **not** independently re-verified in this
  pass and are marked `NOT VERIFIED IN THIS PASS` rather than guessed.

## Broken/missing provenance (not fabricated, not fixed)

Twelve issues found and recorded in `docs/audit/TASK70_ISSUES.md`, not
fixed silently, including: `go.mod`/`go.sum` not tracked in git at all (a
fresh clone cannot build — corrected from this same audit's own earlier,
wrong assumption that they were "force-added"); a `.gitignore`-vs-tracked
mismatch pattern across `experiments/`, `task69.tgz`, `profiles/`,
`data_test/`; two untracked experiment directories (Task55, Task57) at
real risk of loss; Task57's Phase B artifact missing entirely; inconsistent
git-commit provenance across Task58-67; an unverifiable seed/config_hash
claim in Task66's manifest; unknown licensing status for `data/` and
`ivtt/` (a real Task72 blocker); two large untracked tarballs duplicating
`experiments/` content; eight possibly-irreplaceable pre-tooling-era
comparative archives; README.md's pre-existing staleness (documents only
Stage1-23); and one **pre-existing, unrelated** data race in
`internal/conditionalregime`/`internal/positionalcontinuation`'s remote-
worker test harness, confirmed to reproduce identically on unmodified
`main` via an isolated worktree check — not introduced by this
reorganization, not fixed here.

## Legacy artifacts

Documented, not touched, per the resolved scoping decision: `workdir.v0/`
(240MB old scratch snapshot), `workdir-v2/` and `dataset/` (empty),
`task54-audit/` (empty), 8 `comparative-*`/`comporative-*` root archives
(possibly the sole surviving copy of pre-`experiment-compare`-era manual
comparative runs — classified LEGACY, not SCRATCH, for that reason), and
12 other redundant `.tgz`/`.tar.gz` duplicates of already-tracked
`experiments/` content. Full classification with rationale is in
`docs/audit/REPOSITORY_INVENTORY.tsv`.

## Unresolved issues

See `docs/audit/TASK70_ISSUES.md` for all 12, with severity and proposed
future action for each. None block this reorganization; three (untracked
`go.mod`/`go.sum`, untracked Task55/57 experiment directories, and the
`internal/positionalcontinuation` data race) are worth prioritizing in a
follow-up task, since two carry real loss risk and one is a correctness
bug independent of this repository's organization.

## Reproducibility smoke-test results

Five representative commands, run against temporary output locations,
never overwriting a frozen artifact (§16):

1. **Early structural (Stage1-3)**: `cmd/voinich` → `cmd/dict-analyze` →
   `cmd/structural-analyze`, chained into a fresh temp directory. **PASS** —
   all three built, ran, found `data_work/ZL3b-x7.txt`, and produced valid
   YAML output.
2. **Task60 (`research/phase1/token-repetition-analyze`)**: run with
   `-output-dir` pointed at a temp directory. **PASS (time-boxed)** —
   found its default corpus, built, and had written 10+ real TSV outputs
   before a 120s time budget was reached; not run to full completion given
   this pass's time budget.
3. **Task65 (`research/phase1/local-regime-topology-analyze`)**: this tool
   hardcodes its output directory as `experiments/local-regime-topology-v1`
   with no override flag, so it was run inside a from-scratch minimal
   isolated copy of the module (go.mod/go.sum, `internal/`, this one
   `research/phase1/` directory, and the corpus files it reads) rather
   than in place. **PASS, full completion** — produced all 25 documented
   output files; 23 of 25 are byte-identical to the real frozen
   `experiments/local-regime-topology-v1/`. The 2 that differ
   (`NATURAL_CONTROL_TOPOLOGY.tsv`, `TASK62_STATIONARY_CONTROL.tsv`) do so
   only because the minimal isolated copy intentionally omitted
   `data_test/`'s natural-corpus control files to keep the copy small —
   root-caused by inspecting the diff and the tool's own source, not
   assumed.
4. **Task67 (`research/phase1/recoverability-analyze`)**: same
   hardcoded-output-dir situation, plus a real dependency on Task66's
   frozen `experiments/mechanism-space-v1` artifacts as input — run in a
   similar minimal isolated copy that also included a read-only copy of
   `experiments/mechanism-space-v1` and the three `data_test/` control
   corpora it needs. **PASS, full completion** — produced all 26 documented
   output files, **byte-identical to the real frozen
   `experiments/recoverability-v1/`** in every file except `manifest.json`
   (which differs only in host/timestamp metadata, as expected). This is
   the strongest reproducibility confirmation in this pass: full,
   independent, from-scratch recomputation matching the frozen artifact
   exactly.
5. **Task68/69 literature artifacts**: confirmed `task68.tgz` (still at
   repo root — see below) and `experiments/task69-literature/task69.tgz`
   (moved to co-locate with the directory it mirrors) both still extract,
   and their contents still match the corresponding tracked files exactly.
   **PASS.**

One correction made during this smoke test: an earlier pass in this same
audit assumed `task68.tgz` was git-tracked like `task69.tgz`; `git mv`
failed with "not under version control," and direct verification confirmed
it is untracked/gitignored. It was left at the repository root rather than
moved into `docs/literature/`, consistent with the "document only, don't
touch" decision already made for the other untracked root archives.

## Build/vet/test/race/diff-check results

- `go build ./...` — **PASS**.
- `go vet ./...` — **PASS**.
- `go test ./...` — **PASS**, all packages.
- `go test -race ./...` — **one failure**,
  `TestPositionalContinuationTwoRemoteWorkersMatchLocalInAnyCompletionOrder`
  in `internal/conditionalregime`, confirmed pre-existing on unmodified
  `main` (see TASK70_ISSUES.md #11) and not caused by this reorganization.
  Not fixed here.
- `git diff --check` — **PASS** (one round of trailing-whitespace cleanup
  needed in the two newly-generated TSVs, fixed).

## Definition of done — status

1. Repository inventory created — `docs/audit/REPOSITORY_INVENTORY.tsv` (239 rows). ✅
2. Dependency map created — `docs/audit/DEPENDENCY_MAP.md`. ✅
3. Tools/research/pipelines/experiments separated — `cmd/`, `research/phase1/`,
   `pipelines/`, `experiments/` (unchanged). ✅
4. Phase I boundary defined — `research/phase1/README.md` (through Task69). ✅
5. Experiment artifacts preserved — `experiments/` untouched; only
   `task69.tgz` co-located, nothing recomputed. ✅
6. Task → experiment mapping created — `docs/audit/TASK_EXPERIMENT_MAP.tsv`
   (all 70 tasks + known variants). ✅
7. Provenance audit performed — `docs/audit/PROVENANCE_AUDIT.tsv` (13
   Phase I experiments, 13 fields each). ✅
8. Missing provenance explicitly marked — UNKNOWN/ABSENT/NOT VERIFIED used
   throughout rather than reconstructed values. ✅
9. Corpus organization documented — `corpora/README.md`. ✅
10. Hardcoded paths fixed — every `./<old-dir>` reference outside
    `tasks/*.txt` (deliberately untouched historical specs) updated;
    verified clean via a final repo-wide grep sweep. ✅
11. Representative reproducibility smoke tests pass — 5/5 (see above),
    two run to full byte-identical completion in isolated copies. ✅
12. build/vet/test/race/diff-check executed — all pass except one
    confirmed-pre-existing, unrelated race (documented, not masked). ✅
13. Scientific results unchanged — no algorithm, estimator, null model,
    threshold, or conclusion touched; verified by byte-identical
    smoke-test reruns of Task65 and Task67. ✅
14. Scientific/implementation issues found are documented, not silently
    fixed — 12 issues in `docs/audit/TASK70_ISSUES.md`. ✅
15. This report created. ✅

## What's next

This work sits on branch `task70-repo-reorg`, uncommitted, for review.
Recommended next steps once reviewed: commit, decide whether to merge to
`main`, then prioritize the `go.mod`/`go.sum` tracking decision and the
Task55/57 untracked-experiment loss risk before anything else in this
audit — those two are the only findings here with real data-loss exposure.
