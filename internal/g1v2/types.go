// Package g1v2 implements the execution and evidence transport contract for
// the frozen G1-v2 experiment.  It deliberately contains no model fitting,
// generation, metric, threshold-derivation, or corpus-reading code.
package g1v2

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"golang.org/x/text/unicode/norm"
)

const (
	Experiment       = "g1v2/task86c-v2"
	ProtocolVersion  = "g1v2-execution-v1"
	SchemaVersion    = "g1v2-evidence-v1"
	CanonicalVersion = "canonical-json-nfc-v1"
)

var validStages = map[string]bool{
	"FIT": true, "PREDICTIVE": true, "GENERATION": true,
	"STRUCTURAL": true, "AGGREGATION": true,
}

// JobBundle is the immutable scientific portion leased to a worker. DependsOn
// is DAG routing metadata and is itself scientific: the manifest, never the
// scheduler, decides reachability and dependencies.
type JobBundle struct {
	JobID            string   `json:"job_id"`
	Experiment       string   `json:"experiment"`
	ProtocolVersion  string   `json:"protocol_version"`
	Stage            string   `json:"stage"`
	CorpusID         string   `json:"corpus_id"`
	Model            string   `json:"model"`
	Candidate        string   `json:"candidate"`
	Scale            string   `json:"scale"`
	Replicate        int      `json:"replicate"`
	Seed             uint64   `json:"seed"`
	InputHashes      []string `json:"input_hashes"`
	DependencyHashes []string `json:"dependency_hashes"`
	CodeHash         string   `json:"code_hash"`
	ConfigHash       string   `json:"config_hash"`
	OutputSchema     string   `json:"output_schema"`
	DependsOn        []string `json:"depends_on,omitempty"`
	Work             WorkSpec `json:"work"`
}

// WorkSpec is an engineering execution descriptor. A production manifest may
// replace its executor implementation without changing bundle identity.
type WorkSpec struct {
	Kind       string `json:"kind"`
	Payload    string `json:"payload"`
	Iterations int    `json:"iterations,omitempty"`
}

type Manifest struct {
	SchemaVersion    string      `json:"schema_version"`
	CanonicalVersion string      `json:"canonical_version"`
	Jobs             []JobBundle `json:"jobs"`
}

// identity is intentionally explicit so operational fields can never leak
// into JobID. Field order is the frozen canonical order.
type identity struct {
	Experiment       string   `json:"experiment"`
	ProtocolVersion  string   `json:"protocol_version"`
	CorpusID         string   `json:"corpus_id"`
	Model            string   `json:"model"`
	Candidate        string   `json:"candidate"`
	Scale            string   `json:"scale"`
	Replicate        int      `json:"replicate"`
	Seed             uint64   `json:"seed"`
	Stage            string   `json:"stage"`
	InputHashes      []string `json:"input_hashes"`
	DependencyHashes []string `json:"dependency_hashes"`
	CodeHash         string   `json:"code_hash"`
	ConfigHash       string   `json:"config_hash"`
	OutputSchema     string   `json:"output_schema"`
	DependsOn        []string `json:"depends_on,omitempty"`
	Work             WorkSpec `json:"work"`
}

func HashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }

func ValidHash(s string) bool {
	b, err := hex.DecodeString(s)
	return err == nil && len(b) == sha256.Size && s == strings.ToLower(s)
}

func canonicalJSON(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(b, norm.NFC.Bytes(b)) {
		return nil, fmt.Errorf("canonical strings must be UTF-8 NFC")
	}
	return b, nil
}

func (j JobBundle) identity() identity {
	in := append([]string(nil), j.InputHashes...)
	deps := append([]string(nil), j.DependencyHashes...)
	on := append([]string(nil), j.DependsOn...)
	sort.Strings(in)
	sort.Strings(deps)
	sort.Strings(on)
	return identity{j.Experiment, j.ProtocolVersion, j.CorpusID, j.Model,
		j.Candidate, j.Scale, j.Replicate, j.Seed, j.Stage, in, deps,
		j.CodeHash, j.ConfigHash, j.OutputSchema, on, j.Work}
}

func (j JobBundle) ComputedID() (string, error) {
	b, err := canonicalJSON(j.identity())
	if err != nil {
		return "", err
	}
	return HashBytes(b), nil
}

func sortedUnique(xs []string) bool {
	for i, x := range xs {
		if i > 0 && xs[i-1] >= x {
			return false
		}
	}
	return true
}

func (j JobBundle) Validate() error {
	if j.Experiment != Experiment || j.ProtocolVersion != ProtocolVersion {
		return fmt.Errorf("unsupported experiment/protocol")
	}
	if !validStages[j.Stage] {
		return fmt.Errorf("invalid stage %q", j.Stage)
	}
	if j.CorpusID == "" || j.Model == "" || j.Candidate == "" || j.Scale == "" || j.OutputSchema == "" {
		return fmt.Errorf("empty required scientific identity field")
	}
	if j.Replicate < 0 {
		return fmt.Errorf("negative replicate")
	}
	if !ValidHash(j.CodeHash) || !ValidHash(j.ConfigHash) {
		return fmt.Errorf("invalid code/config hash")
	}
	if !sortedUnique(j.InputHashes) || !sortedUnique(j.DependencyHashes) || !sortedUnique(j.DependsOn) {
		return fmt.Errorf("hash/dependency lists must be sorted and unique")
	}
	for _, h := range append(append([]string{}, j.InputHashes...), j.DependencyHashes...) {
		if !ValidHash(h) {
			return fmt.Errorf("invalid input/dependency hash %q", h)
		}
	}
	if j.Work.Kind == "" {
		return fmt.Errorf("missing work descriptor")
	}
	want, err := j.ComputedID()
	if err != nil {
		return err
	}
	if j.JobID != want {
		return fmt.Errorf("job_id mismatch: got %s want %s", j.JobID, want)
	}
	return nil
}

func (m *Manifest) Compile() error {
	if m.SchemaVersion != SchemaVersion || m.CanonicalVersion != CanonicalVersion {
		return fmt.Errorf("unsupported manifest schema/canonical version")
	}
	// IDs are computed in topological manifest order. The order is not an
	// execution order; it merely permits symbolic @<index> fixture edges.
	for i := range m.Jobs {
		j := &m.Jobs[i]
		j.InputHashes = sortedCopy(j.InputHashes)
		j.DependencyHashes = sortedCopy(j.DependencyHashes)
		j.DependsOn = sortedCopy(j.DependsOn)
		id, err := j.ComputedID()
		if err != nil {
			return err
		}
		j.JobID = id
	}
	return m.Validate()
}

func sortedCopy(xs []string) []string {
	out := append([]string(nil), xs...)
	sort.Strings(out)
	return out
}

func (m Manifest) Validate() error {
	seen := map[string]JobBundle{}
	for _, j := range m.Jobs {
		if err := j.Validate(); err != nil {
			return fmt.Errorf("job %s: %w", j.JobID, err)
		}
		if _, ok := seen[j.JobID]; ok {
			return fmt.Errorf("duplicate job_id %s", j.JobID)
		}
		seen[j.JobID] = j
	}
	for _, j := range m.Jobs {
		for _, d := range j.DependsOn {
			if _, ok := seen[d]; !ok {
				return fmt.Errorf("job %s: unknown dependency %s", j.JobID, d)
			}
		}
	}
	visiting, done := map[string]bool{}, map[string]bool{}
	var visit func(string) error
	visit = func(id string) error {
		if visiting[id] {
			return fmt.Errorf("DAG cycle at %s", id)
		}
		if done[id] {
			return nil
		}
		visiting[id] = true
		for _, d := range seen[id].DependsOn {
			if err := visit(d); err != nil {
				return err
			}
		}
		visiting[id] = false
		done[id] = true
		return nil
	}
	for id := range seen {
		if err := visit(id); err != nil {
			return err
		}
	}
	return nil
}
