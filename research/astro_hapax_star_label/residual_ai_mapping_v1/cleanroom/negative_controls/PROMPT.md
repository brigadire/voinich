# Blind negative-control alignment test

Read only files in this directory. Do not inspect any parent directory, repository file, sealed answer, prior mapping, or other agent output. For every CASES.tsv row inspect the crop and only that control_id's candidate rows. The sets are deliberately adverse. Test whether visual evidence supports any candidate; do not assume a match exists. Candidate order is random. Do not use meaning.

Write exactly one compact JSON object per CASES row, in order, with keys: control_id, neutral_label_id, outcome, selected_candidate_id, confidence, abstained, rationale. outcome is RANKED_MATCH, AMBIGUOUS, or NO_MATCH. selected_candidate_id is an empty string on abstention. confidence is HIGH/MEDIUM/LOW. abstained is true for AMBIGUOUS and NO_MATCH. Never force a candidate.
