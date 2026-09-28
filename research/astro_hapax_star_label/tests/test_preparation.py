#!/usr/bin/env python3
from __future__ import annotations

import csv
from collections import Counter
import json
from pathlib import Path
import sys
import unittest

PKG=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(PKG/'scripts'))
import build_preparation as b
import import_human_mapping as importer

def rows(name):return b.read_tsv(PKG/name)

class PreparationTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.candidates=rows('LABEL_TOKEN_CANDIDATES.tsv');cls.tokens=rows('TRANSCRIPTION_TOKEN_CANDIDATES.tsv')

    def test_upstream_checksums_and_manifest(self):b.verify_inputs(PKG/'INPUT_MANIFEST.tsv')

    def test_plan_is_frozen(self):
        lock=json.loads((PKG/'ANALYSIS_PLAN_FREEZE.json').read_text())
        self.assertEqual(lock['sha256'],b.sha(PKG/'HAPAX_STAR_LABEL_ANALYSIS_PLAN.md'))
        self.assertTrue(lock['frozen_before_enrichment'] and lock['frozen_before_verified_mapping'])

    def test_all_92_relation_labels_exactly_once(self):
        source=[r for r in b.read_tsv(b.AUG/'AUGMENTED_HUMAN_OBJECTS.tsv') if r['panel'] in b.PANELS and r['object_type']=='LABEL']
        self.assertEqual(len(self.candidates),92);self.assertEqual(len({r['label_id'] for r in self.candidates}),92)
        self.assertEqual({r['label_id'] for r in self.candidates},{r['object_id'] for r in source})
        self.assertEqual(Counter(r['panel'] for r in self.candidates),{'f68r1':37,'f68r2':33,'f68r3':22})

    def test_candidate_geometry_is_frozen(self):
        src={r['object_id']:r for r in b.read_tsv(b.AUG/'AUGMENTED_HUMAN_OBJECTS.tsv')}
        for r in self.candidates:
            for k in ('panel','geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation'):self.assertEqual(r[k],src[r['label_id']][k])

    def test_group_membership_is_exact_but_internal(self):
        memberships={r['object_id']:r['canonical_group_id'] for r in b.read_tsv(b.AUG/'RELATION_GROUP_MEMBERS_3G1.tsv') if r['object_type']=='LABEL'}
        self.assertEqual(len(memberships),64)
        for r in self.candidates:self.assertEqual(r['group_id'],memberships.get(r['label_id'],'UNGROUPED'))
        self.assertEqual(Counter(r['group_id']!='UNGROUPED' for r in self.candidates),{True:64,False:28})

    def test_no_forced_candidate_mapping(self):
        self.assertTrue(all(r['proposed_transcription_line']==r['proposed_token_position']==r['proposed_raw_token_form']==r['proposed_normalized_token_form']=='NONE' for r in self.candidates))
        self.assertTrue(all(r['candidate_mapping_outcome']=='AMBIGUOUS' and r['mapping_confidence']=='LOW' for r in self.candidates))

    def test_frozen_occurrences_are_byte_derived(self):
        actual={(r['panel'],r['occurrence_id']):r for r in self.tokens};expected={}
        with b.OCC.open(encoding='utf-8') as f:
            for line in f:
                o=json.loads(line)
                if o['folio'] in b.PANELS:expected[(o['folio'],str(o['absolute_token_position']))]=o
        self.assertEqual(len(actual),268);self.assertEqual(set(actual),set(expected))
        for key,r in actual.items():
            token=expected[key]['token'];self.assertEqual(r['canonical_token_key'],'/'.join(token.split('\x1f')));self.assertEqual(r['readable_eva'],b.expand(token))

    def test_candidates_are_same_page_exhaustive(self):
        lines=rows('TRANSCRIPTION_LINE_CANDIDATES.tsv');by=Counter(r['panel'] for r in lines)
        self.assertEqual(by,{'f68r1':37,'f68r2':31,'f68r3':22})
        for r in self.candidates:self.assertEqual(int(r['candidate_line_count']),by[r['panel']])

    def test_blind_ui_has_no_prohibited_metadata(self):
        html=(PKG/'human_review/index.html').read_text().lower();text=(PKG/'human_review/review_data.js').read_text().lower()
        for term in ('hapax','frequency','group_id','group_size','human_added','lexicon','hypothesis'):
            self.assertNotIn(term,html);self.assertNotIn(term,text)

    def test_ui_has_every_crop_and_three_context_pages(self):
        self.assertEqual(len(list((PKG/'human_review/assets/crops').glob('*.jpg'))),92)
        self.assertEqual({p.name for p in (PKG/'human_review/assets/pages').glob('*.jpg')},{'f68r1.jpg','f68r2.jpg','f68r3.jpg'})
        for r in self.candidates:self.assertTrue((PKG/r['visual_crop_reference']).is_file())

    def test_template_is_incomplete_not_no_match(self):
        template=rows('HUMAN_REVIEW_EXPORT_TEMPLATE.tsv')
        self.assertEqual(len(template),92);self.assertTrue(all(r['completion_status']=='INCOMPLETE' and not r['mapping_outcome'] for r in template))
        with self.assertRaises(ValueError):importer.validate_rows(template)

    def test_roundtrip_validator_accepts_explicit_ambiguous_fixture(self):
        fixture=[]
        for row in rows('HUMAN_REVIEW_EXPORT_TEMPLATE.tsv'):
            fixture.append({**row,'review_action':'AMBIGUOUS','mapping_outcome':'AMBIGUOUS','mapping_confidence':'HIGH',
              'reviewer_id':'TEST_REVIEWER','decision_timestamp':'2026-09-15T12:00:00Z','ambiguity_status':'AMBIGUOUS','completion_status':'COMPLETE'})
        verified=importer.validate_rows(fixture)
        self.assertEqual(len(verified),92);self.assertTrue(all(r['raw_token_forms']=='NONE' for r in verified))

    def test_importer_rejects_invented_occurrence(self):
        fixture=[]
        for row in rows('HUMAN_REVIEW_EXPORT_TEMPLATE.tsv'):
            fixture.append({**row,'review_action':'AMBIGUOUS','mapping_outcome':'AMBIGUOUS','mapping_confidence':'HIGH',
              'reviewer_id':'TEST','decision_timestamp':'2026-09-15T12:00:00Z','ambiguity_status':'AMBIGUOUS','completion_status':'COMPLETE'})
        fixture[0]['selected_occurrence_ids']='INVENTED'
        with self.assertRaises(ValueError):importer.validate_rows(fixture)

    def test_multimember_group_remains_one_group_observation(self):
        groups=b.read_tsv(b.AUG/'RELATION_GROUPS_3G1.tsv');multi=[r for r in groups if r['member_count']=='8']
        self.assertEqual(len(multi),1);self.assertEqual((multi[0]['label_count'],multi[0]['star_count']),('1','7'))
        self.assertEqual(multi[0]['cartesian_edges_inferred'],'NO')

    def test_late_analysis_outputs_not_created(self):
        forbidden=('HUMAN_VERIFIED_LABEL_TOKEN_MAPPING.tsv','HAPAX_ANALYTIC_COHORTS.tsv','HAPAX_ENRICHMENT_RESULTS.tsv','LEXICON_MATCH_RESULTS.tsv','FINAL_HAPAX_STAR_LABEL_REPORT.md')
        self.assertFalse(any((PKG/x).exists() for x in forbidden))

    def test_gate_a_status(self):
        report=(PKG/'PREPARATION_REPORT.md').read_text()
        for line in ('HAPAX_STAR_LABEL_PREPARATION=COMPLETE','LABELS_IN_RELATION_SCOPE=92','LABEL_TOKEN_MAPPING_STATUS=READY_FOR_HUMAN_VERIFICATION',
          'HAPAX_DEFINITION_FROZEN=YES','ANALYSIS_PLAN_FROZEN=YES','HAPAX_ENRICHMENT_RUN_AUTHORIZED=NO','LEXICON_MATCH_RUN_AUTHORIZED=NO',
          'FROZEN_INPUTS_UNCHANGED=YES','RESULTS_REPRODUCIBLE=YES'):self.assertIn(line,report)

if __name__=='__main__':unittest.main()
