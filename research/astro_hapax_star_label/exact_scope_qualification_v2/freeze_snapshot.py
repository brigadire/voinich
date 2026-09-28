#!/usr/bin/env python3
"""Create deterministic content-addressed freeze records before sealed generation."""
from __future__ import annotations
import hashlib, json, platform, tarfile
from pathlib import Path
OUT=Path(__file__).resolve().parent
FROZEN=[OUT/"EXACT_SCOPE_SPEC.json",OUT/"PREREGISTERED_GATES.json",OUT/"DEPENDENCY_LOCK.json",OUT/"QUALIFICATION_PROTOCOL.md",OUT/"generate_sealed.py",OUT/"run_qualification.py",OUT/"score_qualification.py"]+sorted((OUT/"snapshot").glob("*.py"))
def sha(p): return hashlib.sha256(p.read_bytes()).hexdigest()
def main():
    rows=[(str(p.relative_to(OUT)),sha(p)) for p in FROZEN]
    canonical="".join(f"{n}\t{h}\n" for n,h in rows); tree=hashlib.sha256(canonical.encode()).hexdigest()
    (OUT/"CANONICAL_TREE_SHA256.txt").write_text(tree+"\n"+canonical)
    archive=OUT/"CP_SAT_SNAPSHOT_V2.tar"; 
    with tarfile.open(archive,"w") as tf:
        for p in sorted((OUT/"snapshot").glob("*.py")):
            info=tf.gettarinfo(str(p),arcname=f"snapshot/{p.name}"); info.uid=0;info.gid=0;info.uname="";info.gname="";info.mtime=0
            with p.open("rb") as fh: tf.addfile(info,fh)
    manifest={"snapshot_id":"EXACT_SCOPE_V2","content_bound":True,"git_commit_binding":"UNAVAILABLE_NOT_REQUIRED","canonical_tree_sha256":tree,"archive_sha256":sha(archive),"files":dict(rows),"python":platform.python_version(),"real_data_accessed":False,"sealed_run_started":False}
    (OUT/"SNAPSHOT_FREEZE_MANIFEST.json").write_text(json.dumps(manifest,indent=2,sort_keys=True)+"\n")
    print(json.dumps(manifest,indent=2))
if __name__=="__main__": main()
