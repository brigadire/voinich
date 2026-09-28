# Independent annotation protocol

## Primary passes

Annotators receive only the eight panel crops, panel IDs, crop dimensions, this
ontology, and an empty copy of the seed schema. They must not receive
transcriptions, translations, Stolfi records, ZL3b data, lexical frequencies,
hapax status, M1/M2/M3 outputs, dictionary candidates, or brute-force results.
Every visible object is boxed independently. Irregular objects may additionally
receive a polygon. Labels are boxed as physical text units and initially receive
anonymous IDs.

The frozen panel image origin is top-left. Images may be zoomed but must not be
rotated, rescaled, enhanced, or cropped without recording the affine transform.
The original Yale bytes and crop coordinates in the source registry remain the
canonical frame.

## Candidate relations

Each annotator may record every plausible label/object relation. No forced
single choice is allowed. `NEAREST_OBJECT` is calculated geometrically and does
not identify what the object means. Confidence must be one of `HIGH`, `MEDIUM`,
`LOW`, or `AMBIGUOUS`, with visible evidence in the relation row.

## Matching and agreement

Passes A and B are retained unchanged. Candidate detections are paired for
agreement only after both passes are frozen, using class equality and maximum
IoU (with centre distance as a diagnostic). Report detection precision/recall,
IoU, normalized centre distance and class agreement. Relations use exact-object
and Jaccard candidate-set agreement; Cohen's kappa is reported only when its
assumptions hold. Ring assignments use exact agreement. Cyclic orders use
Kendall tau in both directions and retain an orientation-ambiguity flag.

At least all STAR labels and 30% of other objects must be double annotated.
Adjudication creates new `ADJUDICATED` rows and never overwrites either pass.
Only after adjudication may an independent reviewer connect anonymous physical
labels to Stolfi and then to frozen ZL3b occurrences using image location and
transcription evidence. Token spelling alone is insufficient.

## Release gates

Production structural use additionally requires STAR label spatial coverage of
at least 0.80, high-or-medium relation coverage of at least 0.70, completed
double annotation, and no unresolved provenance or coordinate-transform issue.
Failure is reported as failure; missing records are never imputed to meet a gate.
