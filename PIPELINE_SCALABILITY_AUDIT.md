# Pipeline scalability audit (Task53)

Дата: 2026-08-21. Источник runtime/CPU/RSS: `experiments/*/run-state.json`, а не
filesystem timestamps. Поэтому период выключенного компьютера в измерения не
включён.

## Результат аудита

Pipeline масштабируется по промежуточным структурам и числу null trials, а не
по одному `N`. T8 имеет `N=43,713`, `V=5,360`, но Stage 13 получает одну family
с 3,315 tokens и 3,170 edges. Это scientific result и он не отбрасывается.

| stage | Voynich wall / RSS | Doyle wall / RSS | T8 wall / RSS | фактический multiplier / bottleneck |
|---|---:|---:|---:|---|
| 13 structural-pair-decompose | 3.72s / 0.61GB | 34.99s / 5.60GB | 1702.24s / 29.8GB | target/family edges, controls, YAML+SVG output |
| 14 distance-context-analyze | 30.59s / 0.37GB | 24.45s / 1.21GB | 1098.06s / 5.82GB | family-edge × distance/context matrix |
| 17 structural-projection-analyze | 12811.0s / 13.95GB | 692.86s / 0.40GB | 8157.66s / 2.42GB | projection trials × structural graph; executor work units |
| 23 token-relation-validate | 54.52s / 0.61GB | 3957.28s / 0.78GB | 6457.29s / 2.22GB | permutation batteries × candidate/family work |
| 27 transition-network-validate | 94.96s / 0.28GB | 1153.72s / 3.01GB | 328.14s / 0.63GB | edge permutation/null work; corpus-dependent |

CPU confirms that wall time is not a timestamp artefact. T8 CPU seconds are
approximately 1,420 (14), 7,542 (17), 1,504 (23), and 323 (27). Stage 17 is
executor-enabled in the recorded pipeline configuration; the run-state alone
does not expose a per-job balance table, so adding workers is not justified by
wall time alone. Existing distribution audit identifies deterministic
replicate-level units for 23/27 and trial-level units for 17; algorithmic
redundancy must be removed before dispatching more jobs.

## Scaling variables

`N` and `V` describe corpus size, but the expensive stages are driven by:

* Stage 13: `P` selected pairs, `F` family edges, `C` controls and serialized
  scientific/presentation bytes. Runtime is approximately proportional to
  decompositions and output cardinality.
* Stage 14: selected pairs/family edges × distances × direction/boundary
  profiles.
* Stage 17: `R` projection trials × structural candidates/classes and the
  cost of each projection; executor preprocessing/postprocessing is serial.
* Stage 23: number of frozen candidates/families × permutation batteries ×
  block/profile work.
* Stage 27: observed directed edges × primary/refinement permutations and
  graph diagnostics.

## Stage 13 diagnostics and semantics

Before analysis the CLI prints `target_pairs`, estimated decompositions,
`family_count`, `largest_family_tokens`, `largest_family_edges`, and
`total_family_edges`. If the family is unusually large it additionally prints
`STRUCTURAL_CARDINALITY_EXPLOSION`. This is a warning only; no threshold,
family, target, control, or scientific row is truncated.

The new `-no-svg` flag only omits derived plots. YAML, TSV and Markdown remain
unchanged. With the default settings SVG behavior is unchanged, so existing
reproducible runs retain their presentation artifacts.

## Task52 candidate metrics

The following are support/diagnostic metrics, not distance features:

| metric | classification | distance eligible? |
|---|---|---|
| `structural.family_count` | scientific-structure support | no |
| `structural.largest_family_tokens` | structural cardinality diagnostic | no |
| `structural.largest_family_edges` | structural cardinality diagnostic | no |
| `structural.total_family_edges` | structural cardinality diagnostic | no |

They are not added to aggregate distance. Raw support values must remain
available for audit and interpretation.

## Distribution recommendation

Stage 17, 23 and 27 have deterministic independent units in the existing
executor. Distribution is justified only after profiling confirms that the
unit payload is large enough to amortize transport and serialization. Stage
13 is principally an in-process decomposition/serialization problem; remote
execution would multiply large payloads and is not the first remedy.
