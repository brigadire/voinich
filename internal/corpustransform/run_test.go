package corpustransform

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeCorpusFile(t *testing.T, dir, name, content string) string {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRunTranspositionEndToEndDeterministic(t *testing.T) {
	dir := t.TempDir()
	corpus := writeCorpusFile(t, dir, "in.txt", "the quick brown fox\njumps over the lazy dog\nthe fox runs away fast\n")
	req := TranspositionRequest{
		CorpusPath: corpus,
		OutputPath: filepath.Join(dir, "out.txt"),
		GitCommit:  "deadbeef",
		LinePolicy: LinePolicyPreserve,
		Params:     TranspositionParams{Width: 4, Order: OrderKeyed, Round: 1, Seed: 3},
	}
	res1, err := RunTransposition(req)
	if err != nil {
		t.Fatal(err)
	}
	out1, err := os.ReadFile(req.OutputPath)
	if err != nil {
		t.Fatal(err)
	}

	req2 := req
	req2.OutputPath = filepath.Join(dir, "out2.txt")
	res2, err := RunTransposition(req2)
	if err != nil {
		t.Fatal(err)
	}
	out2, err := os.ReadFile(req2.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(out1) != string(out2) {
		t.Fatal("same corpus/method/parameters/seed did not produce byte-identical output")
	}
	if res1.Manifest.OutputSHA256 != res2.Manifest.OutputSHA256 {
		t.Fatal("manifest output SHA256 mismatch across identical runs")
	}

	origInput, err := os.ReadFile(corpus)
	if err != nil {
		t.Fatal(err)
	}
	if string(origInput) != "the quick brown fox\njumps over the lazy dog\nthe fox runs away fast\n" {
		t.Fatal("input corpus was modified")
	}

	manifestBytes, err := os.ReadFile(req.OutputPath + ".transform.json")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(manifestBytes), "\t") {
		t.Fatal("manifest unexpectedly contains a tab")
	}
}

func TestRunTranspositionOutputHasNoMetadataMarkers(t *testing.T) {
	dir := t.TempDir()
	corpus := writeCorpusFile(t, dir, "in.txt", "alpha beta gamma delta epsilon zeta\n")
	req := TranspositionRequest{
		CorpusPath: corpus,
		OutputPath: filepath.Join(dir, "out.txt"),
		LinePolicy: LinePolicyReflow,
		Params:     TranspositionParams{Width: 3, Order: OrderNatural, Round: 1},
	}
	if _, err := RunTransposition(req); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(req.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, marker := range []string{"transposition", "seed", "schema_version", "#"} {
		if strings.Contains(string(out), marker) {
			t.Fatalf("output corpus leaked transform metadata marker %q: %q", marker, out)
		}
	}
}

func TestRunHomophonicEndToEndWritesMappingAndManifest(t *testing.T) {
	dir := t.TempDir()
	corpus := writeCorpusFile(t, dir, "in.txt", "the quick brown fox\njumps over the lazy dog\n")
	req := HomophonicRequest{
		CorpusPath: corpus,
		OutputPath: filepath.Join(dir, "out.txt"),
		GitCommit:  "cafef00d",
		LinePolicy: LinePolicyPreserve,
		Params:     HomophonicParams{Model: HomophoneModelFixed, Homophones: 4, Selection: SelectionUniform, Seed: 1},
	}
	res, err := RunHomophonic(req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(req.OutputPath + ".mapping.tsv"); err != nil {
		t.Fatalf("mapping.tsv not written: %v", err)
	}
	if _, err := os.Stat(req.OutputPath + ".transform.json"); err != nil {
		t.Fatalf("manifest not written: %v", err)
	}
	if res.Manifest.MappingSHA256 == "" {
		t.Fatal("manifest missing mapping_sha256")
	}
	mappingBytes, err := os.ReadFile(req.OutputPath + ".mapping.tsv")
	if err != nil {
		t.Fatal(err)
	}
	if got := ShaBytes(mappingBytes); got != res.Manifest.MappingSHA256 {
		t.Fatalf("mapping_sha256 = %s, want %s", res.Manifest.MappingSHA256, got)
	}
}

func TestLinePolicyPreserveKeepsOriginalLineCountForHomophonic(t *testing.T) {
	dir := t.TempDir()
	corpus := writeCorpusFile(t, dir, "in.txt", "one two three\nfour five\nsix seven eight nine\n")
	req := HomophonicRequest{
		CorpusPath: corpus,
		OutputPath: filepath.Join(dir, "out.txt"),
		LinePolicy: LinePolicyPreserve,
		Params:     HomophonicParams{Model: HomophoneModelFixed, Homophones: 2, Selection: SelectionUniform, Seed: 1},
	}
	if _, err := RunHomophonic(req); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(req.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d output lines, want 3 (original line count preserved)", len(lines))
	}
	wantLens := []int{3, 2, 4}
	for i, l := range lines {
		if got := len(strings.Fields(l)); got != wantLens[i] {
			t.Fatalf("line %d: got %d tokens, want %d", i, got, wantLens[i])
		}
	}
}

func TestLinePolicyPreserveKeepsOriginalLineCountForTransposition(t *testing.T) {
	dir := t.TempDir()
	corpus := writeCorpusFile(t, dir, "in.txt", "one two three\nfour five\nsix seven eight nine\n")
	req := TranspositionRequest{
		CorpusPath: corpus,
		OutputPath: filepath.Join(dir, "out.txt"),
		LinePolicy: LinePolicyPreserve,
		Params:     TranspositionParams{Width: 3, Order: OrderNatural, Round: 1},
	}
	if _, err := RunTransposition(req); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(req.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	if len(lines) != 3 {
		t.Fatalf("got %d output lines, want 3 (original line-length sequence preserved)", len(lines))
	}
	wantLens := []int{3, 2, 4}
	total := 0
	for i, l := range lines {
		if got := len(strings.Fields(l)); got != wantLens[i] {
			t.Fatalf("line %d: got %d tokens, want %d", i, got, wantLens[i])
		}
		total += len(strings.Fields(l))
	}
	if total != 9 {
		t.Fatalf("total tokens = %d, want 9", total)
	}
}

func TestLinePolicyReflowUsesFixedWidth(t *testing.T) {
	dir := t.TempDir()
	tokens := make([]string, 0, 25)
	for range 25 {
		tokens = append(tokens, "tok")
	}
	corpus := writeCorpusFile(t, dir, "in.txt", strings.Join(tokens, " ")+"\n")
	req := TranspositionRequest{
		CorpusPath: corpus,
		OutputPath: filepath.Join(dir, "out.txt"),
		LinePolicy: LinePolicyReflow,
		Params:     TranspositionParams{Width: 5, Order: OrderNatural, Round: 1},
	}
	if _, err := RunTransposition(req); err != nil {
		t.Fatal(err)
	}
	out, err := os.ReadFile(req.OutputPath)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(out), "\n"), "\n")
	wantLineCount := 25/ReflowTokensPerLine + 1
	if len(lines) != wantLineCount {
		t.Fatalf("got %d lines, want %d (reflow width %d)", len(lines), wantLineCount, ReflowTokensPerLine)
	}
	for i, l := range lines[:len(lines)-1] {
		if got := len(strings.Fields(l)); got != ReflowTokensPerLine {
			t.Fatalf("line %d: got %d tokens, want %d", i, got, ReflowTokensPerLine)
		}
	}
}
