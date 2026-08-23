package speculumf01

import "gopkg.in/yaml.v3"

// StateDoc is the on-disk serialization of a device State: exactly the
// ring offsets, nothing else. It intentionally excludes message length,
// which rings were "used", or any other fact the physical object does not
// itself expose.
type StateDoc struct {
	NumRings     int   `yaml:"num_rings"`
	AlphabetSize int   `yaml:"alphabet_size"`
	RingOffsets  []int `yaml:"ring_offsets"` // index 0 = innermost ring; -1 = missing/unreadable
}

func (c Config) ToDoc(s State) StateDoc {
	return StateDoc{NumRings: c.NumRings, AlphabetSize: c.Alphabet.Size(), RingOffsets: s.Offsets}
}

func (d StateDoc) State() State { return State{Offsets: d.RingOffsets} }

func MarshalState(c Config, s State) ([]byte, error) {
	return yaml.Marshal(c.ToDoc(s))
}

func UnmarshalState(data []byte) (StateDoc, error) {
	var d StateDoc
	err := yaml.Unmarshal(data, &d)
	return d, err
}
