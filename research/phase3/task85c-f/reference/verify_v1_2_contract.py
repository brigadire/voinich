#!/usr/bin/env python3
"""Verify that the failure package does not falsely claim a V1.2 freeze."""

import csv
import hashlib
import json
from pathlib import Path

HERE = Path(__file__).resolve().parent
OUT = HERE.parent
REPO = HERE.parents[3]

assert hashlib.sha256((REPO / "research/phase3/task85c-c/G1V2_EXECUTABLE_CONTRACT_V1_1.md").read_bytes()).hexdigest() == "5c3cd272c1dbae9bfe1d7a100155faf102e86d34660da239e1cb31704ad470b0"
assert hashlib.sha256((REPO / "research/phase3/task85c-e/G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json").read_bytes()).hexdigest() == "dbfb9a4a7101eed7006f751b9c4631b5f0286c3792f9777cc833c5dcfa42a3d3"
assert hashlib.sha256((REPO / "research/phase3/task85c-c/registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json").read_bytes()).hexdigest() == "fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9"
assert not (OUT / "G1_V2_EXECUTABLE_CONTRACT_V1_2.json").exists()
assert not (OUT / "G1_V2_EXECUTABLE_CONTRACT_V1_2.md").exists()
assert (OUT / "TASK85C_F_SCIENTIFIC_CONTRACT_DEFECT_EXPANDED").exists()
with (OUT / "G1V2_GENERATION_BOUNDARY_AUDIT.tsv").open(newline="") as stream:
    rows = list(csv.DictReader(stream, delimiter="\t"))
assert len(rows) == 25
assert sum(row["ambiguity"] != "NONE" for row in rows) >= 8
manifest = json.loads((OUT / "TASK85C_F_RESULTS_MANIFEST.json").read_text())
assert manifest["v1_2_ready"] == "NOT_SUPPORTED"
assert manifest["terminal_marker"] == "TASK85C_F_SCIENTIFIC_CONTRACT_DEFECT_EXPANDED"
lines = []
for artifact in manifest["artifacts_excluding_manifest"]:
    path = OUT / artifact["path"]
    digest = hashlib.sha256(path.read_bytes()).hexdigest()
    assert digest == artifact["sha256"], artifact["path"]
    repo_relative = path.relative_to(REPO).as_posix()
    lines.append((repo_relative, f"{digest}  {repo_relative}\n"))
assert hashlib.sha256("".join(line for _, line in sorted(lines)).encode()).hexdigest() == manifest["artifact_root_excluding_manifest_sha256"]
print("FAILURE_PACKAGE_VALIDATION=PASS")
print("V1_2_READY=NOT_SUPPORTED")
