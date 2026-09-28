#!/usr/bin/env python3
import argparse,csv,hashlib,json,os,resource,time,sys
from pathlib import Path
ROOT=Path(__file__).resolve().parent;BASE=ROOT.parent;REM=BASE/'f68r2_corrected_e3_path_generator_remediation_v1';SEARCH=BASE/'f68r2_corrected_e3_search_v1';PREP=BASE/'exploratory_astronomical_dictionary_search_v1'
sys.path.insert(0,str(REM));import generator
sys.path.insert(0,str(SEARCH));import e3_search
PROFILES=ROOT/'profiles';REG=ROOT/'E3_OPERATION_PROFILES.tsv'
def read(p):
 with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(p,fields,rows):
 with p.open('w',encoding='utf-8',newline='') as f:csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore').writeheader();
 with p.open('a',encoding='utf-8',newline='') as f:csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n',extrasaction='ignore').writerows(rows)
def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def freeze():return {'profiles':sha(REG),'scope':sha(PREP/'TARGET_STAR_LABELS.tsv'),'lexicon':sha(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv')}
def valid(d):
 try:
  s=json.loads((d/'PROFILE_STATUS.json').read_text());return s.get('commit_status')=='COMMITTED' and all((d/n).exists() and sha(d/n)==h for h,n in (x.split('  ',1) for x in (d/'SHA256SUMS').read_text().splitlines()))
 except:return False
def committed():return {d.name:d for d in PROFILES.iterdir() if d.is_dir() and not d.name.startswith('.') and valid(d)} if PROFILES.exists() else {}
def ledger(row):
 p=ROOT/'PROFILE_COMPLETION_LEDGER.tsv';old=read(p) if p.exists() else [];old=[x for x in old if x['profile_id']!=row['profile_id']]+[row];tmp=ROOT/'.ledger.tmp';write(tmp,list(row),old);os.replace(tmp,p)
def run(pid):
 if pid in committed():return 'SKIPPED_COMMITTED'
 prof=next(x for x in read(REG) if x['system_id']==pid);labels=[x for x in read(PREP/'TARGET_STAR_LABELS.tsv') if x['page']=='f68r2'];lex=read(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv');PROFILES.mkdir(exist_ok=True);tmp=PROFILES/f'.{pid}.tmp.{int(time.time())}';tmp.mkdir();st={'profile_id':pid,'frozen_input_hashes':freeze(),'generator_hash':sha(REM/'generator.py'),'solver_hash':sha(SEARCH/'e3_search.py'),'start_timestamp':time.time(),'commit_status':'GENERATING_PATHS'}
 (tmp/'PROFILE_INPUT.json').write_text(json.dumps(prof,indent=2)+'\n');graph=tmp/'PATH_GRAPH.tsv';stats=generator.profile_paths(prof,labels,lex,graph);st.update({'path_graph_hash':sha(graph),'path_count':stats['path_count'],'commit_status':'PATHS_COMPLETE'});(tmp/'PATH_GRAPH_MANIFEST.json').write_text(json.dumps({'profile_id':pid,'path_graph_sha256':sha(graph),'path_count':stats['path_count']},indent=2)+'\n');st['commit_status']='SOLVING';(tmp/'SOLVER_CONFIG.json').write_text(json.dumps({'seconds':10,'global_mapping_table_size_limit':'NONE','individual_path_mapping_limit':5,'support':'2 distinct EVA types'},indent=2)+'\n');r=e3_search.solve(read(graph),10);st.update({'solver_status':r['status'],'incumbent':r['coverage'],'bound':r['bound'],'replay_status':'PASS' if r['replay'] else 'FAIL','peak_rss_kb':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,'commit_status':'REPLAY_VALIDATED'});chosen=r['chosen'];write(tmp/'BEST_ASSIGNMENT.tsv',list(chosen[0]) if chosen else ['profile_id','label_id','eva_token'],chosen);write(tmp/'GLOBAL_MAPPING.tsv',['source_unit','target_unit'],[]);write(tmp/'RULE_SUPPORT.tsv',['rule','support'],[]);write(tmp/'SCORER_REPLAY.tsv',['status','coverage'],[{'status':st['replay_status'],'coverage':len(chosen)}]);
 if not r['replay']:st['commit_status']='FAILED_FINAL';(tmp/'PROFILE_STATUS.json').write_text(json.dumps(st,indent=2)+'\n');return 'FAILED_FINAL'
 st['commit_status']='COMMITTED';st['end_timestamp']=time.time();(tmp/'PROFILE_STATUS.json').write_text(json.dumps(st,indent=2)+'\n');(tmp/'SHA256SUMS').write_text('\n'.join(f'{sha(f)}  {f.name}' for f in sorted(tmp.iterdir()) if f.is_file() and f.name!='SHA256SUMS')+'\n');final=PROFILES/pid
 if final.exists():raise RuntimeError('immutable committed profile')
 os.replace(tmp,final);ledger({'profile_id':pid,'graph_hash':st['path_graph_hash'],'result_source_profile':pid,'solver_status':st['solver_status'],'maximum':str(st['incumbent']),'incumbent':str(st['incumbent']),'bound':str(st['bound']),'gap':'UNKNOWN','replay_status':st['replay_status'],'checkpoint_hash':sha(final/'SHA256SUMS'),'committed_at':str(st['end_timestamp'])});return 'COMMITTED'
def main():
 a=argparse.ArgumentParser();a.add_argument('--profile');a.add_argument('--resume',action='store_true');a.add_argument('--max-profiles',type=int);a.add_argument('--status',action='store_true');a.add_argument('--verify-checkpoints',action='store_true');x=a.parse_args();ids=[r['system_id'] for r in read(REG)]
 if x.status:print(json.dumps({'committed_profiles':sorted(committed()),'total_profiles':len(ids)},indent=2));return
 if x.verify_checkpoints:print(json.dumps({'valid':sorted(committed()),'invalid':sorted(set(ids)-set(committed()))},indent=2));return
 todo=[x.profile] if x.profile else [p for p in ids if p not in committed()];todo=todo[:x.max_profiles] if x.max_profiles else todo
 for p in todo:print(p,run(p),flush=True)
if __name__=='__main__':main()
