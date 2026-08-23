package fontanafamily

import "fmt"

// Rota is the invariant single cyclic selector supported by F07. Any keyed
// stepping or multiple-wheel coupling would be an unsupported extension.
type Rota struct {
	Alphabet []rune
	Offset   int `json:"offset"`
}

func (r Rota) Rotate(steps int) Rota {
	if len(r.Alphabet) == 0 {
		return r
	}
	r.Offset = Normalize(r.Offset+steps, len(r.Alphabet))
	return r
}

func (r Rota) Observe() (rune, error) {
	if err := ValidateAlphabet(r.Alphabet); err != nil {
		return 0, err
	}
	return r.Alphabet[Normalize(r.Offset, len(r.Alphabet))], nil
}

func (r Rota) Period(step int) int {
	n := len(r.Alphabet)
	if n == 0 {
		return 0
	}
	a, b := n, Normalize(step, n)
	for b != 0 {
		a, b = b, a%b
	}
	return n / a
}

// Cylinder models F10 profile R0: independently positioned circular bands and
// one axial reading line. Band count, alphabet and independence are explicit H
// parameters in the profile, never claims about the manuscript drawing.
type Cylinder struct {
	Alphabet []rune `json:"alphabet"`
	Offsets  []int  `json:"offsets"`
}

func (c Cylinder) Validate() error {
	if err := ValidateAlphabet(c.Alphabet); err != nil {
		return err
	}
	if len(c.Offsets) == 0 {
		return fmt.Errorf("cylinder has no bands")
	}
	return nil
}

func (c Cylinder) RotateBand(band, steps int) (Cylinder, error) {
	if err := c.Validate(); err != nil {
		return c, err
	}
	if band < 0 || band >= len(c.Offsets) {
		return c, fmt.Errorf("band %d outside 0..%d", band, len(c.Offsets)-1)
	}
	out := Cylinder{Alphabet: append([]rune(nil), c.Alphabet...), Offsets: append([]int(nil), c.Offsets...)}
	out.Offsets[band] = Normalize(out.Offsets[band]+steps, len(out.Alphabet))
	return out, nil
}

func (c Cylinder) Read(direction Direction) (string, error) {
	if err := c.Validate(); err != nil {
		return "", err
	}
	out := make([]rune, len(c.Offsets))
	for i := range out {
		band := i
		if direction == Reverse {
			band = len(out) - 1 - i
		}
		out[i] = c.Alphabet[Normalize(c.Offsets[band], len(c.Alphabet))]
	}
	return string(out), nil
}

func (c Cylinder) Encode(message string, direction Direction) (Cylinder, error) {
	if err := c.Validate(); err != nil {
		return c, err
	}
	runes := []rune(message)
	if len(runes) != len(c.Offsets) {
		return c, fmt.Errorf("message length %d != bands %d", len(runes), len(c.Offsets))
	}
	out := c
	out.Offsets = append([]int(nil), c.Offsets...)
	for i, r := range runes {
		idx := -1
		for j, a := range c.Alphabet {
			if a == r {
				idx = j
				break
			}
		}
		if idx < 0 {
			return c, fmt.Errorf("symbol %q outside alphabet", r)
		}
		band := i
		if direction == Reverse {
			band = len(runes) - 1 - i
		}
		out.Offsets[band] = idx
	}
	return out, nil
}

func (c Cylinder) StateSpaceSize() uint64 {
	if len(c.Alphabet) == 0 {
		return 0
	}
	v := uint64(1)
	for range c.Offsets {
		v *= uint64(len(c.Alphabet))
	}
	return v
}
