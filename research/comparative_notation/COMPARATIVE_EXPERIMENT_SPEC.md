# Frozen comparative experiment specification v1

## Units and isolation

The ten independent experiment units are C01 ordinary historical Latin, C02
abbreviated manuscript Latin, C03 historical shorthand, C04 historical
ciphers, C05 Fontana notation, C06 mensural music, C07 tablature, C08
positional numerical records, C09 formal/tabular records, and C10 synthetic
calibration controls. A corpus instance belongs to exactly one class;
mechanisms and traditions remain separate instances. Failure in one class
must not block another.

Each run fixes `(corpus_id, source version, representation_id, adapter
version, analyzer version, metric registry version, parameters, seed)`.
Source → representation → USC → validation → generic analysis → comparison is
strict. Corpus-aware logic ends at USC.

## Frozen supports and sizes

Sequence supports are `frequency>=5`, `frequency>=10`, `top100`, `top250`, and
matched vocabulary `N=553` (the frozen VM frequent-token support). Corpus-size
checkpoints are 5,000, 10,000, 20,000, and 39,380 tokens. Seed `20260830` is
reserved, but the boundary-preserving sampling unit and replicate/CI policy
remain STOP blocker B03; unordered token sampling must not silently destroy
sequence structure. Raw density and rarefied estimates are both retained. A
corpus below a checkpoint is `NOT_COMPARABLE` there.

Accumulation curves are A2(N) symbol-bigram types, A3(N) symbol-trigram types,
and AT(N) token-transition types. Production curves use the same checkpoints
and seed schedule for VM and candidates.

## Outputs

Each corpus/representation independently creates the files listed in
`RESULT_CONTRACT.md`. A class with at least three independent corpora may later
create within-class pair distances and variance summaries. These are not
candidate-vs-VM rankings. `compare-classes` is installed but repository-locked.

Paired sources create two ordinary fingerprints plus `NOTATION_DELTA.tsv`,
matched by metric ID and representation pair. A delta is never filled when
one side is not comparable.

Production execution is disabled. Only manually inspectable fixtures and
synthetic test data may pass through `analyze` during preparation.
