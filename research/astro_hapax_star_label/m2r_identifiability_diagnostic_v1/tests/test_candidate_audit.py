"""Unit tests for candidate generation audit."""
import sys
import unittest
from pathlib import Path

PACKAGE = Path(__file__).resolve().parent.parent
sys.path.insert(0, str(PACKAGE))
import audit_engine


class TestCandidateAudit(unittest.TestCase):
    def test_candidate_generation_support_filter(self):
        # 2 terms with 'ba', 1 with 'ca' -> 'ba' has support 2, 'ca' has support 1
        ts = ['bada', 'bafe', 'cada']
        ls = ['xypq', 'xyrs', 'wzpq']
        cfg = audit_engine.Config(pool=48)
        budget = audit_engine.Budget(cfg)
        
        pool_sup3, _ = audit_engine.candidates(ts, ls, cfg, budget, min_support=3)
        self.assertEqual(len(pool_sup3), 0)
        
        pool_sup2, _ = audit_engine.candidates(ts, ls, cfg, budget, min_support=2)
        self.assertGreater(len(pool_sup2), 0)
        # Verify 'ba' is paired with 'xy'
        inits = [r for r in pool_sup2 if r[0] == 'INITIAL']
        self.assertTrue(any(r[1] == 'ba' and r[2] == 'xy' for r in inits))

    def test_oracle_rule_forcing(self):
        ts = ['bada', 'bafe', 'cada']
        ls = ['xypq', 'xyrs', 'wzpq']
        cfg = audit_engine.Config(pool=48)
        budget = audit_engine.Budget(cfg)
        forced = [('INITIAL', 'forced_src', 'forced_tgt')]
        pool, _ = audit_engine.candidates(ts, ls, cfg, budget, force_rules=forced, min_support=3)
        self.assertIn(('INITIAL', 'forced_src', 'forced_tgt'), pool)


if __name__ == '__main__':
    unittest.main()
