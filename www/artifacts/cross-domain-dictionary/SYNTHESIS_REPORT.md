# Historical-dictionary brute-force comparison

This publication records the final synthesis of the f68r2 cross-domain dictionary branch. It is a report of a bounded experiment, not a decipherment or a thematic identification.

## Two separate results

### Positive structural result: hapax enrichment

The independently matched astronomical `LABEL` inventory contains 112 confirmed occurrences, of which 80 are section-local hapax: `0.714285714`. The panel-conditioned null mean is `0.611207143`; enrichment ratio `1.168647524`; one-sided permutation `p=0.008799120`. All 8/8 leave-one-panel-out exclusions retained positive direction and `p<0.05`.

This supports only the tested statement that independently identified astronomical labels are enriched for section-local hapax. It does not imply that every hapax is a label, and it does not provide semantics or a decipherment.

### Negative result: dictionary search within the tested class

The fresh matched f68r2 search used the corrected global mapping model, 27 frozen LABEL occurrences, 64 frozen E3 profiles, equal budgets and model-selection-aware nulls. Exact certified maxima were:

| panel | maximum |
|---|---:|
| astronomical | 0/27 |
| botanical | 0/27 |
| historical non-botanical control | 3/27 |

All 99 pseudo-controls were valid and were evaluated over all 64 profiles. Botanical and astronomical empirical `p=1.00`; Holm-adjusted values remained `1.00`. The registered interpretation is `RESULT_COMPATIBLE_WITH_GENERIC_FORM_MATCHING`, with `DOMAIN_SPECIFIC_SIGNAL=NONE`.

## Withdrawn or superseded results

- The earlier astronomy `3/27` is not transferable to the matched experiment because it used a different dictionary size/design.
- Caesar-shift controls are withdrawn as the primary null and replaced by non-isomorphic pseudo-lexicons.
- Earlier unverified botanical candidates and provisional historical controls are not evidence.
- The form-derived `134 forms = 134 identities` construction is superseded by source-level identity remediation.

## Limits

The negative result applies only to this corrected-global form-matching model class, panel size, frozen profiles and controls. It does not disprove botanical, astronomical or any other page theme. The hapax result is a token-frequency result under a specified label inventory and null, not a semantic classification.

The search package did not fully enumerate all co-optimal assignments; saved assignments are profile witnesses. The published evidence therefore does not support interpreting particular identity pairs as recovered meanings.

## Closure

The current global-substitution branch is closed without a thematic verdict. Any future work must use a materially different, independently justified model family, such as historically grounded abbreviation, mnemonic/notational encoding, or morpheme-compositional construction. See `NEXT_MODEL_FAMILIES.md`.

The search null distribution was: 72 pseudo-lexicons with maximum 0, 11 with maximum 2, and 16 with maximum 3. The historical control had `p=0.17`; it is retained as a comparison, not interpreted as a domain-specific signal.
