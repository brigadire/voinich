#!/usr/bin/env python3
"""Build the target-blind v2 curation package from the v1 candidate ledger.

This script intentionally reads only the v1 curation files.  It does not
open EVA data, target labels, path graphs, witnesses, generators, or solvers.
"""
from __future__ import annotations

import csv
import hashlib
import json
from pathlib import Path

ROOT = Path(__file__).parent
V1 = ROOT.parent / "f68r2_cross_domain_dictionary_preparation_v1"

ISOREDORE_URL = "https://penelope.uchicago.edu/Thayer/L/Roman/Texts/Isidore/17%2A.html"
DIOSCORIDES_URL = "https://www.dioscorides.org/about/"


def read_tsv(path: Path):
    with path.open(encoding="utf-8", newline="") as fh:
        return list(csv.DictReader(fh, delimiter="\t"))


def write_tsv(name: str, fields: list[str], rows: list[dict[str, object]]):
    with (ROOT / name).open("w", encoding="utf-8", newline="") as fh:
        out = csv.DictWriter(fh, fieldnames=fields, delimiter="\t", lineterminator="\n")
        out.writeheader()
        out.writerows({field: row.get(field, "") for field in fields} for row in rows)


def main():
    candidates = read_tsv(V1 / "BOTANICAL_LEXICON_FULL.tsv")
    controls = read_tsv(V1 / "HISTORICAL_CONTROL_LEXICONS.tsv")
    assert len(candidates) == 77, len(candidates)
    assert len(controls) == 30, len(controls)

    audit_fields = [
        "audit_id", "candidate_id", "historical_form", "canonical_identity",
        "status", "source", "edition", "book_chapter_entry", "url", "locator",
        "language", "text_date", "manuscript_or_edition_date", "evidence_type",
        "normalization_note", "audit_basis", "primary_corpus_inclusion",
    ]
    audits = []
    for row in candidates:
        source = row["source"]
        is_isidore = source.startswith("Isidore")
        audits.append({
            "audit_id": "AUDIT_" + row["lexicon_id"].replace("BOT_", ""),
            "candidate_id": row["lexicon_id"],
            "historical_form": row["historical_form"],
            "canonical_identity": row["canonical_plant_identity"],
            "status": "SOURCE_DESCRIPTION_ONLY",
            "source": source,
            "edition": "Isidore, Etymologiae XVII, ed. W. M. Lindsay (1911)" if is_isidore else "Dioscorides, De materia medica; historical edition not yet line-inspected",
            "book_chapter_entry": "XVII; exact chapter/entry not established" if is_isidore else "exact book/chapter/entry not established",
            "url": ISOREDORE_URL if is_isidore else DIOSCORIDES_URL,
            "locator": "NO_LINE_LOCATOR_AVAILABLE_IN_V1",
            "language": row["language"],
            "text_date": "Isidore: 7th c.; Dioscorides: 1st c. CE tradition",
            "manuscript_or_edition_date": "1911 edition planned; manuscript witness not established",
            "evidence_type": "source-description only; no attested line consulted",
            "normalization_note": "No normalization admitted: source form has not been line-verified.",
            "audit_basis": "v1 candidate ledger only; overview page cannot establish a dictionary line",
            "primary_corpus_inclusion": "NO",
        })
    write_tsv("BOTANICAL_LINE_AUDIT.tsv", audit_fields, audits)

    verified_fields = [
        "lexicon_id", "historical_form", "canonical_identity", "source", "edition",
        "book_chapter_entry", "url", "locator", "language", "text_date",
        "manuscript_or_edition_date", "evidence_type", "normalization_note",
    ]
    write_tsv("BOTANICAL_LEXICON_VERIFIED.tsv", verified_fields, [])

    rejected_fields = ["rejected_id", "candidate_id", "historical_form", "canonical_identity", "status", "reason", "source", "audit_locator"]
    write_tsv("BOTANICAL_REJECTED_V2.tsv", rejected_fields, [
        {"rejected_id": "REJ_V2_" + a["candidate_id"].replace("BOT_", ""), "candidate_id": a["candidate_id"],
         "historical_form": a["historical_form"], "canonical_identity": a["canonical_identity"],
         "status": "SOURCE_DESCRIPTION_ONLY", "reason": "No line-level attestation or locator in the selected source edition; excluded from primary corpus.",
         "source": a["source"], "audit_locator": a["locator"]} for a in audits
    ])

    source_fields = ["source_id", "source", "editor_or_edition", "text_date", "used_witness_or_edition_date", "language", "url", "scope", "status", "selection_rule"]
    write_tsv("BOTANICAL_SOURCE_REGISTRY.tsv", source_fields, [
        {"source_id": "SRC_ISIDORE_XVII", "source": "Etymologiae XVII", "editor_or_edition": "W. M. Lindsay, Oxford Classical Text (1911)", "text_date": "7th c.", "used_witness_or_edition_date": "1911 edition; witness not established", "language": "Latin", "url": ISOREDORE_URL, "scope": "Book XVII line-level audit", "status": "PLANNED_NOT_LINE_INSPECTED", "selection_rule": "preselected before any target work"},
        {"source_id": "SRC_DIOSCORIDES_WELLMANN", "source": "De materia medica", "editor_or_edition": "Max Wellmann, Berlin (1906–1914), five-book critical edition", "text_date": "1st c. CE", "used_witness_or_edition_date": "1906–1914 edition; manuscript witness not established", "language": "Greek", "url": DIOSCORIDES_URL, "scope": "one historical edition; exact entries pending", "status": "PLANNED_NOT_LINE_INSPECTED", "selection_rule": "preselected before any target work"},
    ])

    control_audit_fields = ["audit_id", "control_id", "historical_form", "canonical_identity", "status", "source", "edition", "book_chapter_entry", "url", "locator", "language", "text_date", "manuscript_or_edition_date", "evidence_type", "normalization_note", "primary_corpus_inclusion"]
    control_audits = []
    for row in controls:
        control_audits.append({
            "audit_id": "HCA_" + row["control_id"].replace("HCTRL_", ""), "control_id": row["control_id"], "historical_form": row["form"], "canonical_identity": "NON_DOMAIN_" + row["form"].upper(), "status": "SOURCE_DESCRIPTION_ONLY", "source": row["source"], "edition": "No line-level edition selected in v1", "book_chapter_entry": "not established", "url": row["source_reference"], "locator": "NO_LINE_LOCATOR_AVAILABLE_IN_V1", "language": "Latin", "text_date": "not established", "manuscript_or_edition_date": "not established", "evidence_type": "source-description only", "normalization_note": "No normalization admitted.", "primary_corpus_inclusion": "NO"
        })
    write_tsv("HISTORICAL_CONTROL_AUDIT.tsv", control_audit_fields, control_audits)
    write_tsv("HISTORICAL_CONTROL_VERIFIED.tsv", ["control_id", "historical_form", "canonical_identity", "source", "edition", "locator", "language", "text_date", "manuscript_or_edition_date", "evidence_type", "normalization_note"], [])

    write_tsv("MATCHED_PANEL_PARAMETERS.json", [], []) if False else None
    params = {
        "status": "NOT_READY_NO_VERIFIED_FORMS",
        "I_available_botanical_canonical_identities": 0,
        "F_available_verified_botanical_forms": 0,
        "COMMON_IDENTITIES": 0,
        "COMMON_FORMS": 0,
        "panel_domains": ["botanical", "astronomy", "historical_non_domain", "pseudo_control"],
        "matching": {"algorithm": "deterministic minimum-cost matching", "seed": 20260928, "tie_breaking": ["fixed seed", "lexicographic order"], "weights_frozen": True},
        "astronomy_historical_baseline": "3/27 retained as historical baseline; matched astronomy search not executed",
        "real_search_executed": False,
    }
    (ROOT / "MATCHED_PANEL_PARAMETERS.json").write_text(json.dumps(params, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    panel_fields = ["panel_id", "domain", "seed", "canonical_identities", "forms", "source_lexicon", "selection_method", "status"]
    write_tsv("PRIMARY_MATCHED_PANELS.tsv", panel_fields, [{"panel_id": "PRIMARY_NONE", "domain": "ALL", "seed": 20260928, "canonical_identities": 0, "forms": 0, "source_lexicon": "verified ledgers", "selection_method": "minimum-cost matching; no eligible rows", "status": "NOT_READY"}])
    write_tsv("SENSITIVITY_PANEL_REGISTRY.tsv", ["panel_id", "domain", "seed", "variation", "status", "selection_before_search"], [])

    balance_fields = ["lexicon", "identities", "forms", "mean_forms_per_identity", "length_summary", "alphabet_size", "repeat_statistic", "word_pattern_complexity", "status"]
    write_tsv("LEXICON_BALANCE_REPORT.tsv", balance_fields, [
        {"lexicon": "botanical", "identities": 0, "forms": 0, "mean_forms_per_identity": "NA", "length_summary": "NA", "alphabet_size": 0, "repeat_statistic": "NA", "word_pattern_complexity": "NA", "status": "NOT_READY"},
        {"lexicon": "astronomy", "identities": 0, "forms": 0, "mean_forms_per_identity": "NA", "length_summary": "not loaded", "alphabet_size": "not loaded", "repeat_statistic": "not loaded", "word_pattern_complexity": "not loaded", "status": "NOT_LOADED_BY_BLINDNESS"},
        {"lexicon": "historical_non_domain", "identities": 0, "forms": 0, "mean_forms_per_identity": "NA", "length_summary": "NA", "alphabet_size": 0, "repeat_statistic": "NA", "word_pattern_complexity": "NA", "status": "NOT_READY"},
        {"lexicon": "pseudo_control", "identities": 0, "forms": 0, "mean_forms_per_identity": "NA", "length_summary": "NA", "alphabet_size": 0, "repeat_statistic": "NA", "word_pattern_complexity": "NA", "status": "NOT_GENERATED_BEFORE_VERIFIED_SOURCE"},
    ])

    pseudo_fields = ["pseudo_id", "seed", "method", "source_lexicon", "identities", "forms", "length_distribution_preserved", "alphabet_inventory_preserved", "repeat_distribution_preserved", "global_alphabet_isomorphism_to_source", "exact_form_overlap_with_real_lexicons", "generated", "status"]
    write_tsv("PSEUDO_CONTROL_REGISTRY.tsv", pseudo_fields, [{"pseudo_id": f"PSEUDO_V2_{i:03d}", "seed": i, "method": "within-word deterministic constrained randomization (planned)", "source_lexicon": "BOTANICAL_LEXICON_VERIFIED.tsv", "identities": 0, "forms": 0, "length_distribution_preserved": "NOT_APPLICABLE", "alphabet_inventory_preserved": "NOT_APPLICABLE", "repeat_distribution_preserved": "NOT_APPLICABLE", "global_alphabet_isomorphism_to_source": "NO", "exact_form_overlap_with_real_lexicons": "0 (no lexicon generated)", "generated": "NO", "status": "FROZEN_SEED_NOT_GENERATED"} for i in range(1, 100)])

    (ROOT / "CAESAR_CONTROL_AMENDMENT.md").write_text("""# Caesar control amendment\n\n`CAESAR_PSEUDO_CONTROLS_STATUS=WITHDRAWN_MODEL_INVARIANT`. The v1 Caesar controls are not reused as a primary null. v2 freezes 99 seeds for future target-blind constrained randomization; no pseudo search or generation is executed while the verified source corpus has zero forms.\n""", encoding="utf-8")
    (ROOT / "PSEUDO_CONTROL_METHOD.md").write_text("""# Pseudo-control method\n\nThe preregistered method is deterministic constrained randomization within each historical form, using a seed-specific reproducible permutation of character positions and rejection of unchanged words. It preserves word lengths, uses the source alphabet inventory, approximately preserves character and repeated-letter aggregates, and destroys order and lexical identity. It is not a global alphabet permutation.\n\nIn this preparation run the verified corpus has zero forms, so all 99 pseudo-lexicons remain frozen-but-not-generated. Therefore the registry reports no overlap or isomorphism for generated lexicons because none exists, and no null search is run.\n""", encoding="utf-8")
    (ROOT / "BLINDNESS_AUDIT_V2.md").write_text("""# Blindness audit v2\n\nPASS: the builder reads only v1 curation ledgers. It does not read EVA tokens, path graphs, generator/solver inputs, reachability, astronomy witnesses, coverage, or future search results. It does not run an EVA search.\n\nSelection is source-ledger-only. The astronomy baseline is recorded only as the historical scalar `3/27`; astronomy forms are not loaded. No panel is selected by coverage.\n""", encoding="utf-8")
    (ROOT / "CROSS_DOMAIN_EXPERIMENT_PROTOCOL_V2.md").write_text("""# Cross-domain experiment protocol v2\n\nPreparation only; real cross-domain search is explicitly not executed.\n\nThe 77 v1 candidates are audited line-by-line in `BOTANICAL_LINE_AUDIT.tsv`; because the v1 evidence contains no concrete line locator, every row is `SOURCE_DESCRIPTION_ONLY` and excluded. `COMMON_IDENTITIES=min(I,94)` and `COMMON_FORMS` are computed from verified rows only. Here both are zero, so no primary panel exists.\n\nWhen a later curation run supplies verified rows, panels must be built by the frozen minimum-cost matching specification in `MATCHED_PANEL_PARAMETERS.json`, with fixed seed and lexicographic tie-breaking. Sensitivity panels may be added only before search. The 3/27 astronomy result remains historical baseline; matched astronomy search is a future staged action.\n\nNull execution is staged: if botanical maximum is below 4, run no nulls; if it is at least 4, run all 99 frozen seeds without changing the replicate count.\n""", encoding="utf-8")

    status = {
        "BOTANICAL_FORMS_VERIFIED": 0, "BOTANICAL_IDENTITIES_VERIFIED": 0,
        "DIOSCORIDES_LINE_LEVEL_SOURCE_READY": False, "ISIDORE_LINE_LEVEL_AUDIT_COMPLETE": False,
        "HISTORICAL_CONTROL_AUDIT_COMPLETE": False, "COMMON_IDENTITIES": 0, "COMMON_FORMS": 0,
        "PRIMARY_MATCHED_PANELS_READY": False, "CAESAR_CONTROLS_WITHDRAWN": "YES",
        "NON_ISOMORPHIC_PSEUDO_CONTROLS_READY": False, "PSEUDO_CONTROL_SEEDS_FROZEN": 99,
        "TARGET_BLIND_PREPARATION": True, "REAL_CROSS_DOMAIN_SEARCH_EXECUTED": False,
        "CROSS_DOMAIN_SEARCH_READY": False, "SCIENTIFIC_CLAIM": "NONE",
        "candidate_rows_audited": len(audits), "historical_control_rows_audited": len(control_audits),
        "v1_read_only": True,
    }
    (ROOT / "PREPARATION_STATUS.json").write_text(json.dumps(status, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")
    (ROOT / "VALIDATION_REPORT.md").write_text("""# Validation report\n\n- 77/77 v1 botanical candidate rows have an audit row: PASS.\n- All 77 are `SOURCE_DESCRIPTION_ONLY`; none enters the verified corpus: PASS (conservative).\n- Isidore and Dioscorides line-level audit completion: NO; v1 supplied no line locators.\n- 30/30 provisional historical-control rows are audited, but none is verified.\n- `COMMON_IDENTITIES=0`, `COMMON_FORMS=0`; no matched panel is search-ready.\n- Caesar controls withdrawn; 99 seeds frozen, pseudo-lexicons not generated because the verified source corpus is empty.\n- EVA/path/generator/solver/coverage inputs were not read; real search was not executed.\n- Scientific claim: NONE.\n""", encoding="utf-8")

    sums = []
    for path in sorted(ROOT.iterdir()):
        if path.name in {"SHA256SUMS", "build_v2.py"} or not path.is_file():
            continue
        digest = hashlib.sha256(path.read_bytes()).hexdigest()
        sums.append(f"{digest}  {path.name}")
    (ROOT / "SHA256SUMS").write_text("\n".join(sums) + "\n", encoding="utf-8")


if __name__ == "__main__":
    main()
