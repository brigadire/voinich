# Astronomical spatial annotation ontology

This layer records visible marks and geometry only. Names such as `STAR_OBJECT` and
`MOON_OR_DISC_OBJECT` describe appearance, not an astronomical identification.
No translation, M1/M2/M3 assignment, dictionary candidate, frequency or hapax flag
may be consulted while creating a primary annotation.

| Class | Operational definition |
|---|---|
| `STAR_OBJECT` | One separately drawn star-like mark. |
| `LABEL` | One physically bounded text label. Labels are stored in the label table rather than duplicated as objects. |
| `CIRCLE` | A visible circle or ring stroke/envelope. |
| `RADIAL_LINE` | A visible line extending approximately radially from a diagram centre. |
| `SECTOR` | A bounded sector of a circular diagram. |
| `CENTRAL_OBJECT` | The central visible element of a diagram. |
| `MOON_OR_DISC_OBJECT` | A disc or crescent-like visible element, without identity claim. |
| `TEXT_ARC` | Text visibly following a circular arc. |
| `OTHER_DIAGRAM_OBJECT` | A visible element outside the preceding classes. |

Confidence is one of `HIGH`, `MEDIUM`, `LOW`, `AMBIGUOUS`. Primary passes are
immutable (`ANNOTATOR_A`, `ANNOTATOR_B`); adjudication is a separate
`ADJUDICATED` record. Coordinates use the panel crop as origin, with image x to
the right and y down. Normalized coordinates divide x by crop width and y by
crop height. Polar coordinates use theta zero at image top and increase
clockwise: `theta = atan2(x-cx, -(y-cy)) mod 2*pi`.

`NEAREST_OBJECT` is a geometric statement only and never asserts semantic
identity. Derived rings, orders, neighbours and clusters must remain separate
from primary observations.
