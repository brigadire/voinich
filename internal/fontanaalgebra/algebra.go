// Package fontanaalgebra validates the bounded, provenance-aware operation
// algebra frozen by task80. It intentionally models metadata, not Fontana
// devices themselves; execution remains in the task76/task78 packages.
package fontanaalgebra

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type Provenance string

const (
	EFontana  Provenance = "E-FONTANA"
	IFontana  Provenance = "I-FONTANA"
	VFontana  Provenance = "V-FONTANA"
	CFontana  Provenance = "C-FONTANA"
	GAllowed  Provenance = "G-ALLOWED"
	MOnly     Provenance = "M-ONLY"
	Forbidden Provenance = "FORBIDDEN"
	Unknown   Provenance = "UNKNOWN"
)

type Operation struct {
	ID               string     `json:"operation_id"`
	Version          string     `json:"version"`
	InputTypes       []string   `json:"input_types"`
	OutputTypes      []string   `json:"output_types"`
	Preconditions    []string   `json:"preconditions"`
	StateEffect      string     `json:"state_effect"`
	ObservableEffect string     `json:"observable_effect"`
	Knowledge        []string   `json:"knowledge_requirements"`
	Provenance       Provenance `json:"provenance_status"`
	Models           []string   `json:"validated_models"`
	Tests            []string   `json:"tests"`
}

type Composition struct {
	ID           string     `json:"composition_id"`
	Operations   []string   `json:"operations"`
	Class        string     `json:"class"`
	Provenance   Provenance `json:"provenance_status"`
	Models       []string   `json:"models"`
	Precondition string     `json:"precondition"`
}

type Algebra struct {
	Schema       string        `json:"schema"`
	Version      string        `json:"version"`
	DomainTypes  []DomainType  `json:"domain_types"`
	Operations   []Operation   `json:"operations"`
	Compositions []Composition `json:"compositions"`
	Fixtures     []Fixture     `json:"fixtures"`
}

type DomainType struct {
	ID         string `json:"type_id"`
	Carrier    string `json:"carrier"`
	Observable bool   `json:"directly_observable"`
	Knowledge  bool   `json:"interpretation_requires_prior_knowledge"`
}

type Fixture struct {
	ID     string `json:"fixture_id"`
	Class  string `json:"class"`
	Valid  bool   `json:"type_valid"`
	Status string `json:"historical_status"`
}

type FrozenManifest struct {
	Schema  string        `json:"schema"`
	Version string        `json:"version"`
	Models  []FrozenModel `json:"models"`
}

type FrozenModel struct {
	ID          string     `json:"model_id"`
	FreezeLevel string     `json:"freeze_level"`
	Checksums   []Checksum `json:"checksums"`
}

type Checksum struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

func LoadAlgebra(path string) (Algebra, error) {
	var algebra Algebra
	if err := loadJSON(path, &algebra); err != nil {
		return Algebra{}, err
	}
	if err := ValidateAlgebra(algebra); err != nil {
		return Algebra{}, err
	}
	return algebra, nil
}

func LoadFrozenManifest(path string) (FrozenManifest, error) {
	var manifest FrozenManifest
	if err := loadJSON(path, &manifest); err != nil {
		return FrozenManifest{}, err
	}
	if manifest.Schema != "fontana-models-frozen-v1" || manifest.Version != "V1" {
		return FrozenManifest{}, fmt.Errorf("unexpected frozen manifest %q/%q", manifest.Schema, manifest.Version)
	}
	seen := map[string]bool{}
	for _, model := range manifest.Models {
		if model.ID == "" || seen[model.ID] {
			return FrozenManifest{}, fmt.Errorf("empty or duplicate frozen model ID %q", model.ID)
		}
		seen[model.ID] = true
		if model.FreezeLevel != "FROZEN_CORE" && model.FreezeLevel != "FROZEN_PROFILE" {
			return FrozenManifest{}, fmt.Errorf("%s has non-frozen level %q", model.ID, model.FreezeLevel)
		}
		if len(model.Checksums) == 0 {
			return FrozenManifest{}, fmt.Errorf("%s has no frozen checksums", model.ID)
		}
	}
	return manifest, nil
}

func ValidateAlgebra(a Algebra) error {
	if a.Schema != "fontana-operation-algebra-v1" || a.Version != "V1" {
		return fmt.Errorf("unexpected algebra %q/%q", a.Schema, a.Version)
	}
	if len(a.DomainTypes) == 0 || len(a.Operations) == 0 || len(a.Compositions) == 0 {
		return fmt.Errorf("algebra requires types, operations, and compositions")
	}
	types := map[string]bool{}
	for _, t := range a.DomainTypes {
		if t.ID == "" || types[t.ID] {
			return fmt.Errorf("empty or duplicate type %q", t.ID)
		}
		types[t.ID] = true
	}
	operations := map[string]bool{}
	for _, op := range a.Operations {
		if op.ID == "" || operations[op.ID] || op.Version == "" || op.StateEffect == "" || op.ObservableEffect == "" {
			return fmt.Errorf("incomplete or duplicate operation %q", op.ID)
		}
		operations[op.ID] = true
		if op.Provenance != EFontana && op.Provenance != IFontana && op.Provenance != VFontana {
			return fmt.Errorf("%s is not a historical primitive: %s", op.ID, op.Provenance)
		}
		for _, typ := range append(append([]string{}, op.InputTypes...), op.OutputTypes...) {
			if !types[typ] {
				return fmt.Errorf("%s refers to unknown type %q", op.ID, typ)
			}
		}
	}
	for _, composition := range a.Compositions {
		if composition.ID == "" || len(composition.Operations) == 0 || composition.Class == "" {
			return fmt.Errorf("incomplete composition %q", composition.ID)
		}
		for _, id := range composition.Operations {
			if !operations[id] {
				return fmt.Errorf("%s refers to unknown operation %q", composition.ID, id)
			}
		}
		if composition.Class == "ATTESTED" && composition.Provenance != CFontana {
			return fmt.Errorf("%s: attested composition needs C-FONTANA", composition.ID)
		}
		if composition.Class == "TYPE_VALID_UNATTESTED" && composition.Provenance != GAllowed {
			return fmt.Errorf("%s: unattested composition needs G-ALLOWED", composition.ID)
		}
		if composition.Class == "INVALID" && composition.Provenance != Forbidden {
			return fmt.Errorf("%s: invalid composition needs FORBIDDEN", composition.ID)
		}
	}
	for _, fixture := range a.Fixtures {
		if fixture.ID == "" || fixture.Class == "" || fixture.Status == "" {
			return fmt.Errorf("incomplete fixture %q", fixture.ID)
		}
		if fixture.Valid && fixture.Status == "F-EXACT" {
			return fmt.Errorf("fixture %s must not claim to be Fontana", fixture.ID)
		}
	}
	return nil
}

func VerifyChecksums(root string, manifest FrozenManifest) error {
	var failures []string
	for _, model := range manifest.Models {
		for _, sum := range model.Checksums {
			data, err := os.ReadFile(filepath.Join(root, sum.Path))
			if err != nil {
				failures = append(failures, fmt.Sprintf("%s: %v", sum.Path, err))
				continue
			}
			actual := sha256.Sum256(data)
			if hex.EncodeToString(actual[:]) != sum.SHA256 {
				failures = append(failures, fmt.Sprintf("%s: checksum mismatch", sum.Path))
			}
		}
	}
	sort.Strings(failures)
	if len(failures) > 0 {
		return fmt.Errorf("freeze verification failed: %v", failures)
	}
	return nil
}

func loadJSON(path string, into any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, into)
}
