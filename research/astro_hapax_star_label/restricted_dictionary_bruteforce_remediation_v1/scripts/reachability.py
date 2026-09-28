#!/usr/bin/env python3
"""
Reachability remediation (clean_room.md Section 18). Reuses v3_design's analyze_reachability.py
methodology verbatim for the base 120-model-class sweep (COMPONENT_REUSE_REGISTRY.tsv:
REACHABILITY_CODE, REUSABLE_AFTER_REVALIDATION -- confirmed byte-identical reproduction by
v3_audit), then DECOMPOSES the single "structural_coverage_upper_bound" figure the audit found
mis-cited (F002) into 6 named sub-metrics, computed against the REMEDIATED lexicon (this
package's own output, not v3's), so "reachability" cannot again silently mean "lengths happen to
overlap" -- clean_room.md's explicit warning against treating length match as full reachability.

Definitions:
  LENGTH_REACHABILITY: fraction of real-label length-histogram weight for which the aggregate
    length transform can plausibly produce a matching output length (v3's original computation,
    reused unchanged).
  ALPHABET_REACHABILITY: whether transformed output can be guaranteed to use ONLY the 16-symbol
    EVA target alphabet (KEEP mode structurally cannot for non-EVA source letters; DROP/SELECTIVE
    modes can). 1.0 if guaranteed, else the v2-established 0.25 ceiling for KEEP mode.
  UNIQUE_SYMBOL_REACHABILITY: fraction of real-label distinct-symbol-count histogram weight
    reachable given table_size's cap on distinct output symbols (v3's original computation,
    reused unchanged).
  LEXICON_CONDITIONED_REACHABILITY: same length-reachability QUESTION but asked against the
    REMEDIATED lexicon's ACTUAL attested word lengths (not just the aggregate manifest
    histogram) -- a table_size that is length-reachable in the abstract may still not match how
    long real historical star names actually are.
  GLOBAL_TABLE_REACHABILITY: 1 - collision_rate, computed by actually transforming every live
    remediated-lexicon word through ONE shared representative table (not per-word cherry-picked
    tables) -- this is the metric that penalizes a model class for collapsing distinct historical
    names onto the same short output under a single global system, which is what an actual
    dictionary-substitution hypothesis requires.
  DELETION_ABBREVIATION_REACHABILITY: this model class's structural_coverage_upper_bound,
    normalized against the best upper bound achievable at the SAME table_size/mapping_type by
    varying only deletion_mode/abbreviation -- isolates how much of the ceiling is attributable to
    the deletion/abbreviation choice specifically, holding table size and mapping mode fixed.
  PRIMARY_GLOBAL_PATH_REACHABILITY: min(ALPHABET_REACHABILITY, LEXICON_CONDITIONED_REACHABILITY,
    UNIQUE_SYMBOL_REACHABILITY, GLOBAL_TABLE_REACHABILITY) -- the headline number. It requires
    ALL FOUR conditions to hold SIMULTANEOUSLY under one shared table, which is what "historical
    attestation -> one global system -> LABEL structure" (the task's own phrasing) actually means;
    it is never allowed to equal LENGTH_REACHABILITY alone.
"""
import csv
import json
import sys
from collections import Counter
from pathlib import Path

BASE_DIR = Path(__file__).resolve().parent.parent
MANIFEST_PATH = BASE_DIR.parent / "restricted_dictionary_bruteforce_v3_design" / "REAL_SCOPE_STRUCTURAL_MANIFEST.json"
REMEDIATED_LEXICON_PATH = BASE_DIR / "REMEDIATED_HISTORICAL_STAR_LEXICON.tsv"

EVA_ALPHABET = set("acdefhiklnoprsty")
NON_EVA_SOURCE = set("bgjuvwxz")


def load_manifest():
    with MANIFEST_PATH.open() as f:
        return json.load(f)


def load_remediated_lexicon():
    with REMEDIATED_LEXICON_PATH.open() as f:
        rows = list(csv.DictReader(f, delimiter="\t"))
    words = set()
    for r in rows:
        if r.get("disposition") in ("KEEP", "CORRECT", "MERGE_IDENTITY", "RECLASSIFY_RECONSTRUCTION"):
            w = r["normalized_form"].replace(" ", "")
            if w:
                words.add(w)
    return sorted(words)


def length_reachability_generic(words, length_weights, total_target, u_mode, abbr, t_size):
    reachable = 0
    for l_val, count in length_weights.items():
        if u_mode == "KEEP":
            can_produce = any(
                (len(w) if abbr == "NONE" else
                 max(3, len(w) - 1) if abbr == "SUSPENSION_1" else
                 max(3, len(w) - 2) if abbr == "SUSPENSION_2" else
                 min(4, len(w))) == l_val
                for w in words
            )
        elif u_mode == "DROP_UNMAPPED":
            max_len = t_size if abbr != "PREFIX_4" else min(4, t_size)
            can_produce = l_val <= max_len
        else:  # SELECTIVE_VOWEL_DROP
            can_produce = l_val <= 10
        if can_produce:
            reachable += count
    return reachable / total_target


def lexicon_conditioned_reachability(words, u_mode, abbr, t_size):
    """Same question, asked against the REAL remediated lexicon's actual word-length distribution
    instead of the manifest's label-length histogram."""
    if not words:
        return 0.0
    word_lengths = Counter(len(w) for w in words)
    total = sum(word_lengths.values())
    reachable = 0
    for wlen, count in word_lengths.items():
        if u_mode == "KEEP":
            out_len = (wlen if abbr == "NONE" else
                       max(3, wlen - 1) if abbr == "SUSPENSION_1" else
                       max(3, wlen - 2) if abbr == "SUSPENSION_2" else
                       min(4, wlen))
        elif u_mode == "DROP_UNMAPPED":
            out_len = min(wlen, t_size if abbr != "PREFIX_4" else min(4, t_size))
        else:
            out_len = wlen  # selective drop keeps consonants; upper-bounded by wlen
        if 1 <= out_len <= 10:
            reachable += count
    return reachable / total


def global_table_reachability(words, u_mode, abbr, t_size):
    if not words:
        return 1.0
    rep_source = list("adeilnorstmchbgpqu")[:max(t_size, 1)]
    rep_table = {c: (c if c in EVA_ALPHABET else "a") for c in rep_source}
    outputs = []
    for w in words:
        if u_mode == "DROP_UNMAPPED":
            out = "".join(rep_table[c] for c in w if c in rep_table)
        elif u_mode == "SELECTIVE_VOWEL_DROP":
            out = "".join(rep_table[c] if c in rep_table else (c if c not in "aeiouy" else "") for c in w)
        else:
            out = "".join(rep_table.get(c, c) for c in w)
        if abbr == "SUSPENSION_1" and len(out) > 3:
            out = out[:-1]
        elif abbr == "SUSPENSION_2" and len(out) > 4:
            out = out[:-2]
        elif abbr == "PREFIX_4":
            out = out[:4]
        if out:
            outputs.append(out)
    if not outputs:
        return 0.0
    counts = Counter(outputs)
    collisions = sum(v - 1 for v in counts.values() if v > 1)
    collision_rate = collisions / len(outputs)
    return 1.0 - collision_rate


def main():
    out_path = sys.argv[1] if len(sys.argv) > 1 else "REACHABILITY_REMEDIATION.tsv"
    manifest = load_manifest()
    words = load_remediated_lexicon()

    length_weights = {int(k): v for k, v in manifest["length_distribution"]["histogram"].items()}
    total_target = manifest["occurrences_total"]
    distinct_symbols_hist = {int(k): v for k, v in manifest["distinct_symbols_distribution"]["histogram"].items()}

    table_sizes = [4, 6, 8, 10, 12]
    mapping_types = ["INJECTIVE", "MERGE_1"]
    unmapped_modes = ["KEEP", "DROP_UNMAPPED", "SELECTIVE_VOWEL_DROP"]
    abbreviations = ["NONE", "SUSPENSION_1", "SUSPENSION_2", "PREFIX_4"]

    rows = []
    ceiling_by_table_mapping = {}

    for t_size in table_sizes:
        for m_type in mapping_types:
            best_ceiling = 0.0
            cell_rows = []
            for u_mode in unmapped_modes:
                for abbr in abbreviations:
                    alphabet_r = 1.0 if u_mode != "KEEP" else 0.25
                    length_r = length_reachability_generic(words, length_weights, total_target, u_mode, abbr, t_size)
                    effective_max_symbols = t_size if m_type == "INJECTIVE" else (t_size - 1)
                    symbol_r = sum(c for s, c in distinct_symbols_hist.items() if (u_mode == "KEEP" or s <= effective_max_symbols)) / total_target
                    lex_cond_r = lexicon_conditioned_reachability(words, u_mode, abbr, t_size)
                    global_r = global_table_reachability(words, u_mode, abbr, t_size)
                    ceiling = min(alphabet_r, length_r, symbol_r)
                    best_ceiling = max(best_ceiling, ceiling)
                    cell_rows.append(dict(
                        table_size=t_size, mapping_type=m_type, unmapped_mode=u_mode, abbreviation_rule=abbr,
                        alphabet_reachability=round(alphabet_r, 4), length_reachability=round(length_r, 4),
                        unique_symbol_reachability=round(symbol_r, 4), lexicon_conditioned_reachability=round(lex_cond_r, 4),
                        global_table_reachability=round(global_r, 4), _ceiling=ceiling,
                    ))
            for r in cell_rows:
                r["deletion_abbreviation_reachability"] = round(r["_ceiling"] / best_ceiling, 4) if best_ceiling > 0 else 0.0
                r["primary_global_path_reachability"] = round(
                    min(r["alphabet_reachability"], r["lexicon_conditioned_reachability"], r["unique_symbol_reachability"], r["global_table_reachability"]), 4
                )
                del r["_ceiling"]
                rows.append(r)

    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(rows[0].keys()), delimiter="\t")
        w.writeheader()
        w.writerows(rows)

    best_row = max(rows, key=lambda r: r["primary_global_path_reachability"])
    print(f"wrote {len(rows)} rows to {out_path}")
    print(f"best PRIMARY_GLOBAL_PATH_REACHABILITY = {best_row['primary_global_path_reachability']} "
          f"at table_size={best_row['table_size']} {best_row['mapping_type']} {best_row['unmapped_mode']} {best_row['abbreviation_rule']}")


if __name__ == "__main__":
    main()
