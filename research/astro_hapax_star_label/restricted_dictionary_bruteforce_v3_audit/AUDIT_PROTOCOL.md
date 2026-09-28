# Independent Pre-Production Audit Protocol
## Target: `restricted_dictionary_bruteforce_v3_design`

## 1. Mandate

This audit was commissioned by `tasks_other/preproduction_audit_m3.md` to independently attempt to
**refute** the claimed `LEXICON_GATE=PASS` and `SEARCH_ARCHITECTURE_GATE=PASS` dispositions of
`restricted_dictionary_bruteforce_v3_design`, not to formally confirm them. The audit is barred from
any real-data dictionary search: no similarity, alignment, or match score was computed against the
57 real f68r1/f68r2 EVA labels at any point, and no such computation was needed to reach the
conclusions below.

## 2. Isolation Discipline

All independent code re-execution was performed inside a **read-only reproduction sandbox**
(`restricted_dictionary_bruteforce_v3_audit/repro_sandbox/`), which contains copies of the design
package's own `scripts/*.py`, `HISTORICAL_STAR_LEXICON.tsv`, and `REAL_SCOPE_STRUCTURAL_MANIFEST.json`.
Every script that writes output resolves `BASE_DIR` relative to its own file location
(`Path(__file__).resolve().parent.parent`), so running the copies inside the sandbox never touches
the frozen design package. This was verified before and after every execution via
`sha256sum -c` against `checkpoints/BASELINE_restricted_dictionary_bruteforce_v3_design.sha256`
(see `UPSTREAM_INTEGRITY_REPORT.md`).

All genuinely new audit code (adversarial instance construction, the independent bipartite-matching
cross-check, the fresh sealed hidden benchmark, and the resource pilot) lives in
`restricted_dictionary_bruteforce_v3_audit/scripts/audit_verification.py` and imports the frozen
engine/solver classes from the sandbox rather than reimplementing them, so that findings about
the *actual* certified code are being reported, not findings about a reinterpretation of it.

## 3. Method Summary (mapped to task sections)

| Task section | What was actually done |
|---|---|
| A1/A2 | SHA-256 baseline of v1/v2/v3-design before and after audit; grep-based dependency/leakage audit; `DATA_ACCESS_GRAPH.json` |
| B1/B4/B5 | Programmatic structural audit of all 307 lexicon rows/94 identities (`LEXICON_AUDIT.tsv`); re-derivation of identity/attestation counts; cross-identity form-collision detection |
| B2 | Stratified spot-check (9 `MEDIUM`-confidence rows + 12 random rows + the 6 rows behind the 2 riskiest cross-identity collisions = 27/307 rows, ~8.8%) verified against independent secondary scholarship via web search (`PROVENANCE_VERIFICATION.tsv`). This is **below** the mandated ≥30% stratified minimum for the non-high-risk population — see Finding F011 — and is disclosed as an incomplete-coverage limitation, not presented as a full B2 clearance. |
| B3 | Grep of `build_historical_lexicon.py` for any reference to `TARGET_SCOPE`, `f68r`, or `EVA_ALPHABET`: none found |
| C1/C2/C3 | Grep for truth/planted/hidden references inside `engine.py`/`search_bb.py`/`search_cp.py`/`search_heuristic.py`: none found. Seed-range disjointness verified for the fresh hidden benchmark (new seeds 9001+, 70000+, disjoint from the design's 101-120/999/1000-1425/42000+ ranges) |
| C4 | Structural distributions of a freshly generated synthetic dataset compared against `REAL_SCOPE_STRUCTURAL_MANIFEST.json` aggregates only (no real forms) |
| D1/D2/D3 | Confirmed all benchmark scripts call the single shared `engine.evaluate_table`; independent from-scratch Hopcroft-Karp-style bipartite matcher cross-checked against `engine.maximum_bipartite_matching` on 200 randomized instances × 2 capacity policies = 400 trials |
| E1/E2/E3/E5 | New adversarial "k-fragment-combination" instances (k=2..6) with brute-force/'independently-validated-B&B' ground truth; 100-permutation order-invariance test of the heuristic |
| E4 | Grep across all scripts for any code path wiring `DirectionalHeuristicSolver` output into `BranchAndBoundSolver`/`CPSATSolver` as a restricted-neighborhood verifier: none found |
| F | Freshly generated sealed synthetic datasets on **new, previously-unused seeds**, scored with the real `DirectionalHeuristicSolver`/`CPSATSolver`/`BranchAndBoundSolver` classes (not re-typed constants) |
| G | Independent re-execution of `analyze_reachability.py` inside the sandbox; byte-identical diff against the frozen `REACHABILITY_ANALYSIS.tsv`; exhaustive scan of that file for the headline `0.9625` figure |
| H1/H2 | Real timed pilot of synthetic null-style replicas (heuristic solver on freshly generated 57-label datasets), wall-time/peak-memory measured via `tracemalloc`, extrapolated to 60,000 replicas with min/mean/p95/max statistics |
| I | Manual code-path audit against the 7 required scientific-safety invariants (`SCIENTIFIC_SAFETY_TESTS.md`) — no production pipeline exists yet to unit-test, so this is a design-level review, not an executable test suite |

## 4. Deviations From the Literal Task Specification (Disclosed)

The task specification asks for scales (≥1,000 sealed synthetic null runs, ≥20 seeds per every one
of dozens of table-size × noise × mode combinations, a full 6-family/10,000-replicate-per-family
null pipeline) that are individually reasonable but collectively amount to tens of thousands of
solver invocations — well beyond what a single audit session can execute and independently
re-verify line-by-line. Rather than fabricate compliance with those exact scales (the central
defect this audit was commissioned to catch in the target package), the audit:
- ran real, reduced-but-honestly-reported sample sizes (documented in each results file's header/metadata),
- explicitly reports every reduction as a scope limitation rather than silently rounding up, and
- recommends the specific larger runs that remain outstanding before any future `PASS` decision.

This is itself disclosed as a limitation of this audit pass, not a clean bill of health.
