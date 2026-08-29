# Task86C-v2 production-freeze report — V1.2.1/I2/E3

This was a new clean run. The authority chain and the exact frozen Task85c-h implementation root passed identity checks. The implementation itself failed mandatory pre-materialization scientific conformance:

- PF-IMPL-01: `GenerateSynthetic` ignores normative route parameters. A scientifically different M1 parameter object produces the same corpus bytes/hash.
- PF-IMPL-02: `F2Metrics` gives EF3 scientific weight 0 although the registry marks it weighted, and gives HR1/SKELETON weight 1 although SKELETON is diagnostic-only with weight 0.

These defects can change generated controls, gates and final verdicts. Task86C is forbidden to repair frozen scientific handlers, so the run stops with `IMPLEMENTATION_VALIDATION_FAILURE`. DEVELOPMENT calibration, escrow, 144 blind controls, 36 natural controls, production JobIDs, the 1,321,152/2,617,152 DAG, capacity qualification and production executable were not materialized. The firewall remains intact.
