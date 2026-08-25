# Grammar residual contract (preparation for Task89)

**Design version:** task85-v1. **Authority:** task85 sections 41-44; PHASE3_GOALS sections 11-14. **Frozen before grammar fitting:** yes (design only; the final estimator freeze may occur later, up to Task89, per task85 section 42). **Target-blind:** yes.

Task85 does not compute a residual. This document fixes what Task89 will treat as the residual, so Task89 cannot silently redefine it after seeing G_min.

## 1. What counts as residual

```
V = G_min + R
```

`R` is the information in the observed HELDOUT sequence not predicted by the frozen `G_min` (task85 sections 30-31). `R` is never called, or interpreted as, plaintext, a candidate translation, or a semantic quantity (task85 section 41, section 45 semantics firewall). `R` is a property of the model/data pair, not a claim about the manuscript's content.

## 2. Residual representation (candidates; Task89 selects and freezes one before use)

- **Surprisal sequence**: the per-TOKEN (or per-GLYPH, for a G1-only `G_min`) value `-log2 P(unit_i | G_min, context_i)` over HELDOUT, in manuscript order.
- **Coding residual**: the MDL-style leftover `L(V | G_min)` under the coding scheme of `GRAMMAR_COMPLEXITY_CONTRACT.md`, i.e. the actual bits needed to encode HELDOUT given `G_min`, as opposed to `G_min`'s own predicted expectation.
- **Prediction errors**: for a `G_min` with a well-defined most-probable-next-unit, the sequence of hit/miss (or rank-of-truth) values.
- **Sampled-choice residual**: for a fully stochastic `G_min`, the sequence of quantile positions of the observed unit within `G_min`'s own predictive distribution (a probability-integral-transform sequence), used because it is comparable in distribution across model classes with different alphabets (GLYPH- vs TOKEN-level).

Task89 must record which representation it selects and why; it may report more than one but must nominate exactly one as primary before testing for structure in it (task85 section 42 "estimator... frozen").

## 3. Residual entropy / description-length estimator (candidates)

- `H(V)`: unconditional entropy of HELDOUT under a frozen order-0 or matched-null code.
- `H(V | G_min)`: mean surprisal of the chosen representation (section 2) over HELDOUT.
- `ΔH_G = H(V) - H(V | G_min)`: the frozen reduction quantity (PHASE3_GOALS section 12).
- An MDL-equivalent: `L(V) - (L(G_min) + L(V | G_min))` using `GRAMMAR_COMPLEXITY_CONTRACT.md`'s coding scheme throughout, so the residual-entropy and description-length views use one shared code, not two incompatible accountings.

Whichever estimator Task89 freezes must be: applicable to HELDOUT (not DEVELOPMENT/VALIDATION); deterministic given `(G_min, seed, HELDOUT)`; bias-audited (e.g. checked against the message-free calibration battery `MFC0`-`MFC3`'s own known-zero-residual-beyond-generator-noise case, `GRAMMAR_BASELINE_REGISTRY.tsv`); scale-aware (reported per-unit, not only as a corpus-wide total, so partitions of different size remain comparable).

## 4. Residual structure tests (Task89 must define/run at least these; Task85 fixes only the required scales, not the specific test statistics)

- autocorrelation/dependence of the residual sequence at short lags;
- local-sequence structure (residual dependence on the immediately preceding residual value(s));
- LINE-position dependence of the residual;
- FOLIO/section dependence of the residual;
- any `GENERATION_APPLICABLE`, non-`skeleton_only` F2 family from `GRAMMAR_F2_APPLICABILITY.tsv`, recomputed on the residual representation rather than on raw TOKENs, where the family's definition transfers meaningfully to a residual sequence.

Every such test requires a matched residual null (e.g. the same test statistic computed on the residual of a message-free calibration corpus, `MFC0`-`MFC3`, of matched scale) — a residual test without a matched null cannot distinguish "low residual entropy" from "structured residual information" (PHASE3_GOALS section 14), which is the central distinction this contract exists to protect.

## 5. Boundary to message-free generation (Task89's other half)

`G_min + RNG -> M_i` (many independent seeds, task85 section 44 / PHASE3_GOALS sections 16-18) is a *separate* experiment from residual characterization: it asks whether `G_min` alone, run generatively without HELDOUT's specific content, can repeatedly reproduce VM's Fingerprint V2 profile. `G_min` must never be selected, reselected, or tuned according to how well its message-free samples match VM (task85 section 44); this document and `GRAMMAR_VALIDATION_CONTRACT.md` section 6 are what fixes `G_min` before that experiment runs.

## 6. Firewalls carried into Task89

Residual and message-free findings are reported as `RESIDUAL_INFORMATION_STATUS` and `MESSAGE_FREE_REPRODUCTION_STATUS` (PHASE3_GOALS section 15/18's terms), never translated into a plaintext, translation, or meaning claim (task85 section 45), and never selected to be compatible or incompatible with a Fontana/external-memory mechanism (task85 section 46) or with Phase II's Task81-83r mechanism-fit results (task85 section 47).
