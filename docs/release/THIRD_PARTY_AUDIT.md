# Third-party code audit

## Declared dependencies

`go.mod` declares `gopkg.in/yaml.v3 v3.0.1` and
`golang.org/x/text v0.31.0`. They are module dependencies, not vendored source.
Their upstream license notices must be retained and confirmed against the
exact released module versions when a project license is selected.

## Vendored and generated material

`ivtt/ivtt.c` and the compiled `ivtt/ivtt` are vendored from the voynich.nu
IVTT tool. The tree contains no copyright header or license for either file.
They are **not cleared for public redistribution** and are a release blocker.
The public release should either obtain written permission and retain the
notice, replace the tool, or exclude it and document local acquisition.

Frozen files in `experiments/` are generated research artifacts, not
third-party source code. This audit found no separately vendored Go module
tree. A line-by-line provenance determination for every historical snippet is
not available, so contributions must identify copied material and its license.
