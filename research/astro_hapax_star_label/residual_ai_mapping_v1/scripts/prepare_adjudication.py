#!/usr/bin/env python3
"""Build the answer-free adjudication package after independent passes finish."""

from __future__ import annotations

import csv
import hashlib
import json
import shutil
from pathlib import Path


ROOT = Path(__file__).resolve().parents[4]
OUT = ROOT / "research/astro_hapax_star_label/residual_ai_mapping_v1"
SOURCE = OUT / "cleanroom/visual_a"
DEST = OUT / "cleanroom/adjudicator_d"
AGENTS = OUT / "agent_outputs"


def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def jsonl(path: Path) -> list[dict]:
    result = []
    for line_no, line in enumerate(path.read_text(encoding="utf-8").splitlines(), 1):
        try:
            result.append(json.loads(line))
        except json.JSONDecodeError as exc:
            raise RuntimeError(f"{path}:{line_no}: invalid JSON: {exc}") from exc
    return result


def ids_from_cases() -> list[str]:
    with (SOURCE / "CASES.tsv").open(encoding="utf-8", newline="") as f:
        return [r["neutral_label_id"] for r in csv.DictReader(f, delimiter="\t")]


def main() -> None:
    required = {
        "VISUAL_A_RESULTS.jsonl": AGENTS / "AI_VISUAL_A_RESULTS.jsonl",
        "ALIGNMENT_B_RESULTS.jsonl": AGENTS / "AI_ALIGNMENT_B_RESULTS.jsonl",
        "VISUAL_C_RESULTS.jsonl": AGENTS / "AI_VISUAL_C_RESULTS.jsonl",
        "ALIGNMENT_B_SHUFFLED_RESULTS.jsonl": AGENTS / "AI_ALIGNMENT_B_SHUFFLED_RESULTS.jsonl",
    }
    expected = ids_from_cases()
    for display, path in required.items():
        if not path.exists():
            raise RuntimeError(f"missing independent-pass output: {path}")
        got = [r.get("neutral_label_id") for r in jsonl(path)]
        if got != expected:
            raise RuntimeError(f"{path}: IDs/order do not match CASES.tsv")
    if DEST.exists():
        shutil.rmtree(DEST)
    (DEST / "crops").mkdir(parents=True)
    (DEST / "pages").mkdir(parents=True)
    for path in (SOURCE / "crops").iterdir():
        shutil.copyfile(path, DEST / "crops" / path.name)
    for path in (SOURCE / "pages").iterdir():
        shutil.copyfile(path, DEST / "pages" / path.name)
    for name in ("CASES.tsv", "CONTACT_SHEET.jpg"):
        shutil.copyfile(SOURCE / name, DEST / name)
    shutil.copyfile(OUT / "cleanroom/alignment_b_base/CANDIDATES.tsv", DEST / "CANDIDATES.tsv")
    for display, path in required.items():
        shutil.copyfile(path, DEST / display)

    input_files = sorted(p for p in DEST.rglob("*") if p.is_file())
    h = hashlib.sha256()
    manifest_rows = []
    for path in input_files:
        rel = path.relative_to(DEST).as_posix()
        digest = sha(path)
        h.update(f"{digest}  {rel}\n".encode())
        manifest_rows.append((digest, rel))
    package_hash = h.hexdigest()
    with (DEST / "INPUT_HASHES.tsv").open("w", encoding="utf-8", newline="") as f:
        f.write("sha256\trelative_path\n")
        for digest, rel in manifest_rows:
            f.write(f"{digest}\t{rel}\n")

    prompt = f"""# AI_ADJUDICATOR_D blind adjudication\n\nRead only files in this directory. Do not inspect any parent directory, repository file, sealed answer, prior mapping, or excluded metadata. This is an instruction-enforced clean room. The adjudication input package hash is `{package_hash}`.\n\nFor each CASES.tsv row, inspect the crop/context and compare VISUAL_A_RESULTS.jsonl, VISUAL_C_RESULTS.jsonl, ALIGNMENT_B_RESULTS.jsonl, ALIGNMENT_B_SHUFFLED_RESULTS.jsonl, and the complete CANDIDATES.tsv. You have no correct answers. Do not infer meaning. Do not repair annotator readings silently. Preserve disagreements and abstain whenever evidence is inadequate. A display-order change is evidence: top-1 disagreement or changed abstention is UNSTABLE; same top-1 but changed top-k/confidence/material rationale is ORDER_SENSITIVE; otherwise ORDER_STABLE.\n\nCONSENSUS_EXACT requires compatible A/C visual readings, an existing exact candidate supported by Alignment B, no material glyph or boundary conflict, and ORDER_STABLE. CONSENSUS_WITH_DOCUMENTED_NORMALIZATION permits only the fixed display expansion C=cth, K=ckh, P=cph, F=cfh, N=iin, A=ain, H=ch, S=sh, E=ee, I=in. Otherwise use PROBABLE, AMBIGUOUS, NO_MATCH, UNREADABLE, or RING_CYCLIC_SEQUENCE. Never replace an abstention merely to improve accuracy.\n\nWrite exactly one compact JSON object per line in CASES.tsv order with keys: neutral_label_id, outcome, selected_candidate_ids, raw_tokens, alternative_candidate_ids, visual_a_c_agreement, alignment_b_rank, glyph_level_discrepancies, confidence, stability, forced_answer, rationale, ai_pass_provenance, cleanroom_package_hash. Arrays: selected_candidate_ids, raw_tokens, alternative_candidate_ids, glyph_level_discrepancies, ai_pass_provenance. confidence is HIGH/MEDIUM/LOW. stability is ORDER_STABLE/ORDER_SENSITIVE/UNSTABLE/NOT_APPLICABLE. forced_answer must be false. For no selection, selected_candidate_ids and raw_tokens are empty arrays. Use package hash `{package_hash}` exactly.\n"""
    (DEST / "PROMPT.md").write_text(prompt, encoding="utf-8")
    state = {
        "status": "ADJUDICATION_PACKAGE_READY",
        "input_package_hash": package_hash,
        "records_per_pass": len(expected),
        "hidden_answers_included": False,
    }
    (OUT / "ADJUDICATION_PRE_RUN_STATE.json").write_text(json.dumps(state, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    print(json.dumps(state, sort_keys=True))


if __name__ == "__main__":
    main()
