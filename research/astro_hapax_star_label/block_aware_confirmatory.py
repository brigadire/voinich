#!/usr/bin/env python3
from pathlib import Path
import csv,random,statistics,json
O=Path(__file__).resolve().parent/'restricted_hapax_enrichment_v1'; C=list(csv.DictReader((O/'COHORT_MEMBERS.tsv').open(),delimiter='\t'))
def blocks(cohort):
 d={}
 for x in C:
  if x['cohort']==cohort:d.setdefault(x.get('line_ref',''),[]).append(x)
 return [v for v in d.values() if v]
def rate(b):return sum(int(x['hapax']) for q in b for x in q)/sum(len(q) for q in b)
def boot(a,b,N=100000,seed=20260917):
 q=random.Random(seed); vals=[]
 for _ in range(N):
  aa=[q.choice(a) for _ in a];bb=[q.choice(b) for _ in b];vals.append(rate(aa)-rate(bb))
 vals.sort();obs=rate(a)-rate(b);return obs,statistics.mean(vals),sum(x<=0 for x in vals)/N,vals[int(.025*N)],vals[int(.975*N)]
target=blocks('STAR_LABEL'); intro=blocks('INTRO_PROSE'); circ=blocks('CIRCULAR_TEXT')
rows=[]
for name,b in [('INTRO_PROSE',intro),('CIRCULAR_TEXT',circ)]:
 obs,mu,p,lo,hi=boot(target,b);rows.append({'comparison':name,'target_blocks':len(target),'control_blocks':len(b),'target_rate':rate(target),'control_rate':rate(b),'observed_difference':obs,'bootstrap_mean':mu,'p_one_sided':p,'ci95_low':lo,'ci95_high':hi,'replicates':100000})
with (O/'BLOCK_AWARE_RESULTS.tsv').open('w',newline='') as f:
 w=csv.DictWriter(f,fieldnames=list(rows[0]),delimiter='\t');w.writeheader();w.writerows(rows)
page=[]
for p in ('f68r1','f68r2'):
 a=[q for q in target if q[0]['source'].startswith(p)];page.append({'subset':p,'blocks':len(a),'tokens':sum(len(q) for q in a),'hapax':sum(int(x['hapax']) for q in a for x in q),'rate':rate(a)})
with (O/'BLOCK_AWARE_PAGE_RESULTS.tsv').open('w',newline='') as f:
 w=csv.DictWriter(f,fieldnames=list(page[0]),delimiter='\t');w.writeheader();w.writerows(page)
(O/'BLOCK_AWARE_REPORT.md').write_text('# Block-aware confirmatory analysis\n\nBootstrap resampling uses line/functional blocks, preserving token bundles within each line. STAR LABEL blocks are compared with introductory and circular blocks; no dictionary or M0/M1 score is used. Circular text is the sibling control for any later M2R. Because block counts are small and unequal, intervals are descriptive pilot inference.\n')
(O/'BLOCK_AWARE_MANIFEST.json').write_text(json.dumps({'target_blocks':len(target),'intro_blocks':len(intro),'circular_blocks':len(circ),'replicates':100000,'seed':20260917,'roles':'STAR_LABEL target; CIRCULAR_TEXT sibling; INTRO_PROSE external','M2R_decision':'PENDING_REVIEW'},indent=2)+'\n')
