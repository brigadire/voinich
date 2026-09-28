#!/usr/bin/env python3
"""Write deterministic project manifest and SHA256SUMS after validation."""
from __future__ import annotations

import hashlib
import json
from pathlib import Path

PKG = Path(__file__).resolve().parents[1]
ROOT = PKG.parents[1]
SOURCES = [
    ROOT / "research/astro_spatial_annotation/ASTRO_OBJECTS.tsv",
    ROOT / "research/astro_spatial_annotation/ASTRO_LABELS_SPATIAL.tsv",
    ROOT / "research/astro_spatial_annotation/ASTRO_LABEL_OBJECT_RELATIONS.tsv",
    ROOT / "research/astro_spatial_annotation_ai_b/AI_B_OBJECTS.tsv",
    ROOT / "research/astro_spatial_annotation_ai_b/AI_B_LABELS.tsv",
    ROOT / "research/astro_spatial_annotation_ai_b/AI_B_LABEL_OBJECT_RELATIONS.tsv",
    ROOT / "research/astro_spatial_annotation_ai_b2/AI2_OBJECTS.tsv",
    ROOT / "research/astro_spatial_annotation_ai_b2/AI2_LABELS.tsv",
    ROOT / "research/astro_spatial_annotation_ai_b2/AI2_LABEL_OBJECT_RELATIONS.tsv",
]


def digest(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    protocol = PKG / "HUMAN_ADJUDICATION_PROTOCOL_FROZEN.md"
    pdigest = digest(protocol)
    (PKG / "HUMAN_ADJUDICATION_PROTOCOL_FROZEN.sha256").write_text(f"{pdigest}  {protocol.name}\n", encoding="ascii")
    attachment_protocol = PKG / "HUMAN_LABEL_STAR_ATTACHMENT_PROTOCOL_V2.md"
    attachment_digest = digest(attachment_protocol)
    (PKG / "HUMAN_LABEL_STAR_ATTACHMENT_PROTOCOL_V2.sha256").write_text(
        f"{attachment_digest}  {attachment_protocol.name}\n", encoding="ascii"
    )
    excluded = {"SHA256SUMS", "manifest.json"}
    files = sorted(p for p in PKG.rglob("*") if p.is_file() and p.name not in excluded and "__pycache__" not in p.parts)
    manifest = {
        "manifest_version": "1.0", "package_type": "HUMAN_ASTRONOMICAL_SPATIAL_ADJUDICATION",
        "protocol_status": "FROZEN_FOR_PRODUCTION", "protocol_sha256": pdigest,
        "production_status": "STAR_LABEL_REVIEWED_LAYERS_FROZEN",
        "relation_status": "ATTACHMENT_V2_REVIEWED_LAYER_FROZEN",
        "active_relation_protocol_version": "2",
        "relation_ingest_status": "COMPLETE_FOR_SELECTED_QUEUE",
        "relation_output_sha256": digest(PKG / "HUMAN_LABEL_OBJECT_RELATIONS.tsv"),
        "relation_package_manifest_sha256": digest(PKG / "relations_v2/RELATIONS_V2_MANIFEST.json"),
        "spatial_v1_package_manifest_sha256": digest(PKG / "relations/RELATIONS_MANIFEST.json"),
        "relation_guide_sha256": digest(PKG / "HUMAN_LABEL_STAR_ATTACHMENT_V2_GUIDE.md"),
        "relation_protocol_sha256": attachment_digest,
        "reviewed_layer_sha256": {
            name: digest(PKG / name) for name in (
                "HUMAN_STAR_ADJUDICATION.tsv", "HUMAN_LABEL_ADJUDICATION.tsv",
                "HUMAN_PRODUCTION_UNREVIEWED.tsv",
            )
        },
        "source_original_sha256": {str(p.relative_to(ROOT)): digest(p) for p in SOURCES},
        "files": {str(p.relative_to(PKG)): {"sha256": digest(p), "bytes": p.stat().st_size} for p in files},
    }
    (PKG / "manifest.json").write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    checksummed = sorted(p for p in PKG.rglob("*") if p.is_file() and p.name != "SHA256SUMS" and "__pycache__" not in p.parts)
    (PKG / "SHA256SUMS").write_text("".join(f"{digest(p)}  {p.relative_to(PKG)}\n" for p in checksummed), encoding="ascii")


if __name__ == "__main__":
    main()
