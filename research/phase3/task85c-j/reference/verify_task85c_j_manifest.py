#!/usr/bin/env python3
import hashlib,json
from pathlib import Path

out=Path(__file__).resolve().parent.parent
m=json.loads((out/"TASK85C_J_RESULTS_MANIFEST.json").read_text())
lines=[]
for x in m["artifacts"]:
    p=out/x["path"]
    assert p.stat().st_size==x["bytes"] and hashlib.sha256(p.read_bytes()).hexdigest()==x["sha256"]
    lines.append(f'{x["sha256"]}  {x["path"]}\n')
assert hashlib.sha256("".join(lines).encode()).hexdigest()==m["artifact_root_excluding_manifest_sha256"]
assert m["task85c_h_retry_ready"]=="SUPPORTED" and m["h_sc03"]=="CLOSED"
print("TASK85C_J_MANIFEST=PASS")
