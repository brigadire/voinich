# Task85c-h — scientific contract defect

Finding `H-SC03-M0-UNK-PROBABILITY-CONTRADICTION` blocks a conforming implementation before production materialization.

The mandatory inherited `M0-FIT` vector fixes, for DEVELOPMENT tokens `ab,a` and alpha 1, `p(a)=0.375`, `p(b)=0.25`, and `p(EOS)=0.375`. Those values already sum to 1 and its stated denominator is 8. The mandatory inherited `M0-UNSEEN` vector simultaneously requires an unseen glyph mapped to `UNK` to have positive probability when alpha is positive. Both vectors are embedded as `INHERITED_UNCHANGED` in the V1.2 generation golden suite.

No probability row can satisfy both requirements. The V1.2 M0 definition independently specifies outcomes `{a,b,UNK,EOS}` and additive-alpha smoothing, yielding denominator `5 + 1*4 = 9`, positive `p(UNK)=1/9`, and probabilities `p(a)=p(EOS)=1/3`, `p(b)=2/9`. That conforms to `M0-UNSEEN` but fails `M0-FIT`; denominator 8 conforms to `M0-FIT` but assigns no remaining probability mass to `UNK`.

This changes fitted-model bytes, predictive quantities, RNG-to-symbol generation, and potentially minimality/final verdict. Authority precedence cannot resolve two contradictory cases within the same frozen inherited suite. Task section 134 therefore requires the scientific-contract-defect stop rather than an implementer choice.

Run `python3 research/phase3/task85c-h/reference/reproduce_m0_golden_contradiction.py` for the minimal machine-checkable proof.
