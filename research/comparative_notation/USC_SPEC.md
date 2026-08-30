# Universal Symbolic Corpus (USC) 1.0

USC is UTF-8 JSON Lines with one ordered token per record. Required fields are
`schema_version`, `corpus_id`, `representation_id`, `document`, `section`,
`page`, `locus`, `physical_line`, `token_id`, `token_index`, `token`, and
`symbols`. Each hierarchy field is `{value, observed}`. An absent source level
is `{observed:false}` (the omitted `value` decodes as NULL); it is never
fabricated. `document` must be observed.

`token_id` is stable for fixed source, representation, and source-unit ordinal.
`token_index` starts at zero and is contiguous inside the deepest observed
hierarchy unit. `symbols` is the adapter-declared ordered decomposition of the
token. Optional `attributes` preserves dimensions that the generic analyzer
must not semantically interpret: pitch, duration, voice, staff, course, fret,
simultaneity group, cell coordinates, and similar fields.

Example:

```json
{"schema_version":"usc-1.0","corpus_id":"C07-FIX","representation_id":"TAB-R1","document":{"value":"doc1","observed":true},"section":{"observed":false},"page":{"value":"p1","observed":true},"locus":{"observed":false},"physical_line":{"value":"sys1","observed":true},"token_id":"stable-id","token_index":0,"token":"a2+c0","symbols":["a2","c0"],"attributes":{"tradition":"French","simultaneity_group":"g1","course":"2,4","fret":"a,c"}}
```

Validation rejects mixed corpus/representation IDs, duplicate token IDs,
non-contiguous order, empty tokens/symbols, and observed/value disagreement.
For C06 and C07, adapter validation additionally requires the registered
multidimensional attributes.
