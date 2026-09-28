# Transcription policy

The project uses ZL3b/EVA. The primary corpus file is the deterministic-refreeze-supported `data_work/ZL3b-x7.canonical.txt`, SHA-256 `f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`. Token occurrence IDs, folio, locus/line and positions come from the registered occurrence metadata whose three target pages are reproduced byte-identically in the Task83b multirun audit.

Candidate display expands frozen composite atoms through the already implemented bijection `C→cth`, `K→ckh`, `P→cph`, `F→cfh`, `N→iin`, `A→ain`, `H→ch`, `S→sh`, `E→ee`, `I→in`. This expansion is for readability; `canonical_token_key` and absolute occurrence ID determine identity. Case is preserved in the key. Separators are upstream token boundaries. Comments/metadata are absent from occurrences. Literal uncertain/unreadable markers remain literal. Alternative readings are not adjudicated automatically; the upstream first-reading convention remains fixed, while human ambiguity may be recorded without inventing a form.

No page-local frequency, hapax status or dictionary information enters candidate construction or the review UI.
