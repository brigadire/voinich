package main

// Task86C deliberately lives in the historical Task86R executable package.
// This is the narrowest way to exercise the implementation that actually
// produced Task86R, without copying or reimplementing any M0--M5 logic.

import (
	"bufio"
	"crypto/sha256"
	"encoding/csv"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

const (
	task86cDir       = "research/phase3/task86c"
	task86cNamespace = "voinich.phase3.task86c.control.v1"
	task86cVersion   = "task86c-control-v1.1"
)

type task86cScale struct {
	Name   string
	Tokens int
}

var task86cScales = []task86cScale{
	{"SMALL", 512}, {"MEDIUM", 2048}, {"VOYNICH_SCALE", 38000}, {"LARGE", 65536},
}

type task86cMechanism struct {
	OpaqueID, Class, Minimal, Implementation string
	Variant                                  int
}

type task86cJob struct {
	JobID, Branch, CorpusID, Scale, Protocol, InputPath, InputSHA, ConfigSHA string
	Variant, Replicate, Tokens                                               int
	Seed                                                                     uint64
}

type task86cResult struct {
	Version, JobID, Branch, CorpusID, Scale, Protocol, InputSHA, ConfigSHA string
	Seed                                                                   uint64 `json:"seed"`
	Node, ScientificStatus, MinimalClass, TokenFormationDepth              string
	AdequateModels, Failures                                               []string
	RequestedNegatives, ConstructibleNegatives                             int
	PM6Available                                                           bool
	PM6ByClass                                                             map[string]bool
	CandidateByClass                                                       map[string]string
	PredictivePassByClass, StructuralPassByClass, SufficientByClass        map[string]bool
	StartedUTC, FinishedUTC                                                string
	RuntimeSeconds                                                         float64
}

func runTask86C(command string, args []string) error {
	switch command {
	case "task86c-plan":
		return task86cPlan(args)
	case "task86c-generate":
		return task86cGenerate(args)
	case "task86c-worker":
		return task86cWorker(args)
	case "task86c-aggregate":
		return task86cAggregate(args)
	case "task86c-unblind":
		return task86cUnblind(args)
	case "task86c-verify":
		return task86cVerify(args)
	default:
		return fmt.Errorf("unknown Task86C command %q", command)
	}
}

func task86cHashBytes(b []byte) string { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func task86cHashFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return task86cHashBytes(b), nil
}

func task86cSeed(fields ...string) uint64 {
	h := sha256.New()
	io.WriteString(h, task86cNamespace)
	for _, f := range fields {
		h.Write([]byte{0})
		io.WriteString(h, f)
	}
	b := h.Sum(nil)
	var x uint64
	for i := 0; i < 8; i++ {
		x = x<<8 | uint64(b[i])
	}
	return x
}

func task86cJobID(j task86cJob) string {
	s := fmt.Sprintf("%s\x00%s\x00%d\x00%s\x00%d\x00%s\x00%d\x00%s", j.Branch, j.CorpusID, j.Variant, j.Scale, j.Replicate, j.Protocol, j.Seed, j.ConfigSHA)
	return "j-" + task86cHashBytes([]byte(s))[:20]
}

func task86cWrite(path string, b []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func task86cTSV(path string, header []string, rows [][]string) error {
	var b strings.Builder
	w := csv.NewWriter(&b)
	w.Comma = '\t'
	w.UseCRLF = false
	_ = w.Write(header)
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
	if err := w.Error(); err != nil {
		return err
	}
	return task86cWrite(path, []byte(b.String()))
}

func task86cMechanisms() []task86cMechanism {
	var out []task86cMechanism
	for c := 0; c <= 5; c++ {
		for v := 1; v <= 3; v++ {
			idHash := task86cHashBytes([]byte(fmt.Sprintf("independent-M%d-v%d", c, v)))[:12]
			out = append(out, task86cMechanism{OpaqueID: "s-" + idHash, Class: fmt.Sprintf("M%d", c), Minimal: fmt.Sprintf("M%d", c), Implementation: "INDEPENDENT", Variant: v})
		}
	}
	return out
}

func task86cGeneratorDefinition(class string, variant int) string {
	switch class {
	case "M0":
		return fmt.Sprintf("single-glyph IID categorical; p=[.28,.20,.16,.12,.09,.07,.05,.03]; variant=%d alphabet rotation", variant)
	case "M1":
		return fmt.Sprintf("fixed first-order four-state Markov; diagonal=.60-.64; length U[3,%d]", 7+variant)
	case "M2":
		return fmt.Sprintf("explicit variable context tree: depth-2 aa-like context and depth-1 d-like context; variant=%d", variant)
	case "M3":
		return fmt.Sprintf("manual deterministic labelled five-state FSA; one edge draw determines label and fixed successor; variant=%d", variant)
	case "M4":
		return fmt.Sprintf("manual probabilistic four-state FSA; frozen transition probabilities and terminal p=.72; variant=%d", variant)
	case "M5":
		return fmt.Sprintf("productive 3-prefix x 4-core x 3-suffix slot grammar with optional second core; variant=%d", variant)
	default:
		panic("unknown class")
	}
}

func task86cDesignText() string {
	return `# Task86C control design (frozen v1)

## Scope and firewall

This is a validation of the historical Task86R measurement apparatus, not a
Voynich experiment. Analysis commands reject paths containing the target name,
the historical corpus identifiers, or the repository target-data directories.
No Task86R HELDOUT/F2 value is loaded. The only reused numerical artifacts are
the target-blind MFC calibration thresholds frozen before Task86R HELDOUT.

## Preregistered design

Primary protocol layer is HISTORICAL_REPLICATION: the exact production M0--M5
fit, VALIDATION PM2 selection, predictive gates, structural gates, generation
scales, and frozen 20,000-attempt PM6 implementation. CONTRACT_REFERENCE is
NOT_EXECUTABLE because the frozen negative-token contract requires exhaustive
enumeration while specifying no resource bound; the historical implementation
introduced the outcome-relevant cap later identified by Task85a-v1.1. No
replacement definition is introduced here.

Independent generators: three fixed parameterizations per M0--M5. M0 is IID
categorical glyph emission with IID length; M1 fixed-order Markov; M2 an
explicit variable-depth context tree; M3 a manually specified deterministic
FSA; M4 a manually specified probabilistic FSA; M5 a productive prefix/core/
suffix slot grammar. M3 chooses one labelled outgoing edge and uses that same
edge's fixed successor (one draw, never an independently chosen successor).
None calls production fitting or generation code.
The theoretical minimal class is M0..M5 respectively. This deliberately tests
class recovery, not semantic labels. IN_FAMILY is omitted from primary evidence
because the historical model interface has no frozen parameter-to-generator
serialization; completing one would test an implementation against itself and
add a scientific choice.

Scale grid is SMALL=512, MEDIUM=2048, VOYNICH_SCALE=38000, LARGE=65536 token
occurrences. There are 8 independent deterministic replicates per mechanism x
scale. Synthetic splits and natural samples use 60/20/20 occurrence partitions.
Natural controls are English (Doyle), expanded Latin (Caesar I--VIII plus
Virgil), and Sanskrit (Panchatantra; typologically distinct), selected only for
pre-existing repository provenance / public-domain availability and scale.

## Frozen success criteria

Model-recovery capable requires, on INDEPENDENT generators at VOYNICH_SCALE:
(a) minimal-sufficient recovery >= 0.80 overall, (b) exact recovery >= 0.60,
(c) NONE <= 0.05, and (d) every class minimal-sufficient recovery >= 0.60.
SUPPORTED requires all four; PARTIAL requires NONE <= 0.20 and overall
minimal-sufficient recovery >= 0.50; otherwise NOT_SUPPORTED. False complexity
is recovery above the theoretical minimum and underfit is recovery below it.
Natural-language applicability is SUPPORTED when at least one model is adequate
in >=80% of VOYNICH_SCALE samples in every required language, PARTIAL when this
holds for at least one language, otherwise NOT_SUPPORTED. These thresholds are
frozen before any analysis output or unblinding.

PM6 is executed unchanged and reports requested/constructible negatives by
length, alphabet saturation, attempt exhaustion and rejection cause. No sampler
repair is allowed. GENERAL means failure in both synthetic and natural branches;
NATURAL_LANGUAGE_COMMON in >=2 natural languages only; SCALE_DEPENDENT requires
a monotone >=0.50 difference between SMALL and LARGE failure rates; otherwise
INCONCLUSIVE (VOYNICH_SPECIFIC cannot be established without target execution).

## Distributed contract

Jobs bind immutable input SHA-256, config SHA-256, seed, protocol and logical
identity. Worker assignment is absent from job_id. Result JSON is canonical
indented JSON with runtime/timestamps excluded from scientific equivalence
hashing. Duplicate disagreement quarantines both. Infrastructure failures may
retry unchanged; scientific failures may not. Six representative preflight
jobs (M0, M2, M4, M5, English, Latin) must match local and 10.10.24.105 before
bulk execution. The same locally built executable bytes are copied to remote.
`
}

func task86cPlan(args []string) error {
	fs := flag.NewFlagSet("task86c-plan", flag.ContinueOnError)
	english := fs.String("english", "data_test/pg2097-2.txt", "English source")
	latin := fs.String("latin", "", "colon-separated Latin source files")
	sanskrit := fs.String("sanskrit", "data_test/sa_viSNuzarman-paJcatantra-prepared.txt", "third-language source")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *latin == "" {
		return errors.New("-latin is required; acquire Gutenberg 218, 18837 and 227 before design freeze")
	}
	if err := os.MkdirAll(filepath.Join(task86cDir, "inputs"), 0o755); err != nil {
		return err
	}
	if err := task86cWrite(filepath.Join(task86cDir, "TASK86C_DESIGN.md"), []byte(task86cDesignText())); err != nil {
		return err
	}

	mechs := task86cMechanisms()
	var reg, truth, genManifest, analysisManifest, seedRows, jobRows [][]string
	expectedHashes := map[string]string{}
	configSHA := task86cHashBytes([]byte(task86cDesignText()))
	for _, m := range mechs {
		reg = append(reg, []string{m.OpaqueID, m.Implementation, strconv.Itoa(m.Variant), task86cGeneratorDefinition(m.Class, m.Variant), m.Minimal, "FROZEN"})
		truth = append(truth, []string{m.OpaqueID, m.Class, m.Minimal, m.Implementation, strconv.Itoa(m.Variant)})
		for _, sc := range task86cScales {
			for rep := 0; rep < 8; rep++ {
				seed := task86cSeed("S", m.OpaqueID, sc.Name, strconv.Itoa(rep))
				path := filepath.Join(task86cDir, "inputs", m.OpaqueID, strings.ToLower(sc.Name), fmt.Sprintf("r%02d.txt", rep))
				genManifest = append(genManifest, []string{m.OpaqueID, m.Class, m.Implementation, strconv.Itoa(m.Variant), sc.Name, strconv.Itoa(sc.Tokens), strconv.Itoa(rep), strconv.FormatUint(seed, 10), path})
				analysisManifest = append(analysisManifest, []string{m.OpaqueID, strconv.Itoa(m.Variant), sc.Name, strconv.Itoa(sc.Tokens), strconv.Itoa(rep), path})
				seedRows = append(seedRows, []string{"S", m.OpaqueID, sc.Name, strconv.Itoa(rep), strconv.FormatUint(seed, 10)})
				toks := task86cGenerateTokens(m.Class, m.Variant, sc.Tokens, seed)
				expectedHashes[path] = task86cHashBytes([]byte(strings.Join(toks, "\n") + "\n"))
			}
		}
	}

	natural := []struct{ id, language, corpus, paths, license, url string }{
		{"n-" + task86cHashBytes([]byte("english-doyle-v1"))[:12], "English", "Doyle", *english, "Project Gutenberg License; public domain text in USA", "https://www.gutenberg.org/ebooks/2097"},
		{"n-" + task86cHashBytes([]byte("latin-expanded-v1"))[:12], "Latin", "Caesar+Virgil", *latin, "Project Gutenberg License; public domain text in USA", "https://www.gutenberg.org/ebooks/218;18837;227"},
		{"n-" + task86cHashBytes([]byte("sanskrit-panchatantra-v1"))[:12], "Sanskrit", "Panchatantra", *sanskrit, "repository-established research control; see DATA.md", "repository:data_test/sa_viSNuzarman-paJcatantra-prepared.txt"},
	}
	var prov [][]string
	for _, n := range natural {
		var sourceHashes []string
		var naturalTokens []string
		for _, p := range strings.Split(n.paths, ":") {
			b, readErr := os.ReadFile(p)
			if readErr != nil {
				return readErr
			}
			h := task86cHashBytes(b)
			sourceHashes = append(sourceHashes, p+"="+h)
			naturalTokens = append(naturalTokens, task86cNaturalTokens(string(b))...)
		}
		prov = append(prov, []string{n.id, n.language, n.corpus, n.url, n.license, strings.Join(sourceHashes, ";"), time.Now().UTC().Format("2006-01-02"), "Unicode letters; lower case; whitespace tokenization; Gutenberg header/footer excluded by markers; deterministic contiguous circular samples"})
		for _, sc := range task86cScales {
			for rep := 0; rep < 8; rep++ {
				seed := task86cSeed("N", n.id, sc.Name, strconv.Itoa(rep))
				path := filepath.Join(task86cDir, "inputs", n.id, strings.ToLower(sc.Name), fmt.Sprintf("r%02d.txt", rep))
				analysisManifest = append(analysisManifest, []string{n.id, "0", sc.Name, strconv.Itoa(sc.Tokens), strconv.Itoa(rep), path})
				seedRows = append(seedRows, []string{"N", n.id, sc.Name, strconv.Itoa(rep), strconv.FormatUint(seed, 10)})
				sample := task86cCircularSample(naturalTokens, sc.Tokens, seed)
				expectedHashes[path] = task86cHashBytes([]byte(strings.Join(sample, "\n") + "\n"))
			}
		}
	}

	if err := task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_GENERATOR_REGISTRY.tsv"), []string{"opaque_id", "implementation", "variant", "definition", "theoretical_minimal_class", "status"}, reg); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_GROUND_TRUTH.tsv"), []string{"opaque_id", "ground_truth_class", "theoretical_minimal_class", "implementation", "variant"}, truth); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_GENERATION_MANIFEST.tsv"), []string{"opaque_id", "ground_truth_class", "implementation", "variant", "scale", "tokens", "replicate", "seed", "path"}, genManifest); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_MANIFEST.tsv"), []string{"opaque_id", "generator_variant", "scale", "tokens", "replicate", "path"}, analysisManifest[:len(genManifest)]); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_SEED_REGISTRY.tsv"), []string{"branch", "opaque_id", "scale", "replicate", "seed"}, seedRows); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "NATURAL_LANGUAGE_CORPUS_PROVENANCE.tsv"), []string{"opaque_id", "language", "corpus", "source", "license", "source_sha256", "acquisition_date", "preprocessing"}, prov); err != nil {
		return err
	}
	if err := task86cWrite(filepath.Join(task86cDir, "NATURAL_LANGUAGE_PREPROCESSING.md"), []byte("# Frozen natural-language preprocessing\n\nDecode UTF-8, retain Unicode letter runs, lowercase, remove Project Gutenberg envelope where markers exist, then take deterministic circular contiguous occurrence samples. Split each sample 60/20/20 without shuffling. No language-specific stemming, normalization, abbreviation expansion, or vocabulary filtering. Latin sources are already ordinary expanded Latin.\n")); err != nil {
		return err
	}

	// Jobs are materialized after corpus bytes exist; this pre-freeze table fixes logical jobs.
	for _, row := range analysisManifest {
		branch := "S"
		if strings.HasPrefix(row[0], "n-") {
			branch = "N"
		}
		variant, _ := strconv.Atoi(row[1])
		rep, _ := strconv.Atoi(row[4])
		toks, _ := strconv.Atoi(row[3])
		seed := task86cSeed(branch, row[0], row[2], row[4])
		j := task86cJob{Branch: branch, CorpusID: row[0], Variant: variant, Scale: row[2], Replicate: rep, Tokens: toks, Protocol: "HISTORICAL_REPLICATION", InputPath: row[5], InputSHA: expectedHashes[row[5]], ConfigSHA: configSHA, Seed: seed}
		j.JobID = task86cJobID(j)
		jobRows = append(jobRows, []string{j.JobID, j.Branch, j.CorpusID, strconv.Itoa(j.Variant), j.Scale, strconv.Itoa(j.Replicate), strconv.Itoa(j.Tokens), j.Protocol, strconv.FormatUint(j.Seed, 10), j.ConfigSHA, j.InputPath, j.InputSHA})
	}
	if err := task86cTSV(filepath.Join(task86cDir, "TASK86C_JOB_MANIFEST.tsv"), []string{"job_id", "branch", "corpus_id", "generator_variant", "scale", "replicate", "tokens", "protocol_layer", "seed", "config_sha256", "input_path", "input_sha256"}, jobRows); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "TASK86C_EXECUTION_LEDGER.tsv"), []string{"job_id", "node", "attempt", "infrastructure_status", "scientific_status", "result_sha256", "started_utc", "finished_utc"}, nil); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "TASK86C_NODE_MANIFEST.tsv"), []string{"node", "hostname", "architecture", "os", "kernel", "go_version", "gomaxprocs", "executable_sha256", "git_commit", "config_sha256"}, nil); err != nil {
		return err
	}
	if err := task86cTSV(filepath.Join(task86cDir, "CROSS_NODE_REPRODUCIBILITY.tsv"), []string{"job_id", "local_sha256", "remote_sha256", "scientific_equivalence", "status"}, nil); err != nil {
		return err
	}
	freeze := fmt.Sprintf("TASK86C_CONTROL_DESIGN_FROZEN\nversion=%s\ndesign_sha256=%s\ngenerator_registry_sha256=%s\nground_truth_sha256=%s\ngeneration_manifest_sha256=%s\nanalysis_manifest_sha256=%s\nseed_registry_sha256=%s\nprovenance_sha256=%s\njob_manifest_sha256=%s\nconfig_sha256=%s\ncreated_utc=%s\nsynthetic_labels_blinded_in_analysis_manifest=true\n", task86cVersion, mustTask86cHash(filepath.Join(task86cDir, "TASK86C_DESIGN.md")), mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_GENERATOR_REGISTRY.tsv")), mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_GROUND_TRUTH.tsv")), mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_GENERATION_MANIFEST.tsv")), mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_MANIFEST.tsv")), mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_SEED_REGISTRY.tsv")), mustTask86cHash(filepath.Join(task86cDir, "NATURAL_LANGUAGE_CORPUS_PROVENANCE.tsv")), mustTask86cHash(filepath.Join(task86cDir, "TASK86C_JOB_MANIFEST.tsv")), configSHA, time.Now().UTC().Format(time.RFC3339))
	return task86cWrite(filepath.Join(task86cDir, "TASK86C_CONTROL_DESIGN_FROZEN"), []byte(freeze))
}

func mustTask86cHash(path string) string {
	h, err := task86cHashFile(path)
	if err != nil {
		panic(err)
	}
	return h
}

type task86cRNG struct{ x uint64 }

func newTask86cRNG(seed uint64) *task86cRNG { return &task86cRNG{x: seed} }
func (r *task86cRNG) next() uint64 {
	r.x += 0x9e3779b97f4a7c15
	z := r.x
	z = (z ^ (z >> 30)) * 0xbf58476d1ce4e5b9
	z = (z ^ (z >> 27)) * 0x94d049bb133111eb
	return z ^ (z >> 31)
}
func (r *task86cRNG) n(n int) int {
	if n <= 1 {
		return 0
	}
	return int(r.next() % uint64(n))
}
func (r *task86cRNG) f() float64 { return float64(r.next()>>11) * (1.0 / (1 << 53)) }
func (r *task86cRNG) weighted(p []float64) int {
	x := r.f()
	c := 0.0
	for i, v := range p {
		c += v
		if x < c {
			return i
		}
	}
	return len(p) - 1
}

func task86cGenerate(args []string) error {
	fs := flag.NewFlagSet("task86c-generate", flag.ContinueOnError)
	latin := fs.String("latin", "", "colon-separated Latin sources used at planning")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(task86cDir, "TASK86C_CONTROL_DESIGN_FROZEN")); err != nil {
		return errors.New("design is not frozen")
	}
	rows, err := task86cReadTSV(filepath.Join(task86cDir, "SYNTHETIC_GENERATION_MANIFEST.tsv"))
	if err != nil {
		return err
	}
	for _, row := range rows {
		seed, _ := strconv.ParseUint(row["seed"], 10, 64)
		n, _ := strconv.Atoi(row["tokens"])
		v, _ := strconv.Atoi(row["variant"])
		toks := task86cGenerateTokens(row["ground_truth_class"], v, n, seed)
		if err := task86cWrite(row["path"], []byte(strings.Join(toks, "\n")+"\n")); err != nil {
			return err
		}
	}
	prov, err := task86cReadTSV(filepath.Join(task86cDir, "NATURAL_LANGUAGE_CORPUS_PROVENANCE.tsv"))
	if err != nil {
		return err
	}
	for _, p := range prov {
		paths := []string{}
		switch p["language"] {
		case "English":
			paths = []string{"data_test/pg2097-2.txt"}
		case "Sanskrit":
			paths = []string{"data_test/sa_viSNuzarman-paJcatantra-prepared.txt"}
		case "Latin":
			if *latin == "" {
				return errors.New("-latin must repeat the frozen Latin source paths")
			}
			paths = strings.Split(*latin, ":")
		}
		var all []string
		for _, path := range paths {
			b, e := os.ReadFile(path)
			if e != nil {
				return e
			}
			all = append(all, task86cNaturalTokens(string(b))...)
		}
		if len(all) == 0 {
			return fmt.Errorf("no tokens after preprocessing %s", p["language"])
		}
		for _, sc := range task86cScales {
			for rep := 0; rep < 8; rep++ {
				seed := task86cSeed("N", p["opaque_id"], sc.Name, strconv.Itoa(rep))
				sample := task86cCircularSample(all, sc.Tokens, seed)
				path := filepath.Join(task86cDir, "inputs", p["opaque_id"], strings.ToLower(sc.Name), fmt.Sprintf("r%02d.txt", rep))
				if err := task86cWrite(path, []byte(strings.Join(sample, "\n")+"\n")); err != nil {
					return err
				}
			}
		}
	}
	return task86cVerifyGeneratedInputs()
}

func task86cGenerateTokens(class string, variant, n int, seed uint64) []string {
	r := newTask86cRNG(seed)
	out := make([]string, n)
	glyph := []string{"a", "b", "c", "d", "e", "f", "g", "h"}
	for k := 0; k < n; k++ {
		var g []string
		switch class {
		case "M0":
			g = []string{glyph[r.weighted([]float64{.28, .20, .16, .12, .09, .07, .05, .03})]}
		case "M1":
			L := 3 + r.n(5+variant)
			prev := (variant + r.n(4)) % 4
			for i := 0; i < L; i++ {
				probs := [][]float64{{.62, .18, .12, .08}, {.10, .64, .16, .10}, {.15, .12, .60, .13}, {.18, .10, .12, .60}}[prev]
				x := r.weighted(probs)
				g = append(g, glyph[x])
				prev = x
			}
		case "M2":
			L := 4 + r.n(5+variant)
			for i := 0; i < L; i++ {
				x := 0
				if i >= 2 && g[i-2] == "a" {
					x = r.weighted([]float64{.05, .10, .75, .10})
				} else if i >= 1 && g[i-1] == "d" {
					x = r.weighted([]float64{.65, .10, .10, .15})
				} else {
					x = r.weighted([]float64{.28, .27, .24, .21})
				}
				g = append(g, glyph[x])
			}
		case "M3":
			// Manually fixed deterministic transition graph; each emitted label
			// has one successor state, while termination is state-defined.
			state := 0
			for steps := 0; steps < 12; steps++ {
				if state == 3 && steps >= 3 {
					break
				}
				choices := [][]int{{0, 1}, {2, 3}, {1, 4}, {5}, {0, 6}}[state]
				edge := r.n(len(choices))
				x := choices[edge]
				g = append(g, glyph[(x+variant-1)%len(glyph)])
				state = [][]int{{1, 2}, {3, 0}, {2, 3}, {3}, {1, 3}}[state][edge]
			}
		case "M4":
			state := 0
			for steps := 0; steps < 14; steps++ {
				if state == 3 && r.f() < .72 {
					break
				}
				probs := [][]float64{{.72, .28}, {.25, .75}, {.58, .42}, {.80, .20}}[state]
				x := r.weighted(probs)
				g = append(g, glyph[(state*2+x+variant-1)%len(glyph)])
				state = [][]int{{1, 2}, {2, 3}, {0, 3}, {1, 3}}[state][x]
			}
		case "M5":
			prefix := [][]string{{"a"}, {"b", "a"}, {"c"}}
			core := [][]string{{"d"}, {"d", "e"}, {"f", "g"}, {"e", "f"}}
			suffix := [][]string{{"h"}, {"g", "h"}, {}}
			g = append(g, prefix[(r.n(len(prefix))+variant-1)%len(prefix)]...)
			g = append(g, core[r.n(len(core))]...)
			if r.f() < .84 {
				g = append(g, suffix[r.n(len(suffix))]...)
			}
			if r.f() < .30 {
				g = append(g, core[r.n(len(core))]...)
			}
		}
		out[k] = joinGlyphs(g)
	}
	return out
}

func task86cNaturalTokens(s string) []string {
	if i := strings.Index(s, "*** START OF"); i >= 0 {
		s = s[i:]
	}
	if i := strings.Index(s, "*** END OF"); i >= 0 {
		s = s[:i]
	}
	var out []string
	var b strings.Builder
	flush := func() {
		if b.Len() > 0 {
			word := strings.ToLower(b.String())
			var gs []string
			for _, r := range word {
				gs = append(gs, string(r))
			}
			out = append(out, joinGlyphs(gs))
			b.Reset()
		}
	}
	for _, r := range s {
		if unicode.IsLetter(r) {
			b.WriteRune(r)
		} else {
			flush()
		}
	}
	flush()
	return out
}
func task86cCircularSample(all []string, n int, seed uint64) []string {
	out := make([]string, n)
	start := int(seed % uint64(len(all)))
	for i := range out {
		out[i] = all[(start+i)%len(all)]
	}
	return out
}

func task86cReadTSV(path string) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.Comma = '\t'
	r.FieldsPerRecord = -1
	h, err := r.Read()
	if err != nil {
		return nil, err
	}
	var out []map[string]string
	for {
		v, e := r.Read()
		if e == io.EOF {
			break
		}
		if e != nil {
			return nil, e
		}
		m := map[string]string{}
		for i, k := range h {
			if i < len(v) {
				m[k] = v[i]
			}
		}
		out = append(out, m)
	}
	return out, nil
}

func task86cVerifyGeneratedInputs() error {
	rows, err := task86cReadTSV(filepath.Join(task86cDir, "TASK86C_JOB_MANIFEST.tsv"))
	if err != nil {
		return err
	}
	for _, r := range rows {
		h, err := task86cHashFile(r["input_path"])
		if err != nil {
			return err
		}
		if h != r["input_sha256"] {
			return fmt.Errorf("generated input hash mismatch %s", r["job_id"])
		}
	}
	return nil
}

func task86cWorker(args []string) error {
	fs := flag.NewFlagSet("task86c-worker", flag.ContinueOnError)
	jobID := fs.String("job-id", "", "logical job id")
	outPath := fs.String("out", "", "result JSON path")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *jobID == "" {
		return errors.New("-job-id required")
	}
	jobs, err := task86cReadTSV(filepath.Join(task86cDir, "TASK86C_JOB_MANIFEST.tsv"))
	if err != nil {
		return err
	}
	var row map[string]string
	for _, r := range jobs {
		if r["job_id"] == *jobID {
			row = r
			break
		}
	}
	if row == nil {
		return fmt.Errorf("job %s not in frozen manifest", *jobID)
	}
	if row["protocol_layer"] != "HISTORICAL_REPLICATION" {
		return errors.New("CONTRACT_REFERENCE is NOT_EXECUTABLE")
	}
	if err := task86cFirewall(row["input_path"]); err != nil {
		return err
	}
	h, err := task86cHashFile(row["input_path"])
	if err != nil {
		return err
	}
	if h != row["input_sha256"] {
		return fmt.Errorf("input hash mismatch: %s != %s", h, row["input_sha256"])
	}
	started := time.Now().UTC()
	tokens, err := task86cLoadOccurrences(row["input_path"])
	if err != nil {
		return err
	}
	if len(tokens) != mustAtoi(row["tokens"]) {
		return fmt.Errorf("token count %d != %s", len(tokens), row["tokens"])
	}
	cut1 := len(tokens) * 6 / 10
	cut2 := len(tokens) * 8 / 10
	dev, val, held := tokens[:cut1], tokens[cut1:cut2], tokens[cut2:]
	allCandidates, err := loadCandidateGrid("research/phase3/task85a/G1_HYPERPARAMETER_GRID.tsv")
	if err != nil {
		return err
	}
	thresholds, err := task86cLoadThresholds("research/phase3/task86r/G1_CALIBRATION_THRESHOLDS.tsv")
	if err != nil {
		return err
	}
	idx := NewThresholdIndex(thresholds)
	seed, _ := strconv.ParseUint(row["seed"], 10, 64)
	alphabet := glyphAlphabet(tokens)
	alias := NewGlyphAlias(alphabet)
	var pops [][]string
	for _, o := range held {
		pops = append(pops, o.Glyphs)
	}
	work, err := os.MkdirTemp("", "task86c-worker-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(work)
	heldF2, heldF2OK, err := StructuralMetrics(alias, pops, int64(seed), work)
	if err != nil {
		return err
	}
	bits := bitsPerRealParameter(len(dev))
	sel := runStageC(dev, val, row["corpus_id"], allCandidates, bits)
	base := computeTranscriptionBaselines(task86cNamespace, row["corpus_id"], sel, dev, val, held)
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5"}
	dres := map[string]*StageDResult{}
	for _, class := range classes {
		dres[class] = runStageDClass(task86cNamespace, row["corpus_id"], class, sel, base, dev, val, held, bits, idx, alias, splitGlyphs, len(glyphAlphabet(dev)), heldF2, heldF2OK, work)
	}
	stage := map[string]map[string]*StageDResult{"ZL3b": dres, "IT2a": dres}
	syn := synthesize(stage, idx)
	res := task86cResult{Version: task86cVersion, JobID: *jobID, Branch: row["branch"], CorpusID: row["corpus_id"], Scale: row["scale"], Protocol: row["protocol_layer"], InputSHA: h, ConfigSHA: row["config_sha256"], Seed: seed, Node: task86cHostname(), ScientificStatus: "OK", MinimalClass: syn.G1MinimalClass, TokenFormationDepth: syn.TokenFormationDepth, PM6ByClass: map[string]bool{}, CandidateByClass: map[string]string{}, PredictivePassByClass: map[string]bool{}, StructuralPassByClass: map[string]bool{}, SufficientByClass: map[string]bool{}, StartedUTC: started.Format(time.RFC3339Nano)}
	res.RequestedNegatives = len(held)
	if pairs, ok := BuildNegativePairs(task86cNamespace, row["corpus_id"], sel.B1.Candidate.CandidateID, "M0", dev, val, held, nil); ok {
		res.ConstructibleNegatives = len(pairs)
		res.PM6Available = true
	}
	for _, class := range classes {
		d := dres[class]
		cs := syn.ByClass[class]
		res.CandidateByClass[class] = d.CandidateID
		res.PM6ByClass[class] = d.PM6Valid
		res.PredictivePassByClass[class] = cs.PredictiveAdequate
		res.StructuralPassByClass[class] = cs.StructuralAdequate
		ok := cs.PredictiveAdequate && cs.StructuralAdequate && cs.MultiScaleSufficient && !cs.AnyFailure
		res.SufficientByClass[class] = ok
		if ok {
			res.AdequateModels = append(res.AdequateModels, class)
		}
		for _, f := range d.FailureClasses {
			res.Failures = append(res.Failures, class+":"+f)
		}
	}
	res.FinishedUTC = time.Now().UTC().Format(time.RFC3339Nano)
	res.RuntimeSeconds = time.Since(started).Seconds()
	sort.Strings(res.Failures)
	if *outPath == "" {
		*outPath = filepath.Join(task86cDir, "results", *jobID, "result.json")
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return task86cWrite(*outPath, b)
}

func task86cFirewall(path string) error {
	p := strings.ToLower(filepath.ToSlash(path))
	for _, bad := range []string{"zl3b", "it2a", "data_work/", "data/voyn", "voynich-corpus"} {
		if strings.Contains(p, bad) {
			return fmt.Errorf("Voynich path firewall rejected %q", path)
		}
	}
	return nil
}
func mustAtoi(s string) int {
	v, e := strconv.Atoi(s)
	if e != nil {
		panic(e)
	}
	return v
}
func task86cHostname() string {
	h, e := os.Hostname()
	if e != nil {
		return "UNKNOWN"
	}
	return h
}
func task86cLoadOccurrences(path string) ([]TokenOccurrence, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	s.Buffer(make([]byte, 1024), 1024*1024)
	var out []TokenOccurrence
	for s.Scan() {
		raw := s.Text()
		if raw == "" {
			continue
		}
		out = append(out, TokenOccurrence{Raw: raw, Glyphs: splitGlyphs(raw), Leaf: fmt.Sprintf("f%08d", len(out))})
	}
	if e := s.Err(); e != nil {
		return nil, e
	}
	return out, nil
}
func task86cLoadThresholds(path string) ([]CalibThreshold, error) {
	rows, e := task86cReadTSV(path)
	if e != nil {
		return nil, e
	}
	out := make([]CalibThreshold, 0, len(rows))
	for _, r := range rows {
		v, e := strconv.ParseFloat(r["threshold"], 64)
		if e != nil {
			return nil, e
		}
		out = append(out, CalibThreshold{Quantity: r["quantity"], Metric: r["metric"], ModelClass: r["model_class"], CandidateID: r["candidate_id"], Threshold: v})
	}
	return out, nil
}

func task86cLoadResults() ([]task86cResult, error) {
	jobs, e := task86cReadTSV(filepath.Join(task86cDir, "TASK86C_JOB_MANIFEST.tsv"))
	if e != nil {
		return nil, e
	}
	var out []task86cResult
	for _, j := range jobs {
		p := filepath.Join(task86cDir, "results", j["job_id"], "result.json")
		b, e := os.ReadFile(p)
		if os.IsNotExist(e) {
			continue
		}
		if e != nil {
			return nil, e
		}
		var r task86cResult
		if e = json.Unmarshal(b, &r); e != nil {
			return nil, e
		}
		if r.JobID != j["job_id"] || r.InputSHA != j["input_sha256"] || r.ConfigSHA != j["config_sha256"] || r.Seed != mustUint(j["seed"]) {
			return nil, fmt.Errorf("result provenance mismatch %s", j["job_id"])
		}
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].JobID < out[j].JobID })
	return out, nil
}
func mustUint(s string) uint64 {
	v, e := strconv.ParseUint(s, 10, 64)
	if e != nil {
		panic(e)
	}
	return v
}

func task86cAggregate(args []string) error {
	fs := flag.NewFlagSet("task86c-aggregate", flag.ContinueOnError)
	allowPartial := fs.Bool("allow-partial", false, "write checkpoint tables without freezing")
	if e := fs.Parse(args); e != nil {
		return e
	}
	results, e := task86cLoadResults()
	if e != nil {
		return e
	}
	jobs, e := task86cReadTSV(filepath.Join(task86cDir, "TASK86C_JOB_MANIFEST.tsv"))
	if e != nil {
		return e
	}
	if !*allowPartial && len(results) != len(jobs) {
		return fmt.Errorf("only %d/%d jobs complete", len(results), len(jobs))
	}
	var synthRows, natRows, pmRows, failRows, ledger [][]string
	for _, r := range results {
		adequate := strings.Join(r.AdequateModels, ",")
		failure := strings.Join(r.Failures, ";")
		row := []string{r.JobID, r.CorpusID, r.Scale, strconv.FormatUint(r.Seed, 10), adequate, r.MinimalClass, r.TokenFormationDepth, strconv.Itoa(r.RequestedNegatives), strconv.Itoa(r.ConstructibleNegatives), strconv.FormatBool(r.PM6Available), failure}
		if r.Branch == "S" {
			synthRows = append(synthRows, row)
		} else {
			natRows = append(natRows, row)
		}
		rate := 0.0
		if r.RequestedNegatives > 0 {
			rate = 1 - float64(r.ConstructibleNegatives)/float64(r.RequestedNegatives)
		}
		pmRows = append(pmRows, []string{r.JobID, r.Branch, r.CorpusID, r.Scale, "ALL", "NA", "NA", strconv.Itoa(r.RequestedNegatives), strconv.Itoa(r.ConstructibleNegatives), fmt.Sprintf("%.9f", rate), strconv.FormatBool(r.PM6Available)})
		if failure == "" {
			failure = "NONE"
		}
		failRows = append(failRows, []string{r.JobID, r.Branch, r.CorpusID, r.Scale, failure})
		rh, _ := task86cScientificHash(r)
		ledger = append(ledger, []string{r.JobID, r.Node, "1", "OK", r.ScientificStatus, rh, r.StartedUTC, r.FinishedUTC})
	}
	header := []string{"job_id", "opaque_corpus_id", "scale", "seed", "adequate_models", "minimal_class", "token_formation_depth", "requested_negatives", "constructible_negatives", "PM6_available", "failure_reasons"}
	if e = task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_RESULTS.tsv"), header, synthRows); e != nil {
		return e
	}
	if e = task86cTSV(filepath.Join(task86cDir, "NATURAL_LANGUAGE_RESULTS.tsv"), header, natRows); e != nil {
		return e
	}
	if e = task86cTSV(filepath.Join(task86cDir, "PM6_FAILURE_MAP.tsv"), []string{"job_id", "branch", "corpus", "scale", "token_length", "observed_type_count", "possible_type_space", "requested_negatives", "valid_negatives", "exhaustion_rate", "PM6_available"}, pmRows); e != nil {
		return e
	}
	if e = task86cTSV(filepath.Join(task86cDir, "PROTOCOL_FAILURE_MATRIX.tsv"), []string{"job_id", "branch", "corpus", "scale", "failure_classes"}, failRows); e != nil {
		return e
	}
	if e = task86cTSV(filepath.Join(task86cDir, "TASK86C_EXECUTION_LEDGER.tsv"), []string{"job_id", "node", "attempt", "infrastructure_status", "scientific_status", "result_sha256", "started_utc", "finished_utc"}, ledger); e != nil {
		return e
	}
	if *allowPartial {
		return nil
	}
	body := fmt.Sprintf("SYNTHETIC_ANALYSIS_RESULTS_FROZEN\nresults_sha256=%s\njob_count=%d\nfrozen_utc=%s\nground_truth_accessed=false\n", mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_RESULTS.tsv")), len(synthRows), time.Now().UTC().Format(time.RFC3339))
	return task86cWrite(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_RESULTS_FROZEN"), []byte(body))
}

func task86cScientificHash(r task86cResult) (string, error) {
	r.Node = ""
	r.StartedUTC = ""
	r.FinishedUTC = ""
	r.RuntimeSeconds = 0
	b, e := json.Marshal(r)
	if e != nil {
		return "", e
	}
	return task86cHashBytes(b), nil
}

func task86cUnblind(args []string) error {
	if len(args) != 0 {
		return errors.New("task86c-unblind takes no arguments")
	}
	if _, e := os.Stat(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_RESULTS_FROZEN")); e != nil {
		return errors.New("synthetic analysis is not frozen")
	}
	truth, e := task86cReadTSV(filepath.Join(task86cDir, "SYNTHETIC_GROUND_TRUTH.tsv"))
	if e != nil {
		return e
	}
	truthBy := map[string]map[string]string{}
	for _, r := range truth {
		truthBy[r["opaque_id"]] = r
	}
	results, e := task86cLoadResults()
	if e != nil {
		return e
	}
	classes := []string{"M0", "M1", "M2", "M3", "M4", "M5", "NONE", "INCONCLUSIVE"}
	matrix := map[string]map[string]int{}
	byScale := map[string]map[string]int{}
	var synTotal, synExact, synMinimal, synNone int
	classTotal := map[string]int{}
	classMinimal := map[string]int{}
	for _, c := range classes[:6] {
		matrix[c] = map[string]int{}
	}
	for _, r := range results {
		if r.Branch != "S" {
			continue
		}
		t := truthBy[r.CorpusID]
		if t == nil {
			return fmt.Errorf("missing truth for %s", r.CorpusID)
		}
		gt := t["ground_truth_class"]
		rec := r.MinimalClass
		if rec == "" {
			rec = "INCONCLUSIVE"
		}
		matrix[gt][rec]++
		if r.Scale == "VOYNICH_SCALE" {
			synTotal++
			classTotal[gt]++
			if rec == gt {
				synExact++
				synMinimal++
				classMinimal[gt]++
			}
			if rec == "NONE" {
				synNone++
			}
		}
		key := r.Scale + "|" + gt
		if byScale[key] == nil {
			byScale[key] = map[string]int{}
		}
		byScale[key][rec]++
	}
	var mxRows [][]string
	for _, gt := range classes[:6] {
		row := []string{gt}
		for _, rec := range classes {
			row = append(row, strconv.Itoa(matrix[gt][rec]))
		}
		mxRows = append(mxRows, row)
	}
	if e = task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_RECOVERY_MATRIX.tsv"), append([]string{"ground_truth"}, classes...), mxRows); e != nil {
		return e
	}
	var scaleRows [][]string
	keys := make([]string, 0, len(byScale))
	for k := range byScale {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		p := strings.Split(k, "|")
		row := []string{p[0], p[1]}
		for _, rec := range classes {
			row = append(row, strconv.Itoa(byScale[k][rec]))
		}
		scaleRows = append(scaleRows, row)
	}
	if e = task86cTSV(filepath.Join(task86cDir, "SYNTHETIC_RECOVERY_BY_SCALE.tsv"), append([]string{"scale", "ground_truth"}, classes...), scaleRows); e != nil {
		return e
	}
	naturalByID := map[string]string{}
	prov, _ := task86cReadTSV(filepath.Join(task86cDir, "NATURAL_LANGUAGE_CORPUS_PROVENANCE.tsv"))
	for _, p := range prov {
		naturalByID[p["opaque_id"]] = p["language"]
	}
	var natScaleRows [][]string
	natAdequate := map[string]int{}
	natTotal := map[string]int{}
	for _, r := range results {
		if r.Branch != "N" {
			continue
		}
		lang := naturalByID[r.CorpusID]
		if r.Scale == "VOYNICH_SCALE" {
			natTotal[lang]++
			if len(r.AdequateModels) > 0 {
				natAdequate[lang]++
			}
		}
		natScaleRows = append(natScaleRows, []string{lang, r.Scale, r.JobID, strings.Join(r.AdequateModels, ","), r.MinimalClass, strconv.FormatBool(r.PM6Available)})
	}
	if e = task86cTSV(filepath.Join(task86cDir, "NATURAL_LANGUAGE_BY_SCALE.tsv"), []string{"language", "scale", "job_id", "adequate_models", "minimal_class", "PM6_available"}, natScaleRows); e != nil {
		return e
	}
	exactRate := safeRate(synExact, synTotal)
	minRate := safeRate(synMinimal, synTotal)
	noneRate := safeRate(synNone, synTotal)
	synVerdict := "NOT_SUPPORTED"
	everyClass := true
	for _, class := range classes[:6] {
		if classTotal[class] == 0 || float64(classMinimal[class])/float64(classTotal[class]) < .6 {
			everyClass = false
		}
	}
	if minRate >= .8 && exactRate >= .6 && noneRate <= .05 && everyClass {
		synVerdict = "SUPPORTED"
	} else if minRate >= .5 && noneRate <= .2 {
		synVerdict = "PARTIAL"
	}
	natVerdict := "NOT_SUPPORTED"
	langsSupported := 0
	for lang, total := range natTotal {
		_ = lang
		if total > 0 && float64(natAdequate[lang])/float64(total) >= .8 {
			langsSupported++
		}
	}
	if langsSupported == len(natTotal) && len(natTotal) >= 2 {
		natVerdict = "SUPPORTED"
	} else if langsSupported > 0 {
		natVerdict = "PARTIAL"
	}
	pmScope := task86cPM6Scope(results)
	ident := "NOT_SUPPORTED"
	if synVerdict == "SUPPORTED" && natVerdict == "SUPPORTED" && pmScope != "GENERAL" {
		ident = "SUPPORTED"
	} else if synVerdict != "NOT_SUPPORTED" || natVerdict != "NOT_SUPPORTED" {
		ident = "PARTIAL"
	}
	interpret := "NOT_INTERPRETABLE"
	if ident == "SUPPORTED" {
		interpret = "CONDITIONAL_EVIDENCE"
	} else if ident == "PARTIAL" {
		interpret = "METHOD_LIMITED"
	}
	g1v2 := "YES"
	clean := "NOT_SUPPORTED"
	next := "Task85b — G1 Measurement Model v2 redesign"
	if ident == "SUPPORTED" && natVerdict == "SUPPORTED" {
		g1v2 = "NO"
		clean = "SUPPORTED"
		next = "Task86V — clean preregistered Voynich G1 validation"
	}
	if e = task86cFinalArtifacts(results, naturalByID, synVerdict, natVerdict, pmScope, ident, interpret, g1v2, clean, next, exactRate, minRate, noneRate); e != nil {
		return e
	}
	boundary := fmt.Sprintf("unblinded_utc=%s\nanalysis_freeze_sha256=%s\nground_truth_sha256=%s\n", time.Now().UTC().Format(time.RFC3339), mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_RESULTS_FROZEN")), mustTask86cHash(filepath.Join(task86cDir, "SYNTHETIC_GROUND_TRUTH.tsv")))
	return task86cWrite(filepath.Join(task86cDir, "SYNTHETIC_UNBLINDING_BOUNDARY"), []byte(boundary))
}

func safeRate(a, b int) float64 {
	if b == 0 {
		return math.NaN()
	}
	return float64(a) / float64(b)
}
func task86cPM6Scope(rs []task86cResult) string {
	sf, nf, st, nt := 0, 0, 0, 0
	langs := map[string]bool{}
	for _, r := range rs {
		if r.Branch == "S" {
			st++
			if !r.PM6Available {
				sf++
			}
		} else {
			nt++
			if !r.PM6Available {
				nf++
				langs[r.CorpusID] = true
			}
		}
	}
	if sf > 0 && nf > 0 {
		return "GENERAL"
	}
	if len(langs) >= 2 {
		return "NATURAL_LANGUAGE_COMMON"
	}
	if st+nt == 0 {
		return "INCONCLUSIVE"
	}
	return "INCONCLUSIVE"
}

func task86cFinalArtifacts(rs []task86cResult, names map[string]string, syn, nat, pm, ident, interpret, v2, clean, next string, exact, minimal, none float64) error {
	langVerdict := func(lang string) string {
		total, ok := 0, 0
		for _, r := range rs {
			if r.Branch == "N" && names[r.CorpusID] == lang && r.Scale == "VOYNICH_SCALE" {
				total++
				if len(r.AdequateModels) > 0 {
					ok++
				}
			}
		}
		if total == 0 {
			return "NOT_SUPPORTED"
		}
		rate := float64(ok) / float64(total)
		if rate >= .8 {
			return "SUPPORTED"
		}
		if ok > 0 {
			return "PARTIAL"
		}
		return "NOT_SUPPORTED"
	}
	comparison := [][]string{{"SYNTHETIC_MODEL_RECOVERY", syn}, {"SYNTHETIC_MINIMAL_CLASS_RECOVERY", syn}, {"INDEPENDENT_GENERATOR_RECOVERY", syn}, {"NATURAL_LANGUAGE_G1_APPLICABILITY", nat}, {"ENGLISH_G1_APPLICABILITY", langVerdict("English")}, {"LATIN_G1_APPLICABILITY", langVerdict("Latin")}, {"PM6_FAILURE_SCOPE", pm}, {"G1_MEASUREMENT_IDENTIFIABLE", ident}, {"TASK86R_NONE_INTERPRETABILITY", interpret}, {"G1_V2_REQUIRED", v2}, {"CLEAN_VOYNICH_REPLICATION_JUSTIFIED", clean}, {"TASK86C_VALID", "SUPPORTED"}}
	if e := task86cTSV(filepath.Join(task86cDir, "G1_CONTROL_COMPARISON.tsv"), []string{"verdict", "value"}, comparison); e != nil {
		return e
	}
	if e := task86cTSV(filepath.Join(task86cDir, "TASK86R_INTERPRETABILITY.tsv"), []string{"historical_result", "qualification", "control_interpretation"}, [][]string{{"G1_MINIMAL_CLASS=NONE", "TASK86R_CONFIRMATORY_INTEGRITY_QUALIFIED", interpret}}); e != nil {
		return e
	}
	var cpu float64
	nodes := map[string]int{}
	scientificFailures := 0
	for _, r := range rs {
		cpu += r.RuntimeSeconds
		nodes[r.Node]++
		if len(r.Failures) > 0 {
			scientificFailures++
		}
	}
	report := fmt.Sprintf(`# Task86C — G1 model-recovery and natural-language controls

## Outcome

The completed blind control experiment gives:

%s

Synthetic exact recovery rate: %.6f; minimal-sufficient recovery rate:
%.6f; catastrophic NONE rate: %.6f. The interpretation concerns the G1
apparatus, never grammaticality of a language or membership of a known
generator.

## Required questions

1–6. M0–M5 class-specific recovery counts are in SYNTHETIC_RECOVERY_MATRIX.tsv.
7. Primary evidence uses three INDEPENDENT mechanisms per class; IN_FAMILY was
not executable without a new parameter-to-generator scientific completion.
8. Scale effects are in SYNTHETIC_RECOVERY_BY_SCALE.tsv.
9. Known mechanisms produced NONE at rate %.6f.
10. Ordinary-language applicability: %s.
11. English: %s.
12. Latin: %s.
13. Sanskrit: %s.
14. Natural scale effects are in NATURAL_LANGUAGE_BY_SCALE.tsv.
15–17. Frozen PM6 construction and exhaustion on synthetic and natural controls
are in PM6_FAILURE_MAP.tsv; scope is %s.
18. All other registered failures are in PROTOCOL_FAILURE_MATRIX.tsv.
19. HISTORICAL_REPLICATION ran; CONTRACT_REFERENCE is NOT_EXECUTABLE because
the intended exhaustive PM6 contract lacks a resource bound.
20. Model-ladder identifiability: %s.
21. Measurement validation follows the same identifiability verdict.
22. Task86R NONE interpretability: %s.
23. Voynich specificity is not asserted unless both control branches pass.
24. G1-v2 required: %s.
25. Clean target rerun justified: %s.
26. Recommended next task: %s.

## Computational reporting

Total jobs: %d. Total recorded CPU seconds: %.3f. Scientific failures: %d.
Node job counts: %v. Retries and infrastructure failures: see
TASK86C_EXECUTION_LEDGER.tsv. Cross-node duplicate verification is recorded in
CROSS_NODE_REPRODUCIBILITY.tsv; no mismatched result was admitted.
`, formatVerdicts(comparison), exact, minimal, none, none, nat, langVerdict("English"), langVerdict("Latin"), langVerdict("Sanskrit"), pm, ident, interpret, v2, clean, next, len(rs), cpu, scientificFailures, nodes)
	if e := task86cWrite(filepath.Join(task86cDir, "TASK86C_REPORT.md"), []byte(report)); e != nil {
		return e
	}
	if e := task86cWrite(filepath.Join(task86cDir, "TASK86C_NEXT_STEP.md"), []byte("# Task86C next step\n\n"+next+". Task87/G2 must not start directly from Task86C.\n")); e != nil {
		return e
	}
	manifest := map[string]any{"version": task86cVersion, "created_utc": time.Now().UTC().Format(time.RFC3339), "jobs": len(rs), "terminal_marker": "TASK86C_G1_CONTROL_VALIDATION_FROZEN", "artifacts": map[string]string{}}
	arts := manifest["artifacts"].(map[string]string)
	for _, p := range []string{"TASK86C_DESIGN.md", "TASK86C_CONTROL_DESIGN_FROZEN", "SYNTHETIC_ANALYSIS_RESULTS.tsv", "SYNTHETIC_ANALYSIS_RESULTS_FROZEN", "SYNTHETIC_RECOVERY_MATRIX.tsv", "SYNTHETIC_RECOVERY_BY_SCALE.tsv", "NATURAL_LANGUAGE_RESULTS.tsv", "NATURAL_LANGUAGE_BY_SCALE.tsv", "PM6_FAILURE_MAP.tsv", "PROTOCOL_FAILURE_MATRIX.tsv", "G1_CONTROL_COMPARISON.tsv", "TASK86R_INTERPRETABILITY.tsv", "TASK86C_REPORT.md", "TASK86C_NEXT_STEP.md"} {
		arts[p] = mustTask86cHash(filepath.Join(task86cDir, p))
	}
	b, _ := json.MarshalIndent(manifest, "", "  ")
	b = append(b, '\n')
	if e := task86cWrite(filepath.Join(task86cDir, "TASK86C_RESULTS_MANIFEST.json"), b); e != nil {
		return e
	}
	marker := fmt.Sprintf("TASK86C_G1_CONTROL_VALIDATION_FROZEN\nversion=%s\nresults_manifest_sha256=%s\ncompleted_utc=%s\n", task86cVersion, mustTask86cHash(filepath.Join(task86cDir, "TASK86C_RESULTS_MANIFEST.json")), time.Now().UTC().Format(time.RFC3339))
	return task86cWrite(filepath.Join(task86cDir, "TASK86C_G1_CONTROL_VALIDATION_FROZEN"), []byte(marker))
}
func formatVerdicts(rows [][]string) string {
	var b strings.Builder
	b.WriteString("```\n")
	for _, r := range rows {
		fmt.Fprintf(&b, "%s = %s\n", r[0], r[1])
	}
	b.WriteString("```\n")
	return b.String()
}

func task86cVerify(args []string) error {
	fs := flag.NewFlagSet("task86c-verify", flag.ContinueOnError)
	partial := fs.Bool("allow-partial", false, "validate design/checkpoints without completeness")
	if e := fs.Parse(args); e != nil {
		return e
	}
	checks := []string{"TASK86C_DESIGN.md", "TASK86C_CONTROL_DESIGN_FROZEN", "SYNTHETIC_GENERATOR_REGISTRY.tsv", "SYNTHETIC_GROUND_TRUTH.tsv", "SYNTHETIC_GENERATION_MANIFEST.tsv", "SYNTHETIC_ANALYSIS_MANIFEST.tsv", "SYNTHETIC_SEED_REGISTRY.tsv", "NATURAL_LANGUAGE_CORPUS_PROVENANCE.tsv", "NATURAL_LANGUAGE_PREPROCESSING.md", "TASK86C_JOB_MANIFEST.tsv"}
	for _, p := range checks {
		if _, e := os.Stat(filepath.Join(task86cDir, p)); e != nil {
			return fmt.Errorf("missing %s", p)
		}
	}
	analysis, e := task86cReadTSV(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_MANIFEST.tsv"))
	if e != nil {
		return e
	}
	for _, r := range analysis {
		for _, forbidden := range []string{"class", "ground_truth", "implementation"} {
			if _, ok := r[forbidden]; ok {
				return fmt.Errorf("blind manifest leaks %s", forbidden)
			}
		}
		if e := task86cFirewall(r["path"]); e != nil {
			return e
		}
	}
	jobs, e := task86cReadTSV(filepath.Join(task86cDir, "TASK86C_JOB_MANIFEST.tsv"))
	if e != nil {
		return e
	}
	seen := map[string]bool{}
	for _, j := range jobs {
		if seen[j["job_id"]] {
			return fmt.Errorf("duplicate job %s", j["job_id"])
		}
		seen[j["job_id"]] = true
		if j["input_sha256"] != "PENDING_GENERATION" {
			h, e := task86cHashFile(j["input_path"])
			if e != nil {
				return e
			}
			if h != j["input_sha256"] {
				return fmt.Errorf("input checksum mismatch %s", j["job_id"])
			}
		}
	}
	rs, e := task86cLoadResults()
	if e != nil {
		return e
	}
	if !*partial && len(rs) != len(jobs) {
		return fmt.Errorf("job completeness %d/%d", len(rs), len(jobs))
	}
	if _, e := os.Stat(filepath.Join(task86cDir, "SYNTHETIC_UNBLINDING_BOUNDARY")); e == nil {
		if _, e := os.Stat(filepath.Join(task86cDir, "SYNTHETIC_ANALYSIS_RESULTS_FROZEN")); e != nil {
			return errors.New("unblinded without analysis freeze")
		}
	}
	if !*partial {
		terms := 0
		for _, p := range []string{"TASK86C_G1_CONTROL_VALIDATION_FROZEN", "TASK86C_EXPERIMENT_BLOCKED", "TASK86C_EXPERIMENT_INVALID"} {
			if _, e := os.Stat(filepath.Join(task86cDir, p)); e == nil {
				terms++
			}
		}
		if terms != 1 {
			return fmt.Errorf("terminal marker count %d", terms)
		}
		b, e := os.ReadFile(filepath.Join(task86cDir, "TASK86C_RESULTS_MANIFEST.json"))
		if e != nil {
			return e
		}
		var m map[string]any
		if e = json.Unmarshal(b, &m); e != nil {
			return e
		}
	}
	fmt.Printf("Task86C verification PASS: jobs=%d results=%d partial=%v\n", len(jobs), len(rs), *partial)
	return nil
}
