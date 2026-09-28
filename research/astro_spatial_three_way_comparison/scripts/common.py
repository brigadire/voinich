"""Read-only frozen inputs and deterministic derived-output helpers."""
from __future__ import annotations
import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
PACKAGE = Path(__file__).resolve().parents[1]
HUMAN = ROOT / 'research/astro_spatial_human_adjudication'
AI1 = ROOT / 'research/astro_spatial_annotation_ai_b'
AI2 = ROOT / 'research/astro_spatial_annotation_ai_b2'
A = ROOT / 'research/astro_spatial_annotation'
SEED = 20260914
BOOTSTRAPS = 2000
CALIBRATION = {'f68r1', 'f68r3', 'f68v2'}

def digest(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as f:
        for chunk in iter(lambda: f.read(1024 * 1024), b''):
            h.update(chunk)
    return h.hexdigest()

def read(path):
    with Path(path).open(encoding='utf-8', newline='') as f:
        return list(csv.DictReader(f, delimiter='\t'))

def write(path, rows, fields=None):
    path = Path(path)
    path.parent.mkdir(parents=True, exist_ok=True)
    fields = fields or (list(rows[0]) if rows else [])
    with path.open('w', encoding='utf-8', newline='') as f:
        w = csv.DictWriter(f, fields, delimiter='\t', lineterminator='\n')
        w.writeheader()
        for row in rows:
            w.writerow({k: '' if row.get(k) is None else row.get(k, '') for k in fields})

def write_json(path, obj):
    Path(path).write_text(json.dumps(obj, indent=2, sort_keys=True, allow_nan=False) + '\n', encoding='utf-8')

def check(condition, message):
    if not condition:
        raise ValueError(message)

def unique(rows, key, label):
    index = {r[key]: r for r in rows}
    check(len(index) == len(rows), f'duplicate {key}: {label}')
    return index

def safe_target(out):
    out = Path(out).resolve()
    for frozen in (A, AI1, AI2, HUMAN):
        check(not out.is_relative_to(frozen.resolve()) and not frozen.resolve().is_relative_to(out),
              f'output overlaps frozen package: {out}')
    check(out != ROOT and out != Path('/'), 'unsafe output directory')
    return out

def verify_inventory(existing=None):
    """Every upstream checksum entry, plus ledgers, defines the input snapshot."""
    manifest = []
    hashes = {}
    ledgers = []
    for pkg, version in ((A, 'A-spatial-1'), (AI1, 'AI1-frozen'), (AI2, 'AI2-frozen'), (HUMAN, 'human-1.2/attachment-2')):
        ledger = pkg / 'SHA256SUMS'
        check(ledger.is_file(), f'missing checksum ledger: {ledger}')
        lines = ledger.read_text().splitlines()
        seen = set()
        for line in lines:
            expected, relative = line.split('  ', 1)
            check(relative not in seen, f'duplicate checksum target: {relative}')
            seen.add(relative)
            path = (pkg / relative).resolve()
            check(path.is_relative_to(pkg.resolve()), f'checksum path escapes package: {relative}')
            actual = digest(path)
            check(actual == expected, f'frozen checksum mismatch: {path}')
            hashes[path] = actual
        hashes[ledger.resolve()] = digest(ledger)
        ledgers.append({'package': pkg.name, 'verified_entries': len(lines), 'ledger_sha256': hashes[ledger.resolve()]})
        # Include raw images/manifests that a primary ledger might not cover.
        for extra in pkg.glob('*INPUT_MANIFEST.json'):
            data = json.loads(extra.read_text())
            for rel, meta in data.get('files', {}).items():
                if rel.startswith('crops/'):
                    path = (pkg / 'package' / rel).resolve()
                    check(path.is_file() and digest(path) == meta['sha256'], f'canonical input mismatch: {path}')
                    hashes[path] = meta['sha256']
            hashes[extra.resolve()] = digest(extra)
        for path in sorted(p for p in hashes if p.is_relative_to(pkg.resolve())):
            manifest.append({'path': str(path.relative_to(ROOT)), 'package': pkg.name,
                             'version': version, 'bytes': path.stat().st_size,
                             'sha256': hashes[path], 'role': role(path), 'frozen': 'YES'})
    index = {r['path']: r for r in manifest}
    check(len(index) == len(manifest), 'duplicate inventory paths')
    if existing is not None:
        check(manifest == existing, 'input snapshot differs from registered INPUT_MANIFEST.tsv')
    return manifest, ledgers

def role(path):
    name = path.name
    if name == 'SHA256SUMS': return 'CHECKSUM_LEDGER'
    if path.suffix in {'.jpg', '.png'}: return 'CANONICAL_OR_DERIVED_IMAGE'
    if name.endswith('MANIFEST.json') or name == 'manifest.json': return 'FROZEN_MANIFEST'
    if 'MATCH' in name: return 'FROZEN_PAIRWISE_MATCH'
    if 'CANDIDATE' in name: return 'RECONCILIATION_PROVENANCE'
    if 'ADJUDICATION.tsv' in name: return 'FINAL_HUMAN_DECISIONS_AND_GEOMETRY'
    if 'CALIBRATION' in name or 'HIGH' in name or 'CONSENSUS_QC' in name or 'MEDIUM' in name: return 'REVIEW_PHASE_RECORD'
    if 'RELATION' in name or 'ATTACHMENT' in name: return 'VERSIONED_RELATION_RECORD'
    if name in {'AI_B_OBJECTS.tsv','AI_B_LABELS.tsv','AI2_OBJECTS.tsv','AI2_LABELS.tsv','ASTRO_OBJECTS.tsv','ASTRO_LABELS_SPATIAL.tsv'}: return 'PRIMARY_SOURCE_OBJECTS'
    return 'FROZEN_PACKAGE_AUDIT_INPUT'
