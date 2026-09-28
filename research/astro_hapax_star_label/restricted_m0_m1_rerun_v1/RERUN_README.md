# Restricted M0-R/M1-R rerun

This package invokes the unchanged frozen M0 (640-rule grid) and M1 (beam width 64, global 1–2 EVA-unit mapping) implementations. Only the input adapter and output directory differ. Direction A trains on all f68r1 STAR LABEL tokens and holds out all f68r2 tokens; direction B reverses the pages. M1 null controls use 10,000 deterministic replicates per each of the three original families. The combined run is intentionally not used for model selection. No M2 code is involved.
