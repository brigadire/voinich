# Higher-order sequential dependence validation

Confirmatory test of whether the first token of a frozen n>=3 replicated sequence A B C carries information about the third token C beyond what the second token B alone predicts, i.e. P(C|A,B) vs P(C|B). Corpus SHA256: `d44a2dabd08cf969d23a2ec4e91f719eb9da85f49f966184f31aa40f745dead9`. No new bigrams, trigrams or n-grams were discovered; every candidate below is read programmatically from the previous audit.

## Frozen inventory


## Summary table

| sequence | family | occurrences | eligible blocks | joint classes | P(C\|B) pooled | P(C\|A,B) pooled | enrichment | conditional p | conditional q | CMI (bits) | LOBO M2 advantage | sign consistency | jackknife | status |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|

## Per-candidate discussion


## Interpretation guardrails

A HIGHER_ORDER_REPLICATED status means the sequence exhibits replicated higher-order conditional dependence, not that " is a rule". This audit performs no new sequence discovery, tests only the frozen candidates listed above, and establishes nothing about natural language, grammar, operator/operand structure, or decipherment.
