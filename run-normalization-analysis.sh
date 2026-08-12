#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${TMPDIR:-/tmp}/voinich-go-cache"

mkdir -p bin
go build -o bin/sequence-analyze ./sequence-analyze
go build -o bin/structural-normalize ./structural-normalize
go build -o bin/normalization-compare ./normalization-compare

bin/structural-normalize \
  -input data_work/ivtt_output_1786282555007.txt \
  -structural dataset/structural_analysis.yaml \
  -output normalized.txt \
  -classes structural_classes.yaml \
  -thresholds 0.70,0.75,0.80,0.85,0.90 \
  -singleton-mode preserve \
  -random-baselines 100 \
  -random-seed 1

bin/normalization-compare \
  -classes structural_classes.yaml \
  -input data_work/ivtt_output_1786282555007.txt \
  -raw-analysis sequence_analysis.yaml \
  -normalized-pattern 'normalized_%s.txt' \
  -analysis-pattern 'sequence_analysis_%s.yaml' \
  -sequence-analyzer bin/sequence-analyze \
  -random-baselines 100 \
  -random-seed 1 \
  -output normalization_comparison.yaml
