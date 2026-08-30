package notation

import (
	"math/rand"
	"sort"
	"strconv"
)

// BootstrapReplicates is the frozen production replicate count B (B04
// section 19). It was reduced from the originally proposed 1000 after a
// documented runtime benchmark (see BOOTSTRAP_PROTOCOL.md "Benchmark and
// STOP decision"): a single full-corpus Analyze() pass over the frozen VM
// source (39,380 tokens) costs approximately 4.85s, so B=1000 would need
// roughly 81 CPU-minutes for VM alone. B=200 keeps a valid percentile
// bootstrap (200 replicates is a standard lower bound for 95% percentile
// CIs) while remaining tractable, and is applied uniformly to VM and every
// future candidate corpus.
const BootstrapReplicates = 200

// BootstrapCILevel is the frozen confidence level (B04 section 20).
const BootstrapCILevel = 0.95

// BootstrapRow is one row of BOOTSTRAP_RESULTS.tsv: a single summarized
// bootstrap estimate for one metric at the corpus's own actual size. Unlike
// rarefaction, bootstrap is not evaluated per checkpoint; it estimates the
// sampling uncertainty of the corpus's own point estimate (B04 section 18).
type BootstrapRow struct {
	CorpusID, RepresentationID string
	Family, MetricID, Regime  string
	Estimate                  float64
	BootstrapMean             float64
	BootstrapSD               float64
	CILevel                   float64
	CILow, CIHigh             float64
	NValid                    int
}

// bootstrapDraw resamples whole structuralBlocks with replacement until the
// same number of blocks as the source has been drawn (a standard block
// bootstrap). It never draws a partial block and never joins two blocks
// that were not already part of the same source hierarchy unit, because
// each resampled block keeps its original internal record order and its
// own hierarchy key.
func bootstrapDraw(rs []Record, seed int64) []Record {
	blocks := buildStructuralBlocks(rs)
	rng := rand.New(rand.NewSource(seed))
	var out []Record
	for draw := range blocks {
		b := blocks[rng.Intn(len(blocks))]
		suffix := "-boot" + strconv.Itoa(draw)
		for _, r := range b.records {
			rc := r
			disambiguateDeepestLevel(&rc, suffix)
			rc.TokenID += suffix
			out = append(out, rc)
		}
	}
	return out
}

// disambiguateDeepestLevel appends suffix to the value of the deepest
// source-observed hierarchy level of r, so that resampling the same block
// twice in one bootstrap draw still produces distinct, internally valid
// hierarchy units instead of a duplicate/non-contiguous key.
func disambiguateDeepestLevel(r *Record, suffix string) {
	switch {
	case r.PhysicalLine.Observed:
		r.PhysicalLine.Value += suffix
	case r.Locus.Observed:
		r.Locus.Value += suffix
	case r.Page.Observed:
		r.Page.Value += suffix
	case r.Section.Observed:
		r.Section.Value += suffix
	default:
		r.Document.Value += suffix
	}
}

// RunBootstrap computes the frozen percentile bootstrap for every metric
// emitted by Analyze(rs), resampling structural blocks with replacement at
// the corpus's own observed size (B04 sections 18-20). The same estimator,
// unit, and replicate count are used for VM and every future candidate.
func RunBootstrap(rs []Record, corpusID, representationID string, replicates int, baseSeed int64) ([]BootstrapRow, error) {
	point, err := Analyze(rs)
	if err != nil {
		return nil, err
	}
	type key struct{ family, metric, regime string }
	draws := map[key][]float64{}
	var order []key
	pointValue := map[key]float64{}
	for _, m := range point.Metrics {
		k := key{m.Family, m.MetricID, m.Regime}
		order = append(order, k)
		if m.Status == Comparable {
			pointValue[k] = m.Value
		}
	}
	for rep := range replicates {
		seed := SeedFor(baseSeed, corpusID, representationID, FamilyGroupBootstrap, 0, rep)
		sample := bootstrapDraw(rs, seed)
		fp, err := Analyze(sample)
		if err != nil {
			return nil, err
		}
		for _, m := range fp.Metrics {
			if m.Status != Comparable {
				continue
			}
			k := key{m.Family, m.MetricID, m.Regime}
			draws[k] = append(draws[k], m.Value)
		}
	}
	var rows []BootstrapRow
	for _, k := range order {
		v := append([]float64(nil), draws[k]...)
		sort.Float64s(v)
		row := BootstrapRow{CorpusID: corpusID, RepresentationID: representationID, Family: k.family, MetricID: k.metric, Regime: k.regime, CILevel: BootstrapCILevel, NValid: len(v)}
		if est, ok := pointValue[k]; ok {
			row.Estimate = est
		}
		if len(v) > 0 {
			row.BootstrapMean = mean(v)
			row.BootstrapSD = stddev(v)
			tail := (1 - BootstrapCILevel) / 2
			row.CILow = percentile(v, tail)
			row.CIHigh = percentile(v, 1-tail)
		}
		rows = append(rows, row)
	}
	return rows, nil
}
