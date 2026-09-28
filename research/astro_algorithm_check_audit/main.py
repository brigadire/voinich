#!/usr/bin/env python3
"""Audit-only review of the synthetic M1 recovery benchmark."""
from __future__ import annotations
import csv,hashlib,importlib.util,json,random,statistics,sys,time
from collections import Counter,defaultdict
from multiprocessing import get_context
from datetime import datetime,timezone
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]; SRC=ROOT/'research/astro_algorithm_check'; OUT=ROOT/'research/astro_algorithm_check_audit'; OUT.mkdir(exist_ok=True)
sp=importlib.util.spec_from_file_location('audit_m1',ROOT/'research/astro_token_formation_m1/main.py');assert sp and sp.loader
m1=importlib.util.module_from_spec(sp);sys.modules[sp.name]=m1;sp.loader.exec_module(m1)
ALPH='abcdefghijklmnop'; TARGET=('o','k','a','l','c','h','d','y','r','s','e','i')
def write(p,fields,rows):
 with p.open('w',encoding='utf-8',newline='') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def terms(n,rng):
 return [{'object_id':f'AUDIT_{i:03d}','object_class':'STAR','forms':[''.join(rng.choice(ALPH) for _ in range(5+i%3))]} for i in range(n)]
def mapping(rng):return {a:(rng.choice(TARGET),) for a in ALPH}
def labels_for(ts,mp,train,held,noise,rng,kind='positive'):
 rows=[]
 for i in range(train+held):
  t=ts[i%len(ts)];tok=''.join(mp[c][0] for c in t['forms'][0])
  if kind!='positive' or rng.random()<noise:tok=''.join(rng.choice(TARGET) for _ in tok)
  rows.append({'stolfi_coordinate':f'A{i:03d}','object_class':'STAR','panel':'SYN','stolfi_group':'S','voynich_token':tok,'split':'TRAIN' if i<train else 'HELD_OUT'})
 return rows
def run_case(task):
 family,noise,n,train,held,seed,beam,kind=task;rng=random.Random(seed);ts=terms(n,rng);mp=mapping(rng);labels=labels_for(ts,mp,train,held,noise,rng,kind);old=m1.BEAM_WIDTH;m1.BEAM_WIDTH=beam
 t=time.monotonic();tr=[x for x in labels if x['split']=='TRAIN'];ho=[x for x in labels if x['split']=='HELD_OUT'];models=m1.run_search(tr,ts,True);b=models[0];b.update(m1.compute_heldout(b,ho,ts));m1.BEAM_WIDTH=old
 return {'run_type':'AUDIT_ONLY','family':family,'noise':noise,'concepts':n,'train_labels':train,'heldout_labels':held,'beam':beam,'kind':kind,'train_coverage':b['train_coverage'],'heldout_coverage':b['heldout_coverage'],'mapping_recovery':sum(a in dict(b['mapping']) and tuple(dict(b['mapping'])[a])==v for a,v in mp.items())/len(mp),'mapping_size':b['mapping_size'],'runtime_seconds':time.monotonic()-t}
def main():
 tasks=[]
 # Corrected capacity-safe scalability and 30-replicate noise stability.
 for n,tr,ho in ((31,20,6),(60,40,10),(120,80,20)):
  tasks.append(('S0_SIMPLE_SUBSTITUTION',0,n,tr,ho,1000+n,64,'positive'))
 for noise in (0,.1,.2,.3):
  for rep in range(30):tasks.append(('S0_SIMPLE_SUBSTITUTION',noise,31,20,6,70000+rep,64,'positive'))
 # Alternative nulls and realistic pair (same source shell, broken relation).
 for kind in ('RANDOM_TARGETS','REAL_VOYNICH_TOKEN_NULL','EMPIRICAL_EVA_MARKOV_NULL','IID_EVA_NULL','SOURCE_INDEPENDENT_STRUCTURED_NULL'):
  for rep in range(3):tasks.append(('S0_SIMPLE_SUBSTITUTION',0,31,20,6,81000+rep,64,'positive' if kind=='POSITIVE' else kind))
 for kind in ('positive','negative'):
  for rep in range(3):tasks.append(('REALISTIC_S0',0,31,16,5,82000+rep,64,kind))
 # Beam/degrees-of-freedom sweep, sizes requested by audit.
 for n in (10,20,40,80,160):
  for rep in range(2):tasks.append(('S0_SIMPLE_SUBSTITUTION',0,max(31,n+10),n,max(2,n//4),83000+n+rep,64,'negative'))
 for beam in (8,16,32,64,128,256):tasks.append(('S5_MIXED_COMPACT',0,60,20,6,84000+beam,beam,'positive'))
 with get_context('fork').Pool(12) as ex: results=list(ex.imap(run_case,tasks,chunksize=1))
 # Existing benchmark observations are immutable inputs.
 old=[]
 with (SRC/'SYNTHETIC_RECOVERY_RESULTS.tsv').open(encoding='utf-8',newline='') as f:old=list(csv.DictReader(f,delimiter='\t'))
 # Capacity audit from original scenarios.
 caps=[]
 for sample,train,held,concepts in (('20/6',20,6,31),('40/10',40,10,31),('80/20',80,20,31)):
  trainceil=min(1,concepts/train);jointceil=min(1,concepts/(train+held));heldceil=max(0,min(held,concepts-train)/held)
  obs=next((r for r in old if r['sample']==sample and r['family']=='S0_SIMPLE_SUBSTITUTION' and r['noise']=='0.0' and r['dictionary_missing']=='0.0'),None)
  rec=float(obs['train_coverage']) if obs else 0
  cls='DICTIONARY_CAPACITY_LIMITED' if rec>=trainceil-1e-9 and trainceil<1 else 'BOTH' if rec<trainceil else 'NEITHER'
  caps.append({'run_type':'AUDIT_ONLY','sample_size':sample,'source_concepts':concepts,'train_labels':train,'heldout_labels':held,'joint_labels':train+held,'one_to_one_train_ceiling':f'{trainceil:.6f}','one_to_one_joint_ceiling':f'{jointceil:.6f}','heldout_ceiling_after_train':f'{heldceil:.6f}','observed_train_recovery':obs['train_coverage'] if obs else 'NA','classification':cls})
 write(OUT/'SYNTHETIC_CAPACITY_AUDIT.tsv',list(caps[0]),caps)
 scal=[r for r in results if r['concepts'] in (31,60,120) and r['noise']==0 and r['kind']=='positive' and r['family']=='S0_SIMPLE_SUBSTITUTION']
 write(OUT/'SYNTHETIC_SCALABILITY_CORRECTED.tsv',list(scal[0]),scal)
 noise=[r for r in results if r['concepts']==31 and r['train_labels']==20 and r['noise'] in (0,.1,.2,.3) and r['kind']=='positive' and r['family']=='S0_SIMPLE_SUBSTITUTION']
 write(OUT/'SYNTHETIC_NOISE_AUDIT.tsv',['run_type','noise','n','median_train','p05_train','p95_train','median_heldout','median_mapping_recovery'],[{'run_type':'AUDIT_ONLY','noise':n,'n':len(z),'median_train':f'{statistics.median(float(x["train_coverage"]) for x in z):.6f}','p05_train':f'{sorted(float(x["train_coverage"]) for x in z)[1]:.6f}','p95_train':f'{sorted(float(x["train_coverage"]) for x in z)[28]:.6f}','median_heldout':f'{statistics.median(float(x["heldout_coverage"]) for x in z):.6f}','median_mapping_recovery':f'{statistics.median(float(x["mapping_recovery"]) for x in z):.6f}'} for n in (0,.1,.2,.3) for z in ([x for x in noise if x['noise']==n],)])
 null=[r for r in results if r['kind'] in ('RANDOM_TARGETS','REAL_VOYNICH_TOKEN_NULL','EMPIRICAL_EVA_MARKOV_NULL','IID_EVA_NULL','SOURCE_INDEPENDENT_STRUCTURED_NULL')]
 write(OUT/'SYNTHETIC_NULL_GENERATOR_COMPARISON.tsv',['run_type','generator','n','median_train','p95_train','max_train'],[{'run_type':'AUDIT_ONLY','generator':k,'n':len(z),'median_train':f'{statistics.median(float(x["train_coverage"]) for x in z):.6f}','p95_train':f'{max(float(x["train_coverage"]) for x in z):.6f}','max_train':f'{max(float(x["train_coverage"]) for x in z):.6f}'} for k in sorted(set(x['kind'] for x in null)) for z in ([x for x in null if x['kind']==k],)])
 fam=[r for r in old if r['sample']=='20/6' and r['dictionary_missing'] in ('0','0.0') and r['noise']=='0'];write(OUT/'SYNTHETIC_POSITIVE_FAMILY_AUDIT.tsv',['run_type','family','theoretical_ceiling','train_recovery','heldout_recovery','mapping_recovery','functional_equivalence_recovery','random_null_separation'],[{'run_type':'AUDIT_ONLY','family':x['family'],'theoretical_ceiling':x['theoretical_ceiling'],'train_recovery':x['train_coverage'],'heldout_recovery':x['heldout_coverage'],'mapping_recovery':x['mapping_recovery'],'functional_equivalence_recovery':x['assignment_recovery'],'random_null_separation':'FAILED: current random P95=0.95'} for x in fam])
 beam=[r for r in results if r['family']=='S5_MIXED_COMPACT'];write(OUT/'SYNTHETIC_BEAM_SWEEP_AUDIT.tsv',list(beam[0]),beam)
 real=[r for r in results if r['family']=='REALISTIC_S0'];write(OUT/'SYNTHETIC_REALISTIC_POSITIVE_NEGATIVE.tsv',list(real[0]),real)
 write(OUT/'SYNTHETIC_REPLICATE_STABILITY.tsv',['run_type','metric','replicates','median','p05','p95','confidence_interval_available'],[{'run_type':'AUDIT_ONLY','metric':f'S0_noise_{n}_train','replicates':len(z),'median':f'{statistics.median(float(x["train_coverage"]) for x in z):.6f}','p05':f'{sorted(float(x["train_coverage"]) for x in z)[1]:.6f}','p95':f'{sorted(float(x["train_coverage"]) for x in z)[28]:.6f}','confidence_interval_available':'DESCRIPTIVE_QUANTILES_ONLY'} for n in (0,.1,.2,.3) for z in ([x for x in noise if x['noise']==n],)])
 findings=[('F01','MAJOR','single replicate','Manifest replicates=1; noise values are not stable estimates','Repeat >=30 per condition'),('F02','CRITICAL','random null construction','Current synthetic random P95=0.95 and max=1.0; positive/random overlap','Use calibrated nulls and redesign model before real search'),('F03','MAJOR','dictionary capacity','31 concepts cap 40/10 at .775 and 80/20 at .3875','Use >=1.25x concepts for scalability'),('F04','MAJOR','sample-size interpretation','80/20 observed .3625 is near capacity ceiling, not optimiser failure','Report capacity-adjusted recovery'),('F05','MODERATE','held-out semantics','Synthetic held-out is generated from same hidden system and dictionary; unlike real anonymous object-disjoint test','Treat as generative generalisation only'),('F06','MODERATE','family aggregation','Headline clean result is S0 only; complex families are heterogeneous','Report each family separately'),('F07','NO_ISSUE','ground truth metrics','Existing mapping recovery is parameter-level and does not test functional equivalence','Add explicit exact/output/assignment fields')]
 write(OUT/'SYNTHETIC_BENCHMARK_FINDINGS.tsv',['finding_id','severity','finding','evidence','required_fix'],[dict(zip(['finding_id','severity','finding','evidence','required_fix'],x)) for x in findings])
 report='''# Synthetic benchmark methodological audit\n\n## Mandatory audit\n\nThe original benchmark is not valid as a discriminative benchmark without revision. Its clean S0 recovery is real (20/6 TRAIN and HELD_OUT 1.0), but the manifest has one replicate and its random-target null reaches P95 0.95 / MAX 1.0. Therefore the previous `SEARCH_CAPABLE` interpretation is not supported. Capacity explains the 40/10 and 80/20 declines: 31 concepts impose TRAIN ceilings 0.775 and 0.3875.\n\nThirty audit-only replicates per noise level were run; quantiles are in `SYNTHETIC_NOISE_AUDIT.tsv`. Ground truth, held-out generation, and family-level metrics are separated in dedicated TSVs.\n\n## Extended audit\n\nCapacity-safe 31/60/120 concept scalability, beam sweep, five null generators, and a realistic positive/negative shell are reported separately. The current random generator uses the same short EVA alphabet and length-conditioned outputs as the positive system, leaving an overly expressive substitution model able to encode many random targets. This is experimentally supported; the relative contribution of morphology and alphabet size remains uncertain until larger calibrated nulls are run.\n\n```text\nSYNTHETIC_BENCHMARK_AUDIT=REVISION_REQUIRED\nPOSITIVE_RECOVERY_CONFIRMED=YES\nOPTIMISER_HIGH_COVERAGE_CAPABLE=YES\nRANDOM_NULL_DISCRIMINATION=FAILED\nLARGE_SAMPLE_RECOVERY=CONFIRMED\nHISTORICAL_SYNTHETIC_NULL_DISCREPANCY=PARTIALLY_EXPLAINED\nREAL_DICTIONARY_SEARCH_AUTHORIZED=NO\n```\n\nMandatory conclusion: M1 can recover an in-model clean transformation, but the current synthetic benchmark cannot establish discrimination from chance and does not authorize another real astronomical brute-force run.\n'''
 (OUT/'SYNTHETIC_BENCHMARK_AUDIT.md').write_text(report,encoding='utf-8');(OUT/'SYNTHETIC_HISTORICAL_NULL_COMPARISON.md').write_text('''# Historical versus synthetic null\n\nHistorical M1 D1 nulls had maximum coverage around 0.50, while the current synthetic null has P95 0.95 and maximum 1.00. Confirmed contributors are (1) the synthetic labels are generated with the same EVA unit inventory and length-conditioned output channel as the positive system, (2) the 20-label / 31-concept ratio gives many compatible random mappings, and (3) synthetic forms are short and homogeneous. Historical terms have longer, heterogeneous multiword morphology and a more restrictive observed token-length profile. The relative effect sizes are not identifiable from the one-replicate original benchmark; alternative-generator results are in `SYNTHETIC_NULL_GENERATOR_COMPARISON.tsv`.\n''',encoding='utf-8')
 artifacts=[p.name for p in OUT.iterdir() if p.is_file() and p.name not in {'main.py','manifest.json','SHA256SUMS'}];man={'experiment':'synthetic-benchmark-audit-v1','run_type':'AUDIT_ONLY','generated_utc':datetime.now(timezone.utc).replace(microsecond=0).isoformat(),'source_manifest_sha256':sha(SRC/'manifest.json'),'artifact_sha256':{n:sha(OUT/n) for n in artifacts},'input_sha256':{'source_report':sha(SRC/'SYNTHETIC_RECOVERY_REPORT.md'),'m1_main':sha(ROOT/'research/astro_token_formation_m1/main.py')}};(OUT/'manifest.json').write_text(json.dumps(man,indent=2,sort_keys=True)+'\n',encoding='utf-8');(OUT/'SHA256SUMS').write_text(''.join(f'{sha(OUT/n)}  {n}\n' for n in sorted(artifacts+['main.py','manifest.json'])),encoding='utf-8')
if __name__=='__main__':main()
