"""Synthetic orchestration: no model branches on family, seed, role or filename."""
import json
import os
from pathlib import Path
import subprocess
import sys
from engine import atomic_json, digest

PACKAGE = Path(__file__).resolve().parent


def seed_registry(rows):
    seeds, identities = set(), set()
    for row in rows:
        if row['seed'] in seeds or row['dataset'] in identities:
            raise ValueError('duplicate seed or dataset')
        seeds.add(row['seed'])
        identities.add(row['dataset'])
    return sorted(rows, key=lambda r: (r['family'], r['seed']))


def run_surface(surface, outdir, shard='0', checkpoint=False, mode=None):
    outdir = Path(outdir)
    outdir.mkdir(parents=True, exist_ok=True)
    key = digest(surface)
    path, output = outdir/(key+'.surface.json'), outdir/(key+'.output.json')
    atomic_json(path, surface)
    cp = outdir/(key+'.checkpoint.json') if checkpoint else None
    args = [sys.executable, '-I', '-B', str(PACKAGE/'worker.py'), str(path.resolve()), str(output.resolve()), str(cp.resolve()) if cp else '-']
    if mode:
        args.append('--'+mode)
    try:
        p = subprocess.run(args, capture_output=True, text=True, timeout=255,
                           env={**os.environ, 'PYTHONDONTWRITEBYTECODE': '1'})
        if p.returncode:
            if p.returncode < 0 or 'MemoryError' in p.stderr:
                result = {'train': {'SEARCH_STATUS': 'INCOMPLETE', 'reason': 'worker_resource_limit'},
                          'heldout': {'SEARCH_STATUS': 'INCOMPLETE'}, 'accepted': False}
                atomic_json(output, result)
            else:
                raise RuntimeError('isolated worker failed: '+p.stderr[-3000:])
    except subprocess.TimeoutExpired:
        atomic_json(output, {'train': {'SEARCH_STATUS': 'INCOMPLETE', 'reason': 'worker_timeout'},
                             'heldout': {'SEARCH_STATUS': 'INCOMPLETE'}, 'accepted': False})
    return {'key': key, 'shard': str(shard), 'output': json.loads(output.read_text()), 'path': str(output)}


def aggregate(rows):
    seen, result = set(), []
    for row in sorted(rows, key=lambda x: x['key']):
        if row['key'] in seen:
            raise ValueError('duplicate dataset during aggregation')
        seen.add(row['key'])
        result.append({'key': row['key'], 'output': row['output']})
    return result
