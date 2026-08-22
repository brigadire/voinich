package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// tsvWriter is a minimal, dependency-free TSV writer (tab-separated,
// values written as given - every caller in this package controls its
// own value formatting, so no escaping is needed for these strictly
// numeric/token-id columns).
type tsvWriter struct {
	f *os.File
}

func newTSV(dir, name string, header ...string) *tsvWriter {
	f, err := os.Create(filepath.Join(dir, name))
	if err != nil {
		panic(err) // output directory is created by run() before any writer opens
	}
	w := &tsvWriter{f: f}
	w.row(header...)
	return w
}

func (w *tsvWriter) row(cols ...string) {
	for i, c := range cols {
		if i > 0 {
			fmt.Fprint(w.f, "\t")
		}
		fmt.Fprint(w.f, c)
	}
	fmt.Fprint(w.f, "\n")
}

func (w *tsvWriter) close() { w.f.Close() }

func i(v int) string      { return strconv.Itoa(v) }
func f8(v float64) string { return strconv.FormatFloat(v, 'f', 8, 64) }
func f4(v float64) string { return strconv.FormatFloat(v, 'f', 4, 64) }
