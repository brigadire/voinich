"""Complete package ledgers; SHA256SUMS necessarily excludes itself."""
import hashlib
from pathlib import Path


def file_hash(path):
    h = hashlib.sha256()
    with Path(path).open('rb') as f:
        for block in iter(lambda: f.read(1024*1024), b''):
            h.update(block)
    return h.hexdigest()


def entries(root):
    root = Path(root)
    return {p.relative_to(root).as_posix(): file_hash(p) for p in sorted(root.rglob('*'))
            if p.is_file() and p != root/'SHA256SUMS'}


def write(root):
    root = Path(root)
    (root/'SHA256SUMS').write_text(''.join(f'{h}  {p}\n' for p, h in entries(root).items()))


def verify(root):
    root = Path(root)
    expected = {}
    for line in (root/'SHA256SUMS').read_text().splitlines():
        h, p = line.split('  ', 1)
        if p in expected:
            raise ValueError('duplicate checksum path')
        expected[p] = h
    return expected == entries(root)
