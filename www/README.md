# Voynich Phase I/II publication site

This directory is the complete static publication root for
`https://voynich.lite.systems/`. It has no runtime dependencies, application
database, external font request, analytics, or server-side component.

## Deploy

Serve this directory as the web root with directory-index resolution enabled.
Stable pages use directory URLs such as `/phase-1/` and `/claims/`.

## Validate

From the repository root:

```sh
www/validate.sh
```

The validator checks local HTML links/assets, required terminal-result wording,
artifact checksums, manifest metadata, and the absence of accidentally bundled
private/research working paths.

## Release identity

- publication release: `publication-site-v1.0.0`
- publication date: `2026-08-25`
- canonical scientific language: English
- scientific scope: frozen Phase I and Phase II only
- research-source commit: `56e0a1e97362ac7c2791e9ec1b1574e57ed71570`

The public artifact list is `artifacts/RELEASE_MANIFEST.json`; individual
checksums are in `artifacts/SHA256SUMS`.
The exact static tree (excluding the checksum file itself) is frozen in
`SITE_FILES_SHA256SUMS` so the released directory can be archived immutably.
