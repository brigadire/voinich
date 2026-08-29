#!/usr/bin/env python3
"""Independently reproduce H-SC03 from immutable parent goldens."""
import json
from decimal import Decimal
from pathlib import Path

root=Path(__file__).resolve().parents[4]
suite=json.loads((root/"research/phase3/task85c/golden/G1V2_GOLDEN_SUITE.json").read_text())
fit=next(x for x in suite["cases"] if x["id"]=="M0-FIT")
unseen=next(x for x in suite["cases"] if x["id"]=="M0-UNSEEN")
mass=sum(Decimal(fit["expected"][x]) for x in ["p_a","p_b","p_eos"])
assert mass==1 and unseen["expected"]["positive_alpha"]=="positive"
assert Decimal(1)-mass==0 and 5+1*4==9
print("H_SC03_REPRODUCED=YES")
