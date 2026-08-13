#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${TMPDIR:-/tmp}/voinich-go-cache"

ivtff="data/ZL3b-n.txt"
input="data_work/ZL3b-x7.txt"
workdir="workdir"
dataset="$workdir/dataset"

mkdir -p "$dataset"

# The IVTFF source is authoritative. All analyzers consume this deterministic
# IVTT -x7 derivative and never the former timestamp-named export.
./ivtt/ivtt -x7 "$ivtff" "$input"

go run . "$input" "$dataset/dictionary.yaml"
go run ./dict-analyze "$dataset/dictionary.yaml" "$dataset/tokens_analysis.yaml"
go run ./structural-analyze \
  -dictionary "$dataset/dictionary.yaml" \
  -analysis "$dataset/tokens_analysis.yaml" \
  -output "$dataset/structural_analysis.yaml"
go run ./sequence-analyze \
  -input "$input" \
  -output "$workdir/sequence_analysis.yaml"

./run-normalization-analysis.sh

go run ./structural-validate \
  -input "$input" \
  -classes "$workdir/structural_classes.yaml" \
  -folds 5 \
  -fold-seed 1 \
  -threshold 0.70 \
  -random-baselines 100 \
  -random-seed 1 \
  -output "$workdir/structural_validation.yaml"

go run ./structural-profile-stability \
  -input "$input" \
  -classes "$workdir/structural_classes.yaml" \
  -folds 5 \
  -fold-seed 1 \
  -min-token-count 10 \
  -neighbors 10 \
  -bootstrap-runs 200 \
  -bootstrap-seed 1 \
  -threshold 0.70 \
  -threshold-margin 0.05 \
  -output "$workdir/structural_profile_stability.yaml"

go run ./structural-reliability \
  -input "$input" \
  -classes "$workdir/structural_classes.yaml" \
  -folds 5 \
  -fold-seed 1 \
  -min-token-count 10 \
  -neighbors 10 \
  -bootstrap-runs 200 \
  -bootstrap-seed 1 \
  -threshold 0.70 \
  -threshold-margin 0.05 \
  -count-thresholds 10,20,40,80,160,320 \
  -subsample-min-full-count 160 \
  -subsample-runs 100 \
  -subsample-seed 1 \
  -output "$workdir/structural_reliability.yaml"

go run ./soft-structural-space
go run ./structural-graphemic
go run ./structural-pair-decompose
