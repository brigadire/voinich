# f68r3 limited eligibility — DRAFT, not an approved mask

F68R3_ATTACHMENT_SCOPE=AWAITING_REVIEWER_CONFIRMATION

Previous approved attachment-v2 materials say “all confirmed-endpoint source pairs” on
f68r3, no right filter. That is not an explicit limited region/object rule demanded by
this new task; neither historical positive decisions nor guessed semantics define it.
Object completeness is ready regardless; no f68r3 pair enters current attachment dry run.

Visual map: figures/F68R3_ELIGIBILITY_MAP.png. Neutral IDs S001… / L001… correspond to
F68R3_WHITELIST_CATALOG.tsv, with unchanged canonical IDs for a future whitelist.
Navigation zones T01…T12 are only location references, not selected eligible regions.
The map shows confirmed final geometry only; no source identity or historical relations.

Reviewer can explicitly choose one reproducible option:

1. WHITELIST: enumerate both permitted STAR IDs and LABEL IDs from neutral catalog.
   Future new objects require an explicit whitelist amendment; not added automatically.
2. POLYGON_CENTER_MASK: list polygon vertices in original image pixel coordinates;
   endpoint bbox center must lie inside/on boundary. This geometric convention is explicit
   and allows future confirmed additions inside region without semantic selection.
3. Zones: specify T-zone union; convert it into explicit polygon mask(s) or ID whitelist
   before approval. A textual “some right/lower stars” is not sufficient.

No option is preselected. Approval must include reviewer, UTC timestamp, rule kind/values,
status REVIEWER_CONFIRMED_AND_FROZEN and rule_sha256 over sorted compact JSON excluding
rule_sha256. Template F68R3_SCOPE_TEMPLATE.json is incomplete/non-operational; never use
empty mask/whitelist as whole-panel permission. Polygon/whitelist is included as f68r3_rule
in a newly confirmed ATTACHMENT_ELIGIBILITY.json revision, not modifying old human protocols.
