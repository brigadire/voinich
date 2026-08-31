import csv, json, unittest
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]/'research/visual_context_level_c'

class LevelCTest(unittest.TestCase):
 def test_alignment_and_outputs(self):
    rows=list(csv.DictReader(open(ROOT/'LEVEL_C_PAGE_ALIGNMENT.tsv'),delimiter='\t'))
    self.assertEqual(len(rows),227)
    self.assertTrue(all(r['page_id'] for r in rows))
    manifest=json.load(open(ROOT/'LEVEL_C_RESULTS_MANIFEST.json'))
    self.assertTrue(manifest['status']['LEVEL_C_INPUTS_FROZEN'])
    self.assertFalse(manifest['status']['VISUAL_SCHEMA_MODIFIED'])
    self.assertFalse(manifest['status']['TEXTUAL_FINGERPRINT_MODIFIED'])

 def test_descriptor_output(self):
    rows=list(csv.DictReader(open(ROOT/'LEVEL_C_DESCRIPTOR_ASSOCIATIONS.tsv'),delimiter='\t'))
    self.assertEqual(len(rows),150)
    self.assertTrue(all(r['descriptor_id'] and r['text_metric'] for r in rows))

if __name__ == '__main__':
 unittest.main()
