package mnemonicspace

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
)

type Job struct {
	MechanismID       string
	ParameterSetID    string
	InputID           string
	RecoveryCondition RecoveryCondition
	Replicate         int
	MasterSeed        uint64
}

func (j Job) DerivedSeed() uint64 {
	sum := sha256.Sum256([]byte(fmt.Sprintf("%d|%s|%s|%s|%s|%d", j.MasterSeed, j.MechanismID, j.ParameterSetID, j.InputID, j.RecoveryCondition, j.Replicate)))
	return binary.BigEndian.Uint64(sum[:8])
}

func (j Job) ID(spec MechanismSpec, authority Authority) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%s|%s|%d", spec.ID, spec.Version, authority.AlgebraSHA256[:16], authority.FrozenSHA256[:16], j.ParameterSetID, j.InputID, j.DerivedSeed())
	sum := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(sum[:16])
}

type CheckpointMetadata struct {
	JobID                    string            `json:"job_id"`
	MechanismID              string            `json:"mechanism_id"`
	ParameterSetID           string            `json:"parameter_set_id"`
	SpecVersion              string            `json:"spec_version"`
	RecoveryCondition        RecoveryCondition `json:"recovery_condition"`
	DerivedSeed              uint64            `json:"derived_seed"`
	Task80AlgebraSHA256      string            `json:"task80_algebra_sha256"`
	Task80FrozenSHA256       string            `json:"task80_frozen_sha256"`
	ObservableDocumentSHA256 string            `json:"observable_document_sha256"`
	RecoverySHA256           string            `json:"recovery_sha256"`
}

func NewCheckpointMetadata(job Job, spec MechanismSpec, authority Authority, exec Execution) (CheckpointMetadata, error) {
	recoverySHA, err := hashJSON(exec.Recovery)
	if err != nil {
		return CheckpointMetadata{}, err
	}
	return CheckpointMetadata{
		JobID:                    job.ID(spec, authority),
		MechanismID:              spec.ID,
		ParameterSetID:           job.ParameterSetID,
		SpecVersion:              spec.Version,
		RecoveryCondition:        job.RecoveryCondition,
		DerivedSeed:              job.DerivedSeed(),
		Task80AlgebraSHA256:      authority.AlgebraSHA256,
		Task80FrozenSHA256:       authority.FrozenSHA256,
		ObservableDocumentSHA256: exec.Prepared.Document.Checksum(),
		RecoverySHA256:           recoverySHA,
	}, nil
}

func (c CheckpointMetadata) Matches(job Job, spec MechanismSpec, authority Authority) bool {
	return c.JobID == job.ID(spec, authority) &&
		c.MechanismID == spec.ID &&
		c.ParameterSetID == job.ParameterSetID &&
		c.SpecVersion == spec.Version &&
		c.RecoveryCondition == job.RecoveryCondition &&
		c.DerivedSeed == job.DerivedSeed() &&
		c.Task80AlgebraSHA256 == authority.AlgebraSHA256 &&
		c.Task80FrozenSHA256 == authority.FrozenSHA256
}

func hashJSON(v any) (string, error) {
	payload, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:]), nil
}
