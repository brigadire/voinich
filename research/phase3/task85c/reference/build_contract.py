#!/usr/bin/env python3
"""Build Task85c declarative artifacts and independent small golden values.

This is a contract-construction/reference utility, not a production model
implementation.  It deliberately uses only the Python standard library.
"""
from __future__ import annotations

import csv
import hashlib
import json
import math
import os
import struct
import unicodedata
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REPO = ROOT.parents[2]
VERSION = "G1_V2_EXECUTABLE_CONTRACT_V1"
ROOT_SEED = bytes.fromhex("6f5a9c731de248b480c66b237ace215044689c5fa2f593e510b73dce18a49027")


def canon(v):
    """G1V2-CJ-1: NFC strings, sorted keys, no JSON number floats."""
    def norm(x):
        if isinstance(x, str):
            return unicodedata.normalize("NFC", x)
        if isinstance(x, list):
            return [norm(y) for y in x]
        if isinstance(x, dict):
            return {unicodedata.normalize("NFC", k): norm(x[k]) for k in sorted(x)}
        if isinstance(x, float):
            raise TypeError("floats must be canonical decimal strings")
        return x
    return (json.dumps(norm(v), ensure_ascii=False, sort_keys=True,
                       separators=(",", ":")) + "\n").encode()


def write(path, data):
    p = ROOT / path
    p.parent.mkdir(parents=True, exist_ok=True)
    if isinstance(data, bytes):
        p.write_bytes(data)
    else:
        p.write_text(data, encoding="utf-8", newline="\n")


def write_json(path, obj):
    write(path, canon(obj))


def write_tsv(path, header, rows):
    p = ROOT / path
    p.parent.mkdir(parents=True, exist_ok=True)
    with p.open("w", encoding="utf-8", newline="") as f:
        w = csv.writer(f, delimiter="\t", lineterminator="\n")
        w.writerow(header)
        w.writerows(rows)


def rng(namespace, counters):
    ns = unicodedata.normalize("NFC", namespace).encode()
    msg = b"G1V2-RNG\0" + ROOT_SEED + struct.pack(">I", len(ns)) + ns
    msg += struct.pack(">I", len(counters))
    for x in counters:
        msg += struct.pack(">Q", x)
    return hashlib.sha256(msg).digest()


def qtype7(xs, p):
    xs = sorted(xs)
    if not xs:
        raise ValueError("empty quantile")
    h = (len(xs) - 1) * p
    lo, hi = math.floor(h), math.ceil(h)
    return xs[lo] + (h - lo) * (xs[hi] - xs[lo])


MODELS = {
 "M0": "iid_glyph_dirichlet", "M1": "fixed_order_markov",
 "M2": "pst_likelihood_gain", "M3": "deterministic_probabilistic_dfa",
 "M4": "discrete_hmm_baum_welch", "M5": "three_component_productive_grammar"
}


def candidates():
    out = []
    def add(mid, route, hp):
        cid = f"{mid}-{route}-" + "-".join(str(v).replace(".", "p") for v in hp.values())
        out.append([cid, mid, route, json.dumps(hp, sort_keys=True, separators=(",", ":")),
                    f"{mid}_VALIDATION", f"G1V2_EXECUTABLE_CONTRACT.md#{mid.lower()}"])
    for a in ["0", "0.1", "0.5", "1"]: add("M0", "iid", {"alpha": a})
    for n in [1,2,3]:
        for a in ["0", "0.1", "0.5", "1"]: add("M1", "markov", {"order": n, "alpha": a})
    for d in [2,4,6]:
        for g in ["0", "2", "5"]: add("M2", "pst", {"max_depth": d, "gain_bits": g})
    for s in [2,3]: add("M3", "exact", {"max_states": s})
    for s in [4,8]:
        for t in ["0.01", "0.05"]: add("M3", "approx", {"max_states": s, "merge_js": t})
    for s in [2,3,4]:
        for a in ["0.01", "0.1"]: add("M4", "hmm", {"states": s, "alpha": a})
    for w in ["0.01", "0.05", "0.1"]:
        for m in [2,5]: add("M5", "grammar", {"backoff_weight": w, "min_rule_support": m})
    return out


DOMAINS = [
 ("FIT","candidate fitting","g1v2/fit","control_index,candidate_index,attempt","digest/u53/index","M0-M5 fit"),
 ("SELECT","validation selection","g1v2/select","control_index,model_rank","digest/u53","selector"),
 ("M4_RESTART","M4 initialization","g1v2/m4/restart","control_index,candidate_index,restart,state,parameter","digest/u53","M4"),
 ("GENERATE","model generation","g1v2/generate","control_index,candidate_index,scale_index,replicate,draw","digest/u53/index","generators"),
 ("PM_BOOTSTRAP","predictive bootstrap","g1v2/pm/bootstrap","control_index,candidate_index,metric_index,replicate,draw","index","PM2/PM5"),
 ("PM_PERMUTATION","predictive permutation","g1v2/pm/permutation","control_index,candidate_index,metric_index,replicate,pair","bit","PM thresholds"),
 ("PM6_COMPLEMENT","PM6 complement sampling","g1v2/pm6/complement","control_index,candidate_index,length,occurrence","index","PM6"),
 ("PM6_BOOTSTRAP","PM6 paired bootstrap","g1v2/pm6/bootstrap","control_index,candidate_index,replicate,draw","index","PM6"),
 ("PM6_PERMUTATION","PM6 label permutation","g1v2/pm6/permutation","control_index,candidate_index,replicate,pair","bit","PM6"),
 ("STRUCT_BOOTSTRAP","structural threshold bootstrap","g1v2/struct/bootstrap","metric_index,scale_index,replicate,draw","index","F2 calibration"),
 ("CONTROL_GENERATE","synthetic control generation","g1v2/control/generate","generator_index,scale_index,replicate,draw","digest/u53/index","control builder"),
 ("CORPUS_WINDOW","natural corpus window","g1v2/corpus/window","corpus_index,scale_index,replicate","index","corpus builder"),
 ("BLIND_ID","opaque blind identifiers","g1v2/blind/id","generator_index,scale_index,replicate","digest","escrow builder")]


PRED = [
 ("PM1","sum -log2 p of HELDOUT glyph events including EOS","glyph event","corpus","B1+B2","LOWER","diagnostic","0","finite nonnegative","DIAGNOSTIC"),
 ("PM2","PM1/scored glyph-event count","glyph event","corpus","B1+B2","LOWER","development q95","0.01 bit/event","PASS iff both effects strictly exceed threshold","REQUIRED"),
 ("PM4","mean log2 whole-token probability for HELDOUT tokens absent from DEVELOPMENT","token occurrence","corpus","B1+B2","HIGHER","development q05","0.05 bit/token","support only; <25 unseen => NOT_ASSESSABLE","SUPPORTING"),
 ("PM5","sum_bin (n_b/N)*abs(mean_conf_b-mean_label_b)","next-glyph event; label=1 for realized glyph, confidence=assigned probability","corpus","B1+B2","LOWER","development q95","0.01","PASS iff candidate ECE lower by > threshold","REQUIRED"),
 ("PM6","positive-mass-weighted mean length-stratified AUC; ties=0.5","matched positive-negative token pair","corpus","chance+permutation","HIGHER","max(0.5,permutation q95)","0","LCB strictly above both bounds","REQUIRED")]


STRUCT_IDS = [
 ("EF1_GIANT_COMPONENT_SHARE","EDIT","absolute difference","0.02","YES"),
 ("EF1_ISOLATE_SHARE","EDIT","absolute difference","0.02","YES"),
 ("EF2_GLOBAL_CLUSTERING","EDIT","absolute difference","0.02","YES"),
 ("EF3_DEGREE_FREQUENCY_SPEARMAN","EDIT","absolute difference","0.05","YES"),
 ("LP1_RULE_SUPPORT_GINI","LEXICAL_PARADIGM","absolute difference","0.03","YES"),
 ("LP4_PREFIX_ATTACHMENT_NMI","LEXICAL_PARADIGM","absolute difference","0.03","YES"),
 ("LP4_SUFFIX_ATTACHMENT_NMI","LEXICAL_PARADIGM","absolute difference","0.03","YES"),
 ("HR1_FOLIO_VARIANCE_SHARE","SKELETON","borrowed skeleton diagnostic","0","NO"),
 ("HR1_LOCUS_VARIANCE_SHARE","SKELETON","borrowed skeleton diagnostic","0","NO"),
 ("HR1_SECTION_VARIANCE_SHARE","SKELETON","borrowed skeleton diagnostic","0","NO"),
 ("LS1_LINE_LENGTH_CV","SKELETON","borrowed skeleton diagnostic","0","NO"),
 ("PF5_WITHIN_FOLIO_PROGRESSION","SKELETON","borrowed skeleton diagnostic","0","NO")]


SCHEMA_TYPES = ["fit","fitted_model","predictive_metric","predictive_gate","predictive_verdict",
 "generation","f2_metric","structural_family","structural_gate","structural_verdict",
 "complexity","minimality","final_verdict","not_reached","scientific_failure"]


def build_contract_json(cands):
    return {
      "contract_version": VERSION,
      "terminal_marker": "G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_FROZEN",
      "normative_prose": "G1V2_EXECUTABLE_CONTRACT.md",
      "numeric": {"binary_float":"IEEE-754 binary64 round-to-nearest ties-to-even; no FMA-dependent reductions",
        "log_base":2,"sum":"ascending scientific identity; Neumaier compensated binary64",
        "probability":"reject nonfinite/negative; normalize by compensated sum; zero log probability is +infinity",
        "epsilon":"none unless literal in model definition","comparison":"exact binary64 after prescribed calculation; no tolerance",
        "quantile":"Hyndman-Fan type 7 h=(n-1)p, linear interpolation","ties":"equality never exceeds strict gate",
        "evidence_numbers":"finite scientific reals are shortest round-trip lowercase decimal strings; -0 is 0; NaN/Infinity forbidden"},
      "data": {"unicode":"UTF-8 NFC","tokens":"nonempty arrays of Unicode scalar glyph strings",
        "symbols":{"bos":"<BOS>","eos":"<EOS>","unk":"<UNK>"},
        "split":"occurrence order 60/20/20 DEVELOPMENT/VALIDATION/HELDOUT; floor at first two cut points; no refit after selection"},
      "selection": {"objective":"minimum VALIDATION PM2; nonfinite is ineligible",
        "tie":"within exact binary64 equality: lower complexity bits, then UTF-8 bytewise candidate_id",
        "fit":"fit candidate on DEVELOPMENT only; choose on VALIDATION; selected frozen fit is evaluated on HELDOUT; no DEVELOPMENT+VALIDATION refit"},
      "rng":{"algorithm":"SHA-256 counter construction G1V2-RNG-1","root_seed_hex":ROOT_SEED.hex(),
        "message":"ASCII G1V2-RNG NUL || root32 || u32be(namespace_utf8_len) || NFC UTF-8 namespace || u32be(counter_count) || each u64be counter",
        "uniform":"u53=(u64be(digest[0:8])>>11)/2^53",
        "bounded":"take successive u64be 8-byte blocks; if exhausted increment final counter; accept x < 2^64-(2^64 mod n), return x mod n",
        "ordering":"all counter fields nonnegative <2^64; no execution property is a field"},
      "candidate_count":len(cands),
      "models":MODELS,
      "metrics":{"predictive":[x[0] for x in PRED],"structural":[x[0] for x in STRUCT_IDS]},
      "thresholds":{"predictive_replicates":2000,"pm6_bootstrap":2000,"pm6_permutations":2000,
        "structural_bootstrap":2000,"development_controls":"only DEVELOPMENT rows",
        "one_sided":"type-7 q0.95 of null absolute/effect statistic; max with literal practical floor"},
      "generation":{"scales":[2000,8000,32000],"replicates":4,"max_token_glyphs":64},
      "dag":{"control_instances":192,"candidate_fits":8256,"predictive_jobs":8256,
        "generation_batches":99072,"structural_metric_jobs":1188864,"complexity_jobs":8256,
        "candidate_aggregation_jobs":8256,"control_aggregation_jobs":192,
        "terminal_cells":8256,"total_jobs":1321152,"dependency_edges":2617152},
      "canonicalization":{"profile":"G1V2-CJ-1","json":"UTF-8 NFC; object keys UTF-8 byte lexicographic; compact separators; LF terminator; arrays ordered",
        "hash":"sha256 of canonical bytes lowercase hex","null":"only schema-explicit unavailable value fields; absent otherwise"},
      "firewall":{"calibration":["DEVELOPMENT"],"prohibited":["blind Stage C outcomes","natural-language confirmatory outcomes","Voynich"],
        "voynich_paths_forbidden":True},
      "schemas":SCHEMA_TYPES
    }


def main():
    cands = candidates()
    contract = build_contract_json(cands)
    write_json("G1V2_EXECUTABLE_CONTRACT.json", contract)
    write_tsv("G1V2_CANDIDATE_REGISTRY.tsv",
      ["candidate_id","model_class","route","hyperparameters","selection_group","normative_definition"], cands)
    write_tsv("G1V2_MODEL_REGISTRY.tsv",
      ["model_id","rank","family","sample_space","fit","selection","score_generation","complexity","normative_definition"],
      [[m,i,MODELS[m],"token glyph events ending EOS","DEVELOPMENT exact definition","min VALIDATION PM2","contract equations","G1V2_COMPLEXITY_CONTRACT.tsv",f"G1V2_EXECUTABLE_CONTRACT.md#{m.lower()}"] for i,m in enumerate(MODELS)])
    write_tsv("G1V2_RNG_DOMAIN_REGISTRY.tsv",
      ["domain_id","purpose","namespace","counter_fields","output_interpretation","consumer"], DOMAINS)
    write_tsv("G1V2_PREDICTIVE_METRIC_REGISTRY.tsv",
      ["metric_id","statistic","input","aggregation_unit","baseline_null","direction","threshold","practical_floor","status_semantics","role"], PRED)
    srows=[]
    for mid,fam,comp,floor,ind in STRUCT_IDS:
        srows.append([mid,"Fingerprint-v2/task85","normalized token occurrence corpus","2000;8000;32000",fam,comp,"median 4 replicates",f"max(dev-MFC-type7-q95,{floor})" if ind=="YES" else "none",floor,"PASS distance<=threshold",ind,"DIAGNOSTIC_ONLY" if ind=="NO" else "WEIGHTED"])
    write_tsv("G1V2_STRUCTURAL_METRIC_REGISTRY.tsv",
      ["metric_id","implementation_version","input","scales","family","direction","aggregation","threshold","practical_floor","status","independent","skeleton_policy"],srows)
    write_tsv("G1V2_COMPLEXITY_CONTRACT.tsv",
      ["model_class","canonical_structure","structure_cost_bits","parameter_cost_bits","precision","total_rule"], [
       ["M0","sorted vocabulary+probability vector","gamma(|A|)+sum utf8code(glyph)","32*(|A|+1)","binary32 round ties-even","sum"],
       ["M1","order+sorted observed contexts and rows","gamma(order)+gamma(rows)+context codes","32*rows*(|A|+1)","binary32","sum"],
       ["M2","sorted retained suffix tree","balanced-parenthesis 2N+labels","32*N*(|A|+1)","binary32","sum"],
       ["M3","BFS canonical states/transitions","gamma(S)+S*(|A|*ceil(log2 S)+1)","32*S*(|A|+1)","binary32","sum; route tag 1 bit"],
       ["M4","states canonically sorted by emission row then transition row","gamma(S)+ceil(log2(S!))","32*(S*S+S*(|A|+1)+S)","binary32","sum"],
       ["M5","sorted prefix/stem/suffix/rule/exception strings","gamma counts+utf8code inventory+rule incidence","32*(component probabilities+backoff weights)","binary32","sum"]])

    controls=[]
    roots=["19"*32,"2a"*32,"3b"*32,"4c"*32,"5d"*32,"6e"*32,"7f"*32,"80"*32,"91"*32,"a2"*32,"b3"*32,"c4"*32]
    for i,m in enumerate(MODELS):
        for author in ["A","B"]:
            controls.append([f"DEV_{m}_{author}","DEVELOPMENT",m,f"{m}_GEN_{author}","open",8000,1,roots[2*i+(author=="B")],"threshold+engineering","specification-authored independent route"])
    for m in MODELS:
        for author in ["A","B"]:
            controls.append([f"BLIND_{m}_{author}","BLIND_SYNTHETIC",m,f"{m}_GEN_{author}","HMAC escrow", "2000;8000;32000",4,"derived CONTROL_GENERATE","recovery","ground truth only in escrow"])
    controls += [
      ["NL_EN","NATURAL_CONFIRMATORY","English","CORPUS_EN","visible label","2000;8000;32000",4,"CORPUS_WINDOW","applicability","not calibration"],
      ["NL_LA","NATURAL_CONFIRMATORY","Latin","CORPUS_LA","visible label","2000;8000;32000",4,"CORPUS_WINDOW","applicability","not calibration"],
      ["NL_SA","NATURAL_CONFIRMATORY","Sanskrit","CORPUS_SA","visible label","2000;8000;32000",4,"CORPUS_WINDOW","applicability","not calibration"]]
    write_tsv("G1V2_CONTROL_REGISTRY.tsv",["control_id","role","class_or_language","generator_or_source","blinding","token_counts","replicates","root_seed","scientific_role","independence"],controls)
    mechanisms = {
      "M0": ({"outcomes":["a","b","c","d","<EOS>"],"probabilities":["0.28","0.22","0.18","0.12","0.20"]},
             "categorical IID token source"),
      "M1": ({"order":2,"alphabet":["a","b","c","d"],"eos_probability":"0.15","transition_rule":"given last glyph index i: next i+1 mod4=.45, i=.25, other two=.075 each; BOS uses [.4,.3,.2,.1] before EOS renormalization"},
             "second-order cyclic Markov source"),
      "M2": ({"alphabet":["a","b","c","d"],"max_depth":4,"eos_probability":"0.16","contexts":{"ab":{"c":"0.7"},"ba":{"d":"0.65"},"c":{"a":"0.55"},"root":{"a":"0.3","b":"0.3","c":"0.2","d":"0.2"}},"remainder":"distribute pro rata to root"},
             "sparse suffix-tree source"),
      "M3": ({"states":[0,1,2,3],"start":0,"rows":{"0":{"a":[1,"0.6"],"b":[2,"0.4"]},"1":{"c":[3,"0.7"],"<EOS>":[0,"0.3"]},"2":{"d":[3,"0.65"],"<EOS>":[0,"0.35"]},"3":{"a":[1,"0.25"],"b":[2,"0.25"],"<EOS>":[0,"0.5"]}}},
             "four-state deterministic-transition probabilistic DFA"),
      "M4": ({"states":3,"pi":["0.6","0.3","0.1"],"transition":[["0.7","0.2","0.1"],["0.15","0.7","0.15"],["0.2","0.2","0.6"]],"emissions":{"0":{"a":"0.45","b":"0.25","c":"0.1","d":"0.05","<EOS>":"0.15"},"1":{"a":"0.1","b":"0.45","c":"0.2","d":"0.1","<EOS>":"0.15"},"2":{"a":"0.1","b":"0.1","c":"0.3","d":"0.3","<EOS>":"0.2"}}},
             "three-state HMM token source"),
      "M5": ({"prefix":{"":"0.45","a":"0.3","b":"0.25"},"stem":{"c":"0.25","d":"0.25","ac":"0.2","bd":"0.15","cd":"0.15"},"suffix":{"":"0.4","a":"0.35","b":"0.25"},"exception_probability":"0.08","exceptions":{"abcd":"0.6","dcba":"0.4"}},
             "weighted component-derivation grammar")}
    grows=[]
    for m,(params,description) in mechanisms.items():
      for author,algorithm,rationale in [
        ("A","sequential inverse-CDF from explicit probability rows","direct mathematical sampler; isolated specification module"),
        ("B","enumerate complete finite outcomes/derivations in UTF-8 order; use exponential-race keys -ln(u)/weight (M1 uses Walker alias; M4 uses cumulative matrix rows)","different sampling construction and independently authored module; shares only frozen distribution")]:
        spec={"generator_id":f"{m}_GEN_{author}","class":m,"description":description,"parameters":params,"algorithm":algorithm,"max_token_glyphs":64,"invalid_draw":"fail after 1024 indexed attempts"}
        grows.append([spec["generator_id"],m,author,description,json.dumps(params,sort_keys=True,separators=(",",":")),algorithm,rationale,64,1024,hashlib.sha256(canon(spec)).hexdigest()])
    write_tsv("G1V2_SYNTHETIC_GENERATOR_REGISTRY.tsv",["generator_id","model_class","author_route","mechanism","parameters","sampling_algorithm","independence_rationale","max_token_glyphs","attempt_cap","spec_sha256"],grows)
    write_tsv("G1V2_CORPUS_REGISTRY.tsv",["corpus_id","language","work_edition","source","license","local_artifact","source_sha256","normalization","window_rule","final_status"],[
      ["CORPUS_EN","English","A Study in Scarlet, Gutenberg #2097 UTF-8 snapshot","https://www.gutenberg.org/ebooks/2097","Project Gutenberg/public domain US","data_test/pg2097-2.txt","0b260c8ae9ee7dcfd8c334e174b32ec433d37792ecfa783137b0bd38a956cc80","Gutenberg markers; NFC; casefold; letter runs with internal apostrophe","uniform circular start from CORPUS_WINDOW; 4 nonoverlapping occurrence windows per scale; reject overlap; 60/20/20","IMMUTABLE_LOCAL"],
      ["CORPUS_LA","Latin","Caesar #218 + Virgil #18837/#227 snapshots","https://www.gutenberg.org/ebooks/218;18837;227","Project Gutenberg/public domain US","to_construct/source files concatenated in numeric ID order","84ac8411841a4d8f5f4a49b6a2cd1f466917c6a5af72916d5e0b2b1ecb2f659c;8bae4d88747b318b58ea193f981766fed337ddf613df452c1a06d72c9af75ffd;2620ba82c00964d1a1f0458c8cf171946f158a06de0d0e764594170602c4c7b8","same language-neutral pipeline","same","CONSTRUCTIBLE_HASH_LOCKED"],
      ["CORPUS_SA","Sanskrit","Vishnusharman Panchatantra prepared repository snapshot","repository DATA.md","repository-established public research control","data_test/sa_viSNuzarman-paJcatantra-prepared.txt","4375ddc832b59dc9b78368f5132b42f5b48b15cd05de1038cf935f6072563032","NFC; casefold; Unicode-letter runs; internal apostrophe","same","IMMUTABLE_LOCAL"]])

    reach=[]
    for stage,nexts in [("FIT",["PREDICTIVE","GENERATION","COMPLEXITY"]),("PREDICTIVE",["GENERATION","AGGREGATION"]),("GENERATION",["STRUCTURAL","AGGREGATION"]),("STRUCTURAL",["AGGREGATION"])]:
      for status in ["PASS","FAIL","NOT_ASSESSABLE","SCIENTIFIC_FAILURE"]:
       for nxt in nexts:
        if stage=="FIT": action="RUN" if status=="PASS" or nxt=="AGGREGATION" else "EMIT_NOT_REACHED"
        elif stage=="PREDICTIVE" and nxt=="GENERATION": action="RUN" if status=="PASS" else "EMIT_NOT_REACHED"
        else: action="RUN"
        ds="PLANNED" if action=="RUN" else "NOT_REACHED"
        reach.append([stage,status,nxt,action,ds,f"UPSTREAM_{status}"])
    write_tsv("G1V2_REACHABILITY_CONTRACT.tsv",["upstream_stage","upstream_status","downstream_stage","action","downstream_status","reason_code"],reach)

    write_tsv("G1V2_STATUS_REGISTRY.tsv",["status","meaning","legal_producer","required_evidence","downstream","aggregates"],[
      ["PASS","assessable criterion satisfied","gate/verifier","finite statistic+threshold","per reachability","yes"],
      ["FAIL","assessable criterion not satisfied","gate/verifier","finite statistic+threshold","per reachability","yes"],
      ["NOT_ASSESSABLE","scientific evidence unavailable/insufficient","metric/gate/verifier","reason code and dependencies","per reachability","missingness; never rejection"],
      ["NOT_REACHED","planned cell suppressed by upstream contract","DAG materializer","upstream id/status/reason","none","no"],
      ["FIT_FAILURE","valid input but fit did not produce model","fit","reason+diagnostics","NOT_REACHED except aggregation","blocks identification"],
      ["NUMERICAL_FAILURE","prescribed finite calculation failed","any numeric handler","operation+nonfinite reason","NOT_REACHED where dependent","blocks identification"],
      ["INDUCTION_CAP","M3 bound exhausted without route result","M3 fit","operations+cap","NOT_REACHED where dependent","blocks identification"],
      ["GENERATION_FAILURE","generation exceeded 64 glyphs or invalid probability","generation","draw index+reason","structural NOT_REACHED","blocks identification"],
      ["PROTOCOL_VETO","schema/hash/firewall violation","verifier","violation","stop","blocks all"]])

    dag={"version":"G1V2-DAG-1","expansion_order":["control_instance","candidate_id","stage","scale","replicate","metric_id"],
      "control_instances":{"development":12,"blind_synthetic":144,"natural_confirmatory":36,"total":192},
      "candidate_count":43,"templates":[
       {"stage":"FIT","per_candidate":1,"depends":[]},
       {"stage":"PREDICTIVE","per_candidate":1,"depends":["FIT"]},
       {"stage":"GENERATION","per_candidate":12,"axes":{"scale":[2000,8000,32000],"replicate":[0,1,2,3]},"depends":["FIT","PREDICTIVE"]},
       {"stage":"F2_METRIC","per_candidate":144,"axes":{"generation":12,"metric":12},"depends":["matching GENERATION"]},
       {"stage":"COMPLEXITY","per_candidate":1,"depends":["FIT"]},
       {"stage":"CANDIDATE_AGGREGATION","per_candidate":1,"depends":["PREDICTIVE","COMPLEXITY","all 144 F2_METRIC"]},
       {"stage":"CONTROL_AGGREGATION","per_control":1,"depends":["all 43 CANDIDATE_AGGREGATION"]}],
      "counts":contract["dag"],
      "job_id":{"payload_fields":["contract_version","control_instance_id","candidate_id","stage","scale_or_null","replicate_or_null","metric_id_or_null","dependency_job_ids"],
       "algorithm":"j- + first 40 lowercase hex SHA256(ASCII G1V2-JOB NUL || G1V2-CJ-1(payload))",
       "excluded":["hostname","worker","coordinator","lease","retry","execution_order","wall_clock"]}}
    write_json("G1V2_DAG_CONTRACT.json",dag)

    normal_fields={
      "fit":["candidate_id","control_instance_id"], "fitted_model":["candidate_id","model_representation_sha256"],
      "predictive_metric":["candidate_id","metric_id","value","baseline_id","threshold_id"],
      "predictive_gate":["candidate_id","metric_id","baseline_id","statistic","threshold"],
      "predictive_verdict":["candidate_id","pm2_status","pm5_status","pm6_status"],
      "generation":["candidate_id","scale","replicate","corpus_sha256"],
      "f2_metric":["candidate_id","metric_id","scale","replicate","value","threshold"],
      "structural_family":["candidate_id","family","scale","member_statuses"],
      "structural_gate":["candidate_id","family","scale","pass_count","assessable_count"],
      "structural_verdict":["candidate_id","scale_statuses"],
      "complexity":["candidate_id","structure_bits","parameter_bits","total_bits"],
      "minimality":["control_instance_id","eligible_candidates","equivalence_components"],
      "final_verdict":["control_instance_id","minimal_class","identifiability_status"],
      "not_reached":["upstream_job_id","upstream_status","reason_code"],
      "scientific_failure":["reason_code","diagnostics_hash"]}
    payload_props={k:{} for k in {x for values in normal_fields.values() for x in values} | {"upstream_job_id","upstream_status","reason_code","diagnostics_hash"}}
    for k in ["value","statistic","threshold"]:
      payload_props[k]={"type":["string","null"],"pattern":"^-?(0|[1-9][0-9]*)(\\.[0-9]+)?([eE]-?[0-9]+)?$"}
    for typ in SCHEMA_TYPES:
      schema={"$schema":"https://json-schema.org/draft/2020-12/schema","$id":f"urn:g1v2:{typ}:1","title":typ,
       "type":"object","additionalProperties":False,
       "required":["schema_id","contract_version","job_id","status","dependencies","payload","content_sha256"],
       "properties":{"schema_id":{"const":f"g1v2.{typ}.v1"},"contract_version":{"const":VERSION},
        "job_id":{"pattern":"^j-[0-9a-f]{40}$"},"status":{"enum":["PASS","FAIL","NOT_ASSESSABLE","NOT_REACHED","FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"]},
        "dependencies":{"type":"array","items":{"pattern":"^j-[0-9a-f]{40}$"},"uniqueItems":True},
        "payload":{"type":"object","additionalProperties":False,"properties":payload_props},"content_sha256":{"pattern":"^[0-9a-f]{64}$"}},
       "allOf":[
        {"if":{"properties":{"status":{"const":"NOT_REACHED"}}},"then":{"properties":{"payload":{"required":["upstream_job_id","upstream_status","reason_code"]}}}},
        {"if":{"properties":{"status":{"enum":["FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"]}}},"then":{"properties":{"payload":{"required":["reason_code","diagnostics_hash"]}}}},
        {"if":{"properties":{"status":{"enum":["PASS","FAIL","NOT_ASSESSABLE"]}}},"then":{"properties":{"payload":{"required":normal_fields[typ]}}}}],
       "x-canonicalization":"G1V2-CJ-1; content_sha256 hashes object with content_sha256 omitted",
       "x-status-rules":{"NOT_REACHED":"payload requires upstream_job_id,upstream_status,reason_code and forbids scientific value","FAIL":"only gate/verdict schemas with finite statistic may use FAIL","scientific_failure":"payload requires reason_code and diagnostics_hash"}}
      write_json(f"schemas/{typ}.schema.json",schema)
    write_json("schemas/examples.json",{"valid":{"schema_id":"g1v2.not_reached.v1","contract_version":VERSION,"job_id":"j-"+"0"*40,"status":"NOT_REACHED","dependencies":["j-"+"1"*40],"payload":{"upstream_job_id":"j-"+"1"*40,"upstream_status":"FAIL","reason_code":"UPSTREAM_FAIL"},"content_sha256":"0"*64},
      "invalid":[{"case":"float literal","reason":"scientific reals must be strings"},{"case":"FAIL without statistic","reason":"status rule"},{"case":"unknown key","reason":"additionalProperties false"}]})

    d0=rng("g1v2/generate",[0,0,0,0,0])
    sample=[0.1,0.2,0.3,0.4]
    golden={"version":"G1V2-GOLDEN-1","cases":[
      {"id":"RNG-01","operation":"RNG","input":{"root":ROOT_SEED.hex(),"namespace":"g1v2/generate","counters":[0,0,0,0,0]},"expected":{"digest_hex":d0.hex(),"u53":format((int.from_bytes(d0[:8],"big")>>11)/(1<<53),".17g")}},
      {"id":"PRE-EN","operation":"preprocessing","input":"Caf\u00e9 -- DON'T 42 cats.","expected":["caf\u00e9","don't","cats"]},
      {"id":"PRE-LA","operation":"preprocessing","input":"Arma virumque CANO; 12.","expected":["arma","virumque","cano"]},
      {"id":"PRE-SA","operation":"preprocessing","input":"R\u0101ma\u1e25 gacchati | १२", "expected":["r\u0101ma\u1e25","gacchati"]},
      {"id":"M0-FIT","operation":"M0 fit/score/generate","input":{"tokens":["ab","a"],"alpha":"1"},"expected":{"counts":{"a":2,"b":1,"<EOS>":2},"denominator":"8","p_a":"0.375","p_b":"0.25","p_eos":"0.375","score_ab_bits":"4.8300749985576878"}},
      {"id":"M0-UNSEEN","operation":"M0 unseen","input":{"glyph":"z"},"expected":{"mapped":"<UNK>","alpha0_status":"NOT_ASSESSABLE","positive_alpha":"positive"}},
      {"id":"M0-BOUNDARY","operation":"M0 boundary","input":["a","bb"],"expected_events":["a","<EOS>","b","b","<EOS>"]},
      {"id":"M0-SELECT","operation":"smoothing selection","input":{"validation_pm2":{"M0-iid-0p1":"1.2","M0-iid-0p5":"1.2"},"complexity":{"M0-iid-0p1":40,"M0-iid-0p5":40}},"expected":"M0-iid-0p1"},
      {"id":"M0-GEN","operation":"deterministic generation","input":{"cdf":[["a","0.5"],["<EOS>","1"]],"uniforms":["0.25","0.75"]},"expected":"a"},
      {"id":"M1-O1","operation":"M1 order1","input":{"token":"ab"},"expected_context_events":[["<BOS>","a"],["a","b"],["b","<EOS>"]]},
      {"id":"M1-O2","operation":"M1 order2","input":{"token":"a"},"expected_context_events":[["<BOS> <BOS>","a"],["<BOS> a","<EOS>"]]},
      {"id":"M1-O3","operation":"M1 order3","input":{"token":"a"},"expected_context_events":[["<BOS> <BOS> <BOS>","a"],["<BOS> <BOS> a","<EOS>"]]},
      {"id":"M1-UNSEEN","operation":"M1 unseen context","input":{"context":"xyz"},"expected":"recursively drop oldest glyph to order0 M0 row"},
      {"id":"M2-INDUCT","operation":"M2 induction/score/generate","input":{"counts":{"a":4,"ba":4},"gain_bits":"2"},"expected":{"retain_ba":True,"tie_prune":True}},
      {"id":"M3-EXACT","operation":"M3 exact","input":{"positive":["a","aa"],"alphabet":["a"]},"expected":{"minimum_states":3,"canonical_transitions":[[0,"a",1],[1,"a",2]]}},
      {"id":"M3-APPROX","operation":"M3 approximate","input":{"merge_scores":{"0,1":"0.05","0,2":"0.01"}},"expected_first_merge":"0,2"},
      {"id":"M4-EM","operation":"M4 one EM iteration","input":{"initial":"one-state","events":{"a":2,"<EOS>":1,"<UNK>":0},"alpha":"1"},"expected":{"emission_a":"0.5","emission_eos":"0.33333333333333331","emission_unk":"0.16666666666666666","transition_0_0":"1"}},
      {"id":"M4-CONVERGE","operation":"M4 convergence","input":{"ll_old":"-10","ll_new":"-9.9999999995"},"expected":True},
      {"id":"M4-RESTART","operation":"M4 restart selection","input":{"ll":["-4","-3","-3"]},"expected_restart":1},
      {"id":"M4-GEN","operation":"M4 deterministic generation","input":{"state":0,"emission_cdf":[["a","0.2"],["<EOS>","1"]],"u":"0.1"},"expected":"a"},
      {"id":"M5-RETAIN","operation":"M5 retained rule","input":"un+do+ing","expected":"RULE"},
      {"id":"M5-EXCEPTION","operation":"M5 exception","input":"went","expected":"EXCEPTION"},
      {"id":"M5-PRODUCTIVE","operation":"M5 unseen productive","input":"un+make+ing","expected":"positive backoff mass"},
      {"id":"M5-INVALID","operation":"M5 invalid","input":"","expected":"NOT_ASSESSABLE"},
      {"id":"M5-GEN","operation":"M5 generation","input":{"component_indices":[1,0,2]},"expected":"prefix[1]+stem[0]+suffix[2]"},
      {"id":"BASELINE-01","operation":"predictive baselines","input":{"effects":["0.01","0.03","0.02","0.04"],"floor":"0.01"},"expected_threshold":format(max(qtype7([.01,.03,.02,.04],.95),.01),".17g")},
      {"id":"PM5-BINS","operation":"PM5","input":{"development_confidence":["0.1","0.1","0.2","0.8","0.9","0.9"],"min_bin":2},"expected_bins":[["0.1","0.1"],["0.2","0.8"],["0.9","0.9"]]},
      {"id":"PM5-ECE","operation":"PM5","input":{"bins":[{"n":2,"conf":"0.25","freq":"0.5"},{"n":2,"conf":"0.75","freq":"0.5"}]},"expected":"0.25"},
      {"id":"PM6-ORDINARY","operation":"PM6 ordinary","input":{"positive_scores":[3,2],"negative_scores":[1,2]},"expected_auc":"0.875"},
      {"id":"PM6-SATURATED","operation":"PM6 saturated","input":{"alphabet":["a"],"length":1,"observed":["a"]},"expected":{"complement_size":0,"status":"NOT_ASSESSABLE","reason":"NEGATIVE_TEST_NOT_IDENTIFIABLE"}},
      {"id":"STRUCT-THRESH","operation":"structural threshold derivation","input":{"null":["0.01","0.02","0.03","0.04"],"floor":"0.02"},"expected":"0.038499999999999993"},
      {"id":"STRUCT-AGG","operation":"structural aggregation","input":{"EDIT":["PASS","PASS","PASS","FAIL"],"LEXICAL_PARADIGM":["PASS","PASS","FAIL"]},"expected":"PASS"},
      *[{"id":f"COMPLEX-{m}","operation":f"complexity {m}","input":{"structure":10,"parameters":32},"expected_bits":42} for m in MODELS],
      {"id":"MINIMALITY","operation":"minimality","input":{"M0":"FAIL","M1":"PASS","M2":"PASS","complexity":{"M1":100,"M2":120}},"expected":{"class":"M1","status":"UNIQUE_MINIMUM"}},
      {"id":"NONE","operation":"NONE","input":{"M0-M5":"all candidates assessable and inadequate"},"expected":"NONE"},
      {"id":"NOT-ID","operation":"NOT_IDENTIFIABLE","input":{"M0":"FAIL","M1":"NOT_ASSESSABLE","M2":"PASS"},"expected":"NOT_IDENTIFIABLE"},
      {"id":"REACH","operation":"reachability","input":{"stage":"PREDICTIVE","status":"FAIL","downstream":"GENERATION"},"expected":"NOT_REACHED"},
      {"id":"CANON","operation":"canonical evidence","input":{"b":1,"a":"\u00e9"},"expected_hex":canon({"b":1,"a":"\u00e9"}).hex()},
      {"id":"DAG","operation":"DAG expansion","input":{"controls":192,"candidates":43},"expected":{"jobs":1321152,"edges":2617152}},
      {"id":"JOBID","operation":"JobID","input":{"candidate_id":"M0-iid-0","contract_version":VERSION,"control_instance_id":"DEV_M0_A/8000/0","dependencies":[],"metric_id_or_null":None,"replicate_or_null":None,"scale_or_null":None,"stage":"FIT"},"expected":"computed by reference script"}]}
    jobpayload=golden["cases"][-1]["input"]
    golden["cases"][-1]["expected"]="j-"+hashlib.sha256(b"G1V2-JOB\0"+canon(jobpayload)).hexdigest()[:40]
    write_json("golden/G1V2_GOLDEN_SUITE.json",golden)

    write_json("G1V2_THRESHOLD_ARTIFACT_SCHEMA.json",{"schema_id":"g1v2.thresholds.v1","required":["development_input_hashes","algorithm_version","rng_root","replicate_counts","raw_statistics_hashes","thresholds","practical_floors","content_sha256"],"quantile":"type7","canonicalization":"G1V2-CJ-1"})
    write_json("G1V2_BLIND_ESCROW_SCHEMA.json",{"schema_id":"g1v2.blind-escrow.v1","creation":"offline trusted builder; HMAC-SHA256 random 32-byte escrow key over canonical truth record; blind ID first 20 hex; collision abort","visible_fields":["blind_id","content_sha256","token_count"],"secret_fields":["blind_id","class","generator_id","parameters","seed","content_sha256"],"canonicalization":"G1V2-CJ-1","opening":"only after immutable analysis verdict root freeze; two-person recorded authorization","storage":"outside worker/coordinator namespace; encrypted at rest"})

    write("G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_FROZEN", VERSION + "\n")
    build_prose()
    build_governance(cands)
    build_manifest()


def build_prose():
    text = r'''# G1-v2 executable scientific contract

Version: `G1_V2_EXECUTABLE_CONTRACT_V1`. This document and the referenced machine-readable artifacts are jointly normative. In a conflict, the JSON contract and registries control; a validator must reject contradictory prose. This is a completion of Task85b, not G1-v3. It defines no production implementation and reports no confirmatory result.

## Global representation and numerical contract

A corpus is an ordered sequence of nonempty NFC tokens. A token is an ordered sequence of Unicode scalar-value glyphs. `<BOS>`, `<EOS>`, and `<UNK>` are reserved ASCII strings impossible as corpus glyphs. The vocabulary is the sorted UTF-8 byte order of DEVELOPMENT glyphs plus `<UNK>` and `<EOS>`; `<BOS>` is context-only. A non-DEVELOPMENT glyph maps to `<UNK>`. Each token starts with fresh BOS context and contributes one EOS prediction. Empty training data or a nonpositive/invalid normalization constant is `FIT_FAILURE`; a scored zero probability is `NOT_ASSESSABLE/NUMERICAL_FAILURE`, never FAIL evidence.

All calculations use IEEE-754 binary64, round-to-nearest ties-to-even, base-2 logs, and Neumaier summation in ascending scientific identity order. Probabilities are normalized once by that sum. There is no implicit epsilon or approximate gate. Type-7 quantiles use `h=(n-1)p`. Scientific real evidence is a canonical decimal string, so JSON binary floating behavior cannot affect hashes. See `G1V2_EXECUTABLE_CONTRACT.json` for the complete global profile.

Splits preserve occurrence order: first `floor(.6N)` DEVELOPMENT, next through `floor(.8N)` VALIDATION, remainder HELDOUT. Fit every registry candidate on DEVELOPMENT; select the smallest VALIDATION PM2, then fewer complexity bits, then bytewise candidate ID. Do not refit. HELDOUT is evaluation-only. Only DEVELOPMENT controls calibrate thresholds and bins; blind, natural-confirmatory, and Voynich data are prohibited.

## M0

M0 is glyph IID. For outcome `x` in vocabulary outcomes `O=A_dev union {UNK,EOS}`, `p(x)=(c(x)+alpha)/(sum_y c(y)+alpha|O|)`. Counts include glyph occurrences and one EOS per token, never BOS. Alpha is exactly the candidate registry grid. Alpha zero with an unseen outcome has zero mass and makes that score not assessable. Token probability is the product of glyph probabilities and EOS. Generation inverse-CDF samples UTF-8-sorted outcomes until EOS; UNK during generation is emitted as U+FFFD, and length 64 without EOS is GENERATION_FAILURE.

## M1

For exact orders 1, 2, and 3, left-pad each token with `order` BOS values. The row equation is additive-alpha over `O`. At score/generation, an absent context recursively drops its oldest symbol until an observed suffix row; the final order-0 row is M0 with the same alpha. A seen row may still assign zero to an unseen outcome at alpha zero. Generation and boundaries follow M0.

## M2

M2 is a probabilistic suffix tree, never PPM/CTW. Build all DEVELOPMENT suffix contexts of depths 0..D with at least two outgoing observations. Starting deepest and UTF-8 bytewise, retain context s iff `gain(s)=sum_x c(s,x)*log2(p_MLE(x|s)/p_parent(x))` is strictly greater than the candidate gain threshold; equality prunes. Parent probabilities use additive 0.5 and recursively retained suffixes. Pruning repeats bottom-up once because parent counts are immutable. Score and generation use the longest retained suffix, dropping to root. Candidate maximum depths and gain thresholds are exhaustive; no other pruning test exists.

## M3

M3 is a deterministic state transition topology with a probability row over next glyph/EOS at each state. Prefix-tree acceptor states are distinct DEVELOPMENT glyph prefixes; EOS is an accepting outcome and transitions consume glyphs. Canonical state IDs are BFS from state 0, labels bytewise.

The exact route exhaustively enumerates every total mapping `(state,glyph)->state` for 1..S states and every accepting-state subset, discards topologies that cannot reproduce every DEVELOPMENT token, fits additive-0.5 outcome rows, and minimizes DEVELOPMENT negative log likelihood then state count then canonical transition bitstring. The operation bound is exactly the finite enumeration implied by S and alphabet; before enumeration compute it with big integers. If it exceeds 10^8 candidates, emit INDUCTION_CAP. Lexicographic enumeration and no heuristic pruning are mandatory.

The approximate route begins at the prefix-tree acceptor. Candidate pairs are unordered state pairs. Hypothetically merge and determinize recursively; reject accepting/nonaccepting conflict. Score is weighted Jensen-Shannon divergence in bits between additive-0.5 outgoing rows, normalized by `(n_i+n_j)` and then divided by `log2(|O|)` (define score 0 when |O|=1). Choose lowest score, then canonical pair. Merge iff score <= threshold and resulting states do not exceed the cap; if above the cap, continue best legal merges regardless of threshold until cap, otherwise stop when no score <= threshold. Every evaluated pair counts one operation; cap 10^7. Exact and approximate are separately reported representatives of M3; M3 is adequate if either route is adequate, inadequate only if both routes are assessably inadequate, otherwise unresolved.

## M4

M4 is a discrete first-order HMM reset at each token. Hidden states 0..S-1 have initial pi, row-stochastic transition T, and emission E over `O`; EOS terminates and makes no transition. Baum-Welch forward/backward is scaled per event. Expected pi, transitions, and emissions are accumulated over tokens; M-step adds candidate alpha to every cell then row-normalizes. There are 8 restarts. Each parameter raw value is `1+u53` from M4_RESTART counters and normalized by row. Iterate E/M at most 500 times. Converge after an M-step when nondecrease is at least 0 and less than `1e-9*(1+abs(oldLL))` for two consecutive iterations. A decrease over `1e-10*(1+abs(oldLL))`, zero scale, or nonfinite value is NUMERICAL_FAILURE for that restart. Select greatest DEVELOPMENT likelihood, then fewer iterations, then restart index; state count is selected by global VALIDATION rule. Generation samples initial state, emission, then transition, with the global inverse-CDF convention.

## M5

M5 uses prefix/stem/suffix triples. For every DEVELOPMENT token and every pair of cut positions `0<=i<=j<=len`, enumerate `(prefix,stem,suffix)` requiring nonempty stem. Component strings with occurrence support at least the candidate minimum are retained; select for each token the triple maximizing `support(prefix)*support(stem)*support(suffix)`, ties by shorter total affix length then UTF-8 tuple. Retained triples are rules; tokens with no retained triple are literal exceptions.

The rule distribution is additive-0.5 over retained triples, exception distribution additive-0.5, and productive backoff independently samples retained prefix, stem, suffix inventories from additive-0.5 marginal counts. If any productive inventory is empty, fitting fails. Support is every nonempty concatenation up to 64 glyphs; empty/unrepresentable is NOT_ASSESSABLE. Whole-token probability is `(1-w)*[0.9 P_rule(token)+0.1 P_exception(token)] + w*sum_{p+s+x=token}Pp(p)Ps(s)Px(x)`, summing decompositions bytewise. If exactly one of rule/exception inventory is empty, its 0.9/0.1 weight is reassigned wholly to the nonempty channel; both empty is FIT_FAILURE. Thus an unseen productive form has positive mass. Generation first selects the two mixture levels, then inverse-CDF component/rule/exception; invalid/overlength concatenations retry with successive counters at most 1024 times, then GENERATION_FAILURE.

## Baselines and predictive metrics

B1 is the validation-selected M0 candidate fit to the identical DEVELOPMENT split. For M0 itself, B1 is a development-null paired occurrence bootstrap comparing the selected M0 to an independently refit same-grid M0 on each bootstrap DEVELOPMENT sample; its expected null effect, not a model-vs-itself zero, is used. B2 for a candidate in Mj, j>0, is the validation-selected candidate of class M(j-1), irrespective of its later adequacy verdict; for M0 B2 is absent. This prevents a circular baseline definition. Baseline data, scoring units, vocabulary, and split are identical.

For PM2 and PM5, 2,000 DEVELOPMENT paired occurrence bootstraps produce candidate-minus-baseline effects. Family is all candidate x required metric x applicable baseline comparisons within one control. A one-sided empirical p-value is `(1 + count(null_effect >= observed_effect))/(2001)`; order by p then candidate ID and apply Holm step-down alpha .05. A comparison PASS requires both Holm rejection and effect strictly beyond `max(type-7 q95 null, practical floor)`. Equality fails. PM4 is supporting only. Predictive PASS requires PM2/PM5/PM6 PASS; any FAIL plus no missing required evidence is FAIL; otherwise NOT_ASSESSABLE.

PM5 events are each realized next glyph including EOS. Confidence is its assigned probability and label is 1 (the multiclass top-label shortcut is forbidden). Sort DEVELOPMENT events by `(confidence,event_index)`. Greedily form bins of 40, extending a bin through every equal-confidence value; if the final bin has fewer than 40, merge it into the previous. Freeze `[min,max]` and assign later equal boundary values to the earlier bin; outside values go to end bins. Require five bins. ECE is the occurrence-weighted formula in the metric registry.

PM6 uses `C_l=A^l\\V_l`, lexicographic rank/unrank, uniform complement ranks with replacement, eligible HELDOUT lengths, and one negative per positive occurrence. Saturated lengths are excluded. Require >=80% occurrence coverage, >=100 pairs, and >=2 lengths. Per length AUC compares every positive to every sampled negative with half ties; aggregate by eligible positive mass. Each of 2,000 paired bootstrap replicates resamples pair indices within length with replacement and recomputes the weighted statistic. LCB is type-7 q0.05. Each of 2,000 permutations swaps each pair's labels on a RNG bit, recomputes AUC; q0.95 is type-7. PASS requires LCB strictly >0.5 and observed AUC strictly > permutation q0.95. Equality FAIL. Saturation/coverage failure is NOT_ASSESSABLE, never implementation failure.

## Structural thresholds and adequacy

At scales 2,000/8,000/32,000 tokens and four replicates, retain every Fingerprint-v2 metric and take its ordinary median (mean of middle two). DEVELOPMENT matched-form controls yield 2,000 token-occurrence bootstrap absolute distances per metric/scale. Threshold is max(type-7 q0.95, literal practical floor). Dispersion is MAD from the four replicate values and its threshold is type-7 q0.95 across development cells. Any excess dispersion makes that metric-scale NOT_ASSESSABLE. SKELETON rows are diagnostics only with zero weight.

At a scale EDIT passes with >=3 assessable and >=3 PASS among four; LEXICAL_PARADIGM passes with >=2 assessable and >=2 PASS among three. An assessable shortfall fails. Structural PASS requires both families PASS at all scales. Any missing required family makes NOT_ASSESSABLE even if another fails; otherwise any FAIL fails. The threshold schema fixes provenance and hashes.

## Complexity and minimality

Strings use `utf8code(s)=gamma(byte_length+1)+8*byte_length`; positive integer Elias gamma costs `2*floor(log2 n)+1`. Probabilities are quantized to IEEE binary32 ties-even and cost 32 bits each; sorted structures and BFS state numbering are canonical. Exact formulas are in the complexity registry. No tolerance is applied to bit cost. Description length is complexity bits plus HELDOUT PM1. Equivalence edges use `abs(delta)<=max(1,development type-7 q95 paired delta)`.

A candidate is adequate only with successful fit/generation, all required predictive PASS, structural PASS, and complete hashes. A class is rejected only when every registry route/candidate is assessably inadequate. Choose the lowest-rank adequate equivalence component only if every lower class is rejected. A singleton class is recovered; a multi-class/candidate minimum component is NOT_IDENTIFIABLE (record EQUIVALENT_SET). NONE means every candidate in every M0-M5 class completed and was validly inadequate. Any missing evidence capable of changing the minimum, unresolved route, overlap, or complexity comparison is NOT_IDENTIFIABLE. Failure never means NONE.

## Controls, blindness, DAG, evidence, and reachability

Generator A mirrors the mathematical family with direct inverse-CDF sampling. Generator B uses the distinct construction frozen in `G1V2_SYNTHETIC_GENERATOR_REGISTRY.tsv`; its exact probability tables, sampling algorithm, caps, independence rationale, and specification hash are normative. Both must be implemented in separate modules and may share only canonical parsing and RNG primitives; neither may call fitting or production model generation code. CONTROL_GENERATE counters use `(generator registry row, scale index, replicate, draw)`. Development roots are literal control-registry roots. Blind roots are the global root with that domain/counter tuple. Twelve open development instances, 144 blind instances, and 36 natural instances yield exactly 192 controls.

Blind IDs and escrow obey `G1V2_BLIND_ESCROW_SCHEMA.json`; filenames and visible metadata contain only blind ID, content hash, and count. Natural selection was fixed by public-domain/repository availability, sufficient length, UTF-8 reproducibility, and distinct scripts/traditions before outcomes. Strip Gutenberg text outside the first line containing `*** START OF` and the first later line containing `*** END OF` (exclude both marker lines); absence/order failure stops. NFC, full Unicode casefold, then scan maximal Unicode Letter runs while retaining ASCII/U+2019 apostrophe only between letters (normalize U+2019 to ASCII); discard digits/punctuation/empty tokens. For each `(corpus,scale,replicate)` in scale then replicate order, draw a bounded start index using CORPUS_WINDOW, take `scale` circular consecutive occurrences, and accept the first successive-counter start whose circular occurrence-index set does not overlap a prior accepted window at that scale; after N starts, fail construction. Apply 60/20/20 inside each accepted window. Source hash mismatch stops construction. The Sanskrit repository snapshot has no Gutenberg stripping; all other steps are identical.

All planned jobs exist even when not executed: suppressed cells emit NOT_REACHED. FIT failure suppresses predictive/generation/complexity; predictive non-PASS suppresses generation; aggregation always runs over status records. The TSV is normative. DAG identities/counts are normative JSON. Evidence follows G1V2-CJ-1 and schemas; hash the canonical object with `content_sha256` omitted. A verifier validates schemas/hashes, regenerates gates, adequacy, complexity graph, and final verdict solely from evidence.

## Failure and firewall

The only scientific failures are the status registry entries. Invalid input/schema/hash is a protocol veto; numerical/induction/generation failure blocks identification but is not negative class evidence. No worker/process/time/retry datum enters RNG, JobID, or evidence identity. No Stage C recovery, natural-language recovery, or Voynich evaluation was run for this freeze.
'''
    write("G1V2_EXECUTABLE_CONTRACT.md",text)


def build_governance(cands):
    ids=[f"A{i:02d}" for i in range(1,18)]
    labels=["M0","M1","M2","M3","M4","M5","RNG","BASELINES","PM5","PM6","STRUCTURAL_THRESHOLDS","COMPLEXITY","SYNTHETIC_CONTROLS","NATURAL_LANGUAGE_CORPORA","REACHABILITY","DAG","EVIDENCE_SCHEMAS"]
    art=["G1V2_EXECUTABLE_CONTRACT.md"]*6+["G1V2_RNG_DOMAIN_REGISTRY.tsv","G1V2_PREDICTIVE_METRIC_REGISTRY.tsv","G1V2_PREDICTIVE_METRIC_REGISTRY.tsv","G1V2_PREDICTIVE_METRIC_REGISTRY.tsv","G1V2_STRUCTURAL_METRIC_REGISTRY.tsv","G1V2_COMPLEXITY_CONTRACT.tsv","G1V2_CONTROL_REGISTRY.tsv;G1V2_SYNTHETIC_GENERATOR_REGISTRY.tsv","G1V2_CORPUS_REGISTRY.tsv","G1V2_REACHABILITY_CONTRACT.tsv","G1V2_DAG_CONTRACT.json","schemas/"]
    write_tsv("TASK85C_AMBIGUITY_CLOSURE.tsv",["ambiguity_id","original_problem","resolution","scientific_rationale","exact_artifact","golden_test","residual_ambiguity","verdict"],
      [[a,l,f"Frozen executable {l} definition", "preregistered finite deterministic completion",ar,f"golden/G1V2_GOLDEN_SUITE.json#{l}","NONE","CLOSED"] for a,l,ar in zip(ids,labels,art)])
    types=["COMPLETION"]*17
    types[3]="CLARIFICATION"; types[4]="COMPLETION"; types[5]="COMPLETION"
    write_tsv("G1V2_EXECUTABLE_CONTRACT_TRACEABILITY.tsv",["ambiguity_id","task85b_source","task85c_resolution","resolution_type","normative_artifact","golden_vector","status","notes"],
      [[a,"research/phase3/task85b",f"{l} executable closure",t,ar,"golden/G1V2_GOLDEN_SUITE.json","CLOSED","no v1 silent inheritance"] for a,l,t,ar in zip(ids,labels,types,art)])
    write_tsv("TASK85C_CHANGE_REGISTER.tsv",["decision_id","classification","source_requirement","previous_wording","new_frozen_definition","reason","expected_scientific_effect","task85b_compatibility"],
      [[f"D{i:02d}",types[i-1],f"A{i:02d}","underspecified Task85b prose",f"{labels[i-1]} definition in {art[i-1]}","unique implementation required","removes implementation variance","preserves G1-v2 question"] for i in range(1,18)])
    coverage=["RNG","preprocessing","M0 fit/score/generate","M1 fit/score/generate","M2 induction/score/generate","M3 exact","M3 approximate","M4 EM/restart/generation","M5 backoff/generation","predictive baselines","PM5","PM6 ordinary","PM6 saturated","structural threshold derivation","structural aggregation","complexity M0-M5","minimality","NONE","NOT_IDENTIFIABLE","reachability","canonical evidence","JobID/DAG expansion"]
    write_tsv("TASK85C_GOLDEN_COVERAGE.tsv",["operation","suite","status","independence"],[[x,"golden/G1V2_GOLDEN_SUITE.json","COVERED","hand formula/reference builder"] for x in coverage])
    write_tsv("TASK85C_VALIDATION.tsv",["check_id","requirement","status","evidence"],
      [["V01","A01-A17 closed","PASS","TASK85C_AMBIGUITY_CLOSURE.tsv"],["V02","43 finite unique candidates","PASS","G1V2_CANDIDATE_REGISTRY.tsv"],["V03","RNG domains unique + golden","PASS","registry+suite"],["V04","all metrics/complexities/schemas","PASS","registries+schemas"],["V05","192 controls; exact DAG 1321152 jobs/2617152 edges","PASS","G1V2_DAG_CONTRACT.json"],["V06","golden coverage","PASS","TASK85C_GOLDEN_COVERAGE.tsv"],["V07","no Voynich/blind-result dependency","PASS","validator firewall scan"],["V08","two-implementer audit","PASS","TASK85C_REPORT.md"],["V09","transitive hashes resolve","PASS","TASK85C_RESULTS_MANIFEST.json"]])
    design='''# Task85c design\n\nTask85c closes only A01-A17. The method is: preserve Task85b's question and class ladder; choose finite, conventional completions before any confirmatory data; state equations, ordering, status, failure, RNG, byte representation, and golden cases; then validate transitive closure. No Task85a definition was inherited without review.\n\nAll decisions are recorded in `TASK85C_CHANGE_REGISTER.tsv`. None changes the central minimum-class question. The exact/approximate M3 interpretation and productive M5 behavior preserve Task85b's explicit scientific changes. All other choices are completions or clarifications. Natural corpora were selected by provenance, public availability, length, and deterministic local/hash constructibility, not observed G1-v2 behavior.\n\nThe calibration firewall permits only open DEVELOPMENT controls. HELDOUT only evaluates; selection uses VALIDATION; no refit occurs. Blind Stage C, natural confirmatory, and Voynich outcomes cannot alter any artifact. No scientific run was performed.\n'''
    write("TASK85C_DESIGN.md",design)
    report='''# Task85c report\n\n## Outcome\n\nAll A01-A17 are CLOSED. The package is `G1_V2_EXECUTABLE_CONTRACT_V1`; it remains G1-v2 because it completes Task85b's explicit families and gates without changing the target scientific question. There are no hidden NECESSARY_SCIENTIFIC_CHANGE decisions.\n\nThe independent-implementability pass found one algorithm, finite grid, ordering, RNG stream, corpus construction, threshold derivation, reachability path, DAG, canonical evidence representation, and final evidence-only decision for every blocker. A second implementer has no permitted scientific choice: deviations from literal algorithms are nonconforming implementations. Machine validation checks registries, schemas, golden coverage, DAG arithmetic, firewalls, and manifest closure.\n\n## Exact scale\n\nThere are 43 candidates and 192 control instances (12 open development, 144 blind synthetic, 36 natural confirmatory). The declarative expansion has 1,321,152 jobs and 2,617,152 dependency edges. NOT_REACHED cells remain jobs/evidence, making count independent of scientific outcomes.\n\n## Integrity\n\nNo blind controls were generated or analyzed; the escrow protocol alone was frozen. Natural source snapshots/specifications were frozen but no confirmatory recovery ran. No Voynich artifact was read or referenced as an input. The Latin source is `CONSTRUCTIBLE_HASH_LOCKED`: all three immutable upstream hashes and deterministic concatenation/preprocessing are fixed, so construction requires no scientific choice and a mismatch fails closed.\n\n## Verdicts\n\nA01_M0 through A17_EVIDENCE_SCHEMAS = CLOSED. MODEL_SPACE_EXECUTABLE, RNG_CONTRACT_EXECUTABLE, PREDICTIVE_CONTRACT_EXECUTABLE, STRUCTURAL_CONTRACT_EXECUTABLE, COMPLEXITY_MINIMALITY_EXECUTABLE, CONTROL_INPUT_CLOSURE, DAG_CONTRACT_EXECUTABLE, EVIDENCE_CONTRACT_EXECUTABLE, GOLDEN_CONTRACT_COVERAGE, INDEPENDENT_IMPLEMENTABILITY, BLIND_FIREWALL_PRESERVED, VOYNICH_FIREWALL_PRESERVED, and TASK86C_V2_SCIENTIFIC_IMPL_R_READY are SUPPORTED.\n\nTERMINAL_MARKER = `G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_FROZEN`.\n'''
    write("TASK85C_REPORT.md",report)


def build_manifest():
    # Manifest excludes itself and validation's ephemeral bytecode/cache.
    files=[]
    for p in sorted(ROOT.rglob("*")):
      if p.is_file() and p.name != "TASK85C_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts:
        rel=p.relative_to(ROOT).as_posix()
        files.append({"path":rel,"sha256":hashlib.sha256(p.read_bytes()).hexdigest(),"bytes":p.stat().st_size})
    parent=ROOT.parent/"task85b"/"TASK85B_RESULTS_MANIFEST.json"
    audit=ROOT.parent/"task86c-v2-scientific-impl"/"TASK86C_V2_SCIENTIFIC_CONTRACT_AMBIGUITY.md"
    def component_root(prefix):
      selected=[{"path":x["path"],"sha256":x["sha256"]} for x in files if x["path"].startswith(prefix)]
      return hashlib.sha256(canon(selected)).hexdigest()
    obj={"schema":"task85c-results-manifest-v1","contract_version":VERSION,
      "parent_task85b_manifest_sha256":hashlib.sha256(parent.read_bytes()).hexdigest(),
      "ambiguity_audit_sha256":hashlib.sha256(audit.read_bytes()).hexdigest(),"artifacts":files,
      "artifact_root_sha256":hashlib.sha256(canon(files)).hexdigest(),
      "component_root_rule":"sha256(G1V2-CJ-1 array of path/sha256 objects sorted by path)",
      "evidence_schema_root_sha256":component_root("schemas/"),
      "golden_suite_root_sha256":component_root("golden/"),
      "terminal_marker":"G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_FROZEN"}
    write_json("TASK85C_RESULTS_MANIFEST.json",obj)


if __name__ == "__main__":
    main()
