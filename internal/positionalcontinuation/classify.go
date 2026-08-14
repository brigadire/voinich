package positionalcontinuation

// classificationInput bundles every statistic task23 Part Q's outcome rules
// (section 92) read from.
type classificationInput struct {
	PositionDependenceP      float64 // primary (line_position) I(X;position) empirical p
	StratifiedPredecessorP   float64 // primary (line_position) stratified predecessor empirical p
	M3BetterThanM2Fraction   float64
	CrossBlockSignConsistency float64
	EligibleBlocks           int
	SingleBlockSensitive     bool
	UniqueCheySurroundingContexts int
	CheyOccurrences          int
}

// minEligibleBlocks mirrors higher-order-sequence-validate's own
// cross-block-replication floor (task22 sections 42-45): fewer than 3
// eligible physical blocks is treated as insufficient to claim replication.
const minEligibleBlocks = 3

const significanceLevel = 0.05

// classify implements task23 Part Q (section 92) in a fixed priority order:
// data-sufficiency and fragility checks first (they bound what can be
// concluded at all), then the concrete BOUNDARY_FORMULA alternative, then the
// three positive/negative structural findings from strongest to weakest, with
// NO_POSITIONAL_STRUCTURE as the default when nothing else applies. This
// ordering is a documented implementation choice - task23 defines each status
// independently but does not fix a precedence among them.
func classify(in classificationInput) ValidationRow {
	row := ValidationRow{
		PositionDependenceP: in.PositionDependenceP, PositionDependenceSig: in.PositionDependenceP <= significanceLevel,
		StratifiedPredecessorSig: in.StratifiedPredecessorP <= significanceLevel,
		M3BetterThanM2Fraction:   in.M3BetterThanM2Fraction,
		CrossBlockSignConsistency: in.CrossBlockSignConsistency,
		SingleBlockSensitive:     in.SingleBlockSensitive,
		EligibleBlocks:           in.EligibleBlocks,
		CheyEnrichmentSig:        in.StratifiedPredecessorP <= significanceLevel,
		BoundaryFormulaSupported: in.CheyOccurrences >= 2 && in.UniqueCheySurroundingContexts <= 1,
	}

	switch {
	case in.EligibleBlocks < minEligibleBlocks:
		row.FinalStatus = "INSUFFICIENT_SUPPORT"
	case in.SingleBlockSensitive:
		row.FinalStatus = "SINGLE_BLOCK_SENSITIVE"
	case row.BoundaryFormulaSupported && row.PositionDependenceSig:
		row.FinalStatus = "BOUNDARY_FORMULA"
	case row.StratifiedPredecessorSig && in.M3BetterThanM2Fraction >= 0.5 && in.CrossBlockSignConsistency >= 0.75:
		row.FinalStatus = "GENERAL_HIGHER_ORDER"
	case row.PositionDependenceSig && row.StratifiedPredecessorSig && in.CrossBlockSignConsistency >= 0.5:
		row.FinalStatus = "POSITION_CONDITIONED_HIGHER_ORDER"
	case row.PositionDependenceSig && !row.StratifiedPredecessorSig:
		row.FinalStatus = "AIIN_POSITION_EFFECT"
	default:
		row.FinalStatus = "NO_POSITIONAL_STRUCTURE"
	}
	return row
}
