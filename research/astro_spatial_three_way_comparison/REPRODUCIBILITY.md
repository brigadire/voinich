# Reproducibility

Run from the repository root. Requirements: Python 3.10+ (tested interpreter version
recorded in summary JSON), NumPy and Pillow. No network, model calls, transcription,
external scientific packages or frozen-source writes are required.

```bash
cd research/astro_spatial_three_way_comparison
python3 scripts/analyze.py
python3 scripts/validate.py
python3 -m unittest discover -s tests -v
sha256sum -c --quiet SHA256SUMS
```

`analyze.py` verifies four complete upstream ledgers before computing anything, registers
all inputs and verifies original manifests/images/provenance. It refuses a changed
registered snapshot or analysis-plan lock. It generates only derived files here.
`validate.py` runs tests, performs a full second analysis with figures in a temporary
directory, compares every regenerated file byte-for-byte, rechecks upstream hashes,
writes the validation report, freezes all outputs/scripts/plan/tests and checks SHA256SUMS.
The final independent test run includes output-ledger verification. For an already frozen
package repeat the commands unchanged; generated outputs remain byte-identical.

For isolated replay without touching existing derived outputs:

```bash
python3 scripts/analyze.py --output-dir /tmp/three-way-isolated-replay
```

Use a new directory for an input/protocol/plan revision. Do not refreeze upstream
packages to conceal checksum failures. An actual input inconsistency yields
BLOCKER_REPORT.md and requires reviewer direction, not automatic correction.

Prospective plan hash is locked in ANALYSIS_PLAN_LOCK.json before outcome aggregation.
Bootstrap seed 20260914, 2000 replicates, per-metric deterministic SHA-256 seed offsets.
Panel resampling retains all candidates within each selected cluster. A single panel
cell uses candidate bootstrap, explicitly marked. Raw-box, rotated-proxy and filled
ellipse-envelope representations are separately labelled; none changes frozen matching.

All figures have source TSVs registered in figures/FIGURE_INDEX.tsv. Canonical crops
and source/human polygons used in overlays are recorded in CASE_STUDIES and
CASE_OVERLAY_SOURCE. PNGs are data-derived scientific visualizations, not AI imagery.
