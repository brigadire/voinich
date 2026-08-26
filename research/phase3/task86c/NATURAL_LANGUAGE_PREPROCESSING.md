# Frozen natural-language preprocessing

Decode UTF-8, retain Unicode letter runs, lowercase, remove Project Gutenberg envelope where markers exist, then take deterministic circular contiguous occurrence samples. Split each sample 60/20/20 without shuffling. No language-specific stemming, normalization, abbreviation expansion, or vocabulary filtering. Latin sources are already ordinary expanded Latin.
