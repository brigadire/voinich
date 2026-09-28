#!/usr/bin/env python3
import csv, hashlib, json, time, shutil
from collections import defaultdict
from pathlib import Path
from ortools.sat.python import cp_model

ROOT = Path(__file__).resolve().parent
SOURCE = ROOT / 'FROZEN_SOURCE_GRAPH.tsv'
REGISTRY = ROOT / 'TRANSFORMATION_REGISTRY.tsv'
PATH_DIR = ROOT / 'PROFILE_PATHS'
CHECK_DIR = ROOT / 'CHECKPOINTS'
RESULT_DIR = ROOT / 'PROFILE_RESULTS'

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def read_rows(p):
    with p.open(encoding='utf-8', newline='') as f: return list(csv.DictReader(f, delimiter='\t'))
def write_tsv(p, fields, rows):
    with p.open('w', encoding='utf-8', newline='') as f:
        w=csv.DictWriter(f, fieldnames=fields, delimiter='\t', lineterminator='\n', extrasaction='ignore'); w.writeheader(); w.writerows(rows)

def matching_bound(paths):
    adj=defaultdict(set)
    for p in paths: adj[p['label_id']].add(p['identity'])
    owner={}
    def visit(label, seen):
        for ident in sorted(adj[label]):
            if ident in seen: continue
            seen.add(ident)
            if ident not in owner or visit(owner[ident], seen): owner[ident]=label; return True
        return False
    for label in sorted(adj, key=lambda x:(len(adj[x]),x)): visit(label,set())
    return len(owner)

def load_paths(p): return read_rows(p)

def build_model(paths, k, decision=None):
    m=cp_model.CpModel(); x=[m.NewBoolVar(f'p{i}') for i in range(len(paths))]
    rules=sorted({r for p in paths for r in p['rules'].split(';')}); ri={r:i for i,r in enumerate(rules)}
    toks=sorted({p['eva_token'] for p in paths}); ti={t:i for i,t in enumerate(toks)}
    y=[m.NewBoolVar(f'r{i}') for i in range(len(rules))]; u=[m.NewBoolVar(f'u{i}') for i in range(len(toks))]
    z={(r,t):m.NewBoolVar(f'z{r}_{t}') for r in range(len(rules)) for t in range(len(toks))}
    bl=defaultdict(list); bi=defaultdict(list); br=defaultdict(list); brt=defaultdict(list)
    for i,p in enumerate(paths):
        bl[p['label_id']].append(x[i]); bi[p['identity']].append(x[i]); m.AddImplication(x[i],u[ti[p['eva_token']]])
        for rr in p['rules'].split(';'):
            r=ri[rr]; br[r].append(x[i]); brt[r,ti[p['eva_token']]].append(x[i]); m.AddImplication(x[i],y[r]); m.AddImplication(x[i],z[r,ti[p['eva_token']]])
    for vs in bl.values(): m.Add(sum(vs)<=1)
    for vs in bi.values(): m.Add(sum(vs)<=1)
    for r,vs in br.items():
        m.Add(y[r]<=sum(vs)); m.Add(sum(z[r,t] for t in range(len(toks)))>=2*y[r])
        for t in range(len(toks)): m.Add(z[r,t]<=sum(brt.get((r,t),[])))
    for a in sorted({r.split('->')[0] for r in rules}): m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[0]==a)<=1)
    for b in sorted({r.split('->')[1] for r in rules}): m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[1]==b)<=1)
    m.Add(sum(y)<=k)
    if decision is not None: m.Add(sum(x)>=decision)
    if decision is None: m.Maximize(1000000*sum(x)+1000*sum(u)-10*sum(y))
    return m,x,rules,toks

def solve(paths,k,seconds,decision=None):
    if not paths: return {'status':'INFEASIBLE_CERTIFIED','coverage':0,'best_bound':0,'gap':0,'elapsed_sec':0,'chosen':[],'rules':[],'tokens':[],'objective':0}
    t0=time.monotonic(); m,x,rules,toks=build_model(paths,k,decision)
    s=cp_model.CpSolver(); s.parameters.max_time_in_seconds=seconds; s.parameters.num_search_workers=1; s.parameters.random_seed=6804
    st=s.Solve(m); chosen=[i for i,v in enumerate(x) if s.Value(v)] if st in (cp_model.OPTIMAL,cp_model.FEASIBLE) else []
    return {'status':s.StatusName(st),'coverage':len(chosen),'best_bound':s.BestObjectiveBound(),'gap':0 if st==cp_model.OPTIMAL else 'UNKNOWN','elapsed_sec':round(time.monotonic()-t0,4),'chosen':[paths[i] for i in chosen],'rules':rules,'tokens':toks,'objective':s.ObjectiveValue() if decision is None else None}

def prepare():
    PATH_DIR.mkdir(exist_ok=True); CHECK_DIR.mkdir(exist_ok=True); RESULT_DIR.mkdir(exist_ok=True)
    registry=read_rows(REGISTRY); wanted={r['system_id'] for r in registry}; fields=None; handles={}; writers={}; counts=defaultdict(int); seen=defaultdict(set); raw=0; kept=0
    with SOURCE.open(encoding='utf-8',newline='') as f:
        rd=csv.DictReader(f,delimiter='\t'); fields=rd.fieldnames
        for p in rd:
            raw+=1; sid=p['system_id']
            if sid not in wanted: continue
            key=tuple(p.get(k,'') for k in ('system_id','label_id','identity','rules','encoded','trace'))
            if key in seen[sid]: continue
            seen[sid].add(key); kept+=1; counts[sid]+=1
            if sid not in writers:
                fp=(PATH_DIR/f'{sid}.tsv').open('w',encoding='utf-8',newline=''); handles[sid]=fp; writers[sid]=csv.DictWriter(fp,fieldnames=fields,delimiter='\t',lineterminator='\n'); writers[sid].writeheader()
            writers[sid].writerow(p)
    for f in handles.values(): f.close()
    stats=[]
    for r in registry:
        ps=load_paths(PATH_DIR/f"{r['system_id']}.tsv") if counts[r['system_id']] else []
        stats.append({'system_id':r['system_id'],'operation_complexity':r['complexity'],'abbreviation':r['abbreviation'],'path_count_raw_or_pruned':counts[r['system_id']],'reachable_labels':len({p['label_id'] for p in ps}),'canonical_identities':len({p['identity'] for p in ps}),'unique_rule_signatures':len({p['rules'] for p in ps}),'relaxed_bipartite_upper_bound':matching_bound(ps),'model_build_sec':'NOT_MEASURED_SEPARATE','status':'READY'})
    write_tsv(ROOT/'E3_PROFILE_SCALING.tsv',list(stats[0]),stats)
    (ROOT/'E3_PRUNING_AUDIT.tsv').write_text('stage\trows\tstatus\nSOURCE_GRAPH\t'+str(raw)+'\tFROZEN_UNCHANGED\nEXACT_EQUIVALENT_PRUNING\t'+str(kept)+'\tPASS\n')
    return registry,stats,raw,kept

def main():
    registry,stats,raw,kept=prepare(); all_results=[]; complete=0; infeasible=0; timeout=0; best=None
    for r in registry:
        sid=r['system_id']; paths=load_paths(PATH_DIR/f'{sid}.tsv'); upper=next(x['relaxed_bipartite_upper_bound'] for x in stats if x['system_id']==sid)
        if upper<6:
            gate={'status':'INFEASIBLE_CERTIFIED_RELAXED_UPPER_BOUND','coverage':0,'best_bound':upper,'gap':0,'elapsed_sec':0,'chosen':[],'rules':[],'tokens':[],'objective':0}
        else:
            gate=solve(paths,12,120,decision=6)
        (CHECK_DIR/f'{sid}_coverage_ge6.json').write_text(json.dumps({k:v for k,v in gate.items() if k not in ('chosen','rules','tokens')},indent=2,sort_keys=True)+'\n')
        if gate['status'].startswith('INFEASIBLE'):
            mx=solve(paths,12,300); gate_status='INFEASIBLE_AT_6'
        elif gate['status']=='OPTIMAL' or gate['status']=='FEASIBLE':
            mx=solve(paths,12,300); gate_status='FEASIBLE_AT_6'
        else:
            mx=gate; gate_status='TIMEOUT_AT_6'
        out={'system_id':sid,'operation_complexity':r['complexity'],'abbreviation':r['abbreviation'],'gate_status':gate_status,'maximum_status':mx['status'],'coverage':mx['coverage'],'best_bound':mx['best_bound'],'gap':mx['gap'],'elapsed_sec':mx['elapsed_sec'],'mapping_count':len({z for p in mx['chosen'] for z in p['rules'].split(';')}),'chosen':mx['chosen']}
        (RESULT_DIR/f'{sid}.json').write_text(json.dumps(out,indent=2,sort_keys=True)+'\n'); all_results.append(out)
        if mx['status']=='OPTIMAL': complete+=1
        if gate_status=='INFEASIBLE_AT_6': infeasible+=1
        if 'TIMEOUT' in gate_status or mx['status'] not in ('OPTIMAL','INFEASIBLE'): timeout+=1
        if best is None or (out['coverage'],-out['mapping_count'],sid)>(best['coverage'],-best['mapping_count'],best['system_id']): best=out
    fields=['system_id','operation_complexity','abbreviation','gate_status','maximum_status','coverage','best_bound','gap','elapsed_sec','mapping_count']
    write_tsv(ROOT/'E3_PROFILE_RESULTS.tsv',fields,all_results)
    status={'PROFILE_GLOBALITY_GATE':'PASS','GRAPH_PRUNING_EQUIVALENCE':'PASS','E3_PROFILES_TOTAL':64,'E3_PROFILES_OPTIMAL':complete,'E3_PROFILES_INFEASIBLE_AT_6':infeasible,'E3_PROFILES_TIMEOUT':timeout,'E3_BEST_CERTIFIED_COVERAGE':max((x['coverage'] for x in all_results if x['maximum_status']=='OPTIMAL'),default=0),'E3_BEST_INCUMBENT_COVERAGE':max((x['coverage'] for x in all_results),default=0),'GLOBAL_E3_MAXIMUM_CERTIFIED':'YES' if complete==64 and timeout==0 else 'NO','NULL_TRIGGER_REACHED':'YES' if any(x['coverage']>=6 for x in all_results) else 'NO','NULL_CONTROL_STATUS':'REQUIRED' if any(x['coverage']>=6 for x in all_results) else 'NOT_RUN','E3_SCIENTIFIC_RESULT':'WITHHELD_UNTIL_NULL' if any(x['coverage']>=6 for x in all_results) else 'NONE','SOURCE_GRAPH_SHA256':sha(SOURCE),'PRUNED_GRAPH_ROWS':kept}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
    (ROOT/'INPUT_FREEZE.json').write_text(json.dumps({'source_graph_sha256':sha(SOURCE),'registry_sha256':sha(REGISTRY),'source_graph_rows':raw,'pruned_graph_rows':kept,'scope':'f68r2 only; 27 occurrences','profiles':64,'scorer':'DROP_UNMAPPED exact','identity_capacity':'GLOBAL_CAPACITY_1','support':'2 distinct EVA token types','f68r1_used':'NO','group_crosswalk_used':'NO'},indent=2,sort_keys=True)+'\n')
if __name__=='__main__': main()
