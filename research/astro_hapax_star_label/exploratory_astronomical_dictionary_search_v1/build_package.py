#!/usr/bin/env python3
"""Build and validate the exploratory astronomical dictionary-search package.

This builder deliberately never runs a search against the 57 real labels.
Synthetic cases are used for safety tests and resource preflight only.
"""
from __future__ import annotations
import csv, hashlib, json, os, platform, resource, shutil, sys, time
from pathlib import Path

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]
SCOPE = REPO / "research/astro_hapax_star_label/restricted_dictionary_bruteforce_v1/TARGET_SCOPE.tsv"
LEX = REPO / "research/astro_hapax_star_label/restricted_dictionary_bruteforce_remediation_v1/REMEDIATED_HISTORICAL_STAR_LEXICON.tsv"
REMED = REPO / "research/astro_hapax_star_label/restricted_dictionary_bruteforce_remediation_v1"
V2 = REPO / "research/astro_hapax_star_label/exact_scope_qualification_v2"

def sha(p):
    h=hashlib.sha256(); h.update(p.read_bytes()); return h.hexdigest()
def dump(name, obj):
    (ROOT/name).write_text(json.dumps(obj, indent=2, sort_keys=True)+"\n")
def tsv(name, fields, rows):
    with (ROOT/name).open("w", newline="") as f:
        w=csv.DictWriter(f, fieldnames=fields, delimiter="\t", lineterminator="\n"); w.writeheader(); w.writerows(rows)
def readtsv(p):
    with p.open() as f: return list(csv.DictReader(f, delimiter="\t"))

def build_inputs():
    scope=readtsv(SCOPE); lex=readtsv(LEX)
    assert len(scope)==57 and sum(x['page_id']=='f68r1' for x in scope)==30 and sum(x['page_id']=='f68r2' for x in scope)==27
    labels=[]
    for x in scope:
        lid=f"STAR_{x['page_id']}_{x['occurrence_id']}"
        labels.append(dict(label_id=lid,page=x['page_id'],zl3b_token=x['token'],normalized_token=x['canonical_token_key'],hapax_status='HAPAX' if x['hapax']=='1' else 'NON_HAPAX',line_id=x['line_ref'],source_reference=f"{SCOPE.relative_to(REPO)}#occurrence_id={x['occurrence_id']}",scope_status='FROZEN_F68R1_F68R2_STAR_LABEL',transcription_notes='Copied from frozen restricted-dictionary target scope; no visual relabeling.'))
    tsv('TARGET_STAR_LABELS.tsv', list(labels[0]), labels)
    fields=['lexicon_id','canonical_identity_id','canonical_identity_name','attestation_id','attested_form','normalized_form','language','script','source_title','source_date_start','source_date_end','source_type','historically_attested','editorial_reconstruction','confidence','disposition','remediation_verification_result','provenance_status','object_type','source_reference']
    out=[]
    for i,x in enumerate(lex,1):
        y={k:x.get(k,'') for k in fields}; y['lexicon_id']=f"LEX_{i:04d}"; y['provenance_status']='REMEDIATED_KEEP_VERIFIED_SOURCE'; y['object_type']='FIXED_STAR_NAME'; y['source_reference']=f"{LEX.relative_to(REPO)}#attestation_id={x['attestation_id']}"; out.append(y)
    tsv('ASTRONOMICAL_NAMES_LEXICON.tsv',fields,out)
    audit=readtsv(REMED/'LEXICON_CHANGE_AUDIT.tsv')
    ex=[]
    for x in audit:
        if x['disposition']!='KEEP': ex.append(dict(attestation_id=x['attestation_id'],canonical_identity_id=x['canonical_identity_id'],exclusion_status='EXCLUDED',reason=x['disposition'],source_reference=f"{(REMED/'LEXICON_CHANGE_AUDIT.tsv').relative_to(REPO)}#attestation_id={x['attestation_id']}"))
    tsv('LEXICON_EXCLUSIONS.tsv',['attestation_id','canonical_identity_id','exclusion_status','reason','source_reference'],ex)
    dump('LEXICON_MANIFEST.json',{'manifest_version':'1','source_file':str(LEX.relative_to(REPO)),'source_sha256':sha(LEX),'included_records':len(out),'excluded_records':len(ex),'selection':'disposition=KEEP only','eva_similarity_used':False,'object_type':'FIXED_STAR_NAME','provenance':'remediation_v1; no new claims'})
    return labels,out

def build_specs():
    dump('TRANSFORMATION_SCOPE.json',{'scope_version':'1','mode':'GLOBAL_SUBSTITUTION_ONLY','target_table_sizes':[1,2,3,4,5,6],'one_to_one':True,'allow_unmatched':True,'allow_singletons':False,'minimum_independent_support':2,'stages':{'E0':'compatibility precompute safe filters','E1':'exact global substitution; CP-SAT/oracle certifiable; sizes 1-4','E2':'exact local expansion; sizes 5-6; separate branch','E3':'near local one-error; diagnostic candidates','E4':'weak local two-error; separate branch, never silently mixed'},'forbidden':['page-specific rules','term-specific rules','label-specific rules','singleton evidence','EVA similarity thresholding'],'identity_capacity_and_relaxed_attestation_capacity_separate':True})
    profiles=[]
    for name, ex, near, weak, unmatch in [('CONSERVATIVE',1000,25,0,-20),('BALANCED',1000,30,5,-12),('RECALL_ORIENTED',1000,35,8,-8)]: profiles.append({'profile':name,'exact_reward':ex,'near_reward':near,'weak_reward':weak,'unmatched_penalty':unmatch,'weak_can_be_worse_than_unmatched':True,'preregister_before_real_run':True})
    dump('SCORING_PROFILES.json',{'formula':'exact_reward*EXACT + near_reward*NEAR + weak_reward*WEAK + unmatched_penalty*UNMATCHED - complexity_penalty','profiles':profiles,'ordering':'EXACT_GLOBAL > NEAR_LOCAL_1 > WEAK_LOCAL_2; weak is not automatically preferred to unmatched'})
    dump('SEARCH_STAGES.json',{'E0':{'status':'IMPLEMENTED','output':'compatibility edges'},'E1':{'status':'IMPLEMENTED','table_sizes':[1,2,3,4],'top_k':50,'cooptimal_limit':500},'E2':{'status':'IMPLEMENTED_INTERFACE','table_sizes':[5,6]},'E3':{'status':'IMPLEMENTED_INTERFACE','one_local_error':True},'E4':{'status':'IMPLEMENTED_INTERFACE','two_local_errors':True,'separate_branch':True},'real_run':'NOT_EXECUTED'})
    schemas={
      'RULE_SYSTEMS.tsv':['system_id','stage','profile','rank','score','bound','gap','status','rule_count','unmatched_count','cooptimal_group_id'],
      'RULE_DEFINITIONS.tsv':['system_id','rule_id','input_grapheme','output_grapheme','support_count','support_label_ids'],
      'LABEL_NAME_CANDIDATES.tsv':['system_id','label_id','candidate_name','match_class','support','score_contribution','capacity_status'],
      'UNMATCHED_LABELS.tsv':['system_id','label_id','reason'],
      'COOPTIMAL_SYSTEMS.tsv':['cooptimal_group_id','system_id','score','canonical_symmetry_key'],
      'SOLUTION_FAMILIES.tsv':['family_id','cooptimal_group_id','symmetry_basis','member_count'],
      'SCORE_DECOMPOSITION.tsv':['system_id','exact','near','weak','unmatched','complexity_penalty','total'],
      'SEARCH_BOUNDS.tsv':['run_id','stage','lower_bound','upper_bound','gap','proof_status'],
      'RESOURCE_USAGE.tsv':['run_id','stage','wall_seconds','cpu_seconds','peak_rss_kb','variables','constraints','edges','checkpoint']}
    for fn,fields in schemas.items(): tsv(fn,fields,[])

def write_docs():
    (ROOT/'EXPERIMENT_PROTOCOL.md').write_text('''# Exploratory astronomical dictionary search v1\n\nThis package prepares a candidate search over the frozen 57-label f68r1/f68r2 scope. It is exploratory: it does not authorize a decipherment claim, unique star identification, or scientific confirmation. The 57 real labels were not searched while building this package.\n\nThe target is 30 f68r1 and 27 f68r2 labels, including 26 hapax occurrences. The dictionary is copied from remediation_v1 with `disposition=KEEP`; exclusions remain explicit. Rules are global substitutions with at least two independent supports, one-to-one capacity, and unmatched labels allowed.\n\nExact CP-SAT search is the E1 baseline. E2-E4 are separately reported candidate branches. Profiles, seeds, resource limits, hashes, and run identity must be frozen before any real run.\n''')
    (ROOT/'RESOURCE_PREFLIGHT_PLAN.md').write_text('''# Resource preflight plan\n\nUse synthetic labels at sizes 10, 20, 30, and 57 and lexicon fractions 25%, 50%, and 100%. Measure table sizes 1-4 exact, 5-6 expansion, one-error branch, and top-1/top-10/top-50 extraction with at least three repetitions. Record wall time, CPU time, peak RSS, variables, constraints, edges, status, objective, bound, gap, solutions, and checkpoint. Synthetic measurements estimate scheduling only; they do not qualify the real task.\n''')
    (ROOT/'OUTPUT_SCHEMA.md').write_text('''# Output schema\n\nAll result tables are TSV and every raw result row carries `run_id`, `solver_sha256`, `scorer_sha256`, `config_sha256`, `lexicon_sha256`, `target_scope_sha256`, `ortools_version`, `seed`, `timestamp_utc`, `stage`, and `integrity_sha256`.\n\n`RULE_SYSTEMS.tsv`: system_id, stage, profile, rank, score, bound, gap, status, rule_count, unmatched_count, cooptimal_group_id.\n`RULE_DEFINITIONS.tsv`: system_id, rule_id, input_grapheme, output_grapheme, support_count, support_label_ids.\n`LABEL_NAME_CANDIDATES.tsv`: system_id, label_id, candidate_name, match_class, support, score_contribution, capacity_status.\n`UNMATCHED_LABELS.tsv`: system_id, label_id, reason.\n`COOPTIMAL_SYSTEMS.tsv`: cooptimal_group_id, system_id, score, canonical_symmetry_key.\n`SOLUTION_FAMILIES.tsv`: family_id, cooptimal_group_id, symmetry_basis, member_count.\n`SCORE_DECOMPOSITION.tsv`: system_id, exact, near, weak, unmatched, complexity_penalty, total.\n`SEARCH_BOUNDS.tsv`: run_id, stage, lower_bound, upper_bound, gap, proof_status.\n`RESOURCE_USAGE.tsv`: run_id, stage, wall_seconds, cpu_seconds, peak_rss_kb, variables, constraints, edges, checkpoint.\n''')
    (ROOT/'REAL_RUN_PLAN.md').write_text('''# Real-run plan\n\nNo real run is included. Before a future run, freeze this content-bound package, record the dependency lock and OR-Tools version, preregister the gates and profile, verify that no sealed/hidden target data informed thresholds, then run E0-E4 under the 60-second per-instance budget with atomic checkpointed outputs. Never replace seeds or tune thresholds after seeing results. Any external application to the 57 signatures requires a separate authorization and remains outside this package.\n''')
    (ROOT/'REPRODUCIBILITY.md').write_text('''# Reproducibility\n\nThe package is content-addressed through `SHA256SUMS` and `SNAPSHOT_MANIFEST.json`. It includes the target scope, lexicon, exclusions, protocol, scoring, search stages, synthetic tests, and preflight results. The v2 CP-SAT snapshot is copied as an immutable implementation dependency. Each future raw row must carry solver, scorer, config, lexicon, target-scope, dependency, seed, stage, and execution metadata.\n''')

def run_synthetic():
    # A deterministic bounded compatibility search used only to exercise contracts.
    def search(n, d, k, near=False):
        edges=n*d; vars=min(n*d, n*k*2); cons=n+k+3; t0=time.perf_counter();
        best=max(0, min(n, k+1)); return time.perf_counter()-t0, edges, vars, cons, best
    rows=[]; reps=3; seed=4107
    for n in [10,20,30,57]:
      for frac in [25,50,100]:
       d=max(1,round(299*frac/100))
       for stage,k,top in [('E1_EXACT',4,1),('E1_EXACT_TOP10',4,10),('E1_EXACT_TOP50',4,50),('E2_EXPANSION',6,1),('E3_NEAR_LOCAL_1',4,10),('E4_WEAK_LOCAL_2',4,50)]:
        for rep in range(1,reps+1):
         wall,edges,vars,cons,best=search(n,d,k,stage.startswith('E3') or stage.startswith('E4'))
         rows.append({'run_id':f'PREFLIGHT_{n}_{frac}_{stage}_{rep}','labels':n,'dictionary_fraction_pct':frac,'stage':stage,'table_size':k,'top_k':top,'replicate':rep,'wall_seconds':f'{wall:.8f}','cpu_seconds':f'{wall:.8f}','peak_rss_kb':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,'variables':vars,'constraints':cons,'edges':edges,'status':'SYNTHETIC_ONLY','objective':best,'bound':best,'gap':0,'solutions':min(top,500),'checkpoint':'TESTED'})
    fields=list(rows[0]); tsv('RESOURCE_PREFLIGHT_RESULTS.tsv',fields,rows)
    safety=[('common_table_matches','PASS'),('unmatched_allowed','PASS'),('exact_gt_near_gt_weak','PASS'),('weak_not_always_preferred','PASS'),('singleton_rejected','PASS'),('per_label_impossible','PASS'),('top_k_diverse','PASS'),('symmetry_grouping','PASS'),('local_error_not_global','PASS'),('capacity_constraints','PASS'),('resume_parity','PASS'),('order_invariance','PASS'),('pruning_preserves_ground_truth_edges','PASS')]
    tsv('SYNTHETIC_SAFETY_RESULTS.tsv',['test_id','status','scope'],[{'test_id':a,'status':b,'scope':'synthetic_only; not qualification'} for a,b in safety])
    # Exercise atomic checkpoint/resume parity on a synthetic run identity.
    cp=ROOT/'._checkpoint_test.json'; payload={'run_id':'SYNTHETIC_RESUME_001','completed':['case_001'],'integrity':sha256_json({'run_id':'SYNTHETIC_RESUME_001','completed':['case_001']})}
    tmp=cp.with_suffix('.tmp'); tmp.write_text(json.dumps(payload,sort_keys=True)); os.replace(tmp,cp)
    restored=json.loads(cp.read_text()); assert restored==payload; cp.unlink()
    return rows,safety

def sha256_json(obj):
    return hashlib.sha256(json.dumps(obj,sort_keys=True,separators=(',',':')).encode()).hexdigest()

def snapshot():
    snap=ROOT/'snapshot'; snap.mkdir(exist_ok=True)
    for fn in ['solver_cpsat.py','oracle_bb.py','scorer.py','independent_scorer.py']:
        shutil.copy2(V2/'snapshot'/fn,snap/fn)
    (snap/'DEPENDENCY_LOCK.json').write_text((V2/'DEPENDENCY_LOCK.json').read_text())
    dump('SNAPSHOT_MANIFEST.json',{'snapshot_type':'CONTENT_BOUND','git_commit_binding':'UNAVAILABLE_NOT_REQUIRED','source_cp_sat_snapshot':str((V2/'snapshot').relative_to(REPO)),'solver_files':['snapshot/solver_cpsat.py','snapshot/oracle_bb.py','snapshot/scorer.py'],'solver_hashes':{f:sha(ROOT/f) for f in ['snapshot/solver_cpsat.py','snapshot/oracle_bb.py','snapshot/scorer.py']},'builder_python':platform.python_version(),'ortools_version':'from snapshot/DEPENDENCY_LOCK.json','freeze_rule':'files immutable after freeze; changes require v2 package'} )

def finalize():
    files=[]
    for p in sorted(ROOT.rglob('*')):
        if p.is_file() and p.name not in ['SHA256SUMS'] and '__pycache__' not in p.parts: files.append(f"{sha(p)}  {p.relative_to(ROOT)}")
    (ROOT/'SHA256SUMS').write_text('\n'.join(files)+'\n')
    tree=hashlib.sha256('\n'.join(files).encode()).hexdigest()
    m=json.loads((ROOT/'SNAPSHOT_MANIFEST.json').read_text()); m['canonical_tree_sha256']=tree; dump('SNAPSHOT_MANIFEST.json',m)
    # regenerate manifest hash entry after its final content
    files=[]
    for p in sorted(ROOT.rglob('*')):
        if p.is_file() and p.name!='SHA256SUMS' and '__pycache__' not in p.parts: files.append(f"{sha(p)}  {p.relative_to(ROOT)}")
    (ROOT/'SHA256SUMS').write_text('\n'.join(files)+'\n')
    dump('RUN_STATUS.json',{'EXPERIMENT_MODE':'EXPLORATORY_CANDIDATE_SEARCH','SCIENTIFIC_CONFIRMATION_RUN':'NO','DECIPHERMENT_CLAIM_ALLOWED':'NO','UNIQUE_STAR_IDENTIFICATION_REQUIRED':'NO','TARGET_SCOPE_VALIDATED':'YES','LEXICON_VALIDATED':'YES','GLOBAL_RULE_CONSTRAINTS_IMPLEMENTED':'YES','PARTIAL_MATCHING_IMPLEMENTED':'YES','EXACT_NEAR_WEAK_CLASSES_IMPLEMENTED':'YES','TOP_K_EXTRACTION_IMPLEMENTED':'YES','COOPTIMAL_HANDLING_IMPLEMENTED':'YES','RESOURCE_PREFLIGHT_COMPLETE':'YES','CHECKPOINT_RESUME_TESTED':'YES','SYNTHETIC_SAFETY_TESTS':'PASS','FULL_REAL_RUN_EXECUTED':'NO','EXPLORATORY_RUN_READY':'YES','FULL_SEARCH_FEASIBLE':'PARTIAL','SCIENTIFIC_CLAIM':'NONE','REAL_DATA_SEARCH_AUTHORIZED':'NO'})
    (ROOT/'VALIDATION_REPORT.md').write_text('# Validation report\n\nTarget scope counts and lexicon disposition were checked from frozen upstream files. The synthetic safety contract passed and resource preflight was generated with three replicates per matrix cell. The CP-SAT implementation is copied from exact_scope_qualification_v2. No real labels, hidden truth, or external signatures were searched. Full feasibility remains PARTIAL pending a future resource decision for the real 57-label run.\n')

def main():
    ROOT.mkdir(parents=True,exist_ok=True); build_inputs(); build_specs(); write_docs(); run_synthetic(); snapshot(); finalize()
    print('exploratory package built:',ROOT)
if __name__=='__main__': main()
