#!/usr/bin/env python3
"""Reference B: independently structured G1V2 generation primitives V1."""
import math
import unicodedata


def add(total, correction, value):
    candidate = total + value
    if abs(total) < abs(value):
        correction += (value - candidate) + total
    else:
        correction += (total - candidate) + value
    return candidate, correction


def prepare(labels, probabilities, accepted):
    chosen = []
    total = correction = 0.0
    for i in range(len(labels)):
        if labels[i] not in accepted:
            continue
        p = float(probabilities[i])
        if p < 0.0 or not math.isfinite(p):
            return "NUMERICAL_FAILURE", []
        chosen.append([labels[i], p])
        total, correction = add(total, correction, p)
    divisor = total + correction
    if divisor <= 0.0 or not math.isfinite(divisor):
        return "GENERATION_FAILURE", []
    for item in chosen:
        item[1] /= divisor
    return "OK", chosen


def categorical(labels, probabilities, accepted, uniform):
    status, chosen = prepare(labels, probabilities, accepted)
    if status != "OK": return status, None, 0
    total = correction = 0.0
    fallback = None
    for label, p in chosen:
        total, correction = add(total, correction, p)
        if p > 0.0:
            fallback = label
            if uniform < total + correction:
                return "OK", label, 1
    return "OK", fallback, 1


def exponential_race(labels, probabilities, accepted, uniforms):
    status, chosen = prepare(labels, probabilities, accepted)
    if status != "OK": return status, None, 0
    candidates = [(label, p) for label, p in chosen if p > 0.0]
    winner_label = None; winner_score = math.inf; winner_index = -1
    for i in range(len(candidates)):
        u = uniforms[i]
        score = math.inf if u == 0.0 else -math.log(u) / candidates[i][1]
        if winner_label is None or score < winner_score or (score == winner_score and i < winner_index):
            winner_label, winner_score, winner_index = candidates[i][0], score, i
    return "OK", winner_label, len(candidates)


def alias_table(labels, probabilities, accepted):
    status, chosen = prepare(labels, probabilities, accepted)
    if status != "OK": return status, [], [], []
    names = [x[0] for x in chosen]; count = len(names)
    scaled = [x[1] * count for x in chosen]
    cutoffs = [0.0 for _ in names]; redirects = list(range(count))
    below = [i for i in range(count) if scaled[i] < 1.0]
    above = [i for i in range(count) if scaled[i] >= 1.0]
    while len(below) and len(above):
        below.sort(reverse=True); above.sort(reverse=True)
        low = below.pop(); high = above.pop()
        cutoffs[low] = scaled[low]; redirects[low] = high
        scaled[high] = (scaled[high] + scaled[low]) - 1.0
        (below if scaled[high] < 1.0 else above).append(high)
    for position in sorted(below + above):
        cutoffs[position] = 1.0; redirects[position] = position
    return "OK", names, cutoffs, redirects


def alias_sample(labels, probabilities, accepted, first, second):
    status, names, cutoffs, redirects = alias_table(labels, probabilities, accepted)
    if status != "OK": return status, None, 0
    slot = math.floor(first * len(names))
    answer = slot if second < cutoffs[slot] else redirects[slot]
    return "OK", names[answer], 2


def serialize(tokens):
    output = bytearray()
    for raw in tokens:
        value = raw
        if unicodedata.normalize("NFC", value) != value:
            raise ValueError("token is not NFC")
        if len(value) == 0 or len(value) > 64 or "\n" in value or "\r" in value or value in ("<BOS>", "<EOS>", "<UNK>"):
            raise ValueError("invalid token")
        output.extend(value.encode("utf-8")); output.append(10)
    return bytes(output)


def stream_token(labels, probabilities, uniforms, max_glyphs=64):
    pieces = []; cursor = 0
    for _ in range(max_glyphs):
        accepted = {x for x in labels if x != "<BOS>" and (pieces or x != "<EOS>")}
        status, result, consumed = categorical(labels, probabilities, accepted, uniforms[cursor])
        if status != "OK": return status, "", cursor
        cursor += consumed
        if result == "<EOS>": return "OK", "".join(pieces), cursor
        pieces.append("\ufffd" if result == "<UNK>" else result)
    return "OK", "".join(pieces), cursor
