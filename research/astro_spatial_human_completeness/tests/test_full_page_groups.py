from pathlib import Path
import copy
import json
import sys
import unittest

PKG=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(PKG/'scripts'))
from core import attrs,xml_input
from full_page_groups import validate

TASK=PKG/'job_18_attachment_v3_full_pages_grouped'

def completed():
    root=xml_input(TASK/'ANNOTATIONS.zip')
    for tag in root.findall('.//tag'):
        for a in tag.findall('attribute'):
            if a.get('name')=='completion_status':a.text='COMPLETE'
            if a.get('name')=='human_confidence':a.text='HIGH'
    return root

class FullPageGroupTests(unittest.TestCase):
    def test_manifest_and_initial_groups(self):
        meta=json.loads((TASK/'GROUP_TASK_MANIFEST.json').read_text())
        self.assertEqual(meta['panels'],['f68r1','f68r2','f68r3'])
        self.assertEqual((meta['endpoints'],meta['new_endpoints'],meta['initial_groups']),(256,17,46))
    def test_initial_full_page_copied_tree(self):
        groups,audit,markers=validate(xml_input(TASK/'ANNOTATIONS.zip'),TASK,'PREFLIGHT','2026-09-15T10:29:18Z',True)
        self.assertEqual(len(groups),46);self.assertEqual(len(markers),3)
        self.assertTrue(all(r['cartesian_edges_inferred']=='NO' for r in groups))
    def test_completed_unchanged_groups(self):
        groups,audit,_=validate(completed(),TASK,'H01','2026-09-15T10:29:18Z')
        self.assertEqual(len(groups),46)
        self.assertEqual(sum(r['review_outcome']=='UNCHANGED_PRIOR_GROUP' for r in audit),46)
    def test_single_class_group_blocked(self):
        root=completed();shape=next(s for s in root.findall('.//image/*') if s.get('group_id')=='1')
        shape.attrib.pop('group_id')
        with self.assertRaises(ValueError):validate(root,TASK,'H01','2026-09-15T10:29:18Z')

if __name__=='__main__':unittest.main()
