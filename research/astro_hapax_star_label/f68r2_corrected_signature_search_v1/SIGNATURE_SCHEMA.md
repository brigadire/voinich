# Semantic signature schema

A signature is the exact tuple:

`(label_id, eva_token, identity, required_mappings, forbidden_source_units)`.

`required_mappings` is the complete sorted path mapping set. `forbidden_source_units` is derived from the normalized source stream after the frozen preprocessing contract and records source units that must remain unmapped. Paths are merged only when every tuple component is identical. `PATH_SIGNATURE_MEMBERS.tsv` preserves all original path provenance.
