#!/usr/bin/env python3
"""Build the frozen, non-analytic f68r2 synthesis package."""
from __future__ import annotations

import hashlib
import json
from datetime import date
from pathlib import Path

ROOT = Path(__file__).resolve().parents[3]
OUT = Path(__file__).resolve().parent

INPUTS = [
    "research/stolfi_label_hapax_enrichment/STOLFI_ASTRO_LABEL_HAPAX_REPORT.md",
    "research/stolfi_label_hapax_enrichment/STOLFI_ASTRO_LABEL_HAPAX_ENRICHMENT.tsv",
    "research/stolfi_label_hapax_enrichment/STOLFI_ASTRO_LABEL_HAPAX_MANIFEST.json",
    "research/astro_hapax_star_label/restricted_hapax_enrichment_v1/REPORT.md",
    "research/astro_hapax_star_label/restricted_hapax_enrichment_v1/BLOCK_AWARE_REPORT.md",
    "research/astro_hapax_star_label/f68r2_cross_domain_search_v1/CROSS_DOMAIN_RESULTS_REPORT.md",
    "research/astro_hapax_star_label/f68r2_cross_domain_search_v1/EMPIRICAL_STATISTICS.tsv",
    "research/astro_hapax_star_label/f68r2_cross_domain_search_v1/INPUT_FREEZE.json",
    "research/astro_hapax_star_label/f68r2_cross_domain_search_v1/REAL_PANEL_MAXIMA.tsv",
    "research/astro_hapax_star_label/f68r2_cross_domain_search_v1/RUN_STATUS.json",
    "research/astro_hapax_star_label/f68r2_botanical_identity_remediation_v1/RUN_STATUS.json",
    "research/astro_hapax_star_label/f68r2_botanical_identity_remediation_v1/IDENTITY_SEMANTICS.md",
]

def sha(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as f:
        for block in iter(lambda: f.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()

def write(name: str, text: str) -> None:
    (OUT / name).write_text(text.rstrip() + "\n", encoding="utf-8")

def main() -> None:
    missing = [p for p in INPUTS if not (ROOT / p).is_file()]
    if missing:
        raise SystemExit("missing input: " + ", ".join(missing))
    manifest = {
        "package": "f68r2_cross_domain_synthesis_v1",
        "created": str(date.today()),
        "purpose": "frozen synthesis; no new search or statistical computation",
        "inputs": {p: sha(ROOT / p) for p in INPUTS},
        "read_only_upstream": True,
        "real_cross_domain_search_executed_here": False,
    }
    (OUT / "INPUT_MANIFEST.json").write_text(json.dumps(manifest, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    write("RESULTS_LEDGER.tsv", """result_id\tclass\tstatus\tquantitative_summary\tinterpretation_scope\tsource\nHAPAX_POSITIVE\tstructural_label_enrichment\tPOSITIVE_STRUCTURAL_RESULT\t112 confirmed labels; 80 hapax; fraction 0.714285714 vs null 0.611207143; ratio 1.168647524; one-sided p=0.008799120; LOPO 8/8 positive and p<0.05\tIndependently identified astronomical LABEL occurrences are enriched for section-local hapax; not all hapaxes are labels and no semantic/decryption claim follows\tresearch/stolfi_label_hapax_enrichment/STOLFI_ASTRO_LABEL_HAPAX_REPORT.md\nHAPAX_RESTRICTED_ROBUSTNESS\tstructural_label_enrichment\tPOSITIVE_WITH_SCOPE_LIMITS\tSTAR 46/54=0.851851852 vs null 0.690868519; p=0.002099790; circular-text comparison p=0.443907\tSupports an enrichment signal against specified astronomical/background controls, not a universal label rule; small/non-independent block limits remain\tresearch/astro_hapax_star_label/restricted_hapax_enrichment_v1/REPORT.md\nDICTIONARY_CROSS_DOMAIN\tcorrected_global_form_matching\tNEGATIVE_WITHIN_TESTED_MODEL_CLASS\tbotanical 0/27, astronomical 0/27, historical control 3/27; pseudo maxima 0 for 72, 2 for 11, 3 for 16 of 99; botanical/astronomical p=1.00\tNo domain-specific signal for botanical or astronomical panels within the preregistered matched global-form model; compatible with generic form matching\tresearch/astro_hapax_star_label/f68r2_cross_domain_search_v1/CROSS_DOMAIN_RESULTS_REPORT.md\nOLD_ASTRONOMY_3_27\tlegacy_dictionary_search\tWITHDRAWN_FROM_INFERENCE\t3/27\tHistorical baseline only; not comparable because it used a different dictionary size/panel and was not rerun as the matched astronomy panel\tresearch/astro_hapax_star_label/f68r2_cross_domain_search_v1/CROSS_DOMAIN_RESULTS_REPORT.md\nCAESAR_CONTROLS\tpseudo_null\tWITHDRAWN_MODEL_INVARIANT\tCaesar-shift controls withdrawn\tNot a valid primary null for the final experiment; replaced by 99 non-isomorphic target-blind pseudo-lexicons\tresearch/astro_hapax_star_label/f68r2_cross_domain_search_v1/INPUT_FREEZE.json\nPROVISIONAL_BOTANICAL_CANDIDATES\tsource_curation\tSUPERSEDED_NOT_VERIFIED\t77 earlier candidates and earlier 30 controls lacked line-level proof\tNot evidence and not part of the final verified comparison\tresearch/astro_hapax_star_label/f68r2_cross_domain_source_extraction_v1\nFORM_IDENTITY_EQUALITY\tidentity_semantics\tSUPERSEDED\t134 forms=134 identities was rejected; remediation retained 89 source concepts from 134 forms\tSource-level identity is not the same as modern species identity and is not one-form-one-identity by default\tresearch/astro_hapax_star_label/f68r2_botanical_identity_remediation_v1/RUN_STATUS.json\nPAGE_THEME\tglobal_interpretation\tNOT_DETERMINED\tNo synthesis result identifies or excludes a page theme\tThe tested negative model class cannot refute botanical, astronomical, or another thematic interpretation in general\tthis package\n""")

    write("WITHDRAWN_RESULTS.tsv", """legacy_item\tstatus\treason\treplacement\nAstronomy 3/27\tWITHDRAWN_FOR_CROSS_DOMAIN_INFERENCE\tdifferent dictionary size and unmatched design\tfresh matched astronomical panel; result 0/27\nCaesar pseudo-controls\tWITHDRAWN_MODEL_INVARIANT\tone global alphabet shift is an inappropriate primary null\t99 target-blind non-isomorphic pseudo-lexicons\n77 SOURCE_DESCRIPTION_ONLY botanical candidates\tNOT_VERIFIED\tno exact line-level attestation\tsource-led Isidore audit and remediation\n30 provisional historical controls\tNOT_VERIFIED\tno exact line-level attestation\tline-level historical-control audit\n134-forms-equals-134-identities\tREJECTED_IDENTITY_RULE\tform identity was artificially unique\t89 source-level concepts; unresolved concepts excluded where applicable\nAny f68r1/group-crosswalk carryover\tOUT_OF_SCOPE\tnot an input to the f68r2 matched search\tfrozen f68r2 panels and inputs only\n""")

    write("INTERPRETATION_BOUNDARIES.md", """# Interpretation boundaries

The synthesis fixes two different results at different inferential levels.

1. **Positive structural result.** Independently matched astronomical LABEL occurrences are enriched for section-local hapax. The primary frozen test used 112 occurrences, 80 hapaxes, a 0.714285714 observed fraction against a 0.611207143 panel-conditioned null mean, ratio 1.168647524, and one-sided permutation p=0.008799120. Leave-one-panel-out exclusions retained positive direction and p<0.05 in 8/8 cases. This is a property of the tested token/label inventory and null, not a semantic decoding result.
2. **Negative dictionary result.** Under the corrected global mapping model, fresh matched astronomical and botanical panels each reached 0/27, while the historical control reached 3/27. All 99 pseudo-controls were run with maxima over the same 64 profiles; the botanical and astronomical empirical p-values were 1.00. Thus no domain-specific signal was found in this tested model class.

The two conclusions are not contradictory: hapax enrichment concerns token frequency conditional on an independently frozen label inventory; dictionary search tests a particular global form-matching compatibility model. Neither result establishes the subject matter of the pages. The dictionary result does not disprove botanical, astronomical, or any other theme, and the hapax result does not identify meanings.

The search package did not fully enumerate all co-optimal assignments; retained assignments are profile witnesses. That limits identity-level interpretation even though all real maxima were certified. The hapax result inherits incomplete label-inventory/complement limitations documented by its source reports.
""")

    write("NEXT_MODEL_FAMILIES.md", """# Principally different models that could be tested later

The current branch is closed. A future experiment must have independent historical or linguistic justification before any run is authorized; it must not be selected because it improves coverage.

## Abbreviational model

Test historically documented abbreviation conventions, including positional or scribal contexts, with an explicit expansion inventory and held-out validation. The model should penalize arbitrary expansion choices and use non-target historical controls with the same abbreviation mechanisms.

## Mnemonic or notational model

Test whether labels encode a documented mnemonic system, such as a fixed catalogue, memory aid, or diagrammatic notation. The candidate source tradition and mapping rules must be fixed before inspecting coverage, with a held-out corpus and a complexity penalty. Free-form semantic association is not sufficient evidence.

## Morpheme-compositional model

Test reusable subunits, concatenative or templatic composition, and context-conditioned morpheme meanings only when independently supported by a source tradition. Freeze the morpheme inventory, segmentation rules, composition grammar, and null model in advance; validate on held-out forms and compare against length/alphabet/repetition-preserving controls.

These are different model classes, not extensions of the failed global substitution. They should preserve blind execution, equal budgets, preregistration, scorer replay, and explicit UNKNOWN handling. No next model is authorized by this package.
""")

    write("BRANCH_CLOSURE.md", """# Branch closure

`f68r2_cross_domain_synthesis_v1` closes the current cross-domain dictionary branch.

- `HAPAX_ENRICHMENT_STATUS=POSITIVE_STRUCTURAL_RESULT`
- `DICTIONARY_SEARCH_STATUS=NEGATIVE_WITHIN_TESTED_MODEL_CLASS`
- `PAGE_THEMATIC_INTERPRETATION=NOT_RESOLVED`
- `NEW_SEARCH_EXECUTED=NO`
- `SCIENTIFIC_CLAIM=NONE`

No further global-substitution search is implied. Any future work requires a new, independently justified model family and a new preregistered branch.
""")

    write("SYNTHESIS_REPORT.md", """# f68r2 cross-domain synthesis v1

## Executive conclusion

The current evidence has two separate outcomes:

- **Hapax enrichment: positive structural result.** The frozen Stolfi-label analysis detected enrichment of independently matched astronomical LABEL occurrences among section-local hapax: 80/112 = 0.714285714 versus null mean 0.611207143, ratio 1.168647524, one-sided permutation p=0.008799120, with 8/8 leave-one-panel-out exclusions retaining positive direction and p<0.05.
- **Dictionary search: negative within the tested class.** The fresh matched f68r2 search certified 0/27 for both astronomical and botanical panels, versus 3/27 for historical control; the 99 model-selection-aware pseudo-controls gave botanical and astronomical p=1.00. The preregistered interpretation is `RESULT_COMPATIBLE_WITH_GENERIC_FORM_MATCHING` and `DOMAIN_SPECIFIC_SIGNAL=NONE`.

The old astronomy `3/27` is withdrawn as a cross-domain claim because it came from a different dictionary size/design. Caesar shifts are withdrawn as the primary null. Earlier unverified botanical/control rows and the one-form-one-identity construction are superseded by line-level source auditing and source-level identity remediation.

## What is and is not established

Established: a reproducible hapax enrichment under the specified label inventory and null; and a null/negative result for the specified corrected-global form-matching model class.

Not established: a decipherment, word identification, semantic interpretation, page theme, or evidence that pages are not botanical/astronomical. The negative result is bounded by the 27 occurrences, 64 frozen profiles, 30-form matched panels, corrected global semantics, and the declared controls. The hapax result is not a semantic label detector and does not license treating every hapax as a label.

## Closure and future scope

This branch is closed without a thematic verdict. The only principled next models are materially different and independently justified systems: historically grounded abbreviation, mnemonic/notational, or morpheme-compositional models. Their inventories, rules, controls, complexity penalties, held-out tests, and UNKNOWN procedures must be frozen before results are inspected. See `NEXT_MODEL_FAMILIES.md`.
""")

    status = {
        "HAPAX_ENRICHMENT_STATUS": "POSITIVE_STRUCTURAL_RESULT",
        "DICTIONARY_SEARCH_STATUS": "NEGATIVE_WITHIN_TESTED_MODEL_CLASS",
        "DICTIONARY_INTERPRETATION": "RESULT_COMPATIBLE_WITH_GENERIC_FORM_MATCHING",
        "OLD_ASTRONOMY_3_27": "WITHDRAWN_FROM_CROSS_DOMAIN_INFERENCE",
        "CAESAR_CONTROLS": "WITHDRAWN_MODEL_INVARIANT",
        "BRANCH_CLOSED": "YES",
        "PAGE_THEME_DISPROVED": "NO",
        "NEW_SEARCH_EXECUTED": "NO",
        "SCIENTIFIC_CLAIM": "NONE",
        "SOURCE_MANIFEST": "INPUT_MANIFEST.json",
    }
    (OUT / "RUN_STATUS.json").write_text(json.dumps(status, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")

    excluded = {"SHA256SUMS", "build_synthesis.py"}
    lines = []
    for p in sorted(OUT.iterdir()):
        if p.is_file() and p.name not in excluded:
            lines.append(f"{sha(p)}  {p.name}")
    (OUT / "SHA256SUMS").write_text("\n".join(lines) + "\n", encoding="utf-8")

if __name__ == "__main__":
    main()
