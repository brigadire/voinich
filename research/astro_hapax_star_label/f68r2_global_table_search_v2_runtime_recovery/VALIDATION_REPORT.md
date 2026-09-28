# Validation report

The run used only the frozen f68r2 graph, frozen occurrence scope, lexicon, and exact scorer semantics. No group crosswalk or f68r1 path was used. Deduplication found zero collapsible duplicate paths.

The model enforces one path per occurrence, one canonical identity globally, one injective global mapping, at most five active mappings, active-rule support from at least two distinct EVA token types, and unmatched occurrences.

The recovered backend passed import, synthetic API, warm-start, support, identity-capacity, scorer-parity, and deterministic single-worker configuration checks. CP-SAT reported `OPTIMAL` with objective 3 and best bound 3.
