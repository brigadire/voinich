# Lexicon provenance

Astronomical forms are copied exclusively from the locally frozen D1 expansion,
SHA256 9a2b97b0c5accd66565c8556ba1f857b44a833970c5690caedd6a5f1e2317b59,
STAR rows only. The original local bibliography is
`research/astro_dictionary_expansion_m1/ASTRO_TERM_CORPUS_EXPANSION_SOURCES.md`.
Attestation row IDs and canonical identities are separate; 63 attestations do not
mean 63 independently assignable stars (31 identities). Repeated normalized forms
are deduplicated inside identities. Original script and editorial transliteration
are separate columns. Historically_attested means inherited source attestation,
not independently dated verification of the supplied transliteration.

- National Museums Scotland, [T.1959.62 rete](https://www.nms.ac.uk/discover-catalogue/a-1000-year-old-star-catcher), replacement rete dated 13th c.
- Oxford History of Science Museum, [inventory 47632](https://www.mhs.ox.ac.uk/astrolabe/catalogue/reteReport/Astrolabe_ID=131.html), late ninth c., Arabic inscriptions.
- Oxford History of Science Museum, [inventory 41468](https://www.mhs.ox.ac.uk/astrolabe/catalogue/reteReport/Astrolabe_ID=237.html), ca.1400, Latin and Latinized Arabic inscriptions.
- Paul Kunitzsch (1987), “An unknown Arabic source for star names”, pp.155–163, [DOI](https://doi.org/10.1017/S0252921100105986); inherited ca.1246 bedalgeuze only.

Controls are frozen before production. Local downloaded source bytes are retained:

- Isidore, *Etymologiae* [VII](https://www.thelatinlibrary.com/isidore/7.shtml), early seventh-century text: manually selected attested personal-name forms, verified against the source by prepare.py. Inflected forms remain forms; anonymous capacity blocks are not historical equivalence claims.
- Isidore, *Etymologiae* [XVII](https://www.thelatinlibrary.com/isidore/17.shtml): lowercase alphabetic body-text types, including ordinary inflected/function words. No length- or score-based deletion. Latin Library HTML encoding CP1252; source-byte snapshots are frozen.
- Julius Ruska (ed.), *Das Buch der Alaune und Salze* (1935), [public-domain digitization](https://archive.org/details/buch-der-alaune-und-salze_-_liber_de_aluminibus_et_salibus__q218), Latin text pp.54–83, OCR lines 2401–3948. Reviewed list of 207 ordinary/technical forms, each verified in that excerpt. This is medieval Arabic-Latin translated alchemy, preserved in an early printed witness and editorial edition. The edition discusses its transmission immediately before the Latin text. Modern German commentary, raw unreviewed OCR and astronomical/planetary terms are excluded from the curated vocabulary. No precise medieval manuscript orthography is claimed for the printed forms.

Historical controls use nearest-length sampling; unmatched length tails and mixed
language/chronology remain measurable limitations. No dictionary here is a complete
inventory of medieval vocabulary or star nomenclature. The synthetic controls
provide exact size/length/endpoints and unigram/bigram matching where specified;
the historical controls provide real attestation, not full distributional parity.
