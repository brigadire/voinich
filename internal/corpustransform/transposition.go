package corpustransform

import "fmt"

// TranspositionParams fully determines a deterministic rectangular
// columnar transposition (task46 sections 3-4).
type TranspositionParams struct {
	Width int
	Order string // OrderNatural or OrderKeyed
	Round int    // >=1
	Seed  int64
}

// RemainderPolicy documents, verbatim, the deterministic rule Transpose
// uses for an incomplete final row. It is recorded in every transposition
// manifest.
const RemainderPolicy = "short-columns: token i (0-based) belongs to row i/width, column i%width; " +
	"a column is read by increasing row as long as row*width+column < token_count; " +
	"trailing columns whose only membership would be the missing final row are simply shorter, " +
	"no padding token is introduced or removed"

// ColumnOrder returns the deterministic column read order for width and
// order. OrderNatural is [0,1,...,width-1]; OrderKeyed is a Fisher-Yates
// permutation of the same indices derived purely from seed and width via
// subRand (task46 section 4: no textual password/key, seed only).
func ColumnOrder(width int, order string, seed int64) ([]int, error) {
	cols := make([]int, width)
	for i := range cols {
		cols[i] = i
	}
	switch order {
	case OrderNatural:
		return cols, nil
	case OrderKeyed:
		r := subRand("transposition-keyed-column-order", seed, uint64(width))
		for i := width - 1; i > 0; i-- {
			j := r.IntN(i + 1)
			cols[i], cols[j] = cols[j], cols[i]
		}
		return cols, nil
	default:
		return nil, fmt.Errorf("unsupported transposition order %q", order)
	}
}

// Transpose applies rectangular columnar transposition to tokens according
// to p. It never adds, drops, or alters a token (task46 section 3
// invariants): output is always a permutation of the input.
func Transpose(tokens []string, p TranspositionParams) ([]string, error) {
	if p.Width < 1 {
		return nil, fmt.Errorf("transposition width must be >= 1, got %d", p.Width)
	}
	if p.Round < 1 {
		return nil, fmt.Errorf("transposition rounds must be >= 1, got %d", p.Round)
	}
	colOrder, err := ColumnOrder(p.Width, p.Order, p.Seed)
	if err != nil {
		return nil, err
	}
	out := tokens
	for round := 0; round < p.Round; round++ {
		out = transposeOnce(out, p.Width, colOrder)
	}
	return out, nil
}

// Untranspose is the exact inverse of Transpose, used by round-trip tests
// (task46 section 18) to prove reversibility.
func Untranspose(tokens []string, p TranspositionParams) ([]string, error) {
	if p.Width < 1 {
		return nil, fmt.Errorf("transposition width must be >= 1, got %d", p.Width)
	}
	if p.Round < 1 {
		return nil, fmt.Errorf("transposition rounds must be >= 1, got %d", p.Round)
	}
	colOrder, err := ColumnOrder(p.Width, p.Order, p.Seed)
	if err != nil {
		return nil, err
	}
	out := tokens
	for round := 0; round < p.Round; round++ {
		out = untransposeOnce(out, p.Width, colOrder)
	}
	return out, nil
}

// transposeOnce reads tokens by column, in colOrder, top-to-bottom, exactly
// reproducing the worked example in task46 section 3: for width=4,
// "A B C D E F G H I J K L" -> "A E I B F J C G K D H L".
func transposeOnce(tokens []string, width int, colOrder []int) []string {
	n := len(tokens)
	out := make([]string, 0, n)
	for _, col := range colOrder {
		for row := 0; ; row++ {
			idx := row*width + col
			if idx >= n {
				break
			}
			out = append(out, tokens[idx])
		}
	}
	return out
}

// untransposeOnce inverts transposeOnce: it replays the identical
// column/row enumeration order used to build the transposed stream, and
// scatters each transposed token back to the grid index that produced it.
func untransposeOnce(transposed []string, width int, colOrder []int) []string {
	n := len(transposed)
	orig := make([]string, n)
	pos := 0
	for _, col := range colOrder {
		for row := 0; ; row++ {
			idx := row*width + col
			if idx >= n {
				break
			}
			orig[idx] = transposed[pos]
			pos++
		}
	}
	return orig
}
