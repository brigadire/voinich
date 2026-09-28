from __future__ import annotations
import json
import math
from pathlib import Path
import sys
import unittest
import numpy as np

PACKAGE=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(PACKAGE/'scripts'))
from common import ROOT, HUMAN, read, digest, CALIBRATION
from statistics import bootstrap, wilson, fisher, holm, rng
from geometry import shape, metrics, iou, area, axial

class GeometryTests(unittest.TestCase):
    def row(self,x1=0,y1=0,x2=100,y2=50): return dict(bbox_x1=x1,bbox_y1=y1,bbox_x2=x2,bbox_y2=y2)
    def test_identical(self):
        a=shape(self.row())
        self.assertAlmostEqual(iou(a['polygon'],a['polygon']),1.)
    def test_disjoint_and_touching(self):
        a=shape(self.row())
        for x in (100,101): self.assertEqual(iou(a['polygon'],shape(self.row(x,0,x+100,50))['polygon']),0.)
    def test_rotation(self):
        a=shape(self.row(),'ROTATED_RECTANGLE',0)
        b=shape(self.row(),'ROTATED_RECTANGLE',90)
        self.assertAlmostEqual(iou(a['polygon'],b['polygon']),1/3)
    def test_axis_equivalence(self):
        self.assertEqual(axial(179,1),2)
        self.assertEqual(axial(0,180),0)
    def test_ellipse_proxy(self):
        a=shape(self.row(),'ELLIPSE_ENVELOPE_PROXY')
        self.assertEqual(len(a['polygon']),256)
        self.assertLess(abs(area(a['polygon'])/(math.pi*50*25)-1),.0002)
    def test_proxy_does_not_change_input(self):
        row=self.row(0,0,40,100)
        saved=row.copy()
        g=shape(row,'ROTATED_RECTANGLE',90,'orientation_normalized_proxy')
        self.assertEqual(row,saved)
        self.assertEqual((g['width'],g['height']),(100,40))

class StatisticsTests(unittest.TestCase):
    def test_bootstrap_reproducible_and_call_order_independent(self):
        a=bootstrap([0,1,2,3],['p1','p1','p2','p2'],'fixed',replicates=100)[0]
        bootstrap([4,5],['p3','p4'],'unrelated',replicates=100)
        b=bootstrap([0,1,2,3],['p1','p1','p2','p2'],'fixed',replicates=100)[0]
        np.testing.assert_array_equal(a,b)
    def test_single_panel_warning(self):
        _,method=bootstrap([1,2],['p1','p1'],'single',replicates=10)
        self.assertIn('no_between_panel_inference',method)
    def test_weighted_cluster_bootstrap_matches_literal_resampling(self):
        values=np.array([1.,5.,2.,6.,7.])
        panels=np.array(['p1','p1','p2','p2','p2'])
        draws=rng('literal').integers(2,size=(100,2))
        groups=[values[panels==p] for p in ('p1','p2')]
        for stat,fn in [('mean',np.mean),('median',np.median)]:
            expected=np.array([fn(np.concatenate([groups[i] for i in row])) for row in draws])
            actual,_=bootstrap(values,panels,'literal',stat,replicates=100)
            np.testing.assert_allclose(actual,expected,rtol=0,atol=1e-15)
    def test_empty_denominator(self): self.assertEqual(wilson(0,0),[None,None])
    def test_wilson_zero_success(self):
        ci=wilson(0,20)
        self.assertAlmostEqual(ci[0],0.)
        self.assertGreater(ci[1],0.)
    def test_exact_fisher(self): self.assertAlmostEqual(fisher(1,9,11,3),.0027594561852200836,places=12)
    def test_holm(self):
        rows=[dict(family='A',p_raw=p,test_id=str(i),p_holm=None) for i,p in enumerate([.01,.03,.20])]
        holm(rows)
        np.testing.assert_allclose([r['p_holm'] for r in rows],[.03,.06,.20])

class FrozenResultsTests(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.objects=read(PACKAGE/'THREE_WAY_OBJECT_RESULTS.tsv')
        cls.cohorts=read(PACKAGE/'ANALYSIS_COHORTS.tsv')
        cls.geometry=read(PACKAGE/'THREE_WAY_GEOMETRY_RESULTS.tsv')
        cls.data=json.loads((PACKAGE/'THREE_WAY_SUMMARY_METRICS.json').read_text())
    def test_plan_lock(self):
        lock=json.loads((PACKAGE/'ANALYSIS_PLAN_LOCK.json').read_text())
        self.assertEqual(lock['plan_sha256'],digest(PACKAGE/'THREE_WAY_COMPARISON_PLAN.md'))
        self.assertEqual(lock['plan_sha256'],self.data['analysis_plan_sha256'])
    def test_frozen_snapshot(self):
        for row in read(PACKAGE/'INPUT_MANIFEST.tsv'):
            self.assertEqual(digest(ROOT/row['path']),row['sha256'],row['path'])
    def test_unique_physical_and_source_ids(self):
        self.assertEqual(len(self.objects),742)
        self.assertEqual(len({r['candidate_id'] for r in self.objects}),742)
        for kind in ('STAR','LABEL'):
            nodes=[]
            for row in self.objects:
                if row['object_type']!=kind: continue
                tokens=row['source_annotation_ids'].split(';')
                self.assertEqual(len(tokens),len({t.split(':')[0] for t in tokens}))
                nodes.extend(tokens)
            self.assertEqual(len(nodes),len(set(nodes)))
    def test_provenance_counts(self):
        for kind,ai,n in [('STAR','ai1',293),('STAR','ai2',330),('LABEL','ai1',175),('LABEL','ai2',206)]:
            self.assertEqual(sum(bool(r[ai+'_id']) for r in self.objects if r['object_type']==kind),n)
    def test_review_partitions(self):
        self.assertEqual(sum(r['uncertain_cohort']=='1' for r in self.objects),8)
        self.assertEqual(sum(r['not_reviewed_cohort']=='1' for r in self.objects),8)
        for row in self.objects:
            if row['human_decision'] in {'UNCERTAIN','NOT_REVIEWED'}:
                self.assertEqual(row['primary_strict'],'0')
                self.assertEqual(row['combined_strict'],'0')
    def test_no_A_absence_negative(self):
        self.assertTrue(all(r['a_absence_negative_evidence']=='0' for r in self.objects))
    def test_calibration_split(self):
        for row in self.objects:
            self.assertEqual(row['scope']=='CALIBRATION',row['panel'] in CALIBRATION)
            if row['panel'] in CALIBRATION: self.assertEqual(row['primary_strict'],'0')
    def test_geometry_confirmed_only(self):
        lookup={r['candidate_id']:r for r in self.objects}
        for row in self.geometry:
            obj=lookup[row['candidate_id']]
            self.assertIn(obj['human_decision'],{'ACCEPT','MODIFY'})
            if row['source'] in {'AI1','AI2'}:
                self.assertEqual(row['source_id'],obj[row['source'].lower()+'_id'])
            self.assertEqual(row['human_protocol_version'],'1.2')
            self.assertGreaterEqual(float(row['iou']),0)
            self.assertLessEqual(float(row['iou']),1)
    def test_confirmation_denominators_and_json_counts(self):
        for scope,classes in self.data['object_metrics'].items():
            for kind,stats in classes.items():
                cells=[r for r in self.objects if r['object_type']==kind and (scope=='COMBINED' or r['scope']==scope)]
                self.assertEqual(stats['candidate_union'],len(cells))
                for ai,ai_stats in stats['AI'].items():
                    supported=[r for r in cells if r['support_'+ai.lower()]=='YES']
                    strict=[r for r in supported if r['human_decision'] in {'ACCEPT','MODIFY','REJECT'}]
                    confirmed=[r for r in strict if r['human_decision'] in {'ACCEPT','MODIFY'}]
                    self.assertEqual(ai_stats['confirmation_strict']['denominator'],len(strict))
                    self.assertEqual(ai_stats['confirmation_strict']['numerator'],len(confirmed))
                confirmed=[r for r in cells if r['human_decision'] in {'ACCEPT','MODIFY'}]
                for group,f in stats['conditional_coverage'].items(): self.assertEqual(f['denominator'],len(confirmed))
    def test_no_missing_confusion_categories(self):
        for result in self.data['confidence_agreement_exploratory'].values():
            self.assertFalse({'NOT_REVIEWED','UNCERTAIN','MISSING'} & set(result['matrix']))
            self.assertEqual(result['n'],sum(sum(r.values()) for r in result['matrix'].values()))
            self.assertEqual(result['n'],result['raw_agreement_fraction']['denominator'])
            self.assertTrue(all(v is not None for v in result['raw_agreement_fraction']['ci95']))
    def test_protocol_separation(self):
        graph=self.data['attachment_v2']
        self.assertEqual(graph['analysis'],'DESCRIPTIVE')
        self.assertTrue(all(not a['compatible_v2'] for a in graph['ai_compatibility_audit']))
        self.assertTrue(all(r['relation_protocol_version']=='2' for r in read(PACKAGE/'ATTACHMENT_V2_EDGE_RESULTS.tsv')))
        self.assertTrue(all(r['human_protocol_version']=='1.2' for r in self.objects))
    def test_attachment_graph_and_denominators(self):
        graph=self.data['attachment_v2']
        self.assertEqual(graph['confirmed_links'],53)
        nodes=read(PACKAGE/'ATTACHMENT_V2_NODE_DEGREES.tsv')
        for kind in ('STAR','LABEL'):
            self.assertEqual(sum(int(r['degree']) for r in nodes if r['object_type']==kind),53)
            self.assertEqual(sum(int(r['degree'])>0 for r in nodes if r['object_type']==kind),51)
        self.assertEqual(graph['explicit_review_coverage_eligible']['denominator'],324)
        self.assertEqual(graph['explicit_review_coverage_original']['denominator'],675)
        self.assertEqual(graph['individual_selected_task_coverage']['numerator'],125)
    def test_geometry_summary_matches_json_and_raw(self):
        tab=read(PACKAGE/'GEOMETRY_SUMMARY_METRICS.tsv')
        self.assertEqual(len(tab),len(self.data['geometry_summaries']))
        for row,js in zip(tab,self.data['geometry_summaries']):
            self.assertEqual(int(row['n']),js['n'])
            self.assertAlmostEqual(float(row['median']),js['median'])
            cells=[r for r in self.geometry if r['object_type']==row['object_type'] and r['source']==row['source']
                   and r['geometry_type']==row['geometry_type'] and r['representation']==row['representation']
                   and (row['scope']=='COMBINED' or r['scope']==row['scope'])
                   and (row['grouping']!='PANEL' or r['panel']==row['stratum'])
                   and (row['grouping']!='DECISION' or r['human_decision']==row['stratum']) and r[row['metric']]!='']
            self.assertEqual(len(cells),int(row['n']))
            self.assertAlmostEqual(float(np.median([float(r[row['metric']]) for r in cells])),float(row['median']))
    def test_statistical_families(self):
        rows=read(PACKAGE/'THREE_WAY_STATISTICAL_TESTS.tsv')
        for kind in ('STAR','LABEL'):
            self.assertEqual(sum(r['family']==kind+'_categorical' for r in rows),12)
            self.assertEqual(sum(r['family']==kind+'_continuous' for r in rows),4)
        self.assertTrue(all(r['scope']=='PRODUCTION' for r in rows))
    def test_figure_source_tables(self):
        index=read(PACKAGE/'figures/FIGURE_INDEX.tsv')
        self.assertGreaterEqual(len(index),10)
        for row in index:
            self.assertTrue((PACKAGE/row['figure']).is_file())
            self.assertTrue(row['source_tables'])
            for path in row['source_tables'].split(';'): self.assertTrue((PACKAGE/path).is_file(),path)
    def test_sensitivity_never_uses_not_reviewed(self):
        rows=read(PACKAGE/'SENSITIVITY_RESULTS.tsv')
        for row in rows:
            self.assertLessEqual(int(row['confirmed_or_bound_numerator']),int(row['denominator']))
            self.assertIn(row['uncertain_policy'],{'exclude','negative','positive'})
        for kind in ('STAR','LABEL'):
            for ai in ('AI1','AI2'):
                row=next(r for r in rows if r['policy']=='PRIMARY_PRODUCTION_STRICT' and r['object_type']==kind and r['support_group']==ai)
                self.assertEqual(int(row['denominator']),self.data['object_metrics']['PRODUCTION'][kind]['AI'][ai]['strict_denominator'])
    def test_output_checksums_if_frozen(self):
        ledger=PACKAGE/'SHA256SUMS'
        if not ledger.exists(): self.skipTest('output freeze follows first validation; tested again after freezing')
        for line in ledger.read_text().splitlines():
            expected,path=line.split('  ',1)
            self.assertEqual(digest(PACKAGE/path),expected,path)

if __name__=='__main__': unittest.main()
