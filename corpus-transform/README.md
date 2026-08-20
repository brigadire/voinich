# corpus-transform

Deterministic historical-cipher corpus transformer (task46).

`corpus-transform` is a standalone experiment-input generator. It is:

- **not** a stage of the scientific pipeline;
- **not** invoked automatically by `pipeline-orchestrate`;
- **not** a statistical analysis tool;
- **not** a comparison against the Voynich manuscript;
- **not** a parameter search for "closeness to Voynich";
- never a modifier of its input corpus.

Given an existing plain-text corpus, it deterministically applies one of two
token-level historical cipher mechanisms - rectangular columnar
**transposition** or **homophonic substitution** - and writes a new,
plain, whitespace-tokenized corpus that the existing generic scientific
pipeline can consume unchanged via:

```
pipeline-orchestrate manifest -generic-corpus -corpus <transformed-corpus>
```

See `TRANSFORMATION_METHODS.md` for the mathematical definitions,
determinism/seed derivation, remainder handling, line-boundary policy, and
the scope/limitations of these mechanistic controls.

## Usage

```
go run ./corpus-transform \
    -corpus data_test/pg2097-2.txt \
    -method transposition \
    -output data_test/transformed/doyle__transposition__w008__natural__seed001.txt \
    -seed 1 \
    -transposition-width 8
```

```
go run ./corpus-transform \
    -corpus data_test/pg2097-2.txt \
    -method homophonic \
    -output data_test/transformed/doyle__homophonic__h004__uniform__seed001.txt \
    -seed 1 \
    -homophones 4
```

Every run writes:

- `<output>` - the transformed canonical corpus (plain whitespace-tokenized
  text, no headers, comments, or markers);
- `<output>.transform.json` - the full reproducibility manifest (task46
  section 11);
- for `-method homophonic` only: `<output>.mapping.tsv` - the plaintext to
  cipher-token mapping used for audit/reproducibility. **This file is never
  given to the scientific pipeline.**

The input corpus is never modified.

### Common flags

| Flag | Default | Meaning |
|---|---|---|
| `-corpus` | (required) | input corpus path |
| `-method` | (required) | `transposition` or `homophonic` |
| `-output` | (required) | output corpus path |
| `-seed` | `1` | deterministic PRNG seed |
| `-line-policy` | `preserve` | `preserve` or `reflow` - see TRANSFORMATION_METHODS.md |

### Transposition flags

| Flag | Default | Meaning |
|---|---|---|
| `-transposition-width` | (required) | rectangular columnar width |
| `-transposition-order` | `natural` | `natural` or `keyed` |
| `-rounds` | `1` | repeated transposition rounds |

### Homophonic flags

| Flag | Default | Meaning |
|---|---|---|
| `-homophones` | `4` | H, homophones per plaintext token |
| `-homophone-selection` | `uniform` | `uniform` or `weighted` |
| `-homophone-model` | `fixed` | `fixed` only; `frequency` is backlog (see TRANSFORMATION_METHODS.md) |

## Batch generation

```
go run ./corpus-transform batch \
    -corpus data_test/pg2097-2.txt \
    -output-dir data_test/transformed \
    -label doyle \
    -transposition-widths 2,4,8,16,32 \
    -transposition-order natural \
    -homophone-counts 2,4,8 \
    -homophone-selection uniform \
    -seeds 1
```

generates one corpus + manifest (+ mapping.tsv for homophonic runs) per
combination of width/H x seed, named per the stable experiment-id scheme
(task46 section 14):

```
doyle__transposition__w008__natural__seed001.txt
doyle__homophonic__h004__uniform__seed001.txt
```

Scientific identity is never determined by this filename - only by the
manifest's hashes and parameters.

## Scientific separation

`corpus-transform` ends at corpus creation. It does not run
`pipeline-orchestrate`, does not run `experiment-compare`, computes no
distance to Voynich, selects no "best cipher", and never tunes its
parameters against Voynich or any other comparison target. Any such
analysis is a separate, later experiment.
