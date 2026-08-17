# Doyle isolated smoke validation

This is a new isolation-v1 experiment, not a continuation of contaminated v1.
It uses `data_test/pg2097-2.txt`, SHA256
`0b260c8ae9ee7dcfd8c334e174b32ec433d37792ecfa783137b0bd38a956cc80`.

The smoke scope stops after the first five applicable stages (`dict-gen`,
`dict-analyze`, `structural-analyze`, `sequence-analyze`, and
`begin-end-analyze`); it is deliberately not a production Doyle baseline and
is not frozen as a completed 27-stage experiment.

Validation performed after the run:

- no forbidden Voynich path sentinel or Voynich corpus SHA occurs in
  scientific artifacts;
- all seven artifacts in `artifacts.json` carry this experiment ID and Doyle
  corpus SHA;
- every metadata-dependent `NOT_APPLICABLE` stage owns zero artifacts;
- a resume validation pass accepts the registered hashes and invocation
  fingerprints;
- binaries are operational scratch under `workspace/workdir/bin` and are not
  scientific artifacts or freeze candidates.
