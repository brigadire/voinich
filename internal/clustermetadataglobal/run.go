package clustermetadataglobal

import (
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/workdir"
)

func defaults(c Config) Config {
	if c.DiscoveryDir == "" {
		c.DiscoveryDir = workdir.Dir
	}
	if c.TokenMetadataMap == "" {
		c.TokenMetadataMap = workdir.Path("token_metadata_map.tsv")
	}
	if c.MetadataReportPath == "" {
		c.MetadataReportPath = workdir.Path("metadata_validation_report.md")
	}
	if c.OutputDir == "" {
		c.OutputDir = workdir.Dir
	}
	if c.Permutations <= 0 {
		c.Permutations = 10000
	}
	if c.Seed == 0 {
		c.Seed = 1
	}
	return c
}

// RunAndWrite runs the confirmatory global multiple-comparison correction
// over the complete, unmodified frozen discovery search space and writes its
// outputs. Discovery windows, clustering and cluster assignments are never
// recomputed; only metadata labels are permuted.
func RunAndWrite(c Config) error {
	c = defaults(c)
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)

	p.begin(1, "Loading frozen discovery search space")
	fs, err := loadFrozenSpace(c.DiscoveryDir)
	if err != nil {
		return fmt.Errorf("load frozen search space: %w", err)
	}
	p.update(1, 1, "Loading frozen discovery search space")

	p.begin(2, "Loading token metadata map")
	tokenCount, err := readDiscoveryTokenCount(c.DiscoveryDir)
	if err != nil {
		return err
	}
	currier, hand, err := loadTokenLabels(c.TokenMetadataMap)
	if err != nil {
		return fmt.Errorf("load token metadata map: %w", err)
	}
	if len(currier) != tokenCount {
		return fmt.Errorf("token metadata map has %d tokens but frozen discovery reports token_count %d; observed and null denominators would disagree", len(currier), tokenCount)
	}
	p.update(1, 1, "Loading token metadata map")

	p.begin(3, "Computing observed statistics and block-aware permutation nulls")
	series, byWindowVector := RunSearchSpace(fs, currier, hand, c.Permutations, c.Seed, func(done, total int) {
		p.update(done, total, "Computing observed statistics and block-aware permutation nulls")
	})
	p.update(1, 1, "Computing observed statistics and block-aware permutation nulls")

	p.begin(4, "Aggregating empirical significance")
	p.update(1, 1, "Aggregating empirical significance")

	p.begin(5, "Writing results")
	if err := os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0755); err != nil {
		return err
	}
	writers := []func() error{
		func() error { return writeSummary(filepath.Join(c.OutputDir, "cluster_metadata_global_summary.tsv"), series) },
		func() error {
			return writeScalePersistence(filepath.Join(c.OutputDir, "cluster_metadata_scale_persistence.tsv"), series, byWindowVector)
		},
		func() error {
			return writePermutationsYAML(filepath.Join(c.OutputDir, "cluster_metadata_global_permutations.yaml"), c, series)
		},
		func() error { return writePlots(c.OutputDir, series, byWindowVector) },
		func() error { return updateValidationReport(c.MetadataReportPath, c, series) },
	}
	for _, w := range writers {
		if err := w(); err != nil {
			return err
		}
	}
	p.update(1, 1, "Writing results")
	fmt.Printf("Global multiple-comparison correction completed for %d frozen tokens over %d permutations; results written to %s\n", tokenCount, c.Permutations, c.OutputDir)
	return nil
}
