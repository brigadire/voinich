#!/usr/bin/env python3
"""Reproduce PF-SC01 from frozen V1.1 machine artifacts."""

import csv
import hashlib
import json
import struct
from pathlib import Path

REPO = Path(__file__).resolve().parents[4]
T85 = REPO / "research/phase3/task85c"
T85CC = REPO / "research/phase3/task85c-c"

contract = json.loads((T85CC / "G1V2_EXECUTABLE_CONTRACT_V1_1.json").read_text())
with (T85 / "G1V2_SYNTHETIC_GENERATOR_REGISTRY.tsv").open(newline="") as f:
    generators = {r["generator_id"]: r for r in csv.DictReader(f, delimiter="\t")}
with (T85CC / "registries/G1V2_RNG_DOMAIN_REGISTRY.tsv").open(newline="") as f:
    domains = {r["domain_id"]: r for r in csv.DictReader(f, delimiter="\t")}

assert contract["data"]["tokens"].startswith("nonempty")
row = generators["M0_GEN_A"]
params = json.loads(row["parameters"])
assert params["outcomes"] == ["a", "b", "c", "d", "<EOS>"]
assert params["probabilities"] == ["0.28", "0.22", "0.18", "0.12", "0.20"]
assert domains["CONTROL_GENERATE"]["namespace"] == "g1v2/control/generate"

root = bytes.fromhex(contract["rng"]["root_seed_hex"])
namespace = domains["CONTROL_GENERATE"]["namespace"].encode()


def u53(draw):
    counters = (0, 0, 0, draw)
    message = (
        b"G1V2-RNG\0" + root + struct.pack(">I", len(namespace)) + namespace
        + struct.pack(">I", len(counters))
        + b"".join(struct.pack(">Q", value) for value in counters)
    )
    digest = hashlib.sha256(message).digest()
    return digest.hex(), (int.from_bytes(digest[:8], "big") >> 11) / 2**53


digest0, first = u53(0)
digest1, second = u53(1)
assert first >= 0.8  # EOS is the first sampled outcome.
assert second < 0.28  # retry convention would start the token with "a".

# Conditional first-event convention normalizes non-EOS mass and reuses u0.
conditional = first * 0.8
assert 0.68 <= conditional < 0.8  # starts the token with "d".

print("PF_SC01_EMPTY_TOKEN_CONTRACT_DEFECT=REPRODUCED")
print("DRAW0_SHA256=" + digest0)
print("DRAW0_U53=" + format(first, ".17g"))
print("DRAW0_OUTCOME=<EOS>")
print("RETRY_DRAW1_FIRST_GLYPH=a")
print("CONDITIONAL_RENORMALIZATION_FIRST_GLYPH=d")
print("SCIENTIFIC_CORPUS_VARIANTS=2")
