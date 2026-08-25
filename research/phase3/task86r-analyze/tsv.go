package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

type TSVWriter struct {
	f      *os.File
	header []string
}

func NewTSVWriter(path string, header []string) (*TSVWriter, error) {
	f, err := os.Create(path)
	if err != nil {
		return nil, err
	}
	if _, err := fmt.Fprintln(f, strings.Join(header, "\t")); err != nil {
		f.Close()
		return nil, err
	}
	return &TSVWriter{f: f, header: header}, nil
}

func (w *TSVWriter) Row(fields ...string) error {
	_, err := fmt.Fprintln(w.f, strings.Join(fields, "\t"))
	return err
}

func (w *TSVWriter) Close() error { return w.f.Close() }

func f64(v float64) string {
	if math.IsNaN(v) {
		return "NaN"
	}
	if math.IsInf(v, 1) {
		return "+Inf"
	}
	if math.IsInf(v, -1) {
		return "-Inf"
	}
	return strconv.FormatFloat(v, 'g', 10, 64)
}

func boolStr(b bool) string {
	if b {
		return "TRUE"
	}
	return "FALSE"
}

func i64(v int) string { return strconv.Itoa(v) }
