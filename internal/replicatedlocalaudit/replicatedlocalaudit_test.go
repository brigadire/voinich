package replicatedlocalaudit

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
)

func TestCountSequenceRespectsLineBoundary(t *testing.T) {
	tokens := []token{{Text: "a", Line: "1"}, {Text: "b", Line: "1"}, {Text: "c", Line: "2"}}
	if got := countSequence(tokens, []string{"b", "c"}); got != 0 {
		t.Fatalf("cross-line sequence counted: %d", got)
	}
	if got := countSequence(tokens, []string{"a", "b"}); got != 1 {
		t.Fatalf("within-line count = %d", got)
	}
}

func TestMarkovNeverTrainsOnHeldoutBlock(t *testing.T) {
	blocks := []block{
		{ID: "A#0", Joint: "A/1", Tokens: []token{{Text: "held", Line: "1"}, {Text: "held", Line: "1"}}},
		{ID: "A#1", Joint: "A/1", Tokens: []token{{Text: "train", Line: "2"}, {Text: "train", Line: "2"}}},
	}
	got, available := markovBlocks(blocks, 1)
	if available != 2 || len(got) != 2 {
		t.Fatalf("availability = %d/%d", available, len(got))
	}
	for _, x := range got[0].Tokens {
		if x.Text != "train" {
			t.Fatalf("held-out content leaked into model: %q", x.Text)
		}
	}
}

func TestCheckpointRequiresMatchingFingerprint(t *testing.T) {
	path := filepath.Join(t.TempDir(), "checkpoint.json")
	c := Config{Permutations: 10}
	cp := newCheckpoint("one", c)
	cp.ShuffleCompleted = 4
	if err := saveCheckpoint(path, cp); err != nil {
		t.Fatal(err)
	}
	loaded, ok, err := loadCheckpoint(path, "one", c)
	if err != nil || !ok || loaded.ShuffleCompleted != 4 {
		t.Fatalf("matching checkpoint: ok=%v err=%v cp=%+v", ok, err, loaded)
	}
	loaded, ok, err = loadCheckpoint(path, "two", c)
	if err != nil || ok || loaded.ShuffleCompleted != 0 {
		t.Fatalf("mismatching checkpoint was resumed: ok=%v err=%v", ok, err)
	}
}

func TestProgressBar(t *testing.T) {
	var b bytes.Buffer
	p := newProgress(&b)
	p.begin(3, "Null")
	p.update(10, 10, "Null")
	s := b.String()
	if !strings.Contains(s, "[3/8]") || !strings.Contains(s, "[====================]") || !strings.Contains(s, "elapsed") {
		t.Fatalf("status bar: %q", s)
	}
}
