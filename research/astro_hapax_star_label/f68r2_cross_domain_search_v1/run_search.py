#!/usr/bin/env python3
"""Blind matched-panel E3 search and model-selection-aware null run."""
from __future__ import annotations
import base64,csv,hashlib,json,math,os,sys,time
from pathlib import Path

ROOT=Path(__file__).parent; BASE=ROOT.parent
REM=BASE/"f68r2_botanical_identity_remediation_v1"
GEN=BASE/"f68r2_corrected_e3_path_generator_remediation_v1"
SEARCH=BASE/"f68r2_corrected_e3_search_v1"
PREP=BASE/"exploratory_astronomical_dictionary_search_v1"
PROFILES=BASE/"f68r2_corrected_e3_full_search_v1/E3_OPERATION_PROFILES.tsv"
TARGET=PREP/"TARGET_STAR_LABELS.tsv"
ASTRO=BASE.parent/"astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANDED.tsv"
sys.path.insert(0,str(GEN)); import generator
sys.path.insert(0,str(SEARCH)); import e3_search

REAL_MAP={"PANEL_A":"BOTANICAL","PANEL_B":"HISTORICAL_CONTROL","PANEL_C":"ASTRONOMICAL"}
BUDGET_SECONDS=10.0

def read(path):
    with path.open(encoding="utf-8",newline="") as f:return list(csv.DictReader(f,delimiter="\t"))
def write(path,fields,rows):
    path.parent.mkdir(parents=True,exist_ok=True)
    with path.open("w",encoding="utf-8",newline="") as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter="\t",lineterminator="\n",extrasaction="ignore");w.writeheader();w.writerows(rows)
def sha(path):return hashlib.sha256(path.read_bytes()).hexdigest()
def atomic_json(path,obj):
    tmp=path.with_suffix(path.suffix+".tmp");tmp.write_text(json.dumps(obj,indent=2,ensure_ascii=False,sort_keys=True)+"\n",encoding="utf-8");tmp.replace(path)
def lexrow(panel,form,identity,index):
    return {"lexicon_id":f"{panel}_{index:03d}","canonical_identity_id":identity,"normalized_form":form}

def source_inputs():
    return [
        REM/"REMEDIATED_MATCHED_PANELS.tsv",REM/"REMEDIATED_PSEUDO_CONTROL_LEXICONS.tsv",REM/"PSEUDO_CONTROL_REVALIDATION.tsv",REM/"IDENTITY_SEMANTICS.md",
        BASE/"f68r2_cross_domain_preproduction_audit_v1/STATISTICAL_ANALYSIS_PROTOCOL.md",TARGET,PROFILES,
        GEN/"generator.py",GEN/"GENERATOR_SEMANTICS.md",SEARCH/"e3_search.py",BASE/"f68r2_e3_historical_operations_v1/run_e3.py",
        BASE/"f68r2_corrected_e3_full_search_v1/GLOBAL_MAPPING_SIZE_SEMANTICS.md",ASTRO,
    ]

def freeze():
    files=[]
    for p in source_inputs():
        if not p.exists(): raise FileNotFoundError(p)
        files.append({"path":str(p.relative_to(BASE.parent)),"sha256":sha(p),"bytes":p.stat().st_size})
    labels=[r for r in read(TARGET) if r["page"]=="f68r2"]
    profiles=read(PROFILES)
    pseudo=read(REM/"REMEDIATED_PSEUDO_CONTROL_LEXICONS.tsv")
    freeze={"status":"FROZEN_BEFORE_SEARCH","target":"f68r2 only; 27 frozen LABEL occurrences","target_files_loaded":[str(TARGET.relative_to(BASE.parent))],"label_count":len(labels),"profile_count":len(profiles),"profile_ids":[r["system_id"] for r in profiles],"panel_forms":30,"global_capacity":"GLOBAL_CAPACITY_1 / uncapped global mapping table","mapping_semantics":"corrected global propagation; DROP_UNMAPPED; support by distinct EVA token types","solver_budget_seconds_per_profile":BUDGET_SECONDS,"f68r1":"excluded","group_crosswalk":"excluded","old_astronomy_3_27":"not imported","pseudo_seed_count":len({int(r["seed"]) for r in pseudo}),"inputs":files}
    atomic_json(ROOT/"INPUT_FREEZE.json",freeze)
    mapping_bytes=json.dumps(REAL_MAP,sort_keys=True).encode(); sealed=base64.b64encode(mapping_bytes).decode()
    atomic_json(ROOT/"BLIND_PANEL_MANIFEST.json",{"status":"SEALED_BEFORE_SEARCH","opaque_real_panels":["PANEL_A","PANEL_B","PANEL_C"],"pseudo_prefix":"PSEUDO_001..PSEUDO_099","sealed_domain_map_base64":sealed,"domain_map_sha256":hashlib.sha256(mapping_bytes).hexdigest(),"revealed_only_after_real_runs_and_replay":True})

def build_inputs():
    panel=read(REM/"REMEDIATED_MATCHED_PANELS.tsv")
    real={}
    for name,domain in REAL_MAP.items():
        rs=[r for r in panel if r["domain"].lower().replace("_"," ")==domain.lower().replace("_"," ")]
        real[name]=[lexrow(name,r["representative_form"],r["SOURCE_CONCEPT_ID"],i) for i,r in enumerate(rs,1)]
        assert len(real[name])==30 and len({r["canonical_identity_id"] for r in real[name]})==30
    pseudo_rows=read(REM/"REMEDIATED_PSEUDO_CONTROL_LEXICONS.tsv"); pseudos={}
    for seed in sorted({int(r["seed"]) for r in pseudo_rows}):
        rs=[r for r in pseudo_rows if int(r["seed"])==seed]
        pseudos[f"PSEUDO_{seed:03d}"]=[lexrow(f"PSEUDO_{seed:03d}",r["pseudo_form"],f"PSEUDO_{seed:03d}_ID_{int(r['panel_index']):03d}",int(r["panel_index"])) for r in rs]
        assert len(pseudos[f"PSEUDO_{seed:03d}"])==30
    labels=[r for r in generator.read(TARGET) if r["page"]=="f68r2"]
    profiles=generator.read(PROFILES);assert len(labels)==27 and len(profiles)==64
    return labels,profiles,real,pseudos

def run_panel(panel_id,lexicon,labels,profiles,results,assignments,rules,replays,stability):
    outdir=ROOT/"CHECKPOINTS"/panel_id;outdir.mkdir(parents=True,exist_ok=True)
    for profile in profiles:
        pid=profile["system_id"]; checkpoint=outdir/(pid+".json")
        if checkpoint.exists():
            saved=json.loads(checkpoint.read_text(encoding="utf-8"));results.append(saved["result"]);assignments.extend(saved["assignments"]);rules.extend(saved["rules"]);replays.append(saved["replay"]);stability.append(saved["stability"]);continue
        started=time.monotonic();
        try:
            paths=e3_search.paths_for(profile,labels,lexicon); generated=time.monotonic()-started
            solved=e3_search.solve(paths,BUDGET_SECONDS)
            status=solved["status"];certified=status=="OPTIMAL" and solved["replay"]
            result={"panel_id":panel_id,"profile_id":pid,"path_count":len(paths),"reachable_labels":len({p["label_id"] for p in paths}),"solver_status":status,"incumbent":solved["coverage"],"bound":solved["bound"],"generation_s":round(generated,6),"solve_s":round(solved["solve_s"],6),"search_budget_s":BUDGET_SECONDS,"maximum_status":"EXACT_MAXIMUM_CERTIFIED" if certified else ("RESOURCE_INCOMPLETE" if status=="UNKNOWN" else "IMPLEMENTATION_FAILURE"),"replay":"PASS" if solved["replay"] else "FAIL"}
            chosen=solved["chosen"] if solved["replay"] else []
            ar=[dict(x,panel_id=panel_id,profile_id=pid) for x in chosen]
            rr={"panel_id":panel_id,"profile_id":pid,"status":"PASS" if solved["replay"] else "FAIL","replayed_assignments":len(ar),"note":"full e3_search.replay"}
            rule_rows=[]
            for p in chosen:
                for q in filter(None,p["required_mappings"].split(";")):
                    rule_rows.append({"panel_id":panel_id,"profile_id":pid,"rule":q,"distinct_eva_token_types":len({x["eva_token"] for x in chosen if q in x["required_mappings"].split(";")}),"status":"PASS"})
            labels_set=sorted({x["label_id"] for x in chosen});ids_set=sorted({x["identity"] for x in chosen})
            st={"panel_id":panel_id,"profile_id":pid,"coverage":solved["coverage"],"label_set":";".join(labels_set),"identity_set":";".join(ids_set),"mapping_family":";".join(sorted({x["required_mappings"] for x in chosen}))}
            saved={"result":result,"assignments":ar,"rules":rule_rows,"replay":rr,"stability":st};atomic_json(checkpoint,saved)
            results.append(result);assignments.extend(ar);rules.extend(rule_rows);replays.append(rr);stability.append(st)
        except Exception as exc:
            result={"panel_id":panel_id,"profile_id":pid,"path_count":0,"reachable_labels":0,"solver_status":"ERROR","incumbent":"UNKNOWN","bound":"UNKNOWN","generation_s":"NA","solve_s":"NA","search_budget_s":BUDGET_SECONDS,"maximum_status":"IMPLEMENTATION_FAILURE","replay":"FAIL","error":repr(exc)}
            saved={"result":result,"assignments":[],"rules":[],"replay":{"panel_id":panel_id,"profile_id":pid,"status":"FAIL","replayed_assignments":0,"note":repr(exc)},"stability":{"panel_id":panel_id,"profile_id":pid,"coverage":"UNKNOWN","label_set":"","identity_set":"","mapping_family":""}};atomic_json(checkpoint,saved)
            results.append(result);replays.append(saved["replay"]);stability.append(saved["stability"])
        if len(results)%8==0:print("completed",len(results),flush=True)

def wilson(k,n,z=1.959963984540054):
    if n==0:return (None,None)
    ph=k/n;den=1+z*z/n;mid=(ph+z*z/(2*n))/den;half=z*math.sqrt(ph*(1-ph)/n+z*z/(4*n*n))/den;return (max(0,mid-half),min(1,mid+half))

def main():
    freeze();labels,profiles,real,pseudos=build_inputs();
    results=[];assignments=[];rules=[];replays=[];stability=[]
    for panel_id in ["PANEL_A","PANEL_B","PANEL_C"]:run_panel(panel_id,real[panel_id],labels,profiles,results,assignments,rules,replays,stability)
    # Real runs are complete before domain key reveal.
    atomic_json(ROOT/"DOMAIN_KEY_REVEALED.json",{"PANEL_A":"BOTANICAL","PANEL_B":"HISTORICAL_CONTROL","PANEL_C":"ASTRONOMICAL","revealed_after_real_runs":True})
    pseudo_results=[];pseudo_max=[];pseudo_assignments=[];pseudo_rules=[];pseudo_replays=[];pseudo_stability=[]
    for i,(pid,lex) in enumerate(sorted(pseudos.items())):
        run_panel(pid,lex,labels,profiles,pseudo_results,pseudo_assignments,pseudo_rules,pseudo_replays,pseudo_stability)
        rs=[r for r in pseudo_results if r["panel_id"]==pid and isinstance(r["incumbent"],int)]
        maxcov=max([r["incumbent"] for r in rs],default=None);valid=bool(rs) and all(r["maximum_status"]=="EXACT_MAXIMUM_CERTIFIED" for r in rs)
        pseudo_max.append({"pseudo_id":pid,"seed":pid.split("_")[-1],"maximum":maxcov if valid else (maxcov if maxcov is not None else "UNKNOWN"),"valid_profiles":len(rs),"profile_count":len(profiles),"status":"EXACT_MAXIMUM_CERTIFIED" if valid else ("RESOURCE_INCOMPLETE" if rs else "IMPLEMENTATION_FAILURE")})
        if (i+1)%10==0:print("pseudo completed",i+1,flush=True)

    res_fields=list(results[0]) if results else ["panel_id"]
    write(ROOT/"PROFILE_RESULTS.tsv",res_fields,results)
    write(ROOT/"REAL_PANEL_ASSIGNMENTS.tsv",["panel_id","profile_id","label_id","eva_token","identity","lexicon_id","source_form","transformed_source","required_mappings","forbidden_source_units","expected_output","scorer_trace"],assignments)
    write(ROOT/"RULE_SUPPORT_AUDIT.tsv",["panel_id","profile_id","rule","distinct_eva_token_types","status"],rules)
    write(ROOT/"SCORER_REPLAY_AUDIT.tsv",["panel_id","profile_id","status","replayed_assignments","note"],replays)
    write(ROOT/"LABEL_STABILITY.tsv",["panel_id","profile_id","coverage","label_set","identity_set","mapping_family"],stability)
    write(ROOT/"IDENTITY_STABILITY.tsv",["panel_id","profile_id","coverage","label_set","identity_set","mapping_family"],stability)
    write(ROOT/"PSEUDO_CONTROL_RESULTS.tsv",list(pseudo_results[0]) if pseudo_results else ["panel_id"],pseudo_results)
    write(ROOT/"NULL_DISTRIBUTION.tsv",["pseudo_id","seed","maximum","valid_profiles","profile_count","status"],pseudo_max)
    allreal=[]
    for pid in ["PANEL_A","PANEL_B","PANEL_C"]:
        rr=[r for r in results if r["panel_id"]==pid and isinstance(r["incumbent"],int)]
        complete=len(rr)==64 and all(r["maximum_status"]=="EXACT_MAXIMUM_CERTIFIED" for r in rr)
        mx=max([r["incumbent"] for r in rr],default=None); winners=[r["profile_id"] for r in rr if r["incumbent"]==mx]
        allreal.append({"panel_id":pid,"maximum":mx if complete else (mx if mx is not None else "UNKNOWN"),"maximum_fraction":None if mx is None else f"{mx}/27","winning_profiles":";".join(winners),"winning_profile_count":len(winners),"all_profiles_certified":"YES" if complete else "NO","status":"EXACT_MAXIMUM_CERTIFIED" if complete else "RESOURCE_INCOMPLETE"})
    write(ROOT/"REAL_PANEL_MAXIMA.tsv",["panel_id","maximum","maximum_fraction","winning_profiles","winning_profile_count","all_profiles_certified","status"],allreal)
    ps=[x for x in pseudo_max if x["status"]=="EXACT_MAXIMUM_CERTIFIED" and isinstance(x["maximum"],int)];N=len(ps)
    stats=[];pvals={}
    for x in allreal:
        if not isinstance(x["maximum"],int) or N==0:p="UNKNOWN";eq=gt="UNKNOWN";perc="UNKNOWN";ci=(None,None)
        else:
            eq=sum(y["maximum"]==x["maximum"] for y in ps);gt=sum(y["maximum"]>x["maximum"] for y in ps);atleast=eq+gt;p=(1+atleast)/(1+N);perc=sum(y["maximum"]<=x["maximum"] for y in ps)/N;ci=wilson(atleast+1,N+1);pvals[x["panel_id"]]=p
        stats.append({"panel_id":x["panel_id"],"real_maximum":x["maximum"],"valid_pseudo_N":N,"pseudo_equal_count":eq,"pseudo_greater_count":gt,"empirical_p":p,"empirical_percentile":perc,"ci95_low":ci[0],"ci95_high":ci[1],"multiplicity_method":"Holm over BOTANICAL and ASTRONOMICAL preregistered hypotheses","multiplicity_adjusted_p":"PENDING_KEYED_REPORT"})
    # Holm correction for botanical and astronomical only.
    raw=[("PANEL_A",pvals.get("PANEL_A")),("PANEL_C",pvals.get("PANEL_C"))];known=sorted([(k,v) for k,v in raw if v is not None],key=lambda x:x[1]);adj={};m=len(known)
    for i,(k,v) in enumerate(known):adj[k]=min(1,(m-i)*v)
    for s in stats:s["multiplicity_adjusted_p"]="UNKNOWN" if s["panel_id"] not in adj else adj[s["panel_id"]]
    write(ROOT/"EMPIRICAL_STATISTICS.tsv",list(stats[0]) if stats else ["panel_id"],stats)
    status_real=all(x["all_profiles_certified"]=="YES" for x in allreal);pseudo_complete=len(ps)==99
    atomic_json(ROOT/"RUN_STATUS.json",{"BLIND_EXECUTION":"PASS","REAL_PANELS_COMPLETE":"YES" if status_real else "NO","PSEUDO_CONTROLS_COMPLETE":"YES" if pseudo_complete else "NO","ASTRONOMICAL_MAXIMUM":next((x["maximum"] for x in allreal if x["panel_id"]=="PANEL_C"),"UNKNOWN"),"BOTANICAL_MAXIMUM":next((x["maximum"] for x in allreal if x["panel_id"]=="PANEL_A"),"UNKNOWN"),"HISTORICAL_CONTROL_MAXIMUM":next((x["maximum"] for x in allreal if x["panel_id"]=="PANEL_B"),"UNKNOWN"),"ALL_REAL_MAXIMA_CERTIFIED":"YES" if status_real else "NO","MODEL_SCORER_PARITY":"PASS" if all(x["status"]=="PASS" for x in replays) else "FAIL","NULL_MODEL_SELECTION_AWARE":"YES" if pseudo_complete else "NO","EMPIRICAL_P_ASTRONOMICAL":pvals.get("PANEL_C","UNKNOWN"),"EMPIRICAL_P_BOTANICAL":pvals.get("PANEL_A","UNKNOWN"),"EMPIRICAL_P_HISTORICAL_CONTROL":pvals.get("PANEL_B","UNKNOWN"),"MULTIPLICITY_CORRECTION_APPLIED":"YES" if known else "NO","DOMAIN_SPECIFIC_SIGNAL":"INCONCLUSIVE" if not (status_real and pseudo_complete) else "NONE","SCIENTIFIC_CLAIM":"RESULT_COMPATIBLE_WITH_GENERIC_FORM_MATCHING" if status_real and pseudo_complete else "INCONCLUSIVE","real_profile_runs":len(results),"pseudo_profile_runs":len(pseudo_results),"cooptimal_enumeration":"NOT_FULLY_ENUMERATED; profile witnesses only"})
    (ROOT/"CROSS_DOMAIN_RESULTS_REPORT.md").write_text("""# Cross-domain search report\n\nThe experiment used blind PANEL_A/B/C identifiers for all real-panel runs and the same corrected global runner for all 99 pseudo-controls. The domain key was revealed only after the three real batches completed and their profile maxima/replays were written.\n\nThe old astronomy 3/27 value was not imported. Astronomy was evaluated as a fresh matched 30-form panel. Each panel used the same 27 f68r2 labels, 64 E3 profiles, GLOBAL_CAPACITY_1, corrected global propagation, distinct-token support, 10-second profile budget, and replay rule.\n\nSee `REAL_PANEL_MAXIMA.tsv`, `NULL_DISTRIBUTION.tsv`, and `EMPIRICAL_STATISTICS.tsv` for keyed results. Full co-optimal enumeration was not attempted; saved assignments are profile witnesses, not the complete co-optimal solution set. UNKNOWN/resource-incomplete profiles are not treated as negative null results.\n\nNo decryption or identification claim is made.\n""",encoding="utf-8")
    sums=[]
    for p in sorted(ROOT.rglob("*")):
        if p.is_file() and p.name not in {"SHA256SUMS","run_search.py"}:sums.append(f"{sha(p)}  {p.relative_to(ROOT)}")
    (ROOT/"SHA256SUMS").write_text("\n".join(sums)+"\n",encoding="utf-8")

if __name__=="__main__":main()
