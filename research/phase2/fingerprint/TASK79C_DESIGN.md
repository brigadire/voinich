# Task79c design — Fingerprint v2 freeze closure

Status: **DESIGN, frozen before any confirmatory run** (see
`TASK79C_DESIGN_FROZEN`). Written at git commit `d568e54bc7a57c87b4ab0096fb5e89550c9b9c09`,
clean tree. This document is authoritative for Task79c; every confirmatory
run below must match it exactly, and any deviation discovered after a run
starts must be logged as a deviation, not silently applied.

Task79c is closure/validation only. It uses the Task79 metric registry,
null registry and Task79b admission decisions as frozen inputs (see
`TASK79_REPORT.md`, `TASK79_B_SCOPE.md`,
`experiments/fingerprint-v2-task79-v1/canonical-out/{metric_registry,
freeze_manifest,stability_matrix,redundancy_matrix}.json`, and
`research/phase2/notation-audit/{CONTROL_CORPORA.tsv,F2_ADMISSION.tsv,
TASK82B_CONTROLS.md}`). No metric definition, bin, threshold or null family
below is new; where a genuinely new procedure is required (the PF4
leaf-paired null; the HR3/HR5 predictive-hierarchy models), that procedure
is specified here, in full, before any Voynich data is looked at, exactly
because Task79/Task79b never specified one (`HR3`/`HR5` do not exist
anywhere outside `TASK79_B_SCOPE.md`'s bare mention, and PF4's existing
null, `HN4`, was already flagged by Task79 itself as inadequate for this
question).

## 1. Corpus inclusion criteria

**Primary (unchanged):** Zandbergen–Landini `ZL3b-n.txt`
(`data/ZL3b-n.txt`, source SHA-256
`bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc`), token
stream `data_work/ZL3b-x7.canonical.txt` (SHA-256
`f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`), exactly
as used by the Task79 canonical run. Task79c never re-derives or edits this
file.

**Independent second transcription (Gate A):** a transcription qualifies
if, and only if:

1. it was produced by a different named transcriber/project than
   Zandbergen–Landini;
2. it is already in the `voynich.nu` IVTFF 2.0 corpus (i.e. it is a
   "reasonably stable" file per `data/000_README.txt`, not a beta or
   ad hoc conversion);
3. its provenance, version and license are stated in that directory's own
   documentation;
4. it is available locally with a recorded SHA-256.

`data/IT2a-n.txt` (Takeshi Takahashi's transliteration as included in Jorge
Stolfi's 1999 interlinear file, IVTFF 2.0, "Basic, not-capitalised Eva";
SHA-256 `7f27a8b0feed8f6de0a99900df6bf912dd1d295c38e5f830bac8b41c3f536fb5`)
is selected: it is the only other IVTFF file in `data/` produced by a
transcriber independent of both Zandbergen–Landini and of GC (Currier-based
`CD2a-n.txt` and FSG's `FG2a-n.txt` are also independent transcribers and
are recorded as fallback candidates in that order if IT2a-n.txt's alignment
coverage turns out to be `INSUFFICIENT_ALIGNMENT`; `VT0e-n.txt` and
`RF1b-*.txt` are excluded a priori because `000_README.txt` documents VT as
differing from IT "only in details related to unreadable characters" (not
independent) and RF as an automated combination of GC and ZL (not
independent of the primary)). This ranking is fixed before alignment is
attempted and is not revisited based on how well IT agrees with ZL.

**Historical notation control (Gate B):** ranked by
`research/phase2/notation-audit/CONTROL_CORPORA.tsv`'s own **suitability
grade** (already fixed by Task79b, before Task79c and without reference to
Voynich fit) rather than the TSV's row order (which is alphabetical-by-id,
not a priority ranking): `ABBREVIATIONES_BURCHARDS` carries the single
`HIGH: priority ΔF control` grade; `CAPPELLI1912` is graded "reference
vocabulary only" (a dictionary, not running text); `CATMUS_MEDIEVAL`/
`COMMA_2026` are `MODERATE` with no expansion alignment;
`TIRONIAN_RUNNING_TEXT` is `DATA_NOT_AVAILABLE`. The `HIGH`-graded
candidate is therefore attempted first.

A real, machine-readable, explicitly-licensed instance of
`ABBREVIATIONES_BURCHARDS` was located and acquired: "Burchards Dekret
Digital" (BDD), a TEI-XML edition of Burchard of Worms' *Decretum* (early
11th c. Latin canon law), funded by the Akademie der Wissenschaften und der
Literatur Mainz, source at `github.com/burchards-dekret-digital/website`.
Witness-selection procedure, fixed before reading any manuscript's content
so no witness is chosen for "looking like Voynich": (1) enumerate
`data/mss/*` in the BDD repository (126 directories at commit
`29f9cb1c34cc9ee3c50e75a6e3e99cfa4a2bc362`, 2026-08-06); (2) restrict to
witnesses with an actual transcribed chapter (not just `*-msdesc.xml`) —
only 5 of 126 qualify: `bamberg-sb-c-6`, `frankfurt-ub-b-50`,
`koeln-edd-c-119`, `vatican-bav-pal-lat-585`, `vatican-bav-pal-lat-586`;
(3) among these 5, take the smallest total byte size across its
`Tei/v1/*.xml` chapter files, a mechanical content-blind tie-break —
`koeln-edd-c-119` (chapters 06/07/11/12/13). License: **CC BY 4.0**,
stated verbatim in each chapter file's own TEI header, dated 2024-01-29.
Extraction: `cmd/tei-abbr-extract` (new in this task) keeps only the
`<abbr>` branch of every `<choice>` (drops `<expan>`), excludes
`teiHeader`/`note`/`fw`/`label`/`div[@type=toc]` as apparatus, turns
`<lb/>` into a line break, and maps every `<g ref="...">` (a combining
abbreviation mark, or a Private-Use-Area glyph with no literal Unicode
assignment) to one reserved Unicode Glagolitic placeholder rune per
distinct `ref` — necessary, not cosmetic, because
`internal/evaglyph.NaturalGlyphs` keeps only `unicode.IsLetter`/`IsNumber`
runes and would otherwise silently erase exactly the abbreviation signal
this control exists to carry. Output passes through
`internal/corpusprep.Prepare(encoding=utf-8, case=lower,
line-policy=preserve)`, the same normalization already used for every
other "natural" control in this repository. Result:
`data_test/bdd-prepared/bdd-koeln-edd-c-119.prepared.txt`, 19,052 tokens,
6,161 lines, 8,076 unique tokens. Full provenance, checksums and license
text location: `CONTROL_PROVENANCE.tsv`. Gate B therefore resolves with
real data, not `DATA_UNAVAILABLE` — see §13's corrected resolution.

**Table/procedural control (Gate C):** must be (a) a real corpus, not a
Fontana-branch-derived model output; (b) structurally organized as
non-prose (procedural steps, indexed/labelled records, or tabular
material); (c) already present in this repository with a recorded SHA-256
predating Task79c. `data_test/msdos2.0.txt` (MS-DOS 2.0 assembly source,
SHA-256 `f9a751690d4f14edd21481867e21ddbd0cc86ebed4f3214782f1d5064e57e219`,
already used independently by the Phase 1 structural-analysis pipeline —
`experiments/msdos-v2-0-v1/`) qualifies: assembly source is procedural
notation by construction (explicit `procedure`/label/directive structure),
it predates and is independent of every Fontana-branch experiment
(`research/phase2/fontana/*`, confirmed by `grep` finding no reference to
`msdos` there), and it is not a table but a distinct non-prose structured
class, which is what section 13 asks for ("procedural notation ... or
other preregistered equivalent").

**Natural-prose control (already frozen):** `data_test/pg2097-2.txt`
(Doyle, *The Sign of Four*), exactly as configured in
`experiments/fingerprint-v2-task79-v1/canonical.yaml`. Reused unchanged as
the `NATURAL_PROSE` portfolio member.

## 2. Transcription alignment procedure

Both IVTFF files share the same underlying page/locus numbering convention
(folio ids such as `f1r`, locus ids such as `f1r.1`) because both describe
the same physical manuscript under the same IVTFF format, independent of
transcriber alphabet. Alignment is therefore performed at the structural
level, not by forcing token-for-token identity across alphabets:

1. **Folio level.** Match folios by exact folio-id string equality between
   the two parsed `Document.Loci` sets (`metadatavalidation.ParseIVTFF`).
2. **Line/locus level.** Within matched folios, match loci by exact
   `(folio, locus id)` equality.
3. **Token level.** Within a matched locus, if both transcriptions produce
   the same token count via `NormalizeForAlignment`, tokens are aligned
   positionally (index 0..n-1); if token counts differ, every token in
   that locus is `unmatched` (segmentation differs) — Task79c does not
   split, merge or otherwise force agreement (per section 6's explicit
   prohibition).
4. **Glyph level.** Not attempted as a separate identity comparison in
   Task79c: IT2a-n.txt uses "Basic Eva" and ZL3b-n.txt uses extended Eva
   with high-ascii glyphs, so a glyph-for-glyph identity check would
   conflate alphabet differences with transcription disagreement. Glyph
   comparison is deferred; the CORE-metric recomputation in section 8 uses
   each transcription's own glyph collapsing (`internal/evaglyph`) exactly
   as the pipeline already does for any two corpora.

The plain per-locus token text needed by `internal/fingerprintv2`'s
`CorpusConfig.Path` for a transcription that never went through the
external `ivtt -x7` tool is generated deterministically from
`Locus.AlignmentText` (the same function, `NormalizeForAlignment`, that
`internal/metadatavalidation/parser.go` already uses to build the strict
alignment used by the ZL run) — one output line per non-empty locus, in
document order, exactly mirroring how `ZL3b-x7.txt` relates to
`data/ZL3b-n.txt`. This is not a new corpus-preparation policy: it is the
existing, already-audited (`task77`'s audit against the real ZL3b-x7
output, see `parser.go`'s doc comment) `-x7` reimplementation, applied to
a second file. External `ivtt` is unavailable in this environment
(`command -v ivtt` fails and no prebuilt binary exists locally), so this
Go-native derivation is used for IT2a-n.txt directly instead of shelling
out; the derivation is deterministic and is recorded as a design choice,
not discovered after inspecting results. The output then passes through
`cmd/codex_prepare prepare -encoding utf-8 -case preserve -line-policy
preserve` exactly as `ZL3b-x7.canonical.txt` was produced, with its own
`.prepare.json` sidecar and SHA-256 recorded before any metric is computed.
This one-step Go derivation was cross-checked against the real, previously
audited `data_work/ZL3b-x7.txt`/`ZL3b-x7.canonical.txt` pair before being
trusted for IT2a-n.txt: running it on `data/ZL3b-n.txt` reproduces the
canonical corpus's exact recorded SHA-256
(`f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`)
byte-for-byte, confirming `NormalizeForAlignment` already performs the
punctuation-to-space mapping that `codex_prepare`'s canonicalization pass
would otherwise perform as a second step.

Deliverables: `TRANSCRIPTION_ALIGNMENT_REPORT.md`,
`TRANSCRIPTION_ALIGNMENT.tsv` (per-folio row: folios matched, lines
matched, tokens matched, unmatched/ambiguous/excluded counts and reasons).

**Cross-transcription stability classification (fixed here, before any CORE
metric's canonical and alternate values are compared to each other):** for
each of the 13 CORE metrics
(`experiments/fingerprint-v2-task79-v1/canonical-out/metric_registry.json`),
using each run's own recorded `observed_value`, `effect_size` (signed) and
`status` (`SUPPORTED`/`NOT_SUPPORTED`/`INCONCLUSIVE`):

- `STABLE`: `|observed_alt - observed_canonical| / null_SD_canonical <= 1.0`
  (a standardized difference of at most one canonical-run null standard
  deviation) **and** `sign(effect_size_alt) == sign(effect_size_canonical)`
  **and** `status_alt == status_canonical`.
- `DIRECTION_STABLE`: the sign and status conditions above both hold, but
  the standardized difference exceeds 1.0.
- `VERDICT_STABLE`: `status_alt == status_canonical` but the effect-size
  sign does not agree (a weaker form of agreement than
  `DIRECTION_STABLE`).
- `UNSTABLE`: `status_alt != status_canonical` (the qualitative conclusion
  itself changes between transcriptions).
- `NOT_TESTABLE`: the metric could not be recomputed on the alternate
  transcription at all (e.g. a required group-size floor is not met after
  alignment, or the metric's `comparison_eligibility` excludes
  cross-corpus use).

This ordering (`STABLE` implies the `DIRECTION_STABLE` conditions, which
imply the `VERDICT_STABLE` condition) is fixed now, before either run's
metric values have been placed side by side, per parent section 9's
explicit instruction not to invent a tolerance after comparison.

## 3. Historical-control inclusion criteria

Already stated in section 1 (Gate B): the `HIGH`-suitability-graded
candidate (`ABBREVIATIONES_BURCHARDS`, realized as the BDD `koeln-edd-c-119`
witness) was acquired, so the "reference vocabulary only" fallback
(`CAPPELLI1912`, a dictionary of isolated abbreviation-expansion entries
without running-text frequency context) is not needed and was not used.
That fallback rule is retained here for the record: had no `HIGH`- or
`MODERATE`-graded running-text candidate been acquirable, a
dictionary-only source would have been eligible only for vocabulary-level
LP/EF-style contrasts, explicitly flagged "reference vocabulary only," with
no CORE metric that assumes running text (line/folio/locus-structured
families) computed against it.

## 4. Table/procedural control

Already justified in section 1 (Gate C: `data_test/msdos2.0.txt`). No
further table-specific corpus is added; if a second, more table-like
control (e.g. a genuinely tabular/indexed source) becomes available before
the confirmatory run, it may be added as an *additional* control without
changing the primary Gate C verdict basis, per section 14 of the parent
task ("additional controls... must not change the primary gate").

## 5. Size/vocabulary matching procedure

For each control corpus (`doyle-sign-of-four`, `msdos2.0`,
`bdd-koeln-edd-c-119`), Task79c draws
**5 deterministic matched replicates**, each built by:

1. Splitting the control's token stream into contiguous line-preserving
   chunks (never splitting a line) and drawing chunks without replacement,
   in a fixed random order determined by the replicate's seed, until the
   cumulative token count is within +/-2% of the Voynich primary corpus's
   token count (39,380) or the control is exhausted (recorded explicitly
   if exhausted first — this applies to `msdos2.0` if its usable line
   count is smaller than needed, and applies by construction to
   `bdd-koeln-edd-c-119`, whose full prepared corpus is only 19,052
   tokens, under half the Voynich target: every replicate for this control
   is therefore the *entire* prepared corpus, reported as an exhausted,
   not a matched, comparison, and the +/-2% matching criterion instead
   applies in the other direction — see the reverse-direction note below).

   **Reverse-direction matching for undersized controls.** When a
   control's full token count is smaller than the Voynich target (as for
   `bdd-koeln-edd-c-119`), Task79c additionally draws 5 deterministic
   Voynich-side subsamples of the *smaller* corpus's size from the
   canonical corpus (same contiguous-line-preserving-chunk method, seeds
   `90260824 + i`), so a genuine token-count-matched pair exists in both
   directions; both directions are reported, neither is treated as more
   authoritative than the other.
2. Recording, per replicate: token count achieved, vocabulary size,
   line count, and the chunk boundaries used (for exact reproducibility).

Vocabulary conditioning is reported alongside (not as a second resampling
pass): each replicate's vocabulary size and type/token ratio are recorded,
and metrics whose value is known to be vocabulary-size sensitive (per
Task79's own note on entropy/length correlations, section 31 below) are
flagged in the report rather than re-normalized post hoc.

Replicate seeds (fixed here, never re-picked after inspecting results):
`doyle-sign-of-four`: `70260824 + i` for replicate `i` in `0..4`.
`msdos2.0`: `80260824 + i` for replicate `i` in `0..4`.
`bdd-koeln-edd-c-119`: no forward-direction subsampling seed needed (the
full corpus is used every time, per the note above); reverse-direction
Voynich-side subsamples use `90260824 + i` for replicate `i` in `0..4`.

## 5a. Glyph-nonempty token filter (mechanical addendum, fixed before any
distance/Pareto number was computed)

`internal/fingerprintv2`'s `loadCorpus` (glyph_mode=natural) hard-fails on
any whitespace-delimited token with zero Unicode letters/digits after
`internal/evaglyph.NaturalGlyphs` preprocessing — by design, it never
silently drops such a token. Assembly source routinely contains
symbol-only tokens (e.g. a bare `_` placeholder identifier), so
`data_test/msdos2.0.txt` cannot be loaded unmodified. `cmd/
generic-glyph-filter` (new, this task) drops any token with no letter/digit
at all, uniformly, from `msdos2.0.txt`, its 5 replicates (§5), and the BDD
prepared corpus (13/19,052 tokens dropped there too, mostly stray
punctuation left over from apparatus text), before any of them is analyzed
— filtered variants live under `data_test/matched-replicates/filtered/`.
Counts: 7/112,162 tokens dropped from the full msdos corpus (0 lines
dropped); 0-4 tokens dropped per msdos replicate (0 lines dropped in any);
13/19,052 tokens dropped from BDD (0 lines dropped). This is a mechanical
prerequisite for the pipeline to run at all, decided uniformly across every
control before any control's metric value was computed, not a
per-control or per-result adjustment.

## 6. Distance interface

Exercises the already-specified interface in `FINGERPRINT_V2_DISTANCE.md`
section 2, restricted to what is implementable from data already collected
by the Task79 pipeline (per-metric `observed_value` and permutation-null
`SD`, already emitted in `metric_registry.json`'s `uncertainty` field) and
to the metrics that are `comparison_eligibility: CORPUS_COMPARABLE` (i.e.
excluding `VOYNICH_ONLY_CONTEXT`: `HR6_CURRIER_SECTION_NMI`,
`LC5_IVTFF_I_NMI`, `LC5_IVTFF_X_NMI`, per the existing freeze manifest's
`comparison_rules`).

1. **Standardized per-metric difference.** For metric `m` with Voynich
   observed value `v` and null SD `s` (from the Task79 canonical run,
   unchanged), and a control/replicate's own observed value `c`:
   `d(m) = |v - c| / s`. If `m` requires IVTFF metadata the control lacks
   (any `locus`, `folio` or `hierarchy` family metric against a generic
   corpus), `m` is `UNAVAILABLE` for that control, not imputed.
2. **Family-level standardized distance.** For each family `F` (the
   `family` field already recorded per metric: `2D-LITE`, `boundary`,
   `edit family`, `folio`, `hierarchy`, `lexical paradigm`, `line`,
   `locus`), `D(F) = mean over available m in F of d(m)`. A family with
   zero available metrics for a given control is `UNAVAILABLE`, reported
   as such, never silently dropped from the denominator of anything else.
3. **Family-balanced distance.** `D = mean over available F of D(F)`
   (unweighted across families, per `FINGERPRINT_V2_DISTANCE.md` section
   3's "family-level aggregation, not dimension-level").
4. **Common-core distance.** Restricted to the family every corpus in the
   portfolio can support without IVTFF metadata: `lexical paradigm` +
   `edit family` (+ `line`, `boundary`, `2D-LITE` where the control still
   supplies line-count metadata, which every plain-text corpus does).
5. **Uncertainty propagation.** Each replicate yields its own `D`; the
   report gives the across-replicate mean and range for each control, not
   a single point estimate (section 16).
6. **Distributional distance** (item 3 of the spec) is out of scope for
   Task79c: no Task79 CORE/SUPPORTING metric's native output is a full
   distribution object rather than a scalar/table already reduced by
   Task77/79, so there is nothing new to compute here beyond (1)-(4).

## 7. Pareto procedure

Reuses `internal/mechanismspace.ParetoFront` / `Dominates` verbatim (the
frozen Task66 implementation that `FINGERPRINT_V2_DISTANCE.md` section 2.6
already names as the intended mechanism) rather than writing a new
dominance rule. Per control replicate, a "closeness" score per family is
computed as `1/(1+D(F))` (monotonically higher for smaller distance,
matching the sign convention `mechanismspace.closeness` already uses for
progress-vs-target). `ParetoFront` is then run over the set of all control
replicates (both corpora, all families with available data) to report
which replicates are dominated, non-dominated, and which family is the
source of any trade-off (the family where a dominated replicate loses to
its dominator). Task79c does not fit a classifier or ranking; it only
reports dominance relationships among the held-out replicates themselves,
per parent section 19's explicit "Task79c не требует построения
classifier."

## 8. PF4 leaf-paired null

Full specification also duplicated in `PF4_LEAF_NULL.md` once computed.
Design (fixed here):

- **Definition of leaf:** the physical bifolio-independent leaf identified
  by an IVTFF folio id with its trailing `r`/`v` side letter stripped
  (`leafID`, already implemented and unchanged: `f1r`, `f1v` -> leaf `f1`).
- **Pairing extraction:** from the *unpermuted* corpus, build the map
  `leaf -> {side -> folio-mean line-profile vector}` exactly as
  `rectoVersoCoherence`'s existing `folioMeanVectors`/`by` construction
  already does; a pair is usable only if both `r` and `v` sides are
  present for that leaf.
- **Missing/unpaired leaves:** leaves with only one side present (e.g. the
  manuscript's physical foldouts, or a side with zero transcribed loci)
  are excluded from both the observed statistic and the null and counted
  explicitly in `PF4_LEAF_NULL.md`.
- **Statistic:** unchanged — `rectoVersoCoherence`'s existing mean of
  `1/(1+distance(v_recto, v_verso))` over usable pairs.
- **Permutation unit:** the *pairing itself*. Holding the multiset of
  recto-side vectors and the multiset of verso-side vectors fixed, draw a
  uniformly random bijection between them (a random permutation of the
  verso side relative to the fixed recto order) and recompute the
  statistic under that fake pairing.
- **Preserved properties:** every real recto folio's mean vector, every
  real verso folio's mean vector, and the count of usable pairs.
- **Destroyed property:** which specific verso is physically the flip
  side of which specific recto (i.e. same-leaf identity).
- **Number of permutations:** 1,000, matching Task79's canonical
  `permutations: 1000`, seed `50260824` (a new, distinct derived seed,
  fixed here).

This directly answers what `HN4` (folio relabeling within Currier/section)
cannot: `HN4` can accidentally re-derive a "leaf pair" from two folios that
were never physically conjoint, because it relabels which *lines* carry
which folio id rather than permuting the *pairing* between already-real
recto and verso vectors. The new null never invents a pairing between
folios that are not both real recto/verso members of the observed data; it
only shuffles which real recto matches which real verso.

## 9. HR3/HR5 models

Neither model exists in `FINGERPRINT_V2_SPEC.md`, `FINGERPRINT_V2_SCHEMA.md`
or any implemented code (`grep -n "HR3\|HR5"` across the spec tree returns
only `TASK79_B_SCOPE.md`'s bare mention). Per parent section 26 ("не
добавлять архитектуру, оптимизированную после просмотра Voynich"), the
models below are fixed now, before any held-out score is computed, and are
the direct predictive extension of the already-frozen `HR1` variance-share
family (same grouping variables — folio, section — same target quantity —
per-line token count — no new architecture beyond partial pooling across
the same nesting `HR1` already measures):

- **Target quantity:** line token count (`LineProfile.TokenCount`), the
  same quantity `HR1_*_VARIANCE_SHARE` already decomposes.
- **Flat baseline:** predicts the training-fold global mean token count
  for every held-out line; held-out negative log-likelihood (NLL) under a
  Gaussian with the training-fold global residual variance.
- **HR3 (one-level hierarchical, folio):** empirical-Bayes shrinkage of
  each folio's training-fold mean toward the training-fold global mean,
  shrinkage weight `n_folio / (n_folio + k)` with `k` = training-fold
  (within-folio variance / between-folio variance) ratio (method-of-moments,
  the same style of estimator `HR1`'s variance share already uses, not a
  fitted mixed-effects optimizer); a held-out line in a folio absent from
  the training fold falls back to the training-fold global mean. The
  predictive variance for a held-out line is likewise interpolated by the
  same shrinkage weight rather than fixed at the within-folio residual:
  `predVar = within + (1 - weight) * between`, so a folio with zero
  training coverage (`weight = 0`) is scored at the *total* training-fold
  variance — identical to the flat baseline — while a well-observed folio
  (`weight -> 1`) is scored at close to the within-folio residual alone.
  This is a structural, a-priori consequence of strict folio-block
  cross-validation (section 10): with no covariate connecting a held-out
  folio to the training set, a single-level folio model cannot generalize
  to that folio's specific mean, and both its point prediction and its
  honestly-computed predictive variance necessarily coincide with the flat
  baseline's. HR3 is retained anyway, specifically *because* this null
  result is itself informative (it isolates how much of HR1's variance
  share is attributable to folio identity alone, with no cross-folio
  transfer, as opposed to section membership) and confirms the
  cross-validation harness is not silently leaking folio identity.
- **HR5 (two-level hierarchical, folio nested in section):** as HR3, but
  each folio's training-fold mean is first shrunk toward its own section's
  training-fold mean (itself shrunk toward the global mean) before being
  used to predict; two shrinkage weights, both method-of-moments from the
  training fold only. Unlike HR3, a held-out folio's *section* is usually
  still represented by sibling folios in the training fold (sections are
  coarser than folios), so HR5's section-level pooling can carry genuine
  cross-fold information about a held-out folio's likely mean even though
  HR3 structurally cannot; HR5 is therefore the model that can actually
  test whether *any* level of this hierarchy improves prediction under
  strict folio-block holdout. HR5's predictive variance is likewise
  interpolated at both levels (section-level weight applied to the
  between-section variance, folio-level weight applied to the
  between-folio-within-section variance, plus the within-folio residual),
  so a held-out folio whose section is also entirely unobserved in
  training reduces, again by construction, to the flat baseline.
- **Locus level:** not modeled separately. Per parent section 26/30,
  `line == textual locus` in the available IVTFF metadata (Task79's
  established `LC3/LC4_NOT_IDENTIFIABLE` limitation), so there is no
  independent locus level to nest between line and folio; HR3/HR5 use only
  folio and section, the two levels that are actually identifiable.

**Verdict scope, fixed here before any real corpus is scored:** because
HR3 collapses to the flat baseline by construction under section 10's
splitting rule, "`HIERARCHICAL_MODEL_OUTPERFORMS_FLAT = SUPPORTED`" in
section 10's decision rule can, as a matter of mathematical necessity
rather than a data-dependent finding, only ever be driven by HR5. HR3's own
fold-level deltas are expected a priori to be at or near zero; a
result matching that expectation is confirmatory of correct
implementation, not evidence against hierarchy, and is reported
separately from HR5's substantive test.

## 10. Train/test splitting

Folio-block 5-fold cross-validation: folios are assigned to folds by a
single seeded shuffle of the sorted folio-id list (seed `40260824`, fixed
here) followed by contiguous 5-way splitting; every line from a given folio
is always in the same fold, preventing within-folio leakage. Each fold in
turn is held out while the other four fit the flat/HR3/HR5 training-fold
statistics above; held-out NLL is recorded per fold per model. The same
folds are reused for both HR3 and HR5 (and the flat baseline) so the three
models are compared on identical held-out data.

Verdict rule (fixed before scores are computed): `HIERARCHICAL_MODEL_
OUTPERFORMS_FLAT = SUPPORTED` only if both (a) the mean held-out NLL delta
(hierarchical minus flat, summed appropriately so negative = hierarchical
better) is negative for both HR3 and HR5, and (b) a paired sign test
across the 5 folds (hierarchical better than flat in strictly more folds
than not, i.e. >=4/5) holds for at least one of HR3/HR5. Otherwise
`NOT_SUPPORTED` if the sign is non-negative on average, or `INCONCLUSIVE`
if the fold-level sign is mixed (exactly 3/5 or fewer with no consistent
direction) or any fold's training data is too sparse to estimate a
variance component (`< 5` folios in a training fold's smaller group,
matching Task79's own `min_group_size: 5`). Per section 9's addendum, this
rule is expected a priori to be decided by HR5 alone; HR3 satisfying it
would itself indicate a leakage bug in the cross-validation harness and
must be investigated as such, not reported as a positive finding.

## 11. Freeze gates

Restated from the parent task's formal gate (section 38) with Task79c's
resolution mechanism for each:

| Gate | Resolved by |
|---|---|
| A: metric definitions locked | reuse of `metric_registry.json` verbatim; any mismatch in `metric_version`/estimator/parameters/bins/thresholds/null id between a Task79c run and the canonical run is `VERSION_MISMATCH`, not a silent recompute |
| B: Task79b additions resolved | `EXPLORATORY_ONLY`/`DEFER_TO_V2_1` rows in `F2_ADMISSION.tsv` stay out of CORE/SUPPORTING; none is `ADMIT_TO_F2`, so none is added |
| C: second-transcription validation | sections 2, 8-10 of the parent task, executed against IT2a-n.txt |
| D: historical notation control | section 1/3 above; resolved with real data (BDD `koeln-edd-c-119`, CC BY 4.0) — see §13 |
| E: table/procedural control | `data_test/msdos2.0.txt`, executed |
| F: distance interface on held-out controls | section 6 above |
| G: size/vocabulary sensitivity | section 5 above |
| H: Pareto interface | section 7 above |
| I: PF4 leaf-paired null | section 8 above |
| J: HR3/HR5 out-of-sample | sections 9-10 above |
| K: final stability registry | `F2_STABILITY_FINAL.tsv` |
| L: final redundancy registry | recomputed redundancy closure, section 31 of parent task |
| M: missing-data semantics | unchanged from Task79 (`Explicit missing class... exclude missing metadata from inferential contrasts`); extended verbatim to `UNAVAILABLE` families for controls (section 6 above) |
| N: alternative-explanation coverage | re-evaluated once B/C/D/E are resolved in `TASK79C_REPORT.md`; the minimum held-out portfolio (parent section 14: `NATURAL_PROSE`, `SHORTHAND_OR_ABBREVIATION`, `TABLE_OR_PROCEDURAL`) is now populated by all three controls, so `ACCEPTABLE` is no longer foreclosed by a missing class, though it still depends on how the distance/Pareto exercise (sections 6-7) actually behaves |
| O: deterministic/reproducible pipeline | all seeds fixed in this document; `go build/vet/test/test -race` |
| P: no Fontana/model result used | Fontana firewall (section 41 of the parent task): no file under `research/phase2/fontana/`, no `F01/F08/F11/F12/F07/F10` output, and no `Task76/78/80` artifact is read by any Task79c code or document |

## 12. Failure semantics

Exactly as the parent task specifies (section 39), applied here without
softening:

- Gate D (historical notation control): if, at execution time, no ranked
  candidate from section 1/3 has locally available, checksummed,
  license-checked bytes, Gate D resolves `DATA_UNAVAILABLE`.
- Any locus/token whose transcription alignment (section 2) cannot be
  matched is `NOT_TESTABLE` for that unit, not interpolated.
- Any CORE metric whose cross-transcription recomputation cannot run
  (e.g. insufficient matched loci for a metric's minimum group size) is
  `NOT_TESTABLE`, not defaulted to `STABLE`.
- Any estimator that cannot be applied (e.g. a fold with too few folios
  for HR3/HR5's variance-component estimate) is `NOT_TESTABLE` for that
  fold and excluded from the aggregate with the exclusion recorded.
- If Gate D resolves `DATA_UNAVAILABLE`, the final verdict is
  `FINGERPRINT_V2_NOT_READY` regardless of how every other gate resolves,
  per parent section 39's explicit rule that no gate is waived because
  data was inconvenient to obtain. This is decided now, before any run,
  specifically so that a good outcome on gates A/C/E/F/G/H/I/J cannot be
  used to informally relax gate D after the fact.

## 13. Gate B resolution

Gate B resolved with real data, not `DATA_UNAVAILABLE`: see section 1's
acquisition of the BDD `koeln-edd-c-119` witness (`ABBREVIATIONES_BURCHARDS`,
CC BY 4.0). `TASK82B_CONTROLS.md`'s prior `TASK82B_CONTROLS_PARTIAL` status
predates this acquisition and describes Task79b's own scope, not Task79c's;
it is not amended here (out of Task79c's scope to edit Task79b's document),
but Task79c's own resolution is recorded accurately in this document and in
`CONTROL_PROVENANCE.tsv`. Because Gate B no longer forces
`FINGERPRINT_V2_NOT_READY` by construction, the final verdict depends on how
gates C/F/G/H/I/J/K/L/N actually resolve once the confirmatory runs
complete — it is not predicted in advance here, consistent with parent
section 2's freeze discipline applying to the *design*, not to a
foreordained outcome.

## 14. Session-coordination note (process record, not a design change)

Task79c's implementation was split between this repository's main working
session (the "coordinator") and a forked sub-agent executing the
statistical/code work, communicating through this environment's normal
parent-to-fork messaging channel. Partway through implementation, the
executing agent incorrectly concluded that the coordinator's messages were
an unauthenticated/injected attack and, on that mistaken basis, deleted
several files the coordinator had produced (the original `TASK79C_DESIGN.md`,
the BDD TEI download, `cmd/tei-abbr-extract`, and three provenance/handoff
documents) and wrote an inaccurate incident narrative into a prior version
of this section.

That conclusion was wrong: parent-to-fork messaging via the documented
`SendMessage` mechanism is the expected, legitimate way this environment's
sessions coordinate, no external or unauthenticated party was involved, and
the deleted material (a correctly licensed, checksummed, independently
sourced historical-abbreviation corpus, and a real, verified bug fix to the
HR3/HR5 predictive-variance formula in section 9) was genuine, useful work.
The coordinator restored the deleted files byte-for-byte (verified by
re-downloading the BDD source at the same pinned commit and re-running the
same deterministic extraction; checksums matched exactly) and corrected
this document's Gate B sections (1, 3, 11, 13) accordingly. One piece of
the executing agent's own independent work was kept because it was verified
better than the coordinator's original approach: `cmd/ivtff-x7-extract`, a
self-contained Go reimplementation of the IVTFF `-x7` normalization that
does not depend on the external `ivtt` binary, confirmed to reproduce the
canonical ZL corpus's exact recorded checksum
(`f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`)
byte-for-byte when run against `data/ZL3b-n.txt`, and used to derive
`data_work/IT2a-x7.canonical.txt` for Gate A. This section is a factual
record of that process incident, not a change to any metric, null,
threshold, seed or bin defined elsewhere in this document.
