package mnemonicspace

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
	"strings"
)

type OperationID string

type DomainType string

const (
	TypeSymbol              DomainType = "Symbol"
	TypeSymbolSequence      DomainType = "SymbolSequence"
	TypePosition            DomainType = "Position"
	TypeIndex               DomainType = "Index"
	TypeOrientation         DomainType = "Orientation"
	TypeAlignment           DomainType = "Alignment"
	TypePath                DomainType = "Path"
	TypeExternalState       DomainType = "ExternalState"
	TypeObservation         DomainType = "Observation"
	TypeCue                 DomainType = "Cue"
	TypeInternalMemoryState DomainType = "InternalMemoryState"
	TypeRetrievedItem       DomainType = "RetrievedItem"
	TypeStep                DomainType = "Step"
)

type Carrier string

const (
	CarrierMessage       Carrier = "M"
	CarrierExternalState Carrier = "E"
	CarrierGeometry      Carrier = "G"
	CarrierHistory       Carrier = "H"
	CarrierConvention    Carrier = "K"
	CarrierInternal      Carrier = "I"
	CarrierContext       Carrier = "C"
)

type HistoricalStatus string

const (
	StatusFExact         HistoricalStatus = "F-EXACT"
	StatusFProfile       HistoricalStatus = "F-PROFILE"
	StatusFComposition   HistoricalStatus = "F-COMPOSITION"
	StatusMRestricted    HistoricalStatus = "M-RESTRICTED"
	StatusMExtended      HistoricalStatus = "M-EXTENDED"
	StatusGenericControl HistoricalStatus = "GENERIC_CONTROL"
	StatusReferenceOnly  HistoricalStatus = "REFERENCE_ONLY"
	StatusForbidden      HistoricalStatus = "FORBIDDEN"
)

type Family string

const (
	FamilyF01                    Family = "F01_LITERAL_ROTATIONAL_STORAGE"
	FamilyF08                    Family = "F08_ORDERED_POSITIONAL_STORAGE"
	FamilyF11                    Family = "F11_INDEXED_OPAQUE_CUE"
	FamilyF12                    Family = "F12_TEMPORAL_ASSOCIATIVE_CUE"
	FamilyRestrictedRotation     Family = "M_RESTRICTED_ROTATION_INDEX"
	FamilyRestrictedAssociation  Family = "M_RESTRICTED_STORAGE_ASSOCIATION"
	FamilyGenericLiteral         Family = "GENERIC_LITERAL_STORAGE"
	FamilyGenericCyclic          Family = "GENERIC_CYCLIC_STATE"
	FamilyGenericIndexed         Family = "GENERIC_INDEXED_LOOKUP"
	FamilyGenericCue             Family = "GENERIC_CUE_BASED"
	FamilyGenericAmbiguous       Family = "GENERIC_AMBIGUOUS"
	FamilyNegativeConvention     Family = "NEGATIVE_RANDOMIZED_CONVENTION"
	FamilyNegativePath           Family = "NEGATIVE_RANDOMIZED_PATH"
	FamilyNegativeCueAssociation Family = "NEGATIVE_RANDOMIZED_CUE_ASSOCIATION"
	FamilyNegativeIndexMapping   Family = "NEGATIVE_RANDOMIZED_INDEX_MAPPING"
)

type InformationEffect string

const (
	EffectPreserve      InformationEffect = "PRESERVE"
	EffectSelect        InformationEffect = "SELECT"
	EffectReorder       InformationEffect = "REORDER"
	EffectCombine       InformationEffect = "COMBINE"
	EffectAddContext    InformationEffect = "ADD_CONTEXT"
	EffectSignal        InformationEffect = "SIGNAL"
	EffectRecallTrigger InformationEffect = "RECALL_TRIGGER"
	EffectAmbiguate     InformationEffect = "AMBIGUATE"
)

type RetrievalKind string

const (
	RetrievalExactWithConvention RetrievalKind = "EXACT_WITH_CONVENTION"
	RetrievalCueOnly             RetrievalKind = "CUE_ONLY"
	RetrievalAmbiguous           RetrievalKind = "AMBIGUOUS"
)

type RecoveryCondition string

const (
	RecoveryFullKnowledge RecoveryCondition = "R0_FULL_KNOWLEDGE"
	RecoveryNoContext     RecoveryCondition = "R1_NO_CONTEXT"
	RecoveryNoConvention  RecoveryCondition = "R2_NO_CONVENTION"
	RecoveryNoGeometry    RecoveryCondition = "R3_NO_PATH_GEOMETRY"
	RecoveryNoHistory     RecoveryCondition = "R4_NO_HISTORY"
	RecoveryNoInternal    RecoveryCondition = "R5_NO_INTERNAL_MEMORY"
	RecoveryObservable    RecoveryCondition = "R6_OBSERVABLE_ONLY"
)

type RecoveryResultClass string

const (
	ResultExact         RecoveryResultClass = "EXACT"
	ResultPartial       RecoveryResultClass = "PARTIAL"
	ResultAmbiguitySet  RecoveryResultClass = "AMBIGUITY_SET"
	ResultCueOnly       RecoveryResultClass = "CUE_ONLY"
	ResultNoRecovery    RecoveryResultClass = "NO_RECOVERY"
	ResultNotApplicable RecoveryResultClass = "NOT_APPLICABLE"
)

type BoundaryStatus string

const (
	BoundaryNotDefined     BoundaryStatus = "NOT_DEFINED"
	BoundaryInheritedInput BoundaryStatus = "INHERITED_FROM_INPUT"
	BoundaryGenerated      BoundaryStatus = "GENERATED_BY_MECHANISM"
)

type ParameterOrigin string

const (
	OriginHistoricalFrozen  ParameterOrigin = "HISTORICAL_FROZEN"
	OriginHistoricalBounded ParameterOrigin = "HISTORICAL_BOUNDED"
	OriginGenericGrid       ParameterOrigin = "GENERIC_GRID"
	OriginControlOnly       ParameterOrigin = "CONTROL_ONLY"
)

type ErrorClass string

const (
	ErrorSubstitution       ErrorClass = "substitution"
	ErrorDeletion           ErrorClass = "deletion"
	ErrorInsertion          ErrorClass = "insertion"
	ErrorTransposition      ErrorClass = "transposition"
	ErrorBoundaryCorruption ErrorClass = "boundary_corruption"
	ErrorStateCorruption    ErrorClass = "state_corruption"
	ErrorConventionCorrupt  ErrorClass = "convention_corruption"
)

type SurfaceRole string

const (
	SurfaceLiteralSequence SurfaceRole = "LITERAL_SEQUENCE"
	SurfaceOpaqueCue       SurfaceRole = "OPAQUE_CUE"
)

type Symbol string

type Cue string

type RetrievedItem string

type Position int

type Index int

type DomainValue struct {
	Name string
	Type DomainType
}

type OperationInvocation struct {
	Operation OperationID
	Inputs    []DomainValue
	Outputs   []DomainValue
	Note      string
}

// CueConversion is a declared representation-level interpretation, not a
// Task80 primitive: an opaque Symbol identifier selected from storage is
// retyped as a Cue before association. It has no state or information effect.
type CueConversion struct {
	InputName  string
	OutputName string
	From       DomainType
	To         DomainType
	Rule       string
}

type CarrierRequirements struct {
	Encode   []Carrier
	Retrieve []Carrier
}

type RetrievalRelation struct {
	Kind              RetrievalKind
	SupportsContext   bool
	SupportsAmbiguity bool
}

type ObservableContract struct {
	SymbolSource        string
	TokenBoundary       BoundaryStatus
	LineBoundary        BoundaryStatus
	UnitBoundary        BoundaryStatus
	ExcludesHiddenState bool
}

type F01Parameters struct {
	NumRings           int
	Alphabet           string
	ReadRadius         int
	Order              string
	RingIdentityMarked bool
}

type F08Parameters struct {
	Capacity    int
	Start       int
	Direction   string
	EmptyMarker string
	Alphabet    string
}

type F11Parameters struct{}

type F12Parameters struct {
	Period       int
	InitialTick  int
	AdvanceSteps int
}

type RotationIndexParameters struct {
	Offset int
}

type StorageAssociateParameters struct {
	Capacity int
}

type ParameterSet struct {
	ID          string
	Origin      ParameterOrigin
	Frozen      bool
	Description string
	F01         *F01Parameters
	F08         *F08Parameters
	F11         *F11Parameters
	F12         *F12Parameters
	Rotation    *RotationIndexParameters
	Storage     *StorageAssociateParameters
}

type runtimeKind uint8

const (
	kindUnknown runtimeKind = iota
	kindF01
	kindF08
	kindF11
	kindF12
	kindRotationIndex
	kindStorageAssociate
)

type MechanismSpec struct {
	ID                    string
	Version               string
	Family                Family
	Status                HistoricalStatus
	Provenance            []string
	SourceModels          []string
	Encoding              []OperationInvocation
	Retrieval             []OperationInvocation
	CueConversions        []CueConversion
	StateSchema           string
	InputSchema           string
	Initialization        string
	TransitionRules       []string
	ObservationRule       string
	Serialization         ObservableContract
	RetrievalRelation     RetrievalRelation
	Carriers              CarrierRequirements
	ExpectedReversibility string
	AmbiguitySemantics    string
	ErrorSemantics        []ErrorClass
	Parameters            []ParameterSet
	SurfaceRole           SurfaceRole
	EquivalentTo          string

	runtime runtimeKind
}

func (m MechanismSpec) ParameterSet(id string) (ParameterSet, bool) {
	for _, set := range m.Parameters {
		if set.ID == id {
			return set, true
		}
	}
	return ParameterSet{}, false
}

type InputModel struct {
	ID             string
	Sequence       []Symbol
	IndexedCues    map[Index]Cue
	PositionedCues map[Position]Cue
	TimedCues      map[int]Cue
}

func (i InputModel) SequenceString() string {
	var b strings.Builder
	for _, s := range i.Sequence {
		b.WriteString(string(s))
	}
	return b.String()
}

type ConventionKnowledge struct {
	MessageLength int
	CueMeanings   map[Cue][]RetrievedItem
}

type GeometryKnowledge struct {
	Path []Position
}

type HistoryKnowledge struct {
	Steps int
}

type InternalMemoryState struct {
	Associations map[Cue][]RetrievedItem
}

type ContextKnowledge struct {
	Allowed []RetrievedItem
}

type RecoveryEnvironment struct {
	Geometry       *GeometryKnowledge
	History        *HistoryKnowledge
	Convention     *ConventionKnowledge
	InternalMemory *InternalMemoryState
	Context        *ContextKnowledge
}

type RetrievalRequest struct {
	Condition      RecoveryCondition
	TargetIndex    *Index
	TargetPosition *Position
}

type F01State struct {
	Offsets []int `json:"offsets"`
}

type F08State struct {
	Holes []string `json:"holes"`
}

type F11State struct {
	Slots map[Index]Cue `json:"slots"`
}

type F12State struct {
	Period int         `json:"period"`
	Tick   int         `json:"tick"`
	Cues   map[int]Cue `json:"cues"`
}

type RotationIndexState struct {
	Slots  []Cue `json:"slots"`
	Offset int   `json:"offset"`
}

type StorageAssociateState struct {
	Slots map[Position]Cue `json:"slots"`
}

type ExternalState struct {
	F01              *F01State
	F08              *F08State
	F11              *F11State
	F12              *F12State
	RotationIndex    *RotationIndexState
	StorageAssociate *StorageAssociateState
}

type ObservationStream struct {
	Symbols []string `json:"symbols,omitempty"`
	Cues    []string `json:"cues,omitempty"`
}

type ObservableDocument struct {
	Symbols       []string          `json:"symbols"`
	TokenBoundary BoundaryStatus    `json:"token_boundary"`
	LineBoundary  BoundaryStatus    `json:"line_boundary"`
	UnitBoundary  BoundaryStatus    `json:"unit_boundary"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

func (d ObservableDocument) Checksum() string {
	payload, _ := json.Marshal(d)
	sum := sha256.Sum256(payload)
	return hex.EncodeToString(sum[:])
}

type InformationEvent struct {
	Carrier Carrier           `json:"carrier"`
	Effect  InformationEffect `json:"effect"`
	Detail  string            `json:"detail"`
}

type PreparedRun struct {
	MechanismID    string
	ParameterSetID string
	State          ExternalState
	Observation    ObservationStream
	Document       ObservableDocument
	Trace          []InformationEvent
}

type RecoveryResult struct {
	Class        RecoveryResultClass
	Value        RetrievedItem
	Cue          Cue
	Candidates   []RetrievedItem
	UsedCarriers []Carrier
	Detail       string
}

type Execution struct {
	Prepared PreparedRun
	Recovery RecoveryResult
}

type CollisionSample struct {
	InputID  string
	Intended RetrievedItem
	Document ObservableDocument
}

type Collision struct {
	DocumentChecksum string
	InputIDs         []string
	Intendeds        []RetrievedItem
}

func DetectObservableCollisions(samples []CollisionSample) []Collision {
	type bucket struct {
		inputs   []string
		intended []RetrievedItem
	}
	groups := map[string]*bucket{}
	for _, sample := range samples {
		h := sample.Document.Checksum()
		group := groups[h]
		if group == nil {
			group = &bucket{}
			groups[h] = group
		}
		group.inputs = append(group.inputs, sample.InputID)
		group.intended = append(group.intended, sample.Intended)
	}
	var out []Collision
	for checksum, group := range groups {
		seen := map[RetrievedItem]bool{}
		for _, intended := range group.intended {
			seen[intended] = true
		}
		if len(seen) < 2 {
			continue
		}
		sort.Strings(group.inputs)
		items := make([]string, 0, len(seen))
		for intended := range seen {
			items = append(items, string(intended))
		}
		sort.Strings(items)
		collision := Collision{DocumentChecksum: checksum, InputIDs: group.inputs}
		for _, item := range items {
			collision.Intendeds = append(collision.Intendeds, RetrievedItem(item))
		}
		out = append(out, collision)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].DocumentChecksum < out[j].DocumentChecksum })
	return out
}
