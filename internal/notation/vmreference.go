package notation

// VMReferenceRow is one row of VM_REFERENCE_V2.tsv (B02 section 36).
type VMReferenceRow struct {
	MetricID, MetricVersion, Family, SupportRegime string
	Checkpoint                                     int
	Value                                           float64
	OutputType                                      OutputType
	Comparability                                   Status
	Provenance                                      string
}

// BuildVMReference joins the full-corpus fingerprint with the frozen output
// type registry into the versioned VM_REFERENCE_V2 schema. checkpoint is the
// corpus's own actual token count (the VM reference is computed once, at
// its full observed size, not rarefied).
func BuildVMReference(fp Fingerprint, checkpoint int) []VMReferenceRow {
	types := map[string]OutputType{}
	for _, t := range MetricOutputTypes() {
		types[t.MetricID] = t.OutputType
	}
	out := make([]VMReferenceRow, 0, len(fp.Metrics))
	for _, m := range fp.Metrics {
		out = append(out, VMReferenceRow{
			MetricID: m.MetricID, MetricVersion: MetricRegistryVersion, Family: m.Family, SupportRegime: m.Regime,
			Checkpoint: checkpoint, Value: m.Value, OutputType: types[m.MetricID], Comparability: m.Status, Provenance: fp.InputSHA256,
		})
	}
	return out
}
