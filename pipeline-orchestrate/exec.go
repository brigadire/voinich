package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"
)

// runResult captures wall-time, exit status, and resource usage for one
// subprocess - Task36's "used compute resources" and "wall-time of every
// stage" requirements.
type runResult struct {
	Duration       time.Duration
	ExitCode       int
	UserCPUSeconds float64
	SysCPUSeconds  float64
	MaxRSSKB       int64
	Err            error
}

// runLogged runs name with args, streaming combined stdout/stderr to a
// fresh file at logPath, and reports timing/exit/resource usage. It never
// uses a shell: args are passed directly to the binary, so nothing here is
// vulnerable to shell-metacharacter injection from any manifest value.
func runLogged(dir, name string, args []string, logPath string) runResult {
	start := time.Now()
	logFile, err := os.Create(logPath)
	if err != nil {
		return runResult{Err: fmt.Errorf("create log %s: %w", logPath, err)}
	}
	defer logFile.Close()

	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.Stdin = nil

	err = cmd.Start()
	if err == nil {
		done := make(chan struct{})
		go func() {
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					fmt.Printf("stage %s still running | elapsed %s | log: %s\n", filepath.Base(name), time.Since(start).Round(time.Second), logPath)
				case <-done:
					return
				}
			}
		}()
		err = cmd.Wait()
		close(done)
	}
	res := runResult{Duration: time.Since(start)}
	if cmd.ProcessState != nil {
		res.ExitCode = cmd.ProcessState.ExitCode()
		if usage, ok := cmd.ProcessState.SysUsage().(*syscall.Rusage); ok {
			res.UserCPUSeconds = time.Duration(usage.Utime.Nano()).Seconds()
			res.SysCPUSeconds = time.Duration(usage.Stime.Nano()).Seconds()
			res.MaxRSSKB = usage.Maxrss
		}
	}
	if err != nil {
		res.Err = err
	}
	return res
}

// buildBinary runs `go build -o <dest> ./<sourceDir>` from repoPath,
// producing the exact binary every stage and the manifest reference.
func buildBinary(repoPath, sourceDir, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}
	pattern := "./" + sourceDir
	if sourceDir == "." {
		pattern = "."
	}
	cmd := exec.Command("go", "build", "-buildvcs=false", "-o", dest, pattern)
	cmd.Dir = repoPath
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("go build ./%s: %w\n%s", sourceDir, err, out)
	}
	return nil
}
