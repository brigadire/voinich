package speculumf01

import "math/rand"

// TestMessage is one pre-registered entry in the task76 Block 3 message
// set. Registered before any experiment ran; do not add, remove, or
// reword entries after seeing results (see EXPERIMENTAL_PROTOCOL.md).
type TestMessage struct {
	Text     string
	Category string
	Natural  bool
}

// NaturalMessages is the fixed natural-language message set: classical
// Latin words (V for U, no J/U/W, matching Latin23) chosen to cover a
// deliberate spread of length, repetition, and rareness before any
// decoding experiment was run.
var NaturalMessages = []TestMessage{
	{"PAX", "short_no_repeat", true},
	{"ANNA", "palindrome_high_repeat", true},
	{"DEVS", "short_no_repeat", true},
	{"EXTRA", "rare_symbol_x", true},
	{"NATVRA", "medium_low_repeat", true},
	{"GLORIA", "medium_no_repeat", true},
	{"MEMORIA", "medium_repeat_m", true},
	{"FONTANA", "medium_repeat_an", true},
	{"KALENDAE", "rare_symbol_k", true},
	{"SPECVLVM", "thematic_device_name", true},
	{"EXPERIMENTA", "long_high_repeat_e_rare_x", true},
	{"CONSTANTINA", "long_high_repeat_nt", true},
}

// BaseLexicon is the small, pre-registered reference word list used only
// to model the "known language" contribution C in Block 7 (which
// candidate decodings are plausible words, not which one is *the*
// message). It is fixed before ablation/corruption experiments ran and
// is explicitly small and illustrative, not an exhaustive Latin lexicon:
// results computed against it are order-of-magnitude, not precise.
var BaseLexicon = []string{
	"PAX", "REX", "LEX", "DEVS", "ANNA", "AMOR", "MARE", "TERRA", "AQVA",
	"IGNIS", "EXTRA", "NATVRA", "GLORIA", "MEMORIA", "FONTANA", "FORTVNA",
	"SPECVLVM", "STELLA", "LVNA", "SOL", "ROTA", "CLAVIS", "CATENA",
	"KALENDAE", "EXPERIMENTA", "CONSTANTINA", "SECRETVM", "THESAVRVS",
	"IMAGINATIO", "ARTIFITIVM", "INSTRVMENTVM", "RESERVATIO", "OBLIVIO",
	"CIRCVLVS", "ALFABETVM", "MACHINA", "VENETIA", "PADVA",
}

func Lexicon() map[string]bool {
	m := make(map[string]bool, len(BaseLexicon))
	for _, w := range BaseLexicon {
		m[w] = true
	}
	return m
}

// GenerateRandomControls builds one random string per requested length,
// matched in alphabet to `a`, using a fixed seed. Any draw that lands in
// lexicon (i.e. would accidentally look like a real word) is rejected and
// redrawn, since the control condition's whole point is "same length and
// alphabet, zero language predictability".
func GenerateRandomControls(a Alphabet, lengths []int, seed int64, lexicon map[string]bool) []TestMessage {
	r := rand.New(rand.NewSource(seed))
	out := make([]TestMessage, 0, len(lengths))
	for _, l := range lengths {
		var text string
		for attempt := 0; attempt < 1000; attempt++ {
			letters := make([]rune, l)
			for i := range letters {
				letters[i] = a.At(r.Intn(a.Size()))
			}
			text = string(letters)
			if !lexicon[text] {
				break
			}
		}
		out = append(out, TestMessage{Text: text, Category: "random_control", Natural: false})
	}
	return out
}
