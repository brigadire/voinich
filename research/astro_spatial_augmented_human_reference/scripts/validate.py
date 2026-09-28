#!/usr/bin/env python3
"""Validate and freeze the augmented reference package."""
from __future__ import annotations

import argparse
import io
from pathlib import Path
import shutil
import sys
import tempfile
import unittest

sys.dont_write_bytecode=True
PKG=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(PKG/'scripts'))
import analyze

def package_files(root):
    excluded={'SHA256SUMS','VALIDATION_REPORT.md'}
    return sorted(p.relative_to(root) for p in root.rglob('*') if p.is_file() and p.name not in excluded and '__pycache__' not in p.parts)

def main():
    parser=argparse.ArgumentParser();parser.add_argument('--package-dir',type=Path,default=PKG);args=parser.parse_args()
    root=args.package_dir.resolve()
    analyze.verify_inputs(root/'INPUT_MANIFEST.tsv')

    suite=unittest.defaultTestLoader.discover(str(root/'tests'))
    stream=io.StringIO();result=unittest.TextTestRunner(stream=stream,verbosity=2).run(suite)
    if not result.wasSuccessful():raise RuntimeError('automated tests failed\n'+stream.getvalue())

    with tempfile.TemporaryDirectory(prefix='augmented-reference-replay-') as td:
        replay=Path(td)/'replay';analyze.build(replay)
        replay_files=package_files(replay)
        for rel in replay_files:
            target=root/rel
            if not target.is_file() or target.read_bytes()!=(replay/rel).read_bytes():
                raise RuntimeError(f'byte reproduction mismatch: {rel}')

    report=f'''# Validation report

Status: **PASS**.

- Registered frozen inputs and every upstream checksum ledger: PASS.
- Augmented object membership, canonical IDs, panels, classes and final geometry: PASS (557 = 520 prior confirmed + 37 confirmed additions).
- Relation endpoint snapshot: PASS (256 objects, each present once and byte-equivalent in canonical fields).
- Human-added provenance reconciliation: PASS (37 audited; 28 flagged overlap records affecting 23 objects; no AI support transferred).
- Assisted human-reference recall numerators, denominators, Wilson intervals and deterministic panel bootstrap: PASS.
- Reviewed 3G1 groups and source audit: PASS (64 = 29/24/11; 63 size-2 + one size-8; 45 unchanged + 19 new/modified).
- Multi-member group preserved without Cartesian expansion: PASS.
- Ungrouped object `HNEW_STAR_f68r2_DC27A556209F27B3` preserved as `NO_VISUAL_GROUP_ASSIGNED`: PASS.
- Frozen attachment-v2 and prior report hashes: PASS.
- Automated tests: PASS ({result.testsRun} tests).
- Clean replay of {len(replay_files)} generated analytical/figure files compared byte-for-byte: PASS.

VALIDATION=PASS
BYTE_REPRODUCIBILITY=PASS
UPSTREAM_CHECKSUMS=PASS
'''
    (root/'VALIDATION_REPORT.md').write_text(report,encoding='utf-8')
    files=sorted(p for p in root.rglob('*') if p.is_file() and p.name!='SHA256SUMS' and '__pycache__' not in p.parts)
    lines=[f'{analyze.digest(p)}  {p.relative_to(root)}' for p in files]
    (root/'SHA256SUMS').write_text('\n'.join(lines)+'\n',encoding='utf-8')
    for line in lines:
        expected,rel=line.split('  ',1)
        if analyze.digest(root/rel)!=expected:raise RuntimeError('final checksum mismatch: '+rel)
    print(f'VALIDATION=PASS; TESTS={result.testsRun}; REPLAY_FILES={len(replay_files)}; FROZEN_FILES={len(files)}')

if __name__=='__main__':main()
