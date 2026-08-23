# Task70b report

## Scope

Task70b repaired repository integrity and provenance only. It did not rerun an
experiment or change an algorithm, estimator, null model, threshold, corpus
normalization, randomization procedure, fingerprint, result, classification,
or scientific conclusion.

## Repository starting state

The starting branch was `task70-repo-reorg` at `16bb9bb` (`Task70:
reorganize repository into cmd/, pipelines/, research/phase1/, docs/`). Task70
was therefore already a separate commit. The only untracked non-ignored files
were the future task specifications `tasks/task70b.txt`, `tasks/task71.txt`,
and `tasks/task72.txt`; they were treated as user-owned inputs and not added.
Before edits, existing `go.mod` and `go.sum` hashes and the Task55/57 artifact
trees were recorded. Frozen tracked experiment files were not modified.

## go.mod/go.sum resolution

The existing files were preserved byte-for-byte and added without running
`go mod tidy`, `go get`, or an upgrade:

- `go.mod`: `6c37dee6e418237bc01503af1395bdd4f5beb3892e6e68170c6539abe52d74d9`
- `go.sum`: `e00472688564c3322cd6393c7bf847f618f23e70b172586f48060dec4a02b74d`

Both ignore rules were removed. `git ls-files go.mod go.sum` confirms that a
commit of the Task70b index will carry both files.

## Task55 preservation

Ten original `experiments/doyle__homophonic__*` result directories were
audited. Their manifests contain unique experiment IDs, timestamps, corpus
hashes, repository commits, dirty state, executor, and Go version. All ten
manifest corpus hashes match the referenced on-disk transformation corpora;
each `artifacts.json` agrees with its manifest identity and corpus hash; and
all 1,698 declared artifact SHA256 values match. Run states contain only
`completed` and legitimate `NOT_APPLICABLE` stage outcomes. The tree contains
text/JSON/YAML/TSV/SVG/GraphML reports, data and logs, with no credentials,
sockets, databases, or private keys.

The workspaces also contained 240 reproducible, host-specific ELF binaries
(about 1.70 GB) under `workspace/workdir/bin/`. Those build products were not
preserved in git and are now explicitly ignored. The manifests, logs, input
copies, and scientific outputs were added without changing their bytes.
The two pre-existing `*-analysis.tar.gz` archives were left untouched as
legacy archives, per scope.

## Task57 preservation

`experiments/inverse-homophony/synthetic-validation` is a self-consistent
Phase-A artifact: its manifest records method `inverse-homophony-v1`, config
seed 1, validation seeds 101--105, eleven corpus hashes, valid commit
`9737c991...`, and `gate_pass=false`. Seven raw result tables agree with the
report. The directory contains no scratch binaries, secrets, or private host
material and was added byte-for-byte.

## Task57 Phase-B determination

**NOT_RUN_BY_DESIGN.** The Phase-A report says the gate failed and Voynich was
not analyzed. Task57 §21 and §36, `INVERSE_HOMOPHONY_DESIGN.md`, and
`checkMethodFrozen` all require a passed synthetic gate and a
`METHOD_FROZEN` marker before Phase B. The marker is correctly absent and the
manifest records `gate_pass=false`. Phase B was intentionally prohibited;
there is no evidence of a lost artifact and no new run was made.

## Task58--67 provenance investigation

For each missing manifest-level commit, history gives an unambiguous final
artifact commit and the corresponding tree remains byte-identical through
Task70:

| Task | Reconstructed producing commit | Evidence |
|---|---|---|
| 58 | `ba88b2de62331a6638aacaa24112717ec48dd21b` | introduced at `a6904f0f...`; Task59 commit corrected the homophony-control result and is the last tree change |
| 61 | `1692874b0426e8f5502f42992e1fdab7f777cae0` | artifact introduced here; no later content change |
| 62 | `9e21b01e95237554b6b11c8607a57692ba119a59` | artifact introduced here; no later content change |
| 63 | `6cf53a6dd154161f14c04caa722201750dcaf2a6` | artifact introduced here; no later content change |
| 66 | `d485fe1c28034cb11481af7f4ae408739a92a05f` | artifact introduced here; no later content change |
| 67 | `0e8b3da23cef1fb64895a5ad8605d4952ebc6299` | replaces the `553ef025...` proxy scaffold with measured outputs; no later content change |

These are recorded as `RECONSTRUCTED_VERIFIED` in `PROVENANCE_AUDIT.tsv`.
Frozen manifests were kept byte-stable. Missing historical dirty-status or
top-level corpus fields remain `LEGACY_MISSING`; no provenance was guessed.

## Task66 deep provenance check

Verdict: **NOT_RECORDED** for the manifest's per-mechanism
`seed`/`config_hash` tracking claim.

The implementation does have the stated deterministic identity machinery:
the complete configuration is represented by `mechanismspace.Config`,
`Config.Hash()` SHA256-hashes all fields, and `Job.ID()` hashes experiment ID,
corpus, config hash, seed, and evaluation set. `RunGrid` derives replicate
seeds as frozen base seed plus replicate index; the run call sites fix base
seeds for screening/development and the other batteries, and
`MECHANISM_GRID.tsv` freezes every mechanism parameter combination.

However, the coordinator groups replicate results before writing the frozen
tables. No file in the entire artifact tree records a per-row seed,
configuration hash, or job ID. The values are deterministically derivable
from frozen source and grid configuration, but were not tracked by the
artifact as claimed. No retroactive values were added.

## .gitignore policy

`go.mod` and `go.sum` are repository files. `experiments/` contains
authoritative research artifacts and is intentionally trackable despite its
generated nature. Only embedded historical workspace binaries are excluded.
Comments document that profiles and `data_test/` are local by default while
the Doyle, Longfellow, and prepared Astafiev fixtures required by repository
tests/Phase-I tools are explicitly allowlisted. The fresh-checkout test caught
that Longfellow and prepared Astafiev had previously existed only in the local
ignored tree; both and Astafiev's preparation metadata are now tracked. Legacy
archives remain local unless explicitly reviewed. No bulk migration or
archive deletion was performed. `.gitattributes` disables whitespace diagnostics only
for `experiments/**`: frozen TSVs use trailing tab-separated empty fields, so
normalizing them would corrupt byte stability and recorded hashes. Source and
documentation retain the normal strict whitespace check. The same byte-stable
exception applies to natural `data_test/*.txt` fixtures, whose source line
endings and spacing are corpus data rather than formatting.

## Data-race investigation and resolution

Scope: **TEST_HARNESS_ONLY** for the observed race. Real remote workers are
separate processes and initialize a workspace once. The regression harness
runs two simulated workers in one process; both installed the same generic
target into package-level `Frozen*` variables while concurrent batteries read
them. A synchronized idempotent setter now avoids rewriting an already
installed identical target. This is an ordinary implementation fix: it does
not alter ordering, random streams, occurrence construction, computed values,
or result serialization. The existing two-worker test asserts deterministic
equality with the local result in arbitrary completion order and passes under
the race detector.

## Fresh-clone verification

A detached worktree made from the Task70b commit (not from untracked working
copy files) was used for the acceptance checks. An initial test correctly
failed because two ignored natural-corpus fixtures were absent; they were
added and a new clean worktree was then tested. It required no copied module,
corpus, or source files. Build, vet, ordinary tests, race tests, and
`git diff --check` results are recorded in the final validation section below.

## Remaining Task72 blockers

The redistribution/licensing status of `data/` and vendored `ivtt/` remains
OPEN and is still a public-release blocker. Legacy comparative archives,
large corpus tarballs, `workdir.v0/`, `task68.tgz`, and other archives were not
removed or reclassified in this task.

## Remaining scientific/provenance issues

No remaining issue makes Phase-I evidence materially ambiguous or exposes the
newly preserved Task55/57 results to immediate loss. Legacy dirty-status and
some top-level corpus/transcription fields remain unavailable and are labeled
as such. Task66's recording mismatch is explicit; its values were not
fabricated. Any decision to add those fields would require a separate future
artifact-format change, not a scientific rerun.

## Final validation

- `go build ./...`: PASS
- `go vet ./...`: PASS
- `go test ./...`: PASS
- `go test -race ./...`: PASS
- targeted local/two-remote-worker deterministic-equivalence race test: PASS
- fresh-clone build/vet/test/race: PASS
- `git diff --check`: PASS
- tracked module and Task55/57 preservation checks: PASS
- frozen pre-existing tracked experiment diff: empty

## Final readiness for Task71

**READY_FOR_TASK71.** Repository reconstruction is self-contained, the
at-risk Phase-I artifacts are preserved, Phase B's status is no longer
ambiguous, and all provable commit provenance is recorded without modifying
historical results. The Task72 licensing blocker does not block synthesis.
