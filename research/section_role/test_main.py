import csv, json, unittest
from pathlib import Path

ROOT=Path(__file__).resolve().parent
class RoleTest(unittest.TestCase):
 def test_core_and_profiles(self):
  with open(ROOT/'SECTION_ROLE_SHARED_TOKEN_CORE.tsv') as f:self.assertGreater(len(list(csv.reader(f,delimiter='\t'))),1)
  with open(ROOT/'SECTION_ROLE_PROFILES.tsv') as f:self.assertGreater(len(list(csv.reader(f,delimiter='\t'))),100)
 def test_manifest(self):
  m=json.load(open(ROOT/'SECTION_ROLE_RESULTS_MANIFEST.json'));self.assertTrue(m['status']['SECTION_ROLE_REPRODUCIBLE']);self.assertFalse(m['status']['STRUCTURE_MODEL_UPDATED'])
if __name__=='__main__':unittest.main()
