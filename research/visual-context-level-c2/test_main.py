import csv, json, unittest
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1] / "visual_context_level_c2"

class C2DiagnosisTest(unittest.TestCase):
    def test_required_tables(self):
        checks = {
            "LEVEL_C_IDENTIFIABILITY_ATTRITION.tsv": 227,
            "LEVEL_C_DESCRIPTOR_MISSINGNESS.tsv": 15,
            "LEVEL_C_TEXTUAL_MISSINGNESS.tsv": 10,
            "LEVEL_C_SECTION_ATTRITION.tsv": 8,
        }
        for name, count in checks.items():
            with self.subTest(name=name):
                with open(ROOT / name) as f:
                    self.assertEqual(len(list(csv.reader(f, delimiter="\t"))) - 1, count)

    def test_v2_is_frozen_but_not_run(self):
        with open(ROOT / "LEVEL_C_V2_INPUT_MANIFEST.json") as f:
            m = json.load(f)
        self.assertEqual(m["LEVEL_C_V2_FEASIBILITY"], "FEASIBLE_WITH_LIMITATIONS")
        self.assertTrue(m["v2_protocol_frozen"])
        self.assertFalse(m["v2_production_run_executed"])

if __name__ == "__main__":
    unittest.main()
