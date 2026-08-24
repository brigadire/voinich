// Command fingerprint-v2-verify verifies every file, parent edge, and embedded
// checksum in a Fingerprint V2 transitive provenance manifest.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Manifest struct {
	SchemaVersion int        `json:"schema_version"`
	Version       string     `json:"version"`
	Status        string     `json:"status"`
	Artifacts     []Artifact `json:"artifacts"`
}

type Artifact struct {
	ID             string    `json:"id"`
	Path           string    `json:"path"`
	SHA256         string    `json:"sha256"`
	Kind           string    `json:"kind"`
	Producer       string    `json:"producer"`
	ProducerCommit string    `json:"producer_git_commit"`
	Arguments      []string  `json:"producer_arguments,omitempty"`
	Parents        []Parent  `json:"parents,omitempty"`
	Bindings       []Binding `json:"embedded_bindings,omitempty"`
}

type Parent struct {
	ID     string `json:"id"`
	SHA256 string `json:"sha256"`
}
type Binding struct {
	JSONPath string `json:"json_path"`
	ParentID string `json:"parent_id"`
}

func main() {
	manifest := flag.String("manifest", "research/phase2/task83a/FINGERPRINT_V2_REFREEZE_MANIFEST.json", "transitive provenance manifest")
	root := flag.String("root", ".", "repository root used to resolve relative paths")
	allowNonAuthoritative := flag.Bool("allow-non-authoritative", false, "audit checksums/bindings even when manifest is explicitly non-authoritative")
	flag.Parse()
	if err := verify(*manifest, *root, *allowNonAuthoritative); err != nil {
		fmt.Fprintln(os.Stderr, "fingerprint-v2-verify:", err)
		os.Exit(1)
	}
	if *allowNonAuthoritative {
		fmt.Printf("Fingerprint V2 integrity graph verified (non-authoritative audit mode): %s\n", *manifest)
		return
	}
	fmt.Printf("Fingerprint V2 authoritative provenance verified: %s\n", *manifest)
}

func Verify(manifestPath, root string) error {
	return verify(manifestPath, root, false)
}

func verify(manifestPath, root string, allowNonAuthoritative bool) error {
	b, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	var m Manifest
	if err = json.Unmarshal(b, &m); err != nil {
		return err
	}
	if m.SchemaVersion != 1 {
		return fmt.Errorf("unsupported schema_version %d", m.SchemaVersion)
	}
	byID := map[string]Artifact{}
	for _, a := range m.Artifacts {
		if a.ID == "" || a.Path == "" || a.SHA256 == "" {
			return errors.New("artifact missing id/path/sha256")
		}
		if _, ok := byID[a.ID]; ok {
			return fmt.Errorf("duplicate artifact id %q", a.ID)
		}
		byID[a.ID] = a
	}
	if err := verifyGraph(byID); err != nil {
		return err
	}
	for _, a := range m.Artifacts {
		p := a.Path
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		actual, err := fileSHA(p)
		if err != nil {
			return fmt.Errorf("%s: %w", a.ID, err)
		}
		if actual != a.SHA256 {
			return fmt.Errorf("%s checksum mismatch: expected %s got %s", a.ID, a.SHA256, actual)
		}
		for _, edge := range a.Parents {
			parent, ok := byID[edge.ID]
			if !ok {
				return fmt.Errorf("%s unknown parent %q", a.ID, edge.ID)
			}
			if edge.SHA256 != parent.SHA256 {
				return fmt.Errorf("%s parent binding mismatch for %s", a.ID, edge.ID)
			}
		}
		for _, binding := range a.Bindings {
			parent, ok := byID[binding.ParentID]
			if !ok {
				return fmt.Errorf("%s embedded binding has unknown parent %q", a.ID, binding.ParentID)
			}
			value, err := jsonStringAt(p, binding.JSONPath)
			if err != nil {
				return fmt.Errorf("%s binding %s: %w", a.ID, binding.JSONPath, err)
			}
			if value != parent.SHA256 {
				return fmt.Errorf("%s embedded binding mismatch at %s: expected parent %s got %s", a.ID, binding.JSONPath, parent.SHA256, value)
			}
		}
	}
	if m.Status != "" && m.Status != "AUTHORITATIVE" && !allowNonAuthoritative {
		return fmt.Errorf("manifest status %q is not authoritative", m.Status)
	}
	return nil
}

func verifyGraph(byID map[string]Artifact) error {
	state := make(map[string]uint8, len(byID))
	var visit func(string) error
	visit = func(id string) error {
		switch state[id] {
		case 1:
			return fmt.Errorf("provenance cycle at %q", id)
		case 2:
			return nil
		}
		state[id] = 1
		a := byID[id]
		if a.Kind != "raw" && a.Kind != "historical" && len(a.Parents) == 0 {
			return fmt.Errorf("%s generated/normative artifact has no parents", id)
		}
		for _, edge := range a.Parents {
			if _, ok := byID[edge.ID]; !ok {
				return fmt.Errorf("%s unknown parent %q", id, edge.ID)
			}
			if err := visit(edge.ID); err != nil {
				return err
			}
		}
		state[id] = 2
		return nil
	}
	for id := range byID {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}

func fileSHA(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:]), nil
}
func jsonStringAt(path, dotted string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var v any
	if err = json.Unmarshal(b, &v); err != nil {
		return "", err
	}
	cur := v
	for _, part := range strings.Split(dotted, ".") {
		switch node := cur.(type) {
		case map[string]any:
			var ok bool
			cur, ok = node[part]
			if !ok {
				return "", fmt.Errorf("missing %q", part)
			}
		case []any:
			i, convErr := strconv.Atoi(part)
			if convErr != nil || i < 0 || i >= len(node) {
				return "", fmt.Errorf("invalid array index %q", part)
			}
			cur = node[i]
		default:
			return "", fmt.Errorf("cannot descend through %q", part)
		}
	}
	s, ok := cur.(string)
	if !ok {
		return "", errors.New("bound value is not a string")
	}
	return s, nil
}
