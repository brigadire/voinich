from __future__ import annotations

import csv
import hashlib
import importlib.util
import json
import shutil
import tempfile
import unittest
from collections import Counter
from pathlib import Path


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/residual_ai_mapping_v1"
BASE = ROOT / "research/astro_hapax_star_label"
LEGACY = BASE / "legacy_mapping_migration_v1"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def rows(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as f:
        return list(csv.DictReader(f, delimiter="\t"))


def jsonl(path: Path) -> list[dict]:
    return [json.loads(line) for line in path.read_text(encoding="utf-8").splitlines()]


class ResidualAIMappingTests(unittest.TestCase):
    def test_protocol_was_frozen_before_outputs(self):
        freeze = json.loads((OUT / "PROTOCOL_FREEZE.json").read_text(encoding="utf-8"))
        self.assertEqual(freeze["status"], "FROZEN_BEFORE_AI_PASS")
        self.assertEqual(freeze["sha256"], sha(OUT / "RESIDUAL_AI_MAPPING_PROTOCOL.md"))

    def test_upstream_checksums_and_universes(self):
        for row in rows(OUT / "INPUT_MANIFEST.tsv"):
            self.assertEqual(row["sha256"], sha(ROOT / row["path"]), row["path"])
        mapping = rows(LEGACY / "LABEL_3G1_TRANSCRIPTION_MAPPING.tsv")
        residual = rows(LEGACY / "RESIDUAL_LABEL_MAPPING_QUEUE.tsv")
        self.assertEqual(len(mapping), 92)
        self.assertEqual(sum(r["primary_analysis_inclusion"] == "YES" for r in mapping), 21)
        self.assertEqual(len(residual), 71)
        self.assertEqual(Counter(r["panel"] for r in residual), Counter({"f68r1": 16, "f68r2": 33, "f68r3": 22}))

    def test_deterministic_split(self):
        split = rows(OUT / "CALIBRATION_SPLIT.tsv")
        self.assertEqual(len(split), 21)
        self.assertEqual(Counter(r["split_role"] for r in split), Counter({"CALIBRATION_VISIBLE": 14, "EVALUATION_HIDDEN": 7}))
        self.assertEqual(Counter(r["spatial_stratum"] for r in split), Counter({str(i): 3 for i in range(1, 8)}))
        self.assertTrue(all(sum(r["spatial_stratum"] == str(i) and r["split_role"] == "EVALUATION_HIDDEN" for r in split) == 1 for i in range(1, 8)))

    def test_candidate_universe_and_order_permutation(self):
        base = rows(OUT / "cleanroom/alignment_b_base/CANDIDATES.tsv")
        shuffled = rows(OUT / "cleanroom/alignment_b_shuffled/CANDIDATES.tsv")
        self.assertEqual(len(base), 268)
        self.assertEqual(Counter(r["panel"] for r in base), Counter({"f68r1": 69, "f68r2": 89, "f68r3": 110}))
        self.assertEqual({r["candidate_id"] for r in base}, {r["candidate_id"] for r in shuffled})
        for panel in ("f68r1", "f68r2", "f68r3"):
            a = [r["candidate_id"] for r in base if r["panel"] == panel]
            b = [r["candidate_id"] for r in shuffled if r["panel"] == panel]
            self.assertNotEqual(a, b)
            self.assertEqual(set(a), set(b))

    def test_candidate_byte_identity_with_frozen_occurrences(self):
        candidate = {int(r["occurrence_id"]): r for r in rows(BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv")}
        frozen = {}
        with (ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl").open(encoding="utf-8") as f:
            for line in f:
                record = json.loads(line)
                position = int(record["absolute_token_position"])
                if position in candidate:
                    frozen[position] = record
        self.assertEqual(set(frozen), set(candidate))
        for position, record in frozen.items():
            self.assertEqual(record["token"].replace("\x1f", "/"), candidate[position]["canonical_token_key"])

    def test_cleanroom_blindness(self):
        canonical_ids = {r["canonical_label_id"] for r in rows(OUT / "sealed_answers/NEUTRAL_LABEL_ID_CROSSWALK.tsv")}
        forbidden_headers = ("canonical_label_id", "group_id", "star_count", "human_added", "hapax", "frequency", "correct_candidate_id", "correct_occurrence_id")
        for path in (OUT / "cleanroom").rglob("*"):
            if not path.is_file() or path.suffix.lower() not in {".md", ".tsv", ".json", ".jsonl"}:
                continue
            text = path.read_text(encoding="utf-8")
            self.assertFalse(any(value in text for value in canonical_ids), path)
            header = text.splitlines()[0].lower() if text.splitlines() else ""
            self.assertFalse(any(value in header for value in forbidden_headers), path)
        self.assertIn("Result: **PASS**", (OUT / "BLINDNESS_AUDIT.md").read_text(encoding="utf-8"))

    def test_independent_pass_schemas_and_ids(self):
        expected = [r["neutral_label_id"] for r in rows(OUT / "CALIBRATION_SPLIT.tsv")]
        files = {
            "AI_VISUAL_A_RESULTS.jsonl": {"visual_glyph_sequence", "token_boundaries", "confidence"},
            "AI_VISUAL_C_RESULTS.jsonl": {"visual_glyph_sequence", "token_boundaries", "confidence"},
            "AI_ALIGNMENT_B_RESULTS.jsonl": {"ranked_candidate_ids", "glyph_alignment", "top1_confidence"},
            "AI_ALIGNMENT_B_SHUFFLED_RESULTS.jsonl": {"ranked_candidate_ids", "glyph_alignment", "top1_confidence"},
            "AI_ADJUDICATION_RESULTS.jsonl": {"selected_candidate_ids", "forced_answer", "stability"},
        }
        for name, keys in files.items():
            data = jsonl(OUT / name)
            self.assertEqual([r["neutral_label_id"] for r in data], expected, name)
            self.assertTrue(all(keys <= set(r) for r in data), name)
        sessions = rows(OUT / "AI_SESSION_MANIFEST.tsv")
        annotation = [r for r in sessions if r["pass"] != "NEGATIVE_CONTROLS"]
        self.assertEqual(len({r["session"] for r in annotation}), len(annotation))
        self.assertGreaterEqual(len({r["model_family"] for r in annotation}), 4)

    def test_selected_occurrences_exist_and_raw_binding(self):
        candidate = {f"ZC_{int(r['occurrence_id']):05d}": r for r in rows(BASE / "TRANSCRIPTION_TOKEN_CANDIDATES.tsv")}
        for name in ("AI_ALIGNMENT_B_RESULTS.jsonl", "AI_ALIGNMENT_B_SHUFFLED_RESULTS.jsonl"):
            for row in jsonl(OUT / name):
                self.assertTrue(all(candidate[value]["panel"] == "f68r1" for value in row["ranked_candidate_ids"]))
        for row in jsonl(OUT / "AI_ADJUDICATION_RESULTS.jsonl"):
            selected = row["selected_candidate_ids"]
            self.assertTrue(all(value in candidate for value in selected))
            self.assertTrue(all(candidate[value]["panel"] == "f68r1" for value in selected))
            self.assertFalse(row["forced_answer"])
            if len(selected) == 1 and len(row["raw_tokens"]) == 1:
                self.assertEqual(row["raw_tokens"][0], candidate[selected[0]]["readable_eva"])

    def test_hidden_gate_and_stability(self):
        hidden = rows(OUT / "HIDDEN_EVALUATION_RESULTS.tsv")
        self.assertEqual(len(hidden), 7)
        self.assertEqual(sum(r["exact_occurrence"] == "YES" for r in hidden), 1)
        self.assertEqual(sum(r["stability"] == "ORDER_STABLE" for r in hidden), 0)
        stability = rows(OUT / "CANDIDATE_ORDER_STABILITY.tsv")
        self.assertEqual(len(stability), 21)
        self.assertEqual(Counter(r["stability_class"] for r in stability), Counter({"UNSTABLE": 20, "ORDER_SENSITIVE": 1}))
        report = (OUT / "AI_MAPPING_REPORT.md").read_text(encoding="utf-8")
        self.assertIn("RESIDUAL_AI_MAPPING_RUN_AUTHORIZED=NO", report)
        self.assertIn("HIDDEN_EXACT_ACCURACY=1/7", report)
        self.assertIn("HIDDEN_ORDER_STABLE=0/7", report)

    def test_negative_controls_and_failed_gate_outputs(self):
        controls = rows(OUT / "NEGATIVE_CONTROL_RESULTS.tsv")
        self.assertEqual(len(controls), 21)
        self.assertTrue(all(r["control_pass"] == "YES" and r["abstained"] == "YES" for r in controls))
        self.assertFalse((OUT / "RESIDUAL_AI_MAPPING_RESULTS.tsv").exists())
        self.assertFalse((OUT / "AI_CONSENSUS_LABEL_TOKEN_MAPPING.tsv").exists())
        unresolved = rows(OUT / "UNRESOLVED_LABELS.tsv")
        self.assertEqual(len(unresolved), 71)
        self.assertTrue(all(r["status"] == "NOT_RUN_GATE_FAILED" and r["ai_mapping_attempted"] == "NO" for r in unresolved))

    def test_no_forbidden_analysis_or_cartesian_group_expansion(self):
        forbidden = {
            "HAPAX_ANALYTIC_COHORTS.tsv", "HAPAX_ENRICHMENT_RESULTS.tsv",
            "HAPAX_ENRICHMENT_METRICS.json", "HAPAX_ENRICHMENT_REPORT.md",
            "LEXICON_MATCH_RESULTS.tsv",
        }
        self.assertFalse(any((OUT / name).exists() for name in forbidden))
        # No AI-facing row has group identity; post-processing bias tables have one row per cohort only.
        self.assertTrue(all("group" not in key.lower() for row in jsonl(OUT / "AI_ADJUDICATION_RESULTS.jsonl") for key in row))
        self.assertEqual(len(rows(OUT / "MAPPING_SELECTION_BIAS_AUDIT.tsv")), 4)

    def test_hash_ledgers_and_package_hash(self):
        for line in (OUT / "SHA256SUMS").read_text(encoding="utf-8").splitlines():
            expected, relative = line.split(None, 1)
            self.assertEqual(expected, sha(OUT / relative.strip()), relative)
        state = json.loads((OUT / "ADJUDICATION_PRE_RUN_STATE.json").read_text(encoding="utf-8"))
        adjudication = jsonl(OUT / "AI_ADJUDICATION_RESULTS.jsonl")
        self.assertTrue(all(r["cleanroom_package_hash"] == state["input_package_hash"] for r in adjudication))

    def test_deterministic_preparation_rerun(self):
        script = OUT / "scripts/prepare_pipeline.py"
        spec = importlib.util.spec_from_file_location("prepare_pipeline_rerun", script)
        self.assertIsNotNone(spec)
        module = importlib.util.module_from_spec(spec)
        assert spec.loader is not None
        spec.loader.exec_module(module)
        with tempfile.TemporaryDirectory() as temp:
            temp_out = Path(temp) / "residual_ai_mapping_v1"
            temp_out.mkdir()
            shutil.copyfile(OUT / "RESIDUAL_AI_MAPPING_PROTOCOL.md", temp_out / "RESIDUAL_AI_MAPPING_PROTOCOL.md")
            module.OUT = temp_out
            module.PROTOCOL = temp_out / "RESIDUAL_AI_MAPPING_PROTOCOL.md"
            module.main()
            for relative in (
                "PROTOCOL_FREEZE.json", "INPUT_MANIFEST.tsv", "CALIBRATION_SPLIT.tsv",
                "PRE_RUN_STATE.json",
                "sealed_answers/NEUTRAL_LABEL_ID_CROSSWALK.tsv",
                "sealed_answers/EVALUATION_HIDDEN_ANSWERS.tsv",
            ):
                self.assertEqual(sha(OUT / relative), sha(temp_out / relative), relative)
            rerun_manifest = rows(temp_out / "CLEAN_ROOM_PACKAGE_MANIFEST.tsv")
            current_manifest = rows(OUT / "CLEAN_ROOM_PACKAGE_MANIFEST.tsv")
            base_packages = {"visual_a", "alignment_b_base", "visual_c", "alignment_b_shuffled"}
            current_base = [r for r in current_manifest if r["package"] in base_packages]
            self.assertEqual(rerun_manifest, current_base)


if __name__ == "__main__":
    unittest.main()
