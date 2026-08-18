package positionalcontinuation

// LoadForDistribution reconstructs exactly the SAiinOccurrences/
// AiinOccurrences a local RunAndWrite run would build from the same
// corpus/metadata/higher-order-dir inputs (including, in generic mode,
// resolving the frozen target triple the same way RunAndWrite does).
// Exported so internal/conditionalregime's distributed worker (Task44) can
// reconstruct identical state without duplicating this package's loading
// logic.
func LoadForDistribution(c Config) (sAiinOccs []SAiinOccurrence, aiinOccs []AiinOccurrence, err error) {
	if c.Generic {
		s, aiin, chey, e := resolveGenericTarget(c.HigherOrderDir)
		if e != nil {
			return nil, nil, e
		}
		FrozenS, FrozenAiin, FrozenChey = s, aiin, chey
		FrozenSAiin = FrozenS + " " + FrozenAiin
	}
	_, blocks, lineLength, _, _, err := loadCorpusAndBlocks(c.CorpusPath, c.TokenMetadataMap, c.Generic)
	if err != nil {
		return nil, nil, err
	}
	totalTokens := 0
	for _, b := range blocks {
		totalTokens += len(b.Tokens)
	}
	sAiinOccs = findSAiinOccurrences(blocks, lineLength, totalTokens)
	aiinOccs = findAiinOccurrences(blocks, lineLength, totalTokens)
	return sAiinOccs, aiinOccs, nil
}

// Fingerprint hashes every byte-identity-relevant input this stage reads
// plus every scientific parameter, mirroring
// higherorderseq.Fingerprint/transitionnetwork.Fingerprint. A distributed
// worker must reproduce this exact value from its own staged copies before
// it is trusted to compute a single battery. It reuses the package's own
// pre-existing computeFingerprint (the same one checkpoint resume already
// uses) so there is exactly one formula, never two.
func Fingerprint(c Config) (string, error) {
	_, _, _, corpusSHA, metaSHA, err := loadCorpusAndBlocks(c.CorpusPath, c.TokenMetadataMap, c.Generic)
	if err != nil {
		return "", err
	}
	higherOrderSHA, err := higherOrderFingerprint(c.HigherOrderDir)
	if err != nil {
		return "", err
	}
	if c.Generic {
		s, aiin, chey, e := resolveGenericTarget(c.HigherOrderDir)
		if e != nil {
			return "", e
		}
		FrozenS, FrozenAiin, FrozenChey = s, aiin, chey
		FrozenSAiin = FrozenS + " " + FrozenAiin
	}
	return computeFingerprint(c, corpusSHA, metaSHA, higherOrderSHA), nil
}
