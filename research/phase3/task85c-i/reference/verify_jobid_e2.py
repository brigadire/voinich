#!/usr/bin/env python3
"""Independent Python implementation of the E2 JobID fixture."""
import hashlib
import json
import unicodedata
from pathlib import Path

OUT=Path(__file__).resolve().parent.parent
def norm(x):
    if isinstance(x,str): return unicodedata.normalize("NFC",x)
    if isinstance(x,list): return [norm(v) for v in x]
    if isinstance(x,dict): return {k:norm(x[k]) for k in sorted(x)}
    return x
def canon(x): return (json.dumps(norm(x),ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n").encode()
def jobid(x): return "j-"+hashlib.sha256(b"G1V2-JOB\0"+canon(x)).hexdigest()[:40]
f=json.loads((OUT/"fixtures/G1V2_E2_JOBID_FIXTURE.json").read_text())
for version in ("v1_1","v1_2"):
    assert canon(f[version]["payload"]).hex()==f[version]["canonical_payload_hex"]
    assert jobid(f[version]["payload"])==f[version]["jobid"]
assert f["v1_1"]["jobid"]!=f["v1_2"]["jobid"]
e2=json.loads((OUT/"G1V2_EXECUTION_IDENTITY_ERRATUM_E2.json").read_text())
assert e2["jobid"]["dependency_field"]=="dependency_job_ids"
assert e2["scientific_boundary"]["scientific_identity_depends_on_blind_id"] is False
assert e2["scientific_boundary"]["scientific_identity_depends_on_escrow_key"] is False
print("JOBID_E2_PYTHON=PASS")
print("H_SC02=CLOSED")
