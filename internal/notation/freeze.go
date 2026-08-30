package notation

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
)

func FileSHA256(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return BytesSHA256(b), nil
}

func BytesSHA256(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// MetricRegistryHash is a versioned content hash of the frozen generic
// metric registry, used to fail closed when the registry silently changes
// after a VM reference or calibration freeze (adversarial test A8).
func MetricRegistryHash() (string, error) {
	b, err := json.Marshal(MetricRegistry())
	if err != nil {
		return "", err
	}
	return BytesSHA256(b), nil
}

// VMReferenceManifest mirrors VM_REFERENCE_V2_MANIFEST.json (B02 section 39).
type VMReferenceManifest struct {
	SchemaVersion               string `json:"schema_version"`
	SourceSHA256                string `json:"source_sha256"`
	AdapterVersion              string `json:"adapter_version"`
	AnalyzerVersion             string `json:"analyzer_version"`
	MetricRegistryVersion       string `json:"metric_registry_version"`
	MetricRegistryHash          string `json:"metric_registry_hash"`
	RarefactionProtocolVersion  string `json:"rarefaction_protocol_version"`
	BootstrapProtocolVersion    string `json:"bootstrap_protocol_version"`
	CalibrationScaleVersion     string `json:"calibration_scale_version"`
	SeedSchedule                int64  `json:"seed_schedule"`
	OutputSHA256                string `json:"output_sha256"`
}

// VerifyFrozenVMReference fail-closed checks that raw (the exact bytes of a
// candidate VM_REFERENCE_V2.json file) is the frozen artifact manifest
// describes, and that the live metric registry has not silently changed
// since the freeze. A candidate cannot substitute its own VM reference file
// (adversarial test A4), and a modified registry is rejected rather than
// silently accepted (A8).
func VerifyFrozenVMReference(raw []byte, manifest VMReferenceManifest) error {
	got := BytesSHA256(raw)
	if got != manifest.OutputSHA256 {
		return fmt.Errorf("VM reference hash mismatch: file is %s, frozen manifest expects %s (a candidate-specific or modified VM reference is not authorized)", got, manifest.OutputSHA256)
	}
	live, err := MetricRegistryHash()
	if err != nil {
		return err
	}
	if live != manifest.MetricRegistryHash {
		return fmt.Errorf("metric registry changed since VM reference freeze: live=%s frozen=%s", live, manifest.MetricRegistryHash)
	}
	return nil
}

// VerifyArtifactHash fail-closed checks arbitrary frozen-artifact bytes
// (CALIBRATION_SCALES.tsv, a protocol document, ...) against a recorded
// SHA-256, used for adversarial tests A7 (corrupt provenance SHA) and A9
// (calibration scale changed after freeze).
func VerifyArtifactHash(name string, raw []byte, expectedSHA256 string) error {
	got := BytesSHA256(raw)
	if got != expectedSHA256 {
		return fmt.Errorf("%s hash mismatch: file is %s, frozen manifest expects %s", name, got, expectedSHA256)
	}
	return nil
}
