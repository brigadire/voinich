package main

import "fmt"

var authoritativeHashes = map[string]string{
	"data_work/ZL3b-x7.canonical.txt": "f46f4190af65b85d145ec5bb957c1f56029b567e4bef12ac7baa1797f358d692",
	"data_work/IT2a-x7.canonical.txt": "10286ee7e11ad974e9d0f884e3b0df1b588745a4b77ad428a638a5ff63946a8b",
	"data/ZL3b-n.txt":                 "bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc",
	"data/IT2a-n.txt":                 "7f27a8b0feed8f6de0a99900df6bf912dd1d295c38e5f830bac8b41c3f536fb5",
	"research/phase3/task85/GRAMMAR_CORPUS_SPLIT.tsv":                 "639969bd91daaf362df49afa4fc12c3f5289cea4200cbb71d5f743d5c5bff551",
	"research/phase3/task85/GRAMMAR_CORPUS_SPLIT_MANIFEST.json":       "80b49086623f53968a0c925c3a6780a82d35b9c118747a9ab6f90fe7ce03719a",
	"research/phase3/task85/GRAMMAR_CORPUS_SPLIT_FROZEN":              "f816d3ad9f1c702fe80a9b4314d06aad4b0ad150d3ba6f6b93c3a897a0e46145",
	"research/phase3/task85/GRAMMAR_EXPERIMENT_CONTRACT_FROZEN":       "a7895c4e4c91bcacd215a71d26cfdc13bdac7013e374f28efd7dad832ac8d2c6",
	"research/phase3/task85a/G1_EXECUTABLE_CONTRACT.json":             "816d39b81ade7ea50cb3253ad16ccac7243b8ba05c642a21859bc10389050491",
	"research/phase3/task85a/GRAMMAR_EXPERIMENT_CONTRACT_V1.1_FROZEN": "98acd9394fa9f24317d39387b2d2a08b6d58287fda9342ee69f3c824925b2b8c",
}

func verifyAuthoritativeHashes() error {
	for path, want := range authoritativeHashes {
		got, err := sha256File(path)
		if err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		if got != want {
			return fmt.Errorf("%s: hash mismatch: want %s got %s", path, want, got)
		}
	}
	return nil
}
