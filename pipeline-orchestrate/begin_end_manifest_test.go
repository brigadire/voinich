package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestBeginEndAnalyzeManifestIncludesRemoteExecutorArgs is Task47 section
// 10's explicit requirement: Executor=true on the stage table entry alone
// is not sufficient evidence that begin-end-analyze actually receives
// remote/mTLS configuration - the generated manifest's stage 5 command
// line must really carry -executor remote and every mTLS flag, alongside
// (never instead of) its own -dictionary/-output-dir arguments.
func TestBeginEndAnalyzeManifestIncludesRemoteExecutorArgs(t *testing.T) {
	repo, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	st, ok := stageByName("begin-end-analyze")
	if !ok {
		t.Fatal("begin-end-analyze stage missing from stages table")
	}
	if !st.Executor {
		t.Fatal("begin-end-analyze stage must have Executor: true (Task47)")
	}

	opt := orchestratorOptions{
		Executor: "remote", LocalWorkers: 4,
		RemoteListen: "127.0.0.1:9443", TLSCert: "coord.crt", TLSKey: "coord.key", ClientCA: "ca.crt",
		RemoteTimeout: "5m", RemoteRetries: 3,
	}
	m, err := buildManifest(repo, "generic", "definitely/missing/ZL3b-n.txt", "data_test/pg2097-2.txt", opt, []string{"worker-a", "worker-b"})
	if err != nil {
		t.Fatal(err)
	}

	var sm *StageManifest
	for i := range m.Stages {
		if m.Stages[i].Name == "begin-end-analyze" {
			sm = &m.Stages[i]
		}
	}
	if sm == nil {
		t.Fatal("manifest has no begin-end-analyze stage entry")
	}
	if sm.Status != "PLANNED" {
		t.Fatalf("begin-end-analyze status = %q, want PLANNED", sm.Status)
	}
	joined := " " + strings.Join(sm.Args, " ") + " "
	for _, want := range []string{
		" -executor remote ", " -remote-listen 127.0.0.1:9443 ",
		" -tls-cert coord.crt ", " -tls-key coord.key ", " -client-ca ca.crt ",
		" -remote-timeout 5m ", " -remote-retries 3 ",
		" -dictionary workdir/dataset/dictionary.yaml ", " -output-dir workdir ",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("begin-end-analyze args missing %q\nargs=%v", strings.TrimSpace(want), sm.Args)
		}
	}
}
