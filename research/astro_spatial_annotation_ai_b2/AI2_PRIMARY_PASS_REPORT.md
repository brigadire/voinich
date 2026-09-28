# ANNOTATOR_B_AI_2 — Primary Pass Report

Independent, blind visual annotation of 8 astronomical panel crops.
Worked only from the supplied `crops/*.jpg` images; no transcription, prior
annotation, or identification information was consulted.

## Totals

| Quantity | Count |
|---|---|
| Panels annotated | 8 |
| Objects (`AI2_OBJECTS.tsv`) | 544 |
| Physical text labels (`AI2_LABELS.tsv`) | 206 |
| Label↔object relations (`AI2_LABEL_OBJECT_RELATIONS.tsv`) | 931 |

Both stages were frozen in order: `AI2_STAGE1_MANIFEST.json` (detection) was
written and hashed before any relation work began; `AI2_STAGE2_MANIFEST.json`
covers the relation table. The Stage 1 hashes still match the files on disk, so
no object or label was added, removed or moved during Stage 2.

## Object counts by class and panel

| class | f67r1 | f67r2 | f67v1 | f68r1 | f68r2 | f68r3 | f68v1 | f68v2 | total |
|---|---|---|---|---|---|---|---|---|---|
| STAR_OBJECT | 21 | 0 | 42 | 29 | 59 | 77 | 61 | 41 | 330 |
| CIRCLE | 4 | 2 | 1 | 1 | 0 | 3 | 2 | 1 | 14 |
| RADIAL_LINE | 24 | 12 | 0 | 0 | 0 | 6 | 16 | 8 | 66 |
| SECTOR | 24 | 12 | 0 | 0 | 0 | 0 | 16 | 8 | 60 |
| CENTRAL_OBJECT | 1 | 1 | 1 | 0 | 0 | 1 | 1 | 1 | 6 |
| MOON_OR_DISC_OBJECT | 1 | 12 | 0 | 2 | 2 | 0 | 0 | 0 | 17 |
| TEXT_ARC | 3 | 1 | 1 | 1 | 2 | 2 | 2 | 1 | 13 |
| OTHER_DIAGRAM_OBJECT | 3 | 5 | 21 | 1 | 2 | 0 | 5 | 1 | 38 |
| **objects, all classes** | **81** | **45** | **66** | **34** | **65** | **89** | **103** | **61** | **544** |
| labels | 21 | 39 | 30 | 37 | 33 | 14 | 10 | 22 | 206 |

Confidence distribution — objects: HIGH 339, MEDIUM 163, LOW 34, AMBIGUOUS 8.
Labels: HIGH 143, MEDIUM 61, LOW 2. Relations: HIGH 458, MEDIUM 434, LOW 39.

Relation types: ADJACENT_TO 358, NEAREST_OBJECT 181, BETWEEN_OBJECTS 140,
RING_LABEL 70, INSIDE_OBJECT 60, SECTOR_LABEL 58, ON_OBJECT 36, UNASSIGNED 28.

## Method notes (needed to read the numbers correctly)

- Positions were read off coordinate-gridded crops of the supplied images. For
  the four panels built as ringed diagrams (f67r1, f67r2, f68v1, f68v2) the
  annulus was also re-sampled into a polar "unrolled" strip, in which radial
  lines appear as straight verticals, rings as horizontals, and radial text as
  upright columns; several object positions in those panels are therefore given
  as polar measurements converted back to image pixels. The assumed diagram
  centre used for each such panel is stated in the object notes.
- Star censuses in the dense fields (f68r2, f68r3, f68v1, f68v2, f67v1) were
  made by scanning for the pale ochre core that every drawn star mark carries,
  then checking every candidate by eye on magnified crops and adding marks the
  scan had missed. Candidates that turned out to be text strokes or parchment
  stains were discarded.
- Label granularity: a line of running text outside a diagram is one label; a
  radial text run inside a diagram is one label; a full ring of writing is one
  label; a block of contiguous short lines filling one wedge (f67r2) is one
  label. This is stated in each label's notes. Any comparison of label counts
  against another annotator must allow for this choice.
- `orientation_angle` for radial labels is the bearing of the outward radius in
  degrees measured clockwise from image-horizontal (image y points down); 0 is
  used for horizontal text and for ring text whose orientation rotates along the
  arc, where a single value is not meaningful.
- Stage 2 relations were derived purely from the frozen geometry — bounding-box
  gap, centre distance, containment in a roundel or a wedge, and coincidence with
  a text arc. No astronomical or semantic reasoning entered. Where more than one
  reading was spatially plausible, all of them are recorded as separate candidate
  rows for the same label, which is why most labels carry 2–8 rows.
- `UNASSIGNED` rows still name an `object_id` (the geometrically nearest object),
  because the relation table requires a known object reference; the `evidence`
  field on those rows states explicitly that no unambiguous association is
  discernible and that the named object is only the nearest neighbour. All 28
  UNASSIGNED rows are lines of running text outside a diagram, or a margin
  numeral.
- Known cosmetic defect in the frozen Stage 1 notes: for f67r2, f68v1 and f68v2
  the two bearings quoted in SECTOR notes are printed with mixed sign
  conventions (one as a negative angle, the other reduced modulo 360), e.g.
  "bearings -148 and 247 deg" for the wedge from −148° to −113°. The bounding
  boxes themselves are correct. This was found after Stage 1 was frozen and was
  deliberately left uncorrected rather than break the freeze.

## What I was least confident about, and why

1. **f68r3 (largest panel, 77 stars).** The densest field, and the only panel
   where the wedge dividers are faint curved pencil lines that are visible only
   under strong contrast enhancement. I recorded 6 dividers as RADIAL_LINE at
   LOW confidence and recorded **no** SECTOR objects for this panel, because I
   could not trace any wedge's full boundary — I would have had to invent the
   closing arcs. The tight group of about seven very small star marks near
   (665–830, 1265–1400) is recorded as seven LOW-confidence stars; the marks
   overlap and both the count and the boxes there are genuinely uncertain.
   Seven further stars in this panel have paler cores than their neighbours and
   were found only by eye (MEDIUM).
2. **f67r1 radial lines and sectors.** The 24 spokes are very faint pencil
   lines. Reading them off the unrolled annulus gave clearly *unequal* spacing
   (13–16° for most, with two wider gaps at roughly 42–79° and 155–184° where I
   could resolve only one line). I recorded exactly what I could see and marked
   the doubtful spokes LOW; the sectors that inherit a LOW boundary are LOW too.
   I did **not** insert extra spokes to even out the spacing. The same panel's
   rosette is filled with a dense field of very small star-like specks which I
   did not attempt to enumerate — they are described inside the rosette object's
   notes instead of being invented as individual objects.
3. **f67v1 outer band.** The band is a continuous ring of text groups
   alternating with small boxed tally-stroke glyphs. It runs into the page fold
   on the right-hand side, so my enumeration (13 text groups, 19 boxed glyphs)
   is explicitly partial and all of it is MEDIUM. Three star fragments visible
   beyond the fold at the right crop edge belong to the adjoining leaf and are
   recorded at LOW confidence as fragments.
4. **f67r2 roundels and outer band.** Eleven of the twelve roundels are
   unambiguous; the twelfth, at (835–894, 1053–1122), is much fainter than the
   rest and is LOW. Its position happens to fall in the one wide angular gap in
   the roundel ring, which is why I looked there — but it was recorded because it
   is visible, not because a gap existed. The twelve radial lines include one
   45°-wide gap (between bearings 0° and 45°) where I could not resolve a line;
   I left the gap rather than inserting one. The band's hatch-stroke clusters
   are recorded as a single ring object because I could not delimit them
   individually.
5. **Circle radii generally.** Every diagram is drawn on a slightly curved page,
   so the concentric rings are measurably non-concentric and non-circular: on
   f68v1 the same ring measures radius ~847 px in one direction and ~932 px in
   another. CIRCLE bounding boxes are therefore approximations, marked MEDIUM or
   LOW, with the measured spread recorded in the notes. I did not force any of
   them onto an idealised circle.
6. **Faint stains vs. deliberate marks.** Eight objects are AMBIGUOUS: pale
   blue-green or grey patches on f67r1, f67r2, f68r2 and f68v1 that have no
   drawn outline. They may be accidental pigment or parchment staining rather
   than intended marks; they are recorded because they are visible, and flagged
   because their status is unclear.
7. **Edge and overlap cases.** A handful of stars sit at a crop edge or overlap
   a neighbour so that only part of the outline is separable (f68v2 at
   (55–150, 1565–1660) and (1290–1390, 1635–1720); f68r2 at (85–180, 1560–1650)).
   These carry LOW/MEDIUM confidence and their boxes are the visible extent, not
   a reconstruction.

## Anti-symmetry discipline actually applied

No object was added to make counts match or spacing regular. Concretely: f67r1's
star sectors contain 9 pairs and 3 *single* stars, and the three singles were
left single; f67r1, f67r2, f68v1 and f68v2 all have visibly unequal wedge widths,
recorded as measured; f67r2 has one 45° gap and f67r1 two wide gaps in the spoke
sequence, left as gaps; the number of stars per wedge on f68v1 varies 7–9 and on
f68v2 1–9, recorded as counted; no damaged or partly visible mark was completed.
