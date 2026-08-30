# Deterministic technical pre-run report

Scope: loading, USC validation, structural (boundary-preserving block) traversal, and seed-schedule derivation for the frozen C01, C02, C06 subset. No structural, rarefaction, bootstrap, distribution, calibration, or VM-comparison metric was computed or saved; this is not a production comparative run.

| Candidate | Representation | Records | Structural blocks | Lines observed | Seeds | Seed schedule SHA-256 |
|---|---|---|---|---|---|---|
| C01 | LATIN-EXPANDED | 19132 | 3393 | false | 1000 | `ab2cbf1fe00a7965ae0b2f0886a9fbb8c6828da85cebfbc194ed8be09afc2fab` |
| C02 | LATIN-DIPLOMATIC | 19121 | 3393 | false | 1000 | `cd58b55effebf7ae25e31021bcb5dd12ae88bd7567211f248b784d1ddfba0e54` |
| C06 | MUSIC-R1 | 13117 | 559 | true | 1000 | `eb12f977749fac6816007cafc4935351b82e09db7bb0de72e586b8a29453e23f` |
| C06 | MUSIC-R2 | 11785 | 1118 | true | 1000 | `483a55c0955f1017b988ce2ec34d95cccad1868be3954c20d0a728c6e6fc9329` |
| C06 | MUSIC-R3 | 12475 | 559 | true | 1000 | `f48c65dbac72240bbec741d7ea280660e7829e64ff356754dd3bdd34188cff91` |

Overall technical output SHA-256 (pass 1): `2eea660270039b285b7a2c7a20b9cb31fcad830e8377495be4e7b0488b534645`

Overall technical output SHA-256 (pass 2): `2eea660270039b285b7a2c7a20b9cb31fcad830e8377495be4e7b0488b534645`

Serialized pass1/pass2 byte-identical: `true`
