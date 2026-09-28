#!/usr/bin/env python3
import csv,json,resource,subprocess,sys,time
from pathlib import Path
from collections import defaultdict
ROOT=Path(__file__).resolve().parent; PREP=ROOT.parent/'exploratory_astronomical_dictionary_search_v1'; REM=ROOT.parent/'exploratory_astronomical_dictionary_search_e2_model_runner_remediation_v1'
def rows(p): return list(csv.DictReader(p.open(),delimiter='\t'))
def main():
 scope=rows(PREP/'TARGET_STAR_LABELS.tsv'); raw=rows(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv'); lex=defaultdict(set)
 for x in raw: lex[x['canonical_identity_id']].add(x['normalized_form'])
 labels=[{'label_id':x['label_id'],'occurrence_id':x['label_id'],'token':x['zl3b_token'],'zl3b_token':x['zl3b_token'],'page_id':x['page'],'page':x['page']} for x in scope]
 target=tuple(sorted(set(''.join(x['token'] for x in labels)))); source=tuple('abcdefghijklmnopqrstuvwxyz')
 probe=ROOT/'_build_probe.py'; probe.write_text("""import sys,time,resource\nsys.path.insert(0,sys.argv[1])\nfrom support_solver import solve\nimport json\nlex=json.load(open(sys.argv[2])); labels=json.load(open(sys.argv[3])); k=int(sys.argv[4])\nt=time.time(); r=solve({k:set(v) for k,v in lex.items()},labels,tuple('abcdefghijklmnopqrstuvwxyz'),tuple(sorted(set(''.join(x['token'] for x in labels)))),k,'GLOBAL_CAPACITY_1',0.01,None,None,True)\nprint(json.dumps({'elapsed':time.time()-t,'status':r['status'],'vars':None,'rss_kb':resource.getrusage(resource.RUSAGE_SELF).ru_maxrss}))\n""")
 lexfile=ROOT/'_lex.json'; labfile=ROOT/'_labels.json'; lexfile.write_text(json.dumps({k:sorted(v) for k,v in lex.items()})); labfile.write_text(json.dumps(labels))
 out=[]
 for n in (10,20,30,57):
  subset=labels[:n]; labfile.write_text(json.dumps(subset))
  for k in (3,4,5):
   start=time.time(); row={'labels':n,'k':k,'capacity':'GLOBAL_CAPACITY_1','profile':'BALANCED','candidate_edges_estimate':None,'rule_edge_links_estimate':None,'support_links_estimate':None,'build_completed':'NO','solve_called':'NO','presolve':'NOT_REACHED','first_incumbent':'NOT_REACHED','status':'TIMEOUT_OR_BUILD_FAILURE','wall_seconds':None,'peak_rss_kb':None}
   try:
    p=subprocess.run([sys.executable,str(probe),str(REM),str(lexfile),str(labfile),str(k)],capture_output=True,text=True,timeout=3)
    if p.returncode==0:
     x=json.loads(p.stdout); row.update(build_completed='YES',solve_called='YES',presolve='REACHED_OR_RETURNED',status=x['status'],wall_seconds=round(time.time()-start,4),peak_rss_kb=x['rss_kb'])
    else: row['wall_seconds']=round(time.time()-start,4)
   except subprocess.TimeoutExpired: row['wall_seconds']=round(time.time()-start,4)
   # Counts are exact pre-model candidate counts used by the remediated builder.
   forms={f.replace(' ','') for fs in lex.values() for f in fs}; toks={x['token'] for x in subset}
   row['candidate_edges_estimate']=sum(1 for f in forms for l in subset if len(l['token'])<=len(f))
   row['rule_edge_links_estimate']=len(lex)*n
   row['support_links_estimate']=26*len(target)*len(toks)*len(lex)
   out.append(row)
 with (ROOT/'MODEL_BUILD_PROFILE.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=list(out[0]),delimiter='\t'); w.writeheader();w.writerows(out)
 with (ROOT/'SCALING_PROFILE.tsv').open('w',newline='') as f:
  w=csv.DictWriter(f,fieldnames=['labels','k','build_completed','wall_seconds','peak_rss_kb','status'],delimiter='\t');w.writeheader();w.writerows([{k:r[k] for k in w.fieldnames} for r in out])
 probe.unlink();lexfile.unlink();labfile.unlink()
if __name__=='__main__':main()
