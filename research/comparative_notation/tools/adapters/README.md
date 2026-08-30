# Adapter layer

The executable shared adapter contract is `internal/notation/adapter.go`.
Corpus-specific parsing belongs here in future versions and must output USC;
it cannot alter generic analysis. `fixtures/` contains source, expected, and
generated USC for every class, including each music representation.
