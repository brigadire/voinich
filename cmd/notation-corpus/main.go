package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"zcore.dev/voinich/internal/notation"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "notation-corpus:", err)
		os.Exit(1)
	}
}
func run(args []string) error {
	if len(args) == 0 {
		return usage()
	}
	switch args[0] {
	case "acquire":
		return acquire(args[1:])
	case "normalize":
		return normalize(args[1:])
	case "validate":
		return validate(args[1:])
	case "analyze":
		return analyze(args[1:])
	case "compare-vm":
		return compareVM(args[1:])
	case "report":
		return report(args[1:])
	case "compare-classes":
		return compareClasses(args[1:])
	case "vm-adapter":
		return vmAdapter(args[1:])
	case "rarefy":
		return rarefy(args[1:])
	case "bootstrap":
		return bootstrapCmd(args[1:])
	case "distributions":
		return distributionsCmd(args[1:])
	case "metric-output-types":
		return metricOutputTypesCmd(args[1:])
	case "calibrate":
		return calibrateCmd(args[1:])
	case "vm-reference":
		return vmReferenceCmd(args[1:])
	case "production-preflight":
		return productionPreflightCmd(args[1:])
	case "bdd-usc":
		return bddUSCCmd(args[1:])
	case "musicxml-usc":
		return musicXMLUSCCmd(args[1:])
	case "production-corpus-validate":
		return productionCorpusValidateCmd(args[1:])
	case "production-run-preflight":
		return productionRunPreflightCmd(args[1:])
	case "global-freeze-verify":
		return globalFreezeVerifyCmd(args[1:])
	case "global-freeze-bind":
		return globalFreezeBindCmd(args[1:])
	case "production-run-execute":
		return productionRunExecuteCmd(args[1:])
	default:
		return usage()
	}
}
func usage() error {
	return fmt.Errorf("usage: notation-corpus acquire|normalize|validate|analyze|compare-vm|report|compare-classes|vm-adapter|rarefy|bootstrap|distributions|metric-output-types|calibrate|production-preflight|bdd-usc|musicxml-usc|production-corpus-validate|production-run-preflight|global-freeze-verify|global-freeze-bind|production-run-execute")
}

// guardNotUnauthorizedCandidate refuses any corpus_id in the ten frozen
// class prefixes C01-C09 unless --fixture confirms a small manually
// inspectable fixture (B03/B04/B01/B02 preparation may only ever touch VM,
// calibration, or fixture data; production candidates remain blocked).
func guardNotUnauthorizedCandidate(corpusID string, fixture bool) error {
	if fixture {
		return nil
	}
	for i := 1; i <= 9; i++ {
		if strings.HasPrefix(corpusID, fmt.Sprintf("C%02d", i)) {
			return fmt.Errorf("preparation lock: corpus_id %q looks like a C01-C09 candidate; production candidate runs are not authorized (pass --fixture only for a manually inspectable fixture)", corpusID)
		}
	}
	return nil
}
func oneFile(name string) (*os.File, error) { return os.Open(name) }
func createNew(name string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(name), 0755); err != nil {
		return nil, err
	}
	return os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
}

func acquire(args []string) error {
	fs := flag.NewFlagSet("acquire", flag.ContinueOnError)
	src := fs.String("source-file", "", "local source file")
	root := fs.String("corpus-dir", "", "corpus directory")
	url := fs.String("source-url", "", "canonical source URL")
	license := fs.String("license", "", "verified license statement")
	version := fs.String("version", "", "source revision/version")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *src == "" || *root == "" || *url == "" || *license == "" || *version == "" {
		return fmt.Errorf("acquire requires --source-file, --corpus-dir, --source-url, --license, and --version")
	}
	in, err := os.Open(*src)
	if err != nil {
		return err
	}
	defer in.Close()
	rawDir := filepath.Join(*root, "raw")
	if err := os.MkdirAll(rawDir, 0755); err != nil {
		return err
	}
	dstPath := filepath.Join(rawDir, filepath.Base(*src))
	out, err := createNew(dstPath)
	if err != nil {
		return fmt.Errorf("immutable raw target: %w", err)
	}
	h := sha256.New()
	_, copyErr := io.Copy(io.MultiWriter(out, h), in)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	prov := map[string]any{"schema_version": "notation-source-provenance-1.0", "source_url": *url, "retrieved_at": time.Now().UTC().Format(time.RFC3339), "version": *version, "license": *license, "files": []map[string]any{{"original_filename": filepath.Base(*src), "raw_path": dstPath, "sha256": hex.EncodeToString(h.Sum(nil))}}, "immutable": true}
	p, err := createNew(filepath.Join(*root, "SOURCE_PROVENANCE.json"))
	if err != nil {
		return err
	}
	defer p.Close()
	e := json.NewEncoder(p)
	e.SetIndent("", "  ")
	return e.Encode(prov)
}
func normalize(args []string) error {
	fs := flag.NewFlagSet("normalize", flag.ContinueOnError)
	inName := fs.String("input", "", "fixture/source JSON")
	outName := fs.String("output", "", "USC JSONL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *inName == "" || *outName == "" {
		return fmt.Errorf("normalize requires --input and --output")
	}
	in, err := oneFile(*inName)
	if err != nil {
		return err
	}
	defer in.Close()
	var src notation.SourceDocument
	if err := json.NewDecoder(in).Decode(&src); err != nil {
		return err
	}
	rs, err := notation.NormalizeFixture(src)
	if err != nil {
		return err
	}
	out, err := createNew(*outName)
	if err != nil {
		return err
	}
	defer out.Close()
	return notation.WriteJSONL(out, rs)
}
func validate(args []string) error {
	fs := flag.NewFlagSet("validate", flag.ContinueOnError)
	inName := fs.String("input", "", "USC JSONL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	in, err := oneFile(*inName)
	if err != nil {
		return err
	}
	defer in.Close()
	rs, err := notation.ReadJSONL(in)
	if err != nil {
		return err
	}
	if err := notation.Validate(rs); err != nil {
		return err
	}
	fmt.Printf("VALID records=%d corpus=%s representation=%s\n", len(rs), rs[0].CorpusID, rs[0].Representation)
	return nil
}
func analyze(args []string) error {
	fs := flag.NewFlagSet("analyze", flag.ContinueOnError)
	inName := fs.String("input", "", "USC JSONL")
	outName := fs.String("output", "", "fingerprint JSON")
	manifestName := fs.String("manifest", "", "RUN_MANIFEST.json")
	fixture := fs.Bool("fixture", false, "confirm fixture/synthetic-only preparation run")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *inName == "" || *outName == "" || *manifestName == "" {
		return fmt.Errorf("analyze requires --input, --output, and --manifest")
	}
	if !*fixture {
		return fmt.Errorf("preparation lock: analyze requires --fixture; production candidate analysis is not authorized")
	}
	in, err := oneFile(*inName)
	if err != nil {
		return err
	}
	defer in.Close()
	rs, err := notation.ReadJSONL(in)
	if err != nil {
		return err
	}
	fp, err := notation.Analyze(rs)
	if err != nil {
		return err
	}
	var payload bytes.Buffer
	if err := notation.WriteFingerprintJSON(&payload, fp); err != nil {
		return err
	}
	out, err := createNew(*outName)
	if err != nil {
		return err
	}
	if _, err := out.Write(payload.Bytes()); err != nil {
		out.Close()
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	outHash := sha256.Sum256(payload.Bytes())
	manifest := map[string]any{"schema_version": "notation-run-manifest-1.0", "source_sha256": fp.InputSHA256, "corpus_id": fp.CorpusID, "representation_id": fp.Representation, "adapter_version": "usc-adapter-1.0", "analyzer_version": "notation-analyzer-1.0", "metric_registry_version": "generic-metrics-1.0", "parameters": map[string]any{"rarefaction_checkpoints": notation.RarefactionCheckpoints, "fixture_only": true}, "random_seed": 20260830, "command": append([]string{"notation-corpus", "analyze"}, args...), "started_at_utc": time.Now().UTC().Format(time.RFC3339), "outputs": map[string]string{*outName: hex.EncodeToString(outHash[:])}}
	mf, err := createNew(*manifestName)
	if err != nil {
		return err
	}
	defer mf.Close()
	enc := json.NewEncoder(mf)
	enc.SetIndent("", "  ")
	return enc.Encode(manifest)
}
func compareVM(args []string) error {
	fs := flag.NewFlagSet("compare-vm", flag.ContinueOnError)
	candidate := fs.String("candidate", "", "candidate fingerprint")
	reference := fs.String("vm-reference", "", "frozen VM fingerprint (VM_REFERENCE_V2.fingerprint.json)")
	manifestPath := fs.String("vm-manifest", "", "VM_REFERENCE_V2_MANIFEST.json (required: --vm-reference is hash-verified against it)")
	scale := fs.String("scale", "", "frozen CALIBRATION_SCALES.tsv")
	checkpoint := fs.Int("checkpoint", 39380, "calibration checkpoint stratum to compare at")
	output := fs.String("output", "", "VM_COMPARISON.tsv")
	authorize := fs.Bool("authorize-production", false, "future production authorization")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*authorize {
		return fmt.Errorf("preparation lock: VM comparison is not authorized")
	}
	if !productionAuthorized() {
		return fmt.Errorf("repository production authorization remains false")
	}
	if *manifestPath == "" {
		return fmt.Errorf("compare-vm requires --vm-manifest: a candidate-specific VM reference file is not authorized (adversarial test A4)")
	}
	cf, err := readFP(*candidate)
	if err != nil {
		return err
	}
	refRaw, err := os.ReadFile(*reference)
	if err != nil {
		return err
	}
	mb, err := os.ReadFile(*manifestPath)
	if err != nil {
		return err
	}
	var manifest notation.VMReferenceManifest
	if err := json.Unmarshal(mb, &manifest); err != nil {
		return err
	}
	if err := notation.VerifyFrozenVMReference(refRaw, manifest); err != nil {
		return err
	}
	rf, err := notation.ReadFingerprintJSON(bytes.NewReader(refRaw))
	if err != nil {
		return err
	}
	sf, err := os.Open(*scale)
	if err != nil {
		return err
	}
	rawScales, err := notation.ReadCalibrationScalesTSV(sf)
	sf.Close()
	if err != nil {
		return err
	}
	sc := notation.ScalesFromCalibration(rawScales, *checkpoint)
	rows, fam, err := notation.Compare(cf, rf, sc)
	if err != nil {
		return err
	}
	out, err := createNew(*output)
	if err != nil {
		return err
	}
	defer out.Close()
	return notation.WriteComparisonTSV(out, rows, fam)
}
func report(args []string) error {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fpName := fs.String("fingerprint", "", "fingerprint JSON")
	outName := fs.String("output", "", "report Markdown")
	if err := fs.Parse(args); err != nil {
		return err
	}
	fp, err := readFP(*fpName)
	if err != nil {
		return err
	}
	out, err := createNew(*outName)
	if err != nil {
		return err
	}
	defer out.Close()
	fmt.Fprintf(out, "# CORPUS STRUCTURAL REPORT\n\n- corpus: `%s`\n- representation: `%s`\n- input SHA-256: `%s`\n- records: %d\n\nThis report contains corpus structure only. It contains no VM distance, ranking, or interpretation.\n", fp.CorpusID, fp.Representation, fp.InputSHA256, fp.RecordCount)
	return nil
}
func compareClasses(args []string) error {
	fs := flag.NewFlagSet("compare-classes", flag.ContinueOnError)
	manifestName := fs.String("manifest", "", "JSON manifest containing fingerprints[]")
	scaleName := fs.String("scale", "", "pre-frozen scale TSV")
	output := fs.String("output", "", "pairwise family TSV")
	authorize := fs.Bool("authorize-production", false, "future production authorization")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if !*authorize {
		return fmt.Errorf("prepared but locked: --authorize-production is required")
	}
	if !productionAuthorized() {
		return fmt.Errorf("prepared but locked: cross-class comparison requires PRODUCTION_COMPARATIVE_RUN_AUTHORIZED=true")
	}
	if *manifestName == "" || *scaleName == "" || *output == "" {
		return fmt.Errorf("compare-classes requires --manifest, --scale, and --output")
	}
	var m struct {
		Fingerprints []string `json:"fingerprints"`
	}
	mb, err := os.ReadFile(*manifestName)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(mb, &m); err != nil {
		return err
	}
	if len(m.Fingerprints) < 2 {
		return fmt.Errorf("at least two fingerprints required")
	}
	sf, err := os.Open(*scaleName)
	if err != nil {
		return err
	}
	scales, err := notation.ReadScalesTSV(sf)
	sf.Close()
	if err != nil {
		return err
	}
	fps := make([]notation.Fingerprint, len(m.Fingerprints))
	for i, p := range m.Fingerprints {
		fps[i], err = readFP(p)
		if err != nil {
			return err
		}
	}
	out, err := createNew(*output)
	if err != nil {
		return err
	}
	defer out.Close()
	fmt.Fprintln(out, "left_corpus\tleft_representation\tright_corpus\tright_representation\tfamily\tdistance\tstatus\treason")
	for i := 0; i < len(fps); i++ {
		for j := i + 1; j < len(fps); j++ {
			_, fam, e := notation.Compare(fps[i], fps[j], scales)
			if e != nil {
				return e
			}
			for _, d := range fam {
				fmt.Fprintf(out, "%s\t%s\t%s\t%s\t%s\t%.12g\t%s\t%s\n", fps[i].CorpusID, fps[i].Representation, fps[j].CorpusID, fps[j].Representation, d.Family, d.Distance, d.Status, d.Reason)
			}
		}
	}
	return nil
}
func productionAuthorized() bool {
	b, err := os.ReadFile("research/comparative_notation/PRODUCTION_COMPARATIVE_RUN_AUTHORIZED")
	return err == nil && strings.TrimSpace(string(b)) == "true"
}
func readFP(name string) (notation.Fingerprint, error) {
	f, err := os.Open(name)
	if err != nil {
		return notation.Fingerprint{}, err
	}
	defer f.Close()
	return notation.ReadFingerprintJSON(f)
}

func parseCheckpoints(s string) ([]int, error) {
	if s == "" {
		return notation.RarefactionCheckpoints, nil
	}
	var out []int
	for _, p := range strings.Split(s, ",") {
		n, err := strconv.Atoi(strings.TrimSpace(p))
		if err != nil {
			return nil, fmt.Errorf("invalid checkpoint %q: %w", p, err)
		}
		out = append(out, n)
	}
	return out, nil
}

func readRecords(name string) ([]notation.Record, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return notation.ReadJSONL(f)
}

// vmAdapter builds the frozen VM USC corpus directly from the frozen
// canonical source, verifying its SHA-256 before touching anything else
// (B02 sections 34-35). It never invents section/page/locus levels the
// source does not encode.
func vmAdapter(args []string) error {
	fs := flag.NewFlagSet("vm-adapter", flag.ContinueOnError)
	src := fs.String("source", "", "frozen VM canonical source (data_work/ZL3b-x7.canonical.txt)")
	expectedSHA := fs.String("expected-sha256", "", "frozen source SHA-256")
	corpusID := fs.String("corpus-id", "VM-ZL3b-x7", "VM corpus id")
	repID := fs.String("representation-id", "VM-CANONICAL", "VM representation id")
	out := fs.String("output", "", "output USC JSONL")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *src == "" || *expectedSHA == "" || *out == "" {
		return fmt.Errorf("vm-adapter requires --source, --expected-sha256, and --output")
	}
	raw, err := os.ReadFile(*src)
	if err != nil {
		return err
	}
	got := sha256.Sum256(raw)
	if hex.EncodeToString(got[:]) != *expectedSHA {
		return fmt.Errorf("source SHA-256 mismatch: got %s want %s (adversarial test A7: corrupt provenance SHA)", hex.EncodeToString(got[:]), *expectedSHA)
	}
	var recs []notation.Record
	lineIdx := 0
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		for ti, tok := range fields {
			syms := make([]string, 0, len(tok))
			for _, r := range tok {
				syms = append(syms, string(r))
			}
			recs = append(recs, notation.Record{
				SchemaVersion: notation.SchemaVersion, CorpusID: *corpusID, Representation: *repID,
				Document:     notation.ObservedLevel{Value: *corpusID, Observed: true},
				PhysicalLine: notation.ObservedLevel{Value: strconv.Itoa(lineIdx), Observed: true},
				TokenID:      fmt.Sprintf("%s-%08d", *corpusID, len(recs)),
				TokenIndex:   ti, Token: tok, Symbols: syms,
			})
		}
		lineIdx++
	}
	if err := notation.Validate(recs); err != nil {
		return fmt.Errorf("generated VM USC failed validation: %w", err)
	}
	outFile, err := createNew(*out)
	if err != nil {
		return err
	}
	defer outFile.Close()
	if err := notation.WriteJSONL(outFile, recs); err != nil {
		return err
	}
	fmt.Printf("VM-ADAPTED records=%d lines=%d corpus=%s representation=%s\n", len(recs), lineIdx, *corpusID, *repID)
	return nil
}

func rarefy(args []string) error {
	fs := flag.NewFlagSet("rarefy", flag.ContinueOnError)
	in := fs.String("input", "", "USC JSONL")
	checkpoints := fs.String("checkpoints", "", "comma-separated checkpoints (default frozen 5000,10000,20000,39380)")
	replicates := fs.Int("replicates", notation.RarefactionReplicates, "replicate count R")
	seed := fs.Int64("seed", notation.BaseSeed, "frozen base seed")
	fixture := fs.Bool("fixture", false, "confirm a small manually inspectable fixture (not a C01-C09 candidate)")
	outRows := fs.String("output-rows", "", "RAREFACTION.tsv")
	outSummary := fs.String("output-summary", "", "RAREFACTION_SUMMARY.tsv")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *in == "" || *outRows == "" || *outSummary == "" {
		return fmt.Errorf("rarefy requires --input, --output-rows, and --output-summary")
	}
	rs, err := readRecords(*in)
	if err != nil {
		return err
	}
	if err := guardNotUnauthorizedCandidate(rs[0].CorpusID, *fixture); err != nil {
		return err
	}
	ckpts, err := parseCheckpoints(*checkpoints)
	if err != nil {
		return err
	}
	rows, summary, err := notation.RunRarefaction(rs, rs[0].CorpusID, rs[0].Representation, ckpts, *replicates, *seed)
	if err != nil {
		return err
	}
	rf, err := createNew(*outRows)
	if err != nil {
		return err
	}
	defer rf.Close()
	if err := notation.WriteRarefactionTSV(rf, rows); err != nil {
		return err
	}
	sf, err := createNew(*outSummary)
	if err != nil {
		return err
	}
	defer sf.Close()
	return notation.WriteRarefactionSummaryTSV(sf, summary)
}

func bootstrapCmd(args []string) error {
	fs := flag.NewFlagSet("bootstrap", flag.ContinueOnError)
	in := fs.String("input", "", "USC JSONL")
	replicates := fs.Int("replicates", notation.BootstrapReplicates, "replicate count B")
	seed := fs.Int64("seed", notation.BaseSeed, "frozen base seed")
	fixture := fs.Bool("fixture", false, "confirm a small manually inspectable fixture (not a C01-C09 candidate)")
	out := fs.String("output", "", "BOOTSTRAP_RESULTS.tsv")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *in == "" || *out == "" {
		return fmt.Errorf("bootstrap requires --input and --output")
	}
	rs, err := readRecords(*in)
	if err != nil {
		return err
	}
	if err := guardNotUnauthorizedCandidate(rs[0].CorpusID, *fixture); err != nil {
		return err
	}
	rows, err := notation.RunBootstrap(rs, rs[0].CorpusID, rs[0].Representation, *replicates, *seed)
	if err != nil {
		return err
	}
	outFile, err := createNew(*out)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return notation.WriteBootstrapTSV(outFile, rows)
}

func distributionsCmd(args []string) error {
	fs := flag.NewFlagSet("distributions", flag.ContinueOnError)
	in := fs.String("input", "", "USC JSONL")
	fixture := fs.Bool("fixture", false, "confirm a small manually inspectable fixture (not a C01-C09 candidate)")
	out := fs.String("output", "", "DISTRIBUTIONS.tsv")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *in == "" || *out == "" {
		return fmt.Errorf("distributions requires --input and --output")
	}
	rs, err := readRecords(*in)
	if err != nil {
		return err
	}
	if err := guardNotUnauthorizedCandidate(rs[0].CorpusID, *fixture); err != nil {
		return err
	}
	dist := notation.BuildDistributions(rs, rs[0].CorpusID, rs[0].Representation)
	outFile, err := createNew(*out)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return notation.WriteDistributionsTSV(outFile, dist)
}

func metricOutputTypesCmd(args []string) error {
	fs := flag.NewFlagSet("metric-output-types", flag.ContinueOnError)
	out := fs.String("output", "", "METRIC_OUTPUT_TYPES.tsv")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *out == "" {
		return fmt.Errorf("metric-output-types requires --output")
	}
	outFile, err := createNew(*out)
	if err != nil {
		return err
	}
	defer outFile.Close()
	return notation.WriteMetricOutputTypesTSV(outFile, notation.MetricOutputTypes())
}

// vmReferenceCmd computes the immutable VM reference v2 once from the
// frozen VM USC input: VM_REFERENCE_V2.tsv, VM_ACCUMULATION_CURVES_V2.tsv,
// and VM_REFERENCE_V2_MANIFEST.json (B02 sections 34-39). It is never
// authorized to recompute against a candidate.
func vmReferenceCmd(args []string) error {
	fs := flag.NewFlagSet("vm-reference", flag.ContinueOnError)
	in := fs.String("input", "", "VM USC JSONL (from vm-adapter)")
	sourceSHA := fs.String("source-sha256", "", "frozen raw source SHA-256")
	outRef := fs.String("output-reference", "", "VM_REFERENCE_V2.tsv")
	outCurves := fs.String("output-curves", "", "VM_ACCUMULATION_CURVES_V2.tsv")
	outFingerprint := fs.String("output-fingerprint", "", "VM_REFERENCE_V2.fingerprint.json (the comparator-consumable frozen artifact; its hash is what compare-vm verifies)")
	outManifest := fs.String("output-manifest", "", "VM_REFERENCE_V2_MANIFEST.json")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *in == "" || *sourceSHA == "" || *outRef == "" || *outCurves == "" || *outFingerprint == "" || *outManifest == "" {
		return fmt.Errorf("vm-reference requires --input, --source-sha256, --output-reference, --output-curves, --output-fingerprint, and --output-manifest")
	}
	rs, err := readRecords(*in)
	if err != nil {
		return err
	}
	if err := guardNotUnauthorizedCandidate(rs[0].CorpusID, false); err != nil {
		return err
	}
	fp, err := notation.Analyze(rs)
	if err != nil {
		return err
	}
	refRows := notation.BuildVMReference(fp, fp.RecordCount)
	var refBuf bytes.Buffer
	if err := notation.WriteVMReferenceTSV(&refBuf, refRows); err != nil {
		return err
	}
	refFile, err := createNew(*outRef)
	if err != nil {
		return err
	}
	if _, err := refFile.Write(refBuf.Bytes()); err != nil {
		refFile.Close()
		return err
	}
	if err := refFile.Close(); err != nil {
		return err
	}
	curvesFile, err := createNew(*outCurves)
	if err != nil {
		return err
	}
	if err := notation.WriteCurvesTSV(curvesFile, fp.Curves); err != nil {
		curvesFile.Close()
		return err
	}
	if err := curvesFile.Close(); err != nil {
		return err
	}
	var fpBuf bytes.Buffer
	if err := notation.WriteFingerprintJSON(&fpBuf, fp); err != nil {
		return err
	}
	fpFile, err := createNew(*outFingerprint)
	if err != nil {
		return err
	}
	if _, err := fpFile.Write(fpBuf.Bytes()); err != nil {
		fpFile.Close()
		return err
	}
	if err := fpFile.Close(); err != nil {
		return err
	}
	registryHash, err := notation.MetricRegistryHash()
	if err != nil {
		return err
	}
	outHash := sha256.Sum256(fpBuf.Bytes())
	manifest := notation.VMReferenceManifest{
		SchemaVersion: "vm-reference-v2-manifest-1.0", SourceSHA256: *sourceSHA,
		AdapterVersion: "vm-canonical-adapter-1.0", AnalyzerVersion: "notation-analyzer-1.0",
		MetricRegistryVersion: notation.MetricRegistryVersion, MetricRegistryHash: registryHash,
		RarefactionProtocolVersion: "rarefaction-protocol-1.0", BootstrapProtocolVersion: "bootstrap-protocol-1.0",
		CalibrationScaleVersion: "calibration-scales-1.0", SeedSchedule: notation.BaseSeed,
		OutputSHA256: hex.EncodeToString(outHash[:]),
	}
	mf, err := createNew(*outManifest)
	if err != nil {
		return err
	}
	defer mf.Close()
	enc := json.NewEncoder(mf)
	enc.SetIndent("", "  ")
	return enc.Encode(manifest)
}

// calibrateCmd runs the frozen calibration panel at every checkpoint and
// writes the pooled CALIBRATION_SCALES.tsv plus a leave-one-generator-out
// stability/dominance diagnostic (B01 sections 28-32). It never reads a VM
// or candidate file.
func calibrateCmd(args []string) error {
	fs := flag.NewFlagSet("calibrate", flag.ContinueOnError)
	checkpoints := fs.String("checkpoints", "", "comma-separated checkpoints (default frozen 5000,10000,20000,39380)")
	replicates := fs.Int("replicates", notation.CalibrationCorpora, "independent corpora per generator per checkpoint")
	seed := fs.Int64("seed", notation.BaseSeed, "frozen base seed")
	outScales := fs.String("output-scales", "", "CALIBRATION_SCALES.tsv")
	outDiagnostics := fs.String("output-diagnostics", "", "optional JSON leave-one-out/dominance diagnostic")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *outScales == "" {
		return fmt.Errorf("calibrate requires --output-scales")
	}
	ckpts, err := parseCheckpoints(*checkpoints)
	if err != nil {
		return err
	}
	var allScales []notation.CalibrationScale
	diagnostics := map[string]any{}
	for _, ck := range ckpts {
		runs := notation.RunCalibrationPanel(ck, *replicates, *seed)
		metrics, err := notation.AnalyzeCalibrationRuns(runs)
		if err != nil {
			return err
		}
		allScales = append(allScales, notation.BuildCalibrationScales(metrics, ck)...)
		if *outDiagnostics != "" {
			contribution := map[string]int{}
			for _, m := range metrics {
				contribution[m.GeneratorID] += len(m.Values)
			}
			leaveOneOut := map[string]int{}
			for gen := range contribution {
				reduced := notation.LeaveOneGeneratorFamilyOut(metrics, gen, ck)
				leaveOneOut[gen] = len(reduced)
			}
			diagnostics[strconv.Itoa(ck)] = map[string]any{"generator_observation_counts": contribution, "leave_one_out_stratum_counts": leaveOneOut}
		}
	}
	sf, err := createNew(*outScales)
	if err != nil {
		return err
	}
	defer sf.Close()
	if err := notation.WriteCalibrationScalesTSV(sf, allScales); err != nil {
		return err
	}
	if *outDiagnostics != "" {
		df, err := createNew(*outDiagnostics)
		if err != nil {
			return err
		}
		defer df.Close()
		enc := json.NewEncoder(df)
		enc.SetIndent("", "  ")
		return enc.Encode(diagnostics)
	}
	return nil
}
