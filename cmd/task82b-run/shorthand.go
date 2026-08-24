package main

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"unicode"

	"zcore.dev/voinich/internal/task82b"
)

var (
	bddOnce  sync.Once
	bddPairs []task82b.PairUnit
	bddStats task82b.CharDeletionStats
	bddErr   error
)

func loadBDD(root string) ([]task82b.PairUnit, task82b.CharDeletionStats, error) {
	bddOnce.Do(func() {
		paths, err := filepath.Glob(filepath.Join(root, "data_test/bdd-tei/koeln-edd-c-119/*.xml"))
		if err != nil {
			bddErr = err
			return
		}
		if len(paths) == 0 {
			bddErr = fmt.Errorf("no BDD TEI files found under data_test/bdd-tei/koeln-edd-c-119")
			return
		}
		res, err := task82b.ExtractTEIPairs(paths)
		if err != nil {
			bddErr = err
			return
		}
		bddPairs = res.Pairs
		bddStats = task82b.BuildCharDeletionStats(res.Pairs)
	})
	return bddPairs, bddStats, bddErr
}

// shorthandScales returns the scale-id -> file-basename-filter map
// (task82b.txt sec.17: local blocks / larger blocks / full corpus). Each
// BDD chapter file is one "local block"; "combined" is the full corpus.
func shorthandScaleNames(pairs []task82b.PairUnit) []string {
	set := map[string]bool{}
	for _, p := range pairs {
		set[filepath.Base(p.File)] = true
	}
	var names []string
	for n := range set {
		names = append(names, n)
	}
	sort.Strings(names)
	names = append(names, "combined")
	return names
}

var shorthandVariants = []string{
	"EXPANDED",
	"ABBREVIATED",
	"NULL_RANDOM_DELETION_MATCHED",
	"NULL_FREQUENCY_MATCHED_DELETION",
	"NULL_POSITION_MATCHED",
}

func buildShorthandJobs(root string, baseSeed int64) []Job {
	pairs, _, err := loadBDD(root)
	if err != nil {
		// Recorded as a report-time limitation (TASK82B_REPORT.md), not a
		// panic: Task82b must still be able to freeze the extraction
		// branch even if the shorthand branch's data is unavailable.
		return nil
	}
	var jobs []Job
	i := 0
	for _, scale := range shorthandScaleNames(pairs) {
		for _, variant := range shorthandVariants {
			if !strings.HasPrefix(variant, "NULL_") {
				jobs = append(jobs, Job{
					JobID:            fmt.Sprintf("shorthand__%s__%s", scale, variant),
					Kind:             "shorthand_variant",
					CorpusID:         "BDD_koeln-edd-c-119",
					ShorthandVariant: variant,
					ShorthandScale:   scale,
					Seed:             baseSeed + 5000 + int64(i),
				})
				i++
				continue
			}
			// Null variants are randomized (NullWord draws from a *rand.Rand);
			// shorthandNullSeeds independent replicates let
			// SHORTHAND_NULL_COMPARISON.tsv build a real null distribution
			// for CompareToNull, matching Branch A's 3-seed random null.
			for s := 1; s <= shorthandNullSeeds; s++ {
				jobs = append(jobs, Job{
					JobID:            fmt.Sprintf("shorthand__%s__%s__seed%d", scale, variant, s),
					Kind:             "shorthand_variant",
					CorpusID:         "BDD_koeln-edd-c-119",
					ShorthandVariant: variant,
					ShorthandScale:   scale,
					Seed:             baseSeed + 5000 + int64(i),
				})
				i++
			}
		}
	}
	return jobs
}

const shorthandNullSeeds = 3

// hasNaturalGlyph reports whether s has at least one rune fingerprintv2's
// GlyphMode=natural (internal/evaglyph.NaturalGlyphs) would keep, so a
// degenerate all-punctuation/no-content word is skipped here rather than
// making the frozen F2 extractor itself error on it.
func hasNaturalGlyph(s string) bool {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsNumber(r) {
			return true
		}
	}
	return false
}

func scalePairs(all []task82b.PairUnit, scale string) []task82b.PairUnit {
	if scale == "combined" {
		return all
	}
	var out []task82b.PairUnit
	for _, p := range all {
		if filepath.Base(p.File) == scale {
			out = append(out, p)
		}
	}
	return out
}

func (j Job) runShorthand(root, workDir string) (JobRecord, error) {
	all, stats, err := loadBDD(root)
	if err != nil {
		return JobRecord{}, err
	}
	pairs := scalePairs(all, j.ShorthandScale)

	var groups [][]string
	r := rand.New(rand.NewSource(j.Seed))
	skipped := 0
	for _, p := range pairs {
		expan := strings.ToLower(strings.TrimSpace(p.ExpanText))
		if expan == "" {
			skipped++
			continue
		}
		var word string
		switch j.ShorthandVariant {
		case "EXPANDED":
			word = expan
		case "ABBREVIATED":
			// p.AbbrText can contain an embedded newline (a mid-word
			// <lb/> inside <abbr>, TASK82B_DESIGN.md sec.5); the
			// PAIR_DEFINED one-word-per-line convention needs exactly
			// one physical line per pair, so internal whitespace is
			// stripped rather than preserved as a second line.
			word = strings.Join(strings.Fields(strings.ToLower(p.AbbrText)), "")
		case "NULL_RANDOM_DELETION_MATCHED":
			word = task82b.NullWord("RANDOM_DELETION_MATCHED", expan, task82b.DeletionCount(p), stats, r)
		case "NULL_FREQUENCY_MATCHED_DELETION":
			word = task82b.NullWord("FREQUENCY_MATCHED_DELETION", expan, task82b.DeletionCount(p), stats, r)
		case "NULL_POSITION_MATCHED":
			word = task82b.NullWord("POSITION_MATCHED", expan, task82b.DeletionCount(p), stats, r)
		default:
			return JobRecord{}, fmt.Errorf("unknown shorthand variant %q", j.ShorthandVariant)
		}
		if word == "" || !hasNaturalGlyph(word) {
			skipped++
			continue
		}
		groups = append(groups, []string{word})
	}

	rec := JobRecord{
		JobID: j.JobID, Kind: j.Kind, CorpusID: j.CorpusID,
		ShorthandVariant: j.ShorthandVariant, ShorthandScale: j.ShorthandScale, Seed: j.Seed,
		SkippedGroups: skipped, CandidatePool: len(pairs), ChosenCount: len(groups),
	}
	typeSet := map[string]bool{}
	for _, g := range groups {
		typeSet[g[0]] = true
	}
	if len(groups) < 20 || len(typeSet) < 2 {
		rec.Degenerate = true
		rec.DegenerateNote = fmt.Sprintf("words=%d types=%d", len(groups), len(typeSet))
	}

	f2, ok := safeExtractF2(writeTemp(workDir, groups), rec.JobID, j.CorpusID, 1, workDir+"_f2out", groups)
	if !ok {
		rec.Degenerate = true
		if rec.DegenerateNote == "" {
			rec.DegenerateNote = "F2 extractor error"
		}
	}
	rec.F2 = f2
	ax := task82b.ComputeAX(groups)
	rec.AX = &ax
	return rec, nil
}
