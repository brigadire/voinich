# Frozen E3 profile protocol

The profile set is frozen before the corrected real search: the 64 rows in `E3_OPERATION_PROFILES.tsv`. Each profile applies its registry fields globally and in this order: word normalization, article policy, orthography policy, vowel policy, abbreviation policy, then DROP_UNMAPPED alignment. Each listed operation is applied at most once at its defined stage; no per-term, per-LABEL, page-specific, or hidden deletion is allowed. Profiles are processed independently and are not mixed.
