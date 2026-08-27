package g1v2

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"time"

	"zcore.dev/voinich/internal/pki"
)

const maxProtocolMessage = 32 << 20

// RemoteCoordinator is the G1-v2 adapter over the existing Phase-I security
// boundary: project CA, certificate-derived worker identity, pull-by-lease,
// expiry, and persistent worker reconnect semantics.
type RemoteCoordinator struct {
	Core     *Coordinator
	Server   *http.Server
	Listener net.Listener
}

func StartRemoteCoordinator(core *Coordinator, listen, cert, key, clientCA, denyList string) (*RemoteCoordinator, error) {
	deny, err := pki.LoadDenyList(denyList)
	if err != nil {
		return nil, err
	}
	cfg, err := pki.CoordinatorServerTLSConfig(cert, key, clientCA, deny)
	if err != nil {
		return nil, err
	}
	ln, err := tls.Listen("tcp", listen, cfg)
	if err != nil {
		return nil, err
	}
	r := &RemoteCoordinator{Core: core, Listener: ln}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/lease", r.lease)
	mux.HandleFunc("/v1/result", r.result)
	r.Server = &http.Server{Handler: mux, ReadHeaderTimeout: 5 * time.Second}
	go r.Server.Serve(ln)
	return r, nil
}

func authenticatedWorker(req *http.Request) (string, error) {
	if req.TLS == nil || len(req.TLS.PeerCertificates) == 0 {
		return "", fmt.Errorf("mTLS worker certificate required")
	}
	return pki.WorkerIdentity(req.TLS.PeerCertificates[0])
}
func decodeProtocol(req *http.Request, v any) error {
	b, err := io.ReadAll(io.LimitReader(req.Body, maxProtocolMessage+1))
	if err != nil {
		return err
	}
	if len(b) > maxProtocolMessage {
		return fmt.Errorf("protocol message too large")
	}
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	return d.Decode(v)
}
func writeProtocol(w http.ResponseWriter, status int, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_, _ = w.Write(b)
}

func (r *RemoteCoordinator) lease(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	id, err := authenticatedWorker(req)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	var info Compatibility
	if err = decodeProtocol(req, &info); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	l, err := r.Core.Claim(id, info, time.Now())
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	writeProtocol(w, 200, struct {
		NoWork bool   `json:"no_work"`
		Lease  *Lease `json:"lease,omitempty"`
	}{l == nil, l})
}

type remoteSubmission struct {
	LeaseID   string           `json:"lease_id"`
	Result    ScientificResult `json:"result"`
	Telemetry Telemetry        `json:"telemetry"`
}

func (r *RemoteCoordinator) result(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.Error(w, "method", 405)
		return
	}
	id, err := authenticatedWorker(req)
	if err != nil {
		http.Error(w, err.Error(), 401)
		return
	}
	var x remoteSubmission
	if err = decodeProtocol(req, &x); err != nil {
		http.Error(w, err.Error(), 400)
		return
	}
	idx, err := r.Core.Submit(id, x.LeaseID, x.Result, x.Telemetry)
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	writeProtocol(w, 200, idx)
}
func (r *RemoteCoordinator) Close(ctx context.Context) error { return r.Server.Shutdown(ctx) }

type RemoteWorker struct {
	URL    string
	Client *http.Client
	Info   Compatibility
	Host   string
}

func NewRemoteWorker(url, cert, key, ca, serverName string, info Compatibility, host string) (*RemoteWorker, error) {
	cfg, err := pki.WorkerClientTLSConfig(cert, key, ca)
	if err != nil {
		return nil, err
	}
	if serverName != "" {
		cfg.ServerName = serverName
	}
	return &RemoteWorker{url, &http.Client{Transport: &http.Transport{TLSClientConfig: cfg}, Timeout: 30 * time.Second}, info, host}, nil
}
func postJSON(client *http.Client, url string, in, out any) error {
	b, _ := json.Marshal(in)
	resp, err := client.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	rb, _ := io.ReadAll(io.LimitReader(resp.Body, maxProtocolMessage+1))
	if resp.StatusCode != 200 {
		return fmt.Errorf("remote %s: %s", resp.Status, string(rb))
	}
	return json.Unmarshal(rb, out)
}
func (w *RemoteWorker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		var x struct {
			NoWork bool   `json:"no_work"`
			Lease  *Lease `json:"lease"`
		}
		if err := postJSON(w.Client, w.URL+"/v1/lease", w.Info, &x); err != nil {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
		if x.NoWork {
			time.Sleep(2 * time.Millisecond)
			continue
		}
		start := time.Now()
		cpuStart := processCPUSeconds()
		r, err := ExecuteEngineering(ctx, x.Lease.Job)
		if err != nil {
			return err
		}
		end := time.Now()
		rb, _ := json.Marshal(r)
		t := Telemetry{Host: w.Host, StartUTC: start.UTC().Format(time.RFC3339Nano), EndUTC: end.UTC().Format(time.RFC3339Nano), WallSeconds: end.Sub(start).Seconds(), CPUSeconds: processCPUSeconds() - cpuStart, PeakRSSBytes: peakRSSBytes(), TransferBytes: int64(len(rb)), InfrastructureStatus: "SUCCESS"}
		var idx IndexRecord
		if err = postJSON(w.Client, w.URL+"/v1/result", remoteSubmission{x.Lease.ID, r, t}, &idx); err != nil {
			// A lost result transfer is infrastructure failure. The lease
			// expires and the identical immutable bundle is claimed again.
			continue
		}
		// The protocol intentionally has no privileged "all done" push. A
		// persistent production worker keeps polling; tests/runners cancel it.
	}
}
