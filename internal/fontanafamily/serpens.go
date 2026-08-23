package fontanafamily

import (
	"encoding/json"
	"fmt"
)

// SerpensConfig is the operationally testable core of F08. Capacity is a
// reconstruction parameter; centre-to-edge traversal is source-supported.
// EmptyMarker is serialization only and is never a hidden end-of-message bit.
type SerpensConfig struct {
	Capacity    int
	Alphabet    []rune
	Start       int
	Direction   Direction
	EmptyMarker rune
}

// SerpensState is exactly the externally observable sequence of holes. A nil
// entry is an empty or lost insertion. There is no stored message length.
type SerpensState struct {
	Holes []*rune `json:"holes"`
}

// MarshalJSON writes visible symbols as one-rune strings rather than Go's
// numeric rune representation, keeping trial states independently legible.
func (s SerpensState) MarshalJSON() ([]byte, error) {
	holes := make([]any, len(s.Holes))
	for i, r := range s.Holes {
		if r != nil {
			holes[i] = string(*r)
		}
	}
	return json.Marshal(struct {
		Holes []any `json:"holes"`
	}{Holes: holes})
}

func (c SerpensConfig) Validate() error {
	if c.Capacity <= 0 {
		return fmt.Errorf("capacity must be positive")
	}
	if err := ValidateAlphabet(c.Alphabet); err != nil {
		return err
	}
	if c.Start < 0 || c.Start >= c.Capacity {
		return fmt.Errorf("start %d outside capacity %d", c.Start, c.Capacity)
	}
	if c.Direction != Forward && c.Direction != Reverse {
		return fmt.Errorf("unknown direction %d", c.Direction)
	}
	return nil
}

func runePtr(r rune) *rune { x := r; return &x }

func (s SerpensState) Clone() SerpensState {
	out := make([]*rune, len(s.Holes))
	for i, r := range s.Holes {
		if r != nil {
			out[i] = runePtr(*r)
		}
	}
	return SerpensState{Holes: out}
}

func (c SerpensConfig) physicalIndex(step int) int {
	if c.Direction == Forward {
		return c.Start + step
	}
	return c.Start - step
}

// Encode places a contiguous sequence along the configured spiral path. The
// caller must retain length as K: it is not smuggled into the state.
func (c SerpensConfig) Encode(message string) (SerpensState, error) {
	if err := c.Validate(); err != nil {
		return SerpensState{}, err
	}
	runes := []rune(message)
	if len(runes) == 0 || len(runes) > c.Capacity {
		return SerpensState{}, fmt.Errorf("message length %d outside 1..%d", len(runes), c.Capacity)
	}
	allowed := make(map[rune]bool, len(c.Alphabet))
	for _, r := range c.Alphabet {
		allowed[r] = true
	}
	s := SerpensState{Holes: make([]*rune, c.Capacity)}
	for i, r := range runes {
		if !allowed[r] {
			return SerpensState{}, fmt.Errorf("symbol %q outside alphabet", r)
		}
		physical := c.physicalIndex(i)
		if physical < 0 || physical >= c.Capacity {
			return SerpensState{}, fmt.Errorf("message does not fit linear traversal from start %d", c.Start)
		}
		s.Holes[physical] = runePtr(r)
	}
	return s, nil
}

// Decode traverses exactly length positions. A lost insertion is rendered as
// EmptyMarker so damage remains observable rather than being collapsed away.
func (c SerpensConfig) Decode(s SerpensState, length int) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	if len(s.Holes) != c.Capacity {
		return "", fmt.Errorf("state has %d holes, want %d", len(s.Holes), c.Capacity)
	}
	if length <= 0 || length > c.Capacity {
		return "", fmt.Errorf("invalid length %d", length)
	}
	marker := c.EmptyMarker
	if marker == 0 {
		marker = '?'
	}
	out := make([]rune, length)
	for i := range out {
		physical := c.physicalIndex(i)
		if physical < 0 || physical >= c.Capacity {
			return "", fmt.Errorf("traversal leaves spiral at step %d", i)
		}
		r := s.Holes[physical]
		if r == nil {
			out[i] = marker
		} else {
			out[i] = *r
		}
	}
	return string(out), nil
}

func SerpensRemove(s SerpensState, physical int) SerpensState {
	out := s.Clone()
	if physical >= 0 && physical < len(out.Holes) {
		out.Holes[physical] = nil
	}
	return out
}

func SerpensSubstitute(s SerpensState, physical int, replacement rune) SerpensState {
	out := s.Clone()
	if physical >= 0 && physical < len(out.Holes) {
		out.Holes[physical] = runePtr(replacement)
	}
	return out
}

func SerpensSwap(s SerpensState, a, b int) SerpensState {
	out := s.Clone()
	if a >= 0 && a < len(out.Holes) && b >= 0 && b < len(out.Holes) {
		out.Holes[a], out.Holes[b] = out.Holes[b], out.Holes[a]
	}
	return out
}

// SerpensCollapse removes a physical position and appends a new empty hole,
// modeling loss of the positional frame rather than loss of one marked slot.
func SerpensCollapse(s SerpensState, physical int) SerpensState {
	out := s.Clone()
	if physical < 0 || physical >= len(out.Holes) {
		return out
	}
	copy(out.Holes[physical:], out.Holes[physical+1:])
	out.Holes[len(out.Holes)-1] = nil
	return out
}

// CompatibleTraversals enumerates only the finite start/direction ambiguity;
// unknown association conventions are intentionally reported as unbounded in
// the experiment metadata rather than guessed here.
func (c SerpensConfig) CompatibleTraversals(s SerpensState, length int, startKnown, directionKnown bool) ([]string, error) {
	type traversal struct {
		start int
		dir   Direction
	}
	options := []traversal{{c.Start, c.Direction}}
	if !directionKnown {
		other := c.Start + length - 1
		if c.Direction == Reverse {
			other = c.Start - length + 1
		}
		options = append(options, traversal{other, Direction(1 - int(c.Direction))})
	}
	if !startKnown {
		options = options[:0]
		dirs := []Direction{c.Direction}
		if !directionKnown {
			dirs = []Direction{Forward, Reverse}
		}
		for _, dir := range dirs {
			if dir == Forward {
				for start := 0; start+length <= c.Capacity; start++ {
					options = append(options, traversal{start, dir})
				}
			} else {
				for start := length - 1; start < c.Capacity; start++ {
					options = append(options, traversal{start, dir})
				}
			}
		}
	}
	seen := map[string]bool{}
	var out []string
	for _, option := range options {
		alt := c
		alt.Start, alt.Direction = option.start, option.dir
		v, err := alt.Decode(s, length)
		if err != nil {
			return nil, err
		}
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out, nil
}
