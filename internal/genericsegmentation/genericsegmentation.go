// Package genericsegmentation is task43's single shared generic
// segmentation layer. Stages 23-27 (token-relation-validate,
// replicated-local-structure-audit, higher-order-sequence-validate,
// positional-continuation-validate, transition-network-validate) each
// build a physical "block" partition from IVTFF-derived Currier x hand
// metadata purely as an opaque grouping device for their block-based
// permutation nulls and leave-one-block-out tests - the identity of the
// metadata label is never read by their actual statistics (see
// GENERIC_STAGE_APPLICABILITY_AUDIT.md). This package supplies the same
// two-level partition (fine physical blocks, coarse resampling groups) for
// a plain-text generic corpus, without fabricating hand/Currier/folio
// metadata: the coarse label is explicitly a statistical resampling fold,
// never presented as manuscript provenance.
//
// The algorithm is deterministic, language-agnostic, and scales with
// corpus size rather than any Voynich-specific constant:
//
//  1. L = number of distinct natural corpus lines. L<2 is too small to
//     form even one cross-block comparison -> ErrNotEnoughData.
//  2. targetBlocks = clamp(round(sqrt(L)), minTargetBlocks, maxTargetBlocks)
//     - enough fine blocks for a permutation/leave-one-block-out null to
//     be meaningful, without degenerating into one-line slivers for large
//     corpora.
//  3. Lines are grouped into contiguous fine blocks of
//     ceil(L/targetBlocks) lines each; an undersized trailing remainder
//     (less than half a full block) is merged into the previous block
//     instead of forming its own sliver. A block never splits a natural
//     line.
//  4. fineBlockCount<2 -> ErrNotEnoughData (nothing to compare across).
//  5. K = min(coarseGroups, fineBlockCount) coarse resampling groups.
//     coarseGroups is a fixed small resampling-fold design constant (not
//     derived from the corpus) chosen to match ordinary k-fold practice.
//     Fine blocks are assigned round-robin (block index mod K), so
//     adjacent fine blocks always land in different groups - mirroring how
//     a real Currier x hand "Joint" value already behaves for contiguous
//     manuscript runs, and giving every existing per-package "maximal
//     contiguous run of one Joint value" block-builder exactly one fine
//     block per physical block, unchanged.
package genericsegmentation

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math"
	"os"
	"strings"
)

// Sentinel is the constant value callers should substitute for the
// historical second metadata dimension ("hand") that a generic corpus has
// no analogue for. It is always this one fixed string - never a fabricated
// per-token identity - so that any per-package "is this metadata known?"
// gate (which typically rejects an empty value) still accepts it, while a
// reader can immediately recognize it as a generic-mode marker rather than
// real manuscript metadata.
const Sentinel = "generic"

// ErrNotEnoughData is returned by Build when the corpus has too few
// natural lines to form even one cross-block comparison. Callers must
// surface this as an explicit error, never panic or silently produce a
// degenerate single-block result.
var ErrNotEnoughData = errors.New("genericsegmentation: corpus has too few natural lines for a generic block/group partition")

const (
	minTargetBlocks = 8
	maxTargetBlocks = 64
	coarseGroups    = 4
)

// TokenInfo is the generic, corpus-derived substitute for one token's
// IVTFF-sourced line/hand/Currier metadata columns.
type TokenInfo struct {
	// LineID is a stable, opaque per-natural-line identifier.
	LineID string
	// IndexInLine is the token's zero-based position within its own
	// natural corpus line.
	IndexInLine int
	// Group is the coarse deterministic resampling-fold label ("G0".."Gk-1").
	// It is never a fabricated hand/Currier/folio value.
	Group string
}

// ReadCorpus tokenizes a plain-text corpus exactly like every existing
// pipeline stage already does (scan lines, split each line on whitespace),
// additionally recording each token's natural line number - information
// today's per-package loadCorpus helpers read but discard.
func ReadCorpus(path string) (tokens []string, lineOfToken []int, sha256Hex string, err error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, "", err
	}
	h := sha256.Sum256(b)
	s := bufio.NewScanner(strings.NewReader(string(b)))
	s.Buffer(make([]byte, 64*1024), 1024*1024)
	line := 0
	for s.Scan() {
		for f := range strings.FieldsSeq(s.Text()) {
			tokens = append(tokens, f)
			lineOfToken = append(lineOfToken, line)
		}
		line++
	}
	if err := s.Err(); err != nil {
		return nil, nil, "", err
	}
	return tokens, lineOfToken, hex.EncodeToString(h[:]), nil
}

// Build computes deterministic per-token structural info from each
// token's natural line assignment (see package doc for the algorithm).
// lineOfToken must be non-decreasing, as produced by ReadCorpus.
func Build(lineOfToken []int) ([]TokenInfo, error) {
	if len(lineOfToken) == 0 {
		return nil, ErrNotEnoughData
	}
	lineCount := lineOfToken[len(lineOfToken)-1] + 1
	if lineCount < 2 {
		return nil, ErrNotEnoughData
	}

	targetBlocks := clampInt(int(math.Round(math.Sqrt(float64(lineCount)))), minTargetBlocks, maxTargetBlocks)
	linesPerBlock := max(1, (lineCount+targetBlocks-1)/targetBlocks)

	// blockOfLine[l] = fine block index owning natural line l.
	blockOfLine := make([]int, lineCount)
	blockCount := 0
	for start := 0; start < lineCount; {
		end := min(start+linesPerBlock, lineCount)
		// Merge an undersized trailing remainder into the previous block
		// instead of creating a sliver smaller than half a full block.
		if blockCount > 0 && end == lineCount && end-start < linesPerBlock/2+1 {
			for l := start; l < end; l++ {
				blockOfLine[l] = blockCount - 1
			}
			start = end
			continue
		}
		for l := start; l < end; l++ {
			blockOfLine[l] = blockCount
		}
		blockCount++
		start = end
	}
	if blockCount < 2 {
		return nil, ErrNotEnoughData
	}

	k := min(coarseGroups, blockCount)

	indexInLine := make([]int, lineCount)
	out := make([]TokenInfo, len(lineOfToken))
	for i, l := range lineOfToken {
		b := blockOfLine[l]
		out[i] = TokenInfo{
			LineID:      fmt.Sprintf("L%d", l),
			IndexInLine: indexInLine[l],
			Group:       fmt.Sprintf("G%d", b%k),
		}
		indexInLine[l]++
	}
	return out, nil
}

func clampInt(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
