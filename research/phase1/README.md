# Phase I

Phase I is the completed set of research spanning Task1 through Task69
(inclusive). This file fixes that boundary and maps the scientific
directions Task55-69 opened to their experiments and code; it is not the
full Phase I synthesis (that is Task71's job, per Task70 §12).

Tasks 1-54 built and audited the Stage1-28 analysis pipeline itself
(`cmd/`, orchestrated by `pipelines/pipeline-orchestrate/`) plus a handful
of standalone experiment tools (`corpus-transform` Task46,
`inverse-transposition-search` Task54, `voynich-validation` Task54b). Task55
onward shifted to independent, narrowly-scoped hypothesis tests that never
touch Stage1-28 code (verified: zero import-graph coupling in either
direction — see `docs/audit/DEPENDENCY_MAP.md` §4). Task57 continued that
standalone-experiment pattern (`inverse-homophony`). This directory holds
all of it.

## Scientific directions → experiments

| Direction | Task(s) | Code | Experiment artifact | Report |
|---|---|---|---|---|
| Homophonic-cipher corpus generation (frequency-band and uniform variants) | 55 | `corpus-transform` (`internal/corpustransform`) | `experiments/doyle__homophonic__*` | manifest only, no REPORT.md |
| Homophone-model / cipher-method literature comparison | 56 | — (documentation only) | — | `docs/literature/ROZANOVA_TEMEREV_METHOD.md`, `docs/literature/FREQUENCY_HOMOPHONE_MODEL.md`, `docs/literature/OUR_PIPELINE_METRIC_INVENTORY.md` |
| Blind homophone-class recovery + synthetic validation | 57 | `inverse-homophony/` (`internal/inversehomophony`) | `experiments/inverse-homophony/synthetic-validation` (Phase A only — Phase B/canonical-Voynich application has no located output, see TASK70_ISSUES.md) | `SYNTHETIC_VALIDATION_REPORT.md` |
| Rozanova–Temerev method reproduction | 58 | `rozanova-temerev/` (`internal/evaglyph`) | `experiments/rozanova-temerev-v1` | `REPORT.md` |
| Glyph-position distributional test | 59 | `glyph-position-analyze/` (`internal/evaglyph`) | `experiments/glyph-position-v1` | `REPORT.md` |
| Token-repetition/adjacency test | 60 | `token-repetition-analyze/` (`internal/tokenrepetition`) | `experiments/token-repetition-v1` | `REPORT.md`, `METHOD.md` |
| Character-entropy test | 61 | `character-entropy-analyze/` (`internal/characterentropy`) | `experiments/character-entropy-v1` | `REPORT.md`, `BOWERN_ENTROPY_METHOD.md` |
| Token-formation test | 62 | `token-formation-analyze/` (`internal/tokenformation`) | `experiments/token-formation-v1` | `REPORT.md`, `TOKEN_FORMATION_DESIGN.md` |
| Token-transition test | 63 | `token-transition-analyze/` (`internal/tokentransition`, delegates to `internal/tokenrepetition`) | `experiments/token-transition-v1` | `REPORT.md`, `TOKEN_TRANSITION_DESIGN.md`, `TRANSITION_MODEL.md` |
| Line-level regime test | 64 | `line-regime-analyze/` (`internal/lineregime`) | `experiments/line-regime-v1` | `REPORT.md`, `LINE_REGIME_DESIGN.md`, `REGIME_MODEL.md` |
| Local-regime topology test | 65 | `local-regime-topology-analyze/` (`internal/localregimetopology`) | `experiments/local-regime-topology-v1` | `REPORT.md`, `LOCAL_REGIME_TOPOLOGY_DESIGN.md` |
| Mechanism-space transformation-grid search | 66 | `mechanism-space-analyze/` (`internal/mechanismspace`) | `experiments/mechanism-space-v1` | `REPORT.md`, `TASK66_DESIGN.md` |
| Recoverability / decoder analysis | 67 | `recoverability-analyze/` (`internal/mechanismspace`, `internal/tokenrepetition`) | `experiments/recoverability-v1` | `REPORT.md`, `CANDIDATE_SELECTION.md`, `DECODER_DESIGN.md` |
| Focused literature audit (claims, contradictions, discriminators) | 68 | — (documentation only) | — | `docs/literature/*` (14 files, frozen snapshot at `docs/literature/task68.tgz`\*) |
| Focused literature search (Task58-67 follow-up) | 69 | — (documentation only) | `experiments/task69-literature` | `TASK69_SYNTHESIS.md`, `RESEARCH_RECOMMENDATION.md`, `SEARCH_LOG.md`; frozen snapshot at `experiments/task69-literature/task69.tgz` |

\* `task68.tgz` (root, untracked/gitignored) duplicates the same 15 files
now under `docs/literature/` and was left where it was per the Task70
scratch-cleanup decision — it isn't tracked in git, so it isn't a "canonical
experiment path" to preserve, just a local backup copy.

## Provenance summary

Git-commit provenance is inconsistent across this era: Task59/60/64/65
record a verified git commit; Task58/61/62/63/66/67 do not (despite
otherwise-good seed/corpus-hash records for most of them). Task55's and
Task57's experiment directories are not tracked in git at all — a real
loss-risk flagged in `docs/audit/TASK70_ISSUES.md`, not fixed here. Full
detail is in `docs/audit/PROVENANCE_AUDIT.tsv`.
