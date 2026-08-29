#!/usr/bin/env python3
import hashlib,sys
from pathlib import Path
sys.dont_write_bytecode=True; HERE=Path(__file__).resolve().parent; sys.path.insert(0,str(HERE))
import generation_reference_a as A
import generation_reference_b as B
fixtures=[[],["a"],["a","bb"],["café"],["a"*64]]
for tokens in fixtures:
    left=A.serialize(tokens); right=B.serialize(tokens); assert left==right
    assert not left or left.endswith(b"\n")
assert hashlib.sha256(A.serialize(["a","café"])).hexdigest()=="bfc49c41fe16af9728de8b5d27b3f23670c817de3d5b1c9282e48dc25f39bd0c"
for bad in ([""],["a\nb"],["a"*65],["<EOS>"],["cafe\u0301"]):
    for fn in (A.serialize,B.serialize):
        try: fn(bad); raise AssertionError(bad)
        except ValueError: pass
print("CORPUS_SERIALIZATION=PASS")
