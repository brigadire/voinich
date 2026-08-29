#!/usr/bin/env python3
"""Verify authority node hashes, edge closure, and acyclicity."""
import hashlib
import json
from pathlib import Path

OUT=Path(__file__).resolve().parent.parent
g=json.loads((OUT/"G1V2_V1_2_AUTHORITY_GRAPH.json").read_text())
ids={n["id"] for n in g["nodes"]}
assert len(ids)==len(g["nodes"])
assert all(e["from"] in ids and e["to"] in ids for e in g["edges"])
adj={x:[] for x in ids}
for e in g["edges"]: adj[e["from"]].append(e["to"])
color={x:0 for x in ids}
def visit(x):
    assert color[x]!=1,"authority cycle"
    if color[x]==2:return
    color[x]=1
    for y in adj[x]:visit(y)
    color[x]=2
for x in sorted(ids):visit(x)
assert g["cycles"]==0 and g["unresolved_authority_edges"]==0
root_node=next(n for n in g["nodes"] if n["id"]=="EVIDENCE_SCHEMA_ROOT_V1_2")
root_path=OUT/"G1V2_V1_2_EVIDENCE_SCHEMA_ROOT.json"
assert root_node["sha256"]==hashlib.sha256(root_path.read_bytes()).hexdigest()
assert root_node["schema_family_root_sha256"]==json.loads(root_path.read_text())["root_sha256"]
print("AUTHORITY_GRAPH_CYCLES=0")
print("AUTHORITY_GRAPH_UNRESOLVED_EDGES=0")
