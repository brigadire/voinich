# Reproducibility and future validation commands

From repository root, Python 3.10+, NumPy (read-only existing input validator), Pillow.
No network/model calls in generation. No real human export is consumed in preparation.

```bash
cd research/astro_spatial_human_completeness
python3 scripts/build_package.py
python3 scripts/validate_package.py
python3 -m unittest discover -s tests -v
sha256sum -c --quiet SHA256SUMS
```

Builder verifies upstream and previous comparison ledgers, input identities/provenance,
creates blind overlay, navigation maps, schemas, manifests, deterministic ZIPs and
attachment dry run. ZIP dates fixed 1980-01-01. validate_package tests copied-tree XML
roundtrip (all coordinates/angles rounded to actual CVAT two decimals), synthetic new
proposal/reconciliation and relation decision paths, full separate-directory regeneration
and byte comparison, then rechecks input hashes and freezes this package outputs only.

Independent rebuild into a new directory:

```bash
python3 scripts/build_package.py --output-dir /tmp/human-completeness-replay
```

Future reviewer operations below are instructions, NOT executed during preparation.
After empty runtime CVAT import/export preflight (URL/version not supplied):

```bash
python3 scripts/check_cvat_version.py --base-url https://YOUR-CVAT-HOST
python3 scripts/validate_objects.py --input /path/to/empty-roundtrip.zip --reviewer-id R01 --timestamp 2026-09-14T12:00:00Z --allow-incomplete
```

After authorized real human object review, validation/staging in a fresh directory:

```bash
python3 scripts/validate_objects.py --input /path/to/complete-export.zip --reviewer-id R01 --timestamp 2026-09-14T12:00:00Z --output-dir /tmp/new-completeness-staging
```

This stages proposed additions, markers and reconciliation questions, NEVER accepts new
objects or changes prior data. Actual reviewer/timestamp must replace examples.
Only after externally reviewed object freeze per OBJECT_FREEZE_CONTRACT.md:

```bash
python3 scripts/attachment.py generate --object-freeze-manifest /path/to/new-object-freeze.json --output-dir /tmp/new-production-attachment
python3 scripts/attachment.py validate --input /path/to/relation-export.zip --queue-dir /tmp/new-production-attachment --reviewer-id R01 --timestamp 2026-09-14T12:00:00Z
```

Current-reference dry run, never production:

```bash
python3 scripts/attachment.py generate --dry-run --output-dir /tmp/new-not-for-production-dry-run
```

Changing frozen eligibility, input snapshots or protocol requires a separate revision.
No current script imports real results automatically, performs new AI comparison/recall,
accepts human proposals, or fixes legacy frozen input inconsistencies.
