# Frozen VM reference contract

The primary frozen source remains `research/structure_catalog/`; it must not be
recomputed for individual candidates. Existing direct mappings include VM
alphabet size 36, initial restriction 9/36, final restriction 0/36, bigram
occupancy 379/1296, trigram occupancy 1569/46656, frequent transition zero
density 0.9604589793 at threshold 10, and same-line zero density 0.7693935582.

Generic metrics not already represented in that frozen catalog have now been
formally implemented, tested, and computed once for VM in a separately
versioned freeze (`VM_REFERENCE_V2.tsv`, `VM_REFERENCE_V2_MANIFEST.json`,
`VM_REFERENCE_RECONCILIATION.md`) before any production authorization. They
were not introduced after candidate inspection — no C01-C09 candidate has
been run. This preparation deliberately does not create candidate-specific
VM values or distances, and the comparator (`notation.Compare`) accepts only
a `Fingerprint` matching the frozen manifest's `output_sha256`
(`notation.VerifyFrozenVMReference`); recomputing VM_REFERENCE_V2 for an
individual candidate is not authorized.
