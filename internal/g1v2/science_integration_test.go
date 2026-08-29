package g1v2

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"zcore.dev/voinich/internal/g1v2science"
)

func scienceJob(t *testing.T) JobBundle {
	t.Helper()
	root := filepath.Join("..", "..", "research", "phase3")
	cor, _ := g1v2science.NewCorpus([]string{"ab", "a", "ba", "aba", "bab", "aa", "bb", "abba", "baba", "aab", "bba", "abab", "baab", "aaa", "bbb", "abb", "baa", "ababa", "babab", "aaba"})
	cand := g1v2science.Candidate{ID: "M0-iid-1", Model: "M0", Route: "iid", Hyper: map[string]any{"alpha": "1"}, SelectionGroup: "M0_VALIDATION", Source: "V1.2.1/M0"}
	req := g1v2science.WorkRequest{Job: g1v2science.JobIdentity{ContractVersion: g1v2science.ContractVersion, ControlInstanceID: "OPEN-SCIENCE-INTEGRATION", CandidateID: cand.ID, Stage: "FIT", DependencyJobIDs: []string{}}, Candidate: cand, Corpus: cor}
	p := SciencePayload{filepath.Join(root, "task85c-c", "registries", "G1V2_CANDIDATE_REGISTRY.tsv"), filepath.Join(root, "task85c-g", "G1V2_GENERATION_SEMANTICS_V1.json"), filepath.Join(root, "task85c-c", "registries", "G1V2_STATUS_REACHABILITY_CONTRACT_V2.json"), filepath.Join(root, "task85c-j", "G1V2_V1_2_1_EVIDENCE_SCHEMA_REGISTRY.json"), req}
	b, e := json.Marshal(p)
	if e != nil {
		t.Fatal(e)
	}
	j := JobBundle{"", Experiment, ProtocolVersion, "FIT", "open-science-corpus", "M0", cand.ID, "OPEN", 0, 1, []string{testHash}, []string{}, testHash, testHash, "g1v2-v1_2_1-evidence", nil, WorkSpec{"g1v2-science-v1_2_1", string(b), 0}}
	id, e := j.ComputedID()
	if e != nil {
		t.Fatal(e)
	}
	j.JobID = id
	return j
}
func TestScienceLocalCoordinatorRetryRestartConflict(t *testing.T) {
	j := scienceJob(t)
	a, e := ExecuteEngineering(context.Background(), j)
	if e != nil {
		t.Fatal(e)
	}
	b, e := ExecuteEngineering(context.Background(), j)
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("local determinism", e)
	}
	s := Store{t.TempDir()}
	m := Manifest{SchemaVersion, CanonicalVersion, []JobBundle{j}}
	info := LocalCompatibility(testHash, 0)
	c, e := NewCoordinator(m, s, info, time.Millisecond)
	if e != nil {
		t.Fatal(e)
	}
	now := time.Now()
	lost, e := c.Claim("lost", info, now)
	if e != nil || lost == nil {
		t.Fatal(e)
	}
	retry, e := c.Claim("retry", info, now.Add(2*time.Millisecond))
	if e != nil || retry == nil || retry.Job.JobID != lost.Job.JobID {
		t.Fatal("retry identity", e)
	}
	if _, e = c.Submit("retry", retry.ID, a, telemetry("retry")); e != nil {
		t.Fatal(e)
	}
	c2, e := NewCoordinator(m, s, info, time.Millisecond)
	if e != nil {
		t.Fatal(e)
	}
	if e = RunWorkers(context.Background(), c2, 2, info); e != nil {
		t.Fatal(e)
	}
	done, total, _ := c2.Counts()
	if done != 1 || total != 1 {
		t.Fatal("restart")
	}
	idx, _ := s.ReadIndex(j.JobID)
	if len(idx.Copies) != 1 {
		t.Fatal("recomputed")
	}
	bad := a
	bad.Artifacts = append([]Artifact{}, a.Artifacts...)
	bad.Artifacts[0].Data = json.RawMessage(`{"changed":true}`)
	bad.Artifacts[0].Hash = HashBytes(bad.Artifacts[0].Data)
	x, e := s.Publish(j, bad, telemetry("conflict"))
	if e == nil || x.Status != "CONFLICT" {
		t.Fatal("conflict not quarantined")
	}
}
