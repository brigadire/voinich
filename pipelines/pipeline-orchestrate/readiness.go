package main

import (
	"fmt"
	"os"
	"strings"

	"zcore.dev/voinich/internal/corpusprep"
)

const readinessStageName = "corpus-readiness-check"

func runCorpusReadinessCheck(m *Manifest) (string, error) {
	data, err := os.ReadFile(m.CorpusPath)
	if err != nil {
		return "", fmt.Errorf("read corpus: %w", err)
	}
	report, err := corpusprep.Check(data)
	if err != nil {
		return "", err
	}
	var b strings.Builder
	fmt.Fprintf(&b, "Corpus readiness check\n")
	fmt.Fprintf(&b, "Input: %s\n", m.CorpusPath)
	fmt.Fprintf(&b, "Canonical corpus version: %d\n", corpusprep.CanonicalCorpusVersion)
	fmt.Fprintf(&b, "Valid: %v\n", report.Valid)
	fmt.Fprintf(&b, "Reason: %s\n", report.Reason)
	fmt.Fprintf(&b, "Tokens: %d\n", report.Stats.OutputTokenCount)
	fmt.Fprintf(&b, "Lines: %d\n", report.Stats.LineCount)
	if !report.Valid {
		return b.String(), fmt.Errorf("generic corpus is not canonical: %s", report.Reason)
	}
	prepareMeta, prepareSHA, err := loadPrepareManifestMetadata(m.CorpusPath, m.CorpusSHA256)
	if err != nil {
		return b.String(), err
	}
	if prepareSHA != "" {
		fmt.Fprintf(&b, "Prepare manifest: %s\n", m.CorpusPath+".prepare.json")
		fmt.Fprintf(&b, "Prepared by: %s\n", prepareMeta.PreparedBy)
		fmt.Fprintf(&b, "Prepare manifest sha256: %s\n", prepareSHA)
		fmt.Fprintf(&b, "Canonical corpus version: %d\n", prepareMeta.CanonicalCorpusVersion)
		if m.PreparedBy != "" && m.PreparedBy != prepareMeta.PreparedBy {
			return b.String(), fmt.Errorf("prepare manifest provenance changed: manifest=%q current=%q", m.PreparedBy, prepareMeta.PreparedBy)
		}
		if m.PrepareManifestSHA256 != "" && m.PrepareManifestSHA256 != prepareSHA {
			return b.String(), fmt.Errorf("prepare manifest sha256 changed: manifest=%q current=%q", m.PrepareManifestSHA256, prepareSHA)
		}
		if m.CanonicalCorpusVersion != 0 && m.CanonicalCorpusVersion != prepareMeta.CanonicalCorpusVersion {
			return b.String(), fmt.Errorf("canonical corpus version changed: manifest=%d current=%d", m.CanonicalCorpusVersion, prepareMeta.CanonicalCorpusVersion)
		}
	}
	return b.String(), nil
}
