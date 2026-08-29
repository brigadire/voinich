#!/usr/bin/env python3
import json
import sys
from pathlib import Path

sys.dont_write_bytecode = True

HERE = Path(__file__).resolve().parent
sys.path.insert(0, str(HERE))
from generation_reference_1 import sample as sample1
from generation_reference_2 import sample as sample2

vectors = json.loads((HERE.parent / "G1V2_GENERATION_GOLDEN_VECTORS.json").read_text())
checked = 0
for case in vectors["cases"]:
    if case["operation"] != "candidate_constrained_categorical":
        continue
    args = case["input"]
    left = sample1(args["outcomes"], args["probabilities"], set(args["admissible"]), float(args["u53"]))
    right = sample2(args["outcomes"], args["probabilities"], set(args["admissible"]), float(args["u53"]))
    assert left == right == case["expected"]
    checked += 1
assert checked >= 7
property_checks = 0
for probabilities in ([0.28, 0.22, 0.18, 0.12, 0.20], [0.0, 0.1, 0.2, 0.7, 0.0]):
    for numerator in range(8192):
        u53 = numerator / 8192.0
        args = (["a", "b", "c", "d", "<EOS>"], probabilities, {"a", "b", "c", "d"}, u53)
        left = sample1(*args)
        right = sample2(*args)
        assert left == right
        assert left["status"] == "OK" and left["outcome"] != "<EOS>" and left["draws"] == 1
        property_checks += 1
print(f"CANDIDATE_BOUNDARY_VECTORS={checked}")
print(f"CANDIDATE_NONEMPTY_PROPERTY_CHECKS={property_checks}")
print("CROSS_IMPLEMENTATION_CANDIDATE_RULE=PASS")
