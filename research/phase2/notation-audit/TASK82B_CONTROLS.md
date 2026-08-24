# Task82b independent notation-control portfolio

**Status: `TASK82B_CONTROLS_PARTIAL`.** No third-party bytes are added here;
acquire each source outside the repository, record version/checksum and confirm
its redistribution terms before use.

| Class | Dataset/input | Experiment | F2 availability | Limitation |
|---|---|---|---|---|
| Scribal abbreviation | Abbreviationes raw TEI `<abbr>/<expan>` pairs | Measure `ΔF=F(abbr)-F(expansion)`, length reduction and expansion ambiguity. | LP/EF ready; information-reduction metrics exploratory. | Licensing and corpus acquisition still required. |
| Scribal abbreviation | CATMuS Medieval | Run common F2 core on abbreviated graphematic lines. | Existing common core. | No expansion alignment. |
| Scribal abbreviation | CoMMA | Large-scale robustness baseline. | Existing common core. | HTR noise and no alignment. |
| Historical shorthand | Tironian notes | None until an aligned running-text corpus is found. | Not available. | `DATA_NOT_AVAILABLE`. |
| Positional / grille | Reimplemented Rugg/Zandbergen generator | Compare a preregistered generator family, not fit to Voynich. | F2 run after generator implementation. | Parameters/code incomplete in sources. |
| Omission/context notation | Aligned abbreviation pairs | Use ambiguity and context-conditioned expansion as a proxy control. | Exploratory only. | Not a direct mnemonic corpus. |

Fontana-derived models must remain separate from every row above. Task82b tests
the wider knowledge-dependent-notation class, not a post-hoc optimisation of
Fontana mechanisms.
