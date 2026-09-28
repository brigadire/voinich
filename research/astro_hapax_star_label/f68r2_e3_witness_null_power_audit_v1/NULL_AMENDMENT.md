# Amendment: witness-null power audit

This package does not alter `f68r2_e3_witness_null_v1`. It separates three ledgers:

- `REAL_CALIBRATION_WITNESS`: one historical S043 assignment supplied as a calibration fixture;
- `NULL_CANDIDATE_REJECTED_BY_REPLAY`: the one apparent null candidate from the old package, rejected after global replay;
- `VALID_NULL_WITNESS`: none.

The historical S043 assignment is **not** a valid witness under the corrected global mapping semantics: adding mappings from other selected paths causes extra output characters. Therefore its independent real-data rediscovery rate is zero in this audit.

All intervals below describe the algorithmic detection rate at a stated search budget. They do not estimate the true probability that a null instance has a coverage-5 solution.
