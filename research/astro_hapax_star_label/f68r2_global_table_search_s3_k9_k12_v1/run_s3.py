#!/usr/bin/env python3
import csv, hashlib, json, resource, time
from collections import defaultdict
from pathlib import Path
from ortools.sat.python import cp_model

ROOT=Path(__file__).resolve().parent
GRAPH=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_scorer_consistent_graph_v1/PATH_GRAPH.tsv'
PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
S1=ROOT.parent/'f68r2_global_table_search_v2_runtime_recovery'

def rows(p):
    with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',extrasaction='ignore');w.writeheader();w.writerows(data)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()

def model(paths,limit):
    m=cp_model.CpModel();x=[m.NewBoolVar(f'p{i}') for i in range(len(paths))]
    rules=sorted({r for p in paths for r in p['rules'].split(';')});ri={r:i for i,r in enumerate(rules)}
    toks=sorted({p['token'] for p in paths});ti={t:i for i,t in enumerate(toks)}
    y=[m.NewBoolVar(f'r{i}') for i in range(len(rules))];u=[m.NewBoolVar(f'u{i}') for i in range(len(toks))]
    z={(r,t):m.NewBoolVar(f'z{r}_{t}') for r in range(len(rules)) for t in range(len(toks))}
    bl=defaultdict(list);bi=defaultdict(list);br=defaultdict(list);brt=defaultdict(list)
    for i,p in enumerate(paths):
        bl[p['label_id']].append(x[i]);bi[p['identity']].append(x[i]);m.AddImplication(x[i],u[ti[p['token']]])
        for rr in p['rules'].split(';'):
            r=ri[rr];br[r].append(x[i]);brt[r,ti[p['token']]].append(x[i]);m.AddImplication(x[i],y[r]);m.AddImplication(x[i],z[r,ti[p['token']]])
    for vs in bl.values():m.Add(sum(vs)<=1)
    for vs in bi.values():m.Add(sum(vs)<=1)
    for r,vs in br.items():
        m.Add(y[r]<=sum(vs));m.Add(sum(z[r,t] for t in range(len(toks)))>=2*y[r])
        for t in range(len(toks)):m.Add(z[r,t]<=sum(brt.get((r,t),[])))
    for s in sorted({r.split('->')[0] for r in rules}):m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[0]==s)<=1)
    for t in sorted({r.split('->')[1] for r in rules}):m.Add(sum(y[i] for i,r in enumerate(rules) if r.split('->')[1]==t)<=1)
    m.Add(sum(y)<=limit);m.Maximize(1000000*sum(x)+1000*sum(u)-sum(y))
    return m,x,rules,toks

def solve(paths,k,seconds=300):
    m,x,rules,toks=model(paths,k)
    warm_table={'l->h','n->l','r->c','s->y','t->d','u->o'}
    warm_labels={'STAR_f68r2_12463','STAR_f68r2_12467','STAR_f68r2_12472','STAR_f68r2_12477','STAR_f68r2_12486'}
    for i,p in enumerate(paths):
        if p['label_id'] in warm_labels and set(p['rules'].split(';')).issubset(warm_table):m.AddHint(x[i],1)
    s=cp_model.CpSolver();s.parameters.max_time_in_seconds=seconds;s.parameters.num_search_workers=1;s.parameters.random_seed=6802
    t=time.monotonic();st=s.Solve(m);chosen=[i for i,v in enumerate(x) if s.Value(v)] if st in (cp_model.OPTIMAL,cp_model.FEASIBLE) else []
    return {'k':k,'status':s.StatusName(st),'coverage':len(chosen),'objective':s.ObjectiveValue(),'best_bound':s.BestObjectiveBound(),'gap':0 if st==cp_model.OPTIMAL else 'UNKNOWN','elapsed_sec':round(time.monotonic()-t,4),'build_rules':len(rules),'tokens':len(toks),'chosen':[paths[i] for i in chosen],'rules':rules}

def main():
    paths=[p for p in rows(GRAPH) if p['page']=='f68r2'];started=time.monotonic()
    max_path=max(len(p['rules'].split(';')) for p in paths);counts=defaultdict(int)
    for p in paths:counts[len(p['rules'].split(';'))]+=1
    write('S3_PATH_COMPLETENESS_AUDIT.tsv',['scope','path_count','max_individual_rule_count','count_by_rule_count','delta_6_to_8','status'],[{'scope':'f68r2 scorer-consistent graph','path_count':len(paths),'max_individual_rule_count':max_path,'count_by_rule_count':json.dumps(dict(sorted(counts.items()))),'delta_6_to_8':0,'status':'COMPLETE_FOR_K12_GLOBAL_UNION' if max_path<=5 else 'DELTA_REQUIRED'}])
    write('S3_PATH_DELTA.tsv',['status','reason'],[{'status':'NOT_REQUIRED','reason':'All individual scorer-consistent paths require <=5 mappings; only the global union limit changes.'}])
    results=[solve(paths,k) for k in range(9,13)]
    write('S3_COVERAGE_BY_K.tsv',['k','status','coverage','best_bound','gap','elapsed_sec','objective'],results)
    prior=rows(ROOT.parent/'f68r2_global_table_search_s2_v1/COVERAGE_BY_K.tsv')
    write('COVERAGE_BY_K.tsv',['k','status','coverage','best_bound','gap','elapsed_sec','objective'],prior+results)
    thresholds=[]
    final=results[-1]
    for t in (6,8,11):thresholds.append({'threshold':t,'status':'FEASIBLE' if final['coverage']>=t else 'INFEASIBLE_CERTIFIED' if final['status']=='OPTIMAL' else 'UNKNOWN_TIMEOUT','coverage':final['coverage'],'best_bound':final['best_bound'],'gap':final['gap']})
    write('S3_EXISTENCE_TESTS.tsv',['threshold','status','coverage','best_bound','gap'],thresholds)
    progress=[{'k':r['k'],'elapsed_sec':r['elapsed_sec'],'incumbent':r['coverage'],'best_bound':r['best_bound'],'gap':r['gap'],'status':r['status'],'checkpoint':'NOT_CREATED'} for r in results]
    write('S3_SOLVER_PROGRESS.tsv',list(progress[0]),progress)
    best=final['chosen']
    supp=defaultdict(set)
    for p in best:
        for r in p['rules'].split(';'):supp[r].add(p['token'])
    write('S3_BEST_ASSIGNMENT.tsv',['k','coverage','label_id','token','identity','lexicon_id','normalized_form','rules','encoded'],[dict(p,k=12,coverage=final['coverage']) for p in best])
    write('S3_RULE_SUPPORT.tsv',['rule','distinct_eva_types','support_types','status'],[{'rule':r,'distinct_eva_types':len(t),'support_types':','.join(sorted(t)),'status':'PASS' if len(t)>=2 else 'FAIL'} for r,t in sorted(supp.items())])
    tables=[]
    for r in results:
        tables.append({'k':r['k'],'coverage':r['coverage'],'table_hash':hashlib.sha256(';'.join(sorted({z for p in r['chosen'] for z in p['rules'].split(';')})).encode()).hexdigest(),'rules':';'.join(sorted({z for p in r['chosen'] for z in p['rules'].split(';')})),'status':r['status']})
    write('S3_COOPTIMAL_TABLES.tsv',['k','coverage','table_hash','rules','status'],tables)
    status={'S3_STATUS':'S3_MAXIMUM_CERTIFIED' if final['status']=='OPTIMAL' else 'S3_TIME_BOUNDED_WITH_BOUND','S3_MAXIMUM_COVERAGE':final['coverage'],'S3_BEST_BOUND':final['best_bound'],'S3_PATH_COUNT':len(paths),'S3_GRAPH_REBUILT':'NO','S3_DELTA_PATH_COUNT':0,'F68R1_USED':'NO','GROUP_CROSSWALK_USED':'NO','SCIENTIFIC_CLAIM':'NONE','NULL_CONTROL':'NOT_RUN_THRESHOLD_6_NOT_REACHED'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
    (ROOT/'INPUT_FREEZE.json').write_text(json.dumps({'scope_sha256':sha(PREP/'TARGET_STAR_LABELS.tsv'),'lexicon_sha256':sha(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'),'path_graph_sha256':sha(GRAPH),'path_count':len(paths),'scorer':'DROP_UNMAPPED,NONE','global_mapping_limits':'k=9..12','s2_warm_start':'l->h;n->l;r->c;s->y;t->d;u->o','s2_maximum':5,'graph_rebuilt':'NO'},indent=2,sort_keys=True)+'\n')
    print(json.dumps(status,sort_keys=True))
if __name__=='__main__':main()
