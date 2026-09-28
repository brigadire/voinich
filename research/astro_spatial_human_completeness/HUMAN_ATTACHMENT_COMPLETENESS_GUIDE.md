# Инструкция attachment completeness (будущий этап)

Не начинайте сейчас. Сначала завершите object completeness и отдельную reconciliation;
создание нового object proposal не означает, что он уже допустимый endpoint.

После внешнего object freeze generator создаст новую CVAT task только для f68r1/f68r2
и отдельно утверждённого limited f68r3. Схема CVAT_ATTACHMENT_COMPLETENESS_SCHEMA.json,
изображения IMAGES.zip, annotations ANNOTATIONS.zip. Не использовать старые spatial flags.
Dry-run ZIP в этом пакете содержит в имени task NOT_FOR_PRODUCTION и не предназначен
для human annotation. Не создавайте relation tasks на остальных панелях.

На каждом pair-frame смотрите STAR_REFERENCE и LABEL_REFERENCE одновременно. Это crop
оригинальной панели без AI suggestions; масштаб/offset точно описан CROP_MANIFEST.tsv.
Объекты не перемещать/не создавать/не удалять. Если контекст недостаточен, откройте исходную
панель по нейтральному panel ID из queue manifest или поставьте UNCERTAIN. Ни близость,
ни правое положение подписи сами по себе не означают VISUAL_LABEL_OF.

В tag PAIR_ATTACHMENT выберите VISUAL_LABEL_OF / UNASSIGNED / UNCERTAIN и confidence.
Если scope/region действительно ошибочен — NOT_APPLICABLE_REGION: это отдельный вопрос
reconciliation, не отрицательное наблюдение. Один LABEL может визуально относиться к
нескольким STAR, и наоборот: проверяйте каждую допустимую пару, не навязывая уникальности.
Если видимая пара отсутствует из геометрической очереди, сообщите neutral endpoint IDs;
не переносите endpoints и не добавляйте скрытую semantic интерпретацию. Изменение окна/scope
требует новой очереди с provenance, а не подстановки UNASSIGNED за filtered pairs.

Если queue manifest содержит `pair_shapes_grouped=true`, копии STAR и LABEL на каждом
pair-frame имеют одинаковый `group_id`. В CVAT включите Appearance → Color By → Group:
два объекта пары будут показаны одним цветом. Не сбрасывайте группу. Каждая возможная
пара остаётся отдельным кадром, поэтому один endpoint может независимо участвовать в
нескольких связях и many-to-many не превращается в неоднозначный общий кластер.

Обычно прежние v2 decisions не повторяются. Специальная очередь с
`prior_v2_pairs_included_for_blind_recheck=true` включает их для новой слепой проверки:
старое решение не показывается и не перезаписывается, а результат сохраняется отдельно
как attachment-v3. Окончательный export всей relation task проходит strict validator.
