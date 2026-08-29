#!/usr/bin/env python3
"""Reference A: direct implementation of G1V2 generation primitives V1."""
import math
import unicodedata


def nsum(values):
    s = c = 0.0
    for x in values:
        t = s + x
        c += (s - t) + x if abs(s) >= abs(x) else (x - t) + s
        s = t
    return s + c


def conditioned(outcomes, weights, allowed):
    row = [(x, float(w)) for x, w in zip(outcomes, weights) if x in allowed]
    if any(w < 0 or not math.isfinite(w) for _, w in row):
        return "NUMERICAL_FAILURE", []
    z = nsum([w for _, w in row])
    if z <= 0 or not math.isfinite(z):
        return "GENERATION_FAILURE", []
    return "OK", [(x, w / z) for x, w in row]


def categorical(outcomes, weights, allowed, u):
    status, row = conditioned(outcomes, weights, allowed)
    if status != "OK":
        return status, None, 0
    cumulative = []
    for i in range(len(row)):
        cumulative.append(nsum([w for _, w in row[:i + 1]]))
    for (x, w), bound in zip(row, cumulative):
        if w > 0 and u < bound:
            return "OK", x, 1
    return "OK", [x for x, w in row if w > 0][-1], 1


def exponential_race(outcomes, weights, allowed, uniforms):
    status, row = conditioned(outcomes, weights, allowed)
    if status != "OK":
        return status, None, 0
    positive = [(x, w) for x, w in row if w > 0]
    if len(uniforms) < len(positive):
        raise ValueError("insufficient uniforms")
    best = None
    for index, (x, w) in enumerate(positive):
        u = uniforms[index]
        score = math.inf if u == 0.0 else -math.log(u) / w
        key = (score, index)
        if best is None or key < best[0]:
            best = (key, x)
    return "OK", best[1], len(positive)


def alias_table(outcomes, weights, allowed):
    status, row = conditioned(outcomes, weights, allowed)
    if status != "OK":
        return status, [], [], []
    labels = [x for x, _ in row]
    n = len(row)
    q = [w * n for _, w in row]
    probability = [0.0] * n
    alias = list(range(n))
    small = sorted(i for i, v in enumerate(q) if v < 1.0)
    large = sorted(i for i, v in enumerate(q) if v >= 1.0)
    while small and large:
        s = small.pop(0); l = large.pop(0)
        probability[s] = q[s]; alias[s] = l
        q[l] = (q[l] + q[s]) - 1.0
        (small if q[l] < 1.0 else large).append(l)
        small.sort(); large.sort()
    for i in sorted(small + large):
        probability[i] = 1.0; alias[i] = i
    return "OK", labels, probability, alias


def alias_sample(outcomes, weights, allowed, u_column, u_threshold):
    status, labels, probability, alias = alias_table(outcomes, weights, allowed)
    if status != "OK":
        return status, None, 0
    column = int(u_column * len(labels))
    selected = column if u_threshold < probability[column] else alias[column]
    return "OK", labels[selected], 2


def serialize(tokens):
    normalized = []
    for token in tokens:
        if unicodedata.normalize("NFC", token) != token:
            raise ValueError("token is not NFC")
        if not token or "\n" in token or "\r" in token or token in {"<BOS>", "<EOS>", "<UNK>"}:
            raise ValueError("invalid token")
        if len(token) > 64:
            raise ValueError("token too long")
        normalized.append(token)
    return b"" if not normalized else ("\n".join(normalized) + "\n").encode("utf-8")


def stream_token(outcomes, weights, uniforms, max_glyphs=64):
    emitted = []; used = 0
    while len(emitted) < max_glyphs:
        allowed = set(outcomes) - {"<BOS>"}
        if not emitted:
            allowed.discard("<EOS>")
        status, selected, draws = categorical(outcomes, weights, allowed, uniforms[used])
        if status != "OK":
            return status, "", used
        used += draws
        if selected == "<EOS>":
            return "OK", "".join(emitted), used
        emitted.append("�" if selected == "<UNK>" else selected)
    return "OK", "".join(emitted), used
