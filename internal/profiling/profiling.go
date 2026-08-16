// Package profiling wires optional CPU, heap, and execution-trace profiling
// into CLI analyzers through the standard -cpuprofile/-memprofile/-trace
// flags. It never touches algorithms, statistics, or output; it only
// captures runtime behavior around an otherwise unchanged run.
package profiling

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

// Config holds the profiling flag values. The zero value disables profiling
// entirely.
type Config struct {
	CPUProfile       string
	MemProfile       string
	Trace            string
	MemStatsInterval time.Duration
}

// RegisterFlags registers the standard profiling flags on fs and returns the
// Config they populate. Call flag.Parse (or fs.Parse) afterward.
func RegisterFlags(fs *flag.FlagSet) *Config {
	c := &Config{}
	fs.StringVar(&c.CPUProfile, "cpuprofile", "", "write CPU profile to file (disabled if empty)")
	fs.StringVar(&c.MemProfile, "memprofile", "", "write heap profile to file after completion (disabled if empty)")
	fs.StringVar(&c.Trace, "trace", "", "write execution trace to file (disabled if empty)")
	fs.DurationVar(&c.MemStatsInterval, "memstats-interval", 0, "periodically log runtime.MemStats (HeapAlloc/HeapInuse/HeapIdle/HeapReleased/NumGC) to stderr at this interval while running (disabled if 0) - a live-heap diagnostic, distinct from the end-of-run -memprofile snapshot")
	return c
}

// Session is an active profiling session started by Start.
type Session struct {
	cfg          *Config
	cpuFile      *os.File
	traceFile    *os.File
	memStatsStop chan struct{}
	memStatsDone chan struct{}
}

// Start begins CPU profiling and/or execution tracing as configured in cfg,
// creating any missing parent directories. If cfg is nil or none of its
// fields are set, Start is a no-op: the returned Session's Stop still writes
// a heap profile if MemProfile is set. Callers must defer Stop after a
// successful Start.
func Start(cfg *Config) (*Session, error) {
	s := &Session{cfg: cfg}
	if cfg == nil {
		return s, nil
	}
	if cfg.CPUProfile != "" {
		f, err := createProfileFile(cfg.CPUProfile)
		if err != nil {
			return nil, fmt.Errorf("cpuprofile: %w", err)
		}
		if err := pprof.StartCPUProfile(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("cpuprofile: start: %w", err)
		}
		s.cpuFile = f
	}
	if cfg.Trace != "" {
		f, err := createProfileFile(cfg.Trace)
		if err != nil {
			return nil, fmt.Errorf("trace: %w", err)
		}
		if err := trace.Start(f); err != nil {
			f.Close()
			return nil, fmt.Errorf("trace: start: %w", err)
		}
		s.traceFile = f
	}
	if cfg.MemStatsInterval > 0 {
		s.memStatsStop = make(chan struct{})
		s.memStatsDone = make(chan struct{})
		go logMemStatsPeriodically(cfg.MemStatsInterval, s.memStatsStop, s.memStatsDone)
	}
	return s, nil
}

// logMemStatsPeriodically writes one runtime.MemStats snapshot to stderr
// every interval until stop is closed, then closes done. It is a pure
// read-only diagnostic (runtime.ReadMemStats does not allocate on the
// analyzed program's behalf beyond its own bookkeeping) - it never touches
// program state, RNG, or output.
func logMemStatsPeriodically(interval time.Duration, stop <-chan struct{}, done chan<- struct{}) {
	defer close(done)
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var m runtime.MemStats
	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			runtime.ReadMemStats(&m)
			fmt.Fprintf(os.Stderr, "memstats: HeapAlloc=%dMB HeapInuse=%dMB HeapIdle=%dMB HeapReleased=%dMB HeapObjects=%d NumGC=%d\n",
				m.HeapAlloc>>20, m.HeapInuse>>20, m.HeapIdle>>20, m.HeapReleased>>20, m.HeapObjects, m.NumGC)
		}
	}
}

// Stop finalizes profiling: it stops any active CPU profile and trace, then
// writes a heap profile (after a forced GC, per pprof convention) if
// MemProfile was set.
func (s *Session) Stop() error {
	if s == nil {
		return nil
	}
	if s.memStatsStop != nil {
		close(s.memStatsStop)
		<-s.memStatsDone
	}
	if s.cpuFile != nil {
		pprof.StopCPUProfile()
		if err := s.cpuFile.Close(); err != nil {
			return fmt.Errorf("cpuprofile: close: %w", err)
		}
	}
	if s.traceFile != nil {
		trace.Stop()
		if err := s.traceFile.Close(); err != nil {
			return fmt.Errorf("trace: close: %w", err)
		}
	}
	if s.cfg != nil && s.cfg.MemProfile != "" {
		f, err := createProfileFile(s.cfg.MemProfile)
		if err != nil {
			return fmt.Errorf("memprofile: %w", err)
		}
		runtime.GC()
		if err := pprof.WriteHeapProfile(f); err != nil {
			f.Close()
			return fmt.Errorf("memprofile: write: %w", err)
		}
		if err := f.Close(); err != nil {
			return fmt.Errorf("memprofile: close: %w", err)
		}
	}
	return nil
}

func createProfileFile(path string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("create parent directory: %w", err)
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("create file: %w", err)
	}
	return f, nil
}

// PrintElapsed writes the total elapsed runtime since start to w.
func PrintElapsed(w io.Writer, start time.Time) {
	fmt.Fprintf(w, "elapsed runtime: %s\n", shortDuration(time.Since(start)))
}

func shortDuration(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	d = d.Round(time.Millisecond)
	h := int(d / time.Hour)
	m := int(d/time.Minute) % 60
	s := d - time.Duration(h)*time.Hour - time.Duration(m)*time.Minute
	if h > 0 {
		return fmt.Sprintf("%dh%02dm%s", h, m, s)
	}
	if m > 0 {
		return fmt.Sprintf("%dm%s", m, s)
	}
	return s.String()
}
