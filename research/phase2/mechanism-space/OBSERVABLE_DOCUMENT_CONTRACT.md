# Observable document contract

An OBSERVABLE_DOCUMENT contains a glyph/symbol stream plus explicit token, line, and unit/page boundary statuses: NOT_DEFINED, INHERITED_FROM_INPUT, or GENERATED_BY_MECHANISM. Task81 does not invent absent hierarchy. F01/F08/F11 use undefined token/line boundaries and generated units; F12 generates cue-token/unit boundaries and leaves lines undefined.

Only externally visible symbols or observations may be serialized. InternalMemoryState, association maps, conventions, paths, history, plaintext input, and implementation state are excluded. Metadata is descriptive and cannot encode any hidden carrier. This is accepted by a future frozen F2 reader as a generic symbol stream with missing-data boundary semantics; Task81 performs no F2 computation or comparison.
