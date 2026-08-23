# Task70 issues

Task70 is a repository-reorganization task, not new research. Every item
below was found while building the inventory/dependency-map/provenance
audit and is recorded here, not fixed silently, per Task70 §1/§10. None of
these affect any scientific algorithm, metric, threshold, classification,
or conclusion already published in a report.

## 1. `go.mod` and `go.sum` are not tracked in git at all

**Component**: repository root / build configuration.
**Finding**: `git cat-file -e main:go.mod` fails ("path 'go.mod' exists on
disk, but not in 'main'") and `git ls-files go.mod` returns nothing on both
`main` and this branch. `.gitignore` also lists `go.mod`/`go.sum`, so this
isn't an oversight in the ignore file — the files are genuinely untracked,
gitignored, working-tree-only. (An earlier automated pass in this same
audit incorrectly assumed they were "committed before the ignore rule was
added, or force-added" like `experiments/`; that assumption is corrected
here after directly verifying `git ls-tree`/`git cat-file` — always verify
tracked status directly rather than trusting a plausible-sounding claim.)
**Impact**: a fresh `git clone` of this repository cannot build — there is
no module file. Every build in this session works only because a
pre-existing `go.mod`/`go.sum` already sits in the working tree.
**Severity**: high (collaboration/release blocker).
**Proposed future action**: decide deliberately whether to track
`go.mod`/`go.sum` (removing them from `.gitignore`) or to document that
they must be supplied out-of-band before this repo is usable. Not decided
or changed by this task — it's a policy call, not a path fix.

## 2. `.gitignore` doesn't reflect what's actually tracked

**Component**: `.gitignore`.
**Finding**: `experiments/` is ignored, yet 587 files under it are tracked
(force-added at some point); `*.tgz` is ignored, yet `task69.tgz` is
tracked; `profiles/` is ignored, yet 6 `.pprof` files are tracked;
`data_test/` is ignored, yet 2 files are tracked. `go.mod`/`go.sum` are
listed as ignored (see #1) despite being required to build.
**Impact**: none currently (the force-added files stay tracked regardless
of the ignore rule), but the pattern is fragile — a careless `git add -f`
mistake, or a future contributor trusting `.gitignore` at face value, could
silently drop an experiment artifact from tracking, or silently commit
something meant to stay local.
**Severity**: low (process risk, not a current bug).
**Proposed future action**: either document the force-add convention
explicitly in a comment above each relevant `.gitignore` line, or move to
an explicit allowlist for `experiments/`. Not changed by this task.

## 3. Task55/57 experiment directories are not git-tracked at all

**Component**: `experiments/doyle__homophonic__*` (Task55),
`experiments/inverse-homophony/synthetic-validation` (Task57).
**Finding**: `git ls-files` returns 0 entries for these directories despite
each having a rich, well-formed manifest (git commit, dirty status, corpus
sha256, host, executor, go_version for Task55; git commit + dirty status
for Task57).
**Impact**: these are "historical scientific artifacts" per Task70 §7's
own framing, and they exist only on local disk. If this working copy is
lost, they are gone.
**Severity**: high (data-loss risk for two Phase I experiments).
**Proposed future action**: `git add -f` these directories (matching the
`experiments/` force-add convention already used for 587 other files) in a
follow-up commit, once someone signs off that the current on-disk content
is exactly what should be preserved. Not done here — Task70 doesn't
recompute or newly commit experiment content, only reorganizes what's
already tracked.

## 4. Task57 Phase B (canonical-Voynich application) has no located artifact

**Component**: `research/phase1/inverse-homophony` (`voynich.go`
implements the canonical-Voynich application described by the task spec's
Phase B).
**Finding**: no output directory or artifact for a Phase-B run was found
anywhere in the repository — only Phase A's synthetic-validation output
exists.
**Impact**: unclear whether Phase B was ever run, or whether its output was
lost/never committed.
**Severity**: medium (a described analysis phase appears to be missing its
result, not merely under-documented).
**Proposed future action**: check with whoever ran Task57 whether Phase B
executed; if so, recover its output; if not, note that explicitly in
`research/phase1/inverse-homophony`'s own docs rather than leaving it
ambiguous. Marked UNKNOWN in `docs/audit/PROVENANCE_AUDIT.tsv`, not
reconstructed.

## 5. Inconsistent git-commit provenance across Task58-67

**Component**: `experiments/rozanova-temerev-v1` (58),
`experiments/character-entropy-v1` (61), `experiments/token-formation-v1`
(62), `experiments/token-transition-v1` (63), `experiments/mechanism-space-v1`
(66), `experiments/recoverability-v1` (67).
**Finding**: these six experiments have no `git_commit` field in their
manifests at all, while Task59/60/64/65 do (and those four commit hashes
were independently verified to be real commit objects via
`git cat-file -t`). Seed and corpus-hash records are otherwise good for
most of the six.
**Severity**: medium (provenance gap, not a correctness issue — the
computations themselves aren't in doubt, just which exact commit produced
them).
**Proposed future action**: if the producing commit can be identified via
`git log --follow` timing (created-at timestamps in each manifest, cross-
referenced against commit timestamps), backfill the field. Not attempted
in this pass to avoid guessing at a fact that must be exact or absent.

## 6. Task66's per-mechanism seed/config_hash claim not centrally verifiable

**Component**: `experiments/mechanism-space-v1`.
**Finding**: the top-level manifest's `worker_contract` field claims
per-mechanism `config_hash`/`seed` tracking, but no such column was found
in `MECHANISM_GRID.tsv` or the other grid output files inspected in this
pass. Also has no centralized corpus sha256 at the top-level manifest (only
inherited by reference to Task58/59 artifacts).
**Severity**: medium (a documented guarantee that couldn't be independently
confirmed — may just require a deeper per-file read than this pass did).
**Proposed future action**: a follow-up pass should grep every file under
`experiments/mechanism-space-v1/` (not just the ones sampled here) for
`seed`/`config_hash` before concluding the claim is actually unmet.

## 7. Voynich corpus and vendored `ivtt` tool have unknown licensing status

**Component**: `data/` (IVTFF transcriptions mirrored from voynich.nu),
`ivtt/` (vendored C transliteration tool, also from voynich.nu).
**Finding**: `data/000_README.txt` documents origin and format in detail
but bundles no license text; `ivtt/ivtt.c` has no copyright or license
header at all (verified by grep).
**Severity**: high, but **explicitly out of scope for Task70** (§19: no
public-release cleanup, no licensing decisions). Recorded here as the real
release blocker it is, for Task72.
**Proposed future action**: contact voynich.nu / René Zandbergen for
explicit redistribution terms before any public release, per Task72.

## 8. Two large corpus tarballs contain full experiment workspaces, not just results

**Component**: `longfellow-song-of-hiawatha-v1.tar.gz` (42MB),
`msdos-v2-0-v1.tar.gz` (160MB), both at repo root, both untracked/gitignored.
**Finding**: each unpacks to `experiment/workdir/dataset/...` — a full
experiment workspace snapshot, not a final-results-only archive — and
duplicates content already present in `experiments/longfellow-song-of-hiawatha-v1/`
and `experiments/msdos-v2-0-v1/`.
**Severity**: low (disk usage only; not git-tracked, so no repository-size
impact).
**Proposed future action**: safe to delete once someone confirms the
unpacked `experiments/` directories are complete; not deleted by this task
per the scratch-cleanup decision (document only, touch nothing).

## 9. Eight `comparative-*`/`comporative-*` root archives may be irreplaceable

**Component**: `comparative-doyle-homophonic01/02/04/05.tgz`,
`comporative-report.tgz`, `comporative-voynich4/5/6/7.tgz`,
`comporative-voynich_doyle.tgz`, `comporative-voynich_doyle3.tgz`.
**Finding**: all untracked/gitignored; each unpacks to a `tmp/comparative-*`
naming scheme that doesn't match anything currently in `experiments/` —
these look like manual runs from before the `experiment-compare` tool
existed (Task45+), not duplicates of anything else on disk.
**Severity**: medium (possible sole surviving copy of early comparative
work; classified LEGACY rather than SCRATCH for this reason).
**Proposed future action**: before ever deleting these, confirm whether
their content is superseded by a current `experiments/doyle-sign-of-four-v1*`
run or similar; if not superseded, consider `git add -f` to preserve them
properly. Not touched by this task.

## 10. Root README.md is stale relative to the current pipeline

**Component**: `README.md`.
**Finding**: documents Stage1-23 in detail but never mentions Stage24-28,
`pipelines/pipeline-orchestrate`, `research/phase1/`, `experiments/`, or
the Task28-44 distributed-execution/mTLS infrastructure. This predates
Task70 — it was already true before any files moved.
**Severity**: low (documentation completeness, not correctness — every
path Task70 touched in this file was updated to remain accurate for what
it does describe).
**Proposed future action**: a documentation pass (out of scope here) should
extend README.md to cover Stage24-28 and point readers at
`docs/REPOSITORY_STRUCTURE.md` for everything else; Task70 added exactly
one such pointer paragraph rather than attempting the full rewrite.

## 11. Pre-existing, unrelated data race in `internal/conditionalregime`

**Component**: `internal/positionalcontinuation.LoadForDistribution` /
`buildBoundaryDistanceRows` (read) vs. `findSAiinOccurrences` (write) —
concurrent access to shared state from two simulated remote-worker
goroutines in `internal/conditionalregime`'s
`TestPositionalContinuationTwoRemoteWorkersMatchLocalInAnyCompletionOrder`.
**Finding**: `go test -race ./...` fails this one test. Verified via an
isolated git worktree of unmodified `main` (with `go.mod`/`go.sum` copied
in, since they aren't tracked — see #1) that the same failure reproduces
identically without any Task70 change present. This is a pre-existing
implementation bug, not something the reorganization introduced.
**Severity**: medium (a real race in test-only remote-worker simulation
code; unclear yet whether it also affects the real remote-worker path or
only the two-goroutine-in-one-process test harness).
**Proposed future action**: needs its own investigation and fix as a
regular bug-fix task — not attempted here, per Task70 §1's explicit
instruction not to fix scientific/implementation bugs silently while doing
a reorganization.

## 12. `corpus_size_audit_test.go` deliberately stayed at the repository root

**Component**: `corpus_size_audit_test.go`.
**Finding**: this is a `package main`-declared lint guard (not a unit test
of any specific package) that walks `filepath.WalkDir(".", ...)` from its
own working directory, expecting that to be the repository root. It was
**not** moved into `cmd/voinich/` alongside `main.go`/`main_test.go`
because doing so would silently narrow its scan to `cmd/voinich/`'s
subtree instead of the whole repository — the opposite of what it exists
to guard against.
**Severity**: informational (this is a design decision this task made, not
a problem found).
**Proposed future action**: none required; flagging here so a future
reorganization doesn't "fix" this by moving it without re-checking why it
sits alone at the root.
