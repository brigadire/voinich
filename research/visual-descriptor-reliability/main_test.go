package main

import "testing"

func TestPerfectAgreement(t *testing.T) {
	d := descriptor{ID: "D", Type: "ordinal", Allowed: []string{"LOW", "MEDIUM", "HIGH", "UNCERTAIN"}}
	r := compute(d, [][2]string{{"LOW", "LOW"}, {"MEDIUM", "MEDIUM"}, {"HIGH", "HIGH"}})
	if r.Agreement != 1 || r.Decision != "RETAIN" {
		t.Fatalf("%+v", r)
	}
}

func TestMissingExcluded(t *testing.T) {
	if !missing("NOT_VISIBLE") || missing("PRESENT") {
		t.Fatal("missing policy")
	}
}
