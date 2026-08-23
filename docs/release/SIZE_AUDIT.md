# Artifact size audit

At audit time the reachable Git object database was approximately **527 MiB**
(`git count-objects -vH`), with many individual tracked TSV/YAML artifacts in
the 40–58 MiB range. The largest include structural graphemic-pair,
transition-profile, and structural-profile-stability outputs in historical
homophony experiments, plus the baseline permutation artifact.

These files are frozen research artifacts and may be source-of-truth evidence.
They were not deleted or ignored merely because they are large. A public
release must decide, with scientific review, whether each large artifact stays
in Git, moves to an archival release asset with checksums, or receives a
reproducible regeneration path. Rewriting history to reduce size is out of
scope and requires an explicit decision.

The repository ignores mutable `workdir/`, local corpora, build binaries, and
profiling output. The audit added ignores for common IDE, scratch, temporary,
and coverage files without ignoring authoritative experiment artifacts.
