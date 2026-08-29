#!/usr/bin/env python3
import sys
from pathlib import Path
sys.dont_write_bytecode=True; HERE=Path(__file__).resolve().parent; sys.path.insert(0,str(HERE))
import generation_reference_a as A
import generation_reference_b as B
labels=["a","b","c","d"]; weights=[.4,.3,.2,.1]; allowed=set(labels)
assert A.alias_table(labels,weights,allowed)==B.alias_table(labels,weights,allowed)
checks=0
for i in range(4096):
    u1=i/4096; u2=((i*4051)%4096)/4096
    assert A.alias_sample(labels,weights,allowed,u1,u2)==B.alias_sample(labels,weights,allowed,u1,u2)
    uniforms=[((i*(j+3)+j+1)%4096)/4096 for j in range(4)]
    assert A.exponential_race(labels,weights,allowed,uniforms)==B.exponential_race(labels,weights,allowed,uniforms)
    checks+=2
print(f"GENERATOR_B_PROPERTY_CASES={checks}:PASS")

