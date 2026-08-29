#!/usr/bin/env python3
"""Independent Task85c-d reproduction of R2-G01 and R2-G02."""

import hashlib
import json
from pathlib import Path

REPO = Path(__file__).resolve().parents[4]


def jid(obj):
    raw = (json.dumps(obj, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()
    return "j-" + hashlib.sha256(b"G1V2-JOB\0" + raw).hexdigest()[:40]


suite = json.loads((REPO / "research/phase3/task85c/golden/G1V2_GOLDEN_SUITE.json").read_text())
case = next(x for x in suite["cases"] if x["id"] == "JOBID")
old = case["input"]
v11_old_field = dict(old, contract_version="G1_V2_EXECUTABLE_CONTRACT_V1_1")
v11_registry = dict(v11_old_field)
v11_registry["dependency_job_ids"] = v11_registry.pop("dependencies")
got = [jid(old), jid(v11_old_field), jid(v11_registry)]
assert got[0] == case["expected"]
assert got == [
    "j-d85279815a36c30515b0be66387c99c3303fa09e",
    "j-f7c26e7460fa192e3186873428d5e2a37caa6285",
    "j-186f1406add6d4d4d7f788907efb76500468a5f7",
]
print("R2_G01=REPRODUCED")
print("R2_G02=REPRODUCED")
print("THREE_ID_DIVERGENCE=REPRODUCED")
