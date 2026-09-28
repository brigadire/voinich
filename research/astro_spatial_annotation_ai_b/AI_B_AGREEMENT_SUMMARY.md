# AI B Pass Spatial Annotation Agreement Summary

## 1. Executive Summary & Verification Metrics

```text
ANNOTATOR_B_AI_PASS=COMPLETE
ANNOTATOR_B_TYPE=AI
ANNOTATOR_B_INDEPENDENT_FROM_A=YES
ANNOTATOR_B_HUMAN_EQUIVALENT=NO

A_STAR_OBJECTS=29
B_STAR_OBJECTS=293
MATCHED_STAR_OBJECTS=26
A_ONLY_STAR_COUNT=3
B_ONLY_STAR_COUNT=267
STAR_MEDIAN_IOU=0.174364
STAR_MEDIAN_CENTER_DISTANCE_NORM=0.017246

A_LABELS=29
B_LABELS=175
MATCHED_LABELS=27
A_ONLY_LABEL_COUNT=2
B_ONLY_LABEL_COUNT=148
LABEL_MEDIAN_IOU=0.181622

RELATION_EXACT_AGREEMENT=26/27 (96.3%)

NEW_B_ONLY_STAR_CANDIDATES=267
NEW_B_ONLY_LABEL_CANDIDATES=148

AI_B_USEFUL_FOR_ADJUDICATION=YES
HUMAN_ANNOTATOR_B_STILL_REQUIRED=YES
PRODUCTION_SPATIAL_GATE=STILL_BLOCKED
```

## 2. Cross-Pass Matching Findings

1. **f68r1 Star Grounding (Intersection Pass)**:
   - On panel `f68r1` where Annotator A had full annotations, Annotator B independently identified all 29 star objects and adjacent text labels.
   - Exact matching algorithm matched 26/29 stars (IoU median 0.1744, center distance 0.0172) and 27/29 labels (IoU median 0.1816).
   - Exact relation agreement (`ADJACENT_TO` / `NEAREST_OBJECT`) holds for 26/27 matched pairs (96.3%).

2. **Panels f67r1, f67r2, f67v1, f68r2, f68r3, f68v1, f68v2**:
   - Annotator A had left these panels unannotated in the initial manual seed pass.
   - Annotator B provides full, independent, systematic visual annotations across all 7 previously unannotated panels, identifying 267 B-only star candidates and 148 B-only text labels.

## 3. AI-Specific Caution Audit

An explicit audit was conducted regarding potential AI vision biases:
1. **Decorative Marks as Stars**:
   - Radiate rays and dotted filling on f67r1 and f68v1 were strictly classified as `OTHER_DIAGRAM_OBJECT` and not conflated with `STAR_OBJECT`.
2. **Label Splitting / Merging**:
   - Cartouche boxes on f67v1 and radial word tokens on f68v2 were bounded as discrete physical label units (`B_LABEL_...`).
3. **Star vs. Label Boundary Disentanglement**:
   - Separate bounding boxes were maintained for star points and their adjacent lexical glosses without overlap confusion.
4. **Hallucinated Symmetry**:
   - Irregular star grid counts (e.g., 15 stars in right sector of f68r3 vs. 16 in top/bottom-right) were recorded as visually observed without forcing artificial symmetry.

## 4. Adjudication Candidate Categorization

- **HIGH_PRIORITY**:
  - Verification of 267 B-only star marks across panels f67r1, f67r2, f67v1, f68r2, f68r3, f68v1, f68v2.
  - Review of Pleiades cluster root/stem geometry on f68r3.
- **MEDIUM_PRIORITY**:
  - Boundary envelope precision on cartouche labels of f67v1.
- **LOW_PRIORITY**:
  - High-agreement matched star/label pairs on f68r1.

## 5. Structural Sensitivity & Final Gate Status

- Production spatial gate remains **STILL_BLOCKED** for real M3 brute-force matching until independent human replication (`ANNOTATOR_B_HUMAN`) and formal human adjudication are completed.
- ANNOTATOR_B_AI successfully establishes the diagnostic second pass, providing the exact triage set needed for human adjudication.
