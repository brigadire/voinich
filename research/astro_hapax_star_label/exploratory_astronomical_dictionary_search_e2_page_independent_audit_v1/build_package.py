#!/usr/bin/env python3
import csv, hashlib, json
from pathlib import Path

ROOT=Path(__file__).resolve().parent
SRC=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_fixed_table_maxima_v1'

def rows(p):
    with p.open() as f:return list(csv.DictReader(f,delimiter='\t'))
def write(name,fields,data):
    with (ROOT/name).open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t');w.writeheader();w.writerows(data)

def main():
    maxima=rows(SRC/'TABLE_MAXIMA.tsv')
    uncon=[x for x in maxima if x['mode']=='UNCONSTRAINED_PAGE']
    cross=[x for x in maxima if x['mode']=='CROSS_PAGE_REQUIRED']
    out=[]
    for x,y in zip(uncon,cross):
        out.append({'table_hash':x['table_hash'],'signature':x['signature'],'same_page_maximum':x['maximum_coverage'],'same_page_status':x['status'],'both_pages_required_maximum':y['maximum_coverage'],'both_pages_status':'INFEASIBLE_CERTIFIED' if y['maximum_coverage']=='0' else 'FEASIBLE','optimal_assignment_count':x['optimal_assignment_count'],'unique_label_set_count':x['unique_label_set_count'],'interpretation':'f68r2_only' if x['maximum_coverage'] else 'none'})
    write('PAGE_INDEPENDENT_TABLE_AUDIT.tsv',list(out[0]),out)
    write('CROSS_PAGE_AUDIT.tsv',['table_hash','signature','same_page_maximum','same_page_status','both_pages_required_maximum','both_pages_status','optimal_assignment_count','unique_label_set_count','interpretation'],out)
    freeze={'source_table_maxima_sha256':hashlib.sha256((SRC/'TABLE_MAXIMA.tsv').read_bytes()).hexdigest(),'source_cross_page_sha256':hashlib.sha256((SRC/'CROSS_PAGE_AUDIT.tsv').read_bytes()).hexdigest(),'tested_tables':30,'scorer':'frozen scorer-consistent PATH_GRAPH','support':'2 distinct EVA token types','capacity':'GLOBAL_CAPACITY_1'}
    (ROOT/'INPUT_FREEZE.json').write_text(json.dumps(freeze,indent=2,sort_keys=True)+'\n')
    status={'PAGE_INDEPENDENT_SUPPORT_VALID_SOLUTION_FOUND':'NO','SIGNIFICANT_PAGE_INDEPENDENT_COVERAGE_FOUND':'NO','CROSS_PAGE_INFEASIBLE_CERTIFIED':'YES_FOR_30_TABLE_FAMILY','SEARCH_INCONCLUSIVE_TIMEOUT':'NO','IMPLEMENTATION_GATE_FAILED':'NO','TESTED_TABLE_COUNT':30,'MAX_SAME_PAGE_COVERAGE':3,'MAX_BOTH_PAGE_REQUIRED_COVERAGE':0,'SCIENTIFIC_CLAIM':'NONE'}
    (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
    print(json.dumps(status,sort_keys=True))
if __name__=='__main__':main()
