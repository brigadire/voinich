#!/usr/bin/env python3
import json,subprocess
from pathlib import Path
R=Path(__file__).resolve().parents[1]
i=json.loads((R/"golden/inherited-roots.json").read_text());assert i["task85c_root"]=="b7443a962a82dd5c0cd67b71e24d8acea73fc9be4863fca4078bc53e468c7e51";assert i["status_machine_root"]=="95c0e6bf4c1edeadd4c823b637223cb2440eb6c798e7c16e3bdc7bceb6dbba65"
subprocess.run(["node",str(R/"reference/validate_schemas.mjs")],check=True,stdout=subprocess.DEVNULL)
print("G1V2_GOLDEN_SUITE_V1_1=PASS")
