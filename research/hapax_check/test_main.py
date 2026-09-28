import csv, json, unittest
from pathlib import Path
ROOT=Path(__file__).resolve().parent
class HapaxTest(unittest.TestCase):
 def test_page_coverage(self):
  with open(ROOT/'PAGE_HAPAX_PREVALENCE.tsv') as f:self.assertEqual(len(list(csv.reader(f,delimiter='\t')))-1,227)
 def test_manifest(self):
  m=json.load(open(ROOT/'HAPAX_RESULTS_MANIFEST.json'));self.assertEqual(m['status']['ASTRO_HAPAX_SPATIAL_PATTERN'],'INCONCLUSIVE')
if __name__=='__main__':unittest.main()
