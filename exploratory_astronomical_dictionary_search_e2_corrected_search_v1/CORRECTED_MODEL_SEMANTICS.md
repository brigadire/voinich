# Corrected model semantics

The corrected search uses an injective global substitution table with at most five rules, `DROP_UNMAPPED`, `GLOBAL_CAPACITY_1`, and the existing BALANCED scoring contract.

For each fixed table, the oracle first computes the exact deterministic maximum one-to-one matching between 57 label occurrences and canonical lexicon identities. It then computes support from the final assignment. A rule is considered used only when at least one selected correspondence depends on that source-to-EVA mapping. Used rules require at least two distinct EVA token types. An unused rule has no support violation and does not invalidate the table.

This separates table size from used-rule support and avoids requiring support from unused rule slots. The search is exploratory and does not certify a global optimum.
