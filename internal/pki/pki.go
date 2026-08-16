// Package pki implements the small project-controlled certificate authority
// used by Task34 to add mutual-TLS authentication to the Task33 remote
// distributed executor. It is deliberately minimal: one offline CA, one
// coordinator (server) identity, and one certificate per worker (client)
// identity. It never touches JobID, RNG, scientific computation, scheduling
// or reduction - it only establishes who is allowed to speak to whom over
// the network and where an authenticated WorkerID comes from.
package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

// WorkerURIScheme is the custom URI SAN scheme used to carry an individual
// worker's identity (Task34 phase 4: "unique worker identity in a
// documented certificate field, preferably URI SAN"). The identity is the
// URI's opaque part, e.g. "voinich-worker:worker-7".
const WorkerURIScheme = "voinich-worker"

// DefaultCAValidity, DefaultCoordinatorValidity and DefaultWorkerValidity are
// the "sane default" validities required by phases 2-4. They are
// deliberately short for issued leaf credentials (coordinator/worker) so
// routine renewal is the normal operational path, and long for the offline
// CA since rotating it is the expensive, deliberate operation (phase 7).
const (
	DefaultCAValidity          = 10 * 365 * 24 * time.Hour
	DefaultCoordinatorValidity = 397 * 24 * time.Hour // <= CA/Browser Forum's own server-cert ceiling
	DefaultWorkerValidity      = 90 * 24 * time.Hour
	// clockSkew backdates NotBefore slightly so a freshly issued certificate
	// is not rejected as "not yet valid" by a peer whose clock lags.
	clockSkew = 5 * time.Minute
)

// workerIDPattern is deliberately DNS-label-like: lowercase alphanumeric and
// hyphen, 1-63 characters, no leading/trailing hyphen. This keeps a worker
// identity safe to use in file names, log lines and the URI SAN's opaque
// part without escaping (phase 4: "validate identity format").
var workerIDPattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)

// ValidateWorkerID reports whether id is an acceptable worker identity.
func ValidateWorkerID(id string) error {
	if !workerIDPattern.MatchString(id) {
		return fmt.Errorf("invalid worker id %q: must match %s", id, workerIDPattern.String())
	}
	return nil
}

// workerURI builds the WorkerURIScheme URI SAN carrying a worker's identity
// in its opaque part, e.g. "voinich-worker:worker-7".
func workerURI(workerID string) *url.URL {
	return &url.URL{Scheme: WorkerURIScheme, Opaque: workerID}
}

// KeyPair is a generated ECDSA P-256 key and, once issued, its PEM-encoded
// certificate. P-256/ECDSA-SHA256 is a modern, broadly-supported Go default
// (crypto/tls and every current Go toolchain support it without extra
// flags), avoiding both RSA's key-size bookkeeping and Ed25519's narrower
// tooling support for CA workflows.
type KeyPair struct {
	Private *ecdsa.PrivateKey
	CertDER []byte
}

func newKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
}

func randomSerial() (*big.Int, error) {
	// 128 bits of randomness, RFC 5280-legal (positive, non-zero).
	max := new(big.Int).Lsh(big.NewInt(1), 128)
	serial, err := rand.Int(rand.Reader, max)
	if err != nil {
		return nil, err
	}
	if serial.Sign() == 0 {
		serial.SetInt64(1)
	}
	return serial, nil
}

func subjectKeyID(pub *ecdsa.PublicKey) []byte {
	der, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		// Only reachable if pub is not a well-formed ECDSA key, which never
		// happens for a key this package just generated.
		panic(err)
	}
	sum := sha1.Sum(der)
	return sum[:]
}

// writeRestrictive writes data to path with the given permissions, refusing
// to overwrite an existing file unless force is set (phase 2: "refuse
// overwrite unless explicitly forced").
func writeRestrictive(path string, data []byte, perm os.FileMode, force bool) error {
	flags := os.O_WRONLY | os.O_CREATE
	if force {
		flags |= os.O_TRUNC
	} else {
		flags |= os.O_EXCL
	}
	f, err := os.OpenFile(path, flags, perm)
	if err != nil {
		if os.IsExist(err) {
			return fmt.Errorf("%s already exists (use -force to overwrite)", path)
		}
		return err
	}
	defer f.Close()
	if err := f.Chmod(perm); err != nil { // OpenFile's perm is masked by umask; enforce it explicitly.
		return err
	}
	_, err = f.Write(data)
	return err
}

func encodeKeyPEM(priv *ecdsa.PrivateKey) ([]byte, error) {
	der, err := x509.MarshalPKCS8PrivateKey(priv)
	if err != nil {
		return nil, err
	}
	return pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: der}), nil
}

func encodeCertPEM(der []byte) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
}

// CAPaths returns the two conventional file paths for a project CA rooted
// at dir.
func CAPaths(dir string) (crt, key string) {
	return filepath.Join(dir, "ca.crt"), filepath.Join(dir, "ca.key")
}

// GenerateCA creates a new self-signed project CA at dir/ca.crt and
// dir/ca.key. It must be run once, offline, by an operator: nothing in this
// package ever calls it automatically, and no coordinator/worker runtime
// path ever reads ca.key (phase 2/6: "ca.key is not required on coordinator
// or workers during normal operation").
func GenerateCA(dir string, validity time.Duration, force bool) error {
	if validity <= 0 {
		validity = DefaultCAValidity
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	crtPath, keyPath := CAPaths(dir)
	priv, err := newKey()
	if err != nil {
		return fmt.Errorf("generate CA key: %w", err)
	}
	serial, err := randomSerial()
	if err != nil {
		return err
	}
	skid := subjectKeyID(&priv.PublicKey)
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "voinich conditional-regime project CA", Organization: []string{"voinich"}},
		NotBefore:    now.Add(-clockSkew),
		NotAfter:     now.Add(validity),
		KeyUsage:     x509.KeyUsageCertSign | x509.KeyUsageCRLSign,
		IsCA:         true,
		BasicConstraintsValid: true,
		MaxPathLen:            0,
		MaxPathLenZero:        true,
		SubjectKeyId:          skid,
		AuthorityKeyId:        skid,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &priv.PublicKey, priv)
	if err != nil {
		return fmt.Errorf("create CA certificate: %w", err)
	}
	keyPEM, err := encodeKeyPEM(priv)
	if err != nil {
		return err
	}
	// Key first: if the key write fails, no half-issued CA cert is left
	// behind referencing a key that was never persisted.
	if err := writeRestrictive(keyPath, keyPEM, 0600, force); err != nil {
		return fmt.Errorf("write %s: %w", keyPath, err)
	}
	if err := writeRestrictive(crtPath, encodeCertPEM(der), 0644, force); err != nil {
		return fmt.Errorf("write %s: %w", crtPath, err)
	}
	return nil
}

// LoadCA reads the CA certificate and private key from the given paths.
// Callers issuing certificates need both; runtime coordinator/worker code
// must never call this with a real ca.key path.
func LoadCA(crtPath, keyPath string) (*x509.Certificate, *ecdsa.PrivateKey, error) {
	certPEM, err := os.ReadFile(crtPath)
	if err != nil {
		return nil, nil, err
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, nil, fmt.Errorf("%s: no PEM certificate block found", crtPath)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", crtPath, err)
	}
	keyPEM, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, nil, err
	}
	kb, _ := pem.Decode(keyPEM)
	if kb == nil {
		return nil, nil, fmt.Errorf("%s: no PEM key block found", keyPath)
	}
	key, err := parseECKey(kb.Bytes)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: %w", keyPath, err)
	}
	if !cert.IsCA {
		return nil, nil, fmt.Errorf("%s is not a CA certificate", crtPath)
	}
	return cert, key, nil
}

func parseECKey(der []byte) (*ecdsa.PrivateKey, error) {
	if key, err := x509.ParsePKCS8PrivateKey(der); err == nil {
		if eckey, ok := key.(*ecdsa.PrivateKey); ok {
			return eckey, nil
		}
		return nil, fmt.Errorf("key is not ECDSA")
	}
	return x509.ParseECPrivateKey(der)
}

// IssueCoordinatorPaths returns the conventional coordinator credential
// paths rooted at dir.
func IssueCoordinatorPaths(dir string) (crt, key string) {
	return filepath.Join(dir, "coordinator.crt"), filepath.Join(dir, "coordinator.key")
}

// IssueCoordinator issues coordinator.crt/coordinator.key signed by the CA
// at caCrtPath/caKeyPath. dnsNames/ips are mandatory explicit SANs (phase 3):
// the coordinator's identity is never inferred from Common Name.
func IssueCoordinator(caCrtPath, caKeyPath, dir string, dnsNames []string, ips []string, validity time.Duration, force bool) error {
	if len(dnsNames) == 0 && len(ips) == 0 {
		return fmt.Errorf("coordinator certificate requires at least one -dns or -ip SAN")
	}
	if validity <= 0 {
		validity = DefaultCoordinatorValidity
	}
	caCert, caKey, err := LoadCA(caCrtPath, caKeyPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	var ipAddrs []net.IP
	for _, raw := range ips {
		ip := net.ParseIP(raw)
		if ip == nil {
			return fmt.Errorf("invalid -ip SAN %q", raw)
		}
		ipAddrs = append(ipAddrs, ip)
	}
	priv, err := newKey()
	if err != nil {
		return fmt.Errorf("generate coordinator key: %w", err)
	}
	serial, err := randomSerial()
	if err != nil {
		return err
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "voinich conditional-regime coordinator"},
		NotBefore:             now.Add(-clockSkew),
		NotAfter:              now.Add(validity),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		DNSNames:              dnsNames,
		IPAddresses:           ipAddrs,
		AuthorityKeyId:        caCert.SubjectKeyId,
		SubjectKeyId:          subjectKeyID(&priv.PublicKey),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create coordinator certificate: %w", err)
	}
	keyPEM, err := encodeKeyPEM(priv)
	if err != nil {
		return err
	}
	crtPath, keyPath := IssueCoordinatorPaths(dir)
	if err := writeRestrictive(keyPath, keyPEM, 0600, force); err != nil {
		return fmt.Errorf("write %s: %w", keyPath, err)
	}
	if err := writeRestrictive(crtPath, encodeCertPEM(der), 0644, force); err != nil {
		return fmt.Errorf("write %s: %w", crtPath, err)
	}
	return nil
}

// IssueWorkerPaths returns the conventional per-worker credential paths
// rooted at dir for the given worker identity.
func IssueWorkerPaths(dir, workerID string) (crt, key string) {
	return filepath.Join(dir, "worker-"+workerID+".crt"), filepath.Join(dir, "worker-"+workerID+".key")
}

// IssueWorker issues a unique worker-<id>.crt/.key signed by the CA at
// caCrtPath/caKeyPath, with the worker's identity carried only in a
// WorkerURIScheme URI SAN (phase 4). No two calls ever share a private key
// or serial number, and no shared "worker" certificate is ever produced by
// this function - each call issues exactly one worker's credential.
func IssueWorker(caCrtPath, caKeyPath, dir, workerID string, validity time.Duration, force bool) error {
	if err := ValidateWorkerID(workerID); err != nil {
		return err
	}
	if validity <= 0 {
		validity = DefaultWorkerValidity
	}
	caCert, caKey, err := LoadCA(caCrtPath, caKeyPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	priv, err := newKey()
	if err != nil {
		return fmt.Errorf("generate worker key: %w", err)
	}
	serial, err := randomSerial()
	if err != nil {
		return err
	}
	now := time.Now()
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "voinich conditional-regime worker " + workerID},
		NotBefore:             now.Add(-clockSkew),
		NotAfter:              now.Add(validity),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		IsCA:                  false,
		URIs:                  []*url.URL{workerURI(workerID)},
		AuthorityKeyId:        caCert.SubjectKeyId,
		SubjectKeyId:          subjectKeyID(&priv.PublicKey),
	}
	der, err := x509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		return fmt.Errorf("create worker certificate: %w", err)
	}
	keyPEM, err := encodeKeyPEM(priv)
	if err != nil {
		return err
	}
	crtPath, keyPath := IssueWorkerPaths(dir, workerID)
	if err := writeRestrictive(keyPath, keyPEM, 0600, force); err != nil {
		return fmt.Errorf("write %s: %w", keyPath, err)
	}
	if err := writeRestrictive(crtPath, encodeCertPEM(der), 0644, force); err != nil {
		return fmt.Errorf("write %s: %w", crtPath, err)
	}
	return nil
}
