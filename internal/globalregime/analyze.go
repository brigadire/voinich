package globalregime

import (
	"fmt"
	"sort"
)

func defaults(c Config) Config {
	if len(c.WindowSizes) == 0 {
		c.WindowSizes = []int{50, 100, 200, 500, 1000}
	}
	if c.Seed == 0 {
		c.Seed = 160016
	}
	return c
}

func validate(c Config) error {
	if c.CorpusPath == "" {
		return fmt.Errorf("corpus path is required")
	}
	if c.OutputDir == "" {
		return fmt.Errorf("output directory is required")
	}
	if c.Step < 0 {
		return fmt.Errorf("step must be non-negative")
	}
	if len(c.WindowSizes) == 0 {
		return fmt.Errorf("at least one window size is required")
	}
	seen := map[int]bool{}
	for _, s := range c.WindowSizes {
		if s < 2 {
			return fmt.Errorf("window sizes must be at least 2")
		}
		if seen[s] {
			return fmt.Errorf("duplicate window size %d", s)
		}
		seen[s] = true
	}
	return nil
}

func analyze(c Config, p *progressReporter) (analysis, error) {
	if err := validate(c); err != nil {
		return analysis{}, err
	}
	p.begin(1, "Loading continuous corpus")
	tokens, err := readCorpus(c.CorpusPath)
	if err != nil {
		return analysis{}, fmt.Errorf("read corpus: %w", err)
	}
	p.update(1, 1, "Loading continuous corpus")
	sizes := append([]int(nil), c.WindowSizes...)
	sort.Ints(sizes)
	p.begin(2, "Building multi-scale window distributions")
	scales := make([]scaleAnalysis, 0, len(sizes))
	for i, size := range sizes {
		if size > len(tokens) {
			return analysis{}, fmt.Errorf("window size %d exceeds corpus length %d", size, len(tokens))
		}
		step := c.Step
		if step == 0 {
			step = max(1, size/10)
		}
		scales = append(scales, scaleAnalysis{size: size, step: step, windows: slidingWindows(tokens, size, step)})
		p.update(i+1, len(sizes), "Building multi-scale window distributions")
	}
	p.begin(3, "Detecting distributional change points")
	var allChanges []ChangePoint
	for i := range scales {
		w := scales[i].windows
		scales[i].changes = append(scales[i].changes, thresholdPeaks(w)...)
		scales[i].changes = append(scales[i].changes, pelt(w)...)
		scales[i].changes = append(scales[i].changes, binaryChangePoints(w)...)
		allChanges = append(allChanges, scales[i].changes...)
		p.update(i+1, len(scales), "Detecting distributional change points")
	}
	p.begin(4, "Matching boundaries across scales")
	boundaries := stableBoundaries(allChanges, sizes)
	p.update(1, 1, "Matching boundaries across scales")
	samples := make([][]Window, len(scales))
	sampleIndices := make([][]int, len(scales))
	matrices := make([][][]float64, len(scales))
	trees := make([][]edge, len(scales))
	diagnostics := []ClusterDiagnostic{}
	p.begin(5, "Hierarchical distribution clustering")
	for si := range scales {
		w := scales[si].windows
		samples[si], sampleIndices[si] = clusteringSample(w)
		matrices[si] = distanceMatrix(samples[si])
		trees[si] = mstEdges(matrices[si])
		for k := 2; k <= min(15, len(samples[si])); k++ {
			fitLabels := hierarchicalLabels(len(samples[si]), k, trees[si])
			fullLabels := expandLabels(w, samples[si], fitLabels, k)
			d := diagnosticsFor(w[0].WindowSize, "hierarchical", k, fitLabels, matrices[si])
			diagnostics = append(diagnostics, withFullAssignments(d, fullLabels))
		}
		p.update(si+1, len(scales), "Hierarchical distribution clustering")
	}
	p.begin(6, "K-medoids and contiguous segmentation diagnostics")
	for si := range scales {
		w, sample, d := scales[si].windows, samples[si], matrices[si]
		for k := 2; k <= min(15, len(sample)); k++ {
			fitLabels := kMedoids(d, k, c.Seed+int64(si))
			fullLabels := expandLabels(w, sample, fitLabels, k)
			kd := diagnosticsFor(w[0].WindowSize, "k_medoids", k, fitLabels, d)
			diagnostics = append(diagnostics, withFullAssignments(kd, fullLabels))
			fullLabels = binarySegments(w, k)
			fitLabels = make([]int, len(sample))
			for i, original := range sampleIndices[si] {
				fitLabels[i] = fullLabels[original]
			}
			cd := diagnosticsFor(w[0].WindowSize, "contiguous_segmentation", k, fitLabels, d)
			diagnostics = append(diagnostics, withFullAssignments(cd, fullLabels))
		}
		p.update(si+1, len(scales), "K-medoids and contiguous segmentation diagnostics")
	}
	for i := range scales {
		for _, d := range diagnostics {
			if d.WindowSize == scales[i].size {
				scales[i].diagnostics = append(scales[i].diagnostics, d)
			}
		}
	}
	out := Output{Parameters: map[string]any{"corpus": c.CorpusPath, "token_count": len(tokens), "window_sizes": sizes, "step": c.Step, "default_step": "max(1, window_size / 10)", "rare_token_policy": "no aggregation; full probability mass retained", "primary_metric": "Jensen-Shannon distance", "boundary_tolerance": "0.5 * smaller window size", "cluster_k_sweep": "2..15", "cluster_fit_max_windows": maxClusterFitWindows, "cluster_fit_sampling": "deterministic sequence-wide uniform sample; all windows assigned to fitted distribution centroids", "metadata_used": false, "seed": c.Seed}, Boundaries: boundaries, Diagnostics: diagnostics}
	return analysis{out, scales}, nil
}

// Kept separate from the sweep orchestration so progress stages can mirror the
// command's expensive phases.
func diagnosticsFor(size int, method string, k int, labels []int, d [][]float64) ClusterDiagnostic {
	return diagnostics(size, method, k, labels, d)
}
