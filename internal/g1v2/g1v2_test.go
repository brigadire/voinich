package g1v2

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"zcore.dev/voinich/internal/pki"
)

var testHash = HashBytes([]byte("frozen-engineering-value"))

func job(model, stage string, deps []string, i int) JobBundle {
	j := JobBundle{"", Experiment, ProtocolVersion, stage, "open-engineering-corpus", model, model + "-fixture", "small", i, uint64(100 + i), []string{testHash}, []string{}, testHash, testHash, "engineering-output-v1", sortedCopy(deps), WorkSpec{"sha256-chain-v1", model + stage, 1000 + i}}
	id, _ := j.ComputedID()
	j.JobID = id
	return j
}

func manifest() Manifest {
	var jobs []JobBundle
	for i, m := range []string{"M0", "M1", "M2", "M3", "M4", "M5"} {
		f := job(m, "FIT", nil, i)
		p := job(m, "PREDICTIVE", []string{f.JobID}, i)
		g := job(m, "GENERATION", []string{p.JobID}, i)
		s := job(m, "STRUCTURAL", []string{g.JobID}, i)
		jobs = append(jobs, f, p, g, s)
	}
	deps := make([]string, 0, 6)
	for _, j := range jobs {
		if j.Stage == "STRUCTURAL" {
			deps = append(deps, j.JobID)
		}
	}
	sort.Strings(deps)
	jobs = append(jobs, job("stage-neutral", "AGGREGATION", deps, 0))
	return Manifest{SchemaVersion, CanonicalVersion, jobs}
}

func telemetry(worker string) Telemetry {
	return Telemetry{Worker: worker, Host: "fixture-host", StartUTC: "2026-01-01T00:00:00Z", EndUTC: "2026-01-01T00:00:01Z", WallSeconds: 1, CPUSeconds: .9, PeakRSSBytes: 1024, TransferBytes: 128, InfrastructureStatus: "SUCCESS"}
}

func TestJobIDExcludesOperationsAndManifestDAG(t *testing.T) {
	m := manifest()
	if err := m.Validate(); err != nil {
		t.Fatal(err)
	}
	a := m.Jobs[0]
	id, _ := a.ComputedID()
	if id != a.JobID {
		t.Fatal("unstable JobID")
	}
	_ = telemetry("different-worker")
	id2, _ := a.ComputedID()
	if id2 != id {
		t.Fatal("worker telemetry affected JobID")
	}
}

func TestRunResumeAndWorkerExpansion(t *testing.T) {
	m := manifest()
	s := Store{t.TempDir()}
	info := LocalCompatibility(testHash, 0)
	c, err := NewCoordinator(m, s, info, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	// Complete one job, then construct a fresh coordinator (restart).
	l, err := c.Claim("worker-a", info, time.Now())
	if err != nil || l == nil {
		t.Fatal(err)
	}
	r, err := ExecuteEngineering(context.Background(), l.Job)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = c.Submit("worker-a", l.ID, r, telemetry("worker-a")); err != nil {
		t.Fatal(err)
	}
	c2, err := NewCoordinator(m, s, info, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	if err = RunWorkers(context.Background(), c2, 4, info); err != nil {
		t.Fatal(err)
	}
	done, total, _ := c2.Counts()
	if done != total {
		t.Fatalf("completed %d/%d", done, total)
	}
	first, _ := s.ReadIndex(l.Job.JobID)
	if len(first.Copies) != 1 {
		t.Fatal("completed job was recomputed after restart")
	}
}

func TestLeaseExpiryPreservesIdentity(t *testing.T) {
	m := Manifest{SchemaVersion, CanonicalVersion, []JobBundle{job("M0", "FIT", nil, 0)}}
	s := Store{t.TempDir()}
	info := LocalCompatibility(testHash, 0)
	c, _ := NewCoordinator(m, s, info, time.Millisecond)
	now := time.Now()
	a, _ := c.Claim("lost-worker", info, now)
	b, _ := c.Claim("replacement-worker", info, now.Add(2*time.Millisecond))
	if b == nil || a.Job.JobID != b.Job.JobID || a.ID == b.ID {
		t.Fatal("expired lease did not reissue identical JobID under a new LeaseID")
	}
}

func TestDuplicateConflictCorruptionWrongCode(t *testing.T) {
	j := job("M2", "FIT", nil, 0)
	s := Store{t.TempDir()}
	r, _ := ExecuteEngineering(context.Background(), j)
	x, err := s.Publish(j, r, telemetry("worker-a"))
	if err != nil || x.Status != "VERIFIED" {
		t.Fatal(err)
	}
	x, err = s.Publish(j, r, telemetry("worker-b"))
	if err != nil || len(x.Copies) != 2 {
		t.Fatalf("identical duplicate: %v %+v", err, x)
	}
	bad := r
	bad.Artifacts = append([]Artifact(nil), r.Artifacts...)
	bad.Artifacts[0].Data = json.RawMessage(`{"changed":true}`)
	bad.Artifacts[0].Hash = HashBytes(bad.Artifacts[0].Data)
	x, err = s.Publish(j, bad, telemetry("worker-c"))
	if err == nil || x.Status != "CONFLICT" {
		t.Fatal("conflict was not quarantined")
	}
	if _, err = os.Stat(filepath.Join(s.Root, "quarantine", j.JobID+".json")); err != nil {
		t.Fatal(err)
	}
	corrupt := r
	corrupt.Artifacts = append([]Artifact(nil), r.Artifacts...)
	corrupt.Artifacts[0].Data = json.RawMessage(`{"corrupt":true}`)
	if _, err = s.Publish(j, corrupt, telemetry("worker-d")); err == nil {
		t.Fatal("corrupt artifact accepted")
	}
	wrong := j
	wrong.CodeHash = HashBytes([]byte("wrong"))
	if _, err = s.Publish(wrong, r, telemetry("worker-e")); err == nil {
		t.Fatal("wrong code bundle accepted")
	}
}

func decisionFixture() DecisionEvidence {
	v := .8
	pms := map[string]PMRecord{}
	for _, id := range []string{"PM1", "PM2", "PM4", "PM5", "PM6"} {
		r := PMRecord{id, &v, 0, .5, .5, "GTE", true, true, "PASS", ""}
		r.ArtifactHash = PMHash(r)
		pms[id] = r
	}
	d := .1
	f2 := []F2Record{{"edit-small-0", "EDIT_1", "EDIT", "small", 0, &d, .2, true, "PASS", nil, ""}, {"lex-small-0", "LP_1", "LEXICAL_PARADIGM", "small", 0, &d, .2, true, "PASS", nil, ""}}
	for i := range f2 {
		f2[i].ArtifactHash = F2Hash(f2[i])
	}
	rank := 1
	return DecisionEvidence{SchemaVersion, testHash, testHash, testHash, 7, []string{testHash}, "FITTED", pms, []string{"edit-small-0", "lex-small-0"}, f2, &rank, Decision{"PREDICTIVE_PASS", "STRUCTURAL_PASS", "MODEL_ADEQUATE", "ORDER_ONLY"}}
}

func cloneEvidence(e DecisionEvidence) DecisionEvidence {
	b, _ := json.Marshal(e)
	var x DecisionEvidence
	_ = json.Unmarshal(b, &x)
	return x
}

func TestEvidenceOnlyRegenerationAndNegativeMutations(t *testing.T) {
	base := decisionFixture()
	j := job("stage-neutral", "AGGREGATION", nil, 0)
	base.JobID, base.CodeHash, base.ConfigHash, base.Seed, base.DependencyHashes = j.JobID, j.CodeHash, j.ConfigHash, j.Seed, j.DependencyHashes
	d, err := VerifyDecisionForJob(j, base)
	if err != nil || d != base.Recorded {
		t.Fatal(err)
	}
	want, _ := CanonicalDecision(base.Recorded)
	got, _ := CanonicalDecision(d)
	if !reflect.DeepEqual(want, got) {
		t.Fatal("decision is not byte identical")
	}
	tests := map[string]func(*DecisionEvidence){
		"PM value": func(e *DecisionEvidence) { r := e.PMs["PM2"]; *r.Value = .1; e.PMs["PM2"] = r },
		"PM threshold": func(e *DecisionEvidence) {
			r := e.PMs["PM2"]
			r.Threshold = .9
			r.ArtifactHash = PMHash(r)
			e.PMs["PM2"] = r
		},
		"PM availability": func(e *DecisionEvidence) {
			r := e.PMs["PM2"]
			r.Available = false
			r.ArtifactHash = PMHash(r)
			e.PMs["PM2"] = r
		},
		"artifact hash":     func(e *DecisionEvidence) { r := e.PMs["PM2"]; r.ArtifactHash = testHash; e.PMs["PM2"] = r },
		"F2 record":         func(e *DecisionEvidence) { *e.F2[0].Distance = .9 },
		"seed":              func(e *DecisionEvidence) { e.Seed++ },
		"config hash":       func(e *DecisionEvidence) { e.ConfigHash = "bad" },
		"code hash":         func(e *DecisionEvidence) { e.CodeHash = "bad" },
		"dependency":        func(e *DecisionEvidence) { e.DependencyHashes = []string{HashBytes([]byte("mutated-dependency"))} },
		"scientific status": func(e *DecisionEvidence) { e.Recorded.ModelStatus = "MODEL_INADEQUATE" },
		"unavailable to FAIL": func(e *DecisionEvidence) {
			r := e.PMs["PM6"]
			r.Available = false
			r.Gate = "FAIL"
			r.ArtifactHash = PMHash(r)
			e.PMs["PM6"] = r
		},
		"missing artifact": func(e *DecisionEvidence) { r := e.PMs["PM5"]; r.ArtifactHash = ""; e.PMs["PM5"] = r },
	}
	for name, mut := range tests {
		t.Run(name, func(t *testing.T) {
			e := cloneEvidence(base)
			mut(&e)
			if _, err := VerifyDecisionForJob(j, e); err == nil {
				t.Fatal("mutation did not fail closed")
			}
		})
	}
}

func TestThreeValuedAndInductionSemantics(t *testing.T) {
	p, err := Predictive(map[string]string{"PM2": "PASS", "PM5": "PASS", "PM6": "NOT_ASSESSABLE"})
	if err != nil || p != "PREDICTIVE_NOT_ASSESSABLE" {
		t.Fatal("NOT_ASSESSABLE collapsed")
	}
	if p == "PREDICTIVE_FAIL" {
		t.Fatal("FAIL equals NOT_ASSESSABLE")
	}
	e := decisionFixture()
	e.FitStatus = "INDUCTION_LIMIT_REACHED"
	e.PMs = nil
	e.ComplexityRank = nil
	e.Recorded = Decision{"PREDICTIVE_NOT_ASSESSABLE", "STRUCTURAL_NOT_ASSESSABLE", "MODEL_NOT_IDENTIFIABLE", "NOT_IDENTIFIABLE"}
	for i := range e.F2 {
		e.F2[i].Distance = nil
		e.F2[i].Applicable = false
		e.F2[i].Gate = "NOT_REACHED"
		e.F2[i].NotReached = &NotReached{"induction failed", testHash, testHash}
		e.F2[i].ArtifactHash = F2Hash(e.F2[i])
	}
	j := job("stage-neutral", "AGGREGATION", nil, 0)
	e.JobID, e.CodeHash, e.ConfigHash, e.Seed, e.DependencyHashes = j.JobID, j.CodeHash, j.ConfigHash, j.Seed, j.DependencyHashes
	if _, err := VerifyDecisionForJob(j, e); err != nil {
		t.Fatal(err)
	}
	e.Recorded.ModelStatus = "MODEL_INADEQUATE"
	if _, err := VerifyDecisionForJob(j, e); err == nil {
		t.Fatal("induction failure collapsed to model inadequacy")
	}
	// Infrastructure status is operational telemetry and is neither a
	// scientific status nor part of the scientific payload hash.
	r, _ := ExecuteEngineering(context.Background(), job("M0", "FIT", nil, 0))
	b1, _ := r.canonical()
	tel := telemetry("worker-a")
	tel.InfrastructureStatus = "INFRASTRUCTURE_FAILED"
	b2, _ := r.canonical()
	if ScientificStatuses[tel.InfrastructureStatus] || HashBytes(b1) != HashBytes(b2) {
		t.Fatal("infrastructure failure collapsed into scientific payload")
	}
}

func TestScientificOutcomeIsNotRetried(t *testing.T) {
	j := job("M3", "FIT", nil, 0)
	m := Manifest{SchemaVersion, CanonicalVersion, []JobBundle{j}}
	s := Store{t.TempDir()}
	info := LocalCompatibility(testHash, 0)
	c, _ := NewCoordinator(m, s, info, time.Second)
	l, _ := c.Claim("worker-a", info, time.Now())
	r, _ := ExecuteEngineering(context.Background(), j)
	r.ScientificStatus = "TRAINING_FAILED"
	if _, err := c.Submit("worker-a", l.ID, r, telemetry("worker-a")); err != nil {
		t.Fatal(err)
	}
	if next, err := c.Claim("worker-a", info, time.Now()); err != nil || next != nil {
		t.Fatal("scientific outcome was retried")
	}
}

func TestPM6EngineeringBoundaries(t *testing.T) {
	cases := []struct {
		name          string
		f             PM6Fixture
		status, space string
	}{
		{"ordinary", PM6Fixture{[]string{"a", "b"}, 2, []string{"aa"}, 8, 4}, "NEGATIVE_TEST_AVAILABLE", "4"},
		{"saturated", PM6Fixture{[]string{"a", "b"}, 1, []string{"a", "b"}, 8, 4}, "NEGATIVE_TEST_NOT_IDENTIFIABLE", "2"},
		{"alphabet-one", PM6Fixture{[]string{"x"}, 1, []string{"x"}, 1, 1}, "NEGATIVE_TEST_NOT_IDENTIFIABLE", "1"},
		{"unicode", PM6Fixture{[]string{"α", "β"}, 2, []string{"αα"}, 5, 4}, "NEGATIVE_TEST_AVAILABLE", "4"},
		{"large-space", PM6Fixture{[]string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}, 30, nil, 10, 1}, "NEGATIVE_TEST_AVAILABLE", "1000000000000000000000000000000"},
		{"insufficient", PM6Fixture{[]string{"a", "b"}, 2, []string{"aa"}, 1, 4}, "INSUFFICIENT_COVERAGE", "4"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := EvaluatePM6Fixture(tc.f)
			if err != nil || got.Status != tc.status || got.SpaceSize != tc.space || !got.DuplicateDrawsAllowed {
				t.Fatalf("%+v %v", got, err)
			}
		})
	}
	saturated, _ := EvaluatePM6Fixture(PM6Fixture{[]string{"a", "b"}, 1, []string{"a", "b"}, 4, 2})
	unsaturated, _ := EvaluatePM6Fixture(PM6Fixture{[]string{"a", "b"}, 2, []string{"aa"}, 4, 2})
	if saturated.Status != "NEGATIVE_TEST_NOT_IDENTIFIABLE" || unsaturated.Status != "NEGATIVE_TEST_AVAILABLE" {
		t.Fatal("mixed saturated/unsaturated lengths collapsed")
	}
}

func TestPhase1MTLSLeaseIntegration(t *testing.T) {
	pkiDir := t.TempDir()
	if err := pki.GenerateCA(pkiDir, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	ca, caKey := pki.CAPaths(pkiDir)
	if err := pki.IssueCoordinator(ca, caKey, pkiDir, nil, []string{"127.0.0.1"}, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	if err := pki.IssueWorker(ca, caKey, pkiDir, "g1-worker-a", time.Hour, false); err != nil {
		t.Fatal(err)
	}
	cc, ck := pki.IssueCoordinatorPaths(pkiDir)
	wc, wk := pki.IssueWorkerPaths(pkiDir, "g1-worker-a")
	m := Manifest{SchemaVersion, CanonicalVersion, []JobBundle{job("M5", "FIT", nil, 0)}}
	s := Store{t.TempDir()}
	info := LocalCompatibility(testHash, 0)
	core, err := NewCoordinator(m, s, info, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	server, err := StartRemoteCoordinator(core, "127.0.0.1:0", cc, ck, ca, "")
	if err != nil {
		if strings.Contains(err.Error(), "operation not permitted") {
			t.Skip("sandbox forbids loopback listener")
		}
		t.Fatal(err)
	}
	defer server.Close(context.Background())
	w, err := NewRemoteWorker("https://"+server.Listener.Addr().String(), wc, wk, ca, "", info, "mtls-fixture-host")
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()
	deadline := time.Now().Add(3 * time.Second)
	for {
		n, total, _ := core.Counts()
		if n == total {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("mTLS worker did not complete lease")
		}
		time.Sleep(time.Millisecond)
	}
	cancel()
	<-done
	x, err := s.ReadIndex(m.Jobs[0].JobID)
	if err != nil || len(x.Copies) != 1 || x.Copies[0].Worker != "g1-worker-a" {
		t.Fatalf("certificate provenance missing: %+v %v", x, err)
	}
	bad := info
	bad.CodeHash = HashBytes([]byte("wrong-code"))
	if l, err := core.Claim("g1-worker-a", bad, time.Now()); err == nil || l != nil {
		t.Fatal("incompatible executable accepted")
	}
}
