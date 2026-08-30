package structurecatalog

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func (cat *Catalog) writeSummaryAndDocs() error {
	c := cat.Primary
	cat.Summary["schema_version"] = SchemaVersion
	cat.Summary["corpus_path"] = c.Path
	cat.Summary["corpus_sha256"] = c.SHA
	cat.Summary["transcription"] = c.Transcription
	cat.Summary["token_count"] = si(len(c.Occurrences))
	cat.Summary["unique_token_count"] = si(len(c.Counts))
	cat.Summary["physical_line_count"] = si(len(c.Lines))
	cat.Summary["folio_count"] = si(len(c.Folios))
	cat.Summary["section_count"] = si(len(c.Sections))
	cat.Summary["locus_type_count"] = si(len(c.LocusTypes))
	cat.Summary["locus_metadata_available"] = fmt.Sprint(c.MetadataAvailable)
	cat.Summary["master_rule_count"] = si(len(cat.Rules))
	keys := make([]string, 0, len(cat.Summary))
	for k := range cat.Summary {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	rows := [][]string{}
	for _, k := range keys {
		rows = append(rows, []string{k, cat.Summary[k]})
	}
	if err := writeTSV(filepath.Join(cat.Config.OutputDir, "STRUCTURAL_SUMMARY.tsv"), []string{"metric", "value"}, rows); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(cat.Config.OutputDir, "STRUCTURAL_CATALOG_SPEC.md"), []byte(cat.spec()), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(cat.Config.OutputDir, "VM_STRUCTURAL_CATALOG.md"), []byte(cat.report()), 0644)
}

func (cat *Catalog) spec() string {
	return fmt.Sprintf(`# VM structural catalog specification

Schema: %s. Primary input is literal Unicode transcription symbols from the canonical corpus; a transcription symbol is not necessarily a physical atomic glyph. No visual decomposition is attempted. Multi-character transcription conventions are consequently represented as sequences of literal symbols.

Physical lines are hard boundaries for transitions and co-occurrence. The ZL3b IVTFF source is strictly aligned token-for-token to the canonical corpus to attach folio, locus type and section metadata. IT2a is used only for independent stability checks.

Every tested relation separates **observed_status** (the corpus fact) from **inferred_status** (the marginal-preserving analytic null assessment). Every zero remains UNOBSERVED with a concrete corpus_rule; FDR never removes it. P-values use directional Poisson tails with frozen lower-level marginals: within-token adjacency slots for glyphs, predecessor/successor marginals and physical-line boundaries for transitions, fixed token frequencies plus the exact physical-line slot lengths for co-occurrence, and document-category marginals for metadata. BH-FDR is applied per output family. These analytic approximations are explicitly not universal-mechanism claims.

The frequent-token threshold is fixed before inspection at frequency >= %d. The full token-transition complement is stored as a deterministic gzip JSON adjacency-complement representation; the explicit unobserved TSV covers the fixed frequent subset. Edit families are connected components of literal Levenshtein-distance-one insertions, deletions and substitutions; transformations are descriptive relations, not morphemes.

All TSV files are UTF-8, tab-delimited, have one header row, use NA for inapplicable numeric values, and are deterministically ordered.

Generate with go run ./cmd/vm-structure generate. Query without recomputation with go run ./cmd/vm-structure query glyph q, query follows daiin, query absent-with daiin, query position daiin, or query section daiin. Use --catalog DIR before the query type for a non-default frozen catalog.
`, SchemaVersion, cat.Config.MinFrequency)
}

func (cat *Catalog) report() string {
	c := cat.Primary
	topPreferred := topRules(cat.Rules, func(r Rule) bool {
		return r.RuleType == "TOKEN_TRANSITION" && r.OpportunityCount >= cat.Config.MinFrequency && strings.Contains(r.InferredStatus, "PREFERRED")
	}, 8, true)
	topAvoided := topRules(cat.Rules, func(r Rule) bool {
		return r.RuleType == "TOKEN_TRANSITION" && (strings.Contains(r.InferredStatus, "AVOIDED") || r.InferredStatus == "DEPLETED")
	}, 8, false)
	gInitial, gFinal := glyphAbsences(c)
	neverFirst, neverLast := frequentPositionAbsences(c, cat.Config.MinFrequency)
	prefixes, suffixes := dominantAffixes(c, 8)
	coPreferred := topRules(cat.Rules, func(r Rule) bool {
		return r.RuleType == "TOKEN_LINE_COOCCURRENCE" && strings.Contains(r.InferredStatus, "PREFERRED")
	}, 8, true)
	docRules := topRules(cat.Rules, func(r Rule) bool {
		return (strings.Contains(r.RuleType, "_BY_locus") || strings.Contains(r.RuleType, "_BY_section")) && strings.Contains(r.InferredStatus, "PREFERRED")
	}, 8, true)
	return fmt.Sprintf(`# Voynich Manuscript structural property catalog

## Scope and provenance

This is a descriptive/empirical catalog of what is and is not observed in the ZL3b canonical corpus. “Not observed” never means “impossible in Voynichese.” The primary corpus is %s (SHA-256 %s): %d tokens, %d unique token types, %d physical lines, %d folios, %d sections, and %d literal transcription symbols. Metadata availability: %t.

## Glyph rules

The literal-symbol inventory has %s symbols. Of %s possible directed bigrams, %s are observed and %s are not observed in this corpus (constraint density %s). Position-exclusive facts and every zero-valued pair are listed explicitly in GLYPH_POSITION_RULES.tsv and GLYPH_BIGRAM_RULES.tsv; conditional continuations through n=4 are in GLYPH_NGRAM_RULES.tsv.

Symbols not observed token-initially: %s. Symbols not observed token-finally: %s. These are literal-symbol facts for this corpus, not universal claims.

## Token formation

The catalog contains all %d token types with line position and document coverage. Literal edit-distance-one structure yields %s connected components and %s edges. Prefix/suffix patterns of length 1–4 are reported as patterns, not linguistic morphemes.

Highest-frequency prefix patterns: %s. Highest-frequency suffix patterns: %s. Entries are pattern:token-frequency-count; token-type productivity is retained in TOKEN_AFFIX_PATTERNS.tsv.

## Token sequencing

At frequency >= %d there are %s possible ordered transitions: %s observed and %s unobserved corpus rules (constraint density %s). The full 8,244-type adjacency complement is queryable without recomputation. Higher-order rows give both P(C|A,B) and P(C|B).

Strongest statistically preferred rows (support and effect are retained in the TSV/master catalog):

%s

Strongest statistically avoided/depleted rows:

%s

## Line grammar

There are %s observed frequent unordered same-line pairs and %s exclusions (constraint density %s). Line files preserve token/glyph lengths, endpoints, repetitions and token-family progression. Immediate transitions and same-line co-occurrence remain distinct relations.

Frequent tokens not observed line-initially (first 30): %s. Frequent tokens not observed line-finally (first 30): %s.

Strongest preferred same-line pairs:

%s

## Document grammar

The locus, folio and section files enumerate each supported token/family against every available category, including every absence. An absence is a corpus fact; enrichment/depletion is a separate FDR-controlled inference. Locus-exclusive families: %s. Section-exclusive families: %s.

Strongest locus/section specializations:

%s

## Transcription stability

TRANSCRIPTION_STABILITY.tsv compares key positional, bigram, transition and line-co-occurrence corpus facts with IT2a. Literal relations lacking comparable token/symbol identities are marked NOT_COMPARABLE; no identity is manufactured across incompatible readings.

## Reuse and limitations

The implementation reuses the repository's audited strict IVTFF alignment and canonical provenance, and integrates literal edit-family, sequence, transition, boundary and positional analyses into concrete rules. It makes no linguistic, cryptographic, mnemonic, shorthand, numeric, procedural or generative-mechanism classification.
`, c.Path, c.SHA, len(c.Occurrences), len(c.Counts), len(c.Lines), len(c.Folios), len(c.Sections), len(c.Inventory), c.MetadataAvailable, cat.Summary["number_of_glyphs"], cat.Summary["possible_glyph_bigrams"], cat.Summary["observed_glyph_bigrams"], cat.Summary["unobserved_glyph_bigrams"], cat.Summary["glyph_bigram_constraint_density"], gInitial, gFinal, len(c.Counts), cat.Summary["edit_families"], cat.Summary["edit_distance_1_edges"], prefixes, suffixes, cat.Config.MinFrequency, cat.Summary["possible_frequent_token_transitions"], cat.Summary["observed_frequent_token_transitions"], cat.Summary["unobserved_frequent_token_transitions"], cat.Summary["frequent_transition_constraint_density"], topPreferred, topAvoided, cat.Summary["observed_frequent_same_line_pairs"], cat.Summary["frequent_same_line_exclusions"], cat.Summary["same_line_constraint_density"], neverFirst, neverLast, coPreferred, cat.Summary["locus_exclusive_families"], cat.Summary["section_exclusive_families"], docRules)
}

func glyphAbsences(c Corpus) (string, string) {
	initial, final := map[rune]bool{}, map[rune]bool{}
	for _, o := range c.Occurrences {
		r := []rune(o.Token)
		initial[r[0]] = true
		final[r[len(r)-1]] = true
	}
	a, b := []string{}, []string{}
	for _, g := range c.Inventory {
		if !initial[g] {
			a = append(a, string(g))
		}
		if !final[g] {
			b = append(b, string(g))
		}
	}
	return printable(a), printable(b)
}
func frequentPositionAbsences(c Corpus, minFreq int) (string, string) {
	first, last := map[string]bool{}, map[string]bool{}
	for _, o := range c.Occurrences {
		if o.Index == 0 {
			first[o.Token] = true
		}
		if o.Index == len(c.Lines[o.Line])-1 {
			last[o.Token] = true
		}
	}
	a, b := []string{}, []string{}
	for _, t := range sortedTokens(c.Counts) {
		if c.Counts[t] >= minFreq && !first[t] {
			a = append(a, t)
		}
		if c.Counts[t] >= minFreq && !last[t] {
			b = append(b, t)
		}
	}
	if len(a) > 30 {
		a = a[:30]
	}
	if len(b) > 30 {
		b = b[:30]
	}
	return printable(a), printable(b)
}
func printable(x []string) string {
	if len(x) == 0 {
		return "none"
	}
	for i := range x {
		x[i] = "`" + x[i] + "`"
	}
	return strings.Join(x, ", ")
}
func dominantAffixes(c Corpus, n int) (string, string) {
	type item struct {
		s string
		n int
	}
	p, s := map[string]int{}, map[string]int{}
	for t, f := range c.Counts {
		r := []rune(t)
		for k := 1; k <= 4 && k <= len(r); k++ {
			p[string(r[:k])] += f
			s[string(r[len(r)-k:])] += f
		}
	}
	top := func(m map[string]int) string {
		a := []item{}
		for x, v := range m {
			a = append(a, item{x, v})
		}
		sort.Slice(a, func(i, j int) bool {
			if a[i].n != a[j].n {
				return a[i].n > a[j].n
			}
			return a[i].s < a[j].s
		})
		if len(a) > n {
			a = a[:n]
		}
		z := []string{}
		for _, x := range a {
			z = append(z, fmt.Sprintf("`%s`:%d", x.s, x.n))
		}
		return strings.Join(z, ", ")
	}
	return top(p), top(s)
}

func topRules(rs []Rule, ok func(Rule) bool, n int, descending bool) string {
	a := []Rule{}
	for _, r := range rs {
		if ok(r) {
			a = append(a, r)
		}
	}
	sort.Slice(a, func(i, j int) bool {
		if descending {
			return a[i].EffectSize > a[j].EffectSize
		}
		return a[i].EffectSize < a[j].EffectSize
	})
	if len(a) > n {
		a = a[:n]
	}
	lines := []string{}
	for _, r := range a {
		lines = append(lines, fmt.Sprintf("- `%s → %s` (%s): count %d, expected %s, effect %s, q=%s, %s / %s.", r.LHS, r.RHS, r.RuleType, r.ObservedCount, sf(r.ExpectedCount), sf(r.EffectSize), sf(r.QValue), r.CorpusRule, r.InferredStatus))
	}
	if len(lines) == 0 {
		return "- None at the frozen FDR threshold."
	}
	return strings.Join(lines, "\n")
}
