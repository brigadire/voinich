# M2 rule inventory

Rules map recurrent source units to recurrent EVA units by structural role. A rule is admissible only when support is at least 2 independent concepts; singleton rules are forbidden in the primary profile. Metrics use concept-level support, not spelling variants. Rule score uses `MODEL_DESCRIPTION_LENGTH + UNEXPLAINED_DATA_COST`; repeated rules reduce description length. Context-free order-preserving composition is the baseline; positional/context-sensitive substitutions and post-hoc astronomical segmentation are excluded.
