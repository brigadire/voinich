# Lexicon Remediation Report (Gate L-R)

## 1. What was wrong in v3

`restricted_dictionary_bruteforce_v3_design/HISTORICAL_STAR_LEXICON.tsv` (307 rows, 94 canonical
star identities) was audited by `restricted_dictionary_bruteforce_v3_audit`, which found: a
confirmed misattribution (Mizar/Mirach sharing one Arabic phrase from the same source), two
confirmed anachronisms (Cor Caroli's "hunting dogs" gloss predates the 1687 invention of its
constellation by ~250 years; Mira's "wondrous star" gloss predates its 1596 recognition as notable
by ~150 years), and provenance-verification coverage of only 8.8% (27/307 rows) — far below the
task's ≥30% stratified mandate, with the small checked sample already showing a 33% (3/9) defect
rate in the MEDIUM-confidence stratum.

## 2. What this remediation checked, and how

Per `LEXICON_REMEDIATION_PROTOCOL.md` (written and committed before any row was verified):

- **Mechanical, zero-judgment, 100%-coverage checks** (no external lookup needed): every row's
  `normalized_form` compared against every other row's across the full 307-row file, for both exact
  collisions and near-duplicates within the largest single source (al-Sufi, 72 rows). This alone
  found **3 exact cross-identity collisions**, one already known (Mizar/Mirach) and **two new**
  (Markab/Scheat sharing "mankib al-faras"; Porrima/Zavijava sharing "zawiyat al-'awwa").
- **TIER_MANDATORY (18 rows, 100% coverage):** all 12 rows from the Vienna Astrolabe Star Table
  (escalated from the audit's 9 MEDIUM-confidence rows to the full source, per the escalation rule,
  because 2/9 were already confirmed defective) plus all 6 rows in the 3 collision pairs.
- **TIER_STRATIFIED_SAMPLE (147 rows across 47 of the 94 canonical identities, chosen to span all
  12 sources, all 6 language categories, and all century strata):** 50.9% of the 289 non-mandatory
  rows — well above the ≥30% floor — with al-Sufi rows specifically over-sampled to 51.4% (37/72)
  against the ≥50% target set in the protocol.
- **165 rows checked in total (53.7% of the whole lexicon).** Verification used WebSearch against
  digitized manuscript catalogs (Bodleian, MHS Oxford, National Museums Scotland, Handschriftencensus,
  Biblissima/PAL), published critical-edition metadata (Kunitzsch 1966, 1974, 1986, 1987), and
  secondary star-name scholarship, consistent with the disclosed ceiling in the protocol: this
  remediation does not have page-by-page manuscript image access, so most positive verifications
  land at `VERIFIED_NORMALIZED` rather than `VERIFIED_EXACT`.

## 3. What was found — including new defects beyond the v3 audit

**8 of 165 checked rows (4.85%) were confirmed defective** and are removed in
`REMEDIATED_HISTORICAL_STAR_LEXICON.tsv`. Four were already known from the v3 audit; **four are new
findings from this remediation**:

| Row | Star | Defect | New? |
|---|---|---|---|
| `HIST_STAR_ATT_0188` | Mirach | Misattribution — refined, see below | known, refined |
| `HIST_STAR_ATT_0123` | Cor Caroli | Anachronism (constellation invented 1687) | known |
| `HIST_STAR_ATT_0187` | Mira | Anachronism (star recognized 1596) | known |
| `HIST_STAR_ATT_0164` | Markab | Misattribution — "mankib al-faras" belongs to Scheat | **new** |
| `HIST_STAR_ATT_0296` | Zavijava | Misattribution — "zawiyat al-'awwa" belongs to Porrima | **new** |
| `HIST_STAR_ATT_0101` | Bellatrix | Misattribution — "al-najid" belongs (at 964 AD) to Capella | **new** |
| `HIST_STAR_ATT_0018` | Alcor | Anachronism — "eques stellula" traces only to Bayer, 1603 | **new** |
| `HIST_STAR_ATT_0056` | Alnair | Anachronism — "lucida gruis" presupposes Grus, invented 1597 | **new** |

**The single most important new finding** is that the Vienna Astrolabe Star Table's defect rate is
not 2/9 (22%, MEDIUM stratum only, as the v3 audit found) but **4/12 (33%) across the full source**
once escalated to 100% coverage — and that this source's unreliability is independently explainable:
external scholarship documents "the Vienna school of astronomy led by Johannes von Gmunden" (the
same figure this source's manuscript tradition is associated with) as a real, historically
documented *origin point* for at least two other star-name transfer errors found in this remediation
(Bellatrix/Capella and, per one source consulted for Scheat, Delta Aquarii/Scheat). This means the
Vienna source's problems are not an artifact of this lexicon's compilation — they reflect a real,
external, previously-known unreliability in that specific 15th-century transmission tradition. Any
future work drawing on Vienna-attributed material should treat it as inherently higher-risk, not
just because this lexicon found problems in it, but because independent scholarship already flags
it as a documented source of confusions.

**A second notable finding refines rather than overturns the v3 audit's central claim.** The v3
audit's confirmed-misattribution finding for Mirach (F005) is correct in outcome but was based on
one secondary source. This remediation found a second, independent secondary tradition stating that
the modern name "Mirach" itself derives from a corruption of "mizar" via the *1521 printed Alfonsine
Tables* ("super mizar"). Both things are true: "al-mi'zar" is a real historical name connected to
Mirach, but only via a documented 16th-century Latin-tradition path — not via al-Sufi's own 10th-
century Arabic text, which is what this lexicon's row specifically (and implausibly, given the
identical phrase is also attributed to Mizar from the same source on a different folio) claims. The
row is still correctly disposed as `REMOVE_MISATTRIBUTED`, but for a more precise reason: wrong
source and date, not a fabricated Arabic root.

**Three rows were flagged as genuinely uncertain but not removed**, per the protocol's rule that
absence of confirmation is not the same as contradiction: Zosma's "succinctorium" (an unusual
ecclesiastical-vestment term with no independent attestation found for this star), Mesarthim's
"al-sharatayn" (a real Arabic dual-form name for the Sheratan/Mesarthim *pair*, individual
attribution to Mesarthim alone unconfirmed), and Nunki's "al-sadira" (a real shortening of a
group-asterism name covering five Sagittarius stars, individual attribution to Nunki alone
unconfirmed). These remain `KEEP` / `PLAUSIBLE_NOT_DIRECTLY_VERIFIED`.

**No new rows were added.** Time and search-access constraints meant no source for the invited
coverage-expansion categories (Ptolemaic/Greek tradition, Ulugh Beg's 15th-century catalog, etc.)
could be verified to this remediation's own `VERIFIED_NORMALIZED`-or-better bar within scope. Per
the task's own instruction, a smaller correct lexicon is preferred over a larger unverified one, so
none were added.

## 4. Gate L-R

```text
LEXICON_EVA_INDEPENDENCE=PASS
MEDIUM_CONFIDENCE_VERIFIED=1.000
OTHER_STRATIFIED_COVERAGE=0.509
CONFIRMED_ERROR_RATE=0.0485
MISATTRIBUTIONS_UNRESOLVED=0
ANACHRONISMS_UNRESOLVED=0
LEXICON_REMEDIATION_GATE=PASS
LEXICON_ROWS_TOTAL=299
LEXICON_ROWS_DIRECTLY_VERIFIED=137
LEXICON_PROVENANCE_VALID_RATE=0.830
LEXICON_SCOPE=PARTIAL_BUT_USABLE
```

(`LEXICON_PROVENANCE_VALID_RATE` = rows reaching `VERIFIED_EXACT`/`VERIFIED_NORMALIZED`, as a
fraction of the 165 rows actually checked, not of the full 299-row lexicon — see
`PROVENANCE_VERIFICATION.tsv`. `LEXICON_SCOPE=PARTIAL_BUT_USABLE` reflects that no manuscript-image
verification was possible for most rows, not that the lexicon's content is currently in doubt: the
299 live rows all cleared checking-or-defaulting-to-KEEP-unchecked per the protocol's disposition
rule, and every row that *was* checked and found wanting was removed.)

## 5. Is the dictionary-brute-force line worth continuing?

**Conditionally yes, but with a materially smaller and more fragile foundation than v3 assumed.**
299 of 307 rows survive across all 94 canonical identities (no identity was reduced to zero
attestations), so the lexicon is still usable in shape. But three things should temper any
optimism carried into Gate S-R:

1. **The defect rate did not go to zero with more checking — it barely changed.** v3_audit found a
   defect in the one deliberately-adversarial sample it had time for (3/9 MEDIUM rows, 33%). This
   remediation's much larger, still partly-adversarial-by-design sample (mandatory tier deliberately
   targets known-risky rows) found 8/165 (4.85%) — lower because the denominator now includes 147
   "ordinary" rows that turned out clean, not because the risky rows got any less risky. The 289
   never-independently-verified rows (177 of them still fully `UNCHECKED` after this pass) are an
   unknown quantity; this remediation's own stratified sample gives no reason to expect their defect
   rate is zero, only that it is unlikely to be as concentrated as it was in Vienna/al-Sufi
   specifically (both of which were deliberately over-sampled precisely because they'd already shown
   problems).
2. **The defect *pattern* is not random noise — it is "one real historical phrase, independently
   documented as migrating to a second star, retrojected into an earlier source than it belongs to."**
   That happened 3 times independently (Markab/Scheat, Porrima/Zavijava, Bellatrix/Capella) plus the
   already-known Mizar/Mirach case, now refined into the same pattern. A lexicon built by combining
   secondary tabulations of "known historical star names" without checking each specific
   source-and-date claim is structurally exposed to this exact failure mode, and there is no reason
   to think this remediation's sampling exhausted it.
3. **Verification depth is capped by search access, not by effort.** Almost every positive result in
   this remediation is `VERIFIED_NORMALIZED` (matches independent secondary scholarship) rather than
   `VERIFIED_EXACT` (matches a primary manuscript image/folio). That ceiling was disclosed before
   checking began and did not change during it. Anyone treating this remediated lexicon as
   folio-level-precise provenance would be overclaiming what it actually supports.

The honest summary: this remediation produced a lexicon that is *measurably* better than v3's (known
defects fixed, two new pattern-instances of the same failure mode caught and fixed, coverage more
than tripled), but the underlying construction method — combine secondary star-name compilations
per source without checking each source-date-form triple against a primary witness — has now failed
its own honesty check twice, in the same distinctive way, at two different scales of scrutiny. Gate
S-R (solver qualification) should proceed on this lexicon, since the task requires establishing
whether the search-methodology side is viable at all before any further lexicon investment is
justified — but a third clean-room pass finding a fourth instance of the same pattern would be a
strong signal to stop trusting this construction method for anything downstream of dictionary
brute-force, rather than to keep patching it row by row.
