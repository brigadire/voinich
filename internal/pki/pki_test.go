package pki

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func genCA(t *testing.T, dir string) {
	t.Helper()
	if err := GenerateCA(dir, time.Hour, false); err != nil {
		t.Fatalf("GenerateCA: %v", err)
	}
}

func TestGenerateCARefusesOverwriteUnlessForced(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	if err := GenerateCA(dir, time.Hour, false); err == nil {
		t.Fatal("expected refusal to overwrite existing CA")
	}
	if err := GenerateCA(dir, time.Hour, true); err != nil {
		t.Fatalf("forced overwrite should succeed: %v", err)
	}
}

func TestGenerateCAKeyPermissionsAndBasicConstraints(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crtPath, keyPath := CAPaths(dir)
	info, err := os.Stat(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("ca.key permissions = %o, want 0600", perm)
	}
	cert, _, err := LoadCA(crtPath, keyPath)
	if err != nil {
		t.Fatal(err)
	}
	if !cert.IsCA {
		t.Fatal("CA certificate must have IsCA=true")
	}
	if cert.KeyUsage&x509.KeyUsageCertSign == 0 {
		t.Fatal("CA certificate must have KeyUsageCertSign")
	}
	if cert.MaxPathLen != 0 || !cert.MaxPathLenZero {
		t.Fatal("CA certificate must have MaxPathLen=0 (no intermediates)")
	}
}

func TestIssueCoordinatorRequiresExplicitSAN(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, nil, nil, time.Hour, false); err == nil {
		t.Fatal("expected refusal without any DNS/IP SAN")
	}
	if err := IssueCoordinator(crt, key, dir, []string{"coordinator.internal"}, []string{"127.0.0.1"}, time.Hour, false); err != nil {
		t.Fatalf("IssueCoordinator: %v", err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	info, err := os.Stat(coordKey)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Fatalf("coordinator.key permissions = %o, want 0600", perm)
	}
	cert := mustParseCertFile(t, coordCrt)
	if cert.Subject.CommonName == "coordinator.internal" {
		t.Fatal("identity must not rely on Common Name")
	}
	if len(cert.DNSNames) == 0 || cert.DNSNames[0] != "coordinator.internal" {
		t.Fatalf("expected DNS SAN, got %v", cert.DNSNames)
	}
	if len(cert.IPAddresses) == 0 {
		t.Fatal("expected IP SAN")
	}
	hasServerAuth := false
	for _, eku := range cert.ExtKeyUsage {
		if eku == x509.ExtKeyUsageServerAuth {
			hasServerAuth = true
		}
	}
	if !hasServerAuth {
		t.Fatal("coordinator certificate must have EKU serverAuth")
	}
}

func TestIssueWorkerValidatesIdentityAndIsUnique(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueWorker(crt, key, dir, "Not Valid!", time.Hour, false); err == nil {
		t.Fatal("expected malformed worker id to be rejected")
	}
	if err := IssueWorker(crt, key, dir, "worker-a", time.Hour, false); err != nil {
		t.Fatalf("IssueWorker worker-a: %v", err)
	}
	if err := IssueWorker(crt, key, dir, "worker-b", time.Hour, false); err != nil {
		t.Fatalf("IssueWorker worker-b: %v", err)
	}
	aCrt, _ := IssueWorkerPaths(dir, "worker-a")
	bCrt, _ := IssueWorkerPaths(dir, "worker-b")
	certA := mustParseCertFile(t, aCrt)
	certB := mustParseCertFile(t, bCrt)
	if certA.SerialNumber.Cmp(certB.SerialNumber) == 0 {
		t.Fatal("worker certificates must have distinct serial numbers")
	}
	idA, err := WorkerIdentity(certA)
	if err != nil {
		t.Fatal(err)
	}
	idB, err := WorkerIdentity(certB)
	if err != nil {
		t.Fatal(err)
	}
	if idA != "worker-a" || idB != "worker-b" {
		t.Fatalf("identities = %q, %q; want worker-a, worker-b", idA, idB)
	}
}

func TestWorkerIdentityRejectsCertWithoutURISAN(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, []string{"host"}, nil, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, _ := IssueCoordinatorPaths(dir)
	cert := mustParseCertFile(t, coordCrt)
	if _, err := WorkerIdentity(cert); err == nil {
		t.Fatal("expected rejection: coordinator certificate has no worker URI SAN and no clientAuth EKU")
	}
}

func TestDenyListRevokesBySerialOrWorkerID(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deny.json")
	d := &DenyList{Serials: map[string]bool{"ab12": true}, WorkerIDs: map[string]bool{"worker-x": true}}
	if err := SaveDenyList(path, d); err != nil {
		t.Fatal(err)
	}
	loaded, err := LoadDenyList(path)
	if err != nil {
		t.Fatal(err)
	}
	if !loaded.Revoked("AB12", "worker-y") {
		t.Fatal("serial match (case-insensitive) must be revoked")
	}
	if !loaded.Revoked("ff", "worker-x") {
		t.Fatal("worker id match must be revoked")
	}
	if loaded.Revoked("ff", "worker-y") {
		t.Fatal("unrelated serial/worker must not be revoked")
	}
}

func TestLoadDenyListMissingFileIsEmpty(t *testing.T) {
	d, err := LoadDenyList(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if d.Revoked("anything", "anything") {
		t.Fatal("missing deny-list file must revoke nothing")
	}
}

// TestMTLSHandshakeEndToEnd is the closest thing to a live integration test
// in this package: a real TLS listener using CoordinatorServerTLSConfig
// only accepts a connection from a client using WorkerClientTLSConfig with a
// valid worker certificate, and the coordinator recovers the exact WorkerID
// from the verified peer certificate.
func TestMTLSHandshakeEndToEnd(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, []string{"localhost"}, []string{"127.0.0.1"}, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	if err := IssueWorker(crt, key, dir, "worker-1", time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	workerCrt, workerKey := IssueWorkerPaths(dir, "worker-1")

	serverCfg, err := CoordinatorServerTLSConfig(coordCrt, coordKey, crt, nil)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	var gotWorkerID string
	var serverErr error
	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := listener.Accept()
		if err != nil {
			serverErr = err
			return
		}
		defer conn.Close()
		tlsConn := conn.(*tls.Conn)
		if err := tlsConn.Handshake(); err != nil {
			serverErr = err
			return
		}
		gotWorkerID, serverErr = RequestWorkerID(tlsConn.ConnectionState().VerifiedChains)
	}()

	clientCfg, err := WorkerClientTLSConfig(workerCrt, workerKey, crt)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg.ServerName = "localhost"
	conn, err := tls.Dial("tcp", listener.Addr().String(), clientCfg)
	if err != nil {
		t.Fatalf("client dial/handshake: %v", err)
	}
	conn.Close()
	<-done
	if serverErr != nil {
		t.Fatalf("server side: %v", serverErr)
	}
	if gotWorkerID != "worker-1" {
		t.Fatalf("WorkerID = %q, want worker-1", gotWorkerID)
	}
}

// dialExpectRejected dials addr with cfg and asserts the peer ultimately
// refuses the connection. With TLS 1.3, a server that rejects the client's
// certificate sends its fatal alert only after the client has already sent
// its own Finished message, so tls.Dial itself can return successfully; the
// rejection surfaces on the first subsequent read or write instead.
func dialExpectRejected(t *testing.T, addr string, cfg *tls.Config) {
	t.Helper()
	conn, err := tls.Dial("tcp", addr, cfg)
	if err != nil {
		return // rejected during the handshake itself
	}
	defer conn.Close()
	_, werr := conn.Write([]byte("x"))
	buf := make([]byte, 1)
	_, rerr := conn.Read(buf)
	if werr == nil && rerr == nil {
		t.Fatal("expected the connection to be rejected, but it was accepted")
	}
}

func TestMTLSHandshakeRejectsMissingClientCert(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, []string{"localhost"}, nil, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	serverCfg, err := CoordinatorServerTLSConfig(coordCrt, coordKey, crt, nil)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.(*tls.Conn).Handshake()
			conn.Close()
		}
	}()

	pool, err := loadCACertPool(crt)
	if err != nil {
		t.Fatal(err)
	}
	// A client presenting no certificate at all must fail closed.
	dialExpectRejected(t, listener.Addr().String(), &tls.Config{RootCAs: pool, ServerName: "localhost"})
}

func TestMTLSHandshakeRejectsForeignCA(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, []string{"localhost"}, nil, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	serverCfg, err := CoordinatorServerTLSConfig(coordCrt, coordKey, crt, nil)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.(*tls.Conn).Handshake()
			conn.Close()
		}
	}()

	foreignDir := t.TempDir()
	genCA(t, foreignDir)
	foreignCrt, foreignKey := CAPaths(foreignDir)
	if err := IssueWorker(foreignCrt, foreignKey, foreignDir, "worker-1", time.Hour, false); err != nil {
		t.Fatal(err)
	}
	workerCrt, workerKey := IssueWorkerPaths(foreignDir, "worker-1")
	clientCfg, err := WorkerClientTLSConfig(workerCrt, workerKey, crt)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg.ServerName = "localhost"
	dialExpectRejected(t, listener.Addr().String(), clientCfg)
}

func TestMTLSHandshakeRejectsWrongEKU(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, []string{"localhost"}, nil, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	serverCfg, err := CoordinatorServerTLSConfig(coordCrt, coordKey, crt, nil)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.(*tls.Conn).Handshake()
			conn.Close()
		}
	}()

	// Issue a "worker" certificate signed by the real project CA but with
	// serverAuth (not clientAuth) EKU: the TLS layer itself must reject it
	// as a client certificate before this package's own EKU check ever runs.
	caCert, caKey, err := LoadCA(crt, key)
	if err != nil {
		t.Fatal(err)
	}
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	template := &x509.Certificate{
		SerialNumber:          serialFor(t, 999),
		Subject:               pkix.Name{CommonName: "wrong-eku-worker"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		URIs:                  []*url.URL{{Scheme: WorkerURIScheme, Opaque: "worker-1"}},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg := &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: priv}},
		RootCAs:      mustPool(t, crt),
		ServerName:   "localhost",
	}
	dialExpectRejected(t, listener.Addr().String(), clientCfg)
}

func TestMTLSHandshakeRejectsExpiredCert(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, []string{"localhost"}, nil, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	serverCfg, err := CoordinatorServerTLSConfig(coordCrt, coordKey, crt, nil)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.(*tls.Conn).Handshake()
			conn.Close()
		}
	}()

	caCert, caKey, err := LoadCA(crt, key)
	if err != nil {
		t.Fatal(err)
	}
	priv, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	template := &x509.Certificate{
		SerialNumber:          serialFor(t, 998),
		Subject:               pkix.Name{CommonName: "expired-worker"},
		NotBefore:             time.Now().Add(-48 * time.Hour),
		NotAfter:              time.Now().Add(-24 * time.Hour), // expired yesterday
		KeyUsage:              x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
		BasicConstraintsValid: true,
		URIs:                  []*url.URL{{Scheme: WorkerURIScheme, Opaque: "worker-1"}},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, caCert, &priv.PublicKey, caKey)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg := &tls.Config{
		Certificates: []tls.Certificate{{Certificate: [][]byte{der}, PrivateKey: priv}},
		RootCAs:      mustPool(t, crt),
		ServerName:   "localhost",
	}
	dialExpectRejected(t, listener.Addr().String(), clientCfg)
}

func TestMTLSHandshakeRejectsWrongCoordinatorSAN(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	// Coordinator certificate is only valid for "coordinator.internal", not
	// "localhost" - a worker dialing localhost must reject it.
	if err := IssueCoordinator(crt, key, dir, []string{"coordinator.internal"}, nil, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	if err := IssueWorker(crt, key, dir, "worker-1", time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	serverCfg, err := CoordinatorServerTLSConfig(coordCrt, coordKey, crt, nil)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.(*tls.Conn).Handshake()
			conn.Close()
		}
	}()

	workerCrt, workerKey := IssueWorkerPaths(dir, "worker-1")
	clientCfg, err := WorkerClientTLSConfig(workerCrt, workerKey, crt)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg.ServerName = "localhost" // does not match the coordinator's only SAN
	conn, err := tls.Dial("tcp", listener.Addr().String(), clientCfg)
	if err == nil {
		conn.Close()
		t.Fatal("expected handshake failure: coordinator certificate SAN does not match dialed server name")
	}
}

func TestRevokedWorkerRejected(t *testing.T) {
	dir := t.TempDir()
	genCA(t, dir)
	crt, key := CAPaths(dir)
	if err := IssueCoordinator(crt, key, dir, []string{"localhost"}, nil, time.Hour, false); err != nil {
		t.Fatal(err)
	}
	if err := IssueWorker(crt, key, dir, "worker-1", time.Hour, false); err != nil {
		t.Fatal(err)
	}
	coordCrt, coordKey := IssueCoordinatorPaths(dir)
	deny := &DenyList{Serials: map[string]bool{}, WorkerIDs: map[string]bool{"worker-1": true}}
	serverCfg, err := CoordinatorServerTLSConfig(coordCrt, coordKey, crt, deny)
	if err != nil {
		t.Fatal(err)
	}
	listener, err := tls.Listen("tcp", "127.0.0.1:0", serverCfg)
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err == nil {
			_ = conn.(*tls.Conn).Handshake()
			conn.Close()
		}
	}()

	workerCrt, workerKey := IssueWorkerPaths(dir, "worker-1")
	clientCfg, err := WorkerClientTLSConfig(workerCrt, workerKey, crt)
	if err != nil {
		t.Fatal(err)
	}
	clientCfg.ServerName = "localhost"
	dialExpectRejected(t, listener.Addr().String(), clientCfg)
}

func serialFor(t *testing.T, n int64) *big.Int {
	t.Helper()
	return big.NewInt(n)
}

func mustParseCertFile(t *testing.T, path string) *x509.Certificate {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	block, _ := pem.Decode(b)
	if block == nil {
		t.Fatalf("%s: no PEM block", path)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatal(err)
	}
	return cert
}

func mustPool(t *testing.T, caFile string) *x509.CertPool {
	t.Helper()
	pool, err := loadCACertPool(caFile)
	if err != nil {
		t.Fatal(err)
	}
	return pool
}
