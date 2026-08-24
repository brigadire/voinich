# Task79b report — notation audit

## Scope and firewall

This audit covered shorthand/tachygraphy, scribal abbreviation, notae/sigla,
mnemonic cues, positional extraction, acrostic/telestich, grille/periodic
selection and omission/context recovery. Metric selection used the sources in
`SOURCES.md`, not Fontana results. Definitions, nulls and the finite
positional family were fixed in `POSITIONAL_CHANNELS.md` before the canonical
result was read.

## Literature and controls

The quantitative Voynich literature found is narrow. Rugg (2004) and
Zandbergen (2021) test selected grille/slot-table distributional outcomes;
Matlach et al. (2022) quantify symbol roles and positional behaviour; Parisel
(2026, preprint) proposes boundary/positional diagnostics. No primary
quantitative study was found for a Voynich acrostic, Tironian/notarial
shorthand, mnemonic notation, or omission/context-recovery mechanism.
Academia.edu-only shorthand and mnemonic claims are retained as unsupported
search results, not evidence.

Exact replication is unavailable for Rugg, Zandbergen and Parisel because
their implementations/parameters are not released. The local fixed-channel
run partially replicates the general positional-structure direction, not
their exact estimators. Matlach’s heuristic ligatureness score is not
replicated. `REPLICATION.tsv` records these distinctions.

The best future positive control is the aligned `Abbreviationes`/Burchard
TEI corpus; CATMuS and CoMMA are distributional medieval-abbreviation
controls. No open, machine-readable, plaintext-aligned Tironian running-text
corpus was located. No external bytes were added; provenance, licensing and
acquisition caveats are in `CONTROL_CORPORA.tsv`. Therefore
**`TASK82B_CONTROLS_PARTIAL`**.

## Positional result

`POSITIONAL_RESULTS.tsv` is a deterministic 1,000-permutation analysis of
the canonical `ZL3b-x7` corpus (SHA-256
`f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692`,
seed `20260824`). Each of the 11 predeclared channels has greater adjacent
normalized MI than its matched permutation null after BH correction
(`q=0.000999001`, the resolution limit). This establishes only excess
low-order positional/within-line dependence relative to those nulls.
It neither finds a message nor confirms acrostic, grille, shorthand,
steganography, or external-memory notation.

## F2 admission and freeze

Externally motivated candidates are expansion ratio, ambiguity
`H(expansion|abbreviation)`, positional-channel MI, and boundary-class/Zipf
diagnostics. The first two need aligned notation/plaintext and are
`EXPLORATORY_ONLY` for Voynich. Positional MI and boundary diagnostics are
`DEFER_TO_V2_1`: neither has been tested against an independent notation
control portfolio or the required stability battery; the latter is also
preprint-motivated. Abbreviation productivity is already represented by
LP/EF and is rejected as redundant. See `F2_COVERAGE_MAP.tsv` and
`F2_ADMISSION.tsv`.

No F2 metric, code path or existing value changed. Independently,
`FINGERPRINT_V2_GAPS.md` requires implemented, validated page/2D and
cross-scale families, full stability/redundancy assessment and an exercised
distance/Pareto comparison. `FINGERPRINT_V2_IMPLEMENTATION.tsv` still marks
multiple required metrics as `NEW_IMPLEMENTATION`,
`NEEDS_MINOR_EXTENSION` or `DEFERRED`. Accordingly the final classification
is **`F2_NOT_READY_TO_FREEZE`** and no `FINGERPRINT_V2_FROZEN` marker is
created. Any later candidate remains exploratory / `FINGERPRINT_V2_1`
material and cannot alter the confirmatory comparison.
