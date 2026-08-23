# Orientation corpus transformation

`codex_orientation` creates a new canonical corpus for a directional
experiment. It is not a pipeline stage and never modifies the source corpus.

```sh
go run ./cmd/codex_orientation \
  -input data_work/ZL3b-x7.txt \
  -output data_work/ZL3b-x7.token-reverse.txt \
  -mode TOKEN_REVERSE
```

The command requires a corpus accepted by the shared `codex_prepare`
canonical validator, rejects an existing output unless `-force` is explicit,
and writes `<output>.transform.json`. The manifest records input/output paths
and SHA256 values, N, V, line counts, the exact mode, and deterministic
orientation flags. `pipeline-orchestrate manifest -generic-corpus -corpus
<output>` reads a valid sibling transformation manifest when present and
records its source hash, transformed hash, type, mode, and manifest hash as
optional experiment provenance. The transformed corpus SHA256 remains the
scientific input identity.

## Modes and boundaries

- `TOKEN_REVERSE`: reverse whitespace-delimited token order independently in
  every logical line. Token bytes and the complete token multiset are
  unchanged.
- `GLYPH_REVERSE`: preserve token order and reverse Unicode code points inside
  each token.
- `FULL_REVERSE`: apply both operations per line.

Logical line order, including supported blank lines, is preserved. No tokens
move between lines; page, folio, metadata, and original whitespace
normalization are not interpreted or changed. Output remains canonical because
canonical lines use exactly one ASCII space between tokens and a final newline.

Glyph reversal uses Go Unicode runes, never UTF-8 bytes. The supported EVA
Voynich corpus is ASCII. Canonical generic input may contain combining marks;
rune reversal can separate a base character from its combining sequence.
Therefore this tool documents code-point rather than grapheme-cluster
semantics and should not be used to make grapheme-sensitive claims about such
corpora without a separate, versioned grapheme-mode experiment.

## Correctness and workflow

The transformer validates output invariants after each operation. All modes
preserve token count, logical line count, and token count per line.
`TOKEN_REVERSE` additionally preserves the exact token-frequency map and V.
`GLYPH_REVERSE` and `FULL_REVERSE` preserve the token-length multiset, though
reversed forms can merge vocabulary entries. Every mode is an involution:
applying it twice restores the exact canonical input bytes.

Create the three independent inputs, inspect their manifests, then use the
existing generic-corpus orchestration workflow with otherwise frozen analysis
parameters. Do not add this operation to the scientific pipeline and do not
alter the frozen Voynich baseline.

## Cross-line limitation

Per-line reversal does not reverse a transition from the last token of line N
to the first token of line N+1. Pipeline stages that flatten non-empty lines
into a continuous stream therefore have boundary transitions whose directional
relationship is not a simple global reversal. Interpret those outputs with the
metric audit below rather than as a pure reversed-stream control.
