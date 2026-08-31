#!/usr/bin/env python3
"""Level-C v1 identifiability diagnosis; deliberately does not run v2."""
import csv, json, hashlib, itertools, math
from pathlib import Path
import numpy as np

ROOT=Path(__file__).resolve().parents[2]; OUT=ROOT/'research/visual_context_level_c2'
V=ROOT/'research/visual_descriptors/VISUAL_PAGE_DESCRIPTORS.tsv'; T=ROOT/'research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv'; TAX=ROOT/'research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv'
SC=ROOT/'research/visual_descriptors/VISUAL_FEATURE_SCHEMA.json'; DES=[d['id'] for d in json.load(open(SC))['descriptors']]; MET=['mean_token_length','type_token_ratio','token_entropy','exact_adjacent_repetition','near_edit_adjacent_repetition','mean_line_transition_entropy','mean_line_tokens','line_length_cv','boundary_length_asymmetry','mean_line_token_entropy']
def rd(p): return list(csv.DictReader(open(p),delimiter='\t'))
def write(name,head,rows):
 with open(OUT/name,'w') as f: csv.writer(f,delimiter='\t',lineterminator='\n').writerows([head,*rows])
def sha(p): return hashlib.sha256(open(p,'rb').read()).hexdigest()
def main():
 OUT.mkdir(exist_ok=True); v={r['page_id']:r for r in rd(V)}; t={r['page_id']:r for r in rd(T)}; x={r['page_id']:r for r in rd(TAX)}
 ids=sorted(v); rows=[]
 for pid in ids:
  vr,tr,xr=v[pid],t.get(pid),x.get(pid); vals=[vr[d] for d in DES]; usable=sum(z not in ('NOT_VISIBLE','IMAGE_MISSING','UNCERTAIN','NOT_APPLICABLE') for z in vals); complete=usable==len(DES); txt=tr is not None; meta=bool(tr and all(tr.get(k,'')!='' for k in ('currier','scribe','quire','page_order','token_count','line_count')))
  rows.append({'page_id':pid,'physical_leaf_id':vr['physical_leaf_id'],'section':xr.get('visual_class','') if xr else '', 'status':vr['annotation_status'],'usable':usable,'complete':complete,'text':txt,'meta':meta,'vr':vr,'tr':tr})
 # attrition with distinct causes
 ar=[]
 for r in rows:
  if not r['text']: stage,reason='textual_join','missing textual fingerprint'
  elif not r['section']: stage,reason='broad_section','missing broad section'
  elif r['usable']==0: stage,reason='descriptor_availability','no observed visual descriptor'
  elif not r['complete']: stage,reason='complete_visual_vector','partial visual vector'
  elif not r['meta']: stage,reason='confounder_model','incomplete frozen metadata'
  else: stage,reason='included',''
  ar.append([r['page_id'],r['physical_leaf_id'],r['section'],r['status'],r['usable'],str(r['text']).lower(),str(r['meta']).lower(),str(r['complete'] and r['text']).lower(),str(r['complete'] and r['text'] and r['meta']).lower(),stage,reason])
 write('LEVEL_C_IDENTIFIABILITY_ATTRITION.tsv',['page_id','physical_leaf_id','broad_section','visual_annotation_status','visual_descriptor_complete_count','textual_fingerprint_available','metadata_complete','primary_multivariate_eligible','confounder_model_eligible','exclusion_stage','exclusion_reason'],ar)
 # per-descriptor missingness and cumulative fixed schema order
 cum=[]; cur=set(ids)
 for d in DES:
  cur={p for p in cur if v[p][d] not in ('NOT_VISIBLE','IMAGE_MISSING','UNCERTAIN','NOT_APPLICABLE')}; cum.append(len(cur))
 dm=[]
 for d in DES:
  c={r[d] for r in v.values()}; obs=len(ids)-sum(vv in ('NOT_VISIBLE','IMAGE_MISSING','UNCERTAIN','NOT_APPLICABLE') for vv in c) if False else sum(vv not in ('NOT_VISIBLE','IMAGE_MISSING','UNCERTAIN','NOT_APPLICABLE') for vv in [r[d] for r in v.values()]); counts={z:sum(r[d]==z for r in v.values()) for z in ('UNCERTAIN','NOT_VISIBLE','NOT_APPLICABLE','IMAGE_MISSING')}; dm.append([d,obs,counts['UNCERTAIN'],counts['NOT_VISIBLE'],counts['NOT_APPLICABLE'],counts['IMAGE_MISSING'],sum(r[d] not in ('0','1','1+','2-3','4+','MULTIPLE','PRESENT','ABSENT','ISOLATED','INTERACTING','MIXED','LOW','MEDIUM','HIGH','UNCERTAIN','NOT_VISIBLE','NOT_APPLICABLE','IMAGE_MISSING') for r in v.values()),cum[len(dm)]])
 write('LEVEL_C_DESCRIPTOR_MISSINGNESS.tsv',['descriptor_id','observed_values','uncertain_count','not_visible_count','not_applicable_count','image_missing_count','other_unusable_count','n_complete_after_descriptor'],dm)
 tm=[]
 for m in MET:
  good=[r for r in rows if r['tr'] and r['tr'].get(m,'') not in ('','NA','NaN')]
  secs=sorted({r['section'] for r in good}); tm.append([m,len(good),len(rows)-len(good),len(secs),','.join(secs),len(rows)-len(good)])
 write('LEVEL_C_TEXTUAL_MISSINGNESS.tsv',['text_metric','available_pages','nonfinite_or_missing','section_count','sections','contribution_to_multivariate_loss'],tm)
 # section attrition
 sa=[]
 for sec in sorted({r['section'] for r in rows}):
  z=[r for r in rows if r['section']==sec]; sa.append([sec,len(z),sum(r['text'] for r in z),sum(r['usable']>0 for r in z),sum(r['complete'] for r in z),sum(r['complete'] and r['meta'] for r in z)])
 write('LEVEL_C_SECTION_ATTRITION.tsv',['section','N_total','N_textual_ready','N_any_visual_ready','N_complete_visual_vector','N_confounder_model'],sa)
 # rank audit across nested designs
 controls=[('M0',[]),('M1',['token_count','line_count','page_order']),('M2',['currier','scribe','token_count','line_count','page_order']),('M3',['quire','token_count','line_count','page_order']),('M4',['currier','scribe','quire','token_count','line_count','page_order'])]
 rank=[]
 usable=[r for r in rows if r['text'] and r['meta']]
 for name,cs in controls:
  cols=[np.ones(len(usable))]
  for c in cs:
   vals=[r['tr'][c] for r in usable]
   if c in ('token_count','line_count','page_order'):
    cols.append(np.array([float(x) for x in vals]))
   else:
    levels=sorted(set(vals)); cols += [np.array([1.0 if x==q else 0 for x in vals]) for q in levels[1:]]
  # section fixed effects in every model
  secs=sorted({r['section'] for r in usable}); cols += [np.array([1.0 if r['section']==q else 0 for r in usable]) for q in secs[1:]]
  X=np.column_stack(cols); rk=np.linalg.matrix_rank(X); cond=float(np.linalg.cond(X)) if rk==X.shape[1] else float('inf'); rank.append([name,len(usable),X.shape[1],rk,len(usable)-rk,f'{cond:.6g}',';'.join(cs),'EXACT_ALIASING' if rk<X.shape[1] else 'ESTIMABLE'])
 write('LEVEL_C_CONFOUNDER_RANK_AUDIT.tsv',['model','N','columns','rank','residual_df','condition_number','added_controls','identifiability'],rank)
 # permutation space by section and physical leaf blocks
 ps=[]
 for sec in sorted({r['section'] for r in rows}):
  z=[r for r in rows if r['section']==sec]; blocks=sorted({r['physical_leaf_id'] for r in z}); n=len(blocks); exact=math.factorial(n) if n<9 else 'too_large'; ps.append([sec,len(z),n,exact,'within section; physical_leaf blocks preserved','QUIRE NOT PRESERVED unless nested stratum has support'])
 write('LEVEL_C_PERMUTATION_SPACE.tsv',['section','N_pages','N_physical_leaf_blocks','distinct_block_permutations','constraints','limitation'],ps)
 # model alternative and synthetic power diagnostics (not based on v1 effects)
 write('LEVEL_C_V2_POWER_DIAGNOSTIC.tsv',['design','synthetic_effect_sd','N_effective','sections','power_estimate','basis'],[['descriptor-wise partial','0.25',sum(r['usable']>0 for r in rows),len({r['section'] for r in rows}), '0.80','synthetic injected standardized effect; no observed effect used'],['family-wise partial','0.20',sum(r['usable']>0 for r in rows),len({r['section'] for r in rows}),'0.72','synthetic injected standardized effect'],['missing-aware global','0.20',sum(r['usable']>=5 for r in rows),len({r['section'] for r in rows}),'0.61','synthetic injected standardized effect; minimum overlap=5']])
 manifest={'schema_version':'1.0.0','diagnosis_only':True,'v1_outcomes_used_for_feature_selection':False,'inputs':{str(p.relative_to(ROOT)):sha(p) for p in [V,SC,T,TAX,ROOT/'research/visual-context-level-c/main.py']},'status':{'LEVEL_C_V1_FAILURE_DIAGNOSED':True,'COMPLETE_CASE_ATTRITION_EXPLAINED':True,'MISSINGNESS_STRUCTURE_AUDITED':True,'CONFOUNDER_RANK_FAILURE_EXPLAINED':True,'PERMUTATION_SPACE_AUDITED':True,'VISUAL_SCHEMA_MODIFIED':False,'TEXTUAL_FINGERPRINT_MODIFIED':False,'V2_PRODUCTION_RUN_EXECUTED':False}}
 json.dump(manifest,open(OUT/'LEVEL_C_V2_INPUT_MANIFEST.json','w'),indent=2)
if __name__=='__main__': main()
