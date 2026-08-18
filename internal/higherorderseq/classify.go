package higherorderseq

// classificationInput bundles every statistic task22 Part P's outcome rules
// (section 78) read from.
type classificationInput struct {
	Candidate   Candidate
	Dependence  DependenceRow
	CrossBlock  CrossBlockRow
	LOBO        LOBORow
	Jackknife   JackknifeRow
	BlockPosTVD float64
	LinePosTVD  float64
	// Generic mirrors Config.Generic (see types.go): DistinctHand is always
	// 1 in that mode (there is no real hand dimension), so MetadataLimited
	// must not be gated on it there - doing so would trivially mislabel
	// every candidate that spans >=2 independent generic resampling groups
	// as metadata-limited.
	Generic bool
}

// positionDependentTVD is the descriptive threshold above which a
// higher-order effect is treated as substantially localized to block or
// line boundaries (task22 Part K), rather than a general second-order
// transition.
const positionDependentTVD = 0.3

// classify implements task22 Part P (section 78) in the priority order:
// data-sufficiency and fragility checks come first because they describe
// limits on what can be concluded at all; only then is the full replication
// bar checked, followed by the two narrower positive findings, with
// FIRST_ORDER_EXPLAINED as the default when nothing stronger applies. This
// ordering is a documented implementation choice - the task text defines
// each status independently but does not fix a precedence among them.
func classify(in classificationInput) ValidationRow {
	row := ValidationRow{
		Sequence: in.Candidate.Sequence, Family: in.Candidate.Family,
		ConditionalFDRQ: in.Dependence.FDRQ, EligibleBlocks: in.CrossBlock.EligibleBlocks,
		SignConsistency: in.CrossBlock.SignConsistency, SingleBlockSensitive: in.Jackknife.SingleBlockSensitive,
		DistinctJointClasses: in.CrossBlock.DistinctJoint,
	}
	if in.LOBO.TestedBlocks > 0 {
		row.LOBOAdvantageFraction = float64(in.LOBO.M2BetterBlocks) / float64(in.LOBO.TestedBlocks)
	}
	row.PositionDependent = in.BlockPosTVD > positionDependentTVD || in.LinePosTVD > positionDependentTVD
	if in.Generic {
		row.MetadataLimited = in.CrossBlock.EligibleBlocks >= 3 && in.CrossBlock.DistinctJoint <= 1
	} else {
		row.MetadataLimited = in.CrossBlock.EligibleBlocks >= 3 && (in.CrossBlock.DistinctCurrier <= 1 || in.CrossBlock.DistinctHand <= 1 || in.CrossBlock.DistinctJoint <= 1)
	}

	switch {
	case in.CrossBlock.EligibleBlocks < 3:
		row.FinalStatus = "INSUFFICIENT_SUPPORT"
	case in.Jackknife.SingleBlockSensitive:
		row.FinalStatus = "SINGLE_BLOCK_SENSITIVE"
	case in.Candidate.Family == "primary" && in.Dependence.FDRQ <= 0.05 &&
		in.CrossBlock.SignConsistency >= 0.75 && in.LOBO.TestedBlocks > 0 &&
		row.LOBOAdvantageFraction >= 2.0/3.0 && in.CrossBlock.DistinctJoint >= 2:
		row.FinalStatus = "HIGHER_ORDER_REPLICATED"
	case row.MetadataLimited && in.CrossBlock.SignConsistency >= 0.75:
		if in.Generic {
			row.FinalStatus = "GROUP_LIMITED"
		} else {
			row.FinalStatus = "METADATA_LIMITED"
		}
	case row.PositionDependent && in.CrossBlock.SignConsistency >= 0.5:
		row.FinalStatus = "POSITION_DEPENDENT"
	default:
		row.FinalStatus = "FIRST_ORDER_EXPLAINED"
	}
	return row
}
