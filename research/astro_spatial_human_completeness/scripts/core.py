"""Human-assisted completeness C1.0 / caption attachment 3 preparation helpers."""
from __future__ import annotations
import csv
from datetime import datetime
import hashlib
import importlib.util
import io
import json
import math
from pathlib import Path
import sys
import xml.etree.ElementTree as ET
import zipfile

sys.dont_write_bytecode = True
ROOT=Path(__file__).resolve().parents[3]
PKG=Path(__file__).resolve().parents[1]
HUMAN=ROOT/'research/astro_spatial_human_adjudication'
COMPARISON=ROOT/'research/astro_spatial_three_way_comparison'
OBJECT_VERSION='C1.0'
RELATION_VERSION='3'
TOLERANCE=.0051  # actual CVAT image XML rounds to two decimals; 0.001 would reject valid exports
COMPLETION=['NOT_STARTED','COMPLETE_NO_NEW_OBJECTS','COMPLETE_WITH_NEW_OBJECTS','INCOMPLETE_TECHNICAL','INCOMPLETE_REVIEWER_UNCERTAINTY']
NEW_STATUSES=['PROPOSED_NEW','UNCERTAIN_NEW','POSSIBLE_DUPLICATE_NEW']
REL_DECISIONS=['UNREVIEWED','VISUAL_LABEL_OF','UNASSIGNED','UNCERTAIN','NOT_APPLICABLE_REGION']
CHECKS=['star_review_done','label_review_done','edges_checked','circular_regions_checked','weak_contrast_checked']
REF_FIELDS=['reference_id','canonical_id','panel','object_type','geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation','layer']
NEW_FIELDS=['canonical_id','panel','frame','object_type','geometry_type','bbox_x1','bbox_y1','bbox_x2','bbox_y2','rotation',
            'reviewer_id','review_timestamp','human_confidence','addition_status','completeness_protocol_version','creation_origin',
            'overlap_status','acceptance_status','raw_export_sha256']

def check(ok,message):
    if not ok: raise ValueError(message)

def digest(path):
    h=hashlib.sha256()
    with Path(path).open('rb') as f:
        for chunk in iter(lambda:f.read(1024*1024),b''):h.update(chunk)
    return h.hexdigest()

def read(path):
    with Path(path).open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))

def write(path,rows,fields=None):
    path=Path(path)
    path.parent.mkdir(parents=True,exist_ok=True)
    with path.open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fields or list(rows[0]),delimiter='\t',lineterminator='\n')
        w.writeheader()
        w.writerows(rows)

def jswrite(path,data):Path(path).write_text(json.dumps(data,sort_keys=True,indent=2,allow_nan=False)+'\n',encoding='utf-8')

def unique(rows,key):
    result={r[key]:r for r in rows}
    check(len(result)==len(rows),f'duplicate {key}')
    return result

def utc(value):
    check(value.endswith('Z'),'UTC timestamp must end in Z')
    parsed=datetime.fromisoformat(value[:-1]+'+00:00')
    check(parsed.utcoffset().total_seconds()==0,'invalid UTC timestamp')

def module(name,path):
    spec=importlib.util.spec_from_file_location(name,path)
    obj=importlib.util.module_from_spec(spec)
    spec.loader.exec_module(obj)
    return obj

def inventory():
    """Reuse verified actual upstream inventory; verify comparison freeze as well."""
    common=module('frozen_comparison_common',COMPARISON/'scripts/common.py')
    base,_=common.verify_inventory()
    registered=common.read(COMPARISON/'INPUT_MANIFEST.tsv')
    for row in registered:row['bytes']=int(row['bytes'])
    check(base==registered,'comparison input snapshot no longer agrees with upstream')
    paths={ROOT/r['path']:(r['version'],r['role']) for r in base}
    ledger=COMPARISON/'SHA256SUMS'
    for line in ledger.read_text().splitlines():
        expected,rel=line.split('  ',1)
        path=(COMPARISON/rel).resolve()
        check(path.is_relative_to(COMPARISON), 'unsafe comparison checksum path')
        check(digest(path)==expected,f'comparison frozen checksum mismatch: {rel}')
        paths[path]=('three-way-1.0','FROZEN_THREE_WAY_RESULT')
    paths[ledger]=('three-way-1.0','CHECKSUM_LEDGER')
    result=[]
    for path,(version,role) in sorted(paths.items(),key=lambda item:str(item[0])):
        count=len(read(path)) if path.suffix=='.tsv' else ''
        result.append({'logical_role':role,'path':str(path.relative_to(ROOT)),'protocol_version':version,'row_count':count,
                       'file_size':path.stat().st_size,'sha256':digest(path),'frozen_status':'FROZEN'})
    return result

def verify_inputs(package=PKG,full=False):
    registered=read(package/'INPUT_MANIFEST.tsv')
    for row in registered:
        path=ROOT/row['path']
        check(path.is_file() and digest(path)==row['sha256'],f'frozen input changed: {row["path"]}')
    references,prior,dimensions=load_frozen()
    check(read(package/'REFERENCE_OBJECTS.tsv')==references,'prepared references differ from actual frozen final human source')
    check(read(package/'PRIOR_NONCONFIRMED_PROVENANCE.tsv')==prior,'prior provenance index differs from frozen source')
    prepared=json.loads((package/'PREPARATION_MANIFEST.json').read_text())
    check(prepared['reference_sha256']==digest(package/'REFERENCE_OBJECTS.tsv') and prepared['input_manifest_sha256']==digest(package/'INPUT_MANIFEST.tsv'),'prepared baseline manifest changed')
    if full:
        actual=inventory()
        for row in actual:
            row['row_count']=str(row['row_count']);row['file_size']=str(row['file_size'])
        check(actual==registered,'inventory path set changed')

def load_frozen():
    # Reuse input *validation*, not comparison aggregation or annotation inference.
    sys.path.insert(0,str(COMPARISON/'scripts'))
    analysis=module('read_only_frozen_validation',COMPARISON/'scripts/analyze.py')
    inv=read(COMPARISON/'INPUT_MANIFEST.tsv')
    for r in inv:r['bytes']=int(r['bytes'])
    sources,candidates,human,memberships,dimensions,mappings,audit,phases=analysis.load_inputs(inv)
    references=[]
    prior=[]
    for kind in ('STAR','LABEL'):
        for c in candidates[kind]:
            cid=c['candidate_id']
            h=human[kind].get(cid)
            decision=h['human_decision'] if h else 'NOT_REVIEWED'
            row=h or c
            gtype='BOX' if kind=='STAR' else h['geometry_type'] if h else 'ELLIPSE' if c['geometry_mode']=='ELLIPSE_RING' else 'BOX'
            rotation='0' if kind=='STAR' else h['rotation'] if h else c['provisional_rotation']
            base={'reference_id':'REF_'+hashlib.sha256(cid.encode()).hexdigest()[:16].upper(),'canonical_id':cid,'panel':c['panel'],
                  'object_type':kind,'geometry_type':gtype,**{k:row[k] for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')},
                  'rotation':rotation,'layer':kind+'_REFERENCE' if decision in {'ACCEPT','MODIFY'} else 'UNCERTAIN_REFERENCE'}
            if decision in {'ACCEPT','MODIFY'}: references.append(base)
            else:prior.append(base|{'prior_status':decision,'source_annotation_ids':c['source_annotation_ids'],
                                    'geometry_origin':'FINAL_HUMAN_RECORD' if h else 'UNREVIEWED_CANDIDATE_DISPLAY'})
    references.sort(key=lambda r:(r['panel'],r['canonical_id']))
    prior.sort(key=lambda r:(r['panel'],r['canonical_id']))
    check(len(references)==520 and len(prior)==222,'reference/prior counts mismatch')
    check(len({r['reference_id'] for r in references+prior})==742,'blind reference ID collision')
    return references,prior,dimensions

def attrs(element):
    pairs=[(a.attrib['name'],a.text or '') for a in element.findall('attribute')]
    check(len(dict(pairs))==len(pairs),'duplicate shape/tag attribute')
    return dict(pairs)

def attr(element,name,value):ET.SubElement(element,'attribute',name=name).text=str(value)

def render_xml(root):
    ET.indent(root,space='  ')
    return ET.tostring(root,encoding='utf-8',xml_declaration=True)+b'\n'

def zip_bytes(path,entries):
    with zipfile.ZipFile(path,'w',compression=zipfile.ZIP_DEFLATED) as z:
        for name,data in sorted(entries.items()):
            info=zipfile.ZipInfo(name,date_time=(1980,1,1,0,0,0));info.compress_type=zipfile.ZIP_DEFLATED
            info.external_attr=0o644<<16
            z.writestr(info,data)

def xml_input(path):
    path=Path(path)
    if path.suffix.lower()=='.zip':
        with zipfile.ZipFile(path) as z:
            files=[n for n in z.namelist() if n.endswith('.xml')]
            check(len(files)==1,'export ZIP must contain exactly one XML')
            data=z.read(files[0])
    else:data=path.read_bytes()
    check(b'<!DOCTYPE' not in data.upper() and b'<!ENTITY' not in data.upper(),'DTD/entities prohibited')
    return ET.fromstring(data)

def schema():
    def field(name,kind='text',default='UNSET',values=None,mutable=True):
        return {'name':name,'input_type':kind,'default_value':default,'values':values or [default],'mutable':mutable}
    ref=[field('reference_id',mutable=False),field('object_protocol_version',default=OBJECT_VERSION,mutable=False)]
    new=[field('addition_status','select','PROPOSED_NEW',NEW_STATUSES),field('human_confidence','select','UNSET',['UNSET','LOW','MEDIUM','HIGH']),
         field('object_protocol_version',default=OBJECT_VERSION,mutable=False)]
    marker=[field('completion_status','select','NOT_STARTED',COMPLETION),field('object_protocol_version',default=OBJECT_VERSION,mutable=False)]
    marker += [field(n,'checkbox','false',['false']) for n in CHECKS]
    return [{'name':'STAR_REFERENCE','type':'rectangle','color':'#349866','attributes':ref},
            {'name':'LABEL_REFERENCE','type':'any','color':'#3578b8','attributes':ref},
            {'name':'UNCERTAIN_REFERENCE','type':'any','color':'#888888','attributes':ref},
            {'name':'STAR_HUMAN_ADDED','type':'rectangle','color':'#ed7532','attributes':new},
            {'name':'LABEL_HUMAN_ADDED','type':'any','color':'#b25dda','attributes':new},
            {'name':'PANEL_COMPLETENESS','type':'tag','color':'#555555','attributes':marker}]

def relation_schema():
    ref=[{'name':'reference_id','input_type':'text','default_value':'UNSET','values':['UNSET'],'mutable':False},
         {'name':'relation_protocol_version','input_type':'text','default_value':RELATION_VERSION,'values':[RELATION_VERSION],'mutable':False}]
    pair=[{'name':'pair_id','input_type':'text','default_value':'UNSET','values':['UNSET'],'mutable':False},
          {'name':'relation_protocol_version','input_type':'text','default_value':RELATION_VERSION,'values':[RELATION_VERSION],'mutable':False},
          {'name':'decision','input_type':'select','default_value':'UNREVIEWED','values':REL_DECISIONS,'mutable':True},
          {'name':'human_confidence','input_type':'select','default_value':'UNSET','values':['UNSET','LOW','MEDIUM','HIGH'],'mutable':True}]
    return [{'name':'STAR_REFERENCE','type':'rectangle','color':'#349866','attributes':ref},
            {'name':'LABEL_REFERENCE','type':'any','color':'#3578b8','attributes':ref},
            {'name':'PAIR_ATTACHMENT','type':'tag','color':'#555555','attributes':pair}]

def root_xml(name,labels,size):
    root=ET.Element('annotations');ET.SubElement(root,'version').text='1.1'
    task=ET.SubElement(ET.SubElement(root,'meta'),'task')
    for key,value in [('name',name),('size',str(size)),('mode','annotation'),('overlap','0'),('flipped','False')]:ET.SubElement(task,key).text=value
    labs=ET.SubElement(task,'labels')
    for s in labels:
        label=ET.SubElement(labs,'label')
        for key in ('name','color','type'):ET.SubElement(label,key).text=s[key]
        attributes=ET.SubElement(label,'attributes')
        for a in s['attributes']:
            entry=ET.SubElement(attributes,'attribute')
            for k in ('name','input_type','default_value','mutable','values'):
                value='\n'.join(a[k]) if k=='values' else str(a[k]).lower() if k=='mutable' else a[k]
                ET.SubElement(entry,k).text=value
    return root

def shape_element(image,row,label=None,origin=(0,0),version_field='object_protocol_version',version=OBJECT_VERSION):
    ox,oy=origin
    x1,y1,x2,y2=[float(row[k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')]
    common={'label':label or row['layer'],'source':'manual','occluded':'0','z_order':'0','rotation':row['rotation']}
    if row['geometry_type']=='ELLIPSE':
        element=ET.SubElement(image,'ellipse',common|{'cx':f'{(x1+x2)/2-ox:.6f}','cy':f'{(y1+y2)/2-oy:.6f}',
                               'rx':f'{(x2-x1)/2:.6f}','ry':f'{(y2-y1)/2:.6f}'})
    else:
        element=ET.SubElement(image,'box',common|{'xtl':f'{x1-ox:.6f}','ytl':f'{y1-oy:.6f}',
                         'xbr':f'{x2-ox:.6f}','ybr':f'{y2-oy:.6f}'})
    attr(element,'reference_id',row['reference_id']);attr(element,version_field,version)
    return element

def parse_geometry(element,kind,dimensions,allow_outside=False,origin=(0,0)):
    check(element.tag in {'box','ellipse'},'unsupported geometry')
    check(kind!='STAR' or element.tag=='box','STAR must use box')
    rotation=float(element.get('rotation','0'))%360
    check(math.isfinite(rotation),'nonfinite rotation')
    check(kind!='STAR' or min(rotation,360-rotation)<=TOLERANCE,'STAR rotation forbidden')
    ox,oy=origin
    if element.tag=='box':
        x1,y1,x2,y2=[float(element.attrib[k]) for k in ('xtl','ytl','xbr','ybr')]
    else:
        cx,cy,rx,ry=[float(element.attrib[k]) for k in ('cx','cy','rx','ry')]
        check(rx>0 and ry>0,'invalid ellipse axes')
        x1,y1,x2,y2=cx-rx,cy-ry,cx+rx,cy+ry
    x1+=ox;x2+=ox;y1+=oy;y2+=oy
    check(all(math.isfinite(v) for v in (x1,y1,x2,y2)) and x2>x1 and y2>y1,'invalid/nonfinite shape')
    if not allow_outside:
        w,h=dimensions
        check(0<= (x1+x2)/2 <=w and 0<= (y1+y2)/2 <=h,'new object center outside panel')
    return {'geometry_type':'BOX' if element.tag=='box' else 'ELLIPSE','bbox_x1':f'{x1:.6f}','bbox_y1':f'{y1:.6f}',
            'bbox_x2':f'{x2:.6f}','bbox_y2':f'{y2:.6f}','rotation':f'{rotation:.6f}'}

def same_geometry(observed,expected):
    check(observed['geometry_type']==expected['geometry_type'],'reference shape type changed')
    # Compare ellipse centers/radii, the coordinates actually serialized by CVAT.
    if observed['geometry_type']=='ELLIPSE':
        def serialized(r):
            x1,y1,x2,y2=[float(r[k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')]
            return [(x1+x2)/2,(y1+y2)/2,(x2-x1)/2,(y2-y1)/2]
        deltas=[abs(a-b) for a,b in zip(serialized(observed),serialized(expected))]
    else:deltas=[abs(float(observed[k])-float(expected[k])) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')]
    check(max(deltas)<=TOLERANCE,'reference geometry moved/resized beyond CVAT tolerance')
    error=abs((float(observed['rotation'])-float(expected['rotation'])+180)%360-180)
    check(error<=TOLERANCE,'reference rotation changed')

def stable_id(row):
    signature=[row['object_type'],row['panel'],row['geometry_type']]
    def rounded(value):
        number=round(float(value),2)
        return f'{0. if number==0 else number:.2f}'
    signature += [rounded(row[k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')]
    signature.append(rounded(round(float(row['rotation'])%360,2)%360))
    return f'HNEW_{row["object_type"]}_{row["panel"]}_'+hashlib.sha256('|'.join(signature).encode()).hexdigest()[:16].upper()

def envelope(row):
    x1,y1,x2,y2=[float(row[k]) for k in ('bbox_x1','bbox_y1','bbox_x2','bbox_y2')]
    cx,cy=(x1+x2)/2,(y1+y2)/2;rx,ry=(x2-x1)/2,(y2-y1)/2
    angle=math.radians(float(row['rotation']));c,s=math.cos(angle),math.sin(angle)
    if row['geometry_type']=='ELLIPSE':
        ex,ey=math.hypot(rx*c,ry*s),math.hypot(rx*s,ry*c)
    else:ex,ey=abs(rx*c)+abs(ry*s),abs(rx*s)+abs(ry*c)
    return cx-ex,cy-ey,cx+ex,cy+ey

def overlaps(new,index):
    a=envelope(new);out=[]
    for old in index:
        if old['panel']!=new['panel'] or old['object_type']!=new['object_type']:continue
        b=envelope(old)
        intersection=max(0,min(a[2],b[2])-max(a[0],b[0]))*max(0,min(a[3],b[3])-max(a[1],b[1]))
        aa=(a[2]-a[0])*(a[3]-a[1]);ba=(b[2]-b[0])*(b[3]-b[1])
        iou=intersection/(aa+ba-intersection) if aa+ba>intersection else 0
        containment=intersection/min(aa,ba) if min(aa,ba)>0 else 0
        dist=math.hypot((a[0]+a[2]-b[0]-b[2])/2,(a[1]+a[3]-b[1]-b[3])/2)
        norm=dist/max(1,min(math.hypot(a[2]-a[0],a[3]-a[1]),math.hypot(b[2]-b[0],b[3]-b[1])))
        if iou>=.25 or containment>=.5 or norm<=.25:
            status='POSSIBLE_DUPLICATE_CONFIRMED' if old.get('prior_status') is None else 'MATCHES_PRIOR_'+old['prior_status']
            out.append({'canonical_new_id':new['canonical_id'],'prior_canonical_id':old['canonical_id'],'panel':new['panel'],
                        'status':status,'envelope_iou':iou,'containment':containment,'center_distance_px':dist,
                        'resolution':'AWAITING_EXPLICIT_POST_REVIEW_RECONCILIATION'})
    return out

def safe_output(path):
    path=Path(path).resolve()
    for old in (HUMAN,COMPARISON,ROOT/'research/astro_spatial_annotation',ROOT/'research/astro_spatial_annotation_ai_b',ROOT/'research/astro_spatial_annotation_ai_b2'):
        check(not path.is_relative_to(old) and not old.is_relative_to(path),'output overlaps frozen source')
    check(path not in (ROOT,Path('/')),'unsafe output directory')
    return path
