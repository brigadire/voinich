# Lexicon Coverage Report: Historical Star Nomenclature (v3 Design)

## 1. Executive Summary

This report documents the coverage, provenance, and completeness of the historical star lexicon compiled for the Restricted Dictionary Brute-Force v3 design (`HISTORICAL_STAR_LEXICON.tsv`).

The previous iteration (v2) was halted because its star lexicon was restricted to 31 canonical identities, enforcing an artificial coverage ceiling of 31/57 on the frozen target scope. The v3 lexicon resolves this blocker by incorporating **94 canonical star identities** and **307 distinct historical attestations**, fully satisfying the `CANONICAL_IDENTITIES >= 57` requirement.

| Metric | Target / Requirement | v3 Lexicon Actual | Status |
|---|---|---:|:---:|
| `CANONICAL_IDENTITIES` | $\ge 57$ | **94** | **PASS** |
| `HISTORICAL_ATTESTATIONS` | Unbounded | **307** | **PASS** |
| `PEER_REVIEWED_OR_PRIMARY_SOURCES` | $100\%$ | **100.0%** (307/307) | **PASS** |
| `LEXICON_INDEPENDENT_OF_EVA` | Mandatory `YES` | **YES** | **PASS** |
| `ATTESTATIONS_DEDUPLICATED` | Mandatory `YES` | **YES** | **PASS** |
| `CONFIDENCE_HIGH_FRACTION` | $\ge 90\%$ | **97.1%** (298/307) | **PASS** |

## 2. Included Catalogues and Traditions

The lexicon synthesizes twelve primary and critical scholarly sources representing the four principal medieval astronomical traditions:

1. **Classical Islamic & Ptolemaic Catalogues**:
   - Abd al-Rahman al-Sufi, *Kitāb Ṣuwar al-Kawākib al-Thābita* (964 AD; MS Bodleian Marsh 144 / ed. Schjellerup 1874): 72 attestations.
   - Oxford Museum of the History of Science, Inventory 47632 (late 9th century Syro-Mesopotamian astrolabe rete): 19 epigraphic attestations.
2. **Arabo-Latin Toledo Translations**:
   - Gerard of Cremona, translation of Ptolemy's *Syntaxis Mathematica* (*Almagestum*, Toledo ca. 1175; ed. P. Kunitzsch 1974): 86 attestations.
   - Johannes Hispalensis (John of Seville), astrolabe star list (ca. 1140; ed. P. Kunitzsch 1966, Type III): 8 attestations.
   - Pseudo-Messahalla star list (12th c.; ed. P. Kunitzsch 1966, Type II): 6 attestations.
3. **Alfonsine Royal Workshop**:
   - Alfonso X el Sabio, *Libros del Saber de Astronomía* (esp. *Libro de las estrellas fixas*, Toledo 1252–1270; ed. M. Rico y Sinobas 1863): 57 attestations.
4. **Western European Astrolabes and Star Tables (13th–15th c.)**:
   - National Museums Scotland, Moorish astrolabe replacement rete (T.1959.62, 13th c.): 8 attestations.
   - John of London, Paris astrolabe table (1246; ed. P. Kunitzsch 1987): 1 attestation.
   - Geoffrey Chaucer, *A Treatise on the Astrolabe* (1391; ed. W.W. Skeat 1872): 28 attestations.
   - Oxford Museum of the History of Science, Inventory 41468 (ca. 1400 Paris rete): 10 attestations.
   - Johannes von Gmunden / Vienna School of Astronomy (MS Vienna ÖNB 5415, ca. 1440; ed. P. Kunitzsch 1966, Type XV): 12 attestations.

## 3. Absent Catalogues and Exclusion Justification

The following corpora were deliberately excluded:
- **Ulugh Beg's *Zij-i Sultani* (1437 AD)**: Excluded to maintain focus on the Mediterranean and Western European transmission lines directly contemporary with the manuscript's radiocarbon date range.
- **Byzantine Greek Star Lists**: Excluded due to the absence of Greek loanwords or distinct Greek script variants in Western European astrolabe manufacture.
- **Post-1500 Early Modern Catalogues**: Works by Alessandro Piccolomini (1540), Johannes Bayer (*Uranometria*, 1603), and John Flamsteed (1729) were excluded as outside the chronological boundary.
- **Unverified Pointers**: Pointers with broken legends or missing astronomical identifications were excluded rather than guessed.

## 4. Distributional Breakdowns

### By Language
- **Latin**: 110 attestations (35.8%)
- **Arabic (Standard/Transliterated)**: 98 attestations (31.9%)
- **Latinized Arabic**: 65 attestations (21.2%)
- **Middle English & Latin (Chaucer)**: 28 attestations (9.1%)
- **Latinized Greek**: 4 attestations (1.3%)
- **Old Castilian (Alfonsine)**: 2 attestations (0.7%)

### By Source Type
- **Primary Manuscript & Critical Edition**: 141 attestations (45.9%)
- **Critical Scholarly Edition**: 129 attestations (42.0%)
- **Historical Epigraphy (Physical Instruments)**: 37 attestations (12.1%)

### By Century
- **9th Century (800–899)**: 19 attestations (6.2%)
- **10th Century (900–999)**: 72 attestations (23.5%)
- **12th Century (1100–1199)**: 100 attestations (32.6%)
- **13th Century (1200–1299)**: 66 attestations (21.5%)
- **14th Century (1300–1399)**: 38 attestations (12.4%)
- **15th Century (1400–1450)**: 12 attestations (3.9%)

## 5. Provenance and Scientific Integrity

Every entry in `HISTORICAL_STAR_LEXICON.tsv` includes primary manuscript shelfmarks, critical edition page/folio references, or museum accession numbers. Exactly 100% of rows are derived from authoritative scholarly editions (principally the works of Paul Kunitzsch, the preeminent historian of Arabic and medieval Latin star nomenclature) or direct museum physical records.

No form was modified, shortened, or elongated to fit the 57 Voynich star labels, guaranteeing true blind compilation.
