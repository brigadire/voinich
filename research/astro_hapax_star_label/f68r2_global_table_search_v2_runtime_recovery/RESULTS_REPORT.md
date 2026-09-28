# S1 runtime recovery result

The historical exact runtime was recovered at `/home/brigadire/.venv/bin/python3` with OR-Tools 9.15.6755. The frozen f68r2 S1 model was then built over all 25,517 paths without generating a candidate-table list.

CP-SAT returned `OPTIMAL` in 16.4 seconds. The incumbent and best bound are both 3, so the exact maximum for this frozen k≤5 graph is 3/27. Thresholds 4, 5, 6, 11, 17, and 19 are therefore certified infeasible within S1.

The result narrows the S1 range to an exact value. It does not evaluate k≤8 or k≤12, and it does not make a scientific claim.
