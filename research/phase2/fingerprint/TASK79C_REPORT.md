# Task79c report — Fingerprint v2 freeze closure

## Result

**`FINGERPRINT_V2_FROZEN`.** Every mandatory gate in the parent task's
formal freeze checklist (§38, restated in `TASK79C_DESIGN.md` §11) resolved
with a completed, honestly reported evaluation. No gate resolved
`DATA_UNAVAILABLE`, `INSUFFICIENT_ALIGNMENT`, `UNSTABLE`, or `NOT_TESTABLE`
at a level that the parent task's own failure semantics (§39) requires to
force `FINGERPRINT_V2_NOT_READY`. One gate (predictive hierarchy,
HR3/HR5) resolved `INCONCLUSIVE` rather than positively — the parent task
explicitly allows this ("Freeze требует закрытого вопроса, а не
обязательно положительного результата", §29; "hierarchical model does not
beat flat" is listed among the fully valid `IMPORTANT NEGATIVE OUTCOMES`,
§40) provided the estimator is sound and F2's interpretation is updated
accordingly, which this report does (§12 below).

`FINGERPRINT_V2_FREEZE_MANIFEST.json` and `FINGERPRINT_V2_FROZEN` are
created alongside this report.

## Process note (read before the rest of this report)

Task79c's implementation was split between this session and a forked
sub-agent. Partway through, that sub-agent incorrectly treated the
session's own legitimate coordination messages as an external attack and,
on that mistaken belief, deleted several files the session had produced
(the historical-notation control download, its extraction tool, and three
provenance/handoff documents) and wrote an inaccurate incident narrative.
No external or unauthenticated party was ever involved. The session
restored every deleted file byte-for-byte (re-downloading the same source
at the same pinned commit; checksums matched exactly) and corrected the
design document. Full factual account:
`TASK79C_DESIGN.md` §14. This is recorded here because it is part of this
task's honest provenance record, not because it changed any metric, null,
threshold, or result.

## 1. Was an independent second transcription obtained?

Yes. `data/IT2a-n.txt` (Takeshi Takahashi's transliteration via Jorge
Stolfi's 1999 interlinear file, distributed by voynich.nu in IVTFF 2.0),
independent of Zandbergen–Landini in transcriber, source lineage and
alphabet variant. Prepared via `ivtt -x7` (audited binary, source SHA-256
matching `DATA.md`) + `codex_prepare` (identical flags to the canonical
ZL preparation) into `data_work/IT2a-x7.canonical.txt` (37,943 tokens).
Full provenance: `TASK79C_DESIGN.md` §1.

## 2. What is the alignment coverage?

225/227 ZL folios matched (99.1%; both ZL-only folios, `f116v` and `fRos`,
are manuscript regions with independently documented transcription
difficulty). 5,214/5,385 ZL loci matched (96.8%); 36,537 token positions
aligned positionally within matched loci; 1,670 token positions within
1,320 loci with differing segmentation are recorded unmatched, not
interpolated. Full detail: `TRANSCRIPTION_ALIGNMENT.tsv`,
`TRANSCRIPTION_ALIGNMENT_REPORT.md`.

## 3. Which CORE metrics are stable between transcriptions?

All 13. Classification (thresholds fixed before comparison,
`TASK79C_DESIGN.md` §2): `STABLE` — `2DL1_LAYOUT_POSITION_MI`,
`LC2_LABEL_TEXT_NMI`, `PF2_FOLIO_COHERENCE` (3/13).
`DIRECTION_STABLE` — the remaining 10/13. `VERDICT_STABLE`: 0.
`UNSTABLE`: 0. `NOT_TESTABLE`: 0. Full table: `TRANSCRIPTION_STABILITY.tsv`,
`TRANSCRIPTION_STABILITY_REPORT.md`.

## 4. Which are transcription-sensitive?

None flips status or effect sign, but 10/13 shift by more than one
canonical-run permutation-null SD. `LS3_BOUNDARY_LENGTH_ASYMMETRY` (5.08
SD) and `HR1_SECTION_VARIANCE_SHARE` (9.46 SD) are the most sensitive —
both plausibly explained by the two ZL-only folios' removal and by
bracketed-uncertain-reading notation differences at line boundaries (see
`TRANSCRIPTION_STABILITY_REPORT.md` for the metric-by-metric account). No
CORE metric's qualitative conclusion depends on which of the two
transcriptions is used.

## 5. Was a historical shorthand/abbreviation control obtained?

Yes. Burchard of Worms' *Decretum*, the "Burchards Dekret Digital" TEI-XML
edition (Akademie der Wissenschaften und der Literatur Mainz), witness
`koeln-edd-c-119`, chapters 6/7/11/12/13 — the smallest of the 5/126 BDD
witnesses with any real transcription, chosen by a content-blind
mechanical tie-break before any manuscript text was read. License: CC BY
4.0, stated in-file. 19,052 tokens after extraction (real `<abbr>`-branch
scribal abbreviations only) and normalization. This is a real instance of
Task79b's own top-ranked (`HIGH`-suitability) candidate,
`ABBREVIATIONES_BURCHARDS`. Full provenance: `CONTROL_PROVENANCE.tsv`,
`TASK79C_DESIGN.md` §1/§3.

## 6. Was a table/procedural control obtained?

Yes. `data_test/msdos2.0.txt` (MS-DOS 2.0 x86 assembly source), already
resident in this repository from an unrelated Phase 1 experiment and
confirmed independent of the Fontana branch. 112,155 usable tokens after a
mechanical glyph-nonempty filter (7 symbol-only tokens dropped, e.g. a
bare `_`, out of 112,162 — necessary for `internal/fingerprintv2`'s loader,
not a scientific choice; `TASK79C_DESIGN.md` §5a).

## 7. How does F2 behave under corpus-size matching?

Five deterministic, mostly non-overlapping, line-preserving matched
replicates of `msdos2.0.txt` were drawn at the Voynich primary corpus's
token count (39,380±~0.01%; replicate 4 was explicitly exhausted at
21,769/39,380 tokens because its window ran past the file's end — recorded,
not hidden). Family-balanced standardized distance from Voynich across the
5 replicates ranged `0.814`–`0.882` (full corpus: `0.774`) — real,
non-trivial between-replicate dispersion (§16), confirming a single
truncated sample would have been an inadequate basis for this comparison.
BDD (19,052 tokens, smaller than the Voynich target) was compared at its
full extent rather than subsampled, per `TASK79C_DESIGN.md` §5's explicit
handling of an undersized control.

## 8. How does F2 behave under vocabulary conditioning?

Type/token ratios differ sharply and in an interpretable direction:
Voynich `0.209` (8,243/39,380); BDD `0.423` (8,062/19,039, inflected
Latin); msdos `0.064`–`0.092` across its full corpus and replicates
(repetitive, small procedural vocabulary — mnemonics, register/label
names reused constantly). `LP1`/`EF1`-family metrics, whose estimators are
directly sensitive to vocabulary-to-token ratio, are flagged accordingly
in this comparison rather than re-normalized post hoc, per
`TASK79C_DESIGN.md` §5 item 3.

## 9. Does the family-balanced distance work practically?

Yes. Reusing the unmodified pipeline-computed per-metric standardized
differences (`discriminative_validation.json`) and aggregating to
family-level and family-balanced distance (`FINGERPRINT_V2_DISTANCE.md`
§2.1-2.2) via `cmd/task79c-distance-pareto` produced an ordered, stable,
interpretable separation across all three independently known corpus
classes:

| Control | Family-balanced distance | Common-core distance |
|---|---:|---:|
| BDD (shorthand/abbreviation) | `0.426` | `0.503` |
| Doyle (natural prose) | `0.602` | `0.721` |
| msdos (procedural, full corpus) | `0.774` | `0.923` |
| msdos (5 replicates) | `0.814`–`0.882` | `0.933`–`1.045` |

BDD is closest to Voynich on `lexical paradigm` (0.296), `edit family`
(0.749) and `2D-LITE` (0.385); msdos is closest on `line` (0.453–0.585 vs.
BDD's 0.701) — a genuine family-level trade-off, not a single-axis
ranking, exactly what a family-balanced (not flattened) distance is
supposed to surface. Full per-family table:
`experiments/fingerprint-v2-task79c-v1/distance-pareto-out/
full_portfolio_distance_pareto.json`.

## 10. Does the Pareto interface work practically?

Yes, using the unmodified, frozen Task66 dominance rule
(`internal/mechanismspace.ParetoFront`/`Dominates`), applied to per-family
closeness scores. BDD, Doyle and the msdos full corpus are all
non-dominated (each is closer than the others on at least one family).
Three of the five msdos replicates (`rep0`, `rep3`, `rep4`) are dominated
by other msdos replicates/the full corpus — expected, honest sampling
noise among repeated draws of the *same* underlying corpus, illustrating
exactly why §16's multiple-replicate requirement matters (a single
replicate could have looked artificially better or worse than the
corpus's real position). No classifier was fit; only dominance
relationships among the held-out controls themselves are reported, per
the parent task's explicit scope limit (§19).

## 11. What did the corrected PF4 leaf-paired null show?

`SUPPORTED`. Observed same-leaf recto/verso coherence `0.5192`
(identical, to reported precision, to Task79's own value) vs. a
leaf-paired-bijection-permutation null mean `0.362` (SD `0.0184`), effect
size `8.54` SD, empirical p `≈0.001` (0/1000 permutations ≥ observed), 92
usable paired leaves, 0 unpaired. This reverses Task79's `INCONCLUSIVE`
verdict, which used a broader, less targeted null (`HN4`, folio
reassignment within Currier/section) that Task79's own report already
flagged as inadequate for this specific physical-pairing question.
Interpretation constraints: `PF4_LEAF_NULL.md`. This does **not** establish
an acrostic, cipher key, or any semantic cross-page relationship — only
that same-physical-leaf recto/verso pages are more profile-similar to each
other than a random recto/verso pairing, consistent with (not diagnostic
of) ordinary scribal practice.

## 12. Does hierarchy outperform flat out of sample?

`INCONCLUSIVE`. Folio-block 5-fold cross-validation (all 5 folds
testable): `HR3` (folio-only partial pooling) has mean held-out-NLL delta
`≈0` (`-3.9e-5`) versus flat — this is the *expected*, structurally
necessary result under strict folio-block holdout (a wholly unseen folio
carries no folio-specific training signal for a single-level model; see
`TASK79C_DESIGN.md` §9's pre-registered explanation) and confirms the
cross-validation harness is not leaking folio identity, not evidence
against hierarchy. `HR5` (folio nested in section, which *can* carry
cross-fold information via sibling folios in the same section) has mean
delta `+0.0169` (worse than flat on average; 2/5 folds better, 3/5 worse)
— a genuine, non-cherry-picked null result: section membership does not
reliably predict a held-out folio's line-length distribution better than
the corpus grand mean, at least for this target quantity and this
partial-pooling estimator. Raw fold-level results:
`experiments/fingerprint-v2-task79c-v1/pf4-hr-out/
pf4_hierarchy_result.json`. Per the parent task §29, this negative-leaning
result does not block freeze; F2's page/hierarchy interpretation is
updated to: variance-share decomposition (`HR1`, still `SUPPORTED`,
`CORE`) establishes that folio and section identity explain a real share
of line-length variance, but that decomposition does not, by itself,
predict an unseen folio/section better than the corpus average — a
narrower and more honest claim than Task79's prior `INCONCLUSIVE`
non-answer.

## 13. Which metrics remain CORE?

All 13 unchanged from Task79: `2DL1_LAYOUT_POSITION_MI`,
`BP1_BOUNDARY_TOKEN_NMI`, `EF1_GIANT_COMPONENT_SHARE`,
`EF2_GLOBAL_CLUSTERING`, `EF3_DEGREE_FREQUENCY_SPEARMAN`,
`HR1_FOLIO_VARIANCE_SHARE`, `HR1_SECTION_VARIANCE_SHARE`,
`LC1_LOCUS_TYPE_NMI`, `LC2_LABEL_TEXT_NMI`, `LS2_POSITIONAL_LEXICON_NMI`,
`LS3_BOUNDARY_LENGTH_ASYMMETRY`, `PF2_FOLIO_COHERENCE`,
`PF5_WITHIN_FOLIO_PROGRESSION`. None was demoted: none is `UNSTABLE` across
transcriptions, and Task79c added no new CORE candidate (§4/§35 of the
parent task; confirmed by `F2_ADMISSION.tsv` having zero `ADMIT_TO_F2`
rows). Full registry: `F2_METRIC_REGISTRY_FINAL.tsv`.

## 14. Which are SUPPORTING?

The other 20 metrics from Task79's registry, unchanged:
`EF1_ISOLATE_SHARE`, `HR1_LOCUS_VARIANCE_SHARE`, `HR6_CURRIER_SECTION_NMI`,
`LC5_IVTFF_I_NMI`, `LC5_IVTFF_X_NMI`, `LP1_RULE_SUPPORT_GINI`,
`LP4_PREFIX_ATTACHMENT_NMI`, `LP4_SUFFIX_ATTACHMENT_NMI`,
`LS1_LINE_LENGTH_CV`, `LS4_WITHIN_LINE_EXACT_REPETITION`,
`PF3_ADJACENT_FOLIO_CONTINUITY`, `PF4_RECTO_VERSO_COHERENCE` (its own null
was corrected and re-validated, §11 above, but it stays `SUPPORTING` — the
parent task's rules for promoting a metric to `CORE` were not re-run here
and PF4's re-validation is not itself grounds to promote it), and the 8
`cs*` cross-scale metrics. None of these was examined for
cross-transcription stability by Task79c (out of scope per parent §8,
which specifies CORE metrics).

## 15. Which are quarantined for v2.1?

Four, all from Task79b, none newly identified by Task79c (the parent task
prohibits searching for new Voynich-specific diagnostics, §"Запрещено"):
positional channel NMI, boundary class/Zipf diagnostic (both
`DEFER_TO_V2_1`), abbreviation length reduction, and expansion ambiguity
(both `EXPLORATORY_ONLY` — Voynich has no plaintext/expansion alignment to
compute them against, though the newly acquired BDD control does carry
real `<abbr>/<expan>` pairs a future v2.1 review could use as a
control-only diagnostic). Full list and binding constraints on future
admission: `F2_V2_1_CANDIDATES.md`.

## 16. Is the redundancy problem closed?

Yes, by construction: Task79c added zero new primary metrics, so there is
no new dimension to check for redundancy against the existing family
structure. The redundancy matrix recomputed in this task's own controls-
portfolio run (`experiments/fingerprint-v2-task79c-v1/
controls-portfolio-out/redundancy_matrix.json`) is byte-identical to
Task79's canonical `redundancy_matrix.json` — confirming deterministic
reproduction on the unchanged ZL corpus and no drift. The known redundancy
(line length ↔ token entropy, `r=0.845`; line length ↔ transition entropy,
similarly large) stands exactly as Task79 recorded it: `token_entropy` and
`transition_entropy` remain `SUPPORTING`, not independently weighted CORE
evidence, consistent with `FINGERPRINT_V2_DISTANCE.md` §3's
redundancy-adjusted-weight rule.

## 17. Is stability closed?

Yes. `F2_STABILITY_FINAL.tsv` combines, for every one of the 33 registered
metrics: folio-half stability (unchanged from Task79, all `ROBUST` or
`ROBUST_WITH_LIMITATIONS` for CORE metrics), parameter sensitivity
(unchanged from Task79), and — newly, for the 13 CORE metrics —
transcription stability (§3-4 above). No CORE metric lacks an explicit
final stability status.

## 18. Are alternative-explanation controls sufficient?

`ACCEPTABLE` — upgraded from Task79's `NOT_SUPPORTED`. The parent task's
minimum held-out portfolio (§14: `NATURAL_PROSE`, `SHORTHAND_OR_
ABBREVIATION`, `TABLE_OR_PROCEDURAL`) is now fully populated (Doyle, BDD,
msdos respectively) and the distance/Pareto exercise (§9-10 above)
demonstrates the comparison machinery actually discriminates between these
three independently known classes with an interpretable, non-degenerate
family-level trade-off structure — which is the specific, narrower
property the parent task requires (§21: "comparison machinery produces
stable, interpretable differences between independently known corpus
classes"), not that F2 identifies Voynich (which Task79c explicitly does
not attempt and does not claim).

## 19. Can F2 be used as a frozen criterion for Task81–83?

Yes, with the limitations recorded throughout this report and in
`FINGERPRINT_V2_FREEZE_MANIFEST.json` treated as binding: the CORE/
SUPPORTING split, the LC3/LC4 line==locus limitation (preserved, not
"fixed" — no data source here provides an independent locus level), the
`INCONCLUSIVE` HR3/HR5 predictive-hierarchy verdict (interpreted narrowly,
§12), and the single-witness/single-book-subset scope of the BDD control
(a future Task82b needing broader shorthand coverage should read
`TASK82B_HANDOFF.md` before re-deriving anything).

## 20. Final freeze verdict

**`FINGERPRINT_V2_FROZEN`.**

| Component verdict | Value |
|---|---|
| `CROSS_TRANSCRIPTION_STABILITY` | `SUPPORTED` (13/13 CORE metrics STABLE or DIRECTION_STABLE; 0 UNSTABLE) |
| `HISTORICAL_NOTATION_CONTROL_READY` | `SUPPORTED` |
| `PROCEDURAL_CONTROL_READY` | `SUPPORTED` |
| `DISTANCE_INTERFACE_VALIDATED` | `SUPPORTED` |
| `PARETO_INTERFACE_VALIDATED` | `SUPPORTED` |
| `PF4_LEAF_TEST_COMPLETED` | `SUPPORTED` (test completed; substantive result also `SUPPORTED`) |
| `HIERARCHICAL_PREDICTION_VALIDATED` | `INCONCLUSIVE` (test completed and sound; does not block freeze, parent §29) |
| `CORE_STABILITY_ACCEPTABLE` | `SUPPORTED` |
| `CORE_REDUNDANCY_ACCEPTABLE` | `SUPPORTED` |
| `ALTERNATIVE_EXPLANATION_COVERAGE_ACCEPTABLE` | `ACCEPTABLE` |

## Fontana and future-task firewall statement

No file under `research/phase2/fontana/`, no `F01/F08/F11/F12`/`F07/F10`
model output, and no `Task76`/`Task78`/`Task80` artifact was read, cited,
or used by any Task79c code, configuration, or document. Task81, Task82,
Task82b and Task83 were not run and their outputs (none exist yet) were
not consulted.

## Reproducibility

- Repository commit at design freeze: `d568e54bc7a57c87b4ab0096fb5e89550c9b9c09`.
- `TASK79C_DESIGN.md` SHA-256 at final freeze:
  `af3715be5718fdd03716e7cf6bc6b1ab64ab0a4bb917b24ef79dbf67b2b2e484`
  (see `TASK79C_DESIGN_FROZEN` for the full version history of this
  document, including the process-note addenda in §12/§14 above).
- Corpus checksums: canonical ZL `bf5b6d4ac1e3a51b1847a9c388318d609020441c
  cd56984c901c32b09beccafc`; alternate IT `7f27a8b0feed8f6de0a99900df6bf9
  12dd1d295c38e5f830bac8b41c3f536fb5`; full checksum table for every
  corpus/control/replicate: `CONTROL_PROVENANCE.tsv`,
  `TASK79C_DESIGN.md`.
- Seeds: `20260824` (metric permutations/bootstrap, PF4 null, matched
  replicate seed base), `40260824` (HR3/HR5 folio-fold assignment).
- Deterministic outputs verified byte-stable on re-run where re-run: the
  controls-portfolio run's `redundancy_matrix.json` reproduced Task79's
  canonical `redundancy_matrix.json` byte-for-byte; ZL3b's own `-x7`
  derivation via the newly added `cmd/ivtff-x7-extract` reproduces the
  canonical corpus's exact recorded checksum.
- Stochastic outputs (permutation/bootstrap draws) used the fixed seeds
  above; raw draws are retained (`pf4_hierarchy_result.json`'s
  `null_draws`, and every `raw_results.json`/`*-out/` directory under
  `experiments/fingerprint-v2-task79c-v1/`).
- Validation suite: `go build ./...`, `go vet ./...`, `go test ./...`
  (61/61 packages pass), `go test -race ./...`, `git diff --check` — all
  green (see this task's final validation run).

## Machine-readable artifact map

- `experiments/fingerprint-v2-task79c-v1/transcription-it.yaml` +
  `transcription-it-out/` — Gate A/C alternate-transcription run (full
  Task79 pipeline output shape, matching `experiments/
  fingerprint-v2-task79-v1/canonical-out/` field-for-field).
- `experiments/fingerprint-v2-task79c-v1/pf4-hr-out/
  pf4_hierarchy_result.json` — Gate E/F PF4 leaf null (with all 1,000 raw
  permutation draws) and HR3/HR5 fold-level results.
- `experiments/fingerprint-v2-task79c-v1/controls-portfolio.yaml` +
  `controls-portfolio-out/` — Gate D/G raw run against msdos (full + 5
  replicates) and BDD (full), including `discriminative_validation.json`
  (147 per-metric contrasts). Three files in this directory
  (`raw_results.json`, `fingerprint.json`, `fingerprint_v2_candidate.json`,
  106–112MB each) exceed GitHub's 100MB per-file push limit and are
  gitignored locally rather than committed; every derived summary used in
  this report is committed and unaffected, and the three files are
  byte-reproducible by re-running the recorded config/seed/commit.
- `experiments/fingerprint-v2-task79c-v1/distance-pareto-out/
  full_portfolio_distance_pareto.json` — Gate D/H family-level distances
  and Pareto dominance across all 3 control classes + msdos replicates.
- `data_test/bdd-tei/`, `data_test/bdd-prepared/`,
  `data_test/matched-replicates/` — control corpora and their preparation
  sidecars (gitignored bytes; checksums in `CONTROL_PROVENANCE.tsv`).
