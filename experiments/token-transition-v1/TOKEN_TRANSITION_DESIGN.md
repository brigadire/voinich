# Task63 frozen design

Adjacency is within-line neighboring token occurrences; cross-line pairs are
reported separately. Glyphs use internal/evaglyph and d=1 uses Task60's
Levenshtein/classifier. Non-adjacent controls match token length pairs and are
sampled with seed 63001. Global and within-line shuffles use seed 63002.
Separation is k=1..10. Operation bins are COPY, SUBSTITUTION/INSERTION/
DELETION × BEGIN/MIDDLE/END. Discovery/replication is contiguous 60/40 line
blocks. Model selection is transition cross entropy on validation only; no
Task58–62 fingerprint metric is an objective.
