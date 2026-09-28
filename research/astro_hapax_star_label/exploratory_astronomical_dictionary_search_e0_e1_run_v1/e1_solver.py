"""Reduced exact E1 CP-SAT model.

The reduction removes only form/token pairs that cannot be produced by any
injective global DROP_UNMAPPED table of size <= k. The compatibility predicate
is exhaustive over subsets of source graphemes and is separately safety-tested.
"""
from collections import defaultdict
from itertools import combinations
import time
from ortools.sat.python import cp_model

def pattern(s):
    d={}; out=[]
    for c in s:
        if c not in d: d[c]=len(d)
        out.append(d[c])
    return tuple(out)

def possible(form, token, k):
    form=''.join(form.split()); wanted=pattern(token)
    if len(token)>len(form) or len(set(wanted))>k: return False
    for n in range(1,k+1):
        for selected in combinations(set(form), n):
            projected=''.join(c for c in form if c in selected)
            if len(projected)==len(token) and pattern(projected)==wanted:
                return True
    return False

def compatible_pairs(forms, labels, k):
    return {(f, l['zl3b_token']) for f in forms for l in labels if possible(f,l['zl3b_token'],k)}

def solve(lexicon, labels, source_alphabet, target_alphabet, table_size, capacity_policy, limit_sec=60):
    forms={f.replace(' ','') for fs in lexicon.values() for f in fs}
    eligible=compatible_pairs(forms, labels, table_size)
    token_index={t:i for i,t in enumerate(target_alphabet)}; m=len(target_alphabet)
    model=cp_model.CpModel()
    source=tuple(sorted(source_alphabet)); target=tuple(sorted(target_alphabet))
    x={c:model.NewIntVar(0,m,f'x_{c}') for c in source}
    assigned={c:model.NewBoolVar(f'assigned_{c}') for c in source}
    is_t={}
    for c in source:
        model.Add(x[c]!=m).OnlyEnforceIf(assigned[c]); model.Add(x[c]==m).OnlyEnforceIf(assigned[c].Not())
    model.Add(cp_model.LinearExpr.Sum(list(assigned.values()))==table_size)
    for c in source:
        for t in range(m):
            b=model.NewBoolVar(f'is_{c}_{t}'); model.Add(x[c]==t).OnlyEnforceIf(b); model.Add(x[c]!=t).OnlyEnforceIf(b.Not()); is_t[c,t]=b
    for t in range(m): model.Add(cp_model.LinearExpr.Sum([is_t[c,t] for c in source])<=1)
    def automaton(token):
        idx=[token_index[c] for c in token]; L=len(idx); fail=L+1; rows=[]
        for s in range(L+1):
            for a in range(m+1): rows.append((s,a,s if a==m else (s+1 if s<L and a==idx[s] else fail)))
        rows.extend((fail,a,fail) for a in range(m+1)); return rows,L
    matches={}
    for f in forms:
        for token in {l['zl3b_token'] for l in labels}:
            if (f,token) not in eligible:
                matches[f,token]=model.NewConstant(0); continue
            if any(c not in token_index for c in token): matches[f,token]=model.NewConstant(0); continue
            rows,L=automaton(token); states=[model.NewIntVar(0,L+1,f's_{abs(hash((f,token)))%10**9}_{i}') for i in range(len(f)+1)]
            model.Add(states[0]==0)
            for i,c in enumerate(f):
                if c not in x: model.Add(states[i+1]==L+1)
                else: model.AddAllowedAssignments([states[i],x[c],states[i+1]],rows)
            z=model.NewBoolVar(f'm_{abs(hash((f,token)))%10**9}')
            model.Add(states[-1]==L).OnlyEnforceIf(z); model.Add(states[-1]!=L).OnlyEnforceIf(z.Not()); matches[f,token]=z
    ids=list(lexicon); identities={}
    for ident,fs in lexicon.items():
        for token in {l['zl3b_token'] for l in labels}:
            z=model.NewBoolVar(f'im_{ident}_{abs(hash(token))%10**8}')
            rel=[matches[f.replace(' ','') ,token] for f in fs]
            model.AddMaxEquality(z,rel if rel else [model.NewConstant(0)]); identities[ident,token]=z
    assign={}
    for ident in ids:
        for l in labels:
            a=model.NewBoolVar(f'a_{ident}_{l["label_id"]}')
            model.AddImplication(a,identities[ident,l['zl3b_token']]); assign[ident,l['label_id']]=a
    for l in labels: model.Add(cp_model.LinearExpr.Sum([assign[ident,l['label_id']] for ident in ids])<=1)
    if capacity_policy=='GLOBAL_CAPACITY_1':
        for ident in ids: model.Add(cp_model.LinearExpr.Sum([assign[ident,l['label_id']] for l in labels])<=1)
    else:
        for ident in ids:
            for page in sorted({l['page'] for l in labels}): model.Add(cp_model.LinearExpr.Sum([assign[ident,l['label_id']] for l in labels if l['page']==page])<=1)
    objective=cp_model.LinearExpr.Sum(list(assign.values())); model.Maximize(objective)
    solver=cp_model.CpSolver(); solver.parameters.max_time_in_seconds=limit_sec; solver.parameters.num_workers=1
    start=time.time(); status=solver.Solve(model); elapsed=time.time()-start
    feasible=status in (cp_model.OPTIMAL,cp_model.FEASIBLE)
    table={}
    if feasible:
        for c in source:
            v=solver.Value(x[c])
            if v<m: table[c]=target[v]
    matched=int(solver.Value(objective)) if feasible else 0
    return {'status':solver.StatusName(status),'is_optimal':status==cp_model.OPTIMAL,'table':table,'matched':matched,'runtime_sec':round(elapsed,4),'best_bound':solver.BestObjectiveBound() if feasible else None,'optimality_gap':(solver.BestObjectiveBound()-matched) if feasible else None,'eligible_pairs':len(eligible),'model_forms':len(forms),'model_tokens':len({l['zl3b_token'] for l in labels}),'timed_out':status in (cp_model.FEASIBLE,cp_model.UNKNOWN)}
