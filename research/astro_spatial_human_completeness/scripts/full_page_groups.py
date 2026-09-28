#!/usr/bin/env python3
"""Build/validate full-page CVAT grouping workflow for visual LABEL--STAR associations."""
from __future__ import annotations

import argparse
from collections import defaultdict
import hashlib
import json
from pathlib import Path
import xml.etree.ElementTree as ET

from core import (PKG,HUMAN,REF_FIELDS,attr,attrs,check,digest,jswrite,parse_geometry,
                  read,relation_schema,render_xml,root_xml,safe_output,same_geometry,
                  shape_element,unique,utc,write,xml_input,zip_bytes,verify_inputs)
from attachment import gated_endpoints

GROUP_VERSION='3G1'
PANELS=('f68r1','f68r2','f68r3')
COMPLETION=('NOT_STARTED','COMPLETE','INCOMPLETE_TECHNICAL','INCOMPLETE_REVIEWER_UNCERTAINTY')

def group_schema():
    ref=[{'name':'reference_id','input_type':'text','default_value':'UNSET','values':['UNSET'],'mutable':False},
         {'name':'relation_group_protocol_version','input_type':'text','default_value':GROUP_VERSION,'values':[GROUP_VERSION],'mutable':False}]
    marker=[{'name':'panel','input_type':'select','default_value':'f68r1','values':list(PANELS),'mutable':False},
            {'name':'relation_group_protocol_version','input_type':'text','default_value':GROUP_VERSION,'values':[GROUP_VERSION],'mutable':False},
            {'name':'completion_status','input_type':'select','default_value':'NOT_STARTED','values':list(COMPLETION),'mutable':True},
            {'name':'human_confidence','input_type':'select','default_value':'UNSET','values':['UNSET','LOW','MEDIUM','HIGH'],'mutable':True}]
    return [{'name':'STAR_REFERENCE','type':'rectangle','color':'#349866','attributes':ref},
            {'name':'LABEL_REFERENCE','type':'any','color':'#3578b8','attributes':ref},
            {'name':'PAGE_RELATION_REVIEW','type':'tag','color':'#555555','attributes':marker}]

def prior_positive_groups(endpoints):
    index=unique(endpoints,'canonical_id')
    rows=read(HUMAN/'HUMAN_LABEL_OBJECT_RELATIONS.tsv')
    edges=[]
    for row in rows:
        if row['relation_type']!='VISUAL_LABEL_OF':continue
        lid,sid=row['label_candidate_id'],row['object_candidate_id']
        if lid not in index or sid not in index:continue
        if index[lid]['panel'] not in PANELS:continue
        check(index[lid]['panel']==index[sid]['panel'],'prior relation crosses panels')
        edges.append((lid,sid,row['relation_candidate_id']))
    check(len(edges)==48,'expected 48 frozen positive v2 edges on f68r1/f68r2/f68r3')
    adjacency=defaultdict(set);edge_ids=defaultdict(list)
    for lid,sid,rid in edges:
        adjacency[lid].add(sid);adjacency[sid].add(lid)
        edge_ids[index[lid]['panel']].append((lid,sid,rid))
    groups=[];seen=set()
    for start in sorted(adjacency,key=lambda cid:(index[cid]['panel'],cid)):
        if start in seen:continue
        stack=[start];seen.add(start);members=[]
        while stack:
            cid=stack.pop();members.append(cid)
            for other in adjacency[cid]:
                if other not in seen:seen.add(other);stack.append(other)
        panel=index[start]['panel'];member_set=set(members)
        relation_ids=sorted(rid for lid,sid,rid in edge_ids[panel] if lid in member_set and sid in member_set)
        groups.append({'panel':panel,'members':tuple(sorted(members)),'prior_relation_ids':';'.join(relation_ids)})
    check(len(groups)==46,'expected 46 connected visual-association groups')
    return groups

def build(out,freeze_manifest,package=PKG):
    verify_inputs(package);endpoints,freeze_hash=gated_endpoints(freeze_manifest,package)
    endpoints=[{k:r[k] for k in REF_FIELDS} for r in endpoints if r['panel'] in PANELS]
    check(len(endpoints)==256,'unexpected f68r1/f68r2/f68r3 endpoint count')
    index=unique(endpoints,'canonical_id');groups=prior_positive_groups(endpoints)
    member_group={cid:n for n,g in enumerate(groups,1) for cid in g['members']}
    dimensions=unique(read(package/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv'),'panel')
    root=root_xml('attachment-v3-full-pages-grouped',group_schema(),len(PANELS))
    images={};initial=[]
    for frame,panel in enumerate(PANELS):
        info=dimensions[panel];filename=info['filename']
        im=ET.SubElement(root,'image',id=str(frame),name=filename,width=info['width'],height=info['height'])
        for row in sorted((r for r in endpoints if r['panel']==panel),key=lambda r:r['canonical_id']):
            shape=shape_element(im,row,row['object_type']+'_REFERENCE',version_field='relation_group_protocol_version',version=GROUP_VERSION)
            if row['canonical_id'] in member_group:shape.set('group_id',str(member_group[row['canonical_id']]))
        tag=ET.SubElement(im,'tag',label='PAGE_RELATION_REVIEW',source='manual')
        for name,value in [('panel',panel),('relation_group_protocol_version',GROUP_VERSION),
                           ('completion_status','NOT_STARTED'),('human_confidence','UNSET')]:attr(tag,name,value)
        images[filename]=(HUMAN/'images'/filename).read_bytes()
    for number,g in enumerate(groups,1):
        labels=sorted(cid for cid in g['members'] if index[cid]['object_type']=='LABEL')
        stars=sorted(cid for cid in g['members'] if index[cid]['object_type']=='STAR')
        initial.append({'initial_group_id':number,'panel':g['panel'],'label_ids':';'.join(labels),'star_ids':';'.join(stars),
            'member_count':len(g['members']),'prior_relation_ids':g['prior_relation_ids'],'source':'FROZEN_ATTACHMENT_V2_VISUAL_LABEL_OF_CONNECTED_COMPONENT'})
    out=safe_output(out);check(not out.exists(),'use a new full-page group directory');out.mkdir(parents=True)
    xml=render_xml(root);(out/'annotations.xml').write_bytes(xml)
    zip_bytes(out/'ANNOTATIONS.zip',{'annotations.xml':xml});zip_bytes(out/'IMAGES.zip',images)
    (out/'CVAT_SCHEMA.json').write_text(json.dumps(group_schema(),indent=2)+'\n',encoding='utf-8')
    write(out/'ENDPOINT_SNAPSHOT.tsv',endpoints,REF_FIELDS)
    write(out/'INITIAL_GROUPS.tsv',initial,['initial_group_id','panel','label_ids','star_ids','member_count','prior_relation_ids','source'])
    manifest={'mode':'PRODUCTION_FULL_PAGE_GROUP_REVIEW','relation_group_protocol_version':GROUP_VERSION,
              'object_freeze_manifest_path':str(Path(freeze_manifest).resolve()),'object_freeze_sha256':freeze_hash,
              'panels':list(PANELS),'endpoints':len(endpoints),'initial_groups':len(initial),
              'f68r3_scope':'REVIEWER_CONFIRMED_FULL_CANONICAL_PAGE_GROUP_REVIEW',
              'f68r3_authorization_sha256':digest(package/'F68R3_FULL_PAGE_GROUP_REVIEW_AUTHORIZATION.md'),
              'new_endpoints':sum(r['canonical_id'].startswith('HNEW_') for r in endpoints),
              'group_semantics':'VISUAL_ASSOCIATION_HYPEREDGE_NO_CARTESIAN_EXPANSION',
              'files':{p.name:digest(p) for p in sorted(out.iterdir()) if p.is_file()}}
    jswrite(out/'GROUP_TASK_MANIFEST.json',manifest)
    print(f'FULL_PAGE_GROUP_TASK=READY; PANELS={len(PANELS)}; ENDPOINTS={len(endpoints)}; INITIAL_GROUPS={len(initial)}')

def canonical_group(panel,members):
    return 'HGROUP_'+hashlib.sha256((GROUP_VERSION+'|'+panel+'|'+'|'.join(sorted(members))).encode()).hexdigest()[:16].upper()

def validate(root,task_dir,reviewer,timestamp,allow_incomplete=False):
    utc(timestamp);check(reviewer,'reviewer provenance missing')
    meta=json.loads((task_dir/'GROUP_TASK_MANIFEST.json').read_text())
    check(meta['relation_group_protocol_version']==GROUP_VERSION,'wrong group protocol')
    for name,h in meta['files'].items():check(digest(task_dir/name)==h,'group task input changed: '+name)
    endpoints=unique(read(task_dir/'ENDPOINT_SNAPSHOT.tsv'),'reference_id')
    initial=read(task_dir/'INITIAL_GROUPS.tsv')
    verified,freeze_hash=gated_endpoints(meta['object_freeze_manifest_path'])
    verified=[{k:r[k] for k in REF_FIELDS} for r in verified if r['panel'] in PANELS]
    check(freeze_hash==meta['object_freeze_sha256'] and verified==list(endpoints.values()),'object freeze/snapshot mismatch')
    dimensions=unique(read(PKG/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv'),'panel')
    check(meta.get('f68r3_scope')=='REVIEWER_CONFIRMED_FULL_CANONICAL_PAGE_GROUP_REVIEW' and
          meta.get('f68r3_authorization_sha256')==digest(PKG/'F68R3_FULL_PAGE_GROUP_REVIEW_AUTHORIZATION.md'),'f68r3 full-page authorization missing/changed')
    images=root.findall('image');check(len(images)==3 and not root.findall('track'),'expected three full-page image frames')
    observed=set();groups=defaultdict(list);markers=[]
    for image in images:
        panel=Path(image.get('name','')).stem;check(panel in PANELS,'unexpected panel')
        info=dimensions[panel]
        check(image.get('id')==str(PANELS.index(panel)) and image.get('width')==info['width'] and image.get('height')==info['height'],'full-page frame changed')
        tags=[]
        for element in image:
            a=attrs(element);label=element.get('label')
            if label in {'STAR_REFERENCE','LABEL_REFERENCE'}:
                check(set(a)=={'reference_id','relation_group_protocol_version'} and a['relation_group_protocol_version']==GROUP_VERSION,'endpoint attrs changed')
                rid=a['reference_id'];check(rid in endpoints and rid not in observed,'endpoint missing/duplicated/renamed')
                row=endpoints[rid];check(row['panel']==panel and label==row['object_type']+'_REFERENCE','endpoint panel/class changed')
                check(element.get('source','manual') in {'manual','file'} and element.get('occluded','0')=='0' and element.get('z_order','0')=='0','endpoint service properties changed')
                same_geometry(parse_geometry(element,row['object_type'],(int(info['width']),int(info['height'])),True),row)
                gid=element.get('group_id')
                if gid not in (None,'','0'):
                    check(gid.isdigit() and int(gid)>0,'invalid group_id');groups[(panel,gid)].append(row)
                observed.add(rid)
            else:
                check(element.tag=='tag' and label=='PAGE_RELATION_REVIEW','unexpected object in group task')
                check(set(a)=={'panel','relation_group_protocol_version','completion_status','human_confidence'},'page marker attrs changed')
                check(a['panel']==panel and a['relation_group_protocol_version']==GROUP_VERSION and a['completion_status'] in COMPLETION,'invalid page marker')
                if a['completion_status']=='COMPLETE':check(a['human_confidence'] in {'LOW','MEDIUM','HIGH'},'completed page confidence unset')
                else:check(allow_incomplete,'page group review incomplete')
                tags.append(a)
        check(len(tags)==1,'exactly one page review marker required');markers.append(tags[0])
    check(observed==set(endpoints),'endpoint deleted from full-page task')
    result=[]
    for (panel,gid),members in sorted(groups.items()):
        labels=sorted(r['canonical_id'] for r in members if r['object_type']=='LABEL')
        stars=sorted(r['canonical_id'] for r in members if r['object_type']=='STAR')
        check(labels and stars and len(members)>=2,'every group must contain LABEL and STAR')
        ids=labels+stars
        result.append({'canonical_group_id':canonical_group(panel,ids),'panel':panel,'cvat_group_id':gid,
                       'label_ids':';'.join(labels),'star_ids':';'.join(stars),'member_count':len(ids),
                       'relation_semantics':'VISUAL_ASSOCIATION_GROUP','cartesian_edges_inferred':'NO',
                       'reviewer_id':reviewer,'review_timestamp':timestamp,
                       'human_confidence':next(m['human_confidence'] for m in markers if m['panel']==panel),
                       'relation_group_protocol_version':GROUP_VERSION})
    old={canonical_group(r['panel'],r['label_ids'].split(';')+r['star_ids'].split(';')):r for r in initial}
    current={r['canonical_group_id']:r for r in result}
    audit=[]
    for gid,row in old.items():audit.append({'initial_canonical_group_id':gid,'panel':row['panel'],
        'review_outcome':'UNCHANGED_PRIOR_GROUP' if gid in current else 'REMOVED_OR_REGROUPED_PRIOR_GROUP','prior_v2_unchanged':'YES'})
    for gid,row in current.items():
        if gid not in old:audit.append({'initial_canonical_group_id':gid,'panel':row['panel'],'review_outcome':'NEW_OR_MODIFIED_GROUP','prior_v2_unchanged':'YES'})
    return result,audit,markers

def main():
    p=argparse.ArgumentParser(description=__doc__);sub=p.add_subparsers(dest='command',required=True)
    b=sub.add_parser('build');b.add_argument('--output-dir',type=Path,required=True);b.add_argument('--object-freeze-manifest',type=Path,required=True)
    v=sub.add_parser('validate');v.add_argument('--input',type=Path,required=True);v.add_argument('--task-dir',type=Path,required=True)
    v.add_argument('--reviewer-id',required=True);v.add_argument('--timestamp',required=True);v.add_argument('--output-dir',type=Path);v.add_argument('--allow-incomplete',action='store_true')
    args=p.parse_args()
    if args.command=='build':build(args.output_dir,args.object_freeze_manifest)
    else:
        groups,audit,markers=validate(xml_input(args.input),args.task_dir,args.reviewer_id,args.timestamp,args.allow_incomplete)
        if args.output_dir:
            out=safe_output(args.output_dir);check(not out.exists(),'use a new validation output directory');out.mkdir(parents=True)
            write(out/'REVIEWED_RELATION_GROUPS.tsv',groups)
            write(out/'GROUP_CHANGE_AUDIT.tsv',audit)
            write(out/'PANEL_REVIEW_MARKERS.tsv',markers)
        print(f'GROUP_VALIDATION=PASS; GROUPS={len(groups)}; CARTESIAN_EDGES_INFERRED=NO')
if __name__=='__main__':main()
