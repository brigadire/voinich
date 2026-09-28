# Reproduction

Requires the Python version recorded in FREEZE.json (standard library only), this
package and the unchanged upstream files listed in UPSTREAM_SNAPSHOT.json. Run
from repository root:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 -m unittest discover -s research/astro_hapax_star_label/restricted_dictionary_bruteforce_v1 -p 'test_*.py' -v
PYTHONDONTWRITEBYTECODE=1 python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v1/run.py verify
PYTHONDONTWRITEBYTECODE=1 python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v1/run.py run
```

The last command resumes completed hashed batches; it never modifies the frozen
grid or input files. A fresh independent reproduction uses another output directory:

```sh
PYTHONDONTWRITEBYTECODE=1 python3 research/astro_hapax_star_label/restricted_dictionary_bruteforce_v1/run.py run --out /tmp/bruteforce57-replay
```

The complete run includes 60 synthetic datasets with 99 synthetic nulls each,
90,000 informative null datasets (10,000 for each of nine families), 30,000 extra
historical independent-capacity searches, 57 leave-one-out searches, 200 subsamples,
and exact invariance controls. All search budgets are the same 64 global systems.
The production wall budget is four hours per invocation; a timeout leaves resumable
INCOMPLETE status and withholds best-model interpretation. Seed derivation is in
the protocol. TSVs and compressed checkpoint bytes are deterministic; timestamps
in the command journal and freeze receipt are provenance, not numerical outputs.

Check package checksums from the package directory with `sha256sum -c SHA256SUMS`.
The checksum manifest excludes itself, active lock and transient files. All input
and code checksums are separately protected by FREEZE.json, which precedes the
first production score. No runtime dependency import writes upstream pycache.

RUNNING.lock prevents simultaneous writers to the same output directory. A hard
process kill can leave a stale lock; verify that its recorded PID is no longer
running before removing only that file and resuming. An unregistered shard left
by interruption is deterministically overwritten; a registered corrupt shard is
rejected. Checkpoint validity includes freeze digest, compressed-byte SHA256,
row counts and seed ranges. Never repair a frozen v1 dictionary/grid after seeing
results: make a separately registered v2 instead.

## Initial preparation commands

Sources were downloaded with curl -L --fail --silent --show-error, from:

- https://www.thelatinlibrary.com/isidore/7.shtml → sources/isidore_7.html
- https://www.thelatinlibrary.com/isidore/17.shtml → sources/isidore_17.html
- https://archive.org/download/buch-der-alaune-und-salze_-_liber_de_aluminibus_et_salibus__q218/Buch-der-Alaune-und-Salze_-_Liber_de_Aluminibus_et_Salibus%2C__q218_djvu.txt → sources/ruska_1935.txt

Initial sandbox DNS attempts failed; authorized curl retries succeeded. The source
snapshots remove network dependence from reproduction.

`PYTHONDONTWRITEBYTECODE=1 python3 .../prepare.py` was run before freezing. Its first
validation attempt rejected pre-existing enrichment SHA256 discrepancies. The
preparer was corrected BEFORE any production score to rely on verified target
and corpus metadata directly. Its successful validation is saved in
PREPARATION_VALIDATION.json. No source corpus or frozen upstream output was changed.
The preparation script intentionally refuses to overwrite an existing freeze.
Input reconstruction for auditing should be done in a copied package without a
freeze receipt, never by altering the existing completed experiment.

The exact executed experiment/test/replay commands are recorded in COMMANDS.jsonl.
TEST_RESULTS.txt and VERIFICATION.json record final automated validation.
