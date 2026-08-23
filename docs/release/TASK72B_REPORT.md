# Task72b public release blocker report

## Classification: READY_FOR_PUBLIC_RELEASE

Task72b applied all owner decisions. After the generated-content audit found
two exact corpus copies inside frozen artifacts, the owner explicitly directed
their exclusion while requiring local preservation. No release blocker remains.

## Starting state

- Private/source commit: `19badafdcccd40120e74cc1fa349ac331df200e7`.
- Branch: `main`.
- The source worktree already contained numerous untracked experiment
  workspaces and task specifications. They were not imported wholesale or
  modified as scientific evidence. A minimal reviewed set of linked
  report/manifest/aggregate files was included later, without corpus payloads.
- The original private `.git` and its 91-commit reachable history were not
  rewritten.

## Decisions applied and removed material

- Removed tracked Astafiev source plaintext, normalized plaintext, and its
  preparation sidecar.
- Removed vendored `ivtt/ivtt.c` and compiled `ivtt/ivtt`.
- Removed `tasks/` from the public tree at the user's additional direction;
  its complete local backup is `/tmp/task72b-private-tasks/`.
- IVTFF source/ordinary derivatives were already untracked at HEAD and remain
  ignored. Ignore rules also cover local IVTT installations and Astafiev.
- Added the unmodified official Apache License 2.0 text. Its documented scope
  is original project source code only; external tools/data and non-code
  research/documentation artifacts are not automatically relicensed.

## Preserved provenance and external workflows

[DATA.md](../../DATA.md) is authoritative. It records exact source URLs,
versions/identity, local paths, source hashes, deterministic preparation, and
expected output hashes for IVTFF/ZL3b and Astafiev. Both upstream data pages
returned HTTP 200 on 2026-08-23.

IVTT is external. Its current upstream source matched SHA-256
`d17439299e3ffb9bf0101de611d19ebe472b518fe7dae096f224e0950fa70d45`.
The legacy pipeline now resolves `IVTT_BIN` or `ivtt` on `PATH` and preserves
the exact `-x7 INPUT OUTPUT` invocation. The bootstrap script validates input
and output hashes and never downloads or redistributes third-party bytes. It
also handles IVTT v2.4's observed successful exit status `3`, accepting it only
when the exact expected output checksum matches.

## Dependency licensing

- `gopkg.in/yaml.v3 v3.0.1`: MIT for listed libyaml-derived files and
  Apache-2.0 for remaining files; Canonical NOTICE verified.
- `golang.org/x/text v0.31.0`: BSD 3-Clause plus upstream PATENTS grant.

Neither dependency is vendored. No project NOTICE was invented; known
third-party notices and exact scope are recorded in `THIRD_PARTY_AUDIT.md`.

## Generated artifact content audit

SHA-256 comparison against the excluded IVTFF source, IVTT-derived corpus,
Astafiev source, and prepared Astafiev corpus found no exact Astafiev or IVTFF
source copy in tracked frozen artifacts. It did find two exact full copies of
the IVTT-derived canonical corpus:

- `experiments/voynich-v1/outputs/normalized_085.txt`;
- `experiments/voynich-v1/outputs/normalized_090.txt`.

Both equal SHA-256
`360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`.
They were initially left untouched under Task72b's scientific invariant and
reported as a blocker. The owner then explicitly directed their exclusion.
Their bytes were removed from the release tree and new clean public history;
unchanged local copies remain in `/tmp/task72b-private-frozen/`, and their
paths, hashes, and provenance remain in this report and release manifest.

The documentation-link audit also identified missing links into two untracked
experiment workspaces. The candidate includes only the eight required,
reviewed report/manifest/aggregate files. Inverse-transposition corpus payloads
and the 1.9 GiB Doyle private workspace remain excluded and ignored; none of
the selected files matched excluded corpus hashes or contained private paths.

## Sanitized history and secret scan

A separate clean-root Git repository was recreated at the reviewed release
state after the final corpus exclusions. Removed corpus/tool/task/frozen-copy
paths do not occur in its reachable object/path history. The private history
remains intact and must not be published.

Gitleaks 8.30.1 scanned both the candidate directory and its full reachable
clean history. Four initial `generic-api-key` false positives were reviewed as
scientific token/package/metric fields, recorded narrowly in `.gitleaksignore`,
and disclosed in `SECURITY_AUDIT.md`. Fail-on-finding reruns reported
`no leaks found`; no credential needs rotation.

## Fresh-candidate validation

The fresh clean-root candidate passed:

```text
go build ./...       PASS
go vet ./...         PASS
go test ./...        PASS
go test -race ./...  PASS
git diff --check     PASS
```

Tests were decoupled from bundled Astafiev bytes: the encoding unit test now
uses a synthetic CP1251 fixture, while the all-control integration test skips
with a documented reason only when the external corpus is absent.

## External-input smoke test

Using only the public instructions and external files staged outside the
candidate:

- IVTT v2.4 converted the validated IVTFF source to the expected
  `360d...87c2` output;
- Astafiev preparation produced the expected `ff67...ba5a` output;
- the mechanism-space three-control loader test passed without skipping;
- representative `cmd/voinich` and `cmd/dict-analyze` runs passed;
- `git status --ignored` marked every external source, derivative, sidecar,
  and local-tool path ignored (`!!`), with a clean tracked status.

The validated implementation/bootstrap candidate commit was
`19f8bdb67345a588fda2e7ec8999c4cf9d8b9f0a`; final audit metadata is committed
on top of it. The handoff's final candidate hash is obtained with
`git rev-parse HEAD` and reported to the owner alongside this file.

A code-span-aware whole-repository local Markdown link check passed after the
reviewed artifact subset and corrected relative paths were added. The two
documented upstream corpus/tool pages and the Astafiev source page returned
HTTP 200 during the audit.

## Size

After `git gc`, the final clean candidate contained one 498.95 MiB pack (3,040
objects). Logical tracked content was 3,545,677,319 bytes. Large frozen
evidence was retained; size itself is not a blocker.

## Remaining blockers

None. The owner-selected evidence-preserving strategy was implemented: the two
full-corpus payloads are external to the public artifact set, while their
paths, hashes, provenance, and lawful local regeneration instructions remain.

The final classification is:

**READY_FOR_PUBLIC_RELEASE**

No visibility change, public repository creation, push, release publication,
DOI creation, or announcement was performed.
