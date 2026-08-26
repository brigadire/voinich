# Task86C control design (frozen v1)

## Scope and firewall

This is a validation of the historical Task86R measurement apparatus, not a
Voynich experiment. Analysis commands reject paths containing the target name,
the historical corpus identifiers, or the repository target-data directories.
No Task86R HELDOUT/F2 value is loaded. The only reused numerical artifacts are
the target-blind MFC calibration thresholds frozen before Task86R HELDOUT.

## Preregistered design

Primary protocol layer is HISTORICAL_REPLICATION: the exact production M0--M5
fit, VALIDATION PM2 selection, predictive gates, structural gates, generation
scales, and frozen 20,000-attempt PM6 implementation. CONTRACT_REFERENCE is
NOT_EXECUTABLE because the frozen negative-token contract requires exhaustive
enumeration while specifying no resource bound; the historical implementation
introduced the outcome-relevant cap later identified by Task85a-v1.1. No
replacement definition is introduced here.

Independent generators: three fixed parameterizations per M0--M5. M0 is IID
categorical glyph emission with IID length; M1 fixed-order Markov; M2 an
explicit variable-depth context tree; M3 a manually specified deterministic
FSA; M4 a manually specified probabilistic FSA; M5 a productive prefix/core/
suffix slot grammar. M3 chooses one labelled outgoing edge and uses that same
edge's fixed successor (one draw, never an independently chosen successor).
None calls production fitting or generation code.
The theoretical minimal class is M0..M5 respectively. This deliberately tests
class recovery, not semantic labels. IN_FAMILY is omitted from primary evidence
because the historical model interface has no frozen parameter-to-generator
serialization; completing one would test an implementation against itself and
add a scientific choice.

Scale grid is SMALL=512, MEDIUM=2048, VOYNICH_SCALE=38000, LARGE=65536 token
occurrences. There are 8 independent deterministic replicates per mechanism x
scale. Synthetic splits and natural samples use 60/20/20 occurrence partitions.
Natural controls are English (Doyle), expanded Latin (Caesar I--VIII plus
Virgil), and Sanskrit (Panchatantra; typologically distinct), selected only for
pre-existing repository provenance / public-domain availability and scale.

## Frozen success criteria

Model-recovery capable requires, on INDEPENDENT generators at VOYNICH_SCALE:
(a) minimal-sufficient recovery >= 0.80 overall, (b) exact recovery >= 0.60,
(c) NONE <= 0.05, and (d) every class minimal-sufficient recovery >= 0.60.
SUPPORTED requires all four; PARTIAL requires NONE <= 0.20 and overall
minimal-sufficient recovery >= 0.50; otherwise NOT_SUPPORTED. False complexity
is recovery above the theoretical minimum and underfit is recovery below it.
Natural-language applicability is SUPPORTED when at least one model is adequate
in >=80% of VOYNICH_SCALE samples in every required language, PARTIAL when this
holds for at least one language, otherwise NOT_SUPPORTED. These thresholds are
frozen before any analysis output or unblinding.

PM6 is executed unchanged and reports requested/constructible negatives by
length, alphabet saturation, attempt exhaustion and rejection cause. No sampler
repair is allowed. GENERAL means failure in both synthetic and natural branches;
NATURAL_LANGUAGE_COMMON in >=2 natural languages only; SCALE_DEPENDENT requires
a monotone >=0.50 difference between SMALL and LARGE failure rates; otherwise
INCONCLUSIVE (VOYNICH_SPECIFIC cannot be established without target execution).

## Distributed contract

Jobs bind immutable input SHA-256, config SHA-256, seed, protocol and logical
identity. Worker assignment is absent from job_id. Result JSON is canonical
indented JSON with runtime/timestamps excluded from scientific equivalence
hashing. Duplicate disagreement quarantines both. Infrastructure failures may
retry unchanged; scientific failures may not. Six representative preflight
jobs (M0, M2, M4, M5, English, Latin) must match local and 10.10.24.105 before
bulk execution. The same locally built executable bytes are copied to remote.
