# Fixed positional-channel specification

This is the complete, finite channel family hard-coded in
`positional-analyze/main.go`; no alternative channels were searched before
the canonical execution. It is a statistical screen, **not** a message,
acrostic, or language detector.

| Channel | Unit and statistic | Matched null |
|---|---|---|
| `line_first_glyph`, `line_last_glyph` | adjacent NMI across line order | permute values within line-token-count strata |
| `line_first_token`, `line_last_token` | adjacent NMI across line order | same stratified permutation |
| `token_first_glyph`, `token_last_glyph` | adjacent NMI within physical lines | permute selected values within each line |
| `token_glyph_position_2`, `token_glyph_position_3` | adjacent NMI within lines | same within-line permutation |
| `token_ordinal_2_first_glyph`, `token_ordinal_3_first_glyph` | adjacent NMI across eligible lines | line-token-count-stratified permutation |
| `even_token_ordinal_first_glyph` | adjacent NMI within lines | same within-line permutation |

The estimator reports channel entropy and `2I(X_t;X_(t+1))/(H(X_t)+H(X_(t+1)))`.
The one-sided empirical p-value is `(1 + #{null >= observed})/(1 + R)`;
Benjamini-Hochberg corrects the 11 declared channels. `R=1000`,
seed `20260824`. The null preserves line boundaries, selected-position
marginals and, for one-value-per-line channels, line-length strata.

The canonical input was `data_work/ZL3b-x7.canonical.txt`,
SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`,
with EVA collapse. Run:

```sh
go run ./research/phase2/notation-audit/positional-analyze \
  -input data_work/ZL3b-x7.canonical.txt \
  -output research/phase2/notation-audit/POSITIONAL_RESULTS.tsv \
  -seed 20260824 -repetitions 1000
```

All channels have excess adjacent structure at the resolution of this run
(`p=q=0.000999001`; see `POSITIONAL_RESULTS.tsv`). This is compatible with
known positional/within-line organisation. It does **not** identify a hidden
message or establish acrostic, grille, shorthand, or steganographic encoding.
