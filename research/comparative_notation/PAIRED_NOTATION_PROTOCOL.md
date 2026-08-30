# Paired notation delta

For aligned `(plain, encoded)` or `(expanded, diplomatic)` representations,
run both independently, join exact metric/version/regime IDs, and calculate
`Delta = encoded − plain`. Missing or partially comparable values stay missing.
Alignment coverage, insertions/deletions, and boundary preservation are written
to the normalization report. The executable core is
`notation.NotationDelta`; it does not compare either side with VM and does not
select a representation.
