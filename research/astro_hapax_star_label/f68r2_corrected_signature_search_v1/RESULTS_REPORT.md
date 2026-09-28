# Results report

Signature profiling separated the v2 failure: input loading and variable creation are negligible; support-constraint construction is the dominant model-build phase, followed by CP-SAT search. Full semantic compression is weak: 25,517 paths reduce to 25,469 signatures.

The compact exact existence query `coverage >=4` is `INFEASIBLE` after a 44.5-second search, with 2,050,081 constraints and an independently checked empty assignment. Since coverage 1, 2, and 3 were independently certified infeasible in prior packages, the corrected frozen path space has maximum coverage 0.

This is a result for the frozen 25,517-path graph and its corrected global scorer semantics. E3 and null remain disabled.
