# Task79-b required scope

The Task79 freeze gate is `TASK79_B_REQUIRED`.  The follow-up is deliberately
bounded and must be completed before any model-corpus comparison.

1. Acquire and provenance-lock at least one independently aligned historical
   shorthand/abbreviation corpus and one table/procedural positive control.
2. Add a second independently produced Voynich transcription and align the
   Task79 core metrics without changing their definitions, bins, thresholds or
   null families.
3. Exercise the already specified Fingerprint v2 distance/Pareto interface on
   held-out controls, with corpus size and vocabulary size matched or
   conditioned.
4. Validate or reject `PF4_RECTO_VERSO_COHERENCE` with a leaf-paired null and
   validate predictive hierarchy (`HR3`/`HR5`) out of sample.
5. Report externally motivated boundary/positional additions separately; none
   may enter the Task79 primary family after Voynich/model results are viewed.

No scan annotation, Fontana parameter, F-model, M-model, decryption attempt or
semantic image label is in scope.  If the data cannot be acquired, the result
must remain `NOT_READY`, not be promoted to a freeze by relaxing the gate.
