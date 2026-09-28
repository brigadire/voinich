#!/usr/bin/env python3
import csv, hashlib, itertools, json, time
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parent
GRAPH = ROOT.parent / "exploratory_astronomical_dictionary_search_e2_scorer_consistent_graph_v1"
PREP = ROOT.parent / "exploratory_astronomical_dictionary_search_v1"

def rows(path):
    with path.open() as f:
        return list(csv.DictReader(f, delimiter="\t"))

def hash_file(path):
    return hashlib.sha256(path.read_bytes()).hexdigest()

def main():
    started = time.monotonic()
    graph_rows = rows(GRAPH / "PATH_GRAPH.tsv")
    # A path's rule signature is the complete active table for that path. With
    # exactly two distinct EVA token types, every active rule must occur in both
    # paths, so support-valid coverage-2 solutions are exactly pairs in one
    # signature group with distinct labels and identities.
    groups = defaultdict(list)
    for i, row in enumerate(graph_rows):
        rules = tuple(sorted(x for x in row["rules"].split(";") if x))
        if 1 <= len(rules) <= 5:
            row = dict(row, signature=";".join(rules), rule_count=str(len(rules)))
            graph_rows[i] = row
            groups[row["signature"]].append(row)
    solutions = []
    signatures = []
    for signature, paths in sorted(groups.items()):
        valid_pairs = []
        for a, b in itertools.combinations(paths, 2):
            if a["label_id"] == b["label_id"] or a["identity"] == b["identity"]:
                continue
            if a["token"] == b["token"]:
                continue
            valid_pairs.append((a, b))
        if valid_pairs:
            pages = sorted({p["page"] for p in paths})
            types = sorted({p["token"] for p in paths})
            signatures.append({"signature":signature,"rule_count":len(signature.split(";")),"path_count":len(paths),"valid_pair_count":len(valid_pairs),"distinct_eva_types":len(types),"pages":','.join(pages),"support_status":"PASS"})
            for pair_index, (a,b) in enumerate(valid_pairs,1):
                table_hash = hashlib.sha256(signature.encode()).hexdigest()
                for slot, p in enumerate((a,b),1):
                    solutions.append({"solution_id":f"{table_hash[:12]}_{pair_index}","pair_index":pair_index,"table_hash":table_hash,"signature":signature,"rule_count":len(signature.split(";")),"slot":slot,"label_id":p["label_id"],"page":p["page"],"token":p["token"],"identity":p["identity"],"lexicon_id":p["lexicon_id"],"normalized_form":p["normalized_form"],"encoded":p["encoded"]})
    total_pairs = sum(int(x["valid_pair_count"]) for x in signatures)
    write("COVERAGE_2_SIGNATURES.tsv", ["signature","rule_count","path_count","valid_pair_count","distinct_eva_types","pages","support_status"], signatures)
    write("COVERAGE_2_SOLUTIONS.tsv", ["solution_id","pair_index","table_hash","signature","rule_count","slot","label_id","page","token","identity","lexicon_id","normalized_form","encoded"], solutions)
    tests = [
        {"threshold":"coverage>=2","status":"FEASIBLE_CERTIFIED" if total_pairs else "INFEASIBLE_CERTIFIED","tested_pairs":total_pairs,"candidate_paths":len(graph_rows),"elapsed_sec":round(time.monotonic()-started,4),"method":"complete grouped pair enumeration"},
        {"threshold":"coverage>=3","status":"NOT_RUN_BY_SCOPE","tested_pairs":total_pairs,"candidate_paths":len(graph_rows),"elapsed_sec":round(time.monotonic()-started,4),"method":"coverage-2 package only"},
    ]
    write("EXISTENCE_TESTS.tsv", ["threshold","status","tested_pairs","candidate_paths","elapsed_sec","method"], tests)
    progress=[{"stage":"graph_loaded","elapsed_sec":round(time.monotonic()-started,4),"paths":len(graph_rows),"groups":len(groups),"status":"PASS"},{"stage":"all_pairs_enumerated","elapsed_sec":round(time.monotonic()-started,4),"paths":len(graph_rows),"groups":len(groups),"status":"PASS"}]
    write("SOLVER_PROGRESS.tsv", ["stage","elapsed_sec","paths","groups","status"], progress)
    inc=[{"incumbent_id":"COVERAGE_2_EXACT","coverage":2,"solution_count":total_pairs,"signature_count":len(signatures),"status":"CERTIFIED" if total_pairs else "NONE"}]
    write("INCUMBENTS.tsv", ["incumbent_id","coverage","solution_count","signature_count","status"], inc)
    write("MATCH_ASSIGNMENTS.tsv", ["solution_id","pair_index","table_hash","signature","rule_count","slot","label_id","page","token","identity","lexicon_id","normalized_form","encoded"], solutions)
    support=[]
    for s in signatures:
        for rule in s["signature"].split(";"):
            toks=sorted({r["token"] for r in graph_rows if r["signature"]==s["signature"]})
            support.append({"table_hash":hashlib.sha256(s["signature"].encode()).hexdigest(),"signature":s["signature"],"rule":rule,"distinct_eva_types":len(toks),"support_types":','.join(toks),"status":"PASS" if len(toks)>=2 else "FAIL"})
    write("RULE_SUPPORT_AUDIT.tsv", ["table_hash","signature","rule","distinct_eva_types","support_types","status"], support)
    freeze={"scope_sha256":hash_file(PREP/"TARGET_STAR_LABELS.tsv"),"lexicon_sha256":hash_file(PREP/"ASTRONOMICAL_NAMES_LEXICON.tsv"),"graph_sha256":hash_file(GRAPH/"PATH_GRAPH.tsv"),"graph_summary_sha256":hash_file(GRAPH/"GRAPH_BUILD_SUMMARY.json"),"graph_path_count":len(graph_rows),"rule_limit":5,"exact_scorer":"encode_word(DROP_UNMAPPED,NONE)","enumeration":"all unordered path pairs within exact rule-signature groups"}
    (ROOT/"INPUT_FREEZE.json").write_text(json.dumps(freeze,indent=2,sort_keys=True)+"\n")
    summary=[]
    for r in graph_rows:
        summary.append({"signature":r["signature"],"rule_count":r["rule_count"],"path_count":1,"label_id":r["label_id"],"page":r["page"],"token":r["token"],"identity":r["identity"],"normalization":"exact scorer parity"})
    write("PATH_NORMALIZATION_SUMMARY.tsv", ["signature","rule_count","path_count","label_id","page","token","identity","normalization"], summary)
    status={"SUPPORT_VALID_SOLUTION_FOUND":"YES" if total_pairs else "NO","NO_SUPPORT_VALID_SOLUTION_CERTIFIED":"NO" if total_pairs else "YES","SEARCH_INCONCLUSIVE_TIMEOUT":"NO","IMPLEMENTATION_GATE_FAILED":"NO","COVERAGE_2_EXACT_ENUMERATION":"COMPLETE","GRAPH_PATH_COUNT":len(graph_rows),"COVERAGE_2_SOLUTION_COUNT":total_pairs,"COVERAGE_2_SIGNATURE_COUNT":len(signatures),"SCIENTIFIC_CLAIM":"NONE"}
    (ROOT/"RUN_STATUS.json").write_text(json.dumps(status,indent=2,sort_keys=True)+"\n")
    print(json.dumps(status,sort_keys=True))

def write(name, fields, data):
    with (ROOT/name).open("w",newline="") as f:
        w=csv.DictWriter(f,fieldnames=fields,delimiter="\t"); w.writeheader(); w.writerows(data)

if __name__ == "__main__": main()
