# Scoring contract

For a fixed table, the primary objective is the cardinality of the selected one-to-one assignment. Reports additionally retain raw occurrence coverage, distinct EVA token coverage, distinct identity coverage, unmatched labels, assignment hash, and rule support audit.

The matching oracle and independent scorer use the same `encode_word` semantics and canonical identity capacity. A callback or checkpoint is valid only when objective equals the captured assignment length, every selected edge re-encodes, identity and label uniqueness hold, and every active rule has distinct-EVA support.
