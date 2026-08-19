# Canonical Generic Corpus Format v1

`canonical_generic_corpus_version: 1`

This format is the stable input contract for generic corpus analysis.
The goal is to make text preparation explicit, deterministic, and
inspectable.

## Required properties

- valid UTF-8;
- no U+FFFD replacement rune;
- no NUL;
- no forbidden C0/C1 controls, except line breaks and tab before normalization;
- canonical whitespace;
- one token = one whitespace-delimited sequence;
- punctuation and symbol runes are separators, not content glue;
- deterministic preprocessing;
- no preprocessing metadata/header mixed into the scientific corpus.

## Supported input encodings

- `utf-8`
- `windows-1251`
- `koi8-r`
- `cp866`
- `auto` is supported, but conservative: it only accepts unambiguous UTF-8.

## Punctuation policy

Punctuation and symbol characters are replaced by spaces, not removed.
This prevents token gluing:

- `word,word` → `word word`
- `2—3` → `2 3`
- `well-known` → `well known`
- `"text"` → `text`
- `(abc)` → `abc`

## Numbers

Numbers are preserved as tokens. The tool does not normalize them into
placeholders and does not interpret them semantically.

## Case policy

- `preserve`
- `lower` `lower` is the preferred/default policy for new generic corpora.

Lowercasing uses Unicode-aware, locale-independent case folding from the Go
runtime.

## Line policy

- `preserve`: keep logical line breaks, normalize whitespace inside each line;
- `reflow`: flatten the corpus into a pure token stream and serialize it with
  a fixed representational layout.

The policy must be explicit because some scientific stages use line
boundaries as structure.

## Whitespace policy

Tabs, multiple spaces, CRLF, NBSP, and Unicode space separators are
normalized. Hidden control characters fail closed.

## Reproducibility

The same input bytes, encoding, policies, and tool version must produce
byte-identical output corpus and prepare manifest.

The prepare manifest includes hashes, policies, and version metadata, but no
timestamp in the deterministic identity portion.
