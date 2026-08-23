# inverse-transposition-search

Controlled Task54 inverse search for the rectangular columnar transposition
implemented by Task46. The search is blind: it reads only the transformed
canonical corpus, enumerates the pre-registered `natural` and `keyed` column
orders, and ranks exact inverse reconstructions using the frozen
`structural-v2` objective described in
[`INVERSE_TRANSPOSITION_METRIC_AUDIT.md`](../INVERSE_TRANSPOSITION_METRIC_AUDIT.md).

```bash
go run ./inverse-transposition-search \
  -input data_test/transformed/doyle__transposition__w008__natural__seed001.txt \
  -output-dir /tmp/task54-search
```

The output directory contains `search-manifest.json`, `search-report.md`, and
the top-ranked candidate corpora. It is not a pipeline stage and does not
read an original corpus during search. After ranking, exact oracle recovery
is checked explicitly with:

```bash
go run ./inverse-transposition-search validate \
  -input transformed.txt -oracle doyle.txt \
  -candidate /tmp/task54-search/w008-natural-r01.txt
```
