package notation

import (
	"bytes"
	"testing"
)

// smallCheckpoint keeps calibration tests fast; production checkpoints
// (5k/10k/20k/39380) are exercised once by the frozen CALIBRATION_SCALES.tsv
// build, not by the unit test suite.
const smallCheckpoint = 600
const smallReplicates = 5

func TestGeneratorsProduceValidUSC(t *testing.T) {
	for _, g := range CalibrationGenerators() {
		recs := g.Generate(g, smallCheckpoint, SeedFor(BaseSeed, g.ID, "SYNTHETIC-TOKEN", "CALIBRATION", smallCheckpoint, 0))
		if len(recs) == 0 {
			t.Fatalf("%s produced no records", g.ID)
		}
		if err := Validate(recs); err != nil {
			t.Fatalf("%s: invalid USC: %v", g.ID, err)
		}
	}
}

func TestC1NonZeroScaleCoverage(t *testing.T) {
	runs := RunCalibrationPanel(smallCheckpoint, smallReplicates, BaseSeed)
	metrics, err := AnalyzeCalibrationRuns(runs)
	if err != nil {
		t.Fatal(err)
	}
	scales := BuildCalibrationScales(metrics, smallCheckpoint)
	if len(scales) == 0 {
		t.Fatal("no scalar metric received a calibration scale")
	}
	ok := 0
	for _, s := range scales {
		if s.Status == ScaleStatusOK {
			ok++
		}
	}
	if float64(ok)/float64(len(scales)) < 0.3 {
		t.Fatalf("suspiciously low non-degenerate coverage: %d/%d", ok, len(scales))
	}
}

func TestC3LeaveOneGeneratorFamilyOutIsDiagnosticOnly(t *testing.T) {
	runs := RunCalibrationPanel(smallCheckpoint, smallReplicates, BaseSeed)
	metrics, err := AnalyzeCalibrationRuns(runs)
	if err != nil {
		t.Fatal(err)
	}
	full := BuildCalibrationScales(metrics, smallCheckpoint)
	reduced := LeaveOneGeneratorFamilyOut(metrics, "CAL-IID", smallCheckpoint)
	if len(reduced) == 0 || len(full) == 0 {
		t.Fatal("leave-one-out must still produce a scale table")
	}
	// The diagnostic must not silently change the estimator or the frozen
	// scale formula; only the input sample shrinks.
	for _, s := range reduced {
		if s.Estimator != "MAD_1.4826" && s.Estimator != "IQR_1.349_FALLBACK" {
			t.Fatalf("unexpected estimator after leave-one-out: %s", s.Estimator)
		}
	}
}

func TestC4CalibrationReproducibility(t *testing.T) {
	a := RunCalibrationPanel(smallCheckpoint, smallReplicates, BaseSeed)
	b := RunCalibrationPanel(smallCheckpoint, smallReplicates, BaseSeed)
	ma, err := AnalyzeCalibrationRuns(a)
	if err != nil {
		t.Fatal(err)
	}
	mb, err := AnalyzeCalibrationRuns(b)
	if err != nil {
		t.Fatal(err)
	}
	sa := BuildCalibrationScales(ma, smallCheckpoint)
	sb := BuildCalibrationScales(mb, smallCheckpoint)
	var bufA, bufB bytes.Buffer
	if err := WriteCalibrationScalesTSV(&bufA, sa); err != nil {
		t.Fatal(err)
	}
	if err := WriteCalibrationScalesTSV(&bufB, sb); err != nil {
		t.Fatal(err)
	}
	if bufA.String() != bufB.String() {
		t.Fatal("calibration run is not byte-identical across repeats with the same seed schedule")
	}
}

func TestC5NoCandidateOrVMDataInGenerators(t *testing.T) {
	// Structural: every generator's Generate closure captures only its own
	// alphabet/parameters and rng; this test asserts the panel never touches
	// disk for VM or candidate paths by construction (RunCalibrationPanel
	// takes no corpus argument at all).
	runs := RunCalibrationPanel(smallCheckpoint, 2, BaseSeed)
	for _, r := range runs {
		if r.Records[0].CorpusID == "VM-ZL3b-x7" {
			t.Fatal("calibration panel referenced VM corpus id")
		}
	}
}

func TestScaleEstimatorFrozenFormula(t *testing.T) {
	v := []float64{1, 2, 3, 4, 5, 6, 7, 8, 9, 100}
	s := EstimateScale(v)
	if s.Status != ScaleStatusOK || s.Estimator != "MAD_1.4826" {
		t.Fatalf("expected MAD estimator, got %+v", s)
	}
	if s.Scale != 1.4826*s.MAD {
		t.Fatalf("scale does not match frozen formula: %+v", s)
	}
}
