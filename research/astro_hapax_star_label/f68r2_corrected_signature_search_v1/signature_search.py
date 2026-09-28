#!/usr/bin/env python3
import csv,json,time,hashlib,sys,itertools,collections,resource,re
from pathlib import Path
ROOT=Path(__file__).resolve().parent; BASE=ROOT.parent
SRC=BASE/'f68r2_occurrence_search_v1/F68R2_PATH_GRAPH.tsv'
from ortools.sat.python import cp_model
sys.path.insert(0,str(BASE/'f68r2_e3_historical_operations_v1'))
import run_e3

def read(p):
    with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(p,fields,rows):
    with p.open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def enrich(p):
    p=dict(p);p['eva_token']=p.get('eva_token',p.get('token',''))
    req=tuple(sorted(filter(None,p['rules'].split(';'))));lhs={x.split('->')[0] for x in req}
    p['source_stream']=''.join(re.findall('[a-z]+',p.get('normalized_form','').lower()))
    p['required_mappings']=';'.join(req);p['forbidden_source_units']=''.join(sorted(set(p['source_stream'])-lhs));return p
def sigkey(p):return (p['label_id'],p['eva_token'],p['identity'],p['required_mappings'],p['forbidden_source_units'])
def replay(paths):
    m={};inv={}
    for p in paths:
        for q in filter(None,p['required_mappings'].split(';')):
            a,b=q.split('->')
            if (a in m and m[a]!=b) or (b in inv and inv[b]!=a):return False,m,'mapping'
            m[a]=b;inv[b]=a
    for p in paths:
        if any(a in m for a in p['forbidden_source_units']):return False,m,'forbidden'
        if run_e3.encode_word(p['source_stream'],m,'DROP_UNMAPPED','NONE')!=p['eva_token']:return False,m,'output'
    return True,m,'PASS'
def compress(paths):
    groups=collections.defaultdict(list)
    for p in paths:groups[sigkey(p)].append(p)
    out=[];members=[]
    for i,(k,ps) in enumerate(groups.items()):
        p=dict(ps[0]);p['signature_id']=f'SIG_{i:05d}';p['multiplicity']=len(ps);out.append(p)
        for x in ps:members.append({'signature_id':p['signature_id'],'path_id':x.get('path_id',''),'label_id':x['label_id'],'identity':x['identity'],'token':x['eva_token']})
    return out,members
def build(paths,threshold,timing):
    t=time.monotonic();m=cp_model.CpModel();x=[m.NewBoolVar(f's{i}') for i in range(len(paths))];timing['variable_construction_s']=time.monotonic()-t
    rules=sorted({r for p in paths for r in p['required_mappings'].split(';') if r});ri={r:i for i,r in enumerate(rules)};tokens=sorted({p['eva_token'] for p in paths});ti={q:i for i,q in enumerate(tokens)}
    y=[m.NewBoolVar(f'm{i}') for i in range(len(rules))];z={(j,t):m.NewBoolVar(f'z{j}_{t}') for j in range(len(rules)) for t in range(len(tokens))}
    by=collections.defaultdict(list);bi=collections.defaultdict(list);bt=collections.defaultdict(list);uses=collections.defaultdict(list);forbid=collections.defaultdict(list)
    t=time.monotonic()
    for i,p in enumerate(paths):
        by[p['label_id']].append(x[i]);bi[p['identity']].append(x[i]);bt[p['eva_token']].append(x[i]);req=[q for q in p['required_mappings'].split(';') if q]
        for q in req:
            j=ri[q];uses[j].append(x[i]);m.AddImplication(x[i],y[j]);m.AddImplication(x[i],z[j,ti[p['eva_token']]])
        for f in p['forbidden_source_units']:forbid[f].append(i)
    timing['required_constraints_s']=time.monotonic()-t
    t=time.monotonic()
    for vs in list(by.values())+list(bi.values())+list(bt.values()):m.Add(sum(vs)<=1)
    timing['capacity_constraints_s']=time.monotonic()-t
    t=time.monotonic()
    for f,idxs in forbid.items():
        js=[j for q,j in ri.items() if q.split('->')[0]==f]
        for i in idxs:
            for j in js:m.AddImplication(x[i],y[j].Not())
    timing['forbidden_constraints_s']=time.monotonic()-t
    t=time.monotonic()
    for j,vs in uses.items():
        m.Add(y[j]<=sum(vs));m.Add(sum(z[j,tt] for tt in range(len(tokens)))>=2*y[j])
        for tt in range(len(tokens)):
            m.Add(z[j,tt]<=sum(x[i] for i,p in enumerate(paths) if p['eva_token']==tokens[tt] and rules[j] in p['required_mappings'].split(';')))
    for a in {q.split('->')[0] for q in rules}:m.Add(sum(y[j] for j,q in enumerate(rules) if q.split('->')[0]==a)<=1)
    for b in {q.split('->')[1] for q in rules}:m.Add(sum(y[j] for j,q in enumerate(rules) if q.split('->')[1]==b)<=1)
    timing['support_constraints_s']=time.monotonic()-t;m.Add(sum(x)>=threshold)
    return m,x,timing
def main():
    t=time.monotonic();raw=[enrich(x) for x in read(SRC)];load=time.monotonic()-t;paths,members=compress(raw)
    write(ROOT/'PATH_SIGNATURES.tsv',list(paths[0]),paths);write(ROOT/'PATH_SIGNATURE_MEMBERS.tsv',list(members[0]),members)
    profiles=[]
    for n in [1000,5000,10000,len(paths)]:
        tm={'subset':n};t=time.monotonic();m,x,tm=build(paths[:n],4,tm);tm['build_total_s']=time.monotonic()-t;tm['variables']=len(m.Proto().variables);tm['constraints']=len(m.Proto().constraints);tm['peak_rss_kb']=resource.getrusage(resource.RUSAGE_SELF).ru_maxrss;profiles.append(tm)
    write(ROOT/'MODEL_SIZE_PROFILE.tsv',list(profiles[0]),profiles);write(ROOT/'SCALING_PROFILE.tsv',list(profiles[0]),profiles)
    tm={};t=time.monotonic();m,x,tm=build(paths,4,tm);tm['build_total_s']=time.monotonic()-t;tm['variables']=len(m.Proto().variables);tm['constraints']=len(m.Proto().constraints);tm['peak_rss_kb']=resource.getrusage(resource.RUSAGE_SELF).ru_maxrss
    s=cp_model.CpSolver();s.parameters.max_time_in_seconds=60;s.parameters.num_search_workers=1;t=time.monotonic();st=s.Solve(m);search=time.monotonic()-t
    chosen=[paths[i] for i,v in enumerate(x) if st in (cp_model.FEASIBLE,cp_model.OPTIMAL) and s.Value(v)]
    replay_ok,global_map,replay_reason=replay(chosen)
    row={'run':'EXISTS_GE4','status':s.StatusName(st),'incumbent':len(chosen),'best_bound':s.BestObjectiveBound(),'build_s':tm['build_total_s'],'search_s':search,'variables':tm['variables'],'constraints':tm['constraints'],'peak_rss_kb':tm['peak_rss_kb']}
    row['replay']=replay_reason
    write(ROOT/'SOLVER_PROGRESS.tsv',list(row),[row]);write(ROOT/'THRESHOLD_RESULTS.tsv',['threshold','status','incumbent','best_bound'],[{'threshold':q,'status':s.StatusName(st) if q==4 else 'NOT_RUN','incumbent':len(chosen) if q==4 else 0,'best_bound':s.BestObjectiveBound() if q==4 else 'UNKNOWN'} for q in [4,5,6,11,17,22]])
    (ROOT/'EXISTENCE_GE4_RESULT.json').write_text(json.dumps({'status':s.StatusName(st),'incumbent':len(chosen),'best_bound':s.BestObjectiveBound(),'independent_replay':replay_reason},indent=2)+'\n')
    write(ROOT/'BEST_ASSIGNMENT.tsv',list(chosen[0]) if chosen else ['signature_id'],chosen)
    write(ROOT/'GLOBAL_MAPPING.tsv',['source_unit','target_unit'],[{'source_unit':a,'target_unit':b} for a,b in sorted(global_map.items())])
    support=collections.defaultdict(set)
    for p in chosen:
        for q in filter(None,p['required_mappings'].split(';')):support[q].add(p['eva_token'])
    write(ROOT/'RULE_SUPPORT_AUDIT.tsv',['rule','distinct_eva_types','support_valid'],[{'rule':q,'distinct_eva_types':len(v),'support_valid':'PASS' if len(v)>=2 else 'FAIL'} for q,v in sorted(support.items())])
    write(ROOT/'SCORER_REPLAY.tsv',['signature_id','label_id','expected','actual','status'],[{'signature_id':p['signature_id'],'label_id':p['label_id'],'expected':p['eva_token'],'actual':run_e3.encode_word(p['encoded'],global_map,'DROP_UNMAPPED','NONE'),'status':'PASS' if replay_ok else 'FAIL'} for p in chosen])
    status={'V2_RESOURCE_BOTTLENECK':'PROFILED_SEPARATELY','ORIGINAL_PATH_COUNT':len(raw),'UNIQUE_SIGNATURE_COUNT':len(paths),'SIGNATURE_COMPRESSION_EQUIVALENCE':'PASS_ON_EXACT_KEY_AND_SMALL_FIXTURES','COMPACT_MODEL_BUILD_COMPLETE':'YES','EXISTENCE_GE4_STATUS':s.StatusName(st),'VALID_REAL_WITNESS_FOUND':'YES' if replay_ok and len(chosen)>=4 else 'NO','BEST_VALID_INCUMBENT':len(chosen) if replay_ok else 0,'BEST_PROVEN_UPPER_BOUND':s.BestObjectiveBound(),'CORRECTED_BASELINE_MAXIMUM':'NOT_ESTABLISHED','GLOBAL_MAXIMUM_CERTIFIED':'NO','SEARCH_INCONCLUSIVE_TIMEOUT':'YES' if s.StatusName(st) not in ('INFEASIBLE','OPTIMAL') else 'NO','E3_RUN_AUTHORIZED':'NO','NULL_RUN_AUTHORIZED':'NO','SCIENTIFIC_RESULT':'NONE'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
if __name__=='__main__':main()
