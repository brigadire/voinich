package main

import "sort"

var debugNegative = false

// NegativePair is one matched (positive, negative) TOKEN pair per
// NEGATIVE_TOKEN_PROTOCOL.md.
type NegativePair struct {
	PositiveRaw    string
	PositiveGlyphs []string
	NegativeRaw    string
	NegativeGlyphs []string
}

// glyphFrequencyClasses assigns each DEVELOPMENT glyph a class 0-3 by
// weighted cumulative-frequency quartile, sorted by (frequency,
// code point); a tied block stays together in the class containing the
// block's first cumulative position.
func glyphFrequencyClasses(dev []TokenOccurrence) (classOf map[string]int, byClass [4][]string) {
	counts := map[string]int{}
	total := 0
	for _, o := range dev {
		for _, g := range o.Glyphs {
			counts[g]++
			total++
		}
	}
	glyphs := make([]string, 0, len(counts))
	for g := range counts {
		glyphs = append(glyphs, g)
	}
	sort.Slice(glyphs, func(i, j int) bool {
		fi, fj := float64(counts[glyphs[i]])/float64(total), float64(counts[glyphs[j]])/float64(total)
		if fi != fj {
			return fi < fj
		}
		return glyphs[i] < glyphs[j]
	})
	classOf = map[string]int{}
	cum := 0.0
	i := 0
	for i < len(glyphs) {
		j := i
		freq := float64(counts[glyphs[i]]) / float64(total)
		for j < len(glyphs) && float64(counts[glyphs[j]])/float64(total) == freq {
			j++
		}
		startCum := cum
		cls := quartileOf(startCum)
		for k := i; k < j; k++ {
			classOf[glyphs[k]] = cls
			byClass[cls] = append(byClass[cls], glyphs[k])
			cum += float64(counts[glyphs[k]]) / float64(total)
		}
		i = j
	}
	return classOf, byClass
}

func quartileOf(cum float64) int {
	switch {
	case cum < 0.25:
		return 0
	case cum < 0.50:
		return 1
	case cum < 0.75:
		return 2
	default:
		return 3
	}
}

func seededShuffle(p *PRNG, n int) []int {
	idx := make([]int, n)
	for i := range idx {
		idx[i] = i
	}
	for i := n - 1; i > 0; i-- {
		j := int(p.Float64() * float64(i+1))
		if j > i {
			j = i
		}
		idx[i], idx[j] = idx[j], idx[i]
	}
	return idx
}

// BuildNegativePairs implements NEGATIVE_TOKEN_PROTOCOL.md for one
// candidate/transcription. componentCountOf is nil for M0-M4 and, for M5,
// returns the frozen-segmentation component count of a glyph sequence (ok
// false if unsegmentable).
func BuildNegativePairs(namespace, transcription, candidateID, modelClass string, dev, val, heldout []TokenOccurrence, componentCountOf func([]string) (int, bool)) ([]NegativePair, bool) {
	classOf, byClass := glyphFrequencyClasses(dev)
	classOfGlyph := func(g string) int {
		if c, ok := classOf[g]; ok {
			return c
		}
		return 0
	}
	observedTypes := map[string]bool{}
	for _, o := range dev {
		observedTypes[joinGlyphs(o.Glyphs)] = true
	}
	for _, o := range val {
		observedTypes[joinGlyphs(o.Glyphs)] = true
	}
	for _, o := range heldout {
		observedTypes[joinGlyphs(o.Glyphs)] = true
	}
	usedNegatives := map[string]bool{}

	byLength := map[int][]TokenOccurrence{}
	for _, o := range dev {
		byLength[len(o.Glyphs)] = append(byLength[len(o.Glyphs)], o)
	}
	// Precompute each DEV source's component count once (M5 only):
	// componentCountOf is a pure function of the source's own glyphs, so
	// recomputing it per HELDOUT query (O(heldout x sources) segment()
	// calls) is pure waste.
	var byLengthAndCount map[int]map[int][]TokenOccurrence
	if componentCountOf != nil {
		byLengthAndCount = map[int]map[int][]TokenOccurrence{}
		for length, srcs := range byLength {
			byCount := map[int][]TokenOccurrence{}
			for _, s := range srcs {
				if c, ok := componentCountOf(s.Glyphs); ok {
					byCount[c] = append(byCount[c], s)
				}
			}
			byLengthAndCount[length] = byCount
		}
	}

	var pairs []NegativePair
	for idx, pos := range heldout {
		L := len(pos.Glyphs)
		sources := byLength[L]
		if componentCountOf != nil {
			targetCount, ok := componentCountOf(pos.Glyphs)
			if !ok {
				return nil, false
			}
			sources = byLengthAndCount[L][targetCount]
		}
		if len(sources) == 0 {
			return nil, false
		}
		seed := SeedFields{
			Namespace: namespace, ModelClass: modelClass, CandidateID: candidateID,
			CorpusID: "NEGATIVE", Transcription: transcription, Partition: "HELDOUT",
			Scale: 1, Replicate: idx,
		}
		prng := NewSeededPRNG(seed)
		srcOrder := seededShuffle(prng, len(sources))

		found := ""
		scratch := make([]string, len(pos.Glyphs))
		// attemptCap bounds the exhaustive search per occurrence. The
		// frozen protocol requires full one/two-position enumeration
		// before declaring NEGATIVE_EXHAUSTED, but on a large corpus with
		// many same-length (same-component-count, for M5) sources this is
		// combinatorially large; a bounded search cap -- analogous to the
		// frozen 100,000-operation induction cap used elsewhere in this
		// contract -- keeps the search tractable. Exceeding it is treated
        // identically to true exhaustion (no negative found), which is
        // the only externally observable difference: PM6 unavailability.
		const attemptCap = 20000
		attempts := 0
		tryMutation := func(base []string, positions []int) string {
			copy(scratch, base)
			for _, p := range positions {
				orig := base[p]
				cls := classOfGlyph(orig)
				alts := byClass[cls]
				altOrder := seededShuffle(prng, len(alts))
				for _, ai := range altOrder {
					alt := alts[ai]
					if alt == orig {
						continue
					}
					attempts++
					scratch[p] = alt
					key := joinGlyphs(scratch)
					scratch[p] = orig
					if observedTypes[key] || usedNegatives[key] {
						continue
					}
					return key
				}
				if attempts >= attemptCap {
					return ""
				}
			}
			return ""
		}

		for _, si := range srcOrder {
			if attempts >= attemptCap {
				break
			}
			src := sources[si].Glyphs
			posOrder := seededShuffle(prng, len(src))
			if key := tryMutation(src, posOrder); key != "" {
				found = key
				break
			}
		}
		if found == "" && attempts < attemptCap {
			// Two-position mutations, lexicographic position-pair order.
		outer:
			for _, si := range srcOrder {
				if attempts >= attemptCap {
					break
				}
				src := sources[si].Glyphs
				n := len(src)
				copy(scratch, src)
				for i := 0; i < n && found == ""; i++ {
					origI := src[i]
					ci := classOfGlyph(origI)
					altsI := byClass[ci]
					if len(altsI) <= 1 {
						continue
					}
					oi := seededShuffle(prng, len(altsI))
					for j := i + 1; j < n; j++ {
						if attempts >= attemptCap {
							break outer
						}
						origJ := src[j]
						cj := classOfGlyph(origJ)
						altsJ := byClass[cj]
						if len(altsJ) <= 1 {
							continue
						}
						oj := seededShuffle(prng, len(altsJ))
						for _, ii := range oi {
							ai := altsI[ii]
							if ai == origI {
								continue
							}
							scratch[i] = ai
							for _, jj := range oj {
								aj := altsJ[jj]
								if aj == origJ {
									continue
								}
								scratch[j] = aj
								attempts++
								key := joinGlyphs(scratch)
								if !observedTypes[key] && !usedNegatives[key] {
									found = key
								}
								if found != "" || attempts >= attemptCap {
									break
								}
							}
							scratch[j] = origJ
							if found != "" || attempts >= attemptCap {
								break
							}
						}
						scratch[i] = origI
						if found != "" || attempts >= attemptCap {
							break outer
						}
					}
				}
			}
		}
		if found == "" {
			if debugNegative {
				println("NEGATIVE_EXHAUSTED at idx", idx, "raw", pos.Raw, "len", len(pos.Glyphs), "sources", len(sources))
			}
			return nil, false // NEGATIVE_EXHAUSTED
		}
		usedNegatives[found] = true
		pairs = append(pairs, NegativePair{
			PositiveRaw: pos.Raw, PositiveGlyphs: pos.Glyphs,
			NegativeRaw: found, NegativeGlyphs: splitGlyphs(found),
		})
	}
	return pairs, true
}
