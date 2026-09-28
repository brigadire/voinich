# Lexicon Remediation Protocol (Gate L-R)

Pre-registered BEFORE any verification result is looked at, based only on: (a) the frozen v3
lexicon's structural counts (confidence/source/language/identity distribution) and (b) the frozen
v3 audit's already-published findings (F005, F006, F010, F011). No row's verification outcome
influenced anything written in this file.

## 0. Corpus facts used to design this protocol (structural only, not outcome data)

- 307 attestation rows, 94 canonical identities (2-5 rows each, median 3).
- Confidence: 298 HIGH, 9 MEDIUM. All 9 MEDIUM rows come from one source: "Vienna Astrolabe Star
  Table (MS Vienna ONB 5415)", dated 1435-1445.
- That same Vienna source contributes 12 rows total (9 MEDIUM + 3 HIGH: STAR_ANTARES, STAR_POLARIS,
  STAR_VINDEMIATRIX).
- 12 registered sources; largest strata: al-Sufi's *Kitab Suwar al-Kawakib al-Thabita* (72 rows),
  Gerard of Cremona's *Almagestum* (86 rows), *Libros del Saber de Astronomia* (57 rows), Chaucer's
  *Treatise on the Astrolabe* (28 rows).
- Languages: ARABIC (98), LATIN (110), LATINIZED_ARABIC (65), LATINIZED_GREEK (4),
  MIDDLE_ENGLISH_AND_LATIN (28), OLD_CASTILIAN (2).
- A full scan of `normalized_form` across the whole file (a mechanical, zero-judgment check —
  every row's normalized form against every other row's, not a sample) found **3** cases where the
  identical normalized form is attached to two different `canonical_identity_id`s:
  1. `al mizar` -> STAR_MIRACH (`HIST_STAR_ATT_0188`) / STAR_MIZAR (`HIST_STAR_ATT_0194`) — already
     flagged and confirmed misattributed by the v3 audit (F005).
  2. `mankib al faras` -> STAR_MARKAB (`HIST_STAR_ATT_0164`) / STAR_SCHEAT (`HIST_STAR_ATT_0250`) —
     new, not previously flagged.
  3. `zawiyat al awwa` -> STAR_PORRIMA (`HIST_STAR_ATT_0213`) / STAR_ZAVIJAVA (`HIST_STAR_ATT_0296`)
     — new, not previously flagged.

  This mechanical check costs nothing (no external lookup) and is therefore run at 100% coverage
  regardless of sampling tier — it is not "sampling," it is exhaustive string comparison.

## 1. Verification-category definitions (fixed before any row is checked)

- `VERIFIED_EXACT` — the specific claimed form, at the specific claimed source, date, and
  entry/folio, has been directly confirmed (e.g., against a digitized manuscript image, a named
  page/entry in a critical edition, or a secondary source that itself cites that exact
  source+folio+form combination).
- `VERIFIED_NORMALIZED` — the specific claimed form-at-source-at-date link is confirmed by
  independent secondary scholarship (e.g., Kunitzsch's published star-name typology, which is
  cited as the underlying source for most of this lexicon's Latinized-Arabic and Latin-gloss rows),
  but not against a primary manuscript image or an exact folio citation.
- `PLAUSIBLE_NOT_DIRECTLY_VERIFIED` — the form's etymology/meaning is well-documented in modern
  star-name scholarship and is *consistent* with the claimed source and period, but no independent
  source was found that specifically ties this form to this claimed source at this date. Per the
  task's explicit instruction, a form is never promoted out of this category on etymology alone —
  etymological plausibility is necessary, not sufficient, for `VERIFIED_NORMALIZED`.
- `CONTRADICTED` — independent scholarship gives a different, incompatible attested form/source/date
  for this specific star, or explicitly attests this form to a *different* star than claimed.
- `ANACHRONISTIC` — the claimed date range precedes the historically documented origin of the name,
  the referenced constellation, or the star's recognition, by an amount too large to be editorial
  slack (the working threshold used throughout this protocol is >50 years, matching the two cases
  already on record: ~150 years for Mira/Fabricius, ~250 years for Canes Venatici/Hevelius).
- `SOURCE_UNAVAILABLE` — the claimed source/manuscript could not be found or confirmed to exist at
  all via search, so the claim is neither confirmable nor falsifiable.
- `EDITORIAL_RECONSTRUCTION` — the row itself, or independent scholarship about its source, marks
  the form as a modern scholarly reconstruction/normalization rather than a form actually written in
  the historical witness.

Ceiling acknowledged up front: this remediation has web search access to secondary scholarship and
digitized-catalog metadata, but not to page-by-page manuscript images for most of these sources.
`VERIFIED_EXACT` is therefore expected to be rare; most positive verifications will land at
`VERIFIED_NORMALIZED`. This is stated here, before any row is checked, specifically so a low
`VERIFIED_EXACT` count post-hoc cannot be read as a new failure — the honest ceiling is disclosed
now.

## 2. Verification tiers (fixed before any row is checked)

**TIER_MANDATORY (100% coverage required):**
1. All 9 MEDIUM-confidence rows.
2. All 12 Vienna-source rows (superset of #1, since 3 Vienna rows are HIGH confidence) — the v3
   audit already found a 22% (2/9) confirmed-defect rate in this source's MEDIUM stratum, which
   under the task's own >5%-error escalation rule mandates escalating the *entire* Vienna stratum,
   not just its MEDIUM-confidence subset.
3. All 6 rows involved in an exact `normalized_form` collision across identities (3 pairs, listed
   in §0).

Union of 1-3 = 18 distinct rows (12 Vienna ∪ 6 collision rows, no overlap between the two sets).

**Exhaustive mechanical check (100%, not a sample, no external lookup needed):** every row's
`normalized_form` against every other row's, across all 307 rows, for exact collisions (already run
in §0) — repeated once more after any `CORRECT`/`MERGE_IDENTITY` edits, to confirm no new collision
was introduced by a fix.

**TIER_ALSUFI_SCAN (100% of al-Sufi's 72 rows, near-duplicate check only, no external lookup):**
because al-Sufi's source already produced one of the three exact collisions (Mirach/Mizar) despite
being the second-largest stratum, every al-Sufi row's `normalized_form` is additionally checked for
*near*-duplication (shared 4+ character substrings / edit distance <=2) against every other al-Sufi
row attached to a different identity, to catch transliteration-variant collisions an exact-string
match would miss. This is mechanical and costs no search budget.

**TIER_STRATIFIED_SAMPLE (>=30% of the 289 non-mandatory rows = >=87 rows):** sampled by
identity-cluster (verifying all attestation rows for a sampled identity together, since they share
one well-documented modern star and checking its etymology/history typically resolves multiple rows
in one search) rather than by independent row draws, stratified so the sample's marginal
distribution over source_title, source-century bucket (9th/10th c., 12th c., 13th c., 1252-1270,
1391, 1435-1445), language, and source_type roughly matches the full corpus's marginal
distribution, with one deliberate deviation: **al-Sufi rows are over-sampled to >=50% coverage**
(not just the 30% floor), because al-Sufi is the only source that has already produced one confirmed
exact-form collision among its 72 rows.

## 3. Escalation rule (fixed before any row is checked)

If the checked-row error rate (`CONTRADICTED` + `ANACHRONISTIC` verdicts, over rows actually
checked) in TIER_STRATIFIED_SAMPLE exceeds 5% for any single stratum (source, century-bucket, or
language), OR a second instance of the same defect *pattern* already seen in Vienna/al-Sufi is
found elsewhere, that stratum is escalated to 100% coverage and the escalation and its trigger are
recorded in `LEXICON_REMEDIATION_REPORT.md`.

## 4. Disposition rule (fixed before any row is checked)

Every one of the 307 original rows receives exactly one disposition. Checked rows get a disposition
from their verification result (`CONTRADICTED`/`ANACHRONISTIC` -> `REMOVE_MISATTRIBUTED` /
`REMOVE_ANACHRONISTIC` unless a specific in-scope correction resolves it -> `CORRECT` or
`MERGE_IDENTITY`; `SOURCE_UNAVAILABLE` with no corroboration elsewhere -> `REMOVE_UNVERIFIED`;
`EDITORIAL_RECONSTRUCTION` -> `RECLASSIFY_RECONSTRUCTION`; everything else checked and clean ->
`KEEP`). Unchecked rows default to `KEEP` — the task explicitly forbids blanket removal of rows that
were never examined; sampling is a coverage/confidence statement, not a verdict on the unsampled
rows.

## 5. What this protocol does not attempt

No attempt is made to physically inspect manuscript folios, no folio number is invented or
"rounded" to look more precise than what a search actually returned, and no new row is added to the
remediated lexicon by inference from a star's general fame — every added row (if any) must clear
the same `VERIFIED_NORMALIZED`-or-better bar as an existing row being upgraded.
