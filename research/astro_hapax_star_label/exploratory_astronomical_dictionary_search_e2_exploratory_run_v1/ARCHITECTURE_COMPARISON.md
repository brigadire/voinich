# Architecture comparison

| Architecture | Build behavior | Exactness | Support handling | Result |
|---|---|---|---|---|
| Monolithic CP-SAT | real-scale construction did not reach search | intended exact | embedded, but construction explodes | rejected for current E2 |
| D1 master + matching oracle | realistic synthetic build completes | heuristic table search; exact small fixed-table oracle | final-assignment audit | implemented |
| D2 aggregated CP-SAT | not implemented | not evaluated | not evaluated | deferred |

D1 is suitable for exploratory engineering and produces reviewable incumbents. It is not a replacement for a certified exact E2 optimum.
