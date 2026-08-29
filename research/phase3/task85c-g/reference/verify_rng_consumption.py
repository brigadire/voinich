#!/usr/bin/env python3
import hashlib,json,struct
from pathlib import Path
OUT=Path(__file__).resolve().parent.parent
suite=json.loads((OUT/"G1V2_GENERATION_GOLDEN_SUITE_V1.json").read_text())
case=next(c for c in suite["cases"] if c["id"]=="PF-SC01")
root=bytes.fromhex("6f5a9c731de248b480c66b237ace215044689c5fa2f593e510b73dce18a49027")
ns=b"g1v2/control/generate"; counters=(0,0,0,0)
msg=b"G1V2-RNG\0"+root+struct.pack(">I",len(ns))+ns+struct.pack(">I",4)+b"".join(struct.pack(">Q",x) for x in counters)
d=hashlib.sha256(msg).digest(); u=(int.from_bytes(d[:8],"big")>>11)/2**53
assert format(u,".17g")=="0.92848667210989588"
assert case["expected"]["draws"]==1 and case["rng_trace"][0]["counters_after"]==[0,0,0,1]
assert case["rng_trace"][0]["selected"]=="d"
print("RNG_CONSUMPTION=PASS")

