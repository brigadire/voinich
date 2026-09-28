# Synthetic recovery specification

Source vocabularies contain 31 STAR concepts with 1–2 variants and lengths 5–7. Hidden S0–S5 transformations are generated from the frozen M1 representation; S6 contextual mappings are out of model space. Noise corrupts target labels at 0/10/20/30%; dictionary incompleteness removes 0/10/20/30/40% concepts. TRAIN/HELD_OUT sizes are 20/6, 40/10, and 80/20. Search uses the frozen 640 pipelines, grapheme inventory, beam search and scoring. Random controls rerun complete search.
