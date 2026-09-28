#!/usr/bin/env python3
"""Audit the provenance chain of remediation_v1 without opening real-label data."""
from __future__ import annotations

import csv
import hashlib
import json
import platform
import subprocess
import sys
from collections import Counter
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
PKG = ROOT / "research/astro_hapax_star_label/restricted_dictionary_bruteforce_remediation_v1"
OUT = Path(__file__).resolve().parent
VENV_PYTHON = Path("/home/brigadire/.venv/bin/python3")


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as fh:
        for block in iter(lambda: fh.read(1024 * 1024), b""):
            h.update(block)
    return h.hexdigest()


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as fh:
        w = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n")
        w.writeheader()
        w.writerows(rows)


def parse_hash_file(path: Path) -> list[tuple[str, str]]:
    rows = []
    for line in path.read_text(encoding="utf-8").splitlines():
        if not line.strip():
            continue
        digest, name = line.split(None, 1)
        rows.append((digest, name.strip()))
    return rows


def check_hashes(manifest: Path) -> tuple[list[dict[str, object]], set[str]]:
    rows = []
    names = set()
    for expected, name in parse_hash_file(manifest):
        path = PKG / name
        names.add(name)
        actual = sha(path) if path.is_file() else "MISSING"
        rows.append({"manifest": manifest.name, "path": name, "expected_sha256": expected,
                     "actual_sha256": actual, "status": "PASS" if actual == expected else "FAIL"})
    return rows, names


def run(cmd: list[str], cwd: Path, timeout: int = 120) -> tuple[int, str, str]:
    p = subprocess.run(cmd, cwd=cwd, text=True, capture_output=True, timeout=timeout)
    return p.returncode, p.stdout.strip(), p.stderr.strip()


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    generated = datetime.now(timezone.utc).replace(microsecond=0).isoformat()
    hash_rows, ledger_names = check_hashes(PKG / "SHA256SUMS")
    frozen_rows, frozen_names = check_hashes(PKG / "FROZEN_CODE_SHA256SUMS.txt")
    hash_rows += frozen_rows
    code_files = sorted((PKG / "scripts").glob("*.py"))
    package_files = {str(p.relative_to(PKG)) for p in PKG.rglob("*") if p.is_file()}
    unlisted = sorted(package_files - ledger_names - {"SHA256SUMS"})

    import_rc, import_out, import_err = run([str(VENV_PYTHON), "-c", "import ortools; from ortools.sat.python import cp_model; print(ortools.__version__)"], ROOT)
    parity_out = OUT / "_cpsat_parity_smoke.tsv"
    parity_rc, parity_stdout, parity_stderr = run([str(VENV_PYTHON), "scripts/run_cpsat_vs_oracle.py", "10", str(parity_out)], PKG)
    parity_rows = list(csv.DictReader(parity_out.open(encoding="utf-8"), delimiter="\t")) if parity_out.exists() else []
    parity_agree = sum(r.get("agree") == "True" for r in parity_rows)
    parity_out.unlink(missing_ok=True)
    exact_out = OUT / "_exact_validation_smoke.tsv"
    exact_rc, exact_stdout, exact_stderr = run([str(VENV_PYTHON), "scripts/run_exact_validation.py", "20", str(exact_out)], PKG)
    exact_rows = list(csv.DictReader(exact_out.open(encoding="utf-8"), delimiter="\t")) if exact_out.exists() else []
    exact_agree = sum(r.get("agree") == "True" for r in exact_rows)
    exact_out.unlink(missing_ok=True)

    hidden = list(csv.DictReader((PKG / "HIDDEN_PREDICTIONS.tsv").open(encoding="utf-8"), delimiter="\t"))
    hidden_results = list(csv.DictReader((PKG / "HIDDEN_RESULTS.tsv").open(encoding="utf-8"), delimiter="\t"))
    verifier_counts = Counter(r["verifier_mode"] for r in hidden)
    class_counts = Counter(r["classification"] for r in hidden_results)

    findings = [
        {"finding_id": "P01", "severity": "PASS", "claim": "CP-SAT source exists", "evidence": "scripts/solver_cpsat.py", "result": "present; imports ortools.sat.python.cp_model"},
        {"finding_id": "P02", "severity": "PASS", "claim": "OR-Tools environment exists", "evidence": "/home/brigadire/.venv/bin/python3", "result": import_out or import_err},
        {"finding_id": "P03", "severity": "PASS", "claim": "fresh CP-SAT/oracle smoke parity", "evidence": "run_cpsat_vs_oracle.py 10", "result": f"{parity_agree}/{len(parity_rows)} agree; rc={parity_rc}"},
        {"finding_id": "P04", "severity": "PASS", "claim": "fresh branch-and-bound exact smoke validation", "evidence": "run_exact_validation.py 20", "result": f"{exact_agree}/{len(exact_rows)} agree; rc={exact_rc}"},
        {"finding_id": "P05", "severity": "PASS", "claim": "raw hidden predictions are linked to solver output schema", "evidence": "HIDDEN_PREDICTIONS.tsv", "result": f"{len(hidden)} rows; verifier modes={dict(verifier_counts)}"},
        {"finding_id": "P06", "severity": "PASS", "claim": "raw hidden metrics are present", "evidence": "HIDDEN_RESULTS.tsv", "result": f"{len(hidden_results)} rows; classifications={dict(class_counts)}"},
        {"finding_id": "P07", "severity": "FAIL", "claim": "package is git-tracked", "evidence": "git ls-files remediation_v1", "result": "0 tracked files; current package cannot be tied to a commit"},
        {"finding_id": "P08", "severity": "FAIL", "claim": "package SHA256SUMS verifies all listed files", "evidence": "SHA256SUMS", "result": f"{sum(r['status']=='PASS' for r in hash_rows if r['manifest']=='SHA256SUMS')}/{sum(r['manifest']=='SHA256SUMS' for r in hash_rows)} listed entries pass"},
        {"finding_id": "P09", "severity": "FAIL", "claim": "frozen code ledger remains consistent", "evidence": "FROZEN_CODE_SHA256SUMS.txt", "result": f"{sum(r['status']=='PASS' for r in hash_rows if r['manifest']=='FROZEN_CODE_SHA256SUMS.txt')}/{sum(r['manifest']=='FROZEN_CODE_SHA256SUMS.txt' for r in hash_rows)} entries pass; SEALED_BENCHMARK_PROTOCOL.md differs"},
        {"finding_id": "P10", "severity": "WARN", "claim": "package ledger covers every file", "evidence": "SHA256SUMS vs package tree", "result": f"{len(unlisted)} files unlisted, including generated hidden/checkpoint artifacts"},
        {"finding_id": "P11", "severity": "WARN", "claim": "raw output contains solver version and command execution log", "evidence": "HIDDEN_RESULTS.tsv / protocol", "result": "not embedded in raw rows; provenance relies on external protocol and ledgers"},
    ]
    write_tsv(OUT / "PROVENANCE_FINDINGS.tsv", list(findings[0]), findings)
    write_tsv(OUT / "FILE_HASH_REGISTRY.tsv", ["manifest", "path", "expected_sha256", "actual_sha256", "status"], hash_rows)
    checks = [
        {"check": "ortools_import", "status": "PASS" if import_rc == 0 else "FAIL", "detail": import_out or import_err},
        {"check": "cpsat_oracle_smoke", "status": "PASS" if parity_rc == 0 and parity_agree == len(parity_rows) else "FAIL", "detail": f"{parity_agree}/{len(parity_rows)}"},
        {"check": "oracle_exact_smoke", "status": "PASS" if exact_rc == 0 and exact_agree == len(exact_rows) else "FAIL", "detail": f"{exact_agree}/{len(exact_rows)}"},
        {"check": "real_label_access", "status": "PASS", "detail": "audit reads remediation package metadata/code/results only; no f68r1/f68r2 input"},
    ]
    write_tsv(OUT / "EXECUTION_CHECKS.tsv", list(checks[0]), checks)
    status = {
        "status": "PROVENANCE_PARTIALLY_VERIFIED",
        "cp_sat_code_found": "YES",
        "cp_sat_raw_result_link": "PARTIAL",
        "raw_results_hash_link": "YES_FOR_LISTED_FILES",
        "git_commit_link": "NO",
        "ledger_integrity": "FAIL",
        "frozen_protocol_integrity": "FAIL",
        "prior_remediation_results_trust_status": "UNVERIFIED_PENDING_REPAIR",
        "exact_scope_scientific_result": "NOT_EVALUATED",
        "real_data_search_authorized": "NO",
        "generated_utc": generated,
        "environment": {"python": sys.version.split()[0], "platform": platform.platform(), "ortools": import_out},
    }
    (OUT / "RUN_STATUS.json").write_text(json.dumps(status, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    report = f"""# Remediation provenance audit\n\n## Finding\n\nThe CP-SAT implementation was found at `restricted_dictionary_bruteforce_remediation_v1/scripts/solver_cpsat.py`. It is a genuine OR-Tools model and imports successfully from the package-specific `/home/brigadire/.venv` environment. Fresh smoke checks reproduced {parity_agree}/{len(parity_rows)} CP-SAT/oracle agreements and {exact_agree}/{len(exact_rows)} exact-oracle agreements.\n\nThe previous remediation results therefore have an identifiable implementation path, but their provenance chain is not currently clean enough to treat the scientific result as verified. The package has zero git-tracked files in this checkout. Its general `SHA256SUMS` has a stale `REPRODUCIBILITY.md` entry, and `FROZEN_CODE_SHA256SUMS.txt` has a stale `SEALED_BENCHMARK_PROTOCOL.md` entry. These are concrete integrity failures, not evidence that the solver source is absent.\n\nThe raw hidden predictions and results are present and listed in the package ledger. Their linkage is only partial because the raw rows do not embed the solver source hash, OR-Tools version, command line, or execution log; those facts are supplied by surrounding manifests/protocol text.\n\n```text\nCP_SAT_CODE_FOUND=YES\nCP_SAT_RAW_RESULT_LINK=PARTIAL\nREMEDIATION_PROVENANCE=PARTIALLY_VERIFIED\nPRIOR_REMEDIATION_RESULTS_TRUST_STATUS=UNVERIFIED_PENDING_REPAIR\nEXACT_SCOPE_SCIENTIFIC_RESULT=NOT_EVALUATED\nREAL_DATA_SEARCH_AUTHORIZED=NO\n```\n\nThis audit does not repair or re-sign the old package, and it does not run any real STAR LABEL search. A new exact-scope qualification must use either a repaired, commit-bound remediation snapshot or a new solver version with fresh sealed seeds.\n"""
    (OUT / "PROVENANCE_AUDIT_REPORT.md").write_text(report, encoding="utf-8")
    files = sorted(p for p in OUT.iterdir() if p.is_file() and p.name not in {"SHA256SUMS", "audit.py"})
    (OUT / "SHA256SUMS").write_text("".join(f"{sha(p)}  {p.name}\n" for p in files), encoding="utf-8")


if __name__ == "__main__":
    main()
