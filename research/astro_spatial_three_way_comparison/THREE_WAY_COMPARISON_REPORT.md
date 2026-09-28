# AI1 ↔ AI2 ↔ Human: трёхстороннее пространственное сравнение

Результат: frozen provenance сохранён, основной scope — production-reviewed candidate
union. STAR/LABEL protocol 1.2 и attachment-v2 анализируются отдельно. Исходные пакеты
не изменялись. План зафиксирован до агрегирования: SHA-256 `cf53f275e8485538362eb9b3dd278aa42e7bd719eed2a4f9f79a71b7743e39cf`.

## 1. Входы и аналитическая единица

Зарегистрировано 673 файлов в INPUT_MANIFEST.tsv;
проверены четыре upstream SHA256SUMS. Полная однократная сохранность source ID подтверждена
воспроизведением существующего constrained reconciliation, без нового IoU matching.
6 pairwise edges не объединены constrained
reconciliation (если есть): это сохранённая особенность frozen mapping, а не новая правка.
OTHER diagram candidates без human review не являются отрицательными наблюдениями.
REVIEW_PHASE_INPUTS.tsv регистрирует реальные XML очереди и фазовые финальные таблицы;
отдельные phase manifests не выдумываются вместо общего human manifest.
Полные Consensus-QC samples: 34 STAR/18 LABEL; дополнительные финальные TSV: 7/2.
Остальные 27/16 уже присутствуют в прежних reviewed phases и не дублируются.

Calibration panels: f68r1/f68r3/f68v2, даже если кандидаты повторно входят в High.
Production — остальные панели. Для STAR reviewed 444/451, LABEL 290/291;
5 STAR/3 LABEL UNCERTAIN и 7 STAR/1 LABEL NOT_REVIEWED не входят в strict accuracy.
HIGH confidence всем назначен по указанию reviewer, поэтому не служит измерением уверенности.

## 2. Основные production metrics

| Тип | AI | Confirmation strict | Cluster 95% CI | MODIFY среди confirmed | Условное покрытие confirmed union |
|---|---|---|---|---|---|
| STAR | AI1 | 61.7% (92/149; 95% Wilson 53.7–69.2%) | 49.3–81.2% | 100.0% (92/92; 95% Wilson 96.0–100.0%) | 51.4% (92/179; 95% Wilson 44.1–58.6%) |
| STAR | AI2 | 100.0% (175/175; 95% Wilson 97.9–100.0%) | 100.0–100.0% | 65.1% (114/175; 95% Wilson 57.8–71.8%) | 97.8% (175/179; 95% Wilson 94.4–99.1%) |
| LABEL | AI1 | 44.9% (53/118; 95% Wilson 36.2–53.9%) | 34.0–55.4% | 100.0% (53/53; 95% Wilson 93.2–100.0%) | 41.1% (53/129; 95% Wilson 33.0–49.7%) |
| LABEL | AI2 | 95.5% (126/132; 95% Wilson 90.4–97.9%) | 92.2–99.0% | 79.4% (100/126; 95% Wilson 71.5–85.5%) | 97.7% (126/129; 95% Wilson 93.4–99.2%) |

ACCEPT и MODIFY подтверждают обнаружение, но MODIFY не считается геометрическим совпадением.
Confirmation strict denominator = ACCEPT+MODIFY+REJECT среди кандидатов источника.
Уверенность долей: Wilson conditional candidate CI; cluster CI учитывает зависимость
внутри панелей, но небольшое число панелей ограничивает обобщение. Все reviewed denominators,
uncertainty, ACCEPT/MODIFY fractions и calibration/combined находятся в JSON.

## 3. AI1 ↔ AI2 frozen pairwise agreement

| Тип | Scope | Matched | AI1 / AI2 | Jaccard | Raw bbox IoU median [IQR] |
|---|---|---|---|---|---|
| STAR | PRODUCTION | 96 | 157 / 183 | 39.3% (96/244; 95% Wilson 33.4–45.6%) | 0.107 [0.017, 0.255] |
| STAR | CALIBRATION | 75 | 136 / 147 | 36.1% (75/208; 95% Wilson 29.8–42.8%) | 0.112 [0.000, 0.226] |
| STAR | COMBINED | 171 | 293 / 330 | 37.8% (171/452; 95% Wilson 33.5–42.4%) | 0.109 [0.008, 0.244] |
| LABEL | PRODUCTION | 54 | 119 / 133 | 27.3% (54/198; 95% Wilson 21.5–33.9%) | 0.049 [0.004, 0.086] |
| LABEL | CALIBRATION | 38 | 56 / 73 | 41.8% (38/91; 95% Wilson 32.2–52.0%) | 0.181 [0.107, 0.408] |
| LABEL | COMBINED | 92 | 175 / 206 | 31.8% (92/289; 95% Wilson 26.7–37.4%) | 0.080 [0.031, 0.236] |

Jaccard = matched/(AI1+AI2−matched); per-source matched fractions и CI — в JSON.
Pairwise raw bbox IoU измеряет envelopes. STAR matching было class-constrained:
raw class agreement=1 среди matches обусловлено алгоритмом, не доказывает верный класс.
LABEL имеет один фиксированный тип. Confusion matrices независимых AI confidence
даны в confidence_agreement_exploratory; HIGH человеческой confidence назначен вручную,
поэтому Fleiss kappa/Krippendorff alpha для трёх источников невалидны и не рассчитывались.
NOT_REVIEWED/missing никогда не заполняются категориями в confusion matrix.
Frozen match и both-source support через constrained provenance различаются; оба сохранены.
Unmatched source candidates и их human outcomes: AI1_AI2_PAIRWISE_RESULTS.tsv.

## 4. Геометрия и human correction

| Тип | Источник | Геометрия / representation | n | Median IoU [Q1,Q3] | Bootstrap 95% CI |
|---|---|---|---|---|---|
| LABEL | AI1 | ROTATED_RECTANGLE / orientation_normalized_proxy | 53 | 0.081 [0.012,0.171] | 0.057–0.124 |
| LABEL | AI2 | ELLIPSE_ENVELOPE_PROXY / filled_ellipse_envelope_proxy | 5 | 1.000 [1.000,1.000] | 1.000–1.000 |
| LABEL | AI2 | ROTATED_RECTANGLE / orientation_normalized_proxy | 121 | 0.549 [0.448,0.667] | 0.500–0.631 |
| LABEL | CANDIDATE | ELLIPSE_ENVELOPE_PROXY / actual_candidate_display | 5 | 1.000 [1.000,1.000] | 1.000–1.000 |
| LABEL | CANDIDATE | ROTATED_RECTANGLE / actual_candidate_display | 124 | 0.452 [0.345,0.609] | 0.376–0.557 |
| STAR | AI1 | AABB / raw_AABB | 92 | 0.116 [0.025,0.216] | 0.049–0.160 |
| STAR | AI2 | AABB / raw_AABB | 175 | 0.755 [0.631,1.000] | 0.603–1.000 |
| STAR | CANDIDATE | AABB / actual_candidate_display | 179 | 0.615 [0.375,1.000] | 0.434–1.000 |

Детали IoU/center/size/angle, median/IQR/bootstrap CI отдельно по ACCEPT/MODIFY и панели:
GEOMETRY_SUMMARY_METRICS.tsv. Original source, candidate display/union и final human
coordinates одновременно сохранены. MODIFY — факт необходимости коррекции по reviewer,
а 1−candidate/final IoU — её количественный геометрический proxy.

STAR использует реальные axis-aligned bbox. Для LABEL исходные AI fields не задают
однозначно oriented side lengths: primary orientation-normalized polygon — объявленный
proxy, не точное восстановление AI segmentation. Raw-dimensions-rotated и raw-envelope
метрики/пороги представлены отдельно в GEOMETRY_SENSITIVITY_RESULTS.tsv.
ELLIPSE означает ring path; filled-ellipse overlap оценивает только envelope. Без толщины
и ink segmentation истинный ring/text overlap не определён. Геометрические типы не объединены
в одну сводную точность. При одном panel bootstrap кандидатный, помечен отдельно.

## 5. Эффект both-AI support и систематические различия

| Тип | Support | Confirmation strict | MODIFY среди confirmed |
|---|---|---|---|
| STAR | BOTH_AI | 100.0% (88/88; 95% Wilson 95.8–100.0%) | 100.0% (88/88; 95% Wilson 95.8–100.0%) |
| STAR | AI1_ONLY | 6.6% (4/61; 95% Wilson 2.6–15.7%) | 100.0% (4/4; 95% Wilson 51.0–100.0%) |
| STAR | AI2_ONLY | 100.0% (87/87; 95% Wilson 95.8–100.0%) | 29.9% (26/87; 95% Wilson 21.3–40.2%) |
| LABEL | BOTH_AI | 94.3% (50/53; 95% Wilson 84.6–98.1%) | 100.0% (50/50; 95% Wilson 92.9–100.0%) |
| LABEL | AI1_ONLY | 4.6% (3/65; 95% Wilson 1.6–12.7%) | 100.0% (3/3; 95% Wilson 43.9–100.0%) |
| LABEL | AI2_ONLY | 96.2% (76/79; 95% Wilson 89.4–98.7%) | 65.8% (50/76; 95% Wilson 54.6–75.5%) |

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

| Тип / outcome | Contrast | n left / right | Эффект left−right | Cluster 95% CI | p raw / Holm | Статус |
|---|---|---|---|---|---|---|
| STAR / CONFIRMATION | BOTH_AI − AI1_ONLY | 88 / 61 | 0.934 | 0.870–1.000 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| STAR / REJECTION | BOTH_AI − AI1_ONLY | 88 / 61 | -0.934 | -1.000–-0.870 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| STAR / UNCERTAINTY | BOTH_AI − AI1_ONLY | 89 / 61 | 0.011 | 0.000–0.037 | 1.00000 / 1.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| STAR / MODIFY_AMONG_CONFIRMED | BOTH_AI − AI1_ONLY | 88 / 4 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | INSUFFICIENT_SMALL_GROUP |
| STAR / CONFIRMATION | BOTH_AI − AI2_ONLY | 88 / 87 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | BOUNDARY_DEGENERATE_BOOTSTRAP_NOT_EQUIVALENCE |
| STAR / REJECTION | BOTH_AI − AI2_ONLY | 88 / 87 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | BOUNDARY_DEGENERATE_BOOTSTRAP_NOT_EQUIVALENCE |
| STAR / UNCERTAINTY | BOTH_AI − AI2_ONLY | 89 / 87 | 0.011 | 0.000–0.037 | 1.00000 / 1.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| STAR / MODIFY_AMONG_CONFIRMED | BOTH_AI − AI2_ONLY | 88 / 87 | 0.701 | 0.197–0.979 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| STAR / CONFIRMATION | AI1_ONLY − AI2_ONLY | 61 / 87 | -0.934 | -1.000–-0.870 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| STAR / REJECTION | AI1_ONLY − AI2_ONLY | 61 / 87 | 0.934 | 0.870–1.000 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| STAR / UNCERTAINTY | AI1_ONLY − AI2_ONLY | 61 / 87 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | BOUNDARY_DEGENERATE_BOOTSTRAP_NOT_EQUIVALENCE |
| STAR / MODIFY_AMONG_CONFIRMED | AI1_ONLY − AI2_ONLY | 4 / 87 | 0.701 | 0.197–0.979 | 0.01025 / 0.07178 | INSUFFICIENT_SMALL_GROUP |
| STAR / CORRECTION_1_IOU | BOTH_AI − AI1_ONLY | 88 / 4 | -0.380 | -0.438–-0.332 | 0.00106 / 0.00400 | INSUFFICIENT_SMALL_GROUP |
| STAR / CORRECTION_1_IOU | BOTH_AI − AI2_ONLY | 88 / 87 | 0.619 | 0.342–0.668 | 0.00100 / 0.00400 | EXPLORATORY_FEW_PANEL_CLUSTERS |
| STAR / CORRECTION_1_IOU | AI1_ONLY − AI2_ONLY | 4 / 87 | 0.999 | 0.730–1.000 | 0.00106 / 0.00400 | INSUFFICIENT_SMALL_GROUP |
| STAR / PAIRED_HUMAN_IOU | AI1 − AI2 | 88 / 88 | -0.482 | -0.705–-0.399 | 0.00100 / 0.00400 | EXPLORATORY_FEW_PANEL_CLUSTERS |
| LABEL / CONFIRMATION | BOTH_AI − AI1_ONLY | 53 / 65 | 0.897 | 0.821–1.000 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| LABEL / REJECTION | BOTH_AI − AI1_ONLY | 53 / 65 | -0.897 | -1.000–-0.820 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| LABEL / UNCERTAINTY | BOTH_AI − AI1_ONLY | 53 / 65 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | BOUNDARY_DEGENERATE_BOOTSTRAP_NOT_EQUIVALENCE |
| LABEL / MODIFY_AMONG_CONFIRMED | BOTH_AI − AI1_ONLY | 50 / 3 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | INSUFFICIENT_SMALL_GROUP |
| LABEL / CONFIRMATION | BOTH_AI − AI2_ONLY | 53 / 79 | -0.019 | -0.105–0.064 | 0.68377 / 1.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| LABEL / REJECTION | BOTH_AI − AI2_ONLY | 53 / 79 | 0.019 | -0.067–0.105 | 0.68377 / 1.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| LABEL / UNCERTAINTY | BOTH_AI − AI2_ONLY | 53 / 79 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | BOUNDARY_DEGENERATE_BOOTSTRAP_NOT_EQUIVALENCE |
| LABEL / MODIFY_AMONG_CONFIRMED | BOTH_AI − AI2_ONLY | 50 / 76 | 0.342 | 0.097–0.729 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| LABEL / CONFIRMATION | AI1_ONLY − AI2_ONLY | 65 / 79 | -0.916 | -0.971–-0.886 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| LABEL / REJECTION | AI1_ONLY − AI2_ONLY | 65 / 79 | 0.916 | 0.886–0.971 | 0.00000 / 0.00000 | ESTIMATED_FEW_PANEL_CLUSTERS |
| LABEL / UNCERTAINTY | AI1_ONLY − AI2_ONLY | 65 / 79 | 0.000 | 0.000–0.000 | 1.00000 / 1.00000 | BOUNDARY_DEGENERATE_BOOTSTRAP_NOT_EQUIVALENCE |
| LABEL / MODIFY_AMONG_CONFIRMED | AI1_ONLY − AI2_ONLY | 3 / 76 | 0.342 | 0.097–0.729 | 0.54693 / 1.00000 | INSUFFICIENT_SMALL_GROUP |
| LABEL / CORRECTION_1_IOU | BOTH_AI − AI1_ONLY | 50 / 3 | -0.196 | -0.237–-0.151 | 0.00107 / 0.00400 | INSUFFICIENT_SMALL_GROUP |
| LABEL / CORRECTION_1_IOU | BOTH_AI − AI2_ONLY | 50 / 76 | 0.217 | 0.117–0.701 | 0.00100 / 0.00400 | EXPLORATORY_FEW_PANEL_CLUSTERS |
| LABEL / CORRECTION_1_IOU | AI1_ONLY − AI2_ONLY | 3 / 76 | 0.413 | 0.291–0.829 | 0.00108 / 0.00400 | INSUFFICIENT_SMALL_GROUP |
| LABEL / PAIRED_HUMAN_IOU | AI1 − AI2 | 50 / 50 | -0.445 | -0.482–-0.379 | 0.00100 / 0.00400 | EXPLORATORY_GEOMETRY_PROXY |

Категориальные семьи: 12 contrasts на STAR и 12 на LABEL; Holm внутри каждой семьи.
Continuous families: по 4 contrasts. Fisher p предполагает независимость кандидатов и
служит лишь дополнительным сигналом; основные effect CI — joint panel-cluster bootstrap.
Bootstrap sign-tail p для geometry exploratory, не exact null test. Малые группы и
CI, пересекающие ноль, обозначают недостаточную определённость, не эквивалентность.
Сравнение AI human-IoU paired на одинаковых confirmed candidates не смешивается с
сравнением AI, имеющих разный набор кандидатов.

## 6. Панели, морфология и case studies

| Panel | Тип | Union | Strict | ACCEPT / MODIFY / REJECT / UNCERTAIN / NOT_REVIEWED | Confirmation |
|---|---|---|---|---|---|
| f68r1 (CALIBRATION) | LABEL | 41 | 41 | 3 / 34 / 4 / 0 / 0 | 0.9024390243902439 |
| f68r3 (CALIBRATION) | LABEL | 22 | 20 | 2 / 12 / 6 / 2 / 0 | 0.7 |
| f68v2 (CALIBRATION) | LABEL | 30 | 29 | 5 / 14 / 10 / 1 / 0 | 0.6551724137931034 |
| f67r1 (PRODUCTION) | LABEL | 30 | 30 | 14 / 7 / 9 / 0 / 0 | 0.7 |
| f67r2 (PRODUCTION) | LABEL | 49 | 49 | 3 / 36 / 10 / 0 / 0 | 0.7959183673469388 |
| f67v1 (PRODUCTION) | LABEL | 46 | 46 | 0 / 28 / 18 / 0 / 0 | 0.6086956521739131 |
| f68r2 (PRODUCTION) | LABEL | 58 | 57 | 6 / 25 / 26 / 0 / 1 | 0.543859649122807 |
| f68v1 (PRODUCTION) | LABEL | 15 | 15 | 3 / 7 / 5 / 0 / 0 | 0.6666666666666666 |
| f68r1 (CALIBRATION) | STAR | 32 | 32 | 2 / 27 / 3 / 0 / 0 | 0.90625 |
| f68r3 (CALIBRATION) | STAR | 108 | 105 | 26 / 50 / 29 / 3 / 0 | 0.7238095238095238 |
| f68v2 (CALIBRATION) | STAR | 67 | 66 | 26 / 11 / 29 / 1 / 0 | 0.5606060606060606 |
| f67r1 (PRODUCTION) | STAR | 35 | 35 | 8 / 14 / 13 / 0 / 0 | 0.6285714285714286 |
| f67v1 (PRODUCTION) | STAR | 63 | 63 | 14 / 31 / 18 / 0 / 0 | 0.7142857142857143 |
| f68r2 (PRODUCTION) | STAR | 63 | 56 | 0 / 52 / 4 / 0 / 7 | 0.9285714285714286 |
| f68v1 (PRODUCTION) | STAR | 83 | 82 | 39 / 21 / 22 / 1 / 0 | 0.7317073170731707 |

STAR: максимальная median-коррекция candidate geometry на f68r2 (AABB, median IoU 0.428, n=52); LABEL: максимальная median-коррекция candidate geometry на f67v1 (ROTATED_RECTANGLE, median IoU 0.362, n=28).

Полная стратификация priority/support/geometry/scope/outcome/stage/A-positive:
STRATIFIED_OBJECT_METRICS.tsv. ERROR_FLAGS.tsv и ERROR_TAXONOMY.md отделяют измеряемые
center/extent/orientation issues от неизвестных причин. Frozen поля не позволяют
доказать split/merge, decorative confusion, low contrast или occlusion по одному REJECT.
Калибровка не упражняла SPLIT/MERGE. Граничные STAR на f67v1 подтверждают видимость в скане,
но не принадлежность основной странице; ownership из panel ID не выводится.
Case-study overlays выбираются deterministic экстремумами и neutral ID tie-breaking
по заранее заданному правилу, source selection table — figures/CASE_STUDIES.tsv.

## 7. Sensitivity

| Policy | Тип | AI1 confirmation | AI2 confirmation | Both-AI confirmation |
|---|---|---|---|---|
| PRIMARY_PRODUCTION_STRICT | STAR | 0.617 (92/149) | 1.000 (175/175) | 1.000 (88/88) |
| PRIMARY_PRODUCTION_STRICT | LABEL | 0.449 (53/118) | 0.955 (126/132) | 0.943 (50/53) |
| UNCERTAIN_NEGATIVE_BOUND | STAR | 0.613 (92/150) | 0.994 (175/176) | 0.989 (88/89) |
| UNCERTAIN_NEGATIVE_BOUND | LABEL | 0.449 (53/118) | 0.955 (126/132) | 0.943 (50/53) |
| UNCERTAIN_POSITIVE_BOUND | STAR | 0.620 (93/150) | 1.000 (176/176) | 1.000 (89/89) |
| UNCERTAIN_POSITIVE_BOUND | LABEL | 0.449 (53/118) | 0.955 (126/132) | 0.943 (50/53) |
| CALIBRATION_ONLY_STRICT | STAR | 0.564 (75/133) | 0.966 (140/145) | 0.961 (73/76) |
| CALIBRATION_ONLY_STRICT | LABEL | 0.660 (35/53) | 0.972 (70/72) | 0.972 (35/36) |
| COMBINED_STRICT | STAR | 0.592 (167/282) | 0.984 (315/320) | 0.982 (161/164) |
| COMBINED_STRICT | LABEL | 0.515 (88/171) | 0.961 (196/204) | 0.955 (85/89) |
| PRODUCTION_HIGH_ONLY | STAR | 0.521 (62/119) | 1.000 (145/145) | 1.000 (58/58) |
| PRODUCTION_HIGH_ONLY | LABEL | 0.409 (45/110) | 0.952 (118/124) | 0.933 (42/45) |
| PRODUCTION_HIGH_MEDIUM_DEDUP | STAR | 0.601 (86/143) | 1.000 (169/169) | 1.000 (82/82) |
| PRODUCTION_HIGH_MEDIUM_DEDUP | LABEL | 0.440 (51/116) | 0.954 (124/130) | 0.941 (48/51) |
| PRODUCTION_WITHOUT_A_POSITIVE | STAR | 0.617 (92/149) | 1.000 (175/175) | 1.000 (88/88) |
| PRODUCTION_WITHOUT_A_POSITIVE | LABEL | 0.449 (53/118) | 0.955 (126/132) | 0.943 (50/53) |
| PRODUCTION_A_POSITIVE_ONLY | STAR | NA (0/0) | NA (0/0) | NA (0/0) |
| PRODUCTION_A_POSITIVE_ONLY | LABEL | NA (0/0) | NA (0/0) | NA (0/0) |

STAR: направление сравнения confirmation AI1/AI2 по 6 estimable production policies: AI2. Это описательная устойчивость; не доказательство причинности или эквивалентности.

STAR, BOTH_AI−AI1_ONLY: sensitivity point-effect range 0.923…0.934 across 6 estimable policies.

STAR, BOTH_AI−AI2_ONLY: sensitivity point-effect range -0.011…0.000 across 6 estimable policies.

LABEL: направление сравнения confirmation AI1/AI2 по 6 estimable production policies: AI2. Это описательная устойчивость; не доказательство причинности или эквивалентности.

LABEL, BOTH_AI−AI1_ONLY: sensitivity point-effect range 0.887…0.897 across 6 estimable policies.

LABEL, BOTH_AI−AI2_ONLY: sensitivity point-effect range -0.029…-0.019 across 6 estimable policies.

Primary всегда strict production. Combined/calibration — явно вторичные наборы.
Uncertain bounds не меняют решения. High/Medium deduplicated by candidate ID;
Consensus-QC не добавляется скрыто в High-only. A-positive исключается/стратифицируется
как изменение выборки, но его отсутствие никогда не считается негативным evidence.
Фактически все 29 STAR/29 LABEL с A-positive support находятся в calibration, поэтому
production A-positive-only группа пуста: эффект A-support в production не оценивается.
Geometry threshold sensitivity меняет диагностические флаги, не source matching.
Направления geometry зависят также от неоднозначных AI LABEL conventions; нельзя объявлять
универсального geometry winner по одной прокси-метрике или pooled median.

## 8. Attachment-v2: только описательный human graph

AI compatibility audit в JSON показывает legacy spatial predicates без независимых
caption-v2 predictions; искусственное трёхстороннее relation agreement не рассчитывалось.
267 explicit decisions = 53 VISUAL_LABEL_OF, 171 UNASSIGNED, 43 UNCERTAIN.
125 individual reviews и 142 reviewer panel-rule negatives различаются по origin.
Coverage explicit = 267/324 eligible (неполный pair universe), 267/675 original source pairs;
individual task completion=125/125; individual coverage eligible=125/324.
57 rule-filtered и 351 endpoint-excluded пары остаются без human решения.
Uncertainty=16.1% (43/267; 95% Wilson 12.2–21.0%) среди explicit, и
34.4% (43/125; 95% Wilson 26.6–43.1%) среди individual reviews.

53 positive edges соединяют 51 LABEL и 51 STAR. 2 STAR имеют
несколько LABEL; 1 LABEL имеют несколько STAR.
Positive graph содержит 49 components:
`{'L1_S1_E1': 46, 'L1_S3_E3': 1, 'L2_S1_E2': 2}`. Degree-zero confirmed endpoint означает только отсутствие
positive relation в выбранной очереди. Panеl decisions, degree distributions, component
IDs и все denominators: ATTACHMENT_V2_* tables, graph section JSON и figures.

## 9. Ответы и допустимое применение

1. AI pair agreement раздельно STAR/LABEL показано в разделе 3: source matching и
   geometry agreement — разные свойства; высокий raw agreement не заменяет human review.
2. Частоты human confirmation каждого AI — раздел 2, условные на reviewed proposed scope.
3. MODIFY fraction отдельно among confirmed — раздел 2; величина correction — раздел 4.
4. Both support сопоставлен с каждым singleton отдельно: эффект неодинаков по comparator,
   панели и типу. Раздел 5/7 показывает effect sizes и неопределённость.
5. Наибольшие панели correction указаны в разделе 6; rotated LABEL и ring envelopes
   имеют разные определения ошибок. Нельзя приписать геометрическому proxy визуальную причину.
6. Наблюдаемая асимметрия AI confirmation не доказывает одинаковую независимую page-level
   чувствительность; paired geometry оценивается только на совпадающем confirmed subset.
7. Политики uncertainty и A-positive проверены явно; устойчивость point estimates
   не устраняет uncertainty между малым количеством панелей.
8. Все confirmation/coverage относятся только к adjudicated source-derived candidate union.
9. AI-consensus не является достаточным основанием automatic acceptance: reviewed both-AI
   содержит rejection и geometry modifications, а LOW review/QC отобраны неслучайно.
   Допустим дополнительный сигнал приоритизации проверки, не новый автоматически принятый gold.
   После human review confirmed geometry можно добавлять в spatial корпус с provenance,
   version, scope, uncertain/not-reviewed exclusions и scan-ownership предупреждениями.
10. Для абсолютного recall нужен независимый exhaustive manual census на заранее
    случайно выбранных новых панелях/tiles: annotator blind to AI queues, разрешены новые
    STAR/LABEL, dual review и adjudication, затем заранее fixed matching/geometry protocol.
    Holdout и panel-stratified sampling должны включать фон, слабые/краевые marks и разные
    text geometries. Только этот независимый denominator оценивает пропуски обоих AI.

## 10. Воспроизводимость и gate

Seed=20260914; bootstrap=2000; call-order-independent metric-key seeds.
Результаты не содержат транскрипции, lexical matching, semantic crosswalk, M3 или
астрономической идентификации. Команды — REPRODUCIBILITY.md; автоматические gates,
twice-run byte comparison, upstream/source checksums и output checksums — VALIDATION_REPORT.md.

```text
THREE_WAY_COMPARISON_STATUS=COMPLETE
FROZEN_INPUTS_UNCHANGED=YES
PRIMARY_SCOPE=PRODUCTION_REVIEWED
AI1_AI2_HUMAN_COMPARISON_VALID=YES
ABSOLUTE_PAGE_LEVEL_RECALL_ESTIMATED=NO
ATTACHMENT_V2_ANALYSIS=DESCRIPTIVE
RESULTS_REPRODUCIBLE=YES
```
