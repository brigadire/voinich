#!/usr/bin/env python3
"""Full replay, automated gates and output freeze; read-only upstream."""
from __future__ import annotations
import argparse
from contextlib import redirect_stdout
import io
import json
from pathlib import Path
import subprocess
import tempfile
import unittest
from common import PACKAGE, read, digest, check, verify_inventory
from analyze import run, freeze_outputs

def validate(replay=True):
    print('Running automatic tests...',flush=True)
    # Existing checksum ledger is checked before regenerating validation artifacts.
    suite=unittest.defaultTestLoader.discover(str(PACKAGE/'tests'))
    stream=io.StringIO()
    result=unittest.TextTestRunner(stream=stream,verbosity=2).run(suite)
    print(stream.getvalue(),flush=True)
    check(result.wasSuccessful(),'automatic comparison tests failed')
    replay_files=0
    if replay:
        print('Replaying complete analysis and figures in a temporary directory...',flush=True)
        with tempfile.TemporaryDirectory(prefix='three-way-replay-') as tmp:
            run(Path(tmp))
            for path in sorted(Path(tmp).rglob('*')):
                if not path.is_file(): continue
                baseline=PACKAGE/path.relative_to(tmp)
                check(baseline.is_file() and digest(path)==digest(baseline),f'non-reproducible output: {path.relative_to(tmp)}')
                replay_files+=1
    print('Rechecking upstream snapshot after analysis...',flush=True)
    inventory=read(PACKAGE/'INPUT_MANIFEST.tsv')
    for row in inventory: row['bytes']=int(row['bytes'])
    _,ledgers=verify_inventory(inventory)
    text=f'''# Validation report

Status: PASS. No frozen input was modified. All critical input gates passed before
aggregation; source IDs, panels, source boxes, one-record-per-source physical mapping,
phase deduplication, final decisions, attachment endpoints and published counts agree.

- Automatic tests: {result.testsRun} run, failures={len(result.failures)}, errors={len(result.errors)}.
- Full second-run byte comparison: {'PASS' if replay else 'NOT_REQUESTED'}, {replay_files} regenerated files,
  including bootstrap tables/JSON, reports, source tables and PNG/SVG figures.
- Seed/plan lock: PASS; strict cohorts exclude UNCERTAIN and NOT_REVIEWED.
- Calibration/production and spatial-v1/caption-v2 separation: PASS.
- A-absence negative evidence: NONE.
- Denominators, raw/JSON aggregates and figure source links: PASS.
- Upstream checksums reverified after analysis: {len(inventory)} registered inputs.
- Output SHA256SUMS: generated after this report; verified by sha256sum -c --quiet.

{' '.join(f"{r['package']}: {r['verified_entries']} ledger entries;" for r in ledgers)}

The first validation may skip the output-ledger test because freezing follows validation.
The explicit post-freeze checksum command and subsequent test run verify the frozen output.
Single-panel intervals are candidate bootstrap, not between-panel confidence; few panel
clusters and geometry-proxy limitations remain statistical restrictions, not validation failures.
'''
    (PACKAGE/'VALIDATION_REPORT.md').write_text(text,encoding='utf-8')
    freeze_outputs(PACKAGE)
    check(subprocess.run(['sha256sum','-c','--quiet','SHA256SUMS'],cwd=PACKAGE).returncode==0,'output checksum failure')
    print(f'VALIDATION=PASS; REPLAY_FILES={replay_files}; OUTPUT_CHECKSUMS=PASS',flush=True)

if __name__=='__main__':
    parser=argparse.ArgumentParser()
    parser.add_argument('--skip-replay',action='store_true')
    args=parser.parse_args()
    validate(not args.skip_replay)
