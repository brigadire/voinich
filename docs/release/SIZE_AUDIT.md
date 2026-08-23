# Artifact size audit

**Date:** 2026-08-23

The original private repository reported approximately 526.91 MiB of packed
reachable objects at the Task72b start commit. It was not pruned or rewritten.

The final clean-root public candidate contains 3,207 tracked files with
3,545,677,319 logical bytes. After `git gc`, `git count-objects -vH`
reported one pack, 3,040 objects, and **498.95 MiB** `size-pack`; its
checkout (excluding `.git`) uses approximately 3.4 GiB on disk. The size is
driven by retained frozen research artifacts and repeated/tabular scientific
outputs, not by the removed corpora, IVTT, or task files.

The largest tracked files remain 50–58 MiB transition-profile,
structural-graphemic-pair, and profile-stability artifacts. Size alone is not a
release blocker and no scientific evidence was deleted merely to reduce it.

Content audit is separate from size audit: two small
`experiments/voynich-v1/outputs/normalized_0*.txt` artifacts were exact full
copies of the excluded canonical derivative. They were removed at the owner's
explicit direction because of content, not size; checksums and provenance are
retained in [RELEASE_BLOCKERS.md](RELEASE_BLOCKERS.md) and the Task72b report.
