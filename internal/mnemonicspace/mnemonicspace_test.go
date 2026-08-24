package mnemonicspace

import (
	"path/filepath"
	"strings"
	"testing"
)

func task80Dir() string {
	return filepath.Join("..", "..", "research", "phase2", "fontana", "task80")
}

func loadAuthority(t *testing.T) Authority {
	t.Helper()
	authority, err := LoadTask80Authority(task80Dir())
	if err != nil {
		t.Fatal(err)
	}
	return authority
}

func lookupSpec(t *testing.T, id string) MechanismSpec {
	t.Helper()
	for _, spec := range FrozenRegistry() {
		if spec.ID == id {
			return spec
		}
	}
	t.Fatalf("missing spec %s", id)
	return MechanismSpec{}
}

func lookupParam(t *testing.T, spec MechanismSpec, id string) ParameterSet {
	t.Helper()
	param, ok := spec.ParameterSet(id)
	if !ok {
		t.Fatalf("missing parameter set %s for %s", id, spec.ID)
	}
	return param
}

func TestTask80AuthorityBinding(t *testing.T) {
	authority := loadAuthority(t)
	if authority.AlgebraSHA256 != ExpectedTask80AlgebraSHA256 {
		t.Fatalf("algebra checksum = %s", authority.AlgebraSHA256)
	}
	if authority.FrozenSHA256 != ExpectedTask80FrozenSHA256 {
		t.Fatalf("frozen checksum = %s", authority.FrozenSHA256)
	}
	if !authority.FutureStatus[StatusMRestricted] || !authority.FutureStatus[StatusFProfile] {
		t.Fatal("task80 future statuses missing required entries")
	}
	if !authority.ReferenceOnly["F07_ROTA"] || !authority.ReferenceOnly["F10_CYLINDRUS"] {
		t.Fatal("task80 reference-only markers missing F07/F10")
	}
}

func TestRegistryValidation(t *testing.T) {
	authority := loadAuthority(t)
	if err := ValidateRegistry(authority, FrozenRegistry()); err != nil {
		t.Fatal(err)
	}
	for _, spec := range FrozenRegistry() {
		for _, model := range spec.SourceModels {
			if authority.ReferenceOnly[model] || authority.Excluded[model] {
				t.Fatalf("%s uses forbidden model %s", spec.ID, model)
			}
		}
	}
}

func TestInvalidMechanismRejected(t *testing.T) {
	authority := loadAuthority(t)
	for _, spec := range InvalidControls() {
		if err := ValidateMechanism(authority, spec); err == nil {
			t.Fatalf("expected invalid control %s to be rejected", spec.ID)
		}
	}
}

func TestHistoricalRecoveryPaths(t *testing.T) {
	runner := Runner{}
	f01 := lookupSpec(t, "f01_speculum_profile_latin23_r12")
	f01Param := lookupParam(t, f01, "f01-profile-latin23-r12")
	f01Exec, err := runner.Run(f01, f01Param, InputModel{ID: "m1", Sequence: []Symbol{"M", "A", "G", "I", "S", "T", "E", "R"}}, 7, RetrievalRequest{Condition: RecoveryFullKnowledge}, RecoveryEnvironment{Convention: &ConventionKnowledge{MessageLength: 8}})
	if err != nil {
		t.Fatal(err)
	}
	if f01Exec.Recovery.Class != ResultExact || f01Exec.Recovery.Value != "MAGISTER" {
		t.Fatalf("F01 recovery = %#v", f01Exec.Recovery)
	}

	f08 := lookupSpec(t, "f08_serpens_core")
	f08Param := lookupParam(t, f08, "f08-centre-edge-small")
	f08Exec, err := runner.Run(f08, f08Param, InputModel{ID: "m2", Sequence: []Symbol{"S", "E", "R", "P", "E", "N", "S"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge}, RecoveryEnvironment{Geometry: &GeometryKnowledge{Path: []Position{0, 1, 2}}, Convention: &ConventionKnowledge{MessageLength: 7}})
	if err != nil {
		t.Fatal(err)
	}
	if f08Exec.Recovery.Class != ResultExact || f08Exec.Recovery.Value != "SERPENS" {
		t.Fatalf("F08 recovery = %#v", f08Exec.Recovery)
	}

	f11 := lookupSpec(t, "f11_arismetricum_core")
	f11Param := lookupParam(t, f11, "f11-core-default")
	idx := Index(3)
	f11CueOnly, err := runner.Run(f11, f11Param, InputModel{ID: "m3", IndexedCues: map[Index]Cue{3: "owl"}}, 0, RetrievalRequest{Condition: RecoveryNoConvention, TargetIndex: &idx}, RecoveryEnvironment{})
	if err != nil {
		t.Fatal(err)
	}
	if f11CueOnly.Recovery.Class != ResultCueOnly || f11CueOnly.Recovery.Cue != "owl" {
		t.Fatalf("F11 cue-only = %#v", f11CueOnly.Recovery)
	}
	f11Exact, err := runner.Run(f11, f11Param, InputModel{ID: "m3", IndexedCues: map[Index]Cue{3: "owl"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge, TargetIndex: &idx}, RecoveryEnvironment{Convention: &ConventionKnowledge{CueMeanings: map[Cue][]RetrievedItem{"owl": {"MERCURY"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if f11Exact.Recovery.Class != ResultExact || f11Exact.Recovery.Value != "MERCURY" {
		t.Fatalf("F11 exact = %#v", f11Exact.Recovery)
	}

	f12 := lookupSpec(t, "f12_horalogius_core")
	f12Param := lookupParam(t, f12, "f12-cycle-default")
	f12CueOnly, err := runner.Run(f12, f12Param, InputModel{ID: "m4", TimedCues: map[int]Cue{1: "bell"}}, 0, RetrievalRequest{Condition: RecoveryNoInternal}, RecoveryEnvironment{})
	if err != nil {
		t.Fatal(err)
	}
	if f12CueOnly.Recovery.Class != ResultCueOnly || f12CueOnly.Recovery.Cue != "bell" {
		t.Fatalf("F12 cue-only = %#v", f12CueOnly.Recovery)
	}
	f12Exact, err := runner.Run(f12, f12Param, InputModel{ID: "m4", TimedCues: map[int]Cue{1: "bell"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge}, RecoveryEnvironment{InternalMemory: &InternalMemoryState{Associations: map[Cue][]RetrievedItem{"bell": {"PRIME"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if f12Exact.Recovery.Class != ResultExact || f12Exact.Recovery.Value != "PRIME" {
		t.Fatalf("F12 exact = %#v", f12Exact.Recovery)
	}
}

func TestRestrictedControlsAmbiguityAndContext(t *testing.T) {
	runner := Runner{}
	rotation := lookupSpec(t, "m_restricted_rotation_index")
	rotationParam := lookupParam(t, rotation, "rotation-index-small")
	idx := Index(0)
	rotationExec, err := runner.Run(rotation, rotationParam, InputModel{ID: "m5", IndexedCues: map[Index]Cue{0: "owl", 1: "bell"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge, TargetIndex: &idx}, RecoveryEnvironment{Convention: &ConventionKnowledge{CueMeanings: map[Cue][]RetrievedItem{"bell": {"EAST"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if rotationExec.Recovery.Class != ResultExact || rotationExec.Recovery.Value != "EAST" {
		t.Fatalf("rotation+index recovery = %#v", rotationExec.Recovery)
	}

	storage := lookupSpec(t, "synthetic_ambiguous")
	storageParam := lookupParam(t, storage, "storage-associate-small")
	pos := Position(0)
	ambiguous, err := runner.Run(storage, storageParam, InputModel{ID: "m6", PositionedCues: map[Position]Cue{0: "owl"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge, TargetPosition: &pos}, RecoveryEnvironment{InternalMemory: &InternalMemoryState{Associations: map[Cue][]RetrievedItem{"owl": {"ALPHA", "BETA"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if ambiguous.Recovery.Class != ResultAmbiguitySet || len(ambiguous.Recovery.Candidates) != 2 {
		t.Fatalf("ambiguous recovery = %#v", ambiguous.Recovery)
	}
	narrowed, err := runner.Run(storage, storageParam, InputModel{ID: "m6", PositionedCues: map[Position]Cue{0: "owl"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge, TargetPosition: &pos}, RecoveryEnvironment{InternalMemory: &InternalMemoryState{Associations: map[Cue][]RetrievedItem{"owl": {"ALPHA", "BETA"}}}, Context: &ContextKnowledge{Allowed: []RetrievedItem{"BETA"}}})
	if err != nil {
		t.Fatal(err)
	}
	if narrowed.Recovery.Class != ResultExact || narrowed.Recovery.Value != "BETA" {
		t.Fatalf("narrowed recovery = %#v", narrowed.Recovery)
	}
	collisions := DetectObservableCollisions([]CollisionSample{{InputID: "m6a", Intended: "ALPHA", Document: ambiguous.Prepared.Document}, {InputID: "m6b", Intended: "BETA", Document: narrowed.Prepared.Document}})
	if len(collisions) != 1 || len(collisions[0].Intendeds) != 2 {
		t.Fatalf("collisions = %#v", collisions)
	}
}

func TestNegativeControlsAndLeakageProtection(t *testing.T) {
	runner := Runner{}
	f01 := lookupSpec(t, "f01_speculum_core")
	f01Param := lookupParam(t, f01, "f01-core-bounded")
	prepared, err := runner.Prepare(f01, f01Param, InputModel{ID: "m7", Sequence: []Symbol{"M", "A", "G", "I", "S", "T", "E", "R"}}, 19)
	if err != nil {
		t.Fatal(err)
	}
	observableOnly, err := runner.Recover(f01, f01Param, prepared, RetrievalRequest{Condition: RecoveryObservable}, RecoveryEnvironment{})
	if err != nil {
		t.Fatal(err)
	}
	if observableOnly.Class == ResultExact {
		t.Fatalf("observable-only F01 leaked plaintext: %#v", observableOnly)
	}

	f08 := lookupSpec(t, "negative_randomized_path")
	f08Param := lookupParam(t, f08, "f08-centre-edge-small")
	pathAblation, err := runner.Run(f08, f08Param, InputModel{ID: "m8", Sequence: []Symbol{"S", "E", "R", "P", "E", "N", "S"}}, 0, RetrievalRequest{Condition: RecoveryNoGeometry}, RecoveryEnvironment{Convention: &ConventionKnowledge{MessageLength: 7}})
	if err != nil {
		t.Fatal(err)
	}
	if pathAblation.Recovery.Class != ResultAmbiguitySet {
		t.Fatalf("randomized path should be ambiguous, got %#v", pathAblation.Recovery)
	}

	f11 := lookupSpec(t, "negative_randomized_index_mapping")
	f11Param := lookupParam(t, f11, "f11-core-default")
	idx := Index(2)
	wrongMap, err := runner.Run(f11, f11Param, InputModel{ID: "m9", IndexedCues: map[Index]Cue{2: "owl"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge, TargetIndex: &idx}, RecoveryEnvironment{Convention: &ConventionKnowledge{CueMeanings: map[Cue][]RetrievedItem{"bell": {"WEST"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if wrongMap.Recovery.Class != ResultCueOnly {
		t.Fatalf("wrong index map should fall back to cue-only, got %#v", wrongMap.Recovery)
	}

	f12 := lookupSpec(t, "negative_randomized_cue_association")
	f12Param := lookupParam(t, f12, "f12-cycle-default")
	wrongAssoc, err := runner.Run(f12, f12Param, InputModel{ID: "m10", TimedCues: map[int]Cue{1: "bell"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge}, RecoveryEnvironment{InternalMemory: &InternalMemoryState{Associations: map[Cue][]RetrievedItem{"owl": {"DAWN"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if wrongAssoc.Recovery.Class != ResultCueOnly {
		t.Fatalf("wrong association should remain cue-only, got %#v", wrongAssoc.Recovery)
	}
}

func TestObservableDocumentAndCheckpointDeterminism(t *testing.T) {
	authority := loadAuthority(t)
	runner := Runner{}
	spec := lookupSpec(t, "f11_arismetricum_core")
	param := lookupParam(t, spec, "f11-core-default")
	idx := Index(5)
	execA, err := runner.Run(spec, param, InputModel{ID: "job-input", IndexedCues: map[Index]Cue{5: "owl"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge, TargetIndex: &idx}, RecoveryEnvironment{Convention: &ConventionKnowledge{CueMeanings: map[Cue][]RetrievedItem{"owl": {"MERCURY"}}}})
	if err != nil {
		t.Fatal(err)
	}
	execB, err := runner.Run(spec, param, InputModel{ID: "job-input", IndexedCues: map[Index]Cue{5: "owl"}}, 0, RetrievalRequest{Condition: RecoveryFullKnowledge, TargetIndex: &idx}, RecoveryEnvironment{Convention: &ConventionKnowledge{CueMeanings: map[Cue][]RetrievedItem{"owl": {"MERCURY"}}}})
	if err != nil {
		t.Fatal(err)
	}
	if execA.Prepared.Document.Checksum() != execB.Prepared.Document.Checksum() {
		t.Fatal("observable document checksum is not deterministic")
	}
	serialized := execA.Prepared.Document.Checksum() + "|" + strings.Join(execA.Prepared.Document.Symbols, ",")
	if strings.Contains(serialized, "offsets") || strings.Contains(serialized, "plaintext") || strings.Contains(serialized, "internal_memory") {
		t.Fatalf("observable document leaked hidden state: %s", serialized)
	}
	job := Job{MechanismID: spec.ID, ParameterSetID: param.ID, InputID: "job-input", RecoveryCondition: RecoveryFullKnowledge, Replicate: 0, MasterSeed: 11}
	job2 := Job{MechanismID: spec.ID, ParameterSetID: param.ID, InputID: "job-input", RecoveryCondition: RecoveryFullKnowledge, Replicate: 0, MasterSeed: 11}
	if job.DerivedSeed() != job2.DerivedSeed() || job.ID(spec, authority) != job2.ID(spec, authority) {
		t.Fatal("job identity is not deterministic")
	}
	job3 := Job{MechanismID: spec.ID, ParameterSetID: param.ID, InputID: "job-input", RecoveryCondition: RecoveryFullKnowledge, Replicate: 1, MasterSeed: 11}
	if job3.DerivedSeed() == job.DerivedSeed() || job3.ID(spec, authority) == job.ID(spec, authority) {
		t.Fatal("replicate did not perturb deterministic identity")
	}
	checkpoint, err := NewCheckpointMetadata(job, spec, authority, execA)
	if err != nil {
		t.Fatal(err)
	}
	if !checkpoint.Matches(job, spec, authority) {
		t.Fatalf("checkpoint mismatch: %#v", checkpoint)
	}
}
