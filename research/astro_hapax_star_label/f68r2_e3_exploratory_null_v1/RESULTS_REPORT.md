# Post-hoc exploratory null and five-match audit

The real-data audit is complete. The five E3 matches all use lexicon rows marked historically attested with high confidence and frozen provenance. The winning profile is S043 (`DROP_AL`, `IJ_UV`, `KEEP`, `STRIP_LATIN`); all three non-default operations are active in the profile definition, while the selected paths show that `DROP_AL` has no effect on these five forms and `IJ_UV` is operationally inert for the displayed forms. `STRIP_LATIN` is active for three of the five paths and returns `none` for two.

Nine certified maximum-5 profiles were compared as a profile-family stability diagnostic. The detailed LABEL, pair, and rule frequencies are in `PROFILE_FAMILY_STABILITY.tsv`. This is a comparison across co-optimal profile families, not a complete enumeration of every co-optimal assignment within each profile.

The planned post-hoc null used a global character shuffle preserving the 27 token lengths and repeated full selection across all 64 profiles. It was stopped before the first replicate completed because graph generation exceeded the available runtime window. Consequently `NULL_GE_5` and `NULL_GE_6` are not estimated, and no rarity or astronomical interpretation is claimed.

The result remains: exact real-data maximum 5/27; difference from random is unknown.
