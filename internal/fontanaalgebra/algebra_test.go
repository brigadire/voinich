package fontanaalgebra

import (
	"path/filepath"
	"testing"
)

func task80Path(name string) string {
	return filepath.Join("..", "..", "research", "phase2", "fontana", "task80", name)
}

func TestPublishedAlgebraIsValid(t *testing.T) {
	algebra, err := LoadAlgebra(task80Path("FONTANA_OPERATION_ALGEBRA_V1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(algebra.Fixtures) != 7 {
		t.Fatalf("fixtures = %d, want 7", len(algebra.Fixtures))
	}
}

func TestFrozenManifestChecksums(t *testing.T) {
	manifest, err := LoadFrozenManifest(task80Path("FONTANA_MODELS_FROZEN_V1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyChecksums(filepath.Join("..", ".."), manifest); err != nil {
		t.Fatal(err)
	}
}

func TestRejectsHistoricalFixture(t *testing.T) {
	err := ValidateAlgebra(Algebra{
		Schema: "fontana-operation-algebra-v1", Version: "V1",
		DomainTypes:  []DomainType{{ID: "State"}},
		Operations:   []Operation{{ID: "observe", Version: "v1", InputTypes: []string{"State"}, OutputTypes: []string{"State"}, StateEffect: "none", ObservableEffect: "state", Provenance: EFontana}},
		Compositions: []Composition{{ID: "observe", Operations: []string{"observe"}, Class: "ATTESTED", Provenance: CFontana}},
		Fixtures:     []Fixture{{ID: "bad", Class: "literal", Valid: true, Status: "F-EXACT"}},
	})
	if err == nil {
		t.Fatal("accepted a synthetic fixture claiming F-EXACT")
	}
}
