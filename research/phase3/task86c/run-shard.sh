#!/bin/sh
set -eu

if [ "$#" -ne 5 ]; then
  echo "usage: run-shard.sh BINARY SHARD_COUNT SHARD_INDEX PARALLEL OUTPUT_ROOT" >&2
  exit 2
fi

binary=$1
shard_count=$2
shard_index=$3
parallel=$4
output_root=$5
manifest=research/phase3/task86c/TASK86C_JOB_MANIFEST.tsv

mkdir -p "$output_root"
awk -F '\t' -v count="$shard_count" -v shard="$shard_index" 'NR > 1 && ((NR-2) % count) == shard {print $1}' "$manifest" |
  xargs -r -n 1 -P "$parallel" sh -c '
    binary=$1
    root=$2
    id=$3
    out=$root/$id/result.json
    if [ -f "$out" ]; then exit 0; fi
    mkdir -p "$root/$id"
    "$binary" task86c-worker -job-id "$id" -out "$out"
  ' sh "$binary" "$output_root"
