# Fingerprint v2 lexical-paradigm artifact schema

Status: **implementation schema, not a frozen research result**. It covers
the first LP/EF block implemented by `cmd/fingerprint-v2-analyze`.

## Configuration

The command accepts `-config FILE`. YAML fields are strict; unknown fields
are errors. All paths are caller-supplied.

```yaml
version: fingerprint-v2-lexical-paradigms-v1
output_dir: results/fingerprint-v2-example
seed: 20260823
repetitions: 200
min_rule_support: 3
alpha: 0.05
graph_swaps: 10
diagnostic_tolerance: 0.20
grammar:
  modes: [structure-preserving, frequency-aware]
primary:
  id: corpus-id
  path: corpus.txt
  glyph_mode: eva             # eva or natural
  ivtff_path: corpus.ivtff    # optional; requires strict token alignment
controls:
  - name: prose-control
    corpus:
      path: prose.txt
      glyph_mode: natural
```

Generic inputs are whitespace-tokenized while retaining natural input-line
boundaries. `eva` applies `internal/evaglyph.CollapseEVA`; `natural` retains
lowercase Unicode letters and digits. When `ivtff_path` is supplied,
`internal/metadatavalidation.ParseIVTFF` and `Align` must match every token
in sequence or the run fails.

## Required artifact set

`output_dir` contains:

| File | Contents |
|---|---|
| `config.yaml` | Verbatim input configuration (or normalized programmatic config). |
| `fingerprint.json` | Compact deterministic fingerprint, provenance, controls and verdicts. |
| `raw_results.json` | Fingerprint plus raw null arrays and every C-GRAMMAR replicate. |
| `warnings.json`, `errors.json` | Explicit machine-readable diagnostics. |
| `report.md` | Compact human-readable summary. |

Writes are fail-fast: a failed artifact write makes the command fail.
`fingerprint.json` records implementation commit, input SHA-256, seed,
repetitions, preprocessing profile, metric/null-model versions and generator
settings. Output has no wall-clock field, so identical input/config/commit
and seed produce byte-stable JSON apart from an externally changed git commit.

## Metric fields

`metrics.lp1` is a type-level census of every *directed* edit-distance-one
pair. A rule is
`operation|zone|position_class|source→target`, where operation is
insertion/deletion/substitution, zone is prefix/internal/suffix, and
position class is Task60's begin/middle/end bucket. `support_gini` and
`top_rule_share` are calculated from distinct directed pair instances.
`support_threshold` is declared before analysis.

`metrics.lp2.tests` reports the LP1 support-Gini observed value, null mean,
sample SD, standardized effect, one-sided empirical p-value
`(1 + #{null >= observed})/(1 + repetitions)`, and BH FDR `q_value`.
`effect_defined` is false when a null has zero sample dispersion; the effect
is then mathematically undefined rather than encoded as non-JSON infinity.
The lexical FDR family contains all C-GRAMMAR, C-LEN, C-FREQ and LP4
attachment tests. C-LEN/C-FREQ are sampled random type pairs matched on
source/target length, with C-FREQ also matching log2 frequency bins.

`metrics.lp3` is calculated only when an LP2 C-GRAMMAR test passes the
declared FDR threshold. It contains connected components of the
productive-rule subgraph, exact all-source shortest-path depth, mean
branching, distinct-rule overlap, and same-line (and, where aligned,
same-page) family-occurrence locality. The same-line p-value uses
C-GLOBAL placement of family occurrences.

`metrics.lp4` declares a fixed zone rule: one prefix glyph, one suffix glyph,
and the nonempty interior core; tokens shorter than three glyphs are excluded
and counted. Prefix/core and suffix/core normalized MI are
`2*MI/(H(core)+H(affix))`. The permutation null shuffles affixes within token
length class (C-LEN). `prefix.grammar_tests` and `suffix.grammar_tests`
add every validated C-GRAMMAR mode to the same declared lexical FDR family.
All attachment tables are type-level: repeated token occurrences do not
increase a core/affix cell's support.

`metrics.ef1` reports edit-graph degree/component distributions including
isolates. `metrics.ef2` reports global transitivity, triangles, length-two
paths, and 4-cycles. Its control uses seeded simple-graph double-edge swaps
which preserve the exact degree sequence. `metrics.ef3` is Spearman
correlation between degree and `log1p(token frequency)`, with C-FREQ
frequency-label permutation. `metrics.ef4` applies a separately declared BH
family (including EF2/EF3 controls) to C-GRAMMAR comparisons of
giant-component share, clustering and absolute EF3 correlation, then returns
`CONSISTENT_WITH_GRAMMAR_BOUND`, `EXCEEDS_GRAMMAR_BOUND`, or `MIXED`.

## C-GRAMMAR

Each replicate preserves token count and the complete length distribution
exactly. It draws glyphs using length/absolute-position profiles combined
with a first-order local-transition weight. An alphabet-repair pass ensures
that every observed glyph remains available without replaying source token
strings.

`structure-preserving` generates independent tokens. `frequency-aware`
first generates and de-duplicates the same number of types within each length
class, then assigns the observed within-length frequency ranks to a seeded
permutation of those generated forms and duplicates them. Thus frequency
ranks are independent of generated form identity; no real edit family is
copied deliberately.

Each raw replicate includes token/length/alphabet exactness and total
variation diagnostics for positional glyphs, initial/final glyphs and
bigrams, plus vocabulary, singleton/rare-share and token-frequency
distribution diagnostics. `diagnostic_tolerance` applies to the four TV
diagnostics. A mode that fails validation remains in raw output for audit but
is excluded from FDR, productivity and EF4 inference. If no mode validates,
all grammar-dependent verdicts are `INCONCLUSIVE`.

## Interpretation constraints

All p-values are empirical finite-repetition estimates, not parametric
probabilities. LP and EF results sharing C-GRAMMAR are correlated evidence.
The implementation does not infer morphemes, language, cipher rules or a
Voynich interpretation from graph components. A result belongs to the
configured corpus only; no canonical-Voynich result is represented unless a
provenanced canonical input is actually supplied.
