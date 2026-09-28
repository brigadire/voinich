"""Unit tests for exact exhaustive search and equivalence logic."""
import sys
import unittest
from pathlib import Path

PACKAGE = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PACKAGE))
import audit_engine
import exact_search_audit


class TestExactSearch(unittest.TestCase):
    def test_small_cartesian_task_exhaustion(self):
        inits = [('ba', 'xy'), ('ce', 'wz')]
        finals = [('da', 'pq'), ('fe', 'rs')]
        all_inits = [('INITIAL', s, t) for s in ('ba', 'ce') for t in ('xy', 'wz')]
        all_finals = [('FINAL', s, t) for s in ('da', 'fe') for t in ('pq', 'rs')]
        candidate_rules = all_inits + all_finals
        pairs = [(si + sf, ti + tf) for si, ti in inits for sf, tf in finals]
        
        terms = [p[0] for p in pairs]
        labels = [p[1] for p in pairs]
        
        cfg = audit_engine.Config(pool=16)
        ex_res = audit_engine.exhaustive_search(terms, labels, candidate_rules, cfg, min_support=1, max_rules=4)
        
        self.assertGreater(ex_res['num_global_optima'], 0)
        self.assertGreater(ex_res['global_optimum']['gain'], 0.0)
        # Verify 4 global optima exist due to 2x2 permutation symmetry
        self.assertEqual(ex_res['num_global_optima'], 4)


if __name__ == '__main__':
    unittest.main()
