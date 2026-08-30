# C08 source decision — positional numerical notation

Decision date: 2026-08-30. First-run decision:
`EXCLUDED_ROLE_MISMATCH`; `C08_PRODUCTION_READY=false`.

## Sources considered

| Source | Evidence | Decision |
|---|---|---|
| DReMAR v1, DOI 10.5281/zenodo.13757090 | Stable 115.2 MB snapshot, CC BY 4.0, PAGE/ALTO-like XML; medieval accounts contain running text, Roman numerals, and separately tagged sum lines | Reject for C08: valuable accounting data, but not the preregistered positional-number system. Using it would change the candidate definition. Downloaded temporary archive SHA-256: `8b598da9e6688a94859315f1ad04ee3d2255eb176ae1a17b7ee4ab992cd428fe`. |
| Existing deterministic decimal fixture | Fully reproducible | Reject: synthetic fixture cannot replace a historical production corpus. |
| Unspecified 1400–1600 mensuration/accounting source | Scientifically plausible | Reject until a stable diplomatic structured transcription and license are identified. |

DReMAR resolves neither the positional-notation requirement nor the frozen
field/token semantics. The temporary download was inspection-only and is not
a production raw input.

Positional numerical notation is mandatory because C08 was preregistered as a
control whose sign value depends on numerical position. DReMAR's Roman-numeral
accounting prose does not implement that role. DReMAR may be proposed later as
a new accounting-text C10/Cxx control, but it is not renamed C08 retroactively.

Source checked: `https://doi.org/10.5281/zenodo.13757090` (v1, published
2024-09-13; accessed 2026-08-30).
