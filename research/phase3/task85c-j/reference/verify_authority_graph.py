#!/usr/bin/env python3
import hashlib,json
from pathlib import Path

out=Path(__file__).resolve().parent.parent
g=json.loads((out/"G1V2_TASK85C_J_AUTHORITY_GRAPH.json").read_text())
nodes={x["id"]:x for x in g["nodes"]}
assert len(nodes)==len(g["nodes"])
for x in nodes.values(): assert hashlib.sha256((out/x["path"]).read_bytes()).hexdigest()==x["sha256"]
adj={x:[] for x in nodes}
for e in g["edges"]: assert e["from"] in nodes and e["to"] in nodes; adj[e["from"]].append(e["to"])
seen=set(); active=set()
def visit(x):
    if x in active: raise AssertionError("cycle")
    if x in seen:return
    active.add(x)
    for y in adj[x]:visit(y)
    active.remove(x);seen.add(x)
for x in nodes:visit(x)
assert g["cycles"]==g["unresolved_edges"]==g["mixed_scientific_identities"]==0
print("AUTHORITY_GRAPH=PASS")
