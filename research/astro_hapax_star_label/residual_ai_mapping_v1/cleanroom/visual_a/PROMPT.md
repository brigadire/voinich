# AI_VISUAL_A blind visual transcription

Read only files in this directory. Do not inspect any parent directory, repository file, sealed answer, prior mapping, or other agent output. This is an instruction-enforced clean room.

For every row in CASES.tsv, inspect its crop and page context. Describe visible glyph form only; do not infer meaning or use external Voynich knowledge. You have no candidate list. Preserve uncertainty and abstain when needed. ZL3b/EVA display conventions: C=cth, K=ckh, P=cph, F=cfh, N=iin, A=ain, H=ch, S=sh, E=ee, I=in.

Write one compact JSON object per line, in CASES.tsv order, with exactly these keys: neutral_label_id, outcome, visual_glyph_sequence, token_boundaries, direction, start_point, uncertain_glyph_positions, alternative_visual_readings, confidence, rationale. Allowed outcomes: SINGLE_TOKEN_READING, MULTI_TOKEN_READING, RING_CYCLIC_READING, AMBIGUOUS_READING, UNREADABLE. Confidence: HIGH, MEDIUM, LOW. Use JSON arrays for uncertain_glyph_positions and alternative_visual_readings. Never force a reading.
