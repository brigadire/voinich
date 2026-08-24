package task82a

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

func loadArtifacts(root string, manifest Manifest) (map[string]Artifact, error) {
	rawDir := filepath.Join(outDir(root), "raw")
	out := map[string]Artifact{}
	for _, j := range manifest.Jobs {
		b, err := os.ReadFile(filepath.Join(rawDir, j.JobID+".json"))
		if err != nil {
			return nil, fmt.Errorf("job %s not generated: %w", j.JobID, err)
		}
		var a Artifact
		if err := json.Unmarshal(b, &a); err != nil {
			return nil, err
		}
		out[j.JobID] = a
	}
	return out, nil
}

func f6(v float64) string {
	if math.IsNaN(v) {
		return "NA"
	}
	return strconv.FormatFloat(v, 'f', 6, 64)
}

// Aggregate regenerates every Task82a TSV/JSON/report from the raw
// artifacts on disk (task82a.txt sec.69-70: aggregates must be
// reproducible from raw data alone).
func Aggregate(root string, manifest Manifest) error {
	arts, err := loadArtifacts(root, manifest)
	if err != nil {
		return err
	}
	dir := outDir(root)

	if err := writeTransformation(dir, manifest, arts); err != nil {
		return err
	}
	if err := writeRecovery(dir, manifest, arts); err != nil {
		return err
	}
	if err := writeKnowledgeDependence(root, dir, manifest, arts); err != nil {
		return err
	}
	if err := writeCollisionScaling(dir, manifest, arts); err != nil {
		return err
	}
	if err := writeAmbiguityScaling(dir, manifest, arts); err != nil {
		return err
	}
	if err := writeInputDependence(dir, manifest, arts); err != nil {
		return err
	}
	if err := writeF2RawVectors(dir, manifest, arts); err != nil {
		return err
	}
	if err := writeF2Coverage(dir, manifest, arts); err != nil {
		return err
	}
	stab, err := writeF2Stability(dir, manifest, arts)
	if err != nil {
		return err
	}
	if err := writeMechanismEligibility(dir, manifest, arts, stab); err != nil {
		return err
	}
	if err := writeReport(dir, manifest, arts); err != nil {
		return err
	}
	if err := writeJobLedger(dir, manifest, arts); err != nil {
		return err
	}
	return nil
}

func writeJobLedger(dir string, manifest Manifest, arts map[string]Artifact) error {
	var b strings.Builder
	b.WriteString("job_id\tstatus\tinput_checksum\toutput_checksum\tf2_checksum\truntime_ns\tretries\tfailure_class\n")
	for _, j := range manifest.Jobs {
		a, ok := arts[j.JobID]
		status, failure := "COMPLETE", "NA"
		if !ok {
			status, failure = "MISSING", "NOT_GENERATED"
		} else if a.DocumentSHA256 != a.Document.Checksum() {
			status, failure = "FAILED", "CHECKSUM_MISMATCH"
		}
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\n", j.JobID, status, a.Corpus.SHA256, a.DocumentSHA256, a.F2.CorpusFileChecksum, a.RuntimeNS, 0, failure))
	}
	return os.WriteFile(filepath.Join(dir, "TASK82A_JOB_LEDGER.tsv"), []byte(b.String()), 0o644)
}

func writeTransformation(dir string, manifest Manifest, arts map[string]Artifact) error {
	var b strings.Builder
	b.WriteString("job_id\tmechanism_id\tscaling_policy_id\tcorpus_id\tcorpus_scale\treplicate\tchunks\tobservable_symbol_count\tobservable_token_count\tdistinct_symbols\tdistinct_tokens\tsymbol_entropy_bits\tconditional_symbol_entropy_bits\ttoken_entropy_bits\trepetition_rate\ttoken_repetition_rate\n")
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		m := a.Metrics
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\t%s\t%s\n",
			j.JobID, j.MechanismID, j.ScalingPolicyID, j.InputCorpusID, j.CorpusScale, j.Replicate, j.Chunks,
			m.SymbolCount, m.TokenCount, m.DistinctSymbols, m.DistinctTokens,
			f6(m.SymbolEntropyBits), f6(m.ConditionalEntropyBits), f6(m.TokenEntropyBits), f6(m.RepetitionRate), f6(m.TokenRepetitionRate)))
	}
	return os.WriteFile(filepath.Join(dir, "CORPUS_SCALE_TRANSFORMATION.tsv"), []byte(b.String()), 0o644)
}

var recoveryConditions = []string{"R0_FULL_KNOWLEDGE", "R1_NO_CONTEXT", "R2_NO_CONVENTION", "R3_NO_PATH_GEOMETRY", "R4_NO_HISTORY", "R5_NO_INTERNAL_MEMORY", "R6_OBSERVABLE_ONLY"}

func meanNonNaN(vals []float64) (float64, int) {
	sum, n := 0.0, 0
	for _, v := range vals {
		if !math.IsNaN(v) {
			sum += v
			n++
		}
	}
	if n == 0 {
		return math.NaN(), 0
	}
	return sum / float64(n), n
}

func writeRecovery(dir string, manifest Manifest, arts map[string]Artifact) error {
	var b strings.Builder
	b.WriteString("job_id\tmechanism_id\tscaling_policy_id\tcorpus_id\tcorpus_scale\treplicate\tcondition\tsampled_chunks\tlocal_exact_rate\tlocal_mean_recovery_score\tdocument_recovery_score\tmean_ambiguity_cardinality\n")
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		for _, cond := range recoveryConditions {
			var scores []float64
			exact, n := 0, 0
			ambSum, ambN := 0, 0
			for _, rec := range a.LocalRecoveries {
				sc, ok := rec.Scores[cond]
				if !ok {
					continue
				}
				scores = append(scores, sc)
				n++
				if rec.Classes[cond] == "EXACT" && sc == 1 {
					exact++
				}
				if amb, ok := rec.Ambiguity[cond]; ok && amb > 0 {
					ambSum += amb
					ambN++
				}
			}
			mean, _ := meanNonNaN(scores)
			exactRate := math.NaN()
			if n > 0 {
				exactRate = float64(exact) / float64(n)
			}
			ambMean := math.NaN()
			if ambN > 0 {
				ambMean = float64(ambSum) / float64(ambN)
			}
			// DOCUMENT_RECOVERY is explicitly the mean of LOCAL_RECOVERY
			// over the sampled chunks, never assumed equal to per-chunk
			// exact recovery (task82a.txt sec.36).
			b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%s\t%d\t%s\t%s\t%s\t%s\n",
				j.JobID, j.MechanismID, j.ScalingPolicyID, j.InputCorpusID, j.CorpusScale, j.Replicate, cond, n,
				f6(exactRate), f6(mean), f6(mean), f6(ambMean)))
		}
	}
	return os.WriteFile(filepath.Join(dir, "CORPUS_SCALE_RECOVERY.tsv"), []byte(b.String()), 0o644)
}

// bounded82KD reads Task82's own frozen bounded-adapter KD profile
// (task82a.txt sec.38: compare corpus-scale KD against the Task82 bounded
// profile) keyed by mechanism_id -> carrier delta name -> mean delta.
func bounded82KD(root string) (map[string]map[string]float64, error) {
	path := filepath.Join(root, "research", "phase2", "task82", "KNOWLEDGE_DEPENDENCE.tsv")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) < 2 {
		return map[string]map[string]float64{}, nil
	}
	header := strings.Split(lines[0], "\t")
	col := map[string]int{}
	for i, h := range header {
		col[h] = i
	}
	out := map[string]map[string]float64{}
	deltaCols := []string{"delta_context", "delta_convention", "delta_geometry", "delta_history", "delta_internal_memory"}
	for _, line := range lines[1:] {
		f := strings.Split(line, "\t")
		if len(f) <= col["mechanism_id"] {
			continue
		}
		mech := f[col["mechanism_id"]]
		if out[mech] == nil {
			out[mech] = map[string]float64{}
		}
		for _, dc := range deltaCols {
			idx, ok := col[dc]
			if !ok || idx >= len(f) {
				continue
			}
			v, err := strconv.ParseFloat(f[idx], 64)
			if err != nil {
				continue
			}
			// average across the bounded manifest's rows for this
			// mechanism/carrier (NA rows already fail ParseFloat and are
			// skipped, matching Task82's own NOT_APPLICABLE semantics).
			prev, seen := out[mech][dc+"_sum"]
			cnt := out[mech][dc+"_n"]
			if !seen {
				prev = 0
			}
			out[mech][dc+"_sum"] = prev + v
			out[mech][dc+"_n"] = cnt + 1
		}
	}
	final := map[string]map[string]float64{}
	for mech, vals := range out {
		final[mech] = map[string]float64{}
		for _, dc := range deltaCols {
			if n := vals[dc+"_n"]; n > 0 {
				final[mech][dc] = vals[dc+"_sum"] / n
			}
		}
	}
	return final, nil
}

func kdClass(bounded, corpusScale float64) string {
	if math.IsNaN(bounded) || math.IsNaN(corpusScale) {
		return "NOT_APPLICABLE"
	}
	boundedNonzero := math.Abs(bounded) > 1e-9
	scaleNonzero := math.Abs(corpusScale) > 1e-9
	if !boundedNonzero && !scaleNonzero {
		return "PRESERVED"
	}
	if boundedNonzero != scaleNonzero {
		return "NOT_PRESERVED"
	}
	if (bounded > 0) != (corpusScale > 0) {
		return "NOT_PRESERVED"
	}
	ratio := corpusScale / bounded
	if ratio >= 1.0/3.0 && ratio <= 3.0 {
		return "PRESERVED"
	}
	return "PARTIALLY_PRESERVED"
}

func writeKnowledgeDependence(root, dir string, manifest Manifest, arts map[string]Artifact) error {
	bounded, err := bounded82KD(root)
	if err != nil {
		return err
	}
	condToDelta := map[string]string{"R1_NO_CONTEXT": "delta_context", "R2_NO_CONVENTION": "delta_convention", "R3_NO_PATH_GEOMETRY": "delta_geometry", "R4_NO_HISTORY": "delta_history", "R5_NO_INTERNAL_MEMORY": "delta_internal_memory"}
	type key struct{ mech, policy string }
	r0ByJob := map[string]float64{}
	scoreByJobCond := map[string]map[string]float64{}
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		var r0 []float64
		byCond := map[string][]float64{}
		for _, rec := range a.LocalRecoveries {
			if v, ok := rec.Scores["R0_FULL_KNOWLEDGE"]; ok {
				r0 = append(r0, v)
			}
			for cond, v := range rec.Scores {
				byCond[cond] = append(byCond[cond], v)
			}
		}
		mr0, _ := meanNonNaN(r0)
		r0ByJob[j.JobID] = mr0
		scoreByJobCond[j.JobID] = map[string]float64{}
		for cond, vals := range byCond {
			m, _ := meanNonNaN(vals)
			scoreByJobCond[j.JobID][cond] = m
		}
	}
	groups := map[key][]string{}
	for _, j := range manifest.Jobs {
		k := key{j.MechanismID, j.ScalingPolicyID}
		groups[k] = append(groups[k], j.JobID)
	}
	var keys []key
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].mech != keys[j].mech {
			return keys[i].mech < keys[j].mech
		}
		return keys[i].policy < keys[j].policy
	})
	var b strings.Builder
	b.WriteString("mechanism_id\tscaling_policy_id\tcondition\tcorpus_scale_r0_mean\tcorpus_scale_delta_mean\ttask82_bounded_delta_mean\tkd_stability_class\n")
	for _, k := range keys {
		for cond, deltaCol := range condToDelta {
			var deltas []float64
			var r0s []float64
			for _, jobID := range groups[k] {
				rx, ok := scoreByJobCond[jobID][cond]
				if !ok {
					continue
				}
				r0 := r0ByJob[jobID]
				if math.IsNaN(r0) || math.IsNaN(rx) {
					continue
				}
				deltas = append(deltas, r0-rx)
				r0s = append(r0s, r0)
			}
			meanDelta, _ := meanNonNaN(deltas)
			meanR0, _ := meanNonNaN(r0s)
			b82 := math.NaN()
			if m, ok := bounded[k.mech]; ok {
				if v, ok := m[deltaCol]; ok {
					b82 = v
				}
			}
			b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\n", k.mech, k.policy, cond, f6(meanR0), f6(meanDelta), f6(b82), kdClass(b82, meanDelta)))
		}
	}
	return os.WriteFile(filepath.Join(dir, "KNOWLEDGE_DEPENDENCE_STABILITY.tsv"), []byte(b.String()), 0o644)
}

func writeCollisionScaling(dir string, manifest Manifest, arts map[string]Artifact) error {
	var b strings.Builder
	b.WriteString("job_id\tmechanism_id\tscaling_policy_id\tcorpus_id\tcorpus_scale\treplicate\tchunks\tlocal_collision_groups\tcolliding_chunks\tlocal_collision_rate\n")
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		colliding := 0
		for _, c := range a.LocalCollisions {
			colliding += len(c.ChunkIndices)
		}
		rate := 0.0
		if j.Chunks > 0 {
			rate = float64(colliding) / float64(j.Chunks)
		}
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\n",
			j.JobID, j.MechanismID, j.ScalingPolicyID, j.InputCorpusID, j.CorpusScale, j.Replicate, j.Chunks,
			len(a.LocalCollisions), colliding, f6(rate)))
	}
	if err := os.WriteFile(filepath.Join(dir, "COLLISION_SCALING.tsv"), []byte(b.String()), 0o644); err != nil {
		return err
	}
	return writeCrossCorpusCollisions(dir, manifest, arts)
}

func writeCrossCorpusCollisions(dir string, manifest Manifest, arts map[string]Artifact) error {
	type key struct {
		mech, policy, scale string
		rep                 int
	}
	byKey := map[key]map[string][]string{} // checksum -> corpus ids that produced a chunk with it
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		k := key{j.MechanismID, j.ScalingPolicyID, j.CorpusScale, j.Replicate}
		if byKey[k] == nil {
			byKey[k] = map[string][]string{}
		}
		for _, c := range a.Chunks {
			byKey[k][c.Checksum] = append(byKey[k][c.Checksum], j.InputCorpusID)
		}
	}
	var b strings.Builder
	b.WriteString("mechanism_id\tscaling_policy_id\tcorpus_scale\treplicate\tcross_corpus_collision_checksums\n")
	var keys []key
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j]) })
	for _, k := range keys {
		count := 0
		for _, corpora := range byKey[k] {
			distinct := map[string]bool{}
			for _, c := range corpora {
				distinct[c] = true
			}
			if len(distinct) > 1 {
				count++
			}
		}
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%d\t%d\n", k.mech, k.policy, k.scale, k.rep, count))
	}
	return os.WriteFile(filepath.Join(dir, "CROSS_CORPUS_COLLISIONS.tsv"), []byte(b.String()), 0o644)
}

func writeAmbiguityScaling(dir string, manifest Manifest, arts map[string]Artifact) error {
	var b strings.Builder
	b.WriteString("job_id\tmechanism_id\tscaling_policy_id\tcorpus_id\tcorpus_scale\treplicate\tchunks\tmean_ambiguity_cardinality_r6\tmax_ambiguity_cardinality_r6\n")
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		var vals []int
		for _, rec := range a.LocalRecoveries {
			if v, ok := rec.Ambiguity["R6_OBSERVABLE_ONLY"]; ok && v > 0 {
				vals = append(vals, v)
			}
		}
		mean, maxV := math.NaN(), 0
		if len(vals) > 0 {
			sum := 0
			for _, v := range vals {
				sum += v
				if v > maxV {
					maxV = v
				}
			}
			mean = float64(sum) / float64(len(vals))
		}
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%s\t%d\n", j.JobID, j.MechanismID, j.ScalingPolicyID, j.InputCorpusID, j.CorpusScale, j.Replicate, j.Chunks, f6(mean), maxV))
	}
	return os.WriteFile(filepath.Join(dir, "AMBIGUITY_SCALING.tsv"), []byte(b.String()), 0o644)
}

func writeInputDependence(dir string, manifest Manifest, arts map[string]Artifact) error {
	type key struct {
		mech, policy, scale string
		rep                 int
	}
	byKey := map[key]map[string]string{} // corpus -> document checksum
	for _, j := range manifest.Jobs {
		k := key{j.MechanismID, j.ScalingPolicyID, j.CorpusScale, j.Replicate}
		if byKey[k] == nil {
			byKey[k] = map[string]string{}
		}
		byKey[k][j.InputCorpusID] = arts[j.JobID].DocumentSHA256
	}
	var keys []key
	for k := range byKey {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return fmt.Sprint(keys[i]) < fmt.Sprint(keys[j]) })
	var b strings.Builder
	b.WriteString("mechanism_id\tscaling_policy_id\tcorpus_scale\treplicate\tdistinct_corpus_outputs\tinput_dependence\n")
	for _, k := range keys {
		distinct := map[string]bool{}
		for _, cs := range byKey[k] {
			distinct[cs] = true
		}
		class := "INPUT_INSENSITIVE"
		switch len(distinct) {
		case 3:
			class = "INPUT_SENSITIVE"
		case 2:
			class = "PARTIALLY_INPUT_SENSITIVE"
		}
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%s\n", k.mech, k.policy, k.scale, k.rep, len(distinct), class))
	}
	return os.WriteFile(filepath.Join(dir, "INPUT_DEPENDENCE.tsv"), []byte(b.String()), 0o644)
}

func writeF2RawVectors(dir string, manifest Manifest, arts map[string]Artifact) error {
	f, err := os.Create(filepath.Join(dir, "F2_RAW_VECTORS.jsonl"))
	if err != nil {
		return err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		row := map[string]any{
			"job_id": j.JobID, "mechanism_id": j.MechanismID, "scaling_policy_id": j.ScalingPolicyID,
			"corpus_id": j.InputCorpusID, "corpus_scale": j.CorpusScale, "replicate": j.Replicate,
			"corpus_file_sha256": a.F2.CorpusFileChecksum, "metrics": a.F2.Metrics,
		}
		if err := enc.Encode(row); err != nil {
			return err
		}
	}
	return nil
}

var coreFamilies = []string{"2DL", "BP", "EF", "HR", "LC", "LS", "PF"}

func writeF2Coverage(dir string, manifest Manifest, arts map[string]Artifact) error {
	var b strings.Builder
	b.WriteString("job_id\tmechanism_id\tscaling_policy_id\tcorpus_id\tcorpus_scale\treplicate\tcore_attempted\tcore_available\tsupporting_attempted\tsupporting_available\tattempted_families\tavailable_families\tnot_attempted_families\n")
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		coreAtt, coreAvail, suppAtt, suppAvail := 0, 0, 0, 0
		availFamilies := map[string]bool{}
		attFamilies := map[string]bool{}
		for _, m := range a.F2.Metrics {
			fam := f2FamilyCode(m.MetricID)
			attFamilies[fam] = true
			if m.Classification == "CORE" {
				coreAtt++
				if m.Available {
					coreAvail++
					availFamilies[fam] = true
				}
			} else {
				suppAtt++
				if m.Available {
					suppAvail++
					availFamilies[fam] = true
				}
			}
		}
		var attList, availList []string
		for f := range attFamilies {
			attList = append(attList, f)
		}
		for f := range availFamilies {
			availList = append(availList, f)
		}
		sort.Strings(attList)
		sort.Strings(availList)
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%d\t%s\t%s\t%s\n",
			j.JobID, j.MechanismID, j.ScalingPolicyID, j.InputCorpusID, j.CorpusScale, j.Replicate,
			coreAtt, coreAvail, suppAtt, suppAvail, strings.Join(attList, ","), strings.Join(availList, ","), strings.Join(NotAttemptedFamilies, ",")))
	}
	return os.WriteFile(filepath.Join(dir, "F2_COVERAGE.tsv"), []byte(b.String()), 0o644)
}

func f2FamilyCode(metricID string) string {
	if strings.HasPrefix(metricID, "cs") {
		return "CS"
	}
	if strings.HasPrefix(metricID, "EF") {
		return "EF"
	}
	if strings.HasPrefix(metricID, "LP") {
		return "LP"
	}
	i := strings.IndexAny(metricID, "0123456789_")
	if i > 0 {
		return metricID[:i]
	}
	return metricID
}

func stabilityClass(spread float64) string {
	switch {
	case math.IsNaN(spread):
		return "NOT_APPLICABLE"
	case spread <= 0.01:
		return "STABLE"
	case spread <= 0.10:
		return "PARTIALLY_STABLE"
	default:
		return "UNSTABLE"
	}
}

type stabilityResult struct {
	crossCorpus map[string]string // mechanism|policy|scale|metric -> class
	crossSeed   map[string]string
	crossScale  map[string]string // mechanism|policy|corpus|metric -> CONVERGED/...
}

// writeF2Stability computes cross-corpus, cross-seed, and cross-scale
// dispersion for every attempted, available F2 metric (task82a.txt
// sec.53-55) using the same STABLE/PARTIALLY_STABLE/UNSTABLE thresholds
// TASK82_DESIGN.md already froze for Task82.
func writeF2Stability(dir string, manifest Manifest, arts map[string]Artifact) (stabilityResult, error) {
	type mv struct {
		mech, policy, scale, corpus string
		rep                         int
		metric                      string
		value                       float64
		available                   bool
	}
	var rows []mv
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		for _, m := range a.F2.Metrics {
			rows = append(rows, mv{j.MechanismID, j.ScalingPolicyID, j.CorpusScale, j.InputCorpusID, j.Replicate, m.MetricID, m.Value, m.Available})
		}
	}

	res := stabilityResult{crossCorpus: map[string]string{}, crossSeed: map[string]string{}, crossScale: map[string]string{}}

	// Cross-corpus: group by mechanism/policy/scale/metric/replicate=0, spread across corpus.
	byCC := map[string][]float64{}
	for _, r := range rows {
		if !r.available || r.rep != 0 {
			continue
		}
		k := strings.Join([]string{r.mech, r.policy, r.scale, r.metric}, "|")
		byCC[k] = append(byCC[k], r.value)
	}
	var ccKeys []string
	for k := range byCC {
		ccKeys = append(ccKeys, k)
	}
	sort.Strings(ccKeys)
	var b strings.Builder
	b.WriteString("mechanism_id\tscaling_policy_id\tcorpus_scale\tmetric_id\tn\trange\tstability\n")
	for _, k := range ccKeys {
		vals := byCC[k]
		spread := rangeOf(vals)
		cls := stabilityClass(spread)
		res.crossCorpus[k] = cls
		parts := strings.Split(k, "|")
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%d\t%s\t%s\n", parts[0], parts[1], parts[2], parts[3], len(vals), f6(spread), cls))
	}
	if err := os.WriteFile(filepath.Join(dir, "F2_CROSS_CORPUS_STABILITY.tsv"), []byte(b.String()), 0o644); err != nil {
		return res, err
	}

	// Cross-seed: group by mechanism/policy/scale/corpus/metric, spread across replicate.
	byCS := map[string][]float64{}
	for _, r := range rows {
		if !r.available {
			continue
		}
		k := strings.Join([]string{r.mech, r.policy, r.scale, r.corpus, r.metric}, "|")
		byCS[k] = append(byCS[k], r.value)
	}
	var csKeys []string
	for k := range byCS {
		csKeys = append(csKeys, k)
	}
	sort.Strings(csKeys)
	var b2 strings.Builder
	b2.WriteString("mechanism_id\tscaling_policy_id\tcorpus_scale\tcorpus_id\tmetric_id\tn\trange\tstability\n")
	for _, k := range csKeys {
		vals := byCS[k]
		spread := rangeOf(vals)
		cls := stabilityClass(spread)
		res.crossSeed[k] = cls
		parts := strings.Split(k, "|")
		b2.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n", parts[0], parts[1], parts[2], parts[3], parts[4], len(vals), f6(spread), cls))
	}
	if err := os.WriteFile(filepath.Join(dir, "F2_CROSS_SEED_STABILITY.tsv"), []byte(b2.String()), 0o644); err != nil {
		return res, err
	}

	// Cross-scale: group by mechanism/policy/corpus/replicate/metric across MEDIUM,LARGE (the converged pair; see TASK82A_DESIGN.md).
	byScale := map[string]map[string]float64{} // group -> scaleID -> value
	for _, r := range rows {
		if !r.available {
			continue
		}
		g := strings.Join([]string{r.mech, r.policy, r.corpus, fmt.Sprint(r.rep), r.metric}, "|")
		if byScale[g] == nil {
			byScale[g] = map[string]float64{}
		}
		byScale[g][r.scale] = r.value
	}
	var scaleKeys []string
	for k := range byScale {
		scaleKeys = append(scaleKeys, k)
	}
	sort.Strings(scaleKeys)
	var b3 strings.Builder
	b3.WriteString("mechanism_id\tscaling_policy_id\tcorpus_id\treplicate\tmetric_id\tmedium\tlarge\tdelta\tconvergence\n")
	for _, g := range scaleKeys {
		vals := byScale[g]
		med, hasMed := vals["MEDIUM"]
		lg, hasLg := vals["LARGE"]
		conv := "NOT_APPLICABLE"
		delta := math.NaN()
		if hasMed && hasLg {
			delta = math.Abs(lg - med)
			switch {
			case delta <= 0.01:
				conv = "CONVERGED"
			case delta <= 0.10:
				conv = "PARTIALLY_CONVERGED"
			default:
				conv = "NOT_CONVERGED"
			}
		}
		res.crossScale[g] = conv
		parts := strings.Split(g, "|")
		medS, lgS := "NA", "NA"
		if hasMed {
			medS = f6(med)
		}
		if hasLg {
			lgS = f6(lg)
		}
		b3.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n", parts[0], parts[1], parts[2], parts[3], parts[4], medS, lgS, f6(delta), conv))
	}
	return res, os.WriteFile(filepath.Join(dir, "F2_CROSS_SCALE_STABILITY.tsv"), []byte(b3.String()), 0o644)
}

func rangeOf(vals []float64) float64 {
	if len(vals) == 0 {
		return math.NaN()
	}
	lo, hi := vals[0], vals[0]
	for _, v := range vals {
		if v < lo {
			lo = v
		}
		if v > hi {
			hi = v
		}
	}
	return hi - lo
}

func writeMechanismEligibility(dir string, manifest Manifest, arts map[string]Artifact, _ stabilityResult) error {
	type agg struct {
		coreAvailFamilies map[string]bool
		anyArtifactValid  bool
	}
	byMech := map[string]*agg{}
	for _, j := range manifest.Jobs {
		a := arts[j.JobID]
		ag, ok := byMech[j.MechanismID]
		if !ok {
			ag = &agg{coreAvailFamilies: map[string]bool{}}
			byMech[j.MechanismID] = ag
		}
		ag.anyArtifactValid = ag.anyArtifactValid || a.DocumentSHA256 == a.Document.Checksum()
		for _, m := range a.F2.Metrics {
			if m.Classification == "CORE" && m.Available {
				ag.coreAvailFamilies[f2FamilyCode(m.MetricID)] = true
			}
		}
	}
	var mechs []string
	for m := range byMech {
		mechs = append(mechs, m)
	}
	sort.Strings(mechs)
	var b strings.Builder
	b.WriteString("mechanism_id\tcore_families_available\tcore_family_coverage_ratio\tartifact_valid\teligibility\n")
	for _, mech := range mechs {
		ag := byMech[mech]
		ratio := float64(len(ag.coreAvailFamilies)) / float64(len(coreFamilies))
		elig := "NOT_COMPARABLE"
		switch {
		case ratio >= 0.5:
			elig = "F2_COMPARABLE"
		case ratio > 0:
			elig = "PARTIALLY_COMPARABLE"
		}
		var fams []string
		for f := range ag.coreAvailFamilies {
			fams = append(fams, f)
		}
		sort.Strings(fams)
		b.WriteString(fmt.Sprintf("%s\t%s\t%s\t%v\t%s\n", mech, strings.Join(fams, ","), f6(ratio), ag.anyArtifactValid, elig))
	}
	return os.WriteFile(filepath.Join(dir, "MECHANISM_ELIGIBILITY.tsv"), []byte(b.String()), 0o644)
}

func gitHead(root string) string {
	out, err := exec.Command("git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
