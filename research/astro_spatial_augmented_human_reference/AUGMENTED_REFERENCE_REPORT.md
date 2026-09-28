# Augmented human reference report

## Result

The immutable augmented reference contains 557 confirmed objects: 329 STAR and 228 LABEL. It combines 520 prior ACCEPT/MODIFY final geometries with 37 explicitly confirmed human completeness additions (8 STAR, 29 LABEL). Prior REJECT/UNCERTAIN/NOT_REVIEWED and no unresolved additions enter the denominator. All 37 additions were reconciled as distinct real objects; 28 proximity/overlap records affect 23 objects but transfer no AI provenance and change no old outcome.

## Assisted human-reference recall

| Type | AI1 | AI2 | AI union | Both / AI1-only / AI2-only / human-only |
|---|---:|---:|---:|---:|
| STAR | 50.8% (167/329; Wilson 95% 45.4%–56.1%) | 95.7% (315/329; Wilson 95% 93.0%–97.4%) | 97.6% (321/329; Wilson 95% 95.3%–98.8%) | 161 (48.9%) / 6 (1.8%) / 154 (46.8%) / 8 (2.4%) |
| LABEL | 38.6% (88/228; Wilson 95% 32.5%–45.1%) | 86.0% (196/228; Wilson 95% 80.9%–89.9%) | 87.3% (199/228; Wilson 95% 82.3%–91.0%) | 85 (37.3%) / 3 (1.3%) / 111 (48.7%) / 29 (12.7%) |

AI2's observed advantage persists after augmentation. Relative to the pre-completeness confirmed denominator, AI1 changes from 167/321 (52.0%) to 167/329 (50.8%) for STAR and 88/199 (44.2%) to 88/228 (38.6%) for LABEL; AI2 changes from 315/321 (98.1%) to 315/329 (95.7%) and 196/199 (98.5%) to 196/228 (86.0%). Union common misses are 8/329 STAR and 29/228 LABEL, so the common-miss frequency is appreciably higher for LABEL. AI1's marginal contribution beyond AI2 is 6/329 STAR and 3/228 LABEL; AI2's beyond AI1 is 154/329 and 111/228. Calibration, production, relation-scope, panel rows, Wilson intervals and panel bootstrap intervals are in AI_ASSISTED_RECALL_SUMMARY.tsv. Few-panel bootstrap intervals are descriptive and cannot support manuscript-wide generalization.

## Human additions

All-page human additions number 37, not 17. Seventeen are on f68r1/f68r2/f68r3 and 20 are outside relation scope. Sixteen relation-scope additions are grouped, all in two-member newly created groups. Human-only objects occur on f67r2 (14), f67v1 (1), f68r2 (9), f68r3 (8), and f68v2 (5). HNEW_STAR_f68r2_DC27A556209F27B3 remains an existing, fully reviewed object with `NO_VISUAL_GROUP_ASSIGNED`; no negative LABEL pairs are generated.

## Relation groups 3G1

There are 64 groups: 29 f68r1, 24 f68r2, 11 f68r3. Sixty-three have two members; one f68r3 hyperedge has one LABEL and seven STAR and remains one entity. Membership is empirically non-overlapping: 134 grouped and 122 ungrouped objects within the 256-object relation scope. Exact membership comparison finds 45 unchanged groups, one initial group extended from four to eight members with existing reference STAR, and 18 wholly new groups. Thus `45 unchanged + 19 new/modified = 64`; the 46th initial group is the extended group and is not double counted. Forty objects first receive a group, none lose grouped membership, and four retained members participate in the extended membership.

## Attachment-v2 crosswalk

Attachment-v2 remains frozen pair-level review; 3G1 is full-page group review. The crosswalk records co-membership and representation differences only. Multi-member co-membership is `ENDPOINTS_CO_MEMBER_SAME_3G1_GROUP`, not seven pair claims. No pairwise precision/recall is calculated, no group is expanded, and no absence of grouping becomes an UNASSIGNED decision.

## Scope limitation

The reviewer was assisted by an existing overlay. This is not an exhaustive independent blind gold standard and not absolute recall. Relation structure applies only to f68r1/f68r2/f68r3. No lexical, semantic or astronomical interpretation is made.

```text
AUGMENTED_HUMAN_REFERENCE_STATUS=COMPLETE
HUMAN_ADDED_OBJECTS_RECONCILED=YES
ASSISTED_RECALL_AI1_CALCULATED=YES
ASSISTED_RECALL_AI2_CALCULATED=YES
ASSISTED_RECALL_AI_UNION_CALCULATED=YES
RELATION_GROUP_PROTOCOL=3G1
FINAL_GROUPS=64
MULTIMEMBER_GROUPS_PRESERVED=YES
CARTESIAN_EDGES_INFERRED=NO
UNGROUPED_OBJECTS_PRESERVED=YES
FROZEN_ATTACHMENT_V2_CHANGED=NO
PREVIOUS_REPORTS_OVERWRITTEN=NO
RESULTS_REPRODUCIBLE=YES
```
