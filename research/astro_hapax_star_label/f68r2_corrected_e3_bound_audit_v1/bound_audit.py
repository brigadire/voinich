#!/usr/bin/env python3
"""Audit old E3 bounds and independently solve coverage >= 4 for UNKNOWN profiles."""
import csv, hashlib, json, os, sys, time
from pathlib import Path

HERE=Path(__file__).resolve().parent
OLD=HERE.parent/'f68r2_corrected_e3_resumable_run_v1'
PREV=HERE.parent/'f68r2_corrected_e3_unknown_closure_v1'
sys.path.insert(0,str(PREV))
import closure
from ortools.sat.python import cp_model

def write(path, fields, rows):
    with path.open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore');w.writeheader();w.writerows(rows)
def write_atomic(path, fields, rows):
    tmp=path.with_name('.'+path.name+'.tmp')
    write(tmp,fields,rows)
    os.replace(tmp,path)
def read(path): return closure.read_tsv(path)
def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()
def jwrite(path,obj): path.write_text(json.dumps(obj,indent=2,sort_keys=True)+'\n')

def raw_audit(imported):
    rows=[]
    for r in imported:
        if r['solver_status']!='UNKNOWN': continue
        st=json.loads((OLD/'profiles'/r['profile_id']/'PROFILE_STATUS.json').read_text())
        rows.append({
            'profile_id':r['profile_id'],'status_raw':st.get('solver_status'),'objective_value_raw':'MISSING',
            'best_objective_bound_raw':st.get('bound'),'objective_direction':'MAXIMIZE sum(x) (source code; raw response missing)',
            'scaling_factor':'1','objective_offset':'0','objective_formula':'sum(selected_path_variables)',
            'variables':'MISSING_IN_CHECKPOINT','constraints':'MISSING_IN_CHECKPOINT','branches':'MISSING_IN_CHECKPOINT',
            'conflicts':'MISSING_IN_CHECKPOINT','presolve_time':'MISSING_IN_CHECKPOINT','search_time':'MISSING_IN_CHECKPOINT',
            'stop_reason':'MISSING_IN_CHECKPOINT','solution_info':'MISSING_IN_CHECKPOINT','raw_response':'NOT_SAVED',
            'conversion':'coverage = objective; bound_coverage = best_objective_bound (only if solver response is valid)'
        })
    return rows

def sentinel_tests():
    rows=[]
    # A known positive model. With no time to search CP-SAT can return UNKNOWN
    # while its reported bound remains a default-looking zero.
    for name,budget in [('before_search',0.0),('zero_budget',0.0),('tiny_budget',1e-9),('short_budget',1e-3)]:
        m=cp_model.CpModel(); x=m.NewBoolVar('x'); m.Maximize(x)
        s=cp_model.CpSolver(); s.parameters.max_time_in_seconds=budget; s.parameters.num_search_workers=1
        st=s.Solve(m)
        rows.append({'test':name,'known_optimum':1,'budget_s':budget,'status':s.StatusName(st),'objective_value':s.ObjectiveValue(),'best_objective_bound':s.BestObjectiveBound(),'proves_upper_bound_0':'YES' if s.StatusName(st)=='INFEASIBLE' else 'NO','interpretation':'sentinel/default behavior test'})
    return rows

def classify(raw):
    out=[]
    for r in raw:
        # The old checkpoint does not contain the solver response needed to
        # establish provenance of best_bound. UNKNOWN is therefore metadata
        # incomplete, even though the numeric field is 0.0.
        out.append({'profile_id':r['profile_id'],'raw_status':r['status_raw'],'raw_bound':r['best_objective_bound_raw'],'classification':'BOUND_METADATA_MISSING','reason':'UNKNOWN status and no raw solver response/objective metadata; numeric 0.0 is not accepted as certificate','coverage_conversion':'not certifiable from checkpoint'})
    return out

def main():
    HERE.mkdir(parents=True,exist_ok=True)
    imported,ok=closure.verify_old()
    raw=raw_audit(imported); write(HERE/'RAW_BOUND_AUDIT.tsv',list(raw[0]),raw)
    write(HERE/'BOUND_CLASSIFICATION.tsv',list(classify(raw)[0]),classify(raw))
    write(HERE/'UNKNOWN_ZERO_BOUND_SENTINEL_TESTS.tsv',list(sentinel_tests()[0]),sentinel_tests())
    (HERE/'OBJECTIVE_DIRECTION_AND_SCALING.md').write_text('''# Objective direction and scaling\n\nThe old source model (`f68r2_corrected_e3_search_v1/e3_search.py`) calls `m.Maximize(sum(x))`, where each selected path variable has coefficient 1. Therefore, if a genuine CP-SAT response were available, objective and coverage would have conversion `coverage = objective_value`, and an upper bound would convert as `coverage_upper = best_objective_bound`; scaling is 1 and offset 0.\n\nThe committed `PROFILE_STATUS.json` files preserve only `solver_status`, `incumbent`, and `bound`. They do not preserve objective value, response protobuf, solution info, branches, conflicts, presolve/search timing, or stop reason. Consequently the 24 numeric `0.0` values cannot be independently classified as certified upper bounds. They are classified `BOUND_METADATA_MISSING`.\n\nThe sentinel tests exercise known-positive models under zero and tiny budgets. Their output is the authoritative test of whether `UNKNOWN` plus a zero-looking bound can arise without an infeasibility certificate.\n''')
    unknown=[r['profile_id'] for r in imported if r['solver_status']=='UNKNOWN']
    done={r['profile_id'] for r in read(HERE/'GE4_DIRECT_RESULTS.tsv')} if (HERE/'GE4_DIRECT_RESULTS.tsv').exists() else set()
    result_path=HERE/'GE4_DIRECT_RESULTS.tsv'; replay_path=HERE/'SCORER_REPLAY.tsv'; witness_path=HERE/'VALID_WITNESSES.tsv'
    results=read(result_path) if result_path.exists() else []
    replay=read(replay_path) if replay_path.exists() else []
    witnesses=read(witness_path) if witness_path.exists() else []
    fields=['profile_id','status','coverage','best_bound','objective_value','graph_hash','model_fingerprint','wall_time_s','build_time_s','solve_time_s','variables','constraints','proof_metadata','replay_status']
    for pid in unknown:
        if pid in done: continue
        graph=OLD/'profiles'/pid/'PATH_GRAPH.tsv'; paths=read(graph); graph_hash=sha(graph)
        model_fp=hashlib.sha256((graph_hash+'|threshold=4|support=2 distinct EVA types|global_mapping=uncapped|DROP_UNMAPPED').encode()).hexdigest()
        r=closure.model(paths,4,60,False)
        feasible=r['status'] in ('FEASIBLE','OPTIMAL') and r['coverage']>=4
        rs='PASS' if r['replay'] else 'FAIL'
        row={'profile_id':pid,'status':r['status'],'coverage':r['coverage'],'best_bound':r['bound'],'objective_value':'NOT_APPLICABLE_DECISION','graph_hash':graph_hash,'model_fingerprint':model_fp,'wall_time_s':r['build_s']+r['solve_s'],'build_time_s':r['build_s'],'solve_time_s':r['solve_s'],'variables':r['vars'],'constraints':r['constraints'],'proof_metadata':'INFEASIBLE_CERTIFIED' if r['status']=='INFEASIBLE' else ('FEASIBLE_REPLAY_VALIDATED' if feasible else 'UNKNOWN_TIMEOUT'),'replay_status':rs}
        results.append(row); replay.append({'profile_id':pid,'status':rs,'coverage':r['coverage'],'note':'independent replay of decision result'})
        if feasible:
            for p in r['chosen']: witnesses.append({'profile_id':pid,'label_id':p['label_id'],'eva_token':p['eva_token'],'identity':p['identity'],'required_mappings':p['required_mappings'],'forbidden_source_units':p['forbidden_source_units'],'expected_output':p['expected_output']})
        # Persist after every profile so an interruption loses at most the current profile.
        write_atomic(result_path,fields,results)
        write_atomic(replay_path,['profile_id','status','coverage','note'],replay)
        write_atomic(witness_path,['profile_id','label_id','eva_token','identity','required_mappings','forbidden_source_units','expected_output'],witnesses)
    if not result_path.exists(): write_atomic(result_path,fields,[])
    if not replay_path.exists(): write_atomic(replay_path,['profile_id','status','coverage','note'],[])
    if not witness_path.exists(): write_atomic(witness_path,['profile_id','label_id','eva_token','identity','required_mappings','forbidden_source_units','expected_output'],[])
    ledger=[]
    for r in results: ledger.append({'profile_id':r['profile_id'],'state':'INFEASIBLE_CERTIFIED' if r['status']=='INFEASIBLE' else ('FEASIBLE_REPLAY_VALIDATED' if r['status'] in ('FEASIBLE','OPTIMAL') and r['replay_status']=='PASS' else 'UNKNOWN_TIMEOUT'),'status':r['status'],'coverage':r['coverage'],'bound':r['best_bound'],'graph_hash':r['graph_hash'],'checkpoint_hash':hashlib.sha256(json.dumps(r,sort_keys=True).encode()).hexdigest()})
    write(HERE/'PROFILE_CLOSURE_LEDGER.tsv',['profile_id','state','status','coverage','bound','graph_hash','checkpoint_hash'],ledger)
    feasible=[r for r in results if r['status'] in ('FEASIBLE','OPTIMAL') and r['replay_status']=='PASS' and int(r['coverage'])>=4]
    unresolved=[r for r in results if r['status']=='UNKNOWN']
    final_cert=ok and not unresolved and not feasible
    status={'IMPORTED_CHECKPOINTS_VERIFIED':'YES' if ok else 'NO','UNKNOWN_PROFILES_INITIAL':24,'PROFILES_CLOSED_BY_BOUND':0,'PROFILES_TESTED_GE4':len(results),'GE4_INFEASIBLE_CERTIFIED':sum(r['status']=='INFEASIBLE' for r in results),'GE4_FEASIBLE':len(feasible),'PROFILES_REMAINING_UNKNOWN':len(unresolved),'BEST_VALID_INCUMBENT':4 if feasible else 3,'GLOBAL_E3_MAXIMUM':3 if final_cert else 'NOT_ESTABLISHED','GLOBAL_E3_MAXIMUM_CERTIFIED':'YES' if final_cert else 'NO','MODEL_SCORER_PARITY':'PASS' if all(r['replay_status']=='PASS' for r in results) else 'FAIL','NULL_TRIGGER':4,'NULL_RUN_REQUIRED':'YES' if feasible else 'NO','NULL_RUN_AUTHORIZED':'NO','BOUND_CERTIFICATES_ACCEPTED':0,'E3_SCIENTIFIC_RESULT':'NONE'}
    jwrite(HERE/'RUN_STATUS.json',status)
    print(json.dumps(status,indent=2))

if __name__=='__main__': main()
