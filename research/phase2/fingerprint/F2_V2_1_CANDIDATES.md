# Fingerprint v2.1 candidate quarantine

Source: `research/phase2/notation-audit/F2_ADMISSION.tsv` (Task79b, frozen
before Task79c began). None of these candidates reached `ADMIT_TO_F2`
before the Task79c design freeze
(`research/phase2/fingerprint/TASK79C_DESIGN_FROZEN`), so per the parent
task §4/§35 none of them enters the Task79c confirmatory comparison or any
Task81–83 run. Positional-channel findings in particular do not become
post-hoc primary F2 evidence, regardless of anything Task79c or a later
Voynich/model comparison finds.

## Quarantined for v2.1 consideration

| Candidate | External motivation | Definition/estimator | Why quarantined, not admitted |
|---|---|---|---|
| Positional channel NMI | MATLACH2022; PARISEL2026 | Adjacent normalized MI for a fixed positional channel | `DEFER_TO_V2_1` — single canonical transcription only at admission time, no independent notation control, no stability battery. (Task79c's Gate A second transcription and Gate B/C controls, once complete, could in principle retire the "single canonical transcription" and "no independent notation control" blockers for a future v2.1 admission review — but that review is out of Task79c's scope; Task79c does not retroactively promote this metric.) |
| Boundary class/Zipf diagnostic | PARISEL2026 | Boundary-class transition distribution and Zipf-law fit | `DEFER_TO_V2_1` — preprint-only motivation, no released reference implementation to validate against, not independently replicated. |
| Abbreviation length reduction | CAPPELLI1912; ABBREVIATIONES | Aligned character/token length ratio between abbreviated and expanded forms | `EXPLORATORY_ONLY` — not observable on Voynich itself (no plaintext/expansion alignment exists or is claimed for Voynich); usable only on paired historical-abbreviation data such as the Gate B control acquired for Task79c, never as direct Voynich evidence. |
| Expansion ambiguity (H(expansion\|abbreviation)) | CAPPELLI1912; ABBREVIATIONES | Conditional entropy of expansion given abbreviation | `EXPLORATORY_ONLY` — same reason: requires an abbreviation/expansion pairing Voynich does not have. The Gate B control (Burchards Dekret Digital) does carry real `<abbr>/<expan>` pairs and could compute this as a *control-only* diagnostic in a future v2.1 exploratory note, but that computation is not part of Task79c's confirmatory F2 comparison and was not run here. |

## Excluded, not forwarded

| Candidate | Reason |
|---|---|
| Abbreviation edit-family productivity | `REJECT_REDUNDANT` at Task79b admission — the existing LP/EF (lexical-paradigm / edit-family) dimensions already measure the same diagnostic on the existing C-GRAMMAR/C-LEN/C-FREQ controls. Not carried forward as a v2.1 candidate; there is nothing new for a future version to add here. |

## Binding constraints on any future v2.1 admission review

1. A quarantined candidate may only be promoted by a review that itself
   predates seeing that candidate's Voynich or model-comparison result,
   exactly as Task79b/Task79c's own gates required for the current
   registry.
2. No quarantined candidate may be computed on Voynich and then admitted
   because it happened to discriminate well — that is precisely the
   post-hoc target-fitting `FINGERPRINT_V2_DISTANCE.md` §5 rule 1
   prohibits.
3. This list is exhaustive of everything Task79b flagged as of the Task79c
   design freeze; Task79c did not search for additional new
   Voynich-specific diagnostics (prohibited by the parent task's own
   scope statement).
