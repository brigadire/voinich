package main

import (
	"fmt"
	"math"
	"os"
	"sort"
	"sync/atomic"
	"time"
)

var predictiveMetricNames = []string{"PM1", "PM2", "PM4", "PM5", "PM6"}

type popRun struct {
	base CalibPopulationBaselines
	jobs map[string]CalibJobResult // candidateID -> result
}

// StageAResult holds everything needed to write the calibration output
// tables and freeze marker.
type StageAResult struct {
	Thresholds []CalibThreshold
	Jobs       []CalibJobResult // full ledger, all generators/populations/candidates
	TotalJobs  int
	FailedJobs int
}

func metricValue(pm PredictiveMetrics, name string, pm6 float64, pm6ok bool) (float64, bool) {
	switch name {
	case "PM1":
		return pm.PM1, true
	case "PM2":
		return pm.PM2, true
	case "PM4":
		return pm.PM4, !math.IsNaN(pm.PM4)
	case "PM5":
		return pm.PM5, !math.IsNaN(pm.PM5)
	case "PM6":
		return pm6, pm6ok
	}
	return 0, false
}

func runStageA(namespace string, candidates []Candidate, workDir string) StageAResult {
	generators := []string{"MFC0", "MFC1", "MFC2"}
	runs := map[string][]popRun{} // generator -> 16 popRuns

	var allJobs []CalibJobResult
	var done int64
	total := int64(len(generators) * 16)
	start := time.Now()
	for _, gen := range generators {
		popResults := make([]popRun, 16)
		parallelFor(16, func(idx int) {
			popIdx := idx + 1
			pop := generateMFCPopulation(namespace, gen, popIdx)
			bitsReal := bitsPerRealParameter(len(pop.Dev))
			alias := NewGlyphAlias(mfcAlphabet)
			base := computeCalibBaselines(namespace, candidates, pop.Dev, pop.Val, pop.Heldout, bitsReal, alias, gen, popIdx, workDir)
			jobs := map[string]CalibJobResult{}
			for _, cand := range candidates {
				jobs[cand.CandidateID] = runCalibJob(namespace, cand, pop.Dev, pop.Val, pop.Heldout, bitsReal, alias, gen, popIdx, workDir)
			}
			popResults[idx] = popRun{base: base, jobs: jobs}
			n := atomic.AddInt64(&done, 1)
			fmt.Fprintf(os.Stderr, "[stageA] %s population %d done (%d/%d, elapsed %s)\n", gen, popIdx, n, total, time.Since(start).Round(time.Second))
		})
		runs[gen] = popResults
	}

	byGenerator := map[string]map[string][]float64{}
	for _, gen := range generators {
		byGenerator[gen] = map[string][]float64{}
		for _, cand := range candidates {
			// predictive-gain vs B1 and (where applicable) B2, per metric.
			for _, metric := range predictiveMetricNames {
				var vsB1, vsB2 []float64
				var overfit []float64
				for _, pr := range runs[gen] {
					job, ok := pr.jobs[cand.CandidateID]
					if !ok || job.Failed {
						continue
					}
					cv, cok := metricValue(job.HeldPM, metric, job.PM6, job.PM6Valid)
					if !cok {
						continue
					}
					if !pr.base.B1.TrainingFailed {
						b1v, b1ok := metricValue(pr.base.B1BasePM(), metric, pr.base.B1PM6(), pr.base.B1PM6OK())
						if b1ok {
							vsB1 = append(vsB1, math.Abs(cv-b1v))
						}
					}
					if cand.ModelClass != "M0" && pr.base.B2Applicable && !pr.base.B2.TrainingFailed {
						b2v, b2ok := metricValue(pr.base.B2BasePM(), metric, pr.base.B2PM6(), pr.base.B2PM6OK())
						if b2ok {
							vsB2 = append(vsB2, math.Abs(cv-b2v))
						}
					}
					if metric == "PM2" {
						overfit = append(overfit, job.DevPM2-job.HeldPM.PM2)
					}
				}
				if len(vsB1) == 16 {
					byGenerator[gen][calibKey("predictive_gain_vs_b1", metric, cand.ModelClass, cand.CandidateID)] = vsB1
				}
				if len(vsB2) == 16 {
					byGenerator[gen][calibKey("predictive_gain_vs_b2", metric, cand.ModelClass, cand.CandidateID)] = vsB2
				}
				if metric == "PM2" && len(overfit) == 16 {
					byGenerator[gen][calibKey("overfitting_gap", metric, cand.ModelClass, cand.CandidateID)] = overfit
				}
				// seed variation: raw HELDOUT metric values across the 16
				// populations, deviation from their own median.
				var raw []float64
				for _, pr := range runs[gen] {
					job, ok := pr.jobs[cand.CandidateID]
					if !ok || job.Failed {
						continue
					}
					cv, cok := metricValue(job.HeldPM, metric, job.PM6, job.PM6Valid)
					if cok {
						raw = append(raw, cv)
					}
				}
				if len(raw) == 16 {
					med := median(raw)
					dev := make([]float64, len(raw))
					for i, v := range raw {
						dev[i] = v - med
					}
					byGenerator[gen][calibKey("seed_variation", metric, cand.ModelClass, cand.CandidateID)] = dev
				}
			}
			// structural distance, per F2 metric.
			for _, m := range StructuralMetricIDs {
				var dist []float64
				for _, pr := range runs[gen] {
					job, ok := pr.jobs[cand.CandidateID]
					if !ok || job.Failed || !job.F2Valid || !pr.base.HeldF2Valid {
						continue
					}
					dist = append(dist, math.Abs(job.F2Generated[m]-pr.base.HeldF2[m]))
				}
				if len(dist) == 16 {
					byGenerator[gen][calibKey("structural_distance", m, cand.ModelClass, cand.CandidateID)] = dist
				}
			}
		}
	}

	for _, gen := range generators {
		for _, pr := range runs[gen] {
			for _, cand := range candidates {
				if j, ok := pr.jobs[cand.CandidateID]; ok {
					allJobs = append(allJobs, j)
				}
			}
		}
	}
	sort.Slice(allJobs, func(i, j int) bool {
		if allJobs[i].Generator != allJobs[j].Generator {
			return allJobs[i].Generator < allJobs[j].Generator
		}
		if allJobs[i].Population != allJobs[j].Population {
			return allJobs[i].Population < allJobs[j].Population
		}
		if allJobs[i].ModelClass != allJobs[j].ModelClass {
			return allJobs[i].ModelClass < allJobs[j].ModelClass
		}
		return allJobs[i].CandidateID < allJobs[j].CandidateID
	})

	failed := 0
	for _, j := range allJobs {
		if j.Failed {
			failed++
		}
	}

	thresholds := materializeThresholds(byGenerator)
	return StageAResult{Thresholds: thresholds, Jobs: allJobs, TotalJobs: len(allJobs), FailedJobs: failed}
}
