// Command conditional-regime-pki is the Task34 administrative tool for the
// project CA that backs conditional-regime-analyze's mTLS remote executor
// (Task33 + Task34). It only ever runs by hand, offline: nothing else in
// this repository invokes it, and no coordinator/worker runtime path reads
// ca.key. See DISTRIBUTED_EXECUTION_OPERATIONS.md for the full procedure.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"zcore.dev/voinich/internal/pki"
	"zcore.dev/voinich/internal/workdir"
)

// defaultPKIDir follows the repository-wide convention (internal/workdir)
// that every command's generated output lives under ./workdir by default;
// -out-dir/-ca-cert/-ca-key/-deny-list always accept an explicit override,
// which operators should use to keep ca.key on offline storage rather than
// in a working tree at all.
var defaultPKIDir = workdir.Path("pki")

type stringList []string

func (l *stringList) String() string { return strings.Join(*l, ",") }
func (l *stringList) Set(s string) error {
	*l = append(*l, s)
	return nil
}

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		usage()
		return 2
	}
	switch args[0] {
	case "ca":
		return runCA(args[1:])
	case "issue-coordinator":
		return runIssueCoordinator(args[1:])
	case "issue-worker":
		return runIssueWorker(args[1:])
	case "revoke":
		return runRevoke(args[1:])
	case "-h", "-help", "--help", "help":
		usage()
		return 0
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand %q\n\n", args[0])
		usage()
		return 2
	}
}

func usage() {
	fmt.Fprint(os.Stderr, `conditional-regime-pki: administrative CA/certificate tool (Task34)

Usage:
  conditional-regime-pki ca -out-dir pki [-validity 87600h] [-force]
  conditional-regime-pki issue-coordinator -ca-cert pki/ca.crt -ca-key pki/ca.key -out-dir pki -dns HOST [-dns HOST ...] [-ip IP ...] [-validity 8760h] [-force]
  conditional-regime-pki issue-worker -ca-cert pki/ca.crt -ca-key pki/ca.key -out-dir pki -worker-id worker-1 [-validity 2160h] [-force]
  conditional-regime-pki revoke -deny-list pki/deny.json [-serial HEX ...] [-worker-id ID ...] [-remove]

ca.key must never be copied to a coordinator or worker runtime node; keep it
offline once issuance is done. See DISTRIBUTED_EXECUTION_OPERATIONS.md.
`)
}

func runCA(args []string) int {
	fs := flag.NewFlagSet("ca", flag.ContinueOnError)
	outDir := fs.String("out-dir", defaultPKIDir, "directory to write ca.crt and ca.key into")
	validity := fs.Duration("validity", pki.DefaultCAValidity, "CA certificate validity")
	force := fs.Bool("force", false, "overwrite an existing ca.crt/ca.key")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := pki.GenerateCA(*outDir, *validity, *force); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	crt, key := pki.CAPaths(*outDir)
	fmt.Printf("Wrote %s and %s (validity %s).\nBack up %s offline; it must never be copied to a coordinator or worker.\n", crt, key, *validity, key)
	return 0
}

func runIssueCoordinator(args []string) int {
	fs := flag.NewFlagSet("issue-coordinator", flag.ContinueOnError)
	caCert := fs.String("ca-cert", filepath.Join(defaultPKIDir, "ca.crt"), "project CA certificate")
	caKey := fs.String("ca-key", filepath.Join(defaultPKIDir, "ca.key"), "project CA private key (used only by this offline tool)")
	outDir := fs.String("out-dir", defaultPKIDir, "directory to write coordinator.crt and coordinator.key into")
	validity := fs.Duration("validity", pki.DefaultCoordinatorValidity, "coordinator certificate validity")
	force := fs.Bool("force", false, "overwrite an existing coordinator.crt/coordinator.key")
	var dns, ip stringList
	fs.Var(&dns, "dns", "DNS SAN the coordinator will be dialed as (repeatable)")
	fs.Var(&ip, "ip", "IP SAN the coordinator will be dialed as (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if err := pki.IssueCoordinator(*caCert, *caKey, *outDir, dns, ip, *validity, *force); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	crt, key := pki.IssueCoordinatorPaths(*outDir)
	fmt.Printf("Wrote %s and %s (validity %s, DNS SANs %v, IP SANs %v).\n", crt, key, *validity, []string(dns), []string(ip))
	return 0
}

func runIssueWorker(args []string) int {
	fs := flag.NewFlagSet("issue-worker", flag.ContinueOnError)
	caCert := fs.String("ca-cert", filepath.Join(defaultPKIDir, "ca.crt"), "project CA certificate")
	caKey := fs.String("ca-key", filepath.Join(defaultPKIDir, "ca.key"), "project CA private key (used only by this offline tool)")
	outDir := fs.String("out-dir", defaultPKIDir, "directory to write worker-<id>.crt and worker-<id>.key into")
	workerID := fs.String("worker-id", "", "unique worker identity (DNS-label-like, e.g. worker-1)")
	validity := fs.Duration("validity", pki.DefaultWorkerValidity, "worker certificate validity")
	force := fs.Bool("force", false, "overwrite an existing worker-<id>.crt/.key")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *workerID == "" {
		fmt.Fprintln(os.Stderr, "Error: -worker-id is required")
		return 2
	}
	if err := pki.IssueWorker(*caCert, *caKey, *outDir, *workerID, *validity, *force); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	crt, key := pki.IssueWorkerPaths(*outDir, *workerID)
	fmt.Printf("Wrote %s and %s (validity %s, identity %s://%s).\n", crt, key, *validity, pki.WorkerURIScheme, *workerID)
	return 0
}

func runRevoke(args []string) int {
	fs := flag.NewFlagSet("revoke", flag.ContinueOnError)
	path := fs.String("deny-list", filepath.Join(defaultPKIDir, "deny.json"), "deny-list JSON file to update (created if missing)")
	remove := fs.Bool("remove", false, "remove the given serials/worker ids instead of adding them")
	var serials, workerIDs stringList
	fs.Var(&serials, "serial", "certificate serial number in hex (repeatable)")
	fs.Var(&workerIDs, "worker-id", "authenticated worker identity (repeatable)")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *path == "" {
		fmt.Fprintln(os.Stderr, "Error: -deny-list is required")
		return 2
	}
	if len(serials) == 0 && len(workerIDs) == 0 {
		fmt.Fprintln(os.Stderr, "Error: at least one -serial or -worker-id is required")
		return 2
	}
	deny, err := pki.LoadDenyList(*path)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	for _, s := range serials {
		s = strings.ToLower(strings.TrimSpace(s))
		if *remove {
			delete(deny.Serials, s)
		} else {
			deny.Serials[s] = true
		}
	}
	for _, w := range workerIDs {
		w = strings.TrimSpace(w)
		if *remove {
			delete(deny.WorkerIDs, w)
		} else {
			deny.WorkerIDs[w] = true
		}
	}
	if err := pki.SaveDenyList(*path, deny); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		return 1
	}
	verb := "Revoked"
	if *remove {
		verb = "Un-revoked"
	}
	fmt.Printf("%s %d serial(s) and %d worker id(s) in %s.\n", verb, len(serials), len(workerIDs), *path)
	return 0
}
