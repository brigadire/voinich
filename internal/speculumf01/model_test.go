package speculumf01

import "testing"

func testConfig() Config {
	return Config{NumRings: 10, Alphabet: Latin23, ReadRadius: 5, Order: InnerToOuter, RingIdentityMarked: true}
}

func fixedFiller() func() int {
	i := 0
	return func() int { i++; return i * 7 }
}

func TestBaselineRoundTrip(t *testing.T) {
	c := testConfig()
	for _, msg := range []string{"PAX", "AMOR", "GLORIA", "SPECVLVM"} {
		s, err := c.Encode(msg, fixedFiller())
		if err != nil {
			t.Fatalf("encode %q: %v", msg, err)
		}
		got, err := c.DecodeFull(s, len([]rune(msg)))
		if err != nil {
			t.Fatalf("decode %q: %v", msg, err)
		}
		if got != msg {
			t.Fatalf("D_K(E_K(%q)) = %q, want %q", msg, got, msg)
		}
	}
}

func TestEncodeDeterministic(t *testing.T) {
	c := testConfig()
	s1, _ := c.Encode("FONTANA", fixedFiller())
	s2, _ := c.Encode("FONTANA", fixedFiller())
	for i := range s1.Offsets {
		if s1.Offsets[i] != s2.Offsets[i] {
			t.Fatalf("encoding not deterministic at ring %d: %d vs %d", i, s1.Offsets[i], s2.Offsets[i])
		}
	}
}

func TestMessageTooLongRejected(t *testing.T) {
	c := testConfig()
	if _, err := c.Encode("ABCDEFGHIKL", fixedFiller()); err == nil {
		t.Fatal("expected capacity error for 11-letter word on a 10-ring device")
	}
}

func TestLetterOutsideAlphabetRejected(t *testing.T) {
	c := testConfig()
	if _, err := c.Encode("JUW", fixedFiller()); err == nil {
		t.Fatal("expected rejection: J, U, W are not in Latin23")
	}
}

func TestDirectionAffectsEncoding(t *testing.T) {
	inner := testConfig()
	outer := testConfig()
	outer.Order = OuterToInner
	s1, _ := inner.Encode("PAX", fixedFiller())
	s2, _ := outer.Encode("PAX", fixedFiller())
	if s1.Offsets[0] != s2.Offsets[len(s2.Offsets)-1] {
		t.Fatal("reversing Order should place the first letter on the opposite physical ring")
	}
}

// K4 (read radius unknown) and K5 (initial orientation unknown) must be
// structurally identical: both collapse the same global additive degree
// of freedom, so their candidate sets must have equal size for the same
// message.
func TestK4K5Equivalence(t *testing.T) {
	c := testConfig()
	msg := "MEMORIA"
	s, _ := c.Encode(msg, fixedFiller())
	r4 := Evaluate(c, msg, s, CondReadRadiusUnknown, nil)
	r5 := Evaluate(c, msg, s, CondOrientUnknown, nil)
	if r4.NCandidatesRaw != r5.NCandidatesRaw {
		t.Fatalf("K4/K5 should be equivalent: got %d vs %d", r4.NCandidatesRaw, r5.NCandidatesRaw)
	}
	if r4.NCandidatesRaw != c.Alphabet.Size() {
		t.Fatalf("unknown read radius should yield exactly alphabet-size candidates, got %d", r4.NCandidatesRaw)
	}
}

func TestFullKnowledgeIsUnambiguous(t *testing.T) {
	c := testConfig()
	msg := "NATVRA"
	s, _ := c.Encode(msg, fixedFiller())
	r := Evaluate(c, msg, s, CondFullKnowledge, nil)
	if r.NCandidatesRaw != 1 || !r.TrueMessageInSet || r.ExactBlindP != 1.0 {
		t.Fatalf("full knowledge should be fully unambiguous, got %+v", r)
	}
}

func TestDirectionUnknownGivesTwoCandidates(t *testing.T) {
	c := testConfig()
	msg := "GLORIA"
	s, _ := c.Encode(msg, fixedFiller())
	r := Evaluate(c, msg, s, CondDirectionUnknown, nil)
	if r.NCandidatesRaw != 2 {
		t.Fatalf("direction ablation should give exactly 2 candidates for a message with no palindromic symmetry, got %d", r.NCandidatesRaw)
	}
	if !r.TrueMessageInSet {
		t.Fatal("true message must be a member of its own compatible set")
	}
}

func TestSinglePositionDamageIsLocal(t *testing.T) {
	c := testConfig()
	msg := "MEMORIA"
	s, _ := c.Encode(msg, fixedFiller())
	damaged := SinglePositionDamage(s, c.RingPos(2), (s.Offsets[c.RingPos(2)]+3)%c.Alphabet.Size())
	decoded := c.DecodeWithGap(damaged, len([]rune(msg)), false)
	met := Measure("single-position", msg, decoded, nil, false)
	if met.ExactRecovery {
		t.Fatal("damaged single ring should not exactly recover")
	}
	if met.ErrorClass != "local" {
		t.Fatalf("single substitution should classify as local, got %s", met.ErrorClass)
	}
	if met.FractionAfterFirstError != 1.0 {
		t.Fatalf("everything after the single error should still be correct, got %f", met.FractionAfterFirstError)
	}
}

// A single ring physically collapsing the stack looks, position-by-
// position, like every later character is wrong -- but it is really one
// indel away from ground truth (a frame-sync/bit-slip failure), not a
// scrambled message. That distinction is exactly what "synchronization"
// vs "cascading" is supposed to capture.
func TestDeleteRingCollapseIsSynchronizationNotCascade(t *testing.T) {
	c := testConfig()
	msg := "FONTANAV" // 8 letters, capacity 10
	s, _ := c.Encode(msg, fixedFiller())
	damaged := DeleteRing(s, c.RingPos(2))
	decoded := c.DecodeWithGap(damaged, len([]rune(msg)), true)
	met := Measure("delete-collapse", msg, decoded, nil, false)
	if met.FractionAfterFirstError >= 0.999 {
		t.Fatalf("collapse should visibly disturb positions after the deletion point, got fraction=%f", met.FractionAfterFirstError)
	}
	if met.ErrorClass != "synchronization" {
		t.Fatalf("ring-collapse deletion should classify as synchronization (low edit distance despite high positional mismatch), got class=%s decoded=%q", met.ErrorClass, decoded)
	}
}

// Several independent, unrelated substitutions scattered through the
// message should NOT collapse to a single-indel realignment: that is the
// genuinely "cascading"/scrambled case.
func TestMultipleIndependentDamagesIsCascading(t *testing.T) {
	c := testConfig()
	c.NumRings = 12
	msg := "CONSTANTINA" // 11 letters
	s, _ := c.Encode(msg, fixedFiller())
	damaged := s
	for _, wordIdx := range []int{1, 4, 7, 9} {
		ring := c.RingPos(wordIdx)
		damaged = SinglePositionDamage(damaged, ring, (damaged.Offsets[ring]+11)%c.Alphabet.Size())
	}
	decoded := c.DecodeWithGap(damaged, len([]rune(msg)), false)
	met := Measure("multi-damage", msg, decoded, nil, false)
	if met.ErrorClass != "cascading" {
		t.Fatalf("four scattered independent substitutions should classify as cascading, got class=%s decoded=%q", met.ErrorClass, decoded)
	}
}

func TestDeleteRingNoCollapseIsLocalGap(t *testing.T) {
	c := testConfig()
	msg := "FONTANAV"
	s, _ := c.Encode(msg, fixedFiller())
	damaged := DeleteRing(s, c.RingPos(2))
	decoded := c.DecodeWithGap(damaged, len([]rune(msg)), false)
	if []rune(decoded)[2] != '?' {
		t.Fatalf("expected a gap marker at the deleted position, got %q", decoded)
	}
	for i, r := range []rune(decoded) {
		if i != 2 && r != []rune(msg)[i] {
			t.Fatalf("no-collapse deletion should leave every other position untouched, got %q", decoded)
		}
	}
}

func TestOrientationMarkLossIsGlobal(t *testing.T) {
	c := testConfig()
	msg := "NATVRA"
	s, _ := c.Encode(msg, fixedFiller())
	damaged := LoseOrientationMark(c, s, 4)
	decoded, err := c.DecodeFull(damaged, len([]rune(msg)))
	if err != nil {
		t.Fatal(err)
	}
	if decoded == msg {
		t.Fatal("a nonzero orientation shift must change every letter")
	}
	met := Measure("orientation-loss", msg, decoded, nil, false)
	if met.ErrorClass != "global" {
		t.Fatalf("a uniform additive shift is one degree of freedom, not scrambled independent errors; want class=global, got %s (decoded=%q)", met.ErrorClass, decoded)
	}
}
