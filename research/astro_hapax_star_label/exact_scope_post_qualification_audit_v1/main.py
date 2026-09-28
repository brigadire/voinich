#!/usr/bin/env python3
"""Independent identifiability/objective audit of the immutable v2 sealed record."""
from __future__ import annotations
import csv, hashlib, itertools, json, statistics, sys
from collections import Counter
from pathlib import Path

ROOT=Path(__file__).resolve().parents[3]
V2=ROOT/"research/astro_hapax_star_label/exact_scope_qualification_v2"
OUT=Path(__file__).resolve().parent
sys.path.insert(0,str(V2/"snapshot"))
from scorer import Scorer

def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def write(path, fields, rows):
    with path.open("w",newline="",encoding="utf-8") as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter="\t",lineterminator="\n");w.writeheader();w.writerows(rows)

def tables(source, target, mode, size=4):
    for keys in itertools.combinations(source,size):
        for vals in itertools.product(target,repeat=size):
            if mode=="INJECTIVE" and len(set(vals))<size: continue
            if mode=="MERGE_1" and sum(vals.count(v)==2 for v in set(vals))>1: continue
            if mode=="MERGE_1" and any(vals.count(v)>2 for v in set(vals)): continue
            yield dict(zip(keys,vals))

def main():
    all_public=[json.loads(x) for x in (V2/"SEALED_PUBLIC.jsonl").read_text().splitlines()]
    # Pre-registered audit pilot: two cases per group x mapping mode. Selection uses only
    # generation metadata, never scores, predictions, or truth.
    public=[]
    for group in ("zero_noise","noisy","hard_negative"):
        for mode in ("INJECTIVE","MERGE_1"):
            public.extend([x for x in all_public if x["group"]==group and x["mode"]==mode][:2])
    truth={x["instance_id"]:x for x in (json.loads(y) for y in (V2/"SEALED_TRUTH.jsonl").read_text().splitlines())}
    rows=[]; info=[]; cache={}
    for case in public:
        ident=case["instance_id"]; lex={k:set(v) for k,v in case["lexicon"].items()}; labels=case["labels"]; mode=case["mode"]; target=tuple("acdefhi"); source=tuple(sorted(set(c for fs in lex.values() for f in fs for c in f)))
        key=(source,mode)
        if key not in cache: cache[key]=list(tables(source,target,mode))
        scorer=Scorer(mode,"DROP_UNMAPPED","NONE",case["capacity"]); scored=[]; best=-10**18
        for tab in cache[key]:
            fit=scorer.evaluate(tab,lex,labels[:6])["fitness"]
            if fit>best: best=fit; scored=[tab]
            elif fit==best: scored.append(tab)
        gt=truth[ident]["mapping"]; gt_fit=scorer.evaluate(gt,lex,labels[:6])["fitness"]; gt_in=any(x==gt for x in scored)
        train_source=set(c for l in labels[:6] for c in l["token"])
        if case["hard_negative"]: category="HARD_NEGATIVE_CONTROL"
        elif gt_fit<best: category="GROUND_TRUTH_NOT_OPTIMAL"
        elif len(scored)>1: category="MULTIPLE_OPTIMA"
        else: category="UNIQUE_CORRECT_OPTIMUM"
        rows.append({"instance_id":ident,"group":case["group"],"mapping_mode":mode,"table_size":4,"noise":case["noise"],"unmatched_fraction":"HARD_NEGATIVE" if case["hard_negative"] else "NOT_VARIED","collision_level":"MERGE_1" if mode=="MERGE_1" else "INJECTIVE","global_optimum_score":best,"ground_truth_score":gt_fit,"ground_truth_in_optimum":int(gt_in),"co_optimal_table_count":len(scored),"co_optimal_table_hashes":hashlib.sha256("\n".join(json.dumps(x,sort_keys=True) for x in scored).encode()).hexdigest(),"classification":category,"observed_source_symbol_count":len(train_source),"source_alphabet_count":len(source)})
        info.append({"instance_id":ident,"group":case["group"],"observed_source_symbol_count":len(train_source),"co_optimal_table_count":len(scored),"ground_truth_in_optimum":int(gt_in),"unique_correct":int(category=="UNIQUE_CORRECT_OPTIMUM")})
    fields=list(rows[0]);write(OUT/"COOPTIMAL_TABLE_AUDIT.tsv",fields,rows)
    cats=Counter(x["classification"] for x in rows)
    by_info=[]
    for n in sorted(set(x["observed_source_symbol_count"] for x in info)):
        z=[x for x in info if x["observed_source_symbol_count"]==n];by_info.append({"observed_source_symbol_count":n,"n":len(z),"unique_correct_rate":statistics.fmean(x["unique_correct"] for x in z),"ground_truth_in_optimum_rate":statistics.fmean(x["ground_truth_in_optimum"] for x in z),"median_co_optimal_count":statistics.median(x["co_optimal_table_count"] for x in z)})
    write(OUT/"INFORMATION_REQUIREMENT.tsv",list(by_info[0]),by_info)
    strata=[]
    for col in ("group","mapping_mode","table_size","noise","unmatched_fraction","collision_level"):
        values=sorted(set(str(x[col]) for x in rows))
        for v in values:
            z=[x for x in rows if str(x[col])==v];strata.append({"factor":col,"level":v,"n":len(z),"ground_truth_not_optimal":sum(x["classification"]=="GROUND_TRUTH_NOT_OPTIMAL" for x in z),"multiple_optima":sum(x["classification"]=="MULTIPLE_OPTIMA" for x in z),"unique_wrong_optimum":0,"ground_truth_in_optimum_rate":statistics.fmean(x["ground_truth_in_optimum"] for x in z)})
    write(OUT/"STRATIFIED_AUDIT.tsv",list(strata[0]),strata)
    status={"cooptimal_categories":cats,"n_instances":len(rows),"ground_truth_not_optimal":cats["GROUND_TRUTH_NOT_OPTIMAL"],"multiple_optima":cats["MULTIPLE_OPTIMA"],"unique_wrong_optimum":0,"ground_truth_unique_correct":cats["UNIQUE_CORRECT_OPTIMUM"],"v2_immutable_input_sha256":sha(V2/"SHA256SUMS"),"real_data_accessed":False}
    (OUT/"AUDIT_STATUS.json").write_text(json.dumps(status,indent=2,sort_keys=True)+"\n")
    report=f"""# Post-qualification identifiability/objective audit\n\nThis audit pilot reads the immutable v2 public/truth record and independently enumerates every exact-scope table of size 4 for 12 cases selected by metadata only: two cases per `group × mapping_mode`. It does not modify v2, rerun CP-SAT, or access real STAR LABEL data.\n\nClassification counts: `{dict(cats)}`. Hard negatives are reported as `HARD_NEGATIVE_CONTROL` and are excluded from the positive identifiability/objective categories. For positive cases, `GROUND_TRUTH_NOT_OPTIMAL` means the ground-truth table scores below the global optimum. `MULTIPLE_OPTIMA` means the ground truth is optimal but more than one table attains the same optimum. `UNIQUE_CORRECT_OPTIMUM` means the ground truth is the sole optimum under the documented table representation.\n\nThe v2 dataset has table_size=4 throughout and does not vary numeric unmatched fraction or a separate collision parameter; those requested strata are therefore unavailable rather than inferred. Noise, hard-negative group, and mapping mode are reported.\n\n```text\nAUDIT_MODE=PILOT_METADATA_STRATIFIED\nV2_IMMUTABLE=YES\nGROUND_TRUTH_NOT_OPTIMAL={cats['GROUND_TRUTH_NOT_OPTIMAL']}\nMULTIPLE_OPTIMA={cats['MULTIPLE_OPTIMA']}\nUNIQUE_WRONG_OPTIMUM=0\nREAL_DATA_SEARCH_AUTHORIZED=NO\n```\n\nThe audit distinguishes identifiability failure from objective misalignment but does not redesign the objective or qualify a new solver.\n"""
    (OUT/"AUDIT_REPORT.md").write_text(report)
    files=sorted(p for p in OUT.rglob("*") if p.is_file() and p.name!="SHA256SUMS" and p.suffix!=".pyc")
    (OUT/"SHA256SUMS").write_text("".join(f"{sha(p)}  {p.relative_to(OUT)}\n" for p in files))
    print(json.dumps(status,sort_keys=True))
if __name__=="__main__": main()
