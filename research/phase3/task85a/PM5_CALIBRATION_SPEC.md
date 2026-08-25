# PM5 predictive-calibration functional

PM5 is multiclass top-label expected calibration error. At every scored event,
obtain the model's normalized predictive distribution before observing the
unit. Let confidence `p_i` be its maximum probability; choose the predicted
label by code-point order on a tie; let correctness `y_i` be one when that
label equals the observed unit and zero otherwise.

Use ten fixed bins: bin 0 is `[0,.1]`; bins 1 through 9 are `(j/10,(j+1)/10]`.
For each nonempty bin compute mean confidence and mean correctness. Empty bins
contribute zero. With `N` events,
`PM5 = sum_b (n_b/N)*abs(mean_confidence_b-mean_correctness_b)`.
Accumulate events in corpus order and bin sums in bin-index order using
float64. PM5 is reported per candidate, partition, transcription, and seed;
seed aggregation is the median. A missing or nonnormalized distribution makes
PM5 nonfinite and triggers numerical instability.
