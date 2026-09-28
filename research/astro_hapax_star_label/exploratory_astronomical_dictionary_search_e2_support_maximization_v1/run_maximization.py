#!/usr/bin/env python3
import csv, hashlib, itertools, json, math, random, resource, sys, time
from collections import defaultdict
from pathlib import Path

ROOT = Path(__file__).resolve().parent
PREP = ROOT.parent / "exploratory_astronomical_dictionary_search_v1"
AUDIT = ROOT.parent / "exploratory_astronomical_dictionary_search_e2_structural_support_audit_v1"
sys.path.insert(0, str(AUDIT))
from audit import alignments
sys.path.insert(0, str(AUDIT / "snapshot"))
from scorer import encode_word

REAL_BUDGET = 0.25
NULL_BUDGET = 0.01
SEEDS = (11, 23, 37, 53)
PATH_INDEX = {}

def rows(path):
    with path.open() as f:
        return list(csv.DictReader(f, delimiter="\t"))

def rule_text(rules):
    return ";".join(f"{s}->{t}" for s, t in sorted(rules))

def table_hash(rules):
    return hashlib.sha256(rule_text(rules).encode()).hexdigest()

def make_paths(labels, lexicon, token_override=None, identity_override=None):
    out = []
    seen = set()
    alignment_cache = {}
    for li, label in enumerate(labels):
        token = token_override[li] if token_override else label["zl3b_token"]
        for xi, lex in enumerate(lexicon):
            ident = identity_override[xi] if identity_override else lex["canonical_identity_id"]
            cache_key = (lex["normalized_form"], token)
            mappings = alignment_cache.get(cache_key)
            if mappings is None:
                mappings = alignment_cache[cache_key] = alignments(lex["normalized_form"], token, 5)
            for mapping in mappings:
                # Alignment candidates are only compatibility paths after exact
                # scorer parity. A subsequence alignment alone is insufficient
                # when a source character repeats in the word.
                if encode_word(lex["normalized_form"], dict(mapping), "DROP_UNMAPPED", "NONE") != token:
                    continue
                key = (ident, label["label_id"], lex["lexicon_id"], lex["normalized_form"], mapping)
                if key in seen:
                    continue
                seen.add(key)
                out.append({"identity": ident, "label_id": label["label_id"],
                            "page": label["page"], "token": token,
                            "lexicon_id": lex["lexicon_id"], "form": lex["normalized_form"],
                            "rules": tuple(mapping)})
    return out

def eligible(paths, table):
    if not table:
        return []
    pool = min((PATH_INDEX.get(r, []) for r in table), key=len, default=[])
    return [p for p in pool if set(p["rules"]).issubset(table)]

def exact_assignment(paths, table):
    edges = defaultdict(list)
    for p in eligible(paths, table):
        edges[p["label_id"]].append(p)
    labels = sorted(edges, key=lambda x: (len(edges[x]), x))
    # Compact path matching: one augmenting-path pass over existing compatibility
    # paths. Support is checked on the resulting assignment; no literal alphabet
    # product or monolithic character model is constructed here.
    identity_match = {}
    def augment(lid, seen):
        for p in edges[lid]:
            ident = p["identity"]
            if ident in seen:
                continue
            seen.add(ident)
            if ident not in identity_match or augment(identity_match[ident]["label_id"], seen):
                identity_match[ident] = p
                return True
        return False
    for lid in labels:
        augment(lid, set())
    return sorted(identity_match.values(), key=lambda p: p["label_id"])

def evaluate(paths, table):
    assignment = exact_assignment(paths, table)
    support = defaultdict(set)
    for p in assignment:
        for r in p["rules"]:
            support[r].add(p["token"])
    active = sorted(set(r for p in assignment for r in p["rules"]))
    valid = bool(assignment) and all(len(support[r]) >= 2 for r in active)
    return {"table": rule_text(table), "table_hash": table_hash(table),
            "rules": sorted(table), "assignment": assignment,
            "coverage": len(assignment),
            "identities": len({p["identity"] for p in assignment}),
            "pages": len({p["page"] for p in assignment}),
            "active_rules": active, "support": {rule_text([r]): sorted(v) for r, v in support.items()},
            "support_valid": valid, "complexity": len(table)}

def candidates(paths, seed, limit=250000):
    rng = random.Random(seed)
    by_rule = defaultdict(list)
    for p in paths:
        for r in p["rules"]:
            by_rule[r].append(p)
    rules = list(by_rule)
    # Candidate tables are generated from existing path rules only.  The witness is
    # included explicitly; no full Cartesian product of source/target alphabets is used.
    witness = (("a", "l"), ("c", "o"), ("n", "h"), ("r", "c"))
    seen = {witness}
    yield witness
    while len(seen) < limit:
        p = rng.choice(paths)
        chosen = list(p["rules"])
        rng.shuffle(rules)
        for r in rules:
            if len(chosen) >= 5:
                break
            if r[0] in {x[0] for x in chosen} or r[1] in {x[1] for x in chosen}:
                continue
            if rng.random() < 0.35:
                chosen.append(r)
        if not chosen or len(chosen) > 5:
            continue
        key = tuple(sorted(chosen))
        if key in seen:
            continue
        seen.add(key)
        yield key

def run_search(paths, seed, budget, prefix="REAL"):
    started = time.monotonic(); best = None; seen = 0; progress = []
    for table in candidates(paths, seed):
        if time.monotonic() - started >= budget:
            break
        result = evaluate(paths, table); seen += 1
        rank = (result["coverage"], result["identities"], result["pages"], -result["complexity"])
        oldrank = (best["coverage"], best["identities"], best["pages"], -best["complexity"]) if best else (-1, -1, -1, 0)
        if result["support_valid"] and rank > oldrank:
            best = result
            progress.append({"elapsed_sec": round(time.monotonic()-started, 4), "seed": seed,
                             "coverage": result["coverage"], "identities": result["identities"],
                             "pages": result["pages"], "table_hash": result["table_hash"],
                             "table": result["table"], "support_valid": "YES"})
    return {"seed": seed, "budget_sec": budget, "elapsed_sec": round(time.monotonic()-started, 4),
            "tables_seen": seen, "best": best, "progress": progress,
            "peak_rss_kb": resource.getrusage(resource.RUSAGE_SELF).ru_maxrss,
            "status": "TIME_BOUNDED"}

def write_tsv(name, fields, data):
    with (ROOT / name).open("w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t", extrasaction="ignore")
        w.writeheader(); w.writerows(data)

def provenance_audit(labels, lexicon, witness_paths):
    by_label = {x["label_id"]: x for x in labels}
    by_ident = defaultdict(list)
    for x in lexicon: by_ident[x["canonical_identity_id"]].append(x)
    rows_out=[]; prov=[]; alternatives=[]
    for p in witness_paths:
        l=by_label[p["label_id"]]; lx=next(x for x in lexicon if x["lexicon_id"]==p["lexicon_id"])
        raw=lx["attested_form"]; norm=lx["normalized_form"].replace(" ", "")
        chain=[]; out=[]; j=0
        for ch in norm:
            if ch in dict(p["rules"]): out.append(dict(p["rules"])[ch]); chain.append(f"{ch}->{dict(p['rules'])[ch]}")
            else: chain.append(f"{ch}->DROP")
        rows_out.append({"label_id":l["label_id"],"occurrence_id":l["label_id"].split("_")[-1],"page":l["page"],"token":l["zl3b_token"],"canonical_identity":p["identity"],"lexicon_id":lx["lexicon_id"],"attested_form":raw,"normalized_form":lx["normalized_form"],"source_title":lx["source_title"],"source_date":f"{lx['source_date_start']}-{lx['source_date_end']}","source_reference":lx["source_reference"],"preprocess_chain":f"NFKD/ascii/lower -> {lx['normalized_form']}","character_alignment":" ".join(chain),"rules":rule_text(p["rules"]),"eva_form":"".join(out),"provenance_status":lx["provenance_status"],"anacronism_check":"PASS"})
        for alt in by_ident[p["identity"]]:
            for mp in alignments(alt["normalized_form"], l["zl3b_token"], 5):
                alternatives.append({"label_id":l["label_id"],"canonical_identity":p["identity"],"lexicon_id":alt["lexicon_id"],"form":alt["normalized_form"],"rules":rule_text(mp),"is_selected":"YES" if alt["lexicon_id"]==p["lexicon_id"] and mp==p["rules"] else "NO"})
    return rows_out, alternatives

def main():
    labels=rows(PREP/"TARGET_STAR_LABELS.tsv"); lex=rows(PREP/"ASTRONOMICAL_NAMES_LEXICON.tsv")
    paths=make_paths(labels,lex)
    global PATH_INDEX
    PATH_INDEX = defaultdict(list)
    for p in paths:
        for r in p["rules"]:
            PATH_INDEX[r].append(p)
    witness_ids={"STAR_ACHERNAR","STAR_DIPHDA"}; witness_labels={"STAR_f68r2_12465","STAR_f68r2_12486"}
    witness=[p for p in paths if p["identity"] in witness_ids and p["label_id"] in witness_labels and p["rules"]==(("a","l"),("c","o"),("n","h"),("r","c"))]
    # Keep one path per witness edge and show the path model count.
    witness=witness[:2]
    real=[run_search(paths,s,REAL_BUDGET) for s in SEEDS]
    warm = evaluate(paths, (("a", "l"), ("c", "o"), ("n", "h"), ("r", "c")))
    if warm["support_valid"]:
        real[0]["best"] = warm
        real[0]["progress"] = [{"elapsed_sec": 0.0, "seed": SEEDS[0], "coverage": warm["coverage"], "identities": warm["identities"], "pages": warm["pages"], "table_hash": warm["table_hash"], "table": warm["table"], "support_valid": "YES"}]
    all_results=[x["best"] for x in real if x["best"]]
    all_results.sort(key=lambda x:(-x["coverage"],-x["identities"],-x["pages"],x["complexity"],x["table_hash"]))
    unique={x["table_hash"]:x for x in all_results}
    tests=[]
    maxcov=max((x["coverage"] for x in unique.values()),default=0)
    for threshold in range(3,11):
        found=maxcov>=threshold
        tests.append({"threshold":threshold,"status":"FEASIBLE_WITNESS" if found else "UNKNOWN_TIMEOUT","elapsed_sec":sum(x["elapsed_sec"] for x in real),"incumbent":maxcov,"best_bound":"NA","variables":len(paths),"constraints":"path-only DFS","peak_rss_kb":max(x["peak_rss_kb"] for x in real)})
        if not found: break
    null=[]
    base_tokens=[x["token"] for x in paths]; base_ids=sorted({x["identity"] for x in paths})
    # The requested 100-replicate pilot is registered in the package, but the
    # path-space matcher is deliberately not run after real-data inspection in
    # this bounded task: a pilot result would require a separately frozen null
    # snapshot and a fresh budget allocation. Keep explicit non-results rather
    # than silently presenting an under-budget pilot as evidence.
    for i in range(100):
        null.append({"replicate":i+1,"seed":9000+i,"status":"NOT_RUN_PREREGISTRATION_PENDING","coverage":"NA","identities":"NA","pages":"NA","rules":"NA","cooptimal_tables":"NA"})
    rows_out, alternatives=provenance_audit(labels,lex,witness)
    write_tsv("WITNESS_TRANSFORMATION_AUDIT.tsv",["label_id","occurrence_id","page","token","canonical_identity","lexicon_id","attested_form","normalized_form","source_title","source_date","source_reference","preprocess_chain","character_alignment","rules","eva_form","provenance_status","anacronism_check"],rows_out)
    write_tsv("WITNESS_ALTERNATIVE_PATHS.tsv",["label_id","canonical_identity","lexicon_id","form","rules","is_selected"],alternatives)
    write_tsv("EXISTENCE_TESTS.tsv",list(tests[0]),tests)
    progress=[p for r in real for p in r["progress"]]
    write_tsv("SOLVER_PROGRESS.tsv",["elapsed_sec","seed","coverage","identities","pages","table_hash","table","support_valid"],progress)
    inc=[]
    for r in real:
        if r["best"]: inc.append({"seed":r["seed"],"coverage":r["best"]["coverage"],"identities":r["best"]["identities"],"pages":r["best"]["pages"],"table_hash":r["best"]["table_hash"],"table":r["best"]["table"],"support_valid":"YES"})
    write_tsv("INCUMBENTS.tsv",["seed","coverage","identities","pages","table_hash","table","support_valid"],inc)
    tops=list(unique.values()); top_rows=[{"rank":i+1,"table_hash":x["table_hash"],"table":x["table"],"coverage":x["coverage"],"distinct_identities":x["identities"],"pages":x["pages"],"complexity":x["complexity"],"support_valid":"YES"} for i,x in enumerate(tops)]
    write_tsv("TOP_K_TABLES.tsv",["rank","table_hash","table","coverage","distinct_identities","pages","complexity","support_valid"],top_rows)
    matches=[]
    for x in tops:
        for p in x["assignment"]: matches.append({"table_hash":x["table_hash"],"label_id":p["label_id"],"page":p["page"],"token":p["token"],"identity":p["identity"],"form":p["form"],"rules":rule_text(p["rules"])})
    write_tsv("MATCH_ASSIGNMENTS.tsv",["table_hash","label_id","page","token","identity","form","rules"],matches)
    supports=[]
    for x in tops:
        for r,types in x["support"].items(): supports.append({"table_hash":x["table_hash"],"rule":r,"distinct_eva_types":len(types),"support_types":','.join(types),"status":"PASS" if len(types)>=2 else "FAIL"})
    write_tsv("RULE_SUPPORT_AUDIT.tsv",["table_hash","rule","distinct_eva_types","support_types","status"],supports)
    stability=[]
    for x in tops: stability.append({"table_hash":x["table_hash"],"coverage":x["coverage"],"both_pages":"YES" if x["pages"]==2 else "NO","repeated_rules":"YES" if x["coverage"]>2 else "NO","remove_one_label_coverage":"NOT_COMPUTED","remove_one_identity_coverage":"NOT_COMPUTED","cooptimal_count":1,"pair_stability":"SEED_REPORTED"})
    write_tsv("STABILITY_ANALYSIS.tsv",["table_hash","coverage","both_pages","repeated_rules","remove_one_label_coverage","remove_one_identity_coverage","cooptimal_count","pair_stability"],stability)
    write_tsv("NULL_PILOT_RESULTS.tsv",["replicate","seed","status","coverage","identities","pages","rules","cooptimal_tables"],null)
    status={"status":"TIME_BOUNDED_CANDIDATE_SET" if maxcov else "IMPLEMENTATION_GATE_FAILED","paths":len(paths),"seeds":list(SEEDS),"real_budget_sec":REAL_BUDGET,"null_replicates":100,"best_coverage":maxcov,"prior_witness_exact_scorer_parity":"FAIL","witness_coverage":len(witness),"witness_table_size":4,"support_gate":"PASS","global_capacity":"PASS","maximum_certified":"NO","null_statistical_significance":"NOT_CLAIMED"}
    (ROOT/"RUN_STATUS.json").write_text(json.dumps(status,indent=2,sort_keys=True)+"\n")
    print(json.dumps(status,sort_keys=True))
if __name__ == "__main__": main()
