# Global mapping-size semantics

`MAX_MAPPINGS_PER_INDIVIDUAL_PATH=5` is the frozen generator contract. `GLOBAL_MAPPING_TABLE_SIZE_LIMIT=NONE`: the global table is the union of all selected path mappings. A selected assignment may therefore require more than five mappings. The solver contains no global k=5 constraint.
