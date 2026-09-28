import unittest
import engine

class V2Tests(unittest.TestCase):
    def test_table_is_global_and_bounded(self):
        tables=engine.table_candidates('okdaly',beam=256)
        self.assertEqual(len(tables),256)
        self.assertTrue(all(len(t[0])<=4 and t[2]<=4 for t in tables))
    def test_longest_grapheme_first(self):
        self.assertEqual(engine.encode('ghal',(('gh','x'),('a','y')),False),'xyl')
    def test_one_identity_capacity(self):
        terms=[{'identity':'A','forms':['alpha','beta']}]
        idx=engine.index(terms,{},(('a','x'),),False)
        self.assertLessEqual(len(engine.match([{'occurrence_id':'1','token':'alpha'},{'occurrence_id':'2','token':'beta'}],idx)),1)
    def test_no_singleton_rule_metadata(self):
        for system in engine.global_systems('abc',beam=8):
            self.assertLessEqual(system[-1],8)
    def test_unmapped_policy_is_global(self):
        self.assertEqual(engine.encode('abc',(('a','x'),),False),'xbc')
        self.assertEqual(engine.encode('abc',(('a','x'),),True),'x')

if __name__=='__main__':unittest.main()
