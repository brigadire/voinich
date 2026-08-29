#!/usr/bin/env python3
"""Independent path 1 for the proposed constrained categorical rule."""

import math


def neumaier(values):
    total = 0.0
    correction = 0.0
    for value in values:
        provisional = total + value
        if abs(total) >= abs(value):
            correction += (total - provisional) + value
        else:
            correction += (value - provisional) + total
        total = provisional
    return total + correction


def sample(outcomes, probabilities, admissible, u53):
    """Filter in frozen row order, normalize once, and consume one U53."""
    kept = [(x, float(p)) for x, p in zip(outcomes, probabilities) if x in admissible]
    if any(not math.isfinite(p) or p < 0.0 for _, p in kept):
        return {"status": "NUMERICAL_FAILURE", "draws": 0}
    mass = neumaier([p for _, p in kept])
    if mass == 0.0:
        return {"status": "GENERATION_FAILURE", "draws": 0}
    normalized = [(x, p / mass) for x, p in kept]
    cumulative_values = []
    for end in range(1, len(normalized) + 1):
        cumulative_values.append(neumaier([p for _, p in normalized[:end]]))
    for (outcome, probability), boundary in zip(normalized, cumulative_values):
        if probability > 0.0 and u53 < boundary:
            return {"status": "OK", "outcome": outcome, "draws": 1}
    positive = [x for x, p in normalized if p > 0.0]
    return {"status": "OK", "outcome": positive[-1], "draws": 1}

