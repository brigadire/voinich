#!/usr/bin/env python3
"""Synthetic/copy-tree preparation validation, full replay and own-output freeze."""
import io
from pathlib import Path
import subprocess
import tempfile
import unittest
from core import PKG,check,digest,verify_inputs
from build_package import build

def freeze():
    files=sorted(p for p in PKG.rglob('*') if p.is_file() and p.name!='SHA256SUMS' and '__pycache__' not in p.parts)
    (PKG/'SHA256SUMS').write_text(''.join(f'{digest(p)}  {p.relative_to(PKG)}\n' for p in files),encoding='utf-8')

def main():
    print('Testing native XML copied-tree/two-decimal roundtrip and protection gates...',flush=True)
    output=io.StringIO();suite=unittest.defaultTestLoader.discover(str(PKG/'tests'))
    result=unittest.TextTestRunner(stream=output,verbosity=2).run(suite);print(output.getvalue(),flush=True)
    check(result.wasSuccessful(),'completeness preparation tests failed')
    print('Full isolated regeneration, including ZIPs/maps/dry-run package...',flush=True)
    count=0
    with tempfile.TemporaryDirectory(prefix='completeness-preparation-replay-') as tmp:
        build(Path(tmp))
        for path in sorted(Path(tmp).rglob('*')):
            if not path.is_file():continue
            baseline=PKG/path.relative_to(tmp)
            check(baseline.is_file() and digest(path)==digest(baseline),f'regeneration differs: {path.relative_to(tmp)}')
            count+=1
    verify_inputs(full=True)
    report=f'''# Completeness preparation validation

Status PASS. Tests run: {result.testsRun}; failures={len(result.failures)}, errors={len(result.errors)}.
All tests use prepared/synthetic copied trees, not actual new reviewer annotations.
Full isolated regeneration: {count} files byte-identical, including native XML, deterministic
image/annotation ZIPs, final-reference TSVs, navigation/eligibility maps, protocols, manifests
and current-reference attachment dry run. Upstream + comparison input hashes rechecked after
generation. Only this preparation directory is frozen by SHA256SUMS.

PASS: final human geometry, unchanged IDs/provenance, blindness, optional UNCERTAIN layer,
hidden REJECT/NOT_REVIEWED, exact panel/image coordinates, explicit completion, deterministic
new IDs independent of CVAT shape IDs, no auto acceptance, four prior-overlap queue classes,
mutated/deleted/relabeled/duplicate references blocked, unset confidence forbidden.
Copied-tree roundtrip includes CVAT native ellipse/rotated-box/tag serialization rounded to
two decimals, tolerance 0.0051 on serialized fields. No canonical reference is replaced by
rounded export geometry. All panel markers start NOT_STARTED in actual annotation package.

PASS: f68r1/f68r2 eligibility, unconfirmed f68r3 excluded, non-applicable panels not iterated,
filtered pairs not observations, v2 retained without conversion, v3 version/spatial-field
validation, gated future production blocked without external object freeze, endpoints protected.
NOT_APPLICABLE_REGION becomes a reconciliation question, not a negative relation.

CVAT live instance/version was unavailable; no CVAT writes/review occurred. Roundtrip PASS
means native XML copied-tree compatibility, NOT a claimed live-server integration test.
Empty live import/export preflight remains necessary on reviewer instance (guides/schema).
Lock/hide are optional UI protections; strict reference comparison is mandatory fallback.

HUMAN_REVIEW_STARTED=NO
REAL_HUMAN_EXPORT_IMPORTED=NO
NEW_OBJECTS_AUTO_ACCEPTED=0
NEW_AI_RECALL_CALCULATED=NO
PRODUCTION_ATTACHMENT_PACKAGE_CREATED=NO
FROZEN_INPUTS_UNCHANGED=YES
RESULTS_REPRODUCIBLE=YES
'''
    (PKG/'VALIDATION_REPORT.md').write_text(report,encoding='utf-8');freeze()
    check(subprocess.run(['sha256sum','-c','--quiet','SHA256SUMS'],cwd=PKG).returncode==0,'output checksum failure')
    print(f'VALIDATION=PASS; TESTS={result.testsRun}; REPLAY_FILES={count}; OUTPUT_CHECKSUMS=PASS',flush=True)

if __name__=='__main__':main()
