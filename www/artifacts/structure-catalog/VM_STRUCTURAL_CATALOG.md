# Voynich Manuscript structural property catalog

## Scope and provenance

This is a descriptive/empirical catalog of what is and is not observed in the ZL3b canonical corpus. “Not observed” never means “impossible in Voynichese.” The primary corpus is data_work/ZL3b-x7.canonical.txt (SHA-256 f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692): 39380 tokens, 8244 unique token types, 5385 physical lines, 227 folios, 8 sections, and 36 literal transcription symbols. Metadata availability: true.

## Glyph rules

The literal-symbol inventory has 36 symbols. Of 1296 possible directed bigrams, 379 are observed and 917 are not observed in this corpus (constraint density 0.7075617284). Position-exclusive facts and every zero-valued pair are listed explicitly in GLYPH_POSITION_RULES.tsv and GLYPH_BIGRAM_RULES.tsv; conditional continuations through n=4 are in GLYPH_NGRAM_RULES.tsv.

Symbols not observed token-initially: `0`, `3`, `4`, `5`, `6`, `7`, `8`, `9`, `b`. Symbols not observed token-finally: none. These are literal-symbol facts for this corpus, not universal claims.

## Token formation

The catalog contains all 8244 token types with line position and document coverage. Literal edit-distance-one structure yields 1183 connected components and 29923 edges. Prefix/suffix patterns of length 1–4 are reported as patterns, not linguistic morphemes.

Highest-frequency prefix patterns: `o`:8939, `c`:7063, `ch`:6141, `q`:5401, `qo`:5269, `s`:4603, `d`:3678, `sh`:3243. Highest-frequency suffix patterns: `y`:15732, `dy`:6821, `n`:6093, `l`:6000, `in`:5958, `r`:5866, `iin`:4261, `edy`:4184. Entries are pattern:token-frequency-count; token-type productivity is retained in TOKEN_AFFIX_PATTERNS.tsv.

## Token sequencing

At frequency >= 10 there are 305809 possible ordered transitions: 12092 observed and 293717 unobserved corpus rules (constraint density 0.9604589793). The full 8,244-type adjacency complement is queryable without recomputation. Higher-order rows give both P(C|A,B) and P(C|B).

Strongest statistically preferred rows (support and effect are retained in the TSV/master catalog):

- `v → x` (TOKEN_TRANSITION): count 4, expected 0.002588615973, effect 1545.227273, q=4.824658097e-09, OBSERVED / STRONGLY_PREFERRED.
- `v → t` (TOKEN_TRANSITION): count 4, expected 0.003882923959, effect 1030.151515, q=2.236600696e-08, OBSERVED / STRONGLY_PREFERRED.
- `133 → hy` (TOKEN_TRANSITION): count 5, expected 0.005824385939, effect 858.459596, q=1.756733647e-10, OBSERVED / STRONGLY_PREFERRED.
- `shedain → pol` (TOKEN_TRANSITION): count 1, expected 0.001294307986, effect 772.6136364, q=0.005696762555, OBSERVED / STRONGLY_PREFERRED.
- `checkhey → oain` (TOKEN_TRANSITION): count 1, expected 0.001294307986, effect 772.6136364, q=0.005696762555, OBSERVED / STRONGLY_PREFERRED.
- `qopchdy → dshedy` (TOKEN_TRANSITION): count 1, expected 0.001411972349, effect 708.2291667, q=0.005970968047, OBSERVED / STRONGLY_PREFERRED.
- `c → 132` (TOKEN_TRANSITION): count 10, expected 0.01441388439, effect 693.7755102, q=0, OBSERVED / STRONGLY_PREFERRED.
- `qotal → yshey` (TOKEN_TRANSITION): count 1, expected 0.001588468892, effect 629.537037, q=0.006468299039, OBSERVED / STRONGLY_PREFERRED.

Strongest statistically avoided/depleted rows:

- `aiin → daiin` (TOKEN_TRANSITION): count 1, expected 9.51934108, effect 0.1050492877, q=0.004251964041, OBSERVED / DEPLETED.
- `s → daiin` (TOKEN_TRANSITION): count 1, expected 6.312398882, effect 0.1584183792, q=0.02582774501, OBSERVED / DEPLETED.
- `chol → aiin` (TOKEN_TRANSITION): count 1, expected 5.892395941, effect 0.169710252, q=0.0334859884, OBSERVED / DEPLETED.
- `chedy → ar` (TOKEN_TRANSITION): count 1, expected 5.650948669, effect 0.1769614376, q=0.03881399082, OBSERVED / DEPLETED.
- `dar → daiin` (TOKEN_TRANSITION): count 1, expected 5.622297397, effect 0.1778632344, q=0.03950396909, OBSERVED / DEPLETED.
- `qokain → daiin` (TOKEN_TRANSITION): count 1, expected 5.459920577, effect 0.1831528474, q=0.04396430088, OBSERVED / DEPLETED.
- `daiin → aiin` (TOKEN_TRANSITION): count 2, expected 10.80773643, effect 0.1850526253, q=0.006012605018, OBSERVED / DEPLETED.
- `or → daiin` (TOKEN_TRANSITION): count 2, expected 7.611413443, effect 0.262763285, q=0.03286258867, OBSERVED / DEPLETED.

## Line grammar

There are 35197 observed frequent unordered same-line pairs and 117431 exclusions (constraint density 0.7693935582). Line files preserve token/glyph lengths, endpoints, repetitions and token-family progression. Immediate transitions and same-line co-occurrence remain distinct relations.

Frequent tokens not observed line-initially (first 30): `132`, `133`, `a`, `aiiin`, `aiin`, `aiir`, `alam`, `aldy`, `aly`, `am`, `araiin`, `aram`, `arar`, `arody`, `ary`, `chaiin`, `cham`, `chan`, `chckhdy`, `chckhedy`, `chckhey`, `chcphy`, `chcthy`, `chdaiin`, `chdal`, `chdam`, `chdar`, `cheal`, `checkhey`, `checkhy`. Frequent tokens not observed line-finally (first 30): `132`, `133`, `aiir`, `as`, `ch`, `chckhdy`, `chdaiin`, `chdal`, `checkhey`, `chedar`, `cheeky`, `cheeo`, `cheeody`, `cheeor`, `chees`, `chekar`, `cheo`, `cheockhy`, `chkal`, `chkar`, `chokain`, `chokeey`, `chotar`, `chs`, `cphol`, `ctheol`, `ctho`, `dcheey`, `dchey`, `doiin`.

Strongest preferred same-line pairs:

- `132 → 133` (TOKEN_LINE_COOCCURRENCE): count 7, expected 0.03050975566, effect 229.4348102, q=3.643199076e-10, OBSERVED / STRONGLY_PREFERRED.
- `133 → hy` (TOKEN_LINE_COOCCURRENCE): count 5, expected 0.05483273177, effect 91.18641072, q=0.0001204621877, OBSERVED / STRONGLY_PREFERRED.
- `f → v` (TOKEN_LINE_COOCCURRENCE): count 3, expected 0.03327693447, effect 90.15253502, q=0.02903453604, OBSERVED / PREFERRED.
- `v → x` (TOKEN_LINE_COOCCURRENCE): count 3, expected 0.03659755072, effect 81.97269875, q=0.03358008364, OBSERVED / PREFERRED.
- `132 → c` (TOKEN_LINE_COOCCURRENCE): count 10, expected 0.1376433273, effect 72.65154216, q=1.016706719e-10, OBSERVED / STRONGLY_PREFERRED.
- `k → x` (TOKEN_LINE_COOCCURRENCE): count 4, expected 0.060901744, effect 65.67956412, q=0.004738772008, OBSERVED / STRONGLY_PREFERRED.
- `k → v` (TOKEN_LINE_COOCCURRENCE): count 4, expected 0.06642543796, effect 60.21789427, q=0.005870536171, OBSERVED / STRONGLY_PREFERRED.
- `133 → c` (TOKEN_LINE_COOCCURRENCE): count 8, expected 0.1513785381, effect 52.84765002, q=3.041760568e-07, OBSERVED / STRONGLY_PREFERRED.

## Document grammar

The locus, folio and section files enumerate each supported token/family against every available category, including every absence. An absence is a corpus fact; enrichment/depletion is a separate FDR-controlled inference. Locus-exclusive families: 0. Section-exclusive families: 0.

Strongest locus/section specializations:

- `okody → R` (TOKEN_BY_locus): count 3, expected 0.1549771458, effect 19.35769294, q=0.02079077416, OBSERVED / PREFERRED.
- `v → C` (TOKEN_BY_section): count 12, expected 0.6981208735, effect 17.18900044, q=1.674286351e-09, ONLY_IN_section / STRONGLY_PREFERRED.
- `v → C` (TOKEN_BY_locus): count 12, expected 0.7404773997, effect 16.20576132, q=3.183486474e-08, ONLY_IN_locus / STRONGLY_PREFERRED.
- `f → L` (TOKEN_BY_locus): count 5, expected 0.311833418, effect 16.03420195, q=0.001379776292, OBSERVED / STRONGLY_PREFERRED.
- `shes → A` (TOKEN_BY_section): count 5, expected 0.3203148807, effect 15.60964008, q=0.0005831330308, OBSERVED / STRONGLY_PREFERRED.
- `oteos → Z` (TOKEN_BY_section): count 13, expected 0.8814118842, effect 14.74906367, q=1.606812578e-09, OBSERVED / STRONGLY_PREFERRED.
- `dary → R` (TOKEN_BY_locus): count 3, expected 0.2096749619, effect 14.30786, q=0.04108182751, OBSERVED / PREFERRED.
- `otaly → L` (TOKEN_BY_locus): count 9, expected 0.6548501778, effect 13.74360168, q=8.100271398e-06, OBSERVED / STRONGLY_PREFERRED.

## Transcription stability

TRANSCRIPTION_STABILITY.tsv compares key positional, bigram, transition and line-co-occurrence corpus facts with IT2a. Literal relations lacking comparable token/symbol identities are marked NOT_COMPARABLE; no identity is manufactured across incompatible readings.

## Reuse and limitations

The implementation reuses the repository's audited strict IVTFF alignment and canonical provenance, and integrates literal edit-family, sequence, transition, boundary and positional analyses into concrete rules. It makes no linguistic, cryptographic, mnemonic, shorthand, numeric, procedural or generative-mechanism classification.
