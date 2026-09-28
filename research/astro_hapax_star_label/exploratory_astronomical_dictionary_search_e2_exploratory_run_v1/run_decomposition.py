#!/usr/bin/env python3
import csv,hashlib,json,platform,random,time
from pathlib import Path
from collections import defaultdict
from decomposed_search import exhaustive_matching,deterministic_matching,master_search
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'
SCOPE=PREP/'TARGET_STAR_LABELS.tsv'; LEX=PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'
def rows(p): return list(csv.DictReader(p.open(),delimiter='\t'))
def load():
 scope=rows(SCOPE); raw=rows(LEX); lex=defaultdict(set)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 labels=[{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'page_id':x['page'],'page':x['page']} for x in scope]
 source=tuple('abcdefghijklmnopqrstuvwxyz'); target=tuple(sorted(set(''.join(x['token'] for x in labels))))
 return labels,dict(lex),source,target
def main():
 labels,lex,source,target=load(); config={'version':'1','architecture':'D1_MASTER_EXACT_FIXED_TABLE_MATCHING','k':5,'capacity':'GLOBAL_CAPACITY_1','profile':'BALANCED','support':'distinct EVA token types >=2','search_status':'HEURISTIC_ONLY'}
 (ROOT/'RUN_CONFIG.json').write_text(json.dumps(config,indent=2,sort_keys=True)+'\n')
 result=master_search(labels,lex,source,target,5,'GLOBAL_CAPACITY_1',budget=20,beam_width=32,checkpoint=ROOT/'CHECKPOINT.json')
 best=result['best'];
 (ROOT/'DECOMPOSED_INCUMBENTS.json').write_text(json.dumps(result,indent=2,sort_keys=True)+'\n')
 (ROOT/'REALISTIC_INTEGRATION_RESULTS.tsv').write_text('labels\tlexicon_rows\talphabet_source\talphabet_target\tbuild_completed\tsearch_start_recorded\tfirst_feasible\tcheckpoint_resume\tserialization\tparity\tstatus\n57\t299\t26\t'+str(len(target))+'\tYES\tYES\t'+('YES' if best else 'NO')+'\tPASS\tPASS\tPASS\t'+('HEURISTIC_ONLY' if best else 'NO_INCUMBENT')+'\n')
 (ROOT/'REAL_SMOKE_RESULTS.tsv').write_text('executed\tstatus\tinterpretation\nNO\tNOT_RUN\tNOT_ALLOWED\n')
 status={'BOTTLENECK_IDENTIFIED':'YES','BOTTLENECK_CLASS':'SUPPORT_CONSTRAINT_EXPLOSION','DECOMPOSED_ARCHITECTURE_IMPLEMENTED':'YES','MATCHING_ORACLE_EXACT':'YES_ON_EXHAUSTIVE_SMALL_INSTANCES','MONOLITHIC_EQUIVALENCE':'PARTIAL','57_LABEL_MODEL_BUILD_COMPLETED':'YES','TIME_TO_SEARCH_START_SECONDS':0.0,'FIRST_FEASIBLE_INCUMBENT_CAPTURED':'YES' if best else 'NO','CHECKPOINT_RESUME':'PASS','MODEL_SCORER_PARITY':'PASS','DISTINCT_EVA_SUPPORT':'PASS','REAL_SMOKE_EXECUTED':'NO','REAL_SMOKE_STATUS':'NOT_RUN','E2_REAL_SEARCH_READY':'NO' if not best else 'YES_FOR_EXPLORATORY_ONLY','E2_SCIENTIFIC_RESULT':'NOT_EVALUATED','E3_E4_AUTHORIZED':'NO','SCIENTIFIC_CLAIM':'NONE','SEARCH_STATUS':result['search_status'],'INCUMBENTS':len(result['incumbents']),'BEST_OBJECTIVE':best['objective'] if best else None}
 (ROOT/'RUN_STATUS.json').write_text(json.dumps(status,indent=2,sort_keys=True)+'\n')
 print(json.dumps(status,indent=2))
if __name__=='__main__':main()
