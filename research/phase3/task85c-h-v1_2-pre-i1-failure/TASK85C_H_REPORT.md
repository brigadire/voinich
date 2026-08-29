# Task85c-h report

Task85c-h stops with `SCIENTIFIC_CONTRACT_DEFECT` before production-handler
implementation.

`H-SC01-EVIDENCE-CONTRACT-VERSION`: V1.2 identifies itself as
`G1_V2_EXECUTABLE_CONTRACT_V1_2` but binds the unchanged V1.1 evidence schema
root. All 15 schemas require the literal evidence field
`contract_version=G1_V2_EXECUTABLE_CONTRACT_V1_1`; a generation-success object
changed only to V1.2 is rejected by every schema branch. V1.2's machine
contract also retains V1.1 as first evidence-contract precedence entry.

`H-SC02-E1-JOBID-SCIENTIFIC-VERSION`: E1's machine artifact literally binds
both `contract_version` and `jobid.scientific_identity_version` to V1.1, while
Task85c-h and the updated production-freeze require V1.2 scientific JobID
identity. Selecting either literal changes canonical JobID input, evidence
bytes and hashes.

This is not an ordinary implementation bug: no Go handler can emit one object
that simultaneously has both contract-version literals. Updating schemas,
their root, evidence identifiers and E1's version binding is a scientific
authority refreeze outside Task85c-h. Generation V1.2 identity and PF-SC01
closure remain supported. The firewall is intact and no production material
exists.
