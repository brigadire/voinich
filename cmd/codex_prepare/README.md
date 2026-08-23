# codex_prepare

`codex_prepare` turns a raw text document into a canonical generic corpus.
It is a preparation utility, not a scientific stage.

Default behavior:

- input encoding: `utf-8` unless `-encoding` is set explicitly;
- case policy: `lower`;
- line policy: `preserve`;
- punctuation and symbol runes become separators, never deleted silently;
- whitespace is normalized to canonical ASCII spaces;
- output is always UTF-8.

Commands:

```bash
codex_prepare prepare -input raw.txt -output prepared.txt -encoding windows-1251
codex_prepare inspect -input raw.txt -encoding windows-1251 -json
codex_prepare check -input prepared.txt -json
```

`prepare` writes:

- `prepared.txt`
- `prepared.txt.prepare.json`

The manifest records the canonical corpus version, the exact policies used,
input/output hashes, and deterministic output metadata.

See `CANONICAL_CORPUS_FORMAT.md` for the exact format contract.
