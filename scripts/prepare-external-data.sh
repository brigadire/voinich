#!/usr/bin/env bash
set -euo pipefail

usage() {
  echo "usage: $0 ivtff [SOURCE [OUTPUT]] | astafiev [SOURCE [OUTPUT]]" >&2
  exit 2
}

check_sha256() {
  local expected="$1" file="$2" actual
  actual="$(sha256sum "$file" | awk '{print $1}')"
  if [[ "$actual" != "$expected" ]]; then
    echo "checksum mismatch for $file: expected $expected, got $actual" >&2
    exit 1
  fi
}

case "${1:-}" in
  ivtff)
    source_file="${2:-data/ZL3b-n.txt}"
    output_file="${3:-data_work/ZL3b-x7.txt}"
    ivtt_bin="${IVTT_BIN:-ivtt}"
    check_sha256 bf5b6d4ac1e3a51b1847a9c388318d609020441ccd56984c901c32b09beccafc "$source_file"
    command -v "$ivtt_bin" >/dev/null 2>&1 || {
      echo "IVTT not found: set IVTT_BIN to an external IVTT executable" >&2
      exit 1
    }
    mkdir -p "$(dirname "$output_file")"
    set +e
    "$ivtt_bin" -x7 "$source_file" "$output_file"
    ivtt_status=$?
    set -e
    if [[ "$ivtt_status" -ne 0 && "$ivtt_status" -ne 3 ]]; then
      echo "IVTT failed with exit status $ivtt_status" >&2
      exit "$ivtt_status"
    fi
    check_sha256 360d99583145ec549b80edfafdc3f93534f3a11b85a0d52997ba8425e92b87c2 "$output_file"
    ;;
  astafiev)
    source_file="${2:-data_test/astafiev-1000-culinar-receipts.txt}"
    output_file="${3:-data_test/astafiev-1000-culinar-receipts-prepared.txt}"
    check_sha256 7200ce6cc01398192b05cf7f0d34040f391a07cc55e95b64b94525697b6f1d5c "$source_file"
    go run ./cmd/codex_prepare prepare \
      -input "$source_file" \
      -output "$output_file" \
      -encoding windows-1251 \
      -case lower \
      -line-policy preserve
    check_sha256 ff67a4fbf2606be4409724722e3e4d426aed27bdbeec1698babd92bd2b5eba5a "$output_file"
    ;;
  *) usage ;;
esac
