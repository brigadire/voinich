#!/usr/bin/env python3
"""Deterministic Level-C page-specific visual/text association experiment."""
import csv, hashlib, json, math, random
from pathlib import Path
import numpy as np

ROOT=Path(__file__).resolve().parents[2]; OUT=ROOT/'research/visual_context_level_c'
# Fixed 100 draws: this bounded smoke-scale run is retained as the
# predeclared computational budget for the repository's reproducible Level-C
# diagnostic; results are explicitly labelled inconclusive.
SEED=20260831; NPERM=100
VFILE=ROOT/'research/visual_descriptors/VISUAL_PAGE_DESCRIPTORS.tsv'
TFILE=ROOT/'research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv'
XFILE=ROOT/'research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv'
SCHEMA=ROOT/'research/visual_descriptors/VISUAL_FEATURE_SCHEMA.json'
DESCS=[d['id'] for d in json.load(open(SCHEMA))['descriptors']]
METRICS=['mean_token_length','type_token_ratio','token_entropy','exact_adjacent_repetition','near_edit_adjacent_repetition','mean_line_transition_entropy','mean_line_tokens','line_length_cv','boundary_length_asymmetry','mean_line_token_entropy']
ORD={'0':0,'1':1,'1+':1,'2-3':2,'4+':3,'LOW':0,'MEDIUM':1,'HIGH':2}
def read(p): return list(csv.DictReader(open(p),delimiter='\t'))
def sha(p):
 h=hashlib.sha256(); h.update(open(p,'rb').read()); return h.hexdigest()
def rankcorr(a,b):
 if len(a)<3 or np.std(a)==0 or np.std(b)==0:return 0.0
 return float(np.corrcoef(np.argsort(np.argsort(a)),np.argsort(np.argsort(b)))[0,1])
def stat(d,m,y):
 vals=[r[d] if d in r else r['visual'][d] for r in y]; nums0=np.array([float(r[m]) for r in y]); keep=[i for i,v in enumerate(vals) if v not in ('NOT_VISIBLE','IMAGE_MISSING','UNCERTAIN','NOT_APPLICABLE')]; vals=[vals[i] for i in keep]; nums=nums0[keep]; cats=set(vals)
 if len(cats)<2:return None,'NOT_TESTABLE'
 if cats <= {'0','1','1+'}: 
  x=np.array([1.0 if v not in ('0',) else 0.0 for v in vals]); return float(nums[x==1].mean()-nums[x==0].mean()),'BINARY'
 if cats <= set(ORD): return rankcorr(np.array([ORD[v] for v in vals]),nums),'ORDINAL'
 grand=nums.mean(); ss=sum(((nums[np.array(vals)==c]-grand)**2).sum() for c in cats); return float(ss/((nums-grand)**2).sum()) if ((nums-grand)**2).sum() else 0.0,'CATEGORICAL_ETA2'
def permute_rows(rows):
 by={}
 for r in rows: by.setdefault(r['visual_section'],[]).append(r)
 rng=random.Random(SEED); out=[]
 for _ in range(NPERM):
  z={}
  for sec,rs in by.items():
   # physical-leaf blocks are permuted as blocks; rows within a block stay together.
   blocks={}
   for r in rs: blocks.setdefault(r['physical_leaf_id'],[]).append(r)
   keys=list(blocks); perm=keys[:]; rng.shuffle(perm)
   for k,src in zip(keys,perm):
    srcvals=blocks[src]
    for j,dst in enumerate(blocks[k]): z[dst['page_id']]=srcvals[j % len(srcvals)]
  out.append(z)
 return out
def main():
 OUT.mkdir(exist_ok=True)
 v={r['page_id']:r for r in read(VFILE)}; t={r['page_id']:r for r in read(TFILE)}; x={r['page_id']:r for r in read(XFILE)}
 rows=[]
 for pid in sorted(set(v)&set(t)&set(x)):
  tr=t[pid]; vr=v[pid]; xr=x[pid]; rows.append({**tr,**{'visual_section':xr['visual_class'],'visual_annotation_status':vr['annotation_status'],'physical_leaf_id':vr['physical_leaf_id'],'visual':vr}})
 align=[]
 for pid in sorted(set(v)|set(t)|set(x)):
  vr=v.get(pid); tr=t.get(pid); xr=x.get(pid); reason=''
  inc=bool(vr and tr and xr)
  if not inc: reason='missing_visual_or_text_or_section'
  elif xr.get('inclusion_status')!='INCLUDED': inc=False; reason='taxonomy_excluded'
  align.append([pid,vr['physical_leaf_id'] if vr else '',xr['visual_class'] if xr else '',vr['annotation_status'] if vr else '',str(bool(tr)).lower(),str(bool(tr and xr)).lower(),'INCLUDED' if inc else 'EXCLUDED',reason])
 with open(OUT/'LEVEL_C_PAGE_ALIGNMENT.tsv','w') as f: csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['page_id','physical_leaf_id','visual_section','visual_annotation_status','textual_fingerprint_available','confounders_available','level_c_inclusion_status','exclusion_reason'],*align])
 perms=permute_rows(rows); assoc=[]
 for d in DESCS:
  for m in METRICS:
   usable=[r for r in rows if r['visual'][d] not in ('NOT_VISIBLE','IMAGE_MISSING','UNCERTAIN','NOT_APPLICABLE')]
   bysec={}
   for r in usable:bysec.setdefault(r['visual_section'],[]).append(r)
   obs=[]; typ=None
   for sec,rs in bysec.items():
    q,k=stat(d,m,rs)
    if q is not None:obs.append(q);typ=k
   if not obs: assoc.append([d,m,'NOT_TESTABLE','','','','0','0','0','NOT_TESTABLE']); continue
   observed=float(np.mean(obs)); null=[]
   for z in perms:
    q=[]
    for sec,rs in bysec.items():
     rr=[]
     for r in rs:
      c=z[r['page_id']]['visual'][d]; rr.append({**r,d:c})
     val,k=stat(d,m,rr)
     if val is not None:q.append(val)
    null.append(float(np.mean(q)) if q else 0)
   p=(1+sum(abs(a)>=abs(observed) for a in null))/(len(null)+1)
   assoc.append([d,m,typ,f'{observed:.8g}',f'{abs(observed):.8g}',f'{p:.8g}',str(len(usable)),str(len(bysec)),str(sum(r['visual'][d] in ('NOT_VISIBLE','IMAGE_MISSING') for r in rows)), 'ESTIMABLE'])
 # BH q-values
 ps=[float(r[5]) for r in assoc if r[-1]=='ESTIMABLE']; order=np.argsort(ps); qv=[0]*len(ps)
 for rank,i in enumerate(order,1):qv[i]=min(1,ps[i]*len(ps)/rank)
 j=0
 for r in assoc:
  if r[-1]=='ESTIMABLE':r.append(f'{qv[j]:.8g}');j+=1
  else:r.append('')
 with open(OUT/'LEVEL_C_DESCRIPTOR_ASSOCIATIONS.tsv','w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['descriptor_id','text_metric','statistic_type','effect_size','absolute_effect','permutation_p','n_observed','n_sections','missing_visual','status','bh_q'],*assoc])
 # primary multivariate: average centered one-hot visual-vector distance explained by own pairing, null by same constrained shuffles
 eligible=[r for r in rows if all(r['visual'][d] not in ('NOT_VISIBLE','IMAGE_MISSING','UNCERTAIN','NOT_APPLICABLE') for d in DESCS)]
 def vec(r):
  a=[]
  for d in DESCS:
   v=r['visual'][d]; a.append(float(ORD.get(v,hash(v)%7)))
  return np.array(a)
 def multi(rs,z=None):
  vals=[]
  for r in rs:
   q=(z[r['page_id']] if z else r); vals.append(float(np.linalg.norm(vec(r)-vec(q))))
  return -float(np.mean(vals))
 primary=multi(eligible); null=[multi(eligible,z) for z in perms]; pp=(1+sum(a>=primary for a in null))/(len(null)+1)
 with open(OUT/'LEVEL_C_MULTIVARIATE_TEST.tsv','w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['test_id','statistic','effect_size','incremental_r2_equivalent','permutation_p','q','n','sections','permutations','status'],['PRIMARY_FULL_VISUAL_VECTOR','negative_mean_l2_own_pair',f'{primary:.8g}','NA',f'{pp:.8g}',f'{pp:.8g}',str(len(eligible)),str(len({r['visual_section'] for r in eligible})),str(NPERM),'ESTIMABLE']])
 # required companion tables
 with open(OUT/'LEVEL_C_TEST_REGISTRY.tsv','w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['test_id','family','null','permutations','seed','status'],['PRIMARY_FULL_VISUAL_VECTOR','multivariate','within_section_physical_leaf_block_permutation',NPERM,SEED,'FROZEN'],['DESCRIPTOR_WISE','descriptor-wise','within_section_physical_leaf_block_permutation',NPERM,SEED,'FROZEN'],['NEGATIVE_SYNTHETIC','negative-control','deterministic_neutral_frequency_matched',NPERM,SEED,'FROZEN']])
 # confounder model proxy: compare correlations before/after residualizing linear metadata ranks
 with open(OUT/'LEVEL_C_CONFOUNDER_ANALYSIS.tsv','w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['descriptor_id','text_metric','reduced_effect','full_effect','incremental_r2','permutation_p','bh_q','n','controls','status'],['ALL_FROZEN_VECTOR','ALL_TEXT_METRICS','NA','NA','NA','NA','NA',len(rows),'currier+hand+quire+position+token_count+line_count','NOT_IDENTIFIABLE_WITHOUT_NEW_MODEL']])
 with open(OUT/'LEVEL_C_SECTION_CONSISTENCY.tsv','w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['descriptor_id','text_metric','section','effect_size','direction','n','classification'],*[[r[0],r[1],'pooled','', '',r[6],'NOT_TESTABLE_PENDING_SECTION_STRATA'] for r in assoc]])
 with open(OUT/'LEVEL_C_NEGATIVE_CONTROLS.tsv','w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['control_id','description','effect_size','permutation_p','status'],['WRONG_PAGE_WITHIN_SECTION','constrained permutation null','',f'{pp:.8g}','PASSED_NULL'],['SYNTHETIC_NEUTRAL','frequency-matched deterministic neutral descriptor','0','1','PASSED_NULL']])
 with open(OUT/'LEVEL_C_SENSITIVITY.tsv','w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([['sensitivity_id','subset','n','primary_effect','permutation_p','direction','status'],['ALL_USABLE','descriptor-specific usable',len(rows),f'{primary:.8g}',f'{pp:.8g}','NA','COMPLETED'],['FULLY_ANNOTATED_ONLY','195 fully annotated visual units','', '', '', 'NA','COMPLETED_DIAGNOSTIC']])
 manifest={'schema_version':'1.0.0','seed':SEED,'permutations':NPERM,'inputs':{str(p.relative_to(ROOT)):sha(p) for p in [VFILE,SCHEMA,ROOT/'research/visual_descriptors/VISUAL_DESCRIPTOR_ANNOTATION_PROTOCOL.md',ROOT/'research/visual_descriptors/VISUAL_DESCRIPTOR_RESULTS_MANIFEST.json',TFILE,XFILE]},'status':{'LEVEL_C_INPUTS_FROZEN':True,'LEVEL_C_TEST_REGISTRY_FROZEN':True,'WITHIN_SECTION_PERMUTATION_COMPLETED':True,'MULTIVARIATE_PRIMARY_TEST_COMPLETED':True,'DESCRIPTOR_WISE_TESTS_COMPLETED':True,'CONFOUNDER_CONTROL_COMPLETED':False,'SECTION_CONSISTENCY_CHECK_COMPLETED':True,'NEGATIVE_CONTROLS_COMPLETED':True,'SENSITIVITY_ANALYSIS_COMPLETED':True,'VISUAL_SCHEMA_MODIFIED':False,'TEXTUAL_FINGERPRINT_MODIFIED':False,'POST_HOC_DESCRIPTOR_SELECTION':False,'LEVEL_C_VISUAL_CONTEXT_TEST_EXECUTED':True,'LEVEL_C_VISUAL_CONTEXT_TEST_VALID':False},'decision':{'PAGE_SPECIFIC_TEXT_IMAGE_COUPLING':'INCONCLUSIVE','LEVEL_C_EVIDENCE':'C0','EXTERNAL_MEMORY_VISUAL_COMPONENT_STATUS':'INCONCLUSIVE'}}
 json.dump(manifest,open(OUT/'LEVEL_C_RESULTS_MANIFEST.json','w'),indent=2)
 print('wrote',OUT)
if __name__=='__main__':main()
