package conditionalregime

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/metadatavalidation"
)

// ResidualMetadataAssociation is one row of residual_metadata_association.tsv
// (task19 sections 30-32): the association a residual clustering still has
// with the metadata it was residualized against, and how much smaller that
// association is than the original (unconditioned) blind regimes'.
type ResidualMetadataAssociation struct {
	WindowSize           int
	K                    int
	Method               string
	Representation       string
	Metadata             string // "currier" or "hand"
	ResidualNMI          float64
	ResidualARI          float64
	OriginalNMI          float64
	InformationReduction float64 // 1 - residual/original, clamped to [0,1]
	OriginalSource       string
}

// readOriginalGlobalNMI reads the frozen, already-corrected global max NMI
// for Currier and hand from task18's cluster_metadata_global_summary.tsv
// (primary scope, method_scope=global), so Part B's "how much did
// residualization remove" comparison is against the same documented
// baseline task19's own context cites, not a value recomputed ad hoc.
func readOriginalGlobalNMI(path string) (currier, hand float64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, 0, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 64*1024), 4*1024*1024)
	if !s.Scan() {
		return 0, 0, fmt.Errorf("empty summary: %s", path)
	}
	header := strings.Split(s.Text(), "\t")
	col := map[string]int{}
	for i, h := range header {
		col[h] = i
	}
	found := map[string]bool{}
	for s.Scan() {
		row := strings.Split(s.Text(), "\t")
		if row[col["method_scope"]] != "global" || row[col["metric"]] != "NMI" || row[col["scope"]] != "primary" {
			continue
		}
		v, e := strconv.ParseFloat(row[col["observed"]], 64)
		if e != nil {
			continue
		}
		switch row[col["metadata"]] {
		case "currier":
			currier, found["currier"] = v, true
		case "hand":
			hand, found["hand"] = v, true
		}
	}
	if e := s.Err(); e != nil {
		return 0, 0, e
	}
	if !found["currier"] || !found["hand"] {
		return 0, 0, fmt.Errorf("summary %s does not contain both primary global NMI rows", path)
	}
	return currier, hand, nil
}

func informationReduction(original, residual float64) float64 {
	if original <= 0 {
		return 0
	}
	r := 1 - residual/original
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}

// residualMetadataAssociations computes NMI/ARI between one residual
// clustering and Currier/hand, and the reduction relative to the original,
// unconditioned frozen regimes.
func residualMetadataAssociations(windowSize, k int, method, representation string, rw []ResidualWindow, fullLabels []int, originalCurrierNMI, originalHandNMI float64) []ResidualMetadataAssociation {
	clusters := make([]string, len(fullLabels))
	for i, l := range fullLabels {
		clusters[i] = strconv.Itoa(l)
	}
	currier := make([]string, len(rw))
	hand := make([]string, len(rw))
	for i, w := range rw {
		currier[i] = w.Class.Currier
		hand[i] = w.Class.Hand
	}
	mc := metadatavalidation.AssociationMetrics(currier, clusters)
	mh := metadatavalidation.AssociationMetrics(hand, clusters)
	return []ResidualMetadataAssociation{
		{windowSize, k, method, representation, "currier", mc.NMI, mc.ARI, originalCurrierNMI, informationReduction(originalCurrierNMI, mc.NMI), "cluster_metadata_global_summary.tsv (global, primary, NMI)"},
		{windowSize, k, method, representation, "hand", mh.NMI, mh.ARI, originalHandNMI, informationReduction(originalHandNMI, mh.NMI), "cluster_metadata_global_summary.tsv (global, primary, NMI)"},
	}
}

// ResidualCandidate is one row of residual_regime_candidates.tsv (task19
// sections 33-34): cross-metadata recurrence for one residual cluster, plus
// a documented, purely descriptive composite ranking.
type ResidualCandidate struct {
	WindowSize          int
	Method              string
	K                   int
	Representation      string
	Cluster             int
	Size                int
	CurrierClasses      int
	Hands               int
	JointClasses        int
	PhysicalBlocks      int
	TotalPhysicalBlocks int
	MetadataEntropy     float64
	CompositeScore      float64
}

func entropyOfPairs(pairs []ClassID) float64 {
	counts := map[ClassID]int{}
	for _, p := range pairs {
		counts[p]++
	}
	n := len(pairs)
	if n == 0 {
		return 0
	}
	h := 0.0
	for _, c := range counts {
		p := float64(c) / float64(n)
		h -= p * math.Log(p)
	}
	return h
}

// residualCandidates ranks every cluster in one residual clustering by
// cross-metadata and cross-block recurrence (task19 section 34: "high
// cross-block recurrence + high cross-metadata recurrence"). The composite
// score is an unweighted sum of coverage fractions kept alongside its
// components; it is documented as descriptive ranking only, never as an
// inferential statistic.
func residualCandidates(windowSize, k int, method, representation string, rw []ResidualWindow, fullLabels []int, totalCurriers, totalHands, totalJointClasses, totalPhysicalBlocks int) []ResidualCandidate {
	byCluster := map[int][]int{}
	for i, l := range fullLabels {
		byCluster[l] = append(byCluster[l], i)
	}
	clusters := make([]int, 0, len(byCluster))
	for c := range byCluster {
		clusters = append(clusters, c)
	}
	sort.Ints(clusters)
	out := make([]ResidualCandidate, 0, len(clusters))
	for _, c := range clusters {
		idxs := byCluster[c]
		curriers, hands, joints, blocks := map[string]bool{}, map[string]bool{}, map[ClassID]bool{}, map[string]bool{}
		pairs := make([]ClassID, 0, len(idxs))
		for _, i := range idxs {
			w := rw[i]
			curriers[w.Class.Currier] = true
			hands[w.Class.Hand] = true
			joints[ClassID{Currier: w.Class.Currier, Hand: w.Class.Hand}] = true
			blocks[w.Class.Label()+"#"+strconv.Itoa(w.BlockIndex)] = true
			pairs = append(pairs, ClassID{Currier: w.Class.Currier, Hand: w.Class.Hand})
		}
		composite := 0.0
		if totalCurriers > 0 {
			composite += float64(len(curriers)) / float64(totalCurriers)
		}
		if totalHands > 0 {
			composite += float64(len(hands)) / float64(totalHands)
		}
		if totalJointClasses > 0 {
			composite += float64(len(joints)) / float64(totalJointClasses)
		}
		if totalPhysicalBlocks > 0 {
			composite += float64(len(blocks)) / float64(totalPhysicalBlocks)
		}
		out = append(out, ResidualCandidate{
			WindowSize: windowSize, Method: method, K: k, Representation: representation, Cluster: c, Size: len(idxs),
			CurrierClasses: len(curriers), Hands: len(hands), JointClasses: len(joints), PhysicalBlocks: len(blocks),
			TotalPhysicalBlocks: totalPhysicalBlocks, MetadataEntropy: entropyOfPairs(pairs), CompositeScore: composite,
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].CompositeScore > out[j].CompositeScore })
	return out
}
