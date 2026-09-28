#!/usr/bin/env python3
"""Reproducible read-only AI1/AI2/human spatial comparison."""
from __future__ import annotations
import argparse
from collections import Counter, defaultdict
import importlib.util
import json
import math
import platform
from pathlib import Path
import sys
import xml.etree.ElementTree as ET
import numpy as np

from common import (A, AI1, AI2, HUMAN, ROOT, PACKAGE, CALIBRATION, SEED, BOOTSTRAPS,
                    check, digest, read, write, write_json, unique, verify_inventory, safe_target)
from geometry import box, shape, metrics
from statistics import fraction, summarize, contrast, fisher, holm, categorical, bootstrap, interval

CONFIRMED = {'ACCEPT','MODIFY'}
STRICT = CONFIRMED | {'REJECT'}
STATUS = {'ACCEPT':'CONFIRMED_UNCHANGED','MODIFY':'CONFIRMED_MODIFIED','REJECT':'REJECTED',
          'UNCERTAIN':'UNCERTAIN','NOT_REVIEWED':'NOT_REVIEWED'}
GROUPS = ['BOTH_AI','AI1_ONLY','AI2_ONLY','NEITHER_AI']
METRICS = ['iou','center_distance_px','center_distance_panel_norm','center_distance_shape_norm',
           'width_relative_error','height_relative_error','axial_angle_error_deg']

def load_inputs(manifest):
    registered = {r['path']: r['sha256'] for r in manifest}
    def verified(path):
        check(str(path.relative_to(ROOT)) in registered, f'analysis input not inventoried: {path}')
        return read(path)
    images = unique(verified(HUMAN/'IMAGE_MANIFEST.tsv'),'panel','images')
    dimensions = {p:(int(r['width']),int(r['height'])) for p,r in images.items()}
    for panel,row in images.items():
        check(registered[str((HUMAN/'images'/row['filename']).relative_to(ROOT))] == row['sha256'], f'image digest mismatch: {panel}')
        for pkg in (AI1,AI2):
            check(digest(pkg/'package/crops'/row['filename']) == row['sha256'], f'canonical mismatch: {panel}')
    hm = json.loads((HUMAN/'manifest.json').read_text())
    for relative,expected in hm['source_original_sha256'].items():
        check(registered[relative] == expected, f'upstream digest disagrees with human manifest: {relative}')
    for relative,meta in hm['files'].items():
        check(registered[str((HUMAN/relative).relative_to(ROOT))] == meta['sha256'], f'human file manifest disagreement: {relative}')
    check(hm['active_relation_protocol_version']=='2' and hm['production_status']=='STAR_LABEL_REVIEWED_LAYERS_FROZEN', 'human freeze not ready')
    sources = {'STAR':{},'LABEL':{}}
    for tag,pkg,obj,lab,idlab in [('A',A,'ASTRO_OBJECTS.tsv','ASTRO_LABELS_SPATIAL.tsv','label_occurrence_id'),
                                  ('AI1',AI1,'AI_B_OBJECTS.tsv','AI_B_LABELS.tsv','label_id'),
                                  ('AI2',AI2,'AI2_OBJECTS.tsv','AI2_LABELS.tsv','label_id')]:
        sources['STAR'][tag]=unique(verified(pkg/obj),'object_id',tag+' objects')
        sources['LABEL'][tag]=unique(verified(pkg/lab),idlab,tag+' labels')
        for kind in ('STAR','LABEL'):
            for sid,row in sources[kind][tag].items():
                check(row['panel'] in images, f'unknown source panel: {sid}')
                coords=box(row)
                check(all(math.isfinite(v) for v in coords) and coords[2]>coords[0] and coords[3]>coords[1], f'invalid source bbox: {sid}')
    expected_sources={('STAR','AI1'):293,('STAR','AI2'):330,('LABEL','AI1'):175,('LABEL','AI2'):206}
    for (kind,tag),count in expected_sources.items():
        actual=sum(r['object_class']=='STAR_OBJECT' for r in sources[kind][tag].values()) if kind=='STAR' else len(sources[kind][tag])
        check(actual==count, f'published source count mismatch: {kind}/{tag}: {actual}')
    candidates={'STAR':verified(HUMAN/'HUMAN_CANDIDATE_OBJECTS.tsv'),'LABEL':verified(HUMAN/'HUMAN_CANDIDATE_LABELS.tsv')}
    # Verification only: deterministic constrained reconciliation, no new matching.
    sys.dont_write_bytecode = True  # Never create/update caches inside a frozen package.
    spec=importlib.util.spec_from_file_location('frozen_candidate_builder',HUMAN/'scripts/build_human_candidate_layer.py')
    builder=importlib.util.module_from_spec(spec)
    spec.loader.exec_module(builder)
    pair_specs={
        'STAR':[(AI1/'AI_B_A_MATCH_OBJECTS.tsv','a_object_id','b_object_id','A','AI1'),
                (AI2/'A_AI2_OBJECT_MATCH.tsv','a_object_id','ai2_object_id','A','AI2'),
                (AI2/'AI1_AI2_OBJECT_MATCH.tsv','ai1_object_id','ai2_object_id','AI1','AI2')],
        'LABEL':[(AI1/'AI_B_A_MATCH_LABELS.tsv','a_label_id','b_label_id','A','AI1'),
                 (AI2/'A_AI2_LABEL_MATCH.tsv','a_label_id','ai2_label_id','A','AI2'),
                 (AI2/'AI1_AI2_LABEL_MATCH.tsv','ai1_label_id','ai2_label_id','AI1','AI2')]}
    mappings={}
    audit=[]
    for kind in ('STAR','LABEL'):
        index=unique(candidates[kind],'candidate_id',kind+' physical candidates')
        ids={tag:('object_id' if kind=='STAR' else 'label_occurrence_id' if tag=='A' else 'label_id') for tag in sources[kind]}
        rebuilt, rebuilt_map=builder.build_candidates('OBJECT' if kind=='STAR' else 'LABEL',
                           {t:list(rows.values()) for t,rows in sources[kind].items()},ids,pair_specs[kind])
        rebuilt_index={r['candidate_id']:r for r in rebuilt}
        check(set(rebuilt_index)==set(index), f'frozen reconciliation candidate IDs changed: {kind}')
        source_map={}
        for cid,row in index.items():
            check(row['panel'] in images, f'unknown candidate panel: {cid}')
            nodes=[tuple(t.split(':',1)) for t in row['source_annotation_ids'].split(';')]
            check(len(set(t for t,_ in nodes))==len(nodes),f'multiple records of one source in physical candidate: {cid}')
            check(row['source_annotation_ids']==rebuilt_index[cid]['source_annotation_ids'],f'provenance mismatch: {cid}')
            for field in ('canonical_bbox','provisional_display_bbox','geometry_mode','provisional_rotation'):
                check(row[field]==rebuilt_index[cid][field], f'reconciliation geometry mismatch: {cid}/{field}')
            for tag,sid in nodes:
                check((tag,sid) not in source_map and sid in sources[kind][tag],f'lost/duplicate source: {tag}:{sid}')
                original=sources[kind][tag][sid]
                check(original['panel']==row['panel'],f'cross-panel provenance: {cid}')
                cached=tuple(float(x) for x in row['bbox_'+tag.lower()].split(','))
                check(all(abs(a-b)<=.000001 for a,b in zip(cached,box(original))),f'source bbox provenance mismatch: {cid}')
                check(row['support_'+tag.lower()]=='YES',f'support flag mismatch: {cid}')
                source_map[tag,sid]=cid
            for tag in ('A','AI1','AI2'):
                check((row['support_'+tag.lower()]=='YES')==(tag in {t for t,_ in nodes}),f'support mismatch: {cid}')
            check(int(row['support_count'])==len(nodes),f'support count mismatch: {cid}')
        check(source_map==rebuilt_map, f'complete provenance differs: {kind}')
        mappings[kind]=source_map
        for path,leftcol,rightcol,lt,rt in pair_specs[kind]:
            rows=verified(path)
            seen_left,seen_right=set(),set()
            for row in rows:
                check(row['match_status'] in {'MATCHED','LEFT_ONLY','RIGHT_ONLY','A_ONLY','B_ONLY'}, f'bad pair status: {path}')
                lid,rid=row[leftcol],row[rightcol]
                lc=source_map.get((lt,lid),'')
                rc=source_map.get((rt,rid),'')
                if lid!='UNMATCHED':
                    check(lid not in seen_left and lc, f'pair left missing/duplicate: {lid}')
                    seen_left.add(lid)
                    check(sources[kind][lt][lid]['panel']==row['panel'],f'pair panel mismatch: {lid}')
                if rid!='UNMATCHED':
                    check(rid not in seen_right and rc, f'pair right missing/duplicate: {rid}')
                    seen_right.add(rid)
                    check(sources[kind][rt][rid]['panel']==row['panel'],f'pair panel mismatch: {rid}')
                check((lid!='UNMATCHED' and rid!='UNMATCHED')==(row['match_status']=='MATCHED'), 'inconsistent frozen pair status')
                audit.append({'object_type':kind,'pair_file':str(path.relative_to(ROOT)),'panel':row['panel'],
                              'left_source':lt,'left_id':lid,'right_source':rt,'right_id':rid,'match_status':row['match_status'],
                              'left_candidate_id':lc,'right_candidate_id':rc,
                              'reconciliation_status':'SAME_CANDIDATE' if lc and lc==rc else 'CONSTRAINED_EDGE_NOT_JOINED' if lc and rc else 'UNMATCHED'})
            check(seen_left==set(sources[kind][lt]) and seen_right==set(sources[kind][rt]),f'incomplete frozen pair universe: {path}')
        if kind=='STAR': candidates[kind]=[r for r in candidates[kind] if r['candidate_type']=='STAR_OBJECT']
    check(len(candidates['STAR'])==451 and len(candidates['LABEL'])==291, 'published candidate counts mismatch')
    phase_memberships=defaultdict(list)
    human={}
    phase_counts={}
    for kind in ('STAR','LABEL'):
        final=unique(verified(HUMAN/f'HUMAN_{kind}_ADJUDICATION.tsv'),'candidate_id',kind+' human')
        idx={r['candidate_id']:r for r in candidates[kind]}
        phase_files=[('CALIBRATION','HUMAN_STAR_CALIBRATION_R01.tsv' if kind=='STAR' else 'HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv'),
                     ('HIGH',f'HUMAN_{kind}_HIGH_R01_FINAL.tsv'),('CONSENSUS_QC',f'HUMAN_{kind}_CONSENSUS_QC_R01_FINAL.tsv'),
                     ('MEDIUM',f'HUMAN_{kind}_MEDIUM_R01.tsv')]
        union={}
        for phase,filename in phase_files:
            rows=verified(HUMAN/'exports'/filename)
            unique(rows,'candidate_id',filename)
            phase_counts[kind+'/'+phase]=len(rows)
            for row in rows:
                cid=row['candidate_id']
                check(cid in idx and row['panel']==idx[cid]['panel'], f'unknown/wrong-panel human: {cid}')
                check(row['human_decision'] in STRICT|{'UNCERTAIN'}, f'unhandled human decision: {cid}')
                check(row['final_class']==('STAR_OBJECT' if kind=='STAR' else 'LABEL'),f'unhandled final class: {cid}')
                check(row['human_confidence']=='HIGH', f'confidence override inconsistency: {cid}')
                coords=box(row)
                check(all(math.isfinite(v) for v in coords) and coords[2]>coords[0] and coords[3]>coords[1],f'bad human shape: {cid}')
                check(abs(float(row['center_x'])-(coords[0]+coords[2])/2)<=.001 and abs(float(row['center_y'])-(coords[1]+coords[3])/2)<=.001,f'bad human center: {cid}')
                if kind=='LABEL':
                    check(row['geometry_type'] in {'BOX','ELLIPSE'},f'unknown shape: {cid}')
                    check(row['split_required']=='false' and row['merge_required']=='false',f'unresolved split/merge: {cid}')
                if phase=='CALIBRATION': check(row['panel'] in CALIBRATION, f'calibration outside declared panels: {cid}')
                if phase=='HIGH': check(idx[cid]['priority']=='HIGH', f'High queue priority mismatch: {cid}')
                check(cid not in union or union[cid]==row, f'conflicting phase decisions: {cid}')
                union[cid]=row
                phase_memberships[kind,cid].append(phase)
        check(union==final, f'canonical human union mismatch: {kind}')
        for cid,row in idx.items():
            check((row['panel'] in CALIBRATION)==('CALIBRATION' in phase_memberships[kind,cid]),f'calibration membership mismatch: {cid}')
        human[kind]=final
    expected_decisions={'STAR':{'ACCEPT':115,'MODIFY':206,'REJECT':118,'UNCERTAIN':5},
                        'LABEL':{'ACCEPT':36,'MODIFY':163,'REJECT':88,'UNCERTAIN':3}}
    for kind in human:
        check(dict(Counter(r['human_decision'] for r in human[kind].values()))==expected_decisions[kind],f'published decisions mismatch: {kind}')
    check(phase_counts=={'STAR/CALIBRATION':207,'STAR/HIGH':385,'STAR/CONSENSUS_QC':7,'STAR/MEDIUM':24,
                         'LABEL/CALIBRATION':93,'LABEL/HIGH':271,'LABEL/CONSENSUS_QC':2,'LABEL/MEDIUM':6},'phase manifest counts mismatch')
    unreviewed=verified(HUMAN/'HUMAN_PRODUCTION_UNREVIEWED.tsv')
    check({(r['task'],r['candidate_id']) for r in unreviewed}=={(k,r['candidate_id']) for k,rows in candidates.items() for r in rows if r['candidate_id'] not in human[k]},'not-reviewed universe mismatch')
    check(all(r['status']=='NOT_REVIEWED' and r['priority']=='LOW' for r in unreviewed),'invalid not-reviewed state')
    # Individual review tasks themselves are registered rather than assuming separate manifests exist.
    phases=[]
    for kind in ('STAR','LABEL'):
        for phase,file in [('CALIBRATION',kind.lower()+'_calibration_h1.xml'),('HIGH',kind.lower()+'_high_h1.xml'),
                           ('CONSENSUS_QC',kind.lower()+'_consensus_qc_h1.xml'),('MEDIUM',kind.lower()+'_medium_pending_h1.xml')]:
            path=HUMAN/'cvat'/file
            verified_path=str(path.relative_to(ROOT))
            check(verified_path in registered,f'phase task not registered: {path}')
            xml=ET.parse(path).getroot()
            ids={a.text for a in xml.findall('.//attribute[@name="candidate_id"]')}
            decision_name=('HUMAN_STAR_CALIBRATION_R01.tsv' if kind=='STAR' else 'HUMAN_LABEL_CALIBRATION_R01_FINAL.tsv') if phase=='CALIBRATION' else f'HUMAN_{kind}_{phase}_R01' + ('.tsv' if phase=='MEDIUM' else '_FINAL.tsv')
            decision_path=HUMAN/'exports'/decision_name
            phases.append({'object_type':kind,'phase':phase,'task_path':verified_path,'sha256':registered[verified_path],
                           'decision_table':str(decision_path.relative_to(ROOT)), 'decision_table_sha256':registered[str(decision_path.relative_to(ROOT))],
                           'global_freeze_manifest':str((HUMAN/'manifest.json').relative_to(ROOT)),
                           'task_candidate_count':len(ids),'final_phase_record_count':phase_counts[kind+'/'+phase],
                           'note':'QC task is original full sampled queue; completed pending subset and prior identical reviews deduplicated' if phase=='CONSENSUS_QC' else 'frozen prepared task'})
            if phase in {'CALIBRATION','HIGH','MEDIUM'}:
                reviewed={cid for (k,cid),members in phase_memberships.items() if k==kind and phase in members}
                check(ids==reviewed,f'phase task/records mismatch: {kind}/{phase}')
            else:
                check(len(ids)==(34 if kind=='STAR' else 18) and ids<=set(human[kind]),f'incomplete full QC sample: {kind}')
                for cid in sorted(ids):
                    if 'CONSENSUS_QC' not in phase_memberships[kind,cid]:
                        phase_memberships[kind,cid].append('CONSENSUS_QC')
    return sources,candidates,human,phase_memberships,dimensions,mappings,audit,phases

def build_objects(sources,candidates,human,memberships):
    results=[]
    cohorts=[]
    for kind,rows in candidates.items():
        for c in sorted(rows,key=lambda r:(r['panel'],r['candidate_id'])):
            cid=c['candidate_id']
            h=human[kind].get(cid)
            decision=h['human_decision'] if h else 'NOT_REVIEWED'
            pattern='BOTH_AI' if c['support_ai1']=='YES' and c['support_ai2']=='YES' else 'AI1_ONLY' if c['support_ai1']=='YES' else 'AI2_ONLY' if c['support_ai2']=='YES' else 'NEITHER_AI'
            scope='CALIBRATION' if c['panel'] in CALIBRATION else 'PRODUCTION'
            members=memberships[kind,cid]
            stage=members[0] if members else 'NOT_REVIEWED'
            base={'candidate_id':cid,'panel':c['panel'],'object_type':kind,'scope':scope,'support_pattern':pattern,
                  'a_positive_support':c['support_a'],'support_ai1':c['support_ai1'],'support_ai2':c['support_ai2'],
                  'priority':c['priority'],'priority_reason':c['priority_reason'],'review_stage':stage,
                  'review_phase_memberships':';'.join(members),'human_decision':decision,'analytic_status':STATUS[decision],
                  'geometry_type':h.get('geometry_type','AABB') if h else c['geometry_mode'],
                  'source_annotation_ids':c['source_annotation_ids'],'human_protocol_version':'1.2',
                  'primary_strict':int(scope=='PRODUCTION' and decision in STRICT),
                  'combined_strict':int(decision in STRICT),'uncertain_cohort':int(decision=='UNCERTAIN'),
                  'not_reviewed_cohort':int(decision=='NOT_REVIEWED')}
            cohorts.append(base.copy())
            result=base|{'candidate_display_bbox':c['provisional_display_bbox'],'candidate_union_bbox':c['canonical_bbox'],
                         'candidate_rotation':c['provisional_rotation'],'candidate_geometry_mode':c['geometry_mode'],
                         'human_bbox':','.join(h[k] for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')) if h else '',
                         'human_rotation':h.get('rotation','0') if h else '', 'human_final_class':h['final_class'] if h else '',
                         'a_absence_negative_evidence':0}
            for tag in ('A','AI1','AI2'):
                node=next((sid for t,sid in (x.split(':',1) for x in c['source_annotation_ids'].split(';')) if t==tag),'')
                src=sources[kind][tag].get(node,{})
                result[tag.lower()+'_id']=node
                result[tag.lower()+'_class']=src.get('object_class','LABEL' if node and kind=='LABEL' else '')
                result[tag.lower()+'_bbox']=','.join(str(v) for v in box(src)) if node else ''
                result[tag.lower()+'_orientation']=src.get('orientation_angle','0') if node else ''
                result[tag.lower()+'_recorded_center']=','.join(src[k] for k in ('center_x','center_y')) if node else ''
                result[tag.lower()+'_confidence']=src.get('confidence',src.get('annotation_confidence',''))
            results.append(result)
    return results,cohorts

def build_geometry(objects,candidates,human,sources,dimensions):
    idx={r['candidate_id']:r for rows in candidates.values() for r in rows}
    results=[]
    for obj in objects:
        if obj['human_decision'] not in CONFIRMED: continue
        kind,cid=obj['object_type'],obj['candidate_id']
        h=human[kind][cid]
        gkind='AABB' if kind=='STAR' else 'ELLIPSE_ENVELOPE_PROXY' if h['geometry_type']=='ELLIPSE' else 'ROTATED_RECTANGLE'
        target=shape(h,gkind)
        c=idx[cid]
        ckind='AABB' if kind=='STAR' else 'ELLIPSE_ENVELOPE_PROXY' if c['geometry_mode']=='ELLIPSE_RING' else 'ROTATED_RECTANGLE'
        comparisons=[('CANDIDATE',cid,shape(c,ckind,c['provisional_rotation']),'actual_candidate_display')]
        for tag in ('AI1','AI2'):
            sid=obj[tag.lower()+'_id']
            if not sid: continue
            src=sources[kind][tag][sid]
            if kind=='STAR': comparisons.append((tag,sid,shape(src,'AABB'),'raw_AABB'))
            elif gkind=='ELLIPSE_ENVELOPE_PROXY': comparisons.append((tag,sid,shape(src,gkind,0),'filled_ellipse_envelope_proxy'))
            else:
                comparisons.extend([(tag,sid,shape(src,gkind,convention='orientation_normalized_proxy'),'orientation_normalized_proxy'),
                                    (tag,sid,shape(src,gkind),'raw_dimensions_rotated_proxy')])
        for tag,sid,source,convention in comparisons:
            row={k:obj[k] for k in ('candidate_id','panel','object_type','scope','support_pattern','a_positive_support','priority','human_decision')}
            row.update(source=tag,source_id=sid,human_id=cid,geometry_type=gkind,source_geometry_type=source['kind'],
                       representation=convention,primary_representation=int(convention!='raw_dimensions_rotated_proxy'),
                       source_raw_bbox=obj[tag.lower()+'_bbox'] if tag in {'AI1','AI2'} else obj['candidate_display_bbox'],
                       final_human_bbox=obj['human_bbox'],source_orientation=source['angle'],human_orientation=target['angle'],
                       human_protocol_version='1.2',metric_limit='filled_envelope_not_ring_ink' if gkind=='ELLIPSE_ENVELOPE_PROXY' else 'AI_oriented_side_lengths_not_explicit_proxy' if kind=='LABEL' and tag!='CANDIDATE' else 'actual_frozen_shape')
            row.update(metrics(source,target,dimensions[obj['panel']]))
            # Raw source bbox compared to actual human shape's axis-aligned envelope, separate diagnostic.
            if tag in {'AI1','AI2'}:
                polygon=target['polygon']
                envelope={'bbox_x1':min(p[0] for p in polygon),'bbox_y1':min(p[1] for p in polygon),
                          'bbox_x2':max(p[0] for p in polygon),'bbox_y2':max(p[1] for p in polygon)}
                row['raw_envelope_iou']=metrics(shape(sources[kind][tag][sid],'AABB'),shape(envelope,'AABB'),dimensions[obj['panel']])['iou']
            else: row['raw_envelope_iou']=None
            results.append(row)
    return results

def scope_rows(objects,scope):
    return [r for r in objects if scope=='COMBINED' or r['scope']==scope]

def summarize_objects(objects):
    output={}
    for scope in ('PRODUCTION','CALIBRATION','COMBINED'):
        output[scope]={}
        for kind in ('STAR','LABEL'):
            rows=[r for r in scope_rows(objects,scope) if r['object_type']==kind]
            by_ai={}
            for tag in ('AI1','AI2'):
                supported=[r for r in rows if r['support_'+tag.lower()]=='YES']
                reviewed=[r for r in supported if r['human_decision']!='NOT_REVIEWED']
                strict=[r for r in reviewed if r['human_decision'] in STRICT]
                confirmed=[r for r in strict if r['human_decision'] in CONFIRMED]
                by_ai[tag]={'source_candidates_in_scope':len(supported),'reviewed_candidates':len(reviewed),
                            'not_reviewed':len(supported)-len(reviewed),'strict_denominator':len(strict),
                            'decision_counts':dict(Counter(r['human_decision'] for r in supported))}
                for decision,name in [('CONFIRMED','confirmation_strict'),('REJECT','rejection_strict'),('ACCEPT','unchanged_strict'),('MODIFY','modified_strict')]:
                    values=[int(r['human_decision'] in CONFIRMED if decision=='CONFIRMED' else r['human_decision']==decision) for r in strict]
                    by_ai[tag][name]=fraction(sum(values),len(strict),values if name=='confirmation_strict' else None,
                                             [r['panel'] for r in strict],f'{scope}/{kind}/{tag}/{name}')
                by_ai[tag]['uncertainty_all_reviewed']=fraction(sum(r['human_decision']=='UNCERTAIN' for r in reviewed),len(reviewed))
                by_ai[tag]['confirmation_all_reviewed']=fraction(len(confirmed),len(reviewed))
                by_ai[tag]['modification_among_confirmed']=fraction(sum(r['human_decision']=='MODIFY' for r in confirmed),len(confirmed))
            confirmed=[r for r in rows if r['human_decision'] in CONFIRMED]
            coverage={}
            for name,predicate in [('AI1',lambda r:r['support_ai1']=='YES'),('AI2',lambda r:r['support_ai2']=='YES'),
                                   ('BOTH_AI',lambda r:r['support_pattern']=='BOTH_AI'),('AI1_ONLY',lambda r:r['support_pattern']=='AI1_ONLY'),
                                   ('AI2_ONLY',lambda r:r['support_pattern']=='AI2_ONLY'),('A_POSITIVE',lambda r:r['a_positive_support']=='YES')]:
                values=[int(predicate(r)) for r in confirmed]
                coverage[name]=fraction(sum(values),len(confirmed),values,[r['panel'] for r in confirmed],f'coverage/{scope}/{kind}/{name}')
            by_support={}
            for group in GROUPS:
                cells=[r for r in rows if r['support_pattern']==group]
                strict=[r for r in cells if r['human_decision'] in STRICT]
                confirmed_cells=[r for r in cells if r['human_decision'] in CONFIRMED]
                by_support[group]={'candidate_count':len(cells),'outcomes':dict(Counter(r['human_decision'] for r in cells)),
                                   'confirmation_strict':fraction(len(confirmed_cells),len(strict)),
                                   'modification_among_confirmed':fraction(sum(r['human_decision']=='MODIFY' for r in confirmed_cells),len(confirmed_cells))}
            output[scope][kind]={'candidate_union':len(rows),'reviewed':sum(r['human_decision']!='NOT_REVIEWED' for r in rows),
                                 'confirmed_union':len(confirmed),'outcomes':dict(Counter(r['human_decision'] for r in rows)),
                                 'AI':by_ai,'support_group_metrics':by_support,'conditional_coverage':coverage}
    return output

def stratify(objects):
    rows=[]
    dimensions=['panel','object_type','priority','support_pattern','geometry_type','scope','human_decision','review_stage','a_positive_support']
    for variable in dimensions:
        for (kind,scope,value),cells in sorted(group_rows(objects,lambda r:(r['object_type'],r['scope'],r[variable])).items()):
            counts=Counter(r['human_decision'] for r in cells)
            n=sum(counts[d] for d in STRICT)
            k=sum(counts[d] for d in CONFIRMED)
            ci=fraction(k,n)
            rows.append({'stratum_field':variable,'stratum_value':value,'object_type':kind,'scope':scope,
                         'candidates':len(cells),'strict_reviewed':n,'confirmed':k,'rejected':counts['REJECT'],
                         'uncertain':counts['UNCERTAIN'],'not_reviewed':counts['NOT_REVIEWED'],
                         'accept':counts['ACCEPT'],'modify':counts['MODIFY'],'confirmation_strict':ci['estimate'],
                         'ci95_low':ci['ci95'][0],'ci95_high':ci['ci95'][1],'ci_method':ci['ci_method']})
    return rows

def group_rows(rows,key):
    out=defaultdict(list)
    for row in rows: out[key(row)].append(row)
    return out

def summarize_geometry(rows):
    summaries=[]
    for scope in ('PRODUCTION','CALIBRATION','COMBINED'):
        selected=[r for r in rows if scope=='COMBINED' or r['scope']==scope]
        for grouping in ('ALL','DECISION','PANEL'):
            grouped=group_rows(selected,lambda r:(r['object_type'],r['source'],r['geometry_type'],r['representation'],
                                                  r['human_decision'] if grouping=='DECISION' else r['panel'] if grouping=='PANEL' else 'ALL'))
            for key,group in sorted(grouped.items()):
                kind,source,gtype,rep,stratum=key
                for metric in METRICS+['raw_envelope_iou']:
                    usable=[r for r in group if r.get(metric) is not None]
                    if not usable: continue
                    values=[r[metric] for r in usable]
                    result=summarize(values,[r['panel'] for r in usable],'/'.join((scope,grouping,*key,metric)))
                    ci=result.pop('ci95')
                    summaries.append({'scope':scope,'object_type':kind,'source':source,'geometry_type':gtype,
                                      'representation':rep,'grouping':grouping,'stratum':stratum,'metric':metric,
                                      **result,'ci95_low':ci[0],'ci95_high':ci[1]})
    return summaries

def pairwise_results(sources,mappings,objects,dimensions):
    obj_index={r['candidate_id']:r for r in objects}
    rows=[]
    summaries=[]
    confidences={}
    for kind,path,lc,rc in [('STAR',AI2/'AI1_AI2_OBJECT_MATCH.tsv','ai1_object_id','ai2_object_id'),
                            ('LABEL',AI2/'AI1_AI2_LABEL_MATCH.tsv','ai1_label_id','ai2_label_id')]:
        for pair in read(path):
            left=sources[kind]['AI1'].get(pair[lc])
            right=sources[kind]['AI2'].get(pair[rc])
            if kind=='STAR' and (left or right)['object_class']!='STAR_OBJECT': continue
            lid=mappings[kind].get(('AI1',pair[lc]),'')
            rid=mappings[kind].get(('AI2',pair[rc]),'')
            if left and right:
                actual=metrics(shape(left),shape(right),dimensions[pair['panel']])['iou']
                check(abs(actual-float(pair['iou']))<1e-7,f'frozen pair IoU mismatch: {pair[lc]}')
            rows.append({'object_type':kind,'panel':pair['panel'],'scope':'CALIBRATION' if pair['panel'] in CALIBRATION else 'PRODUCTION',
                         'ai1_id':pair[lc] if left else '', 'ai2_id':pair[rc] if right else '',
                         'ai1_candidate_id':lid,'ai2_candidate_id':rid,'match_status':pair['match_status'],
                         'frozen_iou':float(pair['iou']) if left and right else None,
                         'center_distance_px':float(pair['center_distance']) if left and right else None,
                         'ai1_human_outcome':obj_index[lid]['human_decision'] if lid in obj_index else 'NO_HUMAN_SCOPE',
                         'ai2_human_outcome':obj_index[rid]['human_decision'] if rid in obj_index else 'NO_HUMAN_SCOPE',
                         'class_agreement':int(left['object_class']==right['object_class']) if kind=='STAR' and left and right else None,
                         'ai1_confidence':left.get('confidence','') if left else '', 'ai2_confidence':right.get('confidence','') if right else '',
                         'same_physical_candidate':int(lid==rid) if left and right else None})
        for scope in ('PRODUCTION','CALIBRATION','COMBINED'):
            for panel in ['ALL']+sorted(dimensions):
                subset=[r for r in rows if r['object_type']==kind and (scope=='COMBINED' or r['scope']==scope) and (panel=='ALL' or r['panel']==panel)]
                if not subset: continue
                m=[r for r in subset if r['match_status']=='MATCHED']
                n1=sum(bool(r['ai1_id']) for r in subset)
                n2=sum(bool(r['ai2_id']) for r in subset)
                ci=summarize([r['frozen_iou'] for r in m],[r['panel'] for r in m],f'pair/{kind}/{scope}/{panel}')
                summaries.append({'object_type':kind,'scope':scope,'panel':panel,'ai1_count':n1,'ai2_count':n2,'matched_pairs':len(m),
                                  'left_only':n1-len(m),'right_only':n2-len(m),'matched_fraction_ai1':fraction(len(m),n1),
                                  'matched_fraction_ai2':fraction(len(m),n2),'matched_jaccard':fraction(len(m),n1+n2-len(m)),
                                  'raw_AABB_iou':ci,'class_agreement_conditioned':fraction(len(m),len(m)),
                                  'class_agreement_status':'CONDITIONED_ON_SAME_CLASS_MATCHING' if kind=='STAR' else 'FIXED_LABEL_CLASS'})
            compatible=[(r['ai1_confidence'],r['ai2_confidence']) for r in rows if r['object_type']==kind and r['match_status']=='MATCHED' and
                         (scope=='COMBINED' or r['scope']==scope) and r['ai1_confidence'] and r['ai2_confidence']]
            confidences[f'{scope}/{kind}']=categorical(compatible,sorted({c for pair in compatible for c in pair}))
    return rows,summaries,confidences

def tests(objects,geometry):
    results=[]
    contrasts=[('BOTH_AI','AI1_ONLY'),('BOTH_AI','AI2_ONLY'),('AI1_ONLY','AI2_ONLY')]
    for kind in ('STAR','LABEL'):
        prod=[r for r in objects if r['scope']=='PRODUCTION' and r['object_type']==kind and r['human_decision']!='NOT_REVIEWED']
        for left,right in contrasts:
            for outcome in ('CONFIRMATION','REJECTION','UNCERTAINTY','MODIFY_AMONG_CONFIRMED'):
                eligible=[r for r in prod if r['human_decision'] in CONFIRMED] if outcome=='MODIFY_AMONG_CONFIRMED' else prod if outcome=='UNCERTAINTY' else [r for r in prod if r['human_decision'] in STRICT]
                l=[r for r in eligible if r['support_pattern']==left]
                r=[r for r in eligible if r['support_pattern']==right]
                def value(x):
                    return int(x['human_decision'] in CONFIRMED) if outcome=='CONFIRMATION' else int(x['human_decision']=={'REJECTION':'REJECT','UNCERTAINTY':'UNCERTAIN','MODIFY_AMONG_CONFIRMED':'MODIFY'}[outcome])
                lv,rv=[value(x) for x in l],[value(x) for x in r]
                tid=f'{kind}/{left}-{right}/{outcome}'
                effect,ci,_,valid,method=contrast(lv,[x['panel'] for x in l],rv,[x['panel'] for x in r],tid)
                p=fisher(sum(lv),len(l)-sum(lv),sum(rv),len(r)-sum(rv)) if l and r else None
                status='INSUFFICIENT_SMALL_GROUP' if min(len(l),len(r))<10 or valid<1000 else 'BOUNDARY_DEGENERATE_BOOTSTRAP_NOT_EQUIVALENCE' if ci[0]==ci[1] else 'ESTIMATED_FEW_PANEL_CLUSTERS'
                results.append({'test_id':tid,'family':kind+'_categorical','scope':'PRODUCTION','object_type':kind,'outcome':outcome,
                                'left_group':left,'right_group':right,'n_left':len(l),'n_right':len(r),'events_left':sum(lv),'events_right':sum(rv),
                                'effect_type':'risk_difference_left_minus_right','effect':effect,'ci95_low':ci[0],'ci95_high':ci[1],
                                'ci_method':method,'valid_bootstraps':valid,'p_method':'Fisher_exact_object_independence_supplementary',
                                'p_raw':p,'p_holm':None,'status':status})
        corrections=[r for r in geometry if r['object_type']==kind and r['scope']=='PRODUCTION' and r['source']=='CANDIDATE']
        for left,right in contrasts:
            l=[x for x in corrections if x['support_pattern']==left]
            r=[x for x in corrections if x['support_pattern']==right]
            tid=f'{kind}/{left}-{right}/CORRECTION_1_IOU'
            effect,ci,p,valid,method=contrast([1-x['iou'] for x in l],[x['panel'] for x in l],[1-x['iou'] for x in r],[x['panel'] for x in r],tid,'median')
            results.append({'test_id':tid,'family':kind+'_continuous','scope':'PRODUCTION','object_type':kind,'outcome':'CORRECTION_1_IOU',
                            'left_group':left,'right_group':right,'n_left':len(l),'n_right':len(r),'events_left':None,'events_right':None,
                            'effect_type':'difference_of_medians_left_minus_right','effect':effect,'ci95_low':ci[0],'ci95_high':ci[1],
                            'ci_method':method,'valid_bootstraps':valid,'p_method':'bootstrap_sign_tail_exploratory',
                            'p_raw':p,'p_holm':None,'status':'INSUFFICIENT_SMALL_GROUP' if min(len(l),len(r))<10 or valid<1000 else 'EXPLORATORY_FEW_PANEL_CLUSTERS'})
        paired=group_rows([r for r in geometry if r['object_type']==kind and r['scope']=='PRODUCTION' and r['source'] in {'AI1','AI2'} and r['primary_representation']],lambda r:r['candidate_id'])
        complete=[rows for rows in paired.values() if {r['source'] for r in rows}=={'AI1','AI2'}]
        differences=[]
        panels=[]
        for rows in complete:
            by={r['source']:r for r in rows}
            differences.append(by['AI1']['iou']-by['AI2']['iou'])
            panels.append(rows[0]['panel'])
        tid=f'{kind}/PAIRED_AI1_AI2/HUMAN_IOU'
        boot,method=bootstrap(differences,panels,tid)
        ci=interval(boot)
        p=min(1.,2*min((np.sum(boot<=0)+1)/(len(boot)+1),(np.sum(boot>=0)+1)/(len(boot)+1))) if len(boot) else None
        results.append({'test_id':tid,'family':kind+'_continuous','scope':'PRODUCTION','object_type':kind,'outcome':'PAIRED_HUMAN_IOU',
                        'left_group':'AI1','right_group':'AI2','n_left':len(differences),'n_right':len(differences),'events_left':None,'events_right':None,
                        'effect_type':'median_paired_difference_AI1_minus_AI2','effect':float(np.median(differences)) if differences else None,
                        'ci95_low':ci[0],'ci95_high':ci[1],'ci_method':method,'valid_bootstraps':len(boot),
                        'p_method':'bootstrap_sign_tail_exploratory','p_raw':p,'p_holm':None,
                        'status':'INSUFFICIENT_SMALL_GROUP' if len(differences)<10 else 'EXPLORATORY_GEOMETRY_PROXY' if kind=='LABEL' else 'EXPLORATORY_FEW_PANEL_CLUSTERS'})
    holm(results)
    return results

def sensitivity(objects,geometry):
    policies=[('PRIMARY_PRODUCTION_STRICT',lambda r:r['scope']=='PRODUCTION','exclude'),
              ('UNCERTAIN_NEGATIVE_BOUND',lambda r:r['scope']=='PRODUCTION','negative'),
              ('UNCERTAIN_POSITIVE_BOUND',lambda r:r['scope']=='PRODUCTION','positive'),
              ('CALIBRATION_ONLY_STRICT',lambda r:r['scope']=='CALIBRATION','exclude'),
              ('COMBINED_STRICT',lambda r:True,'exclude'),
              ('PRODUCTION_HIGH_ONLY',lambda r:r['scope']=='PRODUCTION' and 'HIGH' in r['review_phase_memberships'].split(';'),'exclude'),
              ('PRODUCTION_HIGH_MEDIUM_DEDUP',lambda r:r['scope']=='PRODUCTION' and bool({'HIGH','MEDIUM'} & set(r['review_phase_memberships'].split(';'))),'exclude'),
              ('PRODUCTION_WITHOUT_A_POSITIVE',lambda r:r['scope']=='PRODUCTION' and r['a_positive_support']=='NO','exclude'),
              ('PRODUCTION_A_POSITIVE_ONLY',lambda r:r['scope']=='PRODUCTION' and r['a_positive_support']=='YES','exclude')]
    rows=[]
    for name,selector,uncertain in policies:
        for kind in ('STAR','LABEL'):
            selected=[r for r in objects if r['object_type']==kind and selector(r)]
            for group in ['AI1','AI2']+GROUPS:
                supported=[r for r in selected if r['support_'+group.lower()]=='YES'] if group in {'AI1','AI2'} else [r for r in selected if r['support_pattern']==group]
                eligible=[r for r in supported if r['human_decision'] in STRICT or (uncertain!='exclude' and r['human_decision']=='UNCERTAIN')]
                k=sum(r['human_decision'] in CONFIRMED or (uncertain=='positive' and r['human_decision']=='UNCERTAIN') for r in eligible)
                f=fraction(k,len(eligible))
                rows.append({'policy':name,'object_type':kind,'support_group':group,'uncertain_policy':uncertain,
                             'denominator':len(eligible),'confirmed_or_bound_numerator':k,'estimate':f['estimate'],
                             'ci95_low':f['ci95'][0],'ci95_high':f['ci95'][1],'ci_method':f['ci_method'],
                             'excluded_not_reviewed':sum(r['human_decision']=='NOT_REVIEWED' for r in supported),
                             'unresolved_in_selected_scope':sum(r['human_decision']=='UNCERTAIN' for r in supported)})
    geom=[]
    for (kind,source,representation),cells in sorted(group_rows([r for r in geometry if r['scope']=='PRODUCTION' and r['source']!='CANDIDATE'],lambda r:(r['object_type'],r['source'],r['representation'])).items()):
        for threshold in (.25,.5,.75):
            for metric in ('iou','raw_envelope_iou'):
                values=[r[metric] for r in cells if r[metric] is not None]
                f=fraction(sum(v>=threshold for v in values),len(values))
                geom.append({'scope':'PRODUCTION','object_type':kind,'source':source,'representation':representation,'metric':metric,
                             'threshold':threshold,'numerator':f['numerator'],'denominator':f['denominator'],'estimate':f['estimate'],
                             'ci95_low':f['ci95'][0],'ci95_high':f['ci95'][1],'note':'diagnostic thresholds only; provenance matching unchanged'})
    return rows,geom

def attachment(human,candidates):
    records=read(HUMAN/'HUMAN_LABEL_OBJECT_RELATIONS.tsv')
    unique(records,'relation_candidate_id','attachment v2')
    counts=Counter(r['relation_type'] for r in records)
    check(counts=={'VISUAL_LABEL_OF':53,'UNASSIGNED':171,'UNCERTAIN':43},'attachment snapshot count mismatch')
    filtered=read(HUMAN/'relations_v2/RIGHT_RULE_FILTERED.tsv')
    excluded=read(HUMAN/'relations/LABEL_STAR_RELATION_EXCLUSIONS.tsv')
    eligible=read(HUMAN/'relations/LABEL_STAR_RELATION_QUEUE.tsv')
    original=read(HUMAN/'HUMAN_CANDIDATE_RELATIONS.tsv')
    sets=[{r['relation_candidate_id'] for r in data} for data in (records,filtered,excluded)]
    check(len(filtered)==57 and len(excluded)==351 and len(eligible)==324 and len(original)==675,'attachment scope count mismatch')
    check(not any(sets[i]&sets[j] for i in range(3) for j in range(i)), 'relation cohorts overlap')
    check(sets[0]|sets[1]=={r['relation_candidate_id'] for r in eligible} and set.union(*sets)=={r['relation_candidate_id'] for r in original},'relation universe mismatch')
    check(all(r['status']=='RULE_FILTERED_NOT_REVIEWED' for r in filtered),'filtered relation promoted')
    label_idx={r['candidate_id']:r for r in candidates['LABEL']}
    star_idx={r['candidate_id']:r for r in candidates['STAR']}
    panel_counts=defaultdict(Counter)
    adjacency=defaultdict(set)
    edge_rows=[]
    for row in records:
        check(row['relation_protocol_version']=='2','mixed relation protocols')
        lid,sid=row['label_candidate_id'],row['object_candidate_id']
        check(lid in human['LABEL'] and sid in human['STAR'] and human['LABEL'][lid]['human_decision'] in CONFIRMED and human['STAR'][sid]['human_decision'] in CONFIRMED,'unconfirmed attachment endpoint')
        panel=human['LABEL'][lid]['panel']
        check(panel==human['STAR'][sid]['panel'],'cross-panel attachment')
        panel_counts[panel][row['relation_type']]+=1
        if row['relation_type']=='VISUAL_LABEL_OF':
            adjacency['LABEL:'+lid].add('STAR:'+sid)
            adjacency['STAR:'+sid].add('LABEL:'+lid)
        edge_rows.append({'relation_candidate_id':row['relation_candidate_id'],'panel':panel,'label_id':lid,'star_id':sid,
                          'decision':row['relation_type'],'decision_origin':row['decision_origin'],'relation_protocol_version':'2'})
    compatible=[]
    ai_audit=[]
    for tag,path in [('AI1',AI1/'AI_B_LABEL_OBJECT_RELATIONS.tsv'),('AI2',AI2/'AI2_LABEL_OBJECT_RELATIONS.tsv')]:
        predictions=read(path)
        types=sorted({r['relation_type'] for r in predictions})
        versions=sorted({r.get('relation_protocol_version','NOT_DECLARED') for r in predictions})
        valid=all(r.get('relation_protocol_version')=='2' and r['relation_type'] in {'VISUAL_LABEL_OF','UNASSIGNED','UNCERTAIN'} for r in predictions)
        if valid: compatible.append(tag)
        ai_audit.append({'source':tag,'path':str(path.relative_to(ROOT)),'count':len(predictions),'relation_types':types,
                         'declared_protocol_versions':versions,'compatible_v2':valid,
                         'reason':'legacy spatial predicates; no independent caption-v2 predictions' if not valid else 'requires endpoint audit'})
    check(not compatible,'compatible relation-v2 predictions need a separate approved comparison plan')
    nodes=[]
    for kind,index in [('LABEL',human['LABEL']),('STAR',human['STAR'])]:
        for cid,h in sorted(index.items()):
            if h['human_decision'] not in CONFIRMED: continue
            node=kind+':'+cid
            nodes.append({'node_id':node,'object_type':kind,'candidate_id':cid,'panel':h['panel'],
                          'degree':len(adjacency.get(node,set())),'relation_protocol_version':'2',
                          'zero_degree_note':'no confirmed link in selected source-derived queue, not no real caption'})
    visited=set()
    components=[]
    component_map={}
    for start in sorted(adjacency):
        if start in visited: continue
        stack=[start]
        current=set()
        while stack:
            node=stack.pop()
            if node in current: continue
            current.add(node)
            stack.extend(sorted(adjacency[node]-current))
        visited|=current
        component_id=f'COMP_{len(components)+1:03d}'
        for node in current: component_map[node]=component_id
        ls=sum(n.startswith('LABEL:') for n in current)
        ss=len(current)-ls
        edges=sum(len(adjacency[n]) for n in current)//2
        components.append({'component_id':component_id,'labels':ls,'stars':ss,'edges':edges,
                           'node_ids':';'.join(sorted(current)),'relation_protocol_version':'2'})
    for row in edge_rows: row['positive_component_id']=component_map.get('LABEL:'+row['label_id'],'') if row['decision']=='VISUAL_LABEL_OF' else ''
    degrees={kind:dict(sorted(Counter(n['degree'] for n in nodes if n['object_type']==kind).items())) for kind in ('LABEL','STAR')}
    positive_labels=sum(n['degree']>0 for n in nodes if n['object_type']=='LABEL')
    positive_stars=sum(n['degree']>0 for n in nodes if n['object_type']=='STAR')
    check(positive_labels==51 and positive_stars==51,'attachment graph endpoint counts mismatch')
    graph={'analysis':'DESCRIPTIVE','relation_protocol_version':'2','ai_compatibility_audit':ai_audit,
           'decision_counts':dict(counts),'decision_origin_counts':dict(Counter(r['decision_origin'] for r in records)),
           'explicit_decisions':len(records),'eligible_source_pairs':len(eligible),'original_source_pairs':len(original),
           'filtered_not_reviewed':len(filtered),'endpoint_excluded_pairs':len(excluded),
           'explicit_review_coverage_eligible':fraction(len(records),len(eligible)),
           'explicit_review_coverage_original':fraction(len(records),len(original)),
           'individual_review_coverage_eligible':fraction(125,324),'individual_selected_task_coverage':fraction(125,125),
           'uncertainty_explicit':fraction(43,267),'uncertainty_individual_review':fraction(43,125),
           'confirmed_links':53,'positive_labels':51,'positive_stars':51,'degrees_all_confirmed_endpoints':degrees,
           'stars_with_multiple_labels':sum(n['degree']>1 for n in nodes if n['object_type']=='STAR'),
           'labels_with_multiple_stars':sum(n['degree']>1 for n in nodes if n['object_type']=='LABEL'),
           'positive_connected_components':len(components),'component_shapes':dict(Counter(f"L{r['labels']}_S{r['stars']}_E{r['edges']}" for r in components)),
           'panels':{p:dict(c) for p,c in sorted(panel_counts.items())}}
    check(graph['stars_with_multiple_labels']==2,'declared multi-label-star count mismatch')
    descriptive=[]
    def result(section,key,n,den=None,note=''):
        ci=fraction(n,den) if den is not None else {'estimate':None,'ci95':[None,None]}
        descriptive.append({'section':section,'key':str(key),'numerator':n,'denominator':den,'estimate':ci['estimate'],
                            'ci95_low':ci['ci95'][0],'ci95_high':ci['ci95'][1],'note':note,'relation_protocol_version':'2'})
    for p,counts in sorted(panel_counts.items()):
        for d in ('VISUAL_LABEL_OF','UNASSIGNED','UNCERTAIN'): result('panel_decisions',p+'/'+d,counts[d],sum(counts.values()))
    for kind,dist in degrees.items():
        for degree,n in dist.items(): result('degree_all_confirmed_endpoints',kind+'/'+str(degree),n,sum(dist.values()),'degree zero = no positive link in selected queue only')
    for row in components: result('positive_component_shape',row['component_id'],row['edges'],None,f"{row['labels']} LABEL; {row['stars']} STAR")
    for key in ('explicit_review_coverage_eligible','explicit_review_coverage_original','individual_review_coverage_eligible','individual_selected_task_coverage','uncertainty_explicit','uncertainty_individual_review'):
        f=graph[key]
        result('coverage',key,f['numerator'],f['denominator'])
    result('excluded','RULE_FILTERED_NOT_REVIEWED',57)
    result('excluded','ENDPOINT_EXCLUDED_NOT_REVIEWED',351)
    return graph,descriptive,edge_rows,nodes,components

def errors(objects,geometry,sources):
    geo={r['candidate_id']:r for r in geometry if r['source']=='CANDIDATE'}
    rows=[]
    for obj in objects:
        g=geo.get(obj['candidate_id'])
        flags=[]
        if obj['priority_reason']=='major_bbox_or_granularity_disagreement': flags.append('FROZEN_MAJOR_BBOX_OR_GRANULARITY_CONFLICT')
        if obj['human_decision']=='REJECT': flags.append('HUMAN_REJECTED_CANDIDATE_NO_CAUSE_CODE')
        if obj['human_decision']=='UNCERTAIN': flags.append('UNRESOLVED_VISUAL_CANDIDATE')
        if any(obj[t+'_confidence'] in {'LOW','AMBIGUOUS'} for t in ('ai1','ai2','a')): flags.append('LOW_SOURCE_CONFIDENCE_PROXY_NOT_OBSERVED_CONTRAST')
        if g:
            if g['center_distance_shape_norm']>.5: flags.append('LARGE_CENTER_CORRECTION')
            if g['width_ratio']>2 or g['height_ratio']>2: flags.append('EXCESS_CANDIDATE_EXTENT')
            if g['width_ratio']<.5 or g['height_ratio']<.5: flags.append('INSUFFICIENT_CANDIDATE_EXTENT')
            if g['axial_angle_error_deg'] is not None and g['axial_angle_error_deg']>30: flags.append('LARGE_ORIENTATION_CORRECTION')
            if g['geometry_type']=='ELLIPSE_ENVELOPE_PROXY': flags.append('RING_PATH_ENVELOPE_ONLY')
            if g['iou']<.25: flags.append('LOW_SHAPE_OVERLAP_AFTER_REVIEW')
        if obj['panel']=='f67v1' and obj['human_decision'] in CONFIRMED:
            # Ownership is not inferable from panel ID; a geometric boundary flag is only a warning.
            x2=float(obj['human_bbox'].split(',')[2])
            if x2>2500: flags.append('SCAN_RIGHT_BOUNDARY_OWNERSHIP_NOT_INFERRED')
        rows.append({'candidate_id':obj['candidate_id'],'panel':obj['panel'],'object_type':obj['object_type'],'scope':obj['scope'],
                     'human_decision':obj['human_decision'],'support_pattern':obj['support_pattern'],'flags':';'.join(flags),
                     'candidate_human_iou':g['iou'] if g else None,'cause_status':'GEOMETRIC_OR_FROZEN_FLAG_ONLY_NO_SEMANTIC_CAUSE'})
    return rows

def freeze_outputs(out):
    files=sorted(p for p in out.rglob('*') if p.is_file() and p.name!='SHA256SUMS' and '__pycache__' not in p.parts)
    (out/'SHA256SUMS').write_text(''.join(f'{digest(p)}  {p.relative_to(out)}\n' for p in files),encoding='utf-8')

def run(out,make_figures=True):
    out=safe_target(out)
    out.mkdir(parents=True,exist_ok=True)
    plan=PACKAGE/'THREE_WAY_COMPARISON_PLAN.md'
    check(plan.is_file(),'prospective analysis plan missing')
    plan_hash=digest(plan)
    lock=PACKAGE/'ANALYSIS_PLAN_LOCK.json'
    if lock.exists(): check(json.loads(lock.read_text())['plan_sha256']==plan_hash,'prospective plan changed')
    else: write_json(lock,{'plan_sha256':plan_hash,'seed':SEED,'bootstrap_replicates':BOOTSTRAPS,'phase':'BEFORE_OUTCOME_AGGREGATION'})
    print('Verifying frozen inputs...',flush=True)
    manifest,ledgers=verify_inventory()
    registered=PACKAGE/'INPUT_MANIFEST.tsv'
    if registered.exists():
        previous=read(registered)
        for row in previous: row['bytes']=int(row['bytes'])
        check(previous==manifest,'registered frozen snapshot changed')
    write(out/'INPUT_MANIFEST.tsv',manifest)
    print('Validating IDs, provenance and review partitions...',flush=True)
    inputs=load_inputs(manifest)
    sources,candidates,human,memberships,dimensions,mappings,audit,phases=inputs
    # Freeze the explicit data dictionary before any headline results.
    from reporting import dictionary, reports
    (out/'DATA_DICTIONARY.md').write_text(dictionary(),encoding='utf-8')
    write(out/'REVIEW_PHASE_INPUTS.tsv',phases)
    write(out/'PAIRWISE_RECONCILIATION_AUDIT.tsv',audit)
    objects,cohorts=build_objects(sources,candidates,human,memberships)
    write(out/'ANALYSIS_COHORTS.tsv',cohorts)
    write(out/'THREE_WAY_OBJECT_RESULTS.tsv',objects)
    print('Computing object, geometry, bootstrap and attachment results...',flush=True)
    geometry=build_geometry(objects,candidates,human,sources,dimensions)
    write(out/'THREE_WAY_GEOMETRY_RESULTS.tsv',geometry)
    summary=summarize_objects(objects)
    strata=stratify(objects)
    write(out/'STRATIFIED_OBJECT_METRICS.tsv',strata)
    gs=summarize_geometry(geometry)
    write(out/'GEOMETRY_SUMMARY_METRICS.tsv',gs)
    pairs,ps,confidence=pairwise_results(sources,mappings,objects,dimensions)
    write(out/'AI1_AI2_PAIRWISE_RESULTS.tsv',pairs)
    statistical=tests(objects,geometry)
    write(out/'THREE_WAY_STATISTICAL_TESTS.tsv',statistical)
    sens,geom_sens=sensitivity(objects,geometry)
    write(out/'SENSITIVITY_RESULTS.tsv',sens)
    write(out/'GEOMETRY_SENSITIVITY_RESULTS.tsv',geom_sens)
    graph,desc,edges,nodes,components=attachment(human,candidates)
    write(out/'ATTACHMENT_V2_DESCRIPTIVE_RESULTS.tsv',desc)
    write(out/'ATTACHMENT_V2_EDGE_RESULTS.tsv',edges)
    write(out/'ATTACHMENT_V2_NODE_DEGREES.tsv',nodes)
    write(out/'ATTACHMENT_V2_COMPONENTS.tsv',components)
    taxonomy=errors(objects,geometry,sources)
    write(out/'ERROR_FLAGS.tsv',taxonomy)
    output={'analysis_version':'1.0','analysis_plan_sha256':plan_hash,'seed':SEED,'bootstrap_replicates':BOOTSTRAPS,
            'protocols':{'human_STAR_LABEL':'1.2','legacy_spatial_relations':'1','caption_attachment':'2'},
            'primary_scope':'PRODUCTION_REVIEWED','object_metrics':summary,'pairwise_metrics':ps,
            'stratified_object_metrics':strata,
            'geometry_summaries':gs,'confidence_agreement_exploratory':confidence,'attachment_v2':graph,
            'statistical_test_counts':dict(Counter(r['family'] for r in statistical)),
            'input_validation':{'frozen_inputs_unchanged':True,'registered_files':len(manifest),'ledgers':ledgers,
                                'provenance_complete':True,'constrained_edges_not_joined':sum(r['reconciliation_status']=='CONSTRAINED_EDGE_NOT_JOINED' for r in audit)},
            'limitations':['candidate-derived selected human gold, not complete page census','few panel clusters',
                           'human confidence forced HIGH, not an uncertainty measurement','LABEL oriented dimension conventions ambiguous; derived geometry proxies',
                           'ring path thickness unavailable; filled envelope only','legacy relations incompatible with caption protocol 2'],
            'environment':{'python':platform.python_version(),'numpy':np.__version__}}
    write_json(out/'THREE_WAY_SUMMARY_METRICS.json',output)
    if make_figures:
        print('Rendering source-linked figures and deterministic overlays...',flush=True)
        from figures import render
        render(out,objects,geometry,strata,nodes,components,graph,candidates,human,sources)
    reports(out,output,statistical,sens,taxonomy)
    write_json(out/'INPUT_VALIDATION.json',output['input_validation'])
    print('Comparison outputs complete; run validation and freeze.',flush=True)
    return output

def main():
    parser=argparse.ArgumentParser()
    parser.add_argument('--output-dir',type=Path,default=PACKAGE)
    parser.add_argument('--no-figures',action='store_true')
    args=parser.parse_args()
    try: run(args.output_dir,not args.no_figures)
    except (ValueError,KeyError,FileNotFoundError) as exc:
        out=safe_target(args.output_dir)
        out.mkdir(parents=True,exist_ok=True)
        (out/'BLOCKER_REPORT.md').write_text('# Critical input/analysis blocker\n\n'+str(exc)+'\n\nNo frozen input was modified. No automatic correction applied.\n\n'
                 'THREE_WAY_COMPARISON_STATUS=BLOCKED\nFROZEN_INPUTS_UNCHANGED=YES\nPRIMARY_SCOPE=PRODUCTION_REVIEWED\n'
                 'AI1_AI2_HUMAN_COMPARISON_VALID=NO\nABSOLUTE_PAGE_LEVEL_RECALL_ESTIMATED=NO\nATTACHMENT_V2_ANALYSIS=BLOCKED\nRESULTS_REPRODUCIBLE=NO\n',encoding='utf-8')
        raise

if __name__=='__main__': main()
