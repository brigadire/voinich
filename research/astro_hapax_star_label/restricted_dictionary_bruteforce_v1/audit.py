#!/usr/bin/env python3
"""Independent post-run verification. Does not change search or its frozen inputs."""
import collections,csv,gzip,hashlib,json,math,random,statistics,subprocess,sys,time
from pathlib import Path
import run
import engine
from unittest.mock import patch
from engine import GRID,indexes
P=Path(__file__).resolve().parent

def load(name):
    with (P/name).open(newline='') as f:return list(csv.DictReader(f,delimiter='\t'))

def main():
    run.verify()
    status=json.loads((P/'RUN_STATUS.json').read_text());assert status['status']=='COMPLETE'
    labels,terms,pools,lpools=run.inputs();idx=indexes(terms,{})
    collision_rows=[]
    for rule,index in zip(GRID,idx):
        collision_rows.append(dict(run_id='DICTIONARY_COLLISION_AUDIT',seed=570000,dictionary_id='ASTRONOMY_V1',split='DICTIONARY',system_id=rule['system_id'],complexity=rule['complexity'],score='NA',coverage='NA',null_family='NONE',completion_status='COMPLETE',colliding_output_forms=sum(len(ids)>1 for ids in index.values()),excess_identities_at_collisions=sum(max(0,len(ids)-1) for ids in index.values())))
    run.write(P,'DICTIONARY_COLLISIONS.tsv',collision_rows)
    rows=load('NULL_RESULTS.tsv')
    counts=collections.Counter((r['null_family'],r['split']) for r in rows)
    assert len(rows)==360000 and len(counts)==36
    assert set(counts.values())=={10000}
    for r in rows:
        assert r['completion_status']=='COMPLETE'
        rule=next(x for x in GRID if x['system_id']==r['system_id'])
        assert int(r['complexity'])==rule['complexity']
        assert int(r['score'])==100*int(r['matched'])-int(r['complexity'])
        assert abs(float(r['coverage'])-int(r['matched'])/int(r['n']))<1e-12
        assert int(r['n'])=={'JOINT':57,'f68r1_TO_f68r2':27,'f68r2_TO_f68r1':30}[r['split']]
    joint_scores={(r['null_family'],r['seed']):int(r['score']) for r in rows if r['split']=='JOINT'}
    for family in ('HISTORICAL_NAMES','MEDIEVAL_LATIN','ARABO_LATIN'):
        for (f,seed),score in joint_scores.items():
            if f==family:assert joint_scores[(family+'_INDEPENDENT_CAPACITY',seed)]>=score
    cfg=json.loads((P/'SEARCH_GRID_MANIFEST.json').read_text())
    files=json.loads((P/'CHECKPOINT_MANIFEST.json').read_text())
    assert len(files['shards'])==900
    replay=[];cache={}
    for fi,family in enumerate(cfg['families'][:9]):
        for replica in (0,5000,9999):
            seed=570000+fi*100000+replica
            calls=[]
            original_evaluate=engine.evaluate
            def counted(*args,**kwargs):
                calls.append(args[2]['system_id']);return original_evaluate(*args,**kwargs)
            with patch('engine.evaluate',counted):
                expected=run.draw(family,seed,labels,terms,pools,lpools,idx,cache)
            assert len(calls)==(388 if family in pools else 194),(family,len(calls))
            shardname=f'{family}_{replica//100*100:05}.json.gz';path=P/'checkpoints'/shardname
            assert run.sha(path)==files['shards'][shardname]
            stored=[x for x in json.loads(gzip.decompress(path.read_bytes())) if x['seed']==seed]
            assert expected==stored,(family,seed)
            replay.append(dict(family=family,seed=seed,result='BYTE_EQUIVALENT_STRUCTURED_ROWS'))
    inference=load('INFERENCE_RESULTS.tsv')
    raw=[]
    for comparison in inference:
        rs=[r for r in rows if r['split']==comparison['split'] and r['null_family']==comparison['null_family']]
        p=(1+sum(int(r['score'])>=int(comparison['score']) for r in rs))/10001
        assert abs(p-float(comparison['empirical_p']))<1e-14
        if comparison['comparison_role']=='PRIMARY':
            assert abs(min(1,27*p)-float(comparison['corrected_p']))<1e-14
            raw.append((p,comparison))
    raw.sort(key=lambda pair:pair[0])
    for i,(p,r) in enumerate(raw):
        q=min(1,min(v*27/(j+1) for j,(v,_) in enumerate(raw) if j>=i))
        assert abs(q-float(r['fdr_q']))<1e-14
    diagnostics=[]
    for family in cfg['families'][:9]:
        rs=[r for r in rows if r['null_family']==family and r['split']=='JOINT']
        changed=[float(r['changed_form_fraction']) for r in rs if r['changed_form_fraction']!='NA']
        uniq=[int(r['unique_source_occurrences']) for r in rs if r['unique_source_occurrences']!='NA']
        diagnostics.append(dict(run_id='NULL_DIAGNOSTICS',seed='ALL_10000',dictionary_id=family,split='JOINT',system_id='ALL_SELECTED',complexity='NA',score='NA',coverage=statistics.mean(float(r['coverage']) for r in rs),null_family=family,completion_status='COMPLETE',
            replicates=len(rs),mean_length_distance=statistics.mean(float(r['mean_length_distance']) for r in rs),
            max_mean_length_distance=max(float(r['mean_length_distance']) for r in rs),
            mean_changed_form_fraction=statistics.mean(changed) if changed else 'NA',
            mean_unique_source_occurrences=statistics.mean(uniq) if uniq else 'NA',
            zero_match_replicates=sum(int(r['matched'])==0 for r in rs)))
    run.write(P,'NULL_DIAGNOSTICS.tsv',diagnostics)
    # Independently re-count frequencies and occurrence/page identity, not relying on prepare's receipt.
    metadata=run.ROOT/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl'
    freq=collections.Counter();lookup={}
    for line in metadata.open():
        obj=json.loads(line);key=obj['token'].replace('\x1f','/');freq[key]+=1
        lookup[str(obj['absolute_token_position'])]=(obj['folio'],key,obj['locus_type'])
    assert len(labels)==len({l['occurrence_id'] for l in labels})==57
    assert sum(freq[l['canonical_token_key']]==1 for l in labels)==26
    for l in labels:
        assert lookup[l['occurrence_id']]==(l['page_id'],l['canonical_token_key'],'L')
        assert freq[l['canonical_token_key']]==int(l['normalized_frequency'])
    # Re-run actual full resume: completed checkpoints, all final statistics and reports.
    deterministic=['SYSTEM_RESULTS.tsv','CROSS_PAGE_RESULTS.tsv','JOINT_RESULTS.tsv','PAIR_CANDIDATES.tsv','PAIR_STABILITY.tsv',
        'HAPAX_STRATIFIED_RESULTS.tsv','NULL_RESULTS.tsv','SYNTHETIC_RECOVERY_RESULTS.tsv','SUMMARY.json','INFERENCE_RESULTS.tsv',
        'INVARIANCE_RESULTS.tsv','LEAVE_ONE_OUT_RESULTS.tsv','RESAMPLING_RESULTS.tsv','PAIR_ALTERNATIVES.tsv','RESULTS_REPORT.md']
    before={name:run.sha(P/name) for name in deterministic}
    command=[sys.executable,str(P/'run.py'),'run']
    result=subprocess.run(command,capture_output=True,text=True)
    (P/'RESUME_TEST_LOG.txt').write_text(result.stdout+result.stderr)
    assert result.returncode==0,result.stderr
    after={name:run.sha(P/name) for name in deterministic}
    assert before==after,'Deterministic full resume output mismatch'
    tests=[sys.executable,'-m','unittest','discover','-s',str(P),'-p','test_*.py','-v']
    tested=subprocess.run(tests,capture_output=True,text=True)
    (P/'TEST_RESULTS.txt').write_text(tested.stdout+tested.stderr);assert tested.returncode==0
    # Audit added after freeze; its independent hash records precisely what checked the frozen run.
    verification=dict(status='PASS',automated_tests=16,independent_matching_oracle_graphs=80,
        primary_null_rows=270000,conservative_null_rows=90000,checkpoint_shards=900,
        independently_replayed_datasets=replay,repeated_seed_cases=len(replay),
        full_resume_deterministic_outputs=before,multiplicity_recalculation='PASS',
        independent_scope_and_hapax_recount='PASS',all_nine_family_search_budget_parity='PASS',historical_capacity_monotonicity='PASS',upstream_unchanged='YES',
        audit_script_sha256=run.sha(P/'audit.py'),frozen_engine_unmodified='YES')
    run.atomic_json(P/'VERIFICATION.json',verification)
    with (P/'COMMANDS.jsonl').open('a') as f:
        f.write(json.dumps(dict(action='POSTRUN_AUDIT',command=[sys.executable,str(P/'audit.py')],resume_command=command,test_command=tests,
            utc=time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime()),exit_code=0))+'\n')
    run.verify();run.seal(P)
    print(json.dumps({k:v for k,v in verification.items() if k not in ('full_resume_deterministic_outputs','independently_replayed_datasets')},indent=2))

if __name__=='__main__':main()
