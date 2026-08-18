package higherorderseq

// AuditDirFiles/DiscoveryDirFiles are the exact named files auditFingerprint
// (and, transitively, loadFrozenCandidates/structuralRelatives) reads from
// -audit-dir/-discovery-dir. Exported so a distributed coordinator
// (internal/conditionalregime) can stage precisely these by content hash,
// mirroring how replicated-local-structure-audit stages its own
// relation-dir/discovery-dir files.
var (
	AuditDirFiles = []string{
		"universal_sequence_inventory.tsv",
		"sequence_null_validation.tsv",
		"strict_replicated_sequences.tsv",
		"sequence_replication_status.tsv",
		"replicated_local_structure.yaml",
	}
	DiscoveryDirFiles = []string{"structural_classes.yaml"}
)

// LoadForDistribution reconstructs exactly the candidates/blocks/lineLength/
// relatives a local RunAndWrite run would build from the same corpus/
// metadata/audit-dir/discovery-dir inputs. Exported so
// internal/conditionalregime's distributed worker (Task44) can reconstruct
// identical state without duplicating this package's loading logic.
func LoadForDistribution(c Config) (candidates []Candidate, blocks []Block, lineLength map[string]int, relatives map[string][]string, err error) {
	_, blocks, lineLength, _, _, err = loadCorpusAndBlocks(c.CorpusPath, c.TokenMetadataMap, c.Generic)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	candidates, err = loadFrozenCandidates(c.AuditDir)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	relatives, err = structuralRelatives(c.DiscoveryDir)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	return candidates, blocks, lineLength, relatives, nil
}

// Fingerprint hashes every byte-identity-relevant input this stage reads
// plus every scientific parameter, mirroring
// replicatedlocalaudit.Fingerprint/transitionnetwork.Fingerprint. A
// distributed worker must reproduce this exact value from its own staged
// copies before it is trusted to compute a single candidate. It reuses the
// package's own pre-existing computeFingerprint (the same one checkpoint
// resume already uses) so there is exactly one formula, never two.
func Fingerprint(c Config) (string, error) {
	_, _, _, corpusSHA, metaSHA, err := loadCorpusAndBlocks(c.CorpusPath, c.TokenMetadataMap, c.Generic)
	if err != nil {
		return "", err
	}
	auditSHA, err := auditFingerprint(c.AuditDir)
	if err != nil {
		return "", err
	}
	return computeFingerprint(c, corpusSHA, metaSHA, auditSHA), nil
}
