# Reproducibility

From this directory:

```bash
python3 -B scripts/build_preparation.py
python3 -B scripts/validate_preparation.py
python3 -B -m unittest discover -s tests -v
sha256sum -c --quiet SHA256SUMS
```

Open `human_review/index.html` locally, complete all 92 rows and export TSV. Validate/import later with `python3 -B scripts/import_human_mapping.py --input EXPORT.tsv --reviewer-id ID --timestamp ISO8601 --output-dir NEW_DIRECTORY`. The importer refuses incomplete or invented-token mappings. Validation rebuilds generated preparation files in a temporary directory and compares them byte-for-byte.
