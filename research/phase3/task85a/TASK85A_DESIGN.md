# Task85a — executable completion design

**Status:** `GRAMMAR_EXPERIMENT_CONTRACT_V1.1_FROZEN`.

## Scope and provenance

Task85-v1 remains the scientific parent. Task86's blocked result is a
methodological observation: ten operational definitions were absent and no
experiment ran. This layer supplies those definitions without changing the
central question, G1 firewall, M0–M5 classes, corpora, split, metric families,
complexity sum, lexicon/exception accounting, transcription symmetry,
semantics firewall, mechanism firewall, or HELDOUT policy.

The chain is `Task85 -> Task86 BLOCKED -> Task85a -> Task86R`. Exact parent and
blocked-result hashes are frozen in the V1.1 version manifest.

## Normative sources

`G1_EXECUTABLE_CONTRACT.json` is the numeric source of truth. The algorithm,
candidate, calibration, stability, transcription, negative-control, PM5, PM6,
adequacy, ladder, and seed documents are normative expansions. If prose and
JSON disagree, preflight fails; Task86R may not choose an interpretation.

The finite space contains 84 candidates: M0 3, M1 18, M2 9, M3 9, M4 27, M5
18. M3 and M4 share topology induction; M3 uses uniform enabled transitions,
whereas M4 estimates their probabilities. M5 component discovery and all
exception creation are deterministic and charged under the inherited
`StructureCost + LexiconCost + ExceptionCost` rule.

## Why these choices

- Additive smoothing and PPM-C are conventional, closed-form estimators.
- Small logarithmically spaced grids cover weak through unit smoothing without
  target feedback.
- Greedy JS state merging is a single enumerable DFA/PFSA induction procedure;
  shared topology isolates probability estimation in M4.
- Substrings of length one through three plus deterministic dynamic programming
  are the simplest bounded component inventory compatible with frozen M5.
- Caps 64/128/256, 84 candidates, 16 MFC populations, 32 maximum replicates,
  and 64 glyphs per generated token bound compute before target fitting.
- Robust medians, nearest-rank 0.95 quantiles, Theil-Sen slopes, and top-label
  ECE are conventional finite-sample functionals.

These are `OPERATIONAL_COMPLETION`, except serialization, hashing, and sorted
iteration, which are `IMPLEMENTATION_DETAIL`. No `SCIENTIFIC_EXTENSION` is
required.

## Feasibility and firewall

One full transcription fit grid is 84 fits; both transcriptions require 168.
Calibration is 3 generators x 16 populations x 84 candidates = 4,032 pipeline
fits. Maximum structural generation is 84 x 2 transcriptions x 3 scales x 32
replicates = 16,128 bounded corpora. Each induction has a 100,000-operation cap.
These static counts establish feasibility without a Voynich runtime pilot.

`validate_contract.py` has no corpus tokenizer or model-fitting entrypoint. Its
only corpus operation is streaming SHA-256. Task85a tooling therefore cannot
fit M0–M5 on a Voynich corpus.

## Freeze rule

The V1.1 sentinel is valid only while the validator passes all nine preflight
flags, all ten blocker rows are `RESOLVED`, every inherited hash matches, all
84 rows equal the deterministic JSON expansion, and the artifact manifest is
consistent. Otherwise the status is `GRAMMAR_CONTRACT_V1_1_UNRESOLVED`.
