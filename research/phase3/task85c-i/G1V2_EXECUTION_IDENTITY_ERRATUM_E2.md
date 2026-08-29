# G1V2 execution identity erratum E2

`G1V2_EXECUTION_IDENTITY_ERRATUM_E2` is the execution-identity authority for `G1_V2_EXECUTABLE_CONTRACT_V1_2`. It supersedes E1 only for current V1.2 version binding; E1 remains the historical V1.1 authority.

The unambiguous fields are `execution_identity_spec_version = G1V2_EXECUTION_IDENTITY_SPEC_E2` and `scientific_contract_version = G1_V2_EXECUTABLE_CONTRACT_V1_2`. JobID `scientific_identity_version` is `G1_V2_EXECUTABLE_CONTRACT_V1_2`. The JobID hash construction and canonical `dependency_job_ids` field are unchanged.

E1's boundary is preserved: scientific randomness is `G1V2-RNG-1`; blind ID is an opaque execution identifier; escrow provides blindness and truth commitment; JobID identifies a run execution bound to a scientific job. Scientific identity depends on neither escrow-key bytes nor blind-ID bytes. Changing a valid escrow key may change opaque run identifiers but never scientific control content, RNG realization, or truth.

The scientific JobID payload excludes hostname, worker, coordinator, retry, lease, scheduling order, and wall-clock time. `EI01`, `R2-G01`, and `R2-G02` are closed.
