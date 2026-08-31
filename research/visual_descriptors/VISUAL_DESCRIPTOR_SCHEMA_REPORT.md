# Visual descriptor schema report

The visual annotation layer is frozen at schema 1.0.0 with 15 descriptors. The Yale MS408 IIIF manifest (OID 2002046; 213 canvases) and 227 neutral page units are fixed. All 42 multi-panel mappings are `VERIFIED`; an independent second mapping review reports 42/42 agreement.

The final table covers 227/227 units: 195 fully annotated and 32 partially annotated because the image is present but non-diagnostic for at least one descriptor. No image-missing or unresolvable-panel units remain, and panel-map pending count is zero.

The predeclared 24-unit pilot was rerun independently after the schema revision. Six descriptors failed the fixed reliability gate and were removed: `COMP_ILLUSTRATED_AREA`, `COMP_DOMINANT_LAYOUT`, `COMP_IMAGE_INTERLEAVED`, `GEO_CONNECTED_NETWORK`, `GEO_REPEATED_MODULES`, and `TXT_EMBEDDED_IMAGE`. The retained 15 descriptors passed the gate. A separate 46-unit audit pass is preserved and incorporated in the adjudication table.

Blindness is certified by the clean-room protocol and manifests. No textual association analysis was performed, and no textual result was used for descriptor selection. Level C visual-context testing is authorized but not executed here.

```text
VISUAL_FEATURE_SCHEMA_FROZEN=true
VISUAL_PAGE_DESCRIPTORS_READY=true
VISUAL_DESCRIPTOR_COVERAGE=227/227
BLINDNESS_CONSTRAINT_SATISFIED=true
RELIABILITY_GATE_PASSED=true
PANEL_MAP_PENDING_COUNT=0
TEXTUAL_RESULTS_USED_FOR_DESCRIPTOR_SELECTION=false
TEXTUAL_ASSOCIATION_ANALYSIS_PERFORMED=false
LEVEL_C_VISUAL_CONTEXT_TEST_AUTHORIZED=true
LEVEL_C_VISUAL_CONTEXT_TEST_EXECUTED=false
```
