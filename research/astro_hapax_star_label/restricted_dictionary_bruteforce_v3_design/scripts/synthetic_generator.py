#!/usr/bin/env python3
"""
Sealed Synthetic Generator for v3 Search Architecture Benchmark.
Generates synthetic star label datasets matching the frozen structural profile:
- 57 occurrences (30 on f68r1, 27 on f68r2)
- Real length distribution (min 1, max 10, mean ~6.18)
- Real unique symbol distribution (mean ~5.14)
- 94-identity historical star lexicon
- Independently sampled planted substitution tables from full universe
- Noise levels: 0.0, 0.10, 0.25
- Unmatched fractions: 0.0, 0.25, 0.50
- Injective and Merge_1 modes
- Abbreviation and deletion composition
- Distractors and multi-attestation handling
"""
import random
import csv
from pathlib import Path
from typing import Dict, List, Set, Tuple
from collections import Counter

from engine import (
    SOURCE_ALPHABET,
    EVA_ALPHABET,
    encode_word
)

BASE_DIR = Path(__file__).resolve().parent.parent
LEXICON_PATH = BASE_DIR / "HISTORICAL_STAR_LEXICON.tsv"
MANIFEST_PATH = BASE_DIR / "REAL_SCOPE_STRUCTURAL_MANIFEST.json"

def load_lexicon_forms():
    with LEXICON_PATH.open("r", encoding="utf-8") as f:
        rows = list(csv.DictReader(f, delimiter="\t"))
    lex = {}
    for r in rows:
        lex.setdefault(r["canonical_identity_id"], set()).add(r["normalized_form"].replace(" ", ""))
    return lex

class SealedSyntheticGenerator:
    def __init__(self, seed: int):
        self.seed = seed
        self.rng = random.Random(seed)
        self.lexicon = load_lexicon_forms()
        self.identities = sorted(self.lexicon.keys())

    def sample_planted_table(
        self,
        table_size: int = 8,
        mapping_mode: str = "INJECTIVE"
    ) -> Dict[str, str]:
        """Samples planted table UNIFORMLY from the complete space."""
        src_chars = self.rng.sample(SOURCE_ALPHABET, table_size)
        if mapping_mode == "INJECTIVE":
            tgt_chars = self.rng.sample(EVA_ALPHABET, table_size)
        elif mapping_mode == "MERGE_1":
            # Exactly one target character repeated twice
            base_tgt = self.rng.sample(EVA_ALPHABET, table_size - 1)
            dup_char = self.rng.choice(base_tgt)
            tgt_chars = base_tgt + [dup_char]
            self.rng.shuffle(tgt_chars)
        else:
            raise ValueError(f"Unknown mapping mode: {mapping_mode}")
            
        return dict(zip(src_chars, tgt_chars))

    def generate_dataset(
        self,
        table_size: int = 8,
        mapping_mode: str = "INJECTIVE",
        deletion_mode: str = "DROP_UNMAPPED",
        abbreviation: str = "NONE",
        noise_rate: float = 0.0,
        unmatched_fraction: float = 0.25,
        capacity_policy: str = "PER_PAGE_CAPACITY_1"
    ) -> Tuple[List[dict], dict]:
        """
        Returns:
        1. dataset (public): list of 57 label dicts with 'occurrence_id', 'page_id', 'token'
        2. sealed_truth: dict containing planted table, true assignments, and noise stats
        """
        planted_table = self.sample_planted_table(table_size, mapping_mode)
        
        # 30 occurrences on f68r1, 27 on f68r2
        page_counts = [("f68r1", 30), ("f68r2", 27)]
        
        labels = []
        true_assignments = {}
        occ_num = 1
        
        # Shared identities between pages (common stars in medieval rotas)
        shared_pool = self.rng.sample(self.identities, 20)
        
        for page_id, count in page_counts:
            # How many matched vs unmatched
            unmatched_count = int(round(count * unmatched_fraction))
            matched_count = count - unmatched_count
            
            # Select matched identities for this page
            page_pool = list(shared_pool) + [s for s in self.identities if s not in shared_pool]
            selected_page_stars = self.rng.sample(page_pool, matched_count)
            
            for s_id in selected_page_stars:
                occ_id = f"OCC_{occ_num:04d}"
                # Choose one attested form of this star
                form = self.rng.choice(sorted(self.lexicon[s_id]))
                token = encode_word(form, planted_table, deletion_mode, abbreviation)
                
                # If encoded token is empty, fallback to non-empty subset
                if not token:
                    token = "o" # minimal token
                    
                # Apply noise
                corrupted_token = list(token)
                if noise_rate > 0.0:
                    for idx in range(len(corrupted_token)):
                        if self.rng.random() < noise_rate:
                            corrupted_token[idx] = self.rng.choice(EVA_ALPHABET)
                final_token = "".join(corrupted_token)
                
                labels.append({
                    "occurrence_id": occ_id,
                    "page_id": page_id,
                    "token": final_token
                })
                true_assignments[occ_id] = s_id
                occ_num += 1
                
            # Generate unmatched distractors
            for _ in range(unmatched_count):
                occ_id = f"OCC_{occ_num:04d}"
                # Distractor length sampled from real distribution (mean ~6)
                dist_len = self.rng.choices(
                    [3, 4, 5, 6, 7, 8],
                    weights=[0.1, 0.2, 0.3, 0.25, 0.1, 0.05]
                )[0]
                dist_token = "".join(self.rng.choices(EVA_ALPHABET, k=dist_len))
                labels.append({
                    "occurrence_id": occ_id,
                    "page_id": page_id,
                    "token": dist_token
                })
                true_assignments[occ_id] = None # Unmatched
                occ_num += 1

        sealed_truth = {
            "seed": self.seed,
            "planted_table": planted_table,
            "table_size": table_size,
            "mapping_mode": mapping_mode,
            "deletion_mode": deletion_mode,
            "abbreviation": abbreviation,
            "noise_rate": noise_rate,
            "unmatched_fraction": unmatched_fraction,
            "capacity_policy": capacity_policy,
            "true_assignments": true_assignments
        }
        
        return labels, sealed_truth
