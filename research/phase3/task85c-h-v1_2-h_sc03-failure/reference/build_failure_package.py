#!/usr/bin/env python3
"""Build the deterministic Task85c-h defect manifest."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path

TASK = Path(__file__).resolve().parent.parent
MANIFEST = TASK / "TASK85C_H_RESULTS_MANIFEST.json"


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    files = sorted(
        (p for p in TASK.rglob("*") if p.is_file() and p != MANIFEST and "__pycache__" not in p.parts),
        key=lambda p: p.relative_to(TASK).as_posix().encode(),
    )
    artifacts = [
        {"path": p.relative_to(TASK).as_posix(), "bytes": p.stat().st_size, "sha256": digest(p)}
        for p in files
    ]
    root_lines = "".join(f'{x["sha256"]}  {x["path"]}\n' for x in artifacts).encode()
    manifest = {
        "schema": "task85c-h-results-v1",
        "status": "SCIENTIFIC_CONTRACT_DEFECT",
        "finding_ids": ["H-SC03-M0-UNK-PROBABILITY-CONTRADICTION"],
        "scientific_contract_version": "G1_V2_EXECUTABLE_CONTRACT_V1_2",
        "scientific_contract_machine_sha256": "29e39e0c25dc8033f784480fdc537e3ede9eeb69baa0607c9f249d796d6b42dc",
        "scientific_contract_markdown_sha256": "ec60bb23e55ce157fe954b5cafc63d22ab70ecec390822cb63f9ae273142c639",
        "i1_machine_sha256": "35ecf0bfc9a9c27bb63d33b074bc399dd9256692620ee06a54e592e0d1e867b2",
        "e2_machine_sha256": "cfaa8c1cb787380baca5a391d077e5d2b855df8bba767ad5b601457e06eb0070",
        "evidence_schema_root_sha256": "39c4c3bee96ee58ddd38552cbb16fdc2f994a390b86a6645e1420c3cb67eca81",
        "generation_semantics_sha256": "45d533f8b83b24c77a96836fa5c2ef95f9b948003bd2ed725fc2ea97e010b310",
        "generation_golden_suite_sha256": "143954667073a2c10f1bd59ce98b9c93dd84b50632bb67ea80d0d92449480acb",
        "task85c_i_artifact_root_sha256": "4dca7e3bf3d28ad1852ac0ebfee3ae6f4e4410227a1e57d491587874a126f84c",
        "scientific_firewall": "INTACT",
        "production_materialization": {"escrow_keys": 0, "controls": 0, "jobids": 0, "dag_created": False},
        "artifacts": artifacts,
        "artifact_root_definition": "sha256 of sha256sum-format lines over task-relative paths sorted by UTF-8 bytes; excludes results manifest and __pycache__",
        "artifact_root_excluding_manifest_sha256": hashlib.sha256(root_lines).hexdigest(),
    }
    MANIFEST.write_text(json.dumps(manifest, sort_keys=True, separators=(",", ":")) + "\n")
    print(manifest["artifact_root_excluding_manifest_sha256"])


if __name__ == "__main__":
    main()
