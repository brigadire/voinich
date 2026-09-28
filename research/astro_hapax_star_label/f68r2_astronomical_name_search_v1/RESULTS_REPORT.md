# f68r2 standalone search status

The correct human scope contains 24 frozen f68r2 relation groups, while the earlier transcription scope contains 27 f68r2 LABEL occurrences. The difference is preserved in `F68R2_GROUP_SCOPE.tsv` and `LABEL_OCCURRENCE_CROSSWALK.tsv`.

The full f68r2 search was not started because none of the 24 human groups has a unique frozen bridge to an EVA occurrence. The migration package marks the legacy labels `AMBIGUOUS_CROSSWALK`, and the two human-added labels are `HUMAN_ADDED_NOT_IN_LEGACY`. Assigning one of the 27 transcription tokens to these groups would be an unverified provenance choice.

This is an implementation gate failure, not a negative astronomical result. The package intentionally contains no scorer graph or search incumbent until the group-to-occurrence crosswalk is frozen.
