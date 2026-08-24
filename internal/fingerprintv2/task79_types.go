package fingerprintv2

// Task79Result is the versioned, machine-readable freeze candidate.  Empty or
// unavailable fields are represented by explicit status strings, never by an
// implied successful test.
type Task79Result struct {
	Version          string                   `json:"version"`
	InputAudit       []AuditRecord            `json:"task75_task77_audit"`
	MetadataAudit    MetadataAudit            `json:"ivtff_metadata_audit"`
	LineProfiles     []LineProfile            `json:"line_profiles"`
	Metrics          []FreezeMetric           `json:"metric_registry"`
	NullRegistry     []NullModelRegistryEntry `json:"hierarchical_null_registry"`
	StabilityMatrix  []StabilityAssessment    `json:"stability_matrix"`
	RedundancyMatrix []RedundancyRow          `json:"redundancy_matrix"`
	CoverageAudit    []CoverageAssessment     `json:"alternative_explanation_coverage"`
	NegativeEvidence []NegativeEvidence       `json:"negative_evidence_registry"`
	Discriminative   []DiscriminativeResult   `json:"discriminative_validation"`
	Segmentation     SegmentationResult       `json:"segmentation"`
	FreezeManifest   FreezeManifest           `json:"freeze_manifest"`
	Verdicts         []Task79Verdict          `json:"verdicts"`
	Occurrences      []OccurrenceMetadata     `json:"-"`
}

type AuditRecord struct {
	Block      string `json:"block"`
	Status     string `json:"status"`
	Evidence   string `json:"evidence"`
	Limitation string `json:"limitation"`
}

type MetadataAudit struct {
	Status                  string   `json:"status"`
	Folios                  int      `json:"folios"`
	Loci                    int      `json:"loci"`
	Lines                   int      `json:"lines"`
	Tokens                  int      `json:"tokens"`
	Transitions             int      `json:"transitions"`
	DuplicateLineIDs        int      `json:"duplicate_line_ids"`
	NestingViolations       int      `json:"nesting_violations"`
	MissingFolio            int      `json:"missing_folio"`
	MissingLocusType        int      `json:"missing_locus_type"`
	MissingSection          int      `json:"missing_section"`
	MissingCurrier          int      `json:"missing_currier"`
	MissingHand             int      `json:"missing_hand"`
	MissingI                int      `json:"missing_i"`
	MissingX                int      `json:"missing_x"`
	KnownF116r37Occurrences int      `json:"known_f116r_37_occurrences"`
	CorrectionLayer         string   `json:"correction_layer"`
	Issues                  []string `json:"issues"`
}

type OccurrenceMetadata struct {
	Position           int     `json:"absolute_token_position"`
	Token              string  `json:"token"`
	Folio              string  `json:"folio"`
	FolioSide          string  `json:"folio_side"`
	LocusID            string  `json:"locus_identifier"`
	LocusType          string  `json:"locus_type"`
	LineID             string  `json:"line_identifier"`
	IndexInLine        int     `json:"index_in_line"`
	NormalizedPosition float64 `json:"normalized_token_position"`
	LineLength         int     `json:"line_length"`
	IndexInLocus       int     `json:"index_in_locus"`
	IndexInFolio       int     `json:"index_in_folio"`
	ParagraphID        int     `json:"paragraph_id"`
	ParagraphStart     bool    `json:"paragraph_start"`
	Section            string  `json:"section"`
	Currier            string  `json:"currier_language"`
	Scribe             string  `json:"scribe"`
	Quire              string  `json:"quire"`
	IVTFFI             string  `json:"ivtff_i"`
	IVTFFX             string  `json:"ivtff_x"`
	LabelStatus        string  `json:"label_text_status"`
	MissingStatus      string  `json:"missing_status"`
}

type LineProfile struct {
	LineID                 string  `json:"line_id"`
	Folio                  string  `json:"folio"`
	LocusID                string  `json:"locus_id"`
	LocusType              string  `json:"locus_type"`
	Section                string  `json:"section"`
	Currier                string  `json:"currier"`
	Scribe                 string  `json:"scribe"`
	TokenCount             int     `json:"token_count"`
	CharacterCount         int     `json:"character_count"`
	VocabularySize         int     `json:"vocabulary_size"`
	Diversity              float64 `json:"diversity"`
	ExactRepetitionRate    float64 `json:"exact_repetition_rate"`
	NearEditRepetitionRate float64 `json:"near_edit_repetition_rate"`
	TransitionEntropy      float64 `json:"transition_entropy"`
	TokenEntropy           float64 `json:"token_entropy"`
	FirstToken             string  `json:"first_token"`
	SecondToken            string  `json:"second_token"`
	PenultimateToken       string  `json:"penultimate_token"`
	FinalToken             string  `json:"final_token"`
	ParagraphStart         bool    `json:"paragraph_start"`
}

type FreezeMetric struct {
	MetricID               string         `json:"metric_id"`
	MetricVersion          string         `json:"metric_version"`
	Family                 string         `json:"family"`
	Definition             string         `json:"definition"`
	UnitOfAnalysis         string         `json:"unit_of_analysis"`
	Inputs                 []string       `json:"inputs"`
	Parameters             map[string]any `json:"parameters"`
	ObservedValue          float64        `json:"observed_value"`
	Uncertainty            string         `json:"uncertainty"`
	NullModels             []string       `json:"null_models"`
	EffectSize             float64        `json:"effect_size"`
	PValue                 float64        `json:"p_value"`
	QValue                 float64        `json:"q_value"`
	PartitionStability     string         `json:"partition_stability"`
	TranscriptionStability string         `json:"transcription_stability"`
	ParameterSensitivity   string         `json:"parameter_sensitivity"`
	RedundancyClass        string         `json:"redundancy_class"`
	CoverageRole           []string       `json:"coverage_role"`
	ComparisonEligibility  string         `json:"comparison_eligibility"`
	NegativeEvidenceStatus string         `json:"negative_evidence_status"`
	ImplementationVersion  string         `json:"implementation_version"`
	Status                 string         `json:"status"`
	Limitations            string         `json:"limitations"`
}

type StabilityAssessment struct {
	MetricID   string `json:"metric_id"`
	Axis       string `json:"axis"`
	Status     string `json:"status"`
	Evidence   string `json:"evidence"`
	Limitation string `json:"limitation"`
}

type CoverageAssessment struct {
	ExplanationClass         string   `json:"explanation_class"`
	SensitiveMetrics         []string `json:"sensitive_metrics"`
	UncoveredProperties      string   `json:"uncovered_properties"`
	Controls                 string   `json:"controls"`
	NegativeConclusion       string   `json:"negative_conclusion"`
	RequiredWork             string   `json:"required_work"`
	CriticalBeforeExperiment bool     `json:"critical_before_experiment"`
}

type NegativeEvidence struct {
	MetricID            string `json:"metric_id"`
	TestedHypothesis    string `json:"tested_hypothesis"`
	Status              string `json:"status"`
	Sensitivity         string `json:"sensitivity"`
	DetectableThreshold string `json:"detectable_effect_threshold"`
	ConfidenceInterval  string `json:"confidence_interval"`
	Controls            string `json:"controls"`
	Partitions          string `json:"partitions"`
	Limitations         string `json:"limitations"`
}

type DiscriminativeResult struct {
	ControlID              string  `json:"control_id"`
	MetricID               string  `json:"metric_id"`
	PrimaryValue           float64 `json:"primary_value"`
	ControlValue           float64 `json:"control_value"`
	StandardizedDifference float64 `json:"standardized_difference"`
	Status                 string  `json:"status"`
	Limitation             string  `json:"limitation"`
}

type ChangePoint struct {
	Index int     `json:"index"`
	Score float64 `json:"score"`
	Folio string  `json:"folio"`
}
type SegmentationResult struct {
	Method            string        `json:"method"`
	Selection         string        `json:"selection"`
	Status            string        `json:"status"`
	ChangePoints      []ChangePoint `json:"change_points"`
	MetadataAgreement float64       `json:"metadata_agreement"`
	Limitations       string        `json:"limitations"`
}

type FreezeManifest struct {
	CandidateID        string   `json:"candidate_id"`
	Status             string   `json:"status"`
	CorpusSHA256       string   `json:"corpus_sha256"`
	ConfigSHA256       string   `json:"config_sha256"`
	CodeVersion        string   `json:"code_version"`
	MetricVersion      string   `json:"metric_version"`
	Seeds              []int64  `json:"seeds"`
	CoreMetrics        []string `json:"core_metrics"`
	SupportingMetrics  []string `json:"supporting_metrics"`
	UnfrozenExtensions []string `json:"unfrozen_extensions"`
	ProhibitedClaims   []string `json:"prohibited_claims"`
	MissingDataPolicy  string   `json:"missing_data_policy"`
	ComparisonRules    string   `json:"comparison_rules"`
	DecisionBasis      string   `json:"decision_basis"`
}

type Task79Verdict struct {
	ID              string  `json:"id"`
	Value           string  `json:"value"`
	PrimaryEvidence string  `json:"primary_evidence"`
	NullComparison  string  `json:"null_comparison"`
	Stability       string  `json:"stability"`
	HeldOutResult   string  `json:"held_out_result"`
	Limitations     string  `json:"limitations"`
	FreezeImpact    string  `json:"freeze_impact"`
	EffectSize      float64 `json:"effect_size"`
}
