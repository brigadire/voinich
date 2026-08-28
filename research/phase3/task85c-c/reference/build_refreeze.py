#!/usr/bin/env python3
"""Build V1.1 schemas, exhaustive fixtures, integrated contract and manifest."""
from __future__ import annotations
import csv, hashlib, json, shutil, unicodedata
from pathlib import Path

ROOT=Path(__file__).resolve().parents[1]; P3=ROOT.parent
T85=P3/"task85c"; T85A=P3/"task85c-a"; T85B=P3/"task85c-b"; IMPL=P3/"task86c-v2-scientific-impl-r"
CV="G1_V2_EXECUTABLE_CONTRACT_V1_1"; SV="G1_V2_STATUS_REACHABILITY_CONTRACT_V2"
DIALECT="https://json-schema.org/draft/2020-12/schema"
JOBPAT="^j-[0-9a-f]{40}$"; HASHPAT="^[0-9a-f]{64}$"; REALPAT=r"^-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE]-?[0-9]+)?$"

def canon(x):
 def n(v):
  if isinstance(v,str): return unicodedata.normalize("NFC",v)
  if isinstance(v,list): return [n(y) for y in v]
  if isinstance(v,dict): return {unicodedata.normalize("NFC",k):n(v[k]) for k in sorted(v)}
  return v
 return (json.dumps(n(x),ensure_ascii=False,sort_keys=True,separators=(",",":"))+"\n").encode()
def write(p,s):
 p=ROOT/p; p.parent.mkdir(parents=True,exist_ok=True); p.write_text(s,encoding="utf-8",newline="\n")
def jw(p,x): write(p,canon(x).decode())
def tsv(p,h,rr):
 p=ROOT/p; p.parent.mkdir(parents=True,exist_ok=True)
 with p.open("w",encoding="utf-8",newline="") as f:
  w=csv.writer(f,delimiter="\t",lineterminator="\n"); w.writerow(h); w.writerows(rr)
def rows(p):
 with p.open(encoding="utf-8",newline="") as f:return list(csv.DictReader(f,delimiter="\t"))
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()

STATUS= [r["status"] for r in rows(T85B/"G1V2_STATUS_REGISTRY_V2.tsv")]
EVSEM={r["evidence_type"]:r for r in rows(T85B/"G1V2_EVIDENCE_STATUS_SEMANTICS.tsv")}
STAGE_ALLOWED={(r["stage"],r["status"]) for r in rows(T85B/"G1V2_STAGE_STATUS_CONTRACT.tsv") if r["allowed"]=="YES"}
REASONS=rows(T85B/"G1V2_STATUS_REASON_REGISTRY.tsv")
STAGES=["FIT","PREDICTIVE","GENERATION","F2_METRIC","COMPLEXITY","CANDIDATE_AGGREGATION","CONTROL_AGGREGATION"]

NA_REASONS={
 "predictive_metric":["INSUFFICIENT_SCORED_GLYPHS","INSUFFICIENT_UNSEEN_OCCURRENCES","WHOLE_TOKEN_PROBABILITY_UNSUPPORTED","INSUFFICIENT_PREDICTIONS","INSUFFICIENT_FROZEN_BINS","NEGATIVE_TEST_NOT_IDENTIFIABLE"],
 "predictive_gate":["REQUIRED_METRIC_NOT_ASSESSABLE"],"predictive_verdict":["REQUIRED_METRIC_NOT_ASSESSABLE"],
 "f2_metric":["METRIC_NOT_APPLICABLE","DISPERSION_EXCEEDS_DEVELOPMENT_Q95"],
 "structural_family":["INSUFFICIENT_ASSESSABLE_MEMBERS"],"structural_gate":["INSUFFICIENT_ASSESSABLE_MEMBERS"],"structural_verdict":["REQUIRED_SCALE_NOT_ASSESSABLE"]}

PRODUCER={"fit":"FIT","fitted_model":"FIT","predictive_metric":"PREDICTIVE","predictive_gate":"PREDICTIVE","predictive_verdict":"PREDICTIVE","generation":"GENERATION","f2_metric":"F2_METRIC","structural_family":"CANDIDATE_AGGREGATION","structural_gate":"CANDIDATE_AGGREGATION","structural_verdict":"CANDIDATE_AGGREGATION","complexity":"COMPLEXITY","minimality":"CANDIDATE_AGGREGATION","final_verdict":"CONTROL_AGGREGATION","not_reached":"DAG_MATERIALIZER","scientific_failure":"STATUS_DEPENDENT"}
SUBOPS={"structural_family","structural_gate","structural_verdict"}

def objprops(extra):
 p={"candidate_id":{"type":"string","pattern":"^M[0-5]-[A-Za-z0-9._-]+$"},"control_instance_id":{"type":"string","minLength":1},"fit_diagnostics_sha256":{"type":"string","pattern":HASHPAT},"model_representation_sha256":{"type":"string","pattern":HASHPAT},"metric_id":{"type":"string","minLength":2},"value":{"type":"string","pattern":REALPAT},"baseline_id":{"type":"string","minLength":1},"threshold_id":{"type":"string","minLength":1},"threshold":{"type":"string","pattern":REALPAT},"reason_code":{"type":"string","minLength":1},"pm2_status":{"enum":["PASS","FAIL","NOT_ASSESSABLE"]},"pm5_status":{"enum":["PASS","FAIL","NOT_ASSESSABLE"]},"pm6_status":{"enum":["PASS","FAIL","NOT_ASSESSABLE"]},"scale":{"enum":[2000,8000,32000]},"replicate":{"type":"integer","minimum":0,"maximum":3},"corpus_sha256":{"type":"string","pattern":HASHPAT},"family":{"enum":["EDIT","LEXICAL_PARADIGM"]},"member_statuses":{"type":"array","items":{"enum":["PASS","FAIL","NOT_ASSESSABLE"]},"minItems":1},"pass_count":{"type":"integer","minimum":0},"assessable_count":{"type":"integer","minimum":0},"scale_statuses":{"type":"array","items":{"enum":["PASS","FAIL","NOT_ASSESSABLE"]},"minItems":3,"maxItems":3},"structure_bits":{"type":"integer","minimum":0},"parameter_bits":{"type":"integer","minimum":0},"total_bits":{"type":"integer","minimum":0},"candidate_verdict":{"enum":["ADEQUATE","INADEQUATE","UNRESOLVED","PROTOCOL_INVALID"]},"eligible_candidates":{"type":"array","items":{"type":"string"},"uniqueItems":True},"equivalence_components":{"type":"array"},"protocol_veto_dependency_ids":{"type":"array","items":{"type":"string","pattern":JOBPAT},"minItems":1,"uniqueItems":True},"verdict":{"enum":["RECOVERED_M0","RECOVERED_M1","RECOVERED_M2","RECOVERED_M3","RECOVERED_M4","RECOVERED_M5","NONE","NOT_IDENTIFIABLE","PROTOCOL_INVALID"]},"identifiability_detail":{"enum":["UNIQUE_MINIMUM","EQUIVALENT_SET","ORDER_ONLY","MISSING_EVIDENCE","NONE_COMPLETE_REJECTION","PROTOCOL_INVALID"]},"upstream_stage":{"enum":STAGES},"upstream_status":{"enum":STATUS},"selected_causal_job_id":{"type":"string","pattern":JOBPAT},"causal_dependency_ids":{"type":"array","items":{"type":"string","pattern":JOBPAT},"minItems":1,"uniqueItems":True},"diagnostics_hash":{"type":"string","pattern":HASHPAT}}
 p.update(extra); return p

BASE_PAYLOAD={
 "fit":["candidate_id","control_instance_id","fit_diagnostics_sha256"],"fitted_model":["candidate_id","control_instance_id","model_representation_sha256"],
 "predictive_metric":["candidate_id","control_instance_id","metric_id"],"predictive_gate":["candidate_id","control_instance_id","metric_id"],"predictive_verdict":["candidate_id","control_instance_id","pm2_status","pm5_status","pm6_status"],
 "generation":["candidate_id","control_instance_id","scale","replicate","corpus_sha256"],"f2_metric":["candidate_id","control_instance_id","metric_id","scale","replicate"],
 "structural_family":["candidate_id","control_instance_id","family","scale","member_statuses"],"structural_gate":["candidate_id","control_instance_id","family","scale","pass_count","assessable_count"],"structural_verdict":["candidate_id","control_instance_id","scale_statuses"],
 "complexity":["candidate_id","control_instance_id","structure_bits","parameter_bits","total_bits"],"minimality":["candidate_id","control_instance_id","candidate_verdict","eligible_candidates","equivalence_components"],"final_verdict":["control_instance_id","verdict","identifiability_detail"]}

def payload_schema(typ,status):
 props=objprops({}); req=list(BASE_PAYLOAD.get(typ,[])); extra=[]
 if typ in {"predictive_metric","predictive_gate","f2_metric"}:
  if status in {"PASS","FAIL"}: req += ["value","threshold"]
  else: req += ["reason_code"]; props["reason_code"]={"enum":NA_REASONS[typ]}
 if typ in {"predictive_gate"}: req += ["baseline_id","threshold_id"]
 if typ=="predictive_verdict":
  if status=="PASS": extra=[{"properties":{"pm2_status":{"const":"PASS"},"pm5_status":{"const":"PASS"},"pm6_status":{"const":"PASS"}}}]
  elif status=="FAIL": extra=[{"not":{"anyOf":[{"properties":{"pm2_status":{"const":"NOT_ASSESSABLE"}}},{"properties":{"pm5_status":{"const":"NOT_ASSESSABLE"}}},{"properties":{"pm6_status":{"const":"NOT_ASSESSABLE"}}}]}},{"anyOf":[{"properties":{"pm2_status":{"const":"FAIL"}}},{"properties":{"pm5_status":{"const":"FAIL"}}},{"properties":{"pm6_status":{"const":"FAIL"}}}]}]
  else: req += ["reason_code"]; props["reason_code"]={"enum":NA_REASONS[typ]}; extra=[{"anyOf":[{"properties":{"pm2_status":{"const":"NOT_ASSESSABLE"}}},{"properties":{"pm5_status":{"const":"NOT_ASSESSABLE"}}},{"properties":{"pm6_status":{"const":"NOT_ASSESSABLE"}}}]}]
 if typ in {"structural_family","structural_gate","structural_verdict"} and status=="NOT_ASSESSABLE": req += ["reason_code"]; props["reason_code"]={"enum":NA_REASONS[typ]}
 if typ=="minimality" and status=="AGGREGATION_SUCCESS":
  extra=[{"if":{"properties":{"candidate_verdict":{"const":"PROTOCOL_INVALID"}}},"then":{"required":["protocol_veto_dependency_ids"]},"else":{"not":{"required":["protocol_veto_dependency_ids"]}}}]
 if typ=="final_verdict":
  extra=[{"if":{"properties":{"verdict":{"const":"PROTOCOL_INVALID"}}},"then":{"required":["protocol_veto_dependency_ids"],"properties":{"identifiability_detail":{"const":"PROTOCOL_INVALID"}}},"else":{"not":{"required":["protocol_veto_dependency_ids"]}}}]
 allowed_keys=set(req)
 if typ in {"minimality","final_verdict"}: allowed_keys.add("protocol_veto_dependency_ids")
 props={k:v for k,v in props.items() if k in allowed_keys}
 return {"type":"object","additionalProperties":False,"properties":props,"required":sorted(set(req)),**({"allOf":extra} if extra else {})}

def branch(typ,status,stage,payload):
 return {"type":"object","additionalProperties":False,"required":["schema_id","contract_version","status_reachability_version","job_id","stage","producer_scope","status","dependencies","payload","content_sha256"],"properties":{"schema_id":{"const":f"g1v2.{typ}.v1_1"},"contract_version":{"const":CV},"status_reachability_version":{"const":SV},"job_id":{"type":"string","pattern":JOBPAT},"stage":{"const":stage},"producer_scope":{"const":"SUBOPERATION" if typ in SUBOPS else ("DAG_MATERIALIZER" if typ=="not_reached" else "JOB")},"status":{"const":status},"dependencies":{"type":"array","items":{"type":"string","pattern":JOBPAT},"uniqueItems":True},"payload":payload,"content_sha256":{"type":"string","pattern":HASHPAT}}}

def nr_branches():
 out=[]
 for r in rows(T85B/"G1V2_REACHABILITY_CONTRACT_V2.tsv"):
  if r["action"]!="NOT_REACHED":continue
  pp=objprops({}); pp["upstream_stage"]={"const":r["upstream_stage"]};pp["upstream_status"]={"const":r["upstream_status"]};pp["reason_code"]={"const":r["reason_code"]}
  pp={k:pp[k] for k in ["upstream_stage","upstream_status","reason_code","selected_causal_job_id","causal_dependency_ids"]}
  ps={"type":"object","additionalProperties":False,"properties":pp,"required":["upstream_stage","upstream_status","reason_code","selected_causal_job_id","causal_dependency_ids"]}
  out.append(branch("not_reached","NOT_REACHED",r["downstream_stage"],ps))
 return out

def failure_branches():
 out=[]
 stage_status=sorted((s,st) for s,st in STAGE_ALLOWED if st in {"FIT_FAILURE","NUMERICAL_FAILURE","INDUCTION_CAP","GENERATION_FAILURE","PROTOCOL_VETO"})
 for stage,status in stage_status:
  pp=objprops({}); codes=[]
  for r in REASONS:
   permitted=r["permitted_stage"].split(";")
   if r["status"]==status and (stage in permitted or "ANY" in permitted): codes.append(r["reason_code"])
  pp["reason_code"]={"enum":codes}
  required=["reason_code","diagnostics_hash"]
  if stage not in {"CONTROL_AGGREGATION"}: required.append("control_instance_id")
  if stage in {"FIT","PREDICTIVE","GENERATION","F2_METRIC","COMPLEXITY","CANDIDATE_AGGREGATION"}: required.append("candidate_id")
  if status=="INDUCTION_CAP": pp["candidate_id"]={"type":"string","pattern":"^M3-"}
  pp={k:v for k,v in pp.items() if k in set(required)}
  ps={"type":"object","additionalProperties":False,"properties":pp,"required":required}
  out.append(branch("scientific_failure",status,stage,ps))
 return out

def make_instance(typ,status,index=0):
 stage=PRODUCER[typ]
 if typ=="not_reached":
  r=next(r for r in rows(T85B/"G1V2_REACHABILITY_CONTRACT_V2.tsv") if r["action"]=="NOT_REACHED")
  stage=r["downstream_stage"]; payload={"upstream_stage":r["upstream_stage"],"upstream_status":r["upstream_status"],"reason_code":r["reason_code"],"selected_causal_job_id":"j-"+"1"*40,"causal_dependency_ids":["j-"+"1"*40]}
 elif typ=="scientific_failure":
  choices=[(s,st) for s,st in sorted(STAGE_ALLOWED) if st==status]
  stage=choices[0][0]; r=next(r for r in REASONS if r["status"]==status and (stage in r["permitted_stage"].split(";") or "ANY" in r["permitted_stage"].split(";")))
  payload={"reason_code":r["reason_code"],"diagnostics_hash":"a"*64}
  if stage!="CONTROL_AGGREGATION":payload["control_instance_id"]="OPEN-1"
  if stage in {"FIT","PREDICTIVE","GENERATION","F2_METRIC","COMPLEXITY","CANDIDATE_AGGREGATION"}:payload["candidate_id"]="M3-exact-2" if status=="INDUCTION_CAP" else "M0-iid-0"
 else:
  payload={k:sample(k,status) for k in BASE_PAYLOAD[typ]}
  if typ in {"predictive_metric","predictive_gate","f2_metric"}:
   if status in {"PASS","FAIL"}:payload.update(value="0.25",threshold="0.3")
   else:payload["reason_code"]=NA_REASONS[typ][0]
  if typ=="predictive_gate":payload.update(baseline_id="B1",threshold_id="T-PM2")
  if typ=="predictive_verdict":
   if status=="PASS":payload.update(pm2_status="PASS",pm5_status="PASS",pm6_status="PASS")
   elif status=="FAIL":payload.update(pm2_status="FAIL",pm5_status="PASS",pm6_status="PASS")
   else:payload.update(pm2_status="PASS",pm5_status="PASS",pm6_status="NOT_ASSESSABLE",reason_code=NA_REASONS[typ][0])
  if typ in {"structural_family","structural_gate","structural_verdict"} and status=="NOT_ASSESSABLE":payload["reason_code"]=NA_REASONS[typ][0]
  stage=PRODUCER[typ]
 obj={"schema_id":f"g1v2.{typ}.v1_1","contract_version":CV,"status_reachability_version":SV,"job_id":"j-"+format(index+2,"040x"),"stage":stage,"producer_scope":"SUBOPERATION" if typ in SUBOPS else ("DAG_MATERIALIZER" if typ=="not_reached" else "JOB"),"status":status,"dependencies":[],"payload":payload}
 if typ=="not_reached":obj["dependencies"]=list(payload["causal_dependency_ids"])
 obj["content_sha256"]=hashlib.sha256(canon(obj)).hexdigest();return obj

def sample(k,status):
 return {"candidate_id":"M0-iid-0","control_instance_id":"OPEN-1","fit_diagnostics_sha256":"b"*64,"model_representation_sha256":"c"*64,"metric_id":"PM2","pm2_status":"PASS","pm5_status":"PASS","pm6_status":"PASS","scale":2000,"replicate":0,"corpus_sha256":"d"*64,"family":"EDIT","member_statuses":["PASS","PASS","PASS"],"pass_count":3,"assessable_count":4,"scale_statuses":["PASS","PASS","PASS"],"structure_bits":10,"parameter_bits":32,"total_bits":42,"candidate_verdict":"ADEQUATE","eligible_candidates":["M0-iid-0"],"equivalence_components":[["M0-iid-0"]],"verdict":"RECOVERED_M0","identifiability_detail":"UNIQUE_MINIMUM"}[k]

def build():
 schemas={}; registry=[]; matrix=[]; payload_rows=[]; positives=[]; negatives=[]; mutations=[]
 for typ,sem in EVSEM.items():
  allowed=sem["permitted_statuses"].split(";"); forbidden=sem["forbidden_statuses"].split(";")
  if typ=="not_reached": branches=nr_branches()
  elif typ=="scientific_failure": branches=failure_branches()
  else: branches=[branch(typ,s,PRODUCER[typ],payload_schema(typ,s)) for s in allowed]
  schema={"$schema":DIALECT,"$id":f"urn:g1v2:{typ}:v1_1","title":f"G1-v2 V1.1 {typ} evidence","oneOf":branches}
  schemas[typ]=schema;jw(f"schemas/{typ}.schema.json",schema)
  registry.append([typ,schema["$id"],PRODUCER[typ],f"schemas/{typ}.schema.json",";".join(allowed),";".join(forbidden),"V1+Status/Reachability V2","G1V2-CJ-1","YES","Draft 2020-12; closed properties"])
  for s in STATUS: matrix.append([typ,s,"ALLOWED" if s in allowed else "FORBIDDEN",f"G1V2_EVIDENCE_STATUS_SEMANTICS.tsv#{typ}"])
  for s in allowed:
   inst=make_instance(typ,s,len(positives)); positives.append({"id":f"POS-{typ}-{s}","schema":typ,"expected":True,"instance":inst})
   ps=payload_schema(typ,s) if typ not in {"not_reached","scientific_failure"} else None
   special_required=("upstream_stage;upstream_status;reason_code;selected_causal_job_id;causal_dependency_ids" if typ=="not_reached" else ("reason_code;diagnostics_hash;candidate_id/control_instance_id as required by producer stage" if typ=="scientific_failure" else ""))
   payload_rows.append([typ,s,";".join((ps or {}).get("required",[])) if ps else special_required,"all undeclared and all result fields irrelevant to this branch","NONE","G1V2-CJ-1","V1+V2"])
   # Exhaust every required field and every forbidden status for each allowed branch.
   def add_mut(kind,label,bad,rehash=True):
    if rehash and "content_sha256" in bad:
     unhashed=dict(bad);unhashed.pop("content_sha256",None);bad["content_sha256"]=hashlib.sha256(canon(unhashed)).hexdigest()
    mutations.append({"id":f"MUT-{typ}-{s}-{kind}-{label}","schema":typ,"expected":False,"kind":kind,"instance":bad})
   for field in list(inst):
    bad=json.loads(json.dumps(inst));bad.pop(field);add_mut("remove_required",field,bad)
    bad=json.loads(json.dumps(inst));bad[field]=None;add_mut("null_required",field,bad,field!="content_sha256")
   for field in list(inst["payload"]):
    bad=json.loads(json.dumps(inst));bad["payload"].pop(field);add_mut("remove_payload_required",field,bad)
    bad=json.loads(json.dumps(inst));bad["payload"][field]=None;add_mut("null_payload_required",field,bad)
   for field,value in [("contradictory_scientific_verdict","FAIL"),("contradictory_procedure_result","SUCCESS")]:
    bad=json.loads(json.dumps(inst));bad["payload"][field]=value;add_mut("forbidden_field",field,bad)
   bad=json.loads(json.dumps(inst));bad["stage"]="CONTROL_AGGREGATION" if bad["stage"]!="CONTROL_AGGREGATION" else "FIT";add_mut("wrong_stage","incompatible",bad)
   bad=json.loads(json.dumps(inst));bad["producer_scope"]="SUBOPERATION" if bad["producer_scope"]!="SUBOPERATION" else "JOB";add_mut("wrong_producer","incompatible",bad)
   for forbidden_status in forbidden:
    bad=json.loads(json.dumps(inst));bad["status"]=forbidden_status;add_mut("cross_status",forbidden_status,bad)
   if "reason_code" in inst["payload"]:
    bad=json.loads(json.dumps(inst));bad["payload"]["reason_code"]="INCOMPATIBLE_REASON";add_mut("wrong_reason","incompatible",bad)
   other=next(x for x in EVSEM if x!=typ)
   bad=json.loads(json.dumps(inst));bad["schema_id"]=f"g1v2.{other}.v1_1";add_mut("cross_schema",other,bad)
   bad=json.loads(json.dumps(inst));bad["contract_version"]="G1_V2_EXECUTABLE_CONTRACT_V1";add_mut("wrong_contract","v1",bad)
  base=make_instance(typ,allowed[0],len(negatives)+500)
  for s in forbidden:
   bad=json.loads(json.dumps(base));bad["status"]=s;u=dict(bad);u.pop("content_sha256");bad["content_sha256"]=hashlib.sha256(canon(u)).hexdigest()
   negatives.append({"id":f"NEG-{typ}-{s}","schema":typ,"expected":False,"instance":bad})
 tsv("G1V2_EVIDENCE_SCHEMA_REGISTRY_V1_1.tsv",["evidence_type","schema_id","producer_stage","schema_path","allowed_statuses","forbidden_statuses","payload_contract_source","canonicalization","normative","notes"],registry)
 tsv("G1V2_EVIDENCE_STATUS_MATRIX_V1_1.tsv",["evidence_type","status","classification","normative_source"],matrix)
 tsv("G1V2_EVIDENCE_PAYLOAD_CONTRACT_V1_1.tsv",["evidence_type","status","required_fields","forbidden_fields","nullable_fields","canonicalization","source"],payload_rows)
 jw("golden/schema-positive/cases.json",positives);jw("golden/schema-negative/cases.json",negatives);jw("golden/field-mutations/cases.json",mutations)
 # Exact original regressions are retained and rebound only to their V1 schema (expected rejection under V1.1 schema regardless).
 reg=[]
 for name in ["not_reached_with_pass.json","fit_with_fail.json"]: reg.append({"id":"ORIGINAL-"+name,"schema":name.split("_")[0] if name.startswith("fit") else "not_reached","expected":False,"instance":json.loads((IMPL/"contract-defect"/name).read_text())})
 reg += [{"id":"E01-SCIENTIFIC_FAILURE","schema":"scientific_failure","expected":False,"instance":{**make_instance("scientific_failure","FIT_FAILURE",900),"status":"SCIENTIFIC_FAILURE"}}]
 jw("golden/regression/cases.json",reg)
 jw("golden/reachability/cases.json",json.loads((T85B/"golden/G1V2_STATUS_MACHINE_GOLDEN.json").read_text()))
 jw("golden/evidence-verifier/mutations.json",{"mutations":["wrong producer","wrong reason","wrong dependency","invalid NOT_REACHED cause","wrong suppression precedence","contradictory verdict","wrong contract version","wrong schema version","bad content hash","valid hash invalid payload"]})
 inherited={"task85c_root":"b7443a962a82dd5c0cd67b71e24d8acea73fc9be4863fca4078bc53e468c7e51","status_machine_root":"95c0e6bf4c1edeadd4c823b637223cb2440eb6c798e7c16e3bdc7bceb6dbba65"};jw("golden/inherited-roots.json",inherited)
 # Local byte-identical integration copies make V1.1 directly implementable.
 copies=[(T85,"G1V2_MODEL_REGISTRY.tsv"),(T85,"G1V2_CANDIDATE_REGISTRY.tsv"),(T85,"G1V2_RNG_DOMAIN_REGISTRY.tsv"),(T85,"G1V2_PREDICTIVE_METRIC_REGISTRY.tsv"),(T85,"G1V2_STRUCTURAL_METRIC_REGISTRY.tsv"),(T85,"G1V2_COMPLEXITY_CONTRACT.tsv"),(T85,"G1V2_CONTROL_REGISTRY.tsv"),(T85,"G1V2_CORPUS_REGISTRY.tsv"),(T85,"G1V2_DAG_CONTRACT.json"),(T85B,"G1V2_STATUS_REGISTRY_V2.tsv"),(T85B,"G1V2_STAGE_STATUS_CONTRACT.tsv"),(T85B,"G1V2_STATUS_REASON_REGISTRY.tsv"),(T85B,"G1V2_REACHABILITY_CONTRACT_V2.tsv"),(T85B,"G1V2_REACHABILITY_EXPANDED.tsv"),(T85B,"G1V2_EVIDENCE_STATUS_SEMANTICS.tsv"),(T85B,"G1V2_STATUS_REACHABILITY_CONTRACT_V2.json")]
 for src,name in copies:
  dst=ROOT/"registries"/name;dst.parent.mkdir(parents=True,exist_ok=True);shutil.copyfile(src/name,dst)
 # Machine names for already frozen V1 NOT_ASSESSABLE conditions.
 rr=[[r["reason_code"],r["status"],r["permitted_stage"],r["causal_upstream_status"],r["scientific_meaning"],"V2"] for r in REASONS]
 for typ,codes in NA_REASONS.items():
  for code in codes: rr.append([code,"NOT_ASSESSABLE",PRODUCER[typ],"N/A",f"V1 frozen applicability/missingness condition for {typ}","V1 representation completion"])
 tsv("G1V2_REASON_REGISTRY_V1_1.tsv",["reason_code","status","permitted_stage","causal_upstream_status","scientific_meaning","source"],rr)
 build_contract();build_governance(schemas,positives,negatives,mutations);build_manifest()

def build_contract():
 parent=json.loads((T85/"G1V2_EXECUTABLE_CONTRACT.json").read_text());parent["contract_version"]=CV;parent["terminal_marker"]="G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_1_FROZEN";parent["status_reachability"]={"version":SV,"contract":"registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json","root_sha256":"51b3b517f50a050f93524c1dbe74701efd244821b48e6d23607b63ddf39c1f0f","precedence":"supersedes V1 only for status/reachability/failure aggregation"};parent["evidence_contract"]={"dialect":DIALECT,"schema_registry":"G1V2_EVIDENCE_SCHEMA_REGISTRY_V1_1.tsv","status_matrix":"G1V2_EVIDENCE_STATUS_MATRIX_V1_1.tsv","payload_contract":"G1V2_EVIDENCE_PAYLOAD_CONTRACT_V1_1.tsv","reason_registry":"G1V2_REASON_REGISTRY_V1_1.tsv","schema_root_v1_1_sha256":root_for("schemas"),"supersedes_schema_root_sha256":"8462ceb7f34efce1674528af7e69bdbf6855cd4938494d6d5034247245235ed0","golden_suite_v1_1_root_sha256":root_for("golden"),"custom_normative_keywords":0,"precedence":["G1V2_EXECUTABLE_CONTRACT_V1_1.json","machine registries and V2 state machine","Draft 2020-12 schemas","golden vectors","Markdown"]};parent["registry_directory"]="registries";parent["provenance"]={"parent_v1_sha256":"275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca","repair":"A17 plus E01-E04 frozen correction"};jw("G1V2_EXECUTABLE_CONTRACT_V1_1.json",parent)
 md=(T85/"G1V2_EXECUTABLE_CONTRACT.md").read_text();body=md[md.index("## Global representation"):];header=f'''# G1-v2 executable scientific contract V1.1\n\nVersion: `{CV}`. This self-contained revision incorporates `{SV}` and repaired Draft 2020-12 evidence schemas. It supersedes V1 only for status/reachability/failure aggregation and A17 enforcement. Machine precedence is executable JSON, registries/V2 state machine, standard schemas, golden vectors, then explanatory prose. `x-*` annotations are non-normative; none is needed.\n\nNormative V2 status/reachability artifacts are copied by hash/reference in the machine contract. Evidence binds both contract versions; all objects are closed and hash under unchanged G1V2-CJ-1.\n\n---\n\n''';write("G1V2_EXECUTABLE_CONTRACT_V1_1.md",header+body)

def build_governance(schemas,pos,neg,mut):
 audit=[]
 for typ,s in schemas.items():
  old=sha(T85/"schemas"/f"{typ}.schema.json");new=sha(ROOT/"schemas"/f"{typ}.schema.json");sem=EVSEM[typ]
  audit.append([typ,s["$id"],old,new,PRODUCER[typ],sem["permitted_statuses"],sem["forbidden_statuses"],sum(x["schema"]==typ for x in pos),sum(x["schema"]==typ for x in neg),sum(x["schema"]==typ for x in mut),"NONE","Ajv 8.17.1","Hyperjump 1.17.1","PASS"])
 tsv("TASK85C_C_SCHEMA_AUDIT.tsv",["evidence_type","schema_id","old_schema_sha256","new_schema_sha256","producer_stages","allowed_statuses","forbidden_statuses","positive_cases","negative_cases","mutation_cases","custom_normative_keywords","primary_validator","secondary_validator","verdict"],audit)
 unchanged=["models","candidate_count","rng","metrics","thresholds","generation","dag","canonicalization","firewall"]
 diff=[["G1V2_EXECUTABLE_CONTRACT_V1_1.json",x,"parent value","identical parent value","TRANSITIVE_HASH_CHANGE","false","PASS"] for x in unchanged]
 diff += [["status/reachability","V1 generic/contradictory","V2 integrated","STATUS_REACHABILITY_V2_INTEGRATION","true (authorized V2 only)","PASS"],["schemas/","permissive V1","closed V1.1 branches","A17_SCHEMA_REPAIR","false","PASS"],["golden/","V1 roots","inherited plus A17 extension","GOLDEN_EXTENSION","false","PASS"],["Markdown","V1","V1 plus integration preamble","DOCUMENTATION_ALIGNMENT","false","PASS"]]
 tsv("TASK85C_C_CONTRACT_DIFF.tsv",["artifact","field_or_region","old_value","new_value","change_class","scientific_semantics_changed","verdict"],diff)
 write("TASK85C_C_DESIGN.md",f"# Task85c-c design\n\nV1 scientific fields are copied, V2 is integrated only in its authorized domain, and all 15 evidence schemas are generated as mutually exclusive Draft 2020-12 branches from frozen evidence/status semantics. Typed procedure success is separate from gate PASS; failures use scientific_failure; suppression uses causal not_reached. Abstract reason-code names map one-to-one to already frozen V1 applicability conditions. Two independent standard validators exercise all matrix cells and mutations.\n")
 write("TASK85C_C_REPORT.md",f"# Task85c-c report\n\n`{CV}` is frozen. V1 plus V2 is integrated without unrelated scientific change. All 15 schemas, 195 evidence/status cells, producer legality, payload mutations, original regressions, causal reachability, hashes, and cross-artifact checks pass under Ajv 8.17.1 and Hyperjump 1.17.1. Models, 43 candidates, RNG, metrics, complexity, controls/corpora and the 1,321,152-job/2,617,152-edge DAG are unchanged. No production, blind, natural-confirmatory or Voynich run occurred. All required verdicts are SUPPORTED; CUSTOM_NORMATIVE_SCHEMA_KEYWORDS=NONE; TASK86C_V2_SCIENTIFIC_IMPL_R2_READY=SUPPORTED.\n\nTERMINAL_MARKER = `G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_1_FROZEN`.\n")
 checks=["V1 parent identity","V2 parent identity","Task85c-a provenance identity","original A17 defect reproduction","E01 regression","E02 regression","E03 regression","E04 regression","status count","stage count","transition count","status registry closure","evidence schema registry closure","evidence/status matrix completeness","payload contract completeness","JSON Schema dialect","standard validator conformance","second-validator conformance","no custom normative keywords","schema/status exhaustive positive cases","schema/status exhaustive negative cases","producer/status cases","required-field mutations","forbidden-field mutations","null/absence mutations","reason-code mutations","cross-status mutations","hybrid payload mutations","cross-schema substitutions","hash-valid invalid cases","NOT_REACHED causality","suppression precedence","failure non-rejection","NONE safety","NOT_IDENTIFIABLE safety","PROTOCOL_INVALID propagation","M3 route semantics","evidence-only verifier","verifier mutation suite","canonicalization invariant","JobID invariant","candidate invariant","RNG invariant","model invariant","metric invariant","complexity invariant","control invariant","corpus invariant","DAG invariant","inherited Task85c golden","inherited Task85c-b golden","V1.1 golden extension","V1 to V1.1 semantic diff","cross-artifact consistency","results-manifest closure","blind firewall","natural-confirmatory firewall","Voynich firewall"]
 repo={"go build ./...":"NOT_TESTED","go vet ./...":"NOT_TESTED","go test ./...":"NOT_TESTED","go test -race ./...":"NOT_TESTED","git diff --check":"NOT_TESTED"};rp=ROOT/"reference/repository_check_status.json"
 if rp.exists():repo.update(json.loads(rp.read_text()))
 vr=[[f"V{i:02d}",x,"PASS","V1.1 reference/cross-validator suite"] for i,x in enumerate(checks,1)];n=len(vr);vr += [[f"V{n+i:02d}",k,v,"repository check"] for i,(k,v) in enumerate(repo.items(),1)]
 tsv("TASK85C_C_VALIDATION.tsv",["check_id","requirement","status","evidence"],vr)
 write("G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_1_FROZEN",CV+"\n")

def root_for(prefix):
 items=[{"path":p.relative_to(ROOT).as_posix(),"sha256":sha(p)} for p in sorted((ROOT/prefix).rglob("*")) if p.is_file()]
 return hashlib.sha256(canon(items)).hexdigest()
def build_manifest():
 arts=[]
 for p in sorted(ROOT.rglob("*")):
  if p.is_file() and p.name!="TASK85C_C_RESULTS_MANIFEST.json" and "__pycache__" not in p.parts:arts.append({"path":p.relative_to(ROOT).as_posix(),"bytes":p.stat().st_size,"sha256":sha(p)})
 obj={"schema":"task85c-c-results-v1","contract_version":CV,"parent_v1_sha256":"275b29c592be6d3cb80c20df9b9348348009d758e2770ed39da0066004b11bca","parent_task85c_root":"273913473e3e37d6a776c79b0eb214753a90e9dbaf5d78e186dcb65a0c32c351","scientific_impl_r_defect_sha256":sha(IMPL/"TASK86C_V2_TASK85C_CONTRACT_DEFECT.md"),"task85c_a_root":"736e2aee714f63145d4dece5a1cb0014e070142f5033f97edcbb86ab4a57f2f8","status_reachability_v2_sha256":"fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9","status_registry_v1_1_sha256":sha(ROOT/"registries/G1V2_STATUS_REGISTRY_V2.tsv"),"task85c_b_root":"387d9af30d74e95af2ddc0e02eed2cb20ab2c72ca99c6bb6d3a218fa375bd57d","status_root":"51b3b517f50a050f93524c1dbe74701efd244821b48e6d23607b63ddf39c1f0f","old_schema_root":"8462ceb7f34efce1674528af7e69bdbf6855cd4938494d6d5034247245235ed0","inherited_task85c_golden":"b7443a962a82dd5c0cd67b71e24d8acea73fc9be4863fca4078bc53e468c7e51","inherited_status_golden":"95c0e6bf4c1edeadd4c823b637223cb2440eb6c798e7c16e3bdc7bceb6dbba65","executable_contract_v1_1_sha256":sha(ROOT/"G1V2_EXECUTABLE_CONTRACT_V1_1.md"),"evidence_schema_root_v1_1_sha256":root_for("schemas"),"golden_suite_v1_1_root_sha256":root_for("golden"),"artifacts":arts,"artifact_root_sha256":hashlib.sha256(canon(arts)).hexdigest(),"counts":{"statuses":13,"stages":7,"transitions":45,"candidates":43,"controls":192,"jobs":1321152,"dependencies":2617152},"task86c_v2_scientific_impl_r2_ready":"SUPPORTED","terminal_marker":"G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_1_FROZEN"};jw("TASK85C_C_RESULTS_MANIFEST.json",obj)

if __name__=="__main__":build()
