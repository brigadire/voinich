# Independent Rozanova–Temerev analyzer

Build/run a single corpus:

```sh
go run ./independent/rozanova-temerev -corpus PATH -name NAME \
  -glyph-mode text -output-dir OUT
```

Use `-glyph-mode voynich` for the canonical EVA corpus and `-glyph-mode opaque`
for transformed `xNNNNNN` corpora. Batch comparison accepts positional
`NAME=PATH` arguments:

```sh
go run ./independent/rozanova-temerev compare -output-dir OUT \
  Doyle=data_test/pg2097-2.txt \
  Voynich=data_work/ZL3b-x7.canonical.txt
```

`results.tsv`/`comparison.tsv` contain the primary raw and shuffle-corrected
MI columns. JSON records the corpus hash and frozen protocol parameters.
