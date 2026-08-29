#!/usr/bin/env python3
import json,os,subprocess,sys
from pathlib import Path
sys.dont_write_bytecode=True; HERE=Path(__file__).resolve().parent; OUT=HERE.parent; sys.path.insert(0,str(HERE))
import generation_reference_a as A
import generation_reference_b as B
sem=json.loads((OUT/"G1V2_GENERATION_SEMANTICS_V1.json").read_text())
assert len(sem["routes"])==12 and sem["length"]["maximum"]==64
labels=["a","b","c","d","<EOS>"]; weights=[.28,.22,.18,.12,.20]; allowed=set(labels[:-1])
checks=0
for i in range(32768):
    u=i/32768
    left=A.categorical(labels,weights,allowed,u); right=B.categorical(labels,weights,allowed,u)
    assert left==right and left[0]=="OK" and left[1]!="<EOS>" and left[2]==1
    checks+=1
uniforms=[.1,.2,.3,.4,.9,.95]
assert A.stream_token(labels,weights,uniforms)==B.stream_token(labels,weights,uniforms)
env=os.environ.copy(); env["GOCACHE"]="/tmp/task85cg-gocache"; env["GOMODCACHE"]="/tmp/task85cg-gomodcache"
run=subprocess.run(["go","run",str(HERE/"generation_reference_go.go"),"--quiet"],cwd=HERE.parents[3],env=env,text=True,capture_output=True,check=True)
assert "GO_REFERENCE=PASS" in run.stdout and "PF_SC01=d/1" in run.stdout
print(f"COMPLETE_GENERATION_PROPERTY_CASES={checks}:PASS")
print("GO_PYTHON_DIFFERENTIAL=PASS")

