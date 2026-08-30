package notation

import (
	"fmt"
	"math"
	"math/rand"
	"sort"
	"strconv"
)

// CalibrationCorpora is the frozen minimum replicate count per generator
// (B01 section 28), fixed before any scale is computed.
const CalibrationCorpora = 40

// CalibrationGenerator is one frozen, versioned control-corpus generator
// (B01 section 26-27). Generate must depend only on its own parameters and
// seed: it must never read VM or candidate data.
type CalibrationGenerator struct {
	ID, Version         string
	AlphabetSize        int
	LinesPerCorpus      int
	MeanTokenLength     float64
	PreservedProperties string
	DestroyedProperties string
	Generate            func(g CalibrationGenerator, targetTokens int, seed int64) []Record
}

func alphabetSymbols(n int) []string {
	out := make([]string, n)
	for i := range n {
		out[i] = string(rune('a' + i))
	}
	return out
}

// genTokens builds USC records for one generated corpus: lines of
// generator-declared length until targetTokens is reached, using nextSymbol
// to pick each successive symbol from rng. It is shared by every sequence
// generator so their only difference is nextSymbol's memory.
func genTokens(corpusID, repID string, alphabet []string, linesPerCorpus int, meanTokenLen float64, targetTokens int, rng *rand.Rand, nextSymbol func(prevInLine []string) string) []Record {
	var out []Record
	produced := 0
	tokenIdx := 0
	lineIdx := 0
	var lineSyms []string
	flushToken := func() {
		if len(lineSyms) == 0 {
			return
		}
		tok := ""
		for _, s := range lineSyms {
			tok += s
		}
		out = append(out, Record{
			SchemaVersion: SchemaVersion, CorpusID: corpusID, Representation: repID,
			Document:     ObservedLevel{Value: "GEN-1", Observed: true},
			PhysicalLine: ObservedLevel{Value: strconv.Itoa(lineIdx), Observed: true},
			TokenID:      fmt.Sprintf("%s-%08d", corpusID, len(out)),
			TokenIndex:   tokenIdx, Token: tok, Symbols: append([]string(nil), lineSyms...),
		})
		tokenIdx++
		produced++
		lineSyms = nil
	}
	tokensThisLine := 0
	targetLineLen := poissonish(rng, meanTokenLen)
	for produced < targetTokens {
		sym := nextSymbol(lineSyms)
		lineSyms = append(lineSyms, sym)
		if len(lineSyms) >= targetLineLen {
			flushToken()
			tokensThisLine++
			if tokensThisLine >= 8 || lineIdx >= linesPerCorpus*1000 {
				lineIdx++
				tokenIdx = 0
				tokensThisLine = 0
			}
			targetLineLen = poissonish(rng, meanTokenLen)
		}
	}
	if len(lineSyms) > 0 {
		flushToken()
	}
	return out
}

// poissonish returns a small positive integer length with mean m using a
// deterministic geometric-like draw; it never returns less than 1.
func poissonish(rng *rand.Rand, m float64) int {
	if m < 1 {
		m = 1
	}
	p := 1 / m
	n := 1
	for rng.Float64() > p && n < 12 {
		n++
	}
	return n
}

func newRNG(seed int64) *rand.Rand { return rand.New(rand.NewSource(seed)) }

// CalibrationGenerators returns the frozen calibration panel (B01 section
// 26). Parameters are fixed here and must not be tuned against any VM or
// candidate distance.
func CalibrationGenerators() []CalibrationGenerator {
	alphabet := alphabetSymbols(24)
	return []CalibrationGenerator{
		{
			ID: "CAL-IID", Version: "1.0", AlphabetSize: len(alphabet), LinesPerCorpus: 800, MeanTokenLength: 5,
			PreservedProperties: "symbol frequency marginal", DestroyedProperties: "all sequence/token structure",
			Generate: func(g CalibrationGenerator, target int, seed int64) []Record {
				rng := newRNG(seed)
				return genTokens("CAL-IID", "SYNTHETIC-TOKEN", alphabet, g.LinesPerCorpus, g.MeanTokenLength, target, rng, func([]string) string {
					return alphabet[rng.Intn(len(alphabet))]
				})
			},
		},
		{
			ID: "CAL-MARKOV1", Version: "1.0", AlphabetSize: len(alphabet), LinesPerCorpus: 800, MeanTokenLength: 5,
			PreservedProperties: "first-order symbol transition matrix", DestroyedProperties: "second-order+ structure",
			Generate: func(g CalibrationGenerator, target int, seed int64) []Record {
				rng := newRNG(seed)
				trans := markovMatrix(rng, len(alphabet), 1)
				last := 0
				return genTokens("CAL-MARKOV1", "SYNTHETIC-TOKEN", alphabet, g.LinesPerCorpus, g.MeanTokenLength, target, rng, func([]string) string {
					last = weightedPick(rng, trans[last])
					return alphabet[last]
				})
			},
		},
		{
			ID: "CAL-MARKOV2", Version: "1.0", AlphabetSize: len(alphabet), LinesPerCorpus: 800, MeanTokenLength: 5,
			PreservedProperties: "second-order symbol transition structure", DestroyedProperties: "third-order+ structure",
			Generate: func(g CalibrationGenerator, target int, seed int64) []Record {
				rng := newRNG(seed)
				a := len(alphabet)
				trans := map[int][]float64{}
				last1, last2 := 0, 0
				return genTokens("CAL-MARKOV2", "SYNTHETIC-TOKEN", alphabet, g.LinesPerCorpus, g.MeanTokenLength, target, rng, func([]string) string {
					ctx := last1*a + last2
					row, ok := trans[ctx]
					if !ok {
						row = dirichletRow(rng, a, 0.3)
						trans[ctx] = row
					}
					next := weightedPick(rng, row)
					last1, last2 = last2, next
					return alphabet[next]
				})
			},
		},
		{
			ID: "CAL-CGRAMMAR", Version: "1.0", AlphabetSize: len(alphabet), LinesPerCorpus: 800, MeanTokenLength: 5,
			PreservedProperties: "frozen production-rule slot grammar (initial/medial/final symbol classes)", DestroyedProperties: "everything not expressible by the three-slot grammar",
			Generate: func(g CalibrationGenerator, target int, seed int64) []Record {
				rng := newRNG(seed)
				initClass, medClass, finClass := alphabet[:8], alphabet[8:16], alphabet[16:24]
				pos := 0
				return genTokens("CAL-CGRAMMAR", "SYNTHETIC-TOKEN", alphabet, g.LinesPerCorpus, g.MeanTokenLength, target, rng, func(prev []string) string {
					pos = len(prev)
					switch {
					case pos == 0:
						return initClass[rng.Intn(len(initClass))]
					default:
						if rng.Float64() < 0.35 {
							return finClass[rng.Intn(len(finClass))]
						}
						return medClass[rng.Intn(len(medClass))]
					}
				})
			},
		},
	}
}

// CalibrationDerivedGenerators returns the three shuffle-based controls that
// require an existing structured corpus to permute rather than generate
// from scratch (B01 CAL-TOKEN-SHUFFLE / CAL-WITHIN-TOKEN-SHUFFLE /
// CAL-LINE-SHUFFLE / CAL-HIERARCHY-SHUFFLE). base must be a structured,
// non-VM, non-candidate seed corpus (the CAL-MARKOV1 output is used as the
// shared base so no VM or candidate data ever enters the calibration
// panel).
func CalibrationDerivedGenerators(base []Record) []CalibrationGenerator {
	return []CalibrationGenerator{
		{ID: "CAL-TOKEN-SHUFFLE", Version: "1.0", PreservedProperties: "token-frequency marginal", DestroyedProperties: "sequence order", Generate: func(CalibrationGenerator, int, int64) []Record { return nil }},
		{ID: "CAL-WITHIN-TOKEN-SHUFFLE", Version: "1.0", PreservedProperties: "token length distribution", DestroyedProperties: "within-token symbol order", Generate: func(CalibrationGenerator, int, int64) []Record { return nil }},
		{ID: "CAL-LINE-SHUFFLE", Version: "1.0", PreservedProperties: "within-line structure", DestroyedProperties: "document progression", Generate: func(CalibrationGenerator, int, int64) []Record { return nil }},
		{ID: "CAL-HIERARCHY-SHUFFLE", Version: "1.0", PreservedProperties: "lower (line-internal) grammar", DestroyedProperties: "hierarchy/page progression", Generate: func(CalibrationGenerator, int, int64) []Record { return nil }},
	}
}

// DeriveShuffleCorpus applies the named CAL-*-SHUFFLE transform to base at
// the given seed. base itself must already be a calibration-generated
// corpus, never VM or candidate data.
func DeriveShuffleCorpus(id string, base []Record, seed int64) []Record {
	switch id {
	case "CAL-TOKEN-SHUFFLE":
		return ShuffleTokenOrder(base, seed)
	case "CAL-WITHIN-TOKEN-SHUFFLE":
		return ShuffleWithinTokens(base, seed)
	case "CAL-LINE-SHUFFLE":
		return ShuffleLines(base, seed)
	case "CAL-HIERARCHY-SHUFFLE":
		return ShufflePages(base, seed)
	}
	return nil
}

func markovMatrix(rng *rand.Rand, n, _ int) [][]float64 {
	m := make([][]float64, n)
	for i := range n {
		m[i] = dirichletRow(rng, n, 0.5)
	}
	return m
}
func dirichletRow(rng *rand.Rand, n int, concentration float64) []float64 {
	row := make([]float64, n)
	var sum float64
	for i := range n {
		g := gammaSample(rng, concentration)
		row[i] = g
		sum += g
	}
	if sum == 0 {
		for i := range row {
			row[i] = 1 / float64(n)
		}
		return row
	}
	for i := range row {
		row[i] /= sum
	}
	return row
}

// gammaSample draws a crude Gamma(shape,1) sample sufficient for generating
// non-uniform but bounded Dirichlet rows deterministically from rng.
func gammaSample(rng *rand.Rand, shape float64) float64 {
	if shape >= 1 {
		d := shape - 1.0/3
		c := 1 / math.Sqrt(9*d)
		for {
			x := rng.NormFloat64()
			v := 1 + c*x
			if v <= 0 {
				continue
			}
			v = v * v * v
			u := rng.Float64()
			if u < 1-0.0331*x*x*x*x || math.Log(u) < 0.5*x*x+d*(1-v+math.Log(v)) {
				return d * v
			}
		}
	}
	return gammaSample(rng, shape+1) * math.Pow(rng.Float64(), 1/shape)
}
func weightedPick(rng *rand.Rand, weights []float64) int {
	r := rng.Float64()
	var cum float64
	for i, w := range weights {
		cum += w
		if r <= cum {
			return i
		}
	}
	return len(weights) - 1
}

// CalibrationScale is one row of CALIBRATION_SCALES.tsv (B01 section 31).
type CalibrationScale struct {
	MetricID, MetricVersion, Family, SupportRegime string
	Checkpoint                                     int
	Estimator                                      string
	N                                               int
	Median, MAD, IQR, Scale                        float64
	Status                                          string
}

const (
	ScaleStatusOK        = "OK"
	ScaleStatusDegenerate = "DEGENERATE"
)

// CalibrationRun is one generated calibration corpus.
type CalibrationRun struct {
	GeneratorID string
	Replicate   int
	Records     []Record
}

// shuffleDerivedGeneratorIDs are the four CAL-*-SHUFFLE controls, each
// derived from its own replicate's CAL-MARKOV1 draw (B01 section 26).
var shuffleDerivedGeneratorIDs = []string{"CAL-TOKEN-SHUFFLE", "CAL-WITHIN-TOKEN-SHUFFLE", "CAL-LINE-SHUFFLE", "CAL-HIERARCHY-SHUFFLE"}

// RunCalibrationPanel generates the frozen calibration panel at one
// checkpoint size: `replicates` independent corpora for each of the four
// from-scratch generators, plus `replicates` corpora for each of the four
// shuffle-derived controls (each paired to its own replicate's CAL-MARKOV1
// draw so the panel never reads VM or candidate data). Every generator
// parameter and every seed is frozen; nothing here is tuned against a VM or
// candidate distance.
func RunCalibrationPanel(checkpoint, replicates int, baseSeed int64) []CalibrationRun {
	var runs []CalibrationRun
	markov1Base := make([][]Record, replicates)
	for _, g := range CalibrationGenerators() {
		for rep := range replicates {
			seed := SeedFor(baseSeed, g.ID, "SYNTHETIC-TOKEN", "CALIBRATION", checkpoint, rep)
			recs := g.Generate(g, checkpoint, seed)
			runs = append(runs, CalibrationRun{GeneratorID: g.ID, Replicate: rep, Records: recs})
			if g.ID == "CAL-MARKOV1" {
				markov1Base[rep] = recs
			}
		}
	}
	for _, id := range shuffleDerivedGeneratorIDs {
		for rep := range replicates {
			seed := SeedFor(baseSeed, id, "SYNTHETIC-TOKEN", "CALIBRATION", checkpoint, rep)
			recs := DeriveShuffleCorpus(id, markov1Base[rep], seed)
			runs = append(runs, CalibrationRun{GeneratorID: id, Replicate: rep, Records: recs})
		}
	}
	return runs
}

// calKey identifies one scalar calibration stratum below the checkpoint
// level (checkpoint is supplied separately by the caller).
type calKey struct{ metric, family, regime string }

// CalibrationRunMetrics is the cached scalar-metric output of one
// calibration corpus, computed once by AnalyzeCalibrationRuns so that
// BuildCalibrationScales and the C2/C3 leave-one-generator-out diagnostic
// never re-run the (expensive) generic analyzer.
type CalibrationRunMetrics struct {
	GeneratorID string
	Replicate   int
	Values      map[calKey]float64
}

// AnalyzeCalibrationRuns runs the generic analyzer exactly once per
// generated corpus and keeps only Comparable SCALAR-output-type metrics,
// since only those ever receive a calibrated scale (B01 sections 29-30).
func AnalyzeCalibrationRuns(runs []CalibrationRun) ([]CalibrationRunMetrics, error) {
	scalarType := map[string]bool{}
	for _, t := range MetricOutputTypes() {
		scalarType[t.MetricID] = t.OutputType == OutputScalar
	}
	out := make([]CalibrationRunMetrics, len(runs))
	for i, run := range runs {
		fp, err := Analyze(run.Records)
		if err != nil {
			return nil, fmt.Errorf("calibration corpus %s#%d: %w", run.GeneratorID, run.Replicate, err)
		}
		vals := map[calKey]float64{}
		for _, m := range fp.Metrics {
			if m.Status == Comparable && scalarType[m.MetricID] {
				vals[calKey{m.MetricID, m.Family, m.Regime}] = m.Value
			}
		}
		out[i] = CalibrationRunMetrics{GeneratorID: run.GeneratorID, Replicate: run.Replicate, Values: vals}
	}
	return out, nil
}

// BuildCalibrationScales pools every cached SCALAR metric value across the
// entire calibration panel (B01 sections 29-30) and estimates one frozen
// scale per (metric_id, family, support_regime) at this checkpoint.
func BuildCalibrationScales(runs []CalibrationRunMetrics, checkpoint int) []CalibrationScale {
	values := map[calKey][]float64{}
	var order []calKey
	for _, run := range runs {
		for k, v := range run.Values {
			if _, ok := values[k]; !ok {
				order = append(order, k)
			}
			values[k] = append(values[k], v)
		}
	}
	sort.Slice(order, func(i, j int) bool {
		if order[i].metric != order[j].metric {
			return order[i].metric < order[j].metric
		}
		return order[i].regime < order[j].regime
	})
	var out []CalibrationScale
	for _, k := range order {
		s := EstimateScale(values[k]) // EstimateScale sorts its input, so map-iteration order above never affects the result.
		s.MetricID, s.Family, s.SupportRegime, s.Checkpoint, s.MetricVersion = k.metric, k.family, k.regime, checkpoint, MetricRegistryVersion
		out = append(out, s)
	}
	return out
}

// LeaveOneGeneratorFamilyOut recomputes CalibrationScale for one
// (metric,family,regime) stratum after excluding one generator id. It is a
// stability diagnostic only (B01 C3): it must never be used to pick an
// estimator.
func LeaveOneGeneratorFamilyOut(runs []CalibrationRunMetrics, excludeGeneratorID string, checkpoint int) []CalibrationScale {
	var kept []CalibrationRunMetrics
	for _, r := range runs {
		if r.GeneratorID != excludeGeneratorID {
			kept = append(kept, r)
		}
	}
	return BuildCalibrationScales(kept, checkpoint)
}

// ScalesFromCalibration converts the frozen CALIBRATION_SCALES.tsv rows at
// one checkpoint into the []Scale shape Compare consumes, keeping only
// non-degenerate strata. A candidate is compared at the checkpoint closest
// to its own rarefied/actual size; VM itself is compared at its full
// observed size (39,380).
func ScalesFromCalibration(cs []CalibrationScale, checkpoint int) []Scale {
	var out []Scale
	for _, s := range cs {
		if s.Checkpoint != checkpoint || s.Status != ScaleStatusOK {
			continue
		}
		out = append(out, Scale{MetricID: s.MetricID, Regime: s.SupportRegime, Center: s.Median, Spread: s.Scale})
	}
	return out
}

// EstimateScale computes the frozen robust scale s_i = 1.4826*MAD(X); if
// MAD is zero it falls back to IQR/1.349; if that is also zero the metric is
// explicitly DEGENERATE and never produces a normalized scalar distance
// (B01 section 29). No epsilon is invented post hoc.
func EstimateScale(values []float64) CalibrationScale {
	v := append([]float64(nil), values...)
	sort.Float64s(v)
	n := len(v)
	if n == 0 {
		return CalibrationScale{Status: ScaleStatusDegenerate, Estimator: "MAD_1.4826"}
	}
	med := percentile(v, 0.5)
	dev := make([]float64, n)
	for i, x := range v {
		dev[i] = math.Abs(x - med)
	}
	sort.Float64s(dev)
	mad := percentile(dev, 0.5)
	q1, q3 := percentile(v, 0.25), percentile(v, 0.75)
	iqr := q3 - q1
	s := CalibrationScale{Estimator: "MAD_1.4826", N: n, Median: med, MAD: mad, IQR: iqr}
	switch {
	case mad > 0:
		s.Scale = 1.4826 * mad
		s.Status = ScaleStatusOK
	case iqr > 0:
		s.Estimator = "IQR_1.349_FALLBACK"
		s.Scale = iqr / 1.349
		s.Status = ScaleStatusOK
	default:
		s.Status = ScaleStatusDegenerate
	}
	return s
}
