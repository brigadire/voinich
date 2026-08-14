package globalregime

// Exported wrappers around the frozen distributional representation,
// distance metric, clustering and change-point primitives used by
// global-regime-analyze. They add no behavior of their own: every call
// delegates to the same unexported formula already used by this package, so
// metadata-conditioned residual analyses can reuse it exactly rather than
// redefining it. Task 19 requires that Currier/hand-conditioned discovery use
// "the same distributional representation" and "the same method" as this
// package.

// Profile is the token-frequency window representation.
type Profile = profile

// Distribution exposes one window's token-frequency profile.
func (w Window) Distribution() Profile { return w.distribution }

// Labels exposes the full per-window cluster assignment a diagnostic was
// computed from.
func (d ClusterDiagnostic) Labels() []int { return d.labels }

// BuildWindows builds the sliding-window token-frequency representation over
// a token sequence. Callers that need windows confined to one contiguous
// corpus region (never crossing a metadata boundary) must pass only that
// region's tokens.
func BuildWindows(tokens []string, size, step int) []Window { return slidingWindows(tokens, size, step) }

// JSDistance is the Jensen-Shannon distance between two window profiles.
func JSDistance(a, b Profile) float64 { return jsDistance(a, b) }

// DistanceMatrix is the pairwise Jensen-Shannon distance matrix for a set of
// windows.
func DistanceMatrix(w []Window) [][]float64 { return distanceMatrix(w) }

// ClusteringSample deterministically subsamples windows for quadratic model
// fitting, exactly as global-regime-analyze does for long sequences.
func ClusteringSample(w []Window) ([]Window, []int) { return clusteringSample(w) }

// ExpandLabels assigns every window (not just the fitted sample) to its
// nearest fitted centroid.
func ExpandLabels(w, sample []Window, sampleLabels []int, k int) []int {
	return expandLabels(w, sample, sampleLabels, k)
}

// KMedoids is the deterministic seeded k-medoids clustering used by
// global-regime-analyze.
func KMedoids(d [][]float64, k int, seed int64) []int { return kMedoids(d, k, seed) }

// HierarchicalLabels cuts a minimum-spanning-tree hierarchical clustering at
// K clusters.
func HierarchicalLabels(n, k int, d [][]float64) []int { return hierarchicalLabels(n, k, mstEdges(d)) }

// BinarySegments is the contiguous segmentation (never reorders windows)
// used as the frozen "contiguous_segmentation" method.
func BinarySegments(w []Window, k int) []int { return binarySegments(w, k) }

// Diagnostics computes silhouette, within/between dispersion, cluster sizes
// and fragmentation for one clustering.
func Diagnostics(size int, method string, k int, labels []int, d [][]float64) ClusterDiagnostic {
	return diagnostics(size, method, k, labels, d)
}

// WithFullAssignments attaches full per-window labels (beyond the fitted
// sample) to a diagnostic and recomputes size/transition fields from them.
func WithFullAssignments(d ClusterDiagnostic, labels []int) ClusterDiagnostic {
	return withFullAssignments(d, labels)
}

// ThresholdPeaks, Pelt and BinaryChangePoints are the three change-point
// detectors combined by global-regime-analyze.
func ThresholdPeaks(w []Window) []ChangePoint    { return thresholdPeaks(w) }
func Pelt(w []Window) []ChangePoint              { return pelt(w) }
func BinaryChangePoints(w []Window) []ChangePoint { return binaryChangePoints(w) }

// StableBoundaries combines change points from multiple scales into
// multi-scale-supported boundaries.
func StableBoundaries(changes []ChangePoint, sizes []int) []StableBoundary {
	return stableBoundaries(changes, sizes)
}
