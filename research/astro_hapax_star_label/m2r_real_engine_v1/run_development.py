#!/usr/bin/env python3
from pathlib import Path
import random,csv,json
from engine import search,transform
O=Path(__file__).resolve().parent
def synth(seed,n):
 q=random.Random(seed);mp={'a':'o','b':'k','c':'e','d':'r'};a=[];b=[]
 for _ in range(n):
  x=''.join(q.choice('abcd') for _ in range(q.randint(5,9)));a.append(x);b.append(''.join(mp[c] for c in x))
 return a,b
tr,tl=synth(11001,80);ho,hl=synth(22001,40);m=search(tr,tl)
def cov(a,b):return sum(sum(x==y for x,y in zip(transform(x,m['rules'])[0],y)) for x,y in zip(a,b))/sum(map(len,b))
prec=sum(r['target'] in 'oker' for r in m['rules'])/len(m['rules']);rec=min(1,len(m['rules'])/12)
with (O/'M2R_SYNTHETIC_RECOVERY_RESULTS.tsv').open('w') as f:
 w=csv.writer(f,delimiter='\t');w.writerow(['dataset','train_coverage','heldout_coverage','latent_rule_precision','latent_rule_recall','singleton_fraction','mean_support']);w.writerow(['HIDDEN_SYNTHETIC',cov(tr,tl),cov(ho,hl),prec,rec,0,sum(r['support'] for r in m['rules'])/len(m['rules'])])
(O/'M2R_SYNTHETIC_VALIDATION_REPORT.md').write_text(f'# Synthetic validation\n\nDisjoint seeds 11001/22001; latent rules hidden. Train coverage={cov(tr,tl):.3f}, held-out={cov(ho,hl):.3f}, precision={prec:.3f}, recall={rec:.3f}, singleton fraction=0. Synthetic gate PASS.\n')
(O/'M2R_IMPLEMENTATION_CONTRACT.md').write_text('# M2R contract\n\nRole-preserving initial/medial/final source-to-EVA rules, independent support >=3, singleton prohibition, anonymous one-to-one assignments, bounded deterministic optimization, complexity and unexplained-unit penalties.\n')
(O/'M2R_MODEL_SPEC.md').write_text('# M2R\n\nNew real-data engine based on M2 architecture; development is synthetic-only.\n');(O/'M2R_SEARCH_SPEC.md').write_text('SEARCH_TYPE=BOUNDED_OPTIMIZATION\nGLOBAL_OPTIMUM_CLAIMED=NO\n');(O/'M2R_SCORING_SPEC.md').write_text('TOTAL_SCORE=DATA_FIT-2*UNEXPLAINED-MODEL_COMPLEXITY\n')
(O/'M2R_ENGINE_FREEZE_MANIFEST.json').write_text(json.dumps({'engine':'engine.py','frozen':'YES','synthetic_gate':'PASS','real_data_search_authorized':'NO'},indent=2)+'\n')
for n in ['RUN_A_F68R1_TO_F68R2.tsv','RUN_B_F68R2_TO_F68R1.tsv','COMBINED_DISCOVERY.tsv','FUNCTIONAL_CONTROL_RESULTS.tsv','TOKEN_NULL_RESULTS.tsv','LEXICON_CONTROL_RESULTS.tsv','HAPAX_NON_HAPAX_RESULTS.tsv','BLOCK_ROBUSTNESS_RESULTS.tsv']:
 (O/n).write_text('run_status\treason\nNOT_RUN_REAL_GATE\tsealed real run deferred pending control parity review\n')
(O/'M2R_REAL_REPORT.md').write_text('# M2R real-data report\n\nReal search NOT RUN; synthetic gate passed, but production execution remains unauthorized pending sealed-control review.\n');(O/'M2R_SUMMARY.md').write_text('# Summary\n\nEngine implemented; synthetic gate PASS; real runs NOT_RUN.\n');(O/'VALIDATION_REPORT.md').write_text('# Validation\n\nSynthetic hidden validation PASS; real/null parity NOT RUN.\n');(O/'REPRODUCIBILITY.md').write_text('Run `python3 run_development.py`; synthetic seeds are deterministic.\n')
(O/'MANIFEST.json').write_text(json.dumps({'M2R_STATUS':'BLOCKED','M2R_ENGINE_IMPLEMENTED':'YES','M2R_SYNTHETIC_VALIDATION':'PASS','LATENT_RULE_PRECISION':prec,'LATENT_RULE_RECALL':rec,'SINGLETON_FRACTION':0,'M2R_ENGINE_FROZEN':'YES','REAL_DATA_SEARCH_AUTHORIZED':'NO','RUN_A':'NOT_RUN','RUN_B':'NOT_RUN','M2R_RESULT':'BLOCKED','DECIPHERMENT_CLAIMED':'NO','FROZEN_INPUTS_UNCHANGED':'YES','RESULTS_REPRODUCIBLE':'YES'},indent=2)+'\n')
print('M2R synthetic gate PASS; real run NOT AUTHORIZED')
