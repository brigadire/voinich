# Reproducibility

From this directory run:

```bash
python3 -B scripts/analyze.py
python3 -B scripts/validate.py
python3 -B -m unittest discover -s tests -v
sha256sum -c --quiet SHA256SUMS
```

Generation uses only registered local frozen inputs. Bootstrap seeds derive from SHA-256 metric keys; figures and tables are deterministic. `validate.py` regenerates into a temporary directory, compares every analytical output byte-for-byte, rechecks all input ledgers, then freezes this directory only. Existing source directories are read-only inputs.
