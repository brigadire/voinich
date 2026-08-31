# Confounder identifiability diagnosis

The metadata-only nested designs M0–M4 are numerically estimable on 227 joined pages, but adding the visual increment to the complete visual-vector subset leaves only 24 observations concentrated in Biological (19) and Zodiac (5). The visual categories are therefore separated from several section/quire strata and do not provide enough within-stratum replication for a saturated incremental model. This is a structural `INSUFFICIENT_N + DESCRIPTOR_CATEGORY_SEPARATION` failure, not a reason to remove a confounder.

M0–M4 rank and condition diagnostics are in `LEVEL_C_CONFOUNDER_RANK_AUDIT.tsv`. For v2, use restricted exchangeability blocks and label any remaining association `PAGE_SPECIFIC_ASSOCIATION_WITH_RESIDUAL_CONFOUNDING`; do not force the saturated model.
