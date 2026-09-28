from __future__ import annotations

import copy
import csv
import hashlib
import json
import sys
import subprocess
import tempfile
import unittest
import xml.etree.ElementTree as ET
from pathlib import Path

from PIL import Image

PKG = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(PKG / "scripts"))
import build_label_star_relations as builder
import import_label_star_relations as importer


def read(path):
    with path.open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


class RelationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.relations = PKG / "relations"
        cls.expected = ET.parse(cls.relations / "cvat/label_star_relations_h1.xml").getroot()
        cls.queue = read(cls.relations / "LABEL_STAR_RELATION_QUEUE.tsv")

    def assigned(self):
        reviewed = copy.deepcopy(self.expected)
        for shape in reviewed.findall(".//*[@label='STAR_ENDPOINT']"):
            shape.find("attribute[@name='relation_decision']").text = "ASSIGNED"
            shape.find("attribute[@name='adjacent_to']").text = "true"
        return reviewed

    def test_selection_is_confirmed_source_pair_subset(self):
        stars = {row["candidate_id"]: row for row in read(PKG / "HUMAN_STAR_ADJUDICATION.tsv")}
        labels = {row["candidate_id"]: row for row in read(PKG / "HUMAN_LABEL_ADJUDICATION.tsv")}
        source = {row["relation_candidate_id"]: row for row in read(PKG / "HUMAN_CANDIDATE_RELATIONS.tsv")}
        excluded = read(self.relations / "LABEL_STAR_RELATION_EXCLUSIONS.tsv")
        self.assertEqual(len(self.queue), 324)
        self.assertEqual(len(excluded), 351)
        selected_ids = {row["relation_candidate_id"] for row in self.queue}
        excluded_ids = {row["relation_candidate_id"] for row in excluded}
        self.assertTrue(selected_ids.isdisjoint(excluded_ids))
        self.assertEqual(selected_ids | excluded_ids, set(source))
        for row in self.queue:
            self.assertEqual(row["label_candidate_id"], source[row["relation_candidate_id"]]["label_candidate_id"])
            self.assertEqual(row["object_candidate_id"], source[row["relation_candidate_id"]]["object_candidate_id"])
            star, label = stars[row["object_candidate_id"]], labels[row["label_candidate_id"]]
            self.assertIn(star["human_decision"], {"ACCEPT", "MODIFY"})
            self.assertIn(label["human_decision"], {"ACCEPT", "MODIFY"})
            self.assertEqual(star["final_class"], "STAR_OBJECT")
            self.assertEqual(star["panel"], label["panel"])

    def test_schema_blindness_and_frames(self):
        schema = json.loads((self.relations / "CVAT_LABEL_STAR_RELATION_SCHEMA.json").read_text())
        for label in schema:
            for attr in label["attributes"]:
                self.assertTrue(attr["values"])
                if attr["input_type"] == "select": self.assertIn(attr["default_value"], attr["values"])
        attrs = {node.attrib["name"] for node in self.expected.findall(".//*[@label]/attribute")}
        self.assertTrue(attrs.isdisjoint({"support_count", "priority", "source_relation_ids", "provisional_relation_types", "model_name"}))
        frames = self.expected.findall("image")
        self.assertEqual(len(frames), 331)
        for frame in frames:
            if frame.attrib["name"].startswith("ZZ_CONTEXT_"):
                self.assertEqual(len(list(frame)), 0)
            else:
                self.assertEqual({s.attrib["label"] for s in frame}, {"LABEL_ENDPOINT", "STAR_ENDPOINT"})

    def test_pixel_and_endpoint_translation(self):
        frames = importer.frames(self.expected)
        stars = {row["candidate_id"]: row for row in read(PKG / "HUMAN_STAR_ADJUDICATION.tsv")}
        labels = {row["candidate_id"]: row for row in read(PKG / "HUMAN_LABEL_ADJUDICATION.tsv")}
        for row in self.queue:
            frame = frames[row["filename"]]
            expected = ET.Element("image")
            builder.shape(expected, labels[row["label_candidate_id"]], "LABEL_ENDPOINT", row, (int(row["crop_x"]),int(row["crop_y"])), int(frame.attrib["id"])+1)
            builder.shape(expected, stars[row["object_candidate_id"]], "STAR_ENDPOINT", row, (int(row["crop_x"]),int(row["crop_y"])), int(frame.attrib["id"])+1)
            self.assertEqual([s.attrib for s in frame], [s.attrib for s in expected])
        for row in self.queue[::100]:
            with Image.open(PKG / "images" / f"{row['panel']}.jpg") as source:
                ox,oy,w,h = (int(row[k]) for k in ("crop_x","crop_y","width","height"))
                expected = source.convert("RGB").crop((ox,oy,ox+w,oy+h))
            with Image.open(self.relations / "images" / row["filename"]) as crop:
                self.assertEqual(crop.convert("RGB").tobytes(), expected.tobytes())

    def test_valid_decisions_multiple_unassigned_uncertain(self):
        reviewed = self.assigned()
        shapes = reviewed.findall(".//*[@label='STAR_ENDPOINT']")
        shapes[0].find("attribute[@name='nearest_object']").text = "true"
        for shape, decision in zip(shapes[1:3], ("UNASSIGNED", "UNCERTAIN")):
            shape.find("attribute[@name='relation_decision']").text = decision
            shape.find("attribute[@name='adjacent_to']").text = "false"
        rows = importer.validate(self.expected, reviewed, "TEST", "2000-01-01T00:00:00Z")
        self.assertEqual(len(rows), 324)
        self.assertEqual(rows[0]["relation_type"], "ADJACENT_TO;NEAREST_OBJECT")
        self.assertEqual(rows[1]["relation_type"], "UNASSIGNED")
        self.assertEqual(rows[2]["relation_type"], "UNCERTAIN")

    def test_invalid_decisions_and_endpoint_changes_are_rejected(self):
        with self.assertRaisesRegex(ValueError, "unreviewed"):
            importer.validate(self.expected, self.expected, "TEST", "2000-01-01T00:00:00Z")
        edits = (
            lambda node: node.set("xtl", str(float(node.attrib["xtl"])+1)),
            lambda node: setattr(node.find("attribute[@name='object_candidate_id']"), "text", "UNKNOWN"),
            lambda node: setattr(node.find("attribute[@name='adjacent_to']"), "text", "false"),
        )
        for edit in edits:
            reviewed = self.assigned()
            edit(reviewed.find(".//*[@label='STAR_ENDPOINT']"))
            with self.assertRaises(ValueError): importer.validate(self.expected, reviewed, "TEST", "2000-01-01T00:00:00Z")

    def test_cvat_two_decimal_export_is_accepted(self):
        reviewed = self.assigned()
        for shape in reviewed.findall(".//*[@label]"):
            for key in importer.GEOMETRY_KEYS:
                if key in shape.attrib: shape.set(key, f"{float(shape.attrib[key]):.2f}")
        self.assertEqual(len(importer.validate(self.expected, reviewed, "TEST", "2000-01-01T00:00:00Z")), 324)

    def test_missing_new_shapes_and_inconsistent_flags_are_rejected(self):
        reviewed = self.assigned()
        frame = reviewed.find("image")
        frame.remove(list(frame)[0])
        with self.assertRaisesRegex(ValueError, "deleted or new"):
            importer.validate(self.expected, reviewed, "TEST", "2000-01-01T00:00:00Z")
        reviewed = self.assigned()
        context = next(frame for frame in reviewed.findall("image") if frame.attrib["name"].startswith("ZZ_CONTEXT_"))
        ET.SubElement(context, "box", {"label":"STAR_ENDPOINT"})
        with self.assertRaisesRegex(ValueError, "deleted or new"):
            importer.validate(self.expected, reviewed, "TEST", "2000-01-01T00:00:00Z")
        reviewed = self.assigned()
        reviewed.find(".//*[@label='STAR_ENDPOINT']/attribute[@name='relation_decision']").text = "UNASSIGNED"
        with self.assertRaisesRegex(ValueError, "with selected flags"):
            importer.validate(self.expected, reviewed, "TEST", "2000-01-01T00:00:00Z")

    def test_cli_import_with_frozen_reference_manifest(self):
        with tempfile.TemporaryDirectory() as directory:
            xml, output = Path(directory) / "reviewed.xml", Path(directory) / "reviewed.tsv"
            ET.ElementTree(self.assigned()).write(xml, encoding="utf-8", xml_declaration=True)
            subprocess.run([
                sys.executable, str(PKG / "scripts/import_label_star_relations.py"),
                "--input", str(xml), "--output", str(output), "--reviewer-id", "TEST",
                "--timestamp", "2000-01-01T00:00:00Z",
            ], check=True, stdout=subprocess.DEVNULL)
            rows = read(output)
            self.assertEqual(len(rows), 324)
            self.assertTrue(all(row["relation_type"] == "ADJACENT_TO" for row in rows))


if __name__ == "__main__":
    unittest.main()
