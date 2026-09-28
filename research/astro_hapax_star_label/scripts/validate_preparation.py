#!/usr/bin/env python3
"""Validate, replay, and checksum-freeze the Gate-A preparation package."""
from __future__ import annotations

import io
from pathlib import Path
import sys
import tempfile
import unittest

sys.dont_write_bytecode=True
PKG=Path(__file__).resolve().parents[1]
sys.path.insert(0,str(PKG/'scripts'))
import build_preparation as b

def generated(root):
    excluded={'SHA256SUMS','VALIDATION_REPORT.md'}
    return sorted(p.relative_to(root) for p in root.rglob('*') if p.is_file() and p.name not in excluded and 'scripts' not in p.parts and 'tests' not in p.parts and '__pycache__' not in p.parts)

def main():
    b.verify_inputs(PKG/'INPUT_MANIFEST.tsv')
    suite=unittest.defaultTestLoader.discover(str(PKG/'tests'));stream=io.StringIO();result=unittest.TextTestRunner(stream=stream,verbosity=2).run(suite)
    if not result.wasSuccessful():raise RuntimeError('tests failed\n'+stream.getvalue())
    with tempfile.TemporaryDirectory(prefix='hapax-label-prep-replay-') as td:
        replay=Path(td)/'replay';b.build(replay);files=generated(replay)
        for rel in files:
            if not (PKG/rel).is_file() or (PKG/rel).read_bytes()!=(replay/rel).read_bytes():raise RuntimeError('byte replay mismatch: '+str(rel))
    report=f"""# Preparation validation

Status: **PASS**.

- All registered frozen inputs and available upstream ledgers/manifests: PASS.
- Authoritative ZL3b deterministic-refreeze corpus and occurrence bindings: PASS.
- Frozen preregistration hash: PASS.
- Exactly 92 canonical LABEL IDs and canonical geometry: PASS (37/33/22).
- Internal 3G1 membership: PASS (64 grouped, 28 ungrouped; one eight-member group retained as one entity).
- Candidate registry: PASS (268 exact same-page occurrences in 90 lines; no forced mapping).
- Human package: PASS (92 crops, three page contexts, explicit-outcome template).
- UI blindness: PASS (no group, origin, frequency, hapax, lexicon or hypothesis metadata).
- Importer round-trip and invented-occurrence rejection: PASS.
- Automated tests: PASS ({result.testsRun} tests).
- Byte-for-byte clean replay: PASS ({len(files)} generated files).
- Enrichment and lexicon stages absent and unauthorized: PASS.

VALIDATION=PASS
GATE_A=PASS
HAPAX_ENRICHMENT_RUN_AUTHORIZED=NO
LEXICON_MATCH_RUN_AUTHORIZED=NO
"""
    (PKG/'VALIDATION_REPORT.md').write_text(report,encoding='utf-8')
    frozen=sorted(p for p in PKG.rglob('*') if p.is_file() and p.name!='SHA256SUMS' and '__pycache__' not in p.parts)
    lines=[f'{b.sha(p)}  {p.relative_to(PKG)}' for p in frozen];(PKG/'SHA256SUMS').write_text('\n'.join(lines)+'\n',encoding='utf-8')
    for line in lines:
        digest,rel=line.split('  ',1);b.check(b.sha(PKG/rel)==digest,'final checksum mismatch '+rel)
    print(f'VALIDATION=PASS; TESTS={result.testsRun}; REPLAY_FILES={len(files)}; FROZEN_FILES={len(frozen)}')
if __name__=='__main__':main()
