package conditionalregime

import (
	"encoding/json"
	"reflect"
	"testing"

	"zcore.dev/voinich/internal/beginendanalyze"
)

// TestBeginEndWireRoundTripPreservesNonUTF8Tokens guards against the exact
// bug Task47's real-corpus scaling study caught: begin-end-analyze's token
// text is verbatim corpus/dictionary content, not guaranteed to be valid
// UTF-8 (both the Astafiev and Voynich corpora contain non-UTF-8 byte
// sequences in some tokens). encoding/json silently replaces invalid UTF-8
// in a plain string with U+FFFD on Marshal - a naive
// json.Marshal(BatchResult)/json.Unmarshal round trip would corrupt
// BeginCandidate/EndCandidate for exactly these tokens, while leaving the
// numeric fields (and therefore ranking/scores) untouched - the kind of
// bug a small-fixture test with ASCII-only tokens can never catch.
func TestBeginEndWireRoundTripPreservesNonUTF8Tokens(t *testing.T) {
	invalidUTF8Begin := string([]byte{0xE3, 0xEE, 0xEB, 0xEE, 0xE2, 0xE0}) // CP1251-style bytes, invalid UTF-8
	invalidUTF8End := string([]byte{0xE8, 0xE7, 0xEC})
	r := beginendanalyze.BatchResult{Candidates: []beginendanalyze.Candidate{
		{BeginCandidate: invalidUTF8Begin, EndCandidate: invalidUTF8End, Score: 0.5},
		{BeginCandidate: "ascii-a", EndCandidate: "ascii-b", Score: 0.25},
	}}

	if !json.Valid(mustMarshal(t, r)) {
		t.Fatal("sanity: naive marshal should still be valid JSON")
	}
	// Prove the naive path actually corrupts (documenting *why* the wire
	// encoding exists, not just that it works).
	var naive beginendanalyze.BatchResult
	if err := json.Unmarshal(mustMarshal(t, r), &naive); err != nil {
		t.Fatal(err)
	}
	if naive.Candidates[0].BeginCandidate == invalidUTF8Begin {
		t.Fatal("expected naive json round trip to corrupt non-UTF-8 token text (test assumption invalid)")
	}

	wire := encodeBeginEndBatchResult(r)
	b, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var decoded wireBeginEndBatchResult
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	got := decoded.decode()
	if !reflect.DeepEqual(got, r) {
		t.Fatalf("wire round trip diverged\ngot=%#v\nwant=%#v", got, r)
	}
}

func mustMarshal(t *testing.T, r beginendanalyze.BatchResult) []byte {
	t.Helper()
	b, err := json.Marshal(r)
	if err != nil {
		t.Fatal(err)
	}
	return b
}
