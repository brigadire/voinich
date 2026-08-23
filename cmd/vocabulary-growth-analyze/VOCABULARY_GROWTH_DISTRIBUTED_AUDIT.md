# Vocabulary growth distribution audit

| component | natural work unit | complexity | decision |
|---|---|---:|---|
| observed trajectory and frequency counts | ordered token update | O(N) | local single pass |
| positional segments | contiguous segment | O(N) total | local fixed number of segments |
| shuffled null | permutation replicate | O(P·N) | deterministic local ensemble in v1 |

The null permutation index is the natural future distributed work unit. A
reduced-scale profile must first show that permutation work dominates
serialization and transport overhead. Seeds are derived from
`(base_seed, analysis_salt, replicate_index)`, and reduction is ordered by
replicate index, so adding the existing executor later can preserve byte-level
determinism. Task49 deliberately does not fabricate a remote implementation.
