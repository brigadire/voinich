package main

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// tsvWriter accumulates rows and writes them as a tab-separated file.
type tsvWriter struct {
	b strings.Builder
}

func newTSV(header ...string) *tsvWriter {
	w := &tsvWriter{}
	w.row(header...)
	return w
}

func (w *tsvWriter) row(cols ...string) {
	for idx, c := range cols {
		if idx > 0 {
			w.b.WriteByte('\t')
		}
		w.b.WriteString(strings.ReplaceAll(strings.ReplaceAll(c, "\t", " "), "\n", " "))
	}
	w.b.WriteByte('\n')
}

func (w *tsvWriter) write(dir, name string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, name), []byte(w.b.String()), 0o644)
}

func fstr(v float64) string { return strconv.FormatFloat(v, 'g', 6, 64) }
func istr(v int) string     { return strconv.Itoa(v) }
func bstr(v bool) string    { return strconv.FormatBool(v) }
func fOr(v float64, ok bool) string {
	if !ok {
		return ""
	}
	return fstr(v)
}
