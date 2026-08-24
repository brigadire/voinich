package mnemonicspace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"zcore.dev/voinich/internal/fontanaalgebra"
)

const (
	ExpectedTask80AlgebraSHA256 = "5abf3665df5aeb4eee271045b317bd460c8f91422d64fbb80e6efe0b0adc6f76"
	ExpectedTask80FrozenSHA256  = "89993922009fcf53ee64875455cb37078429c1a604b56dd3e430b92f035d7030"
)

type Authority struct {
	Algebra       fontanaalgebra.Algebra
	Frozen        fontanaalgebra.FrozenManifest
	Operations    map[OperationID]fontanaalgebra.Operation
	Models        map[string]fontanaalgebra.FrozenModel
	FutureStatus  map[HistoricalStatus]bool
	ReferenceOnly map[string]bool
	Excluded      map[string]bool
	AlgebraSHA256 string
	FrozenSHA256  string
}

type rawAlgebra struct {
	FutureModelStatuses []string `json:"future_model_statuses"`
}

type rawFrozen struct {
	ReferenceOnly []struct {
		ModelID string `json:"model_id"`
	} `json:"reference_only"`
	Excluded []string `json:"excluded"`
}

func LoadTask80Authority(root string) (Authority, error) {
	algebraPath := filepath.Join(root, "FONTANA_OPERATION_ALGEBRA_V1.json")
	frozenPath := filepath.Join(root, "FONTANA_MODELS_FROZEN_V1.json")
	algebra, err := fontanaalgebra.LoadAlgebra(algebraPath)
	if err != nil {
		return Authority{}, err
	}
	frozen, err := fontanaalgebra.LoadFrozenManifest(frozenPath)
	if err != nil {
		return Authority{}, err
	}
	algebraSHA, err := fileSHA256(algebraPath)
	if err != nil {
		return Authority{}, err
	}
	frozenSHA, err := fileSHA256(frozenPath)
	if err != nil {
		return Authority{}, err
	}
	if algebraSHA != ExpectedTask80AlgebraSHA256 {
		return Authority{}, fmt.Errorf("task80 algebra checksum drift: %s", algebraSHA)
	}
	if frozenSHA != ExpectedTask80FrozenSHA256 {
		return Authority{}, fmt.Errorf("task80 frozen checksum drift: %s", frozenSHA)
	}
	var interfaceRaw rawAlgebra
	if err := loadRaw(algebraPath, &interfaceRaw); err != nil {
		return Authority{}, err
	}
	var frozenRaw rawFrozen
	if err := loadRaw(frozenPath, &frozenRaw); err != nil {
		return Authority{}, err
	}
	a := Authority{
		Algebra:       algebra,
		Frozen:        frozen,
		Operations:    map[OperationID]fontanaalgebra.Operation{},
		Models:        map[string]fontanaalgebra.FrozenModel{},
		FutureStatus:  map[HistoricalStatus]bool{},
		ReferenceOnly: map[string]bool{},
		Excluded:      map[string]bool{},
		AlgebraSHA256: algebraSHA,
		FrozenSHA256:  frozenSHA,
	}
	for _, op := range algebra.Operations {
		a.Operations[OperationID(op.ID)] = op
	}
	for _, model := range frozen.Models {
		a.Models[model.ID] = model
	}
	for _, status := range interfaceRaw.FutureModelStatuses {
		a.FutureStatus[HistoricalStatus(status)] = true
	}
	for _, model := range frozenRaw.ReferenceOnly {
		a.ReferenceOnly[model.ModelID] = true
	}
	for _, id := range frozenRaw.Excluded {
		a.Excluded[id] = true
	}
	return a, nil
}

func fileSHA256(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:]), nil
}

func loadRaw(path string, into any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, into)
}
