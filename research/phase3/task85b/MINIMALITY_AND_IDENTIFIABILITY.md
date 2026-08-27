# Minimality and identifiability

Candidate adequacy requires FITTED, PREDICTIVE_PASS, GENERATION_SUCCESS, STRUCTURAL_PASS, complete hashes, and no protocol veto. A valid predictive or structural fail yields MODEL_INADEQUATE. Induction caps, convergence failure, generation failure, unavailable metrics, or missing evidence yield MODEL_NOT_IDENTIFIABLE; none is evidence that the represented class is inadequate.

Adequate candidates are compared by `model bits + heldout PM1` only under the same representation. The development-frozen equivalence bound forms an undirected observational-equivalence graph. The four outcomes are:

- `UNIQUE_MINIMUM`: one lowest-rank adequate component and every lower class was validly rejected;
- `EQUIVALENT_SET`: multiple candidates/classes occupy the lowest equivalence component;
- `ORDER_ONLY`: valid comparisons establish ordering but cannot isolate the minimum;
- `NOT_IDENTIFIABLE`: missing/computational/protocol evidence or an unresolved comparison can change the minimum.

`G1_MINIMAL_CLASS=NONE` is issued only when every tested class has at least one class-representative, completed candidate and all such candidates are validly inadequate under the frozen coverage rule. If no candidate is adequate but any potentially decisive route is not assessable, output NOT_IDENTIFIABLE. A single M0-M5 label is never selected merely by rank tie-breaking. The normative algorithm is `G1V2_COMPLEXITY_CONTRACT.tsv`.
