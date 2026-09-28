# Audit Errata: `LEXICON_INDEPENDENCE` naming collision

This note resolves an internal naming collision inside the frozen
`restricted_dictionary_bruteforce_v3_audit` package. **The frozen audit package itself is not
modified** — this file only documents which sense of "independence" each of its own statements
uses, so remediation work does not misread `LEXICON_INDEPENDENCE=FAIL` as a contradiction of
`AUDIT_PROTOCOL.md` row B3.

## The apparent contradiction

- `PREPRODUCTION_AUDIT_REPORT.md`'s mandatory final status block states `LEXICON_INDEPENDENCE=FAIL`.
- `AUDIT_PROTOCOL.md` row B3 and `LEXICON_AUDIT_REPORT.md` §B3 ("Independence from EVA") state
  **"B3 holds"**: a grep of `build_historical_lexicon.py` for `TARGET_SCOPE`, `f68r`, or
  `EVA_ALPHABET` found no reference — i.e., no EVA-directed selection was detected in the
  lexicon-construction code.

Read superficially, "FAIL" next to "B3 holds" looks contradictory.

## Resolution

They are not the same claim. `restricted_dictionary_bruteforce_v3_audit` used the single label
`LEXICON_INDEPENDENCE` for two distinct properties, and its own `VALIDATION_REPORT.md` (line 16)
already glosses the final-block value as **"cross-identity separability"**, not EVA-independence:

1. **Independence from EVA (§B3):** whether the lexicon's *construction process* was steered by
   knowledge of the target EVA labels/alphabet. Code-level grep found no such coupling.
   **This is `B3` / what `PREPRODUCTION_AUDIT_REPORT.md` intends by "no EVA-directed selection
   confirmed" in Section 2 ("What Held Up").**
2. **Cross-identity provenance separability (the final-block `LEXICON_INDEPENDENCE=FAIL`):**
   whether distinct canonical star identities have attestations that are cleanly separable from
   each other. 6/94 identities share a normalized form with a different identity; of the two
   checked, one (Mizar/Mirach, `HIST_STAR_ATT_0188`) is a **confirmed misattribution** (F005), not
   genuine historical ambiguity. This is a property of the *lexicon content*, unrelated to whether
   the content was chosen with EVA in mind.

A lexicon can simultaneously be constructed with no knowledge of the target alphabet (property 1,
PASS) and contain internally confused/misattributed entries (property 2, FAIL). These are
orthogonal axes — code-provenance vs. data-content correctness — and both statements in the frozen
audit are correct under their own (different) definitions.

## Disambiguated fields used by this remediation

```text
LEXICON_EVA_INDEPENDENCE=PASS
LEXICON_PROVENANCE_VALIDITY=FAIL
LEXICON_SCOPE_ADEQUACY=PARTIAL_BUT_USABLE
```

- `LEXICON_EVA_INDEPENDENCE=PASS` — carries forward `restricted_dictionary_bruteforce_v3_audit`'s
  §B3 finding verbatim; the frozen audit's code-level check is accepted as sufficient evidence for
  this axis, and this remediation's own lexicon construction (Section 5 of the task) is also
  performed blind to EVA (new rows added without reference to any label token).
- `LEXICON_PROVENANCE_VALIDITY=FAIL` — carries forward the frozen audit's *content-level* finding
  (confirmed misattribution + confirmed anachronisms + 8.8% coverage, below the ≥30% mandate) as the
  pre-remediation baseline. This is the axis Gate L-R (Section 3-6 of the task) exists to fix; its
  post-remediation value is reported separately in `LEXICON_REMEDIATION_REPORT.md` and is **not**
  assumed to still be FAIL just because the v3 baseline was.
- `LEXICON_SCOPE_ADEQUACY=PARTIAL_BUT_USABLE` — carries forward the frozen audit's
  `LEXICON_SCOPE=PARTIAL_BUT_USABLE` unchanged, as a description of the v3 baseline only.

These three fields are reported for the *v3 baseline* in this file. This remediation's own
post-Gate-L-R values (which supersede the baseline for anything built in `remediation_v1`) are
reported in `LEXICON_REMEDIATION_REPORT.md` and echoed in the final status block of
`REMEDIATION_REPORT.md`.
