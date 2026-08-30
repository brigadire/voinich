package notation

import (
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// variableLinedCorpus builds a corpus with a DIFFERENT token/symbol count
// per physical line (unlike syntheticLinedCorpus's uniform lines), which is
// what actually exposes float64-summation-order nondeterminism: summing N
// identical values never depends on order, but summing N different values
// does (IEEE 754 addition is not associative).
func variableLinedCorpus(nLines int) []Record {
	var out []Record
	alphabet := []string{"a", "b", "c", "d", "e", "f", "g"}
	id := 0
	for li := range nLines {
		tokensThisLine := 1 + (li*37+11)%9 // varies 1..9 across lines
		for ti := range tokensThisLine {
			nSyms := 1 + (li*13+ti*5)%4
			syms := make([]string, nSyms)
			var tokB strings.Builder
			for j := range syms {
				syms[j] = alphabet[(li*7+ti*3+j)%len(alphabet)]
				tokB.WriteString(syms[j])
			}
			tok := tokB.String()
			out = append(out, Record{
				SchemaVersion: SchemaVersion, CorpusID: "SYN-VARLINE", Representation: "SYN-R1",
				Document: ObservedLevel{Value: "doc1", Observed: true}, PhysicalLine: ObservedLevel{Value: strconv.Itoa(li), Observed: true},
				TokenID: "SYN-" + strconv.Itoa(id), TokenIndex: ti, Token: tok, Symbols: syms,
			})
			id++
		}
	}
	return out
}

// TestLineMetricsDeterministicAcrossRuns is a regression test for a real
// bug found via task run03's production-run-execute reproducibility check
// on real C06 (MUSIC-R1/R2/R3) data: lineMetrics built its per-line
// token/symbol-count slices by ranging directly over an unsorted
// map[string]int, so mean()/stddev() (L01/L02/L05) summed float64 values
// in Go's per-process-randomized map iteration order — non-associative
// float addition then made repeated Analyze() calls on the identical input
// disagree in the last few significant digits. The fix sorts line keys
// before accumulating.
func TestLineMetricsDeterministicAcrossRuns(t *testing.T) {
	rs := variableLinedCorpus(400)
	var first Fingerprint
	for i := range 20 {
		fp, err := Analyze(rs)
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			first = fp
			continue
		}
		if !reflect.DeepEqual(first.Metrics, fp.Metrics) {
			t.Fatalf("Analyze() produced different metrics on repeated calls over identical input (run %d):\nfirst=%+v\nlater=%+v", i, first.Metrics, fp.Metrics)
		}
	}
}
