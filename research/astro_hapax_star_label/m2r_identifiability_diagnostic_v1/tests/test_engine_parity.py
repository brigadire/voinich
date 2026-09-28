"""Unit tests for audit_engine parity with m2r_real_engine_v2."""
import sys
import unittest
from pathlib import Path

PACKAGE = Path(__file__).resolve().parent.parent
REPO = PACKAGE.parents[2]
sys.path.insert(0, str(PACKAGE))
sys.path.insert(0, str(REPO / 'research/astro_hapax_star_label/m2r_real_engine_v2'))

import audit_engine
import engine
from tests.test_engine import fixture, RULES


class TestEngineParity(unittest.TestCase):
    def test_parity_across_sizes(self):
        for n in (29, 57, 68):
            t_rows, l_rows = fixture(n)
            ts, _ = engine.load_rows(t_rows)
            ls, _ = engine.load_rows(l_rows)
            
            m_v2 = engine.evaluate(ts, ls, RULES)
            m_audit = audit_engine.evaluate(ts, ls, RULES)
            
            self.assertEqual(m_v2['baseline'], m_audit['baseline'])
            self.assertEqual(m_v2['total'], m_audit['total'])
            self.assertAlmostEqual(m_v2['gain'], m_audit['gain'], places=12)
            self.assertEqual(m_v2['components'], m_audit['components'])
            self.assertEqual(m_v2['valid_support'], m_audit['valid_support'])
            self.assertEqual(len(m_v2['assignments']), len(m_audit['assignments']))

    def test_literal_encoding(self):
        self.assertEqual(audit_engine.literal("test"), 32)
        self.assertEqual(audit_engine.literal("αβ"), 32)  # 2 UTF-8 bytes each

    def test_role_boundaries(self):
        self.assertEqual(audit_engine.role(0, 4, 4), 'WHOLE')
        self.assertEqual(audit_engine.role(0, 2, 4), 'INITIAL')
        self.assertEqual(audit_engine.role(2, 4, 4), 'FINAL')
        self.assertEqual(audit_engine.role(1, 3, 4), 'MEDIAL')


if __name__ == '__main__':
    unittest.main()
