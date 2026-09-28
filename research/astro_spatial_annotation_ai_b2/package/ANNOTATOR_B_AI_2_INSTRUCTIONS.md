# ANNOTATOR_B_AI_2 Instructions

## Goal
Perform an independent visual annotation of astronomical marks, text labels, and
structural elements across 8 canonical astronomical panels of a medieval
illustrated manuscript:

f67r1, f67r2, f67v1, f68r1, f68r2, f68r3, f68v1, f68v2

Each panel is provided as `crops/<panel_id>.jpg`. Panel pixel dimensions are given
in `AI2_INPUT_MANIFEST.json`.

## Blindness constraints
You are operating strictly under blind, clean-room conditions:
- You have NOT been shown any prior annotator's results (human or AI).
- You have NOT been given any transcription, translation, dictionary, frequency,
  or proposed celestial-identity information.
- You have NOT been told how many objects or labels any other annotator found.
- Record only what you can see directly in the crop images. Do not import outside
  knowledge about what medieval astronomical diagrams "usually" contain.

## Two-stage protocol (mandatory order)

### Stage 1 — Detection only
Identify and record:
1. Visual objects (see ontology in `OBJECT_ONTOLOGY_CLEAN.md`), using schema
   `schemas/AI2_OBJECTS_SCHEMA.tsv`.
2. Anonymous physical text labels (`C_LABEL_0001`, `C_LABEL_0002`, ...; do NOT
   transcribe or interpret the text), using schema `schemas/AI2_LABELS_SCHEMA.tsv`.

Do NOT attempt relations, ring/cluster structure, or any geometric interpretation
during Stage 1. Write Stage 1 output to `AI2_OBJECTS.tsv` and `AI2_LABELS.tsv`,
then stop and signal Stage 1 is frozen before proceeding.

### Stage 2 — Relations
Only after Stage 1 is frozen, determine visually observable label <-> object
relations using schema `schemas/AI2_LABEL_OBJECT_RELATIONS_SCHEMA.tsv`. If more
than one relation reading is visually plausible, record multiple candidate rows
for the same label rather than forcing a single answer.

## Confidence discipline
Use `HIGH`, `MEDIUM`, `LOW`, `AMBIGUOUS` honestly. Do not default everything to
`HIGH`. Faint, damaged, overlapping, or ambiguous marks must get `LOW` or
`AMBIGUOUS`.

## Anti-symmetry / anti-completion controls
Do not:
- add objects to make sector counts match or achieve visual symmetry;
- assume every sector has the same number of marks as its neighbors;
- reconstruct damaged or missing marks;
- treat a repeating pattern as proof of an unobserved object;
- copy object positions from a neighboring ring/sector;
- force irregular marks onto idealized circular geometry.

Every recorded object and label must correspond to something you can actually see
in the image. If you are not sure whether a mark is really there, record it with
`LOW`/`AMBIGUOUS` confidence rather than omitting or "completing" it silently —
but never fabricate a mark that is not visually present.
