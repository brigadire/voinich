#!/usr/bin/env python3
"""Alternative matching-model audit on the frozen D1/M1 experiment."""

from __future__ import annotations

import argparse, csv, hashlib, importlib.util, json, os, random, statistics, sys
from collections import Counter, defaultdict
from dataclasses import dataclass
from datetime import datetime, timezone
from multiprocessing import get_context
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/astro_matching_models"
D1 = ROOT / "research/astro_dictionary_expansion_m1"
spec = importlib.util.spec_from_file_location("matching_frozen_m1", ROOT / "research/astro_token_formation_m1/main.py")
assert spec and spec.loader
m1 = importlib.util.module_from_spec(spec); sys.modules[spec.name] = m1; spec.loader.exec_module(m1)

REUSE_PENALTY = 0.05
CONTROLS = ("RANDOM_VOYNICH_SET", "SHUFFLED_TERMS", "PSEUDODICTIONARY")
RUN_TERMS: list[dict[str, object]] = []


@dataclass(frozen=True)
class ReuseState:
    mapping: tuple
    assignments: tuple


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as fh: return list(csv.DictReader(fh, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as fh:
        w = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n", extrasaction="ignore"); w.writeheader(); w.writerows(rows)


def sha(path: Path) -> str: return hashlib.sha256(path.read_bytes()).hexdigest()
def fmt(x: float) -> str: return f"{x:.6f}"


def load_terms() -> list[dict[str, object]]:
    grouped: dict[tuple[str, str], set[str]] = defaultdict(set)
    for row in read_tsv(D1 / "ASTRO_TERM_CORPUS_EXPANDED.tsv"):
        if row["object_class"] in {"STAR", "PLANET_MOON"}: grouped[(row["term_id"], row["object_class"])].add(row["normalized_form"])
    return [{"object_id": oid, "object_class": cls, "forms": sorted(forms)} for (oid, cls), forms in sorted(grouped.items())]


def reuse_excess(assignments: tuple) -> int:
    counts = Counter(item[1] for item in assignments)
    return sum(n - 1 for n in counts.values() if n > 1)


def state_score(state: ReuseState, total: int, pipe: object) -> float:
    return m1.base_score(len(state.assignments), total, state.mapping, pipe) - REUSE_PENALTY * reuse_excess(state.assignments)


def search_pipeline(labels: list[dict[str, str]], terms: list[dict[str, object]], pipe: object, pre: dict) -> list[ReuseState]:
    edges = m1.build_edges(labels, terms, pre)
    order = sorted((r["stolfi_coordinate"] for r in labels), key=lambda x: (len(edges[x]), x))
    beam = [ReuseState((), ())]
    for label_id in order:
        candidates: dict[tuple, ReuseState] = {}
        for state in beam:
            usage = tuple(sorted(Counter(item[1] for item in state.assignments).items()))
            candidates.setdefault((state.mapping, usage), state)
            for edge in edges[label_id]:
                merged = m1.compatible_merge(state.mapping, edge.constraint)
                if merged is None: continue
                new = ReuseState(merged, state.assignments + ((edge.label, edge.object_id, edge.source_form, edge.source_units),))
                new_usage = tuple(sorted(Counter(item[1] for item in new.assignments).items()))
                key = (new.mapping, new_usage)
                old = candidates.get(key)
                if old is None or new.assignments < old.assignments: candidates[key] = new
        beam = sorted(candidates.values(), key=lambda s: (-state_score(s, len(labels), pipe), -len(s.assignments), len(s.mapping), s.mapping, s.assignments))[:m1.BEAM_WIDTH]
    return sorted(beam, key=lambda s: (-state_score(s, len(labels), pipe), -len(s.assignments), len(s.mapping), s.mapping, s.assignments))[:m1.FINALISTS_PER_PIPELINE]


def evaluate(labels: list[dict[str, str]], terms: list[dict[str, object]], pipe: object, pre: dict, state: ReuseState) -> dict[str, object]:
    adj, evidence = m1.adjacency(labels, terms, pre, state.mapping)
    match = {row["stolfi_coordinate"]: sorted(adj[row["stolfi_coordinate"]])[0] for row in labels if adj.get(row["stolfi_coordinate"])}
    excess = reuse_excess(tuple((lid, oid, "", ()) for lid, oid in match.items()))
    metrics = m1.mapping_metrics(state.mapping, pipe)
    score = m1.base_score(len(match), len(labels), state.mapping, pipe) - REUSE_PENALTY * excess
    ambiguity = sum(max(0, len(adj.get(r["stolfi_coordinate"], set())) - 1) for r in labels) / len(labels)
    return {"pipeline": pipe, "pre": pre, "mapping": state.mapping, "adjacency": adj, "evidence": evidence,
            "train_match": match, "train_coverage": len(match)/len(labels), "reuse_excess": excess,
            "ambiguity_excess": ambiguity, "score": score, **metrics}


def run_search(labels: list[dict[str, str]], terms: list[dict[str, object]], details: bool) -> list[dict[str, object]]:
    reps = {}
    for pipe in m1.all_pipelines():
        pre = m1.preprocess(terms, pipe); key = m1.representation_key(pre)
        if key not in reps or (pipe.complexity, pipe.rule_string) < (reps[key][0].complexity, reps[key][0].rule_string): reps[key] = (pipe, pre)
    models = []
    for pipe, pre in sorted(reps.values(), key=lambda x: x[0].pipeline_id):
        models.extend(evaluate(labels, terms, pipe, pre, state) for state in search_pipeline(labels, terms, pipe, pre))
    unique = {}
    for model in models:
        key = (model["pipeline"].rule_string, model["mapping"])
        if key not in unique or (-model["score"], -model["train_coverage"]) < (-unique[key]["score"], -unique[key]["train_coverage"]): unique[key] = model
    models = sorted(unique.values(), key=lambda x: (-x["score"], -x["train_coverage"], x["mapping_size"], x["pipeline"].rule_string, x["mapping"]))
    if details: return models[:m1.RETAINED]
    best = dict(models[0]); best["search_max_train_coverage"] = max(x["train_coverage"] for x in models); return [best]


def heldout(model: dict[str, object], labels: list[dict[str, str]], terms: list[dict[str, object]]) -> dict[str, object]:
    adj, evidence = m1.adjacency(labels, terms, model["pre"], model["mapping"])
    match = {r["stolfi_coordinate"]: sorted(adj[r["stolfi_coordinate"]])[0] for r in labels if adj.get(r["stolfi_coordinate"])}
    return {"held_adjacency": adj, "held_evidence": evidence, "held_match": match, "heldout_coverage": len(match)/len(labels),
            "held_reuse_excess": reuse_excess(tuple((lid, oid, "", ()) for lid, oid in match.items()))}


def breakdown(match: dict, labels: list[dict[str, str]], cls: str) -> tuple[int, int, float]:
    rows = [r for r in labels if r["object_class"] == cls]; n = sum(r["stolfi_coordinate"] in match for r in rows)
    return n, len(rows), n/len(rows)


def worker(task: tuple[str, int]) -> dict[str, object]:
    control, rep = task; rng = random.Random(m1.seed_for(control, rep)); labels = m1.load_labels("TRAIN"); terms = RUN_TERMS
    if control == "RANDOM_VOYNICH_SET": labels = m1.randomize_labels(labels, rng)
    elif control == "SHUFFLED_TERMS": terms = m1.alter_terms(terms, rng, False)
    else: terms = m1.alter_terms(terms, rng, True)
    best = run_search(labels, terms, False)[0]
    star = breakdown(best["train_match"], labels, "STAR")[2]; planet = breakdown(best["train_match"], labels, "PLANET_MOON")[2]
    return {"matching_model":"REUSE_ALLOWED", "control":control, "replicate":rep, "seed":m1.seed_for(control,rep),
            "max_score":best["score"], "max_train_coverage":best["search_max_train_coverage"],
            "selected_train_coverage":best["train_coverage"], "star_coverage":star, "planet_moon_coverage":planet,
            "mapping_size":best["mapping_size"], "reuse_excess":best["reuse_excess"], "rule":best["pipeline"].rule_string}


def main() -> None:
    ap=argparse.ArgumentParser(); ap.add_argument("--workers",type=int,default=min(12,os.cpu_count() or 1)); ap.add_argument("--replicates",type=int,default=100); ap.add_argument("--observed-only",action="store_true"); args=ap.parse_args()
    global RUN_TERMS; RUN_TERMS=load_terms(); train=m1.load_labels("TRAIN"); held=m1.load_labels("HELD_OUT")
    models=run_search(train,RUN_TERMS,True)
    if args.observed_only:
        b=models[0]; print(b["train_coverage"],b["score"],b["mapping_size"],b["reuse_excess"],b["pipeline"].rule_string); return
    tasks=[(c,r) for c in CONTROLS for r in range(args.replicates)]
    with get_context("fork").Pool(args.workers) as pool: null=list(pool.imap(worker,tasks,chunksize=1))
    by=defaultdict(list)
    for row in null: by[row["control"]].append(row)
    mean=max(statistics.fmean(r["max_train_coverage"] for r in rows) for rows in by.values())
    for i,model in enumerate(models,1):
        model.update(heldout(model,held,RUN_TERMS)); model["model_id"]=f"REUSE_{i:03d}"; model["null_advantage"]=model["train_coverage"]-mean
        model["empirical_p"]=max((1+sum(r["max_score"]>=model["score"] for r in rows))/(len(rows)+1) for rows in by.values())
    best=models[0]
    baseline_models=read_tsv(D1/"D1_TOKEN_FORMATION_MODELS.tsv"); baseline=baseline_models[0]
    base_null=read_tsv(D1/"D1_NULL_RESULTS.tsv")
    base_mean=max(statistics.fmean(float(r["max_train_coverage"]) for r in base_null if r["control"]==c) for c in CONTROLS)
    rows=[]
    rows.append({"matching_model":"ANONYMOUS_SET","applicability":"APPLICABLE","interpretation":"NULL_COMPATIBLE","train_coverage":baseline["train_coverage"],"heldout_coverage":baseline["heldout_coverage"],"star_train_coverage":baseline["star_train_coverage"],"star_heldout_coverage":baseline["star_heldout_coverage"],"planet_moon_train_coverage":baseline["planet_moon_train_coverage"],"planet_moon_heldout_coverage":baseline["planet_moon_heldout_coverage"],"null_mean_maximum":fmt(base_mean),"null_p95":fmt(sorted(float(r["max_train_coverage"]) for r in base_null)[284]),"null_maximum":fmt(max(float(r["max_train_coverage"]) for r in base_null)),"empirical_p":baseline["empirical_p"],"real_null_advantage":baseline["null_advantage"],"mapping_size":baseline["mapping_size"],"total_complexity":baseline["total_complexity"],"assignment_ambiguity":baseline.get("ambiguity_excess","")})
    bs=breakdown(best["train_match"],train,"STAR"); bp=breakdown(best["train_match"],train,"PLANET_MOON"); hs=breakdown(best["held_match"],held,"STAR"); hp=breakdown(best["held_match"],held,"PLANET_MOON")
    interp="STRUCTURAL_HINT" if best["null_advantage"]>float(baseline["null_advantage"])+.05 else "STRUCTURE_HURTS" if best["null_advantage"]<float(baseline["null_advantage"]) else "NULL_COMPATIBLE"
    vals=[r["max_train_coverage"] for r in null]
    rows.append({"matching_model":"REUSE_ALLOWED","applicability":"APPLICABLE","interpretation":interp,"train_coverage":fmt(best["train_coverage"]),"heldout_coverage":fmt(best["heldout_coverage"]),"star_train_coverage":fmt(bs[2]),"star_heldout_coverage":fmt(hs[2]),"planet_moon_train_coverage":fmt(bp[2]),"planet_moon_heldout_coverage":fmt(hp[2]),"null_mean_maximum":fmt(mean),"null_p95":fmt(sorted(vals)[284]),"null_maximum":fmt(max(vals)),"empirical_p":fmt(best["empirical_p"]),"real_null_advantage":fmt(best["null_advantage"]),"mapping_size":best["mapping_size"],"total_complexity":best["total_complexity"],"assignment_ambiguity":fmt(best["ambiguity_excess"])})
    for name in ("ORDER_PRESERVING","CYCLIC_ORDER","GROUP_CONSTRAINED","SEQUENCE_ALIGNMENT"):
        rows.append({"matching_model":name,"applicability":"NOT_APPLICABLE","interpretation":"NOT_APPLICABLE","train_coverage":"NA","heldout_coverage":"NA","star_train_coverage":"NA","star_heldout_coverage":"NA","planet_moon_train_coverage":"NA","planet_moon_heldout_coverage":"NA","null_mean_maximum":"NA","null_p95":"NA","null_maximum":"NA","empirical_p":"NA","real_null_advantage":"NA","mapping_size":"NA","total_complexity":"NA","assignment_ambiguity":"NA"})
    fields=list(rows[0]); write_tsv(OUT/"MATCHING_MODEL_RESULTS.tsv",fields,rows)
    write_tsv(OUT/"MATCHING_MODEL_STAR_RESULTS.tsv",fields,rows); write_tsv(OUT/"MATCHING_MODEL_PLANET_MOON_RESULTS.tsv",fields,rows)
    nullout=[{**r,"max_score":fmt(r["max_score"]),"max_train_coverage":fmt(r["max_train_coverage"]),"selected_train_coverage":fmt(r["selected_train_coverage"]),"star_coverage":fmt(r["star_coverage"]),"planet_moon_coverage":fmt(r["planet_moon_coverage"])} for r in null]
    write_tsv(OUT/"MATCHING_MODEL_NULL_RESULTS.tsv",list(nullout[0]),nullout)
    heldrows=[]; assignments=[]
    for model in models:
        for r in held: heldrows.append({"matching_model":"REUSE_ALLOWED","model_id":model["model_id"],"label":r["stolfi_coordinate"],"object_class":r["object_class"],"prediction":"MATCHED" if r["stolfi_coordinate"] in model["held_match"] else "UNEXPLAINED","concept":model["held_match"].get(r["stolfi_coordinate"],"")})
        for split, labels, match in (("TRAIN",train,model["train_match"]),("HELD_OUT",held,model["held_match"])):
            for r in labels:
                lid=r["stolfi_coordinate"]
                if lid in match: assignments.append({"matching_model":"REUSE_ALLOWED","model_id":model["model_id"],"split":split,"label":lid,"object_class":r["object_class"],"concept":match[lid],"reused_count":sum(x==match[lid] for x in match.values())})
    write_tsv(OUT/"MATCHING_MODEL_HELDOUT.tsv",["matching_model","model_id","label","object_class","prediction","concept"],heldrows)
    write_tsv(OUT/"MATCHING_MODEL_ASSIGNMENTS.tsv",["matching_model","model_id","split","label","object_class","concept","reused_count"],assignments)
    status="STRUCTURAL_HINT" if interp=="STRUCTURAL_HINT" else "ANONYMOUS_CONFIRMED"
    report=f"""# Alternative astronomical matching models\n\nThe structural audit found no independently documented, commensurable order or grouping for the frozen STAR sample. Stolfi coordinates and transcription positions identify labels but do not establish clockwise/radial traversal. Historical rete order therefore cannot be aligned to them without an arbitrary bridge. ORDER_PRESERVING, CYCLIC_ORDER, GROUP_CONSTRAINED, and SEQUENCE_ALIGNMENT are `NOT_APPLICABLE`.\n\nREUSE_ALLOWED was tested with the complete frozen M1 optimiser and {args.replicates} replicates of each search-level null. Its reuse penalty was frozen at {REUSE_PENALTY:.2f} per assignment beyond the first use of a concept. It achieved {best['train_coverage']:.6f} TRAIN, {best['heldout_coverage']:.6f} HELD_OUT, advantage {best['null_advantage']:.6f}, and p={best['empirical_p']:.6f}; interpretation `{interp}`. PLANET_MOON remains exploratory because the diagram positions are not confirmed planet identities.\n\n```text\nASTRO_MATCHING_MODEL_TEST={status}\nBEST_MATCHING_MODEL={'REUSE_ALLOWED' if best['null_advantage']>float(baseline['null_advantage']) else 'ANONYMOUS_SET'}\nBEST_STAR_TRAIN_COVERAGE={max(float(baseline['star_train_coverage']),bs[2]):.6f}\nBEST_STAR_HELDOUT_COVERAGE={max(float(baseline['star_heldout_coverage']),hs[2]):.6f}\nBEST_REAL_NULL_ADVANTAGE={max(float(baseline['null_advantage']),best['null_advantage']):.6f}\n```\n"""
    (OUT/"MATCHING_MODEL_COMPARISON.md").write_text(report,encoding="utf-8")
    artifacts=[p.name for p in OUT.iterdir() if p.is_file() and p.name not in {"main.py","manifest.json","SHA256SUMS"}]
    manifest={"experiment":"astro-matching-models-v1","generated_utc":datetime.now(timezone.utc).replace(microsecond=0).isoformat(),"seed":m1.SEED,"replicates_per_control":args.replicates,"reuse_penalty":REUSE_PENALTY,"input_sha256":{"D1_TOKEN_FORMATION_MODELS.tsv":sha(D1/"D1_TOKEN_FORMATION_MODELS.tsv"),"D1_NULL_RESULTS.tsv":sha(D1/"D1_NULL_RESULTS.tsv"),"ASTRO_TERM_CORPUS_EXPANDED.tsv":sha(D1/"ASTRO_TERM_CORPUS_EXPANDED.tsv"),"ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv":sha(ROOT/"research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv"),"main.py":sha(Path(__file__))},"artifact_sha256":{n:sha(OUT/n) for n in artifacts}}
    (OUT/"manifest.json").write_text(json.dumps(manifest,indent=2,sort_keys=True)+"\n",encoding="utf-8")
    names=artifacts+["main.py","manifest.json"]; (OUT/"SHA256SUMS").write_text("".join(f"{sha(OUT/n)}  {n}\n" for n in sorted(names)),encoding="utf-8")

if __name__=="__main__": main()
