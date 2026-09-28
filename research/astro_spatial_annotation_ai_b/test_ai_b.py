#!/usr/bin/env python3
import csv
import json
import hashlib
import unittest
from pathlib import Path

DIR = Path(__file__).resolve().parent
PACKAGE = DIR / "package"
CROPS = PACKAGE / "crops"
SCHEMAS = PACKAGE / "schemas"

REQUIRED_FILES = [
    "AI_B_INPUT_MANIFEST.json",
    "AI_B_OBJECTS.tsv",
    "AI_B_LABELS.tsv",
    "AI_B_LABEL_OBJECT_RELATIONS.tsv",
    "AI_B_PRIMARY_PASS_REPORT.md",
    "AI_B_A_MATCH_OBJECTS.tsv",
    "AI_B_A_MATCH_LABELS.tsv",
    "AI_B_A_RELATION_AGREEMENT.tsv",
    "AI_B_A_DISAGREEMENTS.tsv",
    "AI_B_AGREEMENT_SUMMARY.md",
    "AI_B_STRUCTURAL_SENSITIVITY.tsv",
    "AI_B_MANIFEST.json",
    "SHA256SUMS"
]

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

ALLOWED_CLASSES = {
    "STAR_OBJECT", "CIRCLE", "RADIAL_LINE", "SECTOR", "CENTRAL_OBJECT",
    "MOON_OR_DISC_OBJECT", "TEXT_ARC", "OTHER_DIAGRAM_OBJECT"
}

ALLOWED_RELATIONS = {
    "NEAREST_OBJECT", "ADJACENT_TO", "INSIDE_OBJECT", "ON_OBJECT",
    "BETWEEN_OBJECTS", "SECTOR_LABEL", "RING_LABEL", "UNASSIGNED"
}

class TestAIBPass(unittest.TestCase):
    def test_required_files_exist(self):
        for fname in REQUIRED_FILES:
            fpath = DIR / fname
            self.assertTrue(fpath.exists(), f"Missing required file: {fname}")

    def test_package_structure(self):
        self.assertTrue((PACKAGE / "ANNOTATOR_B_AI_INSTRUCTIONS.md").exists())
        self.assertTrue((PACKAGE / "OBJECT_ONTOLOGY_CLEAN.md").exists())
        self.assertTrue((PACKAGE / "AI_B_INPUT_MANIFEST.json").exists())
        for p in PANELS:
            self.assertTrue((CROPS / f"{p}.jpg").exists(), f"Missing crop for {p}")

    def test_sha256sums_validity(self):
        sha_file = DIR / "SHA256SUMS"
        lines = sha_file.read_text(encoding="utf-8").strip().splitlines()
        for line in lines:
            if not line.strip(): continue
            parts = line.strip().split()
            self.assertEqual(len(parts), 2)
            expected_hash, rel_path = parts
            target_path = DIR / rel_path
            self.assertTrue(target_path.exists(), f"Hashed file does not exist: {rel_path}")
            computed_hash = hashlib.sha256(target_path.read_bytes()).hexdigest()
            self.assertEqual(expected_hash, computed_hash, f"Hash mismatch for {rel_path}")

    def test_objects_validity(self):
        with (DIR / "AI_B_OBJECTS.tsv").open(encoding="utf-8") as f:
            reader = list(csv.DictReader(f, delimiter="\t"))
        self.assertGreater(len(reader), 0)
        seen_panels = set()
        for r in reader:
            self.assertIn(r["panel"], PANELS)
            seen_panels.add(r["panel"])
            self.assertIn(r["object_class"], ALLOWED_CLASSES)
            w, h = PANEL_DIMS[r["panel"]]
            x1, y1, x2, y2 = float(r["bbox_x1"]), float(r["bbox_y1"]), float(r["bbox_x2"]), float(r["bbox_y2"])
            self.assertLess(x1, x2)
            self.assertLess(y1, y2)
            self.assertGreaterEqual(x1, 0.0)
            self.assertGreaterEqual(y1, 0.0)
            self.assertLessEqual(x2, float(w))
            self.assertLessEqual(y2, float(h))
            cx_n, cy_n = float(r["center_x_norm"]), float(r["center_y_norm"])
            self.assertTrue(0.0 <= cx_n <= 1.0)
            self.assertTrue(0.0 <= cy_n <= 1.0)
        self.assertEqual(seen_panels, set(PANELS), "Not all 8 panels are represented in AI_B_OBJECTS.tsv")

    def test_labels_validity(self):
        with (DIR / "AI_B_LABELS.tsv").open(encoding="utf-8") as f:
            reader = list(csv.DictReader(f, delimiter="\t"))
        self.assertGreater(len(reader), 0)
        seen_panels = set()
        for r in reader:
            self.assertIn(r["panel"], PANELS)
            seen_panels.add(r["panel"])
            w, h = PANEL_DIMS[r["panel"]]
            x1, y1, x2, y2 = float(r["bbox_x1"]), float(r["bbox_y1"]), float(r["bbox_x2"]), float(r["bbox_y2"])
            self.assertLess(x1, x2)
            self.assertLess(y1, y2)
            cx_n, cy_n = float(r["center_x_norm"]), float(r["center_y_norm"])
            self.assertTrue(0.0 <= cx_n <= 1.0)
            self.assertTrue(0.0 <= cy_n <= 1.0)
        self.assertEqual(seen_panels, set(PANELS), "Not all 8 panels are represented in AI_B_LABELS.tsv")

    def test_relations_validity(self):
        with (DIR / "AI_B_LABEL_OBJECT_RELATIONS.tsv").open(encoding="utf-8") as f:
            reader = list(csv.DictReader(f, delimiter="\t"))
        self.assertGreater(len(reader), 0)
        for r in reader:
            self.assertIn(r["relation_type"], ALLOWED_RELATIONS)
            self.assertIn(r["confidence"], {"HIGH", "MEDIUM", "LOW", "AMBIGUOUS"})

    def test_agreement_summary_contents(self):
        summary_text = (DIR / "AI_B_AGREEMENT_SUMMARY.md").read_text(encoding="utf-8")
        self.assertIn("ANNOTATOR_B_AI_PASS=COMPLETE", summary_text)
        self.assertIn("ANNOTATOR_B_TYPE=AI", summary_text)
        self.assertIn("ANNOTATOR_B_INDEPENDENT_FROM_A=YES", summary_text)
        self.assertIn("ANNOTATOR_B_HUMAN_EQUIVALENT=NO", summary_text)
        self.assertIn("PRODUCTION_SPATIAL_GATE=STILL_BLOCKED", summary_text)
        self.assertIn("HUMAN_ANNOTATOR_B_STILL_REQUIRED=YES", summary_text)

if __name__ == "__main__":
    unittest.main()
