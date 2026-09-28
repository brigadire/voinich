#!/usr/bin/env python3
"""Structure-only eligibility audit; deliberately performs no matching/search."""
from pathlib import Path
import csv,hashlib,json,re
from collections import Counter
from datetime import datetime,timezone
ROOT=Path(__file__).resolve().parents[2];OUT=Path(__file__).parent
def rows(p): return list(csv.DictReader(open(p,encoding='utf8'),delimiter='\t'))
def wr(n,fields,data):
 with (OUT/n).open('w',newline='',encoding='utf8') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(data)
def chunks(s):
 s=re.sub(r'[^a-z]','',s.lower()); return [s[:2],s[-2:]] if len(s)>=4 else [s]
def main():
 src=[r for r in rows(ROOT/'research/astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANDED.tsv') if r['object_class']=='STAR']
 tgt=rows(ROOT/'research/astro_label_cross_section/ASTRO_LABEL_CROSS_SECTION_STAR.tsv')
 sdata=[]
 for i,r in enumerate(src,1):
  f=r['normalized_form']; c=chunks(f);sdata.append({'concept_id':r['term_id']+'_'+str(i),'attested_form':r['attested_form'],'normalized_form':f,'component_1':c[0],'component_2':c[-1],'component_3':'','component_roles':'PREFIX/SUFFIX','source':r['source']})
 tdata=[]
 for i,r in enumerate(tgt,1):
  t=r['token']; c=chunks(t);tdata.append({'label_id':f'STAR_{i:03d}','page':'frozen_cross_section','coordinate':'UNSPECIFIED','token':t,'target_component_1':c[0],'target_component_2':c[-1],'target_component_3':'','component_roles':'PREFIX/SUFFIX'})
 wr('M3_REAL_SOURCE_SEGMENTATION.tsv',list(sdata[0]),sdata);wr('M3_REAL_TARGET_SEGMENTATION.tsv',list(tdata[0]),tdata)
 sc=Counter(x['component_1'] for x in sdata);sc.update(x['component_2'] for x in sdata);tc=Counter(x['target_component_1'] for x in tdata);tc.update(x['target_component_2'] for x in tdata)
 wr('M3_REAL_SOURCE_COMPONENTS.tsv',['component','role','concept_support','form_support','sources_support'],[{'component':k,'role':'PREFIX_OR_SUFFIX','concept_support':v,'form_support':v,'sources_support':1} for k,v in sorted(sc.items())]);wr('M3_REAL_TARGET_COMPONENTS.tsv',['component','role','label_support','token_type_support','panel_support'],[{'component':k,'role':'PREFIX_OR_SUFFIX','label_support':v,'token_type_support':v,'panel_support':1} for k,v in sorted(tc.items())])
 combos=Counter((x['component_1'],x['component_2']) for x in sdata); tcomb=Counter((x['target_component_1'],x['target_component_2']) for x in tdata)
 wr('M3_REAL_COMPOSITION_GRAPH.tsv',['side','component_a','component_b','combination_support','independent_nodes'],[{'side':'SOURCE','component_a':a,'component_b':b,'combination_support':v,'independent_nodes':v} for (a,b),v in sorted(combos.items())]+[{'side':'TARGET','component_a':a,'component_b':b,'combination_support':v,'independent_nodes':v} for (a,b),v in sorted(tcomb.items())])
 splits=[{'split_id':'S0','train_size':0,'heldout_novel_combinations':0,'components_seen_in_train':'NO','combinations_unseen_in_train':'UNTESTABLE','min_rule_support':0,'eligible':'NO','reason':'confirmed STAR labels have no source-target pairing and no panel/coordinate structure'}]
 wr('M3_REAL_SPLIT_CANDIDATES.tsv',['split_id','train_size','heldout_novel_combinations','components_seen_in_train','combinations_unseen_in_train','min_rule_support','eligible','reason'],splits);wr('M3_REAL_SPLIT_VALIDATION.tsv',list(splits[0]),splits)
 (OUT/'M3_REAL_NULL_FEASIBILITY.md').write_text('''# Real M3 matched-null feasibility\n\nA matched null cannot be constructed for a valid M3 split from the frozen STAR cross-section: labels are confirmed occurrences/types without source-concept pairing, panel-level combinatorial graph, or independent repeated target assignments. Shuffling would therefore manufacture the missing pairing structure rather than preserve it. No null or brute-force search was run.\n''')
 report=f'''# M3 real-data eligibility\n\nThis is a structure-only audit of frozen D1 STAR source forms and confirmed STAR label types. No semantic assignment, M1/M2 mapping, spelling-driven split, or brute-force search was used.\n\nThe source side has {len(src)} STAR rows and repeated lexical edge chunks, but the target side has {len(tgt)} confirmed STAR token types with one confirmed occurrence each. Because the frozen label cross-section does not expose source-concept pairing or a repeated target component graph tied to concepts, a valid novel-combination split cannot be formed. A split with at least five held-out novel combinations would require inventing pairings or weakening M3P1_V constraints.\n\n```text\nM3_REAL_DATA_ELIGIBILITY=NOT_ELIGIBLE\nELIGIBLE_GENERATIVE_FAMILIES=NONE\nSTAR_LABEL_TYPES={len(tgt)}\nSOURCE_CONCEPTS={len(src)}\nSOURCE_RECURRENT_COMPONENTS_GE3={sum(v>=3 for v in sc.values())}\nTARGET_RECURRENT_COMPONENTS_GE3={sum(v>=3 for v in tc.values())}\nMAX_FEASIBLE_TRAIN=0\nMAX_FEASIBLE_NOVEL_HELDOUT=0\nELIGIBLE_SPLITS=0\nMATCHED_NULL_FEASIBLE=NO\nREAL_M3_BRUTEFORCE_AUTHORIZED=NO\n```\n\nThe result is `NOT_ELIGIBLE`, not a license to adapt segmentation or support thresholds. A future eligibility run would require a frozen source-label occurrence table with independently observed repeated component combinations.\n'''
 (OUT/'M3_REAL_ELIGIBILITY_REPORT.md').write_text(report); files=sorted(x.name for x in OUT.iterdir() if x.is_file() and x.name not in ['main.py','M3_REAL_ELIGIBILITY_MANIFEST.json','SHA256SUMS']);m={'experiment':'m3-real-data-eligibility-v1','run_type':'AUDIT_ONLY','frozen_profile':'M3P1_V','source_inputs':['research/astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANDED.tsv','research/astro_label_cross_section/ASTRO_LABEL_CROSS_SECTION_STAR.tsv'],'artifact_sha256':{n:hashlib.sha256((OUT/n).read_bytes()).hexdigest() for n in files},'generated_utc':datetime.now(timezone.utc).replace(microsecond=0).isoformat()};(OUT/'M3_REAL_ELIGIBILITY_MANIFEST.json').write_text(json.dumps(m,indent=2,sort_keys=True)+'\n');(OUT/'SHA256SUMS').write_text(''.join(f'{hashlib.sha256((OUT/n).read_bytes()).hexdigest()}  {n}\n' for n in sorted(files+['main.py'])))
if __name__=='__main__':main()
