#!/usr/bin/env python3
"""Permanent EI01 role-separation and fixture regression."""

import csv
import hashlib
import hmac
import importlib.util
import json
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
REPO = ROOT.parents[2]
spec = importlib.util.spec_from_file_location("fixture", ROOT / "reference/blind_id_fixture_generator.py")
fixture = importlib.util.module_from_spec(spec)
spec.loader.exec_module(fixture)

erratum = json.loads((ROOT / "G1V2_EXECUTION_IDENTITY_ERRATUM_E1.json").read_text())
vectors = json.loads((ROOT / "G1V2_BLIND_ID_TEST_VECTORS.json").read_text())
with (REPO / "research/phase3/task85c-c/registries/G1V2_RNG_DOMAIN_REGISTRY.tsv").open(newline="") as f:
    rng = {r["domain_id"]: r for r in csv.DictReader(f, delimiter="\t")}

assert rng["CONTROL_GENERATE"]["consumer"] == "control builder"
assert rng["BLIND_ID"]["consumer"] == "escrow builder"
assert erratum["rng"]["role"] == "SCIENTIFIC_RANDOM_CONTROL_GENERATION"
assert erratum["escrow"]["role"] == "BLINDNESS_AND_TRUTH_COMMITMENT"
assert "escrow_key" in erratum["rng"]["independent_of"]
assert erratum["scientific_boundary"]["scientific_identity_depends_on_blind_id"] is False
assert erratum["scientific_boundary"]["scientific_identity_depends_on_escrow_key"] is False

cases = {x["id"]: x for x in vectors["cases"]}
assert cases["FIXED-A"]["expected_blind_id"] == cases["FIXED-A-REPEAT"]["expected_blind_id"]
assert cases["FIXED-A"]["scientific_control_identity_sha256"] == cases["DIFFERENT-KEY"]["scientific_control_identity_sha256"]
assert cases["FIXED-A"]["expected_blind_id"] != cases["DIFFERENT-KEY"]["expected_blind_id"]
assert cases["FIXED-A"]["scientific_control_identity_sha256"] != cases["CHANGED-SEED"]["scientific_control_identity_sha256"]
assert cases["FIXED-A"]["truth_record"]["content_sha256"] != cases["CHANGED-CORPUS"]["truth_record"]["content_sha256"]

base = cases["FIXED-A"]
key = bytes.fromhex(base["key_hex"])
raw = fixture.canonical(base["truth_record"])
old_unframed = hmac.new(key, raw, hashlib.sha256).hexdigest()[:20]
assert old_unframed != base["expected_blind_id"]
assert base["scientific_control_identity_sha256"] == base["truth_record"]["scientific_control_identity_sha256"]

invalid = dict(base["truth_record"], blind_id="forbidden")
try:
    fixture.blind_id(key, invalid)
except fixture.E1Error as exc:
    assert str(exc) == "truth record field closure"
else:
    raise AssertionError("blind_id accepted in its own HMAC input")

try:
    fixture.blind_id(b"short", base["truth_record"])
except fixture.E1Error as exc:
    assert "32 bytes" in str(exc)
else:
    raise AssertionError("invalid escrow key length accepted")

print("EI01_REGRESSION=PASS")
print("SCIENTIFIC_CONTROL_GENERATION_INDEPENDENT_OF_OPAQUE_ID=YES")
