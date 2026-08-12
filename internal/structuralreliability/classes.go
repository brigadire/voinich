package structuralreliability

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math"
	"os"

	"gopkg.in/yaml.v3"
	"zcore.dev/voinich/internal/normalization"
)

// ReadClasses loads a structural_classes.yaml file, reusing the exact same
// normalization.ClassesOutput schema every other tool in this repository
// reads and writes.
func ReadClasses(path string) (normalization.ClassesOutput, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return normalization.ClassesOutput{}, "", err
	}
	var result normalization.ClassesOutput
	if err := yaml.Unmarshal(data, &result); err != nil {
		return normalization.ClassesOutput{}, "", err
	}
	sum := sha256.Sum256(data)
	return result, hex.EncodeToString(sum[:]), nil
}

// SelectModel returns the model matching the given threshold, exactly as
// structural-profile-stability's own selectModel does.
func SelectModel(models []normalization.Model, threshold float64) (normalization.Model, error) {
	for _, model := range models {
		if math.Abs(model.Threshold-threshold) < 1e-12 {
			return model, nil
		}
	}
	return normalization.Model{}, fmt.Errorf("class model threshold %.2f is absent", threshold)
}
