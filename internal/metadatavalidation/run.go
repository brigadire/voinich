package metadatavalidation

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"gopkg.in/yaml.v3"
)

type runResult struct {
	Doc                 Document
	Alignment           AlignmentResult
	References          map[string][]MetadataBoundary
	Stable              []StableBoundary
	BoundaryRows        []BoundaryValidation
	BoundaryNull        map[string][][]float64
	Associations        []Association
	Assignments         []Assignment
	AssignmentNull      map[string][]float64
	DiscoveryTokenCount int
}

func defaults(c Config) Config {
	if c.IVTFFPath == "" {
		c.IVTFFPath = "data/ZL3b-n.txt"
	}
	if c.FrozenCorpusPath == "" {
		c.FrozenCorpusPath = "data_work/ZL3b-x7.txt"
	}
	if c.DiscoveryDir == "" {
		c.DiscoveryDir = "workdir"
	}
	if c.OutputDir == "" {
		c.OutputDir = "workdir"
	}
	if c.Permutations <= 0 {
		c.Permutations = 10000
	}
	if len(c.Tolerances) == 0 {
		c.Tolerances = []int{10, 25, 50, 100, 200}
	}
	return c
}

func RunAndWrite(c Config) error {
	c = defaults(c)
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)
	p.begin(1, "Parsing IVTFF metadata")
	doc, e := ParseIVTFF(c.IVTFFPath)
	if e != nil {
		return fmt.Errorf("parse IVTFF: %w", e)
	}
	p.update(1, 1, "Parsing IVTFF metadata")
	p.begin(2, "Strictly aligning frozen corpus")
	tokens, hash, e := ReadFrozenCorpus(c.FrozenCorpusPath)
	if e != nil {
		return fmt.Errorf("read frozen corpus: %w", e)
	}
	aligned, e := Align(doc, tokens, hash)
	if e != nil {
		_ = os.MkdirAll(c.OutputDir, 0755)
		_ = writeFailedAlignmentReport(filepath.Join(c.OutputDir, "alignment_report.md"), c, doc, aligned, e)
		return e
	}
	p.update(1, 1, "Strictly aligning frozen corpus")
	p.begin(3, "Loading frozen discovery")
	tokenCount, e := discoveryTokenCount(filepath.Join(c.DiscoveryDir, "global_distributional_regimes.yaml"))
	if e != nil {
		return e
	}
	if tokenCount != len(tokens) {
		return fmt.Errorf("frozen discovery token_count %d is incompatible with corpus token_count %d", tokenCount, len(tokens))
	}
	stable, e := LoadStable(filepath.Join(c.DiscoveryDir, "stable_distributional_boundaries.tsv"))
	if e != nil {
		return fmt.Errorf("load stable boundaries: %w", e)
	}
	assign, e := LoadAssignments(filepath.Join(c.DiscoveryDir, "global_distributional_cluster_assignments.tsv"))
	if e != nil {
		return fmt.Errorf("load cluster assignments: %w", e)
	}
	for _, name := range []string{"global_distributional_change_points.tsv", "global_distributional_clustering.tsv", "global_distributional_windows.tsv", "global_distributional_report.md"} {
		if _, e = os.Stat(filepath.Join(c.DiscoveryDir, name)); e != nil {
			return fmt.Errorf("required frozen result %s: %w", name, e)
		}
	}
	p.update(1, 1, "Loading frozen discovery")
	p.begin(4, "Validating blind boundaries")
	refs := ExtractBoundaries(aligned.Records)
	boundaryRows, bnull := ValidateBoundaries(stable, refs, c.Tolerances, c.Permutations, len(tokens), c.Seed)
	p.update(1, 1, "Validating blind boundaries")
	p.begin(5, "Associating frozen clusters with metadata")
	assoc := AnalyzeAssignments(assign, aligned.Records)
	p.update(1, 1, "Associating frozen clusters with metadata")
	p.begin(6, "Block-aware metadata permutation controls")
	anull := clusterPermutationSummary(assign, aligned.Records, c.Permutations, c.Seed+9001, p)
	p.update(1, 1, "Block-aware metadata permutation controls")
	r := runResult{doc, aligned, refs, stable, boundaryRows, bnull, assoc, assign, anull, tokenCount}
	p.begin(7, "Writing results")
	if e = writeAll(c, r); e != nil {
		return e
	}
	p.update(1, 1, "Writing results")
	fmt.Printf("Blind metadata validation completed for %d frozen tokens; results written to %s\n", len(tokens), c.OutputDir)
	return nil
}

func discoveryTokenCount(path string) (int, error) {
	b, e := os.ReadFile(path)
	if e != nil {
		return 0, fmt.Errorf("read frozen discovery metadata: %w", e)
	}
	var x struct {
		Parameters struct {
			TokenCount int `yaml:"token_count"`
		} `yaml:"parameters"`
	}
	if e = yaml.Unmarshal(b, &x); e != nil {
		return 0, e
	}
	if x.Parameters.TokenCount <= 0 {
		return 0, fmt.Errorf("frozen discovery metadata has no token_count")
	}
	return x.Parameters.TokenCount, nil
}

// To keep the max-over-K null faithful and tractable, use one representative
// window scale (200 when present) and each frozen method. Each permutation is
// shared by all K, then the maximum NMI is retained.
func clusterPermutationSummary(a []Assignment, records []TokenMetadata, n int, seed int64, p *progressReporter) map[string][]float64 {
	out := map[string][]float64{}
	rng := rand.New(rand.NewSource(seed))
	for _, kind := range []string{"currier", "hand"} {
		for _, method := range []string{"hierarchical", "k_medoids", "contiguous_segmentation"} {
			byK := map[int][]Assignment{}
			for _, x := range a {
				if x.WindowSize == 200 && x.Method == method {
					byK[x.K] = append(byK[x.K], x)
				}
			}
			if len(byK) == 0 {
				continue
			}
			ks := make([]int, 0, len(byK))
			for k := range byK {
				ks = append(ks, k)
			}
			sort.Ints(ks)
			base := byK[ks[0]]
			labels := make([]string, len(base))
			for i, x := range base {
				labels[i] = MetadataComposition(records, x.Start, x.End, kind).Label
			}
			vals := make([]float64, n)
			for z := 0; z < n; z++ {
				permuted := PermuteBlockLabels(labels, rng)
				best := 0.
				for _, k := range ks {
					clusters := make([]string, len(byK[k]))
					for i, x := range byK[k] {
						clusters[i] = strconv.Itoa(x.Cluster)
					}
					v := AssociationMetrics(permuted, clusters).NMI
					if v > best {
						best = v
					}
				}
				vals[z] = best
				if p != nil {
					p.update(z+1, n*6, "Block-aware metadata permutation controls")
				}
			}
			out[kind+"/"+method+"/window_200/max_nmi_over_k"] = vals
		}
	}
	return out
}

func writeAll(c Config, r runResult) error {
	if e := os.MkdirAll(filepath.Join(c.OutputDir, "plots"), 0755); e != nil {
		return e
	}
	writers := []func() error{
		func() error { return writeAlignmentReport(filepath.Join(c.OutputDir, "alignment_report.md"), c, r) }, func() error {
			return writeTokenMap(filepath.Join(c.OutputDir, "token_metadata_map.tsv"), r.Alignment.Records)
		}, func() error {
			return writeMetadataBoundaries(filepath.Join(c.OutputDir, "metadata_boundaries.tsv"), r.References)
		}, func() error {
			return writeBoundaryValidation(filepath.Join(c.OutputDir, "boundary_validation.tsv"), r.BoundaryRows)
		}, func() error {
			return writeYAML(filepath.Join(c.OutputDir, "boundary_validation_permutations.yaml"), map[string]any{"seed": c.Seed, "permutations": c.Permutations, "null_models": map[string]string{"uniform": "same count sampled without replacement from positions 1..N-1", "circular_shift": "entire blind-boundary pattern shifted modulo N-1"}, "distributions": r.BoundaryNull})
		}, func() error {
			return writeAssociations(filepath.Join(c.OutputDir, "cluster_metadata_association.tsv"), r.Associations)
		}, func() error {
			return writeWindowMetadata(filepath.Join(c.OutputDir, "window_metadata_composition.tsv"), r.Alignment.Records, r.Assignments)
		}, func() error {
			return writeYAML(filepath.Join(c.OutputDir, "cluster_metadata_permutations.yaml"), map[string]any{"seed": c.Seed + 9001, "permutations": c.Permutations, "method": "shuffle labels among contiguous metadata blocks; preserve block lengths; retain maximum NMI over frozen K=2..15 at window size 200", "max_nmi_null": r.AssignmentNull})
		}, func() error {
			return writeStrongest(filepath.Join(c.OutputDir, "strongest_boundaries_metadata.tsv"), r.Stable, r.References, false)
		}, func() error {
			return writeStrongest(filepath.Join(c.OutputDir, "unexplained_distributional_structure.tsv"), r.Stable, r.References, true)
		}, func() error { return writeMetadataYAML(filepath.Join(c.OutputDir, "metadata_validation.yaml"), c, r) }, func() error {
			return writeValidationReport(filepath.Join(c.OutputDir, "metadata_validation_report.md"), c, r)
		}, func() error {
			return writeKindValidation(filepath.Join(c.OutputDir, "currier_validation.tsv"), "currier", r.BoundaryRows)
		}, func() error {
			return writeKindValidation(filepath.Join(c.OutputDir, "hand_validation.tsv"), "hand", r.BoundaryRows)
		}, func() error {
			return writeKindValidation(filepath.Join(c.OutputDir, "quire_validation.tsv"), "quire", r.BoundaryRows)
		}, func() error { return writePlots(c.OutputDir, r) }}
	for _, w := range writers {
		if e := w(); e != nil {
			return e
		}
	}
	return nil
}

func writeYAML(path string, v any) error {
	b, e := yaml.Marshal(v)
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0644)
}
func tsv(path, header string, rows []string) error {
	f, e := os.Create(path)
	if e != nil {
		return e
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	defer w.Flush()
	fmt.Fprintln(w, header)
	for _, x := range rows {
		fmt.Fprintln(w, x)
	}
	return nil
}
func clean(s string) string { return strings.NewReplacer("\t", " ", "\n", " ", "\r", " ").Replace(s) }
func f(v float64) string    { return strconv.FormatFloat(v, 'g', -1, 64) }

func writeTokenMap(path string, x []TokenMetadata) error {
	rows := make([]string, len(x))
	for i, v := range x {
		quire := v.Quire
		if quire == "" {
			quire = "null"
		}
		rows[i] = fmt.Sprintf("%d\t%s\t%s\t%s\t%s\t%s\t%d\t%t\t%s\t%s\t%s\t%d\t%d\t%d", v.Position, clean(v.Token), v.Folio, v.LocusID, v.LocusType, v.LineID, v.ParagraphID, v.ParagraphStart, v.Currier, v.Hand, quire, v.IndexInLocus, v.IndexInLine, v.IndexInFolio)
	}
	return tsv(path, "token_position\ttoken\tfolio\tlocus_id\tlocus_type\tline_id\tparagraph_id\tparagraph_start\tcurrier\thand\tquire\ttoken_index_in_locus\ttoken_index_in_line\ttoken_index_in_folio", rows)
}
func writeMetadataBoundaries(path string, x map[string][]MetadataBoundary) error {
	rows := []string{}
	for _, kind := range metadataKinds {
		for _, v := range x[kind] {
			rows = append(rows, fmt.Sprintf("%d\t%s\t%s\t%s", v.Position, v.Kind, v.From, v.To))
		}
	}
	return tsv(path, "position\tmetadata_type\tfrom\tto", rows)
}
func writeBoundaryValidation(path string, x []BoundaryValidation) error {
	rows := []string{}
	for _, v := range x {
		rows = append(rows, fmt.Sprintf("%s\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s", v.Kind, v.MinSupport, v.Tolerance, v.BlindCount, v.Matched, f(v.MatchFraction), f(v.MeanDistance), f(v.MedianDistance), f(v.UniformMean), f(v.UniformPercentile), f(v.CircularMean), f(v.CircularPercentile)))
	}
	return tsv(path, "metadata_type\tmin_scale_support\ttolerance\tblind_boundary_count\tmatched_count\tmatch_fraction\tmean_nearest_distance\tmedian_nearest_distance\tuniform_expected_matches\tuniform_percentile\tcircular_expected_matches\tcircular_percentile", rows)
}
func writeKindValidation(path, kind string, x []BoundaryValidation) error {
	rows := []string{}
	for _, v := range x {
		if v.Kind == kind {
			rows = append(rows, fmt.Sprintf("%d\t%d\t%d\t%d\t%s\t%s\t%s", v.MinSupport, v.Tolerance, v.BlindCount, v.Matched, f(v.MatchFraction), f(v.UniformPercentile), f(v.CircularPercentile)))
		}
	}
	return tsv(path, "min_scale_support\ttolerance\tblind_boundary_count\tmatched_count\tmatch_fraction\tuniform_percentile\tcircular_percentile", rows)
}
func writeAssociations(path string, x []Association) error {
	rows := []string{}
	for _, v := range x {
		rows = append(rows, fmt.Sprintf("%d\t%s\t%d\t%s\t%s\t%d\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s", v.WindowSize, v.Method, v.K, v.Metadata, v.Subset, v.Windows, f(v.MI), f(v.NMI), f(v.ARI), f(v.Homogeneity), f(v.Completeness), f(v.ConditionalEntropy), f(v.EntropyReduction), v.Contingency))
	}
	return tsv(path, "window_size\tmethod\tk\tmetadata\tsubset\twindows\tmutual_information\tnormalized_mutual_information\tadjusted_rand_index\thomogeneity\tcompleteness\tconditional_entropy\tnormalized_entropy_reduction\tcontingency", rows)
}

func writeWindowMetadata(path string, records []TokenMetadata, assignments []Assignment) error {
	type window struct{ size, index, start, end int }
	unique := map[window]bool{}
	for _, a := range assignments {
		unique[window{a.WindowSize, a.Index, a.Start, a.End}] = true
	}
	windows := make([]window, 0, len(unique))
	for w := range unique {
		windows = append(windows, w)
	}
	sort.Slice(windows, func(i, j int) bool {
		if windows[i].size != windows[j].size {
			return windows[i].size < windows[j].size
		}
		return windows[i].index < windows[j].index
	})
	rows := make([]string, 0, len(windows))
	composition := func(m map[string]float64) string {
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		x := make([]string, 0, len(keys))
		for _, k := range keys {
			x = append(x, k+":"+f(m[k]))
		}
		return strings.Join(x, ",")
	}
	for _, w := range windows {
		c := MetadataComposition(records, w.start, w.end, "currier")
		h := MetadataComposition(records, w.start, w.end, "hand")
		rows = append(rows, fmt.Sprintf("%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\t%s", w.size, w.index, w.start, w.end, c.Label, f(c.Purity), composition(c.Composition), h.Label, f(h.Purity), composition(h.Composition)))
	}
	return tsv(path, "window_size\twindow_index\tstart\tend\tcurrier_majority_label\tcurrier_purity\tcurrier_composition\thand_majority_label\thand_purity\thand_composition", rows)
}
