# Краткий результат AI1 / AI2 / Human

Основной scope: production reviewed candidate union, не полный census страниц.

| Тип | AI | Confirmation strict | Cluster 95% CI | MODIFY среди confirmed | Условное покрытие confirmed union |
|---|---|---|---|---|---|
| STAR | AI1 | 61.7% (92/149; 95% Wilson 53.7–69.2%) | 49.3–81.2% | 100.0% (92/92; 95% Wilson 96.0–100.0%) | 51.4% (92/179; 95% Wilson 44.1–58.6%) |
| STAR | AI2 | 100.0% (175/175; 95% Wilson 97.9–100.0%) | 100.0–100.0% | 65.1% (114/175; 95% Wilson 57.8–71.8%) | 97.8% (175/179; 95% Wilson 94.4–99.1%) |
| LABEL | AI1 | 44.9% (53/118; 95% Wilson 36.2–53.9%) | 34.0–55.4% | 100.0% (53/53; 95% Wilson 93.2–100.0%) | 41.1% (53/129; 95% Wilson 33.0–49.7%) |
| LABEL | AI2 | 95.5% (126/132; 95% Wilson 90.4–97.9%) | 92.2–99.0% | 79.4% (100/126; 95% Wilson 71.5–85.5%) | 97.7% (126/129; 95% Wilson 93.4–99.2%) |

STAR: AI2 имеет более высокую наблюдаемую conditional confirmation rate в production (92/149 AI1 против 175/175 AI2). Это не оценка абсолютной page-level precision.

STAR, both против AI1_ONLY: risk difference 0.934, cluster CI [0.870,1.000], n=88/61. Направление поддерживается этим ограниченным набором панелей.

STAR, both против AI2_ONLY: risk difference 0.000, cluster CI [0.000,0.000], n=88/87. Разница недостаточно определена между панелями.

LABEL: AI2 имеет более высокую наблюдаемую conditional confirmation rate в production (53/118 AI1 против 126/132 AI2). Это не оценка абсолютной page-level precision.

LABEL, both против AI1_ONLY: risk difference 0.897, cluster CI [0.821,1.000], n=53/65. Направление поддерживается этим ограниченным набором панелей.

LABEL, both против AI2_ONLY: risk difference -0.019, cluster CI [-0.105,0.064], n=53/79. Разница недостаточно определена между панелями.

Совместная поддержка не даёт обнаружимого преимущества confirmation над
AI2-only: для STAR обе группы имеют 100% strict confirmation, для LABEL both=50/53
против AI2-only=76/79 (cluster CI разницы включает ноль). Это не доказательство
эквивалентности. По сравнению с AI1-only преимущество большое.

Гипотеза меньшей human correction у both-AI не поддерживается относительно AI2-only:
в production все confirmed both-AI требуют MODIFY (88/88 STAR, 50/50 LABEL), тогда
как AI2-only — 26/87 STAR и 50/76 LABEL. Candidate geometry из нескольких источников
была усреднённым UI helper; поэтому этот эффект относится к необходимости исправить
именно frozen candidate display, а не доказывает ухудшение исходной AI2 geometry от
наличия второго источника. Singleton AI1 confirmed groups малы (4 STAR/3 LABEL).
Boundary bootstrap CI [0,0] для равных perfect fractions — ограничение эмпирической
выборки, не нулевая неопределённость; Wilson интервалы групп остаются ненулевой ширины.

STAR: направление сравнения confirmation AI1/AI2 по 6 estimable production policies: AI2. Это описательная устойчивость; не доказательство причинности или эквивалентности.

STAR, BOTH_AI−AI1_ONLY: sensitivity point-effect range 0.923…0.934 across 6 estimable policies.

STAR, BOTH_AI−AI2_ONLY: sensitivity point-effect range -0.011…0.000 across 6 estimable policies.

LABEL: направление сравнения confirmation AI1/AI2 по 6 estimable production policies: AI2. Это описательная устойчивость; не доказательство причинности или эквивалентности.

LABEL, BOTH_AI−AI1_ONLY: sensitivity point-effect range 0.887…0.897 across 6 estimable policies.

LABEL, BOTH_AI−AI2_ONLY: sensitivity point-effect range -0.029…-0.019 across 6 estimable policies.

AI-consensus допустим как дополнительный сигнал human-review priority, но не automatic
acceptance: both-source candidates тоже отвергаются и требуют геометрической коррекции.
LABEL geometry сравнивается с явно обозначенными oriented-box/ellipse-envelope proxies;
неоднозначность исходных side-length conventions ограничивает geometry ranking.
Мало panel clusters; Fisher p — дополнительный object-independence показатель.

Attachment-v2: human-only descriptive graph, 53 links/51 LABEL/51 STAR; 267 explicit
decisions, 43 unresolved; 57 filtered +351 endpoint-excluded пары не являются negatives.
Проверены source provenance/checksums; calibration/uncertainty/not-reviewed не смешаны
с основными production strict metrics. Для абсолютного recall нужен независимый blind
manual census с разрешением новых объектов на holdout panels, не ещё одна AI-derived очередь.

Полный отчёт: THREE_WAY_COMPARISON_REPORT.md; данные/CI: THREE_WAY_SUMMARY_METRICS.json.

```text
THREE_WAY_COMPARISON_STATUS=COMPLETE
FROZEN_INPUTS_UNCHANGED=YES
PRIMARY_SCOPE=PRODUCTION_REVIEWED
AI1_AI2_HUMAN_COMPARISON_VALID=YES
ABSOLUTE_PAGE_LEVEL_RECALL_ESTIMATED=NO
ATTACHMENT_V2_ANALYSIS=DESCRIPTIVE
RESULTS_REPRODUCIBLE=YES
```
