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
AI1_DIR = DIR.parent / "astro_spatial_annotation_ai_b"
A_DIR = DIR.parent / "astro_spatial_annotation"

REQUIRED_FILES = [
    "AI2_INPUT_MANIFEST.json",
    "AI2_OBJECTS.tsv",
    "AI2_LABELS.tsv",
    "AI2_LABEL_OBJECT_RELATIONS.tsv",
    "AI2_PRIMARY_PASS_REPORT.md",
    "AI2_STAGE1_MANIFEST.json",
    "AI2_STAGE2_MANIFEST.json",
    "AI1_AI2_OBJECT_MATCH.tsv",
    "AI1_AI2_LABEL_MATCH.tsv",
    "AI1_AI2_RELATION_AGREEMENT.tsv",
    "A_AI2_OBJECT_MATCH.tsv",
    "A_AI2_LABEL_MATCH.tsv",
    "THREE_WAY_OBJECT_SUPPORT.tsv",
    "THREE_WAY_LABEL_SUPPORT.tsv",
    "THREE_WAY_DISAGREEMENTS.tsv",
    "AI_CONSENSUS_STAR_DATASET.tsv",
    "AI_CONSENSUS_LABEL_DATASET.tsv",
    "AI2_STRUCTURAL_SENSITIVITY.tsv",
    "AI2_AGREEMENT_SUMMARY.md",
    "AI2_MANIFEST.json",
    "SHA256SUMS",
]

PANELS = ["f67r1", "f67r2", "f67v1", "f68r1", "f68r2", "f68r3", "f68v1", "f68v2"]
PANEL_DIMS = {
    "f67r1": (2486, 3738), "f67r2": (2486, 3738), "f67v1": (2565, 3753),
    "f68r1": (2462, 3828), "f68r2": (2078, 3828), "f68r3": (3453, 3828),
    "f68v1": (2530, 3843), "f68v2": (2083, 3843),
}
ALLOWED_CLASSES = {
    "STAR_OBJECT", "CIRCLE", "RADIAL_LINE", "SECTOR", "CENTRAL_OBJECT",
    "MOON_OR_DISC_OBJECT", "TEXT_ARC", "OTHER_DIAGRAM_OBJECT",
}
ALLOWED_RELATIONS = {
    "NEAREST_OBJECT", "ADJACENT_TO", "INSIDE_OBJECT", "ON_OBJECT",
    "BETWEEN_OBJECTS", "SECTOR_LABEL", "RING_LABEL", "UNASSIGNED",
}
ALLOWED_CONFIDENCE = {"HIGH", "MEDIUM", "LOW", "AMBIGUOUS"}


def read_tsv(path):
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


class TestAI2Pass(unittest.TestCase):
    def test_required_files_exist(self):
        for fname in REQUIRED_FILES:
            self.assertTrue((DIR / fname).exists(), f"Missing required file: {fname}")

    def test_package_structure_and_bytes_match_ai1(self):
        self.assertTrue((PACKAGE / "ANNOTATOR_B_AI_2_INSTRUCTIONS.md").exists())
        self.assertTrue((PACKAGE / "OBJECT_ONTOLOGY_CLEAN.md").exists())
        for p in PANELS:
            crop = CROPS / f"{p}.jpg"
            self.assertTrue(crop.exists())
            ai1_crop = AI1_DIR / "package" / "crops" / f"{p}.jpg"
            self.assertEqual(
                hashlib.sha256(crop.read_bytes()).hexdigest(),
                hashlib.sha256(ai1_crop.read_bytes()).hexdigest(),
                f"AI2 crop for {p} must be byte-identical to AI1 crop",
            )

    def test_input_manifest_declares_independence_fields(self):
        manifest = json.loads((DIR / "AI2_INPUT_MANIFEST.json").read_text())
        for key in ("annotator_id", "model_name", "model_version_if_known", "provider"):
            self.assertIn(key, manifest)
        self.assertEqual(manifest["annotator_id"], "ANNOTATOR_B_AI_2")

    def test_objects_schema_and_values(self):
        rows = read_tsv(DIR / "AI2_OBJECTS.tsv")
        self.assertGreater(len(rows), 0)
        seen_ids = set()
        for r in rows:
            self.assertIn(r["panel"], PANELS)
            self.assertIn(r["object_class"], ALLOWED_CLASSES)
            self.assertIn(r["confidence"], ALLOWED_CONFIDENCE)
            self.assertNotIn(r["object_id"], seen_ids)
            seen_ids.add(r["object_id"])
            x1, y1, x2, y2 = (float(r[k]) for k in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"))
            w, h = PANEL_DIMS[r["panel"]]
            self.assertLess(x1, x2)
            self.assertLess(y1, y2)
            self.assertGreaterEqual(x1, 0)
            self.assertGreaterEqual(y1, 0)
            self.assertLessEqual(x2, w + 1)
            self.assertLessEqual(y2, h + 1)

    def test_labels_schema_and_anonymous_ids(self):
        rows = read_tsv(DIR / "AI2_LABELS.tsv")
        self.assertGreater(len(rows), 0)
        for r in rows:
            self.assertTrue(r["label_id"].startswith("C_LABEL_"), r["label_id"])
            self.assertIn(r["panel"], PANELS)
            self.assertIn(r["confidence"], ALLOWED_CONFIDENCE)
            for col in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "center_x", "center_y",
                        "center_x_norm", "center_y_norm", "orientation_angle"):
                float(r[col])
            self.assertNotIn("eva", r.get("notes", "").lower())

    def test_relations_reference_known_ids(self):
        objs = {r["object_id"] for r in read_tsv(DIR / "AI2_OBJECTS.tsv")}
        lbls = {r["label_id"] for r in read_tsv(DIR / "AI2_LABELS.tsv")}
        rels = read_tsv(DIR / "AI2_LABEL_OBJECT_RELATIONS.tsv")
        self.assertGreater(len(rels), 0)
        for r in rels:
            self.assertIn(r["relation_type"], ALLOWED_RELATIONS)
            self.assertIn(r["label_id"], lbls)
            self.assertIn(r["object_id"], objs)

    def test_stage_freeze_manifests_hash_match(self):
        s1 = json.loads((DIR / "AI2_STAGE1_MANIFEST.json").read_text())
        self.assertTrue(s1.get("stage1_frozen"))
        for fname, meta in s1["files"].items():
            digest = hashlib.sha256((DIR / fname).read_bytes()).hexdigest()
            self.assertEqual(digest, meta["sha256"], f"{fname} does not match its Stage 1 freeze hash")
        s2 = json.loads((DIR / "AI2_STAGE2_MANIFEST.json").read_text())
        self.assertTrue(s2.get("stage2_frozen"))
        for fname, meta in s2["files"].items():
            digest = hashlib.sha256((DIR / fname).read_bytes()).hexdigest()
            self.assertEqual(digest, meta["sha256"], f"{fname} does not match its Stage 2 freeze hash")

    def test_three_way_support_statuses_are_valid(self):
        allowed = {
            "SUPPORT_3_OF_3", "SUPPORT_2_OF_3", "SUPPORT_AI1_AI2_ONLY",
            "SUPPORT_A_ONLY", "SUPPORT_AI1_ONLY", "SUPPORT_AI2_ONLY",
        }
        for fname in ("THREE_WAY_OBJECT_SUPPORT.tsv", "THREE_WAY_LABEL_SUPPORT.tsv"):
            rows = read_tsv(DIR / fname)
            self.assertGreater(len(rows), 0)
            for r in rows:
                self.assertIn(r["support_status"], allowed)

    def test_ai1_ai2_only_not_treated_as_conflict_where_a_incomplete(self):
        # On panels other than f68r1, ANNOTATOR_A never annotated stars, so
        # SUPPORT_AI1_AI2_ONLY there must not appear as a THREE_WAY disagreement.
        disagreements = read_tsv(DIR / "THREE_WAY_DISAGREEMENTS.tsv")
        for r in disagreements:
            if r["panel"] != "f68r1":
                self.assertNotEqual(
                    r["disagreement_type"], "A_CONFLICT_WITH_AI_CONSENSUS",
                    f"A-conflict flagged on {r['panel']} where A coverage is known incomplete",
                )

    def test_manifest_sha256sums_consistent(self):
        manifest = json.loads((DIR / "AI2_MANIFEST.json").read_text())
        self.assertEqual(manifest["ANNOTATOR_ID"], "ANNOTATOR_B_AI_2")
        self.assertEqual(manifest["INDEPENDENT_FROM_A"], "YES")
        self.assertEqual(manifest["INDEPENDENT_FROM_AI_B_1"], "YES")
        sha_lines = (DIR / "SHA256SUMS").read_text().strip().splitlines()
        by_name = {}
        for line in sha_lines:
            digest, fname = line.split("  ", 1)
            by_name[fname] = digest
        for fname, meta in manifest["files"].items():
            digest = hashlib.sha256((DIR / fname).read_bytes()).hexdigest()
            self.assertEqual(digest, meta["sha256"], f"{fname} content changed since manifest was written")
            self.assertEqual(by_name.get(fname), digest, f"{fname} missing/mismatched in SHA256SUMS")

    def test_no_leaked_prior_pass_vocabulary(self):
        forbidden = ["B_LABEL_", "B_OBJ_", "ASTRO_OBJECTS", "ASTRO_LABELS_SPATIAL", "EVA", "Stolfi"]
        for fname in ("AI2_OBJECTS.tsv", "AI2_LABELS.tsv", "AI2_PRIMARY_PASS_REPORT.md"):
            text = (DIR / fname).read_text(encoding="utf-8")
            for term in forbidden:
                self.assertNotIn(term, text, f"{fname} leaks prior-pass vocabulary: {term}")


if __name__ == "__main__":
    unittest.main()
