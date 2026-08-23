package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

// FrozenMarker is the sentinel task66 sections 39-41 require: Pareto
// candidates must be frozen (written once, never edited) before
// HELDOUT_RESULTS.tsv is opened, and no further candidate changes are
// permitted afterward.
const FrozenMarker = "PARETO_FROZEN"

// FreezeCandidates writes the frontier's names to the frozen-candidates
// sentinel. Calling this a second time with a different frontier is
// refused: task66 section 42 forbids reworking a candidate after
// held-out is opened, and re-freezing silently would defeat that.
func FreezeCandidates(dir string, frontier []string) error {
	path := filepath.Join(dir, FrozenMarker)
	sorted := append([]string(nil), frontier...)
	sort.Strings(sorted)
	content := strings.Join(sorted, "\n") + "\n"
	if existing, err := os.ReadFile(path); err == nil {
		if string(existing) != content {
			return fmt.Errorf("candidates already frozen with a different frontier; task66 section 42 forbids reworking candidates after freeze (existing=%q new=%q)", existing, content)
		}
		return nil
	}
	return os.WriteFile(path, []byte(content), 0644)
}

// ReadFrozenCandidates returns the frozen frontier, or an error if
// nothing has been frozen yet (task66 test 21: held-out evaluation
// rejects unfrozen candidates).
func ReadFrozenCandidates(dir string) ([]string, error) {
	b, err := os.ReadFile(filepath.Join(dir, FrozenMarker))
	if err != nil {
		return nil, fmt.Errorf("no frozen Pareto candidates found (%s missing): held-out evaluation is rejected until candidates are frozen (task66 section 41)", FrozenMarker)
	}
	var out []string
	for _, l := range strings.Split(strings.TrimSpace(string(b)), "\n") {
		if l != "" {
			out = append(out, l)
		}
	}
	return out, nil
}

// RunHeldout computes the HELDOUT-stage family metrics for exactly the
// frozen candidates (task66 sections 41-42): no parameter is touched
// here, only previously-frozen mechanisms are re-evaluated on the
// metrics that were withheld from every development-stage decision.
func RunHeldout(dir string, grid []GridEntry, corpora map[string]mechanismspace.Corpus, baselines map[string]mechanismspace.Fingerprint, targets []Target, devRows []FamilyMetricsRow, replicates int, opt mechanismspace.FingerprintOptions) ([]FamilyMetricsRow, map[string]string, error) {
	frontier, err := ReadFrozenCandidates(dir)
	if err != nil {
		return nil, nil, err
	}
	byName := map[string]GridEntry{}
	for _, e := range grid {
		byName[e.Name] = e
	}
	var frontierGrid []GridEntry
	for _, name := range frontier {
		if e, ok := byName[name]; ok {
			frontierGrid = append(frontierGrid, e)
		}
	}
	results := RunGrid(frontierGrid, corpora, replicates, opt, 800000)
	grouped := GroupByMechanismCorpus(results)
	heldoutRows := ComputeFamilyMetrics(grouped, baselines, targets, "HELDOUT", StageHeldout)

	overfit := map[string]string{}
	devByMech := map[string][]float64{}
	for _, r := range devRows {
		for _, v := range r.FamilyScores {
			devByMech[r.Mechanism] = append(devByMech[r.Mechanism], v)
		}
	}
	heldByMech := map[string][]float64{}
	for _, r := range heldoutRows {
		for _, v := range r.FamilyScores {
			heldByMech[r.Mechanism] = append(heldByMech[r.Mechanism], v)
		}
	}
	for _, name := range frontier {
		overfit[name] = OverfitClass(meanOf(devByMech[name]), meanOf(heldByMech[name]))
	}
	return heldoutRows, overfit, nil
}

func meanOf(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	s := 0.0
	for _, x := range v {
		s += x
	}
	return s / float64(len(v))
}

// WriteHeldoutTSV writes HELDOUT_RESULTS.tsv (task66 section 76).
func WriteHeldoutTSV(path string, rows []FamilyMetricsRow, overfit map[string]string) error {
	var b strings.Builder
	b.WriteString("mechanism\tcorpus\tfamily\tprogress\toverfit_class\n")
	for _, r := range rows {
		fams := make([]string, 0, len(r.FamilyScores))
		for f := range r.FamilyScores {
			fams = append(fams, f)
		}
		sort.Strings(fams)
		for _, f := range fams {
			b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%.9g\t%s\n", r.Mechanism, r.Corpus, f, r.FamilyScores[f], overfit[r.Mechanism]))
		}
	}
	return os.WriteFile(path, []byte(b.String()), 0644)
}

// OverfitClass is task66 section 42's DEVELOPMENT-vs-HELDOUT comparison.
func OverfitClass(devScore, heldoutScore float64) string {
	switch {
	case devScore > 0.15 && heldoutScore < 0:
		return "FINGERPRINT_OVERFIT"
	case devScore > 0.15 && heldoutScore > 0.15:
		return "CONFIRMED"
	default:
		return "INCONCLUSIVE"
	}
}
