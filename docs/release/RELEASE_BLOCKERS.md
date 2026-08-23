# Release blockers

## BLOCKING

1. **Corpus rights:** IVTFF source/derivative rights are not documented for
   redistribution. Do not make `data/` or `data_work/` public with corpus
   bytes until permission or a compatible license is established.
2. **Tracked control corpus:** Astafiev recipe files are tracked with no
   license record. Verify rights or remove them from the reviewed public
   release while retaining source/checksum/procedure documentation.
3. **Vendored IVTT:** `ivtt/ivtt.c` and `ivtt/ivtt` have no recorded license.
   Obtain permission, replace, or exclude them.
4. **Project license:** No license has been selected for code, documentation,
   or generated artifacts.

## SHOULD_FIX

1. Decide a reviewed archival strategy for the approximately 527 MiB Git
   history and large frozen artifacts.
2. Inspect/prune local dangling Git objects according to normal retention
   policy before publication.
3. Run a dedicated secret scanner in the release environment and retain its
   report; no such scanner was installed for this audit.

## INFORMATIONAL

- A clean detached worktree passed build, vet, ordinary tests, and race tests.
- The security pattern scan found no credential-shaped material in reachable
  history or the current tracked tree.
- Repository visibility and remote state were not changed.
