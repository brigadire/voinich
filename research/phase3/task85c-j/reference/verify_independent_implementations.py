#!/usr/bin/env python3
"""Compare CONTRACT_REFERENCE_ONLY Python and Go M0 implementations."""
import json, os, subprocess
from decimal import Decimal
from pathlib import Path
from m0_reference_impl_a import fit

here=Path(__file__).resolve().parent
env=dict(os.environ); env["GOCACHE"]="/tmp/task85cj-gocache"
go=json.loads(subprocess.check_output(["go","run",str(here/"m0_reference_impl_b.go")],text=True,env=env))
outcomes,counts,denominator,p=fit(["ab","a"],Decimal(1))
assert go["cases"]==32768 and go["generation_cases"]==8192
assert go["fixture_denominator"]==float(denominator)
assert go["fixture_probabilities"]==[float(p[x]) for x in outcomes]
print("INDEPENDENT_IMPLEMENTATIONS=PASS LANGUAGES=Python,Go")
