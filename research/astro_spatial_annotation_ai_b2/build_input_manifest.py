#!/usr/bin/env python3
"""Builds AI2_INPUT_MANIFEST.json for the clean-room ANNOTATOR_B_AI_2 package."""
from __future__ import annotations
import hashlib
import json
from pathlib import Path
from datetime import datetime, timezone

OUT = Path(__file__).resolve().parent
PACKAGE_DIR = OUT / "package"
CROPS_DIR = PACKAGE_DIR / "crops"
SCHEMAS_DIR = PACKAGE_DIR / "schemas"

PANELS = ["f67r1", "f67r2", "f67v1", "f68r1", "f68r2", "f68r3", "f68v1", "f68v2"]
PANEL_DIMS = {
    "f67r1": (2486, 3738),
    "f67r2": (2486, 3738),
    "f67v1": (2565, 3753),
    "f68r1": (2462, 3828),
    "f68r2": (2078, 3828),
    "f68r3": (3453, 3828),
    "f68v1": (2530, 3843),
    "f68v2": (2083, 3843),
}


def sha256_file(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    input_files = {}
    for p in PANELS:
        crop_path = CROPS_DIR / f"{p}.jpg"
        w, h = PANEL_DIMS[p]
        input_files[f"crops/{p}.jpg"] = {
            "sha256": sha256_file(crop_path),
            "width": w,
            "height": h,
            "bytes": crop_path.stat().st_size,
        }

    for fpath in [
        PACKAGE_DIR / "ANNOTATOR_B_AI_2_INSTRUCTIONS.md",
        PACKAGE_DIR / "OBJECT_ONTOLOGY_CLEAN.md",
        SCHEMAS_DIR / "AI2_OBJECTS_SCHEMA.tsv",
        SCHEMAS_DIR / "AI2_LABELS_SCHEMA.tsv",
        SCHEMAS_DIR / "AI2_LABEL_OBJECT_RELATIONS_SCHEMA.tsv",
    ]:
        rel = str(fpath.relative_to(PACKAGE_DIR))
        input_files[rel] = {
            "sha256": sha256_file(fpath),
            "bytes": fpath.stat().st_size,
        }

    manifest_data = {
        "manifest_version": "1.0",
        "package_type": "AI2_INPUT_PACKAGE",
        "created_at": datetime.now(timezone.utc).isoformat(),
        "annotator_id": "ANNOTATOR_B_AI_2",
        "model_name": "Claude Opus 5",
        "model_version_if_known": "claude-opus-5",
        "provider": "Anthropic (invoked as a fresh, context-free Claude Code subagent, model override 'opus')",
        "run_timestamp": "see AI2_STAGE1_MANIFEST.json (frozen_at) / AI2_STAGE2_MANIFEST.json (frozen_at) for exact pass timestamps",
        "independence_note": (
            "Distinct model family from the ANNOTATOR_B_AI (AI_1) pass and from the "
            "orchestrating session (Claude Sonnet 5). Launched as a brand-new agent "
            "instance with no conversation history, given only this package directory."
        ),
        "same_canonical_image_bytes_as_ai_b_1": True,
        "files": input_files,
    }

    (PACKAGE_DIR / "AI2_INPUT_MANIFEST.json").write_text(
        json.dumps(manifest_data, indent=2), encoding="utf-8"
    )
    (OUT / "AI2_INPUT_MANIFEST.json").write_text(
        json.dumps(manifest_data, indent=2), encoding="utf-8"
    )
    print("AI2_INPUT_MANIFEST.json written.")


if __name__ == "__main__":
    main()
