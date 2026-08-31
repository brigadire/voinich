# Level-C v1 identifiability diagnosis and v2 design

## Decision

`LEVEL_C_V2_FEASIBILITY=FEASIBLE_WITH_LIMITATIONS`  
`LEVEL_C_V2_PROTOCOL_FROZEN=true`  
`LEVEL_C_V2_TEST_REGISTRY_FROZEN=true`  
`LEVEL_C_V2_PRODUCTION_RUN_AUTHORIZED=true`  
`LEVEL_C_V2_PRODUCTION_RUN_EXECUTED=false`

Q1. The 24-case collapse is structural: only 24 of 227 pages have all 15 descriptors observed; they are concentrated in Biological (19) and Zodiac (5). Textual join loss is zero.

Q2. Missingness is descriptor- and page-structured. The 32 partially annotated units are concentrated in panel/foldout-like page types and affect several descriptors simultaneously; section survival is uneven. No feature was selected by v1 association.

Q3. Textual fingerprint metrics are available for all 227 joined pages, so textual missingness materially contributes nothing to the collapse.

Q4. The saturated visual-plus-confounder model is not estimable on the 24 complete-vector cases: effective N is too small, sections/quires have insufficient overlap, and descriptor categories are separated by page type/section.

Q5. Metadata-only M0–M4 are estimable on the full join, but visual increments are not reliably estimable in the complete-vector subset. This is reported as a limitation rather than a post-hoc control choice.

Q6. Yes, descriptor-wise and family-wise tests can use partial observations. A global test is feasible with a symmetric normalized distance and a minimum-overlap rule; no imputation is permitted.

Q7. v2 retains 227 textual pages, descriptor-specific N up to 227, and all eight broad sections. The proposed global minimum-overlap rule will have a smaller effective N and must report it.

Q8. Physical-leaf block permutation spaces range from 2 to 64 blocks by section; exact spaces include 2, 24, and 720/40320 where small, with larger strata requiring seeded Monte Carlo.

Q9. Production requires exact enumeration for small spaces and 10,000 seeded permutations otherwise. The v1 100 draws are not reused.

Q10. A scientifically useful v2 is identifiable only with limitations: missing-aware distances and restricted exchangeability are feasible, but residual confounding prevents a claim of fully adjusted robust coupling.

Invalid alternatives explicitly rejected: modal imputation, `NOT_APPLICABLE`→`ABSENT`, text-driven visual imputation, v1-p-value feature selection, and PCA on 24 cases.
