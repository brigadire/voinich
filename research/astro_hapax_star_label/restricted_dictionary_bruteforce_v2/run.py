#!/usr/bin/env python3
"""Production gate for v2; preparation is intentionally separate from scoring."""
import argparse,json,hashlib,platform,time
from pathlib import Path
from engine import global_systems
HERE=Path(__file__).resolve().parent
def freeze():
    if (HERE/'FREEZE.json').exists():raise SystemExit('already frozen')
    files={str(p.relative_to(HERE)):hashlib.sha256(p.read_bytes()).hexdigest() for p in HERE.iterdir() if p.is_file() and p.name!='FREEZE.json'}
    (HERE/'FREEZE.json').write_text(json.dumps({'version':2,'files':files,'python':platform.python_version(),'utc':time.strftime('%Y-%m-%dT%H:%M:%SZ',time.gmtime())},indent=2)+'\n')
    print('v2 frozen; production scores withheld until run')
def verify():
    f=json.loads((HERE/'FREEZE.json').read_text())
    for n,h in f['files'].items():
        assert hashlib.sha256((HERE/n).read_bytes()).hexdigest()==h,n
def main():
    ap=argparse.ArgumentParser();ap.add_argument('action',choices=('freeze','verify','run'));a=ap.parse_args()
    if a.action=='freeze':freeze();return
    verify()
    if a.action=='verify':print('v2 frozen inputs verified');return
    raise SystemExit('BLOCKED: v2 production runner is deliberately gated; implement and preregister full null checkpoint execution before scoring')
if __name__=='__main__':main()
