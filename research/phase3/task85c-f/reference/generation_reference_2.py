#!/usr/bin/env python3
"""Independent path 2 for the proposed constrained categorical rule."""

import math


def compensated_add(accumulator, compensation, value):
    new_accumulator = accumulator + value
    if abs(accumulator) < abs(value):
        compensation += (value - new_accumulator) + accumulator
    else:
        compensation += (accumulator - new_accumulator) + value
    return new_accumulator, compensation


def sample(outcomes, probabilities, admissible, u53):
    labels = []
    weights = []
    s = c = 0.0
    for index in range(len(outcomes)):
        if outcomes[index] not in admissible:
            continue
        weight = float(probabilities[index])
        if weight < 0.0 or not math.isfinite(weight):
            return {"status": "NUMERICAL_FAILURE", "draws": 0}
        labels.append(outcomes[index])
        weights.append(weight)
        s, c = compensated_add(s, c, weight)
    mass = s + c
    if mass == 0.0:
        return {"status": "GENERATION_FAILURE", "draws": 0}
    running = correction = 0.0
    last_positive = None
    for label, weight in zip(labels, weights):
        normalized = weight / mass
        running, correction = compensated_add(running, correction, normalized)
        if normalized > 0.0:
            last_positive = label
            if u53 < running + correction:
                return {"status": "OK", "outcome": label, "draws": 1}
    return {"status": "OK", "outcome": last_positive, "draws": 1}

