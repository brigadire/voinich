package g1v2

import (
	"fmt"
	"math/big"
)

type PM6Fixture struct {
	Alphabet    []string `json:"alphabet"`
	Length      int      `json:"length"`
	Observed    []string `json:"observed"`
	Draws       int      `json:"draws"`
	MinCoverage int      `json:"min_coverage"`
}

type PM6Result struct {
	SpaceSize             string `json:"space_size"`
	ObservedTypes         int    `json:"observed_types"`
	ComplementSize        string `json:"complement_size"`
	DuplicateDrawsAllowed bool   `json:"duplicate_draws_allowed"`
	Status                string `json:"status"`
}

// EvaluatePM6Fixture validates only PM6-v2 negative-set constructibility. It
// deliberately computes no model score and derives no scientific threshold.
func EvaluatePM6Fixture(f PM6Fixture) (PM6Result, error) {
	if len(f.Alphabet) == 0 || f.Length < 1 || f.Draws < 0 || f.MinCoverage < 0 {
		return PM6Result{}, fmt.Errorf("invalid PM6 fixture")
	}
	seenAlpha := map[string]bool{}
	for _, a := range f.Alphabet {
		if a == "" || seenAlpha[a] {
			return PM6Result{}, fmt.Errorf("alphabet symbols must be nonempty and unique")
		}
		seenAlpha[a] = true
	}
	space := new(big.Int).Exp(big.NewInt(int64(len(f.Alphabet))), big.NewInt(int64(f.Length)), nil)
	seen := map[string]bool{}
	for _, s := range f.Observed {
		seen[s] = true
	}
	obs := big.NewInt(int64(len(seen)))
	complement := new(big.Int).Sub(space, obs)
	if complement.Sign() < 0 {
		return PM6Result{}, fmt.Errorf("observed types exceed space")
	}
	status := "NEGATIVE_TEST_AVAILABLE"
	if complement.Sign() == 0 {
		status = "NEGATIVE_TEST_NOT_IDENTIFIABLE"
	} else if f.Draws < f.MinCoverage {
		status = "INSUFFICIENT_COVERAGE"
	}
	return PM6Result{space.String(), len(seen), complement.String(), true, status}, nil
}
