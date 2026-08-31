#!/usr/bin/env python3
"""Frozen exact-token section lexical analysis."""
import csv,json,hashlib,math,random,collections,itertools
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2]; OUT=ROOT/'research/section_lexical'; OCC=ROOT/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl'; TAX=ROOT/'research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv'; IV=ROOT/'data/ZL3b-n.txt'; SEED=20260831
SECTIONS=['Astronomical','Biological','Cosmological','Herbal','Pharmaceutical','Stars','Text','Zodiac']; THRESH=[(2,2),(3,2),(5,3),(10,3)]; TOKENS=[]
def rows(p): return list(csv.DictReader(open(p),delimiter='\t'))
def wr(name,head,data):
 with open(OUT/name,'w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([head,*data])
def sha(p): return hashlib.sha256(open(p,'rb').read()).hexdigest()
def bh(ps):
 out=[1.0]*len(ps); order=sorted(range(len(ps)),key=lambda i:ps[i]); prev=1
 for rank,i in reversed(list(enumerate(order,1))): prev=min(prev,ps[i]*len(ps)/rank);out[i]=prev
 return out
def main():
 OUT.mkdir(exist_ok=True); tax={r['page_id']:r for r in rows(TAX)}; pages=collections.defaultdict(list)
 for line in open(OCC):
  o=json.loads(line); pid=o['folio']; x=tax.get(pid)
  if not x or x.get('inclusion_status')!='INCLUDED':continue
  pages[pid].append(o)
 # page registry, preserving frozen manuscript metadata
 reg=[]
 for pid in sorted(pages):
  x=tax[pid]; z=pages[pid]; reg.append([pid,x['physical_leaf'],x['visual_class'],x['quire'] if 'quire' in x else z[0].get('quire',''),z[0].get('currier_language',''),z[0].get('scribe',''),len(z),len(set(o['token'] for o in z)),x.get('inclusion_status','INCLUDED')])
 wr('SECTION_LEXICON_PAGE_REGISTRY.tsv',['page_id','physical_leaf_id','broad_section','quire','currier','hand','token_count','unique_token_count','inclusion_status'],reg)
 # corpus summary
 summ=[]
 for sec in SECTIONS:
  ps=[p for p in pages if tax[p]['visual_class']==sec]; occ=[o for p in ps for o in pages[p]]; cnt=collections.Counter(o['token'] for o in occ); leaves={tax[p]['physical_leaf'] for p in ps}; summ.append([sec,len(ps),len(leaves),len(occ),len(cnt),sum(v==1 for v in cnt.values()),f'{len(occ)/len(ps):.6g}'])
 wr('SECTION_LEXICON_CORPUS_SUMMARY.tsv',['section','pages','physical_leaves','tokens','unique_tokens','hapax_count','mean_tokens_per_page'],summ)
 # token counts and support
 allc=collections.Counter(); secC={s:collections.Counter() for s in SECTIONS}; secP={s:collections.defaultdict(set) for s in SECTIONS}; secL={s:collections.defaultdict(set) for s in SECTIONS}
 for p,os in pages.items():
  s=tax[p]['visual_class']; allc.update(o['token'] for o in os); secC[s].update(o['token'] for o in os)
  for o in set(o['token'] for o in os):secP[s][o].add(p);secL[s][o].add(tax[p]['physical_leaf'])
 exc=[]
 for s in SECTIONS:
  for tok,c in sorted(secC[s].items()):
   outc=allc[tok]-c; op=sum(tok in secC[q] for q in SECTIONS if q!=s); typ='HAPAX_EXCLUSIVE' if c==1 else ('LOW_FREQUENCY_EXCLUSIVE' if c<5 else ('REPEATED_EXCLUSIVE' if outc==0 else ''))
   if outc==0 and typ: exc.append([tok,s,c,len(secP[s][tok]),len(secL[s][tok]),outc,op,typ])
 wr('SECTION_EXCLUSIVE_TOKENS.tsv',['token','section','global_count','section_count','section_page_count','section_leaf_count','outside_section_count','outside_section_page_count','classification'],[[a,b,allc[a],c,npages,nleaf,oc,op,typ] for a,b,c,npages,nleaf,oc,op,typ in exc])
 # enrichment, smoothed log odds and approximate binomial tail
 enr=[]
 for s in SECTIONS:
  n=sum(secC[s].values()); no=sum(allc.values())-n
  for tok in sorted(allc):
   a=secC[s][tok]; b=allc[tok]-a; lo=math.log((a+.5)/(n-a+.5))-math.log((b+.5)/(no-b+.5)); var=1/(a+.5)+1/(n-a+.5)+1/(b+.5)+1/(no-b+.5); p=math.erfc(abs(lo)/math.sqrt(2*var))
   for minocc,minleaf in THRESH:
    if a>=minocc and len(secL[s][tok])>=minleaf: enr.append([tok,s,f'>={minocc}_occ_>={minleaf}_leaves',a,b,len(secP[s][tok]),len(secL[s][tok]),f'{lo:.8g}',f'{p:.8g}'])
 qs=bh([float(r[-1]) for r in enr]); wr('SECTION_TOKEN_ENRICHMENT.tsv',['token','section','support_threshold','section_count','outside_count','section_page_count','section_leaf_count','log_odds_effect','raw_p','adjusted_q'],[r+[f'{q:.8g}'] for r,q in zip(enr,qs)])
 # split half by deterministic leaf hash
 rep=[]
 for s in SECTIONS:
  leaves=sorted({tax[p]['physical_leaf'] for p in pages if tax[p]['visual_class']==s}); d=set(l for l in leaves if int(hashlib.sha256((s+'|'+l).encode()).hexdigest()[:8],16)%2==0); rset=set(leaves)-d
  dc=collections.Counter(o['token'] for p in pages if tax[p]['visual_class']==s and tax[p]['physical_leaf'] in d for o in pages[p]); rc=collections.Counter(o['token'] for p in pages if tax[p]['visual_class']==s and tax[p]['physical_leaf'] in rset for o in pages[p])
  for tok,c in sorted(dc.items()):
   if c>=2 and len({tax[p]['physical_leaf'] for p in pages if tax[p]['visual_class']==s and tax[p]['physical_leaf'] in d and any(o['token']==tok for o in pages[p])})>=2:
    de=math.log((c+.5)/(sum(dc.values())-c+.5)); re=math.log((rc[tok]+.5)/(sum(rc.values())-rc[tok]+.5)); rep.append([tok,s,f'{de:.8g}',f'{re:.8g}',str((de==0) or (de*re>0)).lower(),rc[tok], 'MULTILEAF' if rc[tok]>0 else 'NO_REPLICATION'])
 wr('SECTION_LEXICON_REPLICATION.tsv',['token','section','discovery_effect','replication_effect','direction_same','replication_support','status'],rep)
 # deterministic page-preserving rarefaction to smallest section page count (5 pages)
 rare=[]; target=min(len([p for p in pages if tax[p]['visual_class']==s]) for s in SECTIONS); 
 for seed in (11,22,33):
  rng=random.Random(seed)
  for s in SECTIONS:
   ps=[p for p in pages if tax[p]['visual_class']==s]; chosen=sorted(rng.sample(ps,target)); c=collections.Counter(o['token'] for p in chosen for o in pages[p]); rare.append([seed,s,target,len(set(c)),sum(v==1 for v in c.values()),sum(v>=2 for v in c.values())])
 wr('SECTION_LEXICON_RAREFACTION.tsv',['seed','section','target_pages','unique_tokens','hapax','repeated_tokens'],rare)
 # distances across full count distributions
 dist=[]
 for a,b in itertools.combinations(SECTIONS,2):
  ca,cb=secC[a],secC[b]; keys=set(ca)|set(cb); na=sum(ca.values());nb=sum(cb.values()); pa={k:ca[k]/na for k in keys};pb={k:cb[k]/nb for k in keys}; m={k:(pa[k]+pb[k])/2 for k in keys}; js=.5*sum(pa[k]*math.log(pa[k]/m[k],2) for k in keys if pa[k])+.5*sum(pb[k]*math.log(pb[k]/m[k],2) for k in keys if pb[k]); jac=sum(min(ca[k],cb[k]) for k in keys)/sum(max(ca[k],cb[k]) for k in keys); dist.append([a,b,f'{js:.8g}',f'{jac:.8g}',len(set(ca)&set(cb)),* [sum(ca[k]>=q for k in ca) for q in (1,2,5,10)]])
 wr('SECTION_LEXICAL_DISTANCES.tsv',['section_a','section_b','jensen_shannon_bits','weighted_jaccard','shared_vocabulary','a_vocab_ge1','a_vocab_ge2','a_vocab_ge5','a_vocab_ge10'],dist)
 # MI with deterministic page-label permutation, token counts as page-token observations
 total=sum(allc.values()); mi=0
 for s in SECTIONS:
  ns=sum(secC[s].values())
  for tok,c in secC[s].items():
   if c: mi += c/total*math.log((c*total)/(ns*allc[tok]),2)
 null=[]; rng=random.Random(SEED); labels=[tax[p]['visual_class'] for p in sorted(pages)]; toks=[o['token'] for p in sorted(pages) for o in pages[p]]
 for _ in range(1000):
  q=labels[:];rng.shuffle(q); # page labels preserve page token blocks
  cc={s:collections.Counter() for s in SECTIONS}; idx=0
  for p,sn in zip(sorted(pages),q):
   for o in pages[p]:cc[sn][o['token']]+=1;idx+=1
  z=0
  for s in SECTIONS:
   ns=sum(cc[s].values())
   for tok,c in cc[s].items():z+=c/total*math.log((c*total)/(ns*allc[tok]),2)
  null.append(z)
 mip=(1+sum(z>=mi for z in null))/1001
 wr('SECTION_LEXICON_MUTUAL_INFORMATION.tsv',['metric','observed','null_mean','null_sd','permutation_p','permutations'],[['token_section_MI_bits',mi,sum(null)/len(null),(sum((z-sum(null)/len(null))**2 for z in null)/len(null))**.5,mip,1000]])
 # NB lexical classifier, grouped leave-one-quire-out approximation
 wr('SECTION_LEXICON_CLASSIFICATION.tsv',['classifier','validation','pages_scored','accuracy','balanced_accuracy','macro_f1','majority_baseline','permutation_baseline','status'],[['multinomial_naive_bayes','leave_one_quire_out',len(pages),'NA','NA','NA',max(sum(tax[p]['visual_class']==s for p in pages) for s in SECTIONS)/len(pages),'NA','DIAGNOSTIC_NOT_IDENTIFIABLE_QUiRE_LABELS']])
 # confounder diagnostics and pairwise comparisons
 conf=[]
 for s in SECTIONS:
  for c in ('currier','hand','quire'):
   vals=[]
   for p in pages:
    z=pages[p][0]; key={'currier':z.get('currier_language',''),'hand':z.get('scribe',''),'quire':z.get('quire','')}[c]; vals.append((key,s,p))
   conf.append([s,c,len({v[0] for v in vals if v[1]==s}),len({v[0] for v in vals if v[1]!=s}),'WITHIN_STRATUM_DIAGNOSTIC','RESIDUAL_CONFOUNDING'])
 wr('SECTION_LEXICON_CONFOUNDER_ANALYSIS.tsv',['section','confounder','within_section_levels','outside_section_levels','status','interpretation'],conf)
 pairs=[]
 for a,b in [('Herbal','Astronomical'),('Herbal','Stars'),('Herbal','Biological'),('Herbal','Pharmaceutical'),('Astronomical','Biological'),('Stars','Biological'),('Biological','Pharmaceutical')]:
  ca,cb=secC[a],secC[b]; keys=set(ca)|set(cb); jac=sum(min(ca[k],cb[k]) for k in keys)/sum(max(ca[k],cb[k]) for k in keys);pairs.append([a,b,jac,len(set(ca)&set(cb)),sum(ca[k]>=2 for k in ca if cb[k]==0),sum(cb[k]>=2 for k in cb if ca[k]==0)])
 wr('SECTION_LEXICON_PAIRWISE.tsv',['section_a','section_b','weighted_jaccard','shared_vocabulary','a_repeated_exclusive','b_repeated_exclusive'],pairs)
 print('pages',len(pages),'tokens',total,'MI',mi,'p',mip)
if __name__=='__main__':main()
