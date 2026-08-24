# PF4 leaf-paired null — specification and result

Specification (fixed before this run, `TASK79C_DESIGN.md` §8):

- **Definition of leaf:** `leafID()` (unchanged, `internal/fingerprintv2/
  task79.go`) — an IVTFF folio id with its trailing `r`/`v` side letter
  stripped (`f1r`/`f1v` → leaf `f1`).
- **Pairing extraction:** a leaf is usable iff both its `r` and `v` sides
  have a folio-mean line-profile vector in the corpus (same vectors PF2–PF5
  already use); a leaf with only one side present is excluded and recorded
  in `unpaired_folios`.
- **Statistic:** unchanged `rectoVersoCoherence` arithmetic — mean of
  `1/(1+distance(recto_vector, verso_vector))` over usable pairs.
- **Null model:** hold every folio's own mean vector fixed; draw a
  uniformly random bijection between the recto-side and verso-side vectors
  over the paired leaves (destroying only which verso is the true physical
  flip side of which recto); recompute the statistic.
- **Permutations:** 1,000. **Seed:** `20260824`.
- **Interpretation:** `SUPPORTED` requires p<0.05 one-sided, a defined
  effect size, and ≥5 usable pairs; otherwise `NOT_SUPPORTED` (enough
  pairs, defined effect, not significant) or `INCONCLUSIVE` (<5 pairs or
  degenerate null).

Implementation: `internal/fingerprintv2/task79c_pf4.go`
(`pf4LeafPairedNull`, unit-tested in `task79c_test.go` against synthetic
fixtures with known ground truth before this real run). Driver:
`cmd/task79c-pf4-hr`. Input: `experiments/fingerprint-v2-task79-v1/
canonical-out/line_profiles.json` (the canonical ZL run's real folio-mean
vectors, unchanged from Task79). Raw output (including all 1,000 null
draws): `experiments/fingerprint-v2-task79c-v1/pf4-hr-out/
pf4_hierarchy_result.json`.

## Result

| Quantity | Value |
|---|---:|
| Observed coherence | `0.5191969682` |
| Leaf-paired null mean | `0.3620975557` |
| Leaf-paired null SD | `0.0184022102` |
| Effect size | `8.537` SD |
| Empirical p-value (one-sided) | `0.000999` (0/1000 permutations ≥ observed) |
| Usable paired leaves | `92` |
| Unpaired folios | `0` |
| **Verdict** | **`SUPPORTED`** |

The observed value (`0.51920`) is numerically identical (to the digits
Task79 reported) to Task79's own `PF4_RECTO_VERSO_COHERENCE` observed value
— confirming this run recomputed the same statistic on the same data,
changing only the null. Task79's own `HN4` null ("folio reassignment within
Currier/section") gave a *non-significant* result for this statistic
(`TASK79_REPORT.md`: effect `-2.97` SD, "not supported", recto/verso
verdict recorded as `INCONCLUSIVE`), because `HN4` relabels which lines
carry which folio identity broadly within Currier/section — a much larger
and different randomization than "swap which real verso is paired with
which real recto." The leaf-paired null constructed here answers the
narrower, physically correct question TASK79_B_SCOPE.md item 4 asked for:
same-leaf recto/verso pages are markedly more profile-similar to each
other than a random recto is to a random verso drawn from the same
document, at 92 real paired leaves and p≈0.001.

## Interpretation constraints

This result establishes physical-leaf-pair profile coherence — it does
**not** establish an acrostic, mirrored writing, a shared cipher key across
a bifolio, or any semantic relationship between a leaf's two sides. It is
registered as `RECTO_VERSO_DEPENDENCE_SUPPORTED = SUPPORTED` for Task79c's
own gate, superseding Task79's `INCONCLUSIVE` (which was explicitly
provisional pending exactly this corrected null, per `TASK79_B_SCOPE.md`).
Per the parent task's own discipline (§24), a `SUPPORTED` verdict here is
not "evidence of a coding scheme" — it is evidence that this local
compositional-profile statistic clusters within physical bifolios, which
is itself consistent with (but does not require) ordinary scribal
practice such as writing both sides of a leaf in one sitting or from one
ruled/prepared bifolio.
