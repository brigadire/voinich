# Task61 report

## Scope and estimator

Phase A uses the published Shannon plug-in estimand `h2 = H(X_i | X_{i-1})`,
in bits. The primary raw shared-EVA result is separate from the literal
character representation described in `BOWERN_ENTROPY_METHOD.md`. No
finite-sample correction is applied; order, sample count, context count and
coverage are reported explicitly.

## Results

For the shared-EVA continuous stream, Voynich has h1=4.0058 and h2=2.7726
bits; Doyle has h1=4.1785 and h2=3.5978. Longfellow and Astafiev are in the
same TSV. Thus the low conditional-entropy direction is reproduced relative
to the natural controls, but the exact published magnitude is
representation-dependent and this release is classified **PARTIAL**, not a
claim of exact numerical replication.

The boundary comparison is decisive descriptively: Voynich h2 is 2.7726 in
continuous mode, 2.4458 with an explicit token-boundary symbol, and 2.5026
within tokens only. These are separate estimands. Higher orders fall as
contexts become sparse; h4 coverage is only 0.116 in continuous mode, so no
claim is based on an automatic h5/h6 continuation.

The glyph-shuffle, within-token shuffle, global token shuffle, and within-line
token shuffle outputs are present in `ENTROPY_NULLS.tsv`. They preserve the
specified invariants and are deterministic. The current output is a null
summary rather than a permutation distribution; conclusions requiring a
quantile or p-value are intentionally **NOT_REPORTED**.

## Integration limits

Currier A/B, sections, hands, labels, edit families, and structured positive
controls are explicitly marked `NOT_APPLICABLE` because the required
Task60 metadata/artifacts are not part of the canonical corpus input. Task58
and Task59 values are referenced rather than recomputed. This avoids heuristic
metadata assignment and avoids treating missing data as a negative result.

The homophony rows now use Doyle plaintext for H1 and the shared corrected
position-independent homophony implementation for H2/H4/H8, with fixed seeds
and provenance in `HOMOPHONY_PROVENANCE.tsv`. The transformed opaque-token
files are retained as Task60 provenance references, but are not incorrectly
treated as character streams. Task58/59/60 metrics are joined from their
authoritative artifacts in `HIERARCHICAL_COMPARISON.tsv`.

Task60 labels and the giant edit family are now read from the authoritative
`LABEL_CORPUS.tsv` and `EDIT_DISTANCE_ONE.tsv` artifacts. Their entropy rows
are therefore applicable.

Nothing here implies that low entropy identifies a language, cipher, or
decipherment. It is a property of the stated representation and boundaries.
