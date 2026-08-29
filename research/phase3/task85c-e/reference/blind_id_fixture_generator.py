#!/usr/bin/env python3
"""E1 reference implementation for non-production blind-ID fixtures."""

from __future__ import annotations

import hashlib
import hmac
import json
import struct
import sys
import unicodedata

BLIND_DOMAIN = b"G1V2-BLIND-ID"
SCIENCE_DOMAIN = b"G1V2-SCIENTIFIC-CONTROL\0"


class E1Error(ValueError):
    pass


def normalize(value):
    if isinstance(value, str):
        return unicodedata.normalize("NFC", value)
    if isinstance(value, list):
        return [normalize(x) for x in value]
    if isinstance(value, dict):
        out = {}
        for key in sorted(value, key=lambda x: unicodedata.normalize("NFC", x).encode("utf-8")):
            nk = unicodedata.normalize("NFC", key)
            if nk in out:
                raise E1Error("duplicate key after NFC")
            out[nk] = normalize(value[key])
        return out
    if value is None or isinstance(value, (bool, int)):
        return value
    raise E1Error("unsupported canonical value")


def canonical(value) -> bytes:
    return (json.dumps(normalize(value), ensure_ascii=False, sort_keys=True, separators=(",", ":")) + "\n").encode("utf-8")


def validate_truth_record(record):
    required = {
        "content_sha256", "generator_id", "generator_parameters", "model_class",
        "schema_id", "scientific_control_identity_sha256", "scientific_rng_identity", "token_count",
    }
    if set(record) != required:
        raise E1Error("truth record field closure")
    if record["schema_id"] != "g1v2.blind-truth-record.e1":
        raise E1Error("truth record schema")
    for name in ("content_sha256", "scientific_control_identity_sha256"):
        value = record[name]
        if not isinstance(value, str) or len(value) != 64 or any(c not in "0123456789abcdef" for c in value):
            raise E1Error(name)
    if not isinstance(record["token_count"], int) or record["token_count"] <= 0:
        raise E1Error("token_count")
    scientific_payload = {
        "content_sha256": record["content_sha256"],
        "generator_id": record["generator_id"],
        "generator_parameters": record["generator_parameters"],
        "model_class": record["model_class"],
        "scientific_rng_identity": record["scientific_rng_identity"],
        "token_count": record["token_count"],
    }
    if scientific_control_identity(scientific_payload) != record["scientific_control_identity_sha256"]:
        raise E1Error("scientific control identity binding")
    return record


def scientific_control_identity(payload) -> str:
    return hashlib.sha256(SCIENCE_DOMAIN + canonical(payload)).hexdigest()


def blind_id(key: bytes, truth_record) -> tuple[str, str, str]:
    if len(key) != 32:
        raise E1Error("escrow key must be exactly 32 bytes")
    raw = canonical(validate_truth_record(truth_record))
    message = BLIND_DOMAIN + b"\0" + struct.pack(">Q", len(raw)) + raw
    full = hmac.new(key, message, hashlib.sha256).hexdigest()
    return full[:20], full, raw.hex()


def check_collisions(pairs):
    seen = {}
    for identifier, record in pairs:
        raw = canonical(validate_truth_record(record))
        if identifier in seen and seen[identifier] != raw:
            raise E1Error("BLIND_ID_COLLISION")
        seen[identifier] = raw


def job_id(control_instance_id: str) -> str:
    payload = {
        "candidate_id": "M0-iid-0",
        "contract_version": "G1_V2_EXECUTABLE_CONTRACT_V1_1",
        "control_instance_id": control_instance_id,
        "dependency_job_ids": [],
        "metric_id_or_null": None,
        "replicate_or_null": None,
        "scale_or_null": None,
        "stage": "FIT",
    }
    return "j-" + hashlib.sha256(b"G1V2-JOB\0" + canonical(payload)).hexdigest()[:40]


def main():
    if len(sys.argv) != 2:
        raise SystemExit("usage: blind_id_fixture_generator.py TEST_VECTORS.json")
    vectors = json.loads(open(sys.argv[1], encoding="utf-8").read())
    for case in vectors["cases"]:
        if case["operation"] == "derive":
            got = blind_id(bytes.fromhex(case["key_hex"]), case["truth_record"])
            assert got[0] == case["expected_blind_id"], case["id"]
            assert got[1] == case["expected_hmac_sha256"], case["id"]
            assert job_id(got[0]) == case["expected_fit_job_id"], case["id"]
        elif case["operation"] == "collision":
            try:
                check_collisions([(case["forced_blind_id"], x) for x in case["truth_records"]])
            except E1Error as exc:
                assert str(exc) == case["expected"]
            else:
                raise AssertionError(case["id"])
    print(f"BLIND_ID_E1_FIXTURES=PASS cases={len(vectors['cases'])}")


if __name__ == "__main__":
    main()
