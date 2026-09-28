# Calibration guide

Calibration is mandatory before production. Review all assigned STAR and LABEL candidates on `f68r1`, `f68r3`, and `f68v2` using the H1 blind files.

Record ambiguities in `HUMAN_CALIBRATION_REPORT.md`, especially:

- minimum visible evidence for faint star-like marks;
- partial edge marks and overlaps;
- boundary between a star and decoration;
- LABEL unit: word, physical run, radial run, ring, multi-line block, or cartouche contents;
- whether a continuous ring is one LABEL or must be split into local arc runs;
- alignment of rotated rectangles to local text baselines (ellipses denote complete ring paths);
- repeatable bbox boundary conventions.

Clarify operations, not ontology, and do not optimize rules for agreement with any source. Keep `UNCERTAIN` cases unresolved. If a clarification substantially changes the protocol, increment the protocol version and repeat calibration; never rewrite historical decisions in place.

Production begins only after the calibration report is completed, signed, and a new frozen protocol digest is recorded.
