# Blind pilot disagreement audit

The two initial clean-room passes were hashed and locked before comparison. The reviewer used only those passes, the image-only schema/protocol, and Yale images. The contaminated draft and all textual research outputs were excluded.

## Fixed-gate outcome before revision

Fifteen of 23 descriptors passed immediately. Eight entered the single permitted revision:

| Descriptor | Agreement | Relevant reliability | Diagnosed cause | Controlled action |
|---|---:|---:|---|---|
| `COMP_ILLUSTRATED_AREA` | 0.875 | weighted κ 0.794 | adjacent area thresholds | merge middle/high boundary into a two-level nonzero scale |
| `COMP_ILLUSTRATION_REGIONS` | 0.833 | weighted κ 0.795 | `2-3` versus `4+` boundary | merge as `MULTIPLE` |
| `COMP_DOMINANT_LAYOUT` | 0.833 | α 0.791 | object-centric versus diagram-centric center | merge as `CENTRAL_VISUAL` |
| `OBJ_STAR_LIKE` | 0.875 | weighted κ 0.751 | flower/star boundary and dense counts | coarsen to absence/presence |
| `GEO_REPEATED_MODULES` | 0.875 | weighted κ 0.692 | amount boundary plus prevalence | coarsen to absence/presence |
| `TXT_SURROUNDS_IMAGE` | 0.870 | α 0.726 | three-side rule sensitive to open boundaries | remove |
| `TXT_EMBEDDED_IMAGE` | 0.913 | α 0.779 | open form versus closed boundary | require writing-zone center inside a visibly closed boundary |
| `VIS_MOTIF_REPETITION` | 0.875 | weighted κ 0.692 | redundant with repeated modules | remove |

The principal systematic cases were: sparse marks on otherwise writing-dominant pages; whether a large single form or multiple subforms constituted the layout; decorative ray/flower forms; and open plant-like silhouettes that do not define a true inside. A small number of page-specific disagreements were classified as genuine visual ambiguity and remain eligible for `UNCERTAIN`; no exception rule was created.

## Revision constraints

No descriptor was added. No textual value, section statistic, or predicted association was available to the reviewer. The entire 24-unit pilot is rerun by two fresh zero-history clean-room annotators under 1.0.0-rc; reliability is not recalculated on disagreement rows alone. This consumes the sole revision cycle.
