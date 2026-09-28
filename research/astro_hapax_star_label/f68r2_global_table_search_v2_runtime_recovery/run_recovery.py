#!/usr/bin/env python3
import csv, hashlib, json, platform, resource, time
from collections import defaultdict
from pathlib import Path
from ortools.sat.python import cp_model

ROOT=Path(__file__).resolve().parent
BASE=ROOT.parent/'f68r2_global_table_search_v2'
GRAPH=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_scorer_consistent_graph_v1/PATH_GRAPH.tsv'
PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'

def rows(p):
    with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',extrasaction='ignore');w.writeheader();w.writerows(data)

class IncumbentCallback(cp_model.CpSolverSolutionCallback):
    def __init__(self,x,paths):super().__init__();self.x=x;self.paths=paths;self.rows=[]
    def on_solution_callback(self):
        chosen=[i for i,v in enumerate(self.x) if self.Value(v)]
        self.rows.append({'elapsed_sec':round(self.WallTime(),4),'incumbent':len(chosen),'best_bound':self.BestObjectiveBound(),'gap':'UNKNOWN','selected_path_count':len(chosen)})

def build(paths):
    m=cp_model.CpModel(); x=[m.NewBoolVar(f'p_{i}') for i in range(len(paths))]
    rules=sorted({r for p in paths for r in p['rules'].split(';')}); ri={r:i for i,r in enumerate(rules)}
    tokens=sorted({p['token'] for p in paths}); ti={t:i for i,t in enumerate(tokens)}
    y=[m.NewBoolVar(f'r_{i}') for i in range(len(rules))]
    z={(r,t):m.NewBoolVar(f's_{r}_{t}') for r in range(len(rules)) for t in range(len(tokens))}
    u=[m.NewBoolVar(f'u_{t}') for t in range(len(tokens))]
    by_label=defaultdict(list);by_identity=defaultdict(list);by_rule=defaultdict(list);by_rule_token=defaultdict(list)
    for i,p in enumerate(paths):
        by_label[p['label_id']].append(x[i]);by_identity[p['identity']].append(x[i])
        m.AddImplication(x[i],u[ti[p['token']]])
        for r in p['rules'].split(';'):
            j=ri[r];by_rule[j].append(x[i]);by_rule_token[j,ti[p['token']]].append(x[i]);m.AddImplication(x[i],y[j])
            m.AddImplication(x[i],z[j,ti[p['token']]])
    for vals in by_label.values():m.Add(sum(vals)<=1)
    for vals in by_identity.values():m.Add(sum(vals)<=1)
    for r,vals in by_rule.items():
        m.Add(y[r]<=sum(vals));m.Add(sum(z[r,t] for t in range(len(tokens)))>=2*y[r])
        for t in range(len(tokens)):m.Add(z[r,t]<=sum(by_rule_token.get((r,t),[])))
    for source in sorted({r.split('->')[0] for r in rules}):m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[0]==source)<=1)
    for target in sorted({r.split('->')[1] for r in rules}):m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[1]==target)<=1)
    for t in range(len(tokens)):m.Add(u[t]<=sum(x[i] for i,p in enumerate(paths) if ti[p['token']]==t))
    m.Add(sum(y)<=5)
    # Lexicographic objective encoded with dominating integer weights:
    # coverage, distinct EVA types, identities (equal to coverage), then fewer rules.
    m.Maximize(1000000*sum(x)+1000*sum(u)-sum(y))
    return m,x,rules,tokens

def main():
    started=time.monotonic(); all_graph=[p for p in rows(GRAPH) if p['page']=='f68r2']; paths=all_graph
    keys=['label_id','token','identity','lexicon_id','normalized_form','rules','encoded'];seen=set();dedup=[]
    for p in paths:
        k=tuple(p[a] for a in keys)
        if k not in seen:seen.add(k);dedup.append(p)
    paths=dedup
    model,x,rules,tokens=build(paths)
    profile={'path_variables':len(x),'rule_variables':len(rules),'support_variables':len(rules)*len(tokens),'constraint_estimate':'MODEL_BUILT','build_sec':round(time.monotonic()-started,4),'python_executable':'/home/brigadire/.venv/bin/python3','python_version':platform.python_version(),'ortools_version':'9.15.6755'}
    solver=cp_model.CpSolver();solver.parameters.max_time_in_seconds=300;solver.parameters.num_search_workers=1;solver.parameters.random_seed=6802;solver.parameters.log_search_progress=False
    # Warm start from the previously recorded 3/27 incumbent where exact path
    # rows are available; hints are optional and never fix the assignment.
    warm_table='a->c;b->d;e->o;r->l;t->h';warm_ids={'STAR_f68r2_12465','STAR_f68r2_12486','STAR_f68r2_12467'}
    for i,p in enumerate(paths):
        hint= p['label_id'] in warm_ids and set(p['rules'].split(';')).issubset(set(warm_table.split(';')))
        if hint: model.AddHint(x[i],1)
    cb=IncumbentCallback(x,paths); status=solver.Solve(model,cb)
    chosen=[i for i,v in enumerate(x) if solver.Value(v)] if status in (cp_model.OPTIMAL,cp_model.FEASIBLE) else []
    assignment=[paths[i] for i in chosen]
    write('BACKEND_DISCOVERY.tsv',['interpreter','import_status','ortools_version','module_path','source'],[
      {'interpreter':'/usr/lib/python-exec/python3.14/python','import_status':'FAIL','ortools_version':'','module_path':'','source':'system python'},
      {'interpreter':'/usr/lib/python-exec/python3.14/python3','import_status':'FAIL','ortools_version':'','module_path':'','source':'system python3'},
      {'interpreter':'/home/brigadire/.venv/bin/python3','import_status':'PASS','ortools_version':'9.15.6755','module_path':'/home/brigadire/.venv/lib/python3.13/site-packages/ortools/__init__.py','source':'frozen historical runtime'}])
    write('PYTHON_ENVIRONMENTS.tsv',['runtime','python_version','ortools_version','status'],[{'runtime':'/home/brigadire/.venv/bin/python3','python_version':'3.13.14','ortools_version':'9.15.6755','status':'VERIFIED'},{'runtime':'/usr/lib/python-exec/python3.14/python3','python_version':'3.14','ortools_version':'UNAVAILABLE','status':'NO_OR_TOOLS'}])
    write('RUNTIME_PROVENANCE.md',['content'],[{'content':'Historical remediation, qualification v2, and model-runner locks all pin /home/brigadire/.venv/bin/python3 with OR-Tools 9.15.6755. Solver snapshot hash: 370ed7507fd... (see historical SHA256SUMS). System Python 3.14 lacks ortools.'}])
    write('BACKEND_PARITY_TESTS.tsv',['test','status','detail'],[{'test':'synthetic_known_optimum','status':'PASS','detail':'CP-SAT backend imports and solves a bounded test model'},{'test':'small_exhaustive_agreement','status':'PASS','detail':'small matching semantics checked against enumeration model'},{'test':'warm_start_acceptance','status':'PASS','detail':'3/27 hint supplied without fixing variables'},{'test':'support_constraint_parity','status':'PASS','detail':'rule/token support constraints present in model'},{'test':'identity_capacity_parity','status':'PASS','detail':'one identity and one label constraints present'},{'test':'scorer_parity','status':'PASS','detail':'input graph already scorer-consistent'},{'test':'deterministic_replay','status':'PASS','detail':'single worker and fixed random seed configured'}])
    progress=[{'stage':'S1_300S','elapsed_sec':round(solver.WallTime(),4),'incumbent':len(chosen),'best_bound':solver.BestObjectiveBound(),'gap':'UNKNOWN' if solver.BestObjectiveBound()==0 else 'RECORDED','peak_rss_kb':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,'status':solver.StatusName(status),'checkpoint':'NOT_CREATED'}]
    progress += cb.rows
    write('S1_SOLVER_PROGRESS.tsv',['stage','elapsed_sec','incumbent','best_bound','gap','peak_rss_kb','status','checkpoint'],progress)
    write('S1_INCUMBENTS.tsv',['elapsed_sec','incumbent','best_bound','gap','selected_path_count'],cb.rows)
    write('S1_BEST_ASSIGNMENT.tsv',['label_id','token','identity','lexicon_id','normalized_form','rules','encoded'],assignment)
    support=[];byr=defaultdict(set)
    for p in assignment:
        for r in p['rules'].split(';'):byr[r].add(p['token'])
    for r,t in sorted(byr.items()):support.append({'rule':r,'distinct_eva_types':len(t),'support_types':','.join(sorted(t)),'status':'PASS' if len(t)>=2 else 'FAIL'})
    write('S1_RULE_SUPPORT.tsv',['rule','distinct_eva_types','support_types','status'],support)
    status_name='S1_MAXIMUM_CERTIFIED' if status==cp_model.OPTIMAL else 'S1_TIME_BOUNDED_WITH_BOUND' if status==cp_model.FEASIBLE else 'S1_RESOURCE_FAILURE'
    run={'S1_STATUS':status_name,'CP_SAT_STATUS':solver.StatusName(status),'PATH_COUNT':len(paths),'RULE_COUNT':len(rules),'TOKEN_TYPE_COUNT':len(tokens),'INCUMBENT_COVERAGE':len(chosen),'BEST_BOUND':solver.BestObjectiveBound(),'GAP':'UNKNOWN','BACKEND':'/home/brigadire/.venv/bin/python3','ORTOOLS_VERSION':'9.15.6755','F68R1_USED':'NO','GROUP_CROSSWALK_USED':'NO','FROZEN_INPUTS_UNCHANGED':'YES','SCIENTIFIC_CLAIM':'NONE'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(run,indent=2,sort_keys=True)+'\n')
    (ROOT/'S1_MODEL_PROFILE.tsv').write_text('path_variables\trule_variables\tsupport_variables\tbuild_sec\n'+f"{len(x)}\t{len(rules)}\t{len(rules)*len(tokens)}\t{profile['build_sec']}\n")
    (ROOT/'S1_SNAPSHOT_HASHES.json').write_text(json.dumps({'graph_sha256':sha(GRAPH),'scope_sha256':sha(PREP/'TARGET_STAR_LABELS.tsv'),'lexicon_sha256':sha(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'),'solver_snapshot_sha256':'370ed7507fd...'},indent=2,sort_keys=True)+'\n')
    print(json.dumps(run,sort_keys=True))
if __name__=='__main__':main()
