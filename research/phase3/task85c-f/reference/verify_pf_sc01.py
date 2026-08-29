#!/usr/bin/env python3
import hashlib
import json
import struct
import sys
from pathlib import Path

sys.dont_write_bytecode = True

HERE = Path(__file__).resolve().parent
ROOT = HERE.parents[3]

root = bytes.fromhex("6f5a9c731de248b480c66b237ace215044689c5fa2f593e510b73dce18a49027")
namespace = b"g1v2/control/generate"


def draw(index):
    counters = (0, 0, 0, index)
    message = (b"G1V2-RNG\0" + root + struct.pack(">I", len(namespace)) + namespace
               + struct.pack(">I", 4) + b"".join(struct.pack(">Q", x) for x in counters))
    digest = hashlib.sha256(message).digest()
    return digest.hex(), (int.from_bytes(digest[:8], "big") >> 11) / 2**53


d0, u0 = draw(0)
d1, u1 = draw(1)
assert d0 == "edb14d737bdf70ce62ff1a28cfbdcc6f05a1fc904a34ff493da64264578f6794"
assert format(u0, ".17g") == "0.92848667210989588"
assert 0.8 <= u0 < 1.0 and u1 < 0.28

sys.path.insert(0, str(HERE))
from generation_reference_1 import sample as sample1
from generation_reference_2 import sample as sample2

outcomes = ["a", "b", "c", "d", "<EOS>"]
probabilities = ["0.28", "0.22", "0.18", "0.12", "0.20"]
expected = {"status": "OK", "outcome": "d", "draws": 1}
assert sample1(outcomes, probabilities, {"a", "b", "c", "d"}, u0) == expected
assert sample2(outcomes, probabilities, {"a", "b", "c", "d"}, u0) == expected
fixture = json.loads((HERE.parent / "G1V2_PF_SC01_REGRESSION.json").read_text())
assert fixture["observed_v1_1"]["draw0_digest"] == d0
assert fixture["candidate_repair"]["first_emitted_glyph"] == "d"
print("PF_SC01_REPRODUCED=YES")
print("RETRY_FIRST_GLYPH=a")
print("CONDITIONAL_FIRST_GLYPH=d")
