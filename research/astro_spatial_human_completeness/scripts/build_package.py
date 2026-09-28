#!/usr/bin/env python3
"""Generate blind completeness preparation only; no human review/real export ingestion."""
from __future__ import annotations
import argparse
from collections import Counter
import hashlib
import json
from pathlib import Path
import xml.etree.ElementTree as ET
from PIL import Image,ImageDraw
from core import (ROOT,PKG,HUMAN,OBJECT_VERSION,RELATION_VERSION,REF_FIELDS,CHECKS,inventory,load_frozen,
                  read,write,jswrite,digest,check,safe_output,schema,relation_schema,root_xml,shape_element,
                  attr,render_xml,zip_bytes,envelope)
from attachment import generate

DOCS=['HUMAN_COMPLETENESS_GOALS.md','HUMAN_OBJECT_COMPLETENESS_PROTOCOL.md','HUMAN_OBJECT_COMPLETENESS_GUIDE.md',
      'HUMAN_ATTACHMENT_COMPLETENESS_PROTOCOL.md','HUMAN_ATTACHMENT_COMPLETENESS_GUIDE.md',
      'ATTACHMENT_QUEUE_GENERATION_SPEC.md','OBJECT_FREEZE_CONTRACT.md','F68R3_ATTACHMENT_ELIGIBILITY_DRAFT.md',
      'CVAT_SCHEMA.md','REPRODUCIBILITY.md']

def draw_map(out,panel,refs,tiles,whitelist=False):
    source=Image.open(HUMAN/'images'/f'{panel}.jpg').convert('RGB')
    scale=min(1.,1100/source.width,1400/source.height)
    image=source.resize((round(source.width*scale),round(source.height*scale)),Image.Resampling.LANCZOS)
    canvas=Image.new('RGB',(image.width,image.height+90),'white');canvas.paste(image,(0,90));draw=ImageDraw.Draw(canvas)
    draw.text((12,10),f'{panel}: '+('DRAFT NO APPROVED ELIGIBILITY; all confirmed references shown' if whitelist else 'navigation only; original panel coordinates unchanged'),fill='#222222')
    draw.text((12,30),'T01-T12 row-major STAR then LABEL. Grid is NOT an eligibility mask.',fill='#222222')
    draw.text((12,50),'Object protocol C1.0; no source/model/history; green STAR / blue LABEL.',fill='#222222')
    for tile in tiles:
        if tile['panel']!=panel:continue
        coords=[float(tile[k])*scale for k in ('x1','y1','x2','y2')]
        coords[1]+=90;coords[3]+=90
        draw.rectangle(coords,outline='#dba024',width=2);draw.text((coords[0]+8,coords[1]+8),tile['tile_id'],fill='#835b14')
    catalog=[];counters=Counter()
    for row in refs:
        if row['panel']!=panel:continue
        counters[row['object_type']]+=1
        neutral=('S' if row['object_type']=='STAR' else 'L')+f'{counters[row["object_type"]]:03d}'
        color='#2aa66a' if row['object_type']=='STAR' else '#3578bb'
        x1,y1,x2,y2=[float(row[k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')]
        if row['geometry_type']=='ELLIPSE':
            draw.ellipse((x1*scale,y1*scale+90,x2*scale,y2*scale+90),outline=color,width=2)
        else:
            cx,cy=(x1+x2)/2,(y1+y2)/2
            import math
            c,s=math.cos(math.radians(float(row['rotation']))),math.sin(math.radians(float(row['rotation'])))
            pts=[((cx+x*c-y*s)*scale,(cy+x*s+y*c)*scale+90) for x,y in [(-(x2-x1)/2,-(y2-y1)/2),((x2-x1)/2,-(y2-y1)/2),((x2-x1)/2,(y2-y1)/2),(-(x2-x1)/2,(y2-y1)/2)]]
            draw.line(pts+[pts[0]],fill=color,width=2)
        if whitelist:
            draw.text((x1*scale,y1*scale+90),neutral,fill=color,stroke_width=1,stroke_fill='#ffffff')
            catalog.append({'neutral_map_id':neutral,'canonical_id':row['canonical_id'],'reference_id':row['reference_id'],
                            'object_type':row['object_type'],'panel':panel,'geometry_type':row['geometry_type'],
                            'bbox':','.join(row[k] for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')),
                            'rotation':row['rotation'],'eligibility_status':'NOT_SELECTED_AWAITING_REVIEWER_CONFIRMATION'})
    canvas.save(out,format='PNG',optimize=False)
    return catalog

def examples(out):
    from html import escape
    import math
    source=[];parts=['<svg xmlns="http://www.w3.org/2000/svg" width="1200" height="330" viewBox="0 0 1200 330"><rect width="1200" height="330" fill="white"/>']
    def text(x,y,t):parts.append(f'<text x="{x}" y="{y}" font-family="sans-serif" font-size="14">{escape(t)}</text>')
    scenarios=[('MISSING_OBJECT','Добавить отсутствующий','PROPOSED_NEW'),('REFERENCE_BOUNDARY','Границу reference НЕ менять','NO_REFERENCE_EDIT'),
               ('POSSIBLE_DUPLICATE','Возможный дубликат: только флаг','POSSIBLE_DUPLICATE_NEW'),('DOUBTFUL_NEW','Сомнительный новый объект','UNCERTAIN_NEW')]
    for i,(sid,title,state) in enumerate(scenarios):
        x=i*300
        text(x+12,30,title);text(x+12,280,state)
        star=[]
        for j in range(16):
            angle=j*math.pi/8;r=45 if j%2==0 else 18
            star.append((x+145+r*math.cos(angle),145+r*math.sin(angle)))
        points=' '.join(f'{px:.2f},{py:.2f}' for px,py in star)
        parts.append(f'<polygon points="{points}" fill="{("#aaa" if i==3 else "#777")}"/>')
        if i in {1,2}:parts.append(f'<rect x="{x+95}" y="95" width="100" height="100" fill="none" stroke="#29985f" stroke-width="3"/>')
        if i in {0,2,3}:parts.append(f'<rect x="{x+85}" y="85" width="120" height="120" fill="none" stroke="#df8a37" stroke-dasharray="8 5" stroke-width="3"/>')
        source.append({'scenario_id':sid,'addition_policy':state,'example_type':'SCHEMATIC_NOT_MANUSCRIPT_ANNOTATION','object_protocol_version':OBJECT_VERSION})
    text(12,315,'Зелёный = frozen reference; пунктир = новое proposal, не принятое автоматически. Без транскрипции. C1.0.')
    parts.append('</svg>');(out/'OBJECT_REVIEW_EXAMPLES.svg').write_text('\n'.join(parts)+'\n',encoding='utf-8')
    write(out/'OBJECT_REVIEW_EXAMPLES_SOURCE.tsv',source)

def build(out=PKG):
    out=safe_output(out);out.mkdir(parents=True,exist_ok=True)
    print('Verifying frozen upstream and comparison inputs...',flush=True)
    inv=inventory()
    if (PKG/'INPUT_MANIFEST.tsv').exists():
        current=read(PKG/'INPUT_MANIFEST.tsv')
        rendered=[{k:str(v) for k,v in row.items()} for row in inv]
        check(current==rendered,'prepared input snapshot changed; create a new revision')
    write(out/'INPUT_MANIFEST.tsv',inv)
    references,prior,dimensions=load_frozen()
    write(out/'REFERENCE_OBJECTS.tsv',references,REF_FIELDS)
    write(out/'PRIOR_NONCONFIRMED_PROVENANCE.tsv',prior,REF_FIELDS+['prior_status','source_annotation_ids','geometry_origin'])
    for name in DOCS:
        if out!=PKG:(out/name).write_bytes((PKG/name).read_bytes())
    jswrite(out/'CVAT_OBJECT_COMPLETENESS_SCHEMA.json',schema())
    jswrite(out/'CVAT_ATTACHMENT_COMPLETENESS_SCHEMA.json',relation_schema())
    scope={'relation_protocol_version':RELATION_VERSION,'panels':{panel:'ELIGIBLE' if panel in {'f68r1','f68r2'} else 'AWAITING_REVIEWER_CONFIRMATION' if panel=='f68r3' else 'PANEL_NOT_APPLICABLE_FOR_LABEL_STAR_ATTACHMENT' for panel in sorted(dimensions)},
           'f68r3_rule':None,'basis':'task-specific applicability; f68r1/r2 prior approved right-band screening only; no new automatic outcomes'}
    jswrite(out/'ATTACHMENT_ELIGIBILITY.json',scope)
    jswrite(out/'F68R3_SCOPE_TEMPLATE.json',{'status':'AWAITING_REVIEWER_CONFIRMATION','reviewer_id':'','confirmation_timestamp':'',
                                          'kind':'CHOOSE_WHITELIST_OR_POLYGON_CENTER_MASK','star_ids':[],'label_ids':[],'polygon':[],'rule_sha256':''})
    applicability=[{'panel':panel,'object_completeness':'YES','attachment_completeness':'YES' if panel in {'f68r1','f68r2'} else 'LIMITED_AWAITING_REVIEWER_CONFIRMATION' if panel=='f68r3' else 'PANEL_NOT_APPLICABLE_FOR_LABEL_STAR_ATTACHMENT',
                    'object_protocol_version':OBJECT_VERSION,'attachment_protocol_version':RELATION_VERSION} for panel in sorted(dimensions)]
    write(out/'PANEL_APPLICABILITY.tsv',applicability)
    root=root_xml('human-object-completeness-C1.0-NOT_STARTED',schema(),len(dimensions))
    manifest=[];tiles=[];image_entries={}
    for frame,(panel,(w,h)) in enumerate(sorted(dimensions.items())):
        image=ET.SubElement(root,'image',id=str(frame),name=panel+'.jpg',width=str(w),height=str(h))
        visible=[r for r in references+prior if r['panel']==panel and (r in references or r.get('prior_status')=='UNCERTAIN')]
        for row in sorted(visible,key=lambda r:r['reference_id']):shape_element(image,row)
        tag=ET.SubElement(image,'tag',label='PANEL_COMPLETENESS',source='manual')
        attr(tag,'completion_status','NOT_STARTED');attr(tag,'object_protocol_version',OBJECT_VERSION)
        for k in CHECKS:attr(tag,k,'false')
        path=HUMAN/'images'/f'{panel}.jpg';image_entries[panel+'.jpg']=path.read_bytes()
        manifest.append({'panel':panel,'frame':frame,'filename':panel+'.jpg','width':w,'height':h,'image_sha256':digest(path),
                         'confirmed_references':sum(r['panel']==panel for r in references),
                         'optional_uncertain_references':sum(r['panel']==panel and r['prior_status']=='UNCERTAIN' for r in prior),
                         'completion_marker':'PANEL_COMPLETENESS','initial_status':'NOT_STARTED','object_protocol_version':OBJECT_VERSION})
        for row in range(4):
            for col in range(3):
                number=row*3+col+1
                tiles.append({'panel':panel,'tile_id':f'T{number:02d}','order':number,'x1':col*w//3,'y1':row*h//4,'x2':(col+1)*w//3,'y2':(row+1)*h//4,
                              'review_order':'STAR_THEN_LABEL','purpose':'NAVIGATION_ONLY_NO_COORDINATE_TRANSFORM'})
    write(out/'OBJECT_COMPLETENESS_QUEUE_MANIFEST.tsv',manifest)
    write(out/'NAVIGATION_TILES.tsv',tiles)
    xml=render_xml(root);(out/'OBJECT_COMPLETENESS_CVAT.xml').write_bytes(xml)
    zip_bytes(out/'OBJECT_COMPLETENESS_CVAT.zip',{'annotations.xml':xml});zip_bytes(out/'OBJECT_COMPLETENESS_IMAGES.zip',image_entries)
    print('Drawing navigation, eligibility draft and example figures...',flush=True)
    (out/'navigation').mkdir(exist_ok=True);(out/'figures').mkdir(exist_ok=True)
    for panel in sorted(dimensions):draw_map(out/'navigation'/f'{panel}.png',panel,references,tiles)
    catalog=draw_map(out/'figures/F68R3_ELIGIBILITY_MAP.png','f68r3',references,tiles,True)
    write(out/'F68R3_WHITELIST_CATALOG.tsv',catalog)
    examples(out/'figures')
    print('Creating attachment current-reference dry run NOT_FOR_PRODUCTION...',flush=True)
    pairs,filtered,retained=generate(out/'dry_run/attachment_f68r1_f68r2_NOT_FOR_PRODUCTION',references,scope,out)
    # Explicit synthetic copied-tree data only, never treated as real reviewer results.
    synthetic=ET.fromstring(xml)
    for tag in synthetic.findall('.//tag'):
        for a in tag.findall('attribute'):
            if a.get('name')=='completion_status':a.text='COMPLETE_NO_NEW_OBJECTS'
            elif a.get('name') in CHECKS:a.text='true'
    (out/'dry_run/OBJECT_ROUNDTRIP_SYNTHETIC_NOT_FOR_PRODUCTION.xml').write_bytes(render_xml(synthetic))
    jswrite(out/'dry_run/DRY_RUN_WARNING.json',{'mode':'SYNTHETIC_NOT_FOR_PRODUCTION','human_review_started':False,
                                            'object_markers':'copied-tree simulation, not reviewer assertions','new_objects_accepted':0})
    report=f'''# Human-assisted completeness preparation

Object package C1.0 ready for a future authorized reviewer pass, not started here.
Eight original canonical panels, {len(references)} final human references (321 STAR/199 LABEL),
8 optional unresolved reference shapes and 222 preserved nonconfirmed provenance records
(206 REJECT, 8 UNCERTAIN, 8 NOT_REVIEWED). AI identities/history are excluded from CVAT.
All references use exact final human geometry, not source/candidate display boxes.

Separate object annotation XML/ZIP and image ZIP, navigation order/maps, neutral schemas,
strict reference-preserving validator and proposed-addition staging are prepared. New
objects receive deterministic canonical IDs at future ingestion, but are NOT_ACCEPTED.
No real human export imported, no annotations created by a reviewer, no new AI metrics.
Completion markers start NOT_STARTED; empty export never asserts no missed objects.

Attachment protocol 3 is a future gated workflow, distinct from frozen spatial 1/caption 2.
Dry run only f68r1/f68r2: {len(pairs)} queued UNREVIEWED pairs, {len(filtered)} geometric filters
(not observations), {len(retained)} previous v2 pairs retained without re-review/conversion.
No non-applicable-panel Cartesian product. f68r3 limited scope cannot be reconstructed
from old whole-source-pair policy: map/catalog/options prepared, reviewer confirmation required.
Older f68v2/v2 decisions stay unchanged despite new NOT_APPLICABLE policy.

CVAT instance URL/version is unavailable (no local CVAT container/connector). Documented
UI lock/hide is optional convenience; export validator is authoritative protection.
CVAT_ROUNDTRIP_TEST denotes tested native XML copied-tree/two-decimal serialization,
not a claimed live-server test. Empty live-instance import/export preflight is required
before human review; see CVAT_SCHEMA.md and guides. No invented persistent lock field.

Validation and separate-directory byte reproduction: VALIDATION_REPORT.md. Input hashes
registered in INPUT_MANIFEST.tsv and rechecked after generation. File checksums include
only this new preparation package, never refreeze original inputs. All outputs are separate.

```text
HUMAN_OBJECT_COMPLETENESS_PACKAGE=READY
HUMAN_REVIEW_STARTED=NO
FROZEN_INPUTS_UNCHANGED=YES
REFERENCE_OBJECTS_PROTECTED=YES
AI_SOURCE_IDENTITY_HIDDEN=YES
ALL_PANELS_OBJECT_SCOPE=YES
ATTACHMENT_SCOPE_F68R1=READY
ATTACHMENT_SCOPE_F68R2=READY
ATTACHMENT_SCOPE_F68R3=AWAITING_REVIEWER_CONFIRMATION
NON_APPLICABLE_PANELS_EXCLUDED=YES
PRODUCTION_ATTACHMENT_PACKAGE_CREATED=NO
CVAT_ROUNDTRIP_TEST=PASS
RESULTS_REPRODUCIBLE=YES
```
'''
    (out/'PREPARATION_REPORT.md').write_text(report,encoding='utf-8')
    jswrite(out/'PREPARATION_MANIFEST.json',{'object_protocol_version':OBJECT_VERSION,'attachment_protocol_version':RELATION_VERSION,
                                          'human_review_started':False,'production_attachment_package_created':False,
                                          'reference_objects':len(references),'nonconfirmed_index':len(prior),'panels':len(dimensions),
                                          'f68r3_scope':'AWAITING_REVIEWER_CONFIRMATION','reference_sha256':digest(out/'REFERENCE_OBJECTS.tsv'),
                                          'input_manifest_sha256':digest(out/'INPUT_MANIFEST.tsv')})
    print(f'PREPARED; REFERENCES={len(references)}; DRY_RUN_PAIRS={len(pairs)}; HUMAN_REVIEW_STARTED=NO',flush=True)

def main():
    p=argparse.ArgumentParser();p.add_argument('--output-dir',type=Path,default=PKG);args=p.parse_args()
    try:build(args.output_dir)
    except (ValueError,KeyError,FileNotFoundError) as exc:
        out=safe_output(args.output_dir);out.mkdir(parents=True,exist_ok=True)
        (out/'BLOCKER_REPORT.md').write_text('# Preparation blocker\n\n'+str(exc)+'\n\nHUMAN_OBJECT_COMPLETENESS_PACKAGE=BLOCKED\nHUMAN_REVIEW_STARTED=NO\nNo frozen input was fixed or overwritten.\n',encoding='utf-8')
        raise

if __name__=='__main__':main()
