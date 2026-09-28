#!/usr/bin/env python3
import csv,re,time,json,hashlib,resource,sys,collections
from pathlib import Path
ROOT=Path(__file__).resolve().parent;BASE=ROOT.parent
E3=BASE/'f68r2_e3_historical_operations_v1';PREP=BASE/'exploratory_astronomical_dictionary_search_v1'
sys.path.insert(0,str(E3));import run_e3

def normalize(form):return ''.join(re.findall('[a-z]+',form.lower()))
def key_sig(mapping,source,target):
    req=';'.join(f'{a}->{b}' for a,b in sorted(mapping.items()));lhs=set(mapping)
    forb=''.join(sorted(set(source)-lhs));return (req,forb,target)
def generate(source,target,stats=None,max_mappings=5):
    """Exact global partial injective mapping generator."""
    source,target=str(source),str(target);found=set();memo=set();
    counts=collections.Counter();freq=collections.Counter(source)
    def dfs(i,j,m,used,unmapped):
        state=(i,j,tuple(sorted(m.items())),tuple(sorted(used)),tuple(sorted(unmapped)))
        if state in memo:return
        memo.add(state);counts['states']+=1
        if i==len(source):
            if j==len(target):found.add(key_sig(m,source,target))
            return
        ch=source[i]
        if ch in m:
            if j<len(target) and m[ch]==target[j]:dfs(i+1,j+1,m,used,unmapped)
            else:counts['prune_mapped_repeat']+=1
            return
        if ch in unmapped:
            dfs(i+1,j,m,used,unmapped);return
        # Leave this source type globally unmapped.
        dfs(i+1,j,m,used,unmapped|{ch})
        # Map every occurrence of this source type, consuming current target.
        if j<len(target) and target[j] not in used and len(m)<max_mappings:
            nm=dict(m);nm[ch]=target[j];dfs(i+1,j+1,nm,used|{target[j]},unmapped)
        else:counts['prune_injective_or_target']+=1
    dfs(0,0,{},set(),set());counts['signatures']=len(found)
    if stats is not None:stats.update(counts)
    return sorted(found)
def profile_paths(profile,labels,lexicon,out_path):
    cache={};stats=collections.Counter();seen=set();count=0
    with out_path.open('w',encoding='utf-8',newline='') as f:
        fields=['profile_id','label_id','eva_token','identity','lexicon_id','source_form','transformed_source','required_mappings','forbidden_source_units','expected_output','scorer_trace']
        w=csv.DictWriter(f,fieldnames=fields,delimiter='\t',lineterminator='\n');w.writeheader()
        for lex in lexicon:
            transformed,trace=run_e3.transform(lex['normalized_form'],profile);cache_key=(transformed,)
            if cache_key not in cache:cache[cache_key]=transformed;stats['transformed_cache_misses']+=1
            else:stats['transformed_cache_hits']+=1
            source=cache[cache_key]
            for lab in labels:
                local=collections.Counter();sigs=generate(source,lab['zl3b_token'],local);stats.update({f'dfs_{k}':v for k,v in local.items()})
                for req,forb,target in sigs:
                    k=(lab['label_id'],target,lex['canonical_identity_id'],req,forb)
                    if k in seen:continue
                    seen.add(k);count+=1
                    w.writerow({'profile_id':profile['system_id'],'label_id':lab['label_id'],'eva_token':target,'identity':lex['canonical_identity_id'],'lexicon_id':lex['lexicon_id'],'source_form':lex['normalized_form'],'transformed_source':source,'required_mappings':req,'forbidden_source_units':forb,'expected_output':target,'scorer_trace':json.dumps(trace,separators=(',',':'))})
    stats.update({'path_count':count,'transformed_pairs':stats['transformed_cache_misses']});return stats
def read(p):
    with p.open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def main():
    labels=[x for x in read(PREP/'TARGET_STAR_LABELS.tsv') if x['page']=='f68r2'];lex=read(PREP/'ASTRONOMICAL_NAMES_LEXICON.tsv');reg=read(ROOT/'E3_OPERATION_PROFILES.tsv')
    # Persist the three required smoke profiles only; full E3 is intentionally not run.
    smoke=[reg[0],next(x for x in reg if x['article']=='DROP_AL'),reg[-1]];rows=[]
    for p in smoke:
        out=ROOT/f"SMOKE_{p['system_id']}.tsv";t=time.monotonic();st=profile_paths(p,labels,lex,out);rows.append(dict(st,profile_id=p['system_id'],elapsed_s=time.monotonic()-t,peak_rss_kb=resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,sha256=hashlib.sha256(out.read_bytes()).hexdigest()))
    with (ROOT/'SMOKE_PROFILE_RESULTS.tsv').open('w',newline='') as f:
        w=csv.DictWriter(f,fieldnames=list(rows[0]),delimiter='\t');w.writeheader();w.writerows(rows)
if __name__=='__main__':main()
