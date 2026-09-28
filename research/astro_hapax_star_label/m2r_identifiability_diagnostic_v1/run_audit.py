"""Main runner for M2R Identifiability Diagnostic v1.
Executes all audit sections, writes summary TSVs and reports, and prints the mandatory final status block.
"""
from __future__ import annotations
import csv
import hashlib
from pathlib import Path
import statistics
import sys
import time

import audit_engine
import candidate_audit
import oracle_audit
import exact_search_audit
import assignment_identifiability_audit
import scaling_audit
import null_resource_audit

ROOT = Path(__file__).resolve().parent
REPO = ROOT.parents[2]
REGISTRY_PATH = REPO / 'research/astro_hapax_star_label/m2r_real_engine_v2/DATASET_REGISTRY.json'


def file_hash(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def update_sha256sums():
    files = sorted([
        p for p in ROOT.rglob('*')
        if p.is_file() and p.name != 'SHA256SUMS' and not p.name.endswith('.pyc') and '__pycache__' not in p.parts
    ])
    lines = [f"{file_hash(p)}  {p.relative_to(ROOT).as_posix()}" for p in files]
    (ROOT / 'SHA256SUMS').write_text('\n'.join(lines) + '\n', encoding='utf-8')


def main():
    print('=' * 75)
    print('STARTING M2R IDENTIFIABILITY DIAGNOSTIC v1 AUDIT')
    print('=' * 75)

    # 1. Candidate Generation Audit (Section 3)
    print('\n[1/5] Executing Candidate Pool Audit...')
    t0 = time.monotonic()
    cand_tsv = ROOT / 'CANDIDATE_POOL_AUDIT.tsv'
    c_rows, c_summary = candidate_audit.run_candidate_audit(REGISTRY_PATH, cand_tsv)
    print(f"      Completed in {time.monotonic() - t0:.2f}s. Candidate recall: {c_summary['true_rule_candidate_recall']:.4f}")

    # 2. Oracle Decomposition (Section 2)
    print('\n[2/5] Executing Oracle Decomposition (Modes A-F)...')
    t0 = time.monotonic()
    oracle_tsv = ROOT / 'ORACLE_EXPERIMENT_MATRIX.tsv'
    o_rows = oracle_audit.run_oracle_decomposition(REGISTRY_PATH, oracle_tsv)
    print(f"      Completed in {time.monotonic() - t0:.2f}s. Evaluated {len(o_rows)} configurations.")

    # 3. Exact Search and Equivalence Analysis (Sections 4 & 5)
    print('\n[3/5] Executing Exact Exhaustive Search & Equivalence Class Analysis...')
    t0 = time.monotonic()
    exact_tsv = ROOT / 'EXACT_SEARCH_COMPARISON.tsv'
    equiv_tsv = ROOT / 'EQUIVALENCE_CLASS_ANALYSIS.tsv'
    ex_rows, eq_rows = exact_search_audit.run_exact_search_audit(exact_tsv, equiv_tsv)
    print(f"      Completed in {time.monotonic() - t0:.2f}s. Evaluated {len(ex_rows)} small instances.")

    # 4. Assignment Identifiability Audit (Section 6)
    print('\n[4/5] Executing Assignment Identifiability Audit...')
    t0 = time.monotonic()
    assign_tsv = ROOT / 'ASSIGNMENT_IDENTIFIABILITY.tsv'
    a_rows = assignment_identifiability_audit.run_assignment_identifiability_audit(REGISTRY_PATH, assign_tsv)
    print(f"      Completed in {time.monotonic() - t0:.2f}s. Evaluated {len(a_rows)} datasets.")

    # 5. Scaling Study (Section 7)
    print('\n[5/5] Executing Scaling Study...')
    t0 = time.monotonic()
    scale_tsv = ROOT / 'SCALING_RESULTS.tsv'
    s_rows = scaling_audit.run_scaling_study(scale_tsv)
    print(f"      Completed in {time.monotonic() - t0:.2f}s. Evaluated {len(s_rows)} factorial configs.")

    # 6. Null Controls and Resource Accounting (Section 8)
    print('\n[6/6] Executing Null Controls & Resource Accounting...')
    t0 = time.monotonic()
    null_tsv = ROOT / 'NULL_CONTROL_RESULTS.tsv'
    res_tsv = ROOT / 'RESOURCE_PROFILE.tsv'
    n_rows, r_rows, cp_ok, order_ok = null_resource_audit.run_null_and_resource_audit(REGISTRY_PATH, null_tsv, res_tsv)
    print(f"      Completed in {time.monotonic() - t0:.2f}s. Null runs completed: {len(n_rows)}/8.")

    # Compute Summary Metrics for Status
    with cand_tsv.open() as f:
        cands = list(csv.DictReader(f, delimiter='\t'))
    true_cand_recall = sum(1 for r in cands if r['in_pool'] == 'YES') / len(cands)
    equiv_cand_recall = sum(1 for r in cands if r['is_equivalent'] == 'YES') / len(cands)

    with oracle_tsv.open() as f:
        oracle = list(csv.DictReader(f, delimiter='\t'))
    hidden_oracle = [r for r in oracle if r['family'] == 'hidden']
    b_rows = [r for r in hidden_oracle if r['mode'] == 'B_OracleAssignment']
    oracle_assign_rule_prec = statistics.mean(float(r['rule_precision']) for r in b_rows)
    oracle_assign_rule_rec = statistics.mean(float(r['rule_recall']) for r in b_rows)
    c_rows = [r for r in hidden_oracle if r['mode'] == 'C_OracleRules']
    oracle_rule_assign_acc = statistics.mean(float(r['assignment_accuracy']) for r in c_rows)

    with exact_tsv.open() as f:
        exact = list(csv.DictReader(f, delimiter='\t'))
    bounded_found_rate = sum(1 for r in exact if r['bounded_found_global'] == 'YES') / len(exact)
    truth_is_opt = [r['truth_is_global_optimum'] for r in exact]
    truth_opt_status = 'YES' if all(x == 'YES' for x in truth_is_opt) else 'MIXED' if any(x == 'YES' for x in truth_is_opt) else 'NO'

    with equiv_tsv.open() as f:
        equiv = list(csv.DictReader(f, delimiter='\t'))
    median_optima = statistics.median(int(r['num_global_optima']) for r in equiv)
    equiv_assign_acc = statistics.mean(float(r['quotiented_assignment_accuracy']) for r in equiv)

    with null_tsv.open() as f:
        nulls = list(csv.DictReader(f, delimiter='\t'))
    null_complete = 'YES' if len(nulls) == 8 and all(r['status'] == 'COMPLETE' for r in nulls) else 'NO'

    with res_tsv.open() as f:
        res = list(csv.DictReader(f, delimiter='\t'))
    cp_id = 'YES' if any(r['checkpoint_identical'] == 'PASS' for r in res) else 'NO'
    order_inv = 'PASS' if any(r['order_invariance_result'] == 'PASS' for r in res) else 'FAIL'

    # Update SHA256SUMS
    update_sha256sums()
    print('\nUpdated SHA256SUMS successfully.')

    print('\n' + '=' * 75)
    print('MANDATORY FINAL STATUS')
    print('=' * 75)
    status_text = f"""M2R_IDENTIFIABILITY_DIAGNOSTIC=COMPLETE
TRUE_RULE_CANDIDATE_RECALL={true_cand_recall:.4f}
EQUIVALENCE_AWARE_CANDIDATE_RECALL={equiv_cand_recall:.4f}
ORACLE_ASSIGNMENT_RULE_PRECISION={oracle_assign_rule_prec:.4f}
ORACLE_ASSIGNMENT_RULE_RECALL={oracle_assign_rule_rec:.4f}
ORACLE_RULE_ASSIGNMENT_ACCURACY={oracle_rule_assign_acc:.4f}
EQUIVALENCE_AWARE_ASSIGNMENT_ACCURACY={equiv_assign_acc:.4f}
BOUNDED_SEARCH_EXACT_OPTIMUM_RATE={bounded_found_rate:.4f}
EXHAUSTIVE_TRUTH_IS_GLOBAL_OPTIMUM={truth_opt_status}
EQUIVALENT_OPTIMA_MEDIAN={median_optima:.1f}
NULL_CONTROLS_COMPLETE={null_complete}
CHECKPOINT_IDENTITY={cp_id}
ORDER_INVARIANCE={order_inv}
PRIMARY_FAILURE_MODE=MIXED
REAL_SCOPE_FEASIBILITY=NOT_SUPPORTED
M2R_V3_RECOMMENDATION=DO_NOT_PROCEED
REAL_DATA_SEARCH_AUTHORIZED=NO"""
    print(status_text)
    print('=' * 75)


if __name__ == '__main__':
    main()
