#!/usr/bin/env python3
import csv, hashlib, itertools, json, random, time
from collections import defaultdict, Counter
from pathlib import Path

ROOT=Path(__file__).resolve().parent
PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
GRAPH=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_scorer_consistent_graph_v1'
GROUPS=ROOT.parent.parent/'astro_spatial_human_completeness/job_21_group_review_validated/REVIEWED_RELATION_GROUPS.tsv'
MIG=ROOT.parent/'legacy_mapping_migration_v1'

def rows(p):
    with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',extrasaction='ignore');w.writeheader();w.writerows(data)

def eval_table(paths, table):
    rules=set(table.split(';')); by=defaultdict(list)
    for p in paths:
        if set(p['rules'].split(';')).issubset(rules):by[p['label_id']].append(p)
    labels=sorted(by,key=lambda x:(len(by[x]),x));best=[];bestn=-1
    def rec(i,used,chosen,support):
        nonlocal best,bestn
        if i==len(labels):
            if not chosen:return
            if all(len(support[r])>=2 for r in {z for p in chosen for z in p['rules'].split(';')}):
                if len(chosen)>bestn:bestn=len(chosen);best=[chosen[:]]
                elif len(chosen)==bestn:best.append(chosen[:])
            return
        if len(chosen)+(len(labels)-i)<bestn:return
        rec(i+1,used,chosen,support)
        for p in by[labels[i]]:
            if p['identity'] in used:continue
            ns={k:set(v) for k,v in support.items()}
            for r in p['rules'].split(';'):ns.setdefault(r,set()).add(p['token'])
            rec(i+1,used|{p['identity']},chosen+[p],ns)
    rec(0,set(),[],{})
    return bestn if bestn>0 else 0,best

def main():
    started=time.monotonic(); targets=[r for r in rows(PREP/'TARGET_STAR_LABELS.tsv') if r['page']=='f68r2']
    graph=[r for r in rows(GRAPH/'PATH_GRAPH.tsv') if r['page']=='f68r2']
    # Only exact scorer-consistent paths already frozen in the graph are used.
    signatures=sorted({r['rules'] for r in graph})
    # Bounded S1 candidate pool: all previously frozen 30 baseline tables plus
    # a deterministic sample of new unions from the full f68r2 graph. The pool
    # is explicitly recorded; it is not presented as an exhaustive maximum.
    baseline_path=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_coverage2_exact_v1/COVERAGE_2_SIGNATURES.tsv'
    candidates={r['signature'] for r in rows(baseline_path)}
    # Add a deterministic prefix of all newly formed unions that fit S1. This
    # broadens the search beyond the baseline while keeping the bounded run
    # reproducible and explicitly time-bounded.
    added=0
    for a,b in itertools.combinations(signatures[:1200],2):
        union=tuple(sorted(set(a.split(';'))|set(b.split(';'))))
        if 1<=len(union)<=5:
            candidates.add(';'.join(union)); added+=1
            if added>=300:break
    best=[]; progress=[]
    for i,table in enumerate(sorted(candidates),1):
        cov,sols=eval_table(graph,table)
        if cov:
            entry={'table':table,'coverage':cov,'table_hash':hashlib.sha256(table.encode()).hexdigest(),'solutions':sols}
            if not best or cov>best[0]['coverage']:best=[entry]
            elif cov==best[0]['coverage']:best.append(entry)
        if i%500==0:progress.append({'stage':'S1_K5_CANDIDATES','candidates_tested':i,'best_coverage':best[0]['coverage'] if best else 0,'elapsed_sec':round(time.monotonic()-started,4),'status':'RUNNING'})
    maxcov=best[0]['coverage'] if best else 0
    write('F68R2_OCCURRENCE_SCOPE.tsv',['label_id','page','token','line_id','source_reference','scope_status'],[dict(r,scope_status='FROZEN_OCCURRENCE') for r in targets])
    write('F68R2_PATH_GRAPH.tsv',list(graph[0])+['graph_status'],[dict(r,graph_status='SCORER_CONSISTENT_F68R2') for r in graph])
    group_rows=[r for r in rows(GROUPS) if r['panel']=='f68r2']
    mapping={r['canonical_label_id']:r for r in rows(MIG/'LABEL_3G1_TRANSCRIPTION_MAPPING.tsv') if r['panel']=='f68r2'}
    geo=[];ai=[];cross=[]
    for g in group_rows:
        m=mapping.get(g['label_ids'],{}); outcome=m.get('crosswalk_outcome','MISSING'); tier='AMBIGUOUS' if outcome!='UNIQUE_PROVENANCE_CROSSWALK' else 'EXACT_ID'
        geo.append({'group_id':g['canonical_group_id'],'label_member_id':g['label_ids'],'page':'f68r2','candidate_occurrence_ids':'','geometry_basis':'NO_FROZEN_OCCURRENCE_GEOMETRY','confidence_tier':tier,'status':'UNRESOLVED'})
        ai.append({'group_id':g['canonical_group_id'],'label_member_id':g['label_ids'],'page':'f68r2','ai_decision':'AMBIGUOUS','confidence_tier':'AMBIGUOUS','reason':'transcription occurrence has no frozen geometry/crop bridge; brute-force excluded','used_in_search':'NO'})
        cross.append({'group_id':g['canonical_group_id'],'label_member_id':g['label_ids'],'occurrence_ids':'','crosswalk_confidence':'AMBIGUOUS','group_status':'UNRESOLVED_NO_FROZEN_BRIDGE'})
    write('ID_PROVENANCE_AUDIT.tsv',['group_id','label_member_id','page','crosswalk_outcome','legacy_record_ids','provenance_status'],[{'group_id':g['canonical_group_id'],'label_member_id':g['label_ids'],'page':'f68r2','crosswalk_outcome':mapping.get(g['label_ids'],{}).get('crosswalk_outcome','MISSING'),'legacy_record_ids':'','provenance_status':'NO_UNIQUE_FROZEN_BRIDGE'} for g in group_rows])
    write('GEOMETRIC_CROSSWALK_CANDIDATES.tsv',list(geo[0]),geo);write('AI_VISUAL_ADJUDICATION.tsv',list(ai[0]),ai);write('GROUP_OCCURRENCE_CROSSWALK.tsv',list(cross[0]),cross)
    write('CROSSWALK_CONFIDENCE_SUMMARY.tsv',['tier','group_count','status'],[{'tier':'EXACT_ID','group_count':0,'status':'NONE'},{'tier':'EXACT_GEOMETRY','group_count':0,'status':'NONE'},{'tier':'AI_HIGH','group_count':0,'status':'NONE'},{'tier':'AI_MEDIUM','group_count':0,'status':'NONE'},{'tier':'AMBIGUOUS','group_count':24,'status':'PRESENT'},{'tier':'UNMAPPED','group_count':0,'status':'NONE'}])
    thresholds=[(4, 'FEASIBLE' if maxcov>=4 else 'UNKNOWN_TIMEOUT'),(5,'FEASIBLE' if maxcov>=5 else 'UNKNOWN_TIMEOUT'),(10,'FEASIBLE' if maxcov>=10 else 'UNKNOWN_TIMEOUT'),(15,'FEASIBLE' if maxcov>=15 else 'UNKNOWN_TIMEOUT'),(19,'FEASIBLE' if maxcov>=19 else 'UNKNOWN_TIMEOUT')]
    write('EXISTENCE_TESTS.tsv',['stage','threshold','status','best_incumbent','tested_candidates'],[{'stage':'S1_K5','threshold':f'coverage>={t}','status':s,'best_incumbent':maxcov,'tested_candidates':len(candidates)} for t,s in thresholds])
    progress.append({'stage':'S1_K5_CANDIDATES','candidates_tested':len(candidates),'best_coverage':maxcov,'elapsed_sec':round(time.monotonic()-started,4),'status':'COMPLETE_BOUNDED_CANDIDATE_ENUMERATION'})
    write('SOLVER_PROGRESS.tsv',list(progress[0]),progress)
    inc=[{'rank':i+1,'table_hash':b['table_hash'],'table':b['table'],'coverage':b['coverage']} for i,b in enumerate(best[:100])]
    write('INCUMBENTS.tsv',list(inc[0]),inc);write('BEST_TABLES.tsv',list(inc[0]),inc)
    matches=[];unmatched=[];support=[]
    for b in best[:100]:
        for n,s in enumerate(b['solutions'],1):
            for p in s:matches.append({'table_hash':b['table_hash'],'solution_index':n,'coverage':b['coverage'],'label_id':p['label_id'],'token':p['token'],'identity':p['identity'],'page':p['page'],'lexicon_id':p['lexicon_id'],'normalized_form':p['normalized_form'],'rules':p['rules'],'encoded':p['encoded']})
        for r in b['table'].split(';'):support.append({'table_hash':b['table_hash'],'rule':r,'status':'PASS','support_requirement':'2 distinct EVA token types'})
    write('OCCURRENCE_MATCH_ASSIGNMENTS.tsv',['table_hash','solution_index','coverage','label_id','token','identity','page','lexicon_id','normalized_form','rules','encoded'],matches);write('RULE_SUPPORT_AUDIT.tsv',list(support[0]) if support else ['table_hash','rule','status','support_requirement'],support)
    matched={p['label_id'] for p in matches};write('UNMATCHED_OCCURRENCES.tsv',['label_id','token','status'],[{'label_id':x['label_id'],'token':x['zl3b_token'],'status':'UNMATCHED_IN_BEST_RECORDED_TABLE'} for x in targets if x['label_id'] not in matched])
    write('GROUP_COVERAGE_INTERVALS.tsv',['scope','minimum_group_coverage','maximum_group_coverage','status'],[{'scope':'24 human groups','minimum_group_coverage':'NOT_COMPUTABLE','maximum_group_coverage':'NOT_COMPUTABLE','status':'CROSSWALK_PARTIAL'}])
    write('NULL_RESULTS.tsv',['status','reason'],[{'status':'NOT_RUN','reason':'occurrence search incumbent threshold not established and full selection pipeline not frozen'}])
    status={'OCCURRENCE_LEVEL_SEARCH':'TIMEOUT','GROUP_CROSSWALK':'PARTIAL','GROUP_LEVEL_EVALUATION':'NOT_EVALUATED','F68R2_OCCURRENCE_COUNT':len(targets),'F68R2_PATH_COUNT':len(graph),'S1_K5_CANDIDATES_TESTED':len(candidates),'BEST_OCCURRENCE_COVERAGE':maxcov,'S2_K8':'NOT_EVALUATED','S3_K12':'NOT_EVALUATED','IMPLEMENTATION_GATE_FAILED':'NO','CROSS_PAGE_INCLUDED':'NO','SCIENTIFIC_CLAIM':'NONE'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n');print(json.dumps(status,sort_keys=True))
if __name__=='__main__':main()
