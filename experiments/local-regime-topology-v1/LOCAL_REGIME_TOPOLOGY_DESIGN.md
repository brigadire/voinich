# Task65 frozen design (LOCAL_REGIME_TOPOLOGY_DESIGN)

Analysis units: the unified manuscript coordinate x (GlobalIndex, Line,
PageIndex, Folio/Currier/Hand/Section) built from data_work/ZL3b-x7.canonical.txt
in file order, exactly as Tasks58-64 read it; manuscript order is
explicitly NOT assumed to be the original semantic order (section 5).

Currier A/B is IVTFF $L ("Currier's Language A/B", per the literal
comment attached to $L/$H in data/ZL3b-n.txt's page headers); Hand is
$H; Section is $I (the standard illustration-type code: H=Herbal,
P=Pharmaceutical, S=Stars, A=Astronomical, B=Biological, C=Cosmological,
T=Text, Z=Zodiac). This intentionally differs from the "$C" field other
packages in this repo call Currier (a separate, finer scribal-hand
numbering) - see the code comment on metadata() for the full rationale.

Feature set: the authoritative Task64 Profile (MeanLen, GiantFrac,
TopInit, TopFinal, TypeEnt over internal/lineregime.ComputeProfile) is
the PRIMARY feature/distance for decay, correlation length and boundary
discontinuity (section 11, reusing Task64's distance rather than
inventing a new one). Clustering/hierarchical-variance use a SECONDARY,
7-dimensional extended vector (the 5 Profile dims plus mean d1
neighborhood degree and local near-repeat rate), z-scored using
DISCOVERY-fold windows only (section 10).

Window sizes: primary W=20 tokens, step=W/4=5; sensitivity W=10/40/80,
same step rule; additional physical-line and page scales reuse Task64's
own Line/page grouping. Fixed before Voynich results were inspected
(section 6).

Lag range: token lag 1..30 steps; line/page lag 1..10 (matching Task64's
persistence range).

Change-point algorithms: (1) a deterministic non-overlapping-window
profile-distance scan (ScanChangePoints); (2) the classical CUSUM
maximum-cumulative-deviation statistic applied to that same score series.
Significance is calibrated against a STATIONARY synthetic null's P95/P99
score, never against the Voynich data itself (section 19).

Clustering: deterministic k-medoids (PAM-lite), K=2..8, medoids fit on
DISCOVERY-fold windows only; K is selected by held-out (VALIDATION-fold)
mean distance-to-medoid among K values whose bootstrap co-cluster
stability is >=0.5, never by agreement with Currier/Hand/Section/Page
(section 37/39 - metadata is compared only after this freeze).

Null models: shuffle-order nulls for decay curves; label-preserving
shuffle nulls for transition/dwell; STATIONARY/SMOOTH_DRIFT/DISCRETE/
MIXED synthetic corpora (deterministic contiguous partitions of the
corpus's own sorted vocabulary into "flavors", never hand-picked or
fit to any Voynich effect) for change-point calibration and pipeline
validation (sections 19-22).

Metadata controls: page-level aggregate profiles (not per-window) for
between/within Currier/Hand/Section comparisons, so N is pages, not
pseudo-replicated windows; strata under 5 pages or 200 tokens are
NOT_APPLICABLE / INSUFFICIENT_DATA (section 71).

Discovery/replication protocol: IDENTICAL to Task64 - contiguous folio
blocks in manuscript order, train/validation/test = 50%/20%/30%
of pages, discovery = train+validation, replication = test, seed 64000+1
for the shared nullMean baseline (section 33's byte/metric reproduction
requirement).

Acceptance criteria: as listed in task65 sections 53-56/64; thresholds
(0.5 stability, 30% variance-explained boundary for
PARTIALLY_METADATA_EXPLAINED, +-0.005 Delta gap for COMPOSITION_EXPLAINED)
are fixed here, before Voynich results were inspected.
