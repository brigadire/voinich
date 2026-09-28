# Astronomical spatial quality report

## Verdict

`ANNOTATION_LAYER_STATUS=FIRST_PASS_NOT_PRODUCTION_ELIGIBLE`.

The official Yale facsimile source and all eight crops are frozen, and the derived geometry is reproducible. The primary object pass is intentionally conservative. It contains the 29 separately visible f68r1 star marks plus large structural objects on every panel. It does **not** claim exhaustive star detection on the remaining panels. Anonymous f68r1 label envelopes are approximate and low-confidence; no token was assigned from lexicographic Stolfi order.

```text
STAR_OBJECTS_DETECTED=29
STAR_LABELS_EXPECTED=67
STAR_LABELS_SPATIALLY_MAPPED=29
STAR_LABEL_SPATIAL_COVERAGE=0.432836
HIGH_CONFIDENCE_LABEL_OBJECT_RELATIONS=0
AMBIGUOUS_RELATIONS=0
UNMAPPED_LABELS=38
HIGH_OR_MEDIUM_RELATION_COVERAGE=0.000000
DOUBLE_ANNOTATED_OBJECTS=0
```

Both acceptance thresholds fail (`0.80` spatial coverage and `0.70` high/medium relation coverage). This is a quality result, not a reason to inflate annotation. Independent ANNOTATOR_B and adjudication remain mandatory. Exact object boxes are approximate stroke envelopes at `MEDIUM`; anonymous label boxes and relations are `LOW`.

## Provenance and transformations

Full-resolution Yale IIIF JPEG byte hashes were measured on 2026-09-02. Crops are lossless coordinate definitions against those bytes; generated crop image bytes are not inputs or redistributed. Coordinates use each crop's raw pixel frame. The Yale manifest snapshot is inherited from the independently frozen visual-source review.

## Contamination controls

Primary object geometry was recorded from images. Stolfi/ZL3b data enter only after geometry, in bridge generation. No translation, dictionary candidate, M1/M2/M3 assignment, hapax flag, section frequency or brute-force result is used. Because the executing annotator had repository access, this pass cannot itself certify full experimental blindness; the second pass must be performed from a clean image-only package.
