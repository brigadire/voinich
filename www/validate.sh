#!/bin/sh
set -eu

site_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
fail=0

find "$site_dir" -name '*.html' -type f | while IFS= read -r page; do
  base=$(dirname "$page")
  perl -ne 'while (/(?:href|src)="([^"]+)"/g) { print "$1\n" }' "$page" | while IFS= read -r link; do
    case "$link" in
      http:*|https:*|mailto:*|\#*|'') continue ;;
    esac
    target=${link%%\#*}
    target=${target%%\?*}
    if [ -d "$base/$target" ]; then target="$target/index.html"; fi
    if [ ! -e "$base/$target" ]; then
      printf 'broken local reference: %s -> %s\n' "$page" "$link" >&2
      exit 1
    fi
  done
done || fail=1

(cd "$site_dir/artifacts" && sha256sum -c SHA256SUMS) || fail=1
(cd "$site_dir" && sha256sum -c SITE_FILES_SHA256SUMS) || fail=1

grep -q '"release_id": "publication-site-v1.0.0"' "$site_dir/artifacts/RELEASE_MANIFEST.json" || fail=1
grep -q '"BEST_SUPPORTED_CLASS": "INCONCLUSIVE"' "$site_dir/artifacts/RELEASE_MANIFEST.json" || fail=1
grep -q '"MECHANISM_IDENTIFICATION_FROM_F2": "NOT_IDENTIFIABLE"' "$site_dir/artifacts/RELEASE_MANIFEST.json" || fail=1
grep -q 'BEST_SUPPORTED_CLASS = INCONCLUSIVE' "$site_dir/index.html" || fail=1
grep -q 'MECHANISM_IDENTIFICATION_FROM_F2 = NOT_IDENTIFIABLE' "$site_dir/phase-2/index.html" || fail=1
grep -q 'Why Doyle appears here' "$site_dir/phase-1/index.html" || fail=1
grep -q 'Why external memory was considered' "$site_dir/transition/index.html" || fail=1
grep -q 'Why Fontana was used' "$site_dir/transition/index.html" || fail=1
grep -q 'Mechanism classes were tested. None was identified.' "$site_dir/phase-2/index.html" || fail=1
grep -q 'does not generalize to all possible external-memory systems' "$site_dir/phase-2/index.html" || fail=1

if find "$site_dir" -type f \( -name '*.key' -o -name '*.pem' -o -name '*.crt' \) | grep -q .; then
  printf 'private key/certificate-like file found in publication root\n' >&2
  fail=1
fi

if [ "$fail" -ne 0 ]; then
  printf 'publication validation: FAILED\n' >&2
  exit 1
fi
printf 'publication validation: OK\n'
