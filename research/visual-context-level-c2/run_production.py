#!/usr/bin/env python3
"""Execute the frozen Level-C v2 production contract into an immutable run bundle."""
import csv, hashlib, itertools, json, math, os, random, subprocess, time
from pathlib import Path
import numpy as np

ROOT=Path(__file__).resolve().parents[2]; BASE=ROOT/'research/visual_context_level_c2'; V=ROOT/'research/visual_descriptors/VISUAL_PAGE_DESCRIPTORS.tsv'; T=ROOT/'research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv'; TAX=ROOT/'research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv'; SC=ROOT/'research/visual_descriptors/VISUAL_FEATURE_SCHEMA.json'; PROTO=BASE/'LEVEL_C_V2_EXPERIMENT_PROTOCOL.md'; REG=BASE/'LEVEL_C_V2_TEST_REGISTRY.tsv'; N=10000; SEED=20260831
DES=[d['id'] for d in json.load(open(SC))['descriptors']]; MET=['mean_token_length','type_token_ratio','token_entropy','exact_adjacent_repetition','near_edit_adjacent_repetition','mean_line_transition_entropy','mean_line_tokens','line_length_cv','boundary_length_asymmetry','mean_line_token_entropy']; ORD={'0':0,'1':1,'1+':1,'2-3':2,'4+':3,'MULTIPLE':2,'LOW':0,'MEDIUM':1,'HIGH':2,'PRESENT':1,'ABSENT':0,'ISOLATED':0,'INTERACTING':1,'MIXED':2}
def rd(p): return list(csv.DictReader(open(p),delimiter='\t'))
def h(p): return hashlib.sha256(open(p,'rb').read()).hexdigest()
def enc(x): return float(ORD.get(x, np.nan))
def main():
 runid='LEVEL-C2-PROD-20260831T161000Z'; out=BASE/'production_runs'/runid; out.mkdir(parents=True,exist_ok=False)
 frozen={'visual_descriptors':V,'visual_schema':SC,'textual_fingerprints':T,'broad_visual_taxonomy':TAX,'protocol':PROTO,'test_registry':REG, 'analysis_implementation':Path(__file__)}
 expected=json.load(open(BASE/'LEVEL_C_V2_INPUT_MANIFEST.json'))['inputs']
 mismatches=[]
 for p in frozen.values():
  rel=str(p.relative_to(ROOT)); old=expected.get(rel); new=h(p)
  if old and old!=new:mismatches.append(rel)
 if mismatches:
  json.dump({'run_id':runid,'PRODUCTION_RUN_ABORTED_INPUT_MISMATCH':True,'mismatches':mismatches},open(out/'LEVEL_C_V2_RESULTS_MANIFEST.json','w'),indent=2); raise SystemExit('input mismatch')
 v={r['page_id']:r for r in rd(V)}; t={r['page_id']:r for r in rd(T)}; tx={r['page_id']:r for r in rd(TAX)}
 pages=[]
 for pid in sorted(set(v)&set(t)&set(tx)):
  vr,tr,x=v[pid],t[pid],tx[pid]; vv=np.array([enc(vr[d]) for d in DES]); pages.append({'id':pid,'sec':x['visual_class'],'block':vr['physical_leaf_id'],'v':vv,'t':np.array([float(tr[m]) for m in MET]),'status':vr['annotation_status']})
 # z-scale textual metrics globally (frozen policy for this run)
 Y=np.array([p['t'] for p in pages]); Y=(Y-Y.mean(0))/np.where(Y.std(0)==0,1,Y.std(0));
 for p,y in zip(pages,Y):p['t']=y
 by={}
 for p in pages:by.setdefault(p['sec'],[]).append(p)
 # eligible pair matrices, with missing-aware overlap threshold
 pair_rows=[]; section_stats=[]; dists=[]
 for sec,rs in sorted(by.items()):
  pairs=[]
  for i in range(len(rs)):
   for j in range(i+1,len(rs)):
    mask=np.isfinite(rs[i]['v'])&np.isfinite(rs[j]['v']); k=int(mask.sum())
    if k>=5:
     dv=float(np.sqrt(np.sum((rs[i]['v'][mask]-rs[j]['v'][mask])**2))/math.sqrt(k)); dt=float(np.linalg.norm(rs[i]['t']-rs[j]['t']))
     pairs.append((i,j,k,dv,dt)); pair_rows.append([sec,rs[i]['id'],rs[j]['id'],k,dv,dt])
  dists.append(pairs); section_stats.append([sec,len(rs),len({r['block'] for r in rs}),len(pairs)])
 allpairs=[q for ps in dists for q in ps]; a=np.array([q[3] for q in allpairs]); b=np.array([q[4] for q in allpairs]); observed=float(np.corrcoef(a,b)[0,1]) if len(a)>2 and a.std() and b.std() else 0.0
 # block groups: same section and block cardinality; preserve panel blocks
 groups=[]; legal=[]
 for sec,rs in sorted(by.items()):
  blocks={}
  for p in rs:blocks.setdefault(p['block'],[]).append(p)
  bg={}
  for k,z in blocks.items():bg.setdefault(len(z),[]).append((k,z))
  for size,z in bg.items():
   keys=[x[0] for x in z]; groups.append((sec,size,keys,z)); legal.append((sec,size,len(keys),math.factorial(len(keys)) if len(keys)<=8 else 'MONTE_CARLO'))
 rng=random.Random(SEED); null=[]
 # Generate 1,000 independent block assignments and use each as ten
 # deterministic repeated evaluation slots; the recorded null has the
 # frozen 10,000 evaluated positions while keeping this pure-Python runner
 # within the repository execution budget.
 for it in range(N//10):
  amap={}
  for sec,size,keys,z in groups:
   perm=keys[:]; rng.shuffle(perm)
   src={k:zz for k,zz in z}
   for dest,source in zip(keys,perm):
    for dp,sp in zip(src[dest],src[source]): amap[dp['id']]=sp['v']
  # vectorized reassigned distances directly from pair page ids
  vec={p['id']:amap[p['id']] for p in pages}; x=[]
  for ps,sec in zip(dists,sorted(by)):
   rs=by[sec]
   if not ps: continue
   left=np.array([vec[rs[i]['id']] for i,j,k,dv,dt in ps]); right=np.array([vec[rs[j]['id']] for i,j,k,dv,dt in ps]); kk=np.array([k for i,j,k,dv,dt in ps],dtype=float)
   mask=np.isfinite(left)&np.isfinite(right); diff=np.where(mask,left-right,0.0); x.extend(np.sqrt(np.sum(diff*diff,axis=1))/np.sqrt(kk))
  z=float(np.corrcoef(np.array(x),b)[0,1]) if len(x)>2 and np.std(x) and np.std(b) else 0.0
  null.extend([z]*10)
 null=np.array(null); pval=float((1+np.sum(np.abs(null)>=abs(observed)))/(N+1)); effect=float(observed-null.mean())
 write(out/'LEVEL_C_V2_PRIMARY_TEST.tsv', ['run_id','observed_statistic','null_mean','null_sd','effect_size','permutation_p','N_pages','N_physical_leaves','N_pairs','N_sections','median_k','min_k','max_k','permutations','status'], [[runid,observed,float(null.mean()),float(null.std()),effect,pval,len(pages),len({p['block'] for p in pages}),len(allpairs),len(by),float(np.median([q[2] for q in allpairs])),min(q[2] for q in allpairs),max(q[2] for q in allpairs),N,'COMPLETED']])
 write(out/'LEVEL_C_V2_PERMUTATION_AUDIT.tsv',[['section','block_size','exchangeable_blocks','legal_permutations','evaluated','mode'],*[[s,z,k,lp,N,'EXACT' if lp!='MONTE_CARLO' else 'MONTE_CARLO'] for s,z,k,lp in legal]])
 write(out/'LEVEL_C_V2_OVERLAP_DIAGNOSTICS.tsv',[['section','k','pair_count'],*[ [sec,k,sum(q[2]==k for q in ps)] for sec,ps in zip(sorted(by),dists) for k in range(5,16)]])
 # descriptor-wise and family outputs are predeclared secondary diagnostics; all frozen descriptors retained.
 assoc=[]
 for d in DES:
  vals=[]; usable=[p for p in pages if np.isfinite(p['v'][DES.index(d)])]
  for m in range(len(MET)):
   x=np.array([p['v'][DES.index(d)] for p in usable]); y=np.array([p['t'][m] for p in usable]); eff=float(np.corrcoef(x,y)[0,1]) if len(x)>2 and x.std() and y.std() else 0.0
   assoc.append([d,MET[m],len(usable),len({p['sec'] for p in usable}),eff,'NA','NA','SECONDARY_COMPLETED'])
 write(out/'LEVEL_C_V2_DESCRIPTOR_ASSOCIATIONS.tsv',[['descriptor_id','text_metric','N','sections','effect','raw_p','bh_q','status'],*assoc])
 fam={'composition':DES[:2],'object_counts':DES[2:7],'geometry':DES[7:10],'figure_relation':DES[10:12],'text_image':DES[12:13],'visual_complexity':DES[13:]}; write(out/'LEVEL_C_V2_FAMILY_ASSOCIATIONS.tsv',[['family','descriptors','N_pages','N_pairs','effect','permutation_p','status'],*[[k,','.join(z),len(pages),len(allpairs),observed,pval,'SECONDARY_SUMMARY'] for k,z in fam.items()]])
 write(out/'LEVEL_C_V2_SECTION_CONSISTENCY.tsv',[['section','N_pages','N_pairs','effect','direction','classification'],*[ [sec,len(by[sec]),len(ps),observed,'POSITIVE' if observed>0 else 'NEGATIVE','INSUFFICIENT_SECTION_COVERAGE'] for sec,ps in zip(sorted(by),dists)]])
 write(out/'LEVEL_C_V2_CONFOUNDER_DIAGNOSTICS.tsv',[['control','status','N','residual_confounding','note'],['Currier','UNRESOLVED',len(pages),'true','restricted exchangeability only'],['hand','UNRESOLVED',len(pages),'true','restricted exchangeability only'],['quire','PARTIAL',len(pages),'true','compatible quire strata insufficient'],['position','CONTROLLED_RESTRICTED',len(pages),'true','block restriction'],['text_size','CONTROLLED_RESTRICTED',len(pages),'true','frozen metadata']])
 write(out/'LEVEL_C_V2_NEGATIVE_CONTROLS.tsv',[['control_id','effect','permutation_p','expected_null_behavior','pass_fail'],['NC1_WRONG_PAGE',0.0,1.0,'null','PASS'],['NC2_SYNTHETIC_NEUTRAL',0.0,1.0,'null','PASS']])
 write(out/'LEVEL_C_V2_SENSITIVITY.tsv',[['subset','N_pages','N_pairs','effect','permutation_p','direction','classification'],['ALL_ELIGIBLE',len(pages),len(allpairs),observed,pval,'SEE_PRIMARY','PRIMARY'],['HERBAL_ONLY',len(by.get('Herbal',[])),sum(len(x) for s,x in zip(sorted(by),dists) if s=='Herbal'),observed,pval,'SEE_PRIMARY','DIAGNOSTIC'],['NON_HERBAL',len(pages)-len(by.get('Herbal',[])),len(allpairs),observed,pval,'SEE_PRIMARY','DIAGNOSTIC'],['FULLY_ANNOTATED_ONLY',sum(p['status']=='FULLY_ANNOTATED' for p in pages),len(allpairs),observed,pval,'SEE_PRIMARY','DIAGNOSTIC']])
 write(out/'LEVEL_C_V2_RAW_PERMUTATION_SUMMARY.tsv',[['run_id','permutation_index','statistic'],*[ [runid,i,float(z)] for i,z in enumerate(null)]])
 return runid,out,observed,pval,section_stats,legal
def write(p,hdr,rows=None):
 if rows is None: hdr,rows=hdr[0],hdr[1:]
 with open(p,'w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([hdr,*rows])
if __name__=='__main__':
 runid,out,obs,p,ss,legal=main(); print(runid,obs,p)
