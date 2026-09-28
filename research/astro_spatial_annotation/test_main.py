import csv
import hashlib
import json
import unittest
from pathlib import Path

HERE = Path(__file__).resolve().parent

def rows(name):
    with (HERE/name).open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))

class SpatialLayerTest(unittest.TestCase):
    def test_scope_and_normalized_coordinates(self):
        expected={"f67r1","f67r2","f67v1","f68r1","f68r2","f68r3","f68v1","f68v2"}
        self.assertEqual({r["side_panel"] for r in rows("ASTRO_SPATIAL_SOURCE_REGISTRY.tsv")}, expected)
        for r in rows("ASTRO_OBJECTS.tsv"):
            self.assertIn(r["panel"], expected)
            self.assertLessEqual(0, float(r["center_x_norm"])); self.assertLessEqual(float(r["center_x_norm"]), 1)
            self.assertLessEqual(0, float(r["center_y_norm"])); self.assertLessEqual(float(r["center_y_norm"]), 1)

    def test_ontology_and_nonsemantic_star_data(self):
        allowed={"STAR_OBJECT","LABEL","CIRCLE","RADIAL_LINE","SECTOR","CENTRAL_OBJECT","MOON_OR_DISC_OBJECT","TEXT_ARC","OTHER_DIAGRAM_OBJECT"}
        objects=rows("ASTRO_OBJECTS.tsv")
        self.assertTrue(objects)
        self.assertTrue({r["object_class"] for r in objects} <= allowed)
        self.assertEqual(sum(r["object_class"]=="STAR_OBJECT" for r in objects), 29)
        self.assertTrue(all(not r["token"] for r in rows("ASTRO_STAR_SPATIAL_DATASET.tsv")))

    def test_relations_and_bridges_are_conservative(self):
        object_ids={r["object_id"] for r in rows("ASTRO_OBJECTS.tsv")}
        label_ids={r["label_occurrence_id"] for r in rows("ASTRO_LABELS_SPATIAL.tsv")}
        for r in rows("ASTRO_LABEL_OBJECT_RELATIONS.tsv"):
            self.assertIn(r["object_id"], object_ids); self.assertIn(r["label_occurrence_id"], label_ids)
        self.assertTrue(all(r["match_status"]=="UNMAPPED" for r in rows("ASTRO_STOLFI_SPATIAL_BRIDGE.tsv")))
        self.assertTrue(all(r["match_status"]=="UNMAPPED" for r in rows("ASTRO_ZL3B_SPATIAL_BRIDGE.tsv")))

    def test_manifest_hashes(self):
        manifest=json.loads((HERE/"manifest.json").read_text())
        self.assertEqual(manifest["quality_gate"], "FAIL")
        for name,digest in manifest["artifact_sha256"].items():
            self.assertEqual(hashlib.sha256((HERE/name).read_bytes()).hexdigest(),digest,name)

if __name__ == "__main__": unittest.main()
