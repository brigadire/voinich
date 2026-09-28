#!/usr/bin/env python3
import csv,json,hashlib,collections,math
from pathlib import Path
ROOT=Path(__file__).resolve().parents[2];OUT=ROOT/'research/hapax_check';OCC=ROOT/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl';TAX=ROOT/'research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv';PANEL=ROOT/'research/visual_descriptors/VISUAL_PANEL_CROP_REGISTRY.tsv';SECTIONS=['Astronomical','Biological','Cosmological','Herbal','Pharmaceutical','Stars','Text','Zodiac']
def rd(p):return list(csv.DictReader(open(p),delimiter='\t'))
def wr(n,h,r):
 with open(OUT/n,'w') as f:csv.writer(f,delimiter='\t',lineterminator='\n').writerows([h,*r])
def sha(p):return hashlib.sha256(open(p,'rb').read()).hexdigest()
def main():
 OUT.mkdir(exist_ok=True);tax={r['page_id']:r for r in rd(TAX)};pages=collections.defaultdict(list)
 for line in open(OCC):
  o=json.loads(line);p=o['folio']
  if p in tax and tax[p].get('inclusion_status')=='INCLUDED':pages[p].append(o)
 globalc=collections.Counter(o['token'] for z in pages.values() for o in z); globalhap={t for t,c in globalc.items() if c==1}
 localc={s:collections.Counter(o['token'] for p,z in pages.items() if tax[p]['visual_class']==s for o in z) for s in SECTIONS}; localhap={s:{t for t,c in localc[s].items() if c==1} for s in SECTIONS}
 out=[]
 for p in sorted(pages):
  s=tax[p]['visual_class'];z=pages[p]; toks=[o['token'] for o in z]; g=sum(t in globalhap for t in toks);l=sum(t in localhap[s] for t in toks);out.append([p,tax[p]['physical_leaf'],s,len(toks),g,len(toks)-g,f'{g/len(toks):.8g}' if toks else 'NA',str(g>len(toks)-g).lower(),l,len(toks)-l,f'{l/len(toks):.8g}' if toks else 'NA',str(l>len(toks)-l).lower()])
 wr('PAGE_HAPAX_PREVALENCE.tsv',['page_id','physical_leaf_id','broad_section','total_tokens','global_hapax_occurrences','global_non_hapax_occurrences','global_hapax_fraction','global_hapax_gt_nonhapax','section_local_hapax_occurrences','section_local_non_hapax_occurrences','section_local_hapax_fraction','section_local_hapax_gt_nonhapax'],out)
 summ=[]
 for s in SECTIONS:
  z=[r for r in out if r[2]==s];summ.append([s,len(z),sum(int(r[3]) for r in z),sum(int(r[4]) for r in z),sum(int(r[6]=='true') for r in z),sum(int(r[8]) for r in z),sum(int(r[11]=='true') for r in z),len({r[1] for r in z})])
 wr('SECTION_HAPAX_PREVALENCE.tsv',['broad_section','pages_panels','total_tokens','global_hapax_occurrences','pages_global_hapax_gt_nonhapax','section_local_hapax_occurrences','pages_local_hapax_gt_nonhapax','physical_leaves'],summ)
 astro=['f67r1','f67r2','f67v1','f68r1','f68r2','f68r3','f68v1','f68v2']; zones=['INSIDE_DIAGRAM','RING_OR_RADIAL_TEXT','MARGINAL_COMMENT','OUTER_TEXT_BLOCK','AMBIGUOUS']
 wr('ASTRO_HAPAX_SPATIAL.tsv',['zone','total_tokens','hapax_tokens','non_hapax_tokens','hapax_fraction','unique_token_types','physical_leaves_or_panels','status','source_artifact'],[[q,'DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','research/visual_context/VISUAL_CONTEXT_EXPERIMENT_REVIEW.md'] for q in zones])
 wr('ASTRO_HAPAX_SPATIAL_BY_PAGE.tsv',['page_id','physical_leaf_id','zone','total_tokens','hapax_tokens','non_hapax_tokens','status','reason'],[[p,tax[p]['physical_leaf'],'AMBIGUOUS','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','No token/line-to-spatial-zone coordinates in frozen artifacts'] for p in astro])
 wr('ASTRO_HAPAX_SPATIAL_TEST.tsv',['comparison','statistic','p_value','permutation_unit','status','reason'],[['image-related vs comment/text zones','DATA_NOT_AVAILABLE','DATA_NOT_AVAILABLE','physical-leaf/panel-aware permutation not executable','INCONCLUSIVE','Frozen panel mappings provide crops only; no token spatial coordinates']])
 wr('HAPAX_MANIFEST_PLACEHOLDER.tsv',['artifact','sha256','status'],[[str(p.relative_to(ROOT)),sha(p),'FROZEN_INPUT'] for p in [OCC,TAX,PANEL]])
 json.dump({'inputs':{str(p.relative_to(ROOT)):sha(p) for p in [OCC,TAX,PANEL]},'status':{'ASTRO_HAPAX_SPATIAL_PATTERN':'INCONCLUSIVE','PAGES_WITH_HAPAX_GT_NONHAPAX_GLOBAL':sum(r[7]=='true' for r in out),'PAGES_WITH_HAPAX_GT_NONHAPAX_LOCAL':sum(r[11]=='true' for r in out),'REPRODUCIBLE':True},'outputs':sorted(x.name for x in OUT.iterdir())},open(OUT/'HAPAX_RESULTS_MANIFEST.json','w'),indent=2)
 print('pages',len(out),'global_gt',sum(r[7]=='true' for r in out),'local_gt',sum(r[11]=='true' for r in out))
if __name__=='__main__':main()
