package metadatavalidation

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"strings"
)

func ReadFrozenCorpus(path string) ([]string, string, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	h := fmt.Sprintf("%x", sha256.Sum256(b))
	return strings.Fields(string(b)), h, nil
}

// Align performs a strict sequential comparison. It never skips, splits,
// joins, substitutes, or approximately matches a frozen token.
func Align(d Document, frozen []string, hash string) (AlignmentResult, error) {
	r := AlignmentResult{Tokens: frozen, TotalLoci: len(d.Loci), SkippedLoci: d.SkippedLoci, CorpusSHA256: hash}
	pos, folioIndex := 0, 0
	lastFolio := ""
	for _, l := range d.Loci {
		produced := strings.Fields(l.AlignmentText)
		if len(produced) == 0 {
			continue
		}
		if l.Folio != lastFolio {
			folioIndex = 0
			lastFolio = l.Folio
		}
		for i, token := range produced {
			if pos >= len(frozen) || frozen[pos] != token {
				end := min(len(frozen), pos+max(10, len(produced)))
				return r, &AlignmentError{Position: pos, Locus: l,
					Expected: append([]string(nil), frozen[pos:end]...), Produced: produced,
					Before: append([]string(nil), frozen[max(0, pos-10):pos]...),
					After:  append([]string(nil), frozen[pos:min(len(frozen), pos+10)]...), Reason: "token identity mismatch"}
			}
			r.Records = append(r.Records, TokenMetadata{Position: pos, Token: token,
				Folio: l.Folio, LocusID: l.ID, LocusType: l.Type, LineID: l.LineID,
				ParagraphID: l.ParagraphID, ParagraphStart: l.ParagraphStart && i == 0,
				Currier: l.Variables["C"], Hand: l.Variables["H"], Quire: l.Variables["Q"], Section: l.Variables["I"],
				IndexInLocus: i, IndexInLine: i, IndexInFolio: folioIndex})
			pos++
			folioIndex++
		}
	}
	if pos != len(frozen) {
		var l Locus
		if len(d.Loci) > 0 {
			l = d.Loci[len(d.Loci)-1]
		}
		return r, &AlignmentError{Position: pos, Locus: l, Expected: append([]string(nil), frozen[pos:min(len(frozen), pos+10)]...), Reason: "token count invariant failed", Before: append([]string(nil), frozen[max(0, pos-10):pos]...)}
	}
	return r, nil
}

func FormatAlignmentError(e *AlignmentError) string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("ALIGNMENT ERROR\n\nposition: %d\nfolio: %s\nlocus: %s\n\nIVTFF:\n%s\n\nnormalized:\n%s\n\nexpected frozen:\n%s\n\nproduced/matched tokens:\n%s\n\ncontext before:\n%s\n\ncontext after:\n%s\n\nreason: %s", e.Position, e.Locus.Folio, e.Locus.ID, e.Locus.RawText, e.Locus.AlignmentText, strings.Join(e.Expected, " "), strings.Join(e.Produced, " "), strings.Join(e.Before, " "), strings.Join(e.After, " "), e.Reason)
}

func WriteAlignmentFailure(w io.Writer, e error) {
	bw := bufio.NewWriter(w)
	defer bw.Flush()
	fmt.Fprint(bw, "# IVTFF alignment report\n\nResult: **FAIL**\n\n")
	fmt.Fprintln(bw, e)
}
