# Description length and normalization

See the implementation contract for exact six-component formulas. All arithmetic
uses bits; literal data use UTF-8 byte lengths. No-rule matching selects no positive
edges and reproduces baseline exactly. Unsupported models have `objective=null`
in training; their raw description length remains visible for debugging. Heldout
retains the entire rule cost, including rules not used in that partition.

Assignment costs and literal gap costs affect the exact matching; the model penalty
is constant within that subproblem. Across models, minimizing total description
length is equivalent to maximizing normalized gain on a fixed dataset. No unsupported
rule can enter the admissible beam. Adding unused rules increases code length and
may increase reference width, never improving an equivalent model's compression.
One whole-token rule per pair is both unsupported and more costly than literals.

Scores for 29/57/68 examples are dimensionless, not claimed numerically identical:
fixed model overhead and assignment code length vary with sample size. The analytic
fixture reports raw components and verifies gain from their sum. Exact row duplication
has no effect because normalization and support operate on deduplicated surface bags.
`EMPTY` and `SMALL` are explicit; negative heldout gain remains negative.
