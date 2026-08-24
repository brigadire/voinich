# IT2a frozen preprocessing contract

The pre-Task83 contract is reconstructed from `TASK79C_DESIGN.md`, `cmd/ivtff-x7-extract`, `internal/metadatavalidation`, `codex_prepare`, and the preparation sidecar.

- **Transcription selection:** Takeshi Takahashi IT2a, raw `data/IT2a-n.txt`, in document order. The choice was preregistered independently of agreement with ZL.
- **Locus filtering:** parse page/locus records; ignore blank, `#`, declaration, and non-locus lines; emit each locus whose normalized alignment text is nonempty. No folio/content selection.
- **Alternatives:** for `[first:alternative]`, retain the first top-level branch. Braces inside a branch do not split it.
- **Uncertainty:** remove braces but retain their content; apostrophe and `?` are token boundaries. Inline `@NNN;` is split to ordinary numeric content.
- **Comments/metadata:** remove inline variables, `<!…>` comments, `<%>` and `<$>` controls. `<->` and `<~>` are boundaries. Page/locus metadata determines ordering/alignment but is not emitted.
- **Text normalization:** `. , = - ' ? @ ;` become boundaries; `strings.Fields` collapses whitespace; one ASCII-space-joined record is emitted per nonempty locus, with LF after every record.
- **Canonical preparation:** UTF-8, case preserved, Unicode punctuation/symbols and whitespace map to ASCII spaces, repeated spaces collapse, hidden controls/invalid UTF-8 are rejected, line boundaries preserved.
- **Tokenization/order/encoding:** whitespace tokens, locus and token order preserved, UTF-8, LF newlines.

The current parser/preparer and their versions at recorded Task79c HEAD `d568e54…` are identical. The small extractor command was first committed with Task79c at `6f185579…`; its normalization calls the already-existing audited parser. Thus the sidecar's recorded HEAD and the committed producer source are both disclosed instead of pretending the working-tree command already existed at `d568e54…`.
