# Validation report

The source audit confirms 24 f68r2 human relation groups and 27 f68r2 target LABEL occurrences. The 24 group rows are complete and frozen at high human confidence. The corresponding transcription migration is not usable for this experiment: all 24 group label IDs lack `UNIQUE_PROVENANCE_CROSSWALK` in the current migration snapshot.

Because the target token for each human group is unresolved, exact scorer-consistent paths cannot be assigned to the 24-group denominator. Any optimization would therefore mix human relation scope with an unvalidated token selection. The search is marked `NOT_EVALUATED`; no timeout is claimed and no scientific conclusion is drawn.
