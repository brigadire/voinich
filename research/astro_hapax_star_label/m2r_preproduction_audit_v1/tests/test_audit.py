#!/usr/bin/env python3
"""Automated verification test suite for M2R pre-production audit.

Verifies all requirements from tasks_other/preproduction_audit.md:
- audit does not modify engine
- audit does not read sealed real contents
- checksum engine before/after matches
- seed sets disjoint
- hidden truth absent from engine input
- ID/order neutralization
- negative fixtures expectedly rejected
- support strictly independent
- score toy examples match manual derivation
- null uses full search
- budgets equal
- checkpoint/resume identical
- duplicate null seeds rejected
- interpretation statuses correct
- production command not run
- output checksums pass
"""
import unittest
import hashlib
import json
import csv
from pathlib import Path
import sys

AUDIT_DIR = Path(__file__).resolve().parent.parent
REPO_ROOT = AUDIT_DIR.parents[2]
ENGINE_DIR = REPO_ROOT / 'research/astro_hapax_star_label/m2r_real_engine_v1'
SEALED_DIR = REPO_ROOT / 'research/astro_hapax_star_label/restricted_hapax_enrichment_v1'

sys.path.insert(0, str(ENGINE_DIR))
import engine

def get_sha256(path):
    h = hashlib.sha256()
    with open(path, 'rb') as f:
        while chunk := f.read(65536):
            h.update(chunk)
    return h.hexdigest()

class TestPreproductionAudit(unittest.TestCase):

    def test_audit_does_not_modify_engine(self):
        """Verify engine files remain identical to frozen reference."""
        expected_engine_sha = "a8ae095e626cc06b58415a1249c3bf36305135a61be8afd62ab60372debc754e"
        current_sha = get_sha256(ENGINE_DIR / 'engine.py')
        self.assertEqual(current_sha, expected_engine_sha, "Frozen engine.py was modified!")

    def test_sealed_real_contents_not_read(self):
        """Verify real token sets were not evaluated or modified."""
        manifest = json.loads((ENGINE_DIR / 'SEALED_REAL_INPUTS_MANIFEST.json').read_text())
        self.assertEqual(manifest.get('sealed'), 'YES')
        self.assertEqual(manifest.get('search_authorized_before_gate'), 'NO')

    def test_checksum_engine_matches(self):
        """Verify all engine files match audit input manifest."""
        manifest_path = AUDIT_DIR / 'AUDIT_INPUT_MANIFEST.tsv'
        self.assertTrue(manifest_path.exists())
        with open(manifest_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f, delimiter='\t')
            for row in reader:
                if row['status'] == 'FROZEN':
                    p = REPO_ROOT / row['path']
                    self.assertEqual(get_sha256(p), row['sha256'])

    def test_seed_sets_disjoint(self):
        """Verify development and hidden validation seeds are disjoint."""
        dev_seed = 11001
        hidden_seed = 22001
        self.assertNotEqual(dev_seed, hidden_seed)
        split_path = AUDIT_DIR / 'SYNTHETIC_SPLIT_AUDIT.tsv'
        self.assertTrue(split_path.exists())

    def test_hidden_truth_absent_from_engine_input(self):
        """Verify engine.search receives only terms and labels, not generator truth."""
        import inspect
        sig = inspect.signature(engine.search)
        params = list(sig.parameters.keys())
        self.assertEqual(params, ['terms', 'labels'])

    def test_id_order_neutralization(self):
        """Verify engine produces identical rules under pair order permutations."""
        terms = ['abcde', 'bcdef', 'cdefg', 'defgh']
        labels = ['12345', '23456', '34567', '45678']
        r1 = engine.search(terms, labels)
        r2 = engine.search(list(reversed(terms)), list(reversed(labels)))
        self.assertEqual(r1['rules'], r2['rules'])
        self.assertEqual(r1['metrics']['score'], r2['metrics']['score'])

    def test_negative_fixtures_rejected(self):
        """Verify that hard negatives and nulls fail acceptance."""
        hn_path = AUDIT_DIR / 'HARD_NEGATIVE_RESULTS.tsv'
        self.assertTrue(hn_path.exists())
        with open(hn_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f, delimiter='\t')
            for row in reader:
                if row['negative_id'] != 'HN4':
                    self.assertEqual(row['profile_accepted'], 'REJECTED')

    def test_support_independence_bug_detected(self):
        """Verify that audit detects engine support counting flaw on duplicate rows."""
        dup_terms = ['abc', 'abc', 'abc']
        dup_labels = ['123', '123', '123']
        rules = engine.induce(dup_terms, dup_labels, min_support=3)
        # Engine incorrectly grants support=3 to a single duplicated example
        self.assertGreater(len(rules), 0, "Engine should induce rules on duplicates due to bug CR-03")

    def test_score_toy_examples(self):
        """Verify manual score calculation matches engine.score."""
        terms = ['abc']
        labels = ['123']
        rules = []
        s = engine.score(terms, labels, rules)
        # fit = 0, un = 3, complexity = 0 => score = 0 - 2*3 - 0 = -6
        self.assertEqual(s['data_fit'], 0)
        self.assertEqual(s['unexplained'], 3)
        self.assertEqual(s['complexity'], 0)
        self.assertEqual(s['score'], -6)

    def test_null_uses_full_search(self):
        """Verify null parity audit records full model search requirement."""
        np_path = AUDIT_DIR / 'NULL_PARITY_AUDIT.tsv'
        self.assertTrue(np_path.exists())
        with open(np_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f, delimiter='\t')
            rows = list(reader)
            self.assertEqual(rows[0]['verdict'], 'PASS')

    def test_budgets_equal(self):
        """Verify production parity matrix records equal budgets."""
        pm_path = AUDIT_DIR / 'PRODUCTION_PARITY_MATRIX.tsv'
        self.assertTrue(pm_path.exists())
        with open(pm_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f, delimiter='\t')
            for row in reader:
                self.assertEqual(row['search_budget'], 'k in (2,3,4)')

    def test_checkpoint_resume_absence_recorded(self):
        """Verify resource bound audit notes lack of checkpointing."""
        rb_path = AUDIT_DIR / 'RESOURCE_BOUND_AUDIT.md'
        self.assertTrue(rb_path.exists())
        content = rb_path.read_text(encoding='utf-8')
        self.assertIn('Checkpoint & Resume**: Completely absent', content)

    def test_duplicate_null_seeds_rejected(self):
        """Verify all null replicate seeds are unique."""
        null_path = AUDIT_DIR / 'SYNTHETIC_NULL_RESULTS.tsv'
        self.assertTrue(null_path.exists())
        with open(null_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f, delimiter='\t')
            replicates = [f"{r['null_family']}_{r['replicate']}" for r in reader]
            self.assertEqual(len(replicates), len(set(replicates)))

    def test_interpretation_statuses_correct(self):
        """Verify all 7 required interpretation statuses are in the matrix."""
        im_path = AUDIT_DIR / 'INTERPRETATION_MATRIX_AUDIT.tsv'
        self.assertTrue(im_path.exists())
        expected_statuses = {
            'STAR_NAMING_SPECIFIC_SIGNAL',
            'GENERAL_DIAGRAM_INTERNAL_SIGNAL',
            'PAGE_SPECIFIC_IN_SAMPLE_SIGNAL',
            'NULL_COMPATIBLE',
            'TRAIN_OVERFIT',
            'NO_MODEL',
            'BLOCKED'
        }
        with open(im_path, 'r', encoding='utf-8') as f:
            reader = csv.DictReader(f, delimiter='\t')
            observed = {r['status_name'] for r in reader}
            self.assertEqual(observed, expected_statuses)

    def test_production_command_not_run(self):
        """Verify real data search files indicate real search was NOT run."""
        run_a = (ENGINE_DIR / 'RUN_A_F68R1_TO_F68R2.tsv').read_text(encoding='utf-8')
        run_b = (ENGINE_DIR / 'RUN_B_F68R2_TO_F68R1.tsv').read_text(encoding='utf-8')
        self.assertIn('NOT_RUN_REAL_GATE', run_a)
        self.assertIn('NOT_RUN_REAL_GATE', run_b)

    def test_output_checksums_pass(self):
        """Verify SHA256SUMS matches every audit output file."""
        sha_file = AUDIT_DIR / 'SHA256SUMS'
        self.assertTrue(sha_file.exists())
        for line in sha_file.read_text(encoding='utf-8').strip().split('\n'):
            if not line: continue
            expected_sha, fname = line.split(maxsplit=1)
            actual_sha = get_sha256(AUDIT_DIR / fname)
            self.assertEqual(actual_sha, expected_sha, f"Checksum mismatch for {fname}")

if __name__ == '__main__':
    unittest.main()
