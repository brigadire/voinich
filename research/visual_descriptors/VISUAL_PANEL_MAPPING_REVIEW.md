# Visual panel-to-canvas adjudication

## Scope and clean-room method

This registry covers exactly the 42 numbered IVTFF panel units in `page_identification.tsv`. I used only:

- `/tmp/vd-cleanroom/common/page_identification.tsv` to obtain the target IDs, physical-leaf IDs, and candidate Yale IIIF identifiers;
- `/tmp/vd-cleanroom/common/image_index.tsv` to locate the Yale JPEG corpus and compare adjacent captures;
- the JPEGs in `/tmp/voinich-images` for all actual assignments and crop boundaries; and
- public physical-layout/page-nomenclature descriptions solely to confirm the convention that foldout suffixes count outward from the binding and reverse left-to-right order on the verso.

No repository files, descriptor-pass directories, transcriptions, or prior descriptor values were read. Textual content was not interpreted. Visible folio numerals and quire marks were used only as physical page marks, alongside seams, gutters, tears, parchment edges, artwork continuity, and recto/verso reversal.

The crop syntax is IIIF `pct:x,y,w,h`. Boundaries were placed on visible gutters or dominant fold lines. Full-panel isolated canvases use `pct:0,0,100,100`. Decimal coordinates are visual estimates, normally within about one percentage point of the fold center; they are locators rather than conservation-grade geometric annotations.

## External nomenclature/layout sources

- [Voynich MS description and foldout numbering](https://www.voynich.nu/descr.html): numbered panels begin at the binding and count outward; recto panels are described left-to-right.
- [Quire 9 layout](https://www.voynich.nu/q09/index.html), [Quire 10 layout](https://www.voynich.nu/q10/index.html), and [Quire 11 layout](https://www.voynich.nu/q11/index.html): confirm the physical ordering of the astronomical foldouts.
- [Quire 14 layout](https://www.voynich.nu/q14/index.html): distinguishes the three-panel and complementary two-panel photographs of the complex f85/f86 sheet.
- [Quire 15 layout](https://www.voynich.nu/q15/index.html), [Quire 17 layout](https://www.voynich.nu/q17/index.html), and [Quire 19 layout](https://www.voynich.nu/q19/index.html): confirm which neighboring panels appear together in unfolded captures.

These sources were used as page-nomenclature/physical-layout references, not for transcription or semantic analysis.

## Visual evidence by leaf group

- **f67:** canvases 1006194 and 1006195 are matched two-panel sides. Their single central fold and reversed verso order produce r1/r2 and v2/v1.
- **f68:** canvas 1006196 has two dominant folds forming three unequal panels. Canvas 1006197 mirrors those widths in reverse (broad-narrow-medium), independently confirming v3/v2/v1. A weaker line inside the broad verso panel was rejected because it has no matching panel boundary on the recto.
- **f70:** canvas 1006199 is a three-page opening: f69v at left, followed by two f70 recto panels. The verso panels are not two crops of 1006200: artwork and parchment edges show 1006200 and 1006201 are separate complete panels.
- **f72:** canvas 1006203 visibly contains f71v plus three outward panels. Canvas 1006204 contains two broad verso panels separated at the dominant fold; 1006205 is the isolated remaining panel.
- **f85/f86:** 1006228 is one complete text-only f85 panel. Canvas 1006229 has three distinct panels, while 1006230 is the complementary two-panel view. This physical arrangement resolves f85r2, f86v4, f86v6 on the former and f86v5, f86v3 on the latter.
- **f89/f90:** 1006233 contains the preceding f88v page plus two f89 recto panels. Canvas 1006234 is an isolated marked f89 verso panel; the complementary f89 verso panel is the left page of 1006235. The middle and right pages of 1006235 are visibly different complete herbal panels and therefore f90r1/r2. Canvases 1006236 and 1006237 are the two isolated reverse-side panels.
- **f95:** 1006241 contains f94v plus two distinct f95 recto panels. The two verso panels are isolated in 1006242 and 1006243, not halves of one canvas.
- **f102:** 1006251 contains a partial f101v page followed by two f102 recto panels. The reverse-side units are isolated complete canvases 1006252 and 1006253.

## Uncertainties

All 42 unit-to-canvas assignments are resolved. The only crop-boundary confidence reduced to `MEDIUM` is f72v3/f72v2 on canvas 1006204: overlapping parchment at the far left and uneven perspective make the precise percentage position less crisp, though the panel identities and their order are secure. Internal creases inside broad outer panels (notably f68, f70, f89, and f102) were retained within the numbered panel because the corresponding physical-layout convention and artwork continuity show they do not create an additional numbered unit.

## Reverse-order second review

I performed a separate visual self-check from f102v1 backward to f67r1, deliberately reversing the first-pass sequence:

1. **f102v1 → f95r1:** confirmed isolated verso pairs (1006252/1006253 and 1006242/1006243), then checked that the preceding wide canvases contain an unrelated left page plus two recto panels. Fold boundaries do not bisect the isolated canvases.
2. **f90v1 → f89r1:** matched the folio-marked isolated outer panels to their unmarked complements, then re-opened 1006235 and verified three physical page regions rather than treating 1006236 or 1006237 as two crops. Rechecked 1006233/1006234 by artwork continuity and edge geometry.
3. **f86v3 → f85r1:** compared 1006230 against 1006229 in reverse order. The two-panel view supplies v5/v3; the three-panel view supplies r2/v4/v6; 1006228 remains a complete independent panel.
4. **f72v1 → f70r1:** verified that 1006205 and 1006201 are isolated panels, then walked outward-to-inward across 1006204, 1006203, 1006200, and 1006199. Recto numbering increases away from the binding and verso presentation reverses it.
5. **f68v1 → f67r1:** mirrored the unequal f68 panel widths across the recto/verso pair and rejected the extra weak crease inside f68v3. Finally, the two f67 canvases showed one unambiguous central fold and exact order reversal.

The reverse review reproduced all 42 canvas assignments and all fold choices. It also caught and preserved the crucial use of adjacent isolated/unfolded canvases instead of forcing multiple numbered units onto the same folded-state photograph.
