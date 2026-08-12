#!/usr/bin/env bash
set -euo pipefail

export GOCACHE="${TMPDIR:-/tmp}/voinich-go-cache"

input="data_work/ivtt_output_1786282555007.txt"

go run . "$input" dataset/dictionary.yaml
go run ./dict-analyze dataset/dictionary.yaml dataset/tokens_analysis.yaml
go run ./structural-analyze \
  -dictionary dataset/dictionary.yaml \
  -analysis dataset/tokens_analysis.yaml \
  -output dataset/structural_analysis.yaml
go run ./sequence-analyze \
  -input "$input" \
  -output sequence_analysis.yaml

./run-normalization-analysis.sh

go run ./structural-validate \
  -input "$input" \
  -classes structural_classes.yaml \
  -folds 5 \
  -fold-seed 1 \
  -threshold 0.70 \
  -random-baselines 100 \
  -random-seed 1 \
  -output structural_validation.yaml

go run ./structural-profile-stability \
  -input "$input" \
  -classes structural_classes.yaml \
  -folds 5 \
  -fold-seed 1 \
  -min-token-count 10 \
  -neighbors 10 \
  -bootstrap-runs 200 \
  -bootstrap-seed 1 \
  -threshold 0.70 \
  -threshold-margin 0.05 \
  -output structural_profile_stability.yaml

go run ./structural-reliability \
  -input "$input" \
  -classes structural_classes.yaml \
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
  -output structural_reliability.yaml
