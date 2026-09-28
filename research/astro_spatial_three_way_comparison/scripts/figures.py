"""Repository-native SVG statistical graphics and PIL scientific overlays."""
from __future__ import annotations
from collections import Counter
from html import escape
import math
from PIL import Image, ImageDraw
from common import HUMAN, write
from geometry import shape

COLORS={'ACCEPT':'#278452','MODIFY':'#487bb4','REJECT':'#bc4b42','UNCERTAIN':'#d7a232','NOT_REVIEWED':'#999999'}
SOURCE_COLORS={'AI1':'#e35636','AI2':'#3975df','HUMAN':'#24a76b','CANDIDATE':'#8853aa'}

def text(x,y,value,size=13,color='#20282f',anchor='start'):
    return f'<text x="{x}" y="{y}" font-family="sans-serif" font-size="{size}" fill="{color}" text-anchor="{anchor}">{escape(str(value))}</text>'

def rect(x,y,w,h,color): return f'<rect x="{x}" y="{y}" width="{w}" height="{h}" fill="{color}"/>'

def line(x1,y1,x2,y2,color='#666',width=1): return f'<line x1="{x1}" y1="{y1}" x2="{x2}" y2="{y2}" stroke="{color}" stroke-width="{width}"/>'

def svg(path,title,source,note,elements,height=650,width=1150):
    head=f'<svg xmlns="http://www.w3.org/2000/svg" width="{width}" height="{height}" viewBox="0 0 {width} {height}">'
    body=[rect(0,0,width,height,'#fff'),text(25,32,title,20),text(25,55,'Source: '+source,12),text(25,76,note,12)]
    path.write_text(head+'\n'+'\n'.join(body+elements)+ '\n</svg>\n',encoding='utf-8')

def barplot(directory,filename,title,rows,label,value,note,ci=None,maximum=1.):
    source=filename+'_SOURCE.tsv'
    write(directory/source,rows)
    elements=[]
    left,scale=380,550/maximum
    for i,r in enumerate(rows):
        y=110+i*33
        v=float(r[value]) if r[value] is not None else 0.
        elements.extend([text(25,y+14,r[label],12),rect(left,y,max(0,v*scale),18,r.get('color','#487bb4')),
                         text(950,y+14,f"{v:.3f}; n={r.get('n',r.get('denominator','?'))}",12)])
        if ci and r[ci[0]] is not None:
            a,b=left+r[ci[0]]*scale,left+r[ci[1]]*scale
            elements.extend([line(a,y+9,b,y+9,'#222',2),line(a,y+4,a,y+14,'#222',2),line(b,y+4,b,y+14,'#222',2)])
    bottom=110+len(rows)*33
    for tick in (0,.25,.5,.75,1):
        x=left+550*tick
        elements.extend([line(x,94,x,bottom,'#ddd'),text(x,bottom+22,f'{tick*maximum:.2f}',11,anchor='middle')])
    svg(directory/(filename+'.svg'),title,source,note,elements,max(190,bottom+60))

def render(out,objects,geometry,strata,nodes,components,graph,candidates,human,sources):
    directory=out/'figures'
    directory.mkdir(exist_ok=True)
    primary=[r for r in objects if r['scope']=='PRODUCTION']
    # Count flow: physical candidates, not duplicated source records.
    rows=[]
    for kind in ('STAR','LABEL'):
        for support in ('BOTH_AI','AI1_ONLY','AI2_ONLY','NEITHER_AI'):
            cells=[r for r in primary if r['object_type']==kind and r['support_pattern']==support]
            for decision in COLORS:
                rows.append({'object_type':kind,'support_pattern':support,'human_decision':decision,'count':sum(r['human_decision']==decision for r in cells),
                             'candidate_denominator':len(cells),'scope':'PRODUCTION','human_protocol_version':'1.2'})
    write(directory/'CANDIDATE_FLOW_SOURCE.tsv',rows)
    elements=[]
    nprod=len(primary)
    elements.extend([rect(35,105,310,65,'#edf2f7'),text(50,130,'Frozen AI/A proposed union →'),text(50,152,f'Production physical candidates: n={nprod}')])
    for i,support in enumerate(('BOTH_AI','AI1_ONLY','AI2_ONLY','NEITHER_AI')):
        y=200+i*95
        cells=[r for r in primary if r['support_pattern']==support]
        counts=Counter(r['human_decision'] for r in cells)
        elements.extend([line(185,170,185,y,'#888'),line(185,y,355,y,'#888'),text(35,y+23,f'{support}; n={len(cells)}')])
        total=max(1,len(cells))
        x=355
        for decision,color in COLORS.items():
            w=700*counts[decision]/total
            elements.extend([rect(x,y-18,w,33,color)])
            if w>45: elements.append(text(x+w/2,y+3,counts[decision],12,'#fff','middle'))
            x+=w
        elements.append(text(355,y+37,'; '.join(f'{d}={counts[d]}' for d in COLORS),12))
    svg(directory/'CANDIDATE_FLOW.svg','Candidate flow: support → human outcome','CANDIDATE_FLOW_SOURCE.tsv',
        'Production; protocol 1.2; uncertainty and not-reviewed displayed separately, not negatives.',elements,620)
    # Count heatmap with raw denominators.
    write(directory/'SUPPORT_OUTCOME_SOURCE.tsv',rows)
    elements=[]
    for j,decision in enumerate(COLORS): elements.append(text(320+j*155,108,decision,12))
    i=0
    for kind in ('STAR','LABEL'):
        for support in ('BOTH_AI','AI1_ONLY','AI2_ONLY','NEITHER_AI'):
            cells=[r for r in rows if r['object_type']==kind and r['support_pattern']==support]
            n=cells[0]['candidate_denominator']
            y=130+i*48
            elements.append(text(25,y+25,f'{kind}/{support} n={n}',12))
            for j,r in enumerate(cells):
                shade=255-int(180*r['count']/max(1,n))
                elements.extend([rect(300+j*155,y,145,40,f'rgb({shade},{shade},255)'),text(372+j*155,y+25,r['count'],15,anchor='middle')])
            i+=1
    svg(directory/'SUPPORT_OUTCOME.svg','Support pattern × human decision (counts)','SUPPORT_OUTCOME_SOURCE.tsv',
        'Production candidate denominator per row; protocol 1.2; strict excludes UNCERTAIN/NOT_REVIEWED.',elements,560)
    # Geometry point estimates with bootstrap CIs, type-stratified.
    from common import read
    gs=read(out/'GEOMETRY_SUMMARY_METRICS.tsv')
    g_rows=[]
    for r in gs:
        if r['scope']!='PRODUCTION' or r['grouping']!='ALL' or r['metric']!='iou' or r['representation']=='raw_dimensions_rotated_proxy': continue
        g_rows.append({'label':r['object_type']+'/'+r['source']+'/'+r['geometry_type'],'median_iou':float(r['median']),
                       'ci95_low':float(r['ci95_low']),'ci95_high':float(r['ci95_high']),'n':int(r['n']),
                       'representation':r['representation'],'ci_method':r['ci_method'],'color':SOURCE_COLORS[r['source']],
                       'human_protocol_version':'1.2','scope':'PRODUCTION'})
    barplot(directory,'GEOMETRY_AI_HUMAN','AI/candidate versus final human geometry (median IoU)',g_rows,'label','median_iou',
            'Confirmed only; production; protocol 1.2; LABEL/ellipse envelopes are declared proxies; excluded U/NR.',('ci95_low','ci95_high'))
    # Correction histogram on actual candidate geometry.
    corr=[r for r in geometry if r['scope']=='PRODUCTION' and r['source']=='CANDIDATE']
    hist=[]
    for kind in ('STAR','LABEL'):
        for decision in ('ACCEPT','MODIFY'):
            values=[1-r['iou'] for r in corr if r['object_type']==kind and r['human_decision']==decision]
            for lo,hi in zip([0,.1,.25,.5,.75],[.1,.25,.5,.75,1.00000001]):
                n=sum(lo<=v<hi for v in values)
                hist.append({'label':f'{kind}/{decision} [{lo:.2f},{min(1,hi):.2f})','count':n,'n':len(values),
                             'bin_low':lo,'bin_high':min(1,hi),'color':COLORS[decision],'scope':'PRODUCTION','human_protocol_version':'1.2'})
    barplot(directory,'HUMAN_CORRECTION','Distribution of human correction: 1−candidate/final IoU',hist,'label','count',
            'Actual candidate geometry; confirmed production only; n=confirmed type/outcome; protocol 1.2; U/NR excluded.',maximum=max(1,max(r['count'] for r in hist)))
    # Per-panel per-AI conditional confirmation.
    panels=[]
    from statistics import fraction
    for panel in sorted({r['panel'] for r in primary}):
        for kind in ('STAR','LABEL'):
            for ai in ('AI1','AI2'):
                cells=[r for r in primary if r['panel']==panel and r['object_type']==kind and r['support_'+ai.lower()]=='YES' and r['human_decision'] in {'ACCEPT','MODIFY','REJECT'}]
                f=fraction(sum(r['human_decision'] in {'ACCEPT','MODIFY'} for r in cells),len(cells))
                if not cells: continue
                panels.append({'label':panel+'/'+kind+'/'+ai,'confirmation':f['estimate'],'numerator':f['numerator'],
                               'denominator':f['denominator'],'ci95_low':f['ci95'][0],'ci95_high':f['ci95'][1],
                               'color':SOURCE_COLORS[ai],'scope':'PRODUCTION','human_protocol_version':'1.2'})
    barplot(directory,'PANEL_CONFIRMATION','Production panel-level conditional confirmation',panels,'label','confirmation',
            'Denominator = strict reviewed AI-supported candidates; Wilson CI; protocol 1.2; U/NR excluded.',('ci95_low','ci95_high'))
    # Bipartite component topology with neutral IDs; 53 positive edges only.
    edge_table=read(out/'ATTACHMENT_V2_EDGE_RESULTS.tsv')
    positives=[r for r in edge_table if r['decision']=='VISUAL_LABEL_OF']
    write(directory/'ATTACHMENT_GRAPH_SOURCE.tsv',positives)
    write(directory/'ATTACHMENT_NODE_SOURCE.tsv',nodes)
    elements=[]
    for i,component in enumerate(components):
        col,row=i%6,i//6
        x,y=30+col*180,110+row*110
        members=component['node_ids'].split(';')
        labels=sorted(n for n in members if n.startswith('LABEL:'))
        stars=sorted(n for n in members if n.startswith('STAR:'))
        positions={n:(x+25,y+30+j*24) for j,n in enumerate(labels)}
        positions.update({n:(x+115,y+30+j*24) for j,n in enumerate(stars)})
        elements.append(text(x,y,component['component_id'],12))
        for edge in positives:
            l,s='LABEL:'+edge['label_id'],'STAR:'+edge['star_id']
            if l in positions and s in positions: elements.append(line(*positions[l],*positions[s],'#aaa',2))
        for n,(nx,ny) in positions.items():
            color='#3975df' if n.startswith('LABEL:') else '#e35636'
            elements.append(f'<circle cx="{nx}" cy="{ny}" r="7" fill="{color}"/>')
            elements.append(text(nx,ny+20,'L' if n.startswith('LABEL:') else 'S',10,anchor='middle'))
    bottom=110+math.ceil(len(components)/6)*110
    elements.extend([text(25,bottom+5,f"53 links / 51 LABEL / 51 STAR; components={graph['positive_connected_components']}; {graph['stars_with_multiple_labels']} multi-L stars, {graph['labels_with_multiple_stars']} multi-S labels."),
                     text(25,bottom+27,'Zero-degree confirmed endpoints: see ATTACHMENT_NODE_SOURCE.tsv; zero degree is not no real caption.',12)])
    svg(directory/'ATTACHMENT_V2_GRAPH','Attachment-v2 confirmed bipartite components','ATTACHMENT_GRAPH_SOURCE.tsv + ATTACHMENT_NODE_SOURCE.tsv',
        'Protocol 2; only 53 positives; 171 unassigned/43 uncertain excluded from edges; 57 filtered +351 excluded not reviewed.',elements,bottom+65)
    # Ensure extension consistent.
    (directory/'ATTACHMENT_V2_GRAPH').rename(directory/'ATTACHMENT_V2_GRAPH.svg')
    case_studies(directory,objects,geometry,candidates,human,sources)
    figure_index=[{'figure':str(p.relative_to(out)),'source_tables':';'.join(str(t.relative_to(out)) for t in sorted(directory.glob(p.stem+'*SOURCE.tsv')))}
                  for p in sorted(directory.glob('*.svg'))]
    for row in figure_index:
        if row['figure'].endswith('ATTACHMENT_V2_GRAPH.svg'): row['source_tables']='figures/ATTACHMENT_GRAPH_SOURCE.tsv;figures/ATTACHMENT_NODE_SOURCE.tsv'
    for p in sorted(directory.glob('CASE_*.png')): figure_index.append({'figure':str(p.relative_to(out)),'source_tables':'figures/CASE_STUDIES.tsv;THREE_WAY_OBJECT_RESULTS.tsv;THREE_WAY_GEOMETRY_RESULTS.tsv'})
    write(directory/'FIGURE_INDEX.tsv',figure_index)

def case_studies(directory,objects,geometry,candidates,human,sources):
    by_id={r['candidate_id']:r for r in objects}
    correction={r['candidate_id']:r['iou'] for r in geometry if r['source']=='CANDIDATE'}
    by_geo={}
    for r in geometry:
        if r['source'] in {'AI1','AI2'} and r['primary_representation']:
            by_geo.setdefault(r['candidate_id'],{})[r['source']]=r['iou']
    selections=[]
    selected=set()
    for kind in ('STAR','LABEL'):
        confirmed=[r for r in objects if r['object_type']==kind and r['human_decision'] in {'ACCEPT','MODIFY'} and r['scope']=='PRODUCTION']
        choices=[]
        if confirmed: choices.append(('LOW_CORRECTION',min(confirmed,key=lambda r:(1-correction[r['candidate_id']],r['candidate_id']))))
        modified=[r for r in confirmed if r['human_decision']=='MODIFY']
        if modified: choices.append(('MAX_MODIFY_CORRECTION',min(modified,key=lambda r:(correction[r['candidate_id']],r['candidate_id']))))
        rejected=[r for r in objects if r['object_type']==kind and r['human_decision']=='REJECT' and r['scope']=='PRODUCTION']
        both=[r for r in rejected if r['support_pattern']=='BOTH_AI']
        if rejected: choices.append(('BOTH_REJECT_ELSE_REJECT',min(both or rejected,key=lambda r:r['candidate_id'])))
        uncertain=[r for r in objects if r['object_type']==kind and r['human_decision']=='UNCERTAIN']
        if uncertain: choices.append(('UNCERTAIN',min(uncertain,key=lambda r:r['candidate_id'])))
        paired=[r for r in confirmed if len(by_geo.get(r['candidate_id'],{}))==2]
        if paired: choices.append(('MAX_PAIRED_AI_HUMAN_IOU_DISCREPANCY',min(paired,key=lambda r:(-abs(by_geo[r['candidate_id']]['AI1']-by_geo[r['candidate_id']]['AI2']),r['candidate_id']))))
        for rule,obj in choices:
            if obj['candidate_id'] in selected: continue
            selected.add(obj['candidate_id'])
            selections.append((rule,obj))
    records=[]
    overlay_sources=[]
    for index,(rule,obj) in enumerate(selections,1):
        cid,kind,panel=obj['candidate_id'],obj['object_type'],obj['panel']
        h=human[kind][cid]
        hkind='AABB' if kind=='STAR' else 'ELLIPSE_ENVELOPE_PROXY' if h['geometry_type']=='ELLIPSE' else 'ROTATED_RECTANGLE'
        shapes=[('HUMAN',shape(h,hkind))]
        for tag in ('AI1','AI2'):
            sid=obj[tag.lower()+'_id']
            if sid:
                gkind='AABB' if kind=='STAR' else hkind
                shapes.append((tag,shape(sources[kind][tag][sid],gkind,0 if gkind=='ELLIPSE_ENVELOPE_PROXY' else None,
                                         'orientation_normalized_proxy' if kind=='LABEL' else 'actual')))
        image=Image.open(HUMAN/'images'/f'{panel}.jpg').convert('RGB')
        points=[p for _,g in shapes for p in g['polygon']]
        minx,miny=max(0,int(min(p[0] for p in points))-50),max(0,int(min(p[1] for p in points))-50)
        maxx,maxy=min(image.width,int(max(p[0] for p in points))+51),min(image.height,int(max(p[1] for p in points))+51)
        crop=image.crop((minx,miny,maxx,maxy))
        scale=min(1.,1050/crop.width,750/crop.height)
        crop=crop.resize((max(1,round(crop.width*scale)),max(1,round(crop.height*scale))),Image.Resampling.LANCZOS)
        canvas=Image.new('RGB',(max(650,crop.width),crop.height+110),'white')
        canvas.paste(crop,(0,110))
        draw=ImageDraw.Draw(canvas)
        draw.text((10,10),f'CASE_{index:02d} {cid} / {panel}',fill='#111111')
        draw.text((10,30),f'{rule}; {kind}; {obj["scope"]}; decision={obj["human_decision"]}',fill='#111111')
        draw.text((10,50),'Protocol 1.2; denominator: 1 deterministic case; AI geometry proxy for LABEL.',fill='#111111')
        draw.text((10,70),'AI1 red / AI2 blue / HUMAN green; rejected/uncertain retained shape, NOT gold.',fill='#111111')
        for tag,g in sorted(shapes,key=lambda pair:pair[0]=='HUMAN'):
            color=SOURCE_COLORS[tag]
            pts=[((x-minx)*scale,(y-miny)*scale+110) for x,y in g['polygon']]
            draw.line(pts+[pts[0]],fill=color,width=3)
            overlay_sources.append({'case_id':f'CASE_{index:02d}','candidate_id':cid,'source':tag,
                                    'source_id':cid if tag=='HUMAN' else obj[tag.lower()+'_id'],
                                    'polygon_points':';'.join(f'{x:.6f},{y:.6f}' for x,y in g['polygon']),
                                    'geometry_type':g['kind'],'convention':g['convention'],'human_decision':obj['human_decision'],
                                    'human_protocol_version':'1.2'})
        filename=f'CASE_{index:02d}.png'
        canvas.save(directory/filename,optimize=False)
        records.append({'case_id':f'CASE_{index:02d}','candidate_id':cid,'panel':panel,'object_type':kind,'scope':obj['scope'],
                        'selection_rule':rule,'human_decision':obj['human_decision'],'support_pattern':obj['support_pattern'],
                        'canonical_image':str((HUMAN/'images'/f'{panel}.jpg').relative_to(HUMAN.parents[1])),
                        'crop_x1':minx,'crop_y1':miny,'crop_x2':maxx,'crop_y2':maxy,'scale':scale,
                        'candidate_human_iou':correction.get(cid),'overlay_file':'figures/'+filename,'source_table':'figures/CASE_OVERLAY_SOURCE.tsv',
                        'human_protocol_version':'1.2','excluded_note':'single selected illustration, not a frequency estimate; U/NR excluded from accuracy'})
    write(directory/'CASE_STUDIES.tsv',records)
    write(directory/'CASE_OVERLAY_SOURCE.tsv',overlay_sources)
