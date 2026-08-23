package fontanafamily

import "fmt"

// Arismetricum is F11's invariant indexed-placement/lookup core. Values are
// opaque cues: the device does not claim to derive their meaning.
type Arismetricum struct {
	Slots map[int]string `json:"slots"`
}

func (a Arismetricum) Place(index int, cue string) (Arismetricum, error) {
	if index < 0 {
		return a, fmt.Errorf("negative index")
	}
	if cue == "" {
		return a, fmt.Errorf("empty cue")
	}
	out := Arismetricum{Slots: make(map[int]string, len(a.Slots)+1)}
	for k, v := range a.Slots {
		out.Slots[k] = v
	}
	out.Slots[index] = cue
	return out, nil
}

func (a Arismetricum) Lookup(index int) (string, bool) { v, ok := a.Slots[index]; return v, ok }

func (a Arismetricum) Remove(index int) Arismetricum {
	out := Arismetricum{Slots: make(map[int]string, len(a.Slots))}
	for k, v := range a.Slots {
		if k != index {
			out.Slots[k] = v
		}
	}
	return out
}

func (a Arismetricum) Swap(aIndex, bIndex int) Arismetricum {
	out := Arismetricum{Slots: make(map[int]string, len(a.Slots))}
	for k, v := range a.Slots {
		out.Slots[k] = v
	}
	av, aok := out.Slots[aIndex]
	bv, bok := out.Slots[bIndex]
	if aok {
		out.Slots[bIndex] = av
	} else {
		delete(out.Slots, bIndex)
	}
	if bok {
		out.Slots[aIndex] = bv
	} else {
		delete(out.Slots, aIndex)
	}
	return out
}

// Horalogius is a finite event-trigger model for F12's invariant core. It
// stores a cyclic temporal state and emits cue IDs. Meaning remains in the
// trained convention supplied to Recall.
type Horalogius struct {
	Period int            `json:"period"`
	Tick   int            `json:"tick"`
	Cues   map[int]string `json:"cues"`
}

func (h Horalogius) Validate() error {
	if h.Period <= 0 {
		return fmt.Errorf("period must be positive")
	}
	if h.Tick < 0 || h.Tick >= h.Period {
		return fmt.Errorf("tick %d outside period", h.Tick)
	}
	for tick, cue := range h.Cues {
		if tick < 0 || tick >= h.Period {
			return fmt.Errorf("cue tick %d outside period", tick)
		}
		if cue == "" {
			return fmt.Errorf("empty cue at tick %d", tick)
		}
	}
	return nil
}

func (h Horalogius) Advance(steps int) (Horalogius, []string, error) {
	if err := h.Validate(); err != nil {
		return h, nil, err
	}
	if steps < 0 {
		return h, nil, fmt.Errorf("negative advance")
	}
	var emitted []string
	for i := 0; i < steps; i++ {
		h.Tick = Normalize(h.Tick+1, h.Period)
		if cue, ok := h.Cues[h.Tick]; ok {
			emitted = append(emitted, cue)
		}
	}
	return h, emitted, nil
}

func Recall(cue string, learned map[string]string) (string, bool) {
	v, ok := learned[cue]
	return v, ok
}
