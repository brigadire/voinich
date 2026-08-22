package inversehomophony

// BuildFeatures computes the ciphertext-only feature set (task57 section 7)
// for every distinct token in relabeled tokens. lineOfToken must be the
// same-length, non-decreasing per-token natural line index produced by
// genericsegmentation.ReadCorpus / LoadCorpus.
func BuildFeatures(tokens []string, lineOfToken []int, cfg Config) map[string]*TokenFeatures {
	out := make(map[string]*TokenFeatures)
	get := func(t string) *TokenFeatures {
		f, ok := out[t]
		if !ok {
			f = &TokenFeatures{
				Pred:    make(map[string]int),
				Succ:    make(map[string]int),
				DistCtx: make(map[string]int),
				PosHist: make([]int, cfg.PositionalBuckets),
			}
			out[t] = f
		}
		return f
	}

	lineStart, lineLen := lineBounds(lineOfToken)

	for i, t := range tokens {
		f := get(t)
		f.Freq++

		line := lineOfToken[i]
		start, length := lineStart[line], lineLen[line]
		idxInLine := i - start
		bucket := positionalBucket(idxInLine, length, cfg.PositionalBuckets)
		f.PosHist[bucket]++

		if i > 0 && lineOfToken[i-1] == line {
			f.Pred[tokens[i-1]]++
		}
		if i+1 < len(tokens) && lineOfToken[i+1] == line {
			f.Succ[tokens[i+1]]++
		}
		for _, d := range cfg.Distances {
			if i-d >= 0 && lineOfToken[i-d] == line {
				f.DistCtx[tokens[i-d]]++
			}
			if i+d < len(tokens) && lineOfToken[i+d] == line {
				f.DistCtx[tokens[i+d]]++
			}
		}
	}
	return out
}

func lineBounds(lineOfToken []int) (start, length []int) {
	if len(lineOfToken) == 0 {
		return nil, nil
	}
	n := lineOfToken[len(lineOfToken)-1] + 1
	start = make([]int, n)
	length = make([]int, n)
	for i, l := range lineOfToken {
		if length[l] == 0 {
			start[l] = i
		}
		length[l]++
	}
	return start, length
}

// positionalBucket maps a token's zero-based index within its line to one
// of buckets equal-width bins over [0,1) of index/length. A line of length
// 1 always lands in bucket 0.
func positionalBucket(idx, length, buckets int) int {
	if length <= 1 || buckets <= 1 {
		return 0
	}
	frac := float64(idx) / float64(length)
	b := int(frac * float64(buckets))
	if b >= buckets {
		b = buckets - 1
	}
	return b
}
