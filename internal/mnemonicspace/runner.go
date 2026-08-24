package mnemonicspace

import (
	"fmt"
	"math/rand"
	"sort"
	"strings"

	"zcore.dev/voinich/internal/fontanafamily"
	"zcore.dev/voinich/internal/speculumf01"
)

type Runner struct{}

func (Runner) Run(spec MechanismSpec, params ParameterSet, input InputModel, seed int64, request RetrievalRequest, env RecoveryEnvironment) (Execution, error) {
	prepared, err := (Runner{}).Prepare(spec, params, input, seed)
	if err != nil {
		return Execution{}, err
	}
	recovery, err := (Runner{}).Recover(spec, params, prepared, request, env)
	if err != nil {
		return Execution{}, err
	}
	return Execution{Prepared: prepared, Recovery: recovery}, nil
}

func (Runner) Prepare(spec MechanismSpec, params ParameterSet, input InputModel, seed int64) (PreparedRun, error) {
	switch spec.runtime {
	case kindF01:
		return prepareF01(spec, params, input, seed)
	case kindF08:
		return prepareF08(spec, params, input)
	case kindF11:
		return prepareF11(spec, input)
	case kindF12:
		return prepareF12(spec, params, input)
	case kindRotationIndex:
		return prepareRotationIndex(spec, params, input)
	case kindStorageAssociate:
		return prepareStorageAssociate(spec, params, input)
	default:
		return PreparedRun{}, fmt.Errorf("unsupported runtime kind")
	}
}

func (Runner) Recover(spec MechanismSpec, params ParameterSet, prepared PreparedRun, request RetrievalRequest, env RecoveryEnvironment) (RecoveryResult, error) {
	if request.Condition == "" {
		request.Condition = RecoveryFullKnowledge
	}
	if request.Condition != RecoveryFullKnowledge && request.Condition != RecoveryObservable {
		carrier, ok := removedCarrier(request.Condition)
		if ok && !usesCarrier(spec.Carriers.Retrieve, carrier) {
			return RecoveryResult{Class: ResultNotApplicable, Detail: fmt.Sprintf("%s does not use carrier %s", spec.ID, carrier)}, nil
		}
	}
	pruned := pruneEnvironment(request.Condition, env)
	switch spec.runtime {
	case kindF01:
		return recoverF01(params, prepared, request, pruned)
	case kindF08:
		return recoverF08(params, prepared, request, pruned)
	case kindF11:
		return recoverF11(prepared, request, pruned)
	case kindF12:
		return recoverF12(prepared, request, pruned)
	case kindRotationIndex:
		return recoverRotationIndex(prepared, request, pruned)
	case kindStorageAssociate:
		return recoverStorageAssociate(prepared, request, pruned)
	default:
		return RecoveryResult{}, fmt.Errorf("unsupported runtime kind")
	}
}

func prepareF01(spec MechanismSpec, params ParameterSet, input InputModel, seed int64) (PreparedRun, error) {
	if params.F01 == nil {
		return PreparedRun{}, fmt.Errorf("missing F01 parameters")
	}
	cfg, err := toSpeculumConfig(*params.F01)
	if err != nil {
		return PreparedRun{}, err
	}
	rng := rand.New(rand.NewSource(seed))
	state, err := cfg.Encode(input.SequenceString(), func() int { return rng.Int() })
	if err != nil {
		return PreparedRun{}, err
	}
	visible := make([]string, len(state.Offsets))
	for i, offset := range state.Offsets {
		visible[i] = string(cfg.LetterAtMark(offset))
	}
	return PreparedRun{
		MechanismID:    spec.ID,
		ParameterSetID: params.ID,
		State:          ExternalState{F01: &F01State{Offsets: append([]int(nil), state.Offsets...)}},
		Observation:    ObservationStream{Symbols: append([]string(nil), visible...)},
		Document: ObservableDocument{
			Symbols:       append([]string(nil), visible...),
			TokenBoundary: BoundaryNotDefined,
			LineBoundary:  BoundaryNotDefined,
			UnitBoundary:  BoundaryGenerated,
			Metadata:      map[string]string{"family": string(spec.Family)},
		},
		Trace: []InformationEvent{{Carrier: CarrierMessage, Effect: EffectPreserve, Detail: "input sequence stored as ring offsets"}, {Carrier: CarrierExternalState, Effect: EffectReorder, Detail: "visible letters depend on ring orientation"}, {Carrier: CarrierConvention, Effect: EffectAddContext, Detail: "message length and reading convention are external"}},
	}, nil
}

func prepareF08(spec MechanismSpec, params ParameterSet, input InputModel) (PreparedRun, error) {
	if params.F08 == nil {
		return PreparedRun{}, fmt.Errorf("missing F08 parameters")
	}
	cfg, err := toSerpensConfig(*params.F08)
	if err != nil {
		return PreparedRun{}, err
	}
	state, err := cfg.Encode(input.SequenceString())
	if err != nil {
		return PreparedRun{}, err
	}
	visible := make([]string, len(state.Holes))
	for i, hole := range state.Holes {
		if hole == nil {
			visible[i] = params.F08.EmptyMarker
			if visible[i] == "" {
				visible[i] = "?"
			}
			continue
		}
		visible[i] = string(*hole)
	}
	return PreparedRun{
		MechanismID:    spec.ID,
		ParameterSetID: params.ID,
		State:          ExternalState{F08: &F08State{Holes: visible}},
		Observation:    ObservationStream{Symbols: append([]string(nil), visible...)},
		Document:       ObservableDocument{Symbols: append([]string(nil), visible...), TokenBoundary: BoundaryNotDefined, LineBoundary: BoundaryNotDefined, UnitBoundary: BoundaryGenerated, Metadata: map[string]string{"family": string(spec.Family)}},
		Trace:          []InformationEvent{{Carrier: CarrierMessage, Effect: EffectCombine, Detail: "symbols occupy explicit positions"}, {Carrier: CarrierGeometry, Effect: EffectReorder, Detail: "path knowledge orders visible positions"}, {Carrier: CarrierConvention, Effect: EffectAddContext, Detail: "message boundary remains external"}},
	}, nil
}

func prepareF11(spec MechanismSpec, input InputModel) (PreparedRun, error) {
	state := map[Index]Cue{}
	indices := sortedIndices(input.IndexedCues)
	for _, idx := range indices {
		state[idx] = input.IndexedCues[idx]
	}
	visible := make([]string, 0, len(indices))
	for _, idx := range indices {
		visible = append(visible, fmt.Sprintf("%d:%s", idx, input.IndexedCues[idx]))
	}
	return PreparedRun{
		MechanismID:    spec.ID,
		ParameterSetID: firstParamID(spec),
		State:          ExternalState{F11: &F11State{Slots: state}},
		Observation:    ObservationStream{Cues: valuesFromIndexed(state)},
		Document:       ObservableDocument{Symbols: visible, TokenBoundary: BoundaryNotDefined, LineBoundary: BoundaryNotDefined, UnitBoundary: BoundaryGenerated, Metadata: map[string]string{"family": string(spec.Family)}},
		Trace:          []InformationEvent{{Carrier: CarrierExternalState, Effect: EffectSelect, Detail: "indexing exposes an occupied slot"}, {Carrier: CarrierConvention, Effect: EffectAddContext, Detail: "cue meaning is external to the device"}, {Carrier: CarrierExternalState, Effect: EffectAmbiguate, Detail: "visible cue is not plaintext by itself"}},
	}, nil
}

func prepareF12(spec MechanismSpec, params ParameterSet, input InputModel) (PreparedRun, error) {
	if params.F12 == nil {
		return PreparedRun{}, fmt.Errorf("missing F12 parameters")
	}
	state := fontanafamily.Horalogius{Period: params.F12.Period, Tick: params.F12.InitialTick, Cues: map[int]string{}}
	for tick, cue := range input.TimedCues {
		state.Cues[tick] = string(cue)
	}
	advanced, emitted, err := state.Advance(params.F12.AdvanceSteps)
	if err != nil {
		return PreparedRun{}, err
	}
	visible := append([]string(nil), emitted...)
	cues := map[int]Cue{}
	for tick, cue := range advanced.Cues {
		cues[tick] = Cue(cue)
	}
	return PreparedRun{
		MechanismID:    spec.ID,
		ParameterSetID: params.ID,
		State:          ExternalState{F12: &F12State{Period: advanced.Period, Tick: advanced.Tick, Cues: cues}},
		Observation:    ObservationStream{Cues: visible},
		Document:       ObservableDocument{Symbols: visible, TokenBoundary: BoundaryGenerated, LineBoundary: BoundaryNotDefined, UnitBoundary: BoundaryGenerated, Metadata: map[string]string{"family": string(spec.Family)}},
		Trace:          []InformationEvent{{Carrier: CarrierExternalState, Effect: EffectSignal, Detail: "temporal state emits opaque cues"}, {Carrier: CarrierInternal, Effect: EffectRecallTrigger, Detail: "association is external to observable state"}, {Carrier: CarrierExternalState, Effect: EffectAmbiguate, Detail: "cue-only without internal mapping"}},
	}, nil
}

func prepareRotationIndex(spec MechanismSpec, params ParameterSet, input InputModel) (PreparedRun, error) {
	if params.Rotation == nil {
		return PreparedRun{}, fmt.Errorf("missing rotation-index parameters")
	}
	indices := sortedIndices(input.IndexedCues)
	if len(indices) == 0 {
		return PreparedRun{}, fmt.Errorf("empty indexed cue input")
	}
	slots := make([]Cue, len(indices))
	for i, idx := range indices {
		slots[i] = input.IndexedCues[idx]
	}
	offset := ((params.Rotation.Offset % len(slots)) + len(slots)) % len(slots)
	rotated := make([]Cue, len(slots))
	for i, cue := range slots {
		rotated[(i+offset)%len(slots)] = cue
	}
	visible := make([]string, len(rotated))
	for i, cue := range rotated {
		visible[i] = string(cue)
	}
	return PreparedRun{
		MechanismID:    spec.ID,
		ParameterSetID: params.ID,
		State:          ExternalState{RotationIndex: &RotationIndexState{Slots: rotated, Offset: offset}},
		Observation:    ObservationStream{Cues: visible},
		Document:       ObservableDocument{Symbols: visible, TokenBoundary: BoundaryNotDefined, LineBoundary: BoundaryNotDefined, UnitBoundary: BoundaryGenerated, Metadata: map[string]string{"family": string(spec.Family)}},
		Trace:          []InformationEvent{{Carrier: CarrierExternalState, Effect: EffectReorder, Detail: "rotation changes the visible index-to-cue alignment"}, {Carrier: CarrierConvention, Effect: EffectAddContext, Detail: "cue meaning remains external"}},
	}, nil
}

func prepareStorageAssociate(spec MechanismSpec, params ParameterSet, input InputModel) (PreparedRun, error) {
	if params.Storage == nil {
		return PreparedRun{}, fmt.Errorf("missing storage-associate parameters")
	}
	slots := map[Position]Cue{}
	for pos, cue := range input.PositionedCues {
		slots[pos] = cue
	}
	visible := make([]string, params.Storage.Capacity)
	for i := 0; i < len(visible); i++ {
		visible[i] = "."
	}
	for pos, cue := range slots {
		if int(pos) >= 0 && int(pos) < len(visible) {
			visible[int(pos)] = string(cue)
		}
	}
	return PreparedRun{
		MechanismID:    spec.ID,
		ParameterSetID: params.ID,
		State:          ExternalState{StorageAssociate: &StorageAssociateState{Slots: slots}},
		Observation:    ObservationStream{Cues: flattenVisibleCues(visible)},
		Document:       ObservableDocument{Symbols: visible, TokenBoundary: BoundaryNotDefined, LineBoundary: BoundaryNotDefined, UnitBoundary: BoundaryGenerated, Metadata: map[string]string{"family": string(spec.Family)}},
		Trace:          []InformationEvent{{Carrier: CarrierExternalState, Effect: EffectCombine, Detail: "cue identifiers occupy explicit positions"}, {Carrier: CarrierInternal, Effect: EffectRecallTrigger, Detail: "retrieval remains outside the visible store"}},
	}, nil
}

func recoverF01(params ParameterSet, prepared PreparedRun, _ RetrievalRequest, env RecoveryEnvironment) (RecoveryResult, error) {
	cfg, err := toSpeculumConfig(*params.F01)
	if err != nil {
		return RecoveryResult{}, err
	}
	state := speculumf01.State{Offsets: append([]int(nil), prepared.State.F01.Offsets...)}
	if env.Convention != nil && env.Convention.MessageLength > 0 {
		decoded, err := cfg.DecodeFull(state, env.Convention.MessageLength)
		if err == nil {
			return RecoveryResult{Class: ResultExact, Value: RetrievedItem(decoded), UsedCarriers: []Carrier{CarrierExternalState, CarrierConvention}, Detail: "full F01 convention supplied"}, nil
		}
	}
	candidates := f01Candidates(cfg, state)
	candidates = applyContext(candidates, env.Context)
	if len(candidates) == 1 {
		return RecoveryResult{Class: ResultExact, Value: candidates[0], UsedCarriers: []Carrier{CarrierExternalState, CarrierContext}, Detail: "context narrowed the ambiguity set"}, nil
	}
	if len(candidates) == 0 {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "no compatible F01 candidates"}, nil
	}
	return RecoveryResult{Class: ResultAmbiguitySet, Candidates: candidates, UsedCarriers: []Carrier{CarrierExternalState}, Detail: "convention withheld; multiple F01 candidates remain"}, nil
}

func recoverF08(params ParameterSet, prepared PreparedRun, _ RetrievalRequest, env RecoveryEnvironment) (RecoveryResult, error) {
	cfg, err := toSerpensConfig(*params.F08)
	if err != nil {
		return RecoveryResult{}, err
	}
	state := fontanafamily.SerpensState{Holes: make([]*rune, len(prepared.State.F08.Holes))}
	for i, symbol := range prepared.State.F08.Holes {
		if symbol == "" || symbol == params.F08.EmptyMarker || symbol == "?" || symbol == "." {
			continue
		}
		r := []rune(symbol)
		if len(r) > 0 {
			state.Holes[i] = runePtr(r[0])
		}
	}
	if env.Geometry != nil && env.Convention != nil && env.Convention.MessageLength > 0 {
		decoded, err := cfg.Decode(state, env.Convention.MessageLength)
		if err == nil {
			return RecoveryResult{Class: ResultExact, Value: RetrievedItem(decoded), UsedCarriers: []Carrier{CarrierExternalState, CarrierGeometry, CarrierConvention}, Detail: "declared F08 path supplied"}, nil
		}
	}
	lengths := []int{}
	if env.Convention != nil && env.Convention.MessageLength > 0 {
		lengths = append(lengths, env.Convention.MessageLength)
	} else {
		for n := 1; n <= params.F08.Capacity; n++ {
			lengths = append(lengths, n)
		}
	}
	seen := map[RetrievedItem]bool{}
	for _, n := range lengths {
		alts, err := cfg.CompatibleTraversals(state, n, env.Geometry != nil, env.Geometry != nil)
		if err != nil {
			continue
		}
		for _, alt := range alts {
			seen[RetrievedItem(alt)] = true
		}
	}
	candidates := make([]RetrievedItem, 0, len(seen))
	for candidate := range seen {
		candidates = append(candidates, candidate)
	}
	candidates = applyContext(candidates, env.Context)
	if len(candidates) == 1 {
		return RecoveryResult{Class: ResultExact, Value: candidates[0], UsedCarriers: []Carrier{CarrierExternalState, CarrierContext}, Detail: "context narrowed positional ambiguity"}, nil
	}
	if len(candidates) == 0 {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "no compatible F08 traversals"}, nil
	}
	return RecoveryResult{Class: ResultAmbiguitySet, Candidates: sortCandidates(candidates), UsedCarriers: []Carrier{CarrierExternalState}, Detail: "path or boundary withheld"}, nil
}

func recoverF11(prepared PreparedRun, request RetrievalRequest, env RecoveryEnvironment) (RecoveryResult, error) {
	if request.TargetIndex == nil {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "missing target index"}, nil
	}
	cue, ok := prepared.State.F11.Slots[*request.TargetIndex]
	if !ok {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "no cue at requested index"}, nil
	}
	if env.Convention == nil || len(env.Convention.CueMeanings[cue]) == 0 {
		return RecoveryResult{Class: ResultCueOnly, Cue: cue, UsedCarriers: []Carrier{CarrierExternalState}, Detail: "cue meaning withheld"}, nil
	}
	candidates := applyContext(env.Convention.CueMeanings[cue], env.Context)
	if len(candidates) == 1 {
		return RecoveryResult{Class: ResultExact, Value: candidates[0], Cue: cue, UsedCarriers: []Carrier{CarrierExternalState, CarrierConvention}, Detail: "cue convention supplied"}, nil
	}
	if len(candidates) == 0 {
		return RecoveryResult{Class: ResultNoRecovery, Cue: cue, Detail: "cue convention/context rejected all candidates"}, nil
	}
	return RecoveryResult{Class: ResultAmbiguitySet, Cue: cue, Candidates: sortCandidates(candidates), UsedCarriers: []Carrier{CarrierExternalState, CarrierConvention, CarrierContext}, Detail: "cue convention remains many-to-one"}, nil
}

func recoverF12(prepared PreparedRun, _ RetrievalRequest, env RecoveryEnvironment) (RecoveryResult, error) {
	if len(prepared.Observation.Cues) == 0 {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "no emitted cue"}, nil
	}
	cue := Cue(prepared.Observation.Cues[len(prepared.Observation.Cues)-1])
	if env.InternalMemory == nil || len(env.InternalMemory.Associations[cue]) == 0 {
		return RecoveryResult{Class: ResultCueOnly, Cue: cue, UsedCarriers: []Carrier{CarrierExternalState}, Detail: "internal memory withheld"}, nil
	}
	candidates := applyContext(env.InternalMemory.Associations[cue], env.Context)
	if len(candidates) == 1 {
		return RecoveryResult{Class: ResultExact, Value: candidates[0], Cue: cue, UsedCarriers: []Carrier{CarrierExternalState, CarrierInternal}, Detail: "association supplied"}, nil
	}
	if len(candidates) == 0 {
		return RecoveryResult{Class: ResultNoRecovery, Cue: cue, Detail: "context rejected all associated candidates"}, nil
	}
	return RecoveryResult{Class: ResultAmbiguitySet, Cue: cue, Candidates: sortCandidates(candidates), UsedCarriers: []Carrier{CarrierExternalState, CarrierInternal, CarrierContext}, Detail: "cue keeps multiple learned associations"}, nil
}

func recoverRotationIndex(prepared PreparedRun, request RetrievalRequest, env RecoveryEnvironment) (RecoveryResult, error) {
	if request.TargetIndex == nil {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "missing target index"}, nil
	}
	state := prepared.State.RotationIndex
	if len(state.Slots) == 0 {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "empty rotation-index state"}, nil
	}
	pos := ((int(*request.TargetIndex) % len(state.Slots)) + len(state.Slots)) % len(state.Slots)
	cue := state.Slots[pos]
	if cue == "" {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "no cue at requested rotated index"}, nil
	}
	if env.Convention == nil || len(env.Convention.CueMeanings[cue]) == 0 {
		return RecoveryResult{Class: ResultCueOnly, Cue: cue, UsedCarriers: []Carrier{CarrierExternalState}, Detail: "rotation+index remains cue-only without convention"}, nil
	}
	candidates := applyContext(env.Convention.CueMeanings[cue], env.Context)
	if len(candidates) == 1 {
		return RecoveryResult{Class: ResultExact, Value: candidates[0], Cue: cue, UsedCarriers: []Carrier{CarrierExternalState, CarrierConvention}, Detail: "M-restricted rotation+index with cue map"}, nil
	}
	return RecoveryResult{Class: ResultAmbiguitySet, Cue: cue, Candidates: sortCandidates(candidates), UsedCarriers: []Carrier{CarrierExternalState, CarrierConvention, CarrierContext}, Detail: "M-restricted cue map remains many-to-one"}, nil
}

func recoverStorageAssociate(prepared PreparedRun, request RetrievalRequest, env RecoveryEnvironment) (RecoveryResult, error) {
	if request.TargetPosition == nil {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "missing target position"}, nil
	}
	storedCue, ok := prepared.State.StorageAssociate.Slots[*request.TargetPosition]
	if !ok {
		return RecoveryResult{Class: ResultNoRecovery, Detail: "no cue at requested position"}, nil
	}
	cue := InterpretOpaqueCue(Symbol(storedCue))
	if env.InternalMemory == nil || len(env.InternalMemory.Associations[cue]) == 0 {
		return RecoveryResult{Class: ResultCueOnly, Cue: cue, UsedCarriers: []Carrier{CarrierExternalState}, Detail: "stored cue lacks internal association"}, nil
	}
	candidates := applyContext(env.InternalMemory.Associations[cue], env.Context)
	if len(candidates) == 1 {
		return RecoveryResult{Class: ResultExact, Value: candidates[0], Cue: cue, UsedCarriers: []Carrier{CarrierExternalState, CarrierInternal}, Detail: "stored cue resolved via association"}, nil
	}
	if len(candidates) == 0 {
		return RecoveryResult{Class: ResultNoRecovery, Cue: cue, Detail: "context rejected all candidates"}, nil
	}
	return RecoveryResult{Class: ResultAmbiguitySet, Cue: cue, Candidates: sortCandidates(candidates), UsedCarriers: []Carrier{CarrierExternalState, CarrierInternal, CarrierContext}, Detail: "stored cue remains ambiguous"}, nil
}

// InterpretOpaqueCue performs the registry-declared, zero-effect
// Symbol -> Cue interpretation for M-RESTRICTED explicit cue storage.
func InterpretOpaqueCue(symbol Symbol) Cue {
	return Cue(symbol)
}

func pruneEnvironment(condition RecoveryCondition, env RecoveryEnvironment) RecoveryEnvironment {
	switch condition {
	case RecoveryNoContext:
		env.Context = nil
	case RecoveryNoConvention:
		env.Convention = nil
	case RecoveryNoGeometry:
		env.Geometry = nil
	case RecoveryNoHistory:
		env.History = nil
	case RecoveryNoInternal:
		env.InternalMemory = nil
	case RecoveryObservable:
		env.Geometry = nil
		env.History = nil
		env.Convention = nil
		env.InternalMemory = nil
		env.Context = nil
	}
	return env
}

func removedCarrier(condition RecoveryCondition) (Carrier, bool) {
	switch condition {
	case RecoveryNoContext:
		return CarrierContext, true
	case RecoveryNoConvention:
		return CarrierConvention, true
	case RecoveryNoGeometry:
		return CarrierGeometry, true
	case RecoveryNoHistory:
		return CarrierHistory, true
	case RecoveryNoInternal:
		return CarrierInternal, true
	default:
		return "", false
	}
}

func usesCarrier(carriers []Carrier, want Carrier) bool {
	for _, carrier := range carriers {
		if carrier == want {
			return true
		}
	}
	return false
}

func toSpeculumConfig(p F01Parameters) (speculumf01.Config, error) {
	alphabet := speculumf01.NewAlphabet("task81", p.Alphabet)
	order := speculumf01.InnerToOuter
	if p.Order == "OUTER_TO_INNER" {
		order = speculumf01.OuterToInner
	}
	cfg := speculumf01.Config{NumRings: p.NumRings, Alphabet: alphabet, ReadRadius: p.ReadRadius, Order: order, RingIdentityMarked: p.RingIdentityMarked}
	if cfg.NumRings <= 0 || cfg.Alphabet.Size() == 0 {
		return cfg, fmt.Errorf("invalid F01 parameters")
	}
	return cfg, nil
}

func toSerpensConfig(p F08Parameters) (fontanafamily.SerpensConfig, error) {
	direction := fontanafamily.Forward
	if strings.EqualFold(p.Direction, "REVERSE") {
		direction = fontanafamily.Reverse
	}
	runes := []rune(p.Alphabet)
	cfg := fontanafamily.SerpensConfig{Capacity: p.Capacity, Alphabet: runes, Start: p.Start, Direction: direction}
	if p.EmptyMarker != "" {
		cfg.EmptyMarker = []rune(p.EmptyMarker)[0]
	}
	return cfg, cfg.Validate()
}

func f01Candidates(cfg speculumf01.Config, state speculumf01.State) []RetrievedItem {
	seen := map[RetrievedItem]bool{}
	for _, order := range []speculumf01.Direction{speculumf01.InnerToOuter, speculumf01.OuterToInner} {
		alt := cfg
		alt.Order = order
		for n := 1; n <= cfg.NumRings; n++ {
			decoded, err := alt.DecodeFull(state, n)
			if err == nil {
				seen[RetrievedItem(decoded)] = true
			}
		}
	}
	out := make([]RetrievedItem, 0, len(seen))
	for item := range seen {
		out = append(out, item)
	}
	return sortCandidates(out)
}

func applyContext(candidates []RetrievedItem, ctx *ContextKnowledge) []RetrievedItem {
	if ctx == nil || len(ctx.Allowed) == 0 {
		return sortCandidates(append([]RetrievedItem(nil), candidates...))
	}
	allowed := map[RetrievedItem]bool{}
	for _, candidate := range ctx.Allowed {
		allowed[candidate] = true
	}
	var out []RetrievedItem
	for _, candidate := range candidates {
		if allowed[candidate] {
			out = append(out, candidate)
		}
	}
	return sortCandidates(out)
}

func sortCandidates(candidates []RetrievedItem) []RetrievedItem {
	out := append([]RetrievedItem(nil), candidates...)
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func sortedIndices(values map[Index]Cue) []Index {
	out := make([]Index, 0, len(values))
	for idx := range values {
		out = append(out, idx)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func valuesFromIndexed(values map[Index]Cue) []string {
	idxs := sortedIndices(values)
	out := make([]string, 0, len(idxs))
	for _, idx := range idxs {
		out = append(out, string(values[idx]))
	}
	return out
}

func firstParamID(spec MechanismSpec) string {
	if len(spec.Parameters) == 0 {
		return ""
	}
	return spec.Parameters[0].ID
}

func flattenVisibleCues(visible []string) []string {
	var out []string
	for _, cue := range visible {
		if cue != "" && cue != "." {
			out = append(out, cue)
		}
	}
	return out
}

func runePtr(r rune) *rune {
	x := r
	return &x
}
