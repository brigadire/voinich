from __future__ import annotations

import csv
import hashlib
import importlib.util
import json
import shutil
import tempfile
import unittest
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/global_structural_alignment_v1"
BASE = ROOT / "research/astro_hapax_star_label"
LEGACY = BASE / "legacy_mapping_migration_v1"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def rows(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


class StructuralAlignmentTests(unittest.TestCase):
    def test_inputs_and_upstream_are_frozen(self):
        for row in rows(OUT / "INPUT_MANIFEST.tsv"):
            self.assertEqual(row["sha256"], sha(ROOT / row["path"]), row["path"])
        self.assertEqual(len(rows(OUT / "SPATIAL_LABEL_FEATURES.tsv")), 92)
        self.assertEqual(len(rows(OUT / "LEGACY_ANCHORS.tsv")), 21)
        self.assertEqual(len(rows(OUT / "UNRESOLVED_LABELS.tsv")), 71)

    def test_features_are_blind_and_coordinates_valid(self):
        features = rows(OUT / "SPATIAL_LABEL_FEATURES.tsv")
        forbidden = ("group", "star_count", "human_added", "hapax", "frequency", "lexicon")
        self.assertFalse(any(any(term in key.lower() for term in forbidden) for key in features[0]))
        for row in features:
            self.assertTrue(0 <= float(row["normalized_x"]) <= 1)
            self.assertTrue(0 <= float(row["normalized_y"]) <= 1)
            self.assertTrue(0 <= float(row["polar_angle_radians"]) < 2 * 3.141592654)
            self.assertGreaterEqual(float(row["polar_radius"]), 0)
        self.assertEqual(len({r["canonical_label_id"] for r in features}), 92)

    def test_sequence_roles_and_occurrences(self):
        sequences = rows(OUT / "TRANSCRIPTION_SEQUENCE_REGISTRY.tsv")
        self.assertTrue({"LABEL_LINES", "CYCLIC_TEXT", "RADIAL_TEXT"} <= {r["sequence_role"] for r in sequences})
        self.assertEqual(sum(r["panel"] == "f68r1" and r["sequence_role"] == "LABEL_LINES" for r in sequences), 1)
        occurrence_ids = {r["occurrence_id"] for r in rows(BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv")}
        for sequence in sequences:
            self.assertTrue(set(sequence["occurrence_ids"].split(";")) <= occurrence_ids)
            if sequence["sequence_role"] == "CYCLIC_TEXT":
                self.assertEqual(sequence["cyclic_status"], "CYCLIC")

    def test_anchor_provenance_and_geometry_identity(self):
        anchors = rows(OUT / "LEGACY_ANCHORS.tsv")
        source = {r["canonical_label_id"]: r for r in rows(LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv")}
        occurrence_ids = {r["occurrence_id"] for r in rows(BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv")}
        self.assertTrue(all(r["provenance"] == "LEGACY_CONFIRMED_FROZEN_OCCURRENCE" for r in anchors))
        for row in anchors:
            self.assertIn(row["occurrence_id"], occurrence_ids)
            self.assertEqual(row["occurrence_id"], source[row["canonical_label_id"]]["absolute_token_positions"])
            self.assertEqual(row["raw_token"], source[row["canonical_label_id"]]["raw_token_sequence"])

    def test_models_nulls_and_loao(self):
        models = rows(OUT / "STRUCTURAL_MODEL_RESULTS.tsv")
        self.assertEqual(len(models), 10)
        self.assertEqual(sum(r["gate_s1_candidate"] == "PRIMARY" for r in models), 1)
        self.assertEqual(models[0]["loao_exact"], "4")
        self.assertEqual(models[0]["loao_top3"], "8")
        nulls = rows(OUT / "NULL_MODEL_RESULTS.tsv")
        self.assertEqual(len(nulls), 40)
        self.assertTrue({"ANCHOR_RANK_PERMUTATION", "SPATIAL_RANK_PERMUTATION", "ROTATION_PRESERVING_SHIFT", "RING_PRESERVING_PERMUTATION"} <= {r["null_type"] for r in nulls})
        self.assertTrue(all(r["replicates"] == "10000" for r in nulls if r["null_type"] != "RING_PRESERVING_PERMUTATION"))
        loao = rows(OUT / "LEAVE_ONE_ANCHOR_OUT_RESULTS.tsv")
        blind = rows(OUT / "LEAVE_ONE_ANCHOR_OUT_PREDICTIONS.tsv")
        self.assertEqual(len(loao), len(blind), 21)
        self.assertFalse(any("true_" in key or "occurrence_id" in key for key in blind[0]))
        self.assertTrue(all(r["answers_joined_after_prediction"] == "YES" for r in loao))
        self.assertTrue(all(r["ad_hoc_exception"] == "NO" for r in loao))

    def test_gate_and_no_production_alignment(self):
        report = (OUT / "GLOBAL_ALIGNMENT_REPORT.md").read_text(encoding="utf-8")
        self.assertIn("GLOBAL_STRUCTURAL_ALIGNMENT_STATUS=BLOCKED", report)
        self.assertIn("STRUCTURAL_ALIGNMENT_F68R1_AUTHORIZED=NO", report)
        self.assertIn("STRUCTURAL_ALIGNMENT_F68R2_AUTHORIZED=NO", report)
        self.assertIn("STRUCTURAL_ALIGNMENT_F68R3_AUTHORIZED=NO", report)
        self.assertFalse((OUT / "F68R1_GLOBAL_ALIGNMENT.tsv").exists())
        self.assertFalse((OUT / "F68R2_GLOBAL_ALIGNMENT.tsv").exists())
        self.assertFalse((OUT / "F68R3_GLOBAL_ALIGNMENT.tsv").exists())
        self.assertEqual(Counter(r["mapping_origin"] for r in rows(OUT / "STRUCTURAL_LABEL_TOKEN_MAPPING.tsv")), Counter({"LEGACY_CONFIRMED": 21, "UNRESOLVED": 71}))
        self.assertTrue(all(r["status"] == "UNRESOLVED_GATE_S1_FAILED" for r in rows(OUT / "UNRESOLVED_LABELS.tsv")))

    def test_no_forbidden_analysis_or_local_ai(self):
        self.assertFalse((OUT / "LOCAL_AI_ADJUDICATION.tsv").exists())
        self.assertIn("LOCAL_AI_ADJUDICATION_USED=NO", (OUT / "GLOBAL_ALIGNMENT_REPORT.md").read_text(encoding="utf-8"))
        forbidden = ("HAPAX_ENRICHMENT", "LEXICON_MATCH_RESULTS", "CORPUS_FREQUENCY")
        names = {p.name for p in OUT.rglob("*") if p.is_file()}
        self.assertFalse(any(any(term in name for term in forbidden) for name in names))
        self.assertFalse(any("group_id" in key.lower() or "star_count" in key.lower() for r in rows(OUT / "SPATIAL_LABEL_FEATURES.tsv") for key in r))

    def test_visuals_do_not_use_hapax_fields(self):
        self.assertTrue((OUT / "visualizations/f68r1_spatial_order_overlay.png").exists())
        self.assertTrue((OUT / "visualizations/f68r1_anchor_rank_scatter.png").exists())
        for path in (OUT / "visualizations").iterdir():
            self.assertNotIn("hapax", path.name.lower())

    def test_reproducible_deterministic_build(self):
        script = OUT / "scripts/build_alignment.py"
        spec = importlib.util.spec_from_file_location("build_alignment_rerun", script)
        module = importlib.util.module_from_spec(spec)
        assert spec.loader is not None
        spec.loader.exec_module(module)
        with tempfile.TemporaryDirectory() as temp:
            temp_out = Path(temp) / "global_structural_alignment_v1"
            temp_out.mkdir()
            for name in ("GLOBAL_ALIGNMENT_PROTOCOL.md", "TRANSCRIPTION_LAYOUT_CONVENTIONS.md"):
                shutil.copyfile(OUT / name, temp_out / name)
            module.OUT = temp_out
            module.main()
            for relative in (
                "INPUT_MANIFEST.tsv", "SPATIAL_LABEL_FEATURES.tsv", "TRANSCRIPTION_SEQUENCE_REGISTRY.tsv",
                "LEGACY_ANCHORS.tsv", "STRUCTURAL_MODEL_REGISTRY.tsv", "STRUCTURAL_MODEL_RESULTS.tsv",
                "NULL_MODEL_RESULTS.tsv", "LEAVE_ONE_ANCHOR_OUT_RESULTS.tsv", "LEAVE_ONE_ANCHOR_OUT_PREDICTIONS.tsv",
                "ALIGNMENT_STABILITY_RESULTS.tsv", "STRUCTURAL_LABEL_TOKEN_MAPPING.tsv", "UNRESOLVED_LABELS.tsv",
                "MAPPING_COVERAGE_SUMMARY.tsv", "F68R1_ALIGNMENT_REPORT.md", "GLOBAL_ALIGNMENT_REPORT.md",
            ):
                self.assertEqual(sha(OUT / relative), sha(temp_out / relative), relative)

    def test_sha_ledger(self):
        for line in (OUT / "SHA256SUMS").read_text(encoding="utf-8").splitlines():
            expected, relative = line.split(None, 1)
            self.assertEqual(expected, sha(OUT / relative.strip()), relative)


if __name__ == "__main__":
    unittest.main()
