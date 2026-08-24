# Task82b design — historical shorthand and selective-extraction fingerprint experiment

Status: **DESIGN, frozen before any operator-grid / null / trajectory
result is interpreted** (see `TASK82B_DESIGN_FROZEN`). Written at git
commit `11d8b9cd63d85975bed1e049cffb03eaa3ee951f`, clean tree. Everything
below was fixed before the extraction operator grid was run over
Doyle/Longfellow/Astafiev and before any shorthand ΔF2 was computed;
carrier/BDD F2 timing pilots (`TASK82B_COST_MODEL.tsv` origin, see §9) and
unit-level code-correctness tests (does an operator's own output length
match its own selection count; does the AX5 estimator recover a synthetic
period-2 signal) are not "primary measurements" in the sense this freeze
gates — no operator parameter, null design, metric subset, or statistical
threshold below was chosen by looking at a Doyle/Longfellow/Astafiev/BDD
*result*.

Task82b is independent of Task81/82/82a (Fontana branch) and of Voynich
entirely; see §4-5 (firewalls).

## 1. Two independent branches

- **Branch S (shorthand/abbreviation).** Real historical
  abbreviated-surface ↔ expanded-edition text pairs. Primary object:
  `ΔF2_shorthand = F2(abbreviated) - F2(expanded)`.
- **Branch A (acrostic/selective extraction).** A small frozen grid of
  deterministic extraction operators applied to natural-language
  carriers. Primary object: `ΔF2_A = F2(A(P)) - F2(P)`.

Outputs are never merged into one dataset (task82b.txt sec.26).

## 2. Authoritative F2 subset (frozen; task82b.txt sec.3)

Task82b reuses, verbatim, the F2 subset Task82a.1 already established as
applicable to a non-Voynich, non-manuscript-hierarchy corpus:
`research/phase2/task82a1/F2_COMMON_DIRECT.tsv` (8 metrics, needs only a
glyph/token stream) union `F2_ASSEMBLER_PROJECTION.tsv`'s
`ASSEMBLER_APPLICABLE` rows for `2DL1/BP1/LS1-4/cs1/cs2/cs6/cs7` (9
metrics, needs only token-sequence + assembler-defined lines — which
every task82b corpus genuinely has as real natural lines, stronger than
Task82a's own synthetic chunk-lines). No Fingerprint V2 definition, null,
bin, or threshold is touched; `internal/task82b.CoreMetricIDs` /
`SupportingMetricIDs` list the 17 metric IDs verbatim
(`internal/task82b/types.go`, `internal/task82b/f2.go`).

CORE (7): `EF1_GIANT_COMPONENT_SHARE`, `EF2_GLOBAL_CLUSTERING`,
`EF3_DEGREE_FREQUENCY_SPEARMAN`, `2DL1_LAYOUT_POSITION_MI`,
`BP1_BOUNDARY_TOKEN_NMI`, `LS2_POSITIONAL_LEXICON_NMI`,
`LS3_BOUNDARY_LENGTH_ASYMMETRY`.

SUPPORTING (10): `EF1_ISOLATE_SHARE`, `LP1_RULE_SUPPORT_GINI`,
`LP4_PREFIX_ATTACHMENT_NMI`, `LP4_SUFFIX_ATTACHMENT_NMI`,
`cs2/prev-family-current-family`, `LS1_LINE_LENGTH_CV`,
`LS4_WITHIN_LINE_EXACT_REPETITION`, `cs1/family-line-position`,
`cs6/family-diversity-x-line-length`,
`cs7/edit-distance-x-structural-distance`.

`2DL1/BP1/LS1-4` are computed via `fingerprintv2.OrderedGroupMetrics`
(Task82a.1's own regression-tested generic recovery of the task79-v1
ordered-group estimators; free, zero repetitions, no null). `EF1-3/LP1/
LP4/cs*` go through the normal `fingerprintv2.Run(Config)` path exactly
as `internal/task82a/f2.go` already calls it. `cs1` is empirically
**never** available for any plain-text corpus (it needs `ivtff_path` line
metadata that a non-Voynich corpus cannot honestly provide) — reported
consistently as `NOT_APPLICABLE` throughout, not a bug specific to one
variant.

**F2Repetitions=5** (`internal/task82b/f2.go`), a Task82b-specific,
documented, cost-driven reduction from Task82a's own 30 (itself already
reduced from Task79's canonical 1000). Justification, measured before any
grid job ran (`TASK82B_COST_MODEL.tsv`): a timing pilot over Doyle/
Longfellow/Astafiev at Repetitions∈{1,5,10,30} found the CORE/SUPPORTING
point estimates this package actually reads (`EF1/EF2/EF3/LP1/LP4` values
and every cross-scale `ObservedStatistic`) byte-identical across that
whole range — repetitions only affect the internal null-significance
precision behind the cross-scale `SUPPORTED`/`NOT_SUPPORTED` verdict
(never a CORE metric in this registry) and diagnostic warning text.
Task82b's grid needs on the order of 450 F2 calls (§9), several times
Task82a's 468-job total run over much smaller synthetic corpora, so 5 is
chosen for cost with no effect on any value this study actually compares.

## 3. Voynich and Fontana firewalls (task82b.txt sec.4-5)

`internal/task82b.assertNoVoynichPath` rejects any path containing
`voynich`, `zl3b`, `it2a`, `cd2a`, `fg2a`, `vt0e`, `rf1b`, `eva.txt`,
`data/ivtff`, or `data_work/` (case-insensitive), independently
reimplemented from `internal/task82a/f2.go`'s policy rather than imported,
so Task82b never has an import edge to Voynich-adjacent code. No Task81/
82/82a output file, "best" Fontana F2 value, or Fontana-to-Voynich
similarity number is read anywhere in this package or its generation
command; no operator parameter or corpus choice below was adjusted to
make Task82b's own output resemble a prior Fontana result.

## 4. Branch S: historical corpora and inclusion criteria

**Primary and only real paired control obtained: Burchards Dekret
Digital (BDD)**, reusing Task79c's already-verified provenance/license
unchanged (task82b.txt sec.9): Burchard of Worms' *Decretum*, Köln
Erzbischöfliche Diözesan- und Dombibliothek Cod. 119, books 6/7/11/12/13,
TEI-XML, repo commit `29f9cb1c34cc9ee3c50e75a6e3e99cfa4a2bc362`, CC BY 4.0
(full chain: `research/phase2/fingerprint/CONTROL_PROVENANCE.tsv`). Unlike
Task79c (which kept only the `<abbr>` branch), Task82b extracts **both**
`<abbr>` and `<expan>` branches of every `<choice>` — present in the
source XML for every abbreviation, never previously extracted — via an
independent new tool, `internal/task82b/teipair.go` (not a modification
of Task79c's frozen `cmd/tei-abbr-extract`).

**Additional historical corpora (sec.10):** searched for openly available
diplomatic+normalized/expanded transcription pairs beyond BDD. Given this
project's data policy (`DATA.md`: third-party bytes are never committed,
only provenance/checksums; any additional corpus would need the same
manual acquisition-and-license-verification Task79c performed for BDD)
and the time budget available for a target-independent control search,
no second independent paired corpus was acquired.
**HISTORICAL_SHORTHAND_ADDITIONAL_CORPORA = NOT_OBTAINABLE** within this
task's scope — recorded as a limitation (task82b.txt sec.10's explicit
fallback), not papered over with a synthetic historical pair. The 5 BDD
chapter files are instead treated as 5 quasi-independent **document-level
blocks within one manuscript/scribe/notation tradition** for the
cross-document stability check (§13); this is **not** cross-tradition
evidence, and `SHORTHAND_CROSS_TRADITION_STABILITY` is graded
accordingly in `TASK82B_REPORT.md`.

**Transcription policy:** TEI `<choice><abbr>/<expan></choice>`, teiHeader/
note/fw/label/toc apparatus excluded, `<lb/>` → line break, every `<g
ref="...">` → one reserved Unicode Glagolitic placeholder rune per
distinct ref (identical policy to Task79c's tool, so both letter-based
metrics and abbreviation-mark signal survive `evaglyph`/`NaturalGlyphs`
filtering).

**Expansion policy:** the edition's own `<expan>` text, lowercased,
whitespace-trimmed; no independent normalization is applied beyond that.

## 5. Branch S: pair alignment (sec.12)

One `PairUnit` per `<choice>`: `{File, Order, Line, AbbrText, ExpanText,
HasMark, MarkIsCombining}`. Alignment is n:1/1:n/n:m by construction (a
`<choice>` may itself span or be split by an embedded `<lb/>`; the
abbreviated and expanded streams are **not** forced to the same line
count — a `<lb/>` inside `<abbr>` is a manuscript-surface line break with
no corresponding break in the normalized `<expan>` text, so line counts
between the two full streams legitimately differ, 7654 vs 6640 lines
observed). `PairUnit.Line` is always the position in the **abbreviated**
stream, since abbreviation is fundamentally a manuscript-surface
phenomenon.

**Corpus-building convention (documented simplification):** F2 is
computed over one-word-per-line corpora built directly from the
`PairUnit` list (`EXPANDED`/`ABBREVIATED`/null words, one per physical
line), not by re-splicing words back into the shared TEI running-text
skeleton. This is the same kind of `ASSEMBLER_DEFINED`-line convention
Task82a already used for its own synthetic corpora; it is chosen because
(a) F2's CORE edit-family/lexical-paradigm metrics are vocabulary-graph
statistics that do not depend on running-prose adjacency, and (b)
splicing `<expan>`/null words back into the shared apparatus text without
altering punctuation spacing is nontrivial extra surgery with no expected
metric benefit. `BoundaryProvenance = PAIR_DEFINED` is recorded
throughout (parallel to Task82a's `ASSEMBLER_DEFINED`).

## 6. Branch S: abbreviation operation inventory (sec.13)

Classes are named from the data (`internal/task82b/shorthandops.go`),
not imposed a priori: `SUSPENSION` (retained letters are a strict prefix
of the expansion), `CONTRACTION` (retained letters are a non-prefix
subsequence — interior letters dropped), `SPECIAL_SIGN_WHOLE_WORD` (the
entire word is one `<g>` mark, no literal letters survive),
`MARK_ONLY_ABBREVIATION` (all literal letters survive; the mark alone
signals an omission, e.g. a doubled consonant), `NO_VISIBLE_CHANGE`,
`OTHER_SUBSTITUTION`. Orthogonally, every `<g>` ref is tagged
`MarkIsCombining` (Unicode combining-diacritical range U+0300-U+036F —
paleographically a superscript abbreviation bar/tilde) versus a dedicated
precomposed special-sign letterform (e.g. Latin ligatures `ꝑ`/`ꝝ`).
Sec.23's functional classes (`SELF_SUFFICIENT` / `CONVENTION_DEPENDENT` /
`AMBIGUOUS_WITHOUT_CONTEXT`) are derived mechanically from these plus
observed expansion ambiguity (`AmbiguousExpan`).

## 7. Branch S: null models (sec.18-20)

All three operate on the expansion text of each pair, deleting exactly
`DeletionCount(pair)` runes (the real abbreviation's own retained-letter
count under a documented greedy longest-common-subsequence-style
alignment heuristic — not a paleographic ground truth, `alignDeletions`
in `internal/task82b/shorthandnull.go`), so output length and per-pair
deletion rate are matched to the real abbreviation exactly:

- **NULL 1 `RANDOM_DELETION_MATCHED`** — uniformly random positions.
- **NULL 2 `FREQUENCY_MATCHED_DELETION`** — positions weighted by the
  empirical per-character deletion rate estimated once from all real
  pairs (`CharDeletionStats.DeleteFreq/TotalFreq`).
- **NULL 3 `POSITION_MATCHED`** — positions weighted by the empirical
  relative-within-word-position histogram (5 equal bins) of real
  deletions.

`EXPANDED` (no deletion) is the "before" baseline; `ABBREVIATED` (the
real surface form) is the positive control (sec.21): if F2 shows no
detectable change even here, F2's sensitivity to shorthand is limited,
which is itself the answer, not a design failure. Each null variant is
generated with **3 independent seeds** per scale (matching Branch A's own
random-null replication), so `SHORTHAND_NULL_COMPARISON.tsv` can compare
the single real `ΔF2` against an actual null distribution
(`internal/task82b.CompareToNull`) rather than one arbitrary draw.

## 8. Branch S: SX (sec.51-52)

F2 sees one corpus at a time; none of the seven SX properties (§52) are
representable in an F2 vector at all, because every one of them needs the
abbreviated↔expanded **alignment**, which is outside F2's input contract
regardless of corpus size or Repetitions. **SX is therefore required by
construction, not by a marginal-sensitivity finding** — `SX_REQUIRED =
YES` is decided here, before any F2 number is read.
`internal/task82b/sx.go` implements all seven: `SX1_CONTRACTION_RATE`,
`SX2_EXPANSION_AMBIGUITY`, `SX3_ABBREVIATION_FAMILY_REUSE` (Gini
coefficient of (abbr,expan) pair-occurrence counts), `SX4_POSITIONAL_
ABBREVIATION_PREFERENCE` (line-initial vs non-initial mean contraction
rate), `SX5_CONTEXT_DEPENDENCE`, `SX6_MANY_TO_ONE_MAPPING`, `SX7_
ABBREVIATION_GRAPH_DENSITY`. Validation (sec.50's gate, adapted: SX has
no "positive/negative control" in the AX sense since it always operates
on real alignment data) is a self-consistency check: SX1 must be 0 for an
identity (no-deletion) pair and near the operator's own deletion rate for
its own null words, verified in `internal/task82b` tests.

## 9. Branch A: carriers, size, and cost (sec.31/33)

`Doyle` (`data_test/pg2097-2.txt`, 43,713 tokens), `Longfellow`
(`data_test/pg30795-mod.txt`, 33,077 tokens), `Astafiev`
(`data_test/astafiev-1000-culinar-receipts-prepared.txt`, run through
`cmd/codex_prepare` with the exact DATA.md-documented flags; output
SHA-256 `ff67a4fbf2606be4409724722e3e4d426aed27bdbeec1698babd92bd2b5eba5a`,
matching DATA.md's expected value; 85,280 tokens) — the same three
controls and paths Task82/Task82a already use. "Matched output size"
(sec.33) is satisfied structurally: every null for a given operator
selects exactly as many atoms of the same kind (token or glyph) as that
operator itself selected from that same carrier, never a fixed constant
across carriers/operators.

`TASK82B_COST_MODEL.tsv` records the pilot timing that fixed
`F2Repetitions=5` (§2) and the total job count below.

## 10. Branch A: operator registry and grid (sec.27-30, frozen; no
combinatorial search per sec.29)

`internal/task82b/operator.go`'s `Registry()`, 20 fixed instances,
`ID -> (StructuralClass, Param)`:

| Structural class | Param grid | Extraction class (sec.40) | Provenance (sec.28) |
|---|---|---|---|
| `FIRST_GLYPH_OF_TOKEN` | - | ACROSTIC | HISTORICALLY_ATTESTED |
| `LAST_GLYPH_OF_TOKEN` | - | TELESTIC | HISTORICALLY_ATTESTED |
| `FIRST_TOKEN_OF_LINE` | - | ACROSTIC | HISTORICALLY_ATTESTED |
| `LAST_TOKEN_OF_LINE` | - | TELESTIC | HISTORICALLY_ATTESTED |
| `FIRST_GLYPH_OF_LINE` | - | ACROSTIC | HISTORICALLY_ATTESTED |
| `LAST_GLYPH_OF_LINE` | - | TELESTIC | HISTORICALLY_ATTESTED |
| `NTH_GLYPH_OF_TOKEN` | n∈{2,3} | POSITIONAL_EXTRACTION | HISTORICALLY_PLAUSIBLE |
| `NTH_TOKEN_OF_LINE` | n∈{2,3} | POSITIONAL_EXTRACTION | HISTORICALLY_PLAUSIBLE |
| `PERIODIC_TOKEN` | k∈{2,3,5,7} | PERIODIC_EXTRACTION | FORMAL_CONTROL |
| `PERIODIC_GLYPH` | k∈{2,3,5,7} | PERIODIC_EXTRACTION | FORMAL_CONTROL |
| `FIXED_OFFSET_WITHIN_GROUP` | offset∈{1,2}, window=3 tokens/line | POSITIONAL_EXTRACTION | FORMAL_CONTROL |

`n=1`/`offset=0` instances are never generated (they would duplicate a
`FIRST_*` class already in the registry) — the registry itself is the
small frozen grid, not a post-hoc filter over a larger one (sec.29).
`k∈{2,3,5,7}` is a small, non-Voynich-derived design choice (a short,
easily-preregistered spread, not tuned to any target — sec.30).
`FIXED_OFFSET_WITHIN_GROUP`'s "group" is defined here concretely as a
non-overlapping 3-token window **within each line** (incomplete trailing
window discarded); this is deliberately distinct from `PERIODIC_TOKEN`
(which thins the whole corpus-wide token stream irrespective of line
boundaries) and from `NTH_TOKEN_OF_LINE` (exactly one pick per line): it
can select multiple tokens per line at a fixed within-window offset.

Every operator carries a `NullClass`: `PER_GROUP` (12 operators — every
glyph-of-token/token-of-line/glyph-of-line/fixed-offset instance) or
`PERIODIC` (8 operators — every `PERIODIC_TOKEN`/`PERIODIC_GLYPH`
instance).

## 11. Branch A: null models (sec.34-37)

- **NULL 1 `RANDOM_SUBSEQUENCE_MATCHED`** (all 20 operators, 3 seeds
  each) — as many atoms of the operator's own kind, chosen uniformly at
  random without replacement from the whole carrier, kept in ascending
  (corpus) order.
- **NULL 2 `POSITION_STRATIFIED_RANDOM`** (the 12 `PER_GROUP` operators,
  3 seeds each) — same groups the operator actually picked from (same
  tokens/lines/windows), one random candidate per group instead of the
  fixed rule.
- **NULL 3 `PERIODIC_PHASE`** (the 8 `PERIODIC` operators, every other
  phase of the same period, deterministic, no seed) — tests whether an
  effect is generic periodic thinning or specific to the frozen phase-0
  alignment.

3 random-null seeds (not more) is a documented, cost-driven,
target-blind choice given the grid's total job count (§13), mirroring
Task82a's own Repetitions reduction.

## 12. AX audit and AX battery (sec.43-50)

**Audit finding (before any AX code was validated):** of the 7 candidate
AX diagnostics sec.46 lists as a minimum to *consider*, three (AX1
concentration at structural beginnings/endings; AX2 first-vs-internal
divergence; AX7 MI between structural position and token/glyph class) are
exactly what `BP1_BOUNDARY_TOKEN_NMI`, `LS2_POSITIONAL_LEXICON_NMI`, and
`2DL1_LAYOUT_POSITION_MI` already measure — all three are already in
Task82b's frozen F2 subset (§2) and computed for every carrier/operator/
null in this study. Reimplementing AX1/AX2/AX7 as independent code would
double-count one signal as two metrics, which sec.61 explicitly warns
against. **Task82b therefore does not implement AX1/AX2/AX7**; the
"is F2 blind to positional/acrostic structure" question (sec.43) is
answered directly from BP1/LS2/2DL1's own observed values (which,
pre-grid, are already nonzero on real prose — e.g. Doyle's untouched
`BP1_BOUNDARY_TOKEN_NMI = 0.0316` — so F2 is not wholly blind to it).

`internal/task82b/ax.go` implements the remaining four as independent,
language-blind, corpus-only statistics (no dictionary/anagram/decoding
per sec.47), applied to an operator's or null's own rendered output
stream: `AX3` (Shannon entropy of the stream), `AX4` (type/token ratio,
a decoding-free lexical-coherence proxy), `AX5` (max NMI between stream
identity and index-mod-k over the same `k∈{2,3,5,7}` grid as the periodic
operators — periods outside this grid are undetectable by construction,
a documented limitation, not a bug), `AX6` (adjacent-line first-atom
match rate, relative to the 1/vocabulary-size chance rate).

**Validation (sec.48-50, gate):** `AX5`'s estimator is validated against
a `SYNTHETIC`-marked positive control (`ax_synthetic_test.go`): a
period-2 alternating stream scores `AX5=1.0` at `k=2`; a length-matched
deterministic-shuffle negative control scores `AX5=0.0185`. No openly
available corpus with a documented historical acrostic was located within
this task's time budget, so no non-synthetic positive control was added
(sec.48's fallback). Given one validated positive control, one negative
control, and no cross-corpus replication of the positive control:
**`AX_VALIDATED = PARTIAL`** — sufficient for `AX5` to be used
descriptively in `TASK82B_REPORT.md`/`TASK83_NOTATION_EXTRACTION_HANDOFF.md`,
insufficient to certify the full gate (`positive-control sensitivity AND
null calibration AND cross-corpus robustness`, sec.50) for `AX3/AX4/AX6`,
which are reported as observational rather than gate-validated.

## 13. Job accounting (frozen before generation)

| kind | count |
|---|---|
| `carrier_baseline` | 3 (Doyle, Longfellow, Astafiev) |
| `operator_output` | 60 (20 operators × 3 carriers) |
| `operator_null_random` | 180 (20 × 3 carriers × 3 seeds) |
| `operator_null_stratified` | 108 (12 `PER_GROUP` operators × 3 carriers × 3 seeds) |
| `operator_null_periodic` | 78 (Σ over the 8 periodic operators × 3 carriers of (period-1) other phases) |
| `shorthand_variant` | 66 (6 scales × [2 deterministic variants (`EXPANDED`,`ABBREVIATED`) + 3 null variants × 3 seeds]) |
| **total** | **495** |

Generated by `cmd/task82b-run` (resumable: a job whose raw JSON already
exists under `raw/` is skipped), aggregated by `cmd/task82b-aggregate`
into every TSV/JSON in §"Required outputs" of `tasks_ph2/task82b.txt`
sec.68.

## 14. Statistical criteria (sec.58-61, frozen)

- **Trajectory:** `ΔF = F(after) - F(before)` per metric, per corpus,
  per operator/variant (sec.58).
- **Null comparison:** `internal/task82b.CompareToNull` — two-sided
  empirical p-value `(#|null-mean| >= |observed-mean| + 1)/(n+1)` and a
  z-like effect size `(observed-null_mean)/null_sd`, computed against the
  matched-null replicate distribution for that operator/carrier/metric
  (sec.59).
- **Multiple testing:** Benjamini-Hochberg FDR at α=0.05 across every
  metric×operator×carrier test in one branch (`internal/task82b.
  BenjaminiHochberg`), reported alongside, never instead of, effect size
  and null separation (sec.60).
- **Redundancy:** Spearman correlation between every pair of the 17 F2
  metrics' ΔF values across the full operator×carrier grid (sec.61);
  |ρ|>0.8 pairs are flagged and not double-counted as independent
  evidence in the stability classification.
- **Convergence:** shorthand ΔF2 compared across the 6 scales (5 chapters
  + combined, sec.17/62); an effect present only in a single small-N
  scale and absent/reversed at the combined scale is labeled
  `SMALL_SAMPLE_EFFECT`.
- **Stability classification (sec.24/39):** `SHORTHAND_TRANSFORMATION_
  FINGERPRINT` / `EXTRACTION_TRANSFORMATION_FINGERPRINT` bins:
  `SYSTEM_SPECIFIC_EFFECT` (Branch S: stable within BDD chapters, no
  cross-tradition data to test further) vs `SHORTHAND_GENERAL_EFFECT`
  (would require ≥2 traditions — not reachable given §4's
  `NOT_OBTAINABLE` finding, so this label is never assigned in Task82b);
  `OPERATOR_SPECIFIC` / `EXTRACTION_GENERAL` / `PLAINTEXT_DRIVEN` /
  `NOT_STABLE` (Branch A, from cross-carrier/cross-seed agreement in
  direction and null-separation).
- **Input dependence (sec.57):** `INPUT_DOMINATED` / `MIXED` /
  `MECHANISM_DOMINATED`, from comparing each carrier's own baseline
  variance against the ΔF2 variance across operators for that carrier.

## 13a. A structural caveat shared by both branches: line-family metrics can be trivially destroyed rather than measured

Two related but distinct effects surfaced during aggregation, both
concerning the `2DL1/BP1/LS1-4/cs6` "line family" (§2's `ASSEMBLER_
PROJECTION` metrics), and both are corrected/documented rather than
silently left in the numbers:

- **Branch S:** the `PAIR_DEFINED` one-word-per-line convention (§5) puts
  exactly one token on every line for *both* `EXPANDED` and `ABBREVIATED`,
  so within-line position never varies and the entire line family is
  identically `0` on both sides — a vacuous, not a detected, `ΔF2=0`.
  `cmd/task82b-aggregate` classifies this case
  `STRUCTURALLY_DEGENERATE_NO_VARIATION`, excluded from every stability
  tally and from `SHORTHAND_TRANSFORMATION_DETECTED`'s count (only the 3
  vocabulary-graph CORE metrics, `EF1-3`, and the `SUPPORTING` LP/EF
  metrics carry real signal for this branch).
- **Branch A:** four operators (`FIRST_TOKEN_OF_LINE`, `LAST_TOKEN_OF_
  LINE`, `FIRST_GLYPH_OF_LINE`, `LAST_GLYPH_OF_LINE`) always emit at most
  one output unit per *original* line, which mechanically zeroes the same
  line family regardless of which position was kept. Unlike Branch S this
  is real, reproducible, cross-carrier evidence (the carrier baseline is
  never itself degenerate), so it is kept as a genuine `ΔF2`observation —
  but it means part of any `ACROSTIC`/`TELESTIC`-vs-`PERIODIC_EXTRACTION`
  gap in `EXTRACTION_STABILITY.tsv` reflects which extraction classes
  *can* collapse to <=1 unit/line by construction, not positional
  specificity alone (`TASK82B_REPORT.md` Q17-18 states this explicitly).

## 14a. A second degeneracy mode: F2 extractor errors on near-alphabet-size vocabulary

Beyond the "too few tokens/types" degeneracy already anticipated (sec.63),
generation surfaced a second, related one: several single-glyph-alphabet
operator/null outputs (mostly `PERIODIC_GLYPH_k` and the glyph-of-token
operators, worst on Astafiev's larger Cyrillic-plus-punctuation-derived
alphabet) make `fingerprintv2`'s frequency-aware C-GRAMMAR null generator
fail outright ("could generate N unique length-1 forms after 300000
attempts, need N+1") — the null generator's target unique-form count can
exceed what a near-exhausted single-character alphabet can physically
supply. This is a property of the frozen F2 extractor meeting a
pathological input, not a Task82b bug, and `fingerprintv2` is never
modified to accommodate it. `cmd/task82b-run.safeExtractF2` catches this
case and records the job with every metric `unavailable`
(`missing_reason="F2_EXTRACTOR_ERROR: ..."`) and `degenerate=true`,
exactly like the token/type-count degeneracy — retained, not dropped, per
sec.63.

## 15. Failure criteria

`TASK82B_EXPERIMENT_INVALID` is issued instead of a frozen portfolio if
any of: the Voynich or Fontana firewall is found breached; the BDD
positive control (§7) shows F2 detecting *no* change at all between
`ABBREVIATED` and `EXPANDED` on the CORE metrics (sec.21's own stated
failure mode, which would still be reported as a valid, if negative,
finding rather than invalidate the run — true invalidity is reserved for
firewall breach or an inability to generate the frozen grid/nulls at
all); or the raw job generation cannot complete for a majority of the
frozen grid.

## 16. Reproducibility

Seeds: `-seed` flag (default 82002) to `cmd/task82b-run`; every
per-job/per-seed derived seed is deterministic from the base seed and job
ID. `go test ./internal/task82b/...` covers the operator engine, the TEI
pair extractor (against the real local BDD files, skipped if absent), SX,
and the AX5 synthetic control. Regenerating from raw: `cmd/task82b-
aggregate` reads only `raw/*.json` and recomputes every TSV/report from
those observed values.
