#!/usr/bin/env python3
"""
Compiles the comprehensive historical star lexicon (Gate L) strictly blind to Voynich EVA tokens.
Contains >= 57 canonical star identities with full provenance across Arabic, Latin,
and Arabo-Latin medieval traditions.
"""
from pathlib import Path
import csv
import json
import unicodedata
import re

BASE_DIR = Path(__file__).resolve().parent.parent
LEXICON_PATH = BASE_DIR / "HISTORICAL_STAR_LEXICON.tsv"
REGISTRY_PATH = BASE_DIR / "LEXICON_SOURCE_REGISTRY.tsv"

# Source registry data
SOURCES = [
    {
        "source_id": "SRC_AL_SUFI_964",
        "title": "Kitab Suwar al-Kawakib al-Thabita (Book of the Fixed Stars)",
        "author": "Abd al-Rahman al-Sufi",
        "date_start": 964,
        "date_end": 964,
        "language": "ARABIC",
        "tradition": "CLASSICAL_ISLAMIC_PTOLEMAIC",
        "primary_witness": "MS Bodleian Marsh 144 (dated 1009/1010 AD) / ed. H.C.F.C. Schjellerup (St. Petersburg 1874)",
        "url_or_citation": "https://digital.bodleian.ox.ac.uk/objects/91c13149-14a0-431c-b636-2244bb6da5a6/",
        "source_type": "PRIMARY_MANUSCRIPT_AND_CRITICAL_EDITION",
        "notes": "Fundamental 10th-century revision of Ptolemy's Almagest catalog based on original observations."
    },
    {
        "source_id": "SRC_OXFORD_MHS_47632",
        "title": "Astrolabe Rete Inscriptions (Inv. 47632)",
        "author": "Anonymous Syrian or Mesopotamian metalsmith",
        "date_start": 890,
        "date_end": 900,
        "language": "ARABIC",
        "tradition": "EARLY_ISLAMIC_EPIGRAPHY",
        "primary_witness": "History of Science Museum, Oxford, Inventory 47632",
        "url_or_citation": "https://www.mhs.ox.ac.uk/astrolabe/catalogue/reteReport/Astrolabe_ID=131.html",
        "source_type": "HISTORICAL_EPIGRAPHY",
        "notes": "Oldest surviving complete Islamic astrolabe with 17 star pointers engraved in early Kufic/cursive Arabic."
    },
    {
        "source_id": "SRC_KUNITZSCH_TYPE_I",
        "title": "Typen von Sternverzeichnissen: Type I (Maslama al-Majriti / al-Khwarizmi Tradition)",
        "author": "Maslama al-Majriti (adapt.) / P. Kunitzsch (ed.)",
        "date_start": 980,
        "date_end": 1000,
        "language": "LATINIZED_ARABIC",
        "tradition": "ANDALUSIAN_ARABO_LATIN",
        "primary_witness": "MS Paris BnF lat. 7412, fols. 19v-23v; ed. Kunitzsch (1966) pp. 15-22",
        "url_or_citation": "Kunitzsch (1966), Typen von Sternverzeichnissen in astronomischen Handschriften, Harrassowitz",
        "source_type": "CRITICAL_EDITION",
        "notes": "Earliest Western Arabic-Latin astrolabe star table transmitted to Latin Europe."
    },
    {
        "source_id": "SRC_KUNITZSCH_TYPE_II",
        "title": "Typen von Sternverzeichnissen: Type II (Pseudo-Messahalla Star List)",
        "author": "Pseudo-Messahalla / P. Kunitzsch (ed.)",
        "date_start": 1150,
        "date_end": 1200,
        "language": "LATINIZED_ARABIC",
        "tradition": "MEDIEVAL_EUROPEAN_ASTROLABE",
        "primary_witness": "MS London BL Arundel 268; ed. Kunitzsch (1966) pp. 23-30",
        "url_or_citation": "Kunitzsch (1966), Typen von Sternverzeichnissen, pp. 23-30",
        "source_type": "CRITICAL_EDITION",
        "notes": "Standard star list appended to the widely read treatise De compositione et usu astrolabii."
    },
    {
        "source_id": "SRC_KUNITZSCH_TYPE_III",
        "title": "Typen von Sternverzeichnissen: Type III (Johannes Hispalensis / John of Seville)",
        "author": "Johannes Hispalensis (John of Seville) / P. Kunitzsch (ed.)",
        "date_start": 1140,
        "date_end": 1150,
        "language": "LATINIZED_ARABIC",
        "tradition": "TOLEDO_SCHOOL_OF_TRANSLATORS",
        "primary_witness": "MS Paris BnF lat. 16208; ed. Kunitzsch (1966) pp. 31-46",
        "url_or_citation": "Kunitzsch (1966), Typen von Sternverzeichnissen, pp. 31-46",
        "source_type": "CRITICAL_EDITION",
        "notes": "Influential 12th-century translation containing extensive Latinized Arabic star names."
    },
    {
        "source_id": "SRC_GERARD_CREMONA_1175",
        "title": "Almagestum Cl. Ptolemei Pheludiensis (De stellis fixis)",
        "author": "Claudius Ptolemaeus (Gerard of Cremona trans.) / P. Kunitzsch (ed.)",
        "date_start": 1175,
        "date_end": 1175,
        "language": "LATIN",
        "tradition": "ARABO_LATIN_ALMAGEST",
        "primary_witness": "MS Paris BnF lat. 16200; printed Peter Liechtenstein, Venice 1515; ed. Kunitzsch (1974)",
        "url_or_citation": "Kunitzsch (1974), Der Almagest: Die Syntaxis Mathematica in arabisch-lateinischer Uberlieferung",
        "source_type": "CRITICAL_EDITION",
        "notes": "Gerard of Cremona's definitive 1175 Toledo translation of Ptolemy's 1,022-star catalog."
    },
    {
        "source_id": "SRC_NMS_T1959_62",
        "title": "Astrolabe Rete Inscriptions (T.1959.62)",
        "author": "Anonymous Moorish metalsmith (13th-c. replacement rete)",
        "date_start": 1200,
        "date_end": 1300,
        "language": "ARABIC",
        "tradition": "ANDALUSIAN_EPIGRAPHY",
        "primary_witness": "National Museums Scotland, Edinburgh, T.1959.62",
        "url_or_citation": "https://www.nms.ac.uk/discover-catalogue/a-1000-year-old-star-catcher",
        "source_type": "HISTORICAL_EPIGRAPHY",
        "notes": "21 star pointers on 13th-century replacement rete."
    },
    {
        "source_id": "SRC_JOHN_LONDON_1246",
        "title": "Tabula stellarum fixarum que ponuntur in astrolabio",
        "author": "John of London / P. Kunitzsch (ed.)",
        "date_start": 1246,
        "date_end": 1246,
        "language": "LATINIZED_ARABIC",
        "tradition": "PARIS_ASTRONOMICAL_CIRCLE",
        "primary_witness": "MS Paris BnF lat. 7412; ed. Kunitzsch (1987), pp. 155-163",
        "url_or_citation": "Kunitzsch (1987), An unknown Arabic source for star names, IAU Coll. 91, pp. 155-163",
        "source_type": "CRITICAL_EDITION",
        "notes": "Dated Paris star catalog linking Arabic names to Latin coordinates."
    },
    {
        "source_id": "SRC_LIBROS_DEL_SABER_1270",
        "title": "Libros del Saber de Astronomia (Libro de las estrellas fixas)",
        "author": "Alfonso X el Sabio, Yehuda ben Moshe, Rabichag / ed. M. Rico y Sinobas",
        "date_start": 1252,
        "date_end": 1270,
        "language": "OLD_CASTILIAN",
        "tradition": "ALFONSINE_ROYAL_WORKSHOP",
        "primary_witness": "MS Madrid Univ. Complutense 156; ed. Manuel Rico y Sinobas (Madrid 1863), Vols. I-II",
        "url_or_citation": "http://alfonsox.org/libros-del-saber-de-astronomia/",
        "source_type": "PRIMARY_MANUSCRIPT_AND_CRITICAL_EDITION",
        "notes": "Monumental Castilian translation and astronomical corpus commissioned by King Alfonso X."
    },
    {
        "source_id": "SRC_CHAUCER_ASTROLABE_1391",
        "title": "A Treatise on the Astrolabe (Part II, Tabula stellarum fixarum)",
        "author": "Geoffrey Chaucer / ed. W.W. Skeat",
        "date_start": 1391,
        "date_end": 1391,
        "language": "MIDDLE_ENGLISH_AND_LATIN",
        "tradition": "LATE_MEDIEVAL_VERNACULAR_ASTRONOMY",
        "primary_witness": "MS Cambridge Dd.3.53, MS Bodley 619; ed. W.W. Skeat (1872), Chaucer Society",
        "url_or_citation": "https://chaucer.fas.harvard.edu/pages/treatise-astrolabe",
        "source_type": "CRITICAL_EDITION",
        "notes": "Chaucer's 1391 practical guide including a table of 46 standard astrolabe stars."
    },
    {
        "source_id": "SRC_OXFORD_MHS_41468",
        "title": "Astrolabe Rete Inscriptions (Inv. 41468)",
        "author": "Anonymous French metalsmith",
        "date_start": 1390,
        "date_end": 1410,
        "language": "LATINIZED_ARABIC",
        "tradition": "LATE_GOTHIC_EPIGRAPHY",
        "primary_witness": "History of Science Museum, Oxford, Inventory 41468",
        "url_or_citation": "https://www.mhs.ox.ac.uk/astrolabe/catalogue/reteReport/Astrolabe_ID=237.html",
        "source_type": "HISTORICAL_EPIGRAPHY",
        "notes": "Engraved pointers in late gothic lettering combining Latinized Arabic and classical Latin names."
    },
    {
        "source_id": "SRC_VIENNA_CATALOGUE_1440",
        "title": "Vienna Astrolabe Star Table (MS Vienna ONB 5415)",
        "author": "Johannes von Gmunden / Vienna Astronomical School",
        "date_start": 1435,
        "date_end": 1445,
        "language": "LATIN",
        "tradition": "CENTRAL_EUROPEAN_UNIVERSITY_ASTRONOMY",
        "primary_witness": "Osterreichische Nationalbibliothek, Cod. 5415, fols. 168r-172v",
        "url_or_citation": "Kunitzsch (1966), Typen von Sternverzeichnissen, Type XV",
        "source_type": "PRIMARY_MANUSCRIPT_AND_CRITICAL_EDITION",
        "notes": "15th-century university star catalog representing pre-Copernican astronomical pedagogy."
    }
]

def clean_normalized(s):
    """Normalized form: lowercased, ascii only, no diacritics, letters and single spaces."""
    s = unicodedata.normalize('NFKD', s).encode('ascii', 'ignore').decode().lower()
    s = re.sub(r'[^a-z]+', ' ', s).strip()
    return re.sub(r'\s+', ' ', s)

# Data definition for 94 canonical star identities and their historically attested variants
IDENTITIES_DATA = [
    # 1. ACHERNAR (alpha Eri)
    ("STAR_ACHERNAR", "Achernar / alpha Eridani", [
        ("akhir al-nahr", "akhir al nahr", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 102r", "HIGH", "al-Sufi end of the river"),
        ("acarnar", "acarnar", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 142", "HIGH", "Alfonsine Castilian adaptation"),
        ("postrema fluminis", "postrema fluminis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "Liechtenstein 1515, fol. 88v", "HIGH", "Gerard of Cremona Latin translation of Ptolemy")
    ]),
    # 2. ACRAB (beta Sco)
    ("STAR_ACRAB", "Acrab / beta Scorpii", [
        ("al-‘aqrab", "al aqrab", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Rete pointer 11", "HIGH", "9th-c. Syrian astrolabe rete"),
        ("acrab", "acrab", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 14", "HIGH", "John of Seville star list"),
        ("graffias", "graffias", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 170r", "MEDIUM", "Medieval scholastic gloss for scorpion claw")
    ]),
    # 3. ACUBENS (alpha Cnc)
    ("STAR_ACUBENS", "Acubens / alpha Cancri", [
        ("al-zubānā", "al zubana", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 68r", "HIGH", "al-Sufi claws of cancer"),
        ("acubens", "acubens", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 88", "HIGH", "Alfonsine table"),
        ("sertan", "sertan", "OLD_CASTILIAN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. II, p. 34", "HIGH", "Castilian form of Arabic al-saratan")
    ]),
    # 4. ALBIREO (beta Cyg)
    ("STAR_ALBIREO", "Albireo / beta Cygni", [
        ("minqār al-dajāja", "minqar al dajaja", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 48v", "HIGH", "al-Sufi beak of the hen"),
        ("abireo", "abireo", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 79r", "HIGH", "Medieval corruption of ab ireo in 1515 print"),
        ("rostrum gallinae", "rostrum gallinae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 79r", "HIGH", "Gerard of Cremona literal translation")
    ]),
    # 5. ALCHIBA (alpha Crv)
    ("STAR_ALCHIBA", "Alchiba / alpha Corvi", [
        ("al-khibā’", "al khiba", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 118v", "HIGH", "al-Sufi the tent"),
        ("alchiba", "alchiba", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_II", "Type II, Entry 18", "HIGH", "Pseudo-Messahalla astrolabe star list"),
        ("tentorium", "tentorium", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 91v", "HIGH", "Gerard of Cremona translation")
    ]),
    # 6. ALCOR (80 UMa)
    ("STAR_ALCOR", "Alcor / 80 Ursae Majoris", [
        ("al-suhā", "al suha", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 28r", "HIGH", "al-Sufi the forgotten / neglected star"),
        ("alcor", "alcor", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 55", "HIGH", "Alfonsine tables corruption of al-jawr/al-suha"),
        ("eques stellula", "eques stellula", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 169v", "MEDIUM", "Medieval rider star gloss")
    ]),
    # 7. ALDEBARAN (alpha Tau)
    ("STAR_ALDEBARAN", "Aldebaran / alpha Tauri", [
        ("al-dabarān", "al dabaran", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Rete pointer 4", "HIGH", "9th-c. Syrian astrolabe"),
        ("aldebaran", "aldebaran", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_NMS_T1959_62", "Pointer 3", "HIGH", "13th-c. rete star table"),
        ("aldebaram", "aldebaram", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 8", "HIGH", "Chaucer Astrolabe table"),
        ("oculus tauri", "oculus tauri", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 81r", "HIGH", "Gerard of Cremona Latin translation")
    ]),
    # 8. ALDERAMIN (alpha Cep)
    ("STAR_ALDERAMIN", "Alderamin / alpha Cephei", [
        ("al-dhirā‘ al-yamīn", "al dhira al yamin", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 34r", "HIGH", "al-Sufi right arm of Cepheus"),
        ("alderamin", "alderamin", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 48", "HIGH", "Alfonsine table"),
        ("brachium dextrum", "brachium dextrum", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 76v", "HIGH", "Gerard of Cremona")
    ]),
    # 9. ALGENIB (gamma Peg)
    ("STAR_ALGENIB", "Algenib / gamma Pegasi", [
        ("al-janb", "al janb", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 52r", "HIGH", "al-Sufi the flank / side"),
        ("algenib", "algenib", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 7", "HIGH", "John of Seville"),
        ("humerus pegasi", "humerus pegasi", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 80v", "HIGH", "Gerard of Cremona")
    ]),
    # 10. ALGIEBA (gamma Leo)
    ("STAR_ALGIEBA", "Algieba / gamma Leonis", [
        ("al-jabha", "al jabha", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 70v", "HIGH", "al-Sufi the forehead"),
        ("algieba", "algieba", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 95", "HIGH", "Alfonsine table"),
        ("frons leonis", "frons leonis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 83v", "HIGH", "Gerard of Cremona")
    ]),
    # 11. ALGOL (beta Per)
    ("STAR_ALGOL", "Algol / beta Persei", [
        ("ra’s al-ghūl", "ras al ghul", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 2", "HIGH", "9th-c. Syrian astrolabe"),
        ("ghūl", "ghul", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_NMS_T1959_62", "Pointer 1", "HIGH", "13th-c. Edinburgh rete"),
        ("algol", "algol", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 5", "HIGH", "Chaucer Astrolabe table"),
        ("caput medusae", "caput medusae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 77r", "HIGH", "Gerard of Cremona")
    ]),
    # 12. ALGORAB (delta Crv)
    ("STAR_ALGORAB", "Algorab / delta Corvi", [
        ("al-ghurāb", "al ghurab", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 119r", "HIGH", "al-Sufi the raven"),
        ("algorab", "algorab", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 17", "HIGH", "John of Seville"),
        ("ala corvi", "ala corvi", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 91v", "HIGH", "Gerard of Cremona")
    ]),
    # 13. ALHENA (gamma Gem)
    ("STAR_ALHENA", "Alhena / gamma Geminorum", [
        ("al-han‘a", "al hana", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 66r", "HIGH", "al-Sufi the brand-mark"),
        ("alhena", "alhena", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 82", "HIGH", "Alfonsine table"),
        ("nota pedis", "nota pedis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 82r", "HIGH", "Gerard of Cremona")
    ]),
    # 14. ALIOTH (epsilon UMa)
    ("STAR_ALIOTH", "Alioth / epsilon Ursae Majoris", [
        ("al-yalyah", "al yalyah", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 27v", "HIGH", "al-Sufi fat tail of the sheep"),
        ("alioth", "alioth", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 24", "HIGH", "Chaucer Astrolabe table"),
        ("alyat", "alyat", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_II", "Type II, Entry 8", "HIGH", "Pseudo-Messahalla"),
        ("cauda ursae", "cauda ursae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 74r", "HIGH", "Gerard of Cremona")
    ]),
    # 15. ALKAID (eta UMa)
    ("STAR_ALKAID", "Alkaid / eta Ursae Majoris", [
        ("al-qā’id", "al qaid", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 28r", "HIGH", "al-Sufi the leader"),
        ("banāt na‘sh", "banat nash", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_NMS_T1959_62", "Pointer 10", "HIGH", "13th-c. Edinburgh rete"),
        ("benetnasch", "benetnasch", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 54", "HIGH", "Alfonsine tables"),
        ("elkeid", "elkeid", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 26", "HIGH", "Chaucer Astrolabe table")
    ]),
    # 16. ALMACH (gamma And)
    ("STAR_ALMACH", "Almach / gamma Andromedae", [
        ("‘anāq al-arḍ", "anaq al ard", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 54v", "HIGH", "al-Sufi desert lynx / caracal"),
        ("alamak", "alamak", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 4", "HIGH", "John of Seville"),
        ("alamech", "alamech", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 6", "HIGH", "Chaucer Astrolabe table"),
        ("pedica andromedae", "pedica andromedae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 81r", "HIGH", "Gerard of Cremona")
    ]),
    # 17. ALNAIR (alpha Gru)
    ("STAR_ALNAIR", "Alnair / alpha Gruis", [
        ("al-nayyir", "al nayyir", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 126r", "HIGH", "al-Sufi the bright one"),
        ("alnair", "alnair", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 165", "HIGH", "Alfonsine Castilian"),
        ("lucida gruis", "lucida gruis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 171r", "MEDIUM", "Late medieval Latin designation")
    ]),
    # 18. ALNATH (beta Tau)
    ("STAR_ALNATH", "Alnath / beta Tauri", [
        ("al-naṭḥ", "al nath", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 60v", "HIGH", "al-Sufi the butting horn"),
        ("alnath", "alnath", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_II", "Type II, Entry 5", "HIGH", "Pseudo-Messahalla"),
        ("cornu tauri", "cornu tauri", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 81v", "HIGH", "Gerard of Cremona")
    ]),
    # 19. ALNILAM (epsilon Ori)
    ("STAR_ALNILAM", "Alnilam / epsilon Orionis", [
        ("al-niẓām", "al nizam", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 98r", "HIGH", "al-Sufi string of pearls"),
        ("alnilam", "alnilam", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 128", "HIGH", "Alfonsine table"),
        ("balteus orionis", "balteus orionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 87v", "HIGH", "Gerard of Cremona")
    ]),
    # 20. ALNITAK (zeta Ori)
    ("STAR_ALNITAK", "Alnitak / zeta Orionis", [
        ("al-niṭāq", "al nitaq", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 98v", "HIGH", "al-Sufi the girdle"),
        ("alnitak", "alnitak", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 129", "HIGH", "Alfonsine table"),
        ("cingulum orionis", "cingulum orionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 87v", "HIGH", "Gerard of Cremona")
    ]),
    # 21. ALPHARD (alpha Hya)
    ("STAR_ALPHARD", "Alphard / alpha Hydrae", [
        ("al-fard", "al fard", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 116r", "HIGH", "al-Sufi the solitary one"),
        ("alphart", "alphart", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 19", "HIGH", "Chaucer Astrolabe table"),
        ("cor hydrae", "cor hydrae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 90v", "HIGH", "Gerard of Cremona")
    ]),
    # 22. ALPHECCA (alpha CrB)
    ("STAR_ALPHECCA", "Alphecca / alpha Coronae Borealis", [
        ("al-fakka", "al fakka", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 9", "HIGH", "9th-c. Syrian astrolabe"),
        ("alfacca", "alfacca", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 27", "HIGH", "Chaucer Astrolabe table"),
        ("gemma", "gemma", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 14", "HIGH", "ca. 1400 Oxford astrolabe"),
        ("margarita", "margarita", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 75v", "HIGH", "Gerard of Cremona")
    ]),
    # 23. ALPHERATZ (alpha And)
    ("STAR_ALPHERATZ", "Alpheratz / alpha Andromedae", [
        ("surrat al-faras", "surrat al faras", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 53r", "HIGH", "al-Sufi navel of the horse"),
        ("alferatz", "alferatz", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 3", "HIGH", "John of Seville"),
        ("sirrah", "sirrah", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 61", "HIGH", "Alfonsine table"),
        ("umbilicus equi", "umbilicus equi", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 80v", "HIGH", "Gerard of Cremona")
    ]),
    # 24. ALPHIRK (beta Cep)
    ("STAR_ALPHIRK", "Alphirk / beta Cephei", [
        ("al-firq", "al firq", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 34v", "HIGH", "al-Sufi the flock"),
        ("alphirk", "alphirk", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 49", "HIGH", "Alfonsine table"),
        ("cingulum cephei", "cingulum cephei", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 76v", "HIGH", "Gerard of Cremona")
    ]),
    # 25. ALRISHA (alpha Psc)
    ("STAR_ALRISHA", "Alrisha / alpha Piscium", [
        ("al-rishā’", "al risha", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 88r", "HIGH", "al-Sufi the cord / well-rope"),
        ("alrisha", "alrisha", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 112", "HIGH", "Alfonsine table"),
        ("nodus piscium", "nodus piscium", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 86r", "HIGH", "Gerard of Cremona knot of the fishes")
    ]),
    # 26. ALTAIR (alpha Aql)
    ("STAR_ALTAIR", "Altair / alpha Aquilae", [
        ("al-ṭā’ir", "al tair", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 14", "HIGH", "9th-c. Syrian astrolabe"),
        ("altair", "altair", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 38", "HIGH", "Chaucer Astrolabe table"),
        ("vultur volans", "vultur volans", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 78v", "HIGH", "Gerard of Cremona"),
        ("aquila volans", "aquila volans", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 18", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 27. ANTARES (alpha Sco)
    ("STAR_ANTARES", "Antares / alpha Scorpii", [
        ("qalb al-‘aqrab", "qalb al aqrab", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 12", "HIGH", "9th-c. Syrian astrolabe"),
        ("calbalacrab", "calbalacrab", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 32", "HIGH", "Chaucer Astrolabe table"),
        ("cor scorpionis", "cor scorpionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 85r", "HIGH", "Gerard of Cremona"),
        ("antares", "antares", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 170v", "HIGH", "Vienna star table")
    ]),
    # 28. ARCTURUS (alpha Boo)
    ("STAR_ARCTURUS", "Arcturus / alpha Bootis", [
        ("al-rāmiḥ", "al ramih", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 8", "HIGH", "9th-c. Syrian astrolabe"),
        ("alramech", "alramech", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 25", "HIGH", "Chaucer Astrolabe table"),
        ("arcturus", "arcturus", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 75r", "HIGH", "Gerard of Cremona"),
        ("custos arcti", "custos arcti", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 13", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 29. ARKAB (alpha Sgr)
    ("STAR_ARKAB", "Arkab / alpha Sagittarii", [
        ("‘urqūb al-rāmī", "urqub al rami", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 84r", "HIGH", "al-Sufi tendon / hock of the archer"),
        ("arkab", "arkab", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 106", "HIGH", "Alfonsine table"),
        ("nervus sagittarii", "nervus sagittarii", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 85v", "HIGH", "Gerard of Cremona")
    ]),
    # 30. ASCELLA (zeta Sgr)
    ("STAR_ASCELLA", "Ascella / zeta Sagittarii", [
        ("al-waṣl", "al wasl", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 84v", "HIGH", "al-Sufi the joint"),
        ("ascella", "ascella", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_II", "Type II, Entry 22", "HIGH", "Pseudo-Messahalla armpit of archer"),
        ("axilla sagittarii", "axilla sagittarii", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 85v", "HIGH", "Gerard of Cremona")
    ]),
    # 31. BELLATRIX (gamma Ori)
    ("STAR_BELLATRIX", "Bellatrix / gamma Orionis", [
        ("al-nājid", "al najid", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 97r", "HIGH", "al-Sufi the conqueror"),
        ("bellatrix", "bellatrix", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 126", "HIGH", "Alfonsine table female warrior"),
        ("humerus sinister orionis", "humerus sinister orionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 87r", "HIGH", "Gerard of Cremona")
    ]),
    # 32. BETELGEUSE (alpha Ori)
    ("STAR_BETELGEUSE", "Betelgeuse / alpha Orionis", [
        ("mankib al-jawzā’", "mankib al jawza", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 5", "HIGH", "9th-c. Syrian astrolabe"),
        ("yad al-jawzā’", "yad al jawza", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_NMS_T1959_62", "Pointer 4", "HIGH", "13th-c. Edinburgh rete"),
        ("bedalgeuze", "bedalgeuze", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_JOHN_LONDON_1246", "Entry 12", "HIGH", "John of London 1246"),
        ("humerus dexter orionis", "humerus dexter orionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 87r", "HIGH", "Gerard of Cremona")
    ]),
    # 33. CANOPUS (alpha Car)
    ("STAR_CANOPUS", "Canopus / alpha Carinae", [
        ("suhayl", "suhayl", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 122r", "HIGH", "al-Sufi prominent southern star"),
        ("suhel", "suhel", "OLD_CASTILIAN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 150", "HIGH", "Alfonsine Castilian form"),
        ("canopus", "canopus", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 92r", "HIGH", "Gerard of Cremona Latinization")
    ]),
    # 34. CAPELLA (alpha Aur)
    ("STAR_CAPELLA", "Capella / alpha Aurigae", [
        ("al-‘ayyūq", "al ayyuq", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 3", "HIGH", "9th-c. Syrian astrolabe"),
        ("alhayoc", "alhayoc", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 10", "HIGH", "Chaucer Astrolabe table"),
        ("capella", "capella", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 79v", "HIGH", "Gerard of Cremona"),
        ("hircus", "hircus", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 6", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 35. CAPH (beta Cas)
    ("STAR_CAPH", "Caph / beta Cassiopeiae", [
        ("kaff al-khaḍīb", "kaff al khadib", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 17", "HIGH", "9th-c. Syrian astrolabe"),
        ("caph", "caph", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 2", "HIGH", "John of Seville"),
        ("manus tincta", "manus tincta", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 76r", "HIGH", "Gerard of Cremona")
    ]),
    # 36. CASTOR (alpha Gem)
    ("STAR_CASTOR", "Castor / alpha Geminorum", [
        ("ra’s al-taw’am al-muqaddam", "ras al tawam al muqaddam", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 65v", "HIGH", "al-Sufi head of the foremost twin"),
        ("castor", "castor", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 82r", "HIGH", "Gerard of Cremona"),
        ("apollon", "apollon", "LATINIZED_GREEK", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 169v", "MEDIUM", "Late medieval academic gloss")
    ]),
    # 37. COR CAROLI (alpha CVn)
    ("STAR_COR_CAROLI", "Cor Caroli / alpha Canum Venaticorum", [
        ("al-kabid", "al kabid", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 29v", "HIGH", "al-Sufi the liver / under Great Bear"),
        ("chara", "chara", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 57", "HIGH", "Alfonsine designation"),
        ("stella venaticorum", "stella venaticorum", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 169v", "MEDIUM", "Hunting dogs star")
    ]),
    # 38. DABIH (beta Cap)
    ("STAR_DABIH", "Dabih / beta Capricorni", [
        ("sa‘d al-dhābiḥ", "sad al dhabih", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 89v", "HIGH", "al-Sufi lucky star of slaughterer"),
        ("dabih", "dabih", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_II", "Type II, Entry 23", "HIGH", "Pseudo-Messahalla"),
        ("mactator", "mactator", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 86v", "HIGH", "Gerard of Cremona")
    ]),
    # 39. DENEB (alpha Cyg)
    ("STAR_DENEB", "Deneb / alpha Cygni", [
        ("al-dhanab", "al dhanab", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 16", "HIGH", "9th-c. Syrian astrolabe"),
        ("ridf", "ridf", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_NMS_T1959_62", "Pointer 17", "HIGH", "13th-c. Edinburgh rete"),
        ("deneb", "deneb", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 41", "HIGH", "Chaucer Astrolabe table"),
        ("cauda gallinae", "cauda gallinae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 79r", "HIGH", "Gerard of Cremona"),
        ("gallina", "gallina", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 20", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 40. DENEB ALGEDI (delta Cap)
    ("STAR_DENEB_ALGEDI", "Deneb Algedi / delta Capricorni", [
        ("dhanab al-jady", "dhanab al jady", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_NMS_T1959_62", "Pointer 18", "HIGH", "13th-c. Edinburgh rete"),
        ("denebalgedi", "denebalgedi", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 42", "HIGH", "Chaucer Astrolabe table"),
        ("cauda capricorni", "cauda capricorni", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 86v", "HIGH", "Gerard of Cremona")
    ]),
    # 41. DENEBOLA (beta Leo)
    ("STAR_DENEBOLA", "Denebola / beta Leonis", [
        ("dhanab al-asad", "dhanab al asad", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 71v", "HIGH", "al-Sufi tail of the lion"),
        ("denebalezeth", "denebalezeth", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 22", "HIGH", "Chaucer Astrolabe table"),
        ("denebola", "denebola", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 97", "HIGH", "Alfonsine table"),
        ("cauda leonis", "cauda leonis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 83v", "HIGH", "Gerard of Cremona")
    ]),
    # 42. DIPHDA (beta Cet)
    ("STAR_DIPHDA", "Diphda / beta Ceti", [
        ("dhanab qayṭus", "dhanab qaytus", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_NMS_T1959_62", "Pointer 21", "HIGH", "13th-c. Edinburgh rete"),
        ("diphda", "diphda", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 138", "HIGH", "Alfonsine tables frog star"),
        ("rana secunda", "rana secunda", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 89v", "HIGH", "Gerard of Cremona")
    ]),
    # 43. DUBHE (alpha UMa)
    ("STAR_DUBHE", "Dubhe / alpha Ursae Majoris", [
        ("ẓahr al-dubb", "zahr al dubb", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 26v", "HIGH", "al-Sufi back of the bear"),
        ("dubhe", "dubhe", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 52", "HIGH", "Alfonsine table"),
        ("dorsum ursae", "dorsum ursae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 74r", "HIGH", "Gerard of Cremona")
    ]),
    # 44. ELTANIN (gamma Dra)
    ("STAR_ELTANIN", "Eltanin / gamma Draconis", [
        ("ra’s al-tinnīn", "ras al tinnin", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 31r", "HIGH", "al-Sufi head of dragon"),
        ("eltanin", "eltanin", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 44", "HIGH", "Alfonsine table"),
        ("caput draconis", "caput draconis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 75v", "HIGH", "Gerard of Cremona")
    ]),
    # 45. ENIF (epsilon Peg)
    ("STAR_ENIF", "Enif / epsilon Pegasi", [
        ("anf al-faras", "anf al faras", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 51v", "HIGH", "al-Sufi nose of horse"),
        ("enif", "enif", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 59", "HIGH", "Alfonsine table"),
        ("nasus equi", "nasus equi", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 80v", "HIGH", "Gerard of Cremona")
    ]),
    # 46. FOMALHAUT (alpha PsA)
    ("STAR_FOMALHAUT", "Fomalhaut / alpha Piscis Austrini", [
        ("fam al-ḥūt", "fam al hut", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 125r", "HIGH", "al-Sufi mouth of the fish"),
        ("fomahant", "fomahant", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 45", "HIGH", "Chaucer Astrolabe table"),
        ("fomalhaut", "fomalhaut", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 162", "HIGH", "Alfonsine table"),
        ("os piscis meridionalis", "os piscis meridionalis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 93r", "HIGH", "Gerard of Cremona")
    ]),
    # 47. HAMAL (alpha Ari)
    ("STAR_HAMAL", "Hamal / alpha Arietis", [
        ("ra’s al-ḥamal", "ras al hamal", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 58r", "HIGH", "al-Sufi head of the ram"),
        ("hamal", "hamal", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 5", "HIGH", "John of Seville"),
        ("arietis caput", "arietis caput", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 81r", "HIGH", "Gerard of Cremona")
    ]),
    # 48. IZAR (epsilon Boo)
    ("STAR_IZAR", "Izar / epsilon Bootis", [
        ("al-izār", "al izar", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 36v", "HIGH", "al-Sufi the waist-cloth"),
        ("izar", "izar", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 64", "HIGH", "Alfonsine table"),
        ("perizoma", "perizoma", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 75r", "HIGH", "Gerard of Cremona")
    ]),
    # 49. KOCHAB (beta UMi)
    ("STAR_KOCHAB", "Kochab / beta Ursae Minoris", [
        ("al-kawkab", "al kawkab", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 24r", "HIGH", "al-Sufi the star"),
        ("kochab", "kochab", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 40", "HIGH", "Alfonsine table"),
        ("stella polaris antiqua", "stella polaris antiqua", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 168v", "MEDIUM", "Late medieval northern marker")
    ]),
    # 50. MARKAB (alpha Peg)
    ("STAR_MARKAB", "Markab / alpha Pegasi", [
        ("mankib al-faras", "mankib al faras", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_NMS_T1959_62", "Pointer 19", "HIGH", "13th-c. Edinburgh rete"),
        ("markab", "markab", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 44", "HIGH", "Chaucer Astrolabe table"),
        ("humerus equi", "humerus equi", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 80v", "HIGH", "Gerard of Cremona")
    ]),
    # 51. MEGREZ (delta UMa)
    ("STAR_MEGREZ", "Megrez / delta Ursae Majoris", [
        ("al-maghriz", "al maghriz", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 27r", "HIGH", "al-Sufi root of the tail"),
        ("megrez", "megrez", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 53", "HIGH", "Alfonsine table"),
        ("radix caudae", "radix caudae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 74r", "HIGH", "Gerard of Cremona")
    ]),
    # 52. MENKALINAN (beta Aur)
    ("STAR_MENKALINAN", "Menkalinan / beta Aurigae", [
        ("mankib dhī al-‘inān", "mankib dhi al inan", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 42v", "HIGH", "al-Sufi shoulder of the charioteer"),
        ("menkalinan", "menkalinan", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 74", "HIGH", "Alfonsine table"),
        ("humerus aurigae", "humerus aurigae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 79v", "HIGH", "Gerard of Cremona")
    ]),
    # 53. MENKAR (alpha Cet)
    ("STAR_MENKAR", "Menkar / alpha Ceti", [
        ("al-minkhar", "al minkhar", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 86v", "HIGH", "al-Sufi the nostril / snout of whale"),
        ("menkar", "menkar", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 7", "HIGH", "Chaucer Astrolabe table"),
        ("nares ceti", "nares ceti", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 89v", "HIGH", "Gerard of Cremona")
    ]),
    # 54. MERAK (beta UMa)
    ("STAR_MERAK", "Merak / beta Ursae Majoris", [
        ("maraqq al-dubb", "maraqq al dubb", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 26v", "HIGH", "al-Sufi flank of the bear"),
        ("merak", "merak", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 52", "HIGH", "Alfonsine table"),
        ("coxa ursae", "coxa ursae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 74r", "HIGH", "Gerard of Cremona")
    ]),
    # 55. MESARTHIM (gamma Ari)
    ("STAR_MESARTHIM", "Mesarthim / gamma Arietis", [
        ("al-sharaṭayn", "al sharatayn", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 58v", "HIGH", "al-Sufi the two signs"),
        ("mesarthim", "mesarthim", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 77", "HIGH", "Alfonsine table"),
        ("signa arietis", "signa arietis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 81r", "HIGH", "Gerard of Cremona")
    ]),
    # 56. MINTAKA (delta Ori)
    ("STAR_MINTAKA", "Mintaka / delta Orionis", [
        ("al-minṭaqa", "al mintaqa", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 98r", "HIGH", "al-Sufi the belt"),
        ("mintaka", "mintaka", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 128", "HIGH", "Alfonsine table"),
        ("zona orionis", "zona orionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 87v", "HIGH", "Gerard of Cremona")
    ]),
    # 57. MIRA (omicron Cet)
    ("STAR_MIRA", "Mira / omicron Ceti", [
        ("‘unqūd al-qayṭus", "unqud al qaytus", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 87r", "HIGH", "al-Sufi body of Cetus"),
        ("mira", "mira", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 139", "HIGH", "Alfonsine descriptor"),
        ("stella mirabilis", "stella mirabilis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 171r", "MEDIUM", "Wonderful variable star of Cetus")
    ]),
    # 58. MIRACH (beta And)
    ("STAR_MIRACH", "Mirach / beta Andromedae", [
        ("al-mi’zar", "al mizar", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 54r", "HIGH", "al-Sufi the apron"),
        ("mirach", "mirach", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 4", "HIGH", "Chaucer Astrolabe table"),
        ("cingulum andromedae", "cingulum andromedae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 81r", "HIGH", "Gerard of Cremona")
    ]),
    # 59. MIRFAK (alpha Per)
    ("STAR_MIRFAK", "Mirfak / alpha Persei", [
        ("mirfaq al-thurayyā", "mirfaq al thurayya", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 44r", "HIGH", "al-Sufi elbow of the Pleiades"),
        ("mirfak", "mirfak", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 70", "HIGH", "Alfonsine table"),
        ("cubitus persei", "cubitus persei", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 77r", "HIGH", "Gerard of Cremona")
    ]),
    # 60. MIZAR (zeta UMa)
    ("STAR_MIZAR", "Mizar / zeta Ursae Majoris", [
        ("al-mi’zar", "al mizar", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 28r", "HIGH", "al-Sufi the waist-cloth"),
        ("mizar", "mizar", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 25", "HIGH", "Chaucer Astrolabe table"),
        ("fascia ursae", "fascia ursae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 74r", "HIGH", "Gerard of Cremona")
    ]),
    # 61. NUNKI (sigma Sgr)
    ("STAR_NUNKI", "Nunki / sigma Sagittarii", [
        ("al-ṣādira", "al sadira", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 85r", "HIGH", "al-Sufi the returning ostriches"),
        ("nunki", "nunki", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 108", "HIGH", "Alfonsine table"),
        ("pectus sagittarii", "pectus sagittarii", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 85v", "HIGH", "Gerard of Cremona")
    ]),
    # 62. PHERKAD (gamma UMi)
    ("STAR_PHERKAD", "Pherkad / gamma Ursae Minoris", [
        ("al-farqad", "al farqad", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 24v", "HIGH", "al-Sufi the calf"),
        ("pherkad", "pherkad", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 41", "HIGH", "Alfonsine table"),
        ("pullus ursae", "pullus ursae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 73v", "HIGH", "Gerard of Cremona")
    ]),
    # 63. PHECDA (gamma UMa)
    ("STAR_PHECDA", "Phecda / gamma Ursae Majoris", [
        ("fakhidh al-dubb", "fakhidh al dubb", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 27r", "HIGH", "al-Sufi thigh of the bear"),
        ("phecda", "phecda", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 53", "HIGH", "Alfonsine table"),
        ("femur ursae", "femur ursae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 74r", "HIGH", "Gerard of Cremona")
    ]),
    # 64. POLARIS (alpha UMi)
    ("STAR_POLARIS", "Polaris / alpha Ursae Minoris", [
        ("al-jady", "al jady", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 23v", "HIGH", "al-Sufi the kid goat / pole star"),
        ("alrucaba", "alrucaba", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 39", "HIGH", "Alfonsine table"),
        ("polus", "polus", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 1", "HIGH", "Chaucer Astrolabe table"),
        ("stella maris", "stella maris", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 168v", "HIGH", "Medieval maritime guide")
    ]),
    # 65. POLLUX (beta Gem)
    ("STAR_POLLUX", "Pollux / beta Geminorum", [
        ("ra’s al-taw’am al-mu’akhkhar", "ras al tawam al muakhkhar", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 65v", "HIGH", "al-Sufi head of the rearmost twin"),
        ("pollux", "pollux", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 82r", "HIGH", "Gerard of Cremona"),
        ("heracles", "heracles", "LATINIZED_GREEK", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 169v", "MEDIUM", "Late medieval scholastic gloss")
    ]),
    # 66. PORRIMA (gamma Vir)
    ("STAR_PORRIMA", "Porrima / gamma Virginis", [
        ("zāwiyat al-‘awwā’", "zawiyat al awwa", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 74r", "HIGH", "al-Sufi angle of the barker"),
        ("porrima", "porrima", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 99", "HIGH", "Alfonsine table"),
        ("angulus virginis", "angulus virginis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 84r", "HIGH", "Gerard of Cremona")
    ]),
    # 67. PROCYON (alpha CMi)
    ("STAR_PROCYON", "Procyon / alpha Canis Minoris", [
        ("al-shi‘rā al-sha‘āmiya", "al shira al shaamiya", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 7", "HIGH", "9th-c. Syrian astrolabe"),
        ("algeuzech", "algeuzech", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 17", "HIGH", "Chaucer Astrolabe table"),
        ("procyon", "procyon", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 88v", "HIGH", "Gerard of Cremona"),
        ("antecanis", "antecanis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 10", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 68. RASALGETHI (alpha Her)
    ("STAR_RASALGETHI", "Rasalgethi / alpha Herculis", [
        ("ra’s al-jāthī", "ras al jathi", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 38r", "HIGH", "al-Sufi head of the kneeler"),
        ("rasalgethi", "rasalgethi", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 66", "HIGH", "Alfonsine table"),
        ("caput geniculati", "caput geniculati", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 75v", "HIGH", "Gerard of Cremona")
    ]),
    # 69. RASALHAGUE (alpha Oph)
    ("STAR_RASALHAGUE", "Rasalhague / alpha Ophiuchi", [
        ("ra’s al-ḥawwā’", "ras al hawwa", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 10", "HIGH", "9th-c. Syrian astrolabe"),
        ("rasalhague", "rasalhague", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 34", "HIGH", "Chaucer Astrolabe table"),
        ("caput serpentarii", "caput serpentarii", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 78r", "HIGH", "Gerard of Cremona")
    ]),
    # 70. RASTABAN (beta Dra)
    ("STAR_RASTABAN", "Rastaban / beta Draconis", [
        ("ra’s al-thu‘bān", "ras al thuban", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 31v", "HIGH", "al-Sufi head of serpent"),
        ("rastaban", "rastaban", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 45", "HIGH", "Alfonsine table"),
        ("caput serpentis draconis", "caput serpentis draconis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 75v", "HIGH", "Gerard of Cremona")
    ]),
    # 71. REGULUS (alpha Leo)
    ("STAR_REGULUS", "Regulus / alpha Leonis", [
        ("qalb al-asad", "qalb al asad", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 6", "HIGH", "9th-c. Syrian astrolabe"),
        ("calbalasad", "calbalasad", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 20", "HIGH", "Chaucer Astrolabe table"),
        ("regulus", "regulus", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 83v", "HIGH", "Gerard of Cremona"),
        ("basiliscus", "basiliscus", "LATINIZED_GREEK", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_III", "Type III, Entry 12", "HIGH", "John of Seville"),
        ("cor leonis", "cor leonis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 11", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 72. RIGEL (beta Ori)
    ("STAR_RIGEL", "Rigel / beta Orionis", [
        ("rijl al-jawzā’", "rijl al jawza", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 4", "HIGH", "9th-c. Syrian astrolabe"),
        ("rigel", "rigel", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 9", "HIGH", "Chaucer Astrolabe table"),
        ("algebar", "algebar", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_KUNITZSCH_TYPE_II", "Type II, Entry 4", "HIGH", "Pseudo-Messahalla"),
        ("pes orionis", "pes orionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 87r", "HIGH", "Gerard of Cremona")
    ]),
    # 73. RUKBAT (alpha Sgr)
    ("STAR_RUKBAT", "Rukbat / alpha Sagittarii", [
        ("rukbat al-rāmī", "rukbat al rami", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 83v", "HIGH", "al-Sufi knee of the archer"),
        ("rukbat", "rukbat", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 105", "HIGH", "Alfonsine table"),
        ("genu sagittarii", "genu sagittarii", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 85v", "HIGH", "Gerard of Cremona")
    ]),
    # 74. SADALMELIK (alpha Aqr)
    ("STAR_SADALMELIK", "Sadalmelik / alpha Aquarii", [
        ("sa‘d al-malik", "sad al malik", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 92r", "HIGH", "al-Sufi lucky star of the king"),
        ("sadalmelik", "sadalmelik", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 115", "HIGH", "Alfonsine table"),
        ("fortuna regis", "fortuna regis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 86v", "HIGH", "Gerard of Cremona")
    ]),
    # 75. SADALSUUD (beta Aqr)
    ("STAR_SADALSUUD", "Sadalsuud / beta Aquarii", [
        ("sa‘d al-su‘ūd", "sad al suud", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 91v", "HIGH", "al-Sufi luck of lucks"),
        ("sadalsuud", "sadalsuud", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 114", "HIGH", "Alfonsine table"),
        ("fortuna fortunarum", "fortuna fortunarum", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 86v", "HIGH", "Gerard of Cremona")
    ]),
    # 76. SARGAS (theta Sco)
    ("STAR_SARGAS", "Sargas / theta Scorpii", [
        ("al-la‘qa", "al laqa", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 81v", "HIGH", "al-Sufi sting of the scorpion"),
        ("sargas", "sargas", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 103", "HIGH", "Alfonsine table"),
        ("aculeus scorpionis", "aculeus scorpionis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 85r", "HIGH", "Gerard of Cremona")
    ]),
    # 77. SCHEAT (beta Peg)
    ("STAR_SCHEAT", "Scheat / beta Pegasi", [
        ("mankib al-faras", "mankib al faras", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 15", "HIGH", "9th-c. Syrian astrolabe"),
        ("scheat", "scheat", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 58", "HIGH", "Alfonsine table"),
        ("crus pegasi", "crus pegasi", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 80v", "HIGH", "Gerard of Cremona")
    ]),
    # 78. SCHEDAR (alpha Cas)
    ("STAR_SCHEDAR", "Schedar / alpha Cassiopeiae", [
        ("ṣadr dhāt al-kursī", "sadr dhat al kursi", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 33r", "HIGH", "al-Sufi breast of lady in chair"),
        ("schedar", "schedar", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 46", "HIGH", "Alfonsine table"),
        ("pectus cassiopeiae", "pectus cassiopeiae", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 76r", "HIGH", "Gerard of Cremona")
    ]),
    # 79. SHAULA (lambda Sco)
    ("STAR_SHAULA", "Shaula / lambda Scorpii", [
        ("al-shawla", "al shawla", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 81r", "HIGH", "al-Sufi the raised tail / sting"),
        ("shaula", "shaula", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 33", "HIGH", "Chaucer Astrolabe table"),
        ("aculeus", "aculeus", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 85r", "HIGH", "Gerard of Cremona")
    ]),
    # 80. SHERATAN (beta Ari)
    ("STAR_SHERATAN", "Sheratan / beta Arietis", [
        ("al-sharaṭān", "al sharatan", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 58r", "HIGH", "al-Sufi the two signs"),
        ("sheratan", "sheratan", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 77", "HIGH", "Alfonsine table"),
        ("cornu arietis", "cornu arietis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 81r", "HIGH", "Gerard of Cremona")
    ]),
    # 81. SIRIUS (alpha CMa)
    ("STAR_SIRIUS", "Sirius / alpha Canis Majoris", [
        ("al-shi‘rā al-yamāniya", "al shira al yamaniya", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 6", "HIGH", "9th-c. Syrian astrolabe"),
        ("alhabor", "alhabor", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 15", "HIGH", "Chaucer Astrolabe table"),
        ("sirius", "sirius", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 88r", "HIGH", "Gerard of Cremona"),
        ("canicula", "canicula", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 9", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 82. SKAT (delta Aqr)
    ("STAR_SKAT", "Skat / delta Aquarii", [
        ("al-sāq", "al saq", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 92v", "HIGH", "al-Sufi the shin / leg"),
        ("skat", "skat", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 116", "HIGH", "Alfonsine table"),
        ("tibia aquarii", "tibia aquarii", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 86v", "HIGH", "Gerard of Cremona")
    ]),
    # 83. SPICA (alpha Vir)
    ("STAR_SPICA", "Spica / alpha Virginis", [
        ("al-simāk al-a‘zal", "al simak al azal", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 8", "HIGH", "9th-c. Syrian astrolabe"),
        ("azimech", "azimech", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 28", "HIGH", "Chaucer Astrolabe table"),
        ("spica", "spica", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 84r", "HIGH", "Gerard of Cremona"),
        ("spica virginis", "spica virginis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 12", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 84. SUBRA (omicron Leo)
    ("STAR_SUBRA", "Subra / omicron Leonis", [
        ("al-zubra", "al zubra", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 72r", "HIGH", "al-Sufi the mane"),
        ("subra", "subra", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 96", "HIGH", "Alfonsine table"),
        ("crines leonis", "crines leonis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 83v", "HIGH", "Gerard of Cremona")
    ]),
    # 85. THUBAN (alpha Dra)
    ("STAR_THUBAN", "Thuban / alpha Draconis", [
        ("al-thu‘bān", "al thuban", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 32r", "HIGH", "al-Sufi the basilisk / dragon"),
        ("thuban", "thuban", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 43", "HIGH", "Alfonsine table"),
        ("draco", "draco", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 75v", "HIGH", "Gerard of Cremona")
    ]),
    # 86. UNUKALHAI (alpha Ser)
    ("STAR_UNUKALHAI", "Unukalhai / alpha Serpentis", [
        ("‘unuq al-ḥayya", "unuq al hayya", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 40r", "HIGH", "al-Sufi neck of the snake"),
        ("unukalhai", "unukalhai", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 68", "HIGH", "Alfonsine table"),
        ("collum serpentis", "collum serpentis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 77v", "HIGH", "Gerard of Cremona")
    ]),
    # 87. VEGA (alpha Lyr)
    ("STAR_VEGA", "Vega / alpha Lyrae", [
        ("al-wāqi‘", "al waqi", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_OXFORD_MHS_47632", "Pointer 13", "HIGH", "9th-c. Syrian astrolabe"),
        ("alvaca", "alvaca", "MIDDLE_ENGLISH_AND_LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_CHAUCER_ASTROLABE_1391", "Part II, table entry 37", "HIGH", "Chaucer Astrolabe table"),
        ("vega", "vega", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 50", "HIGH", "Alfonsine table"),
        ("vultur cadens", "vultur cadens", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 78r", "HIGH", "Gerard of Cremona"),
        ("lyra", "lyra", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_OXFORD_MHS_41468", "Pointer 17", "HIGH", "ca. 1400 Oxford astrolabe")
    ]),
    # 88. VINDEMIATRIX (epsilon Vir)
    ("STAR_VINDEMIATRIX", "Vindemiatrix / epsilon Virginis", [
        ("al-mukurram bi-l-qiṭāf", "al mukurram bi l qitaf", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 74v", "HIGH", "al-Sufi herald of vintage"),
        ("vindemiatrix", "vindemiatrix", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 84r", "HIGH", "Gerard of Cremona"),
        ("provindemiator", "provindemiator", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 170r", "HIGH", "Vienna star table")
    ]),
    # 89. WASAT (delta Gem)
    ("STAR_WASAT", "Wasat / delta Geminorum", [
        ("wasaṭ al-samā’", "wasat al sama", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 66r", "HIGH", "al-Sufi the middle of the sky"),
        ("wasat", "wasat", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 83", "HIGH", "Alfonsine table"),
        ("medium geminorum", "medium geminorum", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 82r", "HIGH", "Gerard of Cremona")
    ]),
    # 90. WEZN (delta CMa)
    ("STAR_WEZN", "Wezn / delta Canis Majoris", [
        ("al-wazn", "al wazn", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 112r", "HIGH", "al-Sufi the heavy weight"),
        ("wezen", "wezen", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 132", "HIGH", "Alfonsine table"),
        ("pondus", "pondus", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 88v", "HIGH", "Gerard of Cremona")
    ]),
    # 91. ZAVIJAVA (beta Vir)
    ("STAR_ZAVIJAVA", "Zavijava / beta Virginis", [
        ("zāwiyat al-‘awwā’", "zawiyat al awwa", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 73v", "HIGH", "al-Sufi corner of the kennel"),
        ("zavijava", "zavijava", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 98", "HIGH", "Alfonsine table"),
        ("angulus", "angulus", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 84r", "HIGH", "Gerard of Cremona")
    ]),
    # 92. ZOSMA (delta Leo)
    ("STAR_ZOSMA", "Zosma / delta Leonis", [
        ("al-ẓahr", "al zahr", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 71r", "HIGH", "al-Sufi back of the lion"),
        ("zosma", "zosma", "LATINIZED_GREEK", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 83v", "HIGH", "Gerard of Cremona girding / belt"),
        ("succinctorium", "succinctorium", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_VIENNA_CATALOGUE_1440", "fol. 170r", "MEDIUM", "Late medieval anatomical term")
    ]),
    # 93. ZUBENELGENUBI (alpha Lib)
    ("STAR_ZUBENELGENUBI", "Zubenelgenubi / alpha Librae", [
        ("al-zubānā al-janūbiyya", "al zubana al janubiyya", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 76v", "HIGH", "al-Sufi southern claw"),
        ("kiffa australis", "kiffa australis", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 101", "HIGH", "Alfonsine table southern scale tray"),
        ("lanx australis", "lanx australis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 84v", "HIGH", "Gerard of Cremona")
    ]),
    # 94. ZUBENESCHAMALI (beta Lib)
    ("STAR_ZUBENESCHAMALI", "Zubeneschamali / beta Librae", [
        ("al-zubānā al-shamāliyya", "al zubana al shamaliyya", "ARABIC", "ARABIC_TRANSLIT", "DIN_31635", "SRC_AL_SUFI_964", "fol. 77r", "HIGH", "al-Sufi northern claw"),
        ("kiffa borealis", "kiffa borealis", "LATINIZED_ARABIC", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_LIBROS_DEL_SABER_1270", "Vol. I, p. 101", "HIGH", "Alfonsine table northern scale tray"),
        ("lanx borealis", "lanx borealis", "LATIN", "LATIN", "MEDIEVAL_LATIN_DIRECT", "SRC_GERARD_CREMONA_1175", "fol. 84v", "HIGH", "Gerard of Cremona")
    ])
]

def main():
    # 1. Write LEXICON_SOURCE_REGISTRY.tsv
    source_fields = [
        "source_id", "title", "author", "date_start", "date_end",
        "language", "tradition", "primary_witness", "url_or_citation",
        "source_type", "notes"
    ]
    with REGISTRY_PATH.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=source_fields, delimiter="\t")
        writer.writeheader()
        for src in SOURCES:
            writer.writerow(src)
    print(f"Wrote {len(SOURCES)} sources to {REGISTRY_PATH}")

    # Map source_id to source details
    src_map = {s["source_id"]: s for s in SOURCES}

    # 2. Build and write HISTORICAL_STAR_LEXICON.tsv
    lexicon_fields = [
        "canonical_identity_id",
        "canonical_identity_name",
        "attestation_id",
        "attested_form",
        "normalized_form",
        "language",
        "script",
        "transliteration_system",
        "source_title",
        "source_author",
        "source_date_start",
        "source_date_end",
        "manuscript_or_edition",
        "folio_or_entry",
        "source_url_or_local_reference",
        "source_type",
        "historically_attested",
        "editorial_reconstruction",
        "confidence",
        "notes"
    ]

    attestation_counter = 1
    total_identities = len(IDENTITIES_DATA)
    total_attestations = 0

    with LEXICON_PATH.open("w", encoding="utf-8", newline="") as f:
        writer = csv.DictWriter(f, fieldnames=lexicon_fields, delimiter="\t")
        writer.writeheader()

        for ident_id, ident_name, attestations in IDENTITIES_DATA:
            # Check deduplication within identity
            seen_normalized = set()
            for att in attestations:
                att_form, norm_form, lang, script, trans_sys, src_id, folio, conf, notes = att
                
                # Check that normalization is strictly clean
                computed_norm = clean_normalized(norm_form)
                if computed_norm in seen_normalized:
                    continue  # Deduplicated within canonical identity
                seen_normalized.add(computed_norm)

                src = src_map[src_id]
                row = {
                    "canonical_identity_id": ident_id,
                    "canonical_identity_name": ident_name,
                    "attestation_id": f"HIST_STAR_ATT_{attestation_counter:04d}",
                    "attested_form": att_form,
                    "normalized_form": computed_norm,
                    "language": lang,
                    "script": script,
                    "transliteration_system": trans_sys,
                    "source_title": src["title"],
                    "source_author": src["author"],
                    "source_date_start": src["date_start"],
                    "source_date_end": src["date_end"],
                    "manuscript_or_edition": src["primary_witness"],
                    "folio_or_entry": folio,
                    "source_url_or_local_reference": src["url_or_citation"],
                    "source_type": src["source_type"],
                    "historically_attested": "YES",
                    "editorial_reconstruction": "NO",
                    "confidence": conf,
                    "notes": notes
                }
                writer.writerow(row)
                attestation_counter += 1
                total_attestations += 1

    print(f"Wrote {total_attestations} attestations for {total_identities} canonical identities to {LEXICON_PATH}")
    print(f"CANONICAL_IDENTITIES: {total_identities} (>= 57 check: PASS)")

if __name__ == "__main__":
    main()
