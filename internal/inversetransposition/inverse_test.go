package inversetransposition

import (
	"reflect"
	"testing"

	"zcore.dev/voinich/internal/corpustransform"
)

func TestEveryTask46ModeHasExactInverse(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}
	for _, c := range []Candidate{
		{Width: 1, Order: "natural", Rounds: 1},
		{Width: 2, Order: "natural", Rounds: 2},
		{Width: 4, Order: "natural", Rounds: 1},
		{Width: 3, Order: "keyed", Rounds: 1, Seed: 17},
		{Width: 7, Order: "keyed", Rounds: 3, Seed: 99},
	} {
		forward, err := corpustransform.Transpose(in, corpustransform.TranspositionParams{Width: c.Width, Order: c.Order, Round: c.Rounds, Seed: c.Seed})
		if err != nil {
			t.Fatal(err)
		}
		back, err := c.Apply(forward)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(back, in) {
			t.Fatalf("%+v: got %v, want %v", c, back, in)
		}
	}
}
