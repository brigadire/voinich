# Phase II scientific report: from transformation ambiguity to the external-memory hypothesis

## 1. Abstract

Phase II asked whether the Voynich Manuscript (VM) could be an **external-memory system**: a structured record that does not autonomously contain a complete plaintext, but works together with conventions, context, or knowledge retained by its user. The question followed from Phase I, which established a strongly constrained symbolic system while showing that surface compatibility, plaintext dependence, reversibility, information retention, and autonomous recovery are different properties. Phase I did not establish that the VM itself is lossy or undecodable.

Phase II expanded the measured description into Fingerprint V2, operationalized historically grounded and generalized external-memory mechanisms, and compared frozen portfolios of Fontana-derived mechanisms, natural text, simple nulls, one historical abbreviation tradition, and selective-extraction operators. The authoritative confirmatory experiment is Task83r. Its direct model/target intersection contains only 3 of 13 CORE metrics, all from one edit family. Within that narrow space natural text is descriptively closest, but no class passes the frozen multi-family and null-separation support gates. The final result is therefore `BEST_SUPPORTED_CLASS = INCONCLUSIVE` and `MECHANISM_IDENTIFICATION_FROM_F2 = NOT_IDENTIFIABLE` ([Task83r report](../../../research/phase2/task83r/TASK83R_REPORT.md)).

External memory remains structurally compatible only at `LEVEL_1`; the tested BDD abbreviation tradition is disfavored; selective extraction retains weak partial compatibility but no hidden channel is detected. Phase II identifies neither a generating mechanism nor a decoding. Its decisive result is that the current measurement/model intersection is too narrow for such identification.

## 2. Research context and evidential language

This report is a synthesis of frozen repository evidence. It introduces no new mechanism, metric, threshold, corpus, hypothesis test, or post-target tuning. `OBSERVED` denotes a stored measurement; `SUPPORTED` a passed declared evidential gate; `COMPATIBLE` structural consistency without identification; `DISFAVORED` scoped counterevidence; `NOT_SUPPORTED` a failed declared gate; `NOT_TESTED` absence of a direct Phase II test; and `NOT_IDENTIFIABLE` failure of the available evidence to distinguish generating classes.

“Glyph,” “token,” “line,” “page,” and “locus” are transcription or analytical units, not assertions about letters, words, sentences, or semantic units. “External memory,” “mnemonic mechanism,” “shorthand,” “selective extraction,” “cipher,” and “formal grammar” are not synonyms.

## 3. Phase I → Phase II

The authoritative predecessor is the [Phase I research report](../../phase1/PHASE1_RESEARCH_REPORT.md). Phase I measured a conjunction of strong within-token positional constraints, enriched edit neighborhoods and local repetition, weak average visible-token order, stronger glyph-edge coupling, representation-sensitive low conditional entropy, modest local sequence residuals, and line/page/manuscript heterogeneity. It disfavored only bounded simple variants: the selected inverse transposition, simple position-independent random homophony as a joint explanation, an inverse-homophony method whose synthetic gate failed, and finite-grid mechanisms lacking constrained formation. It did not select natural/formal language, complex transformation, or structured message-free generation.

The logical transition was:

1. **Measured:** VM surface structure is jointly constrained and cannot be summarized by marginals alone.
2. **Measured on synthetic mechanisms:** output may remain plaintext-dependent yet require a codebook, convention, state, or context for recovery; some tested many-to-one mechanisms are intrinsically ambiguous, while reversible mechanisms can be fragile.
3. **Interpretation:** statistical resemblance does not identify information behavior or autonomous decodability.
4. **Motivation:** test a class in which part of the usable information is deliberately outside the visible record—in user knowledge, retrieval convention, or context.

Thus Phase II did not assume information loss in the VM. It tested whether knowledge-dependent representation is operationally definable and whether frozen instances reproduce the measured structure. The conclusion map is in [PHASE1_TO_PHASE2_CONCLUSION_MAP.tsv](PHASE1_TO_PHASE2_CONCLUSION_MAP.tsv).

## 4. Research questions

- **RQ1:** Can a more complete and transcription-validated Fingerprint V2 be defined?
- **RQ2:** Can Fontana’s described memory mechanisms be operationalized without turning reconstruction assumptions into historical facts?
- **RQ3:** Can those mechanisms generate VM-like statistical properties?
- **RQ4:** Does external memory outperform natural-text or simple-transform alternatives?
- **RQ5:** Is a shorthand-like transformation signature detected?
- **RQ6:** Is a selective-extraction or acrostic-like signature detected?
- **RQ7:** Is the generating class identifiable from the available Fingerprint V2 intersection?

RQ1–RQ3 received methodological and partial structural answers. RQ4–RQ7 were answered negatively at the support/identification level, with the scoped compatibility qualifications below.

## 5. Fingerprint V2

### 5.1 Measurement architecture

Task73 audited Fingerprint V1 and found uneven coverage: glyph/token marginals, repetition, and line/local structure were comparatively strong; lexical paradigms, cross-scale dependence, hierarchy, and genuine 2D/page organization were weak or absent. It therefore specified a hierarchical fingerprint and family-balanced comparison rather than a flat omnibus vector ([Task73 report](../../../research/phase2/fingerprint/TASK73_REPORT.md)).

The implemented layers were:

| Layer | Scientific role |
|---|---|
| token/edit structure | edit graphs, connectedness, clustering, degree/frequency relations |
| lexical paradigms | rule-support concentration and core/affix attachment, guarded by grammar-bounded nulls |
| transitions and cross-scale structure | family/context, position, locus, regime, and structural-distance relations |
| positional and physical-line structure | line length, within-line position, boundary asymmetry, repetition |
| page/folio hierarchy | locus, folio coherence/progression, folio/section variance shares |
| recto/verso and 2D-lite | physical-leaf coherence and layout-position information from IVTFF metadata |
| local regimes | changes and conditioned heterogeneity without semantic labels |
| transcription stability | independent ZL3b and IT2a comparison |

Task75 implemented the lexical/edit block but produced no canonical scientific result. Task77 ran it on the VM: a large edit-rule network reproduced across folio halves, but its aggregate concentration did not exceed the stronger length/position/bigram-matched C-GRAMMAR null; the LP2-gated productive-family graph was empty. This refines “edit proximity” into a bounded claim: reproducible edit structure exists, but a productive paradigm system exceeding that grammar-bounded null was not supported ([Task77 report](../../../research/phase2/fingerprint/TASK77_REPORT.md)).

Task79 added line, boundary, locus, folio, hierarchy, recto/verso, and 2D-lite estimators. It found measurable line, boundary, locus, folio, and hierarchical structure but correctly withheld a freeze because transcription and historical-control gates remained open ([Task79 report](../../../research/phase2/fingerprint/TASK79_REPORT.md)). Task79b audited shorthand/extraction diagnostics and kept unvalidated candidates out of F2 ([Task79b report](../../../research/phase2/notation-audit/TASK79B_REPORT.md)). Task79c obtained IT2a, historical BDD abbreviation, and procedural controls and closed the scientific-definition freeze. Task83b later rebuilt its stochastic artifacts deterministically without changing the 13 CORE definitions or statuses.

### 5.2 CORE, supporting, controls, and stability

Fingerprint V2.1 retains 13 CORE metrics: `2DL1`; `BP1`; `EF1–EF3`; folio and section `HR1`; `LC1–LC2`; `LS2–LS3`; `PF2`; and `PF5`. Twenty further metrics are supporting and are not independently weighted as CORE evidence. CORE means admitted to the frozen primary fingerprint after relevance, stability, redundancy, and control review; it does not mean causal or semantic.

Historical controls include natural prose, the BDD abbreviation control, and MS-DOS procedural text. Nulls preserve explicitly declared nuisance structure, including length, frequency, position, line/page membership, or grammar constraints. Distances are standardized by pre-target natural-control scales, aggregated within families, and then family-balanced. Missing values are not imputed.

Across independent ZL3b and IT2a transcriptions, the deterministic refreeze preserved 3 `STABLE`, 10 `DIRECTION_STABLE`, and 0 `UNSTABLE` CORE classifications ([Task83b report](../../../research/phase2/task83b/TASK83B_REPORT.md)). This is stability of the validated CORE conclusions, not identity of every numeric value or every supporting metric.

## 6. Historical motivation: Giovanni Fontana

The source study concerns Giovanni Fontana’s *Secretum de thesauro experimentorum ymaginationis hominum*, preserved as BnF NAL 635 and dated approximately to 1430. The manuscript is a small illustrated Latin codex whose main text uses a simple 23-sign monoalphabetic symbolic writing while preserving ordinary spacing and textual organization. That writing is reversible with its alphabet and is analytically separate from the memory mechanisms ([Task74 report](../../../research/phase2/fontana/TASK74_REPORT.md); [sources](../../../research/phase2/fontana/SOURCES.md)).

**Explicitly described or source-grounded:** Fontana distinguishes natural, intellectual, and artificial memory; discusses mental *loci et imagines*; inventories instrumental devices including the alphabetic-ring *speculum*, serpent/spiral, wheels, cylinder, indexed board, and clock; and uses operations such as selection, rotation, alignment, placement, ordered traversal, indexing, association, repetition, overlay/removal, and signalling.

**Operational reconstruction:** external configurations can retain letters, positions, indices, or temporal state, while retrieval requires such conventions as ring order, reading radius, starting point, direction, path, index table, or cue association.

**Modern interpretation:** some systems satisfy a graded external-memory definition because an external state persists and later supports human retrieval, while some *loci/imago* practice remains primarily internal. The Fontana-only conclusion is `FONTANA_EXTERNAL_MEMORY_PARTIALLY_SUPPORTED`.

**Speculative relation to the VM:** Fontana demonstrates that historically proximate, knowledge-dependent symbolic devices are possible and operationally heterogeneous. Phase II found no evidence of Fontana’s authorship, influence, acquaintance with the VM, or use of these devices to produce it. Statistical compatibility cannot establish any such historical relation.

## 7. Operational definition and distinctions

In Phase II an **external-memory mechanism** has: (1) a persistent observable external state; (2) a declared encoding or state-setting operation; (3) a later retrieval procedure; (4) an explicit accounting of information carried by the state versus convention, path, context, or internal association; and (5) measurable recovery under knowledge ablations. The label is operational, not metaphorical.

| Class | Information/recovery distinction |
|---|---|
| reversible cipher | ciphertext plus key/rule uniquely recovers the input; secrecy is not memory assistance |
| lossy transform | multiple inputs map to the same observable; no knowledge can restore information destroyed by the map |
| shorthand | a conventional shortened notation paired with expansions; recovery may be ambiguous or context-dependent |
| mnemonic cue system | external cue prompts content substantially retained in learned/internal association |
| selective extraction | an operator selects a subsequence from a carrier; discarded carrier information is unavailable from the output alone |
| formal generated text | symbols are produced by explicit grammar/rules; meaningfulness is a separate question |
| natural-language text | linguistic content is represented by a historically used language; surface units need not equal VM tokens |

Autonomous recovery uses only the observable and public mechanism. Knowledge-assisted recovery additionally supplies declared convention/key/path/context/internal association. Ambiguous recovery returns multiple compatible inputs. Unrecoverable information has been destroyed by a many-to-one mapping. These categories extend Phase I’s finding that reversibility, declared-codebook decoding, collision ambiguity, and robustness are distinct.

## 8. Fontana operational models and blind portfolios

The scientific progression was source → bounded reconstruction → typed operation algebra → executable mechanisms → blind generation → corpus scaling → F2 extraction.

| Family | Historical basis and modern formalization | Information/recovery behavior | Main limitation |
|---|---|---|---|
| F01 *speculum* | independently rotated alphabetic rings; literal rotational storage | exact with full convention; prior-knowledge ablation creates combinatorial ambiguity; no built-in error correction | capacity/alphabet and some physical conventions reconstructed |
| F08 *serpens* | ordered spiral positions; positional traversal | exact for declared path/convention; order-changing damage prevents exact recovery | source leaves boundary/capacity/direction partly open |
| F11 *arismetricum* | indexed occupied positions/cues; lookup model | cue-only unless index/cue convention supplies meaning; context can narrow candidates | mapping semantics underdetermined |
| F12 *horalogius* | cyclic temporal state and emitted reminder | cue-only without learned association; signal is not remembered content | drive, calibration, and human retention unvalidated |
| F07/F10 | cyclic wheel and aligned-band sensitivity models | show state/profile dependence | reference-only, not frozen as complete mechanisms |
| M-RESTRICTED | type-valid rotation+index and storage+association compositions | tests bounded generalized external-memory space | explicitly unattested, not historical Fontana claims |

F01’s direct operational test recovered 24/24 messages under intact state/full knowledge; damage to any used ring yielded 0/170 exact recoveries. Those numbers characterize the reconstructed device, not the VM ([Task76 report](../../../research/phase2/fontana/f01_speculum/TASK76_REPORT.md)). Task78 established heterogeneous rather than universal device behavior. Task80 froze a typed algebra separating `C-FONTANA` attested compositions, `G-ALLOWED` counterfactuals, and forbidden constructions ([Task80 report](../../../research/phase2/fontana/task80/TASK80_REPORT.md)).

Task81 froze 672 target-blind mechanism/recovery jobs. Task82 completed them and separated exact, cue-only, ambiguous, wrong-knowledge, and collision behavior; generic F2 extraction was not yet preregistered. Task82a scaled 16 mechanisms into corpus-sized observable documents while declaring assembler-created token/line boundaries and no pages. Task82a.1 extracted the maximum honest intersection: 3/13 direct CORE metrics; an additional 4/13 assembler projections were retained separately and never combined with direct distance ([Task82a.1 report](../../../research/phase2/task82a1/TASK82A1_REPORT.md)). Synthetic outputs therefore model token streams, not manuscripts.

## 9. Alternative mechanism classes

### 9.1 Historical shorthand/abbreviation

Task82b extracted 7,150 paired TEI `<abbr>/<expan>` units from books 6/7/11/12/13 of the Burchards Dekret Digital witness `koeln-edd-c-119`. Classes included suspension, contraction, whole-word special signs, and mark-only abbreviation, with no-visible-change/other-substitution edge cases. It measured before/after F2 trajectories, abbreviation/expansion ambiguity, context dependence (SX diagnostics), and random-, frequency-, and position-matched deletion nulls ([Task82b report](../../../research/phase2/task82b/TASK82B_REPORT.md)).

This is one manuscript/scribe/notation tradition. It is evidence about the tested BDD abbreviation tradition, not historical shorthand in general. SX was validated for paired recovery questions; F2 alone cannot see the hidden expansion alignment.

### 9.2 Selective extraction

The frozen grid contains 20 operators in four classes: `ACROSTIC`, `TELESTIC`, `POSITIONAL_EXTRACTION`, and `PERIODIC_EXTRACTION`. They include first/last token or glyph, fixed within-group positions, and periodic selections. Random-subsequence and position-stratified matched nulls measure whether an observed trajectory is more than generic thinning.

FIRST/LAST line operators can emit at most one unit per source line and mechanically collapse the entire line-position family. Their repeatable F2 signature is therefore not acrostic-specific evidence. AX diagnostics lacked the required cross-corpus robustness gate: `AX_VALIDATED = NOT_SUPPORTED`. No hidden message was read or detected.

## 10. Confirmatory design and provenance

Task83r repeated the frozen Task83 protocol against the authoritative V2.1 target. The design binds: preregistration before target opening; a target firewall; a checksum-bound opening sentinel; frozen natural, simple-null, Fontana, shorthand, extraction, and matched-null portfolios; two VM transcriptions; direct F2 as primary; projections as separate secondary evidence; endpoint/trajectory separation; family-balanced aggregation; no missing-value imputation; no post-opening tuning; and frozen support/equifinality/identifiability gates ([Task83r design](../../../research/phase2/task83r/TASK83R_DESIGN.md)).

An endpoint asks whether a produced output lies near the VM target. A trajectory asks whether a transformation moves a known input toward the VM in direction and magnitude. “Closest” is descriptive; support additionally requires adequate multi-family coverage, relevant trajectory evidence where declared, and matched-null separation.

### 10.1 The Task83 provenance incident

Task83 found that the old freeze manifest expected IT2a prepared SHA-256 `3fb9531a…`, while the present prepared corpus and embedded Task79c artifacts used `10286ee7…`. Because the mismatch was discovered after target opening and the pre-audit had not transitively checked it, input and opening integrity failed. Task83 was marked `TASK83_EXPERIMENT_INVALID`; all comparisons were quarantined and are not evidence ([invalidation report](../../../research/phase2/task83/TASK83_INVALIDATION_REPORT.md)).

Task83a established that `3fb953…` first appeared as unrecoverable manual metadata, while documented preprocessing reproducibly yields `10286e…`. It also discovered a second defect: seeded Monte Carlo draws and floating-point reductions depended on unsorted Go-map traversal. Although CORE statuses and final Task79c verdicts did not change diagnostically, strict reproducibility failed and no refreeze was issued ([Task83a report](../../../research/phase2/task83a/TASK83A_REPORT.md)).

Task83b canonicalized PRNG job ordering, reductions, ordering-derived artifacts, and dominance traversal, then conservatively recomputed the affected closure. Independent RUN_A/B/C processes under `GOMAXPROCS=1`, `2`, and default produced 75 corresponding normative files with identical SHA-256 values. Ten of 66 metric rows changed; all 13 CORE statuses and final structural conclusions remained unchanged. The authoritative verifier exits zero and the marker is `FINGERPRINT_V2_DETERMINISTIC_SCIENTIFIC_REFROZEN` ([Task83b report](../../../research/phase2/task83b/TASK83B_REPORT.md)). This episode is scientific quality control, not evidence about the VM.

## 11. Confirmatory results

### 11.1 Endpoint proximity and coverage

Coverage-adjusted class-median direct distances (ZL3b / IT2a) were:

| Class | Distance | Confirmatory interpretation |
|---|---:|---|
| natural text | 0.674952 / 0.683933 | descriptively closest; `PARTIAL` |
| shorthand | 0.921503 / 0.930485 | endpoint resemblance, but trajectory/null failure; `DISFAVORED` |
| simple null | 1.110219 / 1.102676 | descriptive comparator only |
| extraction | 1.237889 / 1.246870 | weak directional/endpoint resemblance; `PARTIAL` |
| Fontana | 1.247039 / 1.239497 | endpoint resemblance only; `PARTIAL`, `LEVEL_1` |

Source: [Task83r report](../../../research/phase2/task83r/TASK83R_REPORT.md) and [direct endpoint table](../../../research/phase2/task83r/DIRECT_ENDPOINT_COMPARISON.tsv). Every class covers only **3/13 CORE metrics in one edit family**. Consequently the ordering does not adjudicate hierarchy, locus, folio, page, recto/verso, physical lines and boundaries, positional structure, local regimes, or 2D/manuscript organization.

### 11.2 Natural text

`NATURAL_TEXT_COMPATIBILITY = PARTIAL`. Ordinary natural text is closest in the available direct space, but it does not pass multi-family support gates. The result does not establish that the VM is natural language, that visible VM tokens are words, or that semantics exist. Its strongest support is relative descriptive proximity; its strongest counterevidence is inadequate direct coverage and failure to obtain confirmatory multi-family support.

### 11.3 External memory

`EXTERNAL_MEMORY_COMPATIBILITY = PARTIAL` and `EXTERNAL_MEMORY_EVIDENCE_LEVEL = LEVEL_1`. Some Fontana/generated endpoints occupy the same narrow edit-family region as the target. But there is no multi-family evidence, no demonstrated null-separated mechanism advantage, no valid frozen Fontana before/after trajectory, and no clear advantage over natural, shorthand, or extraction classes ([external-memory evidence](../../../research/phase2/task83r/EXTERNAL_MEMORY_EVIDENCE.tsv)). Phase II therefore neither confirms nor globally excludes generic external memory; the Fontana-specific historical relation remains speculative and unsupported.

### 11.4 Shorthand

`SHORTHAND_COMPATIBILITY = DISFAVORED`; `SHORTHAND_EVIDENCE_LEVEL = S0`. The combined real BDD abbreviation trajectory cosine is `0.279072 / 0.262481` (ZL3b / IT2a), versus approximately `0.99` for matched deletions. The minimum registered shorthand-null p-value is `0.25` ([trajectory](../../../research/phase2/task83r/SHORTHAND_TRAJECTORY_COMPARISON.tsv); [null comparison](../../../research/phase2/task83r/SHORTHAND_NULL_TARGET_COMPARISON.tsv)). The tested historical transformation did not move natural Latin toward the VM better than matched deletion controls. This verdict is limited to the BDD tradition.

### 11.5 Selective extraction and acrostic alternatives

`SELECTIVE_EXTRACTION_COMPATIBILITY = PARTIAL`; `EXTRACTION_EVIDENCE_LEVEL = A1`. Some operators align directionally, with maximum cosine about `0.80`, but no matched-null test reaches `p <= 0.05`; the minimum is `0.142857`. FIRST/LAST remains confounded by line collapse, and AX is excluded from scoring because it was not validated ([extraction evidence](../../../research/phase2/task83r/EXTRACTION_EVIDENCE.tsv); [null comparison](../../../research/phase2/task83r/EXTRACTION_NULL_TARGET_COMPARISON.tsv)). Thus Phase II does not detect an acrostic, telestich, or hidden channel.

## 12. Identifiability

`BEST_SUPPORTED_CLASS = INCONCLUSIVE` and `MECHANISM_IDENTIFICATION_FROM_F2 = NOT_IDENTIFIABLE`.

Several classes can resemble the target in a three-dimensional edit projection. That is not evidence that they are scientifically equivalent: equivalence would require adequately covered, supported classes and discriminating pairwise evidence. Here no class passes its own support gate, all pairwise advantages are `NO_CLEAR_ADVANTAGE`, and supported-class equifinality cannot be tested. Non-identifiability therefore means insufficient informative overlap between what F2 measures and what the models generate—not equality of mechanisms and not absence of structure.

The central bottleneck is structural. Fingerprint V2 measures 13 CORE properties across edit, line, boundary, locus, folio, hierarchy, and 2D-lite families. The synthetic portfolios honestly generate only a token-stream intersection of 3 CORE edit metrics. Adding projection metrics would import assembler-created line semantics, so Task83r correctly kept them secondary. Until competing models generate manuscript-scale organization or the direct intersection expands across multiple families, endpoint rankings cannot identify origin class.

## 13. What Phase II established

- Fingerprint V2 substantially expands the measured structural description of the VM.
- The qualitative statuses of all 13 CORE metrics are stable across two independent transcriptions; the deterministic reconstruction is byte-identical across RUN_A/B/C.
- Historically grounded and generalized external-memory mechanisms can be formally separated by state, convention, path, context, association, information loss, and recovery.
- Historical abbreviation and selective extraction produce measurable, reproducible transformation signatures.
- The tested BDD abbreviation trajectory does not approach the VM better than matched deletion nulls.
- Natural text is descriptively closest within the restricted direct intersection.
- External memory retains only structural `LEVEL_1` compatibility.
- The present 3/13 one-family direct intersection is insufficient for mechanism identification.

## 14. What Phase II did not establish

Phase II did **not** establish that the VM was created by Fontana; is external memory; is natural language; is a cipher; is shorthand; contains an acrostic or hidden channel; implements a formal language; or is meaningless generation. It also did not globally exclude most of these classes. Autonomous-transformation compatibility is only partial for tested portfolios; formal-language and meaningless-generation alternatives were not directly resolved.

## 15. Relation to Phase I

Phase II confirmed that constrained structure is real and refined the earlier edit-family interpretation with a grammar-bounded null. It confirmed on synthetic systems that autonomous recovery, knowledge-assisted recovery, ambiguity, and irreversible loss are distinct. It refined “context may matter” into explicit state/convention/path/association ablations, but did not demonstrate that VM interpretation actually depends on such knowledge. It left the formal-language and meaningful-versus-message-free distinction unresolved. Detailed statuses and reasoning appear in [PHASE1_TO_PHASE2_CONCLUSION_MAP.tsv](PHASE1_TO_PHASE2_CONCLUSION_MAP.tsv) and [PHASE2_HYPOTHESIS_STATUS.tsv](PHASE2_HYPOTHESIS_STATUS.tsv).

## 16. Methods and statistical discipline

Corpora were prepared with recorded hashes and frozen normalization; ZL3b and independently sourced IT2a were aligned without interpolating unmatched loci. F2 extraction combined edit graphs, lexical-rule statistics, grammar-bounded nulls, transition/cross-scale estimators, line/boundary metrics, metadata-conditioned locus/folio/page metrics, hierarchical variance shares, permutation tests, and folio bootstraps. Controls included natural prose, historical BDD abbreviation, procedural text, preservation-matched shuffles, C-GRAMMAR, matched deletions, and matched subsequences.

Mechanisms were frozen before VM target access, executed with derived independent seeds, scaled across corpora/scales/replicates, and audited for input dependence, collisions, recovery, and knowledge ablation. Confirmatory distances used only the common direct metrics, pre-target natural-control normalization, within-family aggregation, family balancing, explicit missingness, and both VM transcriptions. Endpoint, trajectory, matched-null, and projection evidence remained separate.

The governing safeguards were target blindness, preregistration/freeze, held-out controls, no post-target tuning, no missing-value imputation, multi-seed and cross-corpus checks, matched nulls, transcription robustness, multi-family support gates, transitive checksum provenance, and a strict distinction between descriptive proximity and confirmatory support.

## 17. Limitations

1. Only 3/13 CORE metrics are directly comparable, all in one edit family.
2. No valid frozen Fontana before/after trajectory exists.
3. Historical shorthand is represented by one BDD manuscript/scribe/tradition.
4. AX did not pass validation, and FIRST/LAST extraction is line-collapse confounded.
5. Synthetic outputs lack real manuscript hierarchy, physical lines, pages, folios, loci, recto/verso relations, and genuine 2D organization.
6. Assembler projections cannot substitute for those physical structures.
7. No semantic information or validated plaintext is available.
8. The historical relationship between Fontana and the VM remains speculative.
9. Statistical compatibility cannot establish authorship, and fingerprint similarity cannot establish decoding.
10. Formal meaningful language and constrained meaningless generation remain observationally unresolved under current tests.

## 18. Open questions and Phase III boundary

The scientific questions that follow concern formal grammar, information residual after conditioning on grammar, meaningful versus message-free formal systems, manuscript-scale generation, and multi-family identifiability. They are stated without a task roadmap in [PHASE_III_OPEN_PROBLEMS.md](PHASE_III_OPEN_PROBLEMS.md).

## 19. Conclusion

Phase II did not identify how the Voynich Manuscript was generated. External-memory mechanisms are operationally coherent and historically motivated, but their VM evidence stops at weak, one-family endpoint compatibility (`LEVEL_1`). The tested BDD abbreviation tradition is disfavored; selective extraction retains weak compatibility without a detected hidden channel. Natural text is descriptively closest, but also unsupported as an identified class.

The authoritative scientific result is therefore not equivalence among hypotheses. It is **non-identifiability under a narrow intersection**: the models generate only 3 of 13 CORE dimensions, while the most distinctive manuscript-scale properties remain outside direct comparison. Phase II has converted a loose historical analogy into testable mechanism classes, measured their information and recovery behavior, rejected overclaims, and located the exact evidential bottleneck that the next phase must address.

## 20. Reproducibility and artifacts

The curated authoritative index is [PHASE2_ARTIFACT_INDEX.md](PHASE2_ARTIFACT_INDEX.md), bibliography is [PHASE2_REFERENCES.md](PHASE2_REFERENCES.md), and claim-level audit is [PHASE2_CLAIM_TRACEABILITY.tsv](PHASE2_CLAIM_TRACEABILITY.tsv). Task84 validation and output inventory are recorded in [TASK84_REPORT.md](TASK84_REPORT.md) and `TASK84_RESULTS_MANIFEST.json`.
