# Third-party code audit

## External IVTT

The former `ivtt/ivtt.c` and Linux binary are excluded from the release tree
and from the clean public history. IVTT is not part of this Apache-2.0 project.
[DATA.md](../../DATA.md) documents upstream acquisition, the exact source
SHA-256 used (`d17439299e3ffb9bf0101de611d19ebe472b518fe7dae096f224e0950fa70d45`),
an external build, `IVTT_BIN`/`PATH` discovery, and the required
`ivtt -x7 INPUT OUTPUT` interface. The upstream source URL returned HTTP 200
and matched that hash on 2026-08-23. No IVTT redistribution issue remains in
the cleaned tree itself.

## Go module dependencies

The exact modules in `go.mod` were checked from their versioned module-cache
`LICENSE`/`NOTICE`/`PATENTS` files on 2026-08-23:

| Module | Exact version | License |
| --- | --- | --- |
| `gopkg.in/yaml.v3` | `v3.0.1` | Dual file-level scope: MIT for the listed libyaml-derived C-port files; Apache-2.0 for remaining files. Its NOTICE says Copyright 2011–2016 Canonical Ltd. |
| `golang.org/x/text` | `v0.31.0` | BSD 3-Clause, Copyright 2009 The Go Authors; separate Go PATENTS grant included upstream. |

They are resolved by Go at build time; dependency source is not vendored.
Their own notices and conditions continue to apply. The module list from
`go list -m all` contained no additional build-list module beyond the main
module and these two dependencies.

## License scope and remaining issue

The root Apache-2.0 `LICENSE` covers original project source code only.
It does not relicense IVTT, IVTFF, Astafiev, other external data, or generated
research artifacts. No organization or institution was invented for NOTICE;
there is no project NOTICE because the owner supplied no additional required
notice and dependency source is not redistributed.

A generated-content audit found exact copies of the excluded IVTT-derived
ZL3b-x7 corpus in two frozen artifacts. At the owner's explicit direction,
both payloads were excluded from the tree and clean public history while their
checksums and provenance were retained. No third-party redistribution blocker
remains; see [RELEASE_BLOCKERS.md](RELEASE_BLOCKERS.md).
