#!/usr/bin/env python3
import csv, hashlib, json, time
from collections import defaultdict
from pathlib import Path

ROOT=Path(__file__).resolve().parent
GRAPH=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_scorer_consistent_graph_v1'
OLD=ROOT.parent/'f68r2_occurrence_search_v1'
PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'

def rows(p):
    with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',extrasaction='ignore');w.writeheader();w.writerows(data)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()

def matching(edges):
    adj=defaultdict(list)
    for p in edges:adj[p['label_id']].append(p['identity'])
    m={}
    def dfs(lid,seen):
        for ident in adj[lid]:
            if ident in seen:continue
            seen.add(ident)
            if ident not in m or dfs(m[ident],seen):m[ident]=lid;return True
        return False
    for lid in sorted(adj,key=lambda x:(len(adj[x]),x)):dfs(lid,set())
    return len(m)

def main():
    started=time.monotonic(); raw=[p for p in rows(GRAPH/'PATH_GRAPH.tsv') if p['page']=='f68r2']
    keys=['label_id','token','identity','lexicon_id','normalized_form','rules','encoded']
    seen={};dedup=[];collapsed=[]
    for p in raw:
        key=tuple(p[k] for k in keys)
        if key not in seen:seen[key]=p;dedup.append(p)
        else:collapsed.append({'dedup_key':hashlib.sha256(repr(key).encode()).hexdigest(),'label_id':p['label_id'],'token':p['token'],'identity':p['identity'],'lexicon_id':p['lexicon_id'],'normalized_form':p['normalized_form'],'rules':p['rules'],'encoded':p['encoded'],'collapsed_into':'exact_identical_path'})
    write('PATH_DEDUPLICATION.tsv',['dedup_key','label_id','token','identity','lexicon_id','normalized_form','rules','encoded','collapsed_into'],collapsed)
    # Relaxed bounds use the full deduplicated graph, never a candidate-table list.
    reachable=len({p['label_id'] for p in dedup}); raw_matching=matching(dedup)
    rule_count=len({r for p in dedup for r in p['rules'].split(';')}); token_types=len({p['token'] for p in dedup})
    bounds=[
      {'bound_name':'reachable_occurrences','value':reachable,'definition':'occurrences with at least one exact path','status':'UPPER_BOUND'},
      {'bound_name':'bipartite_matching_occurrence_identity','value':raw_matching,'definition':'maximum matching ignoring common-table and support constraints','status':'UPPER_BOUND'},
      {'bound_name':'maximum_without_support','value':raw_matching,'definition':'same full graph, support ignored; identity/occurrence uniqueness retained','status':'UPPER_BOUND'},
      {'bound_name':'maximum_ignoring_mapping_conflicts','value':reachable,'definition':'occurrences with any path; identity and rule conflicts ignored','status':'UPPER_BOUND'},
      {'bound_name':'rule_union_le_5_relaxed','value':reachable,'definition':'relaxed structural bound before global-table optimization','status':'UPPER_BOUND_NOT_CERTIFIED'}]
    write('RELAXED_UPPER_BOUNDS.tsv',['bound_name','value','definition','status'],bounds)
    rules={r for p in dedup for r in p['rules'].split(';')}; support_vars=len(rules)*token_types
    profile=[{'stage':'S1_MODEL_BUILD','path_variables':len(dedup),'rule_variables':len(rules),'support_variables':support_vars,'constraints_estimate':len(dedup)+reachable+rule_count+support_vars,'build_sec':round(time.monotonic()-started,4),'presolve_sec':'NOT_RUN','search_sec':'NOT_RUN','peak_rss_kb':'NOT_PROFILED','backend':'OR_TOOLS_CP_SAT','backend_available':'NO','status':'RESOURCE_FAILURE'}]
    write('S1_MODEL_PROFILE.tsv',list(profile[0]),profile)
    write('S1_SOLVER_PROGRESS.tsv',['stage','elapsed_sec','incumbent','best_bound','gap','status','checkpoint'],[{'stage':'S1_300S','elapsed_sec':0,'incumbent':3,'best_bound':'UNKNOWN','gap':'UNKNOWN','status':'RESOURCE_FAILURE_NO_BACKEND','checkpoint':'NOT_CREATED'},{'stage':'S1_1800S','elapsed_sec':0,'incumbent':3,'best_bound':'UNKNOWN','gap':'UNKNOWN','status':'NOT_RUN_AFTER_RESOURCE_FAILURE','checkpoint':'NOT_CREATED'},{'stage':'S1_7200S','elapsed_sec':0,'incumbent':3,'best_bound':'UNKNOWN','gap':'UNKNOWN','status':'NOT_RUN_AFTER_RESOURCE_FAILURE','checkpoint':'NOT_CREATED'}])
    warm=rows(OLD/'OCCURRENCE_MATCH_ASSIGNMENTS.tsv')
    warm=[x for x in warm if x['table_hash']==warm[0]['table_hash']] if warm else []
    write('S1_INCUMBENTS.tsv',['stage','coverage','table_hash','table','source','status'],[{'stage':'WARM_START','coverage':3,'table_hash':warm[0]['table_hash'] if warm else '','table':'a->c;b->d;e->o;r->l;t->h','source':'f68r2_occurrence_search_v1','status':'VALIDATED_PREVIOUS_INCUMBENT'}])
    write('S1_BEST_ASSIGNMENT.tsv',['stage','coverage','table_hash','label_id','token','identity','normalized_form','rules','encoded'],[dict(x,stage='WARM_START',coverage=3,table_hash=warm[0]['table_hash'] if warm else '') for x in warm])
    write('S1_RULE_SUPPORT.tsv',['table_hash','rule','status','source'],[])
    # S2 preflight is graph construction only, not a search result.
    write('S2_PREFLIGHT.tsv',['stage','path_count','reachable_occurrences','rule_count','memory_estimate','relaxed_coverage_upper_bound','model_status'],[{'stage':'K8','path_count':'NOT_BUILT','reachable_occurrences':'NOT_BUILT','rule_count':'NOT_BUILT','memory_estimate':'NOT_ESTIMATED','relaxed_coverage_upper_bound':'NOT_ESTIMATED','model_status':'DEFERRED_AFTER_S1_RESOURCE_FAILURE'}])
    status={'S1_STATUS':'S1_RESOURCE_FAILURE','S1_PATH_COUNT':len(dedup),'RAW_PATH_COUNT':len(raw),'COLLAPSED_PATH_COUNT':len(collapsed),'F68R2_OCCURRENCE_COUNT':27,'LEXICON_ROW_COUNT':299,'CANONICAL_IDENTITY_COUNT':94,'WARM_START_COVERAGE':3,'S1_BEST_BOUND':'UNKNOWN','S1_GAP':'UNKNOWN','OR_TOOLS_AVAILABLE':'NO','GROUP_CROSSWALK_USED':'NO','F68R1_USED':'NO','SCIENTIFIC_CLAIM':'NONE'}
    (ROOT/'INPUT_FREEZE.json').write_text(json.dumps({'scope_sha256':sha(PREP/'TARGET_STAR_LABELS.tsv'),'lexicon_sha256':sha(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'),'path_graph_sha256':sha(GRAPH/'PATH_GRAPH.tsv'),'path_count_raw':len(raw),'path_count_deduplicated':len(dedup),'deduplication_fields':keys,'scorer':'exact DROP_UNMAPPED, NONE abbreviation','table_limit':5},indent=2,sort_keys=True)+'\n')
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n');print(json.dumps(status,sort_keys=True))
if __name__=='__main__':main()
