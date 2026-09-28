# Human adjudication protocol — frozen v1.2

Status: **frozen for production** after completed STAR, LABEL, and ring follow-up calibration. The authoritative SHA-256 is stored in `HUMAN_ADJUDICATION_PROTOCOL_FROZEN.sha256` and `manifest.json`.

## Scope and order

1. H-Cal: blind H1 STAR and LABEL review on `f68r1`, `f68r3`, `f68v2`.
2. H-High: all HIGH candidates after this production freeze.
3. H-Consensus-QC: deterministic 20% sample of candidates supported by both blind sources.
4. H-Medium: only if H-High/QC shows it is necessary.
5. RELATION: only after human STAR and LABEL layers are frozen.

H1 exposes images, neutral candidate IDs, provisional shapes, decisions, and human confidence. It does not expose annotator/model identity, support count, informative priority, transcription, lexical statistics, semantic mappings, M3 output, or astronomical guesses. H2 may expose only neutral support count and priority after H1 is frozen.

## Decisions

Allowed decisions are `ACCEPT`, `REJECT`, `MODIFY`, and `UNCERTAIN`; LABEL additionally permits `SPLIT` and `MERGE`. `UNCERTAIN` is unresolved and must not be coerced into yes/no. AI agreement changes queue priority only and never implies acceptance or ground truth.

For LABEL, local/radial runs use provisional rotated rectangles derived from the frozen orientation fields. Complete circular writing paths use ellipses so the empty diagram centre is not presented as label area. An ellipse denotes a ring path; `SPLIT` is used if calibration decides that local arcs are separate physical units. On `MODIFY`, the reviewer corrects geometry; center is recomputed from the final shape. Original A/AI1/AI2 records and all source IDs remain immutable.

## Frozen operational boundaries

### STAR_OBJECT

- Faint marks are separate objects when their visible form is recognizably star-shaped.
- Partial edge stars are accepted.
- Overlapping stars are always separate objects.
- Star versus decoration is decided by visible star form only.
- If a production bbox/text collision cannot be resolved visually, use `UNCERTAIN`; calibration contained no such case.

### LABEL

- Several words on one continuous line form one physical LABEL run.
- Inclined and local/radial runs follow the same physical-run rule and use rotated rectangles.
- Each complete continuous circular writing path is one LABEL and uses an ellipse display shape. The focused calibration accepted all three ring candidates.
- A candidate may not be moved to an unrelated text location. If no text exists at its location, use `REJECT` or `UNCERTAIN`.
- Do not create candidates for otherwise unproposed text during adjudication.
- `SPLIT` means one candidate contains multiple disconnected physical runs; `MERGE` means adjacent candidates form one uninterrupted physical run. If that cannot be decided visually, use `UNCERTAIN`.

Substantive changes to these boundaries require a new protocol version and repeated affected calibration; historical annotations are never rewritten.

## Data handling

Do not transcribe. Do not introduce EVA, Stolfi, ZL3b, token/frequency/hapax fields, semantic identity, or astronomical interpretation. Reviewer IDs must be pseudonymous. UTC ISO-8601 timestamps are required. CVAT shapes must not be deleted; rejected candidates retain their shape and `REJECT` decision.

Round-trip coordinate tolerance is 0.001 pixel. HUMAN outputs reference immutable candidate IDs; provenance stays in the non-UI candidate TSVs.
