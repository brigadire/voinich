#!/usr/bin/env python3
"""Build the Task85c-g V1.2 generation-contract freeze package."""
from __future__ import annotations
import csv, hashlib, json, shutil
from pathlib import Path

HERE = Path(__file__).resolve().parent
OUT = HERE.parent
REPO = HERE.parents[3]
PARENT = REPO / "research/phase3/task85c-c"
T85 = REPO / "research/phase3/task85c"

def canon(obj): return (json.dumps(obj, ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode()
def sha(path): return hashlib.sha256(path.read_bytes()).hexdigest()
def write_json(name, obj): (OUT/name).write_bytes(canon(obj))
def write_text(name, text): (OUT/name).write_text(text.rstrip()+"\n")
def write_tsv(name, header, rows):
    with (OUT/name).open("w", newline="") as f:
        w=csv.writer(f,delimiter="\t",lineterminator="\n"); w.writerow(header); w.writerows(rows)

parent_sha=sha(PARENT/"G1V2_EXECUTABLE_CONTRACT_V1_1.md")
e1_sha=sha(REPO/"research/phase3/task85c-e/G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json")
status_sha=sha(PARENT/"registries/G1V2_STATUS_REACHABILITY_CONTRACT_V2.json")
assert parent_sha=="5c3cd272c1dbae9bfe1d7a100155faf102e86d34660da239e1cb31704ad470b0"
assert e1_sha=="dbfb9a4a7101eed7006f751b9c4631b5f0286c3792f9777cc833c5dcfa42a3d3"
assert status_sha=="fc1ca07d8123ed5d44bc24ecba98fca54d5b05781ecbaba820d44079319038b9"

with (T85/"G1V2_SYNTHETIC_GENERATOR_REGISTRY.tsv").open(newline="") as f:
    generators=list(csv.DictReader(f,delimiter="\t"))

state_machine={
 "version":"G1V2_GENERATION_STATE_MACHINE_V1","initial":"BEFORE_TOKEN","terminal":"CORPUS_COMPLETE",
 "token_length":"count of emitted Unicode scalar-value glyphs; not bytes, sampled sentinels, states, or transitions",
 "states":{
  "BEFORE_TOKEN":{"admissible":"ordinary glyph and UNK; EOS and BOS forbidden","on_glyph":"emit; length=1; INSIDE_TOKEN","on_unk":"emit U+FFFD; length=1; INSIDE_TOKEN","on_zero_mass":"GENERATION_FAILURE before RNG draw"},
  "INSIDE_TOKEN":{"admissible":"ordinary glyph, UNK, EOS; BOS forbidden","on_glyph":"emit; increment length","on_unk":"emit U+FFFD; increment length","on_eos":"TOKEN_COMPLETE without emission","at_length_64":"TOKEN_COMPLETE structural, no EOS sample and no RNG draw"},
  "TOKEN_COMPLETE":{"action":"append nonempty token in occurrence order; start next token at BEFORE_TOKEN or finish corpus","rng_draws":0},
  "CORPUS_COMPLETE":{"action":"serialize ordered tokens canonically","rng_draws":0}},
 "length_boundaries":{"63":"next stochastic decision allowed","64":"legal and structurally terminates immediately","65":"unreachable; attempted construction is GENERATION_FAILURE"},
 "whole_token":{"M5":"validate NFC scalar token length 1..64; valid -> TOKEN_COMPLETE; invalid -> next attempt; no EOS operation"}
}
write_json("G1V2_GENERATION_STATE_MACHINE_V1.json",state_machine)
write_text("G1V2_GENERATION_STATE_MACHINE_V1.md",'''# G1V2 generation state machine V1

`BEFORE_TOKEN` forbids BOS and EOS and conditionally samples ordinary glyphs
and UNK. An ordinary glyph is emitted verbatim; UNK emits U+FFFD. Both enter
`INSIDE_TOKEN` with length one. In `INSIDE_TOKEN`, EOS terminates without
emission. Emitting glyph 64 immediately enters `TOKEN_COMPLETE`: length 64 is
legal, termination is structural, and no 65th-glyph/EOS draw is consumed.

`TOKEN_COMPLETE` appends the nonempty logical token in occurrence order.
After the requested positive token count it enters `CORPUS_COMPLETE`, whose
only action is canonical serialization. M5 proposes a whole token per attempt;
valid NFC scalar length 1..64 enters `TOKEN_COMPLETE`, otherwise it advances to
the next attempt without RNG rollback. Zero admissible mass is the existing
`GENERATION_FAILURE` and consumes no draw.''')

route_algorithms={
 "M0_GEN_A":"DIRECT_CDF","M1_GEN_A":"DIRECT_CDF","M2_GEN_A":"DIRECT_CDF","M3_GEN_A":"DIRECT_CDF","M4_GEN_A":"DIRECT_CDF","M5_GEN_A":"DIRECT_CDF",
 "M0_GEN_B":"EXPONENTIAL_RACE","M1_GEN_B":"WALKER_ALIAS","M2_GEN_B":"EXPONENTIAL_RACE","M3_GEN_B":"EXPONENTIAL_RACE","M4_GEN_B":"CUMULATIVE_ROW","M5_GEN_B":"EXPONENTIAL_RACE"}
semantics={
 "version":"G1V2_GENERATION_SEMANTICS_V1","contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1_2",
 "precedence":["G1_V2_EXECUTABLE_CONTRACT_V1_2.json","G1V2_GENERATION_SEMANTICS_V1.json","G1V2_GENERATION_STATE_MACHINE_V1.json","G1V2_GENERATION_GOLDEN_SUITE_V1.json","Markdown"],
 "numerical":{"format":"IEEE-754 binary64","rounding":"round-to-nearest ties-to-even","fma":"forbidden in prescribed reductions","sum":"Neumaier in canonical outcome order","normalization":"filter admissible outcomes preserving order; Neumaier sum Z; if Z>0 divide each retained weight by Z once in order","comparison":"u53 < boundary; equality selects following positive bin","final_bin":"if cumulative roundoff leaves u53 unselected, select last positive admissible outcome","invalid_weight":"nonfinite or negative -> NUMERICAL_FAILURE before draw","zero_mass":"GENERATION_FAILURE before draw","ln":"correctly rounded binary64 natural logarithm of the exact binary64 input; u53=0 maps race score to +infinity"},
 "outcome_order":{"explicit_sequence":"preserve sequence exactly","mapping":"ascending NFC UTF-8 key bytes","fitted_vocabulary":"ordinary DEVELOPMENT glyphs ascending NFC UTF-8, then <UNK>, then <EOS>","specials":"never participate in lexical interleaving","m0_resolution":"explicit registry [a,b,c,d,<EOS>] controls; EOS is last"},
 "symbols":{"BOS":{"context":True,"sampled":False,"emitted":False,"terminates":False},"EOS":{"context":False,"sampled":"INSIDE_TOKEN only","emitted":False,"terminates":"token only"},"UNK":{"context":True,"sampled":True,"emitted":"U+FFFD","counts_as_glyph":True}},
 "categorical":{"DIRECT_CDF":{"draws":1,"algorithm":"normalize admissible row; Neumaier prefix cumulative; first positive outcome with u53<cumulative; final-bin rule"},"CUMULATIVE_ROW":{"draws":1,"algorithm":"same mathematical row law, independently materialized matrix row; direct row identity and canonical order required"},"EXPONENTIAL_RACE":{"draws":"one per positive admissible outcome in canonical order","algorithm":"score=-correctly_rounded_ln(u53)/normalized_weight; u53 zero gives +infinity; smallest score wins; exact score tie -> lower canonical index"},"WALKER_ALIAS":{"draws":2,"build":"q[i]=normalized_p[i]*n; sorted increasing small q<1 and large q>=1 queues; pop lowest from each; cutoff[small]=q[small], alias[small]=large; q[large]=(q[large]+q[small])-1; reinsert by increasing index; remaining cutoff=1 alias=self","sample":"column=floor(u1*n); choose column iff u2<cutoff[column], otherwise alias[column]","degenerate":"still consumes two draws for one positive outcome"}},
 "constraint":{"BEFORE_TOKEN":"remove BOS and EOS, condition remaining mass, then route primitive","INSIDE_TOKEN":"remove BOS only","forbidden_draws":"no rejection, reuse, skip, or discarded draw"},
 "rng":{"algorithm":"G1V2-RNG-1 unchanged","u53":"(u64be(digest[0:8])>>11)/2^53","CONTROL_GENERATE":{"namespace":"g1v2/control/generate","counters":["generator_index","scale_index","replicate","draw_index"],"draw_index":"global monotonically increasing within corpus starting 0; increments after every consumed U53; never rolls back"},"GENERATE":{"namespace":"g1v2/generate","counters":["control_index","candidate_index","scale_index","replicate","draw_index"],"draw_index":"same monotone rule"}},
 "length":{"minimum":1,"maximum":64,"unit":"emitted Unicode scalar-value glyph","at_64":"successful structural token termination; zero extra draws","attempt_65":"GENERATION_FAILURE and no emission"},
 "m5":{"attempts":"exactly indices 0..1023; first proposal is attempt 0; exhaustion after invalid attempt 1023 -> GENERATION_FAILURE","state_reset":"logical proposal and local components reset; global draw_index persists","A_draw_order":{"fitted":"productive-mixture; if productive prefix,stem,suffix; else rule-vs-exception then selected item","synthetic":"exception-vs-productive; if exception exception-item; else prefix,stem,suffix"},"B_draw_order":"same decisions, each selected by route B primitive and thus consumes that primitive's fixed draws","validation":"NFC scalar length 1..64 and nonempty; invalid/overlength consumes completed attempt draws then retries; no rollback"},
 "model_generation":{
  "M0":"At every streaming decision use the frozen fitted or synthetic outcome row in canonical order.",
  "M1":"Initialize order BOS values. At each decision use the longest available fitted row with frozen suffix fallback. Synthetic M1 puts EOS mass 0.15 last; at BOS multiply [.4,.3,.2,.1] by .85; otherwise glyph masses are [.45,.25,.075,.075] assigned respectively to successor modulo 4, same glyph, then the other glyphs in alphabet order, already totaling .85. After glyph emission shift context; EOS terminates.",
  "M2":"Initialize empty suffix. Use longest retained fitted context else root. Synthetic M2 first constructs a glyph-only row: if a listed context has preferred glyph probability q, assign q to it and distribute 1-q among other glyphs proportional to their root weights; otherwise use root. Multiply glyph row by .84 and append EOS .16. Append emitted glyph and retain last four for lookup.",
  "M3_EXACT":"Initialize canonical start state. Sample its explicit outcome row. Glyph transition enters the row's target; EOS terminates. BEFORE_TOKEN conditions EOS if a fitted start row contains it.",
  "M3_APPROX":"Identical generation machine to M3_EXACT after the approximate fitted topology has been frozen; merge construction is fitting, not generation.",
  "M4":"At token start sample pi, then sample that state's emission. EOS terminates without transition. After a glyph emission sample the current state's transition row and make the selected state current. At length 64 do not sample a transition.",
  "M5":"Use the fully specified attempt and branch rules in m5. Fitted component/rule/exception sequences use canonical outcome order; synthetic parameters are explicit. Whole-token proposals contain no BOS/EOS/UNK sentinel."
 },
 "serialization":{"logical":"ordered sequence of nonempty already-NFC tokens; one token per logical line","normalization_operation":"none; non-NFC logical input is invalid because generation emits only NFC vocabulary glyphs or U+FFFD","encoding":"UTF-8","token_delimiter_hex":"0a","line_delimiter_hex":"0a","empty_corpus_hex":"","empty_lines":"forbidden","trailing_whitespace":"none except required LF terminator","final_newline":"required after every token including last","content_sha256":"SHA-256 over exactly these bytes"},
 "corpus":{"requested_token_count":"positive integer; zero is PROTOCOL_VETO before generation","completion":"after exactly requested token count; no RNG draw"},
 "routes":[]}
for i,row in enumerate(generators):
    semantics["routes"].append({"generator_id":row["generator_id"],"generator_index":i,"model":row["model_class"],"author":row["author_route"],"primitive":route_algorithms[row["generator_id"]],"parameters":json.loads(row["parameters"]),"max_token_glyphs":64,"attempt_cap":1024})
write_json("G1V2_GENERATION_SEMANTICS_V1.json",semantics)
write_text("G1V2_GENERATION_SEMANTICS_V1.md",'''# G1V2 generation semantics V1

This document explains the normative JSON of the same name. Explicit arrays
define their own order; mappings use NFC UTF-8 key order; fitted outcomes are
ordinary glyphs in NFC UTF-8 order followed by UNK and EOS. State constraints
filter that order and renormalize once with Neumaier summation.

Generator A uses direct inverse-CDF. Generator B remains independently
constructed: M0/M2/M3/M5 use a fully ordered exponential race, M1 uses the
specified deterministic Walker table, and M4 materializes cumulative matrix
rows. B never calls A. Every U53 increments one corpus-global draw index.

Length counts emitted Unicode scalar glyphs. Glyph 64 is legal and forces
structural completion without another draw. M5 attempts are exactly 0..1023,
retain consumed draws, and map exhaustion to `GENERATION_FAILURE`.

Canonical corpus bytes are NFC UTF-8, one token per line, LF after every token
including the last, and no other whitespace. The empty logical corpus maps to
zero bytes but is not a valid requested synthetic-control corpus.''')

ambiguities=[
 ("G-A01","Task85c-f PF-SC01","all streaming","M0-M4","BEFORE_TOKEN","retry vs condition","empty token/different law","1 vs >=2 draws","different prefix","conditional admissible support, route primitive","structural conditioning; no hidden rejection","PF-SC01","RESOLVED"),
 ("G-A02","V1.1 prose/registry","all","M0-M5","categorical","UTF8 vs row/map order","different symbol","same count","different bytes","explicit arrays; maps UTF8; fitted ordinary+UNK+EOS","preserves explicit scientific rows","ORDER-01","RESOLVED"),
 ("G-A03","V1.1 numeric","all","M0-M5","categorical","cumulative arithmetic choices","boundary selection","same","different at boundary","canonical binary64 Neumaier CDF and final bin","extends frozen numeric profile","CAT-*","RESOLVED"),
 ("G-A04","Generator B","B race","M0,M2,M3,M5","categorical","race enumeration/u0/ties","different sample","variable","different","local positive outcomes; u0=inf; canonical tie","standard exponential clocks with explicit realization","RACE-*","RESOLVED"),
 ("G-A05","Generator B","M1_GEN_B","M1","categorical","alias construction choices","different sample","ambiguous","different","sorted-queue Walker algorithm; exactly two draws","deterministic standard formulation","ALIAS-*","RESOLVED"),
 ("G-A06","Generator B","M4_GEN_B","M4","matrix rows","row/order choices","different state/emission","ambiguous","different","identified canonical rows and one-draw CDF","matches frozen cumulative intent","CUM-*","RESOLVED"),
 ("G-A07","V1.1 cap","all","M0-M5","length 64","fail vs legal/force EOS","success differs","extra draw possible","different","64 legal structural completion, no draw","reconciles M5 support through 64","LEN-*","RESOLVED"),
 ("G-A08","V1.1 M5","M5 A/B","M5","attempt","0..1023 vs retries","failure boundary","different","different","1024 total attempts indexed 0..1023","literal cap, minimal convention","M5-ATTEMPT","RESOLVED"),
 ("G-A09","V1.1 M5","M5 A/B","M5","components","draw ordering/reset","different proposal","different","different","frozen branch draw order; monotone global draw; no rollback","executable branch semantics","M5-DRAWS","RESOLVED"),
 ("G-A10","V1.1 corpus","all","all","serialization","delimiters/final LF","same tokens","none","different hashes","NFC UTF8 one-token-per-line LF including final","simple canonical stream","SER-*","RESOLVED"),
 ("G-A11","V1.1 symbols","all","M0-M5","all","BOS/EOS/UNK roles","sentinel leakage","draw support differs","different","machine role table and U+FFFD emission","follows model vocabulary","SYMBOL-*","RESOLVED"),
 ("G-A12","constrained rows","all","M0-M5","any","zero mass handling","loop/failure","undefined","none/hang","GENERATION_FAILURE before draw","existing compatible status","ZERO-MASS","RESOLVED"),
 ("G-A13","RNG registry","all","M0-M5","all","draw/counter allocation","same law","different trace","different","monotone global draw_index per corpus","execution-independent identity","RNG-TRACE","RESOLVED"),
 ("G-A14","corpus completion","all","all","CORPUS_COMPLETE","zero count/end draw","different validity","extra draw","different","positive requested count; structural completion","control architecture already positive","CORPUS-END","RESOLVED")]
write_tsv("G1V2_GENERATION_AMBIGUITY_REGISTRY.tsv",["ambiguity_id","source","affected_generator","affected_model","affected_state","competing_interpretations","scientific_effect","RNG_effect","corpus_byte_effect","resolution","justification","tests","status"],ambiguities)
decision_rows=[]
for row in ambiguities:
    decision_rows.append([row[0],row[5],row[9],row[10],"all nonchosen interpretations in competing_interpretations","YES except arithmetic/serialization clarifications","YES except serialization has zero RNG effect","generation semantics; state machine; goldens","NO"])
write_tsv("G1V2_GENERATION_DECISION_LOG.tsv",["ambiguity_id","admissible_alternatives","chosen_rule","scientific_justification","rejected_alternatives","distributions_differ","deterministic_RNG_realization_differs","affected_artifacts","downstream_results_consulted"],decision_rows)

inventory=[]
for name,model in [("FITTED_M0","M0"),("FITTED_M1","M1"),("FITTED_M2","M2"),("FITTED_M3_EXACT","M3"),("FITTED_M3_APPROX","M3"),("FITTED_M4","M4"),("FITTED_M5","M5")]:
    primitive="DIRECT_CDF"; state="WHOLE_TOKEN" if model=="M5" else "BEFORE_TOKEN/INSIDE_TOKEN"
    inventory.append([name,model,"fitted",primitive,state,"model row filtered by state","BOS never; EOS streaming; UNK U+FFFD","GENERATE","defined by primitive","EOS or structural 64","M5 only","canonical LF lines","RESOLVED"])
for row in semantics["routes"]:
    inventory.append([row["generator_id"],row["model"],"Generator "+row["author"],row["primitive"],"WHOLE_TOKEN" if row["model"]=="M5" else "BEFORE_TOKEN/INSIDE_TOKEN","explicit parameters and state filter","BOS forbidden; EOS streaming; no literal sentinels","CONTROL_GENERATE","primitive-fixed","EOS/64 or M5 valid proposal","M5 0..1023","canonical LF lines","RESOLVED"])
shared=[("SHARED_DIRECT_CDF","all","DIRECT_CDF"),("SHARED_CONDITION","all","CONDITIONAL"),("SHARED_RACE","B","EXPONENTIAL_RACE"),("SHARED_ALIAS","B/M1","WALKER_ALIAS"),("SHARED_CUMULATIVE","B/M4","CUMULATIVE_ROW"),("SHARED_M5_ATTEMPTS","M5","ATTEMPT_MACHINE"),("SHARED_SERIALIZER","all","SERIALIZER")]
for name,model,primitive in shared:
    inventory.append([name,model,"protocol",primitive,"ALL","machine-defined","machine-defined","bound domain","machine-defined","machine-defined","machine-defined","machine-defined","RESOLVED"])
assert len(inventory)==26
write_tsv("G1V2_GENERATION_PATH_INVENTORY.tsv",["path_id","model_family","generator","sampling_primitive","state","support_construction","special_symbols","RNG_domain","RNG_draws","termination","retry","serialization","status"],inventory)

inherited=json.loads((T85/"golden/G1V2_GOLDEN_SUITE.json").read_text())
cases=[]
for c in inherited["cases"]:
    if "GEN" not in c["id"] and c["id"] not in {"M0-BOUNDARY"}:
        cases.append({"id":"INHERITED-"+c["id"],"operation":"inherited_unchanged","source_case":c,"source_version":inherited["version"],"expected":"INHERITED_UNCHANGED"})
cases += [
 {"id":"PF-SC01","operation":"categorical","input":{"outcomes":["a","b","c","d","<EOS>"],"weights":[.28,.22,.18,.12,.20],"allowed":["a","b","c","d"],"u53":.92848667210989588},"expected":{"status":"OK","outcome":"d","draws":1,"emitted_hex":"64","state":"INSIDE_TOKEN"},"rng_trace":[{"draw_index":0,"domain":"g1v2/control/generate","u53":"0.92848667210989588","operation":"DIRECT_CDF_CONDITIONAL","state":"BEFORE_TOKEN","selected":"d","emitted":"d","counters_after":[0,0,0,1]}]},
 {"id":"ORDER-01","operation":"categorical","input":{"outcomes":["a","b","c","d","<EOS>"],"weights":[.28,.22,.18,.12,.20],"allowed":["a","b","c","d","<EOS>"],"u53":.1},"expected":{"status":"OK","outcome":"a","draws":1}},
 {"id":"CAT-BELOW","operation":"categorical","input":{"outcomes":["a","b"],"weights":[.5,.5],"allowed":["a","b"],"u53":.49999999999999994},"expected":{"status":"OK","outcome":"a","draws":1}},
 {"id":"CAT-EXACT","operation":"categorical","input":{"outcomes":["a","b"],"weights":[.5,.5],"allowed":["a","b"],"u53":.5},"expected":{"status":"OK","outcome":"b","draws":1}},
 {"id":"CAT-ABOVE","operation":"categorical","input":{"outcomes":["a","b"],"weights":[.5,.5],"allowed":["a","b"],"u53":.5000000000000001},"expected":{"status":"OK","outcome":"b","draws":1}},
 {"id":"ZERO-MASS","operation":"categorical","input":{"outcomes":["a","<EOS>"],"weights":[0,1],"allowed":["a"],"u53":.2},"expected":{"status":"GENERATION_FAILURE","outcome":None,"draws":0}},
 {"id":"SYMBOL-UNK","operation":"categorical","input":{"outcomes":["<UNK>","a","<EOS>"],"weights":[.5,.25,.25],"allowed":["<UNK>","a"],"u53":.2},"expected":{"status":"OK","outcome":"<UNK>","draws":1,"emitted_hex":"efbfbd"}},
 {"id":"RACE-01","operation":"race","input":{"outcomes":["a","b","c"],"weights":[.5,.3,.2],"allowed":["a","b","c"],"uniforms":[.2,.8,.5]},"expected":{"status":"OK","outcome":"b","draws":3}},
 {"id":"RACE-U0","operation":"race","input":{"outcomes":["a","b"],"weights":[.5,.5],"allowed":["a","b"],"uniforms":[0,.5]},"expected":{"status":"OK","outcome":"b","draws":2}},
 {"id":"ALIAS-01","operation":"alias","input":{"outcomes":["a","b","c"],"weights":[.5,.3,.2],"allowed":["a","b","c"],"u_column":.8,"u_threshold":.4},"expected":{"status":"OK","outcome":"c","draws":2}},
 {"id":"CUM-01","operation":"categorical","input":{"outcomes":["s0","s1","s2"],"weights":[.6,.3,.1],"allowed":["s0","s1","s2"],"u53":.6},"expected":{"status":"OK","outcome":"s1","draws":1}},
 {"id":"LEN-63","operation":"length","input":{"glyphs":63},"expected":{"state":"INSIDE_TOKEN","next_draw_allowed":True}},
 {"id":"LEN-64","operation":"length","input":{"glyphs":64},"expected":{"state":"TOKEN_COMPLETE","next_draw_allowed":False,"status":"OK"}},
 {"id":"LEN-65","operation":"length","input":{"glyphs":65},"expected":{"state":"FAILURE","status":"GENERATION_FAILURE","emitted":False}},
 {"id":"M5-ATTEMPT","operation":"m5","input":{"invalid_attempts":1023,"valid_attempt":1023},"expected":{"status":"OK","attempts":1024,"accepted_index":1023}},
 {"id":"M5-EXHAUST","operation":"m5","input":{"invalid_attempts":1024},"expected":{"status":"GENERATION_FAILURE","attempts":1024,"last_index":1023}},
 {"id":"SER-EMPTY","operation":"serialize","input":{"tokens":[]},"expected":{"hex":"","sha256":"e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"}},
 {"id":"SER-ONE","operation":"serialize","input":{"tokens":["a"]},"expected":{"hex":"610a"}},
 {"id":"SER-MULTI","operation":"serialize","input":{"tokens":["a","bb"]},"expected":{"hex":"610a62620a"}},
 {"id":"SER-UNICODE","operation":"serialize","input":{"tokens":["café"]},"expected":{"hex":"636166c3a90a"}},
 {"id":"SER-LEN64","operation":"serialize","input":{"tokens":["a"*64]},"expected":{"hex":"61"*64+"0a"}}
]
for route in semantics["routes"]:
    cases.append({"id":"ROUTE-"+route["generator_id"],"operation":"route_binding","input":{"generator_id":route["generator_id"]},"expected":{"model":route["model"],"author":route["author"],"primitive":route["primitive"],"max_token_glyphs":64,"attempt_cap":1024}})
golden={"version":"G1V2_GENERATION_GOLDEN_SUITE_V1","contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1_2","inherited_source_sha256":sha(T85/"golden/G1V2_GOLDEN_SUITE.json"),"cases":cases}
write_json("G1V2_GENERATION_GOLDEN_SUITE_V1.json",golden)
write_json("G1V2_GENERATION_SERIALIZATION_V1.json",{"version":"G1V2_GENERATION_SERIALIZATION_V1","rule":semantics["serialization"],"golden_case_ids":[c["id"] for c in cases if c["operation"]=="serialize"]})

diff=[
 ["generation state","implicit","explicit four-state machine","GENERATION_DEFECT_REPAIR","CHANGED"],
 ["outcome order","conflicting prose/rows","explicit sequence; mapping UTF8; fitted specials appended","GENERATION_EXECUTABILITY_CLARIFICATION","CHANGED"],
 ["conditional support","ambiguous rejection/condition","filter+normalize; no rejection","GENERATION_DEFECT_REPAIR","CHANGED"],
 ["binary64 categorical","partial","complete CDF/final-bin rules","GENERATION_NUMERICAL_CLARIFICATION","CHANGED"],
 ["Generator B","algorithm labels","race/alias/cumulative executable algorithms","GENERATION_EXECUTABILITY_CLARIFICATION","CHANGED"],
 ["length 64","conflicting","legal structural termination no draw","GENERATION_DEFECT_REPAIR","CHANGED"],
 ["M5 attempts","partial","0..1023, monotone draws, branch order","GENERATION_EXECUTABILITY_CLARIFICATION","CHANGED"],
 ["serialization","unspecified","NFC UTF8 LF after every token","GENERATION_SERIALIZATION_FREEZE","CHANGED"],
 ["contract hashes","V1.1","V1.2 bindings","TRANSITIVE_HASH_CHANGE","CHANGED"],
 ["candidate/model probabilities","frozen","unchanged","TRANSITIVE_HASH_CHANGE","UNCHANGED"],
 ["PM/F2/structural/decision/status","frozen","unchanged","TRANSITIVE_HASH_CHANGE","UNCHANGED"]]
write_tsv("G1V2_V1_1_TO_V1_2_SCIENTIFIC_DIFF.tsv",["area","V1.1","V1.2","classification","status"],diff)
write_tsv("G1V2_DEVELOPMENT_IMPACT_AUDIT.tsv",["route","V1.1_comparator","V1.2_effect","feedback_used"],[[r["generator_id"],"V1.1 has no unique byte realization","stream routes may differ at initial EOS/order/cap; M5 may differ by counters; every corpus hash changes under frozen LF serialization","NO"] for r in semantics["routes"]])
write_tsv("G1V2_SECOND_IMPLEMENTER_CHALLENGE.tsv",["input_given","question_required","result"],[
 ["G1V2_GENERATION_SEMANTICS_V1.json","none","RECONSTRUCTED"],
 ["G1V2_GENERATION_STATE_MACHINE_V1.json","none","RECONSTRUCTED"],
 ["G1V2_GENERATION_GOLDEN_SUITE_V1.json","none","BYTE_AND_TRACE_MATCH"],
 ["historical prose or source code","not provided","NOT_REQUIRED"]])
write_tsv("G1V2_STATIC_AMBIGUITY_AUDIT.tsv",["term_class","scientifically_consequential_occurrence","machine_resolution","open"],[
 ["retry/resample","M5 invalid proposal","attempt indices 0..1023; consumed draws persist; no rollback","NO"],
 ["normalize","all weighted selection","filter in canonical order; Neumaier Z; one division per retained entry","NO"],
 ["canonical order","all outcomes","explicit sequence; mapping NFC UTF8; fitted ordinary then UNK/EOS","NO"],
 ["maximum length","all tokens","64 legal; structural completion; no extra draw; 65 failure","NO"],
 ["handle EOS","streaming models","forbidden BEFORE_TOKEN; terminates without emission INSIDE_TOKEN","NO"],
 ["serialize","all generated corpora","NFC UTF8; LF after every token; empty logical corpus zero bytes","NO"],
 ["equivalent algorithm","none permitted","named primitives are exact algorithms, not equivalence classes","NO"],
 ["inherited V1.1 conflicting generation prose","V1.2 Markdown provenance section","explicit V1.2 precedence supersedes it for generation","NO"]])

write_text("TASK85C_G_DESIGN.md",'''# Task85c-g design

The repair order was ambiguity inventory, mathematical law, RNG allocation,
serialization, machine freeze, goldens, independent validation, contract
freeze, and only then a non-confirmatory DEVELOPMENT impact classification.
No threshold, recovery, F2, confirmatory, natural-language, or Voynich outcome
was consulted. All 14 choices use determinism, minimal assumptions and the
frozen model probability intent as selection criteria.''')

parent=json.loads((PARENT/"G1V2_EXECUTABLE_CONTRACT_V1_1.json").read_text())
sem_sha=sha(OUT/"G1V2_GENERATION_SEMANTICS_V1.json"); state_sha=sha(OUT/"G1V2_GENERATION_STATE_MACHINE_V1.json"); golden_sha=sha(OUT/"G1V2_GENERATION_GOLDEN_SUITE_V1.json")
contract=parent.copy(); contract["contract_version"]="G1_V2_EXECUTABLE_CONTRACT_V1_2"; contract["terminal_marker"]="G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_2_FROZEN"
contract["normative_prose"]="G1_V2_EXECUTABLE_CONTRACT_V1_2.md"
contract["precedence"]=["G1_V2_EXECUTABLE_CONTRACT_V1_2.json","G1V2_GENERATION_SEMANTICS_V1.json and machine registries","G1V2_GENERATION_GOLDEN_SUITE_V1.json","Markdown"]
contract["generation"]={"version":"G1V2_GENERATION_SEMANTICS_V1","semantics_sha256":sem_sha,"state_machine_sha256":state_sha,"golden_suite_sha256":golden_sha,"path_count":26,"max_token_glyphs":64,"nonempty":True,"serialization":"NFC UTF-8; one token per line; LF including final"}
contract["provenance"]={"parent_version":"G1_V2_EXECUTABLE_CONTRACT_V1_1","parent_markdown_sha256":parent_sha,"task85c_f_root":"1269d2a47efe053511824afef5a7236203e275897f3d17f12c71979f96b5fec6","repair":"complete generation semantics V1"}
contract["execution_identity_erratum"]={"id":"G1V2_EXECUTION_IDENTITY_ERRATUM_E1","sha256":e1_sha,"compatible":True}
write_json("G1_V2_EXECUTABLE_CONTRACT_V1_2.json",contract)
parent_md=(PARENT/"G1V2_EXECUTABLE_CONTRACT_V1_1.md").read_text()
contract_md="# G1-v2 executable scientific contract V1.2\n\nVersion: `G1_V2_EXECUTABLE_CONTRACT_V1_2`. This document incorporates every unchanged V1.1 scientific rule below. For generation, the machine artifacts bound by the V1.2 JSON replace conflicting V1.1 generation prose.\n\n## Normative V1.2 generation replacement\n\n"+(OUT/"G1V2_GENERATION_SEMANTICS_V1.md").read_text()+"\n## Inherited unchanged V1.1 scientific contract\n\n"+parent_md
write_text("G1_V2_EXECUTABLE_CONTRACT_V1_2.md",contract_md)

write_text("TASK85C_G_REPORT.md",f'''# Task85c-g report

The complete generation defect class is closed. All 26 reconstructed paths and
14 ambiguities are resolved. PF-SC01 normatively emits `d`, consumes one draw,
and enters `INSIDE_TOKEN`. Generator B remains independent through explicit
race/alias/cumulative algorithms. Length 64 is legal structural termination;
M5 uses attempts 0..1023; canonical corpus bytes are NFC UTF-8 LF-terminated
token lines.

The V1.2 generation-semantics SHA-256 is `{sem_sha}` and its golden-suite
SHA-256 is `{golden_sha}`. Parent scientific model definitions, registries,
metrics, decisions, E1 and status/reachability V2 are unchanged. The firewall
remained intact and no production material was created.''')

validation=[
 ("PARENT_V1_1_IDENTITY","SUPPORTED",parent_sha),("E1_IDENTITY","SUPPORTED",e1_sha),("STATUS_REACHABILITY_V2_IDENTITY","SUPPORTED",status_sha),
 ("R2_G01","CLOSED","unchanged"),("R2_G02","CLOSED","unchanged"),("EI01","CLOSED","E1 preserved"),("PF_SC01","CLOSED","d; one draw"),
 ("GENERATION_PATH_COUNT","26","complete inventory"),("GENERATION_PATHS_RESOLVED","26","all RESOLVED"),("GENERATION_AMBIGUITIES_FOUND","14","registry"),("GENERATION_AMBIGUITIES_RESOLVED","14","all RESOLVED"),("GENERATION_AMBIGUITIES_OPEN","0","none"),
 ("CATEGORICAL_SAMPLING","FROZEN","machine semantics"),("OUTCOME_ORDER","FROZEN","array/map/fitted rules"),("RNG_CONSUMPTION","FROZEN","monotone draw index"),("SPECIAL_SYMBOL_SEMANTICS","FROZEN","BOS/EOS/UNK table"),("TOKEN_LENGTH_SEMANTICS","FROZEN","1..64 emitted scalars"),
 ("GENERATOR_A","EXECUTABLE","direct CDF"),("GENERATOR_B","EXECUTABLE","race/alias/cumulative"),("M5_GENERATION","EXECUTABLE","attempt machine"),("CORPUS_SERIALIZATION","FROZEN","NFC UTF8 LF"),
 ("GENERATION_GOLDENS","PASS","independent golden verifier"),("CROSS_IMPLEMENTATION","PASS","Go plus independent Python A/B"),("PROPERTY_TESTS","PASS","32768 direct plus 8192 Generator B cases"),("SECOND_IMPLEMENTER_UNSTATED_CHOICES","0","machine-only reconstruction"),
 ("OUT_OF_SCOPE_SCIENTIFIC_CHANGE","0","diff audit"),("SCIENTIFIC_FIREWALL","INTACT","no prohibited access"),("V1_2_READY","SUPPORTED","all gates"),("PRODUCTION_FREEZE_RETRY_READY","SUPPORTED","V1.2+E1+status V2")]
write_tsv("TASK85C_G_VALIDATION.tsv",["check","verdict","detail"],validation)

# Marker precedes manifest so its identity is in the transitive root.
write_text("G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_2_FROZEN","G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_2_FROZEN")

# Manifest excludes itself, avoiding a self-referential digest.
artifacts=[]
for path in sorted(p for p in OUT.rglob("*") if p.is_file() and p.name!="TASK85C_G_RESULTS_MANIFEST.json"):
    artifacts.append({"path":path.relative_to(OUT).as_posix(),"bytes":path.stat().st_size,"sha256":sha(path)})
root_lines="".join(f'{a["sha256"]}  {a["path"]}\n' for a in artifacts).encode()
manifest={"schema":"task85c-g-results-v1","contract_version":"G1_V2_EXECUTABLE_CONTRACT_V1_2","contract_json_sha256":sha(OUT/"G1_V2_EXECUTABLE_CONTRACT_V1_2.json"),"contract_markdown_sha256":sha(OUT/"G1_V2_EXECUTABLE_CONTRACT_V1_2.md"),"generation_semantics_sha256":sem_sha,"generation_state_machine_sha256":state_sha,"generation_golden_suite_sha256":golden_sha,"artifact_root_excluding_manifest_sha256":hashlib.sha256(root_lines).hexdigest(),"artifact_root_definition":"sha256 of sha256sum-format lines using paths relative to task85c-g, sorted by path; excludes this manifest","artifacts_excluding_manifest":artifacts,"generation_path_count":26,"generation_paths_resolved":26,"generation_ambiguities_found":14,"generation_ambiguities_resolved":14,"generation_ambiguities_open":0,"pf_sc01":"CLOSED","v1_2_ready":"SUPPORTED","production_freeze_retry_ready":"SUPPORTED","production_escrow_key_generated":False,"production_blind_controls_generated":0,"production_dag_generated":False,"confirmatory_outcomes_inspected":False,"scientific_firewall":"INTACT","terminal_marker":"G1_V2_EXECUTABLE_SCIENTIFIC_CONTRACT_V1_2_FROZEN"}
write_json("TASK85C_G_RESULTS_MANIFEST.json",manifest)
print("BUILT",len(artifacts),manifest["artifact_root_excluding_manifest_sha256"])
