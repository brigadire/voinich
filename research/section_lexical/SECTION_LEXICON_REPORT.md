# Section-specific lexical report

## Decision

`SECTION_LEXICAL_MODEL=L2_DISTRIBUTIONAL_SPECIALIZATION`  
`SECTION_EXCLUSIVE_VOCABULARY=LIMITED`  
`SECTION_ENRICHED_VOCABULARY=STRONG`

Q1–Q2: Many exclusive tokens are hapax/low-frequency tail; repeated multi-leaf exclusives are limited after fixed support thresholds. Raw unique counts are size-dependent and are not interpreted directly.  
Q3: Smoothed enrichment identifies section-biased tokens, with BH correction retained for every tested token×section.  
Q4: Page-preserving rarefaction is included for all three fixed seeds; specialization remains primarily distributional rather than a large stable exclusive lexicon.  
Q5: Bag-of-tokens classification is retained as a diagnostic, but the available frozen quire split is not identifiable in this implementation; no accuracy claim is promoted.  
Q6: Currier, hand, quire, and position overlap section labels, so residual confounding remains.  
Q7: The stable vocabulary is largely shared; strongest differences are concentrated in frequency distributions and rare-token tail.  
Q8: Deterministic physical-leaf split-half replication is reported without redefining candidates on the replication half.  
Q9: The supported picture is shared general vocabulary plus a smaller section-biased layer.  
Q10: Best supported model is L2, not L1: distributional specialization without strong word-like exclusive terminology.  
Q11: No website/source-of-truth structure artifact is updated; the result is recorded as a research observation because confounder and classifier limitations prevent a publication-grade structural claim.

This does not establish token semantics, decipherment, or that images are unrelated to text.

```text
SECTION_LEXICON_INPUTS_FROZEN=true
SECTION_LEXICON_PROTOCOL_FROZEN=true
SECTION_LEXICON_PRODUCTION_RUN_EXECUTED=true
SECTION_LEXICON_PRODUCTION_RUN_VALID=true
SECTION_LEXICON_REPRODUCIBLE=true
STRUCTURE_MODEL_UPDATED=false
STRUCTURE_PUBLICATION_UPDATE_PREPARED=false
```
