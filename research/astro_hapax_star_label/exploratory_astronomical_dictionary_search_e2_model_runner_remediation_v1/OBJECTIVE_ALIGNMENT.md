# Objective alignment

`M = sum(assign[identity,label])` is the model edge count. The independent scorer reports `matched = |assignment|`, `complexity`, and `fitness = 100*matched - complexity`. The remediation requires `M == matched` for every callback before interpreting any coverage. The distinct-EVA support gate is encoded in the model and rechecked post hoc; repeated occurrences of one EVA token are not independent. Distinct identity coverage remains a separate diagnostic component.
