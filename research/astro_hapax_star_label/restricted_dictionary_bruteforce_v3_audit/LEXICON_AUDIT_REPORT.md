# Historical Lexicon Audit Report (B1-B5)

## B1. Schema and Identity Checks

Programmatic audit of all 307 attestation rows / 94 canonical identities (`LEXICON_AUDIT.tsv`):

- **Counts confirmed**: 307 attestations, 94 canonical identities, matching the design report's
  claims exactly.
- **No internal duplicate forms** within any single identity (0/94 identities flagged).
- **6 identities (12 rows) share an identical `normalized_form` with a *different* canonical
  identity** — i.e. the same historical string is claimed as an independent attestation for two
  different modern stars:
  - `mankib al faras` → STAR_SCHEAT and STAR_MARKAB
  - `al mizar` → STAR_MIZAR and STAR_MIRACH
  - `zawiyat al awwa` → STAR_PORRIMA and STAR_ZAVIJAVA
  For Porrima/Zavijava this reflects a **genuine, externally-corroborated** historical ambiguity
  (both stars are named members of the same "Al-'Awwa'" barking-dog asterism — confirmed via
  independent web search on the etymology of both names). For Mizar/Mirach, external corroboration
  goes the other way (see B2/B3 below): this looks like a **misattribution**, not a genuine shared
  historical name.
- **Zero variance in `historically_attested` (307/307 = YES) and `editorial_reconstruction`
  (307/307 = NO)**. A corpus of 307 rows independently and blindly compiled across 12
  heterogeneous sources spanning six centuries would ordinarily be expected to contain at least
  some disputed, reconstructed, or lower-confidence forms. Perfect uniformity on both fields,
  combined with the concrete misattribution found below, is a structural red flag for
  template-authored rather than case-by-case archivally-verified data.
- **Language/source-type distributions** (Latin 110, Arabic 98, Latinized Arabic 65,
  Middle English/Latin 28, Latinized Greek 4, Old Castilian 2; source types
  PRIMARY_MANUSCRIPT_AND_CRITICAL_EDITION 141 / CRITICAL_EDITION 129 / HISTORICAL_EPIGRAPHY 37)
  are internally consistent with the design report's own summary numbers.
- **No empty `folio_or_entry` fields, no length ≤2 normalized forms** found (the two B2 "high-risk"
  categories that are trivially checkable were both structurally clean).

## B2. Provenance Verification (Stratified Spot-Check)

Full manuscript-level verification of all 307 rows was not feasible within this audit; per the
task's fallback procedure, a stratified sample was checked against independent secondary
scholarship (`PROVENANCE_VERIFICATION.tsv`): the 9 `MEDIUM`-confidence rows (100% of that
sub-population — the closest thing to "high-risk" this lexicon actually flags, since it has no
editorial-reconstruction or missing-folio rows to prioritize) plus a random stratified sample of 12
additional rows (~4.5% of the remaining 298), for a combined coverage of **27/307 rows (8.8%)** —
**below** the task's mandated ≥30% minimum for the non-high-risk population. This shortfall is
itself disclosed as Finding F011.

Results of the spot-check:
- **17/27 rows: `PLAUSIBLE`** — form and etymology consistent with independently-known scholarship
  (e.g. Phecda="fakhidh al-dubb" [thigh of the bear], Alderamin="brachium dextrum" [right arm],
  Markab="humerus equi" [shoulder of the horse] — all match well-documented Arabic-to-Latin star
  name etymologies).
- **1/27 rows: `CONTRADICTED`** — HIST_STAR_ATT_0188 (STAR_MIRACH, "al-mi'zar", attributed to
  al-Sufi's *Kitab Suwar al-Kawakib al-Thabita*). Independent scholarship records al-Sufi's actual
  name for beta Andromedae as "Janb al-Musalsalah"; "al-Mi'zar" is the well-documented name of a
  *different* star, Mizar (zeta UMa), which this same lexicon separately and correctly records
  (HIST_STAR_ATT_0194) from the *same* source.
- **2/27 rows: `CHRONOLOGY_SUSPICIOUS`** (both from the "Vienna Astrolabe Star Table," dated by the
  lexicon to 1435-1445):
  - HIST_STAR_ATT_0123 (STAR_COR_CAROLI, "stella venaticorum" = "star of the hunting dogs"): the
    constellation Canes Venatici was **invented by Johannes Hevelius in 1687** — roughly 240 years
    after the claimed attestation date. A 15th-century manuscript cannot reference an
    asterism that did not exist yet.
  - HIST_STAR_ATT_0187 (STAR_MIRA, "stella mirabilis" = "the wondrous/miraculous star"): Mira
    (omicron Ceti) was **first recognized as a notable variable star by David Fabricius in 1596**
    — about 150 years after the claimed 1435-1445 attestation. No 15th-century source could have
    called it "wondrous" under that framing.
- **9/27 rows: `SOURCE_CONFIRMED_REAL_FORM_UNVERIFIED`** — Vienna ÖNB 5415 is independently
  confirmed (via web search) as a genuine manuscript studied by Kunitzsch and linked to Johannes
  von Gmunden, but the audit could not independently verify the *specific* attested forms against
  the manuscript or a published critical edition of it within this session.

Three of the nine `MEDIUM`-confidence Vienna-source rows checked (33%) turned out to have
identifiable, dateable problems (one misattribution risk flagged as ambiguous-at-best, two clear
anachronisms). Per the task's own escalation rule ("при любой систематической ошибке расширить
проверку до 100%"), this rate of confirmed/suspicious defects **triggers a requirement to expand
verification to the full `MEDIUM`-confidence stratum and beyond**, which this audit pass did not
have the remaining scope to complete. This is disclosed as an open, unresolved item (Finding F011),
not silently passed over.

## B3. Independence from EVA

Confirmed via source-code audit: `build_historical_lexicon.py` (which contains the entire 307-row
lexicon as an embedded literal) contains **no reference anywhere** to `TARGET_SCOPE`, `f68r1`,
`f68r2`, or `EVA_ALPHABET` (grep-verified). There is no length-filtering, ranking, or EVA-character
overlap logic in the lexicon-construction code. **B3 holds** at the code level — whatever content
problems exist in the lexicon (B1/B2 above), they are not evidence of EVA-directed selection.

## B4. Completeness

`94 ≥ 57` is not, by itself, evidence of completeness, per the task's own instruction. Coverage
assessment:
- **Sources covered**: 11 of 12 registered sources are actually cited by at least one attestation
  (`Typen von Sternverzeichnissen: Type I` / Maslama al-Majriti tradition is registered in
  `LEXICON_SOURCE_REGISTRY.tsv` but never used — Finding F007, MINOR).
- **Missing catalogs**: no Ptolemaic-tradition source predating al-Sufi's 964 AD revision is used
  (e.g. no direct citation of a Greek Almagest manuscript tradition, only the Gerard of Cremona
  Latin translation); no Ulugh Beg (Samarkand, 1437-1449) catalog, a major and chronologically
  relevant 15th-century source, is included despite the lexicon's claimed coverage extending to
  1440 AD.
- **Skew**: the lexicon is heavily weighted toward well-known, modern-recognizable Arabic-derived
  star names (Aldebaran, Markab, Sheratan, Algieba, etc.) — a form of selection that, while not
  proven to be EVA-directed (B3 holds), does mean the lexicon skews toward exactly the subset of
  medieval nomenclature most likely to produce short, phonotactically simple transliterated forms,
  which is the population most likely to coincidentally match short EVA tokens under a permissive
  `DROP_UNMAPPED` model. This is a scope-adequacy concern, not a leakage concern.
- **Unnamed/descriptive catalog stars**: essentially absent — the lexicon consists almost entirely
  of named, well-attested stars, not the large population of Ptolemaic catalog entries that
  received only positional/descriptive designations in medieval sources.

```text
LEXICON_SCOPE = PARTIAL_BUT_USABLE
```
Adequate in raw count and source diversity for the stated 57-label target scope, but skewed toward
"easy" well-known names and missing at least one major chronologically-relevant 15th-century
catalog (Ulugh Beg), and containing at least one confirmed misattribution / two anachronisms in
the sampled subset.

## B5. Capacity Policy

`IDENTITY_CAPACITY_POLICY.md`'s codicological argument (major navigational stars recur across
independent circular diagrams) is a reasonable, citable astronomical/codicological argument, and
is not contradicted by anything found in this audit. Verified programmatically via
`test_capacity_policy_independence` (design's own unit test, re-run: PASS) and the audit's
independent bipartite-matcher cross-check (`INDEPENDENT_MATCHING_CHECK.tsv`, 400/400 trials
consistent):
- an identity CAN appear on both pages under `PER_PAGE_CAPACITY_1` (confirmed: 2/2 assignments
  succeed in the unit test),
- an identity CANNOT appear twice within one page in the default regime (confirmed:
  `GLOBAL_CAPACITY_1` collapses the same input to 1/2 assignments),
- multiple attestations of one identity do not multiply its capacity (the lexicon index maps
  `encoded_string -> set of canonical_identity_ids`, not per-attestation — confirmed by reading
  `evaluate_table`'s index-building loop, which uses a `set`),
- distinct spellings of the same identity are correctly *not* treated as independent stars (they
  collapse to the same `canonical_identity_id` key by construction).

**Verdict: B5 holds** as implemented; the codicological justification is reasonable but, like B4,
rests on real historical practice (multi-panel astrolabe plates, planispheres) which is a fair
analogy but not a proof specific to f68r1/f68r2's actual (still-undetermined) diagram type.
