#!/usr/bin/env python3
import csv, hashlib, itertools, json, time
from collections import defaultdict
from pathlib import Path

ROOT=Path(__file__).resolve().parent
GRAPH=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_scorer_consistent_graph_v1'
COV2=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_coverage2_exact_v1'

def rows(path):
    with path.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader();w.writerows(data)

def main():
    started=time.monotonic(); paths=rows(GRAPH/'PATH_GRAPH.tsv'); sigs=rows(COV2/'COVERAGE_2_SIGNATURES.tsv')
    found=[]; tests=[]; progress=[]; match=[]; support=[]
    for idx,s in enumerate(sigs,1):
        table=tuple(s['signature'].split(';')); tset=set(table)
        eligible=[p for p in paths if set(p['rules'].split(';')).issubset(tset)]
        triples=[]
        for comb in itertools.combinations(eligible,3):
            if len({p['label_id'] for p in comb})<3 or len({p['identity'] for p in comb})<3 or len({p['token'] for p in comb})<3:continue
            ok=True
            for rule in table:
                if len({p['token'] for p in comb if rule in p['rules'].split(';')})<2:ok=False;break
            if ok:triples.append(comb)
        cross=[x for x in triples if len({p['page'] for p in x})==2]
        status='FEASIBLE_CERTIFIED' if triples else 'INFEASIBLE_CERTIFIED'
        cross_status='FEASIBLE_CERTIFIED' if cross else 'INFEASIBLE_CERTIFIED'
        tests.append({'signature':s['signature'],'rule_count':len(table),'eligible_paths':len(eligible),'coverage_ge3_status':status,'cross_page_status':cross_status,'triple_count':len(triples),'cross_page_triple_count':len(cross),'elapsed_sec':round(time.monotonic()-started,4)})
        progress.append({'signature':s['signature'],'index':idx,'eligible_paths':len(eligible),'triples':len(triples),'cross_page_triples':len(cross),'status':'COMPLETE'})
        for n,tr in enumerate(triples,1):
            sid=hashlib.sha256((s['signature']+f'#{n}').encode()).hexdigest()[:16]
            found.append({'solution_id':sid,'signature':s['signature'],'table_hash':hashlib.sha256(s['signature'].encode()).hexdigest(),'coverage':3,'cross_page':'YES' if len({p['page'] for p in tr})==2 else 'NO','slot':'1','label_id':tr[0]['label_id'],'page':tr[0]['page'],'token':tr[0]['token'],'identity':tr[0]['identity'],'lexicon_id':tr[0]['lexicon_id'],'normalized_form':tr[0]['normalized_form']})
            for slot,p in enumerate(tr[1:],2):found.append({'solution_id':sid,'signature':s['signature'],'table_hash':hashlib.sha256(s['signature'].encode()).hexdigest(),'coverage':3,'cross_page':'YES' if len({q['page'] for q in tr})==2 else 'NO','slot':str(slot),'label_id':p['label_id'],'page':p['page'],'token':p['token'],'identity':p['identity'],'lexicon_id':p['lexicon_id'],'normalized_form':p['normalized_form']})
    total=sum(x['triple_count'] for x in tests); cross_total=sum(x['cross_page_triple_count'] for x in tests)
    for s in sigs:
        tab=s['signature']; types=sorted({p['token'] for p in paths if set(p['rules'].split(';')).issubset(set(tab.split(';')))})
        for r in tab.split(';'):support.append({'signature':tab,'rule':r,'potential_distinct_eva_types':len(types),'support_gate':'PASS' if len(types)>=2 else 'FAIL'})
    write('EXISTENCE_TESTS.tsv',['signature','rule_count','eligible_paths','coverage_ge3_status','cross_page_status','triple_count','cross_page_triple_count','elapsed_sec'],tests)
    write('SOLVER_PROGRESS.tsv',['signature','index','eligible_paths','triples','cross_page_triples','status'],progress)
    write('INCUMBENTS.tsv',['scope','coverage','cross_page_triples','signature_count','status'],[{'scope':'30_coverage2_signatures','coverage':3 if total else 2,'cross_page_triples':cross_total,'signature_count':sum(1 for x in tests if x['triple_count']),'status':'CERTIFIED' if total else 'NO_GE3_FOUND'}])
    fields=['solution_id','signature','table_hash','coverage','cross_page','slot','label_id','page','token','identity','lexicon_id','normalized_form']
    write('MATCH_ASSIGNMENTS.tsv',fields,found)
    write('RULE_SUPPORT_AUDIT.tsv',['signature','rule','potential_distinct_eva_types','support_gate'],support)
    status={'COVERAGE_GE3_FEASIBLE':'YES' if total else 'NO','COVERAGE_GE3_INFEASIBLE_CERTIFIED':'NO' if total else 'YES','CROSS_PAGE_FEASIBLE':'YES' if cross_total else 'NO','CROSS_PAGE_INFEASIBLE_CERTIFIED':'NO' if cross_total else 'YES','SEARCH_INCONCLUSIVE_TIMEOUT':'NO','IMPLEMENTATION_GATE_FAILED':'NO','TESTED_TABLE_COUNT':len(sigs),'COVERAGE_GE3_TRIPLE_COUNT':total,'CROSS_PAGE_TRIPLE_COUNT':cross_total,'SCIENTIFIC_CLAIM':'NONE'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
    print(json.dumps(status,sort_keys=True))
if __name__=='__main__':main()
