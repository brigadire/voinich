# Experiment isolation audit (Task41)

## Reproduction and root cause

The pre-Task41 orchestrator ran every binary with the repository root as its
current directory. Analyzer defaults such as `workdir/sequence_analysis.yaml`
therefore resolved to the single repository `workdir/`, independently of the
experiment directory. Passing `-corpus` in generic mode changed only the raw
corpus input; implicit dictionary, class, pair, metadata, and report paths
still resolved to that shared tree.

The failure is deterministic with a synthetic corpus: pre-populate repository
`workdir/` with corpus-A results, create a corpus-B generic manifest, and run a
stage which consumes an implicit upstream path. It reads corpus A. A skipped or
`NOT_APPLICABLE` stage leaves its old files untouched. Finally, the old freeze
walked the shared tree and selected files using only `mtime >= runStartedAt`.
That is neither ownership nor provenance: a concurrent run, a rewritten stale
file, or a coarse/changed timestamp can enter the snapshot.

Thus the exact contamination chain was:

`generic corpus B -> repo-root process cwd -> implicit shared workdir input A
-> fixed shared output paths -> mtime-based broad freeze`.

Resume checked only a stage status string. It did not bind completed output to
the experiment, corpus, invocation, or content hash.

## Stage audit

All executables previously used repository-root cwd and fixed `workdir/...`
defaults. In isolated manifests every executable uses
`experiment/workspace/` cwd; consequently every relative `workdir/...` below
is experiment-local. The corpus argument is absolute and explicit. Exact files
created or changed by a stage are discovered after successful/failed execution
and recorded in `artifacts.json`.

| # | stage / executable | generated inputs | output set | old stale-read/output risk |
|---:|---|---|---|---|
| 1 | dict-gen | explicit corpus | `dataset/dictionary.yaml` | shared output |
| 2 | dict-analyze | dictionary | `dataset/tokens_analysis.yaml` | shared input/output |
| 3 | structural-analyze | dictionary, token analysis | `dataset/structural_analysis.yaml` | shared input/output |
| 4 | sequence-analyze | explicit corpus | `sequence_analysis.yaml` | shared output |
| 5 | begin-end-analyze | corpus, dictionary | begin/end YAML, TSV, report | shared dictionary/output |
| 6 | structural-normalize | corpus, structural analysis | classes and normalized corpora | shared input/output |
| 7 | normalization-compare | corpus, classes, sequence analysis, analyzer binary | normalized analyses/comparison | shared inputs/output |
| 8 | structural-validate | corpus, classes | `structural_validation.yaml` | shared classes/output |
| 9 | structural-profile-stability | corpus, classes | stability YAML | shared classes/output |
| 10 | structural-reliability | corpus, classes | reliability YAML | shared classes/output |
| 11 | soft-structural-space | dictionary, analysis, reliability | soft space YAML/pairs TSV | shared inputs/output |
| 12 | structural-graphemic | soft pairs | graphemic pair/ranking/report set | shared input/output |
| 13 | structural-pair-decompose | dictionary and graphemic pairs | distant pairs/families | shared inputs/output |
| 14 | distance-context-analyze | corpus, distant pairs/families/controls | distance-context result set | shared inputs/output |
| 15 | local-regime-analyze | corpus, distance pairs/controls | local-regime result set | shared inputs/output |
| 16 | property-trajectory-analyze | corpus, structural/distance pairs/controls | trajectory result set | shared inputs/output |
| 17 | structural-projection-analyze | corpus, pair/family artifacts | projection result set/checkpoint | shared inputs/output/checkpoint |
| 18 | global-regime-analyze | explicit corpus | global distributional result set | shared output |
| 19 | metadata-validate | explicit IVTFF and corpus, discovery tree | `metadata-validation/*` | Voynich defaults/shared output |
| 20 | cluster-metadata-global | discovery and explicit local metadata paths | metadata-global result set | inconsistent fixed metadata paths |
| 21 | conditional-regime-analyze | corpus and explicit local token map | `conditional-regimes/*` | Voynich/default metadata paths |
| 22 | residual-diagnostic-analyze | conditional results, corpus, token map | `residual-diagnostics/*` | shared inputs/output |
| 23 | token-relation-validate | corpus and token map | `token-relation-validation/*` | shared inputs/output/checkpoint |
| 24 | replicated-local-structure-audit | corpus, token map, relation results | `replicated-local-structure/*` | shared inputs/output/checkpoint |
| 25 | higher-order-sequence-validate | corpus, token map, replicated audit | `higher-order-sequences/*` | shared inputs/output/checkpoint |
| 26 | positional-continuation-validate | corpus, token map, higher-order results | `positional-continuation/*` | shared inputs/output/checkpoint |
| 27 | transition-network-validate | corpus and token map | `transition-network/*` | shared inputs/output/checkpoint |

Stages 1–18 are generic. Stages 19–27 are `NOT_APPLICABLE` for a generic
corpus and execute no process and own no artifact.

## Implemented invariants

- New manifests use isolation version 1 and a private `workspace/workdir`.
- The workspace must contain only registered artifacts. An unknown sentinel is
  a hard error before the next stage, resume, or freeze.
- `artifacts.json` binds path and SHA256 to ExperimentID, corpus SHA256,
  producing stage, and the hash of that stage's arguments.
- Before a stage, registered files are rehashed and their producer must be a
  completed earlier dependency (or the same failed stage being retried).
- Resume rehashes the corpus and requires the saved invocation hash.
- Freeze copies the registry allow-list only; it never walks the repository
  workdir. Verify checks manifest/run-state/registry consistency, producer
  applicability, provenance, registry hashes, and frozen checksums.
- Isolation metadata is prepended to every new stage log.
- Isolation-version-zero frozen experiments remain verifiable through the
  legacy checksum path; `experiments/voynich-v1` is not migrated or modified.

