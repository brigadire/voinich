# Инструкция: связи группами на полных страницах

Создайте новую CVAT-задачу. Вставьте CVAT_SCHEMA.json, загрузите IMAGES.zip, затем
импортируйте ANNOTATIONS.zip как CVAT for images 1.1. Не используйте две отозванные
crop-задачи.

Задача содержит три полные страницы: f68r1, f68r2, f68r3. STAR/LABEL нельзя перемещать,
изменять, удалять или создавать повторно. Разрешено менять только grouping и tag
PAGE_RELATION_REVIEW.

Включите Appearance -> Color By -> Group. Все 48 прежних подтверждённых v2-связей уже
представлены 46 группами: 45 групп содержат одну STAR и один LABEL; одна группа f68r3
содержит один LABEL и три STAR. Прежние UNASSIGNED/UNCERTAIN не сгруппированы.

Если начальная группа неверна, разгруппируйте все её объекты. Для новой визуальной связи
выберите соответствующие LABEL и STAR и примените Group Shapes (`G`). Группа обязана
содержать хотя бы один LABEL и одну STAR. Можно включить несколько LABEL/STAR, если они
образуют одну визуально связанную структуру. Состав группы сохраняется как hyperedge;
валидатор не создаёт скрытое декартово произведение пар.

Негруппированный объект означает, что в итоговой групповой разметке визуальная связь для
него не задана. После систематического просмотра страницы в PAGE_RELATION_REVIEW выберите
completion_status=COMPLETE и human_confidence=HIGH (либо фактическую уверенность).
Экспортируйте всю задачу, все три страницы, как CVAT for images 1.1.

Проверка экспорта:

```bash
python3 scripts/full_page_groups.py validate \
  --input /path/to/export.zip \
  --task-dir job_18_attachment_v3_full_pages_grouped \
  --reviewer-id H01 --timestamp 2026-09-15T12:00:00Z \
  --output-dir /tmp/job18-group-review
```

Не редактируйте TSV. Валидатор сохранит canonical group IDs, membership и отдельный audit
удалённых/изменённых/новых групп; frozen attachment-v2 останется неизменным.
