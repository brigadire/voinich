# Frozen generic metric registry v1

The executable definitions are returned by `notation.MetricRegistry`; the
formulas below are the normative contract. Metrics are label-invariant. Raw
counts remain descriptive; comparisons use frozen scaling and support.

## G — symbol grammar

G01 alphabet size (descriptive); G02/G03 share of the alphabet never initial/
final; G04 observed bigram types / |Σ|²; G05 observed trigram types / |Σ|³;
G06 H(Sᵢ|Sᵢ₋₁)/log₂|Σ|; G07 G06 minus the equivalently normalized
H(Sᵢ|Sᵢ₋₂,Sᵢ₋₁). Bigram/trigram constraint density is one minus occupancy.

## T — token formation

T01/T02 token-length mean/SD plus the full histogram; T03 type/token ratio;
T04 hapax/type ratio; T05/T06 one-symbol prefix/suffix diversity; T07 edit-1
graph density; T08 giant-component share; T09 mean local clustering; T10
Spearman degree/frequency relation; T11 initial/internal/final positional
restriction profile. Histograms use Wasserstein distance; categorical profiles
use Jensen–Shannon divergence.

## S — sequence grammar

For every frozen support regime: S01 observed transition density, S02 zero
density, S03 normalized transition entropy, S04/S05 repeated bigram/trigram
type profile, S06 preferred fraction (observed/independence expectation ≥2),
S07 depleted fraction (≤0.5 with expected count ≥5), and S08 second-order
predictive gain. Production reports both point estimates and bootstrap CIs.

## L — physical-line grammar

Only source-observed physical lines: token-count and symbol-count distributions,
initial/final-token specialization, normalized-position dependence, same-line
co-occurrence and non-cooccurrence density on frozen supports, within-line
progression, and boundary asymmetry. `L06`/`L07` same-line (non-)cooccurrence
density are stratified over the same five frozen sequence-support regimes as
`S01`-`S03`/`S06`-`S08` (`FREQ_GE_5`, `FREQ_GE_10`, `TOP_100`, `TOP_250`,
`MATCHED_VOCAB`); the VM Structural Catalog's "same-line zero density" anchor
is defined on `FREQ_GE_10` (see `VM_REFERENCE_RECONCILIATION.md`), not an
arbitrary top-100 cutoff. If any record lacks the physical-line level, the
family is `NOT_COMPARABLE`; no artificial wrapping is permitted.

## D — document grammar

For each actually observed locus/page/section (and an explicitly justified
voice/staff structural view): within-unit coherence, between-unit
differentiation, exclusive type fraction, hierarchical variance share, and
within-unit progression. Missing levels yield explicit `NOT_COMPARABLE` rows.

## Curves and distances

A2, A3, and AT are evaluated at 5k/10k/20k/39380 and by normalized area
between rarefaction curves. Scalar `d=|x−x_VM|/s` uses `s` frozen from external
calibration controls, never either compared pair. Distribution distances are
JS (categorical) or one-dimensional Wasserstein (ordered). Family distances
`d_G..d_D` are separate; `d_TOTAL` is forbidden.

The executable emits the scalar core, accumulation checkpoints, distribution
serialization (`DISTRIBUTION_OUTPUT_CONTRACT.md`), rarefaction
(`RAREFACTION_PROTOCOL.md`), bootstrap (`BOOTSTRAP_PROTOCOL.md`), and
calibration scales (`CALIBRATION_PANEL_SPEC.md`). All four are now frozen
(B01/B03/B04 closed); this does not authorize any C01-C09 production run,
which remains gated by `PRODUCTION_COMPARATIVE_RUN_AUTHORIZED` and the
per-class B05-B12 blockers in `PREPARATION_BLOCKERS.md`.
