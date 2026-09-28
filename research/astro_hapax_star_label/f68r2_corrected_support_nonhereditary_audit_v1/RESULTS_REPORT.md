# Results report

The prior coverage>=3 subset bound is invalid and has been withdrawn. The synthetic A/B/C regression fixture demonstrates why: each rule has support two in the triple, while every pair contains a rule with support one.

The frozen graph retains certified coverage-1 and coverage-2 infeasibility. The direct coverage-3 signature audit checked 60,592 admissible signature triples and 16,925,107 path triples with distinct LABEL, identity, token, required/forbidden mapping, injectivity, support, and complete scorer replay checks. It found zero valid triples, so `COVERAGE_3_STATUS=INFEASIBLE_CERTIFIED`. This does not imply infeasibility for coverage 4 or higher; those thresholds remain unevaluated.
