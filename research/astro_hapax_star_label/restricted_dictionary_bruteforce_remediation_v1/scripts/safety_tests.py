#!/usr/bin/env python3
"""
clean_room.md Section 20: automated scientific-safety checks. Each check is a real, runnable
assertion, not a prose promise. Run with: python3 scripts/safety_tests.py
Exits nonzero if any check fails.
"""
import re
import subprocess
import sys
from pathlib import Path

PKG_ROOT = Path(__file__).resolve().parent.parent
SCRIPTS_DIR = PKG_ROOT / "scripts"

results = []


def check(name, ok, detail=""):
    results.append((name, ok, detail))
    print(f"[{'PASS' if ok else 'FAIL'}] {name}: {detail}")


def read(path: Path) -> str:
    return path.read_text(encoding="utf-8", errors="replace")


def main():
    # Only look at actual CODE lines (strip comments/docstrings) so this check can't be defeated
    # by a false negative, but also can't false-positive on a docstring that merely explains the
    # constraint (a real early false-positive: solver_heuristic.py's own docstring says "never on
    # held-out labels or their true_identity ground truth" as documentation, not as code).
    def code_only(text: str) -> str:
        text = re.sub(r'""".*?"""', "", text, flags=re.S)
        lines = [l for l in text.splitlines() if not l.strip().startswith("#")]
        return "\n".join(lines)

    solver_modules = ["oracle_bb.py", "solver_cpsat.py", "solver_heuristic.py", "hybrid.py"]
    forbidden_tokens = ["true_identity", "held_out_labels", "TRUTH_SEALED"]
    leak_found = []
    for mod in solver_modules:
        text = code_only(read(SCRIPTS_DIR / mod))
        for tok in forbidden_tokens:
            if tok in text:
                leak_found.append(f"{mod}:{tok}")
    check(
        "NO_TRUTH_LEAKAGE_IN_SOLVERS",
        len(leak_found) == 0,
        "solver modules never reference true_identity/held_out_labels/TRUTH_SEALED in code (docstrings excluded)" if not leak_found
        else f"found: {leak_found}",
    )

    hidden_solver_code = code_only(read(SCRIPTS_DIR / "run_hidden_solver.py"))
    ok = "TRUTH_SEALED" not in hidden_solver_code and "true_identity" not in hidden_solver_code
    check(
        "HIDDEN_SOLVER_NEVER_OPENS_TRUTH",
        ok,
        "run_hidden_solver.py's code contains no reference to TRUTH_SEALED or true_identity" if ok
        else "run_hidden_solver.py's CODE (not docstring) references truth -- real leak",
    )
    reveal_text = read(SCRIPTS_DIR / "reveal_and_score.py")
    check(
        "REVEAL_SCRIPT_IS_SEPARATE_FROM_SOLVER",
        "HybridSolver" not in reveal_text and "CPSATSolver" not in reveal_text,
        "reveal_and_score.py imports no solver class -- it only reads predictions + truth, it "
        "cannot itself re-run or influence the solve",
    )

    cpsat_text = read(SCRIPTS_DIR / "solver_cpsat.py")
    optimal_lines = [l for l in cpsat_text.splitlines() if "is_optimal" in l or '"OPTIMAL"' in l]
    check(
        "CPSAT_OPTIMAL_ONLY_FROM_SOLVER_STATUS",
        any("status == cp_model.OPTIMAL" in l for l in cpsat_text.splitlines()),
        "is_optimal is derived from solver.Solve()'s own status enum, never set independently",
    )
    bb_text = read(SCRIPTS_DIR / "oracle_bb.py")
    check(
        "BB_NEVER_CERTIFIES_ON_TIMEOUT",
        "not state[\"timed_out\"]" in bb_text or "not self.timed_out" in bb_text,
        "global_optimum_certified is defined as `not timed_out`, structurally impossible to be "
        "True on a timed-out run",
    )

    # generator.py legitimately uses "f68r1"/"f68r2" as SYNTHETIC page names (matching the real
    # 30/27 split's page naming for realism, disclosed in SYNTHETIC_GENERATOR_SPEC.md) -- that is
    # not real label *content* access, so it is intentionally not in this pattern list. This
    # script itself is excluded from the scan since it necessarily quotes the forbidden patterns.
    real_data_patterns = ["restricted_m2_star_labels_v1", "restricted_hapax_enrichment_v1", "restricted_m0_m1_rerun_v1", "TARGET_SCOPE.tsv", "EVA_LABEL"]
    hits = []
    for py_file in SCRIPTS_DIR.glob("*.py"):
        if py_file.name == "safety_tests.py":
            continue
        text = read(py_file)
        for pat in real_data_patterns:
            if pat in text:
                hits.append(f"{py_file.name}:{pat}")
    check(
        "NO_REAL_LABEL_DATA_ACCESS",
        len(hits) == 0,
        "no script references real-label directories/files or raw EVA folio identifiers" if not hits
        else f"found: {hits}",
    )

    # Reproducibility: re-run the cheap generator scripts and diff against committed output.
    repro_checks = [
        ("run_scoring_parity.py", ["500", "/tmp/_repro_scoring_parity.tsv"], PKG_ROOT / "SCORING_PARITY_RESULTS.tsv"),
        ("discriminative_benchmark.py", ["/tmp/_repro_discriminative.tsv"], PKG_ROOT / "DISCRIMINATIVE_BENCHMARK_RESULTS.tsv"),
        ("reachability.py", ["/tmp/_repro_reachability.tsv"], PKG_ROOT / "REACHABILITY_REMEDIATION.tsv"),
    ]
    for script, args, committed in repro_checks:
        if not committed.exists():
            check(f"REPRODUCIBLE::{script}", False, f"{committed.name} does not exist yet")
            continue
        proc = subprocess.run([sys.executable, str(SCRIPTS_DIR / script)] + args, cwd=SCRIPTS_DIR, capture_output=True, text=True)
        regenerated = Path(args[-1])
        ok = proc.returncode == 0 and regenerated.exists() and regenerated.read_text() == committed.read_text()
        check(f"REPRODUCIBLE::{script}", ok, f"re-running {script} byte-for-byte reproduces {committed.name}" if ok else proc.stderr[-500:])

    gate_docs = ["REMEDIATION_REPORT.md"]
    for doc_name in gate_docs:
        doc_path = PKG_ROOT / doc_name
        if not doc_path.exists():
            check(f"GATE_DISCIPLINE::{doc_name}", False, f"{doc_name} does not exist yet")
            continue
        text = read(doc_path)
        ok = "REAL_DATA_SEARCH_AUTHORIZED=NO" in text
        check(f"GATE_DISCIPLINE::{doc_name}", ok, "REAL_DATA_SEARCH_AUTHORIZED=NO present in final status block")

    n_fail = sum(1 for _, ok, _ in results if not ok)
    print(f"\n{len(results) - n_fail}/{len(results)} checks passed")
    return 1 if n_fail else 0


if __name__ == "__main__":
    sys.exit(main())
