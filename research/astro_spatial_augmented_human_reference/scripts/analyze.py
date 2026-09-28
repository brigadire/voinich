#!/usr/bin/env python3
"""Freeze augmented human objects, assisted recall, and 3G1 hyperedge analytics."""
from __future__ import annotations

import argparse
import csv
from collections import Counter,defaultdict
import hashlib
import json
import math
from pathlib import Path
import random
import sys

from PIL import Image,ImageDraw,ImageFont

sys.dont_write_bytecode=True
ROOT=Path(__file__).resolve().parents[3]
PKG=Path(__file__).resolve().parents[1]
HUMAN=ROOT/'research/astro_spatial_human_adjudication'
COMPARE=ROOT/'research/astro_spatial_three_way_comparison'
COMPLETE=ROOT/'research/astro_spatial_human_completeness'
FREEZE=COMPLETE/'job_18_object_freeze_c1_0'
GROUP_TASK=COMPLETE/'job_18_attachment_v3_full_pages_grouped'
GROUP_RESULT=COMPLETE/'job_21_group_review_validated'
PANELS=('f67r1','f67r2','f67v1','f68r1','f68r2','f68r3','f68v1','f68v2')
CALIBRATION={'f68r1','f68r3','f68v2'}
RELATION_SCOPE={'f68r1','f68r2','f68r3'}
REF_FIELDS=('reference_id','canonical_id','panel','object_type','geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation','layer')
GEOM_FIELDS=('geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')
VERSION='augmented-human-reference-1.0'

INPUTS=[
 ('AI1_STAR_SOURCE','research/astro_spatial_annotation_ai_b/AI_B_OBJECTS.tsv','AI1-frozen'),
 ('AI1_LABEL_SOURCE','research/astro_spatial_annotation_ai_b/AI_B_LABELS.tsv','AI1-frozen'),
 ('AI1_MANIFEST','research/astro_spatial_annotation_ai_b/AI_B_MANIFEST.json','AI1-frozen'),
 ('AI1_CHECKSUM_LEDGER','research/astro_spatial_annotation_ai_b/SHA256SUMS','AI1-frozen'),
 ('AI2_STAR_SOURCE','research/astro_spatial_annotation_ai_b2/AI2_OBJECTS.tsv','AI2-frozen'),
 ('AI2_LABEL_SOURCE','research/astro_spatial_annotation_ai_b2/AI2_LABELS.tsv','AI2-frozen'),
 ('AI2_MANIFEST','research/astro_spatial_annotation_ai_b2/AI2_MANIFEST.json','AI2-frozen'),
 ('AI2_CHECKSUM_LEDGER','research/astro_spatial_annotation_ai_b2/SHA256SUMS','AI2-frozen'),
 ('RECONCILED_AI_STAR_UNION','research/astro_spatial_human_adjudication/HUMAN_CANDIDATE_OBJECTS.tsv','human-1.2'),
 ('RECONCILED_AI_LABEL_UNION','research/astro_spatial_human_adjudication/HUMAN_CANDIDATE_LABELS.tsv','human-1.2'),
 ('HUMAN_STAR_ADJUDICATION','research/astro_spatial_human_adjudication/HUMAN_STAR_ADJUDICATION.tsv','human-1.2'),
 ('HUMAN_LABEL_ADJUDICATION','research/astro_spatial_human_adjudication/HUMAN_LABEL_ADJUDICATION.tsv','human-1.2'),
 ('ATTACHMENT_V2_FINAL','research/astro_spatial_human_adjudication/HUMAN_LABEL_OBJECT_RELATIONS.tsv','attachment-2'),
 ('HUMAN_MANIFEST','research/astro_spatial_human_adjudication/manifest.json','human-1.2/attachment-2'),
 ('HUMAN_CHECKSUM_LEDGER','research/astro_spatial_human_adjudication/SHA256SUMS','human-1.2/attachment-2'),
 ('THREE_WAY_OBJECT_RESULTS','research/astro_spatial_three_way_comparison/THREE_WAY_OBJECT_RESULTS.tsv','three-way-1.0'),
 ('THREE_WAY_INPUT_MANIFEST','research/astro_spatial_three_way_comparison/INPUT_MANIFEST.tsv','three-way-1.0'),
 ('THREE_WAY_SUMMARY','research/astro_spatial_three_way_comparison/THREE_WAY_SUMMARY_METRICS.json','three-way-1.0'),
 ('THREE_WAY_FROZEN_REPORT','research/astro_spatial_three_way_comparison/THREE_WAY_COMPARISON_REPORT.md','three-way-1.0'),
 ('THREE_WAY_CHECKSUM_LEDGER','research/astro_spatial_three_way_comparison/SHA256SUMS','three-way-1.0'),
 ('COMPLETENESS_BASELINE_REFERENCE','research/astro_spatial_human_completeness/REFERENCE_OBJECTS.tsv','C1.0'),
 ('COMPLETENESS_PRIOR_NONCONFIRMED','research/astro_spatial_human_completeness/PRIOR_NONCONFIRMED_PROVENANCE.tsv','C1.0'),
 ('COMPLETENESS_INPUT_MANIFEST','research/astro_spatial_human_completeness/INPUT_MANIFEST.tsv','C1.0'),
 ('COMPLETENESS_CHECKSUM_LEDGER','research/astro_spatial_human_completeness/SHA256SUMS','C1.0'),
 ('HUMAN_ADDITIONS','research/astro_spatial_human_completeness/job_18_object_freeze_c1_0/HUMAN_ADDITIONS_PROPOSED.tsv','C1.0'),
 ('HUMAN_ADDITION_DECISIONS','research/astro_spatial_human_completeness/job_18_object_freeze_c1_0/POST_REVIEW_OBJECT_DECISIONS.tsv','C1.0'),
 ('AUGMENTED_ENDPOINT_FREEZE','research/astro_spatial_human_completeness/job_18_object_freeze_c1_0/CONFIRMED_ENDPOINTS.tsv','C1.0'),
 ('PANEL_COMPLETION','research/astro_spatial_human_completeness/job_18_object_freeze_c1_0/PANEL_COMPLETION.tsv','C1.0'),
 ('OBJECT_FREEZE_MANIFEST','research/astro_spatial_human_completeness/job_18_object_freeze_c1_0/OBJECT_FREEZE_MANIFEST.json','C1.0'),
 ('OBJECT_FREEZE_CHECKSUM_LEDGER','research/astro_spatial_human_completeness/job_18_object_freeze_c1_0/SHA256SUMS','C1.0'),
 ('PRIOR_OVERLAP_QUEUE','research/astro_spatial_human_completeness/job_18_validated_staging/POST_REVIEW_RECONCILIATION_QUEUE.tsv','C1.0'),
 ('INITIAL_GROUPS','research/astro_spatial_human_completeness/job_18_attachment_v3_full_pages_grouped/INITIAL_GROUPS.tsv','3G1'),
 ('GROUP_ENDPOINT_SNAPSHOT','research/astro_spatial_human_completeness/job_18_attachment_v3_full_pages_grouped/ENDPOINT_SNAPSHOT.tsv','3G1'),
 ('GROUP_TASK_MANIFEST','research/astro_spatial_human_completeness/job_18_attachment_v3_full_pages_grouped/GROUP_TASK_MANIFEST.json','3G1'),
 ('GROUP_TASK_CHECKSUM_LEDGER','research/astro_spatial_human_completeness/job_18_attachment_v3_full_pages_grouped/SHA256SUMS','3G1'),
 ('REVIEWED_GROUPS','research/astro_spatial_human_completeness/job_21_group_review_validated/REVIEWED_RELATION_GROUPS.tsv','3G1'),
 ('GROUP_CHANGE_AUDIT','research/astro_spatial_human_completeness/job_21_group_review_validated/GROUP_CHANGE_AUDIT.tsv','3G1'),
 ('GROUP_REVIEW_PANEL_MARKERS','research/astro_spatial_human_completeness/job_21_group_review_validated/PANEL_REVIEW_MARKERS.tsv','3G1'),
 ('GROUP_REVIEW_MANIFEST','research/astro_spatial_human_completeness/job_21_group_review_validated/GROUP_REVIEW_RESULT_MANIFEST.json','3G1'),
 ('GROUP_REVIEW_VALIDATION','research/astro_spatial_human_completeness/job_21_group_review_validated/VALIDATION_REPORT.md','3G1'),
 ('GROUP_REVIEW_CHECKSUM_LEDGER','research/astro_spatial_human_completeness/job_21_group_review_validated/SHA256SUMS','3G1'),
]
INPUTS += [('CANONICAL_IMAGE_'+p.upper(),f'research/astro_spatial_human_adjudication/images/{p}.jpg','canonical-image') for p in PANELS]

def check(ok,msg):
    if not ok:raise ValueError(msg)
def digest(path):
    h=hashlib.sha256()
    with Path(path).open('rb') as f:
        for b in iter(lambda:f.read(1024*1024),b''):h.update(b)
    return h.hexdigest()
def read(path):
    with Path(path).open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(path,rows,fields=None):
    path=Path(path);path.parent.mkdir(parents=True,exist_ok=True)
    if fields is None:
        check(rows,'fields required for empty table');fields=list(rows[0])
    with path.open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fields,delimiter='\t',lineterminator='\n',extrasaction='raise');w.writeheader();w.writerows(rows)
def jswrite(path,obj):Path(path).write_text(json.dumps(obj,sort_keys=True,indent=2,allow_nan=False)+'\n',encoding='utf-8')
def unique(rows,key):
    d={r[key]:r for r in rows};check(len(d)==len(rows),f'duplicate {key}');return d
def count_rows(path):return len(read(path)) if Path(path).suffix=='.tsv' else ''
def pct(x):return 'NA' if x is None else f'{100*x:.1f}%'
def f6(x):return '' if x is None else f'{x:.6f}'

def verify_ledger(path):
    path=Path(path)
    for line in path.read_text().splitlines():
        if not line.strip():continue
        expected,rel=line.split('  ',1);target=(path.parent/rel).resolve()
        check(target.is_relative_to(path.parent.resolve()),f'unsafe checksum path {rel}')
        check(target.is_file() and digest(target)==expected,f'checksum mismatch {target}')

def inventory():
    rows=[]
    for role,rel,version in INPUTS:
        path=ROOT/rel;check(path.is_file(),f'missing input {rel}')
        rows.append({'logical_role':role,'path':rel,'protocol_version':version,'row_count':count_rows(path),
                     'file_size':path.stat().st_size,'sha256':digest(path),'frozen_status':'FROZEN'})
    check(len({r['path'] for r in rows})==len(rows),'duplicate manifest path')
    return rows

def verify_inputs(manifest=None):
    for ledger in (HUMAN/'SHA256SUMS',COMPARE/'SHA256SUMS',COMPLETE/'SHA256SUMS',FREEZE/'SHA256SUMS',GROUP_TASK/'SHA256SUMS',GROUP_RESULT/'SHA256SUMS'):
        verify_ledger(ledger)
    actual=inventory()
    if manifest is not None:
        registered=read(manifest)
        for r in actual:r['row_count']=str(r['row_count']);r['file_size']=str(r['file_size'])
        check(actual==registered,'input manifest differs from actual registered inputs')
    return actual

def wilson(k,n,z=1.959963984540054):
    if not n:return None,None
    p=k/n;d=1+z*z/n;c=(p+z*z/(2*n))/d;h=z*math.sqrt(p*(1-p)/n+z*z/(4*n*n))/d
    return max(0,c-h),min(1,c+h)
def quantile(values,q):
    values=sorted(values);x=(len(values)-1)*q;lo=math.floor(x);hi=math.ceil(x)
    return values[lo] if lo==hi else values[lo]*(hi-x)+values[hi]*(x-lo)
def cluster_ci(rows,support,key):
    panels=sorted({r['panel'] for r in rows})
    if len(panels)<2:return None,None,'NOT_ESTIMATED_SINGLE_PANEL'
    counts={p:(sum(r[support]=='YES' for r in rows if r['panel']==p),sum(r['panel']==p for r in rows)) for p in panels}
    seed=int(hashlib.sha256((key+'|'+support).encode()).hexdigest()[:16],16);rng=random.Random(seed);vals=[]
    for _ in range(20000):
        sample=[panels[rng.randrange(len(panels))] for _ in panels]
        k=sum(counts[p][0] for p in sample);n=sum(counts[p][1] for p in sample);vals.append(k/n)
    return quantile(vals,.025),quantile(vals,.975),'PANEL_CLUSTER_BOOTSTRAP_FEW_PANELS' if len(panels)<8 else 'PANEL_CLUSTER_BOOTSTRAP_8_PANELS'

def canonical_group(panel,members):
    payload='3G1|'+panel+'|'+'|'.join(sorted(members))
    return 'HGROUP_'+hashlib.sha256(payload.encode()).hexdigest()[:16].upper()

def initial_group_sets(rows):
    return {canonical_group(r['panel'],r['label_ids'].split(';')+r['star_ids'].split(';')):
            set(r['label_ids'].split(';')+r['star_ids'].split(';')) for r in rows}
def final_group_sets(rows):return {r['canonical_group_id']:set(r['label_ids'].split(';')+r['star_ids'].split(';')) for r in rows}

def svg_bars(path,title,categories,series,width=1050,height=520,ymax=None,percent=False):
    margin=(90,65,35,95);left,top,right,bottom=margin;pw=width-left-right;ph=height-top-bottom
    maximum=ymax or max(v for _,vals,_ in series for v in vals)*1.08 or 1
    colors=['#377eb8','#e6862b','#4daf4a','#984ea3']
    parts=[f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">',
           '<rect width="100%" height="100%" fill="white"/>',f'<text x="{width/2}" y="30" text-anchor="middle" font-family="sans-serif" font-size="20">{title}</text>']
    for i in range(6):
        val=maximum*i/5;y=top+ph*(1-i/5);parts.append(f'<line x1="{left}" y1="{y:.1f}" x2="{width-right}" y2="{y:.1f}" stroke="#ddd"/>')
        lab=f'{val*100:.0f}%' if percent else f'{val:.0f}';parts.append(f'<text x="{left-8}" y="{y+4:.1f}" text-anchor="end" font-family="sans-serif" font-size="12">{lab}</text>')
    groupw=pw/max(1,len(categories));barw=groupw/(len(series)+1)
    for si,(name,vals,labels) in enumerate(series):
        for ci,v in enumerate(vals):
            x=left+ci*groupw+(si+.5)*barw;y=top+ph*(1-v/maximum);h=top+ph-y
            parts.append(f'<rect x="{x:.1f}" y="{y:.1f}" width="{barw*.86:.1f}" height="{h:.1f}" fill="{colors[si]}"/>')
            parts.append(f'<text x="{x+barw*.43:.1f}" y="{max(top+11,y-5):.1f}" text-anchor="middle" font-family="sans-serif" font-size="10">{labels[ci]}</text>')
    for ci,c in enumerate(categories):
        x=left+(ci+.5)*groupw;parts.append(f'<text x="{x:.1f}" y="{top+ph+20}" text-anchor="middle" font-family="sans-serif" font-size="12">{c}</text>')
    for si,(name,_,_) in enumerate(series):
        x=left+si*180;parts.append(f'<rect x="{x}" y="{height-30}" width="14" height="14" fill="{colors[si]}"/><text x="{x+20}" y="{height-18}" font-family="sans-serif" font-size="12">{name}</text>')
    parts.append('</svg>');Path(path).write_text('\n'.join(parts)+'\n',encoding='utf-8')

def group_analytics(final_rows,initial_rows,objects):
    init=initial_group_sets(initial_rows);final=final_group_sets(final_rows);obj=unique(objects,'object_id')
    transitions=[]
    for iid,iset in init.items():
        overlaps=[(fid,fset,iset&fset) for fid,fset in final.items() if iset&fset]
        exact=[fid for fid,fset,_ in overlaps if fset==iset]
        if exact:change='UNCHANGED';reviewed=exact
        elif not overlaps:change='REMOVED_WITHOUT_REPLACEMENT';reviewed=[]
        elif len(overlaps)>1:change='SPLIT_OR_DISTRIBUTED_MEMBERSHIP';reviewed=[x[0] for x in overlaps]
        else:
            fid,fset,inter=overlaps[0];reviewed=[fid]
            if iset<fset:change='EXTENDED_WITH_HUMAN_ADDED_OBJECT' if any(obj[x]['origin']=='HUMAN_ADDED_COMPLETENESS' for x in fset-iset) else 'EXTENDED_WITH_EXISTING_REFERENCE_OBJECTS'
            elif fset<iset:change='MEMBERS_REMOVED'
            else:change='MODIFIED_MEMBERSHIP'
        transitions.append({'row_type':'INITIAL_GROUP','initial_group_id':iid,'reviewed_group_ids':';'.join(reviewed) or 'NONE','change_type':change,
                            'initial_members':';'.join(sorted(iset)),'reviewed_overlap_members':';'.join(sorted(set().union(*(x[2] for x in overlaps)))) if overlaps else '',
                            'initial_member_count':len(iset),'reviewed_group_count':len(reviewed)})
    overlapped_final={fid for iset in init.values() for fid,fset in final.items() if iset&fset}
    for fid in sorted(set(final)-overlapped_final):
        transitions.append({'row_type':'FINAL_ONLY_GROUP','initial_group_id':'NONE','reviewed_group_ids':fid,'change_type':'NEW_GROUP',
                            'initial_members':'','reviewed_overlap_members':'','initial_member_count':0,'reviewed_group_count':1})
    group_rows=[];member_rows=[]
    for row in final_rows:
        fid=row['canonical_group_id'];members=final[fid];overlap=[iid for iid,s in init.items() if s&members]
        exact=[iid for iid in overlap if init[iid]==members]
        if exact:category='UNCHANGED_FROM_INITIAL'
        elif not overlap:category='NEW_GROUP'
        elif len(overlap)>1:category='MERGED_INITIAL_GROUPS'
        else:
            old=init[overlap[0]]
            if old<members:category='EXTENDED_WITH_HUMAN_ADDED_OBJECT' if any(obj[x]['origin']=='HUMAN_ADDED_COMPLETENESS' for x in members-old) else 'EXTENDED_WITH_EXISTING_REFERENCE_OBJECTS'
            else:category='MODIFIED_MEMBERSHIP'
        labels=sorted(x for x in members if obj[x]['object_type']=='LABEL');stars=sorted(x for x in members if obj[x]['object_type']=='STAR')
        human=[x for x in members if obj[x]['origin']=='HUMAN_ADDED_COMPLETENESS']
        group_rows.append({'canonical_group_id':fid,'panel':row['panel'],'protocol_version':'3G1','reviewer_id':row['reviewer_id'],
            'human_confidence':row['human_confidence'],'member_count':len(members),'star_count':len(stars),'label_count':len(labels),
            'member_ids':';'.join(sorted(members)),'relation_semantics':'VISUAL_ASSOCIATION_HYPEREDGE','cartesian_edges_inferred':'NO',
            'change_category':category,'initial_group_ids':';'.join(overlap) or 'NONE','human_added_member_count':len(human),
            'contains_human_added':'YES' if human else 'NO'})
        for cid in sorted(members):member_rows.append({'canonical_group_id':fid,'panel':row['panel'],'object_id':cid,'object_type':obj[cid]['object_type'],
            'object_origin':obj[cid]['origin'],'human_added':'YES' if obj[cid]['origin']=='HUMAN_ADDED_COMPLETENESS' else 'NO','protocol_version':'3G1'})
    return group_rows,member_rows,transitions

def build(out=PKG):
    out=Path(out);out.mkdir(parents=True,exist_ok=True);(out/'scripts').mkdir(exist_ok=True);(out/'figures').mkdir(exist_ok=True)
    inputs=verify_inputs();write(out/'INPUT_MANIFEST.tsv',inputs,['logical_role','path','protocol_version','row_count','file_size','sha256','frozen_status'])
    freeze_rows=read(FREEZE/'CONFIRMED_ENDPOINTS.tsv');check(len(freeze_rows)==557,'augmented freeze must contain 557 objects')
    endpoints=unique(freeze_rows,'canonical_id');baseline=unique(read(COMPLETE/'REFERENCE_OBJECTS.tsv'),'canonical_id')
    additions=unique(read(FREEZE/'HUMAN_ADDITIONS_PROPOSED.tsv'),'canonical_id');decisions=unique(read(FREEZE/'POST_REVIEW_OBJECT_DECISIONS.tsv'),'canonical_new_id')
    check(len(baseline)==520 and len(additions)==len(decisions)==37,'baseline/addition counts changed')
    old_results=unique([r for r in read(COMPARE/'THREE_WAY_OBJECT_RESULTS.tsv') if r['human_decision'] in {'ACCEPT','MODIFY'}],'candidate_id')
    check(set(old_results)==set(baseline),'confirmed prior reference mismatch')
    check(set(endpoints)==set(baseline)|set(additions),'freeze membership mismatch')
    objects=[]
    adjud={r['candidate_id']:r for name in ('HUMAN_STAR_ADJUDICATION.tsv','HUMAN_LABEL_ADJUDICATION.tsv') for r in read(HUMAN/name)}
    for cid in sorted(endpoints):
        e=endpoints[cid];scope='CALIBRATION' if e['panel'] in CALIBRATION else 'PRODUCTION'
        if cid in baseline:
            r=old_results[cid];h=adjud[cid]
            check(all(e[k]==baseline[cid][k] for k in REF_FIELDS),'old endpoint changed')
            row={'object_id':cid,'panel':e['panel'],'object_type':e['object_type'],'scope':scope,
                 'relation_scope':'YES' if e['panel'] in RELATION_SCOPE else 'NO','origin':'AI_DERIVED_HUMAN_CONFIRMED',
                 **{k:e[k] for k in GEOM_FIELDS},'human_decision':r['human_decision'],'human_reviewer_id':h['reviewer_id'],
                 'human_confidence':h['human_confidence'],'human_protocol_version':r['human_protocol_version'],
                 'ai1_support':r['support_ai1'],'ai2_support':r['support_ai2'],'support_pattern':r['support_pattern'],
                 'source_annotation_ids':r['source_annotation_ids'],'provenance_reconciliation':'FROZEN_RECONCILED_AI_CANDIDATE',
                 'final_analytical_status':'AUGMENTED_CONFIRMED_REFERENCE'}
        else:
            p=additions[cid];d=decisions[cid]
            check(d['post_review_decision']=='CONFIRMED_NEW' and d['prior_outcomes_changed']=='NO','new object not reconciled')
            check(p['human_confidence']=='HIGH' and p['addition_status']=='PROPOSED_NEW','new object not final eligible')
            check(all(e[k]==p[k] for k in ('panel','object_type',*GEOM_FIELDS)),'new endpoint geometry changed')
            check(cid not in old_results,'human addition collides with AI candidate ID')
            row={'object_id':cid,'panel':e['panel'],'object_type':e['object_type'],'scope':scope,
                 'relation_scope':'YES' if e['panel'] in RELATION_SCOPE else 'NO','origin':'HUMAN_ADDED_COMPLETENESS',
                 **{k:e[k] for k in GEOM_FIELDS},'human_decision':'CONFIRMED_NEW','human_reviewer_id':d['reviewer_id'],
                 'human_confidence':p['human_confidence'],'human_protocol_version':'C1.0',
                 'ai1_support':'NO','ai2_support':'NO','support_pattern':'HUMAN_ONLY',
                 'source_annotation_ids':'NONE','provenance_reconciliation':'EXPLICIT_DISTINCT_NEW_OBJECT_NO_FROZEN_AI_MEMBERSHIP',
                 'final_analytical_status':'AUGMENTED_CONFIRMED_REFERENCE'}
        objects.append(row)
    unique(objects,'object_id');check(Counter(r['object_type'] for r in objects)=={'STAR':329,'LABEL':228},'augmented type totals mismatch')
    obj_fields=list(objects[0]);write(out/'AUGMENTED_HUMAN_OBJECTS.tsv',objects,obj_fields)

    final_groups=read(GROUP_RESULT/'REVIEWED_RELATION_GROUPS.tsv');initial_groups=read(GROUP_TASK/'INITIAL_GROUPS.tsv')
    check(len(final_groups)==64 and Counter(r['panel'] for r in final_groups)=={'f68r1':29,'f68r2':24,'f68r3':11},'final group counts mismatch')
    snapshot=read(GROUP_TASK/'ENDPOINT_SNAPSHOT.tsv')
    expected_snapshot=[{k:r[k] for k in REF_FIELDS} for r in freeze_rows if r['panel'] in RELATION_SCOPE]
    check(len(snapshot)==256 and snapshot==expected_snapshot,'256-object group endpoint snapshot changed')
    group_rows,member_rows,transitions=group_analytics(final_groups,initial_groups,objects)
    source_audit=read(GROUP_RESULT/'GROUP_CHANGE_AUDIT.tsv')
    expected_audit={r['initial_canonical_group_id']:r['review_outcome'] for r in source_audit}
    initial_ids=set(initial_group_sets(initial_groups));final_ids=set(final_group_sets(final_groups))
    check(len(source_audit)==65 and set(expected_audit)==initial_ids|final_ids,'source group audit ID coverage mismatch')
    for gid in initial_ids:
        check(expected_audit[gid]==('UNCHANGED_PRIOR_GROUP' if gid in final_ids else 'REMOVED_OR_REGROUPED_PRIOR_GROUP'),'source group audit initial outcome mismatch')
    for gid in final_ids-initial_ids:check(expected_audit[gid]=='NEW_OR_MODIFIED_GROUP','source group audit final outcome mismatch')
    write(out/'RELATION_GROUPS_3G1.tsv',group_rows);write(out/'RELATION_GROUP_MEMBERS_3G1.tsv',member_rows);write(out/'GROUP_TRANSITION_AUDIT.tsv',transitions)
    membership=defaultdict(list)
    for r in member_rows:membership[r['object_id']].append(r['canonical_group_id'])
    check(max(map(len,membership.values()))==1,'overlapping 3G1 memberships unexpectedly present')

    overlap_rows=read(COMPLETE/'job_18_validated_staging/POST_REVIEW_RECONCILIATION_QUEUE.tsv');overlap_by=defaultdict(list)
    for r in overlap_rows:overlap_by[r['canonical_new_id']].append(r)
    recon=[];added_audit=[]
    for cid,p in sorted(additions.items()):
        e=endpoints[cid];groups=membership.get(cid,[]);flags=overlap_by.get(cid,[]);d=decisions[cid]
        for flag in flags:
            recon.append({'human_added_object_id':cid,'panel':p['panel'],'object_type':p['object_type'],
                          'prior_candidate_id':flag['prior_canonical_id'],'prior_status_or_flag':flag['status'],
                          'envelope_iou':flag['envelope_iou'],'containment':flag['containment'],'center_distance_px':flag['center_distance_px'],
                          'reviewer_resolution':d['overlap_resolution'],'same_object_interpretation':'NO',
                          'ai_support_transferred':'NO','prior_outcome_changed':'NO'})
        added_audit.append({'object_id':cid,'panel':p['panel'],'object_type':p['object_type'],'geometry_type':p['geometry_type'],
            **{k:p[k] for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')},'completeness_reviewer':p['reviewer_id'],
            'human_confidence':p['human_confidence'],'relation_scope_eligibility':'YES' if p['panel'] in RELATION_SCOPE else 'NO',
            'group_id':groups[0] if groups else ('UNGROUPED' if p['panel'] in RELATION_SCOPE else 'NOT_APPLICABLE'),
            'group_status':'GROUPED' if groups else ('NO_VISUAL_GROUP_ASSIGNED' if p['panel'] in RELATION_SCOPE else 'OUTSIDE_RELATION_SCOPE'),
            'group_type':('TWO_MEMBER_GROUP' if groups and next(int(g['member_count']) for g in group_rows if g['canonical_group_id']==groups[0])==2 else
                          'MULTIMEMBER_GROUP' if groups else 'UNGROUPED' if p['panel'] in RELATION_SCOPE else 'OUTSIDE_RELATION_SCOPE'),
            'previous_candidate_overlap_status':';'.join(sorted({x['status'] for x in flags})) or 'NO_FLAGGED_OVERLAP',
            'ai1_provenance':'NO_FROZEN_MEMBERSHIP','ai2_provenance':'NO_FROZEN_MEMBERSHIP',
            'final_analytical_status':'CONFIRMED_HUMAN_ONLY_DISTINCT_FROM_PRIOR_CANDIDATES'})
    write(out/'HUMAN_ADDED_OBJECT_AUDIT.tsv',added_audit);write(out/'PRIOR_CANDIDATE_RECONCILIATION.tsv',recon,
        ['human_added_object_id','panel','object_type','prior_candidate_id','prior_status_or_flag','envelope_iou','containment','center_distance_px','reviewer_resolution','same_object_interpretation','ai_support_transferred','prior_outcome_changed'])
    check(len(added_audit)==37 and sum(r['relation_scope_eligibility']=='YES' for r in added_audit)==17,'human addition scope count mismatch')
    check(sum(r['group_status']=='GROUPED' for r in added_audit)==16,'grouped human addition count mismatch')
    ungroup=[r for r in added_audit if r['group_status']=='NO_VISUAL_GROUP_ASSIGNED'];check([r['object_id'] for r in ungroup]==['HNEW_STAR_f68r2_DC27A556209F27B3'],'ungrouped object mismatch')
    addition_summary=[]
    def add_summary(dimension,rows,key):
        for value,count in sorted(Counter(r[key] for r in rows).items()):
            addition_summary.append({'dimension':dimension,'value':value,'count':count,'denominator':len(rows),'fraction':f6(count/len(rows))})
    add_summary('OBJECT_TYPE',added_audit,'object_type');add_summary('PANEL',added_audit,'panel')
    add_summary('RELATION_SCOPE',added_audit,'relation_scope_eligibility');add_summary('GROUP_STATUS',added_audit,'group_status')
    add_summary('GROUP_TYPE',added_audit,'group_type')
    add_summary('PREVIOUS_CANDIDATE_OVERLAP_STATUS',added_audit,'previous_candidate_overlap_status')
    for status in ('MATCHES_PRIOR_REJECT','MATCHES_PRIOR_UNCERTAIN','MATCHES_PRIOR_NOT_REVIEWED','POSSIBLE_DUPLICATE_CONFIRMED'):
        count=sum(any(flag['status']==status for flag in overlap_by.get(r['object_id'],[])) for r in added_audit)
        addition_summary.append({'dimension':'PRIOR_MATCH_FLAG','value':status,'count':count,'denominator':len(added_audit),'fraction':f6(count/len(added_audit))})
    relation_added=[r for r in added_audit if r['relation_scope_eligibility']=='YES']
    add_summary('RELATION_SCOPE_GROUP_STATUS',relation_added,'group_status')
    write(out/'HUMAN_ADDED_OBJECT_SUMMARY.tsv',addition_summary)

    by_object=[{'object_id':r['object_id'],'panel':r['panel'],'object_type':r['object_type'],'scope':r['scope'],'relation_scope':r['relation_scope'],
                'origin':r['origin'],'ai1_support':r['ai1_support'],'ai2_support':r['ai2_support'],'ai_union_support':'YES' if 'YES' in (r['ai1_support'],r['ai2_support']) else 'NO',
                'support_pattern':r['support_pattern'],'assisted_reference_denominator_member':'YES'} for r in objects]
    write(out/'AI_ASSISTED_RECALL_BY_OBJECT.tsv',by_object)
    scope_defs={'ALL_PANELS':lambda r:True,'CALIBRATION':lambda r:r['scope']=='CALIBRATION','PRODUCTION':lambda r:r['scope']=='PRODUCTION',
                'RELATION_SCOPE':lambda r:r['relation_scope']=='YES'}
    summary=[]
    for typ in ('STAR','LABEL'):
      for scope,fn in scope_defs.items():
        cohort=[r for r in by_object if r['object_type']==typ and fn(r)]
        for source,field in [('AI1','ai1_support'),('AI2','ai2_support'),('AI_UNION','ai_union_support')]:
            k=sum(r[field]=='YES' for r in cohort);n=len(cohort);lo,hi=wilson(k,n);blo,bhi,status=cluster_ci(cohort,field,typ+'|'+scope)
            summary.append({'object_type':typ,'scope':scope,'panel':'ALL','source':source,'numerator':k,'denominator':n,
                            'assisted_recall':f6(k/n),'wilson_95_low':f6(lo),'wilson_95_high':f6(hi),
                            'panel_cluster_bootstrap_95_low':f6(blo),'panel_cluster_bootstrap_95_high':f6(bhi),
                            'cluster_count':len({r['panel'] for r in cohort}),'uncertainty_status':status})
      for panel in PANELS:
        cohort=[r for r in by_object if r['object_type']==typ and r['panel']==panel]
        for source,field in [('AI1','ai1_support'),('AI2','ai2_support'),('AI_UNION','ai_union_support')]:
            k=sum(r[field]=='YES' for r in cohort);n=len(cohort);lo,hi=wilson(k,n)
            summary.append({'object_type':typ,'scope':'PANEL','panel':panel,'source':source,'numerator':k,'denominator':n,
                            'assisted_recall':f6(k/n),'wilson_95_low':f6(lo),'wilson_95_high':f6(hi),
                            'panel_cluster_bootstrap_95_low':'','panel_cluster_bootstrap_95_high':'','cluster_count':1,'uncertainty_status':'NOT_ESTIMATED_SINGLE_PANEL'})
    write(out/'AI_ASSISTED_RECALL_SUMMARY.tsv',summary)
    all_main=[r for r in summary if r['scope']=='ALL_PANELS']
    support_counts={typ:dict(Counter(r['support_pattern'] for r in objects if r['object_type']==typ)) for typ in ('STAR','LABEL')}
    old_den={'STAR':321,'LABEL':199};main={r['object_type']+'_'+r['source']:r for r in all_main}
    support_fractions={typ:{key:{'numerator':value,'denominator':sum(support_counts[typ].values()),'fraction':round(value/sum(support_counts[typ].values()),6)}
                            for key,value in support_counts[typ].items()} for typ in ('STAR','LABEL')}
    metrics={'version':VERSION,'terminology':'assisted human-reference recall','design_limitation':'reviewer saw existing reference overlay and searched for omissions',
      'object_totals':{'all':557,'STAR':329,'LABEL':228,'ai_derived_confirmed':520,'human_added_confirmed':37,'human_added_relation_scope':17},
      'support_counts':support_counts,'support_partition_exact':support_fractions,
      'marginal_contribution':{typ:{'AI1_beyond_AI2':support_fractions[typ]['AI1_ONLY'],'AI2_beyond_AI1':support_fractions[typ]['AI2_ONLY']} for typ in ('STAR','LABEL')},
      'common_misses':{typ:support_fractions[typ]['HUMAN_ONLY'] for typ in ('STAR','LABEL')},'metrics':summary,
      'pre_completeness_denominators':old_den,
      'recall_change_percentage_points':{typ:{s:round(100*(float(main[typ+'_'+s]['assisted_recall'])-((int(main[typ+'_'+s]['numerator']))/old_den[typ])),6) for s in ('AI1','AI2','AI_UNION')} for typ in ('STAR','LABEL')},
      'bootstrap':{'method':'panel-cluster nonparametric bootstrap','replicates':20000,'seed':'SHA256(metric key)','warning':'few panels; descriptive uncertainty, not manuscript-wide inference'}}
    jswrite(out/'AI_ASSISTED_RECALL_METRICS.json',metrics)

    # Group structure metrics and empirical partition properties.
    structure=[]
    objects_by_id=unique(objects,'object_id')
    for panel in ('ALL',*sorted(RELATION_SCOPE)):
        gs=group_rows if panel=='ALL' else [r for r in group_rows if r['panel']==panel]
        members={m['object_id'] for m in member_rows if panel=='ALL' or m['panel']==panel}
        eligible=[r for r in objects if r['panel'] in RELATION_SCOPE and (panel=='ALL' or r['panel']==panel)]
        eligible_human=[r for r in eligible if r['origin']=='HUMAN_ADDED_COMPLETENESS']
        grouped_human=sum(objects_by_id[cid]['origin']=='HUMAN_ADDED_COMPLETENESS' for cid in members)
        counts={'GROUPS':len(gs),'SIZE_2_GROUPS':sum(int(r['member_count'])==2 for r in gs),'MULTIMEMBER_GROUPS':sum(int(r['member_count'])>2 for r in gs),
                'GROUPED_OBJECTS':len(members),'UNGROUPED_OBJECTS':len(eligible)-len(members),'GROUPED_STAR':sum(objects_by_id[cid]['object_type']=='STAR' for cid in members),
                'GROUPED_LABEL':sum(objects_by_id[cid]['object_type']=='LABEL' for cid in members),
                'GROUPS_WITH_HUMAN_ADDED':sum(r['contains_human_added']=='YES' for r in gs),'HUMAN_ADDED_ELIGIBLE':len(eligible_human),'HUMAN_ADDED_GROUPED':grouped_human,
                'HUMAN_ADDED_GROUPED_FRACTION':f6(grouped_human/len(eligible_human)) if eligible_human else '',
                'NEW_GROUPS':sum(r['change_category']=='NEW_GROUP' for r in gs),'EXTENDED_GROUPS':sum(r['change_category'].startswith('EXTENDED_') for r in gs),
                'REGROUPED_OR_MEMBERSHIP_MODIFIED_GROUPS':sum(r['change_category'] not in {'UNCHANGED_FROM_INITIAL','NEW_GROUP'} for r in gs)}
        for metric,value in counts.items():structure.append({'panel':panel,'metric':metric,'value':value,'protocol_version':'3G1','note':'fraction of eligible human additions' if metric.endswith('_FRACTION') else 'hyperedge/group count; no pair expansion'})
    init_members=set().union(*initial_group_sets(initial_groups).values());final_members=set(membership)
    init_sets=initial_group_sets(initial_groups);final_sets=final_group_sets(final_groups)
    exact_initial_members=set().union(*(iset for iset in init_sets.values() if iset in final_sets.values()))
    turnover={'OBJECTS_FIRST_GROUPED':len(final_members-init_members),'OBJECTS_LOST_GROUP':len(init_members-final_members),
              'OBJECTS_RETAINED_IN_SAME_EXACT_GROUP':len(exact_initial_members),
              'OBJECTS_IN_MEMBERSHIP_MODIFIED_INITIAL_GROUP':len(init_members-exact_initial_members),
              'OBJECTS_CHANGED_GROUP_CANONICAL_ID':len(init_members-exact_initial_members),
              'OVERLAPPING_GROUP_MEMBERSHIPS':sum(len(v)>1 for v in membership.values())}
    for metric,value in turnover.items():structure.append({'panel':'ALL','metric':metric,'value':value,'protocol_version':'3G1','note':'membership-set comparison'})
    write(out/'GROUP_STRUCTURE_METRICS.tsv',structure)

    # Attachment-v2 pair audit without converting 3G1 groups into edges.
    group_of={m['object_id']:m['canonical_group_id'] for m in member_rows};group_size={r['canonical_group_id']:int(r['member_count']) for r in group_rows}
    v2=read(HUMAN/'HUMAN_LABEL_OBJECT_RELATIONS.tsv');cross=[];v2keys={}
    for r in v2:
        lid,sid=r['label_candidate_id'],r['object_candidate_id'];panel=endpoints[lid]['panel'];v2keys[(lid,sid)]=r
        if panel not in RELATION_SCOPE:status='ENDPOINTS_OUTSIDE_3G1_SCOPE';claim='NOT_COMPARABLE_OUTSIDE_SCOPE'
        elif r['relation_type']=='UNCERTAIN':status='ATTACHMENT_V2_UNCERTAIN';claim='NO_3G1_PAIRWISE_INFERENCE'
        elif r['relation_type']=='UNASSIGNED':status='ATTACHMENT_V2_UNASSIGNED';claim='NO_3G1_NEGATIVE_INFERENCE'
        elif group_of.get(lid) and group_of.get(lid)==group_of.get(sid):
            status='V2_POSITIVE_CO_MEMBERS_TWO_MEMBER_3G1_GROUP' if group_size[group_of[lid]]==2 else 'V2_POSITIVE_CO_MEMBERS_MULTIMEMBER_3G1_GROUP'
            claim='VISUAL_LABEL_OF_COMPATIBLE_UNITS' if group_size[group_of[lid]]==2 else 'ENDPOINTS_CO_MEMBER_SAME_3G1_GROUP'
        else:status='V2_POSITIVE_ENDPOINTS_NOT_CO_MEMBERS_IN_3G1';claim='REPRESENTATIONS_DIFFER'
        cross.append({'row_type':'ATTACHMENT_V2_PAIR','attachment_v2_relation_id':r['relation_candidate_id'],'panel':panel,'label_id':lid,'star_id':sid,
                      'attachment_v2_decision':r['relation_type'],'group_3g1_id_label':group_of.get(lid,'UNGROUPED'),'group_3g1_id_star':group_of.get(sid,'UNGROUPED'),
                      'crosswalk_status':status,'allowed_analytical_claim':claim,'pairwise_metric_eligible':'NO'})
    for g in group_rows:
        labels=[x for x in g['member_ids'].split(';') if endpoints[x]['object_type']=='LABEL'];stars=[x for x in g['member_ids'].split(';') if endpoints[x]['object_type']=='STAR']
        if len(labels)==len(stars)==1 and (labels[0],stars[0]) not in v2keys:
            cross.append({'row_type':'3G1_TWO_MEMBER_GROUP_NO_V2_PAIR','attachment_v2_relation_id':'NONE','panel':g['panel'],'label_id':labels[0],'star_id':stars[0],
                          'attachment_v2_decision':'PAIR_ABSENT_FROM_V2_QUEUE','group_3g1_id_label':g['canonical_group_id'],'group_3g1_id_star':g['canonical_group_id'],
                          'crosswalk_status':'PAIR_ABSENT_FROM_ATTACHMENT_V2_QUEUE','allowed_analytical_claim':'NEW_3G1_VISUAL_GROUP_ONLY','pairwise_metric_eligible':'NO'})
        if int(g['member_count'])>2:
            cross.append({'row_type':'3G1_MULTIMEMBER_GROUP_SUMMARY','attachment_v2_relation_id':'NONE','panel':g['panel'],'label_id':';'.join(labels),'star_id':';'.join(stars),
                          'attachment_v2_decision':'NOT_A_PAIR','group_3g1_id_label':g['canonical_group_id'],'group_3g1_id_star':g['canonical_group_id'],
                          'crosswalk_status':'MULTIMEMBER_HYPEREDGE_NOT_EXPANDED','allowed_analytical_claim':'ENDPOINTS_CO_MEMBER_SAME_3G1_GROUP','pairwise_metric_eligible':'NO'})
    write(out/'ATTACHMENT_V2_3G1_CROSSWALK.tsv',cross)

    # Machine-readable figure sources and deterministic SVGs.
    support_source=[]
    for typ in ('STAR','LABEL'):
        for category in ('BOTH_AI','AI1_ONLY','AI2_ONLY','HUMAN_ONLY'):
            count=support_counts[typ].get(category,0);denominator=sum(support_counts[typ].values())
            support_source.append({'object_type':typ,'support_category':category,'count':count,'denominator':denominator,'fraction':f6(count/denominator)})
    write(out/'figures/AI_SUPPORT_DIAGRAM_SOURCE.tsv',support_source)
    svg_bars(out/'figures/AI_SUPPORT_DIAGRAM.svg','Augmented reference support',list(('STAR','LABEL')),
             [(cat,[support_counts[t].get(cat,0) for t in ('STAR','LABEL')],[str(support_counts[t].get(cat,0)) for t in ('STAR','LABEL')]) for cat in ('BOTH_AI','AI1_ONLY','AI2_ONLY','HUMAN_ONLY')],ymax=350)
    recall_source=[r for r in summary if r['scope']=='ALL_PANELS'];write(out/'figures/ASSISTED_RECALL_SOURCE.tsv',recall_source)
    svg_bars(out/'figures/ASSISTED_RECALL.svg','Assisted human-reference recall',['STAR','LABEL'],
             [(src,[float(main[t+'_'+src]['assisted_recall']) for t in ('STAR','LABEL')],[str(main[t+'_'+src]['numerator'])+'/'+str(main[t+'_'+src]['denominator']) for t in ('STAR','LABEL')]) for src in ('AI1','AI2','AI_UNION')],ymax=1,percent=True)
    panel_source=[r for r in summary if r['scope']=='PANEL'];write(out/'figures/ASSISTED_RECALL_BY_PANEL_SOURCE.tsv',panel_source)
    cats=[p+' '+t[0] for p in PANELS for t in ('STAR','LABEL')]
    pidx={(r['panel'],r['object_type'],r['source']):r for r in panel_source}
    svg_bars(out/'figures/ASSISTED_RECALL_BY_PANEL.svg','Assisted recall by panel/type',cats,
             [(src,[float(pidx[p,t,src]['assisted_recall']) for p in PANELS for t in ('STAR','LABEL')],[str(pidx[p,t,src]['numerator'])+'/'+str(pidx[p,t,src]['denominator']) for p in PANELS for t in ('STAR','LABEL')]) for src in ('AI1','AI2','AI_UNION')],width=1900,height=600,ymax=1,percent=True)
    transition_counts=Counter(r['change_category'] for r in group_rows);trans_source=[{'change_category':k,'final_group_count':v,'denominator':64} for k,v in sorted(transition_counts.items())];write(out/'figures/GROUP_TRANSITION_SOURCE.tsv',trans_source)
    svg_bars(out/'figures/GROUP_TRANSITION.svg','Initial to reviewed 3G1 groups',list(transition_counts),[('Final groups',list(transition_counts.values()),[str(x) for x in transition_counts.values()])],width=1150)
    size_counts=Counter(int(r['member_count']) for r in group_rows);size_source=[{'group_size':k,'group_count':v,'denominator':64} for k,v in sorted(size_counts.items())];write(out/'figures/GROUP_SIZE_DISTRIBUTION_SOURCE.tsv',size_source)
    svg_bars(out/'figures/GROUP_SIZE_DISTRIBUTION.svg','3G1 group size distribution',[str(k) for k in sorted(size_counts)],[('Groups',[size_counts[k] for k in sorted(size_counts)],[str(size_counts[k]) for k in sorted(size_counts)])],ymax=70)

    overlay=[]
    objidx=unique(objects,'object_id')
    changed=[r for r in group_rows if r['change_category']!='UNCHANGED_FROM_INITIAL']
    for panel in sorted(RELATION_SCOPE):
        image=Image.open(HUMAN/'images'/f'{panel}.jpg').convert('RGB');scale=min(1,1400/image.width,1400/image.height);canvas=image.resize((round(image.width*scale),round(image.height*scale)))
        draw=ImageDraw.Draw(canvas)
        for number,g in enumerate([x for x in changed if x['panel']==panel],1):
            members=g['member_ids'].split(';');boxes=[[float(objidx[c][k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')] for c in members]
            boundary=(min(b[0] for b in boxes),min(b[1] for b in boxes),max(b[2] for b in boxes),max(b[3] for b in boxes));color='#d627a4'
            draw.rounded_rectangle(tuple(v*scale for v in boundary),radius=12,width=5,outline=color)
            for cid,b in zip(members,boxes):draw.rectangle(tuple(v*scale for v in b),width=3,outline='#1b9e77' if objidx[cid]['object_type']=='STAR' else '#377eb8')
            draw.text((boundary[0]*scale,(boundary[1]*scale)-15),f'G{number:02d}',fill=color)
            overlay.append({'panel':panel,'neutral_group_label':f'G{number:02d}','canonical_group_id':g['canonical_group_id'],'change_category':g['change_category'],
                            'member_ids':g['member_ids'],'boundary_x1':f6(boundary[0]),'boundary_y1':f6(boundary[1]),'boundary_x2':f6(boundary[2]),'boundary_y2':f6(boundary[3])})
        canvas.save(out/'figures'/f'CHANGED_GROUPS_{panel}.png')
    write(out/'figures/GROUP_OVERLAY_SOURCE.tsv',overlay)
    multi=next(r for r in group_rows if int(r['member_count'])==8);members=multi['member_ids'].split(';');boxes=[[float(objidx[c][k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')] for c in members]
    x1=max(0,min(b[0] for b in boxes)-140);y1=max(0,min(b[1] for b in boxes)-140);x2=max(b[2] for b in boxes)+140;y2=max(b[3] for b in boxes)+140
    base=Image.open(HUMAN/'images/f68r3.jpg').convert('RGB').crop((int(x1),int(y1),int(x2),int(y2)));scale=min(1,1400/base.width,1000/base.height);canvas=base.resize((round(base.width*scale),round(base.height*scale)));draw=ImageDraw.Draw(canvas)
    bound=(min(b[0] for b in boxes)-x1,min(b[1] for b in boxes)-y1,max(b[2] for b in boxes)-x1,max(b[3] for b in boxes)-y1)
    draw.rounded_rectangle(tuple(v*scale for v in bound),radius=18,width=7,outline='#d627a4')
    for cid,b in zip(members,boxes):draw.rectangle(tuple((b[i]-(x1 if i%2==0 else y1))*scale for i in range(4)),width=4,outline='#1b9e77' if objidx[cid]['object_type']=='STAR' else '#377eb8')
    draw.text((12,12),'ONE 3G1 HYPEREDGE: 1 LABEL + 7 STAR; NO PAIR ARROWS',fill='black',stroke_width=3,stroke_fill='white');canvas.save(out/'figures/F68R3_EIGHT_MEMBER_HYPEREDGE.png')
    write(out/'figures/F68R3_EIGHT_MEMBER_HYPEREDGE_SOURCE.tsv',[{'canonical_group_id':multi['canonical_group_id'],'panel':'f68r3','member_count':8,'label_count':1,'star_count':7,'member_ids':multi['member_ids'],'representation':'ONE_GROUP_BOUNDARY_NO_PAIR_ARROWS'}])

    reports(out,objects,added_audit,summary,group_rows,structure,cross,transition_counts,support_counts,metrics)
    manifest={'version':VERSION,'status':'COMPLETE','augmented_objects':557,'human_added':37,'relation_scope_objects':256,'relation_groups':64,
              'input_manifest_sha256':digest(out/'INPUT_MANIFEST.tsv'),'outputs_generated':True,'previous_reports_overwritten':False}
    jswrite(out/'AUGMENTED_REFERENCE_MANIFEST.json',manifest)
    print('AUGMENTED_REFERENCE=COMPLETE; OBJECTS=557; HUMAN_ADDED=37; GROUPS=64')

def reports(out,objects,added,summary,groups,structure,cross,transition_counts,support_counts,metrics):
    main={(r['object_type'],r['source']):r for r in summary if r['scope']=='ALL_PANELS'}
    def metric(t,s):
        r=main[t,s];return f"{pct(float(r['assisted_recall']))} ({r['numerator']}/{r['denominator']}; Wilson 95% {pct(float(r['wilson_95_low']))}–{pct(float(r['wilson_95_high']))})"
    plan='''# Analysis plan\n\nFreeze the separately validated C1.0 object endpoint snapshot, join prior objects to the frozen constrained AI provenance, classify all confirmed completeness additions only after explicit distinct-object reconciliation, and calculate assisted human-reference recall. Analyze 3G1 as membership sets/hyperedges and crosswalk attachment-v2 without pair expansion. Primary scopes are all eight panels, calibration, production, and the three relation pages. Uncertainty uses Wilson intervals and deterministic panel-cluster bootstrap. No transcription, semantics, astronomical interpretation, new matching, or old-file mutation.\n'''
    (out/'ANALYSIS_PLAN.md').write_text(plan,encoding='utf-8')
    dictionary='''# Data dictionary\n\n`AUGMENTED_HUMAN_OBJECTS.tsv`: one confirmed object; origin distinguishes AI-derived human-confirmed from human completeness additions. `ai*_support` is frozen candidate provenance, not geometric rematching. `HUMAN_ADDED_OBJECT_AUDIT.tsv`: all 37 additions; relation eligibility identifies the 17-object subset and `group_type` distinguishes two-member, multi-member, ungrouped and out-of-scope records. `HUMAN_ADDED_OBJECT_SUMMARY.tsv` gives exact counts, denominators and fractions. `PRIOR_CANDIDATE_RECONCILIATION.tsv`: conservative envelope flags resolved as distinct objects; no AI support transfers. `AI_ASSISTED_RECALL_*`: denominator is confirmed augmented reference and term is assisted human-reference recall. Wilson is object-level binomial descriptive uncertainty; cluster bootstrap resamples panels and is weak with few panels.\n\n`RELATION_GROUPS_3G1.tsv`: one group/hyperedge. `member_ids` is authoritative; `star_count`/`label_count` are membership counts. `RELATION_GROUP_MEMBERS_3G1.tsv`: normalized membership. Groups empirically form a partition of grouped objects (no overlapping membership), not an assumed protocol constraint. Change categories are membership-defined: `UNCHANGED_FROM_INITIAL` means exact equality with one initial member set; `NEW_GROUP` means no member overlaps any initial group; `EXTENDED_WITH_EXISTING_REFERENCE_OBJECTS` means one initial set is a strict subset and every added member is a prior confirmed reference object. The latter is the independently resolved form of the source audit's removed/regrouped initial group plus new/modified final group. `GROUP_TRANSITION_AUDIT.tsv` compares exact membership sets and distinguishes initial rows from final-only groups.\n\n`ATTACHMENT_V2_3G1_CROSSWALK.tsv` is an audit only: v2 rows are pair decisions, 3G1 rows are groups. `ENDPOINTS_CO_MEMBER_SAME_3G1_GROUP` never means seven inferred pair decisions. `NO_VISUAL_GROUP_ASSIGNED` is neither a negative pair nor an error. `pairwise_metric_eligible=NO` prevents incompatible-unit precision/recall.\n'''
    (out/'DATA_DICTIONARY.md').write_text(dictionary,encoding='utf-8')
    old={'STAR':321,'LABEL':199}
    addendum=f'''# Three-way comparison addendum: augmented assisted reference\n\n+This addendum does not overwrite the frozen three-way report. The completeness reviewer saw the existing overlay and searched only for omissions; results are assisted human-reference recall, not absolute or independent blind-gold recall.\n\n+| Type | AI1 | AI2 | AI union | Human-only |\n+|---|---:|---:|---:|---:|\n+| STAR | {metric('STAR','AI1')} | {metric('STAR','AI2')} | {metric('STAR','AI_UNION')} | 8/329 |\n+| LABEL | {metric('LABEL','AI1')} | {metric('LABEL','AI2')} | {metric('LABEL','AI_UNION')} | 29/228 |\n+\n+Before completeness, denominators were 321 STAR and 199 LABEL and the AI union covered them by construction. Adding 8 STAR and 29 LABEL common omissions lowers union assisted recall to 321/329 and 199/228. AI2 remains substantially above AI1; AI1 adds only 6 STAR and 3 LABEL beyond AI2, while AI2 adds 154 STAR and 111 LABEL beyond AI1. The human-only fraction is larger for LABEL (29/228) than STAR (8/329). Omissions occur on f67r2, f67v1, f68r2, f68r3 and f68v2.\n\n+Sixteen of 17 human additions on relation pages belong to a reviewed group. This does not mean all 37 additions are relation-eligible. Group review yields 64 hyperedges, including one 1-LABEL/7-STAR group, and changes the representation from pair candidates to full-page association groups. These conclusions remain limited to assisted review on eight object pages and three relation pages.\n'''
    (out/'THREE_WAY_COMPARISON_ADDENDUM.md').write_text(addendum.replace('\n+','\n'),encoding='utf-8')
    report=f'''# Augmented human reference report\n\n+## Result\n+\n+The immutable augmented reference contains 557 confirmed objects: 329 STAR and 228 LABEL. It combines 520 prior ACCEPT/MODIFY final geometries with 37 explicitly confirmed human completeness additions (8 STAR, 29 LABEL). Prior REJECT/UNCERTAIN/NOT_REVIEWED and no unresolved additions enter the denominator. All 37 additions were reconciled as distinct real objects; 28 proximity/overlap records affect 23 objects but transfer no AI provenance and change no old outcome.\n\n+## Assisted human-reference recall\n+\n+| Type | AI1 | AI2 | AI union | Both / AI1-only / AI2-only / human-only |\n+|---|---:|---:|---:|---:|\n+| STAR | {metric('STAR','AI1')} | {metric('STAR','AI2')} | {metric('STAR','AI_UNION')} | 161 / 6 / 154 / 8 |\n+| LABEL | {metric('LABEL','AI1')} | {metric('LABEL','AI2')} | {metric('LABEL','AI_UNION')} | 85 / 3 / 111 / 29 |\n+\n+AI2's observed advantage persists after augmentation. Relative to the pre-completeness confirmed denominator, AI1 changes from 167/321 to 167/329 for STAR and 88/199 to 88/228 for LABEL; AI2 changes from 315/321 to 315/329 and 196/199 to 196/228. Union common misses are 8 STAR and 29 LABEL. Calibration, production, relation-scope, panel rows, Wilson intervals and panel bootstrap intervals are in AI_ASSISTED_RECALL_SUMMARY.tsv. Few-panel bootstrap intervals are descriptive and cannot support manuscript-wide generalization.\n\n+## Human additions\n+\n+All-page human additions number 37, not 17. Seventeen are on f68r1/f68r2/f68r3 and 20 are outside relation scope. Sixteen relation-scope additions are grouped. HNEW_STAR_f68r2_DC27A556209F27B3 remains an existing, fully reviewed object with `NO_VISUAL_GROUP_ASSIGNED`; no negative LABEL pairs are generated.\n\n+## Relation groups 3G1\n+\n+There are 64 groups: 29 f68r1, 24 f68r2, 11 f68r3. Sixty-three have two members; one f68r3 hyperedge has one LABEL and seven STAR and remains one entity. Membership is empirically non-overlapping: 134 grouped and 122 ungrouped objects within the 256-object relation scope. Exact membership comparison finds 45 unchanged groups, one initial group extended from four to eight members with existing reference STAR, and 18 wholly new groups. Thus `45 unchanged + 19 new/modified = 64`; the 46th initial group is the extended group and is not double counted. Forty objects first receive a group, none lose grouped membership, and four retained members participate in the extended membership.\n\n+## Attachment-v2 crosswalk\n+\n+Attachment-v2 remains frozen pair-level review; 3G1 is full-page group review. The crosswalk records co-membership and representation differences only. Multi-member co-membership is `ENDPOINTS_CO_MEMBER_SAME_3G1_GROUP`, not seven pair claims. No pairwise precision/recall is calculated, no group is expanded, and no absence of grouping becomes an UNASSIGNED decision.\n\n+## Scope limitation\n+\n+The reviewer was assisted by an existing overlay. This is not an exhaustive independent blind gold standard and not absolute recall. Relation structure applies only to f68r1/f68r2/f68r3. No lexical, semantic or astronomical interpretation is made.\n\n+```text\n+AUGMENTED_HUMAN_REFERENCE_STATUS=COMPLETE\n+HUMAN_ADDED_OBJECTS_RECONCILED=YES\n+ASSISTED_RECALL_AI1_CALCULATED=YES\n+ASSISTED_RECALL_AI2_CALCULATED=YES\n+ASSISTED_RECALL_AI_UNION_CALCULATED=YES\n+RELATION_GROUP_PROTOCOL=3G1\n+FINAL_GROUPS=64\n+MULTIMEMBER_GROUPS_PRESERVED=YES\n+CARTESIAN_EDGES_INFERRED=NO\n+UNGROUPED_OBJECTS_PRESERVED=YES\n+FROZEN_ATTACHMENT_V2_CHANGED=NO\n+PREVIOUS_REPORTS_OVERWRITTEN=NO\n+RESULTS_REPRODUCIBLE=YES\n+```\n'''
    clean_report=report.replace('\n+','\n')
    clean_report=clean_report.replace('161 / 6 / 154 / 8','161 (48.9%) / 6 (1.8%) / 154 (46.8%) / 8 (2.4%)')
    clean_report=clean_report.replace('85 / 3 / 111 / 29','85 (37.3%) / 3 (1.3%) / 111 (48.7%) / 29 (12.7%)')
    clean_report=clean_report.replace(
        "AI2's observed advantage persists after augmentation. Relative to the pre-completeness confirmed denominator, AI1 changes from 167/321 to 167/329 for STAR and 88/199 to 88/228 for LABEL; AI2 changes from 315/321 to 315/329 and 196/199 to 196/228. Union common misses are 8 STAR and 29 LABEL.",
        "AI2's observed advantage persists after augmentation. Relative to the pre-completeness confirmed denominator, AI1 changes from 167/321 (52.0%) to 167/329 (50.8%) for STAR and 88/199 (44.2%) to 88/228 (38.6%) for LABEL; AI2 changes from 315/321 (98.1%) to 315/329 (95.7%) and 196/199 (98.5%) to 196/228 (86.0%). Union common misses are 8/329 STAR and 29/228 LABEL, so the common-miss frequency is appreciably higher for LABEL. AI1's marginal contribution beyond AI2 is 6/329 STAR and 3/228 LABEL; AI2's beyond AI1 is 154/329 and 111/228.")
    clean_report=clean_report.replace(
        'Sixteen relation-scope additions are grouped. HNEW_STAR_f68r2_DC27A556209F27B3',
        'Sixteen relation-scope additions are grouped, all in two-member newly created groups. Human-only objects occur on f67r2 (14), f67v1 (1), f68r2 (9), f68r3 (8), and f68v2 (5). HNEW_STAR_f68r2_DC27A556209F27B3')
    (out/'AUGMENTED_REFERENCE_REPORT.md').write_text(clean_report,encoding='utf-8')
    summary_md=f'''# Augmented reference summary\n\n+557 confirmed objects (329 STAR, 228 LABEL), including 37 reconciled human additions. Assisted recall: AI1 {metric('STAR','AI1')} STAR / {metric('LABEL','AI1')} LABEL; AI2 {metric('STAR','AI2')} / {metric('LABEL','AI2')}; union {metric('STAR','AI_UNION')} / {metric('LABEL','AI_UNION')}.\n\n+3G1 freezes 64 visual groups over three pages as hyperedges: 63 size-2 and one size-8. Sixteen of 17 relation-scope human additions are grouped; the one ungrouped STAR remains `NO_VISUAL_GROUP_ASSIGNED`. Frozen attachment-v2 and previous reports are unchanged. This is assisted, not absolute or independent blind-gold recall.\n'''
    (out/'AUGMENTED_REFERENCE_SUMMARY.md').write_text(summary_md.replace('\n+','\n'),encoding='utf-8')
    repro='''# Reproducibility\n\n+From this directory run:\n+\n+```bash\n+python3 -B scripts/analyze.py\n+python3 -B scripts/validate.py\n+python3 -B -m unittest discover -s tests -v\n+sha256sum -c --quiet SHA256SUMS\n+```\n+\n+Generation uses only registered local frozen inputs. Bootstrap seeds derive from SHA-256 metric keys; figures and tables are deterministic. `validate.py` regenerates into a temporary directory, compares every analytical output byte-for-byte, rechecks all input ledgers, then freezes this directory only. Existing source directories are read-only inputs.\n'''
    (out/'REPRODUCIBILITY.md').write_text(repro.replace('\n+','\n'),encoding='utf-8')

def main():
    p=argparse.ArgumentParser();p.add_argument('--output-dir',type=Path,default=PKG);a=p.parse_args();build(a.output_dir)
if __name__=='__main__':main()
