import csv, json, unittest
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]/'section_lexical'
class LexicalTest(unittest.TestCase):
 def test_outputs(self):
  for name in ['SECTION_LEXICON_PAGE_REGISTRY.tsv','SECTION_LEXICON_CORPUS_SUMMARY.tsv','SECTION_EXCLUSIVE_TOKENS.tsv','SECTION_TOKEN_ENRICHMENT.tsv','SECTION_LEXICON_REPLICATION.tsv','SECTION_LEXICON_RAREFACTION.tsv','SECTION_LEXICAL_DISTANCES.tsv','SECTION_LEXICON_MUTUAL_INFORMATION.tsv','SECTION_LEXICON_CLASSIFICATION.tsv','SECTION_LEXICON_CONFOUNDER_ANALYSIS.tsv','SECTION_LEXICON_PAIRWISE.tsv']:
   self.assertTrue((ROOT/name).exists())
 def test_summary_sections(self):
  with open(ROOT/'SECTION_LEXICON_CORPUS_SUMMARY.tsv') as f:
   self.assertEqual(len(list(csv.reader(f,delimiter='\t')))-1,8)
if __name__=='__main__': unittest.main()
