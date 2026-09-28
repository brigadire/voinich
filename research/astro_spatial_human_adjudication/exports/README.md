# Human exports

This directory receives immutable, timestamped imports from reviewed CVAT exports. Completed canonical STAR/LABEL snapshots are stored at package root. Never overwrite a completed review export.

H-High preserves both direct imports (`*_AS_EXPORTED.tsv`) and derived final files (`*_FINAL.tsv`). Final files replace repeated candidates with frozen calibration decisions, apply only the explicit corrections recorded in `HUMAN_HIGH_REVIEW_OVERRIDES_R01.tsv`, and restore original geometry for new non-`MODIFY` decisions. Raw CVAT ZIP exports are never altered.

Consensus-QC follows the same audit pattern: `HUMAN_*_CONSENSUS_QC_R01_AS_EXPORTED.tsv` preserves direct imports with the reviewer-wide confidence override, while `*_FINAL.tsv` validates the exact pending queue and restores frozen geometry for non-`MODIFY` decisions.

Medium exports contain only `MODIFY` decisions and are imported directly as `HUMAN_*_MEDIUM_R01.tsv` with the reviewer-wide `HIGH` override. `scripts/freeze_reviewed_layers.py` merges all completed phases into the canonical root snapshots, rejects conflicting duplicate records, and lists unreviewed LOW-priority candidates separately in `HUMAN_PRODUCTION_UNREVIEWED.tsv`. Completed snapshots may be regenerated only byte-identically; corrections require a new revision.
