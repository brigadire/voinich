#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${TMPDIR:-/tmp}/voinich-go-cache"

workdir="workdir"
bindir="$workdir/bin"
mkdir -p "$bindir"
go build -o "$bindir/sequence-analyze" ./cmd/sequence-analyze
go build -o "$bindir/structural-normalize" ./cmd/structural-normalize
go build -o "$bindir/normalization-compare" ./cmd/normalization-compare

"$bindir/structural-normalize" \
  -input data_work/ZL3b-x7.txt \
  -structural "$workdir/dataset/structural_analysis.yaml" \
  -output "$workdir/normalized.txt" \
  -classes "$workdir/structural_classes.yaml" \
  -thresholds 0.70,0.75,0.80,0.85,0.90 \
  -singleton-mode preserve \
  -random-baselines 100 \
  -random-seed 1

"$bindir/normalization-compare" \
  -classes "$workdir/structural_classes.yaml" \
  -input data_work/ZL3b-x7.txt \
  -raw-analysis "$workdir/sequence_analysis.yaml" \
  -normalized-pattern "$workdir/normalized_%s.txt" \
  -analysis-pattern "$workdir/sequence_analysis_%s.yaml" \
  -sequence-analyzer "$bindir/sequence-analyze" \
  -random-baselines 100 \
  -random-seed 1 \
  -output "$workdir/normalization_comparison.yaml"
