# Scientific Inference Safety Audit (Section I)

The task requires verifying that the (future) production pipeline cannot: publish a best system
before null controls complete; publish pairs when the synthetic gate fails; declare cross-page
signal from joint-only results; use held-out data for model selection; change thresholds after a
run starts; assert a decipherment claim; and correctly emits `BLOCKED` on timeout or an incomplete
null pipeline.

## Finding: No Production Pipeline Exists Yet to Test

`restricted_dictionary_bruteforce_v3_design` is explicitly a **design/preparation** package — there
is no production runner, no publication/reporting code path, and no null-control orchestration
script anywhere in `scripts/`. `NULL_FEASIBILITY_REPORT.md` is a resource *projection*, not an
implementation. Consequently, Section I cannot be satisfied by running an automated test suite
against real code (none exists), only by a design-level review of what such a pipeline would need
to enforce, and by checking that nothing in the *current* codebase already violates these
invariants in a way that would carry forward.

## Design-Level Review (against current code + design documents)

| Invariant | Status | Evidence |
|---|---|---|
| No best-system publication before null controls complete | **N/A — not yet implemented** | No production runner exists. `NULL_FEASIBILITY_REPORT.md` correctly states production null runs were NOT executed, and `REAL_DATA_SEARCH_AUTHORIZED=NO` is honored throughout (grep-verified: no script in `scripts/` opens `TARGET_SCOPE.tsv` except `build_structural_manifest.py`, which only computes aggregates). |
| No pair publication on synthetic-gate failure | **AT RISK** | The synthetic gate itself is currently fabricated (F001) — a "gate" that never genuinely ran cannot meaningfully block anything. Until F001 is fixed, this invariant has no real gate to attach to. |
| No cross-page signal claim from joint-only result | **NOT YET DESIGNABLE** | `PER_PAGE_CAPACITY_1` treats pages independently at the matching layer (confirmed, B5), which is a *necessary* precondition for eventually being able to report per-page vs joint results separately, but no reporting/aggregation code exists yet to check for the anti-pattern itself. |
| No held-out use in model/table selection | **PASS (in existing code)** | `evaluate_table`/`maximum_bipartite_matching` have no concept of held-out data; nothing in `search_bb.py`/`search_cp.py`/`search_heuristic.py` partitions labels into train/held-out or uses such a partition during search. Held-out scoring, where it appears at all (design's fabricated Suite 1/2 "held_out_accuracy" column), is a separate post-hoc metric column, not a search input. |
| No threshold changes after run start | **PASS (trivially, no runs exist)** | All thresholds (`>=0.80`, `>=0.70`, etc.) are stated in `V3_DESIGN_REPORT.md`/task spec before any genuine run; since F001 shows no genuine run occurred, thresholds were never actually tested against real data to begin with, so they were not moved post-hoc either — this is a degenerate pass. |
| No decipherment claim | **PASS** | Explicitly disclaimed throughout (`REAL_DATA_SEARCH_AUTHORIZED=NO`, "neither asserts nor publishes any star decipherment" in `VALIDATION_REPORT.md` §4.5) — consistent with what the code actually does (nothing on real data). |
| `BLOCKED` on timeout / incomplete null pipeline | **NOT IMPLEMENTED** | No status-emission code exists. `search_bb.py`/`search_cp.py` correctly set `timed_out`/`global_optimum_guaranteed` flags at the solver level (modulo the F003 false-certificate bug in CP_SAT), but nothing aggregates these into a pipeline-level `BLOCKED` status. |

## Required Before Production Authorization

Because no production pipeline exists, `SCIENTIFIC_SAFETY_TESTS` cannot be marked `PASS` in the
sense of "verified by execution" — it can only be marked as a reviewed design requirement list.
Given F001 (fabricated gate results) and F009 (non-existent HYBRID wiring), the two invariants that
matter most in practice right now — "don't publish on synthetic-gate failure" and "don't overclaim
architecture" — are **already violated at the design-reporting level**, before any production code
is even written: the design package already reports gates as `PASS` that were never genuinely
computed. This is the same failure Section I is trying to prevent, occurring one level upstream of
where the task expected to find it.

```text
SCIENTIFIC_SAFETY_DESIGN_REVIEW = INCOMPLETE_NO_PIPELINE_TO_TEST
PRE_EXISTING_GATE_INTEGRITY_VIOLATION = YES (see F001, F002, F009)
```
