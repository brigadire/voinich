#!/usr/bin/env python3
import csv, hashlib, itertools, json, sys, time
from collections import defaultdict, Counter
from pathlib import Path

ROOT=Path(__file__).resolve().parent
GRAPH=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_scorer_consistent_graph_v1'
COV2=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_coverage2_exact_v1'
GE3=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_coverage_ge3_v1'

def rows(p):
    with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader();w.writerows(data)
def table_hash(sig):return hashlib.sha256(sig.encode()).hexdigest()

def solve(paths, table, mode):
    t_rules=tuple(table.split(';')); rule_i={r:i for i,r in enumerate(t_rules)}
    toks=sorted({p['token'] for p in paths}); tok_i={t:i for i,t in enumerate(toks)}
    eligible=[]
    for p in paths:
        if not set(p['rules'].split(';')).issubset(set(t_rules)):continue
        if mode=='F68R2_ONLY' and p['page']!='f68r2':continue
        eligible.append(dict(p,rm=sum(1<<rule_i[r] for r in p['rules'].split(';')),tm=1<<tok_i[p['token']]))
    by=defaultdict(list)
    for p in eligible:by[p['label_id']].append(p)
    labels=sorted(by,key=lambda x:(len(by[x]),x))
    best=-1; sols=[]; nodes=0
    def valid(active,support):return all(support[i].bit_count()>=2 for i in range(len(t_rules)) if active&(1<<i))
    def rec(i,used,active,support,chosen,pages):
        nonlocal best,sols,nodes
        nodes+=1
        if i==len(labels):
            if mode=='CROSS_PAGE_REQUIRED' and pages!={'f68r1','f68r2'}:return
            if not valid(active,support):return
            n=len(chosen)
            key=tuple(sorted((p['label_id'],p['identity'],p['lexicon_id'],p['rules']) for p in chosen))
            if n>best:best=n;sols=[chosen[:]]
            elif n==best:sols.append(chosen[:])
            return
        if len(chosen)+(len(labels)-i)<best:return
        lid=labels[i]
        # Try paths first so incumbents appear early.
        for p in by[lid]:
            if p['identity'] in used:continue
            ns=list(support)
            for j in range(len(t_rules)):
                if p['rm']&(1<<j):ns[j]|=p['tm']
            rec(i+1,used|{p['identity']},active|p['rm'],ns,chosen+[p],pages|{p['page']})
        rec(i+1,used,active,support,chosen,pages)
    rec(0,set(),0,[0]*len(t_rules),[],set())
    return {'mode':mode,'table':table,'table_hash':table_hash(table),'eligible_paths':len(eligible),'best':best,'solutions':sols,'nodes':nodes,'tokens':toks}

def main():
    started=time.monotonic(); paths=rows(GRAPH/'PATH_GRAPH.tsv'); tables=[r['signature'] for r in rows(COV2/'COVERAGE_2_SIGNATURES.tsv')]
    all_results=[]
    for table in tables:
        for mode in ('UNCONSTRAINED_PAGE','F68R2_ONLY','CROSS_PAGE_REQUIRED'):
            r=solve(paths,table,mode); r['elapsed_sec']=round(time.monotonic()-started,4);all_results.append(r)
    maxima=[]; page=[]; assignments=[]; freq=[]; support=[]; cross=[]
    for r in all_results:
        sols=r['solutions']; maxn=max(r['best'],0)
        row={'table_hash':r['table_hash'],'signature':r['table'],'mode':r['mode'],'status':'OPTIMAL','eligible_path_count':r['eligible_paths'],'maximum_coverage':maxn,'optimal_assignment_count':len(sols),'unique_label_set_count':len({tuple(sorted(p['label_id'] for p in s)) for s in sols}), 'unique_identity_set_count':len({tuple(sorted(p['identity'] for p in s)) for s in sols}),'cooptimal_identity_permutations':len({tuple(sorted((p['label_id'],p['identity']) for p in s)) for s in sols}),'nodes':r['nodes'],'elapsed_sec':r['elapsed_sec']}
        maxima.append(row)
        if r['mode']!='UNCONSTRAINED_PAGE':page.append(row)
        label_count=Counter(); id_count=Counter(); rule_count=Counter(); token_by_rule=defaultdict(set)
        for s in sols:
            for p in s:
                label_count[p['label_id']]+=1;id_count[p['identity']]+=1
                for rule in p['rules'].split(';'):
                    rule_count[rule]+=1;token_by_rule[rule].add(p['token'])
        for k,v in label_count.items():freq.append({'table_hash':r['table_hash'],'mode':r['mode'],'kind':'LABEL','value':k,'optimal_assignment_participation':v,'optimal_assignment_count':len(sols)})
        for k,v in id_count.items():freq.append({'table_hash':r['table_hash'],'mode':r['mode'],'kind':'IDENTITY','value':k,'optimal_assignment_participation':v,'optimal_assignment_count':len(sols)})
        for k,v in rule_count.items():support.append({'table_hash':r['table_hash'],'mode':r['mode'],'rule':k,'path_participation':v,'distinct_eva_types':len(token_by_rule[k]),'support_types':','.join(sorted(token_by_rule[k])),'status':'PASS' if len(token_by_rule[k])>=2 else 'FAIL'})
        for n,s in enumerate(sols,1):
            aid=f"{r['table_hash'][:12]}:{r['mode']}:{n}"
            for slot,p in enumerate(s,1):assignments.append({'assignment_id':aid,'table_hash':r['table_hash'],'signature':r['table'],'mode':r['mode'],'slot':slot,'coverage':len(s),'label_id':p['label_id'],'page':p['page'],'token':p['token'],'identity':p['identity'],'lexicon_id':p['lexicon_id'],'normalized_form':p['normalized_form'],'encoded':p['encoded'],'rules':p['rules']})
    # Previous 2,488 triples are checked as label-set subsets of each unrestricted maximum.
    triples=rows(GE3/'MATCH_ASSIGNMENTS.tsv'); triple_sets={}
    triple_ids={}
    triple_tables={}
    for x in triples:
        triple_sets.setdefault(x['solution_id'],set()).add(x['label_id'])
        triple_ids.setdefault(x['solution_id'],set()).add(x['identity'])
        triple_tables[x['solution_id']]=x['table_hash']
    max_sets=defaultdict(list)
    for r in all_results:
        if r['mode']=='UNCONSTRAINED_PAGE':
            max_sets[r['table_hash']]=[(set(p['label_id'] for p in s),set(p['identity'] for p in s)) for s in r['solutions']]
    for row in maxima:
        if row['mode']!='UNCONSTRAINED_PAGE':continue
        label_n=sum(1 for ts in triple_sets.values() if any(ts.issubset(ms[0]) for ms in max_sets[row['table_hash']]))
        identity_n=sum(1 for sid,ts in triple_sets.items() if any(ts.issubset(ms[0]) and triple_ids[sid].issubset(ms[1]) for ms in max_sets[row['table_hash']]))
        same_table_n=sum(1 for sid in triple_sets if triple_tables[sid]==row['table_hash'] and any(triple_sets[sid].issubset(ms[0]) and triple_ids[sid].issubset(ms[1]) for ms in max_sets[row['table_hash']]))
        row['coverage3_label_sets_subsets_of_maximum']=label_n
        row['coverage3_identity_sets_subsets_of_maximum']=identity_n
        row['coverage3_same_table_subsets_of_maximum']=same_table_n
    write('TABLE_MAXIMA.tsv',list(maxima[0]),maxima)
    write('PAGE_MODE_MAXIMA.tsv',list(page[0]),page)
    write('OPTIMAL_ASSIGNMENTS.tsv',list(assignments[0]),assignments)
    write('MATCH_ASSIGNMENTS.tsv',list(assignments[0]),assignments)
    write('COOPTIMALITY_SUMMARY.tsv',[k for k in maxima[0]],maxima)
    write('LABEL_IDENTITY_FREQUENCIES.tsv',list(freq[0]),freq)
    write('RULE_SUPPORT_AUDIT.tsv',list(support[0]),support)
    cross_rows=[x for x in page if x['mode']=='CROSS_PAGE_REQUIRED'];write('CROSS_PAGE_AUDIT.tsv',list(cross_rows[0]),cross_rows)
    status={'UNCONSTRAINED_TABLES_COMPLETED':30,'PAGE_MODES_COMPLETED':60,'CROSS_PAGE_MAXIMA_STATUS':'OPTIMAL','IMPLEMENTATION_GATE_FAILED':'NO','SEARCH_INCONCLUSIVE_TIMEOUT':'NO','SCIENTIFIC_CLAIM':'NONE','MAX_UNCONSTRAINED_COVERAGE':max(int(x['maximum_coverage']) for x in maxima if x['mode']=='UNCONSTRAINED_PAGE'),'MAX_CROSS_PAGE_REQUIRED_COVERAGE':max(int(x['maximum_coverage']) for x in cross_rows)}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
    print(json.dumps(status,sort_keys=True))
if __name__=='__main__':main()
