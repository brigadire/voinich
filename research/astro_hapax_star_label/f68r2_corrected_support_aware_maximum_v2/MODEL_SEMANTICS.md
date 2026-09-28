# Compact path-selection model

Each path has required_mappings and forbidden_source_units. Selecting a path activates all required mappings and keeps all forbidden source units unmapped. Global mappings are functional and injective. At most one path is selected for each LABEL, EVA token type, and canonical identity.

Support is evaluated on the selected assignment: every active mapping rule must be used by at least two distinct selected EVA token types. Every accepted assignment is independently replayed with DROP_UNMAPPED.
