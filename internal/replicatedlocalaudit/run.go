package replicatedlocalaudit

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func newCheckpoint(fp string, c Config) checkpoint {
	return checkpoint{Version: 1, Fingerprint: fp, Permutations: c.Permutations, Distance: map[string][]float64{}, ShuffleExceedBlocks: map[string]int{}, ShuffleExceedTotal: map[string]int{}, ShuffleSumBlocks: map[string]float64{}, ShuffleSumTotal: map[string]float64{}, MarkovExceedBlocks: map[string]int{}, MarkovExceedTotal: map[string]int{}, MarkovSumBlocks: map[string]float64{}, MarkovSumTotal: map[string]float64{}}
}
func loadCheckpoint(path, fp string, c Config) (checkpoint, bool, error) {
	cp := newCheckpoint(fp, c)
	if path == "" {
		return cp, false, nil
	}
	b, e := os.ReadFile(path)
	if os.IsNotExist(e) {
		return cp, false, nil
	}
	if e != nil {
		return cp, false, e
	}
	if e = json.Unmarshal(b, &cp); e != nil {
		return newCheckpoint(fp, c), false, nil
	}
	if cp.Version != 1 || cp.Fingerprint != fp || cp.Permutations != c.Permutations {
		return newCheckpoint(fp, c), false, nil
	}
	return cp, true, nil
}
func saveCheckpoint(path string, cp checkpoint) error {
	if path == "" {
		return nil
	}
	if e := os.MkdirAll(filepath.Dir(path), 0755); e != nil {
		return e
	}
	b, e := json.MarshalIndent(cp, "", "  ")
	if e != nil {
		return e
	}
	tmp := path + ".tmp"
	if e = os.WriteFile(tmp, b, 0644); e != nil {
		return e
	}
	return os.Rename(tmp, path)
}

func RunAndWrite(c Config) error {
	if c.Permutations <= 0 {
		return fmt.Errorf("permutations must be positive")
	}
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)
	p.begin(1, "Loading and validating frozen inputs")
	state, corpusSHA, e := buildDistributionState(c)
	if e != nil {
		return e
	}
	tokens, blocks, dc, sc := state.tokens, state.blocks, state.dc, state.sc
	frozenSimilarity, frozenCrossCurrier, frozenCrossHand, frozenLOBOSuccess, e := loadFrozenDistanceDiagnostics(c)
	if e != nil {
		return e
	}
	fp, e := fingerprint(c, corpusSHA)
	if e != nil {
		return e
	}
	cp, resumed, e := loadCheckpoint(c.CheckpointPath, fp, c)
	if e != nil {
		return e
	}
	if resumed && c.ProgressWriter != nil {
		fmt.Fprintf(c.ProgressWriter, "Resuming checkpoint %s: distance %d, shuffle %d, Markov %d replicates complete\n", c.CheckpointPath, cp.DistanceCompleted, cp.ShuffleCompleted, cp.MarkovCompleted)
	}
	p.update(1, 1, "Frozen inputs loaded")

	p.begin(2, "Building block profiles and frozen observations")
	profiles, refs := state.profiles, state.refs
	counts := map[string]map[string]int{}
	for _, b := range blocks {
		counts[b.ID] = map[string]int{}
		for _, t := range b.Tokens {
			counts[b.ID][t.Text]++
		}
	}
	observed := map[string]float64{}
	for _, d := range dc {
		if d.Q > .05 {
			continue
		}
		for _, b := range blocks {
			if counts[b.ID][d.A] >= 10 && counts[b.ID][d.B] >= 10 {
				v, _ := compareProfiles(profiles[b.ID], refs[b.ID], d.A, d.B)
				observed[d.ID+"\x00"+b.ID] = v
			}
		}
	}
	p.update(1, 1, "Block profiles built")

	executor := permutationExecutorFor(c, state)

	p.begin(3, "Frequency-matched distance null")
	if cp.DistanceCompleted >= c.Permutations {
		p.update(cp.DistanceCompleted, c.Permutations, "Distance null replicates")
	}
	e = runBattery(executorContext(c), executor, "distance", cp.DistanceCompleted, c.Permutations, executorWorkers(c), func(run int, res ReplicateResult) error {
		for k, v := range res.Distance {
			cp.Distance[k] = append(cp.Distance[k], v)
		}
		cp.DistanceCompleted = run + 1
		if err := saveCheckpoint(c.CheckpointPath, cp); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}
		p.update(cp.DistanceCompleted, c.Permutations, "Distance null replicates")
		return nil
	})
	if e != nil {
		return e
	}

	p.begin(4, "Within-block shuffle sequence null")
	if cp.ShuffleCompleted >= c.Permutations {
		p.update(cp.ShuffleCompleted, c.Permutations, "Sequence shuffle replicates")
	}
	seqObs := map[string]seqObserved{}
	for _, s := range sc {
		seqObs[s.ID] = sequenceObserved(s, tokens, blocks)
	}
	e = runBattery(executorContext(c), executor, "shuffle", cp.ShuffleCompleted, c.Permutations, executorWorkers(c), func(run int, res ReplicateResult) error {
		for _, s := range sc {
			o := seqObs[s.ID]
			cp.ShuffleSumBlocks[s.ID] += float64(res.ShuffleBlocks[s.ID])
			cp.ShuffleSumTotal[s.ID] += float64(res.ShuffleTotal[s.ID])
			if res.ShuffleBlocks[s.ID] >= o.Blocks {
				cp.ShuffleExceedBlocks[s.ID]++
			}
			if res.ShuffleTotal[s.ID] >= o.Total {
				cp.ShuffleExceedTotal[s.ID]++
			}
		}
		cp.ShuffleCompleted = run + 1
		if err := saveCheckpoint(c.CheckpointPath, cp); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}
		p.update(cp.ShuffleCompleted, c.Permutations, "Sequence shuffle replicates")
		return nil
	})
	if e != nil {
		return e
	}

	p.begin(5, "Leakage-free first-order Markov null")
	if cp.MarkovCompleted >= c.Permutations {
		p.update(cp.MarkovCompleted, c.Permutations, "Markov replicates")
	}
	markovAvailable := len(state.markovTraining)
	e = runBattery(executorContext(c), executor, "markov", cp.MarkovCompleted, c.Permutations, executorWorkers(c), func(run int, res ReplicateResult) error {
		for _, s := range sc {
			cp.MarkovSumBlocks[s.ID] += float64(res.MarkovBlocks[s.ID])
			cp.MarkovSumTotal[s.ID] += float64(res.MarkovTotal[s.ID])
			if res.MarkovBlocks[s.ID] >= res.MarkovObservedBlocks[s.ID] {
				cp.MarkovExceedBlocks[s.ID]++
			}
			if res.MarkovTotal[s.ID] >= res.MarkovObservedTotal[s.ID] {
				cp.MarkovExceedTotal[s.ID]++
			}
		}
		cp.MarkovCompleted = run + 1
		if err := saveCheckpoint(c.CheckpointPath, cp); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}
		p.update(cp.MarkovCompleted, c.Permutations, "Markov replicates")
		return nil
	})
	if e != nil {
		return e
	}

	p.begin(6, "Computing replication diagnostics")
	dres := buildDistanceResults(c, dc, blocks, profiles, refs, counts, observed, frozenSimilarity, frozenCrossCurrier, frozenCrossHand, frozenLOBOSuccess, cp)
	sres := buildSequenceResults(c, sc, seqObs, cp, markovAvailable, blocks)
	p.update(1, 1, "Diagnostics computed")
	p.begin(7, "Writing confirmatory tables")
	if e = writeOutputs(c, corpusSHA, dres, sres); e != nil {
		return e
	}
	p.update(1, 1, "Confirmatory tables written")
	p.begin(8, "Finalizing reproducibility record")
	if c.CheckpointPath != "" {
		if e = os.Remove(c.CheckpointPath); e != nil && !os.IsNotExist(e) {
			return e
		}
	}
	p.update(1, 1, "Audit complete")
	return nil
}

func freqMatch(a, b int) bool { return a > 0 && b > 0 && float64(max(a, b))/float64(min(a, b)) <= 2 }
func sequenceStatsOne(s sequenceCandidate, all, available []block) [2]int {
	ids := map[string]bool{}
	for _, b := range available {
		ids[b.ID] = true
	}
	total, bc := 0, 0
	for _, b := range all {
		if !ids[b.ID] {
			continue
		}
		n := countSequence(b.Tokens, s.Tokens)
		total += n
		if n > 0 {
			bc++
		}
	}
	return [2]int{total, bc}
}

func buildDistanceResults(c Config, cs []distanceCandidate, blocks []block, profiles, refs map[string]profile, counts map[string]map[string]int, obs, frozenSimilarity map[string]float64, frozenCrossCurrier, frozenCrossHand, frozenLOBOSuccess map[string]bool, cp checkpoint) []distanceResult {
	byID := map[string]block{}
	for _, b := range blocks {
		byID[b.ID] = b
	}
	var out []distanceResult
	for _, d := range cs {
		r := distanceResult{Candidate: d}
		if d.Q > .05 {
			r.Status = "NOT_SIGNIFICANT"
			out = append(out, r)
			continue
		}
		totalObs := 0
		for _, b := range blocks {
			if counts[b.ID][d.A] < 10 || counts[b.ID][d.B] < 10 {
				continue
			}
			key := d.ID + "\x00" + b.ID
			nm, ns := meanSD(cp.Distance[key])
			v := obs[key]
			z := 0.
			if ns > 0 {
				z = (v - nm) / ns
			}
			shape := pairShape(profiles[b.ID], d.A, d.B)
			refShape := pairShape(refs[b.ID], d.A, d.B)
			peak, center, asym := shapeStats(shape)
			_, nctx := compareProfiles(profiles[b.ID], refs[b.ID], d.A, d.B)
			row := distanceRow{CandidateID: d.ID, A: d.A, B: d.B, Block: b.ID, Currier: b.Currier, Hand: b.Hand, Joint: b.Joint, CountA: counts[b.ID][d.A], CountB: counts[b.ID][d.B], Observations: counts[b.ID][d.A] + counts[b.ID][d.B], ComparedProfileCells: nctx, LOBO: v, Similarity: frozenSimilarity[key], ShapeSimilarity: cosine(shape, refShape), Peak: peak, Center: center, Asymmetry: asym, Transfer: frozenLOBOSuccess[key], NullMean: nm, NullSD: ns, P95: quantile(cp.Distance[key], .95), P99: quantile(cp.Distance[key], .99), Standardized: z, Effect: v - nm}
			r.Rows = append(r.Rows, row)
			totalObs += nctx
			if row.Effect > 0 {
				r.Positive++
			} else if row.Effect < 0 {
				r.Negative++
			}
		}
		zs := []float64{}
		effects := []float64{}
		weighted := 0.
		effectAbs := 0.
		maxEffect := 0.
		maxObs := 0
		for _, x := range r.Rows {
			zs = append(zs, x.Standardized)
			effects = append(effects, x.Effect)
			weighted += x.Standardized * float64(x.Observations)
			effectAbs += math.Abs(x.Effect)
			if math.Abs(x.Effect) > maxEffect {
				maxEffect = math.Abs(x.Effect)
			}
			if x.Observations > maxObs {
				maxObs = x.Observations
			}
		}
		r.MeanZ, _ = meanSD(zs)
		if totalObs > 0 {
			r.WeightedZ = weighted / float64(totalObs)
			r.MaxObservationFraction = float64(maxObs) / float64(totalObs)
		}
		if effectAbs > 0 {
			r.MaxEffectContribution = maxEffect / effectAbs
		}
		r.FullEffect, _ = meanSD(effects)
		var jk []float64
		allSignificant := len(effects) >= 2
		for i := range effects {
			var z []float64
			z = append(z, effects[:i]...)
			z = append(z, effects[i+1:]...)
			m, _ := meanSD(z)
			jk = append(jk, m)
			observedMean := 0.0
			for j, row := range r.Rows {
				if j != i {
					observedMean += row.LOBO
				}
			}
			observedMean /= float64(max(1, len(r.Rows)-1))
			exceed := 0
			for run := 0; run < c.Permutations; run++ {
				nullMean, n := 0.0, 0
				for j, row := range r.Rows {
					if j == i {
						continue
					}
					vals := cp.Distance[d.ID+"\x00"+row.Block]
					if run < len(vals) {
						nullMean += vals[run]
						n++
					}
				}
				if n > 0 && nullMean/float64(n) >= observedMean {
					exceed++
				}
			}
			jp := float64(exceed+1) / float64(c.Permutations+1)
			if jp > r.MaxJackknifeP {
				r.MaxJackknifeP = jp
			}
			if jp > .05 {
				allSignificant = false
			}
		}
		if len(jk) > 0 {
			sort.Float64s(jk)
			r.MinJackknife, r.MaxJackknife = jk[0], jk[len(jk)-1]
			_, r.JackknifeSD = meanSD(jk)
			r.JackknifeSurvives = allSignificant
		}
		r.WithinCurrier, r.CrossCurrier = metadataEffects(r.Rows, func(x distanceRow) string { return x.Currier })
		r.WithinHand, r.CrossHand = metadataEffects(r.Rows, func(x distanceRow) string { return x.Hand })
		r.WithinJoint, r.CrossJoint = metadataEffects(r.Rows, func(x distanceRow) string { return x.Joint })
		failed := []string{}
		if d.Eligible < 3 {
			failed = append(failed, "eligible_blocks<3")
		}
		if d.Joint < 2 {
			failed = append(failed, "joint_classes<2")
		}
		if d.Median < .7 {
			failed = append(failed, "profile_stability<0.7")
		}
		if d.Transfer < .67 {
			failed = append(failed, "transfer_success<0.67")
		}
		// In generic mode there is no real hand dimension (loadFrozenDistance
		// Diagnostics leaves frozenCrossHand always empty in that mode), so the
		// failure condition and replication tier below are decided on the
		// single group/joint signal alone - never on a degenerate always-false
		// cross-hand check, which would otherwise silently mislabel every
		// finding as metadata-limited rather than honestly untested.
		if c.Generic {
			if !frozenCrossCurrier[d.ID] {
				failed = append(failed, "no_cross_group_transfer")
			}
		} else if !frozenCrossCurrier[d.ID] && !frozenCrossHand[d.ID] {
			failed = append(failed, "no_cross_currier_or_hand_transfer")
		}
		r.FailedConditions = strings.Join(failed, ";")
		positiveFraction := float64(r.Positive) / float64(max(1, len(r.Rows)))
		if r.MaxObservationFraction > .7 || r.MaxEffectContribution > .7 || !r.JackknifeSurvives {
			r.Status = "SINGLE_BLOCK_DRIVEN"
		} else if d.Eligible >= 3 && d.Joint >= 2 && d.Transfer >= .67 && positiveFraction >= .75 && r.JackknifeSurvives {
			if c.Generic {
				if r.CrossJoint {
					r.Status = "GROUP_REPLICATED"
				} else {
					r.Status = "GROUP_LIMITED_REPLICATION"
				}
			} else if r.CrossCurrier && r.CrossHand && r.CrossJoint {
				r.Status = "ROBUST_RELATIVE_REPLICATION"
			} else {
				r.Status = "METADATA_LIMITED_REPLICATION"
			}
		} else {
			r.Status = "STATISTICALLY_SIGNIFICANT_BUT_UNSTABLE"
		}
		out = append(out, r)
	}
	return out
}
func metadataEffects(rows []distanceRow, key func(distanceRow) string) (within, cross bool) {
	groups := map[string][]float64{}
	for _, x := range rows {
		groups[key(x)] = append(groups[key(x)], x.Effect)
	}
	positive := 0
	for _, v := range groups {
		m, _ := meanSD(v)
		if len(v) >= 2 && m > 0 {
			within = true
		}
		if m > 0 {
			positive++
		}
	}
	cross = positive >= 2
	return
}

func buildSequenceResults(c Config, cs []sequenceCandidate, obs map[string]seqObserved, cp checkpoint, av int, blocks []block) []sequenceResult {
	out := make([]sequenceResult, len(cs))
	p := make([]float64, len(cs))
	for i, s := range cs {
		o := obs[s.ID]
		x := sequenceResult{Candidate: s, Observed: o, MarkovAvailableBlocks: av}
		x.ShuffleP = float64(cp.ShuffleExceedBlocks[s.ID]+1) / float64(c.Permutations+1)
		x.ShuffleTotalP = float64(cp.ShuffleExceedTotal[s.ID]+1) / float64(c.Permutations+1)
		x.ShuffleMeanBlocks = cp.ShuffleSumBlocks[s.ID] / float64(c.Permutations)
		x.ShuffleMeanTotal = cp.ShuffleSumTotal[s.ID] / float64(c.Permutations)
		availObs := sequenceStatsOne(s, blocks, availableMarkovBlocks(blocks))
		if av > 0 {
			x.MarkovP = float64(cp.MarkovExceedBlocks[s.ID]+1) / float64(c.Permutations+1)
			x.MarkovTotalP = float64(cp.MarkovExceedTotal[s.ID]+1) / float64(c.Permutations+1)
			x.MarkovMeanBlocks = cp.MarkovSumBlocks[s.ID] / float64(c.Permutations)
			x.MarkovMeanTotal = cp.MarkovSumTotal[s.ID] / float64(c.Permutations)
		} else {
			x.MarkovP = math.NaN()
			x.MarkovTotalP = math.NaN()
		}
		_ = availObs
		p[i] = x.ShuffleP
		out[i] = x
	}
	q := bh(p)
	for i := range out {
		out[i].ShuffleQ = q[i]
		o := out[i].Observed
		if o.Validity == "absent-from-canonical" {
			out[i].Status = "INVALID_CANONICAL_SUPPORT"
		} else if o.Validity == "ambiguous-transcription" {
			out[i].Status = "TRANSCRIPTION_AMBIGUOUS"
		} else if o.MaxFraction > .7 {
			out[i].Status = "BLOCK_CONCENTRATED"
		} else if o.Blocks < 3 || o.Joint < 2 {
			out[i].Status = "INSUFFICIENT_REPLICATION"
		} else if q[i] <= .05 {
			out[i].Status = "REPLICATED_ABOVE_FREQUENCY_NULL"
		} else {
			out[i].Status = "REPLICATED_BUT_EXPECTED_FROM_FREQUENCY"
		}
	}
	return out
}
func availableMarkovBlocks(blocks []block) []block {
	var out []block
	for _, h := range blocks {
		for _, b := range blocks {
			if b.ID != h.ID && b.Joint == h.Joint {
				out = append(out, h)
				break
			}
		}
	}
	return out
}
