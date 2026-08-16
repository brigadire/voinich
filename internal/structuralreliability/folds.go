package structuralreliability

import (
	"zcore.dev/voinich/internal/profilestability"
	"zcore.dev/voinich/internal/validation"
)

// foldProfiles holds the independently-built TRAIN/TEST profiles for one
// deterministic line split. Building profiles is threshold-independent;
// only eligibility (count >= some min) depends on which analysis is asking.
type foldProfiles struct {
	trainProfiles map[string]profilestability.Profile
	testProfiles  map[string]profilestability.Profile
	trainWs       map[string]profilestability.SortedProfile
	testWs        map[string]profilestability.SortedProfile
}

// BuildFolds reuses the exact same deterministic line split and per-sample
// profile construction as structural-profile-stability.
func BuildFolds(corpus validation.Corpus, folds int, seed int64) ([]foldProfiles, error) {
	foldIndexes, err := validation.SplitFolds(corpus.Lines, folds, seed)
	if err != nil {
		return nil, err
	}
	result := make([]foldProfiles, 0, len(foldIndexes))
	for _, indexes := range foldIndexes {
		train, test, err := validation.Partition(corpus, indexes)
		if err != nil {
			return nil, err
		}
		trainProfiles, testProfiles := profilestability.BuildProfiles(train), profilestability.BuildProfiles(test)
		result = append(result, foldProfiles{
			trainProfiles: trainProfiles, testProfiles: testProfiles,
			trainWs: profilestability.PrecomputeAll(trainProfiles), testWs: profilestability.PrecomputeAll(testProfiles),
		})
	}
	return result, nil
}
