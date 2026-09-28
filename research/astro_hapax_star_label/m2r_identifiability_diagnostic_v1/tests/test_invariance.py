"""Unit tests for order invariance and checkpoint determinism."""
import copy
import random
import sys
import unittest
from pathlib import Path

PACKAGE = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PACKAGE))
import audit_engine


class TestInvariance(unittest.TestCase):
    def test_order_invariance(self):
        inits = [('ba', 'xy'), ('ce', 'wz'), ('di', 'uv')]
        finals = [('da', 'pq'), ('fe', 'rs'), ('go', 'jk')]
        pairs = [(si + sf, ti + tf) for si, ti in inits for sf, tf in finals]
        
        terms = [{'id': f't{i}', 'surface': p[0]} for i, p in enumerate(pairs)]
        labels = [{'id': f'l{i}', 'surface': p[1]} for i, p in enumerate(pairs)]
        
        cfg = audit_engine.Config(beam=2, iterations=3, max_states=16, pool=16)
        
        # Base
        res_base = audit_engine.search(terms, labels, cfg, min_support=3)
        h_base = audit_engine.digest(res_base['model'])
        
        # Reversed
        res_rev = audit_engine.search(list(reversed(terms)), list(reversed(labels)), cfg, min_support=3)
        h_rev = audit_engine.digest(res_rev['model'])
        
        # Permuted
        rng = random.Random(999)
        p_terms = list(terms)
        p_labels = list(labels)
        rng.shuffle(p_terms)
        rng.shuffle(p_labels)
        res_perm = audit_engine.search(p_terms, p_labels, cfg, min_support=3)
        h_perm = audit_engine.digest(res_perm['model'])
        
        self.assertEqual(h_base, h_rev)
        self.assertEqual(h_base, h_perm)


if __name__ == '__main__':
    unittest.main()
