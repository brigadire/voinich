import collections,itertools,json,random,tempfile,unittest
from pathlib import Path
from unittest.mock import patch
import engine
import run


def label(i,token,page='f68r1'):
    return dict(occurrence_id=str(i),token=token,page_id=page,hapax='0')

class MatchingTests(unittest.TestCase):
    def test_optimal_not_greedy(self):
        labels=[label(0,'a'),label(1,'b'),label(2,'z')]
        idx={'a':('X','Y'),'b':('X',)}
        actual=engine.matching(labels,idx)
        self.assertEqual(actual,{'1':'X','0':'Y'})
        self.assertNotIn('2',actual)
    def test_exhaustive_small_graph_oracle(self):
        rng=random.Random(77)
        for _ in range(80):
            labels=[label(i,str(i)) for i in range(4)]
            index={str(i):tuple(t for t in ('A','B','C') if rng.random()<.5) for i in range(4)}
            best=0
            for assignment in itertools.product((None,'A','B','C'),repeat=4):
                chosen=[t for t in assignment if t is not None]
                if len(chosen)!=len(set(chosen)):continue
                if all(t is None or t in index[str(i)] for i,t in enumerate(assignment)):best=max(best,len(chosen))
            self.assertEqual(len(engine.matching(labels,index)),best)
    def test_identity_capacity_across_variants(self):
        terms=[dict(identity='X',forms=['alpha','beta'])]
        idx=engine.indexes(terms)[0]
        self.assertEqual(len(engine.matching([label(0,'alpha'),label(1,'beta')],idx)),1)
    def test_forbidden_assignment_margin(self):
        ls=[label(0,'a'),label(1,'b')]
        idx={'a':('A','B'),'b':('A','B')}
        self.assertEqual(len(engine.matching(ls,idx,forbidden=('0','A'))),2)
        self.assertEqual(len(engine.matching(ls,{'a':('A',),'b':('B',)},forbidden=('0','A'))),1)
    def test_unavailable_train_identity(self):
        idx={'a':('A',),'b':('A','B')}
        self.assertEqual(engine.matching([label(0,'a'),label(1,'b')],idx,unavailable=['A']),{'1':'B'})

class SelectionTests(unittest.TestCase):
    def test_grid_and_score(self):
        self.assertEqual(len(engine.GRID),64)
        self.assertEqual(sum(r['level']=='A' for r in engine.GRID),4)
        r=engine.evaluate([label(0,'alpha'),label(1,'x')],{'alpha':('A',)},engine.GRID[0])
        self.assertEqual((r['score'],r['coverage']), (100,.5))
        self.assertLess(max(r['complexity'] for r in engine.GRID),100)
    def test_train_tie_never_uses_heldout(self):
        ts=[dict(identity='A',forms=['abc']),dict(identity='B',forms=['viva'])]
        idx=engine.indexes(ts)
        # Identity and IJ/UV tie on train; IJ/UV would fit heldout 'uiua'.
        first=engine.search([label(0,'abc'),label(1,'uiua','f68r2')],idx)['f68r1_TO_f68r2']
        second=engine.search([label(0,'abc'),label(1,'viva','f68r2')],idx)['f68r1_TO_f68r2']
        self.assertEqual(first['system_id'],'S000')
        self.assertEqual(second['system_id'],'S000')
        self.assertEqual(first['matched'],0)
        self.assertEqual(second['matched'],1)
    def test_hapax_never_affects_selection(self):
        ls=[label(0,'abc'),label(1,'abcd','f68r2')];idx=engine.indexes([dict(identity='A',forms=['abc'])])
        expected=engine.search(ls,idx)
        for l in ls:l['hapax']='1'
        self.assertEqual(engine.search(ls,idx),expected)
    def test_normalization_contract(self):
        self.assertEqual(engine.transform('‘al-Ṭā’ir',engine.GRID[0]),'altair')
        self.assertEqual(engine.transform('viva',engine.GRID[8]),'uiua')
        self.assertEqual(engine.transform('al-Ṭā’ir',engine.GRID[32]),'tair')
    def test_order_invariance(self):
        ls=[label(0,'abc'),label(1,'def','f68r2')];idx=engine.indexes([dict(identity='A',forms=['abc']),dict(identity='B',forms=['def'])])
        self.assertEqual(engine.search(ls,idx),engine.search(ls[::-1],idx))

class NullTests(unittest.TestCase):
    def test_preserving_dictionary_nulls(self):
        ts=[dict(identity='A',forms=['al abracadabra','bookkeeper','abc']),dict(identity='B',forms=['anana'])]
        for seed in range(30):
            for family in ('LENGTH_ENDPOINT','UNIGRAM','BIGRAM'):
                ns=engine.synthetic_dictionary(ts,random.Random(seed),family)
                self.assertEqual([t['identity'] for t in ts],[t['identity'] for t in ns])
                for a,b in zip((f for t in ts for f in t['forms']),(f for t in ns for f in t['forms'])):
                    for x,y in zip(a.split(),b.split()):
                        self.assertEqual((len(x),x[0],x[-1]),(len(y),y[0],y[-1]))
                        if family!='LENGTH_ENDPOINT':self.assertEqual(collections.Counter(x),collections.Counter(y))
                        if family=='BIGRAM':self.assertEqual(collections.Counter(zip(x,x[1:])),collections.Counter(zip(y,y[1:])))
    def test_historical_capacity_and_size(self):
        ts=[dict(identity='A',forms=['abcd','abc']),dict(identity='B',forms=['abcde'])]
        ns,dist=engine.sampled_dictionary(ts,['fred','eva','peter','joe'],random.Random(3))
        self.assertEqual(dist,[0,0,0]);self.assertEqual([len(t['forms']) for t in ns],[2,1])
        self.assertEqual(len(set(f for t in ns for f in t['forms'])),3)
    def test_full_search_parity(self):
        ts=[dict(identity='A',forms=['alpha']),dict(identity='B',forms=['betaa'])]
        ls=[label(0,'alpha'),label(1,'betaa','f68r2')];idx=engine.indexes(ts)
        calls=[]
        original=engine.evaluate
        def tracked(labels,index,rule,*args,**kwargs):
            calls.append(rule['system_id']);return original(labels,index,rule,*args,**kwargs)
        with patch('engine.evaluate',tracked):
            engine.search(ls,idx)
        self.assertEqual(len(calls),3*64+2)
        for family in ('LENGTH_ENDPOINT','UNIGRAM','BIGRAM'):
            calls.clear()
            with patch('engine.evaluate',tracked):
                rows=run.draw(family,99,ls,ts,{}, {},idx,{})
            self.assertEqual(len(calls),3*64+2)
            self.assertEqual(len(rows),3)
            self.assertEqual({r['completion_status'] for r in rows},{'COMPLETE'})
    def test_checkpoint_deterministic_and_seed_replay(self):
        ts=[dict(identity='A',forms=['alpha'])];ls=[label(0,'alpha'),label(1,'alpha','f68r2')];idx=engine.indexes(ts)
        a=run.draw('UNIGRAM',99,ls,ts,{},{},idx,{})
        b=run.draw('UNIGRAM',99,ls,ts,{},{},idx,{})
        self.assertEqual(a,b)
        with tempfile.TemporaryDirectory() as d:
            a_path=Path(d)/'a.gz';b_path=Path(d)/'b.gz'
            self.assertEqual(run.checkpoint_write(a_path,a),run.checkpoint_write(b_path,b))
    def test_same_page_label_sampling(self):
        ls=[label(0,'abc'),label(1,'abcd','f68r2')]
        pool=[label('p1','xyz'),label('p2','lmno','f68r2')]
        sample,delta,n=run.sampled_labels(ls,pool,random.Random(1),True)
        self.assertEqual([x['token'] for x in sample],['xyz','lmno'])
        self.assertEqual(delta,0);self.assertEqual(n,2)
    def test_sensitivity_failure_gate(self):
        rows=[dict(noise=n,split=e,sensitivity_pass=int(i<17)) for n in (0,.1,.25) for e in run.ENDPOINTS for i in range(20)]
        self.assertFalse(run.synthetic_pass(rows))
        for r in rows:r['sensitivity_pass']=1
        self.assertTrue(run.synthetic_pass(rows))

if __name__=='__main__':unittest.main()
