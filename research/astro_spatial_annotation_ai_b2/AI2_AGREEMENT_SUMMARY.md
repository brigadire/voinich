# ANNOTATOR_B_AI_2 Three-Way Agreement Summary

## 1. Executive Summary & Verification Metrics

```text
ANNOTATOR_B_AI_2_PASS=COMPLETE
ANNOTATOR_B_AI_2_INDEPENDENT_FROM_A=YES
ANNOTATOR_B_AI_2_INDEPENDENT_FROM_AI1=YES

AI1_STAR_OBJECTS=293
AI2_STAR_OBJECTS=330
AI1_AI2_MATCHED_STAR_OBJECTS=171
AI1_ONLY_STAR_CANDIDATES=122
AI2_ONLY_STAR_CANDIDATES=159
AI_CONSENSUS_STAR_RATE=0.3783
MEDIAN_STAR_IOU=0.108877
MEDIAN_STAR_CENTER_DISTANCE_PX=63.0714

AI1_LABELS=175
AI2_LABELS=206
AI1_AI2_MATCHED_LABELS=92
LABEL_CONSENSUS_RATE=0.3183

AI1_AI2_STAR_RELATION_AGREEMENT=43/102 (42.2%)
AI1_AI2_RELATION_EXACT_AGREEMENT=51/151 (33.8%)
AI1_AI2_RELATION_MEAN_CANDIDATE_JACCARD=0.2926

THREE_WAY_STAR_MATCH_F68R1=24
THREE_WAY_LABEL_MATCH_F68R1=26

AI_CONSENSUS_STAR_OBJECTS=171
AI_CONSENSUS_LABELS=92

HUMAN_ADJUDICATION_HIGH_PRIORITY_COUNT=529
HUMAN_ADJUDICATION_MEDIUM_PRIORITY_COUNT=448
HUMAN_ADJUDICATION_LOW_PRIORITY_COUNT=9

AI_CONSENSUS_USEFUL_FOR_HUMAN_REVIEW=LIMITED
HUMAN_ANNOTATOR_STILL_REQUIRED=YES
PRODUCTION_SPATIAL_GATE=READY_FOR_HUMAN_ADJUDICATION
```

## 2. STAR_OBJECT Consensus by Panel

| Panel | A_STARS | AI1_STARS | AI2_STARS | AI1_AI2_MATCHED | AI1_ONLY | AI2_ONLY | CONSENSUS_RATE | MEDIAN_IOU | MEDIAN_CENTER_DIST_NORM |
|---|---|---|---|---|---|---|---|---|---|
| `f67r1` | 0 | 27 | 21 | 13 | 14 | 8 | 0.371 | 0.032 | 0.0302 |
| `f67r2` | 0 | 0 | 0 | 0 | 0 | 0 | 0.000 | 0.000 | 0.0000 |
| `f67v1` | 0 | 40 | 42 | 19 | 21 | 23 | 0.302 | 0.111 | 0.0260 |
| `f68r1` | 29 | 29 | 29 | 25 | 4 | 4 | 0.758 | 0.218 | 0.0084 |
| `f68r2` | 0 | 47 | 59 | 43 | 4 | 16 | 0.683 | 0.205 | 0.0163 |
| `f68r3` | 0 | 71 | 77 | 40 | 31 | 37 | 0.370 | 0.075 | 0.0249 |
| `f68v1` | 0 | 43 | 61 | 21 | 22 | 40 | 0.253 | 0.033 | 0.0232 |
| `f68v2` | 0 | 36 | 41 | 10 | 26 | 31 | 0.149 | 0.010 | 0.0287 |

`f68r1` is the one panel where `ANNOTATOR_A` provides a nearly complete star
reference; the other seven panels had an incomplete `ANNOTATOR_A` pass, so
`AI1_AI2_ONLY` support there must not be read as "AI disagrees with A" — A
simply never annotated those panels.

## 3. f68r1 Critical Sanity Check (Part P)

```text
F68R1_A_STARS=29
F68R1_AI1_STARS=29
F68R1_AI2_STARS=29

F68R1_A_AI1_MATCH=26
F68R1_A_AI2_MATCH=28
F68R1_AI1_AI2_MATCH=25
F68R1_THREE_WAY_MATCH=24

F68R1_A_LABELS=29
F68R1_AI1_LABELS=31
F68R1_AI2_LABELS=37
F68R1_A_AI1_LABEL_MATCH=27
F68R1_A_AI2_LABEL_MATCH=25
F68R1_AI1_AI2_LABEL_MATCH=29
F68R1_THREE_WAY_LABEL_MATCH=26
```

## 4. Disagreement Taxonomy Counts

- `AI1_AI2_BBOX_DISAGREEMENT`: 233
- `AI1_ONLY_LABEL`: 83
- `AI1_ONLY_OBJECT`: 201
- `AI2_ONLY_LABEL`: 114
- `AI2_ONLY_OBJECT`: 295
- `A_CONFLICT_WITH_AI_CONSENSUS`: 1
- `CONFIDENCE_DISAGREEMENT`: 9
- `RELATION_DISAGREEMENT`: 50

## 5. Human Adjudication Triage

- **HIGH** (529): AI1/AI2 disagree on STAR_OBJECT existence, relation disagreements,
  AI1_AI2_ONLY labels, and any `A_CONFLICT_WITH_AI_CONSENSUS` case.
- **MEDIUM** (448): bounding-box disagreements on matched objects, non-star
  AI1-only/AI2-only objects.
- **LOW** (9): confidence-only mismatches on otherwise agreeing pairs.

## 6. Interpretation Boundary

The headline result of this pass is that AI1<->AI2 agreement is **partial, not high**
(`AI_CONSENSUS_STAR_RATE=0.378`, `LABEL_CONSENSUS_RATE=0.318`)
outside `f68r1`, the one panel with a real `ANNOTATOR_A` reference (where
AI1<->AI2 star/label matching is far higher — 25/29 stars,
29/37 labels — than the corpus-wide rate).
On the seven panels A never really annotated, AI1 and AI2 found substantially
different star counts and positions (see the per-panel table above): AI1 was
generated as a single earlier procedural pass over these crops, biased toward
regular sector/ring symmetry, while AI2 was a fresh, independently launched
agent instance (different model family, clean-room, two-stage frozen protocol)
that explicitly rejected symmetry completion and reported partial/asymmetric
counts. This divergence is itself informative: it shows that, for this panel
set, single-AI-pass star/label counts are annotator-sensitive off the one
richly-referenced panel, and should not be treated as a converged candidate
layer without human adjudication. The `AI1 ∩ AI2` intersection
(171 stars, 92 labels) is still a materially
stronger candidate set than either raw pass alone — every entry in it was
independently proposed twice — but it is **not** a substitute for human
ground truth (`HUMAN_ANNOTATOR_STILL_REQUIRED=YES`), and its low corpus-wide
coverage means many real objects present in only one pass are excluded from
it. This independence was operationalized as: a fresh, context-free agent
instance from a different model family, given only the clean-room package,
with no visibility into `ANNOTATOR_A`, `ANNOTATOR_B_AI`, or any counts from
either (see `AI2_INPUT_MANIFEST.json` / `AI2_MANIFEST.json`).
Where the two AI passes disagree, or where `ANNOTATOR_A` conflicts with an
AI1/AI2 consensus object, human adjudication is still required.
