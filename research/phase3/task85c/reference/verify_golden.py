#!/usr/bin/env python3
"""Independent arithmetic checks for selected frozen golden vectors."""
import hashlib
import json
import math
import struct
import unicodedata
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
SEED = bytes.fromhex("6f5a9c731de248b480c66b237ace215044689c5fa2f593e510b73dce18a49027")


def canonical(obj):
    def n(x):
        if isinstance(x, str): return unicodedata.normalize("NFC", x)
        if isinstance(x, list): return [n(v) for v in x]
        if isinstance(x, dict): return {unicodedata.normalize("NFC", k): n(v) for k, v in x.items()}
        return x
    return (json.dumps(n(obj), ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()


suite = json.loads((ROOT / "golden/G1V2_GOLDEN_SUITE.json").read_text())
cases = {x["id"]: x for x in suite["cases"]}

ns = b"g1v2/generate"
msg = b"G1V2-RNG\0" + SEED + struct.pack(">I", len(ns)) + ns + struct.pack(">I", 5) + struct.pack(">QQQQQ", 0, 0, 0, 0, 0)
digest = hashlib.sha256(msg).hexdigest()
assert digest == cases["RNG-01"]["expected"]["digest_hex"]

p = (3 / 8) * (2 / 8) * (3 / 8)
assert abs(-math.log2(p) - float(cases["M0-FIT"]["expected"]["score_ab_bits"])) < 1e-15

xs = [.01, .02, .03, .04]
h = (len(xs) - 1) * .95
q = xs[math.floor(h)] + (h - math.floor(h)) * (xs[math.ceil(h)] - xs[math.floor(h)])
assert abs(q - float(cases["STRUCT-THRESH"]["expected"])) < 1e-15

assert 192 + 192 * 43 * 160 == cases["DAG"]["expected"]["jobs"]
assert 192 * 43 * 316 + 192 * 43 == cases["DAG"]["expected"]["edges"]

payload = cases["JOBID"]["input"]
jobid = "j-" + hashlib.sha256(b"G1V2-JOB\0" + canonical(payload)).hexdigest()[:40]
assert jobid == cases["JOBID"]["expected"]
assert canonical({"b": 1, "a": "e\u0301"}).hex() == cases["CANON"]["expected_hex"]

print("G1V2_GOLDEN_REFERENCE=PASS")
