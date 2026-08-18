package replicatedlocalaudit

import (
	"sort"
)

// RelationDirFiles/DiscoveryDirFiles are the exact named files fingerprint()
// (and, transitively, loadInputs/loadFrozenDistanceDiagnostics) reads from
// -relation-dir/-discovery-dir. Exported so a distributed coordinator
// (internal/conditionalregime) can stage precisely these by content hash,
// mirroring how token-relation-validate stages its own discovery-dir files.
var (
	RelationDirFiles = []string{
		"frozen_candidate_inventory.tsv",
		"distance_profile_block_validation.tsv",
		"distance_profile_summary.tsv",
		"sequence_block_recurrence.tsv",
		"relation_classification.tsv",
		"relation_controls.tsv",
		"leave_one_block_out_transfer.tsv",
		"metadata_transfer_matrix.tsv",
		"token_relation_validation.yaml",
	}
	DiscoveryDirFiles = []string{
		"distance_context_pairs.yaml",
		"sequence_analysis.yaml",
	}
)

// DistributionState is every piece of pre-computed, permutation-invariant
// context the three null batteries (distance/shuffle/markov) read: built
// once (mirrors the pre-Task44 setup code that ran once before each
// battery's loop, unchanged), then reused by every replicate regardless of
// which backend (in-process default, subprocess, or remote worker)
// computes it.
type DistributionState struct {
	tokens         []token
	blocks         []block
	dc             []distanceCandidate
	sc             []sequenceCandidate
	profiles, refs map[string]profile
	eligible       map[string][]string
	choices        map[string]matchedVocab
	frozenPairs    map[string]bool
	markovTraining []markovHeldOut
}

// buildDistributionState is the exact pre-Task44 setup code RunAndWrite
// ran once before its distance/shuffle/markov loops (building
// profiles/refs/eligible/choices/frozenPairs/markovTraining from the
// frozen inputs), extracted unchanged so both the local default executor
// and a distributed worker's reconstruction (LoadForDistribution) call the
// same function.
func buildDistributionState(c Config) (*DistributionState, string, error) {
	tokens, blocks, dc, sc, corpusSHA, err := loadInputs(c)
	if err != nil {
		return nil, "", err
	}
	profiles := map[string]profile{}
	counts := map[string]map[string]int{}
	for _, b := range blocks {
		profiles[b.ID] = buildProfile(b)
		counts[b.ID] = map[string]int{}
		for _, t := range b.Tokens {
			counts[b.ID][t.Text]++
		}
	}
	refs := map[string]profile{}
	for _, b := range blocks {
		refs[b.ID] = mergeProfiles(profiles, b.ID)
	}
	eligible := map[string][]string{}
	for _, d := range dc {
		if d.Q > .05 {
			continue
		}
		for _, b := range blocks {
			if counts[b.ID][d.A] >= 10 && counts[b.ID][d.B] >= 10 {
				eligible[d.ID] = append(eligible[d.ID], b.ID)
			}
		}
	}
	global := map[string]int{}
	for _, t := range tokens {
		global[t.Text]++
	}
	vocab := make([]string, 0, len(global))
	for x := range global {
		vocab = append(vocab, x)
	}
	sort.Strings(vocab)
	frozenPairs := map[string]bool{}
	for _, d := range dc {
		a, b := d.A, d.B
		if b < a {
			a, b = b, a
		}
		frozenPairs[a+"\x00"+b] = true
	}
	choices := map[string]matchedVocab{}
	for _, d := range dc {
		if d.Q > .05 {
			continue
		}
		z := matchedVocab{}
		for _, a := range vocab {
			if freqMatch(global[a], global[d.A]) {
				z.a = append(z.a, a)
			}
			if freqMatch(global[a], global[d.B]) {
				z.b = append(z.b, a)
			}
		}
		choices[d.ID] = z
	}
	markovTraining := buildMarkovTraining(blocks)
	return &DistributionState{
		tokens: tokens, blocks: blocks, dc: dc, sc: sc,
		profiles: profiles, refs: refs, eligible: eligible, choices: choices,
		frozenPairs: frozenPairs, markovTraining: markovTraining,
	}, corpusSHA, nil
}

// LoadForDistribution reconstructs exactly the DistributionState a local
// RunAndWrite run would build from the same corpus/metadata/relation-dir
// inputs. Exported so internal/conditionalregime's distributed worker
// (Task44) can reconstruct identical state without duplicating this
// package's loading logic.
func LoadForDistribution(c Config) (*DistributionState, string, error) {
	state, corpusSHA, err := buildDistributionState(c)
	return state, corpusSHA, err
}

// Fingerprint hashes every byte-identity-relevant input this stage reads
// plus every scientific parameter, mirroring
// structuralprojection.Fingerprint/normalizationcompare.Fingerprint. A
// distributed worker must reproduce this exact value from its own staged
// copies before it is trusted to compute a single replicate. It reuses the
// package's own pre-existing fingerprint() (the same one checkpoint resume
// already uses) so there is exactly one formula, never two.
func Fingerprint(c Config) (string, error) {
	_, corpusSHA, err := LoadForDistribution(c)
	if err != nil {
		return "", err
	}
	return fingerprint(c, corpusSHA)
}
