#!/usr/bin/env python3
"""CONTRACT_REFERENCE_ONLY: run 32,768 deterministic M0 properties."""
from decimal import Decimal
import hashlib
import json
from m0_reference_impl_a import fit

h = hashlib.sha256()
for i in range(32768):
    vocab = 1 + i % 8
    count = 1 + (i // 8) % 12
    alpha = [Decimal(0), Decimal("0.1"), Decimal("0.5"), Decimal(1)][i % 4]
    tokens = ["".join(chr(97 + (i+j+k) % vocab) for k in range(1+(i+j)%7)) for j in range(count)]
    outcomes, counts, denominator, probabilities = fit(tokens, alpha)
    assert outcomes[-2:] == ["<UNK>", "<EOS>"]
    assert counts["<UNK>"] == 0 and counts["<EOS>"] == count
    assert abs(sum(probabilities.values()) - Decimal(1)) <= Decimal("1e-70")
    assert all(x >= 0 for x in probabilities.values())
    if alpha > 0: assert all(x > 0 for x in probabilities.values())
    h.update(json.dumps([outcomes, str(denominator), [str(probabilities[x]) for x in outcomes]], separators=(",", ":")).encode())
print(json.dumps({"status":"PASS","cases":32768,"digest":h.hexdigest()}, sort_keys=True))
