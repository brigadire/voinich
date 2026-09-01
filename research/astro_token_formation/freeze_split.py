#!/usr/bin/env python3
"""Freeze a spelling-blind STAR/PLANET_MOON train/held-out sample."""

from __future__ import annotations

import csv
import hashlib
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
MATCHES = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv"
OUTPUT = ROOT / "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv"


def rank(salt: str, coordinate: str) -> str:
    return hashlib.sha256(f"{salt}|{coordinate}".encode()).hexdigest()


def main() -> None:
    with MATCHES.open(encoding="utf-8", newline="") as fh:
        source = list(csv.DictReader(fh, delimiter="\t"))
    grouped: dict[str, list[dict[str, str]]] = defaultdict(list)
    for row in source:
        if row["object_type"] in {"star", "planet?"}:
            grouped[row["stolfi_coordinate"]].append(row)

    rows: list[dict[str, str]] = []
    for coordinate, records in sorted(grouped.items()):
        cls = "STAR" if records[0]["object_type"] == "star" else "PLANET_MOON"
        matched = [r for r in records if r["match_status"] == "MATCHED"]
        positions = {r["absolute_token_positions"] for r in matched}
        tokens = {r["zl3b_eva_tokens"] for r in matched}
        loci = {r["zl3b_locus"] for r in matched}
        exact_anchor = any(r["match_method"] == "UNIQUE_EXACT" for r in matched)
        reasons: list[str] = []
        if len(matched) != len(records):
            reasons.append("HAS_UNMATCHED_RECORD")
        if not exact_anchor:
            reasons.append("NO_UNIQUE_EXACT_ANCHOR")
        if len(positions) != 1 or len(tokens) != 1 or len(loci) != 1:
            reasons.append("RECORD_DISAGREEMENT")
        token = next(iter(tokens), "")
        position = next(iter(positions), "")
        if "," in position or " " in token:
            reasons.append("MULTI_TOKEN_LABEL")
        eligible = not reasons
        rows.append({
            "stolfi_coordinate": coordinate,
            "object_class": cls,
            "panel": records[0]["panel"],
            "stolfi_group": records[0]["stolfi_group"],
            "voynich_token": token,
            "absolute_token_position": position,
            "zl3b_locus": next(iter(loci), ""),
            "source_record_ids": ";".join(r["record_id"] for r in records),
            "eligible": "1" if eligible else "0",
            "sample_rank_sha256": rank("ASTRO_FORMATION_SAMPLE_V1", coordinate),
            "split_rank_sha256": rank("ASTRO_FORMATION_SPLIT_V1", coordinate),
            "selected": "0",
            "split": "EXCLUDED",
            "exclusion_reason": ";".join(reasons) if reasons else "CAPACITY_SAMPLE",
        })

    for cls, sample_n, train_n in (("STAR", 21, 16), ("PLANET_MOON", 5, 4)):
        eligible = sorted(
            (r for r in rows if r["object_class"] == cls and r["eligible"] == "1"),
            key=lambda r: r["sample_rank_sha256"],
        )
        if len(eligible) < sample_n:
            raise RuntimeError(f"only {len(eligible)} eligible {cls} labels")
        chosen = eligible[:sample_n]
        train_coordinates = {
            r["stolfi_coordinate"]
            for r in sorted(chosen, key=lambda r: r["split_rank_sha256"])[:train_n]
        }
        for row in chosen:
            row["selected"] = "1"
            row["split"] = (
                "TRAIN" if row["stolfi_coordinate"] in train_coordinates else "HELD_OUT"
            )
            row["exclusion_reason"] = ""

    fields = list(rows[0])
    with OUTPUT.open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n")
        writer.writeheader()
        writer.writerows(sorted(rows, key=lambda r: (r["panel"], r["stolfi_coordinate"])))


if __name__ == "__main__":
    main()
