package fingerprintv2

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestCS7MapInsertionOrderDoesNotAffectStatisticOrPRNG(t *testing.T) {
	edit := []float64{1, 2, 3, 1, 2, 2, 3, 4, 1, 3, 2, 4}
	structure := []float64{9, 4, 7, 2, 8, 1, 5, 3, 6, 4, 8, 2}
	first := map[int][]int{}
	first[2] = []int{0, 1, 2, 3, 4, 5}
	first[1] = []int{6, 7, 8, 9, 10, 11}
	second := map[int][]int{}
	second[1] = []int{6, 7, 8, 9, 10, 11}
	second[2] = []int{0, 1, 2, 3, 4, 5}

	aSum, aN := cs7PartialSpearman(edit, first, func(i int) float64 { return structure[i] })
	bSum, bN := cs7PartialSpearman(edit, second, func(i int) float64 { return structure[i] })
	if aSum != bSum || aN != bN {
		t.Fatalf("map insertion order changed CS7 observed statistic: (%g,%d) != (%g,%d)", aSum, aN, bSum, bN)
	}
	aNull := cs7PermutationNull(edit, structure, first, 50, rand.New(rand.NewSource(20260824)))
	bNull := cs7PermutationNull(edit, structure, second, 50, rand.New(rand.NewSource(20260824)))
	if !reflect.DeepEqual(aNull, bNull) {
		t.Fatal("map insertion order changed CS7 PRNG consumption or null distribution")
	}
}

func deterministicProcessPayload() ([]byte, error) {
	c := task79Fixture(true)
	cfg := testConfig("unused")
	cfg.Task79 = &Task79Config{Enabled: true, Permutations: 50, BootstrapReplicates: 50, MinGroupSize: 5, ChangePointPenalty: 2}
	base := CorpusResult{Corpus: c.info, Grammar: &GrammarSummary{Validation: "SUPPORTED"}, EditGraph: &EditFamilyValidation{}, CrossScale: &CrossScaleResult{}}
	task79 := runTask79(c, cfg, base, 20260824)
	familyOf := map[string]int{"start": 0, "mid": 0, "end": 1}
	zoneOf := map[string]string{"start": "PREFIX", "mid": "INTERNAL", "end": "SUFFIX"}
	cs2Family, cs2Zone, n := cs2Test(c, familyOf, zoneOf, 50, rand.New(rand.NewSource(30260824)))
	return json.Marshal(struct {
		Task79    Task79Result `json:"task79"`
		CS2Family NullTest     `json:"cs2_family"`
		CS2Zone   NullTest     `json:"cs2_zone"`
		N         int          `json:"n"`
	}{task79, cs2Family, cs2Zone, n})
}

func TestFingerprintDeterminismAcrossProcessRestartAndGOMAXPROCS(t *testing.T) {
	if os.Getenv("FINGERPRINT_V2_DETERMINISM_CHILD") == "1" {
		payload, err := deterministicProcessPayload()
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(os.Getenv("FINGERPRINT_V2_DETERMINISM_OUTPUT"), payload, 0600); err != nil {
			t.Fatal(err)
		}
		return
	}

	var reference []byte
	for _, procs := range []string{"1", "2", "4"} {
		path := t.TempDir() + "/payload.json"
		cmd := exec.Command(os.Args[0], "-test.run=^TestFingerprintDeterminismAcrossProcessRestartAndGOMAXPROCS$")
		cmd.Env = append(os.Environ(),
			"FINGERPRINT_V2_DETERMINISM_CHILD=1",
			"FINGERPRINT_V2_DETERMINISM_OUTPUT="+path,
			"GOMAXPROCS="+procs,
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("GOMAXPROCS=%s child failed: %v\n%s", procs, err, out)
		}
		got, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if reference == nil {
			reference = got
			continue
		}
		if !bytes.Equal(reference, got) {
			t.Fatalf("separate process with GOMAXPROCS=%s changed normative payload", procs)
		}
	}
}
