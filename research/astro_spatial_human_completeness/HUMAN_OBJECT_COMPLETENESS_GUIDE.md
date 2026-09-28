# Инструкция reviewer: добавлять только пропуски

## Импорт в CVAT

Создайте НОВУЮ задачу, загрузите OBJECT_COMPLETENESS_IMAGES.zip (8 исходных JPEG).
Вставьте CVAT_OBJECT_COMPLETENESS_SCHEMA.json в редактор labels. Не используйте схемы
STAR_OBJECT/LABEL или прежних relation tasks: имена классов здесь другие.
Импортируйте OBJECT_COMPLETENESS_CVAT.xml либо OBJECT_COMPLETENESS_CVAT.zip как
**CVAT for images 1.1**. ZIP аннотаций не заменяет загрузку изображений.
До работы сделайте пустой export этого импорта и проверьте валидатором режим
`--allow-incomplete`; это runtime preflight вашей версии CVAT, не human completion.

На вкладке Labels заблокируйте STAR_REFERENCE, LABEL_REFERENCE и UNCERTAIN_REFERENCE,
если ваша версия предоставляет Lock. UNCERTAIN_REFERENCE обычно скрывайте. Не блокируйте
два HUMAN_ADDED класса или PANEL_COMPLETENESS. Даже при Lock изменение references
проверяется экспортным валидатором. Установите малую/нулевую заливку, чтобы видеть marks.
Никакой reference не нужно снова принимать, подтверждать или исправлять.

## Для каждой панели

1. Посмотрите всю панель и карту navigation/<panel>.png.
2. По порядку T01…T12 (слева направо, сверху вниз) ищите отсутствующие STAR; переключайте
   STAR_REFERENCE. Вернитесь к целой панели и отметьте star_review_done.
3. Повторите T01…T12 для LABEL; переключайте LABEL_REFERENCE, отметьте label_review_done.
4. Отдельно проверьте края, круговые области, слабые/повреждённые участки и соответствующие
   три checkboxes. На f67v1 видимые объекты соседней страницы в захваченном скане не
   означают принадлежности основной странице; не создавайте для них semantic relations.
5. Добавляйте только явно отсутствующее: STAR_HUMAN_ADDED = box, LABEL_HUMAN_ADDED =
   rotated box для одной физической строки или ellipse для полного непрерывного кольца.
   Угол меняется rotation handle прямоугольника/ellipse; не заменяйте наклонный run большим
   axis-aligned bbox. Несколько слов одной строки — один LABEL; текст не переписывать.
6. У каждого proposal поставьте confidence LOW/MEDIUM/HIGH. Сомнение — addition_status
   UNCERTAIN_NEW. Возможный дубликат: сначала включите optional UNCERTAIN_REFERENCE;
   если вопрос остаётся, только явно flagged proposal POSSIBLE_DUPLICATE_NEW, не обычное
   повторное создание уже подтверждённого объекта.
7. В существующем tag PANEL_COMPLETENESS выберите COMPLETE_NO_NEW_OBJECTS или
   COMPLETE_WITH_NEW_OBJECTS (учитываются все proposals, включая doubtful). Все пять
   checkboxes должны быть true. Если просмотр нельзя завершить, выберите INCOMPLETE_*.

[Короткие схемы четырёх ситуаций](figures/OBJECT_REVIEW_EXAMPLES.svg): missing object;
неточная на ваш взгляд reference-граница (не изменять); возможный дубликат (флаг);
сомнительный новый объект (UNCERTAIN_NEW).

Экспортируйте всю задачу (все 8 panels), CVAT for images 1.1. Не удаляйте rejected-looking
reference и не меняйте его атрибуты: исправление предыдущего слоя вне этого этапа.
Пустой export без COMPLETE marker не считается отсутствием пропусков. Не загружайте
dry_run материалы: они синтетические NOT_FOR_PRODUCTION. В рамках подготовки работу
reviewer не запускали; следующая стадия запускается только по вашему решению.

Runtime preflight не завершается человеком автоматически; пример команды (после экспорта):

```bash
python3 scripts/validate_objects.py --input /path/to/empty-roundtrip.zip --reviewer-id R01 --timestamp 2026-09-14T12:00:00Z --allow-incomplete
```

После реального review обычная validation без --allow-incomplete, staging только в новый
каталог; это не принятие additions и не object freeze. Точные команды: REPRODUCIBILITY.md.
