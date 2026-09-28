# Statistical analysis protocol

Frozen before any result review. Primary metric: exact certified maximum coverage out of 27, maximizing over the same 64 profiles for every panel. Secondary metrics: number of profiles attaining the maximum, LABEL-set stability, and identity stability.

Compare botanical and fresh matched astronomy against historical controls and all 99 pseudo-controls. The model-selection-aware empirical null is `p=(1 + count(M_pseudo >= M_real))/(1+N)`, with N=99. Ties count as at least as extreme. The maximum is the same exact integer metric for all panels; UNKNOWN is not exact coverage.

Multiplicity correction is applied to the two preregistered real-domain hypotheses (botanical and astronomy) using Holm step-down over their two raw empirical p-values. No unregistered threshold such as 4/27 is imported. Concrete identity pairs are not interpreted when co-optimal assignments are materially ambiguous; report only aggregate coverage and stability.

No statistical result is computed in this audit package.
