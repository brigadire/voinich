import ast
import itertools
import json
import math
from pathlib import Path
import random
import subprocess
import sys
import tempfile
import unittest

ROOT = Path(__file__).resolve().parents[1]
sys.path.insert(0, str(ROOT))
import engine as e
import evaluator
import generator
import integrity
import orchestrator


def rows(strings):
    return [{'id': str(i), 'surface': s} for i, s in enumerate(strings)]


def fixture(n=4):
    return rows(['ab'+chr(0x3400+i)+'cd' for i in range(n)]), rows(['xy'+chr(0x4400+i)+'zw' for i in range(n)])


SMALL = e.Config(max_states=12, iterations=2, pool=8, additions=4, beam=2)
RULES = [('INITIAL', 'ab', 'xy'), ('FINAL', 'cd', 'zw')]


class Regression(unittest.TestCase):
    def test_no_zip_assignment(self):
        matrix = [[1, 9, 2], [8, 2, 3], [1, 4, 10]]
        self.assertEqual(e.hungarian(matrix), ([(0, 1), (1, 0), (2, 2)], 27))
        tree = ast.parse((ROOT/'engine.py').read_text())
        self.assertFalse(any(isinstance(n, ast.Call) and isinstance(n.func, ast.Name) and n.func.id=='zip' for n in ast.walk(tree)))

    def test_exact_matching_bruteforce(self):
        rng = random.Random(912)
        for n, m in [(1, 3), (3, 2), (4, 4)]:
            for _ in range(12):
                w = [[rng.randrange(-3, 12) for j in range(m)] for i in range(n)]
                pairs, score = e.hungarian(w)
                best = 0
                for a in itertools.product(range(-1, m), repeat=n):
                    selected = [x for x in a if x>=0]
                    if len(selected) != len(set(selected)):
                        continue
                    best = max(best, sum(w[i][j] for i, j in enumerate(a) if j>=0))
                self.assertEqual(score, best)
                self.assertEqual(len({i for i,j in pairs}), len(pairs))
                self.assertEqual(len({j for i,j in pairs}), len(pairs))

    def test_unmatched_and_ambiguity(self):
        self.assertEqual(e.hungarian([[-2, 0], [-1, -2]])[0], [])
        ts, ls = fixture()
        m = e.evaluate(*[e.load_rows(x)[0] for x in (ts, ls)], RULES, alternatives=True)
        self.assertTrue(m['alternative_optima'])
        self.assertEqual(min(x['margin'] for x in m['margins']), 0)

    def test_min_support_three(self):
        for n in (1, 2, 3):
            ts, ls = fixture(n)
            m = e.evaluate(*[e.load_rows(x)[0] for x in (ts, ls)], RULES)
            self.assertEqual(m['valid_support'], n>=3)
            self.assertEqual(m['objective'] is None, n<3)
            result = e.search(ts, ls, SMALL)
            self.assertTrue(all(x['pair_support']>=3 for x in result['model']['support']))

    def test_duplicate_not_independent(self):
        ts, ls = fixture(1)
        ts = [{'id':str(i), 'surface':ts[0]['surface']} for i in range(30)]
        ls = [{'id':str(i), 'surface':ls[0]['surface']} for i in range(30)]
        r = e.search(ts, ls, SMALL)
        self.assertEqual(r['model']['rules'], [])
        self.assertEqual(r['data_status'], 'SMALL')
        assignment = {'term': 'aaaaaa', 'label': 'xxxxxx', 'alignment': {'path': [
            ['R', 1, 2, 1, 2, 'MEDIAL', 'a', 'x'], ['R', 2, 3, 2, 3, 'MEDIAL', 'a', 'x']]}}
        c = e.support([('MEDIAL', 'a', 'x')], [assignment]*3)[0]
        self.assertEqual(c['pair_support'], 1)
        self.assertEqual(c['occurrence_support'], 6)

    def test_precision_general_alphabet(self):
        r = [('INITIAL', 'αβ', '◇☆'), ('FINAL', 'Ж', '界')]
        m = evaluator.rule_metrics(r, [r[0], ('MEDIAL','c','d')])
        self.assertEqual(m['precision'], .5)
        self.assertEqual(m['recall'], .5)
        self.assertEqual(evaluator.rule_metrics(r, r)['precision'], 1)
        self.assertEqual(evaluator.rule_metrics([('FINAL','αβ','◇☆')], [r[0]])['precision'], 0)

    def test_conflicting_rules_explicit(self):
        with self.assertRaisesRegex(ValueError, 'forward'):
            e.transform('abc', [('INITIAL','ab','xy'), ('INITIAL','ab','zz')])
        with self.assertRaisesRegex(ValueError, 'inverse'):
            e.validate_rules([('FINAL','ab','xy'), ('FINAL','cd','xy')])
        self.assertEqual(e.transform('abcd', [('INITIAL','ab','xy'),('FINAL','cd','zw')])['outputs'], ['xyzw'])

    def test_sha256_registry_complete(self):
        with tempfile.TemporaryDirectory() as d:
            root = Path(d)
            (root/'source.py').write_text('code')
            (root/'config.json').write_text('{}')
            integrity.write(root)
            self.assertTrue(integrity.verify(root))
            (root/'unregistered').write_text('unexpected')
            self.assertFalse(integrity.verify(root))

    def test_search_resource_bounds(self):
        ts, ls = fixture()
        r = e.search(ts, ls, e.Config(seconds=1e-12))
        self.assertEqual(r['SEARCH_STATUS'], 'INCOMPLETE')
        r = e.search(ts, ls, e.Config(max_states=3, iterations=1))
        self.assertLessEqual(r['states'], 3)
        self.assertLessEqual(r['iterations'], 1)
        self.assertEqual(e.search(rows(['a'*33]), rows(['b']))['SEARCH_STATUS'], 'INCOMPLETE')

    def test_multisymbol_units_active(self):
        u = e.units(['abcdef', 'abcxef', 'abcgef'], 'source')
        self.assertTrue(any(x['length']==4 for x in u))
        self.assertTrue(any(x['role']=='WHOLE' and x['length']==6 for x in u))
        ts, ls = fixture()
        r = e.search(ts, ls, SMALL)
        self.assertTrue(any(len(x[1])>1 and len(x[2])>1 for x in r['model']['rules']))
        self.assertTrue(any(op[0]=='R' and len(op[6])>1 for x in r['model']['assignments'] for op in x['alignment']['path']))

    def test_roles_and_variable_lengths(self):
        r = [('INITIAL','ab','xyz'), ('FINAL','cd','q')]
        a = e.align('abcd', 'xyzq', r)[0]
        self.assertTrue(a['full'])
        self.assertEqual([p[5] for p in a['path']], ['INITIAL','FINAL'])
        self.assertFalse(e.align('cdab', 'qxyz', r)[0]['full'])
        self.assertTrue(e.align('abcd', 'xyz!q', r)[0]['unexplained_units'])

    def test_size_normalized_scoring(self):
        gains = []
        for n in (29, 57, 68):
            ts, ls = fixture(n)
            t, l = e.load_rows(ts)[0], e.load_rows(ls)[0]
            m = e.evaluate(t, l, RULES)
            self.assertAlmostEqual(m['total'], sum(m['components'].values()))
            self.assertAlmostEqual(m['gain'], (m['baseline']-m['total'])/m['baseline'])
            self.assertEqual(e.evaluate(t,l,[])['gain'], 0)
            dup = e.evaluate(e.load_rows(ts+ts)[0], e.load_rows(ls+ls)[0], RULES)
            self.assertEqual(m, dup)
            gains.append(m['gain'])
        self.assertLess(max(gains)-min(gains), .05)
        self.assertEqual(e.search([], [], SMALL)['data_status'], 'EMPTY')

    def test_simpler_model_wins(self):
        t,l = [e.load_rows(x)[0] for x in fixture(29)]
        simple = e.evaluate(t,l,RULES,training=False)
        complex_ = e.evaluate(t,l,RULES+[('WHOLE','unused','unseen')],training=False)
        self.assertGreater(simple['gain'], complex_['gain'])
        trained = e.evaluate(t,l,RULES+[('WHOLE','unused','unseen')])
        self.assertIsNone(trained['objective'])
        singleton = [('WHOLE',t[i],l[i]) for i in range(len(t))]
        self.assertLess(e.evaluate(t,l,singleton,training=False)['gain'],0)

    def test_order_invariance(self):
        ts, ls = fixture()
        before = e.search(ts, ls, SMALL)
        for bag in (ts, ls):
            bag.reverse()
            for i, r in enumerate(bag):
                r['id'] = 'neutral-'+str(19-i)
        self.assertEqual(before, e.search(ts, ls, SMALL))

    def test_checkpoint_resume_identity(self):
        ts, ls = fixture()
        uninterrupted = e.search(ts, ls, SMALL)
        with tempfile.TemporaryDirectory() as d:
            cp = Path(d)/'state.json'
            self.assertEqual(e.search(ts,ls,SMALL,cp,pause_after=1)['SEARCH_STATUS'],'INCOMPLETE')
            self.assertEqual(uninterrupted, e.search(ts,ls,SMALL,cp,resume=True))
            payload = json.loads(cp.read_text())
            payload['payload']['count'] += 1
            cp.write_text(json.dumps(payload))
            with self.assertRaisesRegex(ValueError,'checksum'):
                e.search(ts,ls,SMALL,cp,resume=True)

    def test_heldout_frozen(self):
        ts,ls = fixture()
        tr = e.search(ts,ls,SMALL)
        before = e.canonical(tr)
        ht,hl = rows(['ab'+chr(0x3500+i)+'cd' for i in range(4)]), rows(['xy'+chr(0x4500+i)+'zw' for i in range(4)])
        ho = e.heldout(tr,ts,ls,ht,hl,SMALL)
        self.assertEqual(ho['model']['rules'],tr['model']['rules'])
        self.assertEqual(before,e.canonical(tr))
        with self.assertRaisesRegex(ValueError,'overlap'):
            e.heldout(tr,ts,ls,ts,ls,SMALL)

    def test_full_null_search(self):
        s, _ = generator.generate(73001,'independent_random',9)
        a = e.search(s['train']['terms'],s['train']['labels'],SMALL)
        self.assertEqual(a['SEARCH_TYPE'],'BOUNDED_DETERMINISTIC_OPTIMIZATION')
        self.assertIn('candidate_pool',a)
        self.assertGreaterEqual(a['states'],1)
        # Identical bags with shuffled truth are intentionally indistinguishable.
        self.assertEqual(a,e.search(list(reversed(s['train']['terms'])),s['train']['labels'],SMALL))

    def test_lexicon_loader_parity(self):
        ts, _ = fixture()
        a = e.load_rows(ts)
        self.assertEqual(a, e.load_rows(list(reversed(ts))))
        with self.assertRaises(ValueError):
            e.load_rows([{'id':'x','surface':'a','class':'positive'}])
        with self.assertRaises(ValueError):
            e.load_rows([{'id':'x','surface':'a'},{'id':'x','surface':'b'}])

    def test_neutral_dataset_roles(self):
        source = (ROOT/'engine.py').read_text()
        constants = {n.value for n in ast.walk(ast.parse(source)) if isinstance(n, ast.Constant) and isinstance(n.value, str)}
        self.assertFalse(constants & {'HN1','positive','null_family','STAR','CIRCULAR'})
        imports = [n for n in ast.walk(ast.parse(source)) if isinstance(n,(ast.Import,ast.ImportFrom))]
        self.assertFalse(any('generator' in ast.unparse(n) or 'evaluator' in ast.unparse(n) for n in imports))
        with self.assertRaises(ValueError):
            e.load_rows([{'id':'x','surface':'a','seed':1}])

    def test_isolated_worker_denies_real_and_truth(self):
        # Opens are rejected by the hook before the OS sees either path.
        code = '''import sys
from pathlib import Path
sys.path.insert(0, sys.argv[1])
from worker import install_guard
install_guard(Path('/tmp/surface.json'),Path('/tmp/output.json'),None)
for p in [Path(sys.argv[1])/'latent_truth/secret.json',Path(sys.argv[1]).parent/'restricted_hapax_enrichment_v1/COHORT_MEMBERS.tsv']:
 try: p.read_text()
 except PermissionError: pass
 else: raise AssertionError('isolation failed')
'''
        p = subprocess.run([sys.executable,'-I','-B','-c',code,str(ROOT)],capture_output=True,text=True)
        self.assertEqual(p.returncode,0,p.stderr)

    def test_seed_registry_and_shards(self):
        rows_ = [{'family':'a','seed':1,'dataset':'x'},{'family':'b','seed':2,'dataset':'y'}]
        self.assertEqual(len(orchestrator.seed_registry(rows_)),2)
        with self.assertRaises(ValueError):
            orchestrator.seed_registry(rows_+[rows_[0]])
        a = [{'key':'b','shard':'1','output':2},{'key':'a','shard':'2','output':1}]
        b = [{**r,'shard':'other'} for r in reversed(a)]
        self.assertEqual(orchestrator.aggregate(a),orchestrator.aggregate(b))

    def test_predictive_acceptance_and_train_only_rejection(self):
        rr = [('INITIAL',x,x.upper()) for x in ('ab','cd','ef','gh')]
        rr += [('MEDIAL',x,x.upper()) for x in ('ij','kl')]
        rr += [('FINAL',x,x.upper()) for x in ('mn','op','qr','st')]
        train, test = [], []
        for i,x in enumerate(('ab','cd','ef','gh')):
            for j,y in enumerate(('ij','kl')):
                for k,z in enumerate(('mn','op','qr','st')):
                    (train if (i+j+k)%2 else test).append(x+y+z)
        tr = {'SEARCH_STATUS':'COMPLETE','model':e.evaluate(sorted(train),sorted(x.upper() for x in train),rr,alternatives=True)}
        ho = {'SEARCH_STATUS':'COMPLETE','model':e.evaluate(sorted(test),sorted(x.upper() for x in test),rr,training=False,alternatives=True)}
        self.assertTrue(e.accepted(tr,ho))
        bad = {'SEARCH_STATUS':'COMPLETE','model':e.evaluate(sorted(test),sorted('◇'+str(i)+'◇' for i in range(len(test))),rr,training=False,alternatives=True)}
        self.assertFalse(e.accepted(tr,bad))

    def test_checkpoint_empty_candidate_pool(self):
        t,l = rows(['abc','def','ghi']),rows(['xyz','uvw','rst'])
        baseline = e.search(t,l,SMALL)
        with tempfile.TemporaryDirectory() as d:
            cp = Path(d)/'state.json'
            e.search(t,l,SMALL,cp,pause_after=1)
            self.assertEqual(baseline,e.search(t,l,SMALL,cp,resume=True))

    def test_original_v1_unchanged(self):
        path = ROOT/'V1_BASELINE_HASHES.json'
        self.assertTrue(path.exists())
        self.assertTrue(all(integrity.file_hash(ROOT.parents[2]/p)==h for p,h in json.loads(path.read_text()).items()))

    def test_bigram_null_preserves_edges(self):
        from collections import Counter
        s = 'abacabadabacaba'
        t = generator.bigram_shuffle(s,random.Random(3))
        self.assertEqual(Counter(s[i:i+2] for i in range(len(s)-1)),Counter(t[i:i+2] for i in range(len(t)-1)))


if __name__=='__main__':
    unittest.main()
