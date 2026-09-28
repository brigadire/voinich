# Упрощённая задача LABEL–STAR: визуальная привязка, v2

Вопрос: **«Этот LABEL визуально является подписью к этой STAR?»** Не читайте текст, не определяйте название звезды и не оценивайте просто расстояние/соседство. Старую spatial-задачу на 324 пары больше не используйте для этой оценки: её пакет сохранён отдельно.

## Основная задача

Создайте **новую** задачу CVAT, не продолжайте старую с пятью checkbox-полями.

1. В Raw labels вставьте `relations_v2/CVAT_LABEL_STAR_ATTACHMENT_SCHEMA.json`, нажмите Done. Имена классов: LABEL_ENDPOINT и STAR_ENDPOINT.
2. Загрузите **изображения** из `relations_v2/main/IMAGES.zip`.
3. Импортируйте **аннотации** из `relations_v2/main/ANNOTATIONS.zip` в формате CVAT for images 1.1.

Задача содержит 99 пар: f68r1 — 29, f68r2 — 19, f68r3 — 51. Ещё 3 кадра ZZ_CONTEXT — полные исходные листы только для справки; размечать их не нужно.

Каждый обычный кадр содержит одну голубую подпись и одну оранжевую звезду. Выберите **оранжевый STAR_ENDPOINT**, затем измените только relation_decision:

- ASSIGNED — это визуально обоснованная подпись к данной звезде.
- UNASSIGNED — такой привязки нет.
- UNCERTAIN — по изображению решить нельзя.

UNREVIEWED — начальное значение, в готовом экспорте его оставлять нельзя. Checkbox-полей отношений больше нет. human_confidence по умолчанию HIGH согласно вашему указанию. Геометрию и ID не меняйте; новые объекты или пары не создавайте.

На f68r1 и f68r2 очередь оставляет только LABEL справа от STAR с перекрытием по высоте ориентированных envelope. Это грубый фильтр очереди, а не автоматическое ASSIGNED. Если пара всё же не подходит — UNASSIGNED. Если фильтр исключил нужную пару, сообщите её ID: потребуется новая версия очереди, а не перенос существующего объекта.

## Отдельная задача f68v2

Создайте вторую новую задачу с той же схемой. Данные: `relations_v2/f68v2/IMAGES.zip`; аннотации: `relations_v2/f68v2/ANNOTATIONS.zip`. В ней 26 пар и 1 обзорный кадр.

На f68v2 ничего не принято автоматически: начальное значение UNREVIEWED. Если два существующих блока текста относятся к одной звезде, поставьте ASSIGNED для **обеих пар**. Один STAR может иметь несколько LABEL. LABEL объединять или разделять для этого не нужно. При неясной привязке используйте UNCERTAIN. Если нужная пара отсутствует, сообщите об этом, не перемещайте endpoint на другой объект.

## Уже обработанные и исключённые пары

- 142 пары на f67r1, f67v1 и f68v1 получили UNASSIGNED по вашему явному решению на уровне листа. Они записаны в `relations_v2/REVIEWER_PANEL_DECISIONS.tsv` с происхождением REVIEWER_PANEL_RULE, а не как индивидуально просмотренные пары.
- 57 пар f68r1/f68r2 не прошли правосторонний фильтр и перечислены в `relations_v2/RIGHT_RULE_FILTERED.tsv` как RULE_FILTERED_NOT_REVIEWED. Они не объявлены отрицательными ручными наблюдениями.
- Исходные STAR/LABEL, их frozen-протокол v1.2, старый relation-пакет и raw exports не изменены. Подробные правила новой задачи заморожены в `HUMAN_LABEL_STAR_ATTACHMENT_PROTOCOL_V2.md`.

## Экспорт

Экспортируйте обе задачи как CVAT for images 1.1 и положите ZIP в exports. Импортёр новой версии:

```bash
python3 scripts/import_label_star_attachment_v2.py --subset main --input exports/reviewed_main.zip --output exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_MAIN_R01.tsv --reviewer-id R01
python3 scripts/import_label_star_attachment_v2.py --subset f68v2 --input exports/reviewed_f68v2.zip --output exports/HUMAN_LABEL_STAR_ATTACHMENT_V2_F68V2_R01.tsv --reviewer-id R01
```

ASSIGNED записывается как VISUAL_LABEL_OF, с relation_protocol_version=2. UNASSIGNED/UNCERTAIN сохраняются явно. Импортёр проверяет completeness, неизменность endpoint ID/геометрии, отсутствие старых spatial checkbox-полей и входные frozen-хеши. Результаты новой версии нельзя смешивать с пространственными отношениями версии 1 без явного указания протокола.
