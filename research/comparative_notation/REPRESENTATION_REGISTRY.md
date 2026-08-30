# Frozen representation registry v1

| ID | Class | Token | Symbols | Source-observed line | Preserved dimensions |
|---|---|---|---|---|---|
| LATIN-DIPLOMATIC | C01/C02 | diplomatic word | graphemes | manuscript line if encoded | abbreviation marks |
| LATIN-EXPANDED | C01/C02 | expanded word | graphemes | same aligned source line | expansion alignment |
| SHORTHAND-DIPLOMATIC | C03 | source-delimited sign group | signs | manuscript line if encoded | tradition, hand |
| SHORTHAND-EXPANDED | C03 | aligned expanded word | graphemes | same aligned line | pair ID |
| CIPHER-SIGNS | C04/C05 | source-delimited cipher group | cipher signs | source line | mechanism/cipher alphabet |
| CIPHER-PLAINTEXT | C04/C05 | aligned plaintext word | graphemes | same line only if alignment proves it | pair ID |
| MUSIC-R1 | C06 | one notated event | event components | physical system only | event, voice, staff, system, simultaneity |
| MUSIC-R2 | C06 | melodic interval event | signed interval components | physical system only | interval, voice, staff, system |
| MUSIC-R3 | C06 | pitch×duration event | pitch and duration components | physical system only | pitch, duration, voice, staff, system |
| TAB-R1 | C07 | simultaneity/chord group | ordered course–fret components | physical system | tradition, course, fret, rhythm, simultaneity |
| TAB-R2 | C07 | course event | fret/rhythm components | physical system | same simultaneity ID retained |
| NUMERIC-RECORD | C08 | source field/number | sign/digit sequence | row only when source-observed | field, unit, column |
| TABLE-CELL | C09 | non-empty cell | cell transcription symbols | source row | row, column, header/body role |
| SYNTHETIC-TOKEN | C10 | generated token | generated symbols | generator-declared line | generator and seed |

All registered representations are analyzed independently. No representation
may be selected after observing its VM distance. C06 and C07 fixtures document
the exact source-event-to-USC mapping under their corpus directories.
