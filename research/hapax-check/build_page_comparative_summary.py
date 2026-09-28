#!/usr/bin/env python3
import csv, json, statistics
from collections import Counter, defaultdict
from pathlib import Path

ROOT=Path(__file__).resolve().parents[2]
OUT=ROOT/'research/hapax_check'
OCC=ROOT/'experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl'
TAX=ROOT/'research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv'
FP=ROOT/'research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv'

def read_tsv(p):
    with open(p) as f: return list(csv.DictReader(f,delimiter='\t'))

def main():
    taxonomy={r['page_id']:r for r in read_tsv(TAX)}
    fingerprints={r['page_id']:r for r in read_tsv(FP)}
    pages=defaultdict(list)
    with open(OCC) as f:
        for line in f:
            o=json.loads(line); p=o['folio']
            if p in taxonomy and taxonomy[p].get('inclusion_status')=='INCLUDED': pages[p].append(o)
    global_counts=Counter(o['token'] for z in pages.values() for o in z)
    local_counts={s:Counter(o['token'] for p,z in pages.items() if taxonomy[p]['visual_class']==s for o in z) for s in set(r['visual_class'] for r in taxonomy.values())}
    ordered=sorted(pages, key=lambda p:int(fingerprints[p]['page_order']))
    header=['page_id','physical_leaf','section','tokens','unique_tokens','global_hapax','global_hapax_types','global_non_hapax','section_local_hapax','section_local_hapax_types','section_local_non_hapax','global_hapax_gt_global_non_hapax','section_local_hapax_gt_section_local_non_hapax','hapax_fraction_global','hapax_fraction_section','unique_tokens_per_1000_tokens','visual_object_count','lines','tokens_per_line','Currier','hand','quire','source_artifacts']
    source_artifacts='experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl;research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv;research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv'
    data=[]
    for p in ordered:
        z=pages[p]; s=taxonomy[p]['visual_class']; toks=[o['token'] for o in z]; types=set(toks); gc=global_counts; lc=local_counts[s]
        gh=sum(gc[t]==1 for t in toks); lh=sum(lc[t]==1 for t in toks)
        gtypes={t for t in types if gc[t]==1}; ltypes={t for t in types if lc[t]==1}; fp=fingerprints[p]; lines=int(fp['line_count'])
        gnh=len(toks)-gh; lnh=len(toks)-lh
        data.append([p,taxonomy[p]['physical_leaf'],s,len(toks),len(types),gh,len(gtypes),gnh,lh,len(ltypes),lnh,gh>gnh,lh>lnh,f'{gh/len(toks):.8f}',f'{lh/len(toks):.8f}',f'{1000*len(types)/len(toks):.8f}','DATA_NOT_AVAILABLE',lines,f'{len(toks)/lines:.8f}',fp['currier'],fp['scribe'],fp['quire'],source_artifacts])
    with open(OUT/'PAGE_COMPARATIVE_SUMMARY.tsv','w') as f: csv.writer(f,delimiter='\t',lineterminator='\n').writerows([header,*data])
    def stats(col):
        vals=sorted(float(r[col]) for r in data); return min(vals),statistics.median(vals),max(vals)
    ti=header.index('tokens'); ui=header.index('unique_tokens'); gi=header.index('hapax_fraction_global'); si=header.index('hapax_fraction_section');
    def fmt(x): return f'{x:.8f}' if isinstance(x,float) else str(x)
    maxs=max(data,key=lambda r:float(r[si])); mins=min(data,key=lambda r:float(r[si]))
    global_count=sum(r[11] is True for r in data); section_count=sum(r[12] is True for r in data)
    md=['# Page comparative summary','', 'Descriptive table built only from frozen occurrence metadata, taxonomy, and page fingerprints. Rows follow frozen manuscript `page_order`. `global_hapax` and `section_local_hapax` are occurrence counts; accompanying `*_types` columns count distinct hapax token types on the page. Global and section-local non-hapax counts are calculated independently. Visual object counts are unavailable in the frozen artifacts.', '', '| page_id | physical_leaf | section | tokens | unique_tokens | global_hapax | global_non_hapax | section_local_hapax | section_local_non_hapax | hapax_fraction_global | hapax_fraction_section | unique_tokens_per_1000_tokens | visual_object_count | lines | tokens_per_line | Currier | hand | quire |','|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---|---:|---:|---|---|---|']
    for r in data: md.append('| '+' | '.join(map(str,[r[0],r[1],r[2],r[3],r[4],r[5],r[7],r[8],r[10],r[13],r[14],r[15],r[16],r[17],r[18],r[19],r[20],r[21]]))+' |')
    astro=[r for r in data if r[2]=='Astronomical']
    md += ['', '## Summary', '', f'- tokens min/median/max: {fmt(stats(ti))}', f'- unique_tokens min/median/max: {fmt(stats(ui))}', f'- global hapax fraction min/median/max: {fmt(stats(gi))}', f'- section-local hapax fraction min/median/max: {fmt(stats(si))}', f'- maximum section-local hapax fraction: `{maxs[0]}` ({maxs[14]})', f'- minimum section-local hapax fraction: `{mins[0]}` ({mins[14]})', f'- global_hapax > global_non_hapax: {global_count}/227', f'- section_local_hapax > section_local_non_hapax: {section_count}/227', '- Previous `12/227` section-local result used the correct section-local non-hapax definition; the earlier comparative `1/227` arose from incorrectly comparing section-local hapax with global non-hapax.', '', '## Astronomical panels', '', '| page_id | tokens | section_local_hapax | section_local_non_hapax | hapax_fraction_section | hapax_majority |', '|---|---:|---:|---:|---:|---|']
    for r in astro: md.append('| '+' | '.join(map(str,[r[0],r[3],r[8],r[10],r[14],'TRUE' if r[12] else 'FALSE']))+' |')
    md += ['', '## Source mapping', '', '- `occurrence_metadata.jsonl`: page token occurrences, token-derived counts/fractions, and hapax type/occurrence counts.', '- `VISUAL_CONTEXT_TAXONOMY.tsv`: `physical_leaf`, broad `section` (`visual_class`), and the explicit `DATA_NOT_AVAILABLE` visual-object field.', '- `VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv`: frozen manuscript `page_order`, `lines`, `Currier`, `hand`, and `quire`.', '', 'Source artifacts: `experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl`, `research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv`, and `research/visual_context/VISUAL_CONTEXT_PAGE_FINGERPRINTS.tsv`.']
    (OUT/'PAGE_COMPARATIVE_SUMMARY.md').write_text('\n'.join(md)+'\n')
if __name__=='__main__': main()
