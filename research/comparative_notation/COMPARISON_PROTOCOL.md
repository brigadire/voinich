# Comparison protocol

1. Verify the raw SHA-256 and immutable provenance record.
2. Freeze adapter/representation versions and validate USC.
3. Run the generic analyzer with the frozen metric registry and seed schedule.
4. Verify rename invariance and corpus-size eligibility.
5. Load the independently pre-frozen VM reference and calibration scale.
6. Join only exact metric ID, version, support regime, and representation
   semantics. Missing joins are `NOT_COMPARABLE`, never zero-distance or a
   penalty.
7. Emit individual metric rows and the vector `(d_G,d_T,d_S,d_L,d_D)`.
8. Report rarefaction and representation sensitivity before interpretation.

Scalar normalization is `|x−x_VM|/s`. `s` must come from a frozen external
calibration panel. Categorical distributions use JS divergence, ordered
histograms use Wasserstein distance, and curves use normalized area between
the common domains. These choices cannot be changed per corpus.

`VM_COMPARISON.tsv` columns are `metric_id, family, vm_value,
candidate_value, candidate_ci, distance, comparable,
reason_if_not_comparable`. `VM_COMPARISON.md` starts with G/T/S/L/D rows, then
Strongest similarities, Strongest differences, Shared corpus rules, VM rules
not reproduced, Candidate rules absent in VM, Corpus-size sensitivity,
Representation sensitivity, Comparability limitations, and Result. Result is
only `STRUCTURALLY_CLOSE_ON`, `STRUCTURALLY_DISTANT_ON`, and
`NOT_COMPARABLE_ON`; mechanism claims are prohibited.

Within-class comparisons require at least three independent corpora and are
reported as a distribution before placing VM relative to it. Cross-class
ranking, winner selection, PCA/UMAP, and nearest-neighbour analysis are outside
this preparation and repository-locked.
