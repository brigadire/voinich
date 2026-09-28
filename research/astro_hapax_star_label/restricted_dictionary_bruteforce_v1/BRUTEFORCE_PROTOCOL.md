# Restricted dictionary brute-force v1 — preregistration

This protocol is frozen by `python3 run.py freeze` before any production score.
`FREEZE.json` hashes all inputs, engine, preparation, runner and tests. Production
and resume reject modified frozen bytes. No score-driven changes to v1 are allowed.
This is an anonymous dictionary compatibility test, not decipherment.

## Inputs and scope

Use exactly 57 occurrences in TARGET_SCOPE.tsv: f68r1=30, f68r2=27, hapax=26,
non-hapax=31. Occurrence IDs, line and block IDs, readable EVA and composite keys
are retained. Verify TARGET_SETS.tsv and source metadata against their upstream
registered SHA256; independently recount corpus-wide composite-key frequencies.
No spatial star identity ground truth is assumed. The previous enrichment
COHORT_MEMBERS.tsv and run_enrichment.py do not match that package's SHA256SUMS;
they are NOT trusted inputs. Rebuild strata from verified metadata. Record that
pre-existing discrepancy; do not repair or rewrite upstream.

Astronomy: all 63 STAR attestations of frozen expanded D1; 31 source canonical
identities; deduplicate normalized forms within each identity for search, retain
all attestations for provenance. Arabic transliteration is inherited editorial
transliteration, not new evidence. Latin, medieval Latinized Arabic and Arabic
representations are included. Planets, zodiac and unidentified pointers excluded.
No modern forms or generated historical spellings. No Alfonsine source is added
without direct attestation. The corpus is small and not exhaustive.

Historical controls: frozen pools from Isidore Etymologiae VII (266 nonastronomical
personal-name forms), XVII (ordinary Latin word types), and reviewed vocabulary
from the Arabic-Latin alchemical tradition edited by Ruska 1935, Latin pp.54–83.
The last source is a medieval tradition transmitted in an early printed witness;
its exact letterforms are editorial/early-print evidence, not dated manuscript
readings. Preserve this qualification. Isidore is early medieval; sampling is
length-matched, not matched by century or language. These are explicit limits.

Sample without replacement, in stable astronomical identity/form order, a random
nearest-length source form per astronomical spelling slot. Retain the same 31
capacity blocks and the same number of distinct spelling opportunities per block.
Control blocks are arbitrary anonymous matching units, NOT assertions that the
sampled names or lexical words are synonyms. This tests dictionary string signal
under equal capacity; it does not compare historically reconstructed ontologies.
Report length discrepancy for every replicate. Additionally compare each sampled
historical dictionary with its forms given independent capacities (conservative,
more available identities); a signal cannot pass if it fails this comparison.
No larger-than-target astronomy subsampling is needed: there are 31 identities.
Repeat control sampling over 10,000 seeds. Do not discard hard-to-match samples.

## Finite systems, levels and objective

SEARCH_GRID_MANIFEST.json and TRANSFORMATION_REGISTRY.tsv enumerate **64** systems:
article KEEP/DROP_AL × orthography IDENTITY/IJ_UV/ARABIC_LATIN/VELAR_COLLAPSE ×
vowels KEEP/CONTRACT_INTERNAL × abbreviation NONE/SUSPEND_1/PREFIX_4/STRIP_LATIN.
Apply operations in that order after Unicode NFKD, ASCII transliteration, lowercase,
and word segmentation. Concatenate words; DROP_AL removes the whole word `al`,
retaining the original list if removal would empty it. IJ_UV: j→i,v→u.
ARABIC_LATIN: kh→h,gh→g,sh→s,th→t,dh→d then j→i,w→u. VELAR_COLLAPSE adds q,c→k.
CONTRACT_INTERNAL removes a,e,i,o,u,y between first and last characters when length>2.
SUSPEND_1 removes the last character only when length>3. PREFIX_4 keeps first four.
STRIP_LATIN strips the first matching suffix in
`ibus,orum,arum,ium,ius,ae,is,us,um,ii,am,em,as,es,os,i,o,a,e`, only if ≥3 remain.
No additional repetition rule, learned substitution, edit distance or level C.
EVA strings are unchanged. These are orthographic hypotheses, not an inferred
Latin-to-EVA phonetic alphabet. A negative result concerns only this narrow family.

Level A: four normalization-only systems (KEEP article, KEEP vowels, NONE
abbreviation). Level B: remaining systems. Report separately; global selection
across their frozen union is explicitly preregistered and repeated under every null.
Complexity: +1 per nondefault article/orthography/vowel; abbreviation cost 0 for
NONE, 1 SUSPEND_1/STRIP_LATIN, 2 PREFIX_4. One exact equality edge costs 0;
unmatched occurrence costs 1. Each canonical identity capacity is 1 irrespective
of spelling variants. Use optimal maximum-cardinality bipartite matching, lexical
occurrence and identity order to resolve ties. No forced matches. Pair score is 1
for an equality edge; nonedges cannot be assigned. System selection statistic:
`100 * matched_count - complexity`, ties by lowest system ID. Cost=n−matched.

## Validation and seed schedule

JOINT selects using all 57, solely corpus-level evidence. Both page directions
select using the train page ONLY, then hold system fixed on the other page;
identities used by training assignments are unavailable on held-out. Never break
train ties using held-out scores. Hapax is used only for final stratification.
Report all 64 system results on joint and both train pages, plus selected held-out
results. Leave-one-out: all 57 exclusions, repeating full selection and both page
directions. Compare surviving assignments and system IDs. Resampling: 200 seeded
page-stratified 80% subsamples without replacement, repeat selection. No bootstrap
copies of an occurrence. Do not count repeated identical forms as independent
pair evidence. Record label omission and changed systems.

Master seed 570000. Null seed for family index f, replicate r: 570000+100000*f+r,
where f is the family's zero-based position in the frozen manifest and r=0..9999.
Fixed lexical ordering before random sampling; Python random.Random, no global RNG.
Resampling seeds=900000..900199. Synthetic seeds=800000..800019.

## Null controls and parity

Each of nine informative families has **10,000** complete replicate searches:
1. LENGTH_ENDPOINT: keep per-word length and first/last letters, draw interior
   letters from the astronomical empirical character multiset.
2. UNIGRAM: shuffle each word's interior, preserving per-word unigram, endpoints,
   lengths, number of forms and capacities exactly.
3. BIGRAM: randomized directed Euler trails within each word, preserving all
   bigram counts, unigrams, endpoints and lengths exactly. This distribution is
   algorithmic, not uniform over distinct trails. Retain unchanged draws and
   report their fraction; no rejection until a desired null is obtained.
4–6. HISTORICAL_NAMES, MEDIEVAL_LATIN, ARABO_LATIN, sampled as above.
7. OTHER_SECTION_LABEL: readable L-locus tokens outside astronomy (1057 pool).
8. CIRCULAR_TEXT: all 29 same-page circular-text occurrences in pool.
9. INTRO_PROSE: all 68 same-page introductory prose occurrences in pool.

Text controls fill all 57 target slots with nearest-length pool tokens, independently
with replacement, preserving 30/27 split slot sizes. Same-page circular/prose pools
are used for their respective slots. Sampling with replacement is required for
29 circular tokens; replication cannot be read as 57 independent observed controls.
For other-section labels use the whole frozen pool. Keep per-draw source occurrence
IDs in deterministic replay; report number of unique source occurrences and mean
length mismatch. Use all 64 systems, same optimizer, split exclusion, complexity,
tie policy and stopping rules for real and every null dataset.

SHUFFLED_LABEL means permutation of occurrence iteration order without changing
page membership; SHUFFLED_IDENTITIES means a bijective rename of identity groups.
Anonymous matching objectives are provably invariant. For the identity check map
renamed keys back before tie-breaking, preserving the original stable order. Use
an **exact test on the singleton quotient space**, p=1; these are negative-control
invariance diagnostics, not evidence of specificity. They do not enter informative
p-values. Shuffling tokens across pages is a different hypothesis and not claimed.
Historical independent-capacity sensitivity is also searched completely per draw.

## Inference, multiplicity and stopping

Primary statistic is selected complexity-penalized score; held-out statistic is
held-out matched count*100 minus frozen train-selected complexity. For each of
3 endpoints × 9 informative families, p=(1+#null≥observed)/(10000+1), including ties.
Report mean null score, mean coverage, empirical p and Bonferroni min(1,27*p).
BH FDR over those same 27 comparisons is secondary only. The conservative omnibus
p for each endpoint is max of its nine family p-values; corrected p is the max of
its corrected comparisons. Best-null advantage=observed coverage−largest family
mean coverage; additionally report maximum replicate. Historical independent-
capacity results cannot improve the decision; report and require p<0.05 there too.
Invariance diagnostic p-values are not included in these maxima.

CROSS_PAGE_SIGNAL requires corrected p<.05 for both directions, same selected
article/orthography/vowel family both ways, positive advantage, synthetic PASS,
≥90% of 200 subsamples with positive held-out coverage in both directions, and all
LOO retained endpoint coverages ≥80% of original (no one label drives it).
CORPUS_LEVEL_SIGNAL_ONLY requires joint corrected p<.05, positive advantage,
synthetic PASS and LOO stability, but not cross-page transfer.
NO_SIGNAL if complete validated search has neither signal. ENGINE_INCONCLUSIVE
if sensitivity fails, frozen bytes change, null coverage incomplete or reproduction
fails. Incomplete runs are BLOCKED/INCOMPLETE and may not announce a best system.
No optional stopping. Budget 4 wall-clock hours for production/nulls per invocation,
resume allowed without changing seeds/grid/count. Each family checkpoints every
100 replicas atomically; verify hashes and row counts before resuming. A timeout
must leave explicit INCOMPLETE status. Synthetic execution has the same engine
and is outside this production wall budget.

## Synthetic positive controls

20 seeds × noise fractions 0,.10,.25 = 60 synthetic datasets, 57 labels, 30/27.
Use the real 31-identity dictionary and system S008 (IJ_UV normalization); select
one distinct transformed representative per identity, randomizing variant order.
Inject at most 31 known pairs, so at least 26/57 unmatched: this unavoidable
capacity limit must not be concealed. Distribute injected items across both pages;
fill other slots with random strings from the target's alphabet and length profile,
rejecting any string present in ANY dictionary system. Corrupt floor(noise*n)
injected tokens by one random-character substitution (ensure a real change).
Keep noisy tokens even if accidental matches occur. True pair recovery is measured
on uncorrupted pairs and noisy pairs separately. Full model selection at each seed.

Equivalence means the selected system agrees with the planted transformation on
≥90% of uncorrupted injected forms, not necessarily the same system ID. PASS
requires ≥80% true surviving pair recovery joint and ≥60% in each held-out direction,
and equivalence, for ≥18/20 seeds at EACH noise level. Each synthetic dataset has
99 independent LENGTH_ENDPOINT dictionary nulls, all with full model selection;
require joint and both held-out Monte Carlo p≤.05 for ≥18/20 seeds at each noise
level. These are engineering sensitivity checks, not real-data inference; their
replicate count is explicitly smaller. Failure sets ENGINE_SENSITIVITY_INSUFFICIENT
and PRIMARY_CONCLUSION=ENGINE_INCONCLUSIVE even if real scores are zero.

## Pair publication and outputs

For each assigned pair: save every alternative equality identity (rank 1 tied),
next unmatched score 0, local margin 1 only for a sole equality candidate, and
assignment margin obtained by forbidding the edge and recomputing optimum.
An optional global assignment may have margin 0 even with a local unique spelling.
Record all system support, neighboring systems (one grid axis difference), LOO,
subsample survival conditional on inclusion, and both independent held-out page
directions (the same occurrence can only be held out in its own page direction;
require agreement with the opposite direction's train assignment).
Publish CANDIDATE_FORMAL_COMPATIBILITY only with CROSS_PAGE_SIGNAL, ≥3 supporting
systems including ≥2 neighbors, assignment margin≥1, ≥95% LOO and ≥90% subsampling
survival, consistent train/held-out assignment, and no equally good historical
control string in any transformed control pool. Otherwise empty candidate table.
Report hapax/non-hapax and pages after selection; secondary enrichment uses exact
hypergeometric upper tail on assigned label counts and is descriptive only.
Never label a pair TRANSLATION, DECIPHERED or IDENTIFIED_STAR.

Produce all task-mandated TSV/Markdown/JSON files, machine-readable summary,
checkpoint shards, command journal, tests, input and result SHA256. Every result
row carries run_id,seed,dictionary_id,split,system_id,complexity,score,coverage,
null_family,completion_status (NA where not applicable). Keep null draws and all
selected systems, not only aggregate p-values. Preserve upstream bytes.
