# Output schema

All result tables are TSV and every raw result row carries `run_id`, `solver_sha256`, `scorer_sha256`, `config_sha256`, `lexicon_sha256`, `target_scope_sha256`, `ortools_version`, `seed`, `timestamp_utc`, `stage`, and `integrity_sha256`.

`RULE_SYSTEMS.tsv`: system_id, stage, profile, rank, score, bound, gap, status, rule_count, unmatched_count, cooptimal_group_id.
`RULE_DEFINITIONS.tsv`: system_id, rule_id, input_grapheme, output_grapheme, support_count, support_label_ids.
`LABEL_NAME_CANDIDATES.tsv`: system_id, label_id, candidate_name, match_class, support, score_contribution, capacity_status.
`UNMATCHED_LABELS.tsv`: system_id, label_id, reason.
`COOPTIMAL_SYSTEMS.tsv`: cooptimal_group_id, system_id, score, canonical_symmetry_key.
`SOLUTION_FAMILIES.tsv`: family_id, cooptimal_group_id, symmetry_basis, member_count.
`SCORE_DECOMPOSITION.tsv`: system_id, exact, near, weak, unmatched, complexity_penalty, total.
`SEARCH_BOUNDS.tsv`: run_id, stage, lower_bound, upper_bound, gap, proof_status.
`RESOURCE_USAGE.tsv`: run_id, stage, wall_seconds, cpu_seconds, peak_rss_kb, variables, constraints, edges, checkpoint.
