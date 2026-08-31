# Visual-context / external-memory experiment report

## Executive result

**VISUAL_TEXT_ASSOCIATION_DETECTED at evidence Level B, with substantial confounding and no Level-C test.** Visually defined IVTFF sections have distinguishable textual profiles. The result is present in existing frozen Task65/Task79 evidence and in a supplemental aggregation of frozen Task79 line/occurrence outputs. After simultaneous adjustment for Currier A/B, dominant scribal hand, quire, page order, token count and line count, four of ten fixed scalar metrics retain an FDR-adjusted section increment. This is evidence for textual association with broad visual/functional organization, not evidence that the manuscript is an external-memory system.

The decisive limitation is unchanged: there are no independent page-specific visual descriptors. Consequently the within-section own-image versus shuffled-image test was not run and Level C cannot be evaluated.

## Data and reuse boundary

The broad taxonomy is the source-defined IVTFF `$I` field: Astronomical, Biological, Cosmological, Herbal, Pharmaceutical, Stars, Text and Zodiac. The complete 227-page registry is frozen in `VISUAL_CONTEXT_TAXONOMY.md` and `VISUAL_CONTEXT_TAXONOMY.tsv`. Two pages (`f65r`, three tokens/one line; `f116v`, two tokens/one line) remain in descriptive outputs but fail the prespecified inferential floor of 20 tokens and two lines, leaving 225 pages.

No Stage 1-28 analyzer was rerun. `research/visual-context-analyze` reads the frozen Task79 `line_profiles.json` and `occurrence_metadata.jsonl`, joins the source IVTFF page headers, and aggregates their sufficient summaries. Correct Currier A/B is taken from source `$L`; Task79's field named `currier_language` reflects `$C` and was not silently used as A/B. A mixed-hand page (`f115r`) is represented by its token-weighted dominant hand with the mixture retained in the label.

The generated page dataset has ten fixed scalars in three families:

- token: mean length, type/token ratio, token entropy;
- sequence: exact adjacency repetition, near-edit adjacency repetition, mean line transition entropy, mean line token entropy;
- line/block: mean line length, line-length CV, first/final length asymmetry.

Context/persistence evidence is reused directly from Task65 rather than approximated with a new page metric. No cross-family composite score is used for the primary section-distance tests.

## Existing evidence reused

Task65's preselected Task64 page profile already directly tests the simplest prediction. Its `METADATA_EFFECTS.tsv` has between-section distance greater than within-section distance for Astronomical, Biological, Cosmological, Herbal, Pharmaceutical, Stars and Zodiac; Text is the exception. Its section variance shares overlap those for Currier and hand, and its local-effect/conditioned-decay outputs show that page-local organization exists without assigning it to imagery.

Task79 independently reports:

| Frozen metric | Value | Permutation p | BH q | Meaning |
|---|---:|---:|---:|---|
| HR1_SECTION_VARIANCE_SHARE | 0.125259 | 0.000999 | 0.001306 | section share of line-length variance |
| LC5_IVTFF_I_NMI | 0.008452 | 0.000999 | 0.001306 | `$I` association with token-length/regime class |
| HR6_CURRIER_SECTION_NMI | 0.648621 | 0.000999 | 0.001306 | registry name is misleading: operational field is `$C`, not Currier A/B `$L`; general metadata-confounding warning only |
| PF2_FOLIO_COHERENCE | 0.173040 | 0.000999 | 0.001306 | folio coherence, not a visual effect by itself |

Task83b's later `cs4/family-section` and `cs5/local-adjacency-x-regime` are `NOT_APPLICABLE` because IVTFF metadata was not bound in that execution. They are missing evidence, not contradictory evidence.

## Supplemental section-level comparison

For each scalar separately, every pair of eligible pages on different physical leaves contributes an absolute difference. The statistic is mean between-section difference minus mean within-section difference. The null permutes complete label vectors among same-shaped physical-leaf groups; 1,000 repetitions use seed 20260830. The ten fixed tests use Benjamini-Hochberg correction.

Seven of ten metrics have larger between-section variation and `q=0.001427`: mean token length (ratio 1.35), type/token ratio (1.92), token entropy (2.30), line transition entropy (2.02), mean line length (1.66), line-length CV (5.58), and mean line token entropy (1.60). Exact repetition, near-edit repetition and boundary-length asymmetry do not follow the prediction. Thus the signal is multi-family but explicitly non-unanimous.

Section-specific means, dispersion, medians and ranges are in `VISUAL_CONTEXT_SECTION_SUMMARY.tsv`; all page values are in `VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv`.

## Confounders and incremental section information

The contingency table confirms strong concentration:

- all Biological pages are Currier B and hand 2, all in quire M;
- all Astronomical and Zodiac pages have missing Currier A/B in the source subset and use hands 4;
- Pharmaceutical pages are Currier A/hand 1 and occur only in quires O/S;
- Stars is overwhelmingly Currier B/hand 3 and concentrated in quire T;
- Herbal supplies most cross-Currier and cross-quire overlap.

The full design remains numerically estimable: reduced rank 28, full rank 35, so section adds seven independent columns despite this concentration. For each metric, the reduced model is `Currier + scribe + quire + log(tokens) + lines + page position`; the full model adds section. Freedman-Lane residual permutations move complete physical-leaf blocks only within the same quire, followed by BH correction across the ten fixed metrics.

Four metrics retain adjusted incremental section information:

| Family / metric | Reduced R² | Full R² | Incremental section R² | p | BH q |
|---|---:|---:|---:|---:|---:|
| token / type-token ratio | 0.7829 | 0.7947 | 0.0545 | 0.00699 | 0.01748 |
| token / token entropy | 0.9373 | 0.9414 | 0.0654 | 0.00200 | 0.00999 |
| sequence / near-edit repetition | 0.4871 | 0.6352 | 0.2888 | 0.000999 | 0.00999 |
| line/block / line-length CV | 0.8620 | 0.8979 | 0.2601 | 0.00300 | 0.00999 |

Mean token length and mean transition entropy are borderline after correction (`q=0.05495` and `0.05195`); the other four metrics are not supported. The two diversity metrics remain size-sensitive even after explicit size adjustment, and the larger increments occur in near-edit and line-variation metrics. Because overlap is sparse and model-based residual exchangeability is an assumption, this establishes Level B with limitations, not an unconfounded causal visual effect.

## Secondary classification diagnostic

A fixed nearest-centroid model over all ten standardized scalars, evaluated leave-one-quire-out with physical leaves kept together, obtains accuracy 0.6756 versus a fold-specific majority baseline of 0.5689. Pages whose class is absent from a training fold count as errors. A 1,000-repetition label test permuting physical-leaf label vectors only within quire gives `p=0.000999`. This is a secondary signal diagnostic; it neither selects metrics nor upgrades the evidence level.

## Page-specific visual test

Not run. Repository metadata contains no frozen image-only descriptors such as object/plant/figure/vessel counts, circular-component counts, image area or illustration topology. `$I` is a broad page class, not a page-specific descriptor. Creating such variables now without a separately frozen and blinded annotation protocol would violate the non-circularity constraint. Therefore `H_VISUAL_LOCAL` remains untested rather than rejected.

## Answers to required questions

**Q1: Do visually defined VM sections have distinguishable textual fingerprints?** Yes. Existing Task65/79 results and seven of ten supplemental scalars show a systematic section signal.

**Q2: Is the section effect larger than within-section variation?** Yes in the distributional sense for seven of ten supplemental scalars and seven of eight Task65 section rows; no for exact repetition, near-edit repetition and boundary asymmetry, and no for Task65 Text considered alone.

**Q3: Can the effect be explained by Currier, scribe, quire, corpus size, or other known metadata?** A large share can: reduced-model R² ranges from 0.156 to 0.937 and the contingency table shows severe concentration. It cannot explain all measured section-associated variation under the specified model, but residual confounding remains plausible.

**Q4: Does visual-section identity add information beyond those confounders?** Yes, narrowly: four fixed metrics in token, sequence and line/block families survive within-quire, physical-leaf-block permutations and BH correction. This is predictive/explanatory association, not causality.

**Q5: Within a section, is page-specific textual structure associated with page-specific visual features?** Unknown. No eligible visual-feature data exist, so the key Level-C permutation test is not computable.

**Q6: Which result level was reached?** Level B, with identifiability limitations; not Level C.

**Q7: What does this imply—and not imply—for external memory?** The result is compatible with text being organized together with broad visual/functional sections and weakly supports that prerequisite. Ordinary thematic organization, production sequence, layout, scribal practice or unmeasured material factors can also produce it. It does not demonstrate page-specific text-image coupling, memory operations, semantics, decipherment or that Voynich is an external-memory system.

## Reproducibility and artifacts

- audit: `VISUAL_CONTEXT_EXISTING_RESEARCH_AUDIT.md`, `VISUAL_CONTEXT_EXISTING_EVIDENCE.tsv`;
- frozen taxonomy: `VISUAL_CONTEXT_TAXONOMY.md`, `VISUAL_CONTEXT_TAXONOMY.tsv`;
- page data: `VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv`;
- section results: `VISUAL_CONTEXT_SECTION_SUMMARY.tsv`, `VISUAL_CONTEXT_SECTION_COMPARISONS.tsv`;
- controls: `VISUAL_CONTEXT_CONTINGENCY.tsv`, `VISUAL_CONTEXT_CONFOUNDER_ANALYSIS.tsv`;
- secondary diagnostic: `VISUAL_CONTEXT_CLASSIFICATION.tsv`;
- hashes and parameters: `VISUAL_CONTEXT_RESULTS_MANIFEST.json`;
- implementation: `research/visual-context-analyze`.

## Final status

```text
EXISTING_RESEARCH_AUDITED=true
VISUAL_TAXONOMY_FROZEN=true
SECTION_LEVEL_TEST_COMPLETED=true
CONFOUNDER_CONTROL_COMPLETED=true
PAGE_SPECIFIC_VISUAL_TEST_COMPLETED=false

VISUAL_CONTEXT_EVIDENCE_LEVEL=B

EXTERNAL_MEMORY_HYPOTHESIS_STATUS=
WEAKLY_SUPPORTED_BY_THIS_EVIDENCE
```
