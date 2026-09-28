#!/usr/bin/env python3
"""Run frozen M0/M1 engines on the independently frozen STAR LABEL page scope."""
from pathlib import Path
import csv, importlib.util, shutil, sys, json, hashlib
ROOT=Path(__file__).resolve().parents[3]; OUT=Path(__file__).resolve().parent
BASE=ROOT/'research/astro_hapax_star_label/restricted_m2_star_labels_v1'
def load(name):
 p=ROOT/name; modname='rerun_'+p.stem; s=importlib.util.spec_from_file_location(modname,p); m=importlib.util.module_from_spec(s); sys.modules[modname]=m; s.loader.exec_module(m); return m
m0=load('research/astro_token_formation/main.py'); m1=load('research/astro_token_formation_m1/main.py')
targets=[]
with (BASE/'TARGET_SETS.tsv').open(encoding='utf-8') as f:
 for r in csv.DictReader(f,delimiter='\t'):
  if r['subset']=='ALL_STAR_LABELS' and r['set_id'] in ('F68R1_ALL_STAR_LABELS','F68R2_ALL_STAR_LABELS'): targets.append(r)
def label(r,i):
 return {'stolfi_coordinate':r['occurrence_id'],'object_class':'STAR','panel':r['panel'],'stolfi_group':'L','voynich_token':r['readable_eva'],'absolute_token_position':str(i),'zl3b_locus':r['line_ref']}
def splitfile(path,train,held):
 fields=['stolfi_coordinate','object_class','panel','stolfi_group','voynich_token','absolute_token_position','zl3b_locus','source_record_ids','eligible','sample_rank_sha256','split_rank_sha256','selected','split','exclusion_reason']
 rows=[]
 for i,r in enumerate(train): rows.append({**label(r,i),'source_record_ids':r['occurrence_id'],'eligible':'1','sample_rank_sha256':'RERUN_FROZEN_SCOPE','split_rank_sha256':'PAGE_DIRECTION_A','selected':'1','split':'TRAIN','exclusion_reason':''})
 for i,r in enumerate(held,len(train)): rows.append({**label(r,i),'source_record_ids':r['occurrence_id'],'eligible':'1','sample_rank_sha256':'RERUN_FROZEN_SCOPE','split_rank_sha256':'PAGE_DIRECTION_A','selected':'1','split':'HELD_OUT','exclusion_reason':''})
 with path.open('w',newline='',encoding='utf-8') as f: w=csv.DictWriter(f,fields,delimiter='\t'); w.writeheader();w.writerows(rows)
def run_direction(name,train_panel,held_panel):
 d=OUT/name; d.mkdir(exist_ok=True)
 train=[x for x in targets if x['panel']==train_panel]; held=[x for x in targets if x['panel']==held_panel]
 trainL=[label(x,i) for i,x in enumerate(train)]; heldL=[label(x,i) for i,x in enumerate(held)]
 splitfile(d/'ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv',train,held); shutil.copy(ROOT/'research/astro_token_formation/ASTRO_TERM_CORPUS.tsv',d/'ASTRO_TERM_CORPUS.tsv')
 m0.OUT=d
 terms=m0.load_terms(); models=[]
 for i,rule in enumerate(m0.rule_grid(),1):
  out=m0.transformed(terms,rule); a=m0.max_matching(trainL,terms,out); used=set(a.values()); b=m0.max_matching(heldL,terms,out,used)
  models.append({'model_id':f'M{i:04d}','rule':rule.rule_string,'complexity':rule.complexity,'train_matched':len(a),'train_total':len(train),'train_coverage':len(a)/len(train),'heldout_matched':len(b),'heldout_total':len(held),'heldout_coverage':len(b)/len(held)})
 models.sort(key=lambda x:(-x['train_coverage'],-x['heldout_coverage'],x['complexity'],x['model_id']))
 with (d/'M0_RERUN_MODELS.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(models[0]),delimiter='\t');w.writeheader();w.writerows(models)
 (d/'M0_RERUN_REPORT.md').write_text(f"# M0-R {name}\n\nUnchanged 640-rule M0 grid on page-held-out scope: train={len(train)}, held-out={len(held)}. Best model: {models[0]['train_matched']}/{len(train)} train, {models[0]['heldout_matched']}/{len(held)} held-out. This is an observed rerun; null controls are reported by M1-R.\n")
 # M1 imports the unchanged M0 implementation; only its input/output locations are redirected.
 m1.OUT=d/'M1_RERUN'; m1.OUT.mkdir(exist_ok=True); m1.load_labels=lambda s: (trainL if s=='TRAIN' else heldL)
 m1.load_terms=lambda: m0.load_terms(); m1.NULL_REPLICATES=0
 observed=m1.run_search(trainL,m1.load_terms(),retain_details=True)
 observed.sort(key=lambda x:(-x['train_coverage'],-x['score'],x['mapping_size']))
 top=observed[0]; held_result=m1.compute_heldout(top,heldL,m1.load_terms())
 with (m1.OUT/'M1_RERUN_OBSERVED.tsv').open('w',newline='') as f:
  fields=['direction','train_n','heldout_n','train_matched','train_coverage','heldout_matched','heldout_coverage','mapping_size','total_complexity','score']; w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader();w.writerow({'direction':name,'train_n':len(trainL),'heldout_n':len(heldL),'train_matched':len(top['train_match']),'train_coverage':top['train_coverage'],'heldout_matched':len(held_result['held_match']),'heldout_coverage':held_result['heldout_coverage'],'mapping_size':top['mapping_size'],'total_complexity':top['total_complexity'],'score':top['score']})
 return {'direction':name,'train_panel':train_panel,'heldout_panel':held_panel,'train_n':len(train),'heldout_n':len(held)}
if __name__=='__main__':
 results=[run_direction('RUN_A_F68R1_TO_F68R2','f68r1','f68r2'),run_direction('RUN_B_F68R2_TO_F68R1','f68r2','f68r1')]
 (OUT/'RERUN_DIRECTIONS.json').write_text(json.dumps(results,indent=2)+'\n')
 print(json.dumps(results,indent=2))
(OUT/'RERUN_README.md').write_text('''# Restricted M0-R/M1-R rerun\n\nThis package invokes the unchanged frozen M0 (640-rule grid) and M1 (beam width 64, global 1–2 EVA-unit mapping) implementations. Only the input adapter and output directory differ. Direction A trains on all f68r1 STAR LABEL tokens and holds out all f68r2 tokens; direction B reverses the pages. Observed directional runs completed; expanded null controls are explicitly NOT_RUN pending a dedicated long compute run. The combined run is intentionally not used for model selection. No M2 code is involved.\n''')
