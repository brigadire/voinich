package task82a

import (
	"strings"
	"testing"

	"zcore.dev/voinich/internal/mnemonicspace"
)

// TestManifestDeterministic checks that BuildManifest is a pure function
// of the frozen registry/design constants: two independent builds must
// produce byte-identical job lists (task82a.txt sec.31, "freeze checksum;
// prohibit additions/removals").
func TestManifestDeterministic(t *testing.T) {
	a := BuildManifest()
	b := BuildManifest()
	if len(a.Jobs) != len(b.Jobs) {
		t.Fatalf("job count differs: %d vs %d", len(a.Jobs), len(b.Jobs))
	}
	for i := range a.Jobs {
		if a.Jobs[i] != b.Jobs[i] {
			t.Fatalf("job %d differs: %+v vs %+v", i, a.Jobs[i], b.Jobs[i])
		}
	}
	seen := map[string]bool{}
	for _, j := range a.Jobs {
		if seen[j.JobID] {
			t.Fatalf("duplicate job id %s", j.JobID)
		}
		seen[j.JobID] = true
	}
}

// TestNoVoynichPathGuard is the F2 feature-extraction-only guard/test
// required by task82a.txt sec.47.
func TestNoVoynichPathGuard(t *testing.T) {
	bad := []string{
		"data_work/ZL3b-x7.canonical.txt",
		"data/ZL3b-n.txt",
		"data_work/IT2a-x7.canonical.txt",
		"data/IT2a-n.txt",
		"anything/voynich/foo.txt",
		"data_work/whatever.txt",
	}
	for _, p := range bad {
		if err := assertNoVoynichPath(p); err == nil {
			t.Errorf("expected VOYNICH_FIREWALL rejection for %q", p)
		}
	}
	good := []string{"research/phase2/task82a/raw/f2corpus/abc.txt", "data_test/pg2097-2.txt"}
	for _, p := range good {
		if err := assertNoVoynichPath(p); err != nil {
			t.Errorf("unexpected rejection for %q: %v", p, err)
		}
	}
}

// TestLocalSemanticsInvariant is task82a.txt sec.34: a local application
// inside a corpus-scale run must be equivalent to a standalone Task82
// application of the same frozen mechanism. This checks f01_speculum_core
// chunk 0 against the exact-recovery behavior Task82's own frozen results
// already established for R0 (MECHANISM_SUMMARY.tsv r0_exact_rate=1.0).
func TestLocalSemanticsInvariant(t *testing.T) {
	reg := mnemonicspace.FrozenRegistry()
	var spec mnemonicspace.MechanismSpec
	for _, s := range reg {
		if s.ID == "f01_speculum_core" {
			spec = s
		}
	}
	if spec.ID == "" {
		t.Fatal("f01_speculum_core not found in frozen registry")
	}
	param, ok := spec.ParameterSet("f01-core-bounded")
	if !ok {
		t.Fatal("f01-core-bounded parameter set not found")
	}
	scale, ok := scaleFor(spec.ID)
	if !ok {
		t.Fatal("no capacity table entry for f01_speculum_core")
	}
	corpus := SourceCorpus{ID: "unit-test", Letters: make([]string, scale.Capacity)}
	for i := range corpus.Letters {
		corpus.Letters[i] = string(Latin23[i%len(Latin23)])
	}
	corpus.Items = []string{"Ia", "Ib", "Ic", "Id"}
	in, unitItems := buildChunkInput(scale, PolicyLiteralReset, corpus, 0)
	prepared, err := (mnemonicspace.Runner{}).Prepare(spec, param, in, chunkSeed(1, 0))
	if err != nil {
		t.Fatalf("Prepare: %v", err)
	}
	results := chunkRecovery(spec, param, in, prepared, unitItems, scale.Capacity)
	r0 := results["R0_FULL_KNOWLEDGE"]
	target := strings.Join(unitItems, "")
	if r0.Class != mnemonicspace.ResultExact || string(r0.Value) != target {
		t.Fatalf("expected EXACT R0 recovery matching %q, got class=%s value=%q", target, r0.Class, r0.Value)
	}
}
