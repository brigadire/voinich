package pki

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

// DenyList is the minimum-justified revocation mechanism required by phase
// 7: an explicit coordinator-side deny list keyed by certificate serial
// and/or authenticated worker identity. A full CRL/OCSP stack is
// disproportionate for a project PKI this small (phase 7: "if CRL/OCSP is
// disproportionate"), and this file is the whole mechanism: no networked
// revocation check, no extra service, no timing side channel beyond the
// existing TLS handshake.
type DenyList struct {
	Serials   map[string]bool `json:"serials"`
	WorkerIDs map[string]bool `json:"worker_ids"`
}

// denyListFile is the on-disk JSON shape; DenyList's maps are keyed for O(1)
// lookup but marshal/unmarshal as sorted-ish string slices for a readable,
// diffable file.
type denyListFile struct {
	Serials   []string `json:"serials"`
	WorkerIDs []string `json:"worker_ids"`
}

// LoadDenyList reads a deny-list file. An empty path is not an error: it
// returns an empty list that revokes nothing, so operators who never need
// revocation need not create the file.
func LoadDenyList(path string) (*DenyList, error) {
	d := &DenyList{Serials: map[string]bool{}, WorkerIDs: map[string]bool{}}
	if path == "" {
		return d, nil
	}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return d, nil
	}
	if err != nil {
		return nil, err
	}
	var f denyListFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	for _, s := range f.Serials {
		d.Serials[strings.ToLower(strings.TrimSpace(s))] = true
	}
	for _, w := range f.WorkerIDs {
		d.WorkerIDs[strings.TrimSpace(w)] = true
	}
	return d, nil
}

// SaveDenyList writes d back to path in the same sorted, readable shape
// LoadDenyList accepts.
func SaveDenyList(path string, d *DenyList) error {
	f := denyListFile{}
	for s := range d.Serials {
		f.Serials = append(f.Serials, s)
	}
	for w := range d.WorkerIDs {
		f.WorkerIDs = append(f.WorkerIDs, w)
	}
	sort.Strings(f.Serials)
	sort.Strings(f.WorkerIDs)
	b, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0644)
}

// Revoked reports whether the given certificate serial (lowercase hex, see
// SerialHex) or authenticated worker identity has been revoked. Either match
// is sufficient: revoking by serial lets an operator replace a compromised
// worker's credential while keeping its WorkerID in service on a new
// certificate, and revoking by WorkerID retires the identity outright.
func (d *DenyList) Revoked(serialHex, workerID string) bool {
	if d == nil {
		return false
	}
	return d.Serials[strings.ToLower(serialHex)] || d.WorkerIDs[workerID]
}
