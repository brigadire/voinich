package notation

import (
	"fmt"
	"math/rand"
	"sort"
)

// structuralBlock is the boundary-preserving sampling unit: the contiguous
// run of records sharing the deepest source-observed hierarchy key
// (lineKey). When physical lines are observed this is exactly one physical
// line; when they are not observed it falls back to the deepest level that
// is (locus, page, section, or the whole document), because every deeper
// level collapses to the empty string uniformly. No sampling procedure ever
// splits a block.
type structuralBlock struct {
	key     string
	order   int
	records []Record
	tokens  int
}

func buildStructuralBlocks(rs []Record) []structuralBlock {
	idx := map[string]int{}
	var blocks []structuralBlock
	for i, r := range rs {
		k := lineKey(r)
		if j, ok := idx[k]; ok {
			blocks[j].records = append(blocks[j].records, r)
			blocks[j].tokens++
			continue
		}
		idx[k] = len(blocks)
		blocks = append(blocks, structuralBlock{key: k, order: i, records: []Record{r}, tokens: 1})
	}
	return blocks
}

func linesObserved(rs []Record) bool {
	for _, r := range rs {
		if !r.PhysicalLine.Observed {
			return false
		}
	}
	return true
}

// RarefactionDraw is one boundary-preserving sample.
type RarefactionDraw struct {
	Records    []Record
	RequestedN int
	ActualN    int
}

// Rarefy draws a deterministic, boundary-preserving sample targeting
// requestedN tokens by adding whole structuralBlocks, shuffled by seed,
// until the closest valid total is reached (B03 section 7). It never splits
// a block and never fabricates a transition between blocks that were not
// already adjacent in rs, because every block boundary is a genuine
// hierarchy-key change.
func Rarefy(rs []Record, requestedN int, seed int64) (RarefactionDraw, error) {
	blocks := buildStructuralBlocks(rs)
	total := 0
	for _, b := range blocks {
		total += b.tokens
	}
	if total < requestedN {
		return RarefactionDraw{}, fmt.Errorf("corpus has %d tokens, below checkpoint %d", total, requestedN)
	}
	order := make([]int, len(blocks))
	for i := range order {
		order[i] = i
	}
	rand.New(rand.NewSource(seed)).Shuffle(len(order), func(i, j int) { order[i], order[j] = order[j], order[i] })

	selected := map[int]bool{}
	cum := 0
	for _, bi := range order {
		if cum >= requestedN {
			break
		}
		prev := cum
		next := cum + blocks[bi].tokens
		if next < requestedN {
			selected[bi] = true
			cum = next
			continue
		}
		// next >= requestedN: choose the closer of {prev, next}. Ties and
		// the case prev==0 (a single block already meeting or exceeding the
		// checkpoint) resolve to inclusion, per the frozen overshoot policy
		// "include next complete unit if this gives the closest valid
		// sample" (B03 section 7).
		devIncl := next - requestedN
		devExcl := requestedN - prev
		if prev > 0 && devExcl < devIncl {
			break
		}
		selected[bi] = true
		cum = next
		break
	}
	idxs := make([]int, 0, len(selected))
	for bi := range selected {
		idxs = append(idxs, bi)
	}
	sort.Slice(idxs, func(i, j int) bool { return blocks[idxs[i]].order < blocks[idxs[j]].order })
	var out []Record
	for _, bi := range idxs {
		out = append(out, blocks[bi].records...)
	}
	return RarefactionDraw{Records: out, RequestedN: requestedN, ActualN: cum}, nil
}

// RarefactionRow is one row of RAREFACTION.tsv.
type RarefactionRow struct {
	CorpusID, RepresentationID       string
	Family, MetricID, Regime         string
	CheckpointRequested              int
	CheckpointActual                 int
	Replicate                        int
	Seed                             int64
	Value                            float64
	Comparable                       bool
}

// RarefactionSummaryRow is one row of RAREFACTION_SUMMARY.tsv.
type RarefactionSummaryRow struct {
	CorpusID, RepresentationID string
	Family, MetricID, Regime  string
	Checkpoint                int
	Mean, Median, SD          float64
	CILow, CIHigh             float64
	NValid                    int
}

// RarefactionReplicates is the frozen replicate count R (B03 section 8).
const RarefactionReplicates = 100

// RunRarefaction executes the frozen rarefaction protocol for every
// checkpoint and replicate, and returns both the raw rows and the summary.
// It applies the identical algorithm to VM, calibration controls, and any
// future candidate corpus (R6 symmetry).
func RunRarefaction(rs []Record, corpusID, representationID string, checkpoints []int, replicates int, baseSeed int64) ([]RarefactionRow, []RarefactionSummaryRow, error) {
	total := len(rs)
	haveLines := linesObserved(rs)
	var rows []RarefactionRow
	type key struct{ family, metric, regime string; checkpoint int }
	values := map[key][]float64{}
	var order []key

	record := func(fam, metricID, regime string, checkpoint, actualN, replicate int, seed int64, m Metric) {
		row := RarefactionRow{CorpusID: corpusID, RepresentationID: representationID, Family: fam, MetricID: metricID, Regime: regime, CheckpointRequested: checkpoint, CheckpointActual: actualN, Replicate: replicate, Seed: seed}
		if m.Status == Comparable {
			row.Comparable = true
			row.Value = m.Value
		}
		rows = append(rows, row)
	}

	for _, n := range checkpoints {
		if total < n {
			// NOT_COMPARABLE checkpoint: emit nothing but a single documented
			// marker row per family group so the schema is auditable.
			for _, fam := range []string{"G", "T", "S", "L", "D"} {
				rows = append(rows, RarefactionRow{CorpusID: corpusID, RepresentationID: representationID, Family: fam, MetricID: "NOT_COMPARABLE", CheckpointRequested: n, CheckpointActual: 0, Replicate: -1, Seed: 0, Comparable: false})
			}
			continue
		}
		for rep := range replicates {
			structSeed := SeedFor(baseSeed, corpusID, representationID, FamilyGroupStructural, n, rep)
			draw, err := Rarefy(rs, n, structSeed)
			if err != nil {
				return nil, nil, err
			}
			fp, err := Analyze(draw.Records)
			if err != nil {
				return nil, nil, err
			}
			for _, m := range fp.Metrics {
				if m.Family == "L" {
					continue // L is drawn separately below with its own seed
				}
				record(m.Family, m.MetricID, m.Regime, n, draw.ActualN, rep, structSeed, m)
				k := key{m.Family, m.MetricID, m.Regime, n}
				if _, ok := values[k]; !ok {
					order = append(order, k)
				}
				if m.Status == Comparable {
					values[k] = append(values[k], m.Value)
				} else if _, ok := values[k]; !ok {
					values[k] = []float64{}
				}
			}
			a2, a3, at := AccumulationCounts(draw.Records)
			for id, v := range map[string]float64{"A2_BIGRAM_TYPES": float64(a2), "A3_TRIGRAM_TYPES": float64(a3), "AT_TRANSITION_TYPES": float64(at)} {
				record("CURVE", id, "", n, draw.ActualN, rep, structSeed, Metric{Value: v, Status: Comparable})
				k := key{"CURVE", id, "", n}
				if _, ok := values[k]; !ok {
					order = append(order, k)
				}
				values[k] = append(values[k], v)
			}
			if !haveLines {
				for _, id := range []string{"L01_LINE_TOKEN_COUNT_MEAN", "L02_LINE_SYMBOL_COUNT_MEAN", "L03_BOUNDARY_SPECIALIZATION", "L04_POSITION_PROGRESSION", "L05_LINE_ASYMMETRY", "L06_SAME_LINE_COOCCURRENCE_DENSITY", "L07_SAME_LINE_NONCOOCCURRENCE_DENSITY"} {
					record("L", id, "", n, 0, rep, 0, Metric{Status: NotComparable})
				}
				continue
			}
			lineSeed := SeedFor(baseSeed, corpusID, representationID, FamilyGroupLine, n, rep)
			lineDraw, err := Rarefy(rs, n, lineSeed)
			if err != nil {
				return nil, nil, err
			}
			lfp, err := Analyze(lineDraw.Records)
			if err != nil {
				return nil, nil, err
			}
			for _, m := range lfp.Metrics {
				if m.Family != "L" {
					continue
				}
				record("L", m.MetricID, m.Regime, n, lineDraw.ActualN, rep, lineSeed, m)
				k := key{"L", m.MetricID, m.Regime, n}
				if _, ok := values[k]; !ok {
					order = append(order, k)
				}
				if m.Status == Comparable {
					values[k] = append(values[k], m.Value)
				}
			}
		}
	}
	var summary []RarefactionSummaryRow
	for _, k := range order {
		v := append([]float64(nil), values[k]...)
		sort.Float64s(v)
		s := RarefactionSummaryRow{CorpusID: corpusID, RepresentationID: representationID, Family: k.family, MetricID: k.metric, Regime: k.regime, Checkpoint: k.checkpoint, NValid: len(v)}
		if len(v) > 0 {
			s.Mean, s.Median, s.SD = mean(v), percentile(v, 0.5), stddev(v)
			s.CILow, s.CIHigh = percentile(v, 0.025), percentile(v, 0.975)
		}
		summary = append(summary, s)
	}
	return rows, summary, nil
}

// percentile returns the linear-interpolated percentile p (0..1) of a
// pre-sorted slice.
func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return 0
	}
	if len(sorted) == 1 {
		return sorted[0]
	}
	pos := p * float64(len(sorted)-1)
	lo := int(pos)
	hi := lo + 1
	if hi >= len(sorted) {
		return sorted[len(sorted)-1]
	}
	frac := pos - float64(lo)
	return sorted[lo]*(1-frac) + sorted[hi]*frac
}
