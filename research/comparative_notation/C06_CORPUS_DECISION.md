# C06 corpus decision — mensural music

Decision date: 2026-08-30. Status: `INCLUDED`.

## Frozen source and subset

C06 uses every MusicXML file directly inside the `Josquin (secure)/XML` and
`La Rue (secure)/XML` directories of JLSDD revision
`64bc7461dc564163cab84894bda263e4501b1a0a`. The path rule yields 79 documents
(33 Josquin, 46 La Rue). The upstream README says 33+44; the actual revision
contains 46 La Rue XML files. The path rule, not the prose count, is frozen so
no unexplained files are selected or dropped.

The source repository is GPL-3.0 and includes the license in the frozen raw
subset. Local analysis and derived USC production are accepted under those
terms. The source was chosen before comparative analysis because it provides
secure-attribution Renaissance polyphony, complete event pitch/duration,
voices, and encoded page/system breaks.

## Representation decision

- `MUSIC-R1`: each notated note or rest is an event. Pitch class/register,
  exact rational duration, explicit accidentals, and ties are ordered symbol
  components. Chord members share an onset.
- `MUSIC-R2`: signed semitone intervals are computed only within the same
  source document/part/voice/staff. Rests reset continuity. The source voice
  becomes an observed USC section, preventing generic sequence metrics from
  creating cross-voice adjacency.
- `MUSIC-R3`: each pitched event is a compound of spelled pitch and exact
  rational duration; rests are not pitch-duration events.

All three representations are mandatory. Encoded page and system breaks are
preserved; measures are traceability attributes, not invented physical lines.
The adapter fails closed on missing divisions/duration/pitch, zero-duration
events, negative cursor motion, or absent initial page/system layout.

## Limitation

JLSDD is a modern MusicXML transcription of Renaissance works, not a diplomatic
encoding of original mensural glyph shapes. C06 therefore tests symbolic
musical event/interval/pitch-duration structure, not paleographic glyph form.
This limitation was accepted without changing the frozen global protocol and
must accompany any later interpretation.
