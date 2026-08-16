package pki

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
)

func loadCACertPool(caFile string) (*x509.CertPool, error) {
	pemBytes, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(pemBytes) {
		return nil, fmt.Errorf("%s: no certificates found (CA rotation may list more than one)", caFile)
	}
	return pool, nil
}

// CoordinatorServerTLSConfig builds the coordinator's listening tls.Config:
// it presents certFile/keyFile (coordinator.crt/.key, serverAuth), trusts
// only the project CA(s) in clientCAFile for client certificates, and
// requires and verifies every connecting worker's certificate (phase 5).
// deny may be nil, meaning nothing is revoked. There is no
// InsecureSkipVerify, no optional client-auth mode and no fallback path:
// a worker without a valid, CA-signed, non-revoked client certificate never
// completes the handshake (phase 5/6: "fail closed").
func CoordinatorServerTLSConfig(certFile, keyFile, clientCAFile string, deny *DenyList) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load coordinator certificate: %w", err)
	}
	pool, err := loadCACertPool(clientCAFile)
	if err != nil {
		return nil, fmt.Errorf("load client CA pool: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
		VerifyPeerCertificate: func(_ [][]byte, verifiedChains [][]*x509.Certificate) error {
			if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
				return fmt.Errorf("no verified client certificate chain")
			}
			leaf := verifiedChains[0][0]
			workerID, err := WorkerIdentity(leaf)
			if err != nil {
				return err
			}
			if deny.Revoked(SerialHex(leaf), workerID) {
				return fmt.Errorf("worker %q certificate (serial %s) is revoked", workerID, SerialHex(leaf))
			}
			return nil
		},
	}, nil
}

// WorkerClientTLSConfig builds a worker's dial-out tls.Config: it presents
// certFile/keyFile (worker-<id>.crt/.key, clientAuth) and trusts only the
// project CA(s) in caFile to verify the coordinator's chain and server name
// (phase 6). Go's TLS client already refuses an invalid chain or a
// certificate that does not match the dialed host by default; this
// function never sets InsecureSkipVerify and never overrides that
// verification, so there is no production insecure-skip-verify path.
func WorkerClientTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("load worker certificate: %w", err)
	}
	pool, err := loadCACertPool(caFile)
	if err != nil {
		return nil, fmt.Errorf("load root CA pool: %w", err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      pool,
		MinVersion:   tls.VersionTLS12,
	}, nil
}

// RequestWorkerID extracts the authenticated WorkerID from an already
// TLS-verified connection's peer certificate. It is the only place the
// coordinator ever learns a WorkerID: never from a JSON/request field
// (phase 5: "derive WorkerID from the verified peer certificate").
func RequestWorkerID(verifiedChains [][]*x509.Certificate) (string, error) {
	if len(verifiedChains) == 0 || len(verifiedChains[0]) == 0 {
		return "", fmt.Errorf("request has no verified client certificate")
	}
	return WorkerIdentity(verifiedChains[0][0])
}
