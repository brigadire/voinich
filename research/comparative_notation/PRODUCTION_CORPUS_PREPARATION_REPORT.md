# Production corpus subset preparation report

Freeze timestamp: 2026-08-30T14:04:30Z. Selection version:
`production-corpus-selection-v1`. Base Git revision:
`f6fc78efed8399044ed73069e0637e387d5fa5a9`.

This task froze source and selection artifacts only. It did **not** execute the
generic analyzer, comparative metrics, rarefaction, bootstrap, calibration,
VM comparison, ranking, or interpretation. Selection decisions were based only
on source suitability, licensing, representation semantics, transcription
feasibility, and applicability of the already frozen protocol.

## Outcome

The maximal reasonably achievable first-run subset is C01, C02, and C06.
C01/C02 retain their previously validated paired BDD bundles. C06 is newly
production-ready from all secure MusicXML files in frozen JLSDD revision
`64bc7461dc564163cab84894bda263e4501b1a0a`.

| Candidate | Decision | Basis |
|---|---|---|
| C01 | `INCLUDED` | Expanded BDD USC, 19,132 records, schema and bitwise reproduction pass. |
| C02 | `INCLUDED` | Diplomatic BDD USC, 19,121 records, schema and bitwise reproduction pass. |
| C03 | `EXCLUDED_SOURCE_UNAVAILABLE` | No machine-readable Tironian running text; a new independently validated 5k-token transcription is not feasible in this freeze. |
| C04 | `EXCLUDED_LICENSE` | `LICENSE_UNRESOLVED` for a mechanism-specific DECODE subset and derived USC. |
| C05 | `EXCLUDED_LICENSE` | `LICENSE_UNRESOLVED` for the required complete Fontana transcription. |
| C06 | `INCLUDED` | GPL-3.0 JLSDD secure subset; 79 documents; all MUSIC-R1/R2/R3 USC outputs validate and reproduce bitwise. |
| C07 | `EXCLUDED_LICENSE` | `LICENSE_UNRESOLVED` for ECOLM corpus content and derived USC. |
| C08 | `EXCLUDED_ROLE_MISMATCH` | DReMAR is open but does not implement positional numerical notation. |
| C09 | `DEFERRED` | Preferred public-domain images exist, but no preregistered, double-keyed, uncertainty-aware cell transcription exists. |

All decisions and candidate manifests are SHA-256-bound by
`PRODUCTION_CORPUS_SELECTION.json`. A new selection version is required to
change them; comparative results cannot be a reason for such a change.

## C06 frozen production bundle

The upstream archive SHA-256 is
`3634f3848ee317dc15e77721a2f2e2062fdeeab9f7f129fa1aabd31aa14653ef`.
The deterministic secure-subset archive SHA-256 is
`1149fe616062e7d40a068a62fe8727ad6248ee8c44708e848eccdfc1c99ce64e`.
The path policy selects every XML file directly under both upstream `secure`
MusicXML directories: 33 Josquin and 46 La Rue files. The upstream README says
33+44; the actual frozen revision has 79 matching files, and none was silently
dropped to force agreement with the prose count.

The frozen representations are:

- `MUSIC-R1` — 13,117 note/rest events; SHA-256
  `8521706b957601ec682d2e280299f746f4a03eba4414eac5b046d1bc4de54cfc`.
- `MUSIC-R2` — 11,785 within-voice signed semitone intervals; SHA-256
  `4653040ed06d37b3599ed2f098e24664d3070e08027dcb7fe297844a13d89efd`.
- `MUSIC-R3` — 12,475 pitch-duration events; SHA-256
  `6a48ac028f1181c39a9a1c89fc347cd5fa88105fc4cac755f357fd5bf21173d9`.

R2 makes source part/voice/staff an observed USC section so generic sequence
metrics cannot manufacture cross-voice adjacency. Rests reset interval
continuity. R1/R3 use deterministic onset and source-part tie-breaking while
retaining simultaneity IDs. Encoded page/system boundaries are preserved; no
measure is relabelled as a physical line. A second build in an independent
directory matched all three files byte-for-byte.

JLSDD is a modern symbolic transcription, not a diplomatic encoding of
original mensural glyph shapes. C06 can support event, interval, and
pitch-duration structure; paleographic glyph-shape claims are out of scope.

## Panel and statistical sufficiency

The panel contains two independent notation/source families: paired Latin
manuscript representations from BDD and symbolic Renaissance polyphony from
JLSDD. It is therefore neither a single notation type nor exclusively one
source family. C01 and C02 remain explicitly paired and are not counted as two
independent source families.

The frozen candidate-vs-VM raw metrics and observed-size bootstrap impose no
minimum number of candidates. Each included USC contains at least 10,000
records, so the 5,000 and 10,000 rarefaction checkpoints apply. All are below
20,000 records, so the 20,000 and 39,380 checkpoints must be emitted as
`NOT_COMPARABLE`, exactly as the frozen short-corpus rule requires. No
within-class distance distribution is claimed: that optional procedure
requires at least three independent corpora within one class, which the
selected panel does not provide. This limitation does not block the applicable
candidate-vs-VM procedures.

## Limitations caused by exclusions

The first run will not represent shorthand, historical cipher families,
Fontana, tablature, positional numerals, or layout-bearing historical tables.
Consequently it cannot support broad claims across the full preregistered
notation taxonomy. C01/C02 share one manuscript subset, C06 uses modern
MusicXML transcription, and no included corpus reaches the two largest frozen
rarefaction checkpoints. These absences must remain visible in the eventual
comparative report.

## Corpus-subset preflight

The selection-aware corpus validator verifies all nine mutually exclusive
decisions, every included raw/policy/profile/provenance/USC hash, USC schema and
record counts, all candidate-manifest hashes, every exclusion/deferment decision
hash, panel-assessment booleans, and the negative authorization marker. The
comparative-run authorization preflight remains a separate future operation;
this task does not set authorization true.

## Final status

```text
GLOBAL_COMPARISON_PROTOCOL_FROZEN=true
FULL_CANDIDATE_PANEL_READY=false
PRODUCTION_CORPUS_SUBSET_FROZEN=true
PRODUCTION_CORPUS_INCLUDED_COUNT=3
PRODUCTION_CORPUS_INCLUDED=C01,C02,C06
PRODUCTION_CORPUS_EXCLUDED=C03:EXCLUDED_SOURCE_UNAVAILABLE,C04:EXCLUDED_LICENSE,C05:EXCLUDED_LICENSE,C07:EXCLUDED_LICENSE,C08:EXCLUDED_ROLE_MISMATCH
PRODUCTION_CORPUS_DEFERRED=C09:cell_transcription_and_independent_validation
PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=false
```
