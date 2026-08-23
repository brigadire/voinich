# Task80 — Fontana operation algebra and historical freeze

`FONTANA_OPERATION_ALGEBRA_V1` defines a shared **typed** operational basis,
not a shared encoder/decoder. `FONTANA_MODELS_FROZEN_V1` freezes F01's
invariant core and tested profile plus the bounded cores of F08, F11, and F12.
F07/F10 are retained only as sensitivity references; F02–F06/F09 are
excluded.

## Operational result

The common basis is `select`, `rotate`, `align`, `place`, `traverse`,
`index`, `signal`, `associate`, `repeat`, and `compose`. It supports four
useful mechanism classes: literal external storage (F01/F08), indexed opaque
cue (F11), cyclic-state selector (reference-only F07), and temporal
associative cue (F12). F10 is a profile-dependent aligned-band reference.
No class is inferred merely from a similar mathematical state transition.

External state and prior knowledge remain separate. F01 and F08 offer
literal recovery only with their conventions; F11 exposes a cue at a known
index; F12 emits a cue and requires learned internal memory. Thus the
universal basis is partial and the universal *mechanism* is not supported.

## Composition and boundary

The JSON registry distinguishes `C-FONTANA` attested compositions from
`G-ALLOWED` type-valid counterfactuals and `FORBIDDEN` constructions.
Counterfactual rotation-plus-index and association-over-storage are retained
solely to test the boundary for later `M-RESTRICTED` work. A signal without
an emitting temporal state is type-invalid. The graph in
`PROVENANCE_GRAPH.dot` traces every frozen behavior through source/package,
operation, and composition.

## Information, recovery, and equivalence

F01 is `EXACT_WITH_CONVENTION`; F08 is
`EXACT_WITH_CONVENTION` for its declared path; F11 is `CUE_ONLY` unless the
cue convention supplies meaning; F12 is `CUE_ONLY` without learned memory.
F01's ring rotations obey modular composition only on a declared independent
ring; F10 R1 supplies the counterexample. Traversal is order-sensitive, and
cue emission alone is not retrieval. No pair has full structural,
operational, observational, retrieval, and mnemonic equivalence; the
relevant equivalence is explicitly local to the registry's stated model and
profile.

## Validation

`go test ./internal/fontanaalgebra` validates schema completeness, historical
primitive provenance, composition classes, synthetic-fixture labelling, and
all freeze checksums. `go run ./research/phase2/fontana/task80-analyze`
performs the same published-artifact verification. Existing task76/task78
tests remain the behavioral reproduction tests named in each frozen entry.

## Verdicts

| verdict | value | basis / limitation |
|---|---|---|
| TASK76_RESULTS_ACCEPTED | SUPPORTED | F01 source-disciplined dossier and complete operational evidence |
| TASK78_RESULTS_ACCEPTED | PARTIALLY_SUPPORTED | Four bounded cores accepted; profile claims remain bounded |
| VALIDATED_FONTANA_SET_DEFINED | SUPPORTED | F01, F08, F11, F12 cores explicitly listed |
| COMMON_OPERATIONAL_BASIS_SUPPORTED | SUPPORTED | Typed primitives span the frozen cores |
| SINGLE_UNIVERSAL_BASIS_SUPPORTED | NOT_SUPPORTED | No common literal encoder/decoder or common retrieval relation |
| MECHANISM_CLASS_HETEROGENEITY_SUPPORTED | SUPPORTED | Literal, indexed, cyclic, and associative cue roles remain distinct |
| PRIMITIVE_OPERATIONS_IDENTIFIED | SUPPORTED | Versioned registry with source/model/test annotations |
| COMPOSITION_RULES_VALIDATED | SUPPORTED | Attested, type-valid, and invalid classes are machine-checked |
| INFORMATION_FLOW_FORMALIZED | SUPPORTED | External state, convention, path, and internal memory are separated |
| HISTORICAL_GENERALIZED_BOUNDARY_DEFINED | SUPPORTED | `C-FONTANA` is distinct from `G-ALLOWED` and M statuses |
| FONTANA_OPERATION_ALGEBRA_READY | SUPPORTED | V1 schema, graph, laws, fixtures, and validator supplied |
| FONTANA_MODELS_FROZEN | FROZEN | Immutable V1 manifest/checksums; only explicitly bounded cores/profile |
| READY_FOR_MNEMONIC_MECHANISM_SPACE | SUPPORTED | Interface prevents generalized combinations being historical claims |

Open questions are the un-frozen profile parameters: F01 capacity/alphabet;
F08 boundary/capacity/mobility; F11 cue mapping; F12 drive, calibration, and
human retention; and the fundamental F07/F10 transition rules.
