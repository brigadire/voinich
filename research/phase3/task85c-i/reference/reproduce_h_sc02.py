#!/usr/bin/env python3
"""Small frozen-parent reproduction of H-SC02 and its JobID consequence."""
import hashlib
import json
import unicodedata
from pathlib import Path

HERE = Path(__file__).resolve().parent
REPO = HERE.parents[3]
E1 = REPO / "research/phase3/task85c-e/G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json"

def norm(x):
    if isinstance(x, str): return unicodedata.normalize("NFC", x)
    if isinstance(x, list): return [norm(v) for v in x]
    if isinstance(x, dict): return {k: norm(x[k]) for k in sorted(x)}
    return x

def canon(x): return (json.dumps(norm(x), ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()

e1 = json.loads(E1.read_text())
assert e1["contract_version"] == e1["jobid"]["scientific_identity_version"] == "G1_V2_EXECUTABLE_CONTRACT_V1_1"
p = {"candidate_id":"M0-iid-0","contract_version":e1["jobid"]["scientific_identity_version"],"control_instance_id":"OPEN-INTEGRATION-FIXTURE-1","dependency_job_ids":[],"metric_id_or_null":None,"replicate_or_null":None,"scale_or_null":None,"stage":"FIT"}
q = dict(p, contract_version="G1_V2_EXECUTABLE_CONTRACT_V1_2")
j = lambda x: "j-" + hashlib.sha256(b"G1V2-JOB\0" + canon(x)).hexdigest()[:40]
assert canon(p) != canon(q) and j(p) != j(q)
print("H_SC02_REPRODUCED=YES")
