package fingerprintv2

import (
	"math"
	"strconv"
	"testing"
)

func leafFixtureProfile(folio, section string, tokenCount int) LineProfile {
	return LineProfile{LineID: folio + ".1", Folio: folio, LocusID: folio + ".1", Section: section, TokenCount: tokenCount}
}

// TestPF4LeafPairedNullDetectsRealPairing builds leaves whose recto/verso
// vectors are close within a leaf and far across leaves; the real pairing
// should coherence-beat a random re-pairing.
func TestPF4LeafPairedNullDetectsRealPairing(t *testing.T) {
	var profiles []LineProfile
	for leaf := 0; leaf < 8; leaf++ {
		base := leaf * 40
		profiles = append(profiles, leafFixtureProfile("f"+strconv.Itoa(leaf)+"r", "A", base+10))
		profiles = append(profiles, leafFixtureProfile("f"+strconv.Itoa(leaf)+"v", "A", base+11))
	}
	// One unpaired folio (verso-only).
	profiles = append(profiles, leafFixtureProfile("f99v", "A", 5))

	pairs, unpaired := pf4LeafPairs(profiles)
	if len(pairs) != 8 {
		t.Fatalf("expected 8 leaf pairs, got %d", len(pairs))
	}
	if len(unpaired) != 1 || unpaired[0] != "f99v" {
		t.Fatalf("expected f99v to be the sole unpaired folio, got %v", unpaired)
	}

	result := pf4LeafPairedNull(profiles, 500, 12345)
	if result.PairedLeafCount != 8 {
		t.Fatalf("PairedLeafCount=%d, want 8", result.PairedLeafCount)
	}
	if result.Observed <= result.NullMean {
		t.Fatalf("observed coherence %g should exceed the random-pairing null mean %g for well-separated leaves", result.Observed, result.NullMean)
	}
	if result.Verdict != "SUPPORTED" {
		t.Fatalf("verdict=%s, want SUPPORTED for a strongly leaf-separated fixture", result.Verdict)
	}
}

// TestPF4LeafPairedNullFlatFixtureNotSupported checks that indistinguishable
// leaves (no leaf-specific signal) do not spuriously support PF4.
func TestPF4LeafPairedNullFlatFixtureNotSupported(t *testing.T) {
	var profiles []LineProfile
	for leaf := 0; leaf < 8; leaf++ {
		profiles = append(profiles, leafFixtureProfile("f"+strconv.Itoa(leaf)+"r", "A", 20))
		profiles = append(profiles, leafFixtureProfile("f"+strconv.Itoa(leaf)+"v", "A", 20))
	}
	result := pf4LeafPairedNull(profiles, 500, 12345)
	if result.Verdict == "SUPPORTED" {
		t.Fatalf("flat fixture with no leaf-specific signal should not support PF4, got %+v", result)
	}
}

// TestPF4LeafPairedNullInconclusiveOnSparseData ensures an underpowered
// fixture is reported INCONCLUSIVE rather than a false negative.
func TestPF4LeafPairedNullInconclusiveOnSparseData(t *testing.T) {
	profiles := []LineProfile{
		leafFixtureProfile("f1r", "A", 10),
		leafFixtureProfile("f1v", "A", 11),
	}
	result := pf4LeafPairedNull(profiles, 200, 1)
	if result.Verdict != "INCONCLUSIVE" {
		t.Fatalf("verdict=%s, want INCONCLUSIVE with only one usable pair", result.Verdict)
	}
}

// hierarchyFixture builds nFolios folios split across two sections with a
// large section-level offset (a held-out folio's *section* is still
// represented by sibling folios in training under folio-block CV, so HR5
// can legitimately borrow strength from it), a small folio-level deviation
// within each section, and small within-folio line jitter.
func hierarchyFixture(nFolios, linesPerFolio int) []LineProfile {
	var profiles []LineProfile
	for f := 0; f < nFolios; f++ {
		section := "A"
		sectionBase := 20.0
		if f >= nFolios/2 {
			section = "B"
			sectionBase = 80.0
		}
		folioOffset := float64((f%5)-2) * 2 // small, +/-4 around the section base
		base := sectionBase + folioOffset
		folio := "f" + strconv.Itoa(f) + "r"
		for l := 0; l < linesPerFolio; l++ {
			jitter := float64(l%3) - 1 // -1, 0, 1
			profiles = append(profiles, LineProfile{
				LineID: folio + "." + strconv.Itoa(l), Folio: folio, LocusID: folio + "." + strconv.Itoa(l),
				Section: section, TokenCount: int(base + jitter),
			})
		}
	}
	return profiles
}

// TestHierarchyOutOfSampleHR3CollapsesToFlatUnderFolioBlockCV documents the
// a-priori mathematical property from TASK79C_DESIGN.md section 9: under
// strict folio-block holdout, a held-out folio is never seen in training
// (n_folio_train=0 for every held-out point), so a folio-only hierarchical
// model has zero shrinkage weight everywhere and its prediction/predictive
// variance both collapse exactly to the flat baseline's. This is a
// correctness check on the CV harness (no folio-identity leakage), not a
// test of whether folio-level structure exists in the fixture.
func TestHierarchyOutOfSampleHR3CollapsesToFlatUnderFolioBlockCV(t *testing.T) {
	profiles := hierarchyFixture(30, 12)
	result := hierarchyOutOfSample(profiles, 5, 99, 5)
	if result.TestableFolds == 0 {
		t.Fatalf("expected at least one testable fold, got fold results: %+v", result.FoldResults)
	}
	for _, fr := range result.FoldResults {
		if !fr.Testable {
			continue
		}
		// Not bit-exact: folioWithin+folioBetween and grandVar use slightly
		// different degrees-of-freedom divisors (n-numGroups vs n-1), so a
		// residual of a few 1e-5 NLL units is expected numerical noise, not
		// leakage; leakage would show up as a systematic, much larger,
		// consistently-signed improvement.
		if math.Abs(fr.HR3Delta) > 1e-3 {
			t.Fatalf("fold %d: HR3 should collapse to (approximately) flat under folio-block CV (delta=%g); a large nonzero delta indicates folio-identity leakage in the split", fr.Fold, fr.HR3Delta)
		}
	}
}

// TestHierarchyOutOfSampleHR5BeatsFlatViaSectionPooling checks the
// substantive claim: HR5 can borrow strength from sibling folios in the
// same (observed) section even though the held-out folio itself is never
// observed, and should out-predict flat on a fixture with a real
// section-level effect.
func TestHierarchyOutOfSampleHR5BeatsFlatViaSectionPooling(t *testing.T) {
	profiles := hierarchyFixture(30, 12)
	result := hierarchyOutOfSample(profiles, 5, 99, 5)
	if result.TestableFolds == 0 {
		t.Fatalf("expected at least one testable fold, got fold results: %+v", result.FoldResults)
	}
	if result.MeanHR5Delta >= 0 {
		t.Fatalf("expected HR5 to improve mean held-out NLL over flat via section pooling, got delta=%g", result.MeanHR5Delta)
	}
	if result.Verdict != "SUPPORTED" {
		t.Fatalf("verdict=%s, want SUPPORTED; folds=%+v", result.Verdict, result.FoldResults)
	}
}

func TestHierarchyOutOfSampleFlatFixtureNotSupported(t *testing.T) {
	var profiles []LineProfile
	for f := 0; f < 30; f++ {
		section := "A"
		if f >= 15 {
			section = "B"
		}
		folio := "f" + strconv.Itoa(f) + "r"
		for l := 0; l < 12; l++ {
			jitter := float64(l%3) - 1
			profiles = append(profiles, LineProfile{
				LineID: folio + "." + strconv.Itoa(l), Folio: folio, LocusID: folio + "." + strconv.Itoa(l),
				Section: section, TokenCount: int(20 + jitter),
			})
		}
	}
	result := hierarchyOutOfSample(profiles, 5, 99, 5)
	if result.Verdict == "SUPPORTED" {
		t.Fatalf("flat fixture (no folio-level signal) should not support hierarchy, got %+v", result)
	}
}

func TestFoldFoliosNeverSplitsAFolioAcrossFolds(t *testing.T) {
	profiles := hierarchyFixture(30, 12)
	assign := foldFolios(profiles, 5, 7)
	seen := map[string]map[int]bool{}
	for _, l := range profiles {
		if seen[l.Folio] == nil {
			seen[l.Folio] = map[int]bool{}
		}
		seen[l.Folio][assign[l.Folio]] = true
	}
	for folio, folds := range seen {
		if len(folds) != 1 {
			t.Fatalf("folio %s spans more than one fold: %v", folio, folds)
		}
	}
}
