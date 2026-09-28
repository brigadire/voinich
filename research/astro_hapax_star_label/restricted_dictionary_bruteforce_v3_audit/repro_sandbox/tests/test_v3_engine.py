#!/usr/bin/env python3
"""
Unit and regression tests for v3 writing system model and search components.
Uses standard library unittest.
"""
import unittest
import csv
import json
from pathlib import Path
import sys

SCRIPTS_DIR = Path(__file__).resolve().parent.parent / "scripts"
sys.path.insert(0, str(SCRIPTS_DIR))

from engine import (
    encode_word,
    compute_complexity,
    maximum_bipartite_matching,
    evaluate_table,
    SOURCE_ALPHABET,
    EVA_ALPHABET
)
from synthetic_generator import SealedSyntheticGenerator, load_lexicon_forms

BASE_DIR = Path(__file__).resolve().parent.parent

class TestV3Engine(unittest.TestCase):
    def test_lexicon_gate_l(self):
        lex_file = BASE_DIR / "HISTORICAL_STAR_LEXICON.tsv"
        self.assertTrue(lex_file.exists(), "HISTORICAL_STAR_LEXICON.tsv missing")
        
        with lex_file.open("r", encoding="utf-8") as f:
            reader = list(csv.DictReader(f, delimiter="\t"))
            
        identities = set(r["canonical_identity_id"] for r in reader)
        self.assertGreaterEqual(len(identities), 57, f"Expected >= 57 canonical identities, got {len(identities)}")
        self.assertGreaterEqual(len(reader), 200, f"Expected comprehensive attestations, got {len(reader)}")
        
        # Check 20 columns
        expected_cols = [
            "canonical_identity_id", "canonical_identity_name", "attestation_id",
            "attested_form", "normalized_form", "language", "script",
            "transliteration_system", "source_title", "source_author",
            "source_date_start", "source_date_end", "manuscript_or_edition",
            "folio_or_entry", "source_url_or_local_reference", "source_type",
            "historically_attested", "editorial_reconstruction", "confidence", "notes"
        ]
        for col in expected_cols:
            self.assertIn(col, reader[0], f"Missing column {col}")
            
        # Check deduplication within identity
        for ident in identities:
            forms = [r["normalized_form"] for r in reader if r["canonical_identity_id"] == ident]
            self.assertEqual(len(forms), len(set(forms)), f"Duplicate normalized form in {ident}")

    def test_structural_manifest(self):
        mani_file = BASE_DIR / "REAL_SCOPE_STRUCTURAL_MANIFEST.json"
        self.assertTrue(mani_file.exists())
        with mani_file.open("r", encoding="utf-8") as f:
            data = json.load(f)
        self.assertEqual(data["occurrences_total"], 57)
        self.assertEqual(data["page_split"], {"f68r1": 30, "f68r2": 27})
        self.assertFalse(data["restrictions_enforced"]["dictionary_scores_calculated"])
        self.assertFalse(data["restrictions_enforced"]["real_data_search_authorized"])

    def test_encoding_modes(self):
        table = {"a": "o", "b": "l", "c": "c", "d": "y"}
        # DROP_UNMAPPED
        res_drop = encode_word("abcd efg", table, deletion_mode="DROP_UNMAPPED", abbreviation="NONE")
        self.assertEqual(res_drop, "olcy")
        
        # KEEP
        res_keep = encode_word("abcd efg", table, deletion_mode="NONE", abbreviation="NONE")
        self.assertEqual(res_keep, "olcyefg")
        
        # Abbreviation
        res_susp1 = encode_word("abcd", table, deletion_mode="DROP_UNMAPPED", abbreviation="SUSPENSION_1")
        self.assertEqual(res_susp1, "olc")
        
        res_pref4 = encode_word("abcdefgh", {"a":"o","b":"l","c":"c","d":"y","e":"k","f":"r"}, deletion_mode="DROP_UNMAPPED", abbreviation="PREFIX_4")
        self.assertEqual(res_pref4, "olcy")

    def test_capacity_policy_independence(self):
        labels = [
            {"occurrence_id": "L1", "page_id": "p1", "token": "olcy"},
            {"occurrence_id": "L2", "page_id": "p2", "token": "olcy"}
        ]
        idx = {"olcy": {"STAR_ALDEBARAN"}}
        
        # PER_PAGE_CAPACITY_1
        match_per_page = maximum_bipartite_matching(labels, idx, "PER_PAGE_CAPACITY_1")
        self.assertEqual(len(match_per_page), 2)
        self.assertEqual(match_per_page["L1"], "STAR_ALDEBARAN")
        self.assertEqual(match_per_page["L2"], "STAR_ALDEBARAN")
        
        # GLOBAL_CAPACITY_1
        match_global = maximum_bipartite_matching(labels, idx, "GLOBAL_CAPACITY_1")
        self.assertEqual(len(match_global), 1)

    def test_order_invariance_engine(self):
        labels_a = [
            {"occurrence_id": "L1", "page_id": "p1", "token": "olcy"},
            {"occurrence_id": "L2", "page_id": "p1", "token": "kro"}
        ]
        labels_b = [
            {"occurrence_id": "L2", "page_id": "p1", "token": "kro"},
            {"occurrence_id": "L1", "page_id": "p1", "token": "olcy"}
        ]
        idx = {"olcy": {"STAR_A"}, "kro": {"STAR_B"}}
        
        mA = maximum_bipartite_matching(labels_a, idx, "PER_PAGE_CAPACITY_1")
        mB = maximum_bipartite_matching(labels_b, idx, "PER_PAGE_CAPACITY_1")
        self.assertEqual(mA, mB)

if __name__ == "__main__":
    unittest.main()
