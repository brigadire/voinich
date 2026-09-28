# Inference Path Audit: M2R Engine Architecture

## Data Flow Analysis
The required M2R data flow:
```text
surface terms + surface labels
→ candidate recurrent units
→ candidate rules
→ assignments
→ scoring
```

Actual observed data flow in `engine.py`:
```text
parallel terms & labels (pre-aligned)
→ zip(terms, labels)
→ zip(t, l) at character index i
→ Counter(role(i), char_t, role(i), char_l)
→ filter support >= k (where k in 2, 3, 4)
→ argmax score
```

## Discrepancies & Fatal Omissions
1. **No Recurrent Morphological Units**:
   - `units(s)` returns `tuple(s)` and is never used.
   - The engine operates purely on individual graphemes (unigrams). Multi-character morphemes, syllables, or stems are completely absent.
2. **No Assignment Inference**:
   - The engine has no search over the combinatorial space of pairings between terms and candidate dictionary entries.
   - Given unaligned sets of tokens, the engine fails completely (0% assignment accuracy).
3. **Trivial Search Space**:
   - `search()` evaluates only 3 discrete parameter values: `k in (2, 3, 4)`.
   - There is no optimization over rules, no beam search, no prune step, no latent structure discovery.

## Verdict
```text
INFERENCE_PATH_VALIDITY=FAIL
MORPHOLOGY_INFERENCE=ABSENT_UNIGRAM_ONLY
ASSIGNMENT_INFERENCE=ABSENT_PREALIGNED_ONLY
```
