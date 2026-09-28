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
import analyze as a

def rows(name):
    with (PKG/name).open(encoding='utf-8',newline='') as f:
        return list(csv.DictReader(f,delimiter='\t'))

class AugmentedReferenceTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.objects=rows('AUGMENTED_HUMAN_OBJECTS.tsv')
        cls.added=rows('HUMAN_ADDED_OBJECT_AUDIT.tsv')
        cls.groups=rows('RELATION_GROUPS_3G1.tsv')
        cls.members=rows('RELATION_GROUP_MEMBERS_3G1.tsv')
        cls.metrics=rows('AI_ASSISTED_RECALL_SUMMARY.tsv')

    def test_upstream_checksums_and_registered_inputs(self):
        a.verify_inputs(PKG/'INPUT_MANIFEST.tsv')

    def test_augmented_object_id_uniqueness_and_totals(self):
        self.assertEqual(len(self.objects),557)
        self.assertEqual(len({r['object_id'] for r in self.objects}),557)
        self.assertEqual(Counter(r['object_type'] for r in self.objects),{'STAR':329,'LABEL':228})

    def test_origins_and_prior_decision_eligibility(self):
        self.assertEqual(Counter(r['origin'] for r in self.objects),
                         {'AI_DERIVED_HUMAN_CONFIRMED':520,'HUMAN_ADDED_COMPLETENESS':37})
        self.assertFalse({r['human_decision'] for r in self.objects}&{'REJECT','UNCERTAIN','NOT_REVIEWED','UNCERTAIN_NEW'})

    def test_old_objects_are_exact_frozen_final_geometry(self):
        actual={r['object_id']:r for r in self.objects if r['origin']=='AI_DERIVED_HUMAN_CONFIRMED'}
        baseline={r['canonical_id']:r for r in a.read(a.COMPLETE/'REFERENCE_OBJECTS.tsv')}
        self.assertEqual(set(actual),set(baseline))
        for cid in actual:
            for field in ('panel','object_type',*a.GEOM_FIELDS):self.assertEqual(actual[cid][field],baseline[cid][field])

    def test_human_added_provenance_is_explicitly_reconciled(self):
        new=[r for r in self.objects if r['origin']=='HUMAN_ADDED_COMPLETENESS']
        self.assertEqual(Counter(r['object_type'] for r in new),{'LABEL':29,'STAR':8})
        self.assertTrue(all(r['ai1_support']==r['ai2_support']=='NO' for r in new))
        self.assertTrue(all(r['provenance_reconciliation']=='EXPLICIT_DISTINCT_NEW_OBJECT_NO_FROZEN_AI_MEMBERSHIP' for r in new))
        rec=rows('PRIOR_CANDIDATE_RECONCILIATION.tsv')
        self.assertEqual(len(rec),28)
        self.assertEqual(len({r['human_added_object_id'] for r in rec}),23)
        self.assertTrue(all(r['same_object_interpretation']==r['ai_support_transferred']==r['prior_outcome_changed']=='NO' for r in rec))

    def test_all_37_additions_and_relation_subset_are_audited(self):
        self.assertEqual(len(self.added),37)
        self.assertEqual(sum(r['relation_scope_eligibility']=='YES' for r in self.added),17)
        self.assertEqual(sum(r['group_status']=='GROUPED' for r in self.added),16)
        self.assertEqual(sum(r['group_type']=='TWO_MEMBER_GROUP' for r in self.added),16)

    def test_ungrouped_new_star_is_preserved(self):
        r=next(x for x in self.added if x['object_id']=='HNEW_STAR_f68r2_DC27A556209F27B3')
        self.assertEqual((r['group_id'],r['group_status']),('UNGROUPED','NO_VISUAL_GROUP_ASSIGNED'))

    def test_group_endpoint_snapshot_has_256_unchanged_objects(self):
        snap=a.read(a.GROUP_TASK/'ENDPOINT_SNAPSHOT.tsv')
        freeze=a.read(a.FREEZE/'CONFIRMED_ENDPOINTS.tsv')
        expected=[{k:r[k] for k in a.REF_FIELDS} for r in freeze if r['panel'] in a.RELATION_SCOPE]
        self.assertEqual(len(snap),256)
        self.assertEqual(snap,expected)
        self.assertEqual(len({r['canonical_id'] for r in snap}),256)

    def test_group_totals_panels_sizes_and_confidence(self):
        self.assertEqual(len(self.groups),64)
        self.assertEqual(Counter(r['panel'] for r in self.groups),{'f68r1':29,'f68r2':24,'f68r3':11})
        self.assertEqual(Counter(int(r['member_count']) for r in self.groups),{2:63,8:1})
        self.assertTrue(all(r['human_confidence']=='HIGH' and r['protocol_version']=='3G1' for r in self.groups))

    def test_group_members_are_known_unique_and_nonoverlapping(self):
        known={r['object_id'] for r in self.objects}
        mids=[r['object_id'] for r in self.members]
        self.assertEqual(len(mids),134)
        self.assertEqual(len(set(mids)),134)
        self.assertLessEqual(set(mids),known)

    def test_multimember_group_is_one_label_seven_star(self):
        g=next(r for r in self.groups if int(r['member_count'])==8)
        self.assertEqual((g['panel'],g['label_count'],g['star_count']),('f68r3','1','7'))
        self.assertEqual(g['cartesian_edges_inferred'],'NO')

    def test_initial_to_reviewed_membership_transition(self):
        self.assertEqual(Counter(r['change_category'] for r in self.groups),
                         {'UNCHANGED_FROM_INITIAL':45,'NEW_GROUP':18,'EXTENDED_WITH_EXISTING_REFERENCE_OBJECTS':1})
        trans=rows('GROUP_TRANSITION_AUDIT.tsv')
        self.assertEqual(Counter(r['row_type'] for r in trans),{'INITIAL_GROUP':46,'FINAL_ONLY_GROUP':18})
        self.assertEqual(Counter(r['change_type'] for r in trans),
                         {'UNCHANGED':45,'EXTENDED_WITH_EXISTING_REFERENCE_OBJECTS':1,'NEW_GROUP':18})

    def test_source_group_change_audit_is_reproduced(self):
        source=a.read(a.GROUP_RESULT/'GROUP_CHANGE_AUDIT.tsv')
        self.assertEqual(Counter(r['review_outcome'] for r in source),
                         {'UNCHANGED_PRIOR_GROUP':45,'NEW_OR_MODIFIED_GROUP':19,'REMOVED_OR_REGROUPED_PRIOR_GROUP':1})

    def test_assisted_recall_exact_counts(self):
        lookup={(r['object_type'],r['scope'],r['source']):(int(r['numerator']),int(r['denominator'])) for r in self.metrics}
        expected={('STAR','AI1'):(167,329),('STAR','AI2'):(315,329),('STAR','AI_UNION'):(321,329),
                  ('LABEL','AI1'):(88,228),('LABEL','AI2'):(196,228),('LABEL','AI_UNION'):(199,228)}
        for (typ,src),value in expected.items():self.assertEqual(lookup[typ,'ALL_PANELS',src],value)

    def test_support_partition_and_marginal_contributions(self):
        c={t:Counter(r['support_pattern'] for r in self.objects if r['object_type']==t) for t in ('STAR','LABEL')}
        self.assertEqual(c['STAR'],{'BOTH_AI':161,'AI1_ONLY':6,'AI2_ONLY':154,'HUMAN_ONLY':8})
        self.assertEqual(c['LABEL'],{'BOTH_AI':85,'AI1_ONLY':3,'AI2_ONLY':111,'HUMAN_ONLY':29})

    def test_every_recall_row_has_exact_fraction_and_wilson_interval(self):
        for r in self.metrics:
            k,n=int(r['numerator']),int(r['denominator']);lo,hi=a.wilson(k,n)
            self.assertEqual(r['assisted_recall'],a.f6(k/n))
            self.assertEqual((r['wilson_95_low'],r['wilson_95_high']),(a.f6(lo),a.f6(hi)))

    def test_crosswalk_never_claims_pairwise_metric_or_expands_hyperedge(self):
        cross=rows('ATTACHMENT_V2_3G1_CROSSWALK.tsv')
        self.assertTrue(all(r['pairwise_metric_eligible']=='NO' for r in cross))
        multi=[r for r in cross if r['row_type']=='3G1_MULTIMEMBER_GROUP_SUMMARY']
        self.assertEqual(len(multi),1)
        self.assertEqual(multi[0]['allowed_analytical_claim'],'ENDPOINTS_CO_MEMBER_SAME_3G1_GROUP')
        self.assertFalse(any(r['row_type'].startswith('INFERRED') for r in cross))

    def test_attachment_v2_registered_hash_is_current(self):
        manifest={r['logical_role']:r for r in rows('INPUT_MANIFEST.tsv')}
        row=manifest['ATTACHMENT_V2_FINAL']
        self.assertEqual(row['sha256'],a.digest(a.ROOT/row['path']))

    def test_status_contract_is_exact(self):
        text=(PKG/'AUGMENTED_REFERENCE_REPORT.md').read_text()
        required=['AUGMENTED_HUMAN_REFERENCE_STATUS=COMPLETE','HUMAN_ADDED_OBJECTS_RECONCILED=YES',
          'ASSISTED_RECALL_AI1_CALCULATED=YES','ASSISTED_RECALL_AI2_CALCULATED=YES','ASSISTED_RECALL_AI_UNION_CALCULATED=YES',
          'RELATION_GROUP_PROTOCOL=3G1','FINAL_GROUPS=64','MULTIMEMBER_GROUPS_PRESERVED=YES','CARTESIAN_EDGES_INFERRED=NO',
          'UNGROUPED_OBJECTS_PRESERVED=YES','FROZEN_ATTACHMENT_V2_CHANGED=NO','PREVIOUS_REPORTS_OVERWRITTEN=NO','RESULTS_REPRODUCIBLE=YES']
        self.assertEqual([line for line in text.splitlines() if '=' in line and line.split('=',1)[0] in {x.split('=',1)[0] for x in required}],required)

    def test_machine_readable_sources_exist_for_every_figure(self):
        for figure in (PKG/'figures').glob('*'):
            if figure.suffix in {'.svg','.png'}:
                candidates=[figure.with_name(figure.stem+'_SOURCE.tsv')]
                if figure.name.startswith('CHANGED_GROUPS_'):candidates=[PKG/'figures/GROUP_OVERLAY_SOURCE.tsv']
                self.assertTrue(any(x.is_file() for x in candidates),figure.name)

if __name__=='__main__':unittest.main()
