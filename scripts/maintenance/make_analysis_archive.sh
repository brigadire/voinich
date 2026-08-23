#!/usr/bin/env bash
set -euo pipefail

#EXP="experiments/doyle-sign-of-four-v1-2"
#CORPUS="data_test/pg2097-2.txt"
OUT="${EXP}-analysis.tar.gz"


tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

mkdir -p "$tmp/experiment/workdir"
mkdir -p "$tmp/experiment/logs"
mkdir -p "$tmp/input"

# Experiment metadata
cp "$EXP/manifest.json"    "$tmp/experiment/"
cp "$EXP/run-state.json"   "$tmp/experiment/"
cp "$EXP/artifacts.json"   "$tmp/experiment/" || true

# Stage logs
cp -a "$EXP/logs/." "$tmp/experiment/logs/"

# All machine-readable/scientific outputs and textual reports.
# Deliberately exclude binaries, plots and large normalized corpora.
find "$EXP/workspace/workdir" -type f \
    \( -name '*.yaml' \
       -o -name '*.yml' \
       -o -name '*.tsv' \
       -o -name '*.json' \
       -o -name '*.md' \
       -o -name '*.graphml' \) \
    ! -path '*/bin/*' \
    ! -path '*/plots/*' \
    -print0 |
while IFS= read -r -d '' f; do
    rel="${f#"$EXP/workspace/workdir/"}"
    mkdir -p "$tmp/experiment/workdir/$(dirname "$rel")"
    cp "$f" "$tmp/experiment/workdir/$rel"
done

# The exact canonical input used by the experiment
cp "$CORPUS" "$tmp/input/"

# Checksums of what we are sending
(
    cd "$tmp"
    find . -type f -print0 |
        sort -z |
        xargs -0 sha256sum > SHA256SUMS
)

tar -C "$tmp" -czf "$OUT" .

echo
echo "Created: $OUT"
du -h "$OUT"
echo
echo "Files:"
tar -tzf "$OUT" | wc -l
