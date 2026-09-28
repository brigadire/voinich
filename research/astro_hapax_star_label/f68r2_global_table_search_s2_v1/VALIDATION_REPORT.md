# Validation report

The S2 model was run with `/home/brigadire/.venv/bin/python3` and OR-Tools 9.15.6755. It used the unchanged 25,517 f68r2 paths, exact `DROP_UNMAPPED` scorer parity, `NONE` abbreviation, injective global mappings, one identity per occurrence assignment, at most k active mappings, and support of at least two distinct EVA token types per active rule.

All eight k runs returned `OPTIMAL`. The k=8 result has five assignments, six active mappings, distinct identities, and support PASS for every rule. f68r1 and group crosswalk data were excluded.
