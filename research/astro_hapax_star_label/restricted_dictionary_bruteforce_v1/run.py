#!/usr/bin/env python3
"""Frozen finite dictionary experiment; stdlib-only, deterministic checkpoints."""
import argparse,collections,csv,gzip,hashlib,json,math,os,platform,random,statistics,sys,time
from pathlib import Path
from engine import GRID,indexes,matching,evaluate,select,search,synthetic_dictionary,sampled_dictionary,words,transform
BASE=Path(__file__).resolve().parent
ROOT=BASE.parents[2]
ENDPOINTS=('JOINT','f68r1_TO_f68r2','f68r2_TO_f68r1')
COMMON=['run_id','seed','dictionary_id','split','system_id','complexity','score','coverage','null_family','completion_status']
FROZEN=['engine.py','run.py','prepare.py','test_engine.py','BRUTEFORCE_PROTOCOL.md','LEXICON_PROVENANCE.md',
        'TARGET_SCOPE.tsv','ASTRONOMICAL_LEXICON.tsv','CONTROL_LEXICON_POOL.tsv','LABEL_CONTROL_POOL.tsv',
        'CONTROL_LEXICON_REGISTRY.tsv','TRANSFORMATION_REGISTRY.tsv','SEARCH_GRID_MANIFEST.json',
        'UPSTREAM_INPUTS.tsv','UPSTREAM_SNAPSHOT.json','PREPARATION_VALIDATION.json']

def sha(p):return hashlib.sha256(p.read_bytes()).hexdigest()
def read(name):
    with (BASE/name).open(encoding='utf-8',newline='') as f:return list(csv.DictReader(f,delimiter='\t'))
def write(out,name,rows,fields=None):
    rows=list(rows)
    if fields is None:
        fields=list(dict.fromkeys(k for r in rows for k in r)) if rows else COMMON
    tmp=out/(name+'.tmp')
    with tmp.open('w',encoding='utf-8',newline='') as f:
        w=csv.DictWriter(f,fields,delimiter='\t',lineterminator='\n');w.writeheader();w.writerows(rows)
    tmp.replace(out/name)
def atomic_json(path,value):
    tmp=path.with_suffix('.tmp');tmp.write_text(json.dumps(value,sort_keys=True,indent=2)+'\n');tmp.replace(path)
def journal(out,action):
    with (out/'COMMANDS.jsonl').open('a') as f:f.write(json.dumps(dict(action=action,argv=sys.argv,python=sys.version,utc=time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime())))+'\n')
def freeze():
    dest=BASE/'FREEZE.json'
    assert not dest.exists(),'Already frozen; use run or verify, not freeze.'
    files={n:sha(BASE/n) for n in FROZEN}
    files.update({str(p.relative_to(BASE)):sha(p) for p in (BASE/'sources').iterdir() if p.is_file()})
    atomic_json(dest,dict(version=1,files=files,production_started=False,
                         python_version=platform.python_version(),utc=time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime())))
    journal(BASE,'FREEZE_BEFORE_PRODUCTION')
    print('Frozen',len(files),'files; no production scores computed.',flush=True)
def verify():
    f=json.loads((BASE/'FREEZE.json').read_text())
    assert all(sha(BASE/n)==h for n,h in f['files'].items()),'Frozen input/code changed. New experiment required.'
    upstream=json.loads((BASE/'UPSTREAM_SNAPSHOT.json').read_text())
    bad=[n for n,h in upstream.items() if not (ROOT/n).exists() or sha(ROOT/n)!=h]
    assert not bad,('Upstream files changed',bad)
    assert platform.python_version()==f['python_version'],'Use the frozen Python version for exact replay.'
    return f

def inputs():
    labels=read('TARGET_SCOPE.tsv')
    groups=collections.defaultdict(set)
    for r in read('ASTRONOMICAL_LEXICON.tsv'):groups[r['canonical_identity']].add(r['normalized_form'])
    terms=[dict(identity=k,forms=sorted(v)) for k,v in sorted(groups.items())]
    pools=collections.defaultdict(list)
    for r in read('CONTROL_LEXICON_POOL.tsv'):pools[r['dictionary_id']].append(r['normalized_form'])
    lpools=collections.defaultdict(list)
    for r in read('LABEL_CONTROL_POOL.tsv'):lpools[r['family']].append(r)
    return labels,terms,pools,lpools

def base_row(result,seed,dictionary,split,family='NONE',run='PRODUCTION'):
    return dict(run_id=run,seed=seed,dictionary_id=dictionary,split=split,system_id=result['system_id'],
        complexity=result['complexity'],score=result['score'],coverage=result['coverage'],null_family=family,
        completion_status='COMPLETE',matched=result['matched'],n=result['n'])

def sampled_labels(template,pool,rng,same_page):
    labels=[];distances=[];sources=[]
    for row in template:
        eligible=[x for x in pool if not same_page or x['page_id']==row['page_id']]
        delta=min(abs(len(x['token'])-len(row['token'])) for x in eligible)
        chosen=rng.choice([x for x in eligible if abs(len(x['token'])-len(row['token']))==delta])
        labels.append(dict(row,token=chosen['token']))
        distances.append(delta);sources.append(chosen['occurrence_id'])
    return labels,statistics.mean(distances),len(set(sources))

def independent_capacity(terms):
    return [dict(identity=f'FORM_{i:03}',forms=[f]) for i,f in enumerate(f for t in terms for f in t['forms'])]

def draw(family,seed,labels,terms,pools,lpools,real_indices,cache):
    rng=random.Random(seed);ls=labels;ts=terms
    info=dict(mean_length_distance=0,unique_source_occurrences='NA',changed_form_fraction='NA')
    if family in ('LENGTH_ENDPOINT','UNIGRAM','BIGRAM'):
        ts=synthetic_dictionary(terms,rng,family)
        original=[f for t in terms for f in t['forms']]; altered=[f for t in ts for f in t['forms']]
        info['changed_form_fraction']=sum(a!=b for a,b in zip(original,altered))/len(original)
        idx=indexes(ts)
    elif family in pools:
        ts,dist=sampled_dictionary(terms,pools[family],rng);info['mean_length_distance']=statistics.mean(dist)
        idx=indexes(ts,cache)
    else:
        ls,delta,n=sampled_labels(labels,lpools[family],rng,family!='OTHER_SECTION_LABEL')
        info.update(mean_length_distance=delta,unique_source_occurrences=n);idx=real_indices
    results=search(ls,idx)
    out=[]
    for endpoint,result in results.items():
        out.append(dict(base_row(result,seed,family,endpoint,family,'NULL'),**info))
    if family in pools:
        for endpoint,result in search(labels,indexes(independent_capacity(ts),cache)).items():
            out.append(dict(base_row(result,seed,family+'_INDEPENDENT_CAPACITY',endpoint,family+'_INDEPENDENT_CAPACITY','NULL'),**info))
    return out

def synthetic_dataset(labels,terms,idx,seed,noise):
    rng=random.Random(seed);planted=GRID[8];used=set();injected=[]
    for term in terms:
        fs=term['forms'][:];rng.shuffle(fs)
        for f in fs:
            token=transform(f,planted)
            if token and token not in used:
                used.add(token);injected.append((term['identity'],f,token));break
    rng.shuffle(injected)
    # Fill alternating page slots to guarantee both directions contain planted pairs.
    page1=[x for x in labels if x['page_id']=='f68r1'];page2=[x for x in labels if x['page_id']=='f68r2']
    slots=[row for pair in zip(page1,page2) for row in pair]+page1[len(page2):]
    forbidden=set().union(*(set(x) for x in idx))
    alphabet=sorted(set(''.join(x['token'] for x in labels)))
    truth={};forms={};out=[]
    corrupted=set(rng.sample(range(len(injected)),int(noise*len(injected))))
    for i,slot in enumerate(slots):
        if i<len(injected):
            identity,form,token=injected[i];truth[slot['occurrence_id']]=identity;forms[slot['occurrence_id']]=form
            if i in corrupted:
                p=rng.randrange(len(token));char=rng.choice([c for c in alphabet if c!=token[p]])
                token=token[:p]+char+token[p+1:]
        else:
            while True:
                token=''.join(rng.choice(alphabet) for _ in range(len(slot['token'])))
                if token not in forbidden:break
        out.append(dict(slot,token=token))
    noisy={slots[i]['occurrence_id'] for i in corrupted}
    return out,truth,forms,noisy

def synthetic(out,labels,terms,idx):
    destination=out/'SYNTHETIC_CHECKPOINT.json'
    if destination.exists():
        saved=json.loads(destination.read_text());assert saved['freeze_sha256']==sha(BASE/'FREEZE.json')
        assert saved['rows_sha256']==hashlib.sha256(json.dumps(saved['rows'],sort_keys=True).encode()).hexdigest()
        return saved['rows']
    rows=[]
    for j in range(20):
        seed=800000+j
        for noise in (0,0.1,0.25):
            ls,truth,forms,noisy=synthetic_dataset(labels,terms,idx,seed,noise)
            obs=search(ls,idx);nulls=[]
            for k in range(99):
                nts=synthetic_dictionary(terms,random.Random(seed*1000+k),'LENGTH_ENDPOINT')
                nulls.append(search(ls,indexes(nts)))
            for endpoint,result in obs.items():
                eligible={l['occurrence_id'] for l in ls if endpoint=='JOINT' or l['page_id']==endpoint.split('_TO_')[1]}
                surviving=(set(truth)-noisy)&eligible
                denominator=len(surviving)
                correct=sum(result['pairs'].get(k)==truth[k] for k in surviving)
                equivalent=sum(transform(forms[k],GRID[result['index']])==transform(forms[k],GRID[8]) for k in surviving)/denominator if denominator else 0
                p=(1+sum(n[endpoint]['score']>=result['score'] for n in nulls))/100
                noisy_eligible=noisy&eligible
                recovery=correct/denominator if denominator else 0
                threshold=.8 if endpoint=='JOINT' else .6
                passed=recovery>=threshold and equivalent>=.9 and p<=.05
                rows.append(dict(base_row(result,seed,'ASTRONOMY_V1',endpoint,'SYNTHETIC_LENGTH_ENDPOINT','SYNTHETIC'),
                    noise=noise,injected=len(truth),unmatched=57-len(truth),surviving_pairs=denominator,
                    true_pair_recovery=recovery,equivalent_fraction=equivalent,planted_system='S008',
                    noisy_pair_recovery=sum(result['pairs'].get(k)==truth[k] for k in noisy_eligible)/len(noisy_eligible) if noisy_eligible else 'NA',
                    null_replicates=99,empirical_p=p,sensitivity_pass=int(passed)))
        if (j+1)%5==0:print(f'Synthetic seeds complete: {j+1}/20',flush=True)
    atomic_json(destination,dict(freeze_sha256=sha(BASE/'FREEZE.json'),rows=rows,
        rows_sha256=hashlib.sha256(json.dumps(rows,sort_keys=True).encode()).hexdigest()))
    return rows

def synthetic_pass(rows):
    return all(sum(int(r['sensitivity_pass']) for r in rows if r['noise']==noise and r['split']==ep)>=18
               for noise in (0,0.1,0.25) for ep in ENDPOINTS)

def checkpoint_write(path,rows):
    data=gzip.compress(json.dumps(rows,sort_keys=True,separators=(',',':')).encode(),mtime=0)
    tmp=path.with_suffix('.tmp');tmp.write_bytes(data);tmp.replace(path)
    return hashlib.sha256(data).hexdigest()

def null_runs(out,labels,terms,pools,lpools,idx,deadline):
    directory=out/'checkpoints';directory.mkdir(exist_ok=True)
    manifest_path=out/'CHECKPOINT_MANIFEST.json'
    if manifest_path.exists():
        manifest=json.loads(manifest_path.read_text());assert manifest['freeze_sha256']==sha(BASE/'FREEZE.json')
    else:manifest=dict(freeze_sha256=sha(BASE/'FREEZE.json'),shards={})
    families=json.loads((BASE/'SEARCH_GRID_MANIFEST.json').read_text())['families'][:9]
    cache={};allrows=[]
    for fi,family in enumerate(families):
        for start in range(0,10000,100):
            name=f'{family}_{start:05}.json.gz';p=directory/name
            if name in manifest['shards']:
                assert sha(p)==manifest['shards'][name],'Checkpoint checksum mismatch'
                rows=json.loads(gzip.decompress(p.read_bytes()))
                assert len(rows)==(600 if family in pools else 300)
                assert {r['seed'] for r in rows}==set(range(570000+fi*100000+start,570000+fi*100000+start+100))
            else:
                if time.monotonic()>deadline:return allrows,False
                rows=[]
                for replica in range(start,start+100):
                    rows.extend(draw(family,570000+fi*100000+replica,labels,terms,pools,lpools,idx,cache))
                manifest['shards'][name]=checkpoint_write(p,rows);atomic_json(manifest_path,manifest)
            allrows.extend(rows)
            if (start+100)%1000==0:print(f'{family}: {start+100}/10000 complete',flush=True)
    return allrows,True

def empirical(observed,nullrows):
    comparisons=[]
    families=json.loads((BASE/'SEARCH_GRID_MANIFEST.json').read_text())['families'][:9]
    for ep in ENDPOINTS:
        for family in families+[f+'_INDEPENDENT_CAPACITY' for f in ('HISTORICAL_NAMES','MEDIEVAL_LATIN','ARABO_LATIN')]:
            rs=[r for r in nullrows if r['split']==ep and r['null_family']==family]
            assert len(rs)==10000,(ep,family,len(rs))
            p=(1+sum(r['score']>=observed[ep]['score'] for r in rs))/10001
            comparisons.append(dict(base_row(observed[ep],570000,'ASTRONOMY_V1',ep,family,'INFERENCE'),
                null_replicates=len(rs),empirical_p=p,corrected_p=min(1,p*27) if family in families else 'NA',
                mean_null_score=statistics.mean(r['score'] for r in rs),mean_null_coverage=statistics.mean(r['coverage'] for r in rs),
                max_null_coverage=max(r['coverage'] for r in rs),advantage=observed[ep]['coverage']-statistics.mean(r['coverage'] for r in rs),
                fdr_q='NA',comparison_role='PRIMARY' if family in families else 'CONSERVATIVE_SENSITIVITY'))
    ordered=sorted([r for r in comparisons if r['comparison_role']=='PRIMARY'],key=lambda r:r['empirical_p'])
    running=1.0
    for rank in range(len(ordered),0,-1):
        running=min(running,ordered[rank-1]['empirical_p']*len(ordered)/rank)
        ordered[rank-1]['fdr_q']=running
    return comparisons

def invariance(labels,terms,idx,observed):
    shuffled=labels[:];random.Random(570000).shuffle(shuffled)
    assert {ep:(r['score'],r['coverage']) for ep,r in search(shuffled,idx).items()}=={ep:(r['score'],r['coverage']) for ep,r in observed.items()}
    # Renaming identity groups is a bijection; canonical sort ranks travel with keys.
    permutation=list(range(len(terms)));random.Random(570001).shuffle(permutation)
    rename={t['identity']:f'ANON_{v:03}' for t,v in zip(terms,permutation)}
    inverse={v:k for k,v in rename.items()}
    renamed=[{token:tuple(rename[k] for k in ids) for token,ids in index.items()} for index in idx]
    restored=[{token:tuple(inverse[k] for k in ids) for token,ids in index.items()} for index in renamed]
    assert restored==idx
    assert search(labels,restored)==observed
    rows=[]
    for f in ('SHUFFLED_LABEL','SHUFFLED_IDENTITIES'):
        for ep,r in observed.items():rows.append(dict(base_row(r,570000,f,ep,f,'EXACT_INVARIANCE'),empirical_p=1,replicates=1,quotient_space_size=1))
    return rows

def stability(out,labels,terms,idx,observed,pools):
    loo=[];sub=[];loores=[];subres=[]
    for excluded in labels:
        result=search([x for x in labels if x['occurrence_id']!=excluded['occurrence_id']],idx);loores.append((excluded['occurrence_id'],result))
        for ep,r in result.items():
            original=observed[ep]['pairs'];retained={k:v for k,v in original.items() if k!=excluded['occurrence_id']}
            loo.append(dict(base_row(r,570000,'ASTRONOMY_V1',ep,run='LOO'),excluded_occurrence=excluded['occurrence_id'],
                system_changed=int(r['system_id']!=observed[ep]['system_id']),
                surviving_pair_fraction=sum(r['pairs'].get(k)==v for k,v in retained.items())/len(retained) if retained else 'NA'))
    for seed in range(900000,900200):
        rng=random.Random(seed);sample=[]
        for page in ('f68r1','f68r2'):
            page_labels=[l for l in labels if l['page_id']==page];sample+=rng.sample(page_labels,int(.8*len(page_labels)))
        result=search(sample,idx);subres.append(({x['occurrence_id'] for x in sample},result))
        for ep,r in result.items():sub.append(base_row(r,seed,'ASTRONOMY_V1',ep,run='RESAMPLE'))
    write(out,'LEAVE_ONE_OUT_RESULTS.tsv',loo);write(out,'RESAMPLING_RESULTS.tsv',sub)
    # All equality alternatives, and globally optimal edge-exclusion margins.
    selected=observed['JOINT'];selected_idx=idx[selected['index']];rows=[];alternative_rows=[]
    other_strings=set()
    for pool in pools.values():
        for word in pool:
            for rule in GRID:other_strings.add(transform(word,rule))
    pair_fields=COMMON+['matched','n','occurrence_id','canonical_identity','pair_score','local_margin','assignment_margin','rank','supporting_systems','neighbor_support','loo_survival','resample_survival','cross_direction_agreement','historical_alternative','eligible']
    for label in labels:
        key=label['occurrence_id'];alts=selected_idx.get(label['token'],())
        for t in alts:
            alternative_rows.append(dict(base_row(selected,570000,'ASTRONOMY_V1','JOINT',run='PAIR_ALTERNATIVE'),occurrence_id=key,canonical_identity=t,pair_score=1,rank=1,selected=int(selected['pairs'].get(key)==t)))
        if key not in selected['pairs']:continue
        term=selected['pairs'][key]
        margin=selected['matched']-len(matching(labels,selected_idx,forbidden=(key,term)))
        support=[i for i,index in enumerate(idx) if matching(labels,index).get(key)==term]
        axes=('article','orthography','vowel','abbreviation')
        neighbors=sum(sum(GRID[i][a]!=GRID[selected['index']][a] for a in axes)==1 for i in support)
        loo_fraction=sum(r['JOINT']['pairs'].get(key)==term for excluded,r in loores if excluded!=key)/(len(labels)-1)
        included=[r for ids,r in subres if key in ids]
        sub_fraction=sum(r['JOINT']['pairs'].get(key)==term for r in included)/len(included) if included else 0
        held_ep='f68r2_TO_f68r1' if label['page_id']=='f68r1' else 'f68r1_TO_f68r2'
        train_ep='f68r1_TO_f68r2' if label['page_id']=='f68r1' else 'f68r2_TO_f68r1'
        consistent=observed[held_ep]['pairs'].get(key)==term and observed[train_ep]['train_pairs'].get(key)==term
        alt=label['token'] in other_strings
        eligible=len(support)>=3 and neighbors>=2 and margin>=1 and loo_fraction>=.95 and sub_fraction>=.9 and consistent and not alt
        rows.append(dict(base_row(selected,570000,'ASTRONOMY_V1','JOINT',run='PAIR_STABILITY'),occurrence_id=key,
            canonical_identity=term,pair_score=1,local_margin=int(len(alts)==1),assignment_margin=margin,rank=1,
            supporting_systems=len(support),neighbor_support=neighbors,loo_survival=loo_fraction,resample_survival=sub_fraction,
            cross_direction_agreement=int(consistent),historical_alternative=int(alt),eligible=int(eligible)))
    write(out,'PAIR_STABILITY.tsv',rows,pair_fields)
    write(out,'PAIR_ALTERNATIVES.tsv',alternative_rows,COMMON+['matched','n','occurrence_id','canonical_identity','pair_score','rank','selected'])
    stable=all(r['coverage']>=.8*observed[r['split']]['coverage'] for r in loo)
    resample_ok=all(sum(r['coverage']>0 for r in sub if r['split']==ep)>=180 for ep in ENDPOINTS[1:])
    return rows,stable,resample_ok

def metrics(result,labels,idx):
    scores=[1]*result['matched']
    values=[len(idx.get(l['token'],())) for l in labels]
    return dict(total_assignment_cost=len(labels)-result['matched'],mean_pair_score=statistics.mean(scores) if scores else 'NA',
        median_pair_score=statistics.median(scores) if scores else 'NA',unique_identities=len(set(result['pairs'].values())),
        unmatched=len(labels)-result['matched'],term_collisions=sum(max(0,n-1) for n in values),
        level=GRID[result['index']]['level'])

def finish(out,labels,terms,idx,obs,nullrows,synrows):
    comparisons=empirical(obs,nullrows);write(out,'INFERENCE_RESULTS.tsv',comparisons)
    write(out,'INVARIANCE_RESULTS.tsv',invariance(labels,terms,idx,obs))
    pairrows,loo_ok,resample_ok=stability(out,labels,terms,idx,obs,inputs()[2])
    syn_ok=synthetic_pass(synrows)
    endpoint={}
    for ep in ENDPOINTS:
        primary=[r for r in comparisons if r['split']==ep and r['comparison_role']=='PRIMARY']
        sensitivity=[r for r in comparisons if r['split']==ep and r['comparison_role']!='PRIMARY']
        endpoint[ep]=dict(empirical_p=max(r['empirical_p'] for r in primary),corrected_p=max(r['corrected_p'] for r in primary),
            best_null_advantage=min(r['advantage'] for r in primary),
            conservative_pass=all(r['empirical_p']<.05 for r in sensitivity))
    def qualifies(ep):
        e=endpoint[ep];return e['corrected_p']<.05 and e['best_null_advantage']>0 and e['conservative_pass'] and syn_ok and loo_ok
    rule1=GRID[obs[ENDPOINTS[1]]['index']];rule2=GRID[obs[ENDPOINTS[2]]['index']]
    family_equal=all(rule1[a]==rule2[a] for a in ('article','orthography','vowel'))
    cross=all(qualifies(ep) for ep in ENDPOINTS[1:]) and family_equal and resample_ok
    joint=qualifies('JOINT')
    conclusion='ENGINE_INCONCLUSIVE' if not syn_ok else 'CROSS_PAGE_SIGNAL' if cross else 'CORPUS_LEVEL_SIGNAL_ONLY' if joint else 'NO_SIGNAL'
    candidates=[dict(r,status='CANDIDATE_FORMAL_COMPATIBILITY') for r in pairrows if r['eligible'] and cross]
    fields=list(pairrows[0]) + ['status'] if pairrows else COMMON+['occurrence_id','canonical_identity','status']
    write(out,'PAIR_CANDIDATES.tsv',candidates,fields)
    pairset=set(obs['JOINT']['pairs']);strata=[]
    for name,rs in [('HAPAX',[l for l in labels if l['hapax']=='1']),('NON_HAPAX',[l for l in labels if l['hapax']=='0']),('f68r1',[l for l in labels if l['page_id']=='f68r1']),('f68r2',[l for l in labels if l['page_id']=='f68r2'])]:
        count=sum(l['occurrence_id'] in pairset for l in rs);r=dict(obs['JOINT'],matched=count,n=len(rs),coverage=count/len(rs))
        strata.append(dict(base_row(r,570000,'ASTRONOMY_V1',name,run='POST_SELECTION_STRATA'),model_selection_uses_hapax='NO'))
    write(out,'HAPAX_STRATIFIED_RESULTS.tsv',strata)
    h=next(r for r in strata if r['split']=='HAPAX');nh=next(r for r in strata if r['split']=='NON_HAPAX')
    total=obs['JOINT']['matched']
    hyper=sum(math.comb(26,k)*math.comb(31,total-k) for k in range(h['matched'],min(26,total)+1) if 0<=total-k<=31)/math.comb(57,total)
    summary=dict(conclusion=conclusion,synthetic_recovery='PASS' if syn_ok else 'FAIL',
        sensitivity_status='PASS' if syn_ok else 'ENGINE_SENSITIVITY_INSUFFICIENT',
        target_valid=True,labels=57,pages={'f68r1':30,'f68r2':27},null_complete=True,null_replicates=90000,
        extra_conservative_null_replicates=30000,grid_systems=len(GRID),canonical_identities=len(terms),
        unique_search_forms=sum(len(t['forms']) for t in terms),cross_page_signal=cross,corpus_level_signal=joint,
        stable_candidates=len(candidates),hapax_advantage=h['coverage']-nh['coverage'],hapax_descriptive_p=hyper,
        loo_stability_pass=loo_ok,resampling_transfer_pass=resample_ok,
        endpoints={ep:dict(system_id=obs[ep]['system_id'],complexity=obs[ep]['complexity'],matched=obs[ep]['matched'],n=obs[ep]['n'],coverage=obs[ep]['coverage'],**endpoint[ep]) for ep in ENDPOINTS})
    atomic_json(out/'SUMMARY.json',summary)
    jointrows=[dict(base_row(obs['JOINT'],570000,'ASTRONOMY_V1','JOINT'),**metrics(obs['JOINT'],labels,idx[obs['JOINT']['index']]),**endpoint['JOINT'])]
    crossrows=[]
    for ep in ENDPOINTS[1:]:
        held=[l for l in labels if l['page_id']==ep.split('_TO_')[1]]
        crossrows.append(dict(base_row(obs[ep],570000,'ASTRONOMY_V1',ep),**metrics(obs[ep],held,idx[obs[ep]['index']]),**endpoint[ep],train_matched=obs[ep]['train_matched']))
    write(out,'JOINT_RESULTS.tsv',jointrows);write(out,'CROSS_PAGE_RESULTS.tsv',crossrows)
    report(out,summary,synrows,comparisons)
    return summary

def report(out,s,synrows,comparisons):
    j=s['endpoints']['JOINT'];a=s['endpoints'][ENDPOINTS[1]];b=s['endpoints'][ENDPOINTS[2]]
    table='\n'.join(f"| {ep} | {r['system_id']} | {r['matched']}/{r['n']} | {r['empirical_p']:.6g} | {r['corrected_p']:.6g} | {r['best_null_advantage']:.6g} |" for ep,r in s['endpoints'].items())
    syn='\n'.join(f"| {noise} | {ep} | {sum(int(r['sensitivity_pass']) for r in synrows if r['noise']==noise and r['split']==ep)}/20 |" for noise in (0,.1,.25) for ep in ENDPOINTS)
    limits='''The 64 systems transform historical strings without learning a Latin-to-EVA cipher.
Thus a negative finding rejects only these transformations and this 31-identity
lexicon; it does not reject astronomical naming in the manuscript. Matching has
a ceiling of 31/57, imposed by historical identities, not 63 attestations.
Historical capacity blocks match spelling opportunities but are not historical
synonym sets; independent-capacity sensitivity gives controls more freedom.
Latin/Arabic language and date distributions are not fully balanced. Alchemical
forms come from an edited early-print witness of a medieval tradition. Circular
controls have only 29 independent observed occurrences and are bootstrapped to 57.
Bigram-preserving randomizations can retain most original words; their unchanged
fraction is reported, not concealed. They may offer little discriminatory power.
All comparisons concern anonymous matching, not a labeled star map.
'''
    status=f'''RESTRICTED_DICTIONARY_BRUTEFORCE=COMPLETE
TARGET_SCOPE_VALID=YES
TARGET_LABELS=57
F68R1_LABELS=30
F68R2_LABELS=27
ASTRONOMICAL_LEXICON_FROZEN=YES
TRANSFORMATION_GRID_FROZEN=YES
SYNTHETIC_RECOVERY={s['synthetic_recovery']}
NULL_CONTROLS_COMPLETE=YES
BEST_SYSTEM_ID={j['system_id']}
BEST_SYSTEM_COMPLEXITY={j['complexity']}
JOINT_MATCHED_COVERAGE={j['coverage']}
F68R1_TO_F68R2_HELDOUT_COVERAGE={a['coverage']}
F68R2_TO_F68R1_HELDOUT_COVERAGE={b['coverage']}
BEST_NULL_ADVANTAGE={j['best_null_advantage']}
EMPIRICAL_P={j['empirical_p']}
CORRECTED_P={j['corrected_p']}
CROSS_PAGE_SIGNAL={'YES' if s['cross_page_signal'] else 'NO'}
CORPUS_LEVEL_SIGNAL={'YES' if s['corpus_level_signal'] else 'NO'}
STABLE_PAIR_CANDIDATES={s['stable_candidates']}
HAPAX_SIGNAL_ADVANTAGE={s['hapax_advantage']}
PRIMARY_CONCLUSION={s['conclusion']}
SPECIFIC_STAR_IDENTIFICATIONS_AUTHORIZED=NO
DECIPHERMENT_CLAIM_AUTHORIZED=NO'''
    text=f'''# Restricted dictionary brute-force v1 results

Primary conclusion: **{s['conclusion']}**. Synthetic sensitivity: **{s['sensitivity_status']}**.
No specific star identification or decipherment claim is authorized.

| Endpoint | Selected system | Matched | Worst-family empirical p | FWER p | Best-null coverage advantage |
|---|---|---:|---:|---:|---:|
{table}

The reported system is the deterministic optimum of the finite search. A tied
zero-score optimum is not a discovered correspondence. Selection evaluated all
64 systems, 57 LOO exclusions and 200 page-stratified subsamples. Level A and B
system rows are separated in SYSTEM_RESULTS.tsv; level C was excluded in advance.
No parameter was selected using hapax status or held-out page scores.

There are 90,000 completed informative null datasets (10,000 per family), plus
30,000 conservative historical independent-capacity searches. Every one repeats
joint selection and both train/held-out directions over the complete grid.
INFERENCE_RESULTS.tsv contains all 27 primary comparisons, their Bonferroni and BH
corrections, and nine conservative sensitivity comparisons. Null scores and selected
models are in NULL_RESULTS.tsv; checkpoints allow deterministic seed replay.
The exact order/identity-renaming controls have p=1 by construction and are listed
separately as invariance diagnostics, not informative randomizations.

Answers to the six task questions:

1. Astronomy superiority: {'YES' if s['corpus_level_signal'] or s['cross_page_signal'] else 'NOT ESTABLISHED'}; all required control families and corrections were evaluated.
2. Cross-page transfer meeting protocol: {'YES' if s['cross_page_signal'] else 'NO'}.
3. Model-selection-aware null test: COMPLETE; worst-family joint corrected p={j['corrected_p']}.
4. Hapax minus non-hapax coverage={s['hapax_advantage']}; descriptive exact upper-tail p={s['hapax_descriptive_p']}. This was not used in selection.
5. Published stable formal candidates={s['stable_candidates']}; empty tables are intentional when the criteria fail.
6. Engine recovery={s['synthetic_recovery']}; results below. {'A real-data negative is not interpretable as absence of signal because sensitivity failed.' if s['synthetic_recovery']=='FAIL' else 'The engine detects the preregistered injected exact-normalization signal; this does not establish power for unknown ciphers.'}

| Corruption fraction | Endpoint | Passing seeds |
|---|---|---:|
{syn}

Synthetic controls use 57 slots and 31 or fewer planted identities, at least 26
unmatched items, 30/27 split, 20 seeds per noise level, and 99 fully searched
synthetic nulls per dataset. PASS requires at least 18/20 per cell.

## Evidence boundaries

{limits}
The old enrichment script and COHORT_MEMBERS.tsv already differed from their
registered SHA256 on entry. They were not used as authoritative inputs. The target
and corpus metadata passed frozen checksums, and 26/31 hapax strata were recounted
independently. UPSTREAM_SNAPSHOT.json verifies that this task changed no upstream
files. See VALIDATION_REPORT.md and REPRODUCIBILITY.md.

## Mandatory final status

```text
{status}
```
'''
    (out/'RESULTS_REPORT.md').write_text(text)

def seal(out):
    files=[p for p in out.rglob('*') if p.is_file() and '__pycache__' not in p.parts and p.name not in ('SHA256SUMS','RUNNING.lock') and not p.name.endswith('.tmp')]
    (out/'SHA256SUMS').write_text(''.join(f'{sha(p)}  {p.relative_to(out)}\n' for p in sorted(files)))

def run(out):
    out.mkdir(parents=True,exist_ok=True);verify();journal(out,'RUN_OR_RESUME')
    lock=out/'RUNNING.lock'
    # Prevent concurrent mutation of one checkpoint/output directory.
    fd=os.open(lock,os.O_CREAT|os.O_EXCL|os.O_WRONLY);os.write(fd,str(os.getpid()).encode());os.close(fd)
    try:
        (out/'RUN_STATUS.json').write_text(json.dumps(dict(status='INCOMPLETE',best_system='WITHHELD_UNTIL_CORRECTION'))+'\n')
        labels,terms,pools,lpools=inputs();idx=indexes(terms,{})
        synrows=synthetic(out,labels,terms,idx);write(out,'SYNTHETIC_RECOVERY_RESULTS.tsv',synrows)
        # Results retained in memory until the full null search and correction complete.
        observed=search(labels,idx)
        nullrows,complete=null_runs(out,labels,terms,pools,lpools,idx,time.monotonic()+14400)
        if not complete:
            write(out,'NULL_RESULTS.tsv',nullrows)
            (out/'RESULTS_REPORT.md').write_text('# Incomplete experiment\n\nRESTRICTED_DICTIONARY_BRUTEFORCE=BLOCKED\nNULL_CONTROLS_COMPLETE=NO\nPRIMARY_CONCLUSION=ENGINE_INCONCLUSIVE\nBest system withheld pending full null completion. Resume with the same command.\n')
            return
        write(out,'NULL_RESULTS.tsv',nullrows)
        systemrows=[]
        for split,ls in [('JOINT',labels),('TRAIN_f68r1',[l for l in labels if l['page_id']=='f68r1']),('TRAIN_f68r2',[l for l in labels if l['page_id']=='f68r2'])]:
            for i,rule in enumerate(GRID):
                r=dict(evaluate(ls,idx[i],rule),index=i)
                systemrows.append(dict(base_row(r,570000,'ASTRONOMY_V1',split),**metrics(r,ls,idx[i])))
        write(out,'SYSTEM_RESULTS.tsv',systemrows)
        summary=finish(out,labels,terms,idx,observed,nullrows,synrows)
        verify()
        atomic_json(out/'RUN_STATUS.json',dict(status='COMPLETE',freeze_sha256=sha(BASE/'FREEZE.json'),primary_conclusion=summary['conclusion']))
        (out/'VALIDATION_REPORT.md').write_text('''# Validation

- Verified frozen TARGET_SETS.tsv, D1 lexicon and corpus-metadata SHA256 before preparation.
- Reconstructed 57 unique occurrence IDs, 30/27 pages and 26/31 hapax strata from the full corpus.
- All frozen experiment inputs and executable bytes verified on run/resume and completion.
- 900 atomic hashed checkpoint shards: 90,000 primary and 30,000 independent-capacity datasets.
- Every primary family/end-point has exactly 10,000 rows; 27 multiplicity-corrected comparisons.
- Exact invariance checks passed for within-page order permutation and identity-group bijection.
- 57 complete LOO searches and 200 page-stratified subsamples.
- All upstream snapshot bytes unchanged. Pre-existing enrichment checksum discrepancies documented.
- Automated tests and independent replay checks: see VERIFICATION.json and TEST_RESULTS.txt.

Synthetic sensitivity is separately gated and must be read from SUMMARY.json;
completed computation is not evidence that all sensitivity or signal gates passed.
''')
        print(json.dumps(summary,indent=2),flush=True)
    finally:
        lock.unlink();seal(out)

def main():
    parser=argparse.ArgumentParser();parser.add_argument('action',choices=['freeze','run','verify']);parser.add_argument('--out',type=Path,default=BASE)
    args=parser.parse_args()
    if args.action=='freeze':freeze()
    elif args.action=='verify':verify();print('Frozen inputs and upstream bytes verified.')
    else:run(args.out.resolve())

if __name__=='__main__':main()
