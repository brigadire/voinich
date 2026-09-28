"""Reports derived from the same machine-readable results as the figures."""
from collections import Counter
from common import SEED, BOOTSTRAPS

STATUS_LINES = '''THREE_WAY_COMPARISON_STATUS=COMPLETE
FROZEN_INPUTS_UNCHANGED=YES
PRIMARY_SCOPE=PRODUCTION_REVIEWED
AI1_AI2_HUMAN_COMPARISON_VALID=YES
ABSOLUTE_PAGE_LEVEL_RECALL_ESTIMATED=NO
ATTACHMENT_V2_ANALYSIS=DESCRIPTIVE
RESULTS_REPRODUCIBLE=YES'''

def dictionary():
    return '''# Data dictionary

Unit of object analysis: frozen physical candidate, not pixel, source record, or page.
Primary provenance is HUMAN_CANDIDATE_OBJECTS/LABELS.tsv. Source IDs retain A/AI1/AI2
namespaces. Original source classes/boxes/orientations are kept, not overwritten.

| Frozen decision | Analytic status | Strict metric | Geometry target |
|---|---|---|---|
| ACCEPT | CONFIRMED_UNCHANGED | positive | final frozen shape |
| MODIFY | CONFIRMED_MODIFIED | positive detection, correction separately measured | final frozen shape |
| REJECT | REJECTED | negative proposed candidate | not used as confirmed geometry |
| UNCERTAIN | UNCERTAIN | excluded; separate lower/upper sensitivity bounds | excluded |
| missing human row | NOT_REVIEWED | excluded in every accuracy policy | excluded |

SPLIT/MERGE are allowed in the protocol but absent in the frozen final records.
All observed final classes equal the proposed STAR_OBJECT/LABEL class. A REJECT row
retains a proposed shape/class; it does not independently classify a different object.
Confidence HIGH is the reviewer-wide explicit override, not calibrated certainty.

## Input/output schemas

INPUT_MANIFEST: repository-relative path, upstream package/version, bytes, sha256, role,
frozen=YES. All four ledgers are verified; unused ancillary upstream records remain
registered for source-preservation audit, but their lexical/semantic contents are not
analysed or exported. No nonexistent phase-specific manifest is presumed:
REVIEW_PHASE_INPUTS registers actual CVAT XML queues, final phase tables, counts/hashes;
the global human manifest and BUILD_SUMMARY register the phase freeze.
Consensus-QC final phase TSVs contain only the additional 7 STAR/2 LABEL; full task
samples contain 34 STAR/18 LABEL. Already-reviewed identical observations cover the
remaining 27/16. Full QC sample membership is retained, not counted as new human rows.

ANALYSIS_COHORTS / THREE_WAY_OBJECT_RESULTS:
candidate_id; panel; object_type STAR/LABEL; scope CALIBRATION/PRODUCTION; support_pattern
BOTH_AI/AI1_ONLY/AI2_ONLY/NEITHER_AI; a_positive_support and support_ai1/2 YES/NO;
priority HIGH/MEDIUM/LOW and frozen priority_reason; review_stage (first applicable phase)
and review_phase_memberships (all phases); human_decision; analytic_status; geometry_type;
source_annotation_ids; human_protocol_version=1.2; primary_strict/combined_strict,
uncertain_cohort/not_reviewed_cohort as 0/1. a_absence_negative_evidence is always 0.
Candidate geometry is display bbox plus rotation and union envelope; human geometry
comes from canonical final row; ai1/2/a IDs, raw bbox, class, orientation, confidence
are original records. Detection support does not imply acceptable geometry.

THREE_WAY_GEOMETRY_RESULTS contains confirmed targets only. source AI1/AI2/CANDIDATE,
source_id and human_id preserve provenance. geometry_type stratifies AABB,
ROTATED_RECTANGLE, ELLIPSE_ENVELOPE_PROXY. representation is raw_AABB,
orientation_normalized_proxy, raw_dimensions_rotated_proxy, filled_ellipse_envelope_proxy
or actual_candidate_display. primary_representation=0 only for alternate rotated proxy.
source_raw_bbox/final_human_bbox and orientations are retained. IoU is polygon area
intersection/union (filled envelope for ellipse, NOT ring ink); center_distance_px is
Euclidean, panel_norm divides by panel diagonal and shape_norm by target shape diagonal.
width/height_relative_error = abs(source-target)/target; width/height_ratio = source/target;
axial_angle_error_deg is min modulo-180 angle separation (NA for STAR AABB).
raw_envelope_iou compares raw AI bbox to axis-aligned envelope of actual human shape.
Missing metrics use blank TSV / JSON null, never invented zero.

GEOMETRY_SUMMARY_METRICS: n, median, q1/q3/iqr, mean, ci95_low/high, ci_method,
panel_clusters, bootstrap_replicates by scope/source/shape/representation/grouping/stratum
and metric. JSON embeds identical rows. Single-panel bootstrap cannot infer panel variability.

AI1_AI2_PAIRWISE_RESULTS retains frozen match_status MATCHED/LEFT_ONLY/RIGHT_ONLY,
both original and physical IDs, frozen raw-box IoU, retained human outcome if in scope,
independent source confidence, conditioned class agreement and same_physical_candidate.
PAIRWISE_RECONCILIATION_AUDIT includes all object classes (unreviewed other classes are
not human negatives). Constrained nontransitive pairwise edges can remain unjoined;
the reproduced frozen reconciliation, not a new IoU match, remains the analytic unit.

JSON fractions always contain numerator/denominator/estimate/ci95/ci_method;
headline confirmation/coverage also include cluster_ci95, cluster_ci_method and valid
bootstrap count. Strict confirmation denominators exclude uncertainty; all-reviewed
uncertainty denominators retain it. modification_among_confirmed conditions on ACCEPT/MODIFY.
All rates are conditional on proposed/adjudicated candidate scope, not page precision/recall.

STATISTICAL_TESTS: family class_categorical or class_continuous; group sizes/events;
left-minus-right risk difference or median difference; CI and valid bootstrap count;
Fisher supplementary p (candidate independence) or exploratory bootstrap sign-tail p;
p_holm within predeclared family. Paired human IoU is median of per-candidate differences.
Small groups/degenerate intervals/few clusters do not establish equivalence.
Independent AI-confidence raw agreement and conditioned fixed-class agreement also
include explicit numerator/denominator/Wilson CI in JSON; these are not object validity.

SENSITIVITY_RESULTS policies never change raw decisions. UNCERTAIN positive/negative
are explicit mathematical bounds; not-reviewed remains excluded. Geometry sensitivity
changes representations/diagnostic flags only, not matching or confirmation.

Attachment tables use relation_protocol_version=2, never spatial-v1 categories.
VISUAL_LABEL_OF = positive visible caption attachment, UNASSIGNED = explicit negative
for that pair, UNCERTAIN = unresolved. Origins separate 125 individual decisions from
142 reviewer panel-rule decisions. 57 rule-filtered and 351 endpoint-excluded pairs
are NOT_REVIEWED, not negatives. Graph degrees include all confirmed endpoints;
degree zero means no positive link in the selected source-derived queue only.
Components include positive-edge nodes, not isolated confirmed nodes.
'''

def number(value, digits=3): return 'NA' if value is None else f'{value:.{digits}f}'

def percent(f):
    if f['estimate'] is None: return f"NA (0/{f['denominator']})"
    return f"{100*f['estimate']:.1f}% ({f['numerator']}/{f['denominator']}; 95% Wilson {100*f['ci95'][0]:.1f}–{100*f['ci95'][1]:.1f}%)"

def cluster(f):
    ci=f.get('cluster_ci95',[None,None])
    return 'NA' if ci[0] is None else f'{100*ci[0]:.1f}–{100*ci[1]:.1f}%'

def reports(out,data,tests,sensitivity,errors):
    obj=data['object_metrics']
    graph=data['attachment_v2']
    production=obj['PRODUCTION']
    overview=['| Тип | AI | Confirmation strict | Cluster 95% CI | MODIFY среди confirmed | Условное покрытие confirmed union |',
              '|---|---|---|---|---|---|']
    for kind in ('STAR','LABEL'):
        for ai in ('AI1','AI2'):
            stats=production[kind]['AI'][ai]
            overview.append(f"| {kind} | {ai} | {percent(stats['confirmation_strict'])} | {cluster(stats['confirmation_strict'])} | {percent(stats['modification_among_confirmed'])} | {percent(production[kind]['conditional_coverage'][ai])} |")
    overview='\n'.join(overview)
    support_lines=['| Тип | Support | Confirmation strict | MODIFY среди confirmed |',
                   '|---|---|---|---|']
    for kind in ('STAR','LABEL'):
        for group in ('BOTH_AI','AI1_ONLY','AI2_ONLY'):
            r=production[kind]['support_group_metrics'][group]
            support_lines.append(f"| {kind} | {group} | {percent(r['confirmation_strict'])} | {percent(r['modification_among_confirmed'])} |")
    support_table='\n'.join(support_lines)
    pair_lines=['| Тип | Scope | Matched | AI1 / AI2 | Jaccard | Raw bbox IoU median [IQR] |',
                '|---|---|---|---|---|---|']
    for r in data['pairwise_metrics']:
        if r['panel']!='ALL': continue
        g=r['raw_AABB_iou']
        pair_lines.append(f"| {r['object_type']} | {r['scope']} | {r['matched_pairs']} | {r['ai1_count']} / {r['ai2_count']} | {percent(r['matched_jaccard'])} | {number(g['median'])} [{number(g['q1'])}, {number(g['q3'])}] |")
    pair_table='\n'.join(pair_lines)
    geom_lines=['| Тип | Источник | Геометрия / representation | n | Median IoU [Q1,Q3] | Bootstrap 95% CI |',
                '|---|---|---|---|---|---|']
    for r in data['geometry_summaries']:
        if r['scope']!='PRODUCTION' or r['grouping']!='ALL' or r['metric']!='iou' or r['representation']=='raw_dimensions_rotated_proxy': continue
        geom_lines.append(f"| {r['object_type']} | {r['source']} | {r['geometry_type']} / {r['representation']} | {r['n']} | {number(r['median'])} [{number(r['q1'])},{number(r['q3'])}] | {number(r['ci95_low'])}–{number(r['ci95_high'])} |")
    geometry_table='\n'.join(geom_lines)
    effect_lines=['| Тип / outcome | Contrast | n left / right | Эффект left−right | Cluster 95% CI | p raw / Holm | Статус |',
                  '|---|---|---|---|---|---|---|']
    for r in tests:
        effect_lines.append(f"| {r['object_type']} / {r['outcome']} | {r['left_group']} − {r['right_group']} | {r['n_left']} / {r['n_right']} | {number(r['effect'])} | {number(r['ci95_low'])}–{number(r['ci95_high'])} | {number(r['p_raw'],5)} / {number(r['p_holm'],5)} | {r['status']} |")
    effects='\n'.join(effect_lines)
    sensitivity_lines=['| Policy | Тип | AI1 confirmation | AI2 confirmation | Both-AI confirmation |',
                       '|---|---|---|---|---|']
    by={(r['policy'],r['object_type'],r['support_group']):r for r in sensitivity}
    for policy in dict.fromkeys(r['policy'] for r in sensitivity):
        for kind in ('STAR','LABEL'):
            cells=[]
            for group in ('AI1','AI2','BOTH_AI'):
                r=by[policy,kind,group]
                cells.append(f"{number(r['estimate'])} ({r['confirmed_or_bound_numerator']}/{r['denominator']})")
            sensitivity_lines.append('| '+ ' | '.join([policy,kind]+cells)+' |')
    sensitivity_table='\n'.join(sensitivity_lines)
    panel_lines=['| Panel | Тип | Union | Strict | ACCEPT / MODIFY / REJECT / UNCERTAIN / NOT_REVIEWED | Confirmation |',
                 '|---|---|---|---|---|---|']
    from common import read
    strata=read(out/'STRATIFIED_OBJECT_METRICS.tsv')
    for r in strata:
        if r['stratum_field']!='panel': continue
        panel_lines.append(f"| {r['stratum_value']} ({r['scope']}) | {r['object_type']} | {r['candidates']} | {r['strict_reviewed']} | {r['accept']} / {r['modify']} / {r['rejected']} / {r['uncertain']} / {r['not_reviewed']} | {r['confirmation_strict']} |")
    panel_table='\n'.join(panel_lines)
    worst=[]
    for kind in ('STAR','LABEL'):
        panels=[r for r in data['geometry_summaries'] if r['scope']=='PRODUCTION' and r['grouping']=='PANEL' and r['metric']=='iou' and r['source']=='CANDIDATE' and r['object_type']==kind]
        if panels:
            r=min(panels,key=lambda r:(r['median'],r['stratum']))
            worst.append(f"{kind}: максимальная median-коррекция candidate geometry на {r['stratum']} ({r['geometry_type']}, median IoU {r['median']:.3f}, n={r['n']})")
    flags=Counter(flag for r in errors for flag in r['flags'].split(';') if flag)
    taxonomy='''# Error taxonomy: geometric/frozen proxies only

Flags are reproducible measurements, not manually inferred visual causes. The original
human fields do not code rejection causes, contrast, occlusion, object grouping or
ink/path thickness. Therefore merging, splitting, decorative confusion and faintness
cannot be assigned as established causes. No SPLIT/MERGE was exercised. A class-constrained
AI match does not validate the class of an actual visible mark.

| Flag | Count (combined candidate union; overlaps allowed) | Meaning |
|---|---|---|
'''
    meanings={'FROZEN_MAJOR_BBOX_OR_GRANULARITY_CONFLICT':'Frozen priority_reason; possible geometry/granularity issue, not proof of split/merge',
              'HUMAN_REJECTED_CANDIDATE_NO_CAUSE_CODE':'Reviewer rejected proposed candidate; no cause code',
              'UNRESOLVED_VISUAL_CANDIDATE':'Human UNCERTAIN, not a negative',
              'LOW_SOURCE_CONFIDENCE_PROXY_NOT_OBSERVED_CONTRAST':'LOW/AMBIGUOUS source confidence, not measured image contrast',
              'LARGE_CENTER_CORRECTION':'Candidate center shift >0.5 target shape diagonal among confirmed',
              'EXCESS_CANDIDATE_EXTENT':'Candidate/final width or height >2 among confirmed',
              'INSUFFICIENT_CANDIDATE_EXTENT':'Candidate/final width or height <0.5 among confirmed',
              'LARGE_ORIENTATION_CORRECTION':'Axial angle error >30 degrees among confirmed',
              'RING_PATH_ENVELOPE_ONLY':'Filled ellipse envelope; ink/ring thickness not available',
              'LOW_SHAPE_OVERLAP_AFTER_REVIEW':'Candidate/final polygon IoU <0.25 among confirmed',
              'SCAN_RIGHT_BOUNDARY_OWNERSHIP_NOT_INFERRED':'f67v1 scan boundary warning, not manuscript ownership'}
    for flag,count in sorted(flags.items()): taxonomy+=f'| {flag} | {count} | {meanings[flag]} |\n'
    taxonomy+='\nSource: ERROR_FLAGS.tsv. Deterministic case selection: figures/CASE_STUDIES.tsv and prospective plan. '
    taxonomy+='Partial edge stars may be valid visible objects; f67v1 adjacent-page strip ownership is not inferred from acceptance.\n'
    (out/'ERROR_TAXONOMY.md').write_text(taxonomy,encoding='utf-8')
    interpretations=[]
    for kind in ('STAR','LABEL'):
        a=production[kind]['AI']['AI1']['confirmation_strict']
        b=production[kind]['AI']['AI2']['confirmation_strict']
        direction='AI2' if b['estimate']>a['estimate'] else 'AI1'
        interpretations.append(f"{kind}: {direction} имеет более высокую наблюдаемую conditional confirmation rate в production ({a['numerator']}/{a['denominator']} AI1 против {b['numerator']}/{b['denominator']} AI2). Это не оценка абсолютной page-level precision.")
        for r in tests:
            if r['object_type']==kind and r['outcome']=='CONFIRMATION' and r['left_group']=='BOTH_AI':
                interpretations.append(f"{kind}, both против {r['right_group']}: risk difference {number(r['effect'])}, cluster CI [{number(r['ci95_low'])},{number(r['ci95_high'])}], n={r['n_left']}/{r['n_right']}. "
                    +('Разница недостаточно определена между панелями.' if r['ci95_low'] is None or r['ci95_low']<=0<=r['ci95_high'] or 'INSUFFICIENT' in r['status'] else 'Направление поддерживается этим ограниченным набором панелей.'))
    interpretations='\n\n'.join(interpretations)
    consensus_conclusion='''Совместная поддержка не даёт обнаружимого преимущества confirmation над
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
выборки, не нулевая неопределённость; Wilson интервалы групп остаются ненулевой ширины.'''
    # Check ranking and consensus effects across sensitivity policies, not only p-values.
    stability=[]
    prod_policies=['PRIMARY_PRODUCTION_STRICT','UNCERTAIN_NEGATIVE_BOUND','UNCERTAIN_POSITIVE_BOUND',
                   'PRODUCTION_HIGH_ONLY','PRODUCTION_HIGH_MEDIUM_DEDUP','PRODUCTION_WITHOUT_A_POSITIVE','PRODUCTION_A_POSITIVE_ONLY']
    for kind in ('STAR','LABEL'):
        valid=[(p,by[p,kind,'AI1']['estimate'],by[p,kind,'AI2']['estimate']) for p in prod_policies if by[p,kind,'AI1']['estimate'] is not None and by[p,kind,'AI2']['estimate'] is not None]
        signs={('AI2' if b>a else 'AI1' if a>b else 'tie') for _,a,b in valid}
        stability.append(f"{kind}: направление сравнения confirmation AI1/AI2 по {len(valid)} estimable production policies: {', '.join(sorted(signs))}. Это описательная устойчивость; не доказательство причинности или эквивалентности.")
        for single in ('AI1_ONLY','AI2_ONLY'):
            outcomes=[]
            for p in prod_policies:
                both=by[p,kind,'BOTH_AI']['estimate']
                s=by[p,kind,single]['estimate']
                if both is not None and s is not None: outcomes.append(both-s)
            stability.append(f"{kind}, BOTH_AI−{single}: sensitivity point-effect range {min(outcomes):.3f}…{max(outcomes):.3f} across {len(outcomes)} estimable policies." if outcomes else f'{kind}/{single}: insufficient sensitivity groups.')
    stability='\n\n'.join(stability)
    report=f'''# AI1 ↔ AI2 ↔ Human: трёхстороннее пространственное сравнение

Результат: frozen provenance сохранён, основной scope — production-reviewed candidate
union. STAR/LABEL protocol 1.2 и attachment-v2 анализируются отдельно. Исходные пакеты
не изменялись. План зафиксирован до агрегирования: SHA-256 `{data['analysis_plan_sha256']}`.

## 1. Входы и аналитическая единица

Зарегистрировано {data['input_validation']['registered_files']} файлов в INPUT_MANIFEST.tsv;
проверены четыре upstream SHA256SUMS. Полная однократная сохранность source ID подтверждена
воспроизведением существующего constrained reconciliation, без нового IoU matching.
{data['input_validation']['constrained_edges_not_joined']} pairwise edges не объединены constrained
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

{overview}

ACCEPT и MODIFY подтверждают обнаружение, но MODIFY не считается геометрическим совпадением.
Confirmation strict denominator = ACCEPT+MODIFY+REJECT среди кандидатов источника.
Уверенность долей: Wilson conditional candidate CI; cluster CI учитывает зависимость
внутри панелей, но небольшое число панелей ограничивает обобщение. Все reviewed denominators,
uncertainty, ACCEPT/MODIFY fractions и calibration/combined находятся в JSON.

## 3. AI1 ↔ AI2 frozen pairwise agreement

{pair_table}

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

{geometry_table}

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

{support_table}

{interpretations}

{consensus_conclusion}

{effects}

Категориальные семьи: 12 contrasts на STAR и 12 на LABEL; Holm внутри каждой семьи.
Continuous families: по 4 contrasts. Fisher p предполагает независимость кандидатов и
служит лишь дополнительным сигналом; основные effect CI — joint panel-cluster bootstrap.
Bootstrap sign-tail p для geometry exploratory, не exact null test. Малые группы и
CI, пересекающие ноль, обозначают недостаточную определённость, не эквивалентность.
Сравнение AI human-IoU paired на одинаковых confirmed candidates не смешивается с
сравнением AI, имеющих разный набор кандидатов.

## 6. Панели, морфология и case studies

{panel_table}

{'; '.join(worst)}.

Полная стратификация priority/support/geometry/scope/outcome/stage/A-positive:
STRATIFIED_OBJECT_METRICS.tsv. ERROR_FLAGS.tsv и ERROR_TAXONOMY.md отделяют измеряемые
center/extent/orientation issues от неизвестных причин. Frozen поля не позволяют
доказать split/merge, decorative confusion, low contrast или occlusion по одному REJECT.
Калибровка не упражняла SPLIT/MERGE. Граничные STAR на f67v1 подтверждают видимость в скане,
но не принадлежность основной странице; ownership из panel ID не выводится.
Case-study overlays выбираются deterministic экстремумами и neutral ID tie-breaking
по заранее заданному правилу, source selection table — figures/CASE_STUDIES.tsv.

## 7. Sensitivity

{sensitivity_table}

{stability}

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
Uncertainty={percent(graph['uncertainty_explicit'])} среди explicit, и
{percent(graph['uncertainty_individual_review'])} среди individual reviews.

53 positive edges соединяют 51 LABEL и 51 STAR. {graph['stars_with_multiple_labels']} STAR имеют
несколько LABEL; {graph['labels_with_multiple_stars']} LABEL имеют несколько STAR.
Positive graph содержит {graph['positive_connected_components']} components:
`{graph['component_shapes']}`. Degree-zero confirmed endpoint означает только отсутствие
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

Seed={SEED}; bootstrap={BOOTSTRAPS}; call-order-independent metric-key seeds.
Результаты не содержат транскрипции, lexical matching, semantic crosswalk, M3 или
астрономической идентификации. Команды — REPRODUCIBILITY.md; автоматические gates,
twice-run byte comparison, upstream/source checksums и output checksums — VALIDATION_REPORT.md.

```text
{STATUS_LINES}
```
'''
    (out/'THREE_WAY_COMPARISON_REPORT.md').write_text(report,encoding='utf-8')
    summary=f'''# Краткий результат AI1 / AI2 / Human

Основной scope: production reviewed candidate union, не полный census страниц.

{overview}

{interpretations}

{consensus_conclusion}

{stability}

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
{STATUS_LINES}
```
'''
    (out/'THREE_WAY_COMPARISON_SUMMARY.md').write_text(summary,encoding='utf-8')
