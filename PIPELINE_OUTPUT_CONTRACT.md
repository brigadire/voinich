# Pipeline output contract

`./workdir` is the only repository-local destination for generated pipeline
content. This includes intermediate datasets, final YAML/TSV/Markdown reports,
normalized corpora, plots, temporary pipeline executables, and outputs added by
future analysis applications.

Contract for every existing and future pipeline command:

1. Default output paths must be rooted below `workdir/`.
2. Inputs produced by another pipeline stage must default to their path below
   `workdir/`; immutable source corpora may remain outside it.
3. Commands must create the required output directory when it is absent.
4. Go commands must use `internal/workdir` instead of repeating the directory
   name. The repository test suite checks this import contract for every
   command package.
5. Shell pipeline entry points must declare one `workdir="workdir"` variable
   and derive generated paths from it.
6. `workdir/**` is intentionally ignored by Git. Only `workdir/.gitkeep` is
   versioned; generated results must not be committed.

Explicit CLI output overrides remain available for controlled external runs,
but repository defaults and bundled pipeline scripts must obey this contract.
