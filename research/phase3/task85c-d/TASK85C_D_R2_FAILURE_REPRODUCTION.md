# R2 failure reproduction

The R2 results manifest transitively verifies with artifact root `c8fae8e9df6633e3d0514ff5a8d98661e267ba93b13169baffbbf91422753d79`.

The frozen reproducer independently returns:

- obsolete V1 golden: `j-d85279815a36c30515b0be66387c99c3303fa09e`;
- V1.1 with historical `dependencies`: `j-f7c26e7460fa192e3186873428d5e2a37caa6285`;
- V1.1 with registry `dependency_job_ids`: `j-186f1406add6d4d4d7f788907efb76500468a5f7`.

Thus `R2_FAILURE_CLOSURE_IDENTITY = SUPPORTED`, `R2-G01 = REPRODUCED`, `R2-G02 = REPRODUCED`, and the three-ID divergence is reproduced.

Run `python3 research/phase3/task85c-d/reference/reproduce_r2_jobid_defect.py`.
