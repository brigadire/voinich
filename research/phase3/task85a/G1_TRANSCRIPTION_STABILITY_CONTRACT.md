# Cross-transcription stability contract

For an effect `e` relative to its baseline, direction is `sign(e)` with exact
zero as zero. Cross-transcription relative discrepancy is
`abs(e_ZL3b-e_IT2a)/max(abs(e_ZL3b),abs(e_IT2a),1e-12)`.

- `TRANSCRIPTION_STABLE`: directions agree and discrepancy is at most 0.20.
- `DIRECTION_STABLE`: nonzero directions agree and discrepancy exceeds 0.20.
- `TRANSCRIPTION_SENSITIVE`: directions differ, either value is nonfinite, or
  exactly one value is zero.

For predictive metrics the effect is signed improvement over the required
baseline; for a structural metric it is `MFC_threshold-absolute_distance`, and
the family effect is the median member effect. Apply the rule separately to
every required predictive metric and each G1 structural family. A gate
requiring “at least direction-stable” rejects
`TRANSCRIPTION_SENSITIVE`. Candidate selection is class-stable only when the
two transcription-specific minimality procedures choose the same M0–M5 class.
