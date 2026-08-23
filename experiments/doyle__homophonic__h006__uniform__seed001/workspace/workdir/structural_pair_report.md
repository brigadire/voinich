# Structural pair decomposition

Structural similarity is reproduced unchanged from the existing pair dataset. All statements below are formal corpus descriptions; no token meaning is inferred. Context similarities and differences use full distributions, while tables are display-limited. Entropy uses natural logarithms and effective vocabulary is `exp(entropy)`.

## Negative controls

Controls match unordered log-counts, normalized graphemic distance, and reliability, while favoring structural similarity near the full-corpus median. They are decomposed with exactly the target metrics.

| Target | Control | Structural | Reliability | Distance | Match cost |
|---|---|---:|---:|---:|---:|

## Family decomposition

A family is a connected component; only listed edges define direct structural-distant links. Complete matrices, including non-edge pairs, are in `family_decomposition.yaml`.

## Limits

Observed absence is not proof of a prohibition. Context observations at line boundaries have no neighbor and therefore context totals can be below token counts. Pair rows are statistically dependent because tokens recur across pairs. Control matching is descriptive and does not make pairs independent.
