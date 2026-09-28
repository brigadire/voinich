# AI_ALIGNMENT_B blind candidate ranking (shuffled order)

Read only files in this directory. Do not inspect any parent directory, repository file, sealed answer, prior mapping, or other agent output. This is an instruction-enforced clean room.

For every row in CASES.tsv, inspect its crop and the complete same-panel candidate list in CANDIDATES.tsv. Rank by glyph-by-glyph visual alignment, never by display position, uniqueness, semantics, or expected yield. You do not have Visual A/C. Preserve uncertainty and choose NO_MATCH when no candidate adequately accounts for the image. ZL3b/EVA display conventions: C=cth, K=ckh, P=cph, F=cfh, N=iin, A=ain, H=ch, S=sh, E=ee, I=in.

Write one compact JSON object per line, in CASES.tsv order, with exactly these keys: neutral_label_id, outcome, ranked_candidate_ids, glyph_alignment, unmatched_visible_glyphs, unmatched_transcription_glyphs, boundary_agreement, direction_start_assessment, top1_confidence, top2_margin, alternatives, rationale. outcome is RANKED_MATCH or NO_MATCH. ranked_candidate_ids and alternatives are JSON arrays with at most three candidate IDs. top1_confidence is HIGH, MEDIUM, or LOW; top2_margin is LARGE, SMALL, TIE, or NOT_APPLICABLE. Never force a match.
