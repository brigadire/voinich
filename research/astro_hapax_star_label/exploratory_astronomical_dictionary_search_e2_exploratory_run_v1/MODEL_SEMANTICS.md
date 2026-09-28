# Model semantics

An active rule is an injective source-to-EVA mapping. A label can be selected only when at least one lexicon identity form encodes exactly to its EVA token under `DROP_UNMAPPED`. Each label and identity can occur at most once under `GLOBAL_CAPACITY_1`. Unmatched labels are allowed.

Every active rule must occur in the selected final assignment with at least two distinct EVA token types. Repeated occurrences of one token type do not satisfy support. Multiple attestations of an identity remain one identity capacity unit.

D1 changes only the computational decomposition. It does not add local errors, digraphs, abbreviations, composition, or real-data-dependent pruning.
