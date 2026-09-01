# Research results

Research outputs are grouped by experiment rather than kept in the repository
root.

| Directory | Contents |
|---|---|
| `astro_label_cross_section/` | Astronomical token reuse across manuscript sections |
| `astro_token_formation/` | M0 historical-term corpus, split, models, nulls, and reports |
| `astro_token_formation_m1/` | M1 global Latin/Arabic → EVA substitution search |
| `stolfi_label_inventory/` | Stolfi source-derived Astronomical inventory and matching audit |
| `stolfi_label_hapax_enrichment/` | Positive-label hapax enrichment and permutation results |
| `stolfi_matching_bias/` | Matching-selection-bias audit |
| `archives/` | Historical experiment result bundles and packaged controls |

Each experiment directory keeps its implementation, outputs, manifests, and
SHA-256 checksum file together. External corpus bytes remain governed by
`DATA.md` and are not moved into the repository.

The two large corpus archives in `archives/` are retained locally but ignored
from Git: `longfellow-song-of-hiawatha-v1.tar.gz` and
`msdos-v2-0-v1.tar.gz`.
