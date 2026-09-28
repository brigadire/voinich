# Input audit

57 unique target occurrences; all match verified source IDs, pages, L loci and
composite keys. Corpus-wide hapax was recomputed, not inherited from the old
enrichment table.

| Page | Hapax | Non-hapax |
|---|---:|---:|
| f68r1 | 11 | 19 |
| f68r2 | 15 | 12 |

Astronomy: 63 attestation rows, 54 distinct normalized within-identity
form slots, 31 canonical identities. Attestation counts are not matching capacity.

The following discrepancies existed before this experiment and were excluded from
authoritative input validation. They were not repaired:

| Old enrichment file | Registered SHA256 | Actual entry SHA256 |
|---|---|---|
| run_enrichment.py | `b543cf9b0581da8d8f173c39b04aeba23f409899a40723755e0ea4c2fe100377` | `66e1558e4bed5fe5e8bf7d2abdcf7e0d588bc96de957e477123ce091fdaad1dd` |
| COHORT_MEMBERS.tsv | `9394f229b4f74ca418a03255bf233e15e50d5c76122538dd16fe50373a09b574` | `add922aa85782dd62a57b2d8e5f3981bb2466c84ad77caf3833e8a5f29d4b2ab` |

The full unchanged-upstream snapshot includes these actual entry bytes as well as
all other prior packages. Required TARGET_SETS.tsv, metadata and D1 dictionary
passed their upstream registered checksums. See UPSTREAM_INPUTS.tsv.
