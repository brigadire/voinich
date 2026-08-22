# Task64 frozen design (LINE_REGIME_DESIGN)

Line definition: one physical transcription line = one line of
data_work/ZL3b-x7.canonical.txt, which corpus prep records as line_policy=preserve
relative to the IVTFF-derived data/ZL3b-n.txt (same line count,
verified at runtime, not assumed). Folio/Currier/Hand come from
internal/metadatavalidation.ParseIVTFF(data/ZL3b-n.txt) zipped to
canonical lines by index; if the locus count ever disagrees with the
canonical line count, folio/page metadata falls back to NOT_APPLICABLE
(same-page control degenerates to the corpus-wide control) rather than
being guessed.

Eligible lines: primary minimum 5 tokens; sensitivity thresholds 3/8/10
are reported in manifest.json's min_n_sensitivity, fixed before results
were inspected, not chosen for effect size.

Form-distance metric and feature set: internal/tokentransition.EditDistance
(Task60/63's Levenshtein over internal/evaglyph glyphs) is the only distance
used anywhere in this task. Features are limited to token length, initial
glyph, final glyph, giant-d1-component membership (union-find over the same
edit distance, length-bucketed) and evaglyph.Classify position class; no
alternative glyph/token distance is defined.

Null models: global pseudo-line, same-page pseudo-line, length-preserving
pseudo-line (task64 section 15), within-line shuffle, line-membership
shuffle, within-page line-membership shuffle (sections 16-18) - all in
internal/lineregime, unit-tested for multiset/page/length preservation.

Page controls: different-line/same-page (matched by token-length pair,
excluding the source line) is the PRIMARY control (section 8/33);
different-line/different-page (also excluding the source page) isolates
manuscript-wide vocabulary geometry (section 9).

Statistical tests: primary estimand Delta_line = P(d<=1 | same line,
non-adjacent) - P(d<=1 | different line, same page, matched), bootstrap
CI over lines (500 resamples, seed 64000+2). Scale comparison uses a
page-level bootstrap (20-60 resamples, fewer for the more expensive
PAGE/shifted/window scales) against a single shared null (global
length-matched different-page rate); rep counts are bounded well below
what a naive implementation would use specifically to honor section 67's
"no O(N^2) global matrix, use bucketing/sampling" performance mandate -
each replicate itself is already a full-corpus-scale computation, so the
rep count is the primary cost lever, fixed before results were inspected.

Discovery/replication split: contiguous folio blocks in manuscript order,
train/validation/test = 50%/20%/30% of pages by page-appearance
order (never splitting a page); discovery = train+validation
(70% of pages), replication = test.

Candidate regime models and selection: the minimal model (section 44) is a
K<=4 categorical mixture over TRAIN line mean-token-length quantiles
(K fixed in advance, not tuned against any Voynich fingerprint), with
per-regime token-length / initial-glyph / final-glyph categorical
distributions and a global (non-regime) middle-glyph distribution.
Model selection between global-only and regime-conditioned components
uses held-out (VALIDATION) bits-per-token, not Task58-63 fingerprint
metrics (section 58 safeguard).

Acceptance thresholds (section 40): NO_LINE_REGIME if
Delta_line<=0; PHYSICAL_LINE_REGIME if the LINE scale's bootstrap effect
size exceeds every shifted-line and fixed-window effect size;
BROADER_LOCAL_REGIME if a shifted/window scale matches or exceeds LINE;
otherwise LINE_ASSOCIATED.
