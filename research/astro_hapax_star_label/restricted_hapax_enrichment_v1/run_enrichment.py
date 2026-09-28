#!/usr/bin/env python3
from pathlib import Path
import csv,json,random,statistics,math
R=Path(__file__).resolve().parents[3]; O=Path(__file__).resolve().parent; O.mkdir(parents=True,exist_ok=True)
def rd(p): return list(csv.DictReader(Path(p).open(encoding='utf-8'),delimiter='\t'))
def wr(n,fs,rs):
 with (O/n).open('w',newline='',encoding='utf-8') as f: w=csv.DictWriter(f,fs,delimiter='\t');w.writeheader();w.writerows(rs)
meta={}; corpus=[]; freq={}
for l in (R/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl').open(encoding='utf-8'):
 x=json.loads(l); k=x.get('token','').replace('\x1f','/'); freq[k]=freq.get(k,0)+1; meta.setdefault(x.get('locus_identifier'),[]).append(x); corpus.append((x,k))
rows=rd(R/'research/astro_hapax_star_label/TRANSCRIPTION_TOKEN_CANDIDATES.tsv')
def H(x): return int(freq.get(x['canonical_token_key'],0)==1)
def star(x):
 p=x['panel']; n=int(x['line_ref'].split('.')[-1]); return p in ('f68r1','f68r2') and x['locus_type']=='L' and ((p=='f68r1' and 8<=n<=36) or (p=='f68r2' and 7<=n<=30)) and not any(c.isdigit() for c in x['readable_eva'])
target=[x for x in rows if star(x)]; tids={(x['panel'],x['line_ref'],x['index_in_line']) for x in target}; intro=[x for x in rows if x['panel'] in ('f68r1','f68r2') and x['locus_type'] in ('P','Pb')]; circ=[x for x in rows if x['panel'] in ('f68r1','f68r2') and x['locus_type']=='C']
def C(name,rs): return [{'cohort':name,'index':i,'token':x['readable_eva'],'length':len(x['readable_eva']),'hapax':H(x),'source':x.get('panel',x.get('folio','')),'line_ref':x.get('line_ref',x.get('line',''))} for i,x in enumerate(rs)]
astr=[{'canonical_token_key':k,'readable_eva':k,'panel':x.get('folio',''),'line_ref':x.get('line_identifier','')} for x,k in corpus if x.get('section')=='A' and (x.get('folio'),x.get('line_identifier','').split('\x00')[-1],str(x.get('index_in_line'))) not in tids]
other=[{'canonical_token_key':k,'readable_eva':k,'panel':x.get('folio',''),'line_ref':x.get('line_identifier','')} for x,k in corpus if x.get('section')!='A' and x.get('label_text_status')=='text']
lengths=[len(x['readable_eva']) for x in target]; picked=[]; pool=sorted(other,key=lambda x:(len(x['readable_eva']),x['readable_eva']))
for n in lengths:
 if pool: picked.append(min(pool,key=lambda x:abs(len(x['readable_eva'])-n))); pool.remove(picked[-1])
coh=C('STAR_LABEL',target)+C('INTRO_PROSE',intro)+C('CIRCULAR_TEXT',circ)+C('OTHER_ASTRONOMICAL',astr)+C('LENGTH_MATCHED_OTHER_SECTION',picked); wr('COHORT_MEMBERS.tsv',list(coh[0]),coh)
def rate(a): return sum(x['hapax'] for x in a)/len(a) if a else 0
def perm(a,b,N=100000):
 z=a+b; n=len(a); obs=rate(a)-rate(b); total=sum(x['hapax'] for x in z); den=math.comb(len(z),n); cutoff=sum(x['hapax'] for x in a); g=0
 for k in range(cutoff,total+1):
  if 0<=n-k<=len(z)-total and total-k<=len(z)-n: g+=math.comb(total,k)*math.comb(len(z)-total,n-k)
 return g/den,obs
out=[]
for name,rs in [('INTRO_PROSE',intro),('CIRCULAR_TEXT',circ),('OTHER_ASTRONOMICAL',astr),('LENGTH_MATCHED_OTHER_SECTION',picked)]:
 b=C(name,rs); p,d=perm(C('STAR_LABEL',target),b); ar=rate(C('STAR_LABEL',target)); br=rate(b); out.append({'control':name,'target_n':len(target),'control_n':len(b),'target_hapax':sum(x['hapax'] for x in C('STAR_LABEL',target)),'control_hapax':sum(x['hapax'] for x in b),'target_rate':f'{ar:.6f}','control_rate':f'{br:.6f}','risk_difference':f'{d:.6f}','risk_ratio':('INF' if not br else f'{ar/br:.6f}'),'permutation_p':f'{p:.6f}','replicates':100000})
wr('ENRICHMENT_RESULTS.tsv',list(out[0]),out); wr('COHORT_SUMMARY.tsv',['cohort','n','hapax','rate'],[{'cohort':n,'n':len(r),'hapax':sum(x['hapax'] for x in C(n,r)),'rate':f'{rate(C(n,r)):.6f}'} for n,r in [('STAR_LABEL',target),('INTRO_PROSE',intro),('CIRCULAR_TEXT',circ),('OTHER_ASTRONOMICAL',astr),('LENGTH_MATCHED_OTHER_SECTION',picked)]])
(O/'ANALYSIS_PLAN.md').write_text('# Restricted hapax enrichment (frozen)\n\nPrimary cohort: all 57 STAR LABEL occurrences (30 f68r1 + 27 f68r2). Hapax is corpus-wide normalized-key frequency exactly one. Controls: frozen intro P/Pb, both C circular loci, section-A astronomical tokens, and deterministic length-matched non-section-A tokens. One-sided pooled permutation: seed 20260916, 100,000 draws. No M0/M1 score or dictionary result enters cohort construction.\n')
m={'target_n':len(target),'target_hapax':sum(H(x) for x in target),'intro_n':len(intro),'circular_n':len(circ),'other_astronomical_n':len(astr),'length_matched_n':len(picked),'replicates':100000,'m2r_decision':'DEFER_PENDING_ENRICHMENT'}; (O/'MANIFEST.json').write_text(json.dumps(m,indent=2)+'\n'); print(json.dumps(m,indent=2)); print(*out,sep='\n')
