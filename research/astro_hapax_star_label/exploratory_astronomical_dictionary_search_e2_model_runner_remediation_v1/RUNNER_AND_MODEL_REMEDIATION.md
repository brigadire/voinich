# Remediation result

The parity defect was an alphabet-index canonicalization error: the CP-SAT token index used the incoming target-alphabet order while callback serialization used sorted order. The remediation canonicalizes the target alphabet before both index construction and table serialization.

The remediated solver captures complete assignments in every incumbent, encodes the distinct-EVA support gate in the optimization model, and rechecks objective components with the independent scorer. Small exhaustive cases pass. The legacy 900-second run remains immutable and invalid for interpretation. No real E2 result is produced by this remediation package.
