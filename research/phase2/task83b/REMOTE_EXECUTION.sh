#!/usr/bin/env bash
set -euo pipefail

root=${1:-/home/brigadire/task83b-remote-work}
bin_dir="$root/.task83b-bin"
runs_dir="$root/.task83b-runs"
cache_dir="$root/.task83b-gocache"

mkdir -p "$bin_dir" "$runs_dir" "$cache_dir"
cd "$root"
GOCACHE="$cache_dir" go build -o "$bin_dir/fingerprint-v2-analyze" ./cmd/fingerprint-v2-analyze
GOCACHE="$cache_dir" go build -o "$bin_dir/ivtff-x7-extract" ./cmd/ivtff-x7-extract
GOCACHE="$cache_dir" go build -o "$bin_dir/codex_prepare" ./cmd/codex_prepare
GOCACHE="$cache_dir" go build -o "$bin_dir/task79c-pf4-hr" ./cmd/task79c-pf4-hr
GOCACHE="$cache_dir" go build -o "$bin_dir/task79c-distance-pareto" ./cmd/task79c-distance-pareto

for run in A B C; do
	d="$runs_dir/run${run}"
	mkdir -p "$d/configs" "$d/data_work" "$d/out/aggregates"
	ln -sfn "$root/data" "$d/data"
	ln -sfn "$root/data_test" "$d/data_test"
	ln -sfn "$root/research" "$d/research"
	ln -sfn "$root/experiments" "$d/experiments"
	printf 'gitdir: %s/.git\n' "$root" > "$d/.git"
	sed 's#^output_dir:.*#output_dir: out/zl#' "$root/experiments/fingerprint-v2-task79-v1/canonical.yaml" > "$d/configs/zl.yaml"
	sed 's#^output_dir:.*#output_dir: out/it#' "$root/experiments/fingerprint-v2-task79c-v1/transcription-it.yaml" > "$d/configs/it.yaml"
	sed 's#^output_dir:.*#output_dir: out/controls#' "$root/experiments/fingerprint-v2-task79c-v1/controls-portfolio.yaml" > "$d/configs/controls.yaml"
	(cd "$d" && "$bin_dir/ivtff-x7-extract" -input data/ZL3b-n.txt -output data_work/ZL3b-x7.txt)
	(cd "$d" && "$bin_dir/codex_prepare" prepare -input data_work/ZL3b-x7.txt -output data_work/ZL3b-x7.canonical.txt -encoding utf-8 -case preserve -line-policy preserve)
	(cd "$d" && "$bin_dir/ivtff-x7-extract" -input data/IT2a-n.txt -output data_work/IT2a-x7.txt)
	(cd "$d" && "$bin_dir/codex_prepare" prepare -input data_work/IT2a-x7.txt -output data_work/IT2a-x7.canonical.txt -encoding utf-8 -case preserve -line-policy preserve)
done

run_remaining() {
	d=$1
	workers=$2
	if [[ "$workers" == default ]]; then
		(cd "$d" && "$bin_dir/fingerprint-v2-analyze" -config configs/it.yaml && "$bin_dir/fingerprint-v2-analyze" -config configs/controls.yaml) > "$d/remaining.log" 2>&1
	else
		(cd "$d" && GOMAXPROCS="$workers" "$bin_dir/fingerprint-v2-analyze" -config configs/it.yaml && GOMAXPROCS="$workers" "$bin_dir/fingerprint-v2-analyze" -config configs/controls.yaml) > "$d/remaining.log" 2>&1
	fi
	(cd "$d" && "$bin_dir/task79c-pf4-hr" -line-profiles out/zl/line_profiles.json -output out/aggregates/pf4_hierarchy_result.json)
	jq -s '.[0] + .[1]' "$d/out/zl/discriminative_validation.json" "$d/out/controls/discriminative_validation.json" > "$d/out/aggregates/combined_discriminative_validation.json"
	(cd "$d" && "$bin_dir/task79c-distance-pareto" -discriminative out/aggregates/combined_discriminative_validation.json -registry out/zl/metric_registry.json -output out/aggregates/full_portfolio_distance_pareto.json)
}

run_remaining "$runs_dir/runA" 1 &
pid_a=$!
run_remaining "$runs_dir/runB" 2 &
pid_b=$!
run_remaining "$runs_dir/runC" default &
pid_c=$!
wait "$pid_a"
wait "$pid_b"
wait "$pid_c"

printf 'TASK83B_REMOTE_RUNS_COMPLETE\n' > "$root/REMOTE_COMPLETE"
