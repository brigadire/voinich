#!/usr/bin/env python3
import json, sys
from pathlib import Path
sys.dont_write_bytecode=True
HERE=Path(__file__).resolve().parent; sys.path.insert(0,str(HERE))
import generation_reference_a as A
import generation_reference_b as B

suite=json.loads((HERE.parent/"G1V2_GENERATION_GOLDEN_SUITE_V1.json").read_text())
semantics=json.loads((HERE.parent/"G1V2_GENERATION_SEMANTICS_V1.json").read_text())
routes={r["generator_id"]:r for r in semantics["routes"]}
checked=0
for case in suite["cases"]:
    op=case["operation"]; data=case.get("input",{}); exp=case.get("expected",{})
    if op=="categorical":
        args=(data["outcomes"],data["weights"],set(data["allowed"]),data["u53"])
        values=[A.categorical(*args),B.categorical(*args)]
        for status,outcome,draws in values:
            assert (status,outcome,draws)==(exp["status"],exp.get("outcome"),exp["draws"]),case["id"]
    elif op=="race":
        args=(data["outcomes"],data["weights"],set(data["allowed"]),data["uniforms"])
        for value in (A.exponential_race(*args),B.exponential_race(*args)):
            assert value==(exp["status"],exp["outcome"],exp["draws"]),case["id"]
    elif op=="alias":
        args=(data["outcomes"],data["weights"],set(data["allowed"]),data["u_column"],data["u_threshold"])
        for value in (A.alias_sample(*args),B.alias_sample(*args)):
            assert value==(exp["status"],exp["outcome"],exp["draws"]),case["id"]
    elif op=="serialize":
        for fn in (A.serialize,B.serialize): assert fn(data["tokens"]).hex()==exp["hex"],case["id"]
    elif op=="route_binding":
        route=routes[data["generator_id"]]
        assert {k:route[k] for k in ("model","author","primitive","max_token_glyphs","attempt_cap")}==exp,case["id"]
    else: continue
    checked+=1
assert checked>=15
print(f"GENERATION_GOLDENS={checked}:PASS")
