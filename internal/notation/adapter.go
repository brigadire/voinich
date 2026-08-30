package notation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
)

type SourceDocument struct {
	CorpusID       string       `json:"corpus_id"`
	ClassID        string       `json:"class_id"`
	Representation string       `json:"representation_id"`
	DocumentID     string       `json:"document_id"`
	Units          []SourceUnit `json:"units"`
}

type SourceUnit struct {
	Section      string            `json:"section,omitempty"`
	Page         string            `json:"page,omitempty"`
	Locus        string            `json:"locus,omitempty"`
	PhysicalLine string            `json:"physical_line,omitempty"`
	Token        string            `json:"token"`
	Symbols      []string          `json:"symbols"`
	Attributes   map[string]string `json:"attributes,omitempty"`
}

// NormalizeFixture is the shared declarative adapter used by all class
// adapters. Class-specific code only validates required dimensions; it never
// changes generic analyzer behavior.
func NormalizeFixture(src SourceDocument) ([]Record, error) {
	if src.CorpusID == "" || src.ClassID == "" || src.Representation == "" || src.DocumentID == "" {
		return nil, fmt.Errorf("source identity fields are required")
	}
	if len(src.Units) == 0 {
		return nil, fmt.Errorf("source has no units")
	}
	lineCounters := map[string]int{}
	out := make([]Record, 0, len(src.Units))
	for i, u := range src.Units {
		if u.Token == "" || len(u.Symbols) == 0 {
			return nil, fmt.Errorf("unit %d: token and symbols required", i)
		}
		if src.ClassID == "C06" {
			for _, k := range requiredMusicAttributes(src.Representation) {
				if u.Attributes[k] == "" {
					return nil, fmt.Errorf("unit %d: music representation %s requires %s", i, src.Representation, k)
				}
			}
		}
		if src.ClassID == "C07" {
			for _, k := range []string{"tradition", "simultaneity_group", "course", "fret"} {
				if u.Attributes[k] == "" {
					return nil, fmt.Errorf("unit %d: tablature requires %s", i, k)
				}
			}
		}
		key := src.DocumentID + "\x1f" + u.Section + "\x1f" + u.Page + "\x1f" + u.Locus + "\x1f" + u.PhysicalLine
		idx := lineCounters[key]
		lineCounters[key]++
		h := sha256.Sum256([]byte(fmt.Sprintf("%s\x00%s\x00%d", src.CorpusID, src.Representation, i)))
		out = append(out, Record{SchemaVersion: SchemaVersion, CorpusID: src.CorpusID, Representation: src.Representation, Document: level(src.DocumentID), Section: level(u.Section), Page: level(u.Page), Locus: level(u.Locus), PhysicalLine: level(u.PhysicalLine), TokenID: src.CorpusID + "-" + hex.EncodeToString(h[:6]), TokenIndex: idx, Token: u.Token, Symbols: append([]string(nil), u.Symbols...), Attributes: cloneMap(u.Attributes)})
	}
	if err := Validate(out); err != nil {
		return nil, err
	}
	return out, nil
}

func requiredMusicAttributes(rep string) []string {
	switch rep {
	case "MUSIC-R1":
		return []string{"event", "voice", "staff", "system"}
	case "MUSIC-R2":
		return []string{"interval", "voice", "staff", "system"}
	case "MUSIC-R3":
		return []string{"pitch", "duration", "voice", "staff", "system"}
	default:
		return []string{"voice", "staff", "system"}
	}
}
func level(v string) ObservedLevel { return ObservedLevel{Value: v, Observed: v != ""} }
func cloneMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := map[string]string{}
	for k, v := range m {
		out[k] = v
	}
	return out
}

func ValidateClassCoverage(classes []SourceDocument) error {
	seen := map[string]bool{}
	for _, s := range classes {
		seen[s.ClassID] = true
	}
	var missing []string
	for i := 1; i <= 10; i++ {
		id := fmt.Sprintf("C%02d", i)
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		return fmt.Errorf("missing fixture adapters: %v", missing)
	}
	return nil
}
