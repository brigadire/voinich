// Package speculumf01 implements a source-disciplined reconstruction of
// Fontana's F01 Speculum: a stack of independently rotatable concentric
// rings, each bearing a full alphabet in equal sectors, used to fix a word
// by rotating each ring until the intended letter lies on one marked
// radius (NAL 635, 19v-21r; see research/phase2/fontana/machines/F01_SPECULUM.md).
//
// Evidence labels follow task76's discipline: E = explicitly stated/drawn,
// I = strongly inferred, H = hypothetical completion, U = unknown. They are
// recorded in research/phase2/fontana/f01_speculum/EIHU_TABLE.tsv, not in
// code comments; comments here only note *which* modeling choice a given
// field encodes, not its evidentiary weight.
package speculumf01

import "fmt"

// Direction is the physical reading order of the ring stack.
type Direction int

const (
	InnerToOuter Direction = iota // ring index 0 (innermost) read first
	OuterToInner                  // ring index 0 (innermost) read last
)

// Alphabet is the fixed circular letter sequence carried by every ring.
// index2letter[i] is the letter in sector i; letter2index is its inverse.
type Alphabet struct {
	Name    string
	Letters []rune
}

func NewAlphabet(name, letters string) Alphabet {
	return Alphabet{Name: name, Letters: []rune(letters)}
}

func (a Alphabet) Size() int { return len(a.Letters) }

func (a Alphabet) IndexOf(r rune) (int, bool) {
	for i, l := range a.Letters {
		if l == r {
			return i, true
		}
	}
	return 0, false
}

func (a Alphabet) At(i int) rune {
	m := len(a.Letters)
	return a.Letters[((i%m)+m)%m]
}

// Latin23 is the classical 23-letter Latin alphabet (no J, U, W), chosen
// by analogy with the 23-symbol monoalphabetic cipher script Fontana used
// elsewhere in the same manuscript (TASK74_REPORT.md, section 2). This is
// an [I] assumption for the F01 rings specifically: the source fragment
// for F01 says only "polnyj alfavit", not its size.
var Latin23 = NewAlphabet("latin23", "ABCDEFGHIKLMNOPQRSTVXYZ")

// Modern26 is the plain modern 26-letter alphabet, kept as an explicit
// [H] alternative reconstruction profile to test whether findings depend
// on the 23-vs-26 letter assumption.
var Modern26 = NewAlphabet("modern26", "ABCDEFGHIJKLMNOPQRSTUVWXYZ")

// Config is the full set of construction + convention facts that Speculum
// use presupposes. Everything here is either read off the source (E),
// inferred as a necessary condition for the mechanism to function (I), or
// a labeled reconstruction choice (H) — see EIHU_TABLE.tsv for the per-field
// classification.
type Config struct {
	NumRings   int       // ring stack depth = message capacity in letters [H: 12, capacity unspecified in source]
	Alphabet   Alphabet  // per-ring letter sequence [I: full alphabet per ring; size is H]
	ReadRadius int       // sector index of the marked reading line [I: a fixed mark is required for a repeatable procedure]
	Order      Direction // which physical end holds the first letter [U in source: "the order of rings in this fragment is not specified"]
	// RingIdentityMarked: whether physical rings carry independent,
	// legible identity marks (e.g. engraved numerals) distinct from their
	// concentric position. [H] This is NOT attested; it is a reconstruction
	// fork used to test how state-corruption failure mode depends on it
	// (see Block 5 / STATE_CORRUPTION_RESULTS.tsv).
	RingIdentityMarked bool
}

// State is the complete, directly observable configuration of the device:
// the rotational offset of every ring, with nothing else attached. Offset
// -1 marks a ring whose reading is unavailable (physically destroyed,
// removed, or illegible) — used only by corruption experiments, never by
// Encode.
type State struct {
	Offsets []int
}

func (s State) Clone() State {
	out := make([]int, len(s.Offsets))
	copy(out, s.Offsets)
	return State{Offsets: out}
}

const Missing = -1

// LetterAtRadius returns the letter a ring with the given offset shows at
// an arbitrary sector `radius`. The device only ever exposes this
// operation; it never exposes the offset number itself to a user (offsets
// are a bookkeeping convenience of the digital model, not a hidden
// convenience handed to the decoder — see task76 Block 2's ban on smuggled
// metadata).
func (c Config) LetterAtRadius(offset, radius int) rune {
	return c.Alphabet.At(offset + radius)
}

func (c Config) LetterAtMark(offset int) rune {
	return c.LetterAtRadius(offset, c.ReadRadius)
}

// RingPos maps a 0-based position within the message to a physical ring
// index, honoring Order.
func (c Config) RingPos(i int) int {
	if c.Order == InnerToOuter {
		return i
	}
	return c.NumRings - 1 - i
}

// Encode realizes E_K(M) = S: for each letter of message, rotate the
// corresponding ring so that letter sits on the marked radius. Rings past
// the message length are left at a filler offset drawn from fillerRNG,
// modeling "whatever was left from a previous use" rather than a
// convenient blank/zero state the device does not actually provide.
func (c Config) Encode(message string, fillerRNG func() int) (State, error) {
	letters := []rune(message)
	if len(letters) == 0 {
		return State{}, fmt.Errorf("empty message")
	}
	if len(letters) > c.NumRings {
		return State{}, fmt.Errorf("message length %d exceeds ring capacity %d", len(letters), c.NumRings)
	}
	offsets := make([]int, c.NumRings)
	used := make([]bool, c.NumRings)
	for i, l := range letters {
		idx, ok := c.Alphabet.IndexOf(l)
		if !ok {
			return State{}, fmt.Errorf("letter %q not in alphabet %s", l, c.Alphabet.Name)
		}
		pos := c.RingPos(i)
		offsets[pos] = ((idx-c.ReadRadius)%c.Alphabet.Size() + c.Alphabet.Size()) % c.Alphabet.Size()
		used[pos] = true
	}
	for pos := range offsets {
		if !used[pos] {
			offsets[pos] = ((fillerRNG() % c.Alphabet.Size()) + c.Alphabet.Size()) % c.Alphabet.Size()
		}
	}
	return State{Offsets: offsets}, nil
}

// DecodeFull realizes D_K(S) = M-hat under full knowledge: the read
// radius, ring order, ring identity and message length are all known and
// the state is intact. This is the only decoder that is allowed to assume
// away ambiguity; every other decode path lives in ablation.go /
// corruption.go and must return the full compatible set, not a lucky
// guess.
func (c Config) DecodeFull(state State, length int) (string, error) {
	if length <= 0 || length > c.NumRings {
		return "", fmt.Errorf("invalid length %d for %d rings", length, c.NumRings)
	}
	out := make([]rune, length)
	for i := 0; i < length; i++ {
		pos := c.RingPos(i)
		if pos < 0 || pos >= len(state.Offsets) || state.Offsets[pos] == Missing {
			return "", fmt.Errorf("ring at word position %d (physical ring %d) is missing", i, pos)
		}
		out[i] = c.LetterAtMark(state.Offsets[pos])
	}
	return string(out), nil
}
