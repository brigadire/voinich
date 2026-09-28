#!/usr/bin/env python3
"""Generate the new sealed public/truth pair. No solver is imported or run."""
from __future__ import annotations
import hashlib, json, random, sys
from pathlib import Path

OUT = Path(__file__).resolve().parent
SEED_ROOTS = {"zero_noise": 20261401, "noisy": 20261402, "hard_negative": 20261403}

def sha(p):
    h=hashlib.sha256(); h.update(p.read_bytes()); return h.hexdigest()

def case(seed, i, noise, hard):
    r=random.Random(seed); source=list("abcdefgh"); target=list("acdefhi"); chosen=r.sample(source,4)
    mode=r.choice(["INJECTIVE","MERGE_1"]); table={}; used=[]
    for c in chosen:
        if mode=="MERGE_1" and used and r.random()<.25: table[c]=r.choice(used)
        else:
            avail=[x for x in target if mode=="MERGE_1" or x not in used]
            table[c]=r.choice(avail); used.append(table[c])
    lexicon={f"ID{j}": {"".join(r.choice(source) for _ in range(r.randint(3,5)))} for j in range(5)}
    labels=[]; truths={}
    for j in range(8):
        ident=r.choice(list(lexicon)); form=next(iter(lexicon[ident])); token="".join(table[c] for c in form if c in table)
        unmatched=hard or r.random()<(.18 if noise else 0)
        if unmatched: token="".join(r.choice(target) for _ in range(r.randint(2,6))); truths[f"I{i:03d}_O{j:02d}"]=None
        else: truths[f"I{i:03d}_O{j:02d}"]=ident
        labels.append({"occurrence_id":f"I{i:03d}_O{j:02d}","page_id":f"P{j//8}","token":token})
    return {"instance_id":f"I{i:03d}","seed":seed,"lexicon":{k:sorted(v) for k,v in lexicon.items()},"labels":labels,"mode":mode,"capacity":"PER_PAGE_CAPACITY_1","noise":noise,"hard_negative":hard}, {"instance_id":f"I{i:03d}","mapping":table,"truth_by_occurrence":truths}

def main():
    public=[]; truth=[]
    for group, root in SEED_ROOTS.items():
        for i in range(30):
            p,t=case(root+i, len(public), .20 if group=="noisy" else 0.0, group=="hard_negative"); p["group"]=group; t["group"]=group; public.append(p); truth.append(t)
    (OUT/"SEALED_PUBLIC.jsonl").write_text("".join(json.dumps(x,sort_keys=True)+"\n" for x in public))
    (OUT/"SEALED_TRUTH.jsonl").write_text("".join(json.dumps(x,sort_keys=True)+"\n" for x in truth))
    manifest={"seed_roots":SEED_ROOTS,"counts":{"zero_noise":30,"noisy":30,"hard_negative":30},"instance_scale":{"identities":5,"labels":8,"train":6,"held_out":2},"solver_run_before_manifest":"NO","real_data_accessed":False}
    (OUT/"SEALED_DATASET_MANIFEST.json").write_text(json.dumps({**manifest,"public_sha256":sha(OUT/"SEALED_PUBLIC.jsonl"),"truth_sha256":sha(OUT/"SEALED_TRUTH.jsonl")},indent=2,sort_keys=True)+"\n")
    print("generated",len(public),"cases; solver not run")
if __name__=="__main__": main()
