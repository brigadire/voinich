# Residual astronomical LABEL multi-pass mapping protocol v1.0

Protocol status: `FROZEN_BEFORE_AI_PASS`.

This protocol maps a physical astronomical LABEL to one or more exact frozen
ZL3b occurrences. It tests the procedure on 21 legacy-confirmed f68r1 LABELs
before any use on the 71 residual LABELs. It does not perform hapax analysis,
frequency analysis, dictionary matching, semantic interpretation, or expert
verification.

## Frozen design

The 21 known mappings are ordered by LABEL centre from top to bottom and left
to right, divided into seven consecutive spatial strata of three, and one item
per stratum is selected by the smallest SHA-256 of
`residual-ai-mapping-v1|canonical_label_id`. The selected seven form
`EVALUATION_HIDDEN`; the other fourteen form `CALIBRATION_VISIBLE`. Geometry is
BOX and token count is one for all 21, so those dimensions cannot be further
stratified. Token length and horizontal position are reported as balance
diagnostics but do not alter the deterministic rule. No frequency or hapax
field is read.

AI-facing packages replace canonical IDs with neutral IDs. The private
crosswalk and correct answers are stored outside every clean-room package.
The evaluation answers remain sealed until all four first-run outputs and the
shuffled-order Alignment B output are complete.

Candidate sets contain every frozen target-page occurrence: 69 for f68r1, 89
for f68r2, and 110 for f68r3, totaling 268 occurrences in 90 lines. Candidates
are not filtered by frequency, group, visual similarity, or dictionaries.
The base and shuffled orders are deterministic SHA-256 orders with independent
public seeds `candidate-base-v1` and `candidate-shuffle-v1`; neither is corpus
or transcription order.

## Independent passes

`AI_VISUAL_A` receives a crop, page context, visual conventions, and schema,
but no candidate text. It reports observable glyph sequence, boundaries,
direction, start-point status, uncertain positions, alternatives, confidence,
and one of `SINGLE_TOKEN_READING`, `MULTI_TOKEN_READING`,
`RING_CYCLIC_READING`, `AMBIGUOUS_READING`, or `UNREADABLE`.

`AI_ALIGNMENT_B` receives the same image material and the complete same-page
candidate list in base order, but no Visual A/C output. It reports up to three
ranked candidates, glyph-level alignment, unmatched glyphs, boundary and
direction assessment, top-1 confidence, top-2 margin, alternatives, and
`NO_MATCH` when warranted. A second context-free run receives only the
shuffled order and uses the identical rules.

`AI_VISUAL_C` is a second candidate-free visual-first pass with the same schema
as Visual A. It cannot see either earlier pass.

`AI_ADJUDICATOR_D` receives images, candidates, the three first-run outputs,
and the shuffled Alignment output. It cannot see known answers or any excluded
metadata. It may return `CONSENSUS_EXACT`,
`CONSENSUS_WITH_DOCUMENTED_NORMALIZATION`, `PROBABLE`, `AMBIGUOUS`,
`NO_MATCH`, `UNREADABLE`, or `RING_CYCLIC_SEQUENCE`.

The sessions are context-free and are assigned different available model
families where possible. Filesystem isolation is instruction-enforced: each
session is directed to one clean-room package and one output path. This
environment does not provide an OS-level per-agent filesystem jail, so the
blindness claim is procedural plus automated content auditing, not a claim of
kernel-enforced isolation.

## Reading and alignment rules

The system is ZL3b/EVA. Display expansion is fixed: `C=cth`, `K=ckh`,
`P=cph`, `F=cfh`, `N=iin`, `A=ain`, `H=ch`, `S=sh`, `E=ee`, `I=in`.
Candidate identity is the neutral candidate ID bound to an exact occurrence;
readable EVA is display-only. Case, literal markers, upstream token boundaries,
and the first-reading convention are preserved. No new glyph substitution may
be invented.

A multi-token decision must preserve ordered candidate IDs, exact boundaries,
raw and normalized sequences, per-token confidence, and whole-label
confidence. A ring is a cyclic sequence: an arbitrary first token is forbidden;
all supported rotations and direction uncertainty must be retained. A local
non-ring reading uses normal visual reading direction when support exists and
records uncertainty otherwise.

## Confidence and abstention

Confidence is `HIGH`, `MEDIUM`, or `LOW`. `HIGH` requires every visible glyph
and boundary to be accounted for, compatible Visual A/C readings, a supporting
Alignment B candidate, and no material conflict. `MEDIUM` permits a localized
uncertainty. `LOW` means weak support and cannot enter the primary mapping.
Unreadable, materially conflicting, or absent candidates require abstention;
candidate position, uniqueness, or expected yield is never a reason to choose.

Adjudication cannot repair or silently rewrite annotator output. It records
disagreements. `CONSENSUS_EXACT` requires compatible A/C readings, an exact
existing candidate supported by Alignment B, no material glyph/boundary
conflict, and order stability. Documented normalization is allowed only for the
fixed expansion above. `PROBABLE`, abstentions, boundary conflicts, arbitrary
ring starts, and order-sensitive choices remain non-primary.

## Metrics and gates

Calibration and hidden evaluation report exact occurrence accuracy,
token-sequence accuracy, Alignment B top-k recall, visual glyph agreement,
abstention accuracy, false confident matches, order stability, and confidence
calibration. Calibration is diagnostic only and cannot change this protocol.

The production gate is fixed before answer disclosure. It passes only if the
seven hidden cases have at least 6/7 exact occurrences, zero confident links to
nonexistent occurrences, seven valid output records, at least 6/7 order-stable
decisions, no forced replacement of `NO_MATCH`/`AMBIGUOUS`, and a repeated
frozen run reproduces the result. Failure of any condition sets
`RESIDUAL_AI_MAPPING_RUN_AUTHORIZED=NO`; no production residual result files
are created.

If authorized, all 71 residual LABELs use this unchanged pipeline, including
16 f68r1, 33 f68r2, and 22 f68r3. Cross-page domain shift is explicit: hidden
evaluation validates only f68r1 and provides no direct accuracy estimate for
f68r2/f68r3. Therefore f68r2/f68r3 primary acceptance additionally requires
all exact-consensus and stability conditions and is reported separately.

Negative controls use same-page sets with the known correct candidate removed,
wrong-page candidates, order permutations, and visually unrelated pairs. They
are scored only for abstention and never merged into mappings or used for
post-hidden tuning.

After mapping is frozen, 3G1 and human-added metadata may be joined only by a
separate analytical post-processing step. The annotator records remain byte
unchanged. Hapax enrichment is authorized only if the hidden gate passes,
primary mapping is frozen, ungrouped coverage is nonzero and analytically
usable, provenance is modeled, bounds are possible, and all order-sensitive
cases are excluded. Enrichment and lexicon matching remain prohibited here.

