# Sources for the expanded astronomical term corpus

## Inclusion policy

The corpus records physical inscriptions or edited medieval text forms dated
from the eighth through fifteenth centuries. Each row is an attestation; the
`term_id` is the concept key. Editorial transliteration is stored in
`normalized_form`, while the instrument spelling/script is retained in
`attested_form`. Modern star names are identifiers only and are never search
forms. Unidentified pointers on the instruments were excluded.

## Sources

### `NMS_T1959_62_RETE`

National Museums Scotland, “A 1000-year-old star catcher,” table of the 21
pointers on the thirteenth-century rete of astrolabe T.1959.62. The museum
dates the main instrument to 1026–1027 and the replacement rete to the
thirteenth century.

https://www.nms.ac.uk/discover-catalogue/a-1000-year-old-star-catcher

### `OXFORD_MHS_47632`

History of Science Museum, Oxford, Astrolabe Catalogue, inventory 47632,
late ninth century, Syria. The catalogue transcribes the Arabic zodiac and all
17 named star pointers directly from the rete. Transliteration in this corpus
is a conservative reading of that displayed Arabic, retained alongside the
instrument text.

https://www.mhs.ox.ac.uk/astrolabe/catalogue/reteReport/Astrolabe_ID=131.html

### `OXFORD_MHS_41468`

History of Science Museum, Oxford, Astrolabe Catalogue, inventory 41468,
Paris (?), ca. 1400. Capitalization and bracketed damaged readings in
`attested_form` reproduce the catalogue. Rows without a modern identification
(`ALHAIOT`, `ALDI[raan]`, `LIBEDENEB`) were excluded rather than guessed.

https://www.mhs.ox.ac.uk/astrolabe/catalogue/reteReport/Astrolabe_ID=237.html

### `KUNITZSCH_1987`

Paul Kunitzsch, “An unknown Arabic source for star names,” *History of
Oriental Astronomy*, IAU Colloquium 91 (1987), pp. 155–163,
doi:10.1017/S0252921100105986. Only the explicitly edited ca. 1246 form
*bedalgeuze* inherited from D0 is used.

### `WALTERS_W73`

Walters Art Museum MS W.73 (ca. 1025), cosmographical diagrams. This is the
manuscript source inherited from frozen D0 for the seven luminaries/planets.

https://www.thedigitalwalters.org/Data/WaltersManuscripts/html/W73/

### `ORDO_PLANETARUM`

Pseudo-Bede, *Ordo planetarum*, anonymous eighth-century Latin text, Migne,
*Patrologia Latina* 90, cols. 943D–946A. The forms are inherited unchanged
from D0; inflections remain variants of seven concepts.

## Evidence boundaries

The Oxford catalogue's modern identifications are curatorial mediation, so
rows are `MEDIUM` where the inscription is generic (`VRSA`, `al-asad`) or
damaged (`ALRAM[et]`). No form later than 1500 is admitted. Zodiac rows are
retained for corpus completeness and audit statistics but cannot match the
frozen M1 sample, whose only classes are `STAR` and `PLANET_MOON`.
