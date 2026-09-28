from __future__ import annotations

import copy
import json
import subprocess
import sys
import tempfile
import unittest
import xml.etree.ElementTree as ET
from collections import Counter
from pathlib import Path

PKG = Path(__file__).resolve().parents[1]
sys.path.insert(0,str(PKG / "scripts"))
from build_label_star_relations import digest, envelope, read
from import_label_star_attachment_v2 import validate


class AttachmentV2Tests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.base = PKG / "relations_v2"
        cls.main = read(cls.base / "main/QUEUE.tsv")
        cls.special = read(cls.base / "f68v2/QUEUE.tsv")

    def reviewed(self, subset):
        expected = ET.parse(self.base / subset / "annotations.xml").getroot()
        reviewed = copy.deepcopy(expected)
        for node in reviewed.findall(".//*[@label='STAR_ENDPOINT']/attribute[@name='relation_decision']"):
            node.text = "ASSIGNED"
        return expected,reviewed

    def test_partition_right_rule_and_reviewer_decisions(self):
        auto = read(self.base / "REVIEWER_PANEL_DECISIONS.tsv")
        filtered = read(self.base / "RIGHT_RULE_FILTERED.tsv")
        original = read(PKG / "relations/LABEL_STAR_RELATION_QUEUE.tsv")
        groups = [self.main,self.special,auto,filtered]
        ids = [{r["relation_candidate_id"] for r in rows} for rows in groups]
        self.assertEqual([len(x) for x in ids],[99,26,142,57])
        self.assertEqual(len(set.union(*ids)),sum(map(len,ids)))
        self.assertEqual(set.union(*ids),{r["relation_candidate_id"] for r in original})
        self.assertEqual(Counter(r["panel"] for r in self.main),{"f68r1":29,"f68r2":19,"f68r3":51})
        self.assertTrue(all(r["panel"]=="f68v2" for r in self.special))
        self.assertTrue(all(r["relation_type"]=="UNASSIGNED" and r["decision_origin"]=="REVIEWER_PANEL_RULE" and r["relation_protocol_version"]=="2" for r in auto))
        self.assertTrue(all(r["status"]=="RULE_FILTERED_NOT_REVIEWED" for r in filtered))
        labels = {r["candidate_id"]:r for r in read(PKG / "HUMAN_LABEL_ADJUDICATION.tsv")}
        stars = {r["candidate_id"]:r for r in read(PKG / "HUMAN_STAR_ADJUDICATION.tsv")}
        for row in self.main:
            if row["panel"]=="f68r3": continue
            label,star = labels[row["label_candidate_id"]],stars[row["object_candidate_id"]]
            le,se = envelope(label),envelope(star)
            self.assertGreater(float(label["center_x"]),float(star["center_x"]))
            self.assertLessEqual(max(le[1],se[1]),min(le[3],se[3]))

    def test_simple_schema_default_and_original_pixel_reuse(self):
        schema = json.loads((self.base / "CVAT_LABEL_STAR_ATTACHMENT_SCHEMA.json").read_text())
        names = {a["name"] for label in schema for a in label["attributes"]}
        self.assertTrue(names.isdisjoint({"adjacent_to","nearest_object","between_objects","inside_object","on_object"}))
        for label in schema:
            for attr in label["attributes"]:
                self.assertTrue(attr["values"])
                if attr["input_type"]=="select":self.assertIn(attr["default_value"],attr["values"])
        for subset,rows,count in (("main",self.main,102),("f68v2",self.special,27)):
            expected = ET.parse(self.base / subset / "annotations.xml").getroot()
            self.assertEqual(len(expected.findall("image")),count)
            values = [a.text for a in expected.findall(".//*[@label='STAR_ENDPOINT']/attribute[@name='relation_decision']")]
            self.assertEqual(set(values),{"UNREVIEWED"})
            with self.assertRaisesRegex(ValueError,"unreviewed"):
                validate(expected,expected,"TEST","2000-01-01T00:00:00Z")
            for row in rows:
                self.assertEqual(digest(self.base / subset / "images" / row["filename"]),digest(PKG / "relations/images" / row["filename"]))

    def test_visual_attachment_output_and_many_labels_per_star(self):
        expected,reviewed = self.reviewed("main")
        rows = validate(expected,reviewed,"TEST","2000-01-01T00:00:00Z")
        self.assertEqual(len(rows),99)
        self.assertTrue(all(r["relation_type"]=="VISUAL_LABEL_OF" and r["relation_protocol_version"]=="2" for r in rows))
        pairs = {}
        for row in rows:pairs.setdefault(row["object_candidate_id"],set()).add(row["label_candidate_id"])
        self.assertTrue(any(len(labels)>1 for labels in pairs.values()))
        shapes = reviewed.findall(".//*[@label='STAR_ENDPOINT']")
        for shape,state in zip(shapes,("UNASSIGNED","UNCERTAIN")):
            shape.find("attribute[@name='relation_decision']").text = state
        rows = validate(expected,reviewed,"TEST","2000-01-01T00:00:00Z")
        self.assertEqual([rows[0]["relation_type"],rows[1]["relation_type"]],["UNASSIGNED","UNCERTAIN"])
        shapes[0].find("attribute[@name='panel']").text = "WRONG"
        with self.assertRaisesRegex(ValueError,"immutable"):
            validate(expected,reviewed,"TEST","2000-01-01T00:00:00Z")

    def test_explicit_unused_legacy_schema_adapter(self):
        expected,reviewed = self.reviewed("f68v2")
        for shape in reviewed.findall(".//*[@label]"):
            shape.remove(shape.find("attribute[@name='panel']"))
            if shape.attrib["label"]=="STAR_ENDPOINT":
                for name in ("adjacent_to","nearest_object","between_objects","on_object","inside_object"):
                    ET.SubElement(shape,"attribute",{"name":name}).text = "false"
        with self.assertRaises(ValueError):validate(expected,reviewed,"TEST","2000-01-01T00:00:00Z")
        rows = validate(expected,reviewed,"TEST","2000-01-01T00:00:00Z",allow_unused_legacy_schema=True)
        self.assertEqual(len(rows),26)
        self.assertTrue(all(r["relation_type"]=="VISUAL_LABEL_OF" and r["decision_origin"]=="PAIR_REVIEW_LEGACY_SCHEMA_ADAPTED" for r in rows))
        # Adaptation happens on a copy and never changes raw export XML.
        self.assertIsNone(reviewed.find(".//*[@label]/attribute[@name='panel']"))
        shape = reviewed.find(".//*[@label='STAR_ENDPOINT']")
        shape.find("attribute[@name='adjacent_to']").text = "true"
        with self.assertRaisesRegex(ValueError,"active or invalid"):
            validate(expected,reviewed,"TEST","2000-01-01T00:00:00Z",allow_unused_legacy_schema=True)

    def test_reviewer_completion_and_cumulative_freeze(self):
        from import_label_star_relations import load,attributes
        expected = load(self.base / "main/annotations.xml")
        raw = load(PKG / "exports/job_16_annotations_2026_09_14_10_16_23_cvat for images 1.1.zip")
        overrides = read(PKG / "exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_REVIEWER_OVERRIDES_R01.tsv")
        rows = validate(expected,raw,"R01","2026-09-14T10:16:23Z",reviewer_overrides=overrides)
        target = next(row for row in rows if row["relation_candidate_id"]=="HREL_68BD06396134")
        self.assertEqual(target["relation_type"],"UNCERTAIN")
        self.assertEqual(target["decision_origin"],"REVIEWER_POST_EXPORT_CONFIRMATION")
        self.assertEqual(target["review_timestamp"],overrides[0]["review_timestamp"])
        original = next(s for s in raw.findall(".//*[@label='STAR_ENDPOINT']") if attributes(s)["relation_candidate_id"]==target["relation_candidate_id"])
        self.assertEqual(attributes(original)["relation_decision"],"UNREVIEWED")
        self.assertEqual(rows,read(PKG / "exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_MAIN_R01.tsv"))
        with tempfile.TemporaryDirectory() as directory:
            subprocess.run([sys.executable,str(PKG / "scripts/freeze_attachment_v2.py"),"--output-dir",directory],
                           check=True,stdout=subprocess.DEVNULL)
            for filename in ("HUMAN_LABEL_OBJECT_RELATIONS.tsv","exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_FINAL_R01.tsv","exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_METRICS.json"):
                self.assertEqual((Path(directory)/filename).read_bytes(),(PKG/filename).read_bytes())
        final = read(PKG / "HUMAN_LABEL_OBJECT_RELATIONS.tsv")
        self.assertEqual(len(final),267)
        self.assertEqual(Counter(r["relation_type"] for r in final),{"VISUAL_LABEL_OF":53,"UNASSIGNED":171,"UNCERTAIN":43})
        self.assertTrue(all(r["relation_protocol_version"]=="2" for r in final))

    def test_v1_flags_rejected_and_cli_round_trip(self):
        expected,reviewed = self.reviewed("main")
        shape = reviewed.find(".//*[@label='STAR_ENDPOINT']")
        ET.SubElement(shape,"attribute",{"name":"adjacent_to"}).text = "true"
        with self.assertRaisesRegex(ValueError,"spatial-v1"):
            validate(expected,reviewed,"TEST","2000-01-01T00:00:00Z")
        with tempfile.TemporaryDirectory() as directory:
            for subset,count in (("main",99),("f68v2",26)):
                expected,reviewed = self.reviewed(subset)
                xml,output = Path(directory) / f"{subset}.xml",Path(directory) / f"{subset}.tsv"
                ET.ElementTree(reviewed).write(xml,encoding="utf-8",xml_declaration=True)
                subprocess.run([sys.executable,str(PKG / "scripts/import_label_star_attachment_v2.py"),
                                "--input",str(xml),"--output",str(output),"--subset",subset,
                                "--reviewer-id","TEST","--timestamp","2000-01-01T00:00:00Z"],check=True,stdout=subprocess.DEVNULL)
                self.assertEqual(len(read(output)),count)


if __name__ == "__main__":unittest.main()
