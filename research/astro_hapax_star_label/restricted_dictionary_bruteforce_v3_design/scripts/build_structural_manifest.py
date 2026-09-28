#!/usr/bin/env python3
"""
Extracts structural characteristics of the frozen target scope without accessing any dictionary.
Strictly prohibited: computing any term-label similarity, scoring, or assignments.
"""
from pathlib import Path
import csv
import json
from collections import Counter
import statistics

BASE_DIR = Path(__file__).resolve().parent.parent
TARGET_SCOPE_PATH = BASE_DIR.parent / "restricted_dictionary_bruteforce_v2" / "TARGET_SCOPE.tsv"
OUTPUT_PATH = BASE_DIR / "REAL_SCOPE_STRUCTURAL_MANIFEST.json"

def main():
    if not TARGET_SCOPE_PATH.exists():
        raise FileNotFoundError(f"Missing target scope: {TARGET_SCOPE_PATH}")
    
    with TARGET_SCOPE_PATH.open("r", encoding="utf-8") as f:
        reader = list(csv.DictReader(f, delimiter="\t"))
    
    total_occurrences = len(reader)
    page_counts = Counter(row["page_id"] for row in reader)
    tokens = [row["token"] for row in reader]
    unique_tokens = sorted(set(tokens))
    
    # EVA alphabet
    eva_alphabet = sorted(set("".join(tokens)))
    
    # Length statistics
    lengths = [len(t) for t in tokens]
    unique_lengths = [len(t) for t in unique_tokens]
    
    length_counts = Counter(lengths)
    sorted_length_dist = {str(k): length_counts[k] for k in sorted(length_counts)}
    
    # Distinct character count per token
    distinct_char_counts = [len(set(t)) for t in tokens]
    distinct_counts_map = Counter(distinct_char_counts)
    sorted_distinct_dist = {str(k): distinct_counts_map[k] for k in sorted(distinct_counts_map)}
    
    # Hapax counts
    hapax_total = sum(1 for row in reader if row["hapax"] == "1")
    non_hapax_total = sum(1 for row in reader if row["hapax"] == "0")
    
    page_hapax = {}
    for page in sorted(page_counts):
        p_rows = [r for r in reader if r["page_id"] == page]
        page_hapax[page] = {
            "occurrences": len(p_rows),
            "hapax": sum(1 for r in p_rows if r["hapax"] == "1"),
            "non_hapax": sum(1 for r in p_rows if r["hapax"] == "0")
        }
    
    # Character frequencies in target labels (purely structural)
    char_freq = Counter("".join(tokens))
    sorted_char_freq = {c: char_freq[c] for c in eva_alphabet}
    
    manifest = {
        "manifest_version": "v3_structural_freeze",
        "description": "Structural profile of frozen star label target scope without any dictionary scoring",
        "target_source_file": "research/astro_hapax_star_label/restricted_dictionary_bruteforce_v2/TARGET_SCOPE.tsv",
        "occurrences_total": total_occurrences,
        "unique_tokens_total": len(unique_tokens),
        "page_split": dict(page_counts),
        "eva_alphabet": {
            "symbols": eva_alphabet,
            "size": len(eva_alphabet),
            "symbol_counts": sorted_char_freq
        },
        "length_distribution": {
            "min": min(lengths),
            "max": max(lengths),
            "mean": round(statistics.mean(lengths), 4),
            "median": statistics.median(lengths),
            "mode": statistics.mode(lengths),
            "stdev": round(statistics.stdev(lengths), 4),
            "histogram": sorted_length_dist
        },
        "distinct_symbols_distribution": {
            "min": min(distinct_char_counts),
            "max": max(distinct_char_counts),
            "mean": round(statistics.mean(distinct_char_counts), 4),
            "median": statistics.median(distinct_char_counts),
            "histogram": sorted_distinct_dist
        },
        "hapax_distribution": {
            "total_hapax": hapax_total,
            "total_non_hapax": non_hapax_total,
            "by_page": page_hapax
        },
        "page_profiles": {
            page: {
                "occurrences": len([r for r in reader if r["page_id"] == page]),
                "lengths_mean": round(statistics.mean([len(r["token"]) for r in reader if r["page_id"] == page]), 4),
                "distinct_symbols_mean": round(statistics.mean([len(set(r["token"])) for r in reader if r["page_id"] == page]), 4)
            } for page in sorted(page_counts)
        },
        "restrictions_enforced": {
            "dictionary_scores_calculated": False,
            "assignments_generated": False,
            "substitutions_optimized_on_real_data": False,
            "real_data_search_authorized": False
        }
    }
    
    with OUTPUT_PATH.open("w", encoding="utf-8") as f:
        json.dump(manifest, f, indent=2)
    
    print(f"Successfully generated {OUTPUT_PATH}")
    print(f"Total occurrences: {total_occurrences}, Pages: {dict(page_counts)}")
    print(f"EVA Alphabet: {len(eva_alphabet)} characters ({''.join(eva_alphabet)})")

if __name__ == "__main__":
    main()
