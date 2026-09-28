# f68r1 structural alignment report

The fixed canonical page-center features and ten predefined traversal models were evaluated without group, hapax, frequency, lexicon, or human-added metadata. The best pre-LOAO model by the frozen global score was `SCAN_TB_LR` (Spearman 0.962562551, mean normalized residual 0.099017385).

Leave-one-anchor-out exact accuracy was 4/21 and top-3 accuracy 8/21. The anchor-rank permutation p-value for the primary model was 0.000099990. Stability rows are invariant only under the explicitly recorded deterministic rank recomputation. No individual anchor exception was used.

Gate S1 requires exact ≥16/21, top-3 ≥19/21, permutation p<0.01, and robustness. It is **FAIL**. Because the fixed LOAO model does not meet the thresholds, no f68r1 production structural alignment file is emitted. f68r2 and f68r3 have no anchors and are independently unauthorized under S2; success on f68r1 would not have transferred automatically.

The 21 legacy mappings remain unchanged as `LEGACY_CONFIRMED`; the other 71 LABELs are explicit `UNRESOLVED_GATE_S1_FAILED`. No local visual AI adjudication was run because there were no ≤3-candidate structural alternatives after the failed gate.

```text
STRUCTURAL_ALIGNMENT_F68R1_AUTHORIZED=NO
STRUCTURAL_ALIGNMENT_F68R2_AUTHORIZED=NO
STRUCTURAL_ALIGNMENT_F68R3_AUTHORIZED=NO
```
