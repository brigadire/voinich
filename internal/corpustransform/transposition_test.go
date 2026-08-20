package corpustransform

import (
	"reflect"
	"strings"
	"testing"
)

func words(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Fields(s)
}

func TestTransposeKnownExample(t *testing.T) {
	in := words("A B C D E F G H I J K L")
	out, err := Transpose(in, TranspositionParams{Width: 4, Order: OrderNatural, Round: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := words("A E I B F J C G K D H L")
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("got %v, want %v", out, want)
	}
}

func TestTransposeIncompleteFinalRow(t *testing.T) {
	// width=4 over 10 tokens: rows [0-3][4-7][8-9], columns 0,1 have 3
	// entries, columns 2,3 have only 2 (no padding).
	in := words("A B C D E F G H I J")
	out, err := Transpose(in, TranspositionParams{Width: 4, Order: OrderNatural, Round: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := words("A E I B F J C G D H")
	if !reflect.DeepEqual(out, want) {
		t.Fatalf("got %v, want %v", out, want)
	}
}

func TestTransposeWidthOneIsIdentity(t *testing.T) {
	in := words("A B C D E")
	out, err := Transpose(in, TranspositionParams{Width: 1, Order: OrderNatural, Round: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("got %v, want identity %v", out, in)
	}
}

func TestTransposeWidthGreaterThanTokenCount(t *testing.T) {
	in := words("A B C")
	out, err := Transpose(in, TranspositionParams{Width: 10, Order: OrderNatural, Round: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("width > token count should be identity under natural order: got %v, want %v", out, in)
	}
}

func TestTransposePreservesMultiset(t *testing.T) {
	in := words("A B C D E F G H I J K L M")
	for _, width := range []int{1, 2, 3, 4, 5, 6, 13, 20} {
		out, err := Transpose(in, TranspositionParams{Width: width, Order: OrderNatural, Round: 1})
		if err != nil {
			t.Fatal(err)
		}
		if len(out) != len(in) {
			t.Fatalf("width %d: got %d tokens, want %d", width, len(out), len(in))
		}
		if !MultisetEqual(in, out) {
			t.Fatalf("width %d: multiset not preserved", width)
		}
	}
}

func TestTransposeKeyedIsDeterministic(t *testing.T) {
	in := words("A B C D E F G H I J K L M N O P")
	p := TranspositionParams{Width: 4, Order: OrderKeyed, Round: 1, Seed: 42}
	out1, err := Transpose(in, p)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := Transpose(in, p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out1, out2) {
		t.Fatalf("same seed produced different keyed permutations: %v vs %v", out1, out2)
	}
	natural, err := Transpose(in, TranspositionParams{Width: 4, Order: OrderNatural, Round: 1})
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(out1, natural) {
		t.Fatal("keyed order produced the same result as natural order; permutation is not being applied")
	}
	other, err := Transpose(in, TranspositionParams{Width: 4, Order: OrderKeyed, Round: 1, Seed: 7})
	if err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(out1, other) {
		t.Fatal("different seeds produced the same keyed permutation")
	}
}

func TestTransposeRoundsAreReproducible(t *testing.T) {
	in := words("A B C D E F G H I J K L M N O P")
	p := TranspositionParams{Width: 3, Order: OrderKeyed, Round: 3, Seed: 5}
	out1, err := Transpose(in, p)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := Transpose(in, p)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out1, out2) {
		t.Fatalf("repeated rounds not reproducible: %v vs %v", out1, out2)
	}
}

func TestTransposeEmptyCorpus(t *testing.T) {
	out, err := Transpose(nil, TranspositionParams{Width: 4, Order: OrderNatural, Round: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 0 {
		t.Fatalf("got %v, want empty", out)
	}
}

func TestTransposeOneTokenCorpus(t *testing.T) {
	in := words("A")
	out, err := Transpose(in, TranspositionParams{Width: 4, Order: OrderNatural, Round: 1})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(out, in) {
		t.Fatalf("got %v, want %v", out, in)
	}
}

func TestTransposeRoundTrip(t *testing.T) {
	in := words("A B C D E F G H I J K L M N O P Q R S T U V W X Y Z")
	cases := []TranspositionParams{
		{Width: 1, Order: OrderNatural, Round: 1},
		{Width: 4, Order: OrderNatural, Round: 1},
		{Width: 5, Order: OrderNatural, Round: 1},
		{Width: 4, Order: OrderKeyed, Round: 1, Seed: 99},
		{Width: 3, Order: OrderKeyed, Round: 2, Seed: 3},
		{Width: 100, Order: OrderNatural, Round: 1},
	}
	for _, p := range cases {
		out, err := Transpose(in, p)
		if err != nil {
			t.Fatalf("%+v: %v", p, err)
		}
		back, err := Untranspose(out, p)
		if err != nil {
			t.Fatalf("%+v: %v", p, err)
		}
		if !reflect.DeepEqual(back, in) {
			t.Fatalf("%+v: inverse(transform(T)) != T: got %v, want %v", p, back, in)
		}
	}
}
