#!/usr/bin/env python3
"""Validate future CVAT export; stage proposed additions, never accept/freeze them."""
from __future__ import annotations
import argparse
from collections import Counter
from pathlib import Path
from core import (PKG, OBJECT_VERSION, REF_FIELDS, NEW_FIELDS, NEW_STATUSES, CHECKS, COMPLETION,
                  read, write, check, unique, verify_inputs, xml_input, attrs, parse_geometry,
                  same_geometry, stable_id, overlaps, utc, digest, safe_output)

def validate(root,package=PKG,reviewer='DRY_RUN',timestamp='2000-01-01T00:00:00Z',raw_sha='SYNTHETIC',allow_incomplete=False):
    utc(timestamp)
    check(reviewer and '\t' not in reviewer and '\n' not in reviewer,'invalid pseudonymous reviewer')
    check(root.tag=='annotations' and root.findtext('version')=='1.1','expected CVAT for images 1.1')
    check(not root.findall('track'),'tracks forbidden in image completeness task')
    manifest=read(package/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv')
    panels=unique(manifest,'panel')
    images=root.findall('image')
    check(len(images)==len(panels),'missing/extra panel images; export complete task, not one partial job')
    expected_refs=read(package/'REFERENCE_OBJECTS.tsv')+[r for r in read(package/'PRIOR_NONCONFIRMED_PROVENANCE.tsv') if r['prior_status']=='UNCERTAIN']
    ref_idx=unique(expected_refs,'reference_id')
    observed=set();seen_panels=set();additions=[];completion=[]
    allowed={'STAR_REFERENCE','LABEL_REFERENCE','UNCERTAIN_REFERENCE','STAR_HUMAN_ADDED','LABEL_HUMAN_ADDED','PANEL_COMPLETENESS'}
    meta_labels=root.findall('./meta/task/labels/label')
    if meta_labels:
        check({l.findtext('name') for l in meta_labels}==allowed,'CVAT metadata has forbidden/missing classes')
    for image in images:
        panel=Path(image.attrib['name']).stem
        check(panel in panels and panel not in seen_panels,'unknown/duplicate panel')
        seen_panels.add(panel)
        info=panels[panel];dimensions=(int(info['width']),int(info['height']))
        check(image.get('id')==info['frame'] and int(image.get('width','0'))==dimensions[0] and int(image.get('height','0'))==dimensions[1],'panel/frame/image geometry mismatch')
        markers=[];panel_new=[]
        for element in image:
            label=element.get('label')
            check(label in allowed and element.tag in {'box','ellipse','tag'},'forbidden class/shape element')
            a=attrs(element)
            if label in {'STAR_REFERENCE','LABEL_REFERENCE','UNCERTAIN_REFERENCE'}:
                check(set(a)=={'reference_id','object_protocol_version'},'reference attributes changed/added/removed')
                rid=a['reference_id']
                check(rid in ref_idx and rid not in observed,'reference ID changed/duplicated')
                expected=ref_idx[rid]
                check(expected['panel']==panel and label==expected['layer'],'reference panel/class changed')
                check(a['object_protocol_version']==OBJECT_VERSION,'reference protocol changed')
                # A real CVAT image-task import/export changes imported shapes from
                # source="manual" to source="file". This is serialization provenance,
                # not a human edit; IDs, classes, attributes and geometry remain strict.
                check(element.get('source','manual') in {'manual','file'} and element.get('occluded','0')=='0' and element.get('z_order','0')=='0' and element.get('group_id','0')=='0','reference shape properties changed')
                same_geometry(parse_geometry(element,expected['object_type'],dimensions,True),expected)
                observed.add(rid)
            elif label=='PANEL_COMPLETENESS':
                check(element.tag=='tag','panel completion must be a tag')
                check(set(a)=={'completion_status','object_protocol_version',*CHECKS},'invalid completion attributes')
                check(a['object_protocol_version']==OBJECT_VERSION and a['completion_status'] in COMPLETION,'invalid completion protocol/status')
                check(all(a[k] in {'true','false'} for k in CHECKS),'invalid completion checkbox')
                markers.append(a)
            else:
                kind='STAR' if label=='STAR_HUMAN_ADDED' else 'LABEL'
                check(element.get('source','manual')=='manual','new completeness additions must be human/manual, not auto predictions')
                check(set(a)=={'addition_status','human_confidence','object_protocol_version'},'invalid addition attributes')
                check(a['object_protocol_version']==OBJECT_VERSION and a['addition_status'] in NEW_STATUSES,'invalid addition status/protocol')
                check(a['human_confidence'] in {'LOW','MEDIUM','HIGH'},'new human confidence must be explicitly set')
                geometry=parse_geometry(element,kind,dimensions)
                row={'panel':panel,'frame':image.attrib['id'],'object_type':kind,**geometry}
                row={'canonical_id':stable_id(row),**row,'reviewer_id':reviewer,'review_timestamp':timestamp,
                     'human_confidence':a['human_confidence'],'addition_status':a['addition_status'],
                     'completeness_protocol_version':OBJECT_VERSION,'creation_origin':'HUMAN_COMPLETENESS_PASS',
                     'overlap_status':'PENDING_RECONCILIATION_SCAN','acceptance_status':'NOT_ACCEPTED_REQUIRES_POST_REVIEW',
                     'raw_export_sha256':raw_sha}
                panel_new.append(row)
        check(len(markers)==1,'every frame needs exactly one panel completion marker')
        marker=markers[0];state=marker['completion_status']
        if state.startswith('COMPLETE_'):
            check(all(marker[k]=='true' for k in CHECKS),'complete marker without all systematic review checks')
            check((state=='COMPLETE_WITH_NEW_OBJECTS')==bool(panel_new),'completion marker contradicts additions count')
        else:check(allow_incomplete,'panel not completed; no completeness/no-missing assertion permitted')
        completion.append({'panel':panel,'frame':image.attrib['id'],'completion_status':state,'new_proposals':len(panel_new),
                           **{k:marker[k] for k in CHECKS},'reviewer_id':reviewer,'review_timestamp':timestamp,
                           'completeness_protocol_version':OBJECT_VERSION,'interpretation':'HUMAN_ASSISTED_REVIEW_ONLY_NOT_ABSOLUTE_COMPLETENESS'})
        additions.extend(panel_new)
    check(observed==set(ref_idx),'reference objects deleted/missing')
    unique(additions,'canonical_id')  # exact-geometry duplicate IDs are blocked, not silently disambiguated by CVAT IDs
    index=read(package/'REFERENCE_OBJECTS.tsv')+read(package/'PRIOR_NONCONFIRMED_PROVENANCE.tsv')
    reconciliation=[]
    for new in additions:
        matches=overlaps(new,index)
        # Proposals can also duplicate each other; retain both distinct geometries for explicit review.
        for other in additions:
            if other['canonical_id']==new['canonical_id']:continue
            for match in overlaps(new,[other]):
                match['status']='POSSIBLE_DUPLICATE_NEW';matches.append(match)
        if new['addition_status']=='POSSIBLE_DUPLICATE_NEW' and not matches:
            matches.append({'canonical_new_id':new['canonical_id'],'prior_canonical_id':'','panel':new['panel'],
                            'status':'REVIEWER_POSSIBLE_DUPLICATE_FLAG','envelope_iou':'','containment':'','center_distance_px':'',
                            'resolution':'AWAITING_EXPLICIT_POST_REVIEW_RECONCILIATION'})
        reconciliation.extend(matches)
        new['overlap_status']=';'.join(sorted({m['status'] for m in matches})) or 'NO_FLAGGED_OVERLAP_NOT_ACCEPTANCE'
    return sorted(additions,key=lambda r:r['canonical_id']),completion,reconciliation

def main():
    p=argparse.ArgumentParser(description=__doc__)
    p.add_argument('--input',type=Path,required=True);p.add_argument('--package',type=Path,default=PKG)
    p.add_argument('--reviewer-id',required=True);p.add_argument('--timestamp',required=True)
    p.add_argument('--output-dir',type=Path);p.add_argument('--allow-incomplete',action='store_true')
    args=p.parse_args()
    verify_inputs(args.package)
    additions,completion,queue=validate(xml_input(args.input),args.package,args.reviewer_id,args.timestamp,digest(args.input),args.allow_incomplete)
    if args.output_dir:
        out=safe_output(args.output_dir);check(not out.exists(),'use a new staging directory; never overwrite staged review')
        out.mkdir(parents=True)
        write(out/'HUMAN_ADDITIONS_PROPOSED.tsv',additions,NEW_FIELDS)
        write(out/'PANEL_COMPLETION.tsv',completion)
        write(out/'POST_REVIEW_RECONCILIATION_QUEUE.tsv',queue,['canonical_new_id','prior_canonical_id','panel','status','envelope_iou','containment','center_distance_px','resolution'])
    print(f'VALIDATION=PASS; NEW_PROPOSALS={len(additions)}; RECONCILIATION_FLAGS={len(queue)}; AUTO_ACCEPTED=0')

if __name__=='__main__':main()
