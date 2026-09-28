# Reproduction and integrity

Python 3.14 on Linux; standard library only. Always disable bytecode so the complete
file ledger remains stable. From this directory:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s tests -v
PYTHONDONTWRITEBYTECODE=1 python3 validation.py verify
sha256sum -c SHA256SUMS
```

The original execution sequence was `validation.py prepare`, `validation.py development`,
then `validation.py validate` with PYTHONDONTWRITEBYTECODE=1. The last command is one-shot:
it refuses to run after HIDDEN_VALIDATION_STARTED.json exists. Do not delete that marker,
regenerate hidden data, tune thresholds or rerun validation for development purposes.

For independent reproduction of a specific committed inference output, pass one
`surface_inputs/<hash>.json` to the isolated worker and write output/checkpoint outside
the package, for example in /tmp. Compare canonical JSON hashes to HIDDEN_RESULT_MANIFEST.

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -I -B worker.py /absolute/path/to/surface.json /tmp/m2r-output.json /tmp/m2r-checkpoint.json
```

Checkpoint reproduction: append `--pause` for the initial call, then `--resume` with
the same inputs and checkpoint. Output has no timestamps, seed, shard or input IDs.
The validation artifacts contain original/permuted/reversed/renamed-ID hashes and
continuous/resumed hashes for each validation dataset. A differing resource environment
may cause INCOMPLETE instead of a completed result; never relabel that as NO_MODEL.

SHA256SUMS covers every file except itself. CANDIDATE_SEAL.json binds implementation
and evaluation code before hidden disclosure; it is not a successful engine freeze.
Only a passing full gate can create M2R_V2_FREEZE_MANIFEST.json. Any failed hidden gate
requires M2R-v3 with fresh hidden seeds. No command in this package authorizes real data.
