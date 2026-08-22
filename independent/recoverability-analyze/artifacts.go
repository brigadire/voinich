package main

import (
	"bufio"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"zcore.dev/voinich/internal/mechanismspace"
)

type experimentData struct {
	cand       candidate
	corpus     string
	model      *decoder
	test       mechanismspace.Corpus
	encoded    encoded
	clean      recovery
	hp, hc, mi float64
	ri         float64
}

type curveKey struct {
	candidate, corpus, channel string
	rate                       float64
}
type curveAgg struct {
	sumGlyph, sumToken, sumExact float64
	n                            int
}

func runExperiments(cs []candidate) error {
	files := map[string][]string{
		"CLEAN_RECOVERABILITY.tsv":             {"candidate\tcorpus\tpartition\tinput_units\toutput_tokens\tclean_glyph_recovery\tclean_token_recovery\tsequence_edit_distance\tnormalized_char_edit_distance\tword_error_rate\texact_message_recovery\trecovered_plaintext_entropy\tdecoder_level\tmeasurement"},
		"INFORMATION_RETENTION.tsv":            {"candidate\tcorpus\tpartition\tpaired_units\tH_plaintext_bits\tH_conditional_bits\tmutual_information_bits\tR_I\testimator\tlimitation"},
		"PREIMAGE_MULTIPLICITY.tsv":            {"candidate\tcorpus\tblock_length\tlog2_plausible_preimages\tlog2_N_per_input\texactly_counted_units\tbound_status\tmeasurement"},
		"AMBIGUITY_GROWTH.tsv":                 {"candidate\tcorpus\tblock_length\tlog2_N\tlog2_N_per_input\tgrowth_class\tsource"},
		"ERROR_RECOVERABILITY.tsv":             {"candidate\tcorpus\terror_channel\terror_rate\treplicate\tseed\tinjected_errors\tglyph_recovery\ttoken_recovery\tword_error_rate\texact_message_recovery\toutput_tokens"},
		"ERROR_PROPAGATION.tsv":                {"candidate\tcorpus\terror_type\terror_position\tposition_fraction\tlocation_class\tseed\tdamaged_units\tcorrect_units_after_error\trecovery_after_error\tL_sync\tcatastrophic_desync\tclean_output_tokens\tdamaged_output_tokens\tpropagation_baseline"},
		"ERROR_DETECTABILITY.tsv":              {"candidate\tcorpus\terror_type\terror_position\tcorrupted_form_valid\tdetectable\tsilent_error\tdecoded_unit_matches\tmeasurement"},
		"RESYNCHRONIZATION.tsv":                {"candidate\tcorpus\terror_type\terror_position\tL_sync\tcatastrophic_desync\trecovery_after_error\tmeasurement"},
		"TRANSCRIPTION_CONFLATION.tsv":         {"candidate\tcorpus\tconflation_fraction\treplicate\tseed\tconflated_class_pairs\tchanged_glyphs\tglyph_recovery\ttoken_recovery\tword_error_rate\texact_message_recovery\tdecoder"},
		"TRANSCRIPTION_SPLITTING.tsv":          {"candidate\tcorpus\tsplitting_fraction\treplicate\tseed\tsplit_classes\tchanged_glyphs\tglyph_recovery_frozen\ttoken_recovery_frozen\tglyph_recovery_collapsed_oracle\tword_error_rate_frozen\tdecoder"},
		"SEGMENTATION_DAMAGE.tsv":              {"candidate\tcorpus\tcondition\treplicate\tseed\tboundary_operations\tobserved_tokens\tboundary_precision\tboundary_recall\tboundary_f1\tglyph_recovery\ttoken_recovery\tword_error_rate\tdecoder"},
		"RESET_EXPERIMENT.tsv":                 {"candidate\tcorpus\treset_condition\treset_interval_tokens\terror_type\terror_position\tseed\tdamaged_units\trecovery_after_error\tL_sync\tcatastrophic_desync\tglyph_recovery\ttoken_recovery\tcheckpoint_observation"},
		"CASCADE_DAMAGE.tsv":                   {"candidate\tcorpus\treplicate\tseed\tcopying_errors\tconflated_pairs\tsegmentation_errors\tglyph_recovery\ttoken_recovery\tword_error_rate\texact_message_recovery\tdamage_order"},
		"PLAINTEXT_LANGUAGE_PRIOR.tsv":         {"candidate\tcorpus\tdecoder_condition\tmodel_corpus\tglyph_recovery\ttoken_recovery\tword_error_rate\texact_message_recovery\ttraining_scope"},
		"GENERATOR_CONTROL.tsv":                {"candidate\tcorpus\tseed\tcontrol\tinput_units\tglyph_recovery_against_original\ttoken_recovery_against_original\tword_error_rate\texact_message_recovery\toutput_grammar"},
		"ORACLE_DECOMPOSITION.tsv":             {"candidate\tcorpus\tcondition\toracle_key\toracle_boundaries\toracle_error_positions\tglyph_recovery\ttoken_recovery\tword_error_rate\texact_message_recovery\tnote"},
		"FINGERPRINT_INFORMATION_FRONTIER.tsv": {"candidate\tcorpus\ttask66_source_mechanism\ttask66_artifact_basis\ttask66_family_count\ttask66_families_above_0_15\tfingerprint_compatibility\tmedian_task66_progress\ttask66_family_progress_vector\ttask67_clean_glyph_recovery\ttask67_clean_token_recovery\ttask67_R_I\ton_pareto_front"},
		"RECOVERABILITY_PARETO.tsv":            {"candidate\tmean_fingerprint_compatibility\tmean_clean_glyph_recovery\tmean_clean_token_recovery\tmean_R_I\ton_pareto_front\tdominated_by"},
		"FINAL_CLASSIFICATION.tsv":             {"mechanism\tclean_recoverability\tinformation_retention\tambiguity_class\terror_fragility\tsynchronization_class\ttranscription_fragility\tsegmentation_fragility\tplaintext_dependence\tfingerprint_compatibility\tfinal_classification"},
	}

	var data []*experimentData
	for _, c := range corpora {
		full, err := mechanismspace.LoadNatural(c.Name, c.Path)
		if err != nil {
			return err
		}
		parts := splitParts(full)
		for _, cand := range cs {
			model := trainDecoder(cand, parts["TRAIN"], parts["VALIDATION"])
			for _, partition := range []string{"TRAIN", "VALIDATION", "TEST"} {
				block := limitCorpus(parts[partition], testWords)
				e := encodeAligned(cand, block)
				model.setCipherKey(e.cipher)
				r := measureRecovery(e.plain, model.decode(e.cipher))
				level := "LEVEL_3"
				if cand.Name == "M0_IDENTITY" || cand.Name == "M2_HOMOPHONY_H2" {
					level = "LEVEL_1"
				}
				files["CLEAN_RECOVERABILITY.tsv"] = append(files["CLEAN_RECOVERABILITY.tsv"], fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%.8f\t%.8f\t%.0f\t%.8f\t%.8f\t%.0f\t%.8f\t%s\tactual encode->decode", cand.Name, c.Name, partition, len(e.plain), len(e.cipher), r.glyph, r.token, r.seqEdit, r.normChar, r.wer, r.exact, r.entropy, level))
				if partition == "TEST" {
					for _, blockSize := range []int{8, 16, 32, 64, 128} {
						n := min(blockSize, min(len(e.plain), len(e.cipher)))
						br := measureRecovery(e.plain[:n], model.decode(e.cipher[:n]))
						files["CLEAN_RECOVERABILITY.tsv"] = append(files["CLEAN_RECOVERABILITY.tsv"], fmt.Sprintf("%s\t%s\tTEST_BLOCK_%d\t%d\t%d\t%.8f\t%.8f\t%.0f\t%.8f\t%.8f\t%.0f\t%.8f\t%s\tactual short-block encode->decode", cand.Name, c.Name, blockSize, n, n, br.glyph, br.token, br.seqEdit, br.normChar, br.wer, br.exact, br.entropy, level))
					}
					hp, hc, mi, ri := pluginInformation(e)
					files["INFORMATION_RETENTION.tsv"] = append(files["INFORMATION_RETENTION.tsv"], fmt.Sprintf("%s\t%s\tTEST\t%d\t%.8f\t%.8f\t%.8f\t%.8f\tempirical paired-unit plug-in\tfinite TEST sample; unit alignment follows frozen encoder boundaries", cand.Name, c.Name, min(len(e.plain), len(e.cipher)), hp, hc, mi, ri))
					data = append(data, &experimentData{cand: cand, corpus: c.Name, model: model, test: block, encoded: e, clean: r, hp: hp, hc: hc, mi: mi, ri: ri})
				}
			}
		}
	}

	writeAmbiguity(files, data)
	fmt.Fprintln(os.Stderr, "task67: clean/ambiguity complete")
	curves := runErrorExperiments(files, data)
	fmt.Fprintln(os.Stderr, "task67: stochastic corruption complete")
	writeCurves(files, curves)
	runSingleErrors(files, data)
	fmt.Fprintln(os.Stderr, "task67: single-error propagation complete")
	runTranscription(files, data)
	fmt.Fprintln(os.Stderr, "task67: transcription controls complete")
	runSegmentation(files, data)
	fmt.Fprintln(os.Stderr, "task67: segmentation complete")
	runResets(files, data)
	runSecondaryControls(files, data)
	fmt.Fprintln(os.Stderr, "task67: reset/cascade/oracle controls complete")
	if err := writeFrontier(files, data); err != nil {
		return err
	}
	writeClassification(files, data, curves)

	for name, rows := range files {
		if err := writeTSV(name, rows); err != nil {
			return err
		}
	}
	return nil
}

func writeAmbiguity(files map[string][]string, data []*experimentData) {
	for _, d := range data {
		for _, n := range []int{8, 16, 32, 64, 128} {
			n = min(n, len(d.encoded.cipher))
			logN, exactUnits := 0.0, 0
			for _, tok := range d.encoded.cipher[:n] {
				pcs := d.model.counts[tokenKey(tok)]
				multiplicity := len(pcs)
				if multiplicity == 0 {
					multiplicity = 1
				} else {
					exactUnits++
				}
				logN += math.Log2(float64(multiplicity))
			}
			status := "OBSERVED_CODEBOOK_LOWER_BOUND"
			if exactUnits == n && n <= 16 {
				status = "EXACT_WITHIN_TRAIN_VALIDATION_CODEBOOK"
			}
			files["PREIMAGE_MULTIPLICITY.tsv"] = append(files["PREIMAGE_MULTIPLICITY.tsv"], fmt.Sprintf("%s\t%s\t%d\t%.8f\t%.8f\t%d\t%s\tproduct of independently observed local preimage sets", d.cand.Name, d.corpus, n, logN, logN/float64(max(1, n)), exactUnits, status))
			growth := "SUBEXPONENTIAL_IN_OBSERVED_BLOCK"
			if logN/float64(max(1, n)) > .05 {
				growth = "EXPONENTIAL_OBSERVED_LOWER_BOUND"
			}
			files["AMBIGUITY_GROWTH.tsv"] = append(files["AMBIGUITY_GROWTH.tsv"], fmt.Sprintf("%s\t%s\t%d\t%.8f\t%.8f\t%s\tTRAIN+VALIDATION codebook applied to TEST ciphertext", d.cand.Name, d.corpus, n, logN, logN/float64(max(1, n)), growth))
		}
	}
}

func runErrorExperiments(files map[string][]string, data []*experimentData) map[curveKey]*curveAgg {
	channels := []string{"GLYPH_SUBSTITUTION", "GLYPH_DELETION", "GLYPH_INSERTION", "TOKEN_BOUNDARY_INSERTION", "TOKEN_BOUNDARY_DELETION", "TOKEN_MERGE", "TOKEN_SPLIT"}
	rates := []float64{0, .001, .0025, .005, .01, .02, .05}
	agg := map[curveKey]*curveAgg{}
	for _, d := range data {
		for _, channel := range channels {
			for _, rate := range rates {
				for rep := 0; rep < 100; rep++ {
					seed := int64(67000000+rep*1009+int(rate*1e6)) + int64(len(channel)*97+len(d.corpus)*31+len(d.cand.Name))
					damaged, injected := corruptRate(d.encoded.cipher, channel, rate, seed)
					r := measureRecovery(d.encoded.plain, d.model.decode(damaged))
					files["ERROR_RECOVERABILITY.tsv"] = append(files["ERROR_RECOVERABILITY.tsv"], fmt.Sprintf("%s\t%s\t%s\t%.5f\t%d\t%d\t%d\t%.8f\t%.8f\t%.8f\t%.0f\t%d", d.cand.Name, d.corpus, channel, rate, rep, seed, injected, r.glyph, r.token, r.wer, r.exact, len(damaged)))
					k := curveKey{d.cand.Name, d.corpus, channel, rate}
					if agg[k] == nil {
						agg[k] = &curveAgg{}
					}
					a := agg[k]
					a.sumGlyph += r.glyph
					a.sumToken += r.token
					a.sumExact += r.exact
					a.n++
				}
			}
		}
	}
	return agg
}

func writeCurves(files map[string][]string, agg map[curveKey]*curveAgg) {
	files["RECOVERABILITY_CURVES.tsv"] = []string{"candidate\tcorpus\terror_channel\terror_rate\treplicates\tmean_glyph_recovery\tmean_token_recovery\texact_message_probability\tE90\tE50\tE10\tthreshold_basis"}
	keys := make([]curveKey, 0, len(agg))
	for k := range agg {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		a, b := keys[i], keys[j]
		if a.candidate != b.candidate {
			return a.candidate < b.candidate
		}
		if a.corpus != b.corpus {
			return a.corpus < b.corpus
		}
		if a.channel != b.channel {
			return a.channel < b.channel
		}
		return a.rate < b.rate
	})
	thresholds := map[string][3]string{}
	for _, k := range keys {
		id := k.candidate + "\x00" + k.corpus + "\x00" + k.channel
		t := thresholds[id]
		a := agg[k]
		mean := a.sumGlyph / float64(a.n)
		vals := []float64{.9, .5, .1}
		for i, target := range vals {
			if t[i] == "" && mean < target {
				t[i] = fmt.Sprintf("%.5f", k.rate)
			}
		}
		thresholds[id] = t
	}
	for _, k := range keys {
		a := agg[k]
		id := k.candidate + "\x00" + k.corpus + "\x00" + k.channel
		t := thresholds[id]
		for i := range t {
			if t[i] == "" {
				t[i] = "NOT_REACHED"
			}
		}
		files["RECOVERABILITY_CURVES.tsv"] = append(files["RECOVERABILITY_CURVES.tsv"], fmt.Sprintf("%s\t%s\t%s\t%.5f\t%d\t%.8f\t%.8f\t%.8f\t%s\t%s\t%s\tempirical 100-replicate mean; no extrapolation", k.candidate, k.corpus, k.channel, k.rate, a.n, a.sumGlyph/float64(a.n), a.sumToken/float64(a.n), a.sumExact/float64(a.n), t[0], t[1], t[2]))
	}
}

func runSingleErrors(files map[string][]string, data []*experimentData) {
	channels := []string{"GLYPH_SUBSTITUTION", "GLYPH_DELETION", "GLYPH_INSERTION", "TOKEN_BOUNDARY_INSERTION", "TOKEN_BOUNDARY_DELETION", "TOKEN_MERGE", "TOKEN_SPLIT"}
	for _, d := range data {
		cleanDecoded := d.model.decode(d.encoded.cipher)
		positions := []int{0, len(d.encoded.cipher) / 4, len(d.encoded.cipher) / 2, max(0, len(d.encoded.cipher)-2)}
		for _, ch := range channels {
			for pi, requested := range positions {
				seed := int64(67100000 + pi*1009 + len(ch)*97 + len(d.cand.Name)*13 + len(d.corpus))
				damaged, pos, loc := corruptOne(d.encoded.cipher, ch, requested, seed)
				decoded := d.model.decode(damaged)
				p := propagationMetrics(cleanDecoded, decoded, pos)
				files["ERROR_PROPAGATION.tsv"] = append(files["ERROR_PROPAGATION.tsv"], fmt.Sprintf("%s\t%s\t%s\t%d\t%.6f\t%s\t%d\t%d\t%d\t%.8f\t%d\t%s\t%d\t%d\tCLEAN_DECODED_SEQUENCE", d.cand.Name, d.corpus, ch, pos, float64(pos)/float64(max(1, len(d.encoded.cipher))), loc, seed, p.damaged, p.correctAfter, p.recoveryAfter, p.lSync, boolText(p.catastrophic), len(d.encoded.cipher), len(damaged)))
				valid := true
				for i := max(0, pos-1); i < min(len(damaged), pos+2); i++ {
					if _, ok := d.model.best[tokenKey(damaged[i])]; !ok && d.cand.Name != "M0_IDENTITY" && d.cand.Name != "M1_MONOALPHABETIC" && d.cand.Name != "M2_HOMOPHONY_H2" {
						valid = false
					}
				}
				matches := pos < len(decoded) && pos < len(d.encoded.plain) && tokenKey(decoded[pos]) == tokenKey(d.encoded.plain[pos])
				files["ERROR_DETECTABILITY.tsv"] = append(files["ERROR_DETECTABILITY.tsv"], fmt.Sprintf("%s\t%s\t%s\t%d\t%s\t%s\t%s\t%s\tfrozen valid-form dictionary; no oracle error location", d.cand.Name, d.corpus, ch, pos, boolText(valid), boolText(!valid), boolText(valid && !matches), boolText(matches)))
				files["RESYNCHRONIZATION.tsv"] = append(files["RESYNCHRONIZATION.tsv"], fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%s\t%.8f\tthree consecutive positionally correct decoded units", d.cand.Name, d.corpus, ch, pos, p.lSync, boolText(p.catastrophic), p.recoveryAfter))
			}
		}
	}
}

func runTranscription(files map[string][]string, data []*experimentData) {
	for _, d := range data {
		for _, fraction := range []float64{.05, .10, .25} {
			for rep := 0; rep < 30; rep++ {
				seed := int64(67200000 + rep*1009 + int(fraction*1000) + len(d.cand.Name)*17 + len(d.corpus))
				conflated, pairs := applyConflation(d.encoded.cipher, fraction, seed)
				changed := changedGlyphs(d.encoded.cipher, conflated)
				r := measureRecovery(d.encoded.plain, d.model.decode(conflated))
				files["TRANSCRIPTION_CONFLATION.tsv"] = append(files["TRANSCRIPTION_CONFLATION.tsv"], fmt.Sprintf("%s\t%s\t%.3f\t%d\t%d\t%d\t%d\t%.8f\t%.8f\t%.8f\t%.0f\tfrozen practical decoder", d.cand.Name, d.corpus, fraction, rep, seed, pairs, changed, r.glyph, r.token, r.wer, r.exact))
				split, classes := applySplitting(d.encoded.cipher, fraction, seed+1)
				changed = changedGlyphs(d.encoded.cipher, split)
				practical := measureRecovery(d.encoded.plain, d.model.decode(split))
				oracle := measureRecovery(d.encoded.plain, d.model.decode(removeSplittingMarks(split)))
				files["TRANSCRIPTION_SPLITTING.tsv"] = append(files["TRANSCRIPTION_SPLITTING.tsv"], fmt.Sprintf("%s\t%s\t%.3f\t%d\t%d\t%d\t%d\t%.8f\t%.8f\t%.8f\t%.8f\tfrozen decoder plus explicitly marked collapse oracle", d.cand.Name, d.corpus, fraction, rep, seed+1, classes, changed, practical.glyph, practical.token, oracle.glyph, practical.wer))
			}
		}
	}
}

func runSegmentation(files map[string][]string, data []*experimentData) {
	for _, d := range data {
		base := d.encoded.cipher
		pos := len(base) / 2
		conditions := []struct {
			name    string
			tokens  [][]string
			ops     int
			decoder string
		}{
			{"CORRECT_BOUNDARIES", cloneTokens(base), 0, "frozen"},
		}
		removed := [][]string{flatten(base)}
		conditions = append(conditions, struct {
			name    string
			tokens  [][]string
			ops     int
			decoder string
		}{"BOUNDARIES_REMOVED", removed, max(0, len(base)-1), "frozen without reconstruction"})
		merged, _, _ := corruptOne(base, "TOKEN_BOUNDARY_DELETION", pos, 67300001)
		conditions = append(conditions, struct {
			name    string
			tokens  [][]string
			ops     int
			decoder string
		}{"DELETE_ONE_BOUNDARY", merged, 1, "frozen"})
		split, _, _ := corruptOne(base, "TOKEN_BOUNDARY_INSERTION", pos, 67300002)
		conditions = append(conditions, struct {
			name    string
			tokens  [][]string
			ops     int
			decoder string
		}{"INSERT_ONE_BOUNDARY", split, 1, "frozen"})
		shifted := shiftBoundary(base, pos)
		conditions = append(conditions, struct {
			name    string
			tokens  [][]string
			ops     int
			decoder string
		}{"SHIFT_BOUNDARY_PLUS_ONE", shifted, 1, "frozen"})
		reconstructed := reconstructBoundaries(flatten(base), d.model)
		conditions = append(conditions, struct {
			name    string
			tokens  [][]string
			ops     int
			decoder string
		}{"BOUNDARIES_RECONSTRUCTED", reconstructed, max(0, len(base)-1), "ciphertext-only dynamic programming"})
		for i, c := range conditions {
			p, rc, f1 := boundaryScores(base, c.tokens)
			r := measureRecovery(d.encoded.plain, d.model.decode(c.tokens))
			files["SEGMENTATION_DAMAGE.tsv"] = append(files["SEGMENTATION_DAMAGE.tsv"], fmt.Sprintf("%s\t%s\t%s\t%d\t%d\t%d\t%d\t%.8f\t%.8f\t%.8f\t%.8f\t%.8f\t%.8f\t%s", d.cand.Name, d.corpus, c.name, i, 67300000+i, c.ops, len(c.tokens), p, rc, f1, r.glyph, r.token, r.wer, c.decoder))
		}
	}
}

func runResets(files map[string][]string, data []*experimentData) {
	for _, d := range data {
		if d.cand.Name != "M10_STATEFUL_FORM_K2" && d.cand.Name != "M11_MIXED_FORM_K2" {
			continue
		}
		pos := len(d.encoded.cipher) / 2
		damaged, pos, _ := corruptOne(d.encoded.cipher, "TOKEN_BOUNDARY_DELETION", pos, 67400001)
		cleanDecoded := d.model.decode(d.encoded.cipher)
		lineSize := medianLineSize(d.test)
		conditions := []struct {
			name     string
			interval int
		}{{"NO_RESET", 0}, {"RESET_EVERY_TOKEN", 1}, {"RESET_LINE_SIZED", lineSize}, {"RESET_PAGE_SIZED", lineSize * 20}, {"RESET_FIXED_N", 32}}
		for _, c := range conditions {
			decoded := resetDecode(d.model, d.encoded.cipher, damaged, pos, c.interval)
			p := propagationMetrics(cleanDecoded, decoded, pos)
			r := measureRecovery(d.encoded.plain, decoded)
			files["RESET_EXPERIMENT.tsv"] = append(files["RESET_EXPERIMENT.tsv"], fmt.Sprintf("%s\t%s\t%s\t%d\tTOKEN_BOUNDARY_DELETION\t%d\t67400001\t%d\t%.8f\t%d\t%s\t%.8f\t%.8f\texplicit synthetic reset markers survive the same injected error", d.cand.Name, d.corpus, c.name, c.interval, pos, p.damaged, p.recoveryAfter, p.lSync, boolText(p.catastrophic), r.glyph, r.token))
		}
	}
}

func runSecondaryControls(files map[string][]string, data []*experimentData) {
	models := map[string]map[string]*decoder{}
	for _, d := range data {
		if models[d.cand.Name] == nil {
			models[d.cand.Name] = map[string]*decoder{}
		}
		models[d.cand.Name][d.corpus] = d.model
	}
	wrong := map[string]string{"Doyle": "Astafiev", "Longfellow": "Doyle", "Astafiev": "Longfellow"}
	for _, d := range data {
		local := measureRecovery(d.encoded.plain, d.model.decode(d.encoded.cipher))
		wrongModel := models[d.cand.Name][wrong[d.corpus]]
		wr := measureRecovery(d.encoded.plain, wrongModel.decode(d.encoded.cipher))
		files["PLAINTEXT_LANGUAGE_PRIOR.tsv"] = append(files["PLAINTEXT_LANGUAGE_PRIOR.tsv"], fmt.Sprintf("%s\t%s\tTRAIN_VALIDATION_LOCAL\t%s\t%.8f\t%.8f\t%.8f\t%.0f\tTRAIN+VALIDATION only", d.cand.Name, d.corpus, d.corpus, local.glyph, local.token, local.wer, local.exact), fmt.Sprintf("%s\t%s\tWRONG_LANGUAGE_CONTROL\t%s\t%.8f\t%.8f\t%.8f\t%.0f\tother corpus TRAIN+VALIDATION only", d.cand.Name, d.corpus, wrong[d.corpus], wr.glyph, wr.token, wr.wer, wr.exact))

		shuffled := d.test.ShufflePlaintextWords(67500001)
		gen := encodeAligned(d.cand, shuffled)
		gr := measureRecovery(d.encoded.plain, d.model.decode(gen.cipher))
		files["GENERATOR_CONTROL.tsv"] = append(files["GENERATOR_CONTROL.tsv"], fmt.Sprintf("%s\t%s\t67500001\tSHUFFLED_PLAINTEXT\t%d\t%.8f\t%.8f\t%.8f\t%.0f\tsame frozen encoder grammar", d.cand.Name, d.corpus, len(gen.plain), gr.glyph, gr.token, gr.wer, gr.exact))

		cascade, copying := corruptRate(d.encoded.cipher, "GLYPH_SUBSTITUTION", .01, 67600001)
		cascade, pairs := applyConflation(cascade, .10, 67600002)
		cascade, _, _ = corruptOne(cascade, "TOKEN_BOUNDARY_DELETION", len(cascade)/2, 67600003)
		cr := measureRecovery(d.encoded.plain, d.model.decode(cascade))
		files["CASCADE_DAMAGE.tsv"] = append(files["CASCADE_DAMAGE.tsv"], fmt.Sprintf("%s\t%s\t0\t67600001\t%d\t%d\t1\t%.8f\t%.8f\t%.8f\t%.0f\tcorruption->transcription->segmentation", d.cand.Name, d.corpus, copying, pairs, cr.glyph, cr.token, cr.wer, cr.exact))

		oracle := trainDecoder(d.cand, d.test, d.test)
		or := measureRecovery(d.encoded.plain, oracle.decode(d.encoded.cipher))
		files["ORACLE_DECOMPOSITION.tsv"] = append(files["ORACLE_DECOMPOSITION.tsv"], fmt.Sprintf("%s\t%s\tPRACTICAL\tfalse\tfalse\tfalse\t%.8f\t%.8f\t%.8f\t%.0f\tfrozen TRAIN+VALIDATION decoder", d.cand.Name, d.corpus, local.glyph, local.token, local.wer, local.exact), fmt.Sprintf("%s\t%s\tORACLE_TEST_CODEBOOK\ttrue\ttrue\tfalse\t%.8f\t%.8f\t%.8f\t%.0f\texplicit upper bound; TEST plaintext used only here", d.cand.Name, d.corpus, or.glyph, or.token, or.wer, or.exact))
	}
}

func writeFrontier(files map[string][]string, data []*experimentData) error {
	auth, err := task66Compatibility("experiments/mechanism-space-v1/FAMILY_METRICS.tsv")
	if err != nil {
		return err
	}
	heldout, err := task66Compatibility("experiments/mechanism-space-v1/HELDOUT_RESULTS.tsv")
	if err != nil {
		return err
	}
	// HELDOUT is authoritative where Task66 evaluated a frozen frontier
	// candidate; DEVELOPMENT supplies the frozen pre-heldout axis for controls
	// and representatives that were not advanced to HELDOUT.
	for mechanism, corpora := range heldout {
		if auth[mechanism] == nil {
			auth[mechanism] = map[string]map[string]float64{}
		}
		for corpus, families := range corpora {
			auth[mechanism][corpus] = families
		}
	}
	type point struct {
		fp, g, t, ri float64
		n            int
	}
	means := map[string]*point{}
	for _, d := range data {
		name := candidateTask66Name(d.cand.Name)
		fam := auth[name][d.corpus]
		keys := make([]string, 0, len(fam))
		for k := range fam {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		above := 0
		vals := make([]float64, 0, len(keys))
		vector := make([]string, 0, len(keys))
		for _, k := range keys {
			v := fam[k]
			vals = append(vals, v)
			if v > .15 {
				above++
			}
			vector = append(vector, fmt.Sprintf("%s=%.6g", k, v))
		}
		fp := 0.0
		if len(keys) > 0 {
			fp = float64(above) / float64(len(keys))
		}
		med := medianFloat(vals)
		p := means[d.cand.Name]
		if p == nil {
			p = &point{}
			means[d.cand.Name] = p
		}
		p.fp += fp
		p.g += d.clean.glyph
		p.t += d.clean.token
		p.ri += d.ri
		p.n++
		files["FINGERPRINT_INFORMATION_FRONTIER.tsv"] = append(files["FINGERPRINT_INFORMATION_FRONTIER.tsv"], fmt.Sprintf("%s\t%s\t%s\tHELDOUT_IF_AVAILABLE_ELSE_DEVELOPMENT\t%d\t%d\t%.8f\t%.8f\t%s\t%.8f\t%.8f\t%.8f\tPENDING_AGGREGATE", d.cand.Name, d.corpus, name, len(keys), above, fp, med, strings.Join(vector, ";"), d.clean.glyph, d.clean.token, d.ri))
	}
	front := map[string]bool{}
	dominatedBy := map[string]string{}
	for a, pa := range means {
		front[a] = true
		for b, pb := range means {
			if a == b {
				continue
			}
			af, ag := pa.fp/float64(pa.n), pa.g/float64(pa.n)
			bf, bg := pb.fp/float64(pb.n), pb.g/float64(pb.n)
			if bf >= af && bg >= ag && (bf > af || bg > ag) {
				front[a] = false
				dominatedBy[a] = b
				break
			}
		}
	}
	for i, row := range files["FINGERPRINT_INFORMATION_FRONTIER.tsv"][1:] {
		p := strings.Split(row, "\t")
		p[len(p)-1] = boolText(front[p[0]])
		files["FINGERPRINT_INFORMATION_FRONTIER.tsv"][i+1] = strings.Join(p, "\t")
	}
	names := make([]string, 0, len(means))
	for n := range means {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		p := means[n]
		den := float64(p.n)
		dominator := dominatedBy[n]
		if dominator == "" {
			dominator = "NONE"
		}
		files["RECOVERABILITY_PARETO.tsv"] = append(files["RECOVERABILITY_PARETO.tsv"], fmt.Sprintf("%s\t%.8f\t%.8f\t%.8f\t%.8f\t%s\t%s", n, p.fp/den, p.g/den, p.t/den, p.ri/den, boolText(front[n]), dominator))
	}
	return nil
}

func writeClassification(files map[string][]string, data []*experimentData, curves map[curveKey]*curveAgg) {
	by := map[string][]*experimentData{}
	for _, d := range data {
		by[d.cand.Name] = append(by[d.cand.Name], d)
	}
	for _, cand := range candidates() {
		ds := by[cand.Name]
		if len(ds) == 0 {
			continue
		}
		clean, ri := 0.0, 0.0
		for _, d := range ds {
			clean += d.clean.glyph
			ri += d.ri
		}
		clean /= float64(len(ds))
		ri /= float64(len(ds))
		frag := "ROBUST_IN_TESTED_RANGE"
		for k, a := range curves {
			if k.candidate == cand.Name && k.channel == "GLYPH_SUBSTITUTION" && k.rate == .005 && a.sumGlyph/float64(a.n) < .9*clean {
				frag = "FRAGILE_AT_OR_BELOW_0.5_PERCENT"
			}
		}
		sync := "LOCAL_RESYNCHRONIZATION"
		if strings.HasPrefix(cand.Name, "M10") || strings.HasPrefix(cand.Name, "M11") {
			sync = "RESET_TEST_REPORTED"
		}
		trans := "MEASURED_LOW"
		if clean > 0 && ri < .999 {
			trans = "REPRESENTATION_INDUCED_INFORMATION_LOSS_TESTED"
		}
		seg := "MEASURED_IN_SEGMENTATION_DAMAGE"
		final := cand.Class
		if clean > .999 {
			final = "MATHEMATICALLY_REVERSIBLE"
		} else if ri < .25 {
			final = "INTRINSICALLY_LOSSY"
		}
		files["FINAL_CLASSIFICATION.tsv"] = append(files["FINAL_CLASSIFICATION.tsv"], fmt.Sprintf("%s\t%.8f\t%.8f\t%s\t%s\t%s\t%s\t%s\t%s\tAUTHORITATIVE_TASK66_AXIS\t%s", cand.Name, clean, ri, cand.Class, frag, sync, trans, seg, "MEASURED", final))
	}
}

func validateExperimentalArtifacts(dir string) error {
	required := map[string][]string{
		"ERROR_PROPAGATION.tsv":                {"error_type", "error_position", "damaged_units", "recovery_after_error", "L_sync", "catastrophic_desync"},
		"RESET_EXPERIMENT.tsv":                 {"reset_condition", "error_position", "damaged_units", "L_sync"},
		"TRANSCRIPTION_CONFLATION.tsv":         {"replicate", "seed", "conflated_class_pairs", "changed_glyphs", "glyph_recovery"},
		"TRANSCRIPTION_SPLITTING.tsv":          {"replicate", "seed", "split_classes", "changed_glyphs", "glyph_recovery_frozen"},
		"SEGMENTATION_DAMAGE.tsv":              {"condition", "boundary_precision", "boundary_recall", "glyph_recovery"},
		"FINGERPRINT_INFORMATION_FRONTIER.tsv": {"fingerprint_compatibility", "task67_clean_glyph_recovery", "task66_family_progress_vector"},
	}
	for name, cols := range required {
		rows, err := readTSV(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if len(rows) < 2 {
			return fmt.Errorf("%s has no experimental rows", name)
		}
		for _, c := range cols {
			if _, ok := rows[0][c]; !ok {
				return fmt.Errorf("%s missing %s", name, c)
			}
		}
	}
	front, _ := readTSV(filepath.Join(dir, "FINGERPRINT_INFORMATION_FRONTIER.tsv"))
	different := false
	for _, r := range front[1:] {
		families, _ := strconv.Atoi(r["task66_family_count"])
		if families < 7 {
			return fmt.Errorf("frontier row %s/%s has only %d Task66 families", r["candidate"], r["corpus"], families)
		}
		a, _ := strconv.ParseFloat(r["fingerprint_compatibility"], 64)
		b, _ := strconv.ParseFloat(r["task67_clean_glyph_recovery"], 64)
		if math.Abs(a-b) > 1e-9 {
			different = true
			break
		}
	}
	if !different {
		return fmt.Errorf("frontier axes are reducible to the same scalar")
	}
	errorRows, _ := readTSV(filepath.Join(dir, "ERROR_RECOVERABILITY.tsv"))
	if len(errorRows) != 102901 {
		return fmt.Errorf("ERROR_RECOVERABILITY.tsv has %d rows, want 102901 including header", len(errorRows))
	}
	resetRows, _ := readTSV(filepath.Join(dir, "RESET_EXPERIMENT.tsv"))
	resetConditions := map[string]bool{}
	for _, r := range resetRows[1:] {
		resetConditions[r["reset_condition"]] = true
	}
	for _, condition := range []string{"NO_RESET", "RESET_EVERY_TOKEN", "RESET_LINE_SIZED", "RESET_PAGE_SIZED", "RESET_FIXED_N"} {
		if !resetConditions[condition] {
			return fmt.Errorf("RESET_EXPERIMENT.tsv lacks %s", condition)
		}
	}
	clean, _ := readTSV(filepath.Join(dir, "CLEAN_RECOVERABILITY.tsv"))
	base := map[string]string{}
	for _, r := range clean[1:] {
		if r["partition"] == "TEST" {
			base[r["candidate"]] = r["clean_glyph_recovery"]
		}
	}
	for _, name := range []string{"ERROR_PROPAGATION.tsv", "RESET_EXPERIMENT.tsv", "TRANSCRIPTION_CONFLATION.tsv", "TRANSCRIPTION_SPLITTING.tsv", "SEGMENTATION_DAMAGE.tsv"} {
		rows, _ := readTSV(filepath.Join(dir, name))
		allCopied := true
		for _, r := range rows[1:] {
			value := ""
			for _, k := range []string{"recovery_after_error", "glyph_recovery", "glyph_recovery_frozen"} {
				if r[k] != "" {
					value = r[k]
					break
				}
			}
			if value != "" && value != base[r["candidate"]] {
				allCopied = false
				break
			}
		}
		if allCopied {
			return fmt.Errorf("%s is reducible to CLEAN_RECOVERABILITY candidate scalar", name)
		}
	}
	return nil
}

func readTSV(path string) ([]map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	if !s.Scan() {
		return nil, fmt.Errorf("empty %s", path)
	}
	h := strings.Split(s.Text(), "\t")
	rows := []map[string]string{{}}
	for _, x := range h {
		rows[0][x] = x
	}
	for s.Scan() {
		p := strings.Split(s.Text(), "\t")
		r := map[string]string{}
		for i, k := range h {
			if i < len(p) {
				r[k] = p[i]
			}
		}
		rows = append(rows, r)
	}
	return rows, s.Err()
}
func changedGlyphs(a, b [][]string) int {
	x, y := flatten(a), flatten(b)
	n := 0
	for i := 0; i < min(len(x), len(y)); i++ {
		if x[i] != y[i] {
			n++
		}
	}
	n += max(len(x), len(y)) - min(len(x), len(y))
	return n
}
func shiftBoundary(in [][]string, pos int) [][]string {
	out := cloneTokens(in)
	if len(out) < 2 {
		return out
	}
	pos = min(max(0, pos), len(out)-2)
	if len(out[pos+1]) == 0 {
		return out
	}
	out[pos] = append(out[pos], out[pos+1][0])
	out[pos+1] = cloneToken(out[pos+1][1:])
	return out
}
func medianLineSize(c mechanismspace.Corpus) int {
	counts := map[int]int{}
	for _, l := range c.Lines {
		counts[l]++
	}
	v := make([]int, 0, len(counts))
	for _, n := range counts {
		if n > 0 {
			v = append(v, n)
		}
	}
	sort.Ints(v)
	if len(v) == 0 {
		return 12
	}
	return max(1, v[len(v)/2])
}
func medianFloat(v []float64) float64 {
	if len(v) == 0 {
		return 0
	}
	x := append([]float64(nil), v...)
	sort.Float64s(x)
	if len(x)%2 == 1 {
		return x[len(x)/2]
	}
	return (x[len(x)/2-1] + x[len(x)/2]) / 2
}
