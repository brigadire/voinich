#!/usr/bin/env python3
"""
Core v3 writing system and matching engine.
Implements:
- Global table encoding (injective and merge)
- Bounded deletion and abbreviation
- Maximum bipartite matching under capacity constraints (PER_PAGE_CAPACITY_1, GLOBAL_CAPACITY_1)
- Strict order-invariant evaluation
- Deterministic scoring with complexity regularization
"""
import unicodedata
import re
from collections import defaultdict
from typing import Dict, List, Tuple, Set, Optional

SOURCE_ALPHABET = tuple("abcdefghijklmnopqrstuvwxyz") + ("kh", "gh", "sh", "th", "dh")
EVA_ALPHABET = tuple("acdefhiklnoprsty")

def normalize_text(s: str) -> str:
    """ASCII lowercase, single spaces between words."""
    s = unicodedata.normalize("NFKD", s).encode("ascii", "ignore").decode().lower()
    return re.sub(r"[^a-z]+", " ", s).strip()

def encode_word(
    word: str,
    table: Dict[str, str],
    deletion_mode: str = "DROP_UNMAPPED",
    abbreviation: str = "NONE"
) -> str:
    """
    Encodes a source word through the global substitution table.
    - table: mapping from source graphemes to EVA characters.
    - deletion_mode: 'NONE' (keep), 'DROP_UNMAPPED' (delete unmapped), 'SELECTIVE_VOWEL_DROP'.
    - abbreviation: 'NONE', 'SUSPENSION_1', 'SUSPENSION_2', 'PREFIX_4'.
    """
    clean_w = word.replace(" ", "")
    # Sort keys by length descending to match digraphs ('kh', 'gh', etc.) before single letters
    keys = sorted(table.keys(), key=len, reverse=True)
    out = []
    i = 0
    n = len(clean_w)
    
    while i < n:
        matched_key = None
        for k in keys:
            if clean_w.startswith(k, i):
                matched_key = k
                break
        
        if matched_key:
            out.append(table[matched_key])
            i += len(matched_key)
        else:
            ch = clean_w[i]
            if deletion_mode == "DROP_UNMAPPED":
                # Drop unmapped character
                i += 1
            elif deletion_mode == "SELECTIVE_VOWEL_DROP":
                # Drop unmapped vowels, retain unmapped consonants
                if ch in "aeiouy":
                    i += 1
                else:
                    out.append(ch)
                    i += 1
            else: # KEEP
                out.append(ch)
                i += 1
                
    encoded = "".join(out)
    
    # Abbreviation
    if abbreviation == "SUSPENSION_1" and len(encoded) > 3:
        encoded = encoded[:-1]
    elif abbreviation == "SUSPENSION_2" and len(encoded) > 4:
        encoded = encoded[:-2]
    elif abbreviation == "PREFIX_4":
        encoded = encoded[:4]
        
    return encoded

def compute_complexity(
    table: Dict[str, str],
    mapping_mode: str = "INJECTIVE",
    deletion_mode: str = "DROP_UNMAPPED",
    abbreviation: str = "NONE"
) -> int:
    """Calculates formal complexity cost of the model."""
    cost = len(table)
    if mapping_mode == "MERGE_1":
        cost += 1
    if deletion_mode == "DROP_UNMAPPED":
        cost += 2
    elif deletion_mode == "SELECTIVE_VOWEL_DROP":
        cost += 1
    
    if abbreviation == "SUSPENSION_1":
        cost += 1
    elif abbreviation in ("SUSPENSION_2", "PREFIX_4"):
        cost += 2
    return cost

def maximum_bipartite_matching(
    labels: List[dict],
    lexicon_index: Dict[str, Set[str]],
    capacity_policy: str = "PER_PAGE_CAPACITY_1"
) -> Dict[str, str]:
    """
    Computes maximum bipartite matching between labels and star identities.
    Deterministically sorted to guarantee order invariance.
    - labels: list of dicts with 'occurrence_id', 'page_id', 'token'.
    - lexicon_index: dict mapping encoded_string -> set of canonical_identity_ids.
    - capacity_policy: 'PER_PAGE_CAPACITY_1' or 'GLOBAL_CAPACITY_1'.
    Returns: mapping occurrence_id -> canonical_identity_id.
    """
    # Deterministic label order
    sorted_labels = sorted(labels, key=lambda l: (l.get("page_id", ""), l["occurrence_id"]))
    
    if capacity_policy == "PER_PAGE_CAPACITY_1":
        # Solve each page independently
        full_assignments = {}
        pages = sorted(set(l.get("page_id", "default") for l in sorted_labels))
        for p in pages:
            p_labels = [l for l in sorted_labels if l.get("page_id", "default") == p]
            p_match = _match_single_capacity(p_labels, lexicon_index)
            full_assignments.update(p_match)
        return full_assignments
    else: # GLOBAL_CAPACITY_1
        return _match_single_capacity(sorted_labels, lexicon_index)

def _match_single_capacity(
    labels: List[dict],
    lexicon_index: Dict[str, Set[str]]
) -> Dict[str, str]:
    """Maximum bipartite matching where each star identity has capacity 1."""
    # Adjacency: label_id -> list of candidate identities
    # Identities sorted deterministically
    adj = {}
    for l in labels:
        token = l["token"]
        identities = sorted(lexicon_index.get(token, set()))
        adj[l["occurrence_id"]] = identities
        
    # Standard augmenting-path algorithm (Hopcroft-Karp / Kuhn's algorithm)
    # matching: identity -> label_id
    match_star = {}
    
    def dfs(lbl_id: str, visited: Set[str]) -> bool:
        for star in adj[lbl_id]:
            if star in visited:
                continue
            visited.add(star)
            if star not in match_star or dfs(match_star[star], visited):
                match_star[star] = lbl_id
                return True
        return False

    for l in labels:
        dfs(l["occurrence_id"], set())
        
    # Invert to label_id -> identity
    return {lbl_id: star for star, lbl_id in match_star.items()}

def evaluate_table(
    table: Dict[str, str],
    lexicon: Dict[str, Set[str]], # identity -> set of forms
    labels: List[dict],
    mapping_mode: str = "INJECTIVE",
    deletion_mode: str = "DROP_UNMAPPED",
    abbreviation: str = "NONE",
    capacity_policy: str = "PER_PAGE_CAPACITY_1"
) -> dict:
    """Evaluates a concrete substitution table on a label set."""
    # 1. Build inverted index: encoded_string -> set of identities
    idx = defaultdict(set)
    for ident, forms in lexicon.items():
        for form in forms:
            enc = encode_word(form, table, deletion_mode, abbreviation)
            if enc:
                idx[enc].add(ident)
                
    # 2. Maximum bipartite matching
    assignments = maximum_bipartite_matching(labels, idx, capacity_policy)
    matched_count = len(assignments)
    total_labels = len(labels)
    coverage = matched_count / total_labels if total_labels > 0 else 0.0
    
    # 3. Complexity and Score
    complexity = compute_complexity(table, mapping_mode, deletion_mode, abbreviation)
    fitness = 100 * matched_count - complexity
    
    return {
        "matched": matched_count,
        "total": total_labels,
        "coverage": coverage,
        "complexity": complexity,
        "fitness": fitness,
        "assignments": assignments,
        "table": dict(table),
        "mapping_mode": mapping_mode,
        "deletion_mode": deletion_mode,
        "abbreviation": abbreviation
    }
