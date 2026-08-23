# Interface for mnemonic-mechanism space

`FONTANA_OPERATION_ALGEBRA_V1.json` is the machine-readable interface. A
consumer must preserve operation `version`, input/output types, preconditions,
invariants, information effects, provenance, and state serialization supplied
by its selected historical core.

| status | permitted meaning |
|---|---|
| `F-EXACT` | exact frozen historical model |
| `F-PROFILE` | frozen reconstruction profile |
| `F-COMPOSITION` | composition attested in a frozen Fontana model |
| `M-RESTRICTED` | new, type-valid combination of listed primitives |
| `M-EXTENDED` | adds a primitive or rule |
| `GENERIC_CONTROL` | synthetic validation fixture |

A generator must reject missing prior knowledge, an operation with mismatched
types, a failed precondition, or a composition marked `FORBIDDEN`. It must
never promote `G-ALLOWED`, `M-RESTRICTED`, or `M-EXTENDED` to a Fontana
claim. The external state, geometry/path, history, convention, and internal
memory are separate carriers:

`RetrievedItem = R(ExternalState, geometry/path, history, convention, InternalMemoryState, context)`.

Literal readout is exact only where the frozen retrieval relation says so;
cue-based retrieval remains `CUE_ONLY` without a learned association.
