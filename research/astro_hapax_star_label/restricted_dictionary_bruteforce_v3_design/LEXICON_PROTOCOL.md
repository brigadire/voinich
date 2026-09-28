# Lexicon Protocol: Historical Astronomical Star Nomenclature (v3 Design)

## 1. Objective and Scope

This protocol governs the compilation and maintenance of the historical star lexicon for the Restricted Dictionary Brute-Force v3 experiment. The primary goal is to provide a scientifically defensible, historically attested corpus of star names circulating in Arabic, Latin, and Arabo-Latin astronomical traditions prior to and around the assumed creation of the Voynich manuscript (circa 8th to 15th centuries), completely free from the 31-identity ceiling of previous iterations.

## 2. Blindness to EVA (Target Independence)

To prevent circular reasoning, overfitting, and confirmation bias, the lexicon was compiled under strict blind conditions:

1. **No Target Token Consulting**: The inclusion of a star, name, or orthographic variant was determined solely by medieval astronomical provenance. No form was added, removed, or altered because it resembled or failed to resemble any Voynich EVA token.
2. **Preservation of Historical Orthography**: Historical manuscript and epigraphic spellings are recorded faithfully. No ad-hoc phonetic approximations, arbitrary dialectal permutations, or speculative transliterations were introduced.
3. **Segregation of Reconstruction**: Modern reconstructive readings are explicitly marked (`editorial_reconstruction=YES`), while historically attested forms are verified (`historically_attested=YES`).
4. **Canonical Identity Grouping**: Different attested spellings, dialectal variants, and multilingual translations of a single physical celestial object are linked to a unique `canonical_identity_id`. They do not constitute independent search capacity.

## 3. Schema Specifications

The lexicon is stored in `HISTORICAL_STAR_LEXICON.tsv` with 20 mandatory columns:

| Column | Description |
|---|---|
| `canonical_identity_id` | Unique identifier of the astronomical body (e.g. `STAR_ALDEBARAN`) |
| `canonical_identity_name` | Common modern identifier and Bayer designation (e.g. `Aldebaran / alpha Tauri`) |
| `attestation_id` | Unique sequential identifier for the attested entry (`HIST_STAR_ATT_xxxx`) |
| `attested_form` | Exact attested textual string from the historical witness or edition |
| `normalized_form` | Algorithmic lower-case ASCII form without diacritics or non-alphabetic characters |
| `language` | Source language (`ARABIC`, `LATIN`, `LATINIZED_ARABIC`, `OLD_CASTILIAN`, etc.) |
| `script` | Historical writing system (`LATIN`, `ARABIC_TRANSLIT`, `ARABIC_SCRIPT`) |
| `transliteration_system` | Standard used (`DIN_31635`, `MEDIEVAL_LATIN_DIRECT`, `SCHOLARLY_STANDARDIZED`) |
| `source_title` | Title of the treatise, star catalog, or instrument record |
| `source_author` | Historical author, translator, or modern editor |
| `source_date_start` | Earliest attested date of the witness (Julian year) |
| `source_date_end` | Latest attested date of the witness (Julian year) |
| `manuscript_or_edition` | Shelfmark of primary manuscript, museum accession number, or critical edition |
| `folio_or_entry` | Specific folio, page, or table pointer entry |
| `source_url_or_local_reference` | Verifiable digital repository link, DOI, or publication citation |
| `source_type` | Nature of witness (`PRIMARY_MANUSCRIPT_AND_CRITICAL_EDITION`, `HISTORICAL_EPIGRAPHY`, `CRITICAL_EDITION`) |
| `historically_attested` | `YES` or `NO` |
| `editorial_reconstruction` | `YES` or `NO` |
| `confidence` | Provenance certainty rating (`HIGH`, `MEDIUM`, `LOW`) |
| `notes` | Contextual philological, astronomical, or paleographical notes |

## 4. Normalization Rules

1. Strip Unicode combining diacritical marks (e.g. macrons, dots, accents: `ā` -> `a`, `ḥ` -> `h`, `ḍ` -> `d`).
2. Transliterate Arabic glottal stops (`’`) and pharyngeal fricatives (`‘`) to null or separate spaces.
3. Replace punctuation and non-alphabetic delimiters with single spaces.
4. Convert all Latin letters to lower case.
5. Deduplicate identical normalized forms within the same canonical identity.

## 5. Inclusion and Exclusion Boundaries

### Included
- Stars catalogued in Ptolemy's *Almagest* (books VII-VIII) and preserved in medieval Arabic or Latin translations.
- Stars described in Abd al-Rahman al-Sufi's *Kitāb Ṣuwar al-Kawākib al-Thābita* (964 AD).
- Pointers engraved on surviving medieval astrolabes (e.g. Oxford MHS 47632, Oxford MHS 41468, NMS T.1959.62).
- Astrolabe star tables circulating in Western Europe (Kunitzsch Types I, II, III, XV, John of London 1246, Chaucer 1391).
- Astronomical tables and catalogs produced under Alfonso X of Castile (*Libros del Saber de Astronomía*, 1252–1270).

### Excluded
- Post-1500 early modern and modern celestial cartography (Bayer 1603, Flamsteed 1729).
- Non-astronomical terms, astrological mansions without specific star references, and mythological figures without point-star coordinates.
- Damaged, illegible, or unidentifiable astrolabe pointers without secure celestial counterparts.
