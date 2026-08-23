# Аудит stage 13: `structural-pair-decompose`

Дата проверки: 2026-08-20

## Итог

Увеличение времени выполнения вызвано не зависанием и не проблемой удалённых workers. В текущем эксперименте stage 13 получил существенно больше пар для обработки из-за одной аномально большой structural family.

Проверенный эксперимент:

`experiments/doyle__transposition__w008__natural__seed001-v1`

Stage 13 завершился успешно и занял:

- wall time: `1702.238664243` секунд, примерно 28 минут;
- user CPU: `658.947832` секунд;
- system CPU: `257.203436` секунд;
- максимальная RSS: `31,257,956 KB`, примерно 31 ГБ.

Источник измерений: [run-state.json](../../experiments/doyle__transposition__w008__natural__seed001-v1/run-state.json), запись `structural-pair-decompose`.

## Что отличалось от обычных запусков

Обычные Voynich-запуски обрабатывали примерно `26–29` пар, а текущий запуск обработал:

```text
Decomposed 3170 pairs and 1 families into workdir
```

В текущем запуске:

- `structural_graphemic_pairs.tsv`: `114,961` строк пар;
- `structural_distant_top.tsv`: около 200 кандидатов;
- одна family содержит `3,315` токенов;
- эта family содержит `3,170` рёбер.

Данные family находятся в [structural_distant_families.yaml](../../experiments/doyle__transposition__w008__natural__seed001-v1/workspace/workdir/structural_distant_families.yaml).

## Причина роста объёма

В [internal/pairdecomposition/analyze.go](../../internal/pairdecomposition/analyze.go) stage 13 сначала выбирает обычные top-N distant pairs, но затем без ограничения добавляет все рёбра всех structural families:

```go
for _, k := range distant[:n] {
    add(key(k[0], k[1]))
}
for _, f := range families {
    for _, e := range f.Edges {
        add(key(e.TokenA, e.TokenB))
    }
}
```

Поэтому family из `3,170` рёбер превратила ожидаемый небольшой анализ в анализ тысяч пар.

Для каждой целевой пары дополнительно выбираются controls. `chooseControls` проходит по всему массиву из примерно `115 тыс.` пар и сортирует кандидатов — [internal/pairdecomposition/analyze.go:492](../../internal/pairdecomposition/analyze.go#L492). При трёх controls это добавляет ещё примерно `9,510` decompositions.

После вычислений stage записывает:

- `pair_decomposition.yaml` размером около `1.1 ГБ`;
- `structural_pair_report.md` размером около `11 МБ`;
- SVG-график для каждой из `3,170` пар.

Генерация графиков выполняется для всех `out.Pairs` в [internal/pairdecomposition/write.go:29](../../internal/pairdecomposition/write.go#L29).

## Вывод

Stage 13 выполнялся долго из-за комбинации факторов:

1. Большая family с `3,315` токенами и `3,170` рёбрами.
2. Автоматическое добавление всех family edges к top-N парам.
3. Полный проход по всем парам для выбора controls для каждой target pair.
4. Генерация тысяч SVG и сериализация гигантского YAML.

Это объясняет отличие от обычных запусков. Эксперимент не был заблокирован: stage 13 завершился с exit code `0`; на момент фиксации аудита уже выполнялся stage 14 `distance-context-analyze`.
