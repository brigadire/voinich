# Clean Object Ontology for Independent Visual Annotation

This layer records visible marks and diagram geometry only. Class names describe visual appearance, not semantic astronomical identification.

| Class | Operational Definition |
|---|---|
| `STAR_OBJECT` | One separately drawn star-like mark. |
| `LABEL` | One physically bounded text label. Stored in label table with anonymous ID. |
| `CIRCLE` | A visible circle or ring stroke/envelope. |
| `RADIAL_LINE` | A visible line extending approximately radially from a diagram centre. |
| `SECTOR` | A bounded sector of a circular diagram. |
| `CENTRAL_OBJECT` | The central visible element of a diagram. |
| `MOON_OR_DISC_OBJECT` | A disc or crescent-like visible element, without identity claim. |
| `TEXT_ARC` | Text visibly following a circular arc. |
| `OTHER_DIAGRAM_OBJECT` | A visible element outside the preceding classes. |

Confidence values: `HIGH`, `MEDIUM`, `LOW`, `AMBIGUOUS`.

Relation Types:
- `NEAREST_OBJECT`: Geometric nearest neighbor.
- `ADJACENT_TO`: Positioned directly beside the object.
- `INSIDE_OBJECT`: Positioned inside a geometric enclosure.
- `ON_OBJECT`: Superimposed or aligned directly along the object boundary/stroke.
- `BETWEEN_OBJECTS`: Situated between two reference objects.
- `SECTOR_LABEL`: Text label designating an entire diagram sector.
- `RING_LABEL`: Text label associated with a circular band or ring.
- `UNASSIGNED`: No unambiguous association discernible.
