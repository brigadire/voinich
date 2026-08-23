# Data

This repository separates source corpus identity from redistribution rights.
Public availability of a source does not establish a right to redistribute it.
The authoritative status table is
[docs/release/DATA_LICENSE_AUDIT.tsv](docs/release/DATA_LICENSE_AUDIT.tsv).

| Corpus | Origin and transcription | Included | Checksum / preparation |
| --- | --- | --- | --- |
| Voynich IVTFF / ZL3b | René Zandbergen's voynich.nu IVTFF distribution; `ZL3b-n.txt` is the pipeline source and `-x7` is its IVTT derivative | No (only placeholders are tracked) | Obtain locally, record SHA-256, then run `./ivtt/ivtt -x7 data/ZL3b-n.txt data_work/ZL3b-x7.txt`. |
| Doyle, *The Sign of Four* | Project Gutenberg text, prepared control | Yes: `data_test/pg2097-2.txt` | SHA-256 `0b2608…a956cc80`; preparation details are in its experiment manifests. |
| Longfellow, *The Song of Hiawatha* | Project Gutenberg text, modified prepared control | Yes: `data_test/pg30795-mod.txt` | SHA-256 `95bdc8…7416c398`. |
| Astafiev culinary receipts | Historical local control corpus, including prepared form | Yes, pending rights review | SHA-256 `7200ce…b6f1d5c` (source) and `ff67a4…b5eba5a` (prepared); do not redistribute until rights are established. |
| MS-DOS, Sanskrit, and Viṣṇusahasranāma controls | Local controls used by historical experiments | No | Obtain and license independently; their local preparation sidecars record exact input hashes. |

The shortened hash display is for orientation only; use the full values in
the audit table when validating an input. The `ivtt` translator is separately
audited as unlicensed vendored code and must also be obtained or cleared before
a public canonical reproduction.
