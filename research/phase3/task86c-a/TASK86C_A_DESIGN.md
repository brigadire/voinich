# Task86C-a design

Task86C-a is a read-only post-hoc diagnostic audit of frozen Task86C.  The unit is job × model class; synthetic rows retain generator, scale and replicate.  The generator reads only Task86C manifests, ledger, frozen aggregate manifest and result JSON.  It never opens a corpus, invokes fitting/generation, changes a threshold, or accesses Voynich data.

Statuses are not collapsed: `TRAINING_FAILED`, `UNAVAILABLE_NEGATIVE_SET_EXHAUSTION`, `NOT_REACHED`, and `NOT_RECORDED` remain distinct. Absence of `TRAINING_FAILED` establishes that Stage D received a fitted model; it does not establish adequacy. `PM6ByClass` establishes score validity only, not PM6 gate passage. `PredictivePassByClass` and `StructuralPassByClass` are retained aggregate gates and are not used to invent individual metric outcomes.

All aggregation here is `DERIVED_RECOMPUTATION`. No `NEW_DIAGNOSTIC_COMPUTATION` was run. Deterministic regeneration command:

```sh
python3 research/phase3/task86c-a/generate.py
```

The audit terminal state is evidence-insufficient because the required PM-by-PM and F2 decompositions and PM6 ablation cannot be obtained without prohibited rerunning.
