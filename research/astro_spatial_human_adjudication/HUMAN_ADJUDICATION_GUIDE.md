# Human spatial adjudication — quick guide

You need no transcription system, language knowledge, or astronomy knowledge. Judge only visible ink and geometry.

## Before you start

1. Use the assigned task only: **STAR**, **LABEL**, or (later) **RELATION**.
2. In H1, do not seek source/model information. Zoom in and out before deciding.
3. Never delete a supplied rectangle. Record a decision; move/resize it only for `MODIFY`.

## STAR task

- `ACCEPT`: one visually real star-like mark; the rectangle is adequate.
- `REJECT`: text stroke, stain, decoration of another class, fragment of a neighbour, or detection artefact.
- `MODIFY`: the mark exists, but resize/move the rectangle or change `STAR_OBJECT` to `OTHER_OBJECT`.
- `UNCERTAIN`: the image does not support a confident decision. Leave it unresolved.

Treat a faint mark consistently with the frozen calibration rules. Do not infer an astronomical identity. Include a clipped edge mark only if the visible part is independently object-like. Overlapping stars are separate only when distinct bodies/boundaries are visible.

## LABEL task

A LABEL is one physical text unit, not a transcription. Do not read or type its content.

Local and radial text runs are shown as **rotated rectangles**. Align the long axis with the visible writing direction; resize it to cover the physical run, not the empty sector around it. A full circular/ring text candidate is shown as an **ellipse**: the ellipse denotes the circular writing path, not the contents of its centre. If a full-ring candidate should instead be several physical runs, mark `SPLIT` rather than forcing one rectangular enclosure.

- Use `ACCEPT`, `REJECT`, `MODIFY`, or `UNCERTAIN` as above.
- Use `SPLIT` / `split_required=true` when one rectangle contains multiple physical units.
- Use `MERGE` / `merge_required=true` when neighbouring candidates form one physical unit. Put partner candidate IDs in notes.

Follow the frozen calibration choice for word, text run, radial run, ring text, multi-line block, and cartouche contents.

## RELATION task (only after STAR and LABEL freeze)

Select only the observed spatial relation between frozen candidates. Do not name or identify the depicted object. If no relation is visually defensible, use `UNASSIGNED`; if ambiguous, leave it unresolved.

## Confidence and notes

Confidence describes the reviewer's certainty in the selected decision, not candidate priority, source quality, or mark darkness:

- `HIGH`: visually unambiguous; another careful inspection is very unlikely to change the decision or corrected geometry.
- `MEDIUM`: the decision is more likely than the alternatives, but boundaries, class, or LABEL granularity have a meaningful ambiguity.
- `LOW`: tentative decision based on weak/ambiguous visual evidence; a second reviewer could reasonably disagree.

For `UNCERTAIN`, confidence qualifies the uncertainty decision itself: `HIGH` means the reviewer is confident that the available image is intrinsically insufficient to resolve the case; `LOW` means even the choice to leave it unresolved is tentative. A faint mark can still receive `HIGH` if its existence/class is visually unambiguous, and a dark mark can receive `LOW` if its class or boundaries are ambiguous.

Notes should describe visible evidence only. Never add transcription, model guesses, astronomical names, EVA/Stolfi/ZL3b terms, or M1/M2/M3 results.
