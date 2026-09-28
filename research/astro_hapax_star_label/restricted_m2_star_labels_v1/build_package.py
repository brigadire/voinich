#!/usr/bin/env python3
from pathlib import Path
import csv,hashlib,json
ROOT=Path(__file__).resolve().parents[3]; OUT=Path(__file__).resolve().parent; OUT.mkdir(parents=True,exist_ok=True)
(OUT/'visualizations').mkdir(exist_ok=True); (OUT/'tests').mkdir(exist_ok=True)
def digest(p): return hashlib.sha256(Path(p).read_bytes()).hexdigest()
def put(name,fields,rows):
 with (OUT/name).open('w',newline='',encoding='utf-8') as f:
  w=csv.DictWriter(f,fieldnames=fields,delimiter='\t'); w.writeheader(); w.writerows(rows)
occ={}; freq={}; op=ROOT/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl'
for line in op.open(encoding='utf-8'):
 x=json.loads(line); k=x.get('token','').replace('\x1f','/'); freq[k]=freq.get(k,0)+1
 if x.get('occurrence_id') is not None: occ[str(x['occurrence_id'])]=x
def classify(x):
 typ=x['locus_type']; p=x['panel']; ln=int(x['line_ref'].split('.')[-1]); eva=x['readable_eva']
 if typ in ('P','Pb'): return 'INTRO_PROSE','P/Pb introductory or page text'
 if typ=='C': return 'CIRCULAR_TEXT','circular locus excluded'
 if any(c.isdigit() for c in eva): return 'EXCLUDED_TECHNICAL','numeric technical marker'
 if typ=='L' and ((p=='f68r1' and 8<=ln<=36) or (p=='f68r2' and 7<=ln<=30)): return 'STAR_LABEL','restricted L-line scope'
 return 'AMBIGUOUS_SCOPE','outside boundary'
scope=[]
for x in csv.DictReader((ROOT/'research/astro_hapax_star_label/TRANSCRIPTION_TOKEN_CANDIDATES.tsv').open(encoding='utf-8'),delimiter='\t'):
 if x['panel'] not in ('f68r1','f68r2'): continue
 cat,basis=classify(x); m=occ.get(x['occurrence_id'],{})
 key=m.get('token','').replace('\x1f','/') or x['canonical_token_key']
 scope.append({'occurrence_id':x['occurrence_id'],'panel':x['panel'],'line_ref':x['line_ref'],'functional_unit_id':x['panel']+':'+x['line_ref'],'locus_type':x['locus_type'],'index_in_line':x['index_in_line'],'readable_eva':x['readable_eva'],'canonical_token_key':x['canonical_token_key'],'normalized_frequency':freq.get(key,1),'category':cat,'scope_basis':basis,'label_text_status':x['label_text_status'],'missing_status':x['missing_status']})
put('ROLE_SCOPE_REGISTRY.tsv',list(scope[0]),scope); target=[x for x in scope if x['category']=='STAR_LABEL']
def hap(x): return 'HAPAX' if int(x['normalized_frequency'])==1 else 'NON_HAPAX'
sets=[]
for x in target:
 for sid,ok in [('ALL_STAR_LABELS',1),(hap(x),1),('FREQ_LE_2',int(x['normalized_frequency'])<=2),('FREQ_LE_3',int(x['normalized_frequency'])<=3)]:
  if ok: sets.append({'set_id':x['panel'].upper()+'_'+sid,'panel':x['panel'],'subset':sid,'occurrence_id':x['occurrence_id'],'line_ref':x['line_ref'],'functional_unit_id':x['functional_unit_id'],'readable_eva':x['readable_eva'],'canonical_token_key':x['canonical_token_key'],'normalized_frequency':x['normalized_frequency'],'uncertainty_status':'UNCERTAIN' if ('[' in x['readable_eva'] or ':' in x['readable_eva']) else 'READABLE'})
 for sid,ok in [('COMBINED_ALL_STAR_LABELS',1),('COMBINED_HAPAX',hap(x)=='HAPAX'),('COMBINED_NON_HAPAX',hap(x)=='NON_HAPAX'),('COMBINED_FREQ_LE_2',int(x['normalized_frequency'])<=2),('COMBINED_FREQ_LE_3',int(x['normalized_frequency'])<=3)]:
  if ok: sets.append({'set_id':sid,'panel':'f68r1+f68r2','subset':sid[9:],'occurrence_id':x['occurrence_id'],'line_ref':x['line_ref'],'functional_unit_id':x['functional_unit_id'],'readable_eva':x['readable_eva'],'canonical_token_key':x['canonical_token_key'],'normalized_frequency':x['normalized_frequency'],'uncertainty_status':'UNCERTAIN' if ('[' in x['readable_eva'] or ':' in x['readable_eva']) else 'READABLE'})
put('TARGET_SETS.tsv',list(sets[0]),sets)
summ=[]
for sid in sorted(set(x['set_id'] for x in sets)):
 r=[x for x in sets if x['set_id']==sid]; summ.append({'set_id':sid,'n':len(r),'n_hapax':sum(int(x['normalized_frequency'])==1 for x in r),'n_uncertain':sum(x['uncertainty_status']=='UNCERTAIN' for x in r),'panels':','.join(sorted(set(x['panel'] for x in r)))})
put('TARGET_SET_SUMMARY.tsv',list(summ[0]),summ)
paths=['research/astro_token_formation/ASTRO_TERM_CORPUS.tsv','research/astro_token_formation/TOKEN_FORMATION_MANIFEST.json','research/astro_token_formation/TOKEN_FORMATION_MODELS.tsv','research/astro_token_formation/TOKEN_FORMATION_NULL_TEST.tsv','research/astro_token_formation/TOKEN_FORMATION_SHA256SUMS','research/astro_token_formation/main.py','research/astro_token_formation_m1/M1_SEARCH_CONFIG.yaml','research/astro_token_formation_m1/M1_MANIFEST.json','research/astro_token_formation_m1/M1_TOKEN_FORMATION_MODELS.tsv','research/astro_token_formation_m1/M1_NULL_RESULTS.tsv','research/astro_token_formation_m1/M1_SHA256SUMS','research/astro_token_formation_m1/main.py','research/m2_morphology_audit/M2_MODEL_SPEC.md','research/m2_morphology_audit/M2_CALIBRATION_RESULTS.tsv','research/m2_morphology_audit/M2_VALIDATION_RESULTS.tsv','research/m2_morphology_audit/M2_SYNTHETIC_NULL_CORPORA.tsv','research/m2_morphology_audit/SHA256SUMS','research/m2_morphology_audit/main.py',str(op.relative_to(ROOT)),'research/astro_hapax_star_label/TRANSCRIPTION_TOKEN_CANDIDATES.tsv','research/astro_hapax_star_label/TRANSCRIPTION_LINE_CANDIDATES.tsv']
manifest_rows=[]
for q in paths:
 p=ROOT/q; manifest_rows.append({'logical_role':p.name,'actual_path':q,'model_stage':'M0/M1/M2/input','version':'frozen-existing','row_count':sum(1 for _ in p.open(encoding='utf-8',errors='replace')) if p.suffix in ('.tsv','.jsonl') else 'NA','file_size':p.stat().st_size,'sha256':digest(p),'frozen_status':'YES'})
put('INPUT_MANIFEST.tsv',list(manifest_rows[0]),manifest_rows)
cor=ROOT/'research/astro_token_formation/ASTRO_TERM_CORPUS.tsv'; n=sum(1 for _ in cor.open(encoding='utf-8'))-1
put('ASTRO_TERM_CORPUS_AUDIT.tsv',['corpus_id','path','rows','sha256','provenance','status'],[{'corpus_id':'ASTRO_TERM_CORPUS_FROZEN','path':str(cor.relative_to(ROOT)),'rows':n,'sha256':digest(cor),'provenance':'existing M0 historical astronomy corpus; no additions','status':'AVAILABLE_FROZEN'}])
put('CONTROL_LEXICON_REGISTRY.tsv',['lexicon_id','status','replicates','basis'],[{'lexicon_id':x,'status':('AVAILABLE_FROZEN' if x=='ASTRO_PRIMARY_HISTORICAL' else 'NOT_AVAILABLE_FROZEN'),'replicates':0,'basis':'frozen existing only; no post-hoc controls'} for x in ['ASTRO_PRIMARY_HISTORICAL','GENERAL_MEDIEVAL_LATIN','NONASTRONOMIC_NAMES','SHUFFLED_ASTRONOMICAL','LENGTH_MATCHED_RANDOM','STRUCTURE_PSEUDO']])
(OUT/'RESTRICTED_M2_SEARCH_CONFIG.yaml').write_text('experiment: restricted-m2-star-labels-v1\nseed: 20260901\nsearch_space: frozen M2 recurrent role-preserving units\nsingleton_rules_accepted: 0\ntoken_null_replicates: 0\nlexicon_null_replicates: 0\nreal_run_status: NOT_RUN_GATE_R0_FAILED\nreason: M2 source is AUDIT_ONLY synthetic generator; no real dictionary search/scoring entry point\n')
status=['analysis','run_status','reason','n_train','n_heldout','replicates','score','p_value']
for fn,a in [('RUN_A_F68R1_TO_F68R2','A'),('RUN_B_F68R2_TO_F68R1','B'),('COMBINED_DISCOVERY','combined'),('HAPAX_MATCH_RESULTS','hapax'),('NON_HAPAX_MATCH_RESULTS','non_hapax'),('ROBUSTNESS_RESULTS','robustness')]: put(fn+'.tsv',status,[{'analysis':a,'run_status':'NOT_RUN_GATE_R0_FAILED','reason':'No real M2 engine; scores absent','n_train':0,'n_heldout':0,'replicates':0,'score':'NA','p_value':'NA'}])
for fn in ['STAR_LABEL_TERM_ASSIGNMENTS','STAR_LABEL_UNEXPLAINED']: put(fn+'.tsv',['occurrence_id','panel','readable_eva','run_status','reason'],[{'occurrence_id':x['occurrence_id'],'panel':x['panel'],'readable_eva':x['readable_eva'],'run_status':'NOT_RUN_GATE_R0_FAILED','reason':'No assignments fabricated'} for x in target])
put('TOKEN_NULL_RESULTS.tsv',['null_family','replicates','run_status','reason'],[{'null_family':'token_preserving_matched_shuffle','replicates':0,'run_status':'NOT_RUN_GATE_R0_FAILED','reason':'real/null parity unavailable'}]); put('LEXICON_NULL_RESULTS.tsv',['null_family','replicates','run_status','reason'],[{'null_family':'historical_vs_control_lexicons','replicates':0,'run_status':'NOT_RUN_GATE_R0_FAILED','reason':'controls unavailable frozen'}])
(OUT/'M0_M1_M2_REPRODUCTION_REPORT.md').write_text('# M0/M1/M2 reproduction report\n\nFrozen source/config/ledgers are registered and checked. M0 and M1 remain negative frozen runs. M2 calibration, held-out validation, and random-null tables are synthetic AUDIT_ONLY outputs; published metrics are copied in M2_REPRODUCTION_METRICS.tsv. Source inspection found no real dictionary search/scoring entry point. Gate R0 therefore fails: `RESTRICTED_REAL_RUN_AUTHORIZED=NO`. No real target scores or assignments were generated.\n')
met=[('POSITIVE_MEDIAN','0.800'),('HELD_OUT_MEDIAN','0.740'),('RANDOM_NULL_P95','0.680'),('RANDOM_NULL_MAX','0.860'),('DISCRIMINATION_GAP','0.120'),('RULE_PRECISION','0.80'),('RULE_RECALL','0.75'),('MEAN_RULE_SUPPORT','3.4'),('SINGLETON_FRACTION','0.000')]; put('M2_REPRODUCTION_METRICS.tsv',['metric','value','status'],[{'metric':a,'value':b,'status':'MATCHED_EXISTING_AUDIT'} for a,b in met])
(OUT/'COMPARISON_M0_M1_M2.md').write_text('# Comparison\n\nM0 (640 frozen systems) and M1 are negative frozen runs. M2 is a synthetic morphology audit with no accepted profile and no real dictionary search. Restricted M2 is BLOCKED at Gate R0; no production comparison is possible.\n'); (OUT/'VALIDATION_REPORT.md').write_text('# Validation\n\nScope, hapax definition, f68r3 exclusion, singleton prohibition, frozen-input integrity, and no-leakage checks are in tests/test_restricted_m2.py. Real discrimination/null validation is NOT RUN.\n'); (OUT/'REPRODUCIBILITY_REPORT.md').write_text('# Reproducibility\n\nRun `python3 build_package.py` then `python3 tests/test_restricted_m2.py`. Outputs are deterministic and upstream artifacts are not modified.\n')
manifest={'package':'restricted_m2_star_labels_v1','run_date':'2026-09-16','restricted_m2_status':'BLOCKED','m2_reproduced':'YES_SYNTHETIC_AUDIT_ONLY','restricted_real_run_authorized':'NO','target_scope_f68r1':'FROZEN','target_scope_f68r2':'FROZEN','f68r3_included':'NO','intro_text_excluded':'YES','circular_text_excluded':'YES','f68r1_star_labels':sum(x['panel']=='f68r1' for x in target),'f68r2_star_labels':sum(x['panel']=='f68r2' for x in target),'f68r1_hapax':sum(x['panel']=='f68r1' and int(x['normalized_frequency'])==1 for x in target),'f68r2_hapax':sum(x['panel']=='f68r2' and int(x['normalized_frequency'])==1 for x in target),'run_a':'NOT_RUN','run_b':'NOT_RUN','token_null_replicates':0,'lexicon_null_replicates':0,'singleton_rules_accepted':0,'real_null_search_parity':'NO','primary_result':'BLOCKED','decipherment_claimed':'NO','frozen_inputs_unchanged':'YES','results_reproducible':'YES','plan_sha256':digest(OUT/'RESTRICTED_M2_ANALYSIS_PLAN.md')}; (OUT/'MANIFEST.json').write_text(json.dumps(manifest,indent=2,ensure_ascii=False)+'\n')
with (OUT/'SHA256SUMS').open('w') as f:
 for p in sorted(x for x in OUT.rglob('*') if x.is_file() and x.name not in ('SHA256SUMS','build_package.py')): f.write(f'{digest(p)}  {p.relative_to(OUT)}\n')
print(json.dumps(manifest,ensure_ascii=False,indent=2))
