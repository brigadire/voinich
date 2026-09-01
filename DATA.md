# External data guide

Corpus bytes are deliberately not bundled because their redistribution rights
are not established. Public availability is not permission to redistribute.
The repository tracks only provenance, checksums, preparation code, and
scientific results. Local inputs under `data/`, `data_work/`, and
`data_test/` are ignored by Git.

## Voynich IVTFF (canonical input)

- **Purpose:** canonical Voynich analysis and IVTFF metadata-aware stages.
- **Identity:** Zandbergen–Landini `ZL3b-n.txt`, IVTFF format 2.0, as used
  for the frozen Phase I run.
- **Authoritative source:** René Zandbergen's
  [IVTFF data directory](https://www.voynich.nu/data/) and
  [format/readme](https://www.voynich.nu/data/000_README.txt).
- **Acquire:** download
  [ZL3b-n.txt](https://www.voynich.nu/data/ZL3b-n.txt) yourself and place it
  at `data/ZL3b-n.txt`. The URL returned HTTP 200 on 2026-08-23.
- **Source SHA-256:** `bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc`.
- **Prepare:** install IVTT as described below, then run
  `IVTT_BIN=/path/to/ivtt scripts/prepare-external-data.sh ivtff`.
  This validates the source, executes the historical `-x7` conversion, and
  writes `data_work/ZL3b-x7.txt`.
- **Expected derived SHA-256:** `360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2`.
- **Redistribution:** `EXTERNAL_NOT_REDISTRIBUTED`; terms were not
  established, so neither source nor derivative belongs in Git.

## Astafiev natural-language control

- **Purpose:** Russian natural-language control used by historical experiments.
- **Identity:** V. Astafiev, *1000 culinary recipes*, Windows-1251 plaintext
  downloaded from RoyalLib; the historical file itself embeds the RoyalLib
  source URL.
- **Source page:** [RoyalLib book page](https://royallib.com/book/astafev_v/1000_kulinarnih_retseptov.html).
  It returned HTTP 200 on 2026-08-23 and offers download formats. Users must
  determine whether their acquisition and use are lawful.
- **Acquire:** obtain the plaintext from that page without routing it through
  this repository and place it at
  `data_test/astafiev-1000-culinar-receipts.txt`.
- **Source SHA-256:** `7200ce6cc01398192b05cf7f0d34040f391a07cc55e95b64b94525697b6f1d5c`;
  size 549,935 bytes; encoding Windows-1251. A changed source is rejected.
- **Prepare:** run `scripts/prepare-external-data.sh astafiev`. It invokes
  the repository's deterministic `cmd/codex_prepare` implementation:
  decode Windows-1251, lowercase, map Unicode punctuation/symbols and
  whitespace to ASCII spaces, collapse repeated spaces, preserve lines, and
  reject forbidden controls.
- **Expected output:** `data_test/astafiev-1000-culinar-receipts-prepared.txt`,
  SHA-256 `ff67a4fbf2606be4409724722e3e4d426aed27bdbeec1698babd92bd2b5eba5a`.
  The preparer also writes a local `.prepare.json` sidecar.
- **Redistribution:** `EXTERNAL_NOT_REDISTRIBUTED`; source and normalized
  plaintext are excluded because redistribution rights were not established.

## IVTT external tool

IVTT reads IVTFF and supplies the canonical `-x7` conversion used by the
legacy pipeline. It is required only for workflows that start from IVTFF.

Download `ivtt.c` and the manual from the upstream
[IVTT directory](https://www.voynich.nu/software/ivtt/). The source URL
returned HTTP 200 on 2026-08-23 and its SHA-256 was
`d17439299e3ffb9bf0101de611d19ebe472b518fe7dae096f224e0950fa70d45`,
matching the source formerly used locally. Review upstream terms yourself,
then build outside the repository, for example:

```bash
cc -O2 -o /path/outside/repository/ivtt /path/to/ivtt.c
export IVTT_BIN=/path/outside/repository/ivtt
scripts/prepare-external-data.sh ivtff
```

The public project neither vendors IVTT nor licenses it under Apache-2.0.
`IVTT_BIN` may be an absolute path or a command discoverable via `PATH`.
The required interface is `ivtt -x7 INPUT OUTPUT`.
The audited IVTT v2.4 returns status `3` after a successful conversion; the
bootstrap accepts only status `0` or `3` and still requires the exact expected
output SHA-256, so a nonconforming result cannot pass silently.

## Stolfi labels and titles inventory

- **Purpose:** independent label-status checks, including cross-section reuse
  of token forms identified as Astronomical labels.
- **Identity:** Jorge Stolfi, `labtit-98-07-20.idx`, the machine-readable
  release linked from *A large list of labels and titles* (last edited
  1998-07-20).
- **Authoritative source:**
  `https://www.ic.unicamp.br/~stolfi/EXPORT/voynich/98-02-01-lotsa-labels/labtit-98-07-20.idx`.
- **Acquire:** download the file yourself and place it at
  `data/stolfi-labtit-98-07-20.idx`.
- **Source SHA-256:**
  `cb210aaa75dfd2e9d86e63fd4cff1684acdfc2669bd6a6f9969f4e6bfe10071c`;
  91,861 bytes; 1,485 pipe-delimited records; Latin-1/ASCII-compatible text.
- **Redistribution:** `EXTERNAL_NOT_REDISTRIBUTED`; the source is locally
  acquired and ignored because redistribution terms were not established.

## Bundled controls

`data_test/pg2097-2.txt` (Doyle, Project Gutenberg; SHA-256
`0b2608f105ead7b17fe286a1d4ba32787e17e22d457818f93eb18572a956cc80`)
and `data_test/pg30795-mod.txt` (Longfellow, Project Gutenberg; SHA-256
`95bdc80f19b078aef2471250b19a87126a98d89912ca435580efc9db7416c398`)
remain bundled with their provenance. See
[the data-license audit](docs/release/DATA_LICENSE_AUDIT.tsv) for status.
