from __future__ import annotations
import copy
import hashlib
import json
from pathlib import Path
import sys
import tempfile
import unittest
import xml.etree.ElementTree as ET
import zipfile

PKG=Path(__file__).resolve().parents[1];sys.path.insert(0,str(PKG/'scripts'))
from core import (ROOT,HUMAN,OBJECT_VERSION,RELATION_VERSION,REF_FIELDS,CHECKS,TOLERANCE,attrs,attr,check,read,write,
                  digest,schema,relation_schema,xml_input,render_xml,stable_id,overlaps,shape_element,jswrite,verify_inputs)
from validate_objects import validate
from attachment import load_scope,queue,validate_relation,gated_endpoints,generate
DRY=PKG/'dry_run/attachment_f68r1_f68r2_NOT_FOR_PRODUCTION'

def original():return xml_input(PKG/'OBJECT_COMPLETENESS_CVAT.xml')

def completed():
    return xml_input(PKG/'dry_run/OBJECT_ROUNDTRIP_SYNTHETIC_NOT_FOR_PRODUCTION.xml')

def add(root,kind='STAR',row=None,status='PROPOSED_NEW'):
    image=root.find('image')
    geometry=row or dict(geometry_type='BOX',bbox_x1='100',bbox_y1='100',bbox_x2='135',bbox_y2='135',rotation='0')
    shape=shape_element(image,geometry|{'reference_id':'NOT_A_REFERENCE','layer':kind+'_HUMAN_ADDED'},kind+'_HUMAN_ADDED')
    for element in list(shape):shape.remove(element)
    for name,value in [('addition_status',status),('human_confidence','LOW'),('object_protocol_version',OBJECT_VERSION)]:attr(shape,name,value)
    for a in image.find('tag').findall('attribute'):
        if a.get('name')=='completion_status':a.text='COMPLETE_WITH_NEW_OBJECTS'
    return shape

def relation_completed():
    root=xml_input(DRY/'annotations.xml')
    for tag in root.findall('.//tag'):
        for a in tag:
            if a.get('name')=='decision':a.text='UNCERTAIN'
            if a.get('name')=='human_confidence':a.text='LOW'
    return root

def synthetic_freeze(directory,with_new=False):
    """Only temporary test data; never accept an actual reviewer export or generate production."""
    root=completed()
    if with_new:add(root)
    proposals,markers,_=validate(root)
    endpoints=read(PKG/'REFERENCE_OBJECTS.tsv')
    if with_new:
        p=proposals[0]
        endpoints.append({k:p[k] for k in REF_FIELDS if k in p}|{
            'reference_id':'REF_'+hashlib.sha256(p['canonical_id'].encode()).hexdigest()[:16].upper(),
            'layer':p['object_type']+'_REFERENCE','post_review_decision':'CONFIRMED_NEW',
            'confirmation_reviewer_id':'SYNTHETIC_TEST_ONLY','confirmation_timestamp':'2000-01-01T00:00:00Z',
            'reconciliation_status':'EXPLICITLY_RECONCILED_NO_UNRESOLVED_CONFLICT'})
    fields=REF_FIELDS+['post_review_decision','confirmation_reviewer_id','confirmation_timestamp','reconciliation_status']
    write(directory/'endpoints.tsv',endpoints,fields)
    from core import NEW_FIELDS
    write(directory/'proposals.tsv',proposals,NEW_FIELDS)
    write(directory/'completion.tsv',markers)
    meta={'status':'OBJECT_COMPLETENESS_REVIEWED_AND_FROZEN','object_protocol_version':OBJECT_VERSION,
          'reviewer_id':'SYNTHETIC_TEST_ONLY','freeze_timestamp':'2000-01-01T00:00:00Z',
          'prepared_reference_sha256':digest(PKG/'REFERENCE_OBJECTS.tsv')}
    for key,name in [('confirmed_endpoints','endpoints.tsv'),('proposed_additions','proposals.tsv'),('panel_completion','completion.tsv')]:
        meta[key]={'path':name,'sha256':digest(directory/name)}
    jswrite(directory/'SYNTHETIC_ONLY.json',meta)
    return directory/'SYNTHETIC_ONLY.json'

def rehash_fixture(manifest,key):
    meta=json.loads(manifest.read_text());entry=meta[key]
    entry['sha256']=digest(manifest.parent/entry['path']);jswrite(manifest,meta)

def confirmed_scope(kind='WHITELIST'):
    scope,_=load_scope();scope=copy.deepcopy(scope)
    rows=[r for r in read(PKG/'REFERENCE_OBJECTS.tsv') if r['panel']=='f68r3']
    rule={'status':'REVIEWER_CONFIRMED_AND_FROZEN','kind':kind,'reviewer_id':'SYNTHETIC_TEST_ONLY',
          'confirmation_timestamp':'2000-01-01T00:00:00Z'}
    if kind=='WHITELIST':
        rule.update(star_ids=[next(r['canonical_id'] for r in rows if r['object_type']=='STAR')],
                    label_ids=[next(r['canonical_id'] for r in rows if r['object_type']=='LABEL')])
    else:rule['polygon']=[[100,100],[500,100],[500,500],[100,500]]
    rule['rule_sha256']=hashlib.sha256(json.dumps(rule,sort_keys=True,separators=(',',':')).encode()).hexdigest()
    scope['panels']['f68r3']='ELIGIBLE_LIMITED';scope['f68r3_rule']=rule
    return scope

def rehash_rule(scope):
    rule=scope['f68r3_rule'];body={k:v for k,v in rule.items() if k!='rule_sha256'}
    rule['rule_sha256']=hashlib.sha256(json.dumps(body,sort_keys=True,separators=(',',':')).encode()).hexdigest()

class PackageTests(unittest.TestCase):
    def test_inputs_unchanged(self):
        for row in read(PKG/'INPUT_MANIFEST.tsv'):self.assertEqual(digest(ROOT/row['path']),row['sha256'],row['path'])
    def test_reference_final_not_candidate_geometry(self):
        rows=read(PKG/'REFERENCE_OBJECTS.tsv');self.assertEqual(len(rows),520)
        lookup={r['candidate_id']:r for f in ('HUMAN_STAR_ADJUDICATION.tsv','HUMAN_LABEL_ADJUDICATION.tsv') for r in read(HUMAN/f) if r['human_decision'] in {'ACCEPT','MODIFY'}}
        self.assertEqual(set(lookup),{r['canonical_id'] for r in rows})
        for row in rows:
            for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2'):self.assertEqual(row[k],lookup[row['canonical_id']][k])
    def test_prior_states_preserved(self):
        from collections import Counter
        rows=read(PKG/'PRIOR_NONCONFIRMED_PROVENANCE.tsv')
        self.assertEqual(Counter(r['prior_status'] for r in rows),{'REJECT':206,'UNCERTAIN':8,'NOT_REVIEWED':8})
        self.assertTrue(all(r['source_annotation_ids'] for r in rows))
    def test_ui_blindness(self):
        xml=(PKG/'OBJECT_COMPLETENESS_CVAT.xml').read_text()
        for forbidden in ('AI1','AI2','support_pattern','priority','source_annotation_ids','HOBJ_','HLABEL_','REJECT','ACCEPT','MODIFY','lexical','semantic'):
            self.assertNotIn(forbidden,xml)
        for shape in original().findall('.//image/*'):
            if shape.get('label','').endswith('REFERENCE'):
                self.assertEqual(set(attrs(shape)),{'reference_id','object_protocol_version'})
    def test_uncertain_optional_and_rejected_hidden(self):
        shapes=[s for i in original().findall('image') for s in i if s.tag!='tag']
        self.assertEqual(sum(s.get('label')=='UNCERTAIN_REFERENCE' for s in shapes),8)
        self.assertEqual(len(shapes),528)
    def test_all_schema_values_nonempty(self):
        for labels in (schema(),relation_schema()):
            for label in labels:
                for a in label['attributes']:
                    self.assertTrue(a['values']);self.assertIn(a['default_value'],a['values'])
    def test_eight_original_frames_and_markers(self):
        images=original().findall('image');self.assertEqual(len(images),8)
        self.assertTrue(all(len(i.findall('tag'))==1 for i in images))
        self.assertTrue(all(attrs(i.find('tag'))['completion_status']=='NOT_STARTED' for i in images))
    def test_navigation_exact_cover(self):
        for panel in read(PKG/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv'):
            tiles=[t for t in read(PKG/'NAVIGATION_TILES.tsv') if t['panel']==panel['panel']]
            self.assertEqual(len(tiles),12)
            area=sum((int(t['x2'])-int(t['x1']))*(int(t['y2'])-int(t['y1'])) for t in tiles)
            self.assertEqual(area,int(panel['width'])*int(panel['height']))
    def test_zip_structure_and_images(self):
        with zipfile.ZipFile(PKG/'OBJECT_COMPLETENESS_CVAT.zip') as z:self.assertEqual(z.read('annotations.xml'),(PKG/'OBJECT_COMPLETENESS_CVAT.xml').read_bytes())
        with zipfile.ZipFile(PKG/'OBJECT_COMPLETENESS_IMAGES.zip') as z:
            self.assertEqual(len(z.namelist()),8)
            for row in read(PKG/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv'):self.assertEqual(hashlib.sha256(z.read(row['filename'])).hexdigest(),row['image_sha256'])
    def test_no_review_or_production_started(self):
        m=json.loads((PKG/'PREPARATION_MANIFEST.json').read_text())
        self.assertFalse(m['human_review_started']);self.assertFalse(m['production_attachment_package_created'])
        self.assertEqual(json.loads((DRY/'QUEUE_MANIFEST.json').read_text())['mode'],'NOT_FOR_PRODUCTION')

class ObjectValidationTests(unittest.TestCase):
    def test_auto_prediction_not_a_human_addition(self):
        root=completed();add(root).set('source','auto')
        with self.assertRaises(ValueError):validate(root)
    def test_cvat_file_source_roundtrip_for_existing_references(self):
        root=completed()
        for image in root.findall('image'):
            for shape in image:
                if shape.get('label','').endswith('REFERENCE'):shape.set('source','file')
        self.assertEqual(validate(root)[0],[])
    def test_empty_uncompleted_is_not_completeness(self):
        with self.assertRaises(ValueError):validate(original())
        self.assertEqual(len(validate(original(),allow_incomplete=True)[0]),0)
    def test_explicit_zero_new_completion(self):
        additions,markers,flags=validate(completed())
        self.assertEqual(additions,[]);self.assertEqual(len(markers),8);self.assertEqual(flags,[])
    def test_actual_two_decimal_roundtrip(self):
        root=completed()
        for image in root.findall('image'):
            for shape in image:
                if shape.tag=='tag':continue
                for key in ('xtl','ytl','xbr','ybr','cx','cy','rx','ry','rotation'):
                    if key in shape.attrib:shape.set(key,f'{float(shape.get(key)):.2f}')
        self.assertEqual(len(validate(root)[0]),0)
    def test_reference_delete(self):
        root=completed();image=root.find('image');image.remove(image.find('box'))
        with self.assertRaises(ValueError):validate(root)
    def test_reference_move(self):
        root=completed();s=root.find('.//box');s.set('xtl',str(float(s.get('xtl'))+.1))
        with self.assertRaises(ValueError):validate(root)
    def test_reference_rotate(self):
        root=completed();s=next(s for i in root.findall('image') for s in i if s.get('label')=='LABEL_REFERENCE');s.set('rotation',str(float(s.get('rotation'))+1))
        with self.assertRaises(ValueError):validate(root)
    def test_reference_attribute_id(self):
        root=completed();s=root.find('.//box');s.find('attribute').text='RENAMED'
        with self.assertRaises(ValueError):validate(root)
    def test_reference_extra_attribute(self):
        root=completed();attr(root.find('.//box'),'history','ACCEPT')
        with self.assertRaises(ValueError):validate(root)
    def test_reference_class(self):
        root=completed();root.find('.//box').set('label','FORBIDDEN')
        with self.assertRaises(ValueError):validate(root)
    def test_reference_duplicate(self):
        root=completed();root.find('image').append(copy.deepcopy(root.find('.//box')))
        with self.assertRaises(ValueError):validate(root)
    def test_missing_marker(self):
        root=completed();i=root.find('image');i.remove(i.find('tag'))
        with self.assertRaises(ValueError):validate(root)
    def test_marker_checks_required(self):
        root=completed();next(a for a in root.find('.//tag') if a.get('name')=='star_review_done').text='false'
        with self.assertRaises(ValueError):validate(root)
    def test_new_objects_not_autoaccepted(self):
        root=completed();add(root)
        additions,_,_=validate(root)
        self.assertEqual(len(additions),1)
        self.assertTrue(additions[0]['canonical_id'].startswith('HNEW_STAR_f67r1_'))
        self.assertEqual(additions[0]['acceptance_status'],'NOT_ACCEPTED_REQUIRES_POST_REVIEW')
        self.assertEqual(additions[0]['creation_origin'],'HUMAN_COMPLETENESS_PASS')
    def test_new_uncertain_not_promoted(self):
        root=completed();add(root,status='UNCERTAIN_NEW')
        additions,_,_=validate(root);self.assertEqual(additions[0]['addition_status'],'UNCERTAIN_NEW')
        self.assertEqual(additions[0]['acceptance_status'],'NOT_ACCEPTED_REQUIRES_POST_REVIEW')
    def test_canonical_ids_ignore_cvat_internal_ids(self):
        root=completed();s=add(root);s.set('id','42');first=validate(root)[0]
        s.set('id','999');second=validate(root)[0]
        self.assertEqual(first,second)
    def test_duplicate_new_signature_blocks(self):
        root=completed();s=add(root);root.find('image').append(copy.deepcopy(s))
        with self.assertRaises(ValueError):validate(root)
    def test_new_confidence_required(self):
        root=completed();s=add(root);next(a for a in s if a.get('name')=='human_confidence').text='UNSET'
        with self.assertRaises(ValueError):validate(root)
    def test_new_star_rotation_forbidden(self):
        root=completed();s=add(root);s.set('rotation','20')
        with self.assertRaises(ValueError):validate(root)
    def test_new_unsupported_geometry(self):
        root=completed();s=add(root);s.tag='polygon'
        with self.assertRaises(ValueError):validate(root)
    def test_wrong_panel_or_frame(self):
        root=completed();root.find('image').set('id','555')
        with self.assertRaises(ValueError):validate(root)
    def test_completion_matches_proposal_count(self):
        root=completed();add(root);next(a for a in root.find('.//tag') if a.get('name')=='completion_status').text='COMPLETE_NO_NEW_OBJECTS'
        with self.assertRaises(ValueError):validate(root)
    def test_all_prior_overlap_classes_queue(self):
        base=read(PKG/'REFERENCE_OBJECTS.tsv')[0]
        for status in ('REJECT','UNCERTAIN','NOT_REVIEWED'):
            old=base|{'prior_status':status};new=base|{'canonical_id':'HNEW_TEST'}
            self.assertEqual(overlaps(new,[old])[0]['status'],'MATCHES_PRIOR_'+status)
        self.assertEqual(overlaps(base,[base])[0]['status'],'POSSIBLE_DUPLICATE_CONFIRMED')

class RelationTests(unittest.TestCase):
    def test_optional_blind_recheck_includes_all_prior_v2_pairs(self):
        scope,_=load_scope();endpoints=read(PKG/'REFERENCE_OBJECTS.tsv')
        ordinary,_,retained=queue(endpoints,scope)
        recheck,_,retained_again=queue(endpoints,scope,recheck_prior=True)
        self.assertEqual(len(retained),48);self.assertEqual(retained,retained_again)
        self.assertEqual(len(recheck),len(ordinary)+48)
        self.assertEqual(sum(r['inclusion_reason']=='FROZEN_V2_PAIR_RECHECK' for r in recheck),48)
    def test_targeted_queue_keeps_prior_and_one_nearest_pair_for_new_endpoint(self):
        scope,_=load_scope();endpoints=read(PKG/'REFERENCE_OBJECTS.tsv')
        seed=next(r for r in endpoints if r['panel']=='f68r2' and r['object_type']=='STAR')
        endpoints=endpoints+[seed|{'canonical_id':'HNEW_STAR_f68r2_SYNTHETIC','reference_id':'REF_SYNTHETIC'}]
        pairs,_,retained=queue(endpoints,scope,recheck_prior=True,cover_new_endpoints=True,targeted_new_and_prior_only=True)
        involving=[r for r in pairs if 'HNEW_STAR_f68r2_SYNTHETIC' in {r['label_id'],r['star_id']}]
        self.assertEqual(len(retained),48);self.assertEqual(len(involving),1);self.assertEqual(len(pairs),49)
    def test_f68r3_awaits_confirmation(self):
        scope,_=load_scope();self.assertIsNone(scope['f68r3_rule'])
        self.assertEqual(scope['panels']['f68r3'],'AWAITING_REVIEWER_CONFIRMATION')
        self.assertTrue(all(r['panel'] in {'f68r1','f68r2'} for r in read(DRY/'QUEUE.tsv')))
    def test_no_automatic_negatives(self):
        self.assertTrue(all(r['initial_decision']=='UNREVIEWED' for r in read(DRY/'QUEUE.tsv')))
        self.assertTrue(all(r['status']=='FILTERED_PAIR_NOT_A_HUMAN_OBSERVATION' for r in read(DRY/'FILTERED_PAIRS.tsv')))
    def test_no_deprecated_panel_products(self):
        for filename in ('QUEUE.tsv','FILTERED_PAIRS.tsv','RETAINED_PRIOR_V2.tsv'):
            self.assertTrue(all(r['panel'] in {'f68r1','f68r2'} for r in read(DRY/filename)))
    def test_prior_v2_not_reinterpreted(self):
        self.assertTrue(all(r['handling']=='RETAIN_FROZEN_DECISION_NO_REVIEW_NO_CONVERSION' for r in read(DRY/'RETAINED_PRIOR_V2.tsv')))
    def test_raw_relation_unreviewed_rejected(self):
        with self.assertRaises(ValueError):validate_relation(xml_input(DRY/'annotations.xml'),DRY)
    def test_synthetic_relation_roundtrip(self):
        root=relation_completed()
        for i in root.findall('image'):
            for shape in i:
                for key in ('xtl','ytl','xbr','ybr','cx','cy','rx','ry','rotation'):
                    if key in shape.attrib:shape.set(key,f'{float(shape.get(key)):.2f}')
        rows=validate_relation(root,DRY)
        self.assertEqual(len(rows),len(read(DRY/'QUEUE.tsv')))
        self.assertTrue(all(r['relation_type']=='UNCERTAIN' and r['relation_protocol_version']=='3' for r in rows))
    def test_moved_endpoint(self):
        root=relation_completed();s=root.find('.//box');s.set('xbr',str(float(s.get('xbr'))+1))
        with self.assertRaises(ValueError):validate_relation(root,DRY)
    def test_legacy_spatial_field_blocks(self):
        root=relation_completed();attr(root.find('.//tag'),'inside_sector','false')
        with self.assertRaises(ValueError):validate_relation(root,DRY)
    def test_wrong_relation_version(self):
        root=relation_completed();next(a for a in root.find('.//tag') if a.get('name')=='relation_protocol_version').text='2'
        with self.assertRaises(ValueError):validate_relation(root,DRY)
    def test_missing_pair_is_not_negative(self):
        root=relation_completed();root.remove(root.find('image'))
        with self.assertRaises(ValueError):validate_relation(root,DRY)
    def test_region_question_not_negative(self):
        root=relation_completed();next(a for a in root.find('.//tag') if a.get('name')=='decision').text='NOT_APPLICABLE_REGION'
        rows=validate_relation(root,DRY)
        self.assertTrue(any(r['eligibility_reconciliation_status']=='AWAITING_REVIEWER_REGION_RECONCILIATION_NOT_A_NEGATIVE' for r in rows))
    def test_dryrun_cannot_be_review_import(self):
        with self.assertRaises(ValueError):validate_relation(relation_completed(),DRY,reviewer='R01')
    def test_production_generation_without_freeze_blocked(self):
        scope,_=load_scope()
        with tempfile.TemporaryDirectory() as tmp:
            with self.assertRaises(ValueError):generate(Path(tmp)/'FORBIDDEN_PRODUCTION',read(PKG/'REFERENCE_OBJECTS.tsv'),scope,mode='PRODUCTION_AFTER_OBJECT_FREEZE')
            self.assertFalse((Path(tmp)/'FORBIDDEN_PRODUCTION').exists())
    def test_proposal_manifest_not_object_freeze(self):
        with tempfile.TemporaryDirectory() as tmp:
            path=Path(tmp)/'proposal.json';jswrite(path,{'status':'PROPOSED_ONLY','object_protocol_version':OBJECT_VERSION})
            with self.assertRaises(ValueError):gated_endpoints(path)

class FutureFreezeGateTests(unittest.TestCase):
    def test_explicit_zero_addition_synthetic_freeze_gate(self):
        with tempfile.TemporaryDirectory() as tmp:
            path=synthetic_freeze(Path(tmp));rows,h=gated_endpoints(path)
            self.assertEqual(len(rows),520);self.assertEqual(h,digest(path))
    def test_explicit_synthetic_confirmed_new_gate(self):
        with tempfile.TemporaryDirectory() as tmp:
            rows,_=gated_endpoints(synthetic_freeze(Path(tmp),True))
            self.assertEqual(len(rows),521);self.assertEqual(rows[-1]['post_review_decision'],'CONFIRMED_NEW')
    def mutation_blocked(self,key,mutate,with_new=True):
        with tempfile.TemporaryDirectory() as tmp:
            path=synthetic_freeze(Path(tmp),with_new);meta=json.loads(path.read_text())
            table=Path(tmp)/meta[key]['path'];rows=read(table);mutate(rows)
            write(table,rows);rehash_fixture(path,key)
            with self.assertRaises(ValueError):gated_endpoints(path)
    def test_new_uncertain_endpoint_blocked(self):
        self.mutation_blocked('proposed_additions',lambda rows:rows[0].update(addition_status='UNCERTAIN_NEW'))
    def test_new_without_confirmation_blocked(self):
        self.mutation_blocked('confirmed_endpoints',lambda rows:rows[-1].update(post_review_decision=''))
    def test_new_without_reconciliation_blocked(self):
        self.mutation_blocked('confirmed_endpoints',lambda rows:rows[-1].update(reconciliation_status=''))
    def test_new_stable_identity_mismatch_blocked(self):
        self.mutation_blocked('proposed_additions',lambda rows:rows[0].update(bbox_x1='101'))
    def test_baseline_change_blocked_even_with_rehashed_manifest(self):
        self.mutation_blocked('confirmed_endpoints',lambda rows:rows[0].update(bbox_x1='101'))
    def test_incomplete_panel_blocked(self):
        self.mutation_blocked('panel_completion',lambda rows:rows[0].update(completion_status='INCOMPLETE_TECHNICAL'))
    def test_missing_panel_blocked(self):
        self.mutation_blocked('panel_completion',lambda rows:rows.pop())
    def test_wrong_completion_frame_blocked(self):
        self.mutation_blocked('panel_completion',lambda rows:rows[0].update(frame='999'))

class EligibilityAndBaselineTests(unittest.TestCase):
    def test_synthetic_confirmed_whitelist_valid(self):
        scope=confirmed_scope()
        with tempfile.TemporaryDirectory() as tmp:
            path=Path(tmp)/'SYNTHETIC_SCOPE.json';jswrite(path,scope)
            loaded,_=load_scope(scope_file=path)
            self.assertEqual(loaded['panels']['f68r3'],'ELIGIBLE_LIMITED')
            pairs,filtered,retained=queue(read(PKG/'REFERENCE_OBJECTS.tsv'),loaded)
            all_r3=[r for r in pairs+filtered+retained if r['panel']=='f68r3']
            self.assertEqual(len(all_r3),1)
    def test_unknown_whitelist_endpoint_blocked(self):
        scope=confirmed_scope();scope['f68r3_rule']['star_ids']=['HNEW_MISSING'];rehash_rule(scope)
        with self.assertRaises(ValueError):queue(read(PKG/'REFERENCE_OBJECTS.tsv'),scope)
    def test_mask_hash_and_geometry_validation(self):
        for mutation in ('hash','zero_area','outside','string'):
            scope=confirmed_scope('POLYGON_CENTER_MASK');rule=scope['f68r3_rule']
            if mutation=='hash':rule['rule_sha256']='INVALID'
            elif mutation=='zero_area':rule['polygon']=[[1,1],[2,2],[3,3]];rehash_rule(scope)
            elif mutation=='outside':rule['polygon']=[[-1,1],[100,100],[100,1]];rehash_rule(scope)
            else:rule['polygon']=[['1',1],[100,100],[100,1]];rehash_rule(scope)
            with self.subTest(mutation=mutation),tempfile.TemporaryDirectory() as tmp:
                path=Path(tmp)/'SYNTHETIC_SCOPE.json';jswrite(path,scope)
                with self.assertRaises(ValueError):load_scope(scope_file=path)
    def test_synthetic_valid_mask(self):
        with tempfile.TemporaryDirectory() as tmp:
            path=Path(tmp)/'SYNTHETIC_SCOPE.json';jswrite(path,confirmed_scope('POLYGON_CENTER_MASK'))
            self.assertEqual(load_scope(scope_file=path)[0]['f68r3_rule']['kind'],'POLYGON_CENTER_MASK')
    def test_baseline_poisoning_not_hidden_by_updated_preparation_hash(self):
        with tempfile.TemporaryDirectory() as tmp:
            directory=Path(tmp)
            for name in ('INPUT_MANIFEST.tsv','REFERENCE_OBJECTS.tsv','PRIOR_NONCONFIRMED_PROVENANCE.tsv','PREPARATION_MANIFEST.json'):
                import shutil
                shutil.copyfile(PKG/name,directory/name)
            rows=read(directory/'REFERENCE_OBJECTS.tsv');rows[0]['bbox_x1']='101';write(directory/'REFERENCE_OBJECTS.tsv',rows)
            meta=json.loads((directory/'PREPARATION_MANIFEST.json').read_text())
            meta['reference_sha256']=digest(directory/'REFERENCE_OBJECTS.tsv');jswrite(directory/'PREPARATION_MANIFEST.json',meta)
            with self.assertRaises(ValueError):verify_inputs(directory)

if __name__=='__main__':unittest.main()
