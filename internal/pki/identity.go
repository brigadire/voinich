package pki

import (
	"crypto/x509"
	"fmt"
)

// WorkerIdentity extracts and validates the authenticated worker identity
// carried by cert's WorkerURIScheme URI SAN (phase 4/5: "Worker identity
// comes from the verified certificate, not an untrusted request field").
// It requires exactly one such URI and a clientAuth EKU, so a certificate
// that is ambiguous, foreign, or issued for a different purpose is rejected
// with an explicit reason rather than silently picking an identity.
func WorkerIdentity(cert *x509.Certificate) (string, error) {
	if cert == nil {
		return "", fmt.Errorf("no certificate presented")
	}
	var found []string
	for _, u := range cert.URIs {
		if u.Scheme == WorkerURIScheme {
			found = append(found, u.Opaque)
		}
	}
	switch len(found) {
	case 0:
		return "", fmt.Errorf("certificate carries no %s:// URI SAN", WorkerURIScheme)
	case 1:
		// fine
	default:
		return "", fmt.Errorf("certificate carries %d %s:// URI SANs, want exactly 1", len(found), WorkerURIScheme)
	}
	id := found[0]
	if err := ValidateWorkerID(id); err != nil {
		return "", fmt.Errorf("certificate worker identity: %w", err)
	}
	hasClientAuth := false
	for _, eku := range cert.ExtKeyUsage {
		if eku == x509.ExtKeyUsageClientAuth {
			hasClientAuth = true
		}
	}
	if !hasClientAuth {
		return "", fmt.Errorf("certificate for worker %q lacks the clientAuth extended key usage", id)
	}
	return id, nil
}

// SerialHex returns the lowercase hex form of a certificate's serial
// number, the canonical key used by the revocation deny-list.
func SerialHex(cert *x509.Certificate) string {
	if cert == nil || cert.SerialNumber == nil {
		return ""
	}
	return cert.SerialNumber.Text(16)
}
