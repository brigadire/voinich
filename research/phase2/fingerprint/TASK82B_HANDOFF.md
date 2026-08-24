# Task82b handoff — independent notation controls available after Task79c

Task79c used the corpora below only to validate the Fingerprint v2
comparison *interface* (Gates B–D: does the distance/Pareto machinery
produce stable, interpretable differences between independently known
corpus classes). It did **not** run the Task82b/Task83 hypothesis
comparison against Voynich itself. This document records what is available
so Task82b does not have to re-derive acquisition/provenance/normalization
from scratch. Full field-by-field provenance is in
`CONTROL_PROVENANCE.tsv`; this is the summary a future Task82b run needs to
decide what to reuse.

## Corpus inventory

| Corpus ID | Class | Local path | Tokens | License | Alignment to plaintext |
|---|---|---|---|---|---|
| `bdd-koeln-edd-c-119` | Scribal abbreviation (Latin, 11th c.) | `data_test/bdd-prepared/bdd-koeln-edd-c-119.prepared.txt` (raw TEI in `data_test/bdd-tei/koeln-edd-c-119/`) | 19,052 | CC BY 4.0 (stated in-file) | Yes, in the source TEI (`<abbr>`/`<expan>` pairs), but the pairing was **not extracted or used** by Task79c — only the abbreviated stream. A future Task82b wanting `ΔF = F(abbr) - F(expansion)` or expansion-ambiguity metrics (`F2_V2_1_CANDIDATES.md`) would need to extend `cmd/tei-abbr-extract` to also emit the aligned `<expan>` stream with the same line/position keys, which it currently discards. |
| `msdos2.0` | Table/procedural (x86 assembly) | `data_test/msdos2.0.txt` (unmodified, pre-existing from Phase 1) | ~112,162 raw words | Not re-verified (predates Task79c) | Not applicable — not an abbreviation of a natural-language plaintext. |
| `doyle-sign-of-four` | Natural prose (English) | `data_test/pg2097-2.txt` | per Task79's canonical run | Project Gutenberg | Not applicable. |

Four other BDD manuscript witnesses have transcribed chapters but were not
acquired for Task79c (only `koeln-edd-c-119` was, per the fixed
smallest-witness tie-break in `TASK79C_DESIGN.md` §3): `bamberg-sb-c-6`,
`frankfurt-ub-b-50`, `vatican-bav-pal-lat-585`, `vatican-bav-pal-lat-586`,
all under the same repository/commit/license. If Task82b needs a larger or
multi-witness shorthand sample (e.g. to test cross-witness/cross-scribe
stability the way Task79c tests cross-transcription stability for
Voynich), these are the next deterministically-available witnesses, listed
in the same GitHub tree snapshot (`burchards-dekret-digital/website`
commit `29f9cb1c34cc9ee3c50e75a6e3e99cfa4a2bc362`).

## Normalization pipeline (for exact reproduction)

- BDD: `go run ./cmd/tei-abbr-extract -diplomatic-output DIPL -prepared-output PREP -manifest-output MANIFEST INPUT.xml [INPUT.xml ...]`
  (abbr-branch only; teiHeader/note/fw/label/toc excluded; `<g>` marks →
  reserved Glagolitic placeholders; then
  `internal/corpusprep.Prepare(encoding=utf-8, case=lower,
  line-policy=preserve)`).
- msdos/Doyle: used as-is, `glyph_mode: natural` in a
  `fingerprint-v2-analyze` `controls` entry, no `ivtff_path`.

## Available metrics for these corpora

None of the three has IVTFF metadata, so **only the metadata-free
"common core" F2 families** (`FINGERPRINT_V2_DISTANCE.md` §2.5) are
computable — the same restriction Task79's canonical run already applied
to the Doyle control (21 eligible contrasts there). Page/locus/folio/
hierarchy/cross-scale families (`LC*`, `PF*`, `HR*`, `2DL1`, `cs*`) are
`NOT_APPLICABLE` for all three, not `NOT_SUPPORTED` — there is no
manuscript-page structure to measure, not a failed measurement.

## Known limitations for Task82b to inherit

1. **BDD sample size and scope.** One manuscript witness, 5 of its
   13 books, ~19k tokens — a bounded, deterministic sample, not the full
   Burchard's Decretum corpus (126 witnesses exist; 5 have any
   transcription at all as of the commit used). Task82b should decide
   deliberately, before looking at any comparison result, whether a larger
   BDD sample changes its own power/precision needs.
2. **No expansion alignment used.** See the corpus inventory table above —
   `<expan>` is present in the source but was not extracted. Any Task82b
   metric that needs abbreviation→expansion pairs needs new extraction
   code, not just a new metric.
3. **msdos2.0 license/provenance was not re-verified by Task79c.** It was
   already resident in this repository for an unrelated Phase 1 purpose;
   Task79c reused it as-is and did not re-derive its acquisition terms.
   Task82b should treat its redistribution status as unresolved, the same
   as every other `data_test/` input except the two whitelisted Gutenberg
   files.
4. **Single-witness, single-hand scope.** BDD's `koeln-edd-c-119` sample
   is (per its manifest description) one scribal hand's work; it is not a
   claim about scribal-abbreviation notation in general, only about this
   one witness. Cross-witness generalization is exactly what the four
   unacquired witnesses above would let Task82b test, if needed.
5. **Task79c's own Gate B/C results (distance, Pareto, size-matched
   replicate dispersion) are reported in `TASK79C_REPORT.md` and the raw
   JSON under `experiments/fingerprint-v2-task79c-v1/`.** Task82b should
   read those before re-deriving comparable numbers, but must not treat
   Task79c's interface-validation exercise as itself a hypothesis test
   about Voynich against these controls (parent task §21: Task79c's
   purpose is narrower than that).
