# Prior-work comparison and control audit

This is a metric-level comparison of Greshko (2025), Rozanova & Temerev
(2026), Averyanov (2026) with Stages 1–28 and Tasks 46/49/52/54/55.
The scientific pipeline was not changed.

## Provenance and evidence weight

| work | status at audit date (2026-08-21) | primary source | code/data |
|---|---|---|---|
| Michael A. Greshko, *The Naibbe cipher: a substitution cipher that encrypts Latin and Italian as Voynich Manuscript-like ciphertext* (2025) | **peer-reviewed journal article**, Cryptologia, online 2025-11-26 | [publisher/DOI](https://doi.org/10.1080/01611194.2025.2566408) | [author repository](https://github.com/greshko/naibbe-cipher), [Zenodo supplementary data](https://doi.org/10.5281/zenodo.16415087) |
| Liudmila Rozanova & Alexander Temerev, *A Glyph Is Not a Letter, a Token Is Not a Word, a Space Is Not a Space: What the Units of Voynichese Are Not* (2026) | **arXiv preprint**, dated 2026-08-17; no peer-reviewed venue in the primary record | [arXiv:2608.17096](https://arxiv.org/abs/2608.17096) | no repository stated in the arXiv record |
| Vitaly Averyanov, *A Workshop Cipher: a Generative Model Reproducing the Statics and Dynamics of the Voynich Manuscript Text* (2026) | **independent author-hosted web publication / preprint archive**, not a peer-reviewed journal article in the inspected source | [author page](https://voynich.site/paper-1-mechanism?lang=en), claimed [Zenodo archive](https://doi.org/10.5281/zenodo.21761192) | no public code repository identified |

Evidence weight is therefore: Greshko publication plus released artifacts >
R&T preprint > Averyanov independent preprint. Date alone is not a peer-review
claim.

## Greshko: model audit

The publisher specifies 20 reference runs: ten 52-card and ten 78-card runs,
each encrypting the first 8,000 letters of four texts (Dante, Pliny Book 16,
Grosseteste, and a 14th-century Latin alchemical herbal). Standard respacing
uses 47.2% unigrams and 52.8% bigrams; pooled summaries are about 20,000
ciphertext tokens. These are letters, not plaintext words or syllables.

Naibbe removes plaintext spaces, splits into single letters or adjacent letter
pairs, and emits each unit through one of six structured substitution tables.
The card/deck variant supplies table-choice randomness. The result is a
verbose, homophonic, reversible substitution with Voynich-like strings. It is
stochastic in representation choice but is not previous-token-conditioned.
Besides homophony it has verbosity, respacing, structured prefix/suffix/table
inventories and plaintext letter/bigram frequency leakage. This is not the
same mechanism as Task55's frequency-ranked allocation.

### Naibbe quantitative metrics and mapping

| Greshko property / metric | exact object or interpretation | closest implementation | status |
|---|---|---|---|
| character entropy | Shannon entropy of glyph frequencies, token-level EVA/Cuva | Stages 2/3 inventories; Stage 27 entropy is graph/transition entropy | CLOSE_ANALOGUE |
| conditional character entropy | first-order glyph conditional entropy | Stages 24/27 only with the same glyph stream | PARTIAL_OVERLAP |
| glyph/glyph-pair frequencies | marginal and within-token adjacent glyph-pair frequencies | Stages 2/3 context summaries | CLOSE_ANALOGUE |
| token-length distribution | glyph count per whitespace token | Stage 2 token lengths | CLOSE_ANALOGUE |
| word-type-length distribution | lengths of distinct token types | Stage 2 type inventory | PARTIAL_OVERLAP |
| Zipf/frequency-rank spectrum | ranked type frequencies and commonest-token share | Stage 2 frequency spectrum; Task52 fields | CLOSE_ANALOGUE |
| Heaps law | V(N) versus token count and VMS-like diversity | Stage 28 V(n), K, beta, R2 | DIRECT_EQUIVALENT |
| hapax legomena | singleton type count/proportion | Stage 28 hapax/fractions | DIRECT_EQUIVALENT |
| skewed word pairings | over/under-represented adjacent token-type pairs | Stage 4 n-grams; Stage 27 adds null inference | CLOSE_ANALOGUE |
| repeated tokens/runs | consecutive repetition and token autocorrelation claims | Stage 4 repetitions; Stage 27 self edges | CLOSE_ANALOGUE |
| edit-distance sequential clusters | Levenshtein-near tokens in sequence | Stages 3, 12, 13, different edge definition | PARTIAL_OVERLAP |
| edit-distance-1 network | token types joined by edit distance 1 | Stages 12/13 structural/graphemic graph | PARTIAL_OVERLAP |
| moving-average TTR / morphological complexity | moving TTR and common-word measures | Stage 28 TTR/growth, different definition | PARTIAL_OVERLAP |
| 42-metric Z-scored Euclidean distance | distance to VMS over Gaskell–Bowern's 42 metrics | Task52 schema is different | NO_EQUIVALENT |
| random-forest classification | classifier label in the 932-document comparison | no classifier/oracle in our fingerprint | NO_EQUIVALENT |

The publisher material does not evidence a Naibbe-specific transition-network
permutation test, held-out prediction, higher-order CMI, boundary AUC,
position-stratified continuation, or shuffled Heaps null. They must not be
added to its metric list because our pipeline has them.

### Naibbe versus Task46/55 controls

| property | Naibbe | global-H | triangular-v1 | frequency-v1 |
|---|---|---|---|---|
| plaintext unit | 1 or 2 letters | existing whitespace token | existing whitespace token | existing whitespace token |
| allocation | six structured tables/deck variant | fixed H | fixed H | rank-quantile H in [1,Hmax] |
| selection | card/table randomness | uniform | decreasing (H-k)/sum | independent uniform/triangular |
| frequency-aware | plaintext frequencies affect outputs/table fit | no | no | allocation only |
| context-aware | no learned previous-token state | no | no | no |
| variable H | table/unit mechanism | no | no | yes |
| reversible | yes with tables | yes with mapping | yes | yes |
| historical claim | 15th-century hand-cipher motivation | control only | control only | control only |
| token expansion | glyph strings; spaces deleted/respaced | one opaque token/input | same | same |

Task55 is therefore an **OUR_EXTENSION**, not a Naibbe replication or a claim
of historical plausibility. Stages 23–27 are also an extension: they test
replication, higher-order dependence, positional structure and network nulls
that Greshko's 42-metric distance does not establish.

## Rozanova & Temerev: units, controls and mapping

The paper tests three assumptions: transcription glyphs are letters,
blank-delimited strings are words, and every separator is a word boundary. It
uses Zandbergen–Landini, matched continuous-text/cipher/pseudotext controls,
within-line shuffles, fitted order-3 surrogates, five cipher seeds/configs and
quire-level resampling.

Reported quantities include 7,022 types, 69.7% singleton types, top-50 coverage
33.2%, shuffle-corrected token-identity order share 0.79%, adjacent edge-glyph
MI 0.197 bits, uncertain-vs-certain separator internality difference 0.465,
paragraph-first line-start JSD 0.528 versus 0.179 for other lines, and
composite-EVA H(X)=3.98, H(X[t+1]|X[t])=2.69, perplexity 6.45. These are
different estimands: glyph conditional entropy is not token-order MI.

Controls named in the primary record include Pliny botanical Latin,
Celsus/Caesar Latin, Culpeper English herbal, Italian prose, a catalogue,
Naibbe, a self-citation generator, within-line shuffles and calibrated
synthetic ciphers. Matching is per test (token count, line template or quire),
not one universal pooled null.

| R&T metric | definition | closest ours | status |
|---|---|---|---|
| marginal/conditional glyph entropy | H(X) and first-order glyph conditional entropy | Stages 2/3 and 24/27 | PARTIAL_OVERLAP |
| token identity order share | shuffle-corrected adjacent token MI, 2,000-type cap / capped entropy | Stages 4/27, no identical capped ratio | CLOSE_ANALOGUE |
| edge-glyph MI | last glyph of token to first glyph of next token | Stage 27 only with glyph stream; production graph is token-level | PARTIAL_OVERLAP |
| BPE learned-unit curve | dependence as recurrent units are merged | no BPE stage | NO_EQUIVALENT |
| boundary internality | normalized edge cohesion against within-token anchor | Stages 5/26 | PARTIAL_OVERLAP |
| separator physical AUC | uncertain separator classification from image gap coordinates | no image/AUC stage | NO_EQUIVALENT |
| line-start JSD/enrichment | line-initial vs mid-line glyph distributions | Stages 5/26 | CLOSE_ANALOGUE |
| vocabulary/singletons/top-50 | counts/coverage at matched N | Stages 2/28 | DIRECT_EQUIVALENT for counts; PARTIAL for protocol |
| calibrated substitution attack | mapping/language-match differential against fitted null | no equivalent; Task54 is inverse transposition | NO_EQUIVALENT |

Our result “homophony weakens transition structure but does not reproduce
sequence organization and the vocabulary spectrum” is **partial replication +
extension**, not independent replication of the full R&T profile and not a
contradiction. Exact replication would require their glyph-edge MI, capped
token MI, BPE and separator estimators on their representations and controls.

## Averyanov: model and ablation audit

The inspected source calls this independent research and points to a Zenodo
preprint archive; no peer-reviewed venue was identified. The public framing
describes a workshop/table generator with homophonic tables, serial reuse,
boundary/sandhi-like operations, anti-repetition and section variation, plus
static and dynamic validation. The inspected public material does not provide
a reproducible component-by-component algorithm and ablation table sufficient
to attribute each property to one mechanism.

| mechanism/property | source-level finding | closest ours | status |
|---|---|---|---|
| homophonic tables | component stated; exact key/table parameters require archive | Task46/55 homophony | CLOSE_ANALOGUE |
| serial reuse | dynamic component stated; no isolated effect | no serial-reuse transform | NO_EQUIVALENT |
| sandhi/boundary operation | public framing; exact estimator/parameter not established | Stages 5/26 | PARTIAL_OVERLAP |
| anti-repetition | claimed mechanism; no located ablation | Stages 4/27 response measures | PRIOR_WORK_GAP |
| section variation | model-level variation | Stages 15/16/18/21 | CLOSE_ANALOGUE |
| static Zipf/length/TTR/entropy | validation family | Stages 2/28; entropy partly same | CLOSE_ANALOGUE |
| dynamic signatures | validation family | Stages 23–27 | CLOSE_ANALOGUE |
| component ablation | not located in public source | one transform to response family is explicit in Task46/55 | PRIOR_WORK_GAP / OUR_EXTENSION |

Current evidence warrants the label
GENERATIVE_MODEL_WITHOUT_COMPONENT_IDENTIFICATION: a model-level fit is not
evidence that one named mechanism caused a particular match.

## Master metric matrix

EXACT means the same estimand and unit, not merely a shared name.

| property | Greshko | R&T | Averyanov | our stage | status |
|---|---|---|---|---|---|
| corpus size | matched 8k letters/source, ~20k tokens pooled | matched token/line templates | matched model comparisons stated | 1/2/28 | ANALOGUE |
| vocabulary | Zipf/Heaps/type inventory | 7,022 types | static target | 2/28 | EXACT |
| hapax/dis-legomena | hapax | 69.7% singleton types | hapax-rich target | 28 | EXACT |
| Heaps growth | explicit VMS-like relationship | static V/singletons | statics | 28 | EXACT / ANALOGUE |
| Zipf/frequency spectrum | explicit | top-50/type spectrum | explicit target | 2/28 | EXACT |
| token length | explicit | inventory | explicit target | 2 | ANALOGUE |
| repetition | repeats/autocorrelation | order/shuffle | dynamics | 4/27 | ANALOGUE |
| n-grams | skewed pairs | token MI | dynamics | 4/27 | ANALOGUE |
| conditional dependence | glyph H | glyph H and token MI | conditional H target | 24/25/27 | PARTIAL |
| transition network | no exact graph null | edge/token MI | dynamics | 27 | OUR_EXTENSION |
| higher-order dependence | not evidenced | order-3 surrogate, not CMI | not isolated | 25 | OUR_EXTENSION |
| positional/boundary structure | not primary | line-start JSD, separator AUC | boundary mechanism | 5/26 | PARTIAL |
| local/section regimes | plaintext topic/language effects | quire/Currier stratification | section variation | 15/16/18/21 | CLOSE |
| structural token families | table grammar/edit clusters | BPE units/grammar | generator units | 3/11/12/13 | PARTIAL |
| graphemic similarity | Levenshtein | learned units/edge glyphs | generator strings | 12/13 | CLOSE |
| vocabulary openness | hapax limitation remains | open hapax-rich vocabulary | target property | 28 | PRIOR_WORK_GAP |
| entropy | character H and conditional H | glyph H and token MI | conditional H | 3/24/27 | PARTIAL |

## Master control matrix

| control | Greshko | R&T | Averyanov | ours |
|---|---|---|---|---|
| natural prose | four Latin/Italian historical texts | matched Latin/English/Italian prose | stated model corpora | Doyle, Longfellow, Astafiev and canonical controls |
| cipher | Naibbe 52/78-card variants | Naibbe and synthetic ciphers | Workshop variants | global-H, triangular-v1, frequency-v1 |
| pseudo-text | 932-document reference corpus | self-citation and fitted surrogates | generator comparisons | transformed controls, no semantic oracle |
| shuffle | not universal central null | 100 within-line shuffles | model-specific calibration | Stage 28 token shuffle; 23–27 block permutations |
| matched size | 8k letters/source | token count/line template per test | claimed matched validation | Task52 common checkpoints; missing = NOT_APPLICABLE |
| boundary | respacing | certain/uncertain/erased separators | boundary mechanism | begin/end, line/block and positional policies |
| held-out | not central | calibrated attack surrogates | not component-specific | Task54 80/20 limitation; 24/27 held-out/LOBO |
| multiplicity | 42-vector distance/classifier | calibrated controls/bootstrap | not sufficiently documented | frozen families/FDR/global permutation rules |

## Conclusions and gaps

1. Greshko is a strong static analogue for frequency spectrum, hapax, Heaps,
   lengths and glyph entropy, but not for our order, higher-order or boundary
   validation. Task55 is an explicitly mechanistic extension, not history.
2. R&T is closest to our sequence/boundary direction, but glyph-edge MI,
   token-MI caps, BPE scale and separator AUC are absent from our production
   code. Our homophony result is partial replication plus extension.
3. Averyanov is a lower-weight independent preprint. Without ablations,
   mechanism-to-property attribution is not established.
4. Important prior-work gaps are component ablation, block/held-out
   replication, higher-order nulls, transition-network nulls and
   vocabulary-growth nulls. Important measurements missing in ours are exact
   R&T glyph-edge MI, BPE curves, separator physical AUC and calibrated
   substitution-attack differential.

## Local evidence

[OUR_PIPELINE_METRIC_INVENTORY.md](OUR_PIPELINE_METRIC_INVENTORY.md),
[corpus-transform/TRANSFORMATION_METHODS.md](../../research/phase1/corpus-transform/TRANSFORMATION_METHODS.md),
[FREQUENCY_HOMOPHONE_MODEL.md](FREQUENCY_HOMOPHONE_MODEL.md),
[VOCABULARY_GROWTH_ANALYSIS.md](../../cmd/vocabulary-growth-analyze/VOCABULARY_GROWTH_ANALYSIS.md),
[EXPERIMENT_COMPARISON.md](../../cmd/experiment-compare/EXPERIMENT_COMPARISON.md), and
[INVERSE_TRANSPOSITION_METRIC_AUDIT.md](../../research/phase1/inverse-transposition-search/INVERSE_TRANSPOSITION_METRIC_AUDIT.md).
