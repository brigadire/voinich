package speculumf01

import "math"

// Condition names the nine minimal prior-knowledge ablations from
// task76 Block 4. Each condition removes exactly one component of K
// from an otherwise-full baseline; it does not stack with the others
// (Condition7 is the one deliberate exception, and is a named compound,
// not a generic combinator, per the source-discipline note in model.go).
type Condition string

const (
	CondFullKnowledge     Condition = "K1_full_knowledge"
	CondLengthUnknown     Condition = "K2_length_unknown"
	CondRingOrderUnknown  Condition = "K3_ring_identity_unknown"
	CondReadRadiusUnknown Condition = "K4_read_radius_unknown"
	CondOrientUnknown     Condition = "K5_initial_orientation_unknown"
	CondDirectionUnknown  Condition = "K6_traversal_direction_unknown"
	CondConventionUnknown Condition = "K7_convention_unknown_compound"
	CondNoInstruction     Condition = "K8_state_only_no_instruction"
	CondStateGapWithFullK Condition = "K9_instruction_only_state_gap"
)

// AblationResult is the per-message, per-condition outcome.
type AblationResult struct {
	Condition         Condition
	Message           string
	NCandidatesRaw    int     // combinatorial compatible-set size, no language filter
	NCandidatesLex    int     // subset also present in the small reference word list (== NRaw if enumeration was capped or message is a random control)
	EnumerationCapped bool    // true if NRaw is analytic-only (full string enumeration skipped for tractability)
	ExactBlindP       float64 // 1/NCandidatesRaw: best a decoder can do with no further information
	ExactLexP         float64 // 1/NCandidatesLex: best a decoder can do using the reference word list as C
	EntropyRawBits    float64 // log2(NCandidatesRaw)
	TrueMessageInSet  bool    // sanity check: the encoded message must always be a member of its own compatible set
}

// permutations returns all permutations of [0,n) as index slices. Used
// only for the ring-identity ablation, where n is a message length
// (bounded by NumRings), never the full alphabet.
func permutations(n int) [][]int {
	if n == 0 {
		return [][]int{{}}
	}
	base := make([]int, n)
	for i := range base {
		base[i] = i
	}
	var out [][]int
	var rec func(prefix, rest []int)
	rec = func(prefix, rest []int) {
		if len(rest) == 0 {
			cp := make([]int, len(prefix))
			copy(cp, prefix)
			out = append(out, cp)
			return
		}
		for i := range rest {
			nextRest := make([]int, 0, len(rest)-1)
			nextRest = append(nextRest, rest[:i]...)
			nextRest = append(nextRest, rest[i+1:]...)
			rec(append(prefix, rest[i]), nextRest)
		}
	}
	rec(nil, base)
	return out
}

func usedRingPositions(c Config, length int) []int {
	pos := make([]int, length)
	for i := 0; i < length; i++ {
		pos[i] = c.RingPos(i)
	}
	return pos
}

func readAt(c Config, s State, ringPositions []int, radius int, order []int) string {
	out := make([]rune, len(ringPositions))
	for wordIdx, ringSlot := range order {
		ring := ringPositions[ringSlot]
		out[wordIdx] = c.LetterAtRadius(s.Offsets[ring], radius)
	}
	return string(out)
}

func identityOrder(n int) []int {
	o := make([]int, n)
	for i := range o {
		o[i] = i
	}
	return o
}
func reversedOrder(n int) []int {
	o := make([]int, n)
	for i := range o {
		o[i] = n - 1 - i
	}
	return o
}

const permutationEnumCap = 8 // 8! = 40320, still tractable; beyond this we report analytic N only

// Evaluate runs one ablation condition against a known ground-truth
// message and its intact encoded state, returning the full compatible
// set's size (not a single lucky guess) plus derived probabilities.
// lexicon is the small pre-registered reference word list used to model
// the "natural language predictability" contribution C from Block 7; pass
// nil to skip lexical filtering (appropriate for random-string controls).
func Evaluate(c Config, message string, s State, cond Condition, lexicon map[string]bool) AblationResult {
	L := len([]rune(message))
	res := AblationResult{Condition: cond, Message: message}

	var raw []string
	capped := false

	switch cond {
	case CondFullKnowledge, CondStateGapWithFullK:
		got, err := c.DecodeFull(s, L)
		if err == nil {
			raw = []string{got}
		}

	case CondLengthUnknown:
		for length := 1; length <= c.NumRings; length++ {
			got, err := c.DecodeFull(s, length)
			if err == nil {
				raw = append(raw, got)
			}
		}

	case CondRingOrderUnknown:
		pos := usedRingPositions(c, L)
		if L > permutationEnumCap {
			capped = true
			res.NCandidatesRaw = factorial(L)
		} else {
			for _, perm := range permutations(L) {
				raw = append(raw, readAt(c, s, pos, c.ReadRadius, perm))
			}
		}

	case CondReadRadiusUnknown, CondOrientUnknown:
		// Structurally identical: not knowing the marked reading line and
		// not knowing the ring's pre-rotation reference letter both
		// collapse the same degree of freedom (a global additive shift
		// applied uniformly to every ring's readout). See
		// F01_RECONSTRUCTION_DOSSIER.md sec. "K4/K5 equivalence".
		pos := usedRingPositions(c, L)
		order := identityOrder(L)
		for radius := 0; radius < c.Alphabet.Size(); radius++ {
			raw = append(raw, readAt(c, s, pos, radius, order))
		}

	case CondDirectionUnknown:
		pos := usedRingPositions(c, L)
		raw = append(raw, readAt(c, s, pos, c.ReadRadius, identityOrder(L)))
		raw = append(raw, readAt(c, s, pos, c.ReadRadius, reversedOrder(L)))

	case CondConventionUnknown:
		pos := usedRingPositions(c, L)
		total := c.Alphabet.Size() * 2 * factorial(L)
		if L > 5 { // 23*2*120 = 5520 is the largest we fully enumerate
			capped = true
			res.NCandidatesRaw = total
		} else {
			for _, perm := range permutations(L) {
				for _, order := range [][]int{identityOrder(L), reversedOrder(L)} {
					mapped := make([]int, L)
					for i, p := range perm {
						mapped[i] = order[p]
					}
					for radius := 0; radius < c.Alphabet.Size(); radius++ {
						raw = append(raw, readAt(c, s, pos, radius, mapped))
					}
				}
			}
		}
	}

	res.EnumerationCapped = capped
	if !capped {
		res.NCandidatesRaw = len(dedupe(raw))
		for _, cand := range dedupe(raw) {
			if cand == message {
				res.TrueMessageInSet = true
			}
		}
		res.NCandidatesLex = res.NCandidatesRaw
		if lexicon != nil {
			lexN := 0
			for _, cand := range dedupe(raw) {
				if lexicon[cand] {
					lexN++
				}
			}
			if lexN > 0 {
				res.NCandidatesLex = lexN
			}
		}
	} else {
		res.TrueMessageInSet = true // true by construction: encoding used one of the enumerated axis values
		res.NCandidatesLex = res.NCandidatesRaw
	}

	if res.NCandidatesRaw <= 0 {
		res.NCandidatesRaw = 1
	}
	res.ExactBlindP = 1.0 / float64(res.NCandidatesRaw)
	res.ExactLexP = 1.0 / float64(res.NCandidatesLex)
	res.EntropyRawBits = math.Log2(float64(res.NCandidatesRaw))
	return res
}

func dedupe(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func factorial(n int) int {
	f := 1
	for i := 2; i <= n; i++ {
		f *= i
	}
	return f
}
