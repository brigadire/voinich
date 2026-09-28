#!/usr/bin/env python3
"""Attachment v3: current-reference dry run or strictly gated future queue."""
from __future__ import annotations
import argparse
from collections import Counter
from datetime import datetime
import hashlib
import io
import json
import math
from pathlib import Path
import xml.etree.ElementTree as ET
from PIL import Image
from core import (PKG, HUMAN, OBJECT_VERSION, RELATION_VERSION, REF_FIELDS, CHECKS,
                  read, write, jswrite, check, unique, digest, utc, verify_inputs, safe_output,
                  envelope, root_xml, relation_schema, attr, shape_element, render_xml, zip_bytes,
                  xml_input, attrs, parse_geometry, same_geometry, stable_id)

def load_scope(package=PKG,scope_file=None):
    path=scope_file or package/'ATTACHMENT_ELIGIBILITY.json'
    scope=json.loads(Path(path).read_text())
    panels=unique(read(package/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv'),'panel')
    check(set(scope['panels'])==set(panels),'eligibility scope does not cover actual canonical panels')
    check(scope['relation_protocol_version']==RELATION_VERSION,'scope must be attachment 3, not legacy spatial/v2')
    check(scope['panels'].get('f68r1')=='ELIGIBLE' and scope['panels'].get('f68r2')=='ELIGIBLE','main eligible panels missing')
    for panel,state in scope['panels'].items():
        if panel not in {'f68r1','f68r2','f68r3'}:check(state=='PANEL_NOT_APPLICABLE_FOR_LABEL_STAR_ATTACHMENT','non-applicable panel enabled')
    rule=scope.get('f68r3_rule')
    if scope['panels'].get('f68r3')=='ELIGIBLE_LIMITED':
        check(rule and rule.get('status')=='REVIEWER_CONFIRMED_AND_FROZEN','f68r3 has no frozen eligibility confirmation')
        check(rule.get('reviewer_id') and rule.get('confirmation_timestamp'),'f68r3 confirmation provenance missing')
        utc(rule['confirmation_timestamp'])
        body={k:v for k,v in rule.items() if k!='rule_sha256'}
        h=hashlib.sha256(json.dumps(body,sort_keys=True,separators=(',',':')).encode()).hexdigest()
        check(rule.get('rule_sha256')==h,'f68r3 eligibility rule hash mismatch')
        check(rule['kind'] in {'WHITELIST','POLYGON_CENTER_MASK'},'unsupported limited scope')
        if rule['kind']=='WHITELIST':
            check(rule.get('star_ids') and rule.get('label_ids'),'whitelist must specify STAR and LABEL IDs')
            check(len(rule['star_ids'])==len(set(rule['star_ids'])) and len(rule['label_ids'])==len(set(rule['label_ids'])),'duplicate whitelist IDs')
        else:
            points=rule.get('polygon',[])
            check(len(points)>=3 and all(isinstance(p,list) and len(p)==2 and all(isinstance(v,(int,float)) and not isinstance(v,bool) and math.isfinite(v) for v in p) for p in points),'invalid polygon mask')
            w,h=int(panels['f68r3']['width']),int(panels['f68r3']['height'])
            check(all(0<=x<=w and 0<=y<=h for x,y in points),'eligibility mask outside canonical image')
            polygon_area=abs(sum(a[0]*b[1]-a[1]*b[0] for a,b in zip(points,points[1:]+points[:1])))/2
            check(polygon_area>0 and len({tuple(p) for p in points})>=3,'zero-area/degenerate eligibility mask')
    else:
        check(scope['panels'].get('f68r3')=='AWAITING_REVIEWER_CONFIRMATION' and rule is None,'unconfirmed f68r3 rule must remain disabled')
    return scope,digest(path)

def in_polygon(x,y,points):
    inside=False
    for (ax,ay),(bx,by) in zip(points,points[1:]+points[:1]):
        if abs((bx-ax)*(y-ay)-(by-ay)*(x-ax))<1e-6 and min(ax,bx)-1e-6<=x<=max(ax,bx)+1e-6 and min(ay,by)-1e-6<=y<=max(ay,by)+1e-6:return True
        if (ay>y)!=(by>y) and x<(bx-ax)*(y-ay)/(by-ay)+ax:inside=not inside
    return inside

def eligible(endpoint,scope):
    panel=endpoint['panel']
    if panel in {'f68r1','f68r2'}:return True
    if panel!='f68r3' or scope['panels'][panel]!='ELIGIBLE_LIMITED':return False
    rule=scope['f68r3_rule']
    if rule['kind']=='WHITELIST':return endpoint['canonical_id'] in rule['star_ids' if endpoint['object_type']=='STAR' else 'label_ids']
    return in_polygon((float(endpoint['bbox_x1'])+float(endpoint['bbox_x2']))/2,
                      (float(endpoint['bbox_y1'])+float(endpoint['bbox_y2']))/2,rule['polygon'])

def gated_endpoints(manifest_file,package=PKG):
    """Accept only an externally reviewed freeze. This code never creates that freeze."""
    manifest_file=Path(manifest_file)
    m=json.loads(manifest_file.read_text())
    check(m.get('status')=='OBJECT_COMPLETENESS_REVIEWED_AND_FROZEN' and m.get('object_protocol_version')==OBJECT_VERSION,'object completeness freeze gate not satisfied')
    check(m.get('reviewer_id') and m.get('freeze_timestamp'),'object freeze has no reviewer provenance')
    utc(m['freeze_timestamp'])
    check(m.get('prepared_reference_sha256')==digest(package/'REFERENCE_OBJECTS.tsv'),'object freeze references another baseline')
    def file(key):
        entry=m[key];path=(manifest_file.parent/entry['path']).resolve()
        check(path.is_relative_to(manifest_file.parent.resolve()),'freeze file escapes manifest directory')
        check(path.is_file() and digest(path)==entry['sha256'],f'object freeze file checksum invalid: {key}')
        return read(path)
    endpoints=file('confirmed_endpoints');proposals=file('proposed_additions');completion=file('panel_completion')
    base=unique(read(package/'REFERENCE_OBJECTS.tsv'),'canonical_id')
    index=unique(endpoints,'canonical_id');new_index=unique(proposals,'canonical_id')
    check(set(base)<=set(index),'object freeze deleted confirmed baseline endpoints')
    for cid,r in base.items():
        current=index[cid]
        check(all(current[k]==r[k] for k in REF_FIELDS),'object freeze changed baseline endpoint geometry/attributes')
    panels=unique(read(package/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv'),'panel')
    completed=unique(completion,'panel')
    check(set(completed)==set(panels),'object completeness panels missing')
    for panel,row in completed.items():
        count=sum(p['panel']==panel for p in proposals)
        check(row['completion_status']==('COMPLETE_WITH_NEW_OBJECTS' if count else 'COMPLETE_NO_NEW_OBJECTS'),'object freeze panel not explicitly complete')
        check(all(row[k]=='true' for k in CHECKS) and row['completeness_protocol_version']==OBJECT_VERSION,'object systematic review checks missing')
        check(row['frame']==panels[panel]['frame'],'object completion frame mismatch')
    for cid,row in index.items():
        check(row['object_type'] in {'STAR','LABEL'} and row['panel'] in panels,'invalid endpoint class/panel')
        check(row['geometry_type'] in {'BOX','ELLIPSE'} and (row['object_type']!='STAR' or row['geometry_type']=='BOX'),'invalid endpoint geometry')
        if cid in base:continue
        check(cid in new_index,'new endpoint has no ingested human proposal')
        p=new_index[cid]
        check(cid==stable_id(p) and p['frame']==panels[row['panel']]['frame'],'new canonical ID/frame is not the ingested stable identity')
        check(p['human_confidence'] in {'LOW','MEDIUM','HIGH'} and p['reviewer_id'],'proposal confidence/reviewer provenance missing')
        utc(p['review_timestamp'])
        check(p['creation_origin']=='HUMAN_COMPLETENESS_PASS' and p['completeness_protocol_version']==OBJECT_VERSION,'foreign-origin proposal')
        check(p['addition_status'] in {'PROPOSED_NEW','POSSIBLE_DUPLICATE_NEW'},'UNCERTAIN_NEW/unknown status cannot be an ordinary endpoint')
        probe=ET.Element('image')
        serialized=shape_element(probe,row,row['object_type']+'_REFERENCE')
        parse_geometry(serialized,row['object_type'],(int(panels[row['panel']]['width']),int(panels[row['panel']]['height'])))
        check(row.get('post_review_decision')=='CONFIRMED_NEW' and row.get('confirmation_reviewer_id') and row.get('confirmation_timestamp'),'new object not explicitly post-reviewed')
        utc(row['confirmation_timestamp'])
        check(all(row[k]==p[k] for k in ('panel','object_type','geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')),'new confirmed endpoint differs from proposal; use an explicit protocol amendment for geometry revision')
        check(row.get('reconciliation_status')=='EXPLICITLY_RECONCILED_NO_UNRESOLVED_CONFLICT','new endpoint post-review reconciliation missing')
        check(row['reference_id']=='REF_'+hashlib.sha256(cid.encode()).hexdigest()[:16].upper(),'new endpoint blind ID mismatch')
        check(row['layer']==row['object_type']+'_REFERENCE','confirmed new endpoint layer/class mismatch')
    return endpoints,digest(manifest_file)

def queue(endpoints,scope,package=PKG,recheck_prior=False,cover_new_endpoints=False,targeted_new_and_prior_only=False):
    check(not targeted_new_and_prior_only or (recheck_prior and cover_new_endpoints),
          'targeted review requires prior recheck plus complete new-endpoint coverage')
    unique(endpoints,'canonical_id');unique(endpoints,'reference_id')
    prior={(r['label_candidate_id'],r['object_candidate_id']):r for r in read(HUMAN/'HUMAN_LABEL_OBJECT_RELATIONS.tsv')}
    pairs=[];filtered=[];retained=[]
    idx=unique(endpoints,'canonical_id')
    if scope['panels']['f68r3']=='ELIGIBLE_LIMITED' and scope['f68r3_rule']['kind']=='WHITELIST':
        rule=scope['f68r3_rule']
        for kind,key in [('STAR','star_ids'),('LABEL','label_ids')]:
            check(all(cid in idx and idx[cid]['panel']=='f68r3' and idx[cid]['object_type']==kind for cid in rule[key]),'f68r3 whitelist contains missing/wrong-class endpoints')
    # Spatial indexes by panel avoid generating the non-applicable Cartesian universe.
    for panel in ('f68r1','f68r2','f68r3'):
        stars=[r for r in endpoints if r['panel']==panel and r['object_type']=='STAR' and eligible(r,scope)]
        labels=[r for r in endpoints if r['panel']==panel and r['object_type']=='LABEL' and eligible(r,scope)]
        for label in sorted(labels,key=lambda r:r['canonical_id']):
            le=envelope(label);lcx=(float(label['bbox_x1'])+float(label['bbox_x2']))/2
            for star in sorted(stars,key=lambda r:r['canonical_id']):
                se=envelope(star);scx=(float(star['bbox_x1'])+float(star['bbox_x2']))/2
                key=(label['canonical_id'],star['canonical_id'])
                reason='RIGHT_SIDE_AND_ORIENTED_Y_BAND' if panel!='f68r3' else 'CONFIRMED_F68R3_LIMITED_SCOPE_AND_ORIENTED_Y_BAND'
                accepted=(lcx>scx if panel!='f68r3' else True) and min(le[3],se[3])>=max(le[1],se[1])
                prior_pair=prior.get(key)
                if prior_pair:
                    retained.append({'panel':panel,'label_id':key[0],'star_id':key[1],'prior_relation_id':prior[key]['relation_candidate_id'],
                                     'prior_protocol_version':'2','handling':'RETAIN_FROZEN_DECISION_NO_REVIEW_NO_CONVERSION'})
                    if not recheck_prior:continue
                if not accepted and not prior_pair:
                    filtered.append({'panel':panel,'label_id':key[0],'star_id':key[1],
                                     'reason':'LABEL_NOT_RIGHT' if panel!='f68r3' and lcx<=scx else 'NO_ORIENTED_Y_OVERLAP',
                                     'status':'FILTERED_PAIR_NOT_A_HUMAN_OBSERVATION','relation_protocol_version':RELATION_VERSION})
                    continue
                distance=math.hypot(lcx-scx,(float(label['bbox_y1'])+float(label['bbox_y2'])-float(star['bbox_y1'])-float(star['bbox_y2']))/2)
                pair_id='HC_REL_'+hashlib.sha256((RELATION_VERSION+'|'+key[0]+'|'+key[1]).encode()).hexdigest()[:16].upper()
                pairs.append({'pair_id':pair_id,'panel':panel,'label_id':key[0],'star_id':key[1],
                              'label_reference_id':label['reference_id'],'star_reference_id':star['reference_id'],
                              'inclusion_reason':'FROZEN_V2_PAIR_RECHECK' if prior_pair else reason,'center_distance_px':f'{distance:.6f}',
                              'label_geometry':json.dumps({k:label[k] for k in ('geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')},sort_keys=True),
                              'star_geometry':json.dumps({k:star[k] for k in ('geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')},sort_keys=True),
                              'panel_eligibility':'ELIGIBLE' if panel!='f68r3' else 'ELIGIBLE_LIMITED',
                              'relation_protocol_version':RELATION_VERSION,'initial_decision':'UNREVIEWED'})
    if cover_new_endpoints:
        queued_keys={(r['label_id'],r['star_id']) for r in pairs}
        represented={cid for key in queued_keys for cid in key}
        missing=[r for r in endpoints if r['canonical_id'].startswith('HNEW_') and eligible(r,scope) and r['canonical_id'] not in represented]
        for endpoint in sorted(missing,key=lambda r:r['canonical_id']):
            panel=endpoint['panel'];ecx=(float(endpoint['bbox_x1'])+float(endpoint['bbox_x2']))/2
            other_kind='LABEL' if endpoint['object_type']=='STAR' else 'STAR'
            candidates=[];all_candidates=[]
            for other in endpoints:
                if other['panel']!=panel or other['object_type']!=other_kind or not eligible(other,scope):continue
                ocx=(float(other['bbox_x1'])+float(other['bbox_x2']))/2
                ecy=(float(endpoint['bbox_y1'])+float(endpoint['bbox_y2']))/2
                ocy=(float(other['bbox_y1'])+float(other['bbox_y2']))/2
                candidate=(math.hypot(ocx-ecx,ocy-ecy),other);all_candidates.append(candidate)
                if panel not in {'f68r1','f68r2'} or (ocx>ecx if other_kind=='LABEL' else ecx>ocx):candidates.append(candidate)
            used_any_direction=not candidates
            if used_any_direction:candidates=all_candidates
            check(candidates,f'new eligible endpoint has no opposite-class fallback counterpart: {endpoint["canonical_id"]}')
            distance,other=min(candidates,key=lambda item:(item[0],item[1]['canonical_id']))
            label=other if other_kind=='LABEL' else endpoint;star=endpoint if endpoint['object_type']=='STAR' else other
            key=(label['canonical_id'],star['canonical_id'])
            check(key not in queued_keys,'fallback duplicated queued pair')
            pair_id='HC_REL_'+hashlib.sha256((RELATION_VERSION+'|'+key[0]+'|'+key[1]).encode()).hexdigest()[:16].upper()
            pairs.append({'pair_id':pair_id,'panel':panel,'label_id':key[0],'star_id':key[1],
                          'label_reference_id':label['reference_id'],'star_reference_id':star['reference_id'],
                          'inclusion_reason':('NEW_ENDPOINT_NEAREST_ANY_DIRECTION_FALLBACK_NO_DIRECTIONAL_CANDIDATE' if used_any_direction else
                                              'NEW_ENDPOINT_NEAREST_DIRECTIONAL_FALLBACK_NO_Y_OVERLAP'),'center_distance_px':f'{distance:.6f}',
                          'label_geometry':json.dumps({k:label[k] for k in ('geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')},sort_keys=True),
                          'star_geometry':json.dumps({k:star[k] for k in ('geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation')},sort_keys=True),
                          'panel_eligibility':'ELIGIBLE' if panel!='f68r3' else 'ELIGIBLE_LIMITED',
                          'relation_protocol_version':RELATION_VERSION,'initial_decision':'UNREVIEWED'})
            queued_keys.add(key);represented.update(key)
            filtered=[r for r in filtered if (r['label_id'],r['star_id'])!=key]
    if targeted_new_and_prior_only:
        prior_ids={r['prior_relation_id'] for r in retained}
        prior_keys={(r['label_id'],r['star_id']) for r in retained}
        new_ids={r['canonical_id'] for r in endpoints if r['canonical_id'].startswith('HNEW_') and eligible(r,scope)}
        selected={r['pair_id'] for r in pairs if (r['label_id'],r['star_id']) in prior_keys}
        for cid in sorted(new_ids):
            candidates=[r for r in pairs if cid in {r['label_id'],r['star_id']}]
            check(candidates,f'new eligible endpoint has no review candidate: {cid}')
            selected.add(min(candidates,key=lambda r:(float(r['center_distance_px']),r['pair_id']))['pair_id'])
        removed=[r for r in pairs if r['pair_id'] not in selected]
        for r in removed:
            filtered.append({'panel':r['panel'],'label_id':r['label_id'],'star_id':r['star_id'],
                             'reason':'NOT_NEAREST_NEW_ENDPOINT_CANDIDATE_OR_NOT_FROZEN_V2_PAIR',
                             'status':'FILTERED_PAIR_NOT_A_HUMAN_OBSERVATION','relation_protocol_version':RELATION_VERSION})
        pairs=[r for r in pairs if r['pair_id'] in selected]
        represented={cid for r in pairs for cid in (r['label_id'],r['star_id'])}
        check(new_ids<=represented,'targeted queue lost a new eligible endpoint')
        check(sum((r['label_id'],r['star_id']) in prior_keys for r in pairs)==len(prior_keys),'targeted queue lost/duplicated a frozen v2 pair')
    return sorted(pairs,key=lambda r:(r['panel'],r['pair_id'])),filtered,retained

def generate(out,endpoints,scope,package=PKG,mode='NOT_FOR_PRODUCTION',object_hash='CURRENT_REFERENCE_ONLY',object_manifest_path=None,
             recheck_prior=False,group_pairs=False,cover_new_endpoints=False,targeted_new_and_prior_only=False):
    if mode!='NOT_FOR_PRODUCTION':
        check(object_manifest_path is not None,'production generation requires the actual object freeze manifest')
        verified,h=gated_endpoints(object_manifest_path,package)
        check(verified==endpoints and h==object_hash,'production endpoints are not the verified object snapshot')
    out=safe_output(out);out.mkdir(parents=True,exist_ok=True)
    pairs,filtered,retained=queue(endpoints,scope,package,recheck_prior,cover_new_endpoints,targeted_new_and_prior_only)
    index={r['canonical_id']:r for r in endpoints}
    root=root_xml('human-attachment-completeness-v3-'+mode,relation_schema(),len(pairs))
    image_entries={};crops=[]
    dimensions={r['panel']:(int(r['width']),int(r['height'])) for r in read(package/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv')}
    for frame,pair in enumerate(pairs):
        label,star=index[pair['label_id']],index[pair['star_id']]
        le,se=envelope(label),envelope(star);w,h=dimensions[pair['panel']]
        ox,oy=max(0,math.floor(min(le[0],se[0]))-80),max(0,math.floor(min(le[1],se[1]))-80)
        x2,y2=min(w,math.ceil(max(le[2],se[2]))+80),min(h,math.ceil(max(le[3],se[3]))+80)
        image=Image.open(HUMAN/'images'/f'{pair["panel"]}.jpg').convert('RGB').crop((ox,oy,x2,y2))
        buffer=io.BytesIO();image.save(buffer,format='JPEG',quality=92,subsampling=0)
        filename=pair['pair_id']+'.jpg';image_entries['images/'+filename]=buffer.getvalue()
        im=ET.SubElement(root,'image',id=str(frame),name=filename,width=str(x2-ox),height=str(y2-oy))
        star_shape=shape_element(im,star,'STAR_REFERENCE',(ox,oy),'relation_protocol_version',RELATION_VERSION)
        label_shape=shape_element(im,label,'LABEL_REFERENCE',(ox,oy),'relation_protocol_version',RELATION_VERSION)
        if group_pairs:
            # One group per pair-frame preserves independent many-to-many edges.
            star_shape.set('group_id',str(frame+1));label_shape.set('group_id',str(frame+1))
        tag=ET.SubElement(im,'tag',label='PAIR_ATTACHMENT',source='manual')
        for name,value in [('pair_id',pair['pair_id']),('relation_protocol_version',RELATION_VERSION),('decision','UNREVIEWED'),('human_confidence','UNSET')]:attr(tag,name,value)
        crops.append({'pair_id':pair['pair_id'],'frame':frame,'panel':pair['panel'],'filename':filename,'origin_x':ox,'origin_y':oy,
                      'width':x2-ox,'height':y2-oy,'image_sha256':hashlib.sha256(buffer.getvalue()).hexdigest(),'mode':mode})
    write(out/'QUEUE.tsv',pairs,['pair_id','panel','label_id','star_id','label_reference_id','star_reference_id','inclusion_reason','center_distance_px','label_geometry','star_geometry','panel_eligibility','relation_protocol_version','initial_decision'])
    write(out/'FILTERED_PAIRS.tsv',filtered,['panel','label_id','star_id','reason','status','relation_protocol_version'])
    write(out/'RETAINED_PRIOR_V2.tsv',retained,['panel','label_id','star_id','prior_relation_id','prior_protocol_version','handling'])
    write(out/'CROP_MANIFEST.tsv',crops,['pair_id','frame','panel','filename','origin_x','origin_y','width','height','image_sha256','mode'])
    # The reviewer-authorized freeze carries additional confirmation columns;
    # the blind relation snapshot intentionally exposes only neutral REF_FIELDS.
    write(out/'ENDPOINT_SNAPSHOT.tsv',[{k:r[k] for k in REF_FIELDS} for r in endpoints],REF_FIELDS)
    jswrite(out/'ELIGIBILITY_SNAPSHOT.json',scope)
    xml=render_xml(root);(out/'annotations.xml').write_bytes(xml)
    zip_bytes(out/'ANNOTATIONS.zip',{'annotations.xml':xml});zip_bytes(out/'IMAGES.zip',image_entries)
    jswrite(out/'QUEUE_MANIFEST.json',{'mode':mode,'relation_protocol_version':RELATION_VERSION,'object_freeze_sha256':object_hash,
                                    'object_freeze_manifest_path':str(Path(object_manifest_path).resolve()) if object_manifest_path else None,
                                    'queued_pairs':len(pairs),'filtered_pairs_not_observations':len(filtered),'retained_v2_pairs':len(retained),
                                    'prior_v2_pairs_included_for_blind_recheck':recheck_prior,'pair_shapes_grouped':group_pairs,
                                    'new_endpoint_fallback_coverage':cover_new_endpoints,
                                    'targeted_new_and_frozen_prior_only':targeted_new_and_prior_only,
                                    'eligible_panels':sorted({p['panel'] for p in pairs}),'f68r3_included':any(p['panel']=='f68r3' for p in pairs),
                                    'files':{p.name:digest(p) for p in sorted(out.iterdir()) if p.is_file() and p.name!='QUEUE_MANIFEST.json'}})
    return pairs,filtered,retained

def validate_relation(root,queue_dir,reviewer='DRY_RUN',timestamp='2000-01-01T00:00:00Z',raw_sha='SYNTHETIC_COPIED_TREE'):
    utc(timestamp);check(reviewer,'reviewer provenance missing')
    meta=json.loads((queue_dir/'QUEUE_MANIFEST.json').read_text())
    check(meta['relation_protocol_version']==RELATION_VERSION,'queue is not attachment v3')
    for name,h in meta['files'].items():check(digest(queue_dir/name)==h,'queue input changed')
    if meta['mode']=='NOT_FOR_PRODUCTION':check(reviewer=='DRY_RUN','dry-run review cannot be treated as production')
    else:
        check(meta['object_freeze_sha256']!='CURRENT_REFERENCE_ONLY' and meta.get('object_freeze_manifest_path'),'production queue lacks object freeze')
        verified,h=gated_endpoints(meta['object_freeze_manifest_path'])
        check(h==meta['object_freeze_sha256'],'object freeze manifest changed')
    scope,_=load_scope(scope_file=queue_dir/'ELIGIBILITY_SNAPSHOT.json')
    endpoints=unique(read(queue_dir/'ENDPOINT_SNAPSHOT.tsv'),'canonical_id')
    if meta['mode']=='NOT_FOR_PRODUCTION':
        check(list(endpoints.values())==read(PKG/'REFERENCE_OBJECTS.tsv'),'dry-run endpoints differ from confirmed baseline')
    if meta['mode']!='NOT_FOR_PRODUCTION':
        check([{k:r[k] for k in REF_FIELDS} for r in verified]==list(endpoints.values()),'production endpoint snapshot not confirmed by object freeze')
    pairs=unique(read(queue_dir/'QUEUE.tsv'),'pair_id');crops=unique(read(queue_dir/'CROP_MANIFEST.tsv'),'pair_id')
    check(root.findtext('version')=='1.1' and not root.findall('track'),'expected image XML 1.1')
    seen=set();results=[]
    for image in root.findall('image'):
        pair_id=Path(image.get('name','')).stem
        check(pair_id in pairs and pair_id not in seen,'unknown/duplicate pair frame')
        pair=pairs[pair_id];crop=crops[pair_id]
        check(image.get('id')==str(crop['frame']) and image.get('width')==str(crop['width']) and image.get('height')==str(crop['height']),'pair frame/crop changed')
        origin=(int(crop['origin_x']),int(crop['origin_y']))
        check(pair['panel'] in {'f68r1','f68r2','f68r3'},'non-applicable panel in relation task')
        check(pair['panel']!='f68r3' or scope['panels']['f68r3']=='ELIGIBLE_LIMITED','f68r3 scope is not reviewer-confirmed')
        label,star=endpoints[pair['label_id']],endpoints[pair['star_id']]
        check(label['object_type']=='LABEL' and star['object_type']=='STAR' and label['panel']==star['panel']==pair['panel'],'pair endpoint mapping invalid')
        endpoints_in_scope=eligible(label,scope) and eligible(star,scope)
        refs={label['reference_id']:label,star['reference_id']:star};observed=set();tags=[]
        for element in image:
            a=attrs(element)
            if element.get('label') in {'STAR_REFERENCE','LABEL_REFERENCE'}:
                check(set(a)=={'reference_id','relation_protocol_version'} and a['relation_protocol_version']==RELATION_VERSION,'endpoint relation attrs/protocol changed')
                rid=a['reference_id'];check(rid in refs and rid not in observed,'endpoint deleted/replaced/duplicated')
                endpoint=refs[rid]
                check(element.get('label')==endpoint['object_type']+'_REFERENCE','endpoint class changed')
                expected_group=str(int(crop['frame'])+1) if meta.get('pair_shapes_grouped') else None
                check(element.get('group_id')==expected_group,'pair grouping changed/missing')
                same_geometry(parse_geometry(element,endpoint['object_type'],(100000,100000),True,origin),endpoint)
                observed.add(rid)
            else:
                check(element.tag=='tag' and element.get('label')=='PAIR_ATTACHMENT','unexpected relation geometry/class')
                check(set(a)=={'pair_id','relation_protocol_version','decision','human_confidence'},'legacy spatial fields or unknown attachment attrs')
                check(a['pair_id']==pair_id and a['relation_protocol_version']==RELATION_VERSION,'pair ID/protocol changed')
                check(a['decision'] in {'VISUAL_LABEL_OF','UNASSIGNED','UNCERTAIN','NOT_APPLICABLE_REGION'},'relation decision must be explicitly reviewed')
                check(a['human_confidence'] in {'LOW','MEDIUM','HIGH'},'relation confidence unset')
                tags.append(a)
        check(observed==set(refs) and len(tags)==1,'missing reference endpoint or pair decision tag')
        check(endpoints_in_scope or tags[0]['decision']=='NOT_APPLICABLE_REGION','out-of-mask endpoint needs an explicit region-exclusion flag, not ordinary attachment decision')
        results.append({'pair_id':pair_id,'panel':pair['panel'],'label_id':pair['label_id'],'star_id':pair['star_id'],
                        'relation_type':tags[0]['decision'],'human_confidence':tags[0]['human_confidence'],'reviewer_id':reviewer,
                        'review_timestamp':timestamp,'relation_protocol_version':RELATION_VERSION,'decision_origin':'EXPLICIT_HUMAN_ATTACHMENT_COMPLETENESS_REVIEW',
                        'queue_mode':meta['mode']})
        results[-1]['raw_export_sha256']=raw_sha
        results[-1]['queue_manifest_sha256']=digest(queue_dir/'QUEUE_MANIFEST.json')
        results[-1]['eligibility_reconciliation_status']='AWAITING_REVIEWER_REGION_RECONCILIATION_NOT_A_NEGATIVE' if tags[0]['decision']=='NOT_APPLICABLE_REGION' else 'IN_FROZEN_SCOPE'
        seen.add(pair_id)
    check(seen==set(pairs),'missing pair frames; absent relations never become negatives')
    return sorted(results,key=lambda r:r['pair_id'])

def main():
    p=argparse.ArgumentParser(description=__doc__);sub=p.add_subparsers(dest='command',required=True)
    gen=sub.add_parser('generate');gen.add_argument('--package',type=Path,default=PKG);gen.add_argument('--output-dir',type=Path,required=True)
    gen.add_argument('--dry-run',action='store_true');gen.add_argument('--object-freeze-manifest',type=Path);gen.add_argument('--scope-file',type=Path)
    gen.add_argument('--recheck-prior-v2',action='store_true',help='include frozen v2 pairs as blind v3 recheck frames; never rewrite v2')
    gen.add_argument('--group-pairs',action='store_true',help='give the two endpoint copies in each pair-frame the same CVAT group_id')
    gen.add_argument('--cover-new-endpoints',action='store_true',help='add one explicit nearest directional fallback for an otherwise unrepresented new eligible endpoint')
    gen.add_argument('--targeted-new-and-prior-only',action='store_true',help='keep exact frozen v2 pairs and nearest coverage for each new endpoint; exclude expanded old-old candidates')
    val=sub.add_parser('validate');val.add_argument('--input',type=Path,required=True);val.add_argument('--queue-dir',type=Path,required=True)
    val.add_argument('--reviewer-id',required=True);val.add_argument('--timestamp',required=True)
    args=p.parse_args()
    if args.command=='generate':
        verify_inputs(args.package);scope,_=load_scope(args.package,args.scope_file)
        check(not args.output_dir.exists(),'use a new queue directory')
        if args.dry_run:
            endpoints=read(args.package/'REFERENCE_OBJECTS.tsv');mode='NOT_FOR_PRODUCTION';h='CURRENT_REFERENCE_ONLY'
        else:
            check(args.object_freeze_manifest is not None,'production generation blocked until object completeness freeze')
            endpoints,h=gated_endpoints(args.object_freeze_manifest,args.package);mode='PRODUCTION_AFTER_OBJECT_FREEZE'
        pairs,_,_=generate(args.output_dir,endpoints,scope,args.package,mode,h,args.object_freeze_manifest,
                           args.recheck_prior_v2,args.group_pairs,args.cover_new_endpoints,args.targeted_new_and_prior_only)
        print(f'QUEUED_PAIRS={len(pairs)}; MODE={mode}')
    else:
        verify_inputs();rows=validate_relation(xml_input(args.input),args.queue_dir,args.reviewer_id,args.timestamp,digest(args.input))
        print(f'VALIDATION=PASS; EXPLICIT_DECISIONS={len(rows)}; NO_AUTOMATIC_NEGATIVES=YES')

if __name__=='__main__':main()
