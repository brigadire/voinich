package notation

import (
	"testing"
)

func regFP(metrics ...Metric) Fingerprint {
	return Fingerprint{Metadata: map[string]string{"metric_registry_version": MetricRegistryVersion}, Metrics: metrics}
}

func TestA1MissingCalibrationScaleFailsClosed(t *testing.T) {
	c := regFP(Metric{MetricID: "G01", Family: "G", Value: 2, Status: Comparable})
	r := regFP(Metric{MetricID: "G01", Family: "G", Value: 1, Status: Comparable})
	rows, _, err := Compare(c, r, nil) // no scale supplied at all
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Status != NotComparable {
		t.Fatal("missing scale must never fall back to scale=1")
	}
	if rows[0].Distance != 0 {
		t.Fatalf("missing scale row must not carry a fabricated distance: %+v", rows[0])
	}
}

func TestA2WrongMetricVersionFails(t *testing.T) {
	c := Fingerprint{Metadata: map[string]string{"metric_registry_version": "generic-metrics-1.0"}, Metrics: []Metric{{MetricID: "G01", Family: "G", Value: 2, Status: Comparable}}}
	r := Fingerprint{Metadata: map[string]string{"metric_registry_version": "generic-metrics-2.0"}, Metrics: []Metric{{MetricID: "G01", Family: "G", Value: 1, Status: Comparable}}}
	if _, _, err := Compare(c, r, []Scale{{MetricID: "G01", Center: 0, Spread: 1}}); err == nil {
		t.Fatal("expected hard failure on metric registry version mismatch")
	}
}

func TestA3WrongSupportRegimeNeverApproximateJoins(t *testing.T) {
	c := regFP(Metric{MetricID: "S01_TRANSITION_DENSITY", Family: "S", Regime: "TOP_100", Value: 0.5, Status: Comparable})
	r := regFP(Metric{MetricID: "S01_TRANSITION_DENSITY", Family: "S", Regime: "TOP_250", Value: 0.5, Status: Comparable})
	rows, _, err := Compare(c, r, []Scale{{MetricID: "S01_TRANSITION_DENSITY", Regime: "TOP_100", Center: 0, Spread: 1}, {MetricID: "S01_TRANSITION_DENSITY", Regime: "TOP_250", Center: 0, Spread: 1}})
	if err != nil {
		t.Fatal(err)
	}
	if rows[0].Status != NotComparable {
		t.Fatal("mismatched support regime must not be approximately joined")
	}
}

func TestA4CandidateSpecificVMReferenceRejected(t *testing.T) {
	raw := []byte(`{"schema_version":"notation-fingerprint-1.0"}`)
	registryHash, err := MetricRegistryHash()
	if err != nil {
		t.Fatal(err)
	}
	manifest := VMReferenceManifest{OutputSHA256: BytesSHA256([]byte("the real frozen file")), MetricRegistryHash: registryHash}
	if err := VerifyFrozenVMReference(raw, manifest); err == nil {
		t.Fatal("a substituted VM reference file must be rejected")
	}
}

func TestA5MissingPhysicalLinesForcesNotComparable(t *testing.T) {
	rs := []Record{{
		SchemaVersion: SchemaVersion, CorpusID: "C-NOLINES", Representation: "R1",
		Document: ObservedLevel{Value: "d", Observed: true},
		TokenID:  "t1", TokenIndex: 0, Token: "ab", Symbols: []string{"a", "b"},
	}, {
		SchemaVersion: SchemaVersion, CorpusID: "C-NOLINES", Representation: "R1",
		Document: ObservedLevel{Value: "d", Observed: true},
		TokenID:  "t2", TokenIndex: 1, Token: "cd", Symbols: []string{"c", "d"},
	}}
	fp, err := Analyze(rs)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, m := range fp.Metrics {
		if m.Family == "L" {
			found = true
			if m.Status != NotComparable {
				t.Fatalf("L metric %s should be NOT_COMPARABLE without lines, got %s", m.MetricID, m.Status)
			}
		}
	}
	if !found {
		t.Fatal("expected L family rows even when NOT_COMPARABLE")
	}
}

func TestA6ShortCorpusCheckpointNotComparable(t *testing.T) {
	rs := syntheticLinedCorpus(5, 4) // 20 tokens total, far below any checkpoint
	fp, err := Analyze(rs)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range fp.Curves {
		if c.Checkpoint >= 5000 && c.Status != NotComparable {
			t.Fatalf("curve %+v should be NOT_COMPARABLE below its checkpoint", c)
		}
	}
	if _, err := Rarefy(rs, 5000, 1); err == nil {
		t.Fatal("rarefaction above corpus size must fail rather than silently truncate")
	}
}

func TestA7CorruptProvenanceShaFails(t *testing.T) {
	raw := []byte("frozen artifact bytes")
	if err := VerifyArtifactHash("TEST_ARTIFACT", raw, "0000000000000000000000000000000000000000000000000000000000000000"); err == nil {
		t.Fatal("a wrong recorded SHA must fail verification")
	}
	if err := VerifyArtifactHash("TEST_ARTIFACT", raw, BytesSHA256(raw)); err != nil {
		t.Fatalf("correct hash must verify: %v", err)
	}
}

func TestA8ChangedMetricRegistryAfterFreezeFails(t *testing.T) {
	raw := []byte("vm reference bytes")
	manifest := VMReferenceManifest{OutputSHA256: BytesSHA256(raw), MetricRegistryHash: "not-the-live-registry-hash"}
	if err := VerifyFrozenVMReference(raw, manifest); err == nil {
		t.Fatal("expected failure when the live metric registry hash no longer matches the frozen manifest")
	}
}

func TestA9ChangedCalibrationScaleAfterFreezeFails(t *testing.T) {
	original := []byte("metric_id\tscale\nG01\t1.0\n")
	tampered := []byte("metric_id\tscale\nG01\t999.0\n")
	frozenHash := BytesSHA256(original)
	if err := VerifyArtifactHash("CALIBRATION_SCALES.tsv", tampered, frozenHash); err == nil {
		t.Fatal("a calibration scale table changed after freeze must fail hash verification")
	}
}
