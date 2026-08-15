package transitionnetwork

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
)

func RunAndWrite(c Config) error {
	if err := validateConfig(c); err != nil {
		return err
	}
	if c.CheckpointPath == "" {
		c.CheckpointPath = filepath.Join(c.OutputDir, "checkpoint.json")
	}
	if c.ProgressWriter == nil && !c.Quiet {
		c.ProgressWriter = os.Stderr
	}
	p := newProgress(c.ProgressWriter)
	p.begin(1, "load corpus and freeze transition matrix")
	tokens, blocks, corpusSHA, metaSHA, err := loadCorpusAndBlocks(c.CorpusPath, c.MetadataPath)
	if err != nil {
		return err
	}
	counts, vocab, edges, data := buildData(tokens, blocks, c.MinTokenCount)
	a := &analysis{Tokens: tokens, Blocks: blocks, Counts: counts, Vocab: vocab, Edges: edges, Data: data}
	p.update(1, 1, "loaded")
	p.begin(2, "compute block-normalized edge effects")
	summarizeEdges(a, c.MinBlockTokenCount)
	p.update(1, 1, "edge effects")
	p.begin(3, "compute outgoing/incoming profiles and entropy")
	observedProfiles(a, c.MinBlockTokenCount)
	p.update(1, 1, "profiles")
	fp := fingerprint(c, corpusSHA, metaSHA)
	cp, resumed, err := loadCheckpoint(c.CheckpointPath, fp)
	if err != nil {
		return err
	}
	if resumed && c.ProgressWriter != nil {
		fmt.Fprintf(c.ProgressWriter, "Resuming from checkpoint %s: %d/%d primary replicates, %d refinement replicates complete\n", c.CheckpointPath, cp.Completed, c.Permutations, cp.RefineCompleted)
	}
	p.begin(4, "within-block destination permutation null")
	summaryByKey := map[string]*EdgeSummary{}
	for _, r := range a.Summaries {
		summaryByKey[r.String()] = r
	}
	stab := map[string]*ProfileStability{}
	for i := range a.Stability {
		stab[a.Stability[i].Direction+"\x00"+a.Stability[i].Token] = &a.Stability[i]
	}
	ws := newPermWorkspace(a, c.MinBlockTokenCount)
	for rep := cp.Completed; rep < c.Permutations; rep++ {
		es, outs, ins := ws.run(c.Seed, rep, true)
		for e, v := range es {
			r := summaryByKey[e.String()]
			if r == nil {
				continue
			}
			if (r.ExpectedSign == "preferred" && v >= r.MedianLog2) || (r.ExpectedSign == "depleted" && v <= r.MedianLog2) {
				cp.EdgeExceed[e.String()]++
			}
		}
		for t, v := range outs {
			if r := stab["outgoing\x00"+t]; r != nil {
				if v.Correlation >= r.LOBOMedianCorrelation {
					cp.OutExceed[t]++
				}
				if v.SignAgreement >= r.SignAgreement {
					cp.OutSignExceed[t]++
				}
				if math.Abs(v.EntropyEffect) >= math.Abs(r.EntropyEffect) {
					cp.OutEntropyExceed[t]++
				}
			}
		}
		for t, v := range ins {
			if r := stab["incoming\x00"+t]; r != nil {
				if v.Correlation >= r.LOBOMedianCorrelation {
					cp.InExceed[t]++
				}
				if v.SignAgreement >= r.SignAgreement {
					cp.InSignExceed[t]++
				}
				if math.Abs(v.EntropyEffect) >= math.Abs(r.EntropyEffect) {
					cp.InEntropyExceed[t]++
				}
			}
		}
		cp.Completed = rep + 1
		if err = saveCheckpoint(c.CheckpointPath, cp); err != nil {
			return fmt.Errorf("save checkpoint: %w", err)
		}
		p.update(rep+1, c.Permutations, "primary null")
	}
	for _, r := range a.Summaries {
		r.EmpiricalP = float64(cp.EdgeExceed[r.String()]+1) / float64(c.Permutations+1)
		r.Permutations = c.Permutations
	}
	for i := range a.Stability {
		r := &a.Stability[i]
		ex, sx, hx := cp.OutExceed[r.Token], cp.OutSignExceed[r.Token], cp.OutEntropyExceed[r.Token]
		if r.Direction == "incoming" {
			ex = cp.InExceed[r.Token]
			sx, hx = cp.InSignExceed[r.Token], cp.InEntropyExceed[r.Token]
		}
		r.PermutationP = float64(ex+1) / float64(c.Permutations+1)
		r.SignPermutationP = float64(sx+1) / float64(c.Permutations+1)
		r.EntropyPermutationP = float64(hx+1) / float64(c.Permutations+1)
	}
	// Refinement eligibility is frozen from the primary run before any refined p is examined.
	if len(cp.RefineCandidates) == 0 {
		for _, r := range a.Summaries {
			if r.EmpiricalP < .01 && r.EligibleBlocks >= 3 && r.JointClasses >= 2 {
				cp.RefineCandidates = append(cp.RefineCandidates, r.String())
			}
		}
		sort.Strings(cp.RefineCandidates)
		cp.Phase = "refine"
		if err = saveCheckpoint(c.CheckpointPath, cp); err != nil {
			return err
		}
	}
	if len(cp.RefineCandidates) > 0 && c.RefinePermutations > c.Permutations {
		set := map[string]bool{}
		for _, k := range cp.RefineCandidates {
			set[k] = true
		}
		extra := c.RefinePermutations - c.Permutations
		for n := cp.RefineCompleted; n < extra; n++ {
			rep := c.Permutations + n
			es, _, _ := ws.run(c.Seed, rep, false)
			for e, v := range es {
				r := summaryByKey[e.String()]
				if r != nil && set[e.String()] && ((r.ExpectedSign == "preferred" && v >= r.MedianLog2) || (r.ExpectedSign == "depleted" && v <= r.MedianLog2)) {
					cp.RefineExceed[e.String()]++
				}
			}
			cp.RefineCompleted = n + 1
			if err = saveCheckpoint(c.CheckpointPath, cp); err != nil {
				return err
			}
			p.update(c.Permutations+n+1, c.RefinePermutations, "refined null")
		}
		for _, k := range cp.RefineCandidates {
			r := summaryByKey[k]
			r.EmpiricalP = float64(cp.EdgeExceed[k]+cp.RefineExceed[k]+1) / float64(c.RefinePermutations+1)
			r.Permutations = c.RefinePermutations
		}
	}
	bh(a.Summaries, true)
	bh(a.Summaries, false)
	for _, r := range a.Summaries {
		classify(r)
	}
	for i := range a.Stability {
		r := &a.Stability[i]
		r.Replicated = r.EligibleBlocks >= 3 && r.JointClasses >= 2 && r.LOBOMedianCorrelation > 0 && r.PermutationP <= .05 && r.SignPermutationP <= .05
		r.EntropyStatus = "INSUFFICIENT_SUPPORT"
		if r.EligibleBlocks >= 3 && r.JointClasses >= 2 && r.EntropySignConsistency >= .75 && r.EntropyPermutationP <= .05 {
			if r.EntropyEffect < 0 {
				r.EntropyStatus = "REPRODUCIBLY_RESTRICTED"
			} else {
				r.EntropyStatus = "REPRODUCIBLY_BROAD"
			}
		}
	}
	p.begin(5, "graph, metadata, and block transfer")
	computeGraphDiagnostics(a, c.MinBlockTokenCount)
	p.update(1, 1, "graph transfer")
	p.begin(6, "held-out prediction and model order")
	computePredictions(a, c.MinBlockTokenCount)
	p.update(1, 1, "prediction")
	p.begin(7, "write TSV, GraphML, YAML, Markdown, and SVG")
	if err = writeAll(c, a, corpusSHA, metaSHA); err != nil {
		return err
	}
	p.update(1, 1, "outputs")
	p.begin(8, "finalize")
	if c.CheckpointPath != "-" {
		if err = os.Remove(c.CheckpointPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	p.update(1, 1, "complete")
	return nil
}
