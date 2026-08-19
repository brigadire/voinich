package beginendanalyze

import (
	"crypto/sha256"
	"fmt"
	"os"
)

// Fingerprint hashes the dictionary and corpus file contents plus every
// scientific parameter. A distributed worker must reproduce this exact
// value from its own staged copies before it is trusted to compute a
// single candidate batch.
func Fingerprint(c Config) (string, error) {
	dictionaryBytes, err := os.ReadFile(c.DictionaryPath)
	if err != nil {
		return "", fmt.Errorf("fingerprint: read dictionary: %w", err)
	}
	corpusBytes, err := os.ReadFile(c.CorpusPath)
	if err != nil {
		return "", fmt.Errorf("fingerprint: read corpus: %w", err)
	}
	p := c.parameters()
	h := sha256.New()
	fmt.Fprintf(h, "begin-end-analyze\x00%x\x00%x\x00%d\x00%d\x00%d\x00%d\x00%s\x00%t\x00%d",
		sha256.Sum256(dictionaryBytes), sha256.Sum256(corpusBytes),
		p.MaxWindow, p.Permutations, p.MinTokenCount, p.RandomSeed, p.PermutationMode, p.IncludeUnclear, p.MaxCandidates)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
