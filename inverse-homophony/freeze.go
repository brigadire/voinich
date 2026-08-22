package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"zcore.dev/voinich/internal/inversehomophony"
)

// writeFreeze writes task57 section 22's immutable freeze record once the
// validation gate has passed: INVERSE_HOMOPHONY_METHOD.md (full method
// description + validation results) and METHOD_FROZEN (a short marker
// with the manifest hash, following the project's existing FROZEN-marker
// convention - see experiments/voynich-v1/FROZEN).
func writeFreeze(outDir string, r *inversehomophony.ValidationReport) error {
	manifestPath := filepath.Join(outDir, "manifest.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return err
	}
	sum := sha256.Sum256(manifestBytes)
	manifestSHA := hex.EncodeToString(sum[:])

	methodDoc := renderMethodDoc(r, manifestSHA)
	if err := os.WriteFile(filepath.Join(outDir, "INVERSE_HOMOPHONY_METHOD.md"), []byte(methodDoc), 0o644); err != nil {
		return err
	}

	marker := fmt.Sprintf("Inverse Homophony Recovery method frozen at %s\nMethodVersion: %s\nGitCommit: %s\nManifestSHA256: %s\nValidationGate: PASS\n",
		time.Now().UTC().Format(time.RFC3339), r.MethodVersion, "see manifest.json", manifestSHA)
	return os.WriteFile(filepath.Join(outDir, "METHOD_FROZEN"), []byte(marker), 0o644)
}

func renderMethodDoc(r *inversehomophony.ValidationReport, manifestSHA string) string {
	cfg := r.Config
	return fmt.Sprintf(`# INVERSE_HOMOPHONY_METHOD.md (task57 section 22)

Frozen after the synthetic validation gate passed. After this record is
written, no parameter, feature, or threshold below may change on the basis
of any Voynich result (task57 section 22).

## Identity

- Method version: %s
- Manifest SHA256: %s

## Features (frozen)

Predecessor distribution (distance 1), successor distribution (distance
1), distance-context profile (distances %v, both sides), positional
histogram (%d buckets of index-in-line/line-length). Graphemic similarity
is never computed (task57 section 8) - all input is opaque-relabeled
before any feature is built.

## Normalization

Every distribution L1-normalized before comparison.

## Similarity formula

Per component: 1 - JensenShannonDivergence(base 2). Combined score is the
unweighted mean of the four component similarities.

## Clustering / search algorithm

Similarity graph restricted to pairs sharing at least one predecessor or
successor context token (support >= %d observations on the lower side),
sorted by descending combined score, greedy union-find merge subject to:

- score > tau (%.6f)
- merged class occurrence-fraction <= %.4f (MaxClassFraction)
- resulting occurrence-weighted partition entropy >= %.4f x NO_COLLAPSE entropy (MinEntropyFraction)

## Thresholds

- tau = %.6f, fit on DEVELOPMENT corpora only (Doyle H4 uniform + Doyle H4
  triangular-v1/weighted) as the Youden-J-optimal point of the pooled
  true/false pair discrimination diagnostic (development AUC = %.4f,
  %d true pairs, %d false pairs).
- MinSupport = %d, MaxClassFraction = %.4f, MinEntropyFraction = %.4f -
  fixed before any corpus was scored (task57 section 16).

## Complexity penalty / anti-collapse

MaxClassFraction and MinEntropyFraction jointly forbid the all-tokens-to-
one-class degenerate solution (task57 section 15); enforced inline during
clustering, not as a post-hoc filter.

## Stopping rule

Process the fixed, pre-sorted candidate edge list exactly once; no
iterative re-scoring, no simulated annealing, no ML optimizer (task57
section 17).

## Random seeds

RANDOM_PARTITION baseline seeds: %v.

## Validation results

Validation gate: %s. See SYNTHETIC_VALIDATION_REPORT.md, class_recovery.tsv,
structural_recovery.tsv, baseline_comparison.tsv, null_distribution.tsv in
this directory for full results.
`, r.MethodVersion, manifestSHA, cfg.Distances, cfg.PositionalBuckets, cfg.MinSupport,
		cfg.Threshold, cfg.MaxClassFraction, cfg.MinEntropyFraction,
		cfg.Threshold, r.DevelopmentAUC, r.DevelopmentTruePairs, r.DevelopmentFalsePairs,
		cfg.MinSupport, cfg.MaxClassFraction, cfg.MinEntropyFraction,
		inversehomophony.RandomSeeds, verdictWord(r.Gate.Pass))
}
