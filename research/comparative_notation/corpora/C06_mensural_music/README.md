# C06 — Mensural musical notation

The production source is the frozen secure MusicXML subset of the JLSDD
Renaissance polyphony corpus. Every work remains a source-identified document
and sequence boundaries never cross documents. Run MUSIC-R1
(notated event), MUSIC-R2 (within-voice relative interval), and MUSIC-R3
(pitch×duration) independently. A physical line is a source-encoded system,
not a measure or arbitrary event count. Preserve voice, staff, system,
simultaneity, pitch, duration, accidental, and rest/note metadata. JLSDD is a
modern symbolic transcription and does not encode original mensural glyph
shape; that limitation is not repaired by inference.

Dimensionality: two-dimensional, simultaneous/polyphonic, system-structured,
and hierarchical. Linearization follows source event order within voice and a
frozen voice/staff tie-break; it does not imply that a note is a word.
