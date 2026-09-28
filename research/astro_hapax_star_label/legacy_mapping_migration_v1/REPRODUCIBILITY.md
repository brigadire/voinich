# Reproducibility

Run from repository root:

```bash
python3 research/astro_hapax_star_label/legacy_mapping_migration_v1/scripts/build_migration.py
python3 -m unittest research/astro_hapax_star_label/legacy_mapping_migration_v1/tests/test_migration.py
(cd research/astro_hapax_star_label/legacy_mapping_migration_v1 && sha256sum -c SHA256SUMS)
```

The builder is deterministic and standard-library-only except Pillow, used for
audit images. It verifies all upstream ledgers before reading analytical data.
The amendment is copied unchanged for isolated reproducibility runs and its
fixed SHA is checked before the gate. The original preparation directory and
its `SHA256SUMS` are never rewritten. Output ordering, the audit sample seed,
PNG compression, and report precision are fixed.

No network request, OCR model, full-92 AI reading, corpus-frequency join, hapax
analysis, or lexicon lookup is performed by this build.
