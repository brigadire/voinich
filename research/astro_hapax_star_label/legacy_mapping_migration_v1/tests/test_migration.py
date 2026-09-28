from __future__ import annotations

import csv
import hashlib
import json
import subprocess
import tempfile
import unittest
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/legacy_mapping_migration_v1"
MATCHES = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv"
OCCURRENCES = ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"


def rows(name: str):
    with (OUT / name).open(encoding="utf-8", newline="") as handle:
        return list(csv.DictReader(handle, delimiter="\t"))


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class MigrationTests(unittest.TestCase):
    def test_required_outputs_and_gate(self):
        required = {
            "LEGACY_MAPPING_INPUT_MANIFEST.tsv",
            "LEGACY_MAPPING_REPRODUCTION.md",
            "HAPAX_STAR_LABEL_ANALYSIS_PLAN_AMENDMENT_01.md",
            "LEGACY_MAPPING_NORMALIZED.tsv",
            "LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv",
            "CROSSWALK_GEOMETRY_METRICS.tsv",
            "CROSSWALK_VISUAL_AUDIT.md",
            "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv",
            "MAPPING_COVERAGE_SUMMARY.tsv",
            "MAPPING_SELECTION_BIAS_AUDIT.tsv",
            "RESIDUAL_LABEL_MAPPING_QUEUE.tsv",
            "LEGACY_MAPPING_MIGRATION_REPORT.md",
            "VALIDATION_REPORT.md",
            "REPRODUCIBILITY.md",
            "SHA256SUMS",
        }
        self.assertTrue(required <= {path.name for path in OUT.iterdir() if path.is_file()})
        report = (OUT / "LEGACY_MAPPING_MIGRATION_REPORT.md").read_text(encoding="utf-8")
        self.assertIn("HAPAX_ENRICHMENT_RUN_AUTHORIZED=NO", report)
        self.assertIn("LEXICON_MATCH_RUN_AUTHORIZED=NO", report)
        for forbidden in (
            "HAPAX_ANALYTIC_COHORTS.tsv",
            "HAPAX_ENRICHMENT_RESULTS.tsv",
            "HAPAX_ENRICHMENT_METRICS.json",
            "HAPAX_ENRICHMENT_REPORT.md",
        ):
            self.assertFalse((OUT / forbidden).exists())

    def test_reproduction_and_normalization(self):
        normalized = rows("LEGACY_MAPPING_NORMALIZED.tsv")
        self.assertEqual(len(normalized), 191)
        self.assertEqual(Counter(row["match_status"] for row in normalized), Counter({"MATCHED": 130, "UNMATCHED": 61}))
        self.assertEqual(len({row["legacy_coordinates"] for row in normalized}), 143)
        self.assertEqual(sum(int(row["sensitivity_only_potential_occurrences"]) for row in normalized), 77)
        self.assertEqual(sum(row["confirmation_status"] == "RULE_CONFIRMED_FROZEN_OCCURRENCE" for row in normalized), 130)
        report = (OUT / "LEGACY_MAPPING_REPRODUCTION.md").read_text(encoding="utf-8")
        for value in ("7/7", "10/12", "53/67", "54", "77"):
            self.assertIn(value, report)

    def test_target_and_outcomes(self):
        crosswalk = rows("LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv")
        self.assertEqual(len(crosswalk), 92)
        self.assertEqual(len({row["canonical_label_id"] for row in crosswalk}), 92)
        self.assertEqual(Counter(row["panel"] for row in crosswalk), Counter({"f68r1": 37, "f68r2": 33, "f68r3": 22}))
        self.assertEqual(Counter(row["crosswalk_outcome"] for row in crosswalk), Counter({
            "NO_LEGACY_CANDIDATE": 28,
            "AMBIGUOUS_CROSSWALK": 28,
            "UNIQUE_PROVENANCE_CROSSWALK": 21,
            "HUMAN_ADDED_NOT_IN_LEGACY": 10,
            "LEGACY_UNMATCHED": 5,
        }))
        primary = [row for row in crosswalk if row["primary_mapping_eligible"] == "YES"]
        self.assertEqual(len(primary), 21)
        self.assertTrue(all(row["panel"] == "f68r1" for row in primary))
        self.assertTrue(all(row["crosswalk_outcome"] == "UNIQUE_PROVENANCE_CROSSWALK" for row in primary))
        self.assertTrue(all(len(row["legacy_coordinates"].split(";")) == 1 for row in primary))

    def test_3g1_and_human_added(self):
        mapping = rows("LABEL_3G1_TRANSCRIPTION_MAPPING.tsv")
        self.assertEqual(Counter(row["grouped"] for row in mapping), Counter({"YES": 64, "NO": 28}))
        self.assertEqual(sum(row["primary_analysis_inclusion"] == "YES" for row in mapping), 21)
        self.assertEqual(sum(row["grouped"] == "YES" and row["primary_analysis_inclusion"] == "YES" for row in mapping), 21)
        self.assertEqual(sum(row["grouped"] == "NO" and row["primary_analysis_inclusion"] == "YES" for row in mapping), 0)
        human = [row for row in mapping if row["human_added_membership"] == "YES"]
        self.assertEqual(len(human), 10)
        self.assertTrue(all(row["primary_analysis_inclusion"] == "NO" and not row["absolute_token_positions"] for row in human))
        eight = [row for row in mapping if row["group_size"] == "8"]
        self.assertEqual(len(eight), 1)
        self.assertEqual(eight[0]["star_count"], "7")
        self.assertEqual(eight[0]["label_count"], "1")
        self.assertEqual(eight[0]["multi_member_group_flag"], "YES")

    def test_occurrences_and_source_rows_exist(self):
        with MATCHES.open(encoding="utf-8", newline="") as handle:
            source = {row["record_id"]: row for row in csv.DictReader(handle, delimiter="\t")}
        occurrence_ids = set()
        with OCCURRENCES.open(encoding="utf-8") as handle:
            for line in handle:
                occurrence_ids.add(str(json.loads(line)["absolute_token_position"]))
        for row in rows("LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv"):
            if row["primary_mapping_eligible"] != "YES":
                continue
            for record_id in row["legacy_record_ids"].split(";"):
                self.assertIn(record_id, source)
                self.assertEqual(source[record_id]["match_status"], "MATCHED")
                self.assertEqual(source[record_id]["panel"], row["panel"])
            for occurrence_id in row["absolute_token_positions"].split(";"):
                self.assertIn(occurrence_id, occurrence_ids)

    def test_geometry_and_no_cross_panel(self):
        metrics = rows("CROSSWALK_GEOMETRY_METRICS.tsv")
        self.assertEqual(len(metrics), 92)
        self.assertTrue(all(row["panel_identity"] == "PASS" for row in metrics))
        primary = [row for row in metrics if row["crosswalk_outcome"] == "UNIQUE_PROVENANCE_CROSSWALK"]
        self.assertEqual(len(primary), 21)
        self.assertTrue(all(row["coordinate_transform"] == "IDENTITY_PANEL_PIXEL_FRAME" for row in primary))
        self.assertTrue(all(row["token_exists"] == "YES" and row["raw_token_byte_identity"] == "YES" for row in primary))

    def test_freezes_and_checksums(self):
        self.assertEqual(sha(ROOT / "research/astro_hapax_star_label/HAPAX_STAR_LABEL_ANALYSIS_PLAN.md"), "02e1f553dd77a60e40b8ae3cae3b3aa50abe0fef50997455ce463fc63d246e7a")
        self.assertEqual(sha(OUT / "HAPAX_STAR_LABEL_ANALYSIS_PLAN_AMENDMENT_01.md"), "9f413ff52b5ba065a2b896e2096c008688400072bfd1dc1a0aa8688ffa25763f")
        self.assertEqual(sha(OCCURRENCES), "ba0342e15d8c468ec4e9f741e97cdb4a11938fe1f0ae3ac4338b73aaf1bd773a")
        for line in (OUT / "SHA256SUMS").read_text(encoding="utf-8").splitlines():
            expected, relative = line.split(None, 1)
            self.assertEqual(sha(OUT / relative.strip()), expected, relative)

    def test_visual_audit_scope(self):
        audit = rows("VISUAL_AUDIT_SELECTION.tsv")
        self.assertEqual(sum(row["audit_class"] == "AMBIGUOUS_OR_HUMAN_ADDED" for row in audit), 38)
        self.assertEqual(sum(row["audit_class"] == "DETERMINISTIC_PRIMARY_SAMPLE" for row in audit), 8)
        self.assertEqual(sum(row["audit_class"] == "ALL_F68R1_ORDINAL" for row in audit), 29)
        self.assertEqual(len(list((OUT / "visual_audit/crops").glob("*.png"))), 21)

    def test_byte_reproducibility_of_outputs(self):
        core = [
            "LEGACY_MAPPING_INPUT_MANIFEST.tsv",
            "LEGACY_MAPPING_REPRODUCTION.md",
            "LEGACY_MAPPING_NORMALIZED.tsv",
            "LEGACY_TO_CANONICAL_LABEL_CROSSWALK.tsv",
            "CROSSWALK_GEOMETRY_METRICS.tsv",
            "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv",
            "MAPPING_COVERAGE_SUMMARY.tsv",
            "MAPPING_SELECTION_BIAS_AUDIT.tsv",
            "RESIDUAL_LABEL_MAPPING_QUEUE.tsv",
            "LEGACY_MAPPING_MIGRATION_REPORT.md",
            "VALIDATION_REPORT.md",
            "REPRODUCIBILITY.md",
            "VISUAL_AUDIT_SELECTION.tsv",
            "visual_audit/F68R1_PROVENANCE_AUDIT.png",
            "visual_audit/AMBIGUOUS_AND_HUMAN_ADDED.png",
            "visual_audit/DETERMINISTIC_PRIMARY_SAMPLE.png",
        ]
        with tempfile.TemporaryDirectory() as temp:
            temp_out = Path(temp) / "migration"
            subprocess.run([
                "python3",
                str(OUT / "scripts/build_migration.py"),
                "--output",
                str(temp_out),
            ], cwd=ROOT, check=True)
            for relative in core:
                self.assertEqual(sha(OUT / relative), sha(temp_out / relative), relative)


if __name__ == "__main__":
    unittest.main()
