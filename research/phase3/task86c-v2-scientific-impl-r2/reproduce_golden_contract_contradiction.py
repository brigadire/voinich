#!/usr/bin/env python3
"""Reproduce the frozen V1.1 JobID/golden contradiction."""

from __future__ import annotations

import hashlib
import json
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]
GOLDEN = REPO / "research/phase3/task85c/golden/G1V2_GOLDEN_SUITE.json"
DAG = REPO / "research/phase3/task85c-c/registries/G1V2_DAG_CONTRACT.json"
CONTRACT = REPO / "research/phase3/task85c-c/G1V2_EXECUTABLE_CONTRACT_V1_1.json"


def canonical(value: object) -> bytes:
    return (json.dumps(value, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode("utf-8")


def job_id(payload: dict[str, object]) -> str:
    digest = hashlib.sha256(b"G1V2-JOB\0" + canonical(payload)).hexdigest()
    return "j-" + digest[:40]


def main() -> None:
    contract = json.loads(CONTRACT.read_text(encoding="utf-8"))
    dag = json.loads(DAG.read_text(encoding="utf-8"))
    suite = json.loads(GOLDEN.read_text(encoding="utf-8"))
    case = next(item for item in suite["cases"] if item["id"] == "JOBID")

    frozen_payload = case["input"]
    frozen_actual = job_id(frozen_payload)
    assert frozen_actual == case["expected"]

    v11_payload = dict(frozen_payload)
    v11_payload["contract_version"] = contract["contract_version"]
    v11_actual = job_id(v11_payload)

    registry_payload = dict(v11_payload)
    registry_payload["dependency_job_ids"] = registry_payload.pop("dependencies")
    registry_actual = job_id(registry_payload)

    expected_fields = dag["job_id"]["payload_fields"]
    assert "dependency_job_ids" in expected_fields
    assert "dependencies" not in expected_fields
    assert frozen_payload["contract_version"] != contract["contract_version"]
    assert len({frozen_actual, v11_actual, registry_actual}) == 3

    print(f"GOLDEN_CONTRACT_VERSION={frozen_payload['contract_version']}")
    print(f"V1_1_CONTRACT_VERSION={contract['contract_version']}")
    print(f"GOLDEN_JOB_ID={frozen_actual}")
    print(f"V1_1_SAME_FIELDS_JOB_ID={v11_actual}")
    print(f"V1_1_DAG_REGISTRY_FIELDS_JOB_ID={registry_actual}")
    print("GOLDEN_CONTRACT_CONTRADICTION=REPRODUCED")


if __name__ == "__main__":
    main()
