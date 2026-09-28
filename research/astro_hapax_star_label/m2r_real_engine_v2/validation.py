"""One-shot synthetic validation, reporting and candidate sealing (no real inputs)."""
from __future__ import annotations
import argparse
from concurrent.futures import ThreadPoolExecutor, as_completed
import copy
import csv
import hashlib
import json
from pathlib import Path
import platform
import random
import statistics
import sys
import tempfile
import unittest

import engine
import evaluator
import generator
import integrity
import orchestrator

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]


def tsv(root, name, rows, fields=None):
    fields = fields or sorted({k for row in rows for k in row})
    with (root/name).open('w', newline='') as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter='\t', lineterminator='\n')
        w.writeheader()
        w.writerows(rows)


def source_hashes():
    files = sorted(ROOT.glob('*.py'))+sorted((ROOT/'tests').glob('*.py'))+sorted(ROOT.glob('*_SPEC.md'))+[ROOT/'M2R_V2_IMPLEMENTATION_CONTRACT.md']
    return {p.relative_to(ROOT).as_posix(): integrity.file_hash(p) for p in files}


def inventory():
    rows = []
    packages = [('m2r_preproduction_audit_v1','audit'), ('m2r_real_engine_v1','rejected_v1'),
                ('restricted_m2_star_labels_v1','M2_specification_metadata_only')]
    for directory, category in packages:
        for p in sorted((ROOT.parent/directory).rglob('*')):
            if p.is_file():
                rows.append({'path':str(p.relative_to(REPO)), 'category':category,
                             'bytes':p.stat().st_size, 'sha256':integrity.file_hash(p),
                             'access':'HASH_ONLY'})
    for p in sorted((REPO/'research/m2_morphology_audit').glob('*')):
        if p.is_file():
            rows.append({'path':str(p.relative_to(REPO)), 'category':'M2_generator_and_frozen_spec',
                         'bytes':p.stat().st_size,'sha256':integrity.file_hash(p),'access':'HASH_ONLY'})
    task = REPO/'tasks_other/m2r-v2.md'
    rows.append({'path':str(task.relative_to(REPO)),'category':'task','bytes':task.stat().st_size,
                 'sha256':integrity.file_hash(task),'access':'REQUIREMENTS'})
    tsv(ROOT,'INPUT_MANIFEST.tsv',rows)
    v1 = {r['path']:r['sha256'] for r in rows if r['category']=='rejected_v1'}
    engine.atomic_json(ROOT/'V1_BASELINE_HASHES.json',v1)
    manifest = ROOT.parent/'m2r_real_engine_v1/SEALED_REAL_INPUTS_MANIFEST.json'
    # This file is expressly authorized metadata. It does not contain token values.
    engine.atomic_json(ROOT/'SEALED_REAL_INPUTS_MANIFEST.json',json.loads(manifest.read_text()))
    engine.atomic_json(ROOT/'ENVIRONMENT_MANIFEST.json',{
        'python':sys.version,'platform':platform.platform(),'dependencies':'Python standard library only',
        'real_data_authorized':False,'workers':3,'memory_mb_per_worker':512})
    engine.atomic_json(ROOT/'CONFIG.json',engine.asdict(engine.Config()))


def verify_v1():
    expected = json.loads((ROOT/'V1_BASELINE_HASHES.json').read_text())
    return all(integrity.file_hash(REPO/p)==h for p,h in expected.items())


def datasets():
    cases = []
    for family, seed, split in [('mixed',41001,'development'),('composition',51001,'calibration')]:
        cases.append({'family':split,'seed':seed,'kind':family,'size':29,'alphabet':0})
    for i, family in enumerate(generator.POSITIVE_FAMILIES):
        cases.append({'family':'hidden','seed':61001+i,'kind':family,'size':(29,57,68)[i%3],'alphabet':i%3})
    cases.append({'family':'out_of_family','seed':91001,'kind':'composition','size':57,'alphabet':3})
    for i,family in enumerate(generator.NULL_FAMILIES):
        cases.append({'family':'null','seed':71001+i,'kind':family,'size':(29,57,68)[i%3],'alphabet':i%3})
    for i,family in enumerate(generator.HARD_FAMILIES):
        cases.append({'family':'hard_negative','seed':81001+i,'kind':family,'size':29,'alphabet':i%3})
    for case in cases:
        case['dataset'] = 'd'+str(case['seed'])
    return orchestrator.seed_registry(cases)


def prepare():
    if (ROOT/'CANDIDATE_SEAL.json').exists():
        raise RuntimeError('candidate sealed; preparation cannot overwrite it')
    inventory()
    registry = datasets()
    tsv(ROOT,'SYNTHETIC_SEED_REGISTRY.tsv',registry)
    surfaces, truths = ROOT/'surface_inputs', ROOT/'latent_truth'
    surfaces.mkdir(exist_ok=True)
    truths.mkdir(exist_ok=True)
    records = []
    for c in registry:
        surface, truth = generator.generate(c['seed'],c['kind'],c['size'],c['alphabet'])
        key = engine.digest(surface)
        sp, tp = surfaces/(key+'.json'), truths/(key+'.json')
        engine.atomic_json(sp,surface)
        engine.atomic_json(tp,truth)
        records.append({**c,'key':key,'surface':str(sp.relative_to(ROOT)),
                        'truth':str(tp.relative_to(ROOT)), 'surface_sha256':integrity.file_hash(sp),
                        'truth_sha256':integrity.file_hash(tp),
                        'train_labels':len(surface['train']['labels']),
                        'heldout_labels':len(surface['heldout']['labels'])})
    tsv(ROOT,'SYNTHETIC_SPLIT_MANIFEST.tsv',records)
    engine.atomic_json(ROOT/'DATASET_REGISTRY.json',records)
    return records


class RecordedResult(unittest.TextTestResult):
    def __init__(self,*args,**kwargs):
        super().__init__(*args,**kwargs)
        self.rows = []
    def addSuccess(self,test):
        super().addSuccess(test)
        self.rows.append({'test':test._testMethodName,'result':'PASS','details':''})
    def addFailure(self,test,err):
        super().addFailure(test,err)
        self.rows.append({'test':test._testMethodName,'result':'FAIL','details':str(err[1])})
    def addError(self,test,err):
        super().addError(test,err)
        self.rows.append({'test':test._testMethodName,'result':'ERROR','details':str(err[1])})


def regressions():
    suite = unittest.defaultTestLoader.discover(str(ROOT/'tests'))
    with (ROOT/'REGRESSION_TEST_LOG.txt').open('w') as stream:
        result = unittest.TextTestRunner(stream=stream,verbosity=2,resultclass=RecordedResult).run(suite)
    tsv(ROOT,'REGRESSION_TEST_RESULTS.tsv',result.rows)
    return result.wasSuccessful()


def run_case(case, variants=True):
    surface = json.loads((ROOT/case['surface']).read_text())
    with tempfile.TemporaryDirectory(prefix='m2rv2-') as tmp:
        base = orchestrator.run_surface(surface,tmp)
        output = base['output']
        # Commit inference output before the evaluator can read latent truth.
        result_path = ROOT/'inference_outputs'/(case['key']+'.json')
        engine.atomic_json(result_path,output)
        base_hash = engine.digest(output)
        deterministic, checkpoint = [], []
        if variants:
            for mode in ('reversed','permuted','renamed_ids'):
                changed = copy.deepcopy(surface)
                for split in changed.values():
                    for side in split.values():
                        if mode=='reversed':
                            side.reverse()
                        elif mode=='permuted':
                            random.Random(10203).shuffle(side)
                        else:
                            for i,row in enumerate(side):
                                row['id'] = 'neutral-'+str(len(side)-i)
                r = orchestrator.run_surface(changed,tmp,shard=mode)
                actual = engine.digest(r['output'])
                deterministic.append({'dataset':case['dataset'],'variant':mode,'original_hash':base_hash,
                                      'variant_hash':actual,'result':'PASS' if actual==base_hash else 'FAIL'})
            paused = orchestrator.run_surface(surface,tmp,checkpoint=True,mode='pause')
            if paused['output']['train'].get('reason') == 'checkpoint_pause':
                resumed = orchestrator.run_surface(surface,tmp,checkpoint=True,mode='resume')
                actual = engine.digest(resumed['output'])
                checkpoint.append({'dataset':case['dataset'],'continuous_hash':base_hash,'resumed_hash':actual,
                                   'result':'PASS' if actual==base_hash else 'FAIL'})
            else:
                checkpoint.append({'dataset':case['dataset'],'continuous_hash':base_hash,'resumed_hash':'NA','result':'FAIL'})
        truth = json.loads((ROOT/case['truth']).read_text())
        metrics = evaluator.evaluate(output,truth)
    return case, metrics, deterministic, checkpoint, base_hash


def size_results():
    from tests.test_engine import fixture, RULES
    result = []
    for n in (29,57,68):
        t,l = [engine.load_rows(x)[0] for x in fixture(n)]
        m = engine.evaluate(t,l,RULES)
        result.append({'size':n,'gain':m['gain'],'baseline':m['baseline'],'total':m['total'],
                       **m['components'],'result':'PASS' if abs(m['gain']-(m['baseline']-m['total'])/m['baseline'])<1e-12 else 'FAIL'})
    tsv(ROOT,'SIZE_NORMALIZATION_RESULTS.tsv',result)


def traceability(passed):
    mapping = [
        ('CR-01','unknown assignments','Hungarian dummy-node matching','hungarian / evaluate','test_no_zip_assignment','test_exact_matching_bruteforce','exact optimum equals exhaustive small oracle'),
        ('CR-02','minimum support 3','reject unsupported state','support / evaluate','test_min_support_three','test_heldout_frozen','all retained rules >=3 distinct pairs'),
        ('CR-03','independent examples','canonical unique surfaces and pair sets','load_rows / support','test_duplicate_not_independent','test_order_invariance','duplicates cannot increase independent support'),
        ('CR-04','general-alphabet precision','exact latent tuple intersection','evaluator.rule_metrics','test_precision_general_alphabet','test_neutral_dataset_roles','no alphabet-dependent correctness metric'),
        ('CR-05','explicit rule conflicts','reject forward and inverse conflict','validate_rules / transform','test_conflicting_rules_explicit','test_roles_and_variable_lengths','conflict raises and compositions retain roles'),
        ('CR-06','complete SHA-256 registry','all regular files except ledger itself','integrity','test_sha256_registry_complete','final integrity.verify','exact inventory equality and correct digest for every file'),
        ('CR-07','bounded search/checkpoint','beam limits and checksummed resume','search / worker / orchestrator','test_search_resource_bounds','test_checkpoint_resume_identity','bounded completion or INCOMPLETE; resume byte-identical'),
        ('CR-08','multisymbol morphology','substring/whole units and ordered DAG paths','units / align','test_multisymbol_units_active','test_roles_and_variable_lengths','multichar rules used in inferred derivations'),
        ('CR-09','size normalization','six bit-length components divided by baseline','evaluate','test_size_normalized_scoring','test_simpler_model_wins','duplicate invariance; correct arithmetic; complexity costs')]
    rows = []
    for r in mapping:
        rows.append(dict(zip(['finding','requirement','v2_design_decision','implementation_component','unit_test','integration_test','acceptance_criterion'],r),verification_result='PASS' if passed else 'FAIL'))
    tsv(ROOT,'M2R_V2_REQUIREMENTS_TRACEABILITY.tsv',rows)


def report(results, regression_ok, hidden):
    metrics = [{**{k:c[k] for k in ('dataset','family','kind','size')},**m} for c,m,_,_,_ in results]
    def subset(keys, source=metrics):
        return [{k:r.get(k,'NA') for k in keys} for r in source]
    tsv(ROOT,'SYNTHETIC_RULE_RECOVERY.tsv',subset(['dataset','family','precision','recall','role_aware_precision','role_aware_recall','role_agnostic_precision','role_agnostic_recall','train_boundary_accuracy','heldout_boundary_accuracy','train_path_accuracy','heldout_path_accuracy','mean_support','singleton_fraction','unsupported_rules']))
    tsv(ROOT,'SYNTHETIC_ASSIGNMENT_RECOVERY.tsv',subset(['dataset','family','train_assignment_accuracy','heldout_assignment_accuracy','train_top_k_accuracy','heldout_top_k_accuracy','train_margin','heldout_margin','train_ambiguous','heldout_ambiguous']))
    tsv(ROOT,'SYNTHETIC_PREDICTIVE_RESULTS.tsv',subset(['dataset','family','status','train_coverage','heldout_coverage','train_gain','heldout_gain','transfer_retention','accepted']))
    nulls = [m for m in metrics if m['family']=='null']
    negatives = [m for m in metrics if m['family']=='hard_negative']
    tsv(ROOT,'SYNTHETIC_NULL_RESULTS.tsv',nulls, sorted(set().union(*(r.keys() for r in metrics))))
    tsv(ROOT,'HARD_NEGATIVE_RESULTS.tsv',negatives, sorted(set().union(*(r.keys() for r in metrics))))
    determinism = [r for _,_,rr,_,_ in results for r in rr]
    checkpoints = [r for _,_,_,rr,_ in results for r in rr]
    # Shards carry scheduling metadata only. Aggregation must erase it canonically.
    aggregate_a = orchestrator.aggregate([{'key':c['key'],'shard':str(i%3),'output':h} for i,(c,_,_,_,h) in enumerate(results)])
    aggregate_b = orchestrator.aggregate([{'key':c['key'],'shard':str(i%2),'output':h} for i,(c,_,_,_,h) in enumerate(reversed(results))])
    determinism.append({'dataset':'ALL','variant':'shard_layout','original_hash':engine.digest(aggregate_a),
                        'variant_hash':engine.digest(aggregate_b),'result':'PASS' if aggregate_a==aggregate_b else 'FAIL'})
    tsv(ROOT,'DETERMINISM_RESULTS.tsv',determinism)
    tsv(ROOT,'CHECKPOINT_RESUME_RESULTS.tsv',checkpoints,['dataset','continuous_hash','resumed_hash','result'])
    positives = [m for m in metrics if m['family']=='hidden']
    def mean(key):
        values = [m[key] for m in positives if m['status']=='COMPLETE']
        return statistics.mean(values) if values else None
    values = [m['heldout_gain'] for m in nulls if m['status']=='COMPLETE']
    p95 = evaluator.quantile(values,.95)
    hnaccepted = sum(bool(m['accepted']) for m in negatives)
    pg = [m['heldout_gain'] for m in positives if m['status']=='COMPLETE']
    gap = statistics.median(pg)-p95 if pg and p95 is not None else None
    all_complete = all(m['status']=='COMPLETE' for m in metrics)
    order_ok = all(r['result']=='PASS' for r in determinism)
    cp_ok = bool(checkpoints) and all(r['result']=='PASS' for r in checkpoints)
    v1_ok = verify_v1()
    gates = {
        'precision':mean('precision') is not None and mean('precision')>=.8,
        'recall':mean('recall') is not None and mean('recall')>=.75,
        'assignment':mean('train_assignment_accuracy') is not None and mean('train_assignment_accuracy')>=.75,
        'heldout_coverage':mean('heldout_coverage') is not None and mean('heldout_coverage')>=.70,
        'heldout_gain':mean('heldout_gain') is not None and mean('heldout_gain')>0,
        'null_separation':gap is not None and gap>=.05,
        'independent_support':mean('mean_support') is not None and mean('mean_support')>=3,
        'singleton_zero':bool(positives) and all(m.get('singleton_fraction')==0 for m in positives),
        'unsupported_zero':bool(positives) and all(m.get('unsupported_rules')==0 for m in positives),
        'hard_negatives':len(negatives)==10 and hnaccepted==0,
        'regressions':regression_ok,'order_invariance':order_ok,'checkpoint':cp_ok,
        'complete_searches':all_complete, 'v1_unchanged':v1_ok}
    passed = hidden and all(gates.values())
    status = {
        'M2R_V2_STATUS':'COMPLETE' if passed else 'BLOCKED',
        'M2R_V1_MODIFIED':'NO' if v1_ok else 'YES','REAL_DATA_CONTENTS_ACCESSED':'NO',
        'ASSIGNMENT_SEARCH':'IMPLEMENTED','MULTISYMBOL_MORPHOLOGY':'IMPLEMENTED',
        'MIN_INDEPENDENT_SUPPORT':3,'DUPLICATE_SUPPORT_GUARD':'PASS' if regression_ok else 'FAIL',
        'RULE_CONFLICT_HANDLING':'PASS' if regression_ok else 'FAIL','SIZE_NORMALIZED_SCORING':'PASS' if regression_ok else 'FAIL',
        'SEARCH_TYPE':'BOUNDED_DETERMINISTIC_OPTIMIZATION','ORDER_INVARIANCE':'PASS' if order_ok else 'FAIL',
        'CHECKPOINT_RESUME_IDENTICAL':'YES' if cp_ok else 'NO','LATENT_RULE_PRECISION':mean('precision'),
        'LATENT_RULE_RECALL':mean('recall'),'ASSIGNMENT_ACCURACY':mean('train_assignment_accuracy'),
        'HELDOUT_COVERAGE':mean('heldout_coverage'),'HELDOUT_COMPRESSION_GAIN':mean('heldout_gain'),
        'SYNTHETIC_NULL_P95':p95,'HARD_NEGATIVES_ACCEPTED':f'{hnaccepted}/{len(negatives)}',
        'SINGLETON_FRACTION':mean('singleton_fraction'),'CR01_CR09_REGRESSION':'PASS' if regression_ok else 'FAIL',
        'M2R_V2_SYNTHETIC_VALIDATION':('PASS' if passed else 'FAIL') if hidden else 'NOT_RUN',
        'M2R_V2_ENGINE_FROZEN':'YES' if passed else 'NO',
        'READY_FOR_INDEPENDENT_PREPRODUCTION_AUDIT':'YES' if passed else 'NO',
        'REAL_DATA_SEARCH_AUTHORIZED':'NO','RESULTS_REPRODUCIBLE':'YES' if order_ok and cp_ok and v1_ok else 'NO'}
    engine.atomic_json(ROOT/'FINAL_STATUS.json',status)
    engine.atomic_json(ROOT/'GATE_RESULTS.json',{'gates':gates,'positive_median_minus_null_p95':gap,
        'null_p99':evaluator.quantile(values,.99),'null_max':max(values) if values else None,
        'null_accepted_rate':sum(m['accepted'] for m in nulls)/len(nulls) if nulls else None})
    failed = [k for k,v in gates.items() if not v]
    text = '# M2R-v2 synthetic validation\n\n'
    text += ('PASS. Candidate meets the declared synthetic gates.\n' if passed else
             'BLOCKED. The synthetic gates do not justify an engine freeze or independent pre-production readiness.\n')
    text += '\nFailed gates: '+', '.join(failed)+'.\n'
    text += '\nMetrics use full latent rule tuples and exact surface assignments, macro-averaged over the six hidden positive datasets. '
    text += 'Negative and out-of-family datasets are reported separately. Every dataset uses the same isolated worker and fixed configuration. '
    text += 'Only one replicate per listed condition was run; tail quantiles from eight nulls are descriptive and do not establish a population false-positive guarantee.\n'
    text += '\nThe shuffled-assignment null preserves the bags of surfaces and cannot be distinguished by an assignment-search engine. '
    text += 'Bigram-preserving nulls may also preserve some entire tokens. No truth-aware score adjustment is used. '
    text += 'The bounded candidate pool and rule search are heuristic; full global optimality is not claimed. '
    text += 'All-pairs top-k derivations are retained for the selected model; search pool pruning can exclude true morphology.\n'
    text += '\nA zero hard-negative acceptance count cannot compensate for poor positive recovery. '
    text += 'The original v1 is checksum-verified unchanged. Real inputs were neither loaded nor searched; only authorized manifests and SHA-256 metadata were inspected.\n'
    if hidden:
        text += '\nHidden validation has been disclosed. Do not repair or tune this v2 implementation. Further development requires M2R-v3 and fresh hidden seeds.\n'
    text += '\n```text\n'+'\n'.join(f'{k}={"NA" if v is None else v}' for k,v in status.items())+'\n```\n'
    (ROOT/'M2R_V2_SYNTHETIC_VALIDATION_REPORT.md').write_text(text)
    traceability(regression_ok)
    size_results()
    if passed:
        # Production remains unauthorized even after successful synthetic freeze.
        snapshot = ROOT/'frozen_source'
        snapshot.mkdir(exist_ok=True)
        for rel in source_hashes():
            target = snapshot/rel
            target.parent.mkdir(parents=True,exist_ok=True)
            target.write_bytes((ROOT/rel).read_bytes())
        engine.atomic_json(ROOT/'M2R_V2_FREEZE_MANIFEST.json',{
            'status':status,'sources':source_hashes(), 'executable_hash':integrity.file_hash(ROOT/'worker.py'),
            'environment_hash':integrity.file_hash(ROOT/'ENVIRONMENT_MANIFEST.json'),
            'registry_hash':integrity.file_hash(ROOT/'SYNTHETIC_SEED_REGISTRY.tsv'),
            'config_hash':integrity.file_hash(ROOT/'CONFIG.json'),
            'hidden_result_hash':integrity.file_hash(ROOT/'GATE_RESULTS.json')})
    integrity.write(ROOT)
    assert integrity.verify(ROOT)
    print(engine.canonical(status),flush=True)


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument('phase',choices=['prepare','development','validate','verify'])
    args = ap.parse_args()
    if args.phase=='verify':
        assert integrity.verify(ROOT)
        assert verify_v1()
        seal = json.loads((ROOT/'CANDIDATE_SEAL.json').read_text())
        assert seal['sources']==source_hashes()
        print('INTEGRITY=PASS; V1_UNCHANGED=YES; CANDIDATE_UNCHANGED=YES')
        return
    if args.phase=='prepare':
        prepare()
        print('prepared synthetic-only input bags; no evaluation')
        return
    if (ROOT/'HIDDEN_VALIDATION_STARTED.json').exists():
        raise RuntimeError('hidden validation already started; v2 must not be changed or revalidated')
    cases = json.loads((ROOT/'DATASET_REGISTRY.json').read_text())
    regression_ok = regressions()
    if not regression_ok:
        raise RuntimeError('regression failure; no hidden validation permitted')
    (ROOT/'inference_outputs').mkdir(exist_ok=True)
    if args.phase=='development':
        cases = [c for c in cases if c['family'] in {'development','calibration'}]
    else:
        cases = [c for c in cases if c['family'] not in {'development','calibration'}]
        seal = {'sources':source_hashes(),'config_sha256':integrity.file_hash(ROOT/'CONFIG.json'),
                'datasets_sha256':integrity.file_hash(ROOT/'DATASET_REGISTRY.json'),
                'contract_sha256':integrity.file_hash(ROOT/'M2R_V2_IMPLEMENTATION_CONTRACT.md'),
                'stage':'CANDIDATE_ONLY_NOT_ENGINE_FREEZE'}
        engine.atomic_json(ROOT/'CANDIDATE_SEAL.json',seal)
        with (ROOT/'HIDDEN_VALIDATION_STARTED.json').open('x') as f:
            f.write(engine.canonical({'candidate_hash':engine.digest(seal),'one_shot':True})+'\n')
    results = []
    with ThreadPoolExecutor(max_workers=3) as executor:
        futures = [executor.submit(run_case,c) for c in cases]
        for future in as_completed(futures):
            row = future.result()
            results.append(row)
            c,m,_,_,_=row
            print(f"finished {c['dataset']} {c['kind']}: {m['status']}",flush=True)
    results.sort(key=lambda x:x[0]['dataset'])
    if args.phase=='development':
        engine.atomic_json(ROOT/'DEVELOPMENT_RESULTS.json',[{'dataset':c['dataset'],**m} for c,m,_,_,_ in results])
        print('Development complete; fixed profile unchanged.',flush=True)
    else:
        assert source_hashes()==json.loads((ROOT/'CANDIDATE_SEAL.json').read_text())['sources']
        engine.atomic_json(ROOT/'HIDDEN_RESULT_MANIFEST.json',[
            {'dataset':c['dataset'],'surface_sha256':c['surface_sha256'],'truth_sha256':c['truth_sha256'],
             'output_hash':h,'metrics':m} for c,m,_,_,h in results])
        report(results,regression_ok,True)


if __name__=='__main__':
    main()
