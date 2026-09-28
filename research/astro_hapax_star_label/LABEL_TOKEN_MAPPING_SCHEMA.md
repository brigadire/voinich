# LABEL/token mapping schema

`LABEL_TOKEN_CANDIDATES.tsv` contains one row per physical LABEL and is not a verified mapping. `proposed_*` fields are deliberately `NONE`: no frozen 2D-to-transcription alignment exists. `alternative_line_refs` is the exhaustive same-page candidate line set. Internal group and provenance fields are retained outside the blinded UI for later cohort construction.

Human export fields:

- `review_id`, `label_id`, `panel`, `candidate_snapshot_sha256`: immutable identifiers;
- `review_action`: `CONFIRMED_TOKEN`, `CONFIRMED_SEQUENCE`, `CORRECTED_MAPPING`, `AMBIGUOUS`, `UNREADABLE`, or `NO_MATCH`;
- `mapping_outcome`: `SINGLE_TOKEN`, `MULTI_TOKEN`, `RING_SEQUENCE`, `AMBIGUOUS`, `UNREADABLE`, `NO_TRANSCRIPTION_MATCH`, or `OUT_OF_TRANSCRIPTION_SCOPE`;
- `selected_occurrence_ids`: semicolon-separated absolute occurrence IDs in intended order;
- `reading_order`: `CORPUS_ORDER`, `CLOCKWISE`, `COUNTERCLOCKWISE`, `VISUAL_LEFT_TO_RIGHT`, or `UNDETERMINED`;
- confidence, reviewer, ISO-8601 timestamp, ambiguity and notes;
- `completion_status=COMPLETE` only after an explicit outcome. Blank fields never mean NO_MATCH.

For token-bearing outcomes every selected occurrence must exist on the LABEL panel. SINGLE_TOKEN requires one occurrence; MULTI_TOKEN and RING_SEQUENCE require at least two. Ambiguous/unreadable/no-match/out-of-scope outcomes cannot carry asserted token forms. The importer derives exact raw/normalized forms from occurrence IDs; reviewers cannot type a novel token into the frozen mapping.
