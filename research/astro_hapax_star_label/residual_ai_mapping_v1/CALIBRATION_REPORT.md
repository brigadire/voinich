# Calibration and hidden evaluation report

The frozen four-pass procedure completed on 14 calibration-visible and seven evaluation-hidden f68r1 cases. Answers were joined only after the independent outputs, shuffled ranking, and adjudication had been frozen by content hash. The two partitions were processed in one unchanged per-pass batch rather than using visible-answer feedback to tune prompts between phases. Thus hidden answers remained sealed and no calibration overfitting was possible, but this run does not claim a separate feedback/tuning stage.

Calibration-visible adjudicated exact occurrence accuracy was 0/14. Hidden adjudicated exact occurrence accuracy was **1/7** and exact token-sequence accuracy was 1/7. Alignment B base-order top-k recall on hidden cases was 3/7; base top-1 recall was 3/7. Candidate-order stability was **0/7**. Visual A contained the exact known reading in its primary/alternatives for 2/7; Visual C did so for 0/7; their primary strings exactly agreed in 0/7.

All seven known cases truly have a frozen occurrence, so adjudicator abstentions count as incorrect abstentions for this diagnostic. Abstention decision accuracy was 1/7; 6/7 cases abstained. There were zero forced answers, 0 false HIGH-confidence selections, and 0 confident references to nonexistent candidates. Hidden confidence calibration (correct/total) was LOW=0/6, MEDIUM=1/1.

Negative controls produced the required abstention in 21/21 cases (SAME_PAGE_CORRECT_OMITTED=7/7, VISUALLY_UNRELATED_RESTRICTED_SET=7/7, WRONG_PAGE_CANDIDATES=7/7). This shows the pipeline can abstain on adverse sets, but it does not rescue the failed positive-case order stability or hidden accuracy.

The candidate package stored the complete 268-occurrence registered universe with a panel field; the frozen prompt restricted ranking to the same panel, giving 69 admissible f68r1 candidates for every evaluation case. Every selected/ranked candidate was checked to be f68r1.

The mandatory gate failed before the repeated-run condition: accuracy 1/7 < 6/7 and stability 0/7 < 6/7. The threshold was not changed, abstentions were not replaced, and a repeated expensive frozen run was not initiated after decisive first-run failure.

The evaluation is internal to f68r1; it provides no accuracy estimate for f68r2 or f68r3.

```text
RESIDUAL_AI_MAPPING_RUN_AUTHORIZED=NO
```
