# Phase III Line A research questions

**Design version:** task85-v1. **Authority:** tasks_ph3/00-PHASE3_GOALS.txt sections 4-18, 45; tasks_ph3/task85.txt section 2. **Scope:** Line A (Task85-89) only; Line B (Task90-94) research questions are out of scope here and are not answered, anticipated, or tuned against here (Line A/Line B firewall, PHASE3_GOALS section 32).

## Central Line A question

> Can the observed structure of Voynichese be described by a compact formal generative system, capable of predicting previously unused parts of the corpus?

This is explicitly **not** "What is the grammar of the Voynich language?" and **not** "What does the text mean?" (task85 section 2).

## Line A research questions, scoped from PHASE3_GOALS section 45

- **RQ1 (Task86-88).** Can Voynichese be represented by a compact, held-out-validated formal grammar at some level G1/G2/G3?
- **RQ2 (Task88).** What is the minimum sufficient grammar, `G_min`, under the frozen `GRAMMAR_MINIMAL`/`GRAMMAR_SUFFICIENT` criteria (`GRAMMAR_VALIDATION_CONTRACT.md`)?
- **RQ3 (Task89).** How much information remains after conditioning on `G_min` (`GRAMMAR_RESIDUAL_CONTRACT.md`)?
- **RQ4 (Task89).** Is that residual statistically structured, or does it resemble a matched null?
- **RQ9 (Task89, partial).** Can formal grammar and message-free generation be statistically distinguished without assuming a plaintext? (Task89 only tests whether `G_min` alone, run generatively, reproduces VM's fingerprint across many seeds; a full RQ9 verdict also needs Line B / Task95-98.)
- **RQ10 (Task89, partial for Line A's own model space).** Which observed VM properties remain unexplained by every Line A model tested? (`GRAMMAR_FAILURE_REGISTRY.tsv`'s `STRUCTURAL_VALIDATION_FAILED` rows and the residual-structure tests in `GRAMMAR_RESIDUAL_CONTRACT.md` section 4 are Line A's contribution to this question; a full verdict also needs Line B.)

## Task85-specific sub-questions this design must leave unambiguous for Task86

1. What counts as a TOKEN, GLYPH, LINE, FOLIO for grammar-fitting purposes? -> `GRAMMAR_UNIT_REGISTRY.tsv`.
2. Which model classes are eligible, and why exactly these? -> `GRAMMAR_MODEL_REGISTRY.tsv`, `TASK85_DESIGN.md` section 4.
3. Which partition may Task86 read, and under what role? -> `GRAMMAR_CORPUS_SPLIT.tsv`/`GRAMMAR_CORPUS_SPLIT_MANIFEST.json`, `GRAMMAR_VALIDATION_CONTRACT.md` section 1.
4. What must Task86 report before selecting a G1 model? -> `TASK86_HANDOFF.md`.

## Explicit non-questions (task85 section 63, restated for Line A as a whole)

Line A does not ask, and no Task85-89 artifact may be interpreted as answering: what Voynichese means; whether it is natural language; whether it has meaning at all; what any illustrated page depicts; whether any proposed translation or plaintext candidate is correct. These remain closed by the semantics firewall (`TASK85_DESIGN.md` section 10) throughout Task85-89.
