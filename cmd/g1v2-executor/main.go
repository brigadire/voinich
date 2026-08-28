// Command g1v2-executor validates and executes frozen G1-v2 scientific
// manifests. It contains no fitting, generation, metric, or gate logic.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"zcore.dev/voinich/internal/g1v2"
)

func fatal(err error) { fmt.Fprintln(os.Stderr, "g1v2-executor:", err); os.Exit(1) }
func readManifest(path string) (g1v2.Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return g1v2.Manifest{}, err
	}
	var m g1v2.Manifest
	d := json.NewDecoder(bytes.NewReader(b))
	d.DisallowUnknownFields()
	if err = d.Decode(&m); err != nil {
		return m, err
	}
	return m, m.Validate()
}
func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0644)
}

func main() {
	if len(os.Args) < 2 {
		fatal(fmt.Errorf("usage: g1v2-executor fixture|validate|run|coordinator|worker"))
	}
	switch os.Args[1] {
	case "fixture":
		fixture(os.Args[2:])
	case "validate":
		validate(os.Args[2:])
	case "run":
		run(os.Args[2:])
	case "coordinator":
		coordinator(os.Args[2:])
	case "worker":
		worker(os.Args[2:])
	default:
		fatal(fmt.Errorf("unknown command %q", os.Args[1]))
	}
}

func coordinator(args []string) {
	fs := flag.NewFlagSet("coordinator", flag.ExitOnError)
	path := fs.String("manifest", "", "")
	store := fs.String("store", "", "")
	listen := fs.String("listen", "", "")
	cert := fs.String("tls-cert", "", "")
	key := fs.String("tls-key", "", "")
	ca := fs.String("client-ca", "", "")
	deny := fs.String("deny-list", "", "")
	timeout := fs.Duration("lease-timeout", 30*time.Second, "")
	fs.Parse(args)
	m, err := readManifest(*path)
	if err != nil {
		fatal(err)
	}
	if len(m.Jobs) == 0 {
		fatal(fmt.Errorf("empty manifest"))
	}
	info := g1v2.LocalCompatibility(m.Jobs[0].CodeHash, 0)
	core, err := g1v2.NewCoordinator(m, g1v2.Store{Root: *store}, info, *timeout)
	if err != nil {
		fatal(err)
	}
	remote, err := g1v2.StartRemoteCoordinator(core, *listen, *cert, *key, *ca, *deny)
	if err != nil {
		fatal(err)
	}
	defer remote.Close(context.Background())
	fmt.Printf("G1V2_COORDINATOR_READY addr=%s jobs=%d\n", remote.Listener.Addr(), len(m.Jobs))
	for {
		done, total, _ := core.Counts()
		if done == total {
			fmt.Printf("G1V2_COORDINATOR_COMPLETE jobs=%d/%d\n", done, total)
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}
func worker(args []string) {
	fs := flag.NewFlagSet("worker", flag.ExitOnError)
	url := fs.String("url", "", "")
	cert := fs.String("tls-cert", "", "")
	key := fs.String("tls-key", "", "")
	ca := fs.String("ca", "", "")
	serverName := fs.String("server-name", "", "")
	code := fs.String("code-hash", "", "")
	host := fs.String("host", "", "")
	free := fs.Int64("free-storage", 0, "")
	fs.Parse(args)
	if !g1v2.ValidHash(*code) {
		fatal(fmt.Errorf("valid code hash required"))
	}
	if *host == "" {
		*host, _ = os.Hostname()
	}
	w, err := g1v2.NewRemoteWorker(*url, *cert, *key, *ca, *serverName, g1v2.LocalCompatibility(*code, *free), *host)
	if err != nil {
		fatal(err)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err = w.Run(ctx)
	if err != nil && ctx.Err() == nil {
		fatal(err)
	}
}
func fixture(args []string) {
	fs := flag.NewFlagSet("fixture", flag.ExitOnError)
	out := fs.String("out", "ENGINEERING_FIXTURE_MANIFEST.json", "")
	per := fs.Int("per-model", 4, "")
	iters := fs.Int("iterations", 100000, "")
	code := fs.String("code-hash", "", "")
	config := fs.String("config-hash", "", "")
	fs.Parse(args)
	if !g1v2.ValidHash(*code) || !g1v2.ValidHash(*config) {
		fatal(fmt.Errorf("valid code/config SHA-256 required"))
	}
	m := g1v2.NewEngineeringManifest(*code, *config, *per, *iters)
	if err := m.Validate(); err != nil {
		fatal(err)
	}
	if err := writeJSON(*out, m); err != nil {
		fatal(err)
	}
}
func validate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	path := fs.String("manifest", "", "")
	fs.Parse(args)
	m, err := readManifest(*path)
	if err != nil {
		fatal(err)
	}
	fmt.Printf("G1V2_MANIFEST_VALID jobs=%d\n", len(m.Jobs))
}
func run(args []string) {
	fs := flag.NewFlagSet("run", flag.ExitOnError)
	path := fs.String("manifest", "", "")
	store := fs.String("store", "", "")
	workers := fs.Int("workers", 1, "")
	fs.Parse(args)
	m, err := readManifest(*path)
	if err != nil {
		fatal(err)
	}
	if *store == "" {
		fatal(fmt.Errorf("store required"))
	}
	code := m.Jobs[0].CodeHash
	info := g1v2.LocalCompatibility(code, 0)
	c, err := g1v2.NewCoordinator(m, g1v2.Store{Root: *store}, info, 30*time.Second)
	if err != nil {
		fatal(err)
	}
	start := time.Now()
	if err = g1v2.RunWorkers(context.Background(), c, *workers, info); err != nil {
		fatal(err)
	}
	done, total, _ := c.Counts()
	fmt.Printf("G1V2_RUN_COMPLETE jobs=%d/%d workers=%d wall_seconds=%.9f\n", done, total, *workers, time.Since(start).Seconds())
}
