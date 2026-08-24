package task82a

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"math"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/mnemonicspace"
)

// chunkSeed derives a per-chunk seed from the job seed and chunk index,
// deterministically and without any dependence on chunk content.
func chunkSeed(jobSeed uint64, chunk int) int64 {
	h := sha256.Sum256(fmt.Appendf(nil, "%d|%d", jobSeed, chunk))
	return int64(binary.BigEndian.Uint64(h[:8]))
}

// buildChunkInput constructs the frozen mechanism's InputModel for local
// chunk index i, under the given surface/cue-namespace policy. It never
// changes primitive/operation/carrier semantics: it only supplies which
// literal symbols or which opaque cue labels occupy chunk i's local
// Index/Position/Tick domain (always 0..capacity-1, per task82a.txt
// sec.15-16 -- cue namespace only changes the cue *label*, not the
// mechanism's own addressing).
func buildChunkInput(scale MechanismScale, policy string, corpus SourceCorpus, chunkIndex int) (mnemonicspace.InputModel, []string) {
	id := fmt.Sprintf("%s-chunk-%06d", corpus.ID, chunkIndex)
	in := mnemonicspace.InputModel{ID: id, IndexedCues: map[mnemonicspace.Index]mnemonicspace.Cue{}, PositionedCues: map[mnemonicspace.Position]mnemonicspace.Cue{}, TimedCues: map[int]mnemonicspace.Cue{}}
	if scale.Surface == SurfaceLiteral {
		seq := corpus.Letters[chunkIndex*scale.Capacity : (chunkIndex+1)*scale.Capacity]
		for _, s := range seq {
			in.Sequence = append(in.Sequence, mnemonicspace.Symbol(s))
		}
		return in, seq
	}
	items := corpus.Items[chunkIndex*scale.Capacity : (chunkIndex+1)*scale.Capacity]
	for j := 0; j < scale.Capacity; j++ {
		label := j
		if policy == PolicyCueResetGlobal {
			label = chunkIndex*scale.Capacity + j
		}
		cue := cueLabel(label)
		in.IndexedCues[mnemonicspace.Index(j)] = cue
		in.PositionedCues[mnemonicspace.Position(j)] = cue
		in.TimedCues[(j+1)%scale.Capacity] = cue
	}
	return in, items
}

// chunkRecovery reconstructs a Task82-equivalent RecoveryEnvironment for
// one sampled chunk and runs all seven frozen recovery conditions,
// reusing exactly the candidate/environment/negative-control-corruption
// pattern internal/task82.runJob uses (task82a.txt sec.34-35: every local
// application must be equivalent to a standalone Task82 application of the
// same frozen mechanism).
func chunkRecovery(spec mnemonicspace.MechanismSpec, param mnemonicspace.ParameterSet, in mnemonicspace.InputModel, prepared mnemonicspace.PreparedRun, chunkItems []string, capacity int) map[string]mnemonicspace.RecoveryResult {
	targetIdx := mnemonicspace.Index(0)
	if param.Rotation != nil {
		targetIdx = mnemonicspace.Index((((0 + param.Rotation.Offset) % capacity) + capacity) % capacity)
	}
	targetPos := mnemonicspace.Position(0)
	path := make([]mnemonicspace.Position, capacity)
	for i := range path {
		path[i] = mnemonicspace.Position(i)
	}
	candidates := []mnemonicspace.RetrievedItem{mnemonicspace.RetrievedItem(chunkItems[0]), mnemonicspace.RetrievedItem("ALT_" + chunkItems[1])}
	firstCue := mnemonicspace.Cue("")
	if c, ok := in.IndexedCues[0]; ok {
		firstCue = c
	}
	env := mnemonicspace.RecoveryEnvironment{
		Geometry:       &mnemonicspace.GeometryKnowledge{Path: path},
		History:        &mnemonicspace.HistoryKnowledge{Steps: 1},
		Convention:     &mnemonicspace.ConventionKnowledge{MessageLength: len(in.Sequence), CueMeanings: map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{firstCue: candidates}},
		InternalMemory: &mnemonicspace.InternalMemoryState{Associations: map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{firstCue: candidates}},
		Context:        &mnemonicspace.ContextKnowledge{Allowed: []mnemonicspace.RetrievedItem{mnemonicspace.RetrievedItem(chunkItems[0])}},
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierGeometry) {
		env.Geometry = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierHistory) {
		env.History = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierConvention) {
		env.Convention = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierInternal) {
		env.InternalMemory = nil
	}
	if !hasCarrier(spec.Carriers.Retrieve, mnemonicspace.CarrierContext) {
		env.Context = nil
	}
	switch spec.ID {
	case "negative_randomized_convention":
		if env.Convention != nil {
			env.Convention.MessageLength = len(in.Sequence) - 1
		}
	case "negative_randomized_path":
		env.Geometry = nil
	case "negative_randomized_cue_association":
		if env.InternalMemory != nil {
			env.InternalMemory.Associations = map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{"WRONG": {mnemonicspace.RetrievedItem(chunkItems[0])}}
		}
	case "negative_randomized_index_mapping":
		if env.Convention != nil {
			env.Convention.CueMeanings = map[mnemonicspace.Cue][]mnemonicspace.RetrievedItem{"WRONG": {mnemonicspace.RetrievedItem(chunkItems[0])}}
		}
	}
	conditions := []mnemonicspace.RecoveryCondition{
		mnemonicspace.RecoveryFullKnowledge, mnemonicspace.RecoveryNoContext, mnemonicspace.RecoveryNoConvention,
		mnemonicspace.RecoveryNoGeometry, mnemonicspace.RecoveryNoHistory, mnemonicspace.RecoveryNoInternal, mnemonicspace.RecoveryObservable,
	}
	out := map[string]mnemonicspace.RecoveryResult{}
	for _, cond := range conditions {
		req := mnemonicspace.RetrievalRequest{Condition: cond, TargetIndex: &targetIdx, TargetPosition: &targetPos}
		r, err := (mnemonicspace.Runner{}).Recover(spec, param, prepared, req, env)
		if err != nil {
			r = mnemonicspace.RecoveryResult{Class: mnemonicspace.ResultNoRecovery, Detail: "runner error: " + err.Error()}
		}
		out[string(cond)] = r
	}
	return out
}

func hasCarrier(list []mnemonicspace.Carrier, c mnemonicspace.Carrier) bool {
	for _, v := range list {
		if v == c {
			return true
		}
	}
	return false
}

// sampleIndices picks a deterministic, preregistered sample of chunk
// indices for recovery replication (task82a.txt sec.37): a fixed stride
// targeting at most maxSamples points, always including chunk 0. The
// formula is fixed before generation and never adjusted after viewing
// output.
func sampleIndices(chunks, maxSamples int) []int {
	if chunks <= maxSamples {
		out := make([]int, chunks)
		for i := range out {
			out[i] = i
		}
		return out
	}
	stride := chunks / maxSamples
	if stride < 1 {
		stride = 1
	}
	var out []int
	for i := 0; i < chunks && len(out) < maxSamples; i += stride {
		out = append(out, i)
	}
	return out
}

const maxRecoverySamples = 8

// recoveryScore mirrors internal/task82's frozen estimand definition
// (TASK82_DESIGN.md "Estimands"): 1 for EXACT, the reported fraction for
// PARTIAL, 1/n for an n-way ambiguity set, 0 for CUE_ONLY/NO_RECOVERY, and
// NaN (excluded from means) for NOT_APPLICABLE.
func recoveryScore(r mnemonicspace.RecoveryResult, exactMatch bool) float64 {
	switch r.Class {
	case mnemonicspace.ResultExact:
		if exactMatch {
			return 1
		}
		return 0
	case mnemonicspace.ResultPartial:
		return 0.5
	case mnemonicspace.ResultAmbiguitySet:
		if len(r.Candidates) == 0 {
			return 0
		}
		return 1 / float64(len(r.Candidates))
	case mnemonicspace.ResultNotApplicable:
		return math.NaN()
	default:
		return 0
	}
}

// ChunkSummary is the lightweight per-chunk record kept for every chunk in
// the corpus (not just sampled ones), enough to assemble the observable
// document and detect collisions without retaining full mechanism state.
type ChunkSummary struct {
	Index      int      `json:"index"`
	Symbols    []string `json:"symbols"`
	Checksum   string   `json:"checksum"`
	IntendedID string   `json:"intended_id"`
}

// SampledRecovery is the per-sampled-chunk local recovery record
// (task82a.txt sec.35, LOCAL_RECOVERY).
type SampledRecovery struct {
	ChunkIndex int                `json:"chunk_index"`
	Scores     map[string]float64 `json:"recovery_score_by_condition"`
	Classes    map[string]string  `json:"recovery_class_by_condition"`
	Ambiguity  map[string]int     `json:"ambiguity_cardinality_by_condition"`
}

// AssembledJob is everything one manifest job produces before F2
// extraction: the assembled document, chunk-level records, and local
// recovery replication.
type AssembledJob struct {
	Chunks []ChunkSummary
	// AssembledLines has one entry per chunk (ASSEMBLER_DEFINED line == one
	// chunk). For literal mechanisms a line holds exactly one token, the
	// chunk's symbols concatenated with no separator (so vocabulary grows
	// with chunk count, like a real corpus's distinct words). For cue
	// mechanisms a line holds one token per emitted cue (so LOCAL_NAMESPACE
	// gives a small, bounded vocabulary and GLOBAL_NAMESPACE gives a
	// vocabulary growing with chunk count x capacity) -- collapsing a
	// whole cue chunk into a single joined token would trivially degenerate
	// LOCAL_NAMESPACE to one repeated word, which is a corpus-file
	// artifact of the assembler, not a property of the frozen mechanism.
	AssembledLines [][]string
	Document       mnemonicspace.ObservableDocument
	Recoveries     []SampledRecovery
	Warnings       []string
}

// assembleJob runs the CorpusScaleAssembler for one job: it applies the
// frozen local mechanism once per chunk (always RESET_EACH_CHUNK -- see
// PoliciesFor), assembles the corpus-scale OBSERVABLE_DOCUMENT purely from
// each chunk's own observable output, and samples local recovery.
func assembleJob(spec mnemonicspace.MechanismSpec, param mnemonicspace.ParameterSet, scale MechanismScale, policy string, corpus SourceCorpus, chunks int, jobSeed uint64) (AssembledJob, error) {
	var out AssembledJob
	sampled := map[int]bool{}
	for _, i := range sampleIndices(chunks, maxRecoverySamples) {
		sampled[i] = true
	}
	for i := 0; i < chunks; i++ {
		in, unitItems := buildChunkInput(scale, policy, corpus, i)
		seed := chunkSeed(jobSeed, i)
		prepared, err := (mnemonicspace.Runner{}).Prepare(spec, param, in, seed)
		if err != nil {
			return AssembledJob{}, fmt.Errorf("chunk %d prepare: %w", i, err)
		}
		if err := validateLocalDocument(spec, prepared.Document); err != nil {
			return AssembledJob{}, fmt.Errorf("LEAKAGE_FAILURE: chunk %d: %w", i, err)
		}
		intended := strings.Join(unitItems, "")
		if scale.Surface == SurfaceCue {
			intended = unitItems[0]
		}
		out.Chunks = append(out.Chunks, ChunkSummary{Index: i, Symbols: prepared.Document.Symbols, Checksum: prepared.Document.Checksum(), IntendedID: intended})
		if scale.Surface == SurfaceLiteral {
			out.AssembledLines = append(out.AssembledLines, []string{strings.Join(prepared.Document.Symbols, "")})
		} else {
			out.AssembledLines = append(out.AssembledLines, append([]string(nil), prepared.Document.Symbols...))
		}
		if sampled[i] {
			results := chunkRecovery(spec, param, in, prepared, unitItems, scale.Capacity)
			rec := SampledRecovery{ChunkIndex: i, Scores: map[string]float64{}, Classes: map[string]string{}, Ambiguity: map[string]int{}}
			target := intended
			for cond, r := range results {
				exact := r.Class == mnemonicspace.ResultExact && string(r.Value) == target
				if score := recoveryScore(r, exact); !math.IsNaN(score) {
					rec.Scores[cond] = score
				}
				rec.Classes[cond] = string(r.Class)
				rec.Ambiguity[cond] = len(r.Candidates)
			}
			out.Recoveries = append(out.Recoveries, rec)
		}
	}
	sort.Slice(out.Recoveries, func(i, j int) bool { return out.Recoveries[i].ChunkIndex < out.Recoveries[j].ChunkIndex })
	allSymbols := make([]string, 0, len(out.Chunks))
	for _, c := range out.Chunks {
		allSymbols = append(allSymbols, c.Symbols...)
	}
	family := ""
	if len(spec.Family) > 0 {
		family = string(spec.Family)
	}
	out.Document = mnemonicspace.ObservableDocument{
		Symbols:       allSymbols,
		TokenBoundary: mnemonicspace.BoundaryGenerated, // ASSEMBLER_DEFINED: one token per chunk (see AssembledTokens)
		LineBoundary:  mnemonicspace.BoundaryGenerated, // ASSEMBLER_DEFINED: one line per chunk
		UnitBoundary:  spec.Serialization.UnitBoundary,
		Metadata:      map[string]string{"family": family, "token_boundary_provenance": "ASSEMBLER_DEFINED", "line_boundary_provenance": "ASSEMBLER_DEFINED", "page_boundary_provenance": "NOT_DEFINED"},
	}
	return out, nil
}

// validateLocalDocument is the same leakage guard internal/task82 applies
// to every local application (task82a.txt sec.34: local semantics
// invariant).
func validateLocalDocument(spec mnemonicspace.MechanismSpec, d mnemonicspace.ObservableDocument) error {
	if !spec.Serialization.ExcludesHiddenState {
		return fmt.Errorf("registry permits hidden state")
	}
	if d.UnitBoundary != spec.Serialization.UnitBoundary {
		return fmt.Errorf("unit boundary provenance mismatch")
	}
	for k := range d.Metadata {
		s := strings.ToLower(k)
		for _, bad := range []string{"plaintext", "offset", "association", "internal_memory", "message_length", "path", "history"} {
			if strings.Contains(s, bad) {
				return fmt.Errorf("hidden metadata %q", k)
			}
		}
	}
	return nil
}
