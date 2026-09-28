#!/usr/bin/env python3
import csv, hashlib, json, sys, time
from pathlib import Path

ROOT = Path(__file__).resolve().parent
PREP = ROOT.parent / "exploratory_astronomical_dictionary_search_v1"
AUDIT = ROOT.parent / "exploratory_astronomical_dictionary_search_e2_structural_support_audit_v1"
sys.path.insert(0, str(AUDIT))
from audit import alignments
sys.path.insert(0, str(AUDIT / "snapshot"))
from scorer import encode_word

def rows(path):
    with path.open() as f:
        return list(csv.DictReader(f, delimiter="\t"))

def rule_text(mapping):
    return ";".join(f"{s}->{t}" for s, t in mapping)

def main():
    started = time.monotonic()
    labels = rows(PREP / "TARGET_STAR_LABELS.tsv")
    lexicon = rows(PREP / "ASTRONOMICAL_NAMES_LEXICON.tsv")
    cache = {}
    paths = []
    for label in labels:
        for lex in lexicon:
            key = (lex["normalized_form"], label["zl3b_token"])
            if key not in cache:
                cache[key] = alignments(*key, 5)
            for mapping in cache[key]:
                table = dict(mapping)
                encoded = encode_word(lex["normalized_form"], table, "DROP_UNMAPPED", "NONE")
                if encoded != label["zl3b_token"]:
                    continue
                paths.append({
                    "label_id": label["label_id"], "page": label["page"],
                    "token": label["zl3b_token"], "identity": lex["canonical_identity_id"],
                    "lexicon_id": lex["lexicon_id"], "normalized_form": lex["normalized_form"],
                    "rules": rule_text(mapping), "encoded": encoded,
                })
    fields = ["label_id", "page", "token", "identity", "lexicon_id", "normalized_form", "rules", "encoded"]
    with (ROOT / "PATH_GRAPH.tsv").open("w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=fields, delimiter="\t")
        w.writeheader(); w.writerows(paths)
    summary = {
        "status": "SCORER_CONSISTENT_GRAPH_BUILT",
        "path_count": len(paths),
        "label_count": len(labels),
        "lexicon_row_count": len(lexicon),
        "canonical_identity_count": len({x["canonical_identity_id"] for x in lexicon}),
        "exact_scorer": "encode_word(DROP_UNMAPPED,NONE)",
        "elapsed_sec": round(time.monotonic() - started, 4),
    }
    (ROOT / "GRAPH_BUILD_SUMMARY.json").write_text(json.dumps(summary, indent=2, sort_keys=True) + "\n")
    print(json.dumps(summary, sort_keys=True))

if __name__ == "__main__":
    main()
