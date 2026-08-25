# G1 calibration contract

MFC generation precedes every Voynich fit. Each of MFC0, MFC1, and MFC2 has 16
independent populations with 20,000 DEVELOPMENT, 5,000 VALIDATION, and 5,000
HELDOUT-analogue TOKENs. The fixed alphabet and exact generators are in
`G1_EXECUTABLE_CONTRACT.json`; they use no Voynich-derived alphabet, length,
frequency, component, or transition parameter.

MFC0 calibrates false predictive improvement and overfitting under IID
formation. MFC1 calibrates the same quantities under known short memory. MFC2
calibrates finite-state recovery and structural-distance variation. All three
calibrate seed variation. The known structure labels are respectively
`IID_GLYPH`, `ORDER2_MARKOV`, and `SIX_STATE_PFSA`; none carries a message.

For scalar values `x_1..x_16`, center is the sample median (even sample: mean of
the two central sorted values). Dispersion is the nearest-rank 0.95 quantile of
`abs(x_i-median)`, at one-based index `ceil(.95*n)`. For a metric threshold,
take the maximum applicable generator-specific dispersion. Predictive-gain
nulls use the absolute candidate-minus-baseline difference; overfitting uses
`PM2_DEVELOPMENT-PM2_HELDOUT`; structural nulls use absolute generated-minus-
that-population's HELDOUT-analogue statistic; seed variation uses replicate-
minus-replicate-median. All three MFC generators apply to every G1 metric, and
the threshold is their maximum.
Nonfinite input invalidates calibration. No interpolation, tail switch, human
override, or post-MFC threshold choice is permitted.
