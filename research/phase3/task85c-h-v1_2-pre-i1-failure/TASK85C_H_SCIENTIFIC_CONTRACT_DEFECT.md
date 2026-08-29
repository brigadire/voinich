# Scientific contract defect

Blocking findings: `H-SC01-EVIDENCE-CONTRACT-VERSION` and
`H-SC02-E1-JOBID-SCIENTIFIC-VERSION`.

The smallest evidence fixture is the frozen positive generation-success object.
With its original V1.1 contract-version literal it matches the schema golden;
changing only that field to the current V1.2 identity makes it fail every
branch. Independently, E1's machine JobID binding remains V1.1 while current
production authority requires V1.2. Both choices alter canonical bytes and
scientific identity hashes.

Required upstream repair: issue V1.2-compatible evidence schemas/schema root
and resolve E1's current-contract scientific identity binding, with new
cross-version regression vectors. Task85c-h cannot select these values in code.
