#!/usr/bin/env python3
"""CONTRACT_REFERENCE_ONLY: Decimal/rational M0 reference implementation A."""
from __future__ import annotations

from decimal import Decimal, getcontext

getcontext().prec = 80


def fit(tokens: list[str], alpha: Decimal):
    ordinary = sorted({g for token in tokens for g in token}, key=lambda x: x.encode("utf-8"))
    outcomes = ordinary + ["<UNK>", "<EOS>"]
    counts = {x: Decimal(0) for x in outcomes}
    for token in tokens:
        for glyph in token:
            counts[glyph] += 1
        counts["<EOS>"] += 1
    denominator = sum(counts.values()) + alpha * len(outcomes)
    if denominator <= 0:
        raise ValueError("nonpositive denominator")
    probabilities = {x: (counts[x] + alpha) / denominator for x in outcomes}
    return outcomes, counts, denominator, probabilities


def direct_cdf(outcomes, weights, allowed, u: Decimal):
    local = [(x, weights[x]) for x in outcomes if x in allowed and weights[x] > 0]
    total = sum((w for _, w in local), Decimal(0))
    if total <= 0:
        raise ValueError("zero admissible mass")
    cumulative = Decimal(0)
    for outcome, weight in local:
        cumulative += weight / total
        if u < cumulative:
            return outcome
    return local[-1][0]


if __name__ == "__main__":
    o, c, d, p = fit(["ab", "a"], Decimal(1))
    assert o == ["a", "b", "<UNK>", "<EOS>"] and d == 9
    assert abs(sum(p.values()) - Decimal(1)) <= Decimal("1e-70")
    print("PASS CONTRACT_REFERENCE_ONLY A")
