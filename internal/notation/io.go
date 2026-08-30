package notation

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
)

func WriteFingerprintJSON(w io.Writer, fp Fingerprint) error {
	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	return e.Encode(fp)
}
func ReadFingerprintJSON(r io.Reader) (Fingerprint, error) {
	var fp Fingerprint
	err := json.NewDecoder(r).Decode(&fp)
	return fp, err
}

func WriteMetricsTSV(w io.Writer, fp Fingerprint) error {
	c := csv.NewWriter(w)
	c.Comma = '\t'
	defer c.Flush()
	if err := c.Write([]string{"metric_id", "family", "regime", "value", "status", "reason"}); err != nil {
		return err
	}
	for _, m := range SortedMetrics(fp) {
		v := ""
		if m.Status == Comparable || m.Status == PartiallyComparable {
			v = strconv.FormatFloat(m.Value, 'g', 12, 64)
		}
		if err := c.Write([]string{m.MetricID, m.Family, m.Regime, v, string(m.Status), m.Reason}); err != nil {
			return err
		}
	}
	return c.Error()
}

func ReadScalesTSV(r io.Reader) ([]Scale, error) {
	c := csv.NewReader(r)
	c.Comma = '\t'
	c.FieldsPerRecord = -1
	head, err := c.Read()
	if err != nil {
		return nil, err
	}
	idx := map[string]int{}
	for i, h := range head {
		idx[h] = i
	}
	for _, k := range []string{"metric_id", "regime", "center", "spread"} {
		if _, ok := idx[k]; !ok {
			return nil, fmt.Errorf("scale missing column %s", k)
		}
	}
	var out []Scale
	for {
		row, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		center, e := strconv.ParseFloat(row[idx["center"]], 64)
		if e != nil {
			return nil, e
		}
		spread, e := strconv.ParseFloat(row[idx["spread"]], 64)
		if e != nil {
			return nil, e
		}
		out = append(out, Scale{row[idx["metric_id"]], row[idx["regime"]], center, spread})
	}
	return out, nil
}

func WriteRarefactionTSV(w io.Writer, rows []RarefactionRow) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "corpus_id\trepresentation_id\tfamily\tmetric_id\tregime\tcheckpoint_requested\tcheckpoint_actual\treplicate\tseed\tvalue\tcomparable")
	for _, r := range rows {
		v := ""
		if r.Comparable {
			v = strconv.FormatFloat(r.Value, 'g', 12, 64)
		}
		fmt.Fprintf(b, "%s\t%s\t%s\t%s\t%s\t%d\t%d\t%d\t%d\t%s\t%t\n", r.CorpusID, r.RepresentationID, r.Family, r.MetricID, r.Regime, r.CheckpointRequested, r.CheckpointActual, r.Replicate, r.Seed, v, r.Comparable)
	}
	return nil
}

func WriteRarefactionSummaryTSV(w io.Writer, rows []RarefactionSummaryRow) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "corpus_id\trepresentation_id\tfamily\tmetric_id\tregime\tcheckpoint\tmean\tmedian\tsd\tci_low\tci_high\tn_valid")
	for _, r := range rows {
		fmt.Fprintf(b, "%s\t%s\t%s\t%s\t%s\t%d\t%.12g\t%.12g\t%.12g\t%.12g\t%.12g\t%d\n", r.CorpusID, r.RepresentationID, r.Family, r.MetricID, r.Regime, r.Checkpoint, r.Mean, r.Median, r.SD, r.CILow, r.CIHigh, r.NValid)
	}
	return nil
}

func WriteBootstrapTSV(w io.Writer, rows []BootstrapRow) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "corpus_id\trepresentation_id\tfamily\tmetric_id\tregime\testimate\tbootstrap_mean\tbootstrap_sd\tci_level\tci_low\tci_high\tn_valid")
	for _, r := range rows {
		fmt.Fprintf(b, "%s\t%s\t%s\t%s\t%s\t%.12g\t%.12g\t%.12g\t%.12g\t%.12g\t%.12g\t%d\n", r.CorpusID, r.RepresentationID, r.Family, r.MetricID, r.Regime, r.Estimate, r.BootstrapMean, r.BootstrapSD, r.CILevel, r.CILow, r.CIHigh, r.NValid)
	}
	return nil
}

func ReadBootstrapTSV(r io.Reader) ([]BootstrapRow, error) {
	c := csv.NewReader(r)
	c.Comma = '\t'
	c.FieldsPerRecord = -1
	if _, err := c.Read(); err != nil {
		return nil, err
	}
	var out []BootstrapRow
	for {
		row, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		est, _ := strconv.ParseFloat(row[5], 64)
		bm, _ := strconv.ParseFloat(row[6], 64)
		bsd, _ := strconv.ParseFloat(row[7], 64)
		cl, _ := strconv.ParseFloat(row[8], 64)
		lo, _ := strconv.ParseFloat(row[9], 64)
		hi, _ := strconv.ParseFloat(row[10], 64)
		nv, _ := strconv.Atoi(row[11])
		out = append(out, BootstrapRow{CorpusID: row[0], RepresentationID: row[1], Family: row[2], MetricID: row[3], Regime: row[4], Estimate: est, BootstrapMean: bm, BootstrapSD: bsd, CILevel: cl, CILow: lo, CIHigh: hi, NValid: nv})
	}
	return out, nil
}

func WriteDistributionsTSV(w io.Writer, rows []DistributionPoint) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "corpus_id\trepresentation_id\tmetric_id\tsupport_id\tbin_or_category\tvalue\tprobability\tcomparable\treason")
	for _, r := range rows {
		fmt.Fprintf(b, "%s\t%s\t%s\t%s\t%s\t%.12g\t%.12g\t%t\t%s\n", r.CorpusID, r.RepresentationID, r.MetricID, r.SupportID, r.BinOrCategory, r.Value, r.Probability, r.Comparable, r.Reason)
	}
	return nil
}

func ReadDistributionsTSV(r io.Reader) ([]DistributionPoint, error) {
	c := csv.NewReader(r)
	c.Comma = '\t'
	c.FieldsPerRecord = -1
	if _, err := c.Read(); err != nil {
		return nil, err
	}
	var out []DistributionPoint
	for {
		row, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		v, _ := strconv.ParseFloat(row[5], 64)
		p, _ := strconv.ParseFloat(row[6], 64)
		out = append(out, DistributionPoint{CorpusID: row[0], RepresentationID: row[1], MetricID: row[2], SupportID: row[3], BinOrCategory: row[4], Value: v, Probability: p, Comparable: row[7] == "true", Reason: row[8]})
	}
	return out, nil
}

func WriteMetricOutputTypesTSV(w io.Writer, rows []MetricOutputTypeRow) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "metric_id\toutput_type\tnote")
	for _, r := range rows {
		fmt.Fprintf(b, "%s\t%s\t%s\n", r.MetricID, r.OutputType, r.Note)
	}
	return nil
}

func WriteCalibrationScalesTSV(w io.Writer, rows []CalibrationScale) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "metric_id\tmetric_version\tfamily\tsupport_regime\tcheckpoint\testimator\tn\tmedian\tmad\tiqr\tscale\tstatus")
	for _, r := range rows {
		fmt.Fprintf(b, "%s\t%s\t%s\t%s\t%d\t%s\t%d\t%.12g\t%.12g\t%.12g\t%.12g\t%s\n", r.MetricID, r.MetricVersion, r.Family, r.SupportRegime, r.Checkpoint, r.Estimator, r.N, r.Median, r.MAD, r.IQR, r.Scale, r.Status)
	}
	return nil
}

func ReadCalibrationScalesTSV(r io.Reader) ([]CalibrationScale, error) {
	c := csv.NewReader(r)
	c.Comma = '\t'
	c.FieldsPerRecord = -1
	if _, err := c.Read(); err != nil {
		return nil, err
	}
	var out []CalibrationScale
	for {
		row, err := c.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		ckpt, _ := strconv.Atoi(row[4])
		n, _ := strconv.Atoi(row[6])
		med, _ := strconv.ParseFloat(row[7], 64)
		mad, _ := strconv.ParseFloat(row[8], 64)
		iqr, _ := strconv.ParseFloat(row[9], 64)
		scale, _ := strconv.ParseFloat(row[10], 64)
		out = append(out, CalibrationScale{MetricID: row[0], MetricVersion: row[1], Family: row[2], SupportRegime: row[3], Checkpoint: ckpt, Estimator: row[5], N: n, Median: med, MAD: mad, IQR: iqr, Scale: scale, Status: row[11]})
	}
	return out, nil
}

func WriteVMReferenceTSV(w io.Writer, rows []VMReferenceRow) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "metric_id\tmetric_version\tfamily\tsupport_regime\tcheckpoint\tvalue\toutput_type\tcomparability\tprovenance")
	for _, r := range rows {
		v := ""
		if r.Comparability == Comparable || r.Comparability == PartiallyComparable {
			v = strconv.FormatFloat(r.Value, 'g', 12, 64)
		}
		fmt.Fprintf(b, "%s\t%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\n", r.MetricID, r.MetricVersion, r.Family, r.SupportRegime, r.Checkpoint, v, r.OutputType, r.Comparability, r.Provenance)
	}
	return nil
}

func WriteCurvesTSV(w io.Writer, curves []CurvePoint) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "curve_id\tcheckpoint\tvalue\tstatus\treason")
	for _, c := range curves {
		v := ""
		if c.Status == Comparable {
			v = strconv.FormatFloat(c.Value, 'g', 12, 64)
		}
		fmt.Fprintf(b, "%s\t%d\t%s\t%s\t%s\n", c.CurveID, c.Checkpoint, v, c.Status, c.Reason)
	}
	return nil
}

func WriteComparisonTSV(w io.Writer, rows []ComparisonRow, families []FamilyDistance) error {
	b := bufio.NewWriter(w)
	defer b.Flush()
	fmt.Fprintln(b, "metric_id\tfamily\tvm_value\tcandidate_value\tcandidate_ci\tdistance\tcomparable\treason_if_not_comparable")
	for _, r := range rows {
		id := r.MetricID
		if r.Regime != "" {
			id += "[" + r.Regime + "]"
		}
		fmt.Fprintf(b, "%s\t%s\t%.12g\t%.12g\t\t%.12g\t%s\t%s\n", id, r.Family, r.Reference, r.Candidate, r.Distance, r.Status, r.Reason)
	}
	for _, f := range families {
		fmt.Fprintf(b, "d_%s\t%s\t\t\t\t%.12g\t%s\t%s\n", f.Family, f.Family, f.Distance, f.Status, f.Reason)
	}
	return nil
}
