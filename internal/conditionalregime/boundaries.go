package conditionalregime

import (
	"math"
	"strconv"

	"zcore.dev/voinich/internal/globalregime"
)

// ConditionalStableBoundary is one row of conditional_stable_boundaries.tsv
// (task19 sections 35-36): a multi-scale-supported change point found using
// exactly global-regime-analyze's own change-point method, but confined to
// one Currier x hand controlled physical block, so the detector can never
// cross a metadata transition.
type ConditionalStableBoundary struct {
	Class               ClassID
	BlockIndex          int
	BlockLen            int
	Position            int // block-relative
	AbsolutePosition    int
	SupportCount        int
	SupportFraction     float64
	MeanJumpStrength    float64
	MaxJumpStrength     float64
	PositionUncertainty float64
	SignatureToken      string
	SignatureDirection  string
	SignatureMagnitude  float64
}

// blockStableBoundaries finds change points within one physical block using
// the same threshold/PELT/binary-segmentation ensemble and multi-scale
// combination global-regime-analyze uses, restricted to that block's own
// tokens so it can never see across a metadata boundary (task19 section 35).
func blockStableBoundaries(tokens []string, class ClassID, blockIdx int, block Block, windowSizes []int) []ConditionalStableBoundary {
	seg := tokens[block.Start:block.End]
	var scalesUsed []int
	byScale := map[int][]globalregime.Window{}
	var allChanges []globalregime.ChangePoint
	for _, size := range windowSizes {
		if size > len(seg) {
			continue
		}
		w := globalregime.BuildWindows(seg, size, 0)
		if len(w) < 3 {
			continue
		}
		byScale[size] = w
		scalesUsed = append(scalesUsed, size)
		allChanges = append(allChanges, globalregime.ThresholdPeaks(w)...)
		allChanges = append(allChanges, globalregime.Pelt(w)...)
		allChanges = append(allChanges, globalregime.BinaryChangePoints(w)...)
	}
	if len(scalesUsed) == 0 {
		return nil
	}
	stable := globalregime.StableBoundaries(allChanges, scalesUsed)
	smallest := scalesUsed[0]
	for _, s := range scalesUsed {
		if s < smallest {
			smallest = s
		}
	}
	out := make([]ConditionalStableBoundary, 0, len(stable))
	for _, b := range stable {
		tok, dir, mag := boundarySignature(byScale[smallest], b.Position)
		out = append(out, ConditionalStableBoundary{
			Class: class, BlockIndex: blockIdx, BlockLen: block.Len(), Position: b.Position, AbsolutePosition: block.Start + b.Position,
			SupportCount: b.SupportCount, SupportFraction: b.SupportFraction, MeanJumpStrength: b.MeanJumpStrength,
			MaxJumpStrength: b.MaxJumpStrength, PositionUncertainty: b.PositionUncertainty,
			SignatureToken: tok, SignatureDirection: dir, SignatureMagnitude: mag,
		})
	}
	return out
}

// boundarySignature characterizes a boundary by which single token's
// frequency changed most across it, rather than by its absolute position,
// so recurring TYPES of transitions can be compared across physically
// incomparable blocks (task19 section 37).
func boundarySignature(windows []globalregime.Window, center int) (token, direction string, magnitude float64) {
	var before, after globalregime.Profile
	for i, w := range windows {
		if w.Center <= center {
			before = w.Distribution()
		} else if after == nil {
			after = windows[i].Distribution()
		}
	}
	if before == nil || after == nil {
		return "", "", 0
	}
	bestTok, bestDelta := "", 0.0
	seen := map[string]bool{}
	for tok, av := range after {
		d := av - before[tok]
		seen[tok] = true
		if math.Abs(d) > math.Abs(bestDelta) {
			bestTok, bestDelta = tok, d
		}
	}
	for tok, bv := range before {
		if seen[tok] {
			continue
		}
		d := -bv
		if math.Abs(d) > math.Abs(bestDelta) {
			bestTok, bestDelta = tok, d
		}
	}
	dir := "increase"
	if bestDelta < 0 {
		dir = "decrease"
	}
	return bestTok, dir, math.Abs(bestDelta)
}

// conditionalBoundaries collects blockStableBoundaries over every physical
// block of every eligible joint class.
func conditionalBoundaries(tokens []string, classes []ClassID, blocksByClass map[ClassID][]Block, windowSizes []int) []ConditionalStableBoundary {
	var out []ConditionalStableBoundary
	for _, class := range classes {
		for i, b := range blocksByClass[class] {
			out = append(out, blockStableBoundaries(tokens, class, i, b, windowSizes)...)
		}
	}
	return out
}

// RecurringBoundaryType groups conditional stable boundaries by their
// (signature token, direction) type rather than by absolute position, per
// task19 section 37.
type RecurringBoundaryType struct {
	SignatureToken     string
	SignatureDirection string
	Occurrences        int
	DistinctClasses    int
	DistinctBlocks     int
}

func recurringBoundaryTypes(boundaries []ConditionalStableBoundary) []RecurringBoundaryType {
	type key struct{ tok, dir string }
	groups := map[key][]ConditionalStableBoundary{}
	for _, b := range boundaries {
		if b.SignatureToken == "" {
			continue
		}
		k := key{b.SignatureToken, b.SignatureDirection}
		groups[k] = append(groups[k], b)
	}
	out := make([]RecurringBoundaryType, 0, len(groups))
	for k, bs := range groups {
		classes := map[ClassID]bool{}
		blocks := map[string]bool{}
		for _, b := range bs {
			classes[b.Class] = true
			blocks[b.Class.Label()+"#"+strconv.Itoa(b.BlockIndex)] = true
		}
		out = append(out, RecurringBoundaryType{k.tok, k.dir, len(bs), len(classes), len(blocks)})
	}
	return out
}
