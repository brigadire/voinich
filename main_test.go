package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileTokenFrequencyList(t *testing.T) {
	fileName := writeTestInput(t, "daiin: 847\nol: 557\naiin: 504\n")

	if _, err := readFileToken(fileName); err == nil {
		t.Fatal("readFileToken() error = nil, want an error for aggregated input")
	}
}

func TestReadFileTokenPlainText(t *testing.T) {
	fileName := writeTestInput(t, "before alpha after\nbefore alpha next\nother alpha after\nthird alpha last\n")

	tokens, err := readFileToken(fileName)
	if err != nil {
		t.Fatalf("readFileToken() error = %v", err)
	}

	alpha := findToken(t, tokens, "alpha")
	if alpha.Count != 4 {
		t.Fatalf("alpha.Count = %d, want 4", alpha.Count)
	}

	wantBefore := []Token{{Token: "before", Count: 2}, {Token: "other", Count: 1}, {Token: "third", Count: 1}}
	wantAfter := []Token{{Token: "after", Count: 2}, {Token: "last", Count: 1}, {Token: "next", Count: 1}}
	wantPositions := []Position{{Position: 1, Count: 4}}
	assertRelatedTokens(t, "WordBefore", alpha.WordBefore, wantBefore)
	assertRelatedTokens(t, "WordAfter", alpha.WordAfter, wantAfter)
	assertPositions(t, alpha.PositionInString, wantPositions)
}

func TestReadFileTokenTopThreePositions(t *testing.T) {
	fileName := writeTestInput(t, "alpha x\ny alpha\ny alpha z\na b alpha\nalpha\nd e f alpha\n")

	tokens, err := readFileToken(fileName)
	if err != nil {
		t.Fatalf("readFileToken() error = %v", err)
	}

	alpha := findToken(t, tokens, "alpha")
	want := []Position{
		{Position: 0, Count: 2},
		{Position: 1, Count: 2},
		{Position: 2, Count: 1},
	}
	assertPositions(t, alpha.PositionInString, want)
}

func findToken(t *testing.T, tokens []Tokens, token string) Tokens {
	t.Helper()
	for _, item := range tokens {
		if item.Token == token {
			return item
		}
	}
	t.Fatalf("token %q not found", token)
	return Tokens{}
}

func assertRelatedTokens(t *testing.T, field string, got, want []Token) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("%s length = %d, want %d: %+v", field, len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("%s[%d] = %+v, want %+v", field, i, got[i], want[i])
		}
	}
}

func assertPositions(t *testing.T, got, want []Position) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("PositionInString length = %d, want %d: %+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("PositionInString[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func writeTestInput(t *testing.T, content string) string {
	t.Helper()
	fileName := filepath.Join(t.TempDir(), "input.txt")
	if err := os.WriteFile(fileName, []byte(content), 0o600); err != nil {
		t.Fatalf("write input: %v", err)
	}
	return fileName
}
