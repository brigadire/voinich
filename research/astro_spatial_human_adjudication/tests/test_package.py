from __future__ import annotations

import csv
import hashlib
import json
import subprocess
import sys
import tempfile
import unittest
import xml.etree.ElementTree as ET
from pathlib import Path

PKG = Path(__file__).resolve().parents[1]
ROOT = PKG.parents[1]
sys.path.insert(0, str(PKG / "scripts"))
import export_to_cvat  # noqa: E402


def read(path):
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


class PackageTests(unittest.TestCase):
    def test_cvat_schema_attribute_values_are_nonempty(self):
        for filename in ("CVAT_STAR_LABEL_SCHEMA.json", "CVAT_TEXT_LABEL_SCHEMA.json"):
            labels = json.loads((PKG / filename).read_text())
            for label in labels:
                for attribute in label.get("attributes", []):
                    self.assertIsInstance(attribute.get("values"), list)
                    self.assertGreater(len(attribute["values"]), 0, f"{filename}: {label['name']}.{attribute['name']}")
                    self.assertIn("default_value", attribute)
                    if attribute["input_type"] == "select":
                        self.assertIn(attribute["default_value"], attribute["values"])

    def test_canonical_images_match_both_manifests(self):
        m1 = json.loads((ROOT / "research/astro_spatial_annotation_ai_b/AI_B_INPUT_MANIFEST.json").read_text())["files"]
        m2 = json.loads((ROOT / "research/astro_spatial_annotation_ai_b2/AI2_INPUT_MANIFEST.json").read_text())["files"]
        for image in sorted((PKG / "images").glob("*.jpg")):
            key = f"crops/{image.name}"
            actual = hashlib.sha256(image.read_bytes()).hexdigest()
            self.assertEqual(actual, m1[key]["sha256"])
            self.assertEqual(actual, m2[key]["sha256"])

    def test_ids_are_stable_unique_and_sources_preserved(self):
        before = {p.name: p.read_bytes() for p in (PKG / "HUMAN_CANDIDATE_OBJECTS.tsv", PKG / "HUMAN_CANDIDATE_LABELS.tsv")}
        subprocess.run([sys.executable, str(PKG / "scripts/build_human_candidate_layer.py")], check=True)
        for path in (PKG / "HUMAN_CANDIDATE_OBJECTS.tsv", PKG / "HUMAN_CANDIDATE_LABELS.tsv"):
            self.assertEqual(path.read_bytes(), before[path.name])
            rows = read(path)
            self.assertEqual(len(rows), len({r["candidate_id"] for r in rows}))
            self.assertTrue(all(r["source_annotation_ids"] for r in rows))
            for row in rows:
                tags = [part.split(":", 1)[0] for part in row["source_annotation_ids"].split(";")]
                self.assertEqual(len(tags), len(set(tags)), row["candidate_id"])

    def test_every_primary_annotation_is_preserved_once(self):
        specs = [
            (ROOT / "research/astro_spatial_annotation/ASTRO_OBJECTS.tsv", "object_id", "A", PKG / "HUMAN_CANDIDATE_OBJECTS.tsv"),
            (ROOT / "research/astro_spatial_annotation_ai_b/AI_B_OBJECTS.tsv", "object_id", "AI1", PKG / "HUMAN_CANDIDATE_OBJECTS.tsv"),
            (ROOT / "research/astro_spatial_annotation_ai_b2/AI2_OBJECTS.tsv", "object_id", "AI2", PKG / "HUMAN_CANDIDATE_OBJECTS.tsv"),
            (ROOT / "research/astro_spatial_annotation/ASTRO_LABELS_SPATIAL.tsv", "label_occurrence_id", "A", PKG / "HUMAN_CANDIDATE_LABELS.tsv"),
            (ROOT / "research/astro_spatial_annotation_ai_b/AI_B_LABELS.tsv", "label_id", "AI1", PKG / "HUMAN_CANDIDATE_LABELS.tsv"),
            (ROOT / "research/astro_spatial_annotation_ai_b2/AI2_LABELS.tsv", "label_id", "AI2", PKG / "HUMAN_CANDIDATE_LABELS.tsv"),
        ]
        for source, id_field, tag, candidate_file in specs:
            expected = {f"{tag}:{r[id_field]}" for r in read(source)}
            observed = [token for row in read(candidate_file) for token in row["source_annotation_ids"].split(";") if token.startswith(f"{tag}:")]
            self.assertEqual(expected, set(observed))
            self.assertEqual(len(observed), len(set(observed)))

    def test_frozen_source_checksums(self):
        for directory, names in (
            (ROOT / "research/astro_spatial_annotation", {"ASTRO_OBJECTS.tsv", "ASTRO_LABELS_SPATIAL.tsv", "ASTRO_LABEL_OBJECT_RELATIONS.tsv"}),
            (ROOT / "research/astro_spatial_annotation_ai_b", {"AI_B_OBJECTS.tsv", "AI_B_LABELS.tsv", "AI_B_LABEL_OBJECT_RELATIONS.tsv"}),
            (ROOT / "research/astro_spatial_annotation_ai_b2", {"AI2_OBJECTS.tsv", "AI2_LABELS.tsv", "AI2_LABEL_OBJECT_RELATIONS.tsv"}),
        ):
            frozen = {}
            for line in (directory / "SHA256SUMS").read_text().splitlines():
                digest, name = line.split(None, 1)
                frozen[name.strip()] = digest
            for name in names:
                self.assertEqual(hashlib.sha256((directory / name).read_bytes()).hexdigest(), frozen[name])

    def test_no_sensitive_fields_in_review_ui(self):
        forbidden_fields = {"token_transcription", "stolfi", "zl3b", "eva", "model_name", "ai1", "ai2", "annotator"}
        for task in ("star", "label"):
            root = export_to_cvat.build(task, "all", "H1").getroot()
            text = ET.tostring(root, encoding="unicode").lower()
            self.assertFalse(any(term in text for term in forbidden_fields))
            attrs = {a.attrib["name"].lower() for a in root.findall(".//*[@label]/attribute")}
            self.assertTrue(forbidden_fields.isdisjoint(attrs))
            self.assertNotIn("candidate_support", attrs)
            priorities = [a.text for a in root.findall(".//*[@label]/attribute[@name='priority']")]
            self.assertTrue(priorities and set(priorities) == {"UNSET"})

    def test_cvat_xml_declares_the_required_task_labels(self):
        expected = {"star": {"STAR_OBJECT", "OTHER_OBJECT"}, "label": {"LABEL"}}
        for task, names in expected.items():
            root = export_to_cvat.build(task, "calibration", "H1").getroot()
            declared = {node.text for node in root.findall("./meta/task/labels/label/name")}
            used = {node.attrib["label"] for node in root.findall(".//*[@label]")}
            self.assertEqual(declared, names)
            self.assertTrue(used <= declared)

    def test_label_calibration_uses_oriented_and_ring_geometry(self):
        root = export_to_cvat.build("label", "calibration", "H1").getroot()
        rotated = [box for box in root.findall(".//box") if abs(float(box.attrib.get("rotation", "0"))) > 0.01]
        self.assertGreater(len(rotated), 0)
        self.assertGreater(len(root.findall(".//ellipse")), 0)
        followup = read(PKG / "HUMAN_LABEL_CALIBRATION_FOLLOWUP.tsv")
        self.assertTrue(all(r["geometry_mode"] == "ELLIPSE_RING" for r in followup))
        self.assertEqual(len(followup), len(root.findall(".//ellipse")))
        self.assertEqual(len(followup), 3)

    def test_cvat_round_trip_coordinates(self):
        with tempfile.TemporaryDirectory() as directory:
            xml = Path(directory) / "in.xml"
            export_to_cvat.build("star", "calibration", "H1").write(xml, encoding="utf-8", xml_declaration=True)
            out = Path(directory) / "out.tsv"
            subprocess.run([sys.executable, str(PKG / "scripts/import_from_cvat.py"), "--task", "star", "--input", str(xml), "--output", str(out), "--reviewer-id", "TEST", "--timestamp", "2000-01-01T00:00:00Z"], check=True)
            source = {r["candidate_id"]: r for r in read(PKG / "HUMAN_CANDIDATE_OBJECTS.tsv")}
            imported = read(out)
            self.assertGreater(len(imported), 0)
            for row in imported:
                for key in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"):
                    self.assertLessEqual(abs(float(row[key]) - float(source[row["candidate_id"]][key])), 0.001)

    def test_label_cvat_round_trip_mixed_geometry(self):
        with tempfile.TemporaryDirectory() as directory:
            xml = Path(directory) / "in.xml"
            export_to_cvat.build("label", "calibration", "H1").write(xml, encoding="utf-8", xml_declaration=True)
            out = Path(directory) / "out.tsv"
            subprocess.run([sys.executable, str(PKG / "scripts/import_from_cvat.py"), "--task", "label", "--input", str(xml), "--output", str(out), "--reviewer-id", "TEST", "--timestamp", "2000-01-01T00:00:00Z"], check=True)
            source = {r["candidate_id"]: r for r in read(PKG / "HUMAN_CANDIDATE_LABELS.tsv")}
            imported = read(out)
            self.assertEqual({r["geometry_type"] for r in imported}, {"BOX", "ELLIPSE"})
            self.assertTrue(all(r["uncertain"] == ("true" if r["human_decision"] == "UNCERTAIN" else "false") for r in imported))
            for row in imported:
                for key in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2"):
                    self.assertLessEqual(abs(float(row[key]) - float(source[row["candidate_id"]][key])), 0.001)

    def test_relation_references_and_calibration_panels(self):
        objects = {r["candidate_id"] for r in read(PKG / "HUMAN_CANDIDATE_OBJECTS.tsv")}
        labels = {r["candidate_id"] for r in read(PKG / "HUMAN_CANDIDATE_LABELS.tsv")}
        for row in read(PKG / "HUMAN_CANDIDATE_RELATIONS.tsv"):
            self.assertIn(row["object_candidate_id"], objects)
            self.assertIn(row["label_candidate_id"], labels)
        self.assertEqual({r["panel"] for r in read(PKG / "HUMAN_CALIBRATION_SET.tsv")}, {"f68r1", "f68r3", "f68v2"})

    def test_completed_human_calibration_outputs(self):
        stars = read(PKG / "exports/HUMAN_STAR_CALIBRATION_R01.tsv")
        labels = read(PKG / "exports/HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv")
        followup = read(PKG / "exports/HUMAN_LABEL_CALIBRATION_RING_FOLLOWUP_R01.tsv")
        self.assertEqual(len(stars), 207)
        self.assertEqual(len(labels), 93)
        self.assertEqual(len(followup), 3)
        self.assertTrue(all(r["human_confidence"] == "HIGH" for r in stars + labels + followup))
        self.assertTrue(all(r["human_decision"] == "ACCEPT" for r in followup))
        self.assertEqual({r["candidate_id"] for r in followup} & {r["candidate_id"] for r in labels}, {r["candidate_id"] for r in followup})

    def test_completed_human_high_outputs(self):
        stars = read(PKG / "exports/HUMAN_STAR_HIGH_R01_FINAL.tsv")
        labels = read(PKG / "exports/HUMAN_LABEL_HIGH_R01_FINAL.tsv")
        star_calibration = {r["candidate_id"]: r for r in read(PKG / "exports/HUMAN_STAR_CALIBRATION_R01.tsv")}
        label_calibration = {r["candidate_id"]: r for r in read(PKG / "exports/HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv")}
        star_candidates = {
            r["candidate_id"]: r for r in read(PKG / "HUMAN_CANDIDATE_OBJECTS.tsv")
            if r["candidate_type"] == "STAR_OBJECT" and r["priority"] == "HIGH"
        }
        label_candidates = {
            r["candidate_id"]: r for r in read(PKG / "HUMAN_CANDIDATE_LABELS.tsv") if r["priority"] == "HIGH"
        }
        self.assertEqual({r["candidate_id"] for r in stars}, set(star_candidates))
        self.assertEqual({r["candidate_id"] for r in labels}, set(label_candidates))
        self.assertEqual(len(stars), 385)
        self.assertEqual(len(labels), 271)
        self.assertTrue(all(r["human_confidence"] == "HIGH" for r in stars + labels))

        stars_by_id = {r["candidate_id"]: r for r in stars}
        labels_by_id = {r["candidate_id"]: r for r in labels}
        for candidate_id in set(stars_by_id) & set(star_calibration):
            self.assertEqual(stars_by_id[candidate_id], star_calibration[candidate_id])
        for candidate_id in set(labels_by_id) & set(label_calibration):
            self.assertEqual(labels_by_id[candidate_id], label_calibration[candidate_id])

        resized = stars_by_id["HOBJ_f67r1_3C48B2BAAA37"]
        self.assertEqual(resized["human_decision"], "MODIFY")
        self.assertNotEqual(resized["bbox_x1"], star_candidates[resized["candidate_id"]]["bbox_x1"])
        ring = labels_by_id["HLABEL_f68v1_3B985EFBA4D9"]
        self.assertEqual(ring["human_decision"], "ACCEPT")
        for field in ("bbox_x1", "bbox_y1", "bbox_x2", "bbox_y2", "center_x", "center_y"):
            self.assertEqual(ring[field], label_candidates[ring["candidate_id"]][field])
        edge_ids = {
            "HOBJ_f67v1_7D4D96035479",
            "HOBJ_f67v1_AC0352A0A7AB",
            "HOBJ_f67v1_D3A2C03DFB8C",
        }
        self.assertTrue(all(stars_by_id[candidate_id]["human_decision"] == "ACCEPT" for candidate_id in edge_ids))

    def test_pending_consensus_qc_excludes_completed_reviews(self):
        cases = (
            (
                "star",
                PKG / "cvat/star_consensus_qc_pending_h1.xml",
                PKG / "exports/HUMAN_STAR_HIGH_R01_FINAL.tsv",
                PKG / "exports/HUMAN_STAR_CALIBRATION_R01.tsv",
                7,
            ),
            (
                "label",
                PKG / "cvat/label_consensus_qc_pending_h1.xml",
                PKG / "exports/HUMAN_LABEL_HIGH_R01_FINAL.tsv",
                PKG / "exports/HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv",
                2,
            ),
        )
        for _, xml_path, high_path, calibration_path, expected_count in cases:
            root = ET.parse(xml_path).getroot()
            pending = {
                node.text for node in root.findall(".//attribute[@name='candidate_id']")
            }
            reviewed = {r["candidate_id"] for r in read(high_path)} | {
                r["candidate_id"] for r in read(calibration_path)
            }
            self.assertEqual(len(pending), expected_count)
            self.assertTrue(pending.isdisjoint(reviewed))

    def test_high_finalization_is_reproducible(self):
        cases = (
            (
                "star",
                PKG / "exports/HUMAN_STAR_HIGH_R01_AS_EXPORTED.tsv",
                PKG / "exports/HUMAN_STAR_CALIBRATION_R01.tsv",
                PKG / "HUMAN_CANDIDATE_OBJECTS.tsv",
                PKG / "exports/HUMAN_STAR_HIGH_R01_FINAL.tsv",
            ),
            (
                "label",
                PKG / "exports/HUMAN_LABEL_HIGH_R01_AS_EXPORTED.tsv",
                PKG / "exports/HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv",
                PKG / "HUMAN_CANDIDATE_LABELS.tsv",
                PKG / "exports/HUMAN_LABEL_HIGH_R01_FINAL.tsv",
            ),
        )
        with tempfile.TemporaryDirectory() as directory:
            for task, review, calibration, candidates, expected in cases:
                output = Path(directory) / f"{task}.tsv"
                subprocess.run([
                    sys.executable,
                    str(PKG / "scripts/finalize_high_review.py"),
                    "--task", task,
                    "--review", str(review),
                    "--calibration", str(calibration),
                    "--candidates", str(candidates),
                    "--overrides", str(PKG / "exports/HUMAN_HIGH_REVIEW_OVERRIDES_R01.tsv"),
                    "--output", str(output),
                ], check=True, stdout=subprocess.DEVNULL)
                self.assertEqual(output.read_bytes(), expected.read_bytes())

    def test_completed_consensus_qc_and_pending_medium(self):
        cases = (
            (
                "star",
                PKG / "exports/HUMAN_STAR_CONSENSUS_QC_R01_AS_EXPORTED.tsv",
                PKG / "exports/HUMAN_STAR_CONSENSUS_QC_R01_FINAL.tsv",
                PKG / "cvat/star_consensus_qc_pending_h1.xml",
                PKG / "HUMAN_CANDIDATE_OBJECTS.tsv",
                PKG / "cvat/star_medium_pending_h1.xml",
                7,
                24,
            ),
            (
                "label",
                PKG / "exports/HUMAN_LABEL_CONSENSUS_QC_R01_AS_EXPORTED.tsv",
                PKG / "exports/HUMAN_LABEL_CONSENSUS_QC_R01_FINAL.tsv",
                PKG / "cvat/label_consensus_qc_pending_h1.xml",
                PKG / "HUMAN_CANDIDATE_LABELS.tsv",
                PKG / "cvat/label_medium_pending_h1.xml",
                2,
                6,
            ),
        )
        with tempfile.TemporaryDirectory() as directory:
            for task, imported, final, expected_xml, candidates, medium_xml, qc_count, medium_count in cases:
                final_rows = read(final)
                self.assertEqual(len(final_rows), qc_count)
                self.assertTrue(all(row["human_confidence"] == "HIGH" for row in final_rows))
                output = Path(directory) / f"{task}-qc.tsv"
                subprocess.run([
                    sys.executable,
                    str(PKG / "scripts/finalize_consensus_qc.py"),
                    "--task", task,
                    "--review", str(imported),
                    "--expected-xml", str(expected_xml),
                    "--candidates", str(candidates),
                    "--output", str(output),
                ], check=True, stdout=subprocess.DEVNULL)
                self.assertEqual(output.read_bytes(), final.read_bytes())

                candidate_rows = {row["candidate_id"]: row for row in read(candidates)}
                medium_root = ET.parse(medium_xml).getroot()
                medium_ids = {
                    node.text for node in medium_root.findall(".//attribute[@name='candidate_id']")
                }
                self.assertEqual(len(medium_ids), medium_count)
                self.assertTrue(all(candidate_rows[candidate_id]["priority"] == "MEDIUM" for candidate_id in medium_ids))

        metrics = json.loads((PKG / "exports/HUMAN_CONSENSUS_QC_METRICS.json").read_text())
        self.assertEqual(metrics["STAR"]["sample_records"], 34)
        self.assertEqual(metrics["LABEL"]["sample_records"], 18)
        self.assertEqual(metrics["STAR"]["reject_rate"], 0.0)
        self.assertEqual(metrics["LABEL"]["reject_rate"], 0.0)
        self.assertEqual(metrics["STAR"]["decision_counts"], {"MODIFY": 32, "UNCERTAIN": 2})
        self.assertEqual(metrics["LABEL"]["decision_counts"], {"MODIFY": 18})

    def test_medium_imports_and_production_freeze(self):
        cases = (
            ("star", 24, "task_12_annotations_2026_09_13_10_53_04_cvat for images 1.1.zip", "2026-09-13T10:53:04Z"),
            ("label", 6, "task_13_annotations_2026_09_13_11_00_30_cvat for images 1.1.zip", "2026-09-13T11:00:30Z"),
        )
        with tempfile.TemporaryDirectory() as directory:
            for task, count, filename, stamp in cases:
                expected = PKG / f"exports/HUMAN_{task.upper()}_MEDIUM_R01.tsv"
                rows = read(expected)
                self.assertEqual(len(rows), count)
                self.assertTrue(all(row["human_decision"] == "MODIFY" and row["human_confidence"] == "HIGH" for row in rows))
                xml = ET.parse(PKG / f"cvat/{task}_medium_pending_h1.xml").getroot()
                expected_ids = {node.text for node in xml.findall(".//attribute[@name='candidate_id']")}
                self.assertEqual({row["candidate_id"] for row in rows}, expected_ids)
                output = Path(directory) / f"{task}-medium.tsv"
                subprocess.run([
                    sys.executable, str(PKG / "scripts/import_from_cvat.py"),
                    "--task", task, "--input", str(PKG / "exports" / filename),
                    "--output", str(output), "--reviewer-id", "R01",
                    "--timestamp", stamp, "--human-confidence-override", "HIGH",
                ], check=True)
                self.assertEqual(output.read_bytes(), expected.read_bytes())
            subprocess.run([
                sys.executable, str(PKG / "scripts/freeze_reviewed_layers.py"),
                "--output-dir", directory,
            ], check=True, stdout=subprocess.DEVNULL)
            for filename in ("HUMAN_STAR_ADJUDICATION.tsv", "HUMAN_LABEL_ADJUDICATION.tsv", "HUMAN_PRODUCTION_UNREVIEWED.tsv"):
                self.assertEqual((Path(directory) / filename).read_bytes(), (PKG / filename).read_bytes())

        stars = read(PKG / "HUMAN_STAR_ADJUDICATION.tsv")
        labels = read(PKG / "HUMAN_LABEL_ADJUDICATION.tsv")
        remaining = read(PKG / "HUMAN_PRODUCTION_UNREVIEWED.tsv")
        self.assertEqual(len(stars), 444)
        self.assertEqual(len(labels), 290)
        self.assertEqual(len(remaining), 8)
        self.assertTrue(all(row["priority"] == "LOW" and row["status"] == "NOT_REVIEWED" for row in remaining))
        self.assertEqual(sum(row["human_decision"] == "UNCERTAIN" for row in stars), 5)
        self.assertEqual(sum(row["human_decision"] == "UNCERTAIN" for row in labels), 3)
        for task, rows, candidate_filename in (
            ("STAR", stars, "HUMAN_CANDIDATE_OBJECTS.tsv"),
            ("LABEL", labels, "HUMAN_CANDIDATE_LABELS.tsv"),
        ):
            universe = {row["candidate_id"] for row in read(PKG / candidate_filename) if task == "LABEL" or row["candidate_type"] == "STAR_OBJECT"}
            reviewed = {row["candidate_id"] for row in rows}
            unreviewed = {row["candidate_id"] for row in remaining if row["task"] == task}
            self.assertTrue(reviewed.isdisjoint(unreviewed))
            self.assertEqual(reviewed | unreviewed, universe)

    def test_originals_match_frozen_manifest(self):
        manifest = json.loads((PKG / "manifest.json").read_text())
        for relative, expected in manifest["source_original_sha256"].items():
            self.assertEqual(hashlib.sha256((ROOT / relative).read_bytes()).hexdigest(), expected)


if __name__ == "__main__":
    unittest.main()
