package fingerprintv2

import (
	"math/rand"
	"sort"
)

// permuteWithinGroups returns a copy of values with entries permuted only
// among indices that share the same group key. A single constant group key
// for every index reduces this to a global (N1) shuffle; line/folio/regime
// group keys realize N2/N4/N5; a position-bucket key realizes N3.
func permuteWithinGroups(values []string, group []string, rng *rand.Rand) []string {
	out := append([]string(nil), values...)
	byGroup := map[string][]int{}
	for i, g := range group {
		byGroup[g] = append(byGroup[g], i)
	}
	for _, key := range orderedKeys(byGroup) {
		idx := byGroup[key]
		perm := rng.Perm(len(idx))
		original := make([]string, len(idx))
		for i, pos := range idx {
			original[i] = values[pos]
		}
		for i, pos := range idx {
			out[pos] = original[perm[i]]
		}
	}
	return out
}

// nmiPermutationTest measures categorical association between x and y with
// normalizedMI, then compares it to a null built by permuting y within the
// supplied group keys (see permuteWithinGroups). It underlies most CS1-CS6
// tests; only the grouping key differs between null models N1-N5.
func nmiPermutationTest(id, model string, x, y, group []string, repetitions int, rng *rand.Rand) NullTest {
	observed := normalizedMI(x, y)
	null := make([]float64, repetitions)
	for r := range null {
		shuffled := permuteWithinGroups(y, group, rng)
		null[r] = normalizedMI(x, shuffled)
	}
	return nullTest(id, model, observed, null)
}

// familyLabelPermutation implements N8: it permutes which member tokens
// carry which family label while holding the multiset of family sizes (and
// therefore each token's own frequency, since tokens keep their own
// identity and simply trade family numbers) exactly fixed.
func familyLabelPermutation(familyOfToken map[string]int, rng *rand.Rand) map[string]int {
	tokens := orderedKeysInt(familyOfToken)
	labels := make([]int, len(tokens))
	for i, t := range tokens {
		labels[i] = familyOfToken[t]
	}
	perm := rng.Perm(len(labels))
	out := map[string]int{}
	for i, t := range tokens {
		out[t] = labels[perm[i]]
	}
	return out
}

func orderedKeysInt(m map[string]int) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

func nullModelRegistry() []NullModelRegistryEntry {
	return []NullModelRegistryEntry{
		{
			ID: "N1", Name: "Global token shuffle",
			Preserves:            "corpus-wide token multiset (frequency), token length distribution",
			Destroys:             "essentially all sequence, line, locus, folio and regime structure",
			TestsHypotheses:      []string{"weak baseline only"},
			RemainingConfounders: "none removed beyond the global marginal; too weak to be a sole control",
			Justification:        "task77 §7 N1: usable only as a weak baseline, never the only control for a cross-scale claim",
		},
		{
			ID: "N2", Name: "Within-line shuffle",
			Preserves:            "each line's token multiset, line length, folio/section composition",
			Destroys:             "within-line order, position-within-line, local adjacency",
			TestsHypotheses:      []string{"CS1 (family x line position)", "CS2 (transformation x local context)", "CS8 (line-position conditioning)"},
			RemainingConfounders: "folio/regime-level composition differences remain (deliberately, since the hypothesis is about within-line position, not between-line composition)",
			Justification:        "isolates a within-line positional/adjacency effect from any line- or folio-level compositional confound, because every permutation is drawn from the same physical line",
		},
		{
			ID: "N3", Name: "Within-position-bucket shuffle",
			Preserves:            "the marginal distribution of token classes at each normalized line-position bucket (initial/interior/final), and the fixed sequence of line lengths",
			Destroys:             "which specific tokens/families occupy a given line's slots and the corpus's true line-composition structure",
			TestsHypotheses:      []string{"CS6 (family composition x line/folio structure)"},
			RemainingConfounders: "position-class base rates are held fixed by construction, so this null targets composition/diversity effects net of the position-class marginal",
			Justification:        "task77 §7 N3: used where the hypothesis concerns composition given position, not the position marginal itself",
		},
		{
			ID: "N4", Name: "Within-folio shuffle",
			Preserves:            "folio vocabulary and frequency composition",
			Destroys:             "line/local organization within the folio",
			TestsHypotheses:      []string{"CS3 (family x locus type)"},
			RemainingConfounders: "regime-level (Currier/section) composition, since folios are mostly regime-homogeneous",
			Justification:        "isolates a within-folio locus-type effect from cross-folio composition differences",
		},
		{
			ID: "N5", Name: "Within-regime shuffle",
			Preserves:            "Currier/section partition sizes and per-regime vocabulary composition",
			Destroys:             "within-regime local/line/folio placement",
			TestsHypotheses:      []string{"CS5 (local context x larger regime), regime-stratum component of CS4"},
			RemainingConfounders: "does not by itself test whether the regime label matters, only whether structure exists within a regime once the regime is fixed",
			Justification:        "task77 §7 N5: the direct within-partition analogue of N2 at the regime scale",
		},
		{
			ID: "N6", Name: "Frequency/length-matched random reassignment",
			Preserves:            "token length and (for the frequency-matched variant) log-frequency bin of both endpoints",
			Destroys:             "which specific same-length/frequency token pair is edit-adjacent",
			TestsHypotheses:      []string{"LP2 C-LEN/C-FREQ (inherited from task75)", "CS7 frequency control"},
			RemainingConfounders: "residual within-bin frequency variation; bins are coarse (log2)",
			Justification:        "directly controls the frequency confound task77 §8 flags for CS7 (\"высокочастотные токены по определению встречаются ближе друг к другу\")",
		},
		{
			ID: "N7", Name: "C-GRAMMAR (structure-preserving / frequency-aware)",
			Preserves:            "token count, exact length distribution, alphabet, positional/endpoint/bigram glyph profiles within configured tolerance",
			Destroys:             "any lexical/paradigm structure beyond bounded token-formation constraints",
			TestsHypotheses:      []string{"EF4 (inherited)", "EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL"},
			RemainingConfounders: "fairness of the C-GRAMMAR construction itself is a modeling choice, documented in FINGERPRINT_V2_CONTROLS.md",
			Justification:        "task77 §7 N7: the only null that asks whether cross-scale/edit-family structure could arise purely from bounded token formation",
		},
		{
			ID: "N8", Name: "Family-label permutation",
			Preserves:            "family size distribution and each token's own frequency (tokens keep their identity; only which family number they carry is permuted)",
			Destroys:             "which specific tokens belong to which family",
			TestsHypotheses:      []string{"incremental value of family identity over family-size structure (CS1/CS4 robustness check)"},
			RemainingConfounders: "graph topology within a family is not re-randomized, only membership labels",
			Justification:        "task77 §7 N8: tests whether specific families carry information beyond size and frequency",
		},
	}
}
