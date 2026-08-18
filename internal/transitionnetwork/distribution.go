package transitionnetwork

// LoadForDistribution reconstructs exactly the *PermWorkspace a local
// RunAndWrite run would build from the same corpus/metadata inputs -
// loadCorpusAndBlocks + buildData + newPermWorkspace, unchanged. It is
// exported so internal/conditionalregime's distributed worker (Task44) can
// reconstruct identical state from staged copies of the same input files
// without duplicating this package's loading logic.
func LoadForDistribution(c Config) (*PermWorkspace, string, string, error) {
	tokens, blocks, corpusSHA, metaSHA, err := loadCorpusAndBlocks(c.CorpusPath, c.MetadataPath, c.Generic)
	if err != nil {
		return nil, "", "", err
	}
	counts, vocab, edges, data := buildData(tokens, blocks, c.MinTokenCount)
	a := &analysis{Tokens: tokens, Blocks: blocks, Counts: counts, Vocab: vocab, Edges: edges, Data: data}
	return newPermWorkspace(a, c.MinBlockTokenCount), corpusSHA, metaSHA, nil
}

// Fingerprint hashes the corpus (and metadata map, unless Generic) plus
// every scientific parameter, via the exact same fingerprint() formula
// checkpoint resume already uses. A distributed worker must reproduce this
// exact value from its own staged copies before it is trusted to compute a
// single replicate.
func Fingerprint(c Config) (string, error) {
	_, corpusSHA, metaSHA, err := LoadForDistribution(c)
	if err != nil {
		return "", err
	}
	return fingerprint(c, corpusSHA, metaSHA), nil
}
