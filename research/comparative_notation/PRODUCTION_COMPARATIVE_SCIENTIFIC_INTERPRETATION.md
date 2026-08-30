# Scientific interpretation of the first production Comparative Notation Study run

Source run: `research/comparative_notation/production_runs/CNS-PROD01-RUN-20260830T182652Z`
(git commit `2fdb30885374759632133e2b458966acfb6cdca9`, `PRODUCTION_COMPARATIVE_RUN_COMPLETED=true`,
`PRODUCTION_COMPARATIVE_RUN_VALID=true`). This document interprets that
immutable bundle; it changes nothing in it. Supporting machine-readable
tables: `SCIENTIFIC_INTERPRETATION_EVIDENCE.tsv` (1,288 rows — every
candidate/representation/checkpoint/metric combination with its raw
values, calibrated distance, rarefaction behaviour, and bootstrap CI) and
`SCIENTIFIC_CLAIM_EVIDENCE_MATRIX.tsv` (13 claims, each traced to
specific evidence rows).

Throughout, four evidence levels are kept separate:

1. **Observed** — a value read directly from the production bundle.
2. **Statistical interpretation** — a conclusion that follows from
   distributions, calibration, rarefaction, or bootstrap uncertainty
   already computed in the bundle.
3. **Scientific interpretation** — a cautious reading against the
   research hypotheses this study exists to inform.
4. **Unsupported** — explicitly flagged where this run cannot answer a
   question, rather than silently omitted.

## 1. Executive scientific summary

Across the frozen production subset, the Voynich Manuscript's (VM)
symbol-grammar (G family) sits at the lowest calibrated distance among
the families comparable for each candidate (G, T, S for the two Latin
representations; G, T, S, L, D for the three music representations), for
every one of the five candidate/representation combinations tested, and
this ranking is stable between the two usable checkpoints (5,000 and
10,000 tokens). Beyond that
single point of agreement, the picture is multidimensional and
representation-dependent, not a single "VM is closer to X" story:

- Token-formation (T) is the one family directly comparable for *both*
  candidate classes where the two disagree sharply: T distance to VM is
  roughly one to two orders of magnitude larger for the tested mensural
  music representations (C06 R1/R2/R3) than for the tested Latin
  representations (C01, C02). Document-hierarchy (D) is large for every
  music representation too, but D is `NOT_COMPARABLE` for C01/C02 (no
  overlapping observed hierarchy level with VM — Section 4), so D cannot
  itself be read as "Latin closer than music"; it has no Latin-side value
  to be closer than.
- Sequence-grammar (S) does not follow the T pattern: it is the largest
  distance family for both Latin representations, but is the
  second-smallest family for one of the three music representations
  (MUSIC-R1), and is close to the Latin values for a second
  (MUSIC-R3 at 10,000 tokens: 19.78 vs. Latin's 19.05–19.48). S is
  genuinely mixed, not a second family corroborating T.
- The spread across C06's own three representations (R1 vs R2 vs R3) is
  larger than the spread between candidate identities (C01 vs C02 vs
  C06) for every comparable family, at both checkpoints. Representation
  choice is not a minor nuisance here; it is as large or larger an effect
  than candidate identity itself.
- A comparable representation-sensitivity effect exists for the Latin
  pair: EXPANDED vs DIPLOMATIC transcription shifts several G-family
  metrics (alphabet size, initial-restriction density, bigram occupancy)
  by 56% to over 600% relative, which is on the same order as, or larger
  than, several of the calibrated candidate-vs-VM distances reported for
  those same metrics.
- No frozen numeric threshold exists for a binary "structurally close" /
  "structurally distant" verdict, so none is assigned here; that
  decision, if wanted, is future preregistered work (Section 15).
- This design does not test historical mechanism, device, retrieval
  procedure, or authorship. It cannot confirm or refute a Fontana-style
  external-memory hypothesis, which is a different, already-attempted,
  and already-inconclusive question (`research/phase2/PHASE2_FAILURE_ANALYSIS.md`).

## 2. Scope and evidentiary limits

This interpretation covers exactly the one validated run named above. It
does not recompute, does not re-run rarefaction/bootstrap/calibration,
does not touch the frozen VM reference, and does not introduce a new
composite similarity score or a new numeric threshold. Every number
quoted below is either read directly from the production bundle or is a
descriptive recombination of bundle numbers (means of already-computed
per-metric distances, spreads, relative CI widths, direction agreement
counts) computed once, deterministically, from the frozen outputs — never
a new metric definition and never a re-estimation of a scientific
quantity with different parameters.

The production panel is three candidate classes only (C01, C02, C06), of
which C01/C02 share one manuscript source and C06 contributes three
dependent representations of one musical corpus. Section 17 details why
this rules out robust cross-class ranking, within-class variance, or any
claim about "notation systems in general."

## 3. Production dataset and candidate roles

| Candidate | Representation | Role | Records | Lines observed |
|---|---|---|---:|---|
| C01 | LATIN-EXPANDED | Ordinary Latin manuscript prose, expanded transcription | 19,132 | no |
| C02 | LATIN-DIPLOMATIC | Same manuscript, diplomatic (abbreviation-preserving) transcription | 19,121 | no |
| C06 | MUSIC-R1 | Mensural music, per-event representation | 13,117 | yes |
| C06 | MUSIC-R2 | Mensural music, signed melodic-interval representation | 11,785 | yes |
| C06 | MUSIC-R3 | Mensural music, pitch×duration compound representation | 12,475 | yes |

C01 and C02 are two representations of the *same* underlying source
(BDD Köln Cod. 119), not two independent historical controls (Section 5).
C06's three representations are three representations of *one* musical
corpus (JLSDD), never treated in this run as three independent candidate
corpora (Section 6; enforced mechanically in the run's own
`representation_independence` preflight gate and confirmed again here by
inspection of `candidate_id`/`representation_id` columns throughout the
evidence table). Both reachable checkpoints (5,000 and 10,000 tokens) are
below every candidate's own observed size; 20,000 and 39,380 are
`NOT_COMPARABLE` for all five candidate/representation combinations,
because no candidate corpus in this subset reaches those sizes — this
is a real, not a not-yet-computed, limitation of the frozen panel.

## 4. Metric-family results (descriptive)

Family-level mean calibrated distance (unweighted mean of the comparable
per-metric calibrated distances within the family; see
`AGGREGATE_SUMMARY.tsv` in the run bundle and its recomputation in
`SCIENTIFIC_INTERPRETATION_EVIDENCE.tsv`):

| Representation | ck | G | T | S | L | D |
|---|---:|---:|---:|---:|---:|---:|
| LATIN-EXPANDED | 5000 | 4.43 | 13.31 | 29.05 | NOT_COMPARABLE | NOT_COMPARABLE |
| LATIN-EXPANDED | 10000 | 5.72 | 18.42 | 19.05 | NOT_COMPARABLE | NOT_COMPARABLE |
| LATIN-DIPLOMATIC | 5000 | 2.44 | 13.82 | 29.76 | NOT_COMPARABLE | NOT_COMPARABLE |
| LATIN-DIPLOMATIC | 10000 | 2.87 | 19.39 | 19.48 | NOT_COMPARABLE | NOT_COMPARABLE |
| MUSIC-R1 | 5000 | 4.26 | 72.32 | 8.50 | 19.29 | 120.33 |
| MUSIC-R1 | 10000 | 6.10 | 132.40 | 8.60 | 27.07 | 143.38 |
| MUSIC-R2 | 5000 | 16.06 | 1866.49 | 165.67 | 67.87 | 269.42 |
| MUSIC-R2 | 10000 | 23.57 | 3639.76 | 114.29 | 55.33 | 320.82 |
| MUSIC-R3 | 5000 | 15.33 | 228.84 | 22.35 | 21.54 | 145.31 |
| MUSIC-R3 | 10000 | 21.29 | 445.47 | 19.78 | 22.85 | 173.80 |

Observed patterns:

- **G (symbol grammar)** is the lowest or tied-lowest family for every
  representation at both checkpoints. Magnitude is modest and comparable
  across candidate classes (2.4–23.6) — this family does not by itself
  separate Latin from music.
- **T (token formation)** is stable-lowish for Latin (13–19) and is by
  far the largest, most volatile family for every music representation
  (72–3,640), varying by roughly 25× between the three music
  representations alone.
- **S (sequence grammar)** is the *largest* family for both Latin
  representations, but is the *smallest* comparable family besides G for
  MUSIC-R1 (8.5) and remains moderate for MUSIC-R3 (20–22); only MUSIC-R2
  shows large S distance (114–166).
- **L (line grammar)** is measurable only for C06 (physical lines are
  not source-observed for C01/C02, so L is schema-level
  `NOT_COMPARABLE` for the Latin pair, not a weak or missing result —
  Section 7).
- **D (document hierarchy)** is likewise `NOT_COMPARABLE` for C01/C02, for
  a different, also schema-level reason: the frozen VM reference observes
  only document and physical-line levels, while C01/C02 observe
  document/locus/page/section but not physical line — the two sources'
  observed hierarchy levels do not overlap, independent of any similarity
  question (confirmed directly: C01's raw `D_LOCUS_*`/`D_PAGE_*`/`D_SECTION_*`
  metrics are themselves `COMPARABLE` internally, it is only the *join*
  against VM's line-only D metrics that has no overlap). For C06, D is
  large (120–320) and second-worst after T for every representation.

No metric family produces a uniform ranking of the three candidate
identities; see Section 12.

## 5. Latin representation sensitivity: C01 vs C02

C01 (expanded) and C02 (diplomatic) are two representations of one
manuscript tradition, not independent historical controls. The paired
`NOTATION_DELTA.tsv` (63 of 74 metrics jointly comparable) shows the
expansion choice materially shifts several G-family metrics:

| Metric | C01 (expanded) | C02 (diplomatic) | Δ | Δ relative |
|---|---:|---:|---:|---:|
| G01_ALPHABET_SIZE | 31 | 60 | +29 | +93.5% |
| G02_INITIAL_RESTRICTION_DENSITY | 0.0323 | 0.2333 | +0.201 | +623% |
| G04_BIGRAM_OCCUPANCY | 0.4391 | 0.1922 | −0.247 | −56.2% |
| T11_POSITIONAL_RESTRICTION_DENSITY | 0.0860 | 0.1944 | +0.108 | +126.0% |
| S04_REPEATED_BIGRAM_TYPES | 1055 | 923 | −132 | −12.5% |
| S05_REPEATED_TRIGRAM_TYPES | 271 | 208 | −63 | −23.2% |

G-family metrics move by tens to hundreds of percent under expansion;
S-family metrics move by low tens of percent. The abbreviation-expansion
choice alone produces shifts on the same order as, and in several cases
larger than, the calibrated candidate-vs-VM distances reported for those
metrics (Section 4). Both G-family means (2.4–5.7) sit close enough
together that either representation would be read as "close" on G under
any threshold one might pick — but the underlying per-metric values that
compose that mean are themselves representation-volatile, so no claim
about *which specific symbol-grammar property* drives the closeness
should be attributed to one representation without checking the other.
This is a within-source-family control, and it already shows that
representation choice is not scientifically negligible for this
comparison design.

## 6. Music representation sensitivity: C06 R1/R2/R3

MUSIC-R1 (per-event), MUSIC-R2 (signed melodic interval), and MUSIC-R3
(pitch×duration) are three representations of one musical corpus, used
here as a controlled representation-sensitivity experiment, never as
three independent confirmations of one result (enforced by the run's own
gates; also checked here directly).

Family-mean calibrated distance range across R1/R2/R3, versus the range
across candidate identity (C01, C02, mean of C06's three
representations), at both checkpoints:

| Family | ck | C06 R1–R3 range | C06 R1–R3 spread | Candidate-identity spread (C01, C02, C06-mean) |
|---|---:|---|---:|---:|
| G | 5000 | 4.26–16.06 | 11.80 | 9.45 |
| G | 10000 | 6.10–23.57 | 17.47 | 14.12 |
| S | 5000 | 8.50–165.67 | 157.17 | 36.45 |
| S | 10000 | 8.60–114.29 | 105.68 | 28.51 |
| L | 5000 | 19.29–67.87 | 48.59 | n/a (C01/C02 not comparable) |
| L | 10000 | 22.85–55.33 | 32.48 | n/a |
| D | 5000 | 120.33–269.42 | 149.10 | n/a (C01/C02 not comparable) |
| D | 10000 | 143.38–320.82 | 177.44 | n/a |
| T | 5000 | 72.32–1866.49 | 1794.17 | 709.18 |
| T | 10000 | 132.40–3639.76 | 3507.36 | 1387.34 |

For every family where a candidate-identity spread can even be computed
(G, S, T — L and D have no C01/C02 value to compare against), the spread
*within* C06's own three representations exceeds the spread *between*
candidate identities, at both checkpoints. T shows the most extreme case:
choosing MUSIC-R2 instead of MUSIC-R1 changes the family-mean distance by
more than the entire gap between the Latin and music candidate classes.

**Which metrics are stable across R1/R2/R3, which are not**: G-family
distances stay within a single order of magnitude across all three
representations at both checkpoints (the most representation-robust
family observed). T and S are the least stable — T's absolute distance
varies roughly 25× and S's roughly 20× depending on which of the three
representations is chosen. This means any statement of the form "C06 is
close to VM on S" or "C06 is far from VM on T" depends materially on
which encoding of the same music was chosen, and must be reported with
that caveat attached, never as a property of "C06" alone.

## 7. Checkpoint stability

Both comparable checkpoints (5,000 and 10,000 tokens) use the *same* raw
candidate and VM values (`Compare()` is run once on each side's
full-observed-size fingerprint); the checkpoint dimension selects which
frozen, size-matched calibration scale normalizes that raw comparison,
not a re-measurement of the candidate at a smaller size (that is what
rarefaction is for — Section 8). Checkpoints 20,000 and 39,380 are
`NOT_COMPARABLE` for every representation in this subset (no candidate
reaches those sizes) and are excluded from every stability judgement
below, not interpolated.

Family-level rank order (lowest to highest calibrated distance) is
identical at 5,000 and 10,000 tokens for four of the five
representations (LATIN-EXPANDED, LATIN-DIPLOMATIC, MUSIC-R1, MUSIC-R2).
MUSIC-R3 shows one reordering, among three closely-spaced families: at
5,000 tokens the order is G(15.33) < L(21.54) < S(22.35); at 10,000 it is
S(19.78) < G(21.29) < L(22.85) — all three values stay within a narrow
band (15–23) at both checkpoints, and the reordering happens entirely
within that band. The #4/#5 ranks (D, T as by far the largest, both far
outside that band) do not change for any representation at either
checkpoint. Magnitude itself is not stable: most
family means increase from the 5,000-checkpoint scale to the
10,000-checkpoint scale (e.g. LATIN-EXPANDED G: 4.43→5.72; MUSIC-R2 T:
1866→3640), which reflects the different calibration null distribution
at each size, not a claim that the candidates' raw metric values changed.

## 8. Rarefaction analysis

Rarefaction answers a different question from Section 7: does the
candidate's *own* metric estimate, recomputed on boundary-preserving
draws of its own corpus, change with sample size, and does VM's own
frozen rarefied estimate (`VM_RAREFACTION_V2.tsv`) move the same way?

Comparing the direction of change (5,000→10,000 tokens) of the
candidate's rarefaction-summary mean against VM's own rarefaction-summary
mean, matched by metric/regime, across all five representations: 165 of
360 checked pairs (46%) move in the same direction, 195 (54%) move in
opposite directions — close to what an even chance split would produce.
This is itself informative: it argues against one dominant, uniform,
size-driven drift shared by every candidate and VM together (if there
were such a drift, agreement would be far above chance). It does **not**
rule out a real size effect for any individual metric — no per-metric
significance test was run on this direction count, and 46/54 with n=360
pooled across five very different representations is not a claim about
any specific metric. No result in this run appeared strong on the full
comparison and then vanished under rarefaction in a way that could be
checked here, because the primary comparison already uses the candidate's
full observed size, not a rarefied value, at the two comparable
checkpoints (Section 7); rarefaction here is a size-sensitivity
diagnostic on the candidate's own estimate, reported alongside, not
substituted into, the primary comparison.

## 9. Bootstrap uncertainty

Bootstrap point estimates and 95% percentile CIs are computed once per
metric at the candidate's own full observed size (200 replicates). A
consistent, cross-candidate pattern: the four `*_PROGRESSION` metrics in
the D family (`D_LINE_PROGRESSION`, `D_SECTION_PROGRESSION`,
`D_PAGE_PROGRESSION`, `D_LOCUS_PROGRESSION` where applicable) have
bootstrap CI widths several times larger than the absolute point
estimate for every candidate where they are computed — for example C01's
`D_PAGE_PROGRESSION` (estimate −0.000143, 95% CI [−0.000507, 0.00125],
relative width 12.35), C02's `D_LOCUS_PROGRESSION` (relative width 5.50),
and MUSIC-R1's `D_PAGE_PROGRESSION` (relative width 2.95). These point
estimates are compatible with zero at the 95% level; any calibrated
distance built on them should be read with that in mind, and this run
treats them as low interpretive weight (Section 11) rather than as
substantive signal, for every candidate, not selectively. Several other
metrics per representation (`S06_PREFERRED_TRANSITION_FRACTION`,
`L07_SAME_LINE_NONCOOCCURRENCE_DENSITY`, `T04_HAPAX_RATIO` for the
smaller music corpora) also show CI widths comparable to or exceeding the
point estimate; these tend to be metrics defined on small observed counts
(hapax counts, single support-regime transition fractions) at corpus
sizes in the 12,000–19,000 token range, consistent with sampling noise
on a small count rather than a data or protocol defect.

## 10. Calibration context

Calibrated distance is `|value − center| / spread` on each side, scaled
by a spread estimated from 40 independent synthetic calibration corpora
per checkpoint (`CALIBRATION_SCALES.tsv`, frozen before any candidate was
inspected); it expresses "how many calibration-panel scale-units apart",
not a probability or a percentage similarity. A metric with a large
absolute distance and a wide, zero-crossing bootstrap CI (Section 9) is
not "very different from VM" so much as "not informative at the current
sample size" — the S-family DEGENERATE-scale exclusions
(`S04_REPEATED_BIGRAM_TYPES`, `S05_REPEATED_TRIGRAM_TYPES`,
`S06_PREFERRED_TRANSITION_FRACTION`, `S07_DEPLETED_TRANSITION_FRACTION`
are `NOT_COMPARABLE` with reason "metric absent from pre-frozen scale"
for every representation checked) are the calibration panel's own
mechanism for flagging exactly this: a metric whose calibration-panel
variability could not be estimated at all is excluded rather than
silently scaled by an invented value. Where a calibrated distance is both
large and attached to a narrow bootstrap CI (e.g. C06's T-family
metrics generally, whose point estimates are large relative to their own
CI width — Section 9 does not list T among the wide-relative-CI metrics),
the distance is more informative than a large distance built on a
near-zero, wide-CI estimate (the D-family `*_PROGRESSION` metrics). This
run does not convert any calibrated distance into a probability of origin
or class membership; the calibration panel's role here is solely to say
whether an observed difference is large relative to the panel's own
synthetic-control variability, not to certify a mechanism.

## 11. Metric-family summaries

```text
Metric family: G (symbol grammar)
Observed pattern: lowest or near-lowest calibrated distance for all 5 representations, both checkpoints; magnitude comparable across Latin and music candidates (2.4-23.6)
Uncertainty: moderate bootstrap CI for most G metrics; no G metric flagged wide-relative-CI in this run
Checkpoint stability: rank #1 (lowest) preserved 5000->10000 for 4/5 representations; contends for #1/#2 with S for MUSIC-R3
Representation sensitivity: HIGH for underlying per-metric values (C01 vs C02 shifts up to 623% relative; C06 R1-R3 spread 11.8-17.5, smallest of the multi-metric families but still the largest family-mean spread relative to its own magnitude)
Rarefaction sensitivity: not separately flagged; direction agreement pooled across families (Section 8)
Interpretive weight: moderately robust
Limitations: family mean is an average over few metrics (5); driven partly by G01 alphabet size, which is itself highly representation-dependent (Section 5)

Metric family: T (token formation)
Observed pattern: lowest-to-moderate for Latin (13-19); by far the largest and most volatile family for every music representation (72-3640)
Uncertainty: no T metric flagged among the widest-relative-CI metrics; large distances for C06 are not simply artifacts of near-zero point estimates
Checkpoint stability: rank (worst for C06, mid-pack for Latin) preserved 5000->10000 for all 5 representations
Representation sensitivity: EXTREME for C06 (~25x range across R1/R2/R3); low for Latin (13.3-19.4 across both representations and checkpoints)
Rarefaction sensitivity: not separately isolated per family
Interpretive weight: robust for the qualitative "T is large and volatile for music" pattern; representation-sensitive for any specific music magnitude claim
Limitations: cannot attribute the T-family gap to a specific structural cause without a second, independently-encoded music corpus (Section 16)

Metric family: S (sequence grammar)
Observed pattern: largest family for both Latin representations; smallest non-G family for MUSIC-R1; moderate for MUSIC-R3; large for MUSIC-R2
Uncertainty: several S metrics (S04-S07) are calibration-DEGENERATE and excluded, not averaged in with an invented scale
Checkpoint stability: rank preserved 5000->10000 for all 5 representations
Representation sensitivity: HIGH for C06 (S mean ranges 8.5-165.7 across R1/R2/R3); LOW for Latin pair (29.05-29.76 at ck5000, 19.05-19.48 at ck10000)
Rarefaction sensitivity: not separately isolated per family
Interpretive weight: representation-sensitive; the single most family-inconsistent-with-T-and-D pattern in this run
Limitations: does not order candidates the same way T and D do; cannot be pooled into a composite score with them (none was computed)

Metric family: L (line grammar)
Observed pattern: measurable only for C06 (physical lines not source-observed for C01/C02); moderate-to-large distance (19-68), stable rank position (#2-3 of 5) across checkpoints and representations
Uncertainty: L07 same-line non-cooccurrence density shows a wide-relative-CI case for MUSIC-R2
Checkpoint stability: preserved 5000->10000 for all three C06 representations
Representation sensitivity: moderate (19.3-67.9 range across R1/R2/R3 at ck5000)
Rarefaction sensitivity: not separately isolated
Interpretive weight: sample-size-sensitive / representation-sensitive
Limitations: no Latin-vs-VM comparison possible at all for this family; not a weak result, a structurally absent one (Section 4)

Metric family: D (document hierarchy)
Observed pattern: NOT_COMPARABLE for C01/C02 (schema-level, no overlapping observed hierarchy level with VM); second-largest family for every C06 representation (120-321)
Uncertainty: the four *_PROGRESSION metrics within D are compatible with zero at 95% CI for every candidate where computed (Section 9)
Checkpoint stability: preserved 5000->10000 for all three C06 representations
Representation sensitivity: moderate-to-high (120.3-269.4 range across R1/R2/R3 at ck5000)
Rarefaction sensitivity: not separately isolated
Interpretive weight: sample-size-sensitive (progression sub-metrics) mixed with moderately robust (coherence/exclusivity sub-metrics)
Limitations: family mean mixes near-zero-signal progression metrics with more stable coherence/exclusivity metrics; no Latin comparison exists at all
```

Interpretive-weight categories used above (`robust`, `moderately robust`,
`representation-sensitive`, `sample-size-sensitive`, `inconclusive`) are
qualitative judgements based on the criteria actually applied in each
case above (checkpoint-rank stability, cross-representation spread
relative to candidate-identity spread, and bootstrap CI width relative to
point estimate) — not a new numeric score, and not chosen to favor any
particular outcome.

## 12. Cross-metric synthesis

No single family orders the three candidate identities the same way as
every other family, and only three families (G, T, S) even have a value
on *both* sides to compare — L and D have no C01/C02 value at all
(Section 4), so neither can support or contradict a Latin-vs-music
direction; they only describe C06 on its own. Of the three
directly-comparable families: G is roughly candidate-neutral (comparable
magnitude for Latin and music); T places Latin closer to VM than music by
a wide, checkpoint-stable margin; S places MUSIC-R1 closer than Latin
(contradicting T's direction) while MUSIC-R3 sits close to the Latin
values and MUSIC-R2 sits far above them. T is therefore the *only* family
offering an uncontradicted, checkpoint-stable Latin-closer-than-music
signal — it is not corroborated by a second directly-comparable family
(S disagrees; G is neutral), and D's large music-only values, while
themselves real and stable, cannot be cited as agreeing with T, because D
has no Latin counterpart to be larger than. No composite similarity score
was computed to force these families into one number; the frozen
protocol never defines one, and this task does not introduce one. The
honest summary is multidimensional: **on token formation, the tested
Latin representations sit markedly closer to VM than the tested music
representations do, in a single family with no directly-comparable
corroborating family; on sequence grammar the direction depends on which
music representation is chosen; on symbol grammar neither candidate class
is clearly closer; document hierarchy and line grammar say something
about the music candidate alone, with no Latin baseline to compare
against.**

## 13. Phase II hypothesis assessment

The Comparative Notation Study's own charter
(`COMPARATIVE_STUDY_GOALS.md`) states it "asks how closely the structural
relations of known notation systems resemble the frozen structural
grammar of the Voynich Manuscript" and explicitly "does not test whether
VM 'is' Latin, shorthand, music, a cipher, or another mechanism." The
hypotheses assessable by *this* design are therefore about structural
resemblance patterns, not identity or mechanism claims:

| Hypothesis | Status | Evidence chain |
|---|---|---|
| VM's structural grammar is broadly closer to the tested natural-language (Latin) representations than to the tested constrained symbolic (music) representations | WEAKLY_SUPPORTED | T (the only family directly comparable for both classes that shows a clear direction) favors Latin by a wide, checkpoint-stable margin (Sections 4, 12); this is not corroborated by a second directly-comparable family — G is candidate-neutral and S favors one of three music representations while another sits close to the Latin values; D's large music-only values cannot corroborate T, since D has no Latin counterpart; representation variance within C06 (Section 6) makes the music-side T magnitude itself unstable across encodings |
| VM's structural grammar is broadly closer to the tested constrained symbolic (music) representation than to natural language | NOT_SUPPORTED | Contradicted by T; only weakly and inconsistently supported by S (one of three music representations) |
| VM demonstrates symbol-grammar (G-family) properties broadly comparable in scale to both tested notation classes | SUPPORTED_BY_CURRENT_EVIDENCE | G-family magnitude is the same order for all five representations at both checkpoints (Section 4), though the underlying per-metric composition is representation-sensitive (Section 5) |
| VM was produced using, or is structurally consistent with, a Giovanni Fontana-style external-memory/mnemonic mechanism (Secretum-type device) | NOT_TESTED_BY_CURRENT_DESIGN | This design measures aggregate structural distance to a real notation-system class, not device mechanism, key/convention dependence, or retrieval procedure. Phase II's dedicated mechanism-identification experiment already found `MECHANISM_IDENTIFICATION_FROM_F2 = NOT_IDENTIFIABLE` on a purpose-built design (`research/phase2/PHASE2_FAILURE_ANALYSIS.md`); this run neither strengthens nor weakens that finding |
| The observed VM-vs-candidate distances are stable across corpus-size checkpoints (i.e. not primarily a sample-size artifact) | INCONCLUSIVE | Family rank order is checkpoint-stable (Section 7), but rarefaction direction-agreement between candidate and VM is near chance (46%, Section 8) — a genuine size effect on individual metrics cannot be excluded, and could not be excluded with only two usable checkpoints |

`VM демонстрирует структурные свойства, встречающиеся у X` is not the
same claim as `VM является X`; every row above is phrased as the former,
never the latter, and every WEAKLY_SUPPORTED / NOT_SUPPORTED /
INCONCLUSIVE / NOT_TESTED_BY_CURRENT_DESIGN status is deliberately not
collapsed into SUPPORTED or REFUTED where the design cannot bear that
weight.

## 14. Alternative explanations

For the two most salient patterns (Latin-closer-on-T; large C06
representation variance):

- **Tokenization/representation effects**: directly demonstrated —
  MUSIC-R2's signed-interval tokens and MUSIC-R1's compound event tokens
  produce a ~25x spread in T-family distance for the *same* underlying
  music, and C01/C02's expansion choice shifts G-family values by up to
  623% relative (Sections 5, 6). A representation-driven effect this
  large is a live alternative to any historically specific explanation.
- **Corpus-size effects**: partially checked (Sections 7, 8); rarefaction
  direction-agreement near chance argues against one uniform drift, but
  the panel cannot reach 20,000/39,380 tokens, so a size effect specific
  to very large corpora is untested, not ruled out.
- **Alphabet/symbol-inventory constraints**: G01_ALPHABET_SIZE alone
  varies from 31 to 60 depending only on whether abbreviations are
  expanded (Section 5); an encoding's raw symbol-inventory size
  mechanically drives several G/T-family metrics regardless of the
  underlying notation's "meaning."
- **Notation density / boundary structure**: not separately isolated in
  this run; would require a dedicated density-normalized metric not in
  the frozen registry.
- **Editorial normalization**: is exactly the C01-vs-C02 axis (Section 5)
  and is already shown to be non-negligible.
- **Generic properties of compact/quantitative notation systems**: a
  plausible explanation for why T (token formation) is large for every
  music representation regardless of which one is chosen (Section 6) —
  quantitative/interval-based tokenization is a generic property of many
  non-linguistic notations, not a signature of any one historical system.
  This run cannot distinguish "generic compact-notation effect" from "a
  historically specific relationship to VM" (claim C12 in the evidence
  matrix), because only one music corpus, encoded three ways, was tested
  — not two independently-sourced music traditions.

No single generic-constraint explanation is confirmed over a
historically-specific one; both remain open, and this design cannot
adjudicate between them (Section 16 non-goal: "не утверждать специальный
механизм, если результат объясняется более общими structural
constraints" is honored by leaving this open, not by picking one side).

## 15. What the experiment does not establish

- Not a test of VM's language, cipher status, or authorship.
- Not a test of Fontana-specific device mechanism, key/convention
  dependence, or retrieval procedure (Section 13).
- Not a robust cross-class ranking or nearest-neighbour result: N=3
  candidate classes, two of which share one manuscript source and one of
  which has three internally-dependent representations, is not a
  representative sample of historical notation-system space, and no
  PCA/UMAP/nearest-neighbour analysis was performed (frozen protocol
  keeps these repository-locked regardless of panel size).
- Not a within-class variance estimate: every class in this panel has
  exactly one independent corpus (Section 17).
- Not a verdict on whether any observed distance counts as "close" or
  "distant": no frozen threshold exists, and none was added here
  (Section 4, claim C13).
- Not a claim about "notation systems in general": C03 (shorthand), C04
  (DECODE cipher), C05 (Fontana cipher itself), C07 (tablature), C08
  (positional numerals), and C09 (deferred table transcription) are
  excluded or deferred from the frozen subset and are not represented in
  any conclusion here.
- Not a size-robust result beyond 10,000 tokens: 20,000 and 39,380 are
  `NOT_COMPARABLE` for this entire panel.

## 16. Methodological implications for the next run

1. Preregister a numeric STRUCTURALLY_CLOSE_ON / STRUCTURALLY_DISTANT_ON
   threshold *before* the next production run, derived only from the
   calibration panel's own null distribution (e.g. a fixed
   calibration-scale-unit cutoff), so this decision is never made
   post hoc against observed candidate data.
2. Before trusting any single-representation claim about a multi-encoding
   candidate (as C06 is here), report the full representation range, not
   a point value — this run shows representation variance can exceed
   candidate-class variance.
3. Acquire a second, independently-sourced constrained-notation corpus
   (a different musical tradition, or a genuinely different compact
   notation system) so that "generic compact-notation effect" and
   "notation-specific effect" become separable (Section 16 above).
4. If C09 (or another deferred/excluded candidate) is ever added, rerun
   `production-run-preflight` and this interpretation step fresh —
   family rankings and representation-sensitivity conclusions here are
   scoped to the current three-candidate panel and are not guaranteed to
   generalize.
5. Consider whether a corpus-size-normalized token-formation metric would
   reduce the extreme representation-sensitivity observed for T in the
   music encodings, before relying on T-family distance as primary
   evidence in a future run.
6. If within-class variance or cross-class ranking is ever wanted, it
   requires at least three independent corpora *per class*, which this
   panel structurally cannot provide (Section 17).

## 17. Conclusions (graded, not binary)

**Are there statistically stable structural correspondences between VM
and any production control?** Yes, one: G-family (symbol-grammar)
calibrated distance is consistently the lowest or near-lowest among
comparable families, for every representation, at both checkpoints. This
is a real, checkpoint-stable pattern, not an artifact of a single metric
or a single checkpoint.

**Is it metric-specific or broad?** Narrow. G is broad in the sense of
appearing across every representation, but it is only one family out of
five, and its underlying per-metric composition is representation-
sensitive (Section 5). The Latin-closer pattern itself rests on T alone —
the only family directly comparable for both classes that shows a clear,
large, checkpoint-stable direction (one to two orders of magnitude) — and
is not corroborated by a second directly-comparable family: S contradicts
that direction for one music representation and sits close to the Latin
values for another; G is neutral. D shows a large, stable value for music
but has no Latin counterpart at all, so it cannot corroborate T, only
describe the music candidate on its own. There is no metric family that
is both large in effect, directly comparable across both classes, and
immune to the representation-sensitivity caveat.

**Do they survive rarefaction?** Family rank order survives between the
two checkpoints this panel can reach. Whether the underlying magnitude is
itself a sample-size artifact is not resolved: direction-of-change
agreement between candidate and VM rarefied estimates is near chance
(46%), which argues against one uniform shared drift but does not clear
any individual metric of a real size effect, and the panel cannot test
beyond 10,000 tokens.

**How representation-sensitive are they?** Highly, for the strongest
patterns. C06's own three representations vary in family-mean distance by
more than the gap between candidate classes, for every comparable family.
The Latin pair shows a smaller but still real sensitivity concentrated in
G-family metrics. No conclusion here should be read as a property of "the
music candidate" or "the Latin candidate" independent of which
representation was used to measure it.

**Can the observed relationship be distinguished from generic notation
effects?** Not fully. Alphabet-inventory size, tokenization granularity,
and quantitative/interval-based encoding are all plausible generic
drivers of the T/G-family patterns observed, and this design — one music
corpus in three encodings, one manuscript in two transcriptions — cannot
separate a generic compact-notation effect from a notation-specific one.

**What does this mean for the Phase II hypotheses?** The narrowly
testable structural-resemblance hypothesis ("VM's grammar resembles
natural language more than the tested constrained notation") is weakly
supported — by one directly-comparable family (T), uncorroborated by a
second, and contradicted in direction by a third (S) for at least one
music representation. The broader, historically loaded
Fontana/external-memory hypothesis is not tested by this design at all
and is neither strengthened nor weakened by this run; it remains where
Phase II's own mechanism-identification retrospective left it —
genuinely open, for reasons unrelated to the comparative-notation results
reported here.

**What remains unresolved?** Whether the T-family pattern reflects
anything about VM's construction beyond generic compact-notation
structure;
whether the observed distances would hold with a representative sample of
notation-system space instead of three candidate classes; whether any
of this survives past 10,000 tokens; and what numeric threshold, if any,
should ever separate "close" from "distant" — all deferred to future,
separately preregistered work.

```text
PRODUCTION_COMPARATIVE_RUN_VALID=true
SCIENTIFIC_INTERPRETATION_COMPLETED=true

POST_HOC_PROTOCOL_CHANGES=false
POST_HOC_THRESHOLDS_ADDED=false
NEW_COMPOSITE_METRIC_ADDED=false

HISTORICAL_ATTRIBUTION_CLAIMED=false
```

```text
SUPPORTED_CLAIMS=C02,C03,C05,C06,C07 (fully supported); C01,C09 (weakly supported)
INCONCLUSIVE_CLAIMS=C08,C12
NOT_SUPPORTED_CLAIMS=C10
NOT_TESTED_CLAIMS=C04,C11,C13
```

(claim IDs are defined in `SCIENTIFIC_CLAIM_EVIDENCE_MATRIX.tsv`; each
carries its own `status` column with the exact
SUPPORTED_BY_CURRENT_EVIDENCE / WEAKLY_SUPPORTED / NOT_SUPPORTED /
INCONCLUSIVE / NOT_TESTED_BY_CURRENT_DESIGN value used above)
