package corpustransform

import (
	"fmt"
	"path/filepath"
	"strings"
)

// Label derives the human-readable experiment-id label from a corpus path
// when the caller does not supply an explicit -label: the file's base name
// with its extension stripped (task46 section 14). Scientific identity
// never depends on this label - only on the manifest's hashes and
// parameters.
func Label(corpusPath string) string {
	base := filepath.Base(corpusPath)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// TranspositionExperimentID builds the stable filename
// "<label>__transposition__w<NNN>__<order>__seed<NNN>" (task46 section 14).
func TranspositionExperimentID(label string, width int, order string, seed int64) string {
	return fmt.Sprintf("%s__transposition__w%03d__%s__seed%03d", label, width, order, seed)
}

// HomophonicExperimentID builds the stable filename
// "<label>__homophonic__h<NNN>__<selection>__seed<NNN>" (task46 section 14).
func HomophonicExperimentID(label string, homophones int, selection string, seed int64) string {
	return fmt.Sprintf("%s__homophonic__h%03d__%s__seed%03d", label, homophones, selection, seed)
}

// FrequencyHomophonicExperimentID includes the allocation model so fixed-H
// and frequency-v1 runs cannot overwrite one another in a batch directory.
func FrequencyHomophonicExperimentID(label string, homophones int, selection string, seed int64) string {
	return fmt.Sprintf("%s__homophonic-frequency-v1__hmax%03d__%s__seed%03d", label, homophones, selection, seed)
}
