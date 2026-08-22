package inversehomophony

import (
	"encoding/json"
	"fmt"
	"os"
)

// SyntheticCorpusSpec names one synthetic ciphertext corpus for the
// validation harness: its ciphertext, its evaluator-only oracle mapping,
// and (found via the corpus-transform manifest, also evaluator-only) the
// original plaintext it was derived from.
type SyntheticCorpusSpec struct {
	Label       string // stable experiment id, e.g. "doyle_h004_uniform"
	CipherPath  string
	MappingPath string
	Genre       string // "doyle", "longfellow", "astafiev" - for cross-genre gate check (task57 section 20.3)
}

// DevelopmentSpecs is the fixed DEVELOPMENT split (task57 section 4/9):
// used only to fit Threshold via the pair-discrimination diagnostic. Never
// used to compute a final validation-gate number.
func DevelopmentSpecs() []SyntheticCorpusSpec {
	return []SyntheticCorpusSpec{
		spec("doyle_h004_uniform", "doyle__homophonic__h004__uniform__seed001", "doyle"),
		spec("doyle_h004_weighted", "doyle__homophonic__h004__weighted__seed001", "doyle"),
	}
}

// ValidationSpecs is the fixed VALIDATION split: unseen H values, the
// unseen frequency-v1 allocation model, and two unseen plaintext genres
// (Longfellow, Astafiev) - task57 section 4's "unseen H/model combinations
// and/or other plaintext corpora", preferring genre transfer.
func ValidationSpecs() []SyntheticCorpusSpec {
	return []SyntheticCorpusSpec{
		spec("doyle_h006_uniform", "doyle__homophonic__h006__uniform__seed001", "doyle"),
		spec("doyle_h008_uniform", "doyle__homophonic__h008__uniform__seed001", "doyle"),
		spec("doyle_h006_weighted", "doyle__homophonic__h006__weighted__seed001", "doyle"),
		spec("doyle_h008_weighted", "doyle__homophonic__h008__weighted__seed001", "doyle"),
		specNamed("doyle_freq_v1_hmax004_uniform", "doyle__homophonic-frequency-v1__hmax004__uniform__seed001", "doyle"),
		specNamed("doyle_freq_v1_hmax006_uniform", "doyle__homophonic-frequency-v1__hmax006__uniform__seed001", "doyle"),
		specNamed("doyle_freq_v1_hmax008_uniform", "doyle__homophonic-frequency-v1__hmax008__uniform__seed001", "doyle"),
		spec("longfellow_h004_uniform", "longfellow__homophonic__h004__uniform__seed001", "longfellow"),
		spec("astafiev_h004_uniform", "astafiev__homophonic__h004__uniform__seed001", "astafiev"),
	}
}

func spec(label, base, genre string) SyntheticCorpusSpec {
	return specNamed(label, base, genre)
}

func specNamed(label, base, genre string) SyntheticCorpusSpec {
	dir := "data_test/transformed/"
	return SyntheticCorpusSpec{
		Label:       label,
		CipherPath:  dir + base + ".txt",
		MappingPath: dir + base + ".txt.mapping.tsv",
		Genre:       genre,
	}
}

type transformManifestStub struct {
	InputPath string `json:"input_path"`
}

// PlaintextPathFromManifest reads "<cipherPath>.transform.json" and returns
// its recorded input_path - the original plaintext corpus this ciphertext
// was derived from. Evaluator-only.
func PlaintextPathFromManifest(cipherPath string) (string, error) {
	b, err := os.ReadFile(cipherPath + ".transform.json")
	if err != nil {
		return "", err
	}
	var m transformManifestStub
	if err := json.Unmarshal(b, &m); err != nil {
		return "", err
	}
	if m.InputPath == "" {
		return "", fmt.Errorf("inversehomophony: %s.transform.json has no input_path", cipherPath)
	}
	return m.InputPath, nil
}
