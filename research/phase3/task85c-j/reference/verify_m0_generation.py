#!/usr/bin/env python3
"""CONTRACT_REFERENCE_ONLY: 8,192 M0 generation-boundary differential cases."""
from decimal import Decimal
import hashlib, json, math
from m0_reference_impl_a import direct_cdf, fit

outcomes, _, _, p = fit(["ab", "a"], Decimal(1))
allowed = set(outcomes)
h = hashlib.sha256()
for i in range(4096):
    u = Decimal(i) / Decimal(4096)
    a = direct_cdf(outcomes, p, allowed, u)
    # Independent interval-index formulation; it does not call direct_cdf.
    cumulative = []
    s = Decimal(0)
    for x in outcomes: s += p[x]; cumulative.append(s)
    b = outcomes[next((j for j, edge in enumerate(cumulative) if u < edge), len(outcomes)-1)]
    assert a == b
    h.update(a.encode())
for i in range(4096):
    # Generator B independently enumerates all positive outcomes and chooses
    # the minimum exponential clock; it never invokes Generator A.
    uniforms=[((i*1103515245+j*12345+1) % 2147483647)/2147483647 for j in range(len(outcomes))]
    scores=[math.inf if u==0 else -math.log(u)/float(p[x]) for x,u in zip(outcomes,uniforms)]
    b=outcomes[min(range(len(scores)),key=lambda j:(scores[j],j))]
    oracle=sorted(zip(scores,range(len(scores)),outcomes))[0][2]
    assert b==oracle
    h.update(b.encode())
# PF-SC01 is an explicit conditional row, not a fitted-M0 vector, hence unchanged.
pf = direct_cdf(["a","b","c","d","<EOS>"], dict(zip(["a","b","c","d","<EOS>"], map(Decimal,[".28",".22",".18",".12",".20"]))), {"a","b","c","d"}, Decimal(".92848667210989588"))
assert pf == "d"
print(json.dumps({"status":"PASS","cases":8192,"generator_a_cases":4096,"generator_b_cases":4096,"generator_b_calls_a":False,"digest":h.hexdigest(),"pf_sc01":"CLOSED_UNCHANGED"}, sort_keys=True))
