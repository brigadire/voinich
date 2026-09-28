# Latent-Rule Leakage Audit

## Evaluation of Input Representations
The engine interface `engine.search(terms, labels)` receives two lists of strings.

## Critical Architectural Leakage Identified
1. **Input Structural Pre-Alignment**:
   - `run_development.py` generates `tr, tl = synth(11001, 80)`.
   - Here `tr[i]` and `tl[i]` are pre-matched 1:1.
   - Furthermore, `engine.py` executes `zip(t, l)` inside `induce()`.
   - This relies on exact character-by-character alignment, encoding the generator's latent truth directly into the index sequence!
2. **Latent Rule Identicality**:
   - Training seed 11001 and held-out seed 22001 share the exact same 12 latent substitution rules (`a->o, b->k, c->e, d->r` across 3 roles).
   - Held-out validation is merely an in-distribution sample, not an out-of-distribution generalization test.
3. **Metric Hardcoding**:
   - `run_development.py` measures precision via `sum(r['target'] in 'oker') / len(rules)`.
   - Target alphabet `'oker'` is hardcoded into the validation harness (CR-04).

## Verdict
```text
LATENT_RULE_LEAKAGE=DETECTED_VIA_INPUT_STRUCTURAL_ALIGNMENT
METRIC_HARDCODING=DETECTED
SPLIT_GENERALIZATION=INSUFFICIENT
```
