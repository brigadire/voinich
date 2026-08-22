package tokentransition

import "testing"

func TestCanonicalEdits(t *testing.T) {
	for _, x := range []struct {
		a, b []string
		op   string
		pos  int
	}{{[]string{"a", "b"}, []string{"a", "c"}, "SUBSTITUTION", 1}, {[]string{"a", "b"}, []string{"a", "x", "b"}, "INSERTION", 1}, {[]string{"a", "x", "b"}, []string{"a", "b"}, "DELETION", 1}} {
		p := Analyze(x.a, x.b)
		if p.Distance != 1 || p.Operation != x.op || p.Position != x.pos {
			t.Fatalf("%+v want %s/%d", p, x.op, x.pos)
		}
	}
}
