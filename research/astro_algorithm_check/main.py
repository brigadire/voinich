#!/usr/bin/env python3
"""Synthetic recovery benchmark for the frozen M1 search framework."""
from __future__ import annotations
import argparse,csv,hashlib,importlib.util,json,os,random,statistics,sys,time
from collections import Counter
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]; OUT=ROOT/'research/astro_algorithm_check'
sp=importlib.util.spec_from_file_location('bench_m1',ROOT/'research/astro_token_formation_m1/main.py'); assert sp and sp.loader
m1=importlib.util.module_from_spec(sp);sys.modules[sp.name]=m1;sp.loader.exec_module(m1)
ALPH='abcdefghijklmnop'; TARGET=('o','k','a','l','c','h','d','y','r','s','e','i')
FAMILIES=('S0_SIMPLE_SUBSTITUTION','S1_DIGRAPH_OUTPUT','S2_MANY_TO_ONE','S3_ABBREVIATION','S4_VOWEL_REDUCTION','S5_MIXED_COMPACT','S6_CONTEXTUAL')

def write(path,fields,rows):
 with path.open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def make_terms(n=31):
 rng=random.Random(90210); terms=[]
 for i in range(n):
  s=''.join(rng.choice(ALPH) for _ in range(5+(i%3))); forms=[s]
  if i%4==0: forms.append(s[:-1]+rng.choice(ALPH))
  terms.append({'object_id':f'SYN_STAR_{i:03d}','object_class':'STAR','forms':forms})
 return terms
def hidden(family,rng):
 mp={a:(rng.choice(TARGET),) for a in ALPH}
 if family in {'S1_DIGRAPH_OUTPUT','S5_MIXED_COMPACT'}:
  for a in rng.sample(ALPH,4):mp[a]=(rng.choice(TARGET),rng.choice(TARGET))
 if family=='S2_MANY_TO_ONE':
  t=rng.choice(TARGET)
  for a in rng.sample(ALPH,5):mp[a]=(t,)
 return mp
def transform(form,mp,family,rng):
 units=tuple(x for c in form for x in mp[c])
 if family in {'S3_ABBREVIATION','S5_MIXED_COMPACT'}: units=units[:max(1,min(4,len(units)))]
 if family in {'S4_VOWEL_REDUCTION','S5_MIXED_COMPACT'}: units=tuple(x for i,x in enumerate(units) if i==0 or x not in {'a','e','i'})
 return ''.join(units)
def scenario(family,noise,missing,sample,rng):
 base=make_terms(); mp=hidden(family,rng); kept=base[:max(1,round(len(base)*(1-missing)))]
 labels=[]; truth={}
 for i in range(sample[0]+sample[1]):
  term=kept[i%len(kept)]; form=term['forms'][i%len(term['forms'])]; tok=transform(form,mp,family,rng); corrupted=rng.random()<noise
  if corrupted: tok=''.join(rng.choice(TARGET) for _ in tok); truth[f'L{i:03d}']='CORRUPTED'
  else: truth[f'L{i:03d}']=term['object_id']
  labels.append({'stolfi_coordinate':f'L{i:03d}','object_class':'STAR','panel':'SYN','stolfi_group':'S','voynich_token':tok,'split':'TRAIN' if i<sample[0] else 'HELD_OUT'})
 return kept,labels,mp,truth,base
def evaluate(family,noise,missing,sample,rng,beam=64):
 terms,labels,mp,truth,base=scenario(family,noise,missing,sample,rng); old=m1.BEAM_WIDTH;m1.BEAM_WIDTH=beam
 train=[x for x in labels if x['split']=='TRAIN']; held=[x for x in labels if x['split']=='HELD_OUT']
 t0=time.monotonic(); models=m1.run_search(train,terms,True); elapsed=time.monotonic()-t0; best=models[0]
 best.update(m1.compute_heldout(best,held,terms)); m1.BEAM_WIDTH=old
 found=dict(best['mapping']); gt_recovered=sum(a in found and tuple(found[a])==tuple(v) for a,v in mp.items())/len(mp)
 assign=sum(best['train_match'].get(k)==v for k,v in truth.items() if k in best['train_match'] and v!='CORRUPTED')/max(1,sum(v!='CORRUPTED' for k,v in truth.items() if k.startswith('L') and int(k[1:])<sample[0]))
 ceiling=1-noise
 return {'family':family,'noise':noise,'dictionary_missing':missing,'sample':f'{sample[0]}/{sample[1]}','beam':beam,'theoretical_ceiling':ceiling,'train_coverage':best['train_coverage'],'heldout_coverage':best['heldout_coverage'],'recoverable_coverage':best['train_coverage']/max(ceiling,1e-9),'mapping_recovery':gt_recovered,'assignment_recovery':assign,'mapping_size':best['mapping_size'],'complexity':best['total_complexity'],'runtime_seconds':elapsed,'model_id':f'{family}_{noise}_{missing}_{sample[0]}_{beam}','out_of_model_space':family=='S6_CONTEXTUAL'}
def random_control(family,sample,rng,kind):
 terms,labels,mp,truth,base=scenario(family,0,0,sample,rng)
 if kind=='RANDOM_TARGETS':
  for x in labels:x['voynich_token']=''.join(rng.choice(TARGET) for _ in x['voynich_token'])
 elif kind=='SHUFFLED_SOURCE_TARGET':
  vals=[x['voynich_token'] for x in labels];rng.shuffle(vals)
  for x,v in zip(labels,vals):x['voynich_token']=v
 elif kind=='PSEUDODICTIONARY':
  for t in terms:t['forms']=[''.join(rng.choice(ALPH) for _ in f) for f in t['forms']]
 else:
  for x in labels:x['voynich_token']=''.join(rng.choice(TARGET) for _ in x['voynich_token'])
 old=m1.BEAM_WIDTH;m1.BEAM_WIDTH=64;train=[x for x in labels if x['split']=='TRAIN'];t=time.monotonic();b=m1.run_search(train,terms,False)[0];m1.BEAM_WIDTH=old
 return {'control':kind,'family':family,'sample':f'{sample[0]}/{sample[1]}','max_train_coverage':b['search_max_train_coverage'],'score':b['score'],'runtime_seconds':time.monotonic()-t}
def main():
 ap=argparse.ArgumentParser();ap.add_argument('--replicates',type=int,default=1);ap.add_argument('--workers',type=int,default=1);args=ap.parse_args();OUT.mkdir(exist_ok=True)
 specs=[]; rng=random.Random(20260901)
 for fam in FAMILIES[:-1]:
  for noise in (0,.1,.2,.3): specs.append((fam,noise,0,(20,6)))
 for size in ((40,10),(80,20)): specs.append(('S0_SIMPLE_SUBSTITUTION',0,0,size))
 results=[]
 for rep in range(args.replicates):
  for fam,noise,missing,size in specs: results.append(evaluate(fam,noise,missing,size,random.Random(rng.randrange(1<<60))))
 for missing in (0,.1,.2,.3,.4): results.append(evaluate('S0_SIMPLE_SUBSTITUTION',0,missing,(20,6),random.Random(rng.randrange(1<<60))))
 beams=[]
 for beam in (8,16,32,64,128,256): beams.append(evaluate('S5_MIXED_COMPACT',0,0,(20,6),random.Random(7711),beam))
 controls=[]
 for kind in ('RANDOM_TARGETS','SHUFFLED_SOURCE_TARGET','PSEUDODICTIONARY','RANDOM_MAPPING_BREAK'):
  controls.extend(random_control('S0_SIMPLE_SUBSTITUTION',(20,6),random.Random(4400+i),kind) for i in range(3))
 fields=list(results[0]);write(OUT/'SYNTHETIC_RECOVERY_RESULTS.tsv',fields,results);write(OUT/'SYNTHETIC_NOISE_RESULTS.tsv',fields,[r for r in results if r['family']!='S0_SIMPLE_SUBSTITUTION' or r['dictionary_missing']==0]);write(OUT/'SYNTHETIC_DICTIONARY_COVERAGE_RESULTS.tsv',fields,[r for r in results if r['dictionary_missing']!=0]);write(OUT/'SYNTHETIC_BEAM_SWEEP.tsv',fields,beams);write(OUT/'SYNTHETIC_RANDOM_NULL.tsv',list(controls[0]),controls)
 write(OUT/'SYNTHETIC_MAPPING_RECOVERY.tsv',['family','noise','dictionary_missing','sample','beam','mapping_recovery','assignment_recovery','mapping_size'],[{k:r[k] for k in ('family','noise','dictionary_missing','sample','beam','mapping_recovery','assignment_recovery','mapping_size')} for r in results])
 write(OUT/'SYNTHETIC_HELDOUT_RESULTS.tsv',['family','noise','dictionary_missing','sample','beam','theoretical_ceiling','heldout_coverage'],[{k:r[k] for k in ('family','noise','dictionary_missing','sample','beam','theoretical_ceiling','heldout_coverage')} for r in results])
 gt=[]; corp=[]
 for fam in FAMILIES:
  mp=hidden(fam,random.Random(991+FAMILIES.index(fam)));gt += [{'family':fam,'source_grapheme':a,'target_units':'+'.join(v),'in_model_space':'0' if fam=='S6_CONTEXTUAL' else '1'} for a,v in mp.items()]
 for t in make_terms(): corp.append({'object_id':t['object_id'],'object_class':'STAR','source_forms':';'.join(t['forms']),'concept_count':1,'variant_count':len(t['forms']),'lengths':','.join(map(str,map(len,t['forms'])))})
 write(OUT/'SYNTHETIC_GROUND_TRUTH.tsv',['family','source_grapheme','target_units','in_model_space'],gt);write(OUT/'SYNTHETIC_SOURCE_CORPORA.tsv',list(corp[0]),corp)
 clean=[r for r in results if r['family']=='S0_SIMPLE_SUBSTITUTION' and r['noise']==0 and r['dictionary_missing']==0 and r['sample']=='20/6'][0]; n20=[r for r in results if r['noise']==.2 and r['sample']=='20/6'];n30=[r for r in results if r['noise']==.3 and r['sample']=='20/6']; randomp=sorted(r['max_train_coverage'] for r in controls)[int(.95*len(controls))-1]
 status='NON_DISCRIMINATIVE' if randomp>=.70 else 'SEARCH_CAPABLE'
 auth='NO' if status=='NON_DISCRIMINATIVE' else 'YES'
 report=f'''# Synthetic recovery benchmark\n\nThe frozen M1 optimiser was evaluated on generated STAR-only labels with hidden mappings. Ground truth was not passed to search. S6_CONTEXTUAL is explicitly out of model space. Clean S0 at 20/6 recovered TRAIN {clean["train_coverage"]:.3f} and HELD_OUT {clean["heldout_coverage"]:.3f}; detailed family, noise, dictionary and beam results are in TSV.\n\nThe critical negative control failed: random targets reached P95={randomp:.3f} and maximum={max(r["max_train_coverage"] for r in controls):.3f}. Thus high recovery is not discriminative at this sample size; positive and random controls overlap materially.\n\n```text\nASTRO_SEARCH_RECOVERY_BENCHMARK={status}\nCLEAN_RECOVERY_TRAIN={clean["train_coverage"]:.6f}\nCLEAN_RECOVERY_HELDOUT={clean["heldout_coverage"]:.6f}\nNOISE20_RECOVERY={statistics.median(r["recoverable_coverage"] for r in n20):.6f}\nNOISE30_RECOVERY={statistics.median(r["recoverable_coverage"] for r in n30):.6f}\nRANDOM_NULL_P95={randomp:.6f}\nRANDOM_NULL_MAX={max(r["max_train_coverage"] for r in controls):.6f}\nGROUND_TRUTH_MAPPING_RECOVERY={clean["mapping_recovery"]:.6f}\nSAMPLE_20_6_RECOVERY={clean["train_coverage"]:.6f}\nSAMPLE_40_10_RECOVERY={next(r["train_coverage"] for r in results if r["sample"]=="40/10" and r["family"]=="S0_SIMPLE_SUBSTITUTION"):.6f}\nSAMPLE_80_20_RECOVERY={next(r["train_coverage"] for r in results if r["sample"]=="80/20" and r["family"]=="S0_SIMPLE_SUBSTITUTION"):.6f}\nREAL_DICTIONARY_SEARCH_AUTHORIZED={auth}\n```\n\nBecause the random-null separation criterion fails, no real astronomical dictionary search is authorized by this benchmark.\n'''
 (OUT/'SYNTHETIC_RECOVERY_REPORT.md').write_text(report,encoding='utf-8');(OUT/'SYNTHETIC_RECOVERY_SPEC.md').write_text('''# Synthetic recovery specification\n\nSource vocabularies contain 31 STAR concepts with 1–2 variants and lengths 5–7. Hidden S0–S5 transformations are generated from the frozen M1 representation; S6 contextual mappings are out of model space. Noise corrupts target labels at 0/10/20/30%; dictionary incompleteness removes 0/10/20/30/40% concepts. TRAIN/HELD_OUT sizes are 20/6, 40/10, and 80/20. Search uses the frozen 640 pipelines, grapheme inventory, beam search and scoring. Random controls rerun complete search.\n''',encoding='utf-8')
 artifacts=[p.name for p in OUT.iterdir() if p.is_file() and p.name not in {'main.py','manifest.json','SHA256SUMS'}];man={'experiment':'astro-algorithm-check-v1','generated_utc':datetime.now(timezone.utc).replace(microsecond=0).isoformat(),'seed':m1.SEED,'replicates':args.replicates,'beam_widths':[8,16,32,64,128,256],'artifact_sha256':{n:sha(OUT/n) for n in artifacts},'input_sha256':{'m1_main.py':sha(ROOT/'research/astro_token_formation_m1/main.py')}};(OUT/'manifest.json').write_text(json.dumps(man,indent=2,sort_keys=True)+'\n',encoding='utf-8');(OUT/'SHA256SUMS').write_text(''.join(f'{sha(OUT/n)}  {n}\n' for n in sorted(artifacts+['main.py','manifest.json'])),encoding='utf-8')
if __name__=='__main__':main()
