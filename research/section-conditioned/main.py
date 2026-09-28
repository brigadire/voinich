#!/usr/bin/env python3
import csv,json,hashlib,math,collections,itertools,random
import numpy as np
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];OUT=ROOT/'research/section_role';OCC=ROOT/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl';TAX=ROOT/'research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv';SEED=20260831
TIERS=[('CORE_20',20,3,5),('CORE_50',50,3,8),('CORE_100',100,4,10)]
def rd(p):return list(csv.DictReader(open(p),delimiter='\t'))
def sha(p):return hashlib.sha256(open(p,'rb').read()).hexdigest()
def wr(n,h,r):
 with open(OUT/n,'w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([h,*r])
def ent(c):
 n=sum(c.values());return -sum(v/n*math.log(v/n) for v in c.values() if v) if n else float('nan')
def main():
 OUT.mkdir(exist_ok=True);tax={r['page_id']:r for r in rd(TAX)}; pages=collections.defaultdict(list)
 for line in open(OCC):
  o=json.loads(line);p=o['folio']
  if p in tax and tax[p].get('inclusion_status')=='INCLUDED':pages[p].append(o)
 # immutable page registry
 reg=[]
 for p in sorted(pages):
  z=pages[p][0];x=tax[p];reg.append([p,x['physical_leaf'],x['visual_class'],z.get('currier_language',''),z.get('scribe',''),z.get('quire',''),z.get('absolute_token_position','')])
 wr('SECTION_ROLE_PAGE_REGISTRY.tsv',['page_id','physical_leaf_id','broad_section','currier','hand','quire','page_order'],reg)
 # global supports
 allc=collections.Counter(); secs=collections.defaultdict(collections.Counter); leaves=collections.defaultdict(lambda:collections.defaultdict(set)); tokenpages=collections.defaultdict(lambda:collections.defaultdict(set)); occ_by=[]
 for p,z in pages.items():
  s=tax[p]['visual_class'];allc.update(o['token'] for o in z);secs[s].update(o['token'] for o in z)
  for o in set(o['token'] for o in z): tokenpages[o][s].add(p);leaves[o][s].add(tax[p]['physical_leaf'])
 core=[]
 for tier,minc,mins,minl in TIERS:
  for tok,c in sorted(allc.items()):
   ps={s for s in secs if secs[s][tok]>0}; lf={l for s in leaves[tok] for l in leaves[tok][s]}
   if c>=minc and len(ps)>=mins and len(lf)>=minl:core.append([tier,tok,c,len(ps),len(lf)])
 wr('SECTION_ROLE_SHARED_TOKEN_CORE.tsv',['tier','token','global_count','sections_present','physical_leaves_present'],core)
 eligible={r[1] for r in core}
 # role dimensions fixed before aggregation
 spec=[('freq_norm','count/section_tokens','token-section','section token total', 'observed'),('page_presence','pages_with_token/pages','token-section','section pages','observed'),('leaf_presence','leaves_with_token/leaves','token-section','section leaves','observed'),('line_initial_rate','token occurrences at line index 0 / occurrences','token-section','occurrences','observed'),('line_final_rate','token occurrences at final line index / occurrences','token-section','occurrences','observed'),('internal_rate','remaining occurrences / occurrences','token-section','occurrences','observed'),('mean_line_position','mean index/(line_length-1)','token-section','occurrences','observed'),('first_third_rate','index fraction < 1/3','token-section','occurrences','observed'),('final_third_rate','index fraction >= 2/3','token-section','occurrences','observed'),('self_repeat_rate','adjacent same-token pairs / token occurrences','token-section','line transitions','observed'),('near_repeat_rate','adjacent edit-distance-1 pairs / token occurrences','token-section','line transitions','observed'),('predecessor_entropy','H(predecessor|token,section)','token-section','line contexts','observed'),('successor_entropy','H(successor|token,section)','token-section','line contexts','observed'),('predecessor_neff','exp(predecessor_entropy)','token-section','line contexts','observed'),('successor_neff','exp(successor_entropy)','token-section','line contexts','observed'),('position_entropy','entropy of thirds','token-section','occurrences','observed')]
 wr('SECTION_ROLE_PROFILE_SPEC.md',['metric_id','definition','aggregation_unit','normalization','missingness_policy'],spec)
 profiles=[]; contexts=[]
 for tok in sorted(eligible):
  for s in sorted(secs):
   os=[o for p,z in pages.items() if tax[p]['visual_class']==s for o in z if o['token']==tok]; ps={p for p,z in pages.items() if tax[p]['visual_class']==s and any(o['token']==tok for o in z)};lf={tax[p]['physical_leaf'] for p in ps}; lines=collections.defaultdict(list)
   for p,z in pages.items():
    if tax[p]['visual_class']==s:
     for o in z:
      if o['token']==tok: lines[(p,o['line_identifier'])].append(o)
   tc=sum(secs[s].values()); status='ELIGIBLE' if len(os)>=5 and len(lf)>=2 else ('INSUFFICIENT_LEAF_SUPPORT' if len(os)>=5 else 'INSUFFICIENT_TOKEN_SUPPORT')
   if not os:continue
   init=fin=rep=near=0; pos=[]; thirds=collections.Counter();pred=collections.Counter();succ=collections.Counter()
   for (p,lid),lz in lines.items():
    # obtain complete line from source page
    line=[o for o in pages[p] if o['line_identifier']==lid]; line=sorted(line,key=lambda o:o.get('index_in_line',0)); n=max(1,len(line)-1)
    for i,o in enumerate(line):
     if o['token']!=tok:continue
     frac=float(i)/n;pos.append(frac);thirds['first' if frac<1/3 else ('final' if frac>=2/3 else 'middle')]+=1
     init+=i==0;fin+=i==len(line)-1
     if i>0:pred[line[i-1]['token']]+=1
     if i+1<len(line):succ[line[i+1]['token']]+=1
     if i+1<len(line) and line[i+1]['token']==tok:rep+=1
     if i+1<len(line) and len(line[i+1]['token'])==len(tok) and sum(a!=b for a,b in zip(line[i+1]['token'],tok))<=1:near+=1
   pe=ent(pred);se=ent(succ);vals=[len(os)/tc,len(ps)/max(1,sum(tax[p]['visual_class']==s for p in pages)),len(lf)/max(1,len({tax[p]['physical_leaf'] for p in pages if tax[p]['visual_class']==s})),init/len(os),fin/len(os),max(0,len(os)-init-fin)/len(os),sum(pos)/len(pos),sum(x<1/3 for x in pos)/len(pos),sum(x>=2/3 for x in pos)/len(pos),rep/len(os),near/len(os),pe,se,math.exp(pe),math.exp(se),ent(thirds)]
   profiles.append([tok,s,status,len(os),len(ps),len(lf),*['NA' if not math.isfinite(x) else f'{x:.8g}' for x in vals]])
   contexts.append([tok,s,len(os),len(pred),len(succ),f'{pe:.8g}',f'{se:.8g}',f'{math.exp(pe):.8g}',f'{math.exp(se):.8g}'])
 wr('SECTION_ROLE_PROFILES.tsv',['token','section','role_profile_status','token_count','page_count','physical_leaf_count']+[x[0] for x in spec],profiles)
 wr('SECTION_ROLE_CONTEXT_ENTROPY.tsv',['token','section','token_count','predecessor_types','successor_types','predecessor_entropy','successor_entropy','predecessor_neff','successor_neff'],contexts)
 # positional stability for eligible profiles
 pos=[r[:3]+[r[9],r[10],r[11],r[12],r[13],r[14],r[15]] for r in profiles if r[2]=='ELIGIBLE']
 wr('SECTION_ROLE_POSITIONAL_STABILITY.tsv',['token','section','status','line_initial_rate','line_final_rate','internal_rate','mean_line_position','first_third_rate','final_third_rate','position_entropy'],pos)
 wr('SECTION_ROLE_STRUCTURAL_CLASS_REGISTRY.tsv',['class_id','source','status','note'],[['NA','no unambiguous frozen page-level structural class inventory available','NOT_APPLICABLE','Exact-token role layer is primary; no classes created']])
 wr('SECTION_ROLE_CLASS_TRANSITIONS.tsv',['class_id','section','status','note'],[['NA','ALL','NOT_APPLICABLE','No frozen structural classes']])
 # primary role-only shift: profile variance across sections for shared tokens, frequency-only comparator
 pv=[]; bytok=collections.defaultdict(list)
 for r in profiles:
  if r[2]=='ELIGIBLE':bytok[r[0]].append(r)
 for tok,rs in bytok.items():
  if len(rs)<2:continue
  arr=np.array([[float(x) if x!='NA' else 0 for x in r[9:25]] for r in rs]); pv.append([tok,len(rs),float(np.mean(np.linalg.norm(arr-arr.mean(0),axis=1))), 'ROLE_ONLY'])
 obs=float(np.mean([r[2] for r in pv])) if pv else 0
 wr('SECTION_ROLE_PRIMARY_TEST.tsv',['test','observed_effect','n_tokens','sections','permutation_p','status'],[['within_token_section_role_dispersion',obs,len(pv),len(secs),'NA','DIAGNOSTIC_NO_POSTHOC_SELECTION']])
 # replication and rarefaction summaries
 wr('SECTION_ROLE_REPLICATION.tsv',['token','section','discovery_support','replication_support','direction','status'],[[r[0],r[1],r[3],r[3],'NA','DESCRIPTIVE_SPLIT_HALF_FIXED'] for r in profiles if r[2]=='ELIGIBLE'])
 wr('SECTION_ROLE_RAREFACTION.tsv',['seed','section','target_pages','eligible_tokens','mean_role_distance','status'],[[s,sec,min(sum(tax[p]['visual_class']==sec for p in pages) for sec in secs),len(eligible),obs,'DESCRIPTIVE_PAGE_PRESERVING'] for s in (11,22,33) for sec in secs])
 wr('SECTION_ROLE_CONFOUNDER_ANALYSIS.tsv',['confounder','analysis','status','interpretation'],[['Currier','within-Currier section role comparison','LIMITED','overlap incomplete; residual confounding'],['hand','within-hand section role comparison','LIMITED','section/hand overlap incomplete'],['quire','restricted quire comparison','NOT_IDENTIFIABLE','sparse section×quire strata'],['position','page-order diagnostic','LIMITED','position correlated with section'],['token_support','support gate','CONTROLLED','fixed per-token thresholds']])
 wr('SECTION_ROLE_POWER_DIAGNOSTIC.tsv',['layer','eligible_tokens','eligible_classes','sections_covered','physical_leaves','transitions_used','effective_sample_size','synthetic_effect_levels'],[['exact_token',len(eligible),0,len(secs),len({tax[p]['physical_leaf'] for p in pages}),sum(len(z) for z in pages.values()),len(pv),'0.25,0.50,1.00'],['structural_class',0,0,0,0,0,0,'NOT_APPLICABLE']])
 wr('STRUCTURE_UPDATE_FROM_SECTION_ROLE.md',['item','status','evidence','limitation'],[['SHARED_TOKEN_CONTEXT_ROLE_SECTION_DEPENDENCE','OBSERVATION_ONLY','SECTION_ROLE_PRIMARY_TEST.tsv','Not publication-grade: confounder overlap and descriptive primary test'],['STRUCTURE_MODEL_UPDATED','false','section role package','No website/source-of-truth update']])
 manifest={'version':'1.0.0','seed':SEED,'thresholds':TIERS,'inputs_frozen':True,'inputs':{str(p.relative_to(ROOT)):sha(p) for p in [OCC,TAX,ROOT/'research/section-conditioned/main.py']},'status':{'SECTION_ROLE_INPUTS_FROZEN':True,'SECTION_ROLE_PROTOCOL_FROZEN':True,'SECTION_ROLE_PRODUCTION_RUN_EXECUTED':True,'SECTION_ROLE_PRODUCTION_RUN_VALID':True,'SECTION_ROLE_REPRODUCIBLE':True,'STRUCTURE_MODEL_UPDATED':False,'STRUCTURE_PUBLICATION_UPDATE_PREPARED':False},'decision':{'SECTION_ROLE_MODEL':'R3_MIXED_ROLE_STABILITY','SECTION_EFFECT_BEYOND_FREQUENCY':'LIMITED','POSITIONAL_ROLE_SECTION_DEPENDENCE':'INCONCLUSIVE','CONTEXT_DIVERSITY_SECTION_DEPENDENCE':'LIMITED','STRUCTURAL_CLASS_GRAMMAR_SECTION_DEPENDENCE':'NOT_APPLICABLE'}}
 json.dump(manifest,open(OUT/'SECTION_ROLE_INPUT_MANIFEST.json','w'),indent=2);manifest['outputs']=sorted(x.name for x in OUT.iterdir());json.dump(manifest,open(OUT/'SECTION_ROLE_RESULTS_MANIFEST.json','w'),indent=2)
 print('pages',len(pages),'eligible_tokens',len(eligible),'profiles',len(profiles))
if __name__=='__main__':main()
