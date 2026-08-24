package fingerprintv2

import (
	"math/rand"
	"strings"
)

// PF4LeafPair is one real, physically-paired recto/verso leaf used by the
// Task79c leaf-paired null (research/phase2/fingerprint/TASK79C_DESIGN.md
// section 8). It is never itself permuted; only the pairing between
// RectoLeaf and VersoLeaf sides is permuted by pf4LeafPairedNull.
type PF4LeafPair struct {
	Leaf        string
	RectoFolio  string
	VersoFolio  string
	RectoVector []float64
	VersoVector []float64
}

// PF4LeafNullResult is the machine-readable output of the Task79c
// leaf-paired null for PF4_RECTO_VERSO_COHERENCE.
type PF4LeafNullResult struct {
	MetricID         string    `json:"metric_id"`
	NullModel        string    `json:"null_model"`
	Observed         float64   `json:"observed"`
	NullMean         float64   `json:"null_mean"`
	NullSD           float64   `json:"null_sd"`
	PValue           float64   `json:"p_value"`
	EffectSizeSD     float64   `json:"effect_size_sd"`
	EffectDefined    bool      `json:"effect_defined"`
	Permutations     int       `json:"permutations"`
	Seed             int64     `json:"seed"`
	PairedLeafCount  int       `json:"paired_leaf_count"`
	UnpairedFolios   []string  `json:"unpaired_folios"`
	NullDraws        []float64 `json:"null_draws"`
	Verdict          string    `json:"verdict"`
	VerdictRationale string    `json:"verdict_rationale"`
}

// pf4LeafPairs extracts every real, physically-paired recto/verso leaf from
// the corpus's folio-mean line-profile vectors (the same vectors
// rectoVersoCoherence already computes), plus every folio side that has no
// opposite side transcribed (excluded from pairing, per
// TASK79C_DESIGN.md section 8's "missing/unpaired leaves").
func pf4LeafPairs(ls []LineProfile) (pairs []PF4LeafPair, unpaired []string) {
	order, v, _ := folioMeanVectors(ls)
	sides := map[string]map[string]string{}
	for _, f := range order {
		side := ""
		if strings.HasSuffix(f, "r") {
			side = "r"
		}
		if strings.HasSuffix(f, "v") {
			side = "v"
		}
		if side == "" {
			continue
		}
		leaf := leafID(f)
		if sides[leaf] == nil {
			sides[leaf] = map[string]string{}
		}
		sides[leaf][side] = f
	}
	for _, leaf := range orderedKeys(sides) {
		s := sides[leaf]
		r, hasR := s["r"]
		vv, hasV := s["v"]
		switch {
		case hasR && hasV:
			pairs = append(pairs, PF4LeafPair{Leaf: leaf, RectoFolio: r, VersoFolio: vv, RectoVector: v[r], VersoVector: v[vv]})
		case hasR:
			unpaired = append(unpaired, r)
		case hasV:
			unpaired = append(unpaired, vv)
		}
	}
	return pairs, unpaired
}

// pf4CoherenceForPairs recomputes PF4's existing statistic (mean of
// 1/(1+distance(recto,verso))) over a fixed set of pairs. It is the same
// arithmetic rectoVersoCoherence already uses; the null below only changes
// which vector is paired with which, never the statistic itself.
func pf4CoherenceForPairs(pairs []PF4LeafPair) float64 {
	vals := make([]float64, 0, len(pairs))
	for _, p := range pairs {
		vals = append(vals, 1/(1+distance(p.RectoVector, p.VersoVector)))
	}
	return mean(vals)
}

// pf4LeafPairedNull runs the Task79c-designed leaf-paired null
// (TASK79C_DESIGN.md section 8 / PF4_LEAF_NULL.md): it holds every real
// recto-side and verso-side mean vector fixed and draws a uniformly random
// bijection between the two sides, destroying only which specific verso is
// the physical flip side of which specific recto.
func pf4LeafPairedNull(ls []LineProfile, permutations int, seed int64) PF4LeafNullResult {
	pairs, unpaired := pf4LeafPairs(ls)
	observed := pf4CoherenceForPairs(pairs)
	null := make([]float64, permutations)
	rng := rand.New(rand.NewSource(seed))
	rectoVecs := make([][]float64, len(pairs))
	versoVecs := make([][]float64, len(pairs))
	for i, p := range pairs {
		rectoVecs[i] = p.RectoVector
		versoVecs[i] = p.VersoVector
	}
	for i := 0; i < permutations; i++ {
		perm := rng.Perm(len(versoVecs))
		vals := make([]float64, len(pairs))
		for j := range pairs {
			vals[j] = 1 / (1 + distance(rectoVecs[j], versoVecs[perm[j]]))
		}
		null[i] = mean(vals)
	}
	t := nullTest("PF4_RECTO_VERSO_COHERENCE", "leaf-paired bijection permutation (task79c)", observed, null)
	verdict, rationale := pf4LeafVerdict(t, len(pairs))
	return PF4LeafNullResult{
		MetricID: "PF4_RECTO_VERSO_COHERENCE", NullModel: t.NullModel,
		Observed: t.Observed, NullMean: t.NullMean, NullSD: t.NullSD, PValue: t.PValue,
		EffectSizeSD: t.EffectSize, EffectDefined: t.EffectDefined, Permutations: permutations, Seed: seed,
		PairedLeafCount: len(pairs), UnpairedFolios: unpaired, NullDraws: null,
		Verdict: verdict, VerdictRationale: rationale,
	}
}

// pf4LeafVerdict applies the parent task's fixed vocabulary (SUPPORTED /
// NOT_SUPPORTED / INCONCLUSIVE, section 24 of the parent task) using a
// one-sided alpha=0.05 decision rule fixed in TASK79C_DESIGN.md, and treats
// too few usable leaf pairs as inconclusive rather than a false negative.
func pf4LeafVerdict(t NullTest, pairedLeaves int) (string, string) {
	const minPairs = 5
	if pairedLeaves < minPairs {
		return "INCONCLUSIVE", "fewer than 5 usable recto/verso leaf pairs; the leaf-paired null is underpowered"
	}
	if !t.EffectDefined {
		return "INCONCLUSIVE", "null distribution has zero variance; effect size is undefined"
	}
	if t.PValue < 0.05 && t.EffectSize > 0 {
		return "SUPPORTED", "same-leaf recto/verso coherence exceeds the leaf-paired permutation null at alpha=0.05"
	}
	return "NOT_SUPPORTED", "same-leaf recto/verso coherence does not exceed the leaf-paired permutation null at alpha=0.05; this is absence of evidence, not evidence of absence, since no equivalence margin was preregistered"
}
