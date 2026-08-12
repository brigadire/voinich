# Анализ структуры корпуса

Репозиторий содержит основной конвейер анализа и эксперимент структурной нормализации:

1. корневая программа строит частотный словарь и статистику окружения из текста;
2. `dict-analyze` рассчитывает метрики для каждого токена и перехода;
3. `structural-analyze` строит интерпретируемые рейтинги структурных особенностей.
4. `sequence-analyze` независимо исследует точные последовательности непосредственно в исходном тексте.
5. `structural-normalize` строит complete-link классы по готовому структурному сходству.
6. `normalization-compare` сравнивает raw/normalized последовательности с matched random baseline.
7. `structural-validate` проверяет структурную нормализацию вне обучающей выборки и оценивает вклад отдельных классов.
8. `structural-profile-stability` разделяет нестабильность профилей, similarity, ближайших соседей и hard classes.

Типичный конвейер:

```text
исходный текст
    → dictionary.yaml
    → tokens_analysis.yaml
    → structural_analysis.yaml

исходный текст
    → sequence_analysis.yaml
```

## Требования

- Go 1.25.5 или новее;
- входной текст, в котором одна строка файла является одной анализируемой строкой корпуса;
- токены должны быть заранее очищены и разделены пробельными символами.

Установка зависимостей и проверка проекта:

```bash
go mod download
go test ./...
go vet ./...
```

При необходимости программы можно собрать отдельно:

```bash
mkdir -p bin
go build -o bin/dictionary-build .
go build -o bin/dict-analyze ./dict-analyze
go build -o bin/structural-analyze ./structural-analyze
go build -o bin/sequence-analyze ./sequence-analyze
go build -o bin/structural-normalize ./structural-normalize
go build -o bin/normalization-compare ./normalization-compare
go build -o bin/structural-validate ./structural-validate
go build -o bin/structural-profile-stability ./structural-profile-stability
```

## Быстрый запуск полного анализа

Из корня репозитория:

```bash
go run . data_work/ivtt_output_1786282555007.txt dataset/dictionary.yaml
go run ./dict-analyze dataset/dictionary.yaml dataset/tokens_analysis.yaml
go run ./structural-analyze -output structural_analysis.yaml
go run ./sequence-analyze \
  -input data_work/ivtt_output_1786282555007.txt \
  -output sequence_analysis.yaml
```

`structural-analyze` по умолчанию читает файлы из `dataset/`, поэтому первые два этапа можно пропустить, если готовый набор данных уже актуален.

`sequence-analyze` является независимой ветвью конвейера. Он читает исходный текст и не пытается восстанавливать цепочки из агрегированных соседей `dictionary.yaml`.

Полный пересчёт всех этапов, включая 100 random baselines и out-of-sample validation:

```bash
./run-full-analysis.sh
```

## 1. Генератор словаря

Исходный код находится в корневом [main.go](main.go). Программа читает обычный текст и создаёт YAML-словарь.

Запуск:

```bash
go run . <input.txt> [dictionary.yaml]
```

Если второй аргумент не указан, результат записывается в `output.yaml`.

Для каждого токена сохраняются:

- общее число появлений `count`;
- полное распределение абсолютных позиций в строке;
- полные таблицы непосредственных предшественников и последователей;
- число появлений в начале и конце строки.

Пример результата:

```yaml
- token: daiin
  count: 847
  position_in_string:
    - position: 0
      count: 160
  word_before:
    - token: chol
      count: 33
  word_after:
    - token: chey
      count: 13
  line_start_count: 160
  line_end_count: 131
```

Токенизация выполняется функцией `strings.Fields`: программа не удаляет пунктуацию и не приводит регистр. Например, `word`, `word,` и `Word` считаются разными токенами. Пустые строки не считаются строками корпуса.

После создания словаря должны выполняться инварианты:

```text
Σ line_start_count = Σ line_end_count = число непустых строк
Σ count − число строк = Σ переходов
```

## 2. Анализатор словаря `dict-analyze`

Исходный код находится в [dict-analyze/main.go](dict-analyze/main.go).

Запуск:

```bash
go run ./dict-analyze <dictionary.yaml> [tokens_analysis.yaml]
```

Выходной файл по умолчанию — `tokens_analysis.yaml`.

Для каждого токена рассчитываются:

- `P(start|token)` и `P(end|token)`;
- число уникальных предшественников и последователей;
- энтропия Шеннона левого и правого окружения в битах;
- все наблюдавшиеся переходы `A → B` и `P(B|A)`;
- обратный переход `B → A` и его вероятность;
- направленная асимметрия перехода;
- число и вероятность самопереходов;
- предварительные структурные оценки.

Основные формулы:

```text
P(start|A) = line_start_count(A) / count(A)
P(end|A)   = line_end_count(A) / count(A)
P(B|A)     = count(A→B) / ΣX count(A→X)

asymmetry(A,B) = (P(B|A) − P(A|B)) / (P(B|A) + P(A|B))
restriction(A) = 1 − H(neighbor|A) / log2(unique observed neighbors)
```

Асимметрия находится в диапазоне `[-1, 1]`: положительное значение означает преимущество направления `A → B`, отрицательное — `B → A`.

Вход проверяется на неизвестные YAML-поля, отрицательные значения, некорректные граничные частоты и дубликаты токенов.

## 3. Структурный анализатор `structural-analyze`

Реализация находится в директории [structural-analyze](structural-analyze). Программа совместно использует `dictionary.yaml` и `tokens_analysis.yaml`.

Запуск с параметрами по умолчанию:

```bash
go run ./structural-analyze
```

Полная форма:

```bash
go run ./structural-analyze \
  -dictionary dataset/dictionary.yaml \
  -analysis dataset/tokens_analysis.yaml \
  -output structural_analysis.yaml \
  -min-token-count 10 \
  -min-transition-count 3 \
  -min-context-observations 10 \
  -min-self-transition-count 3 \
  -reliability-prior 10 \
  -min-similarity 0.7 \
  -max-items 100 \
  -dominant-context-limit 5
```

### Параметры

| Флаг | Значение по умолчанию | Назначение |
| --- | ---: | --- |
| `-dictionary` | `dataset/dictionary.yaml` | Входной словарь |
| `-analysis` | `dataset/tokens_analysis.yaml` | Результат `dict-analyze` |
| `-output` | `structural_analysis.yaml` | Выходной YAML |
| `-min-token-count` | `10` | Минимальная частота токена для рейтингов |
| `-min-transition-count` | `3` | Минимальная частота перехода для раздела значимых переходов |
| `-min-context-observations` | `10` | Минимум наблюдаемых соседей для рейтингов predictability |
| `-min-self-transition-count` | `3` | Минимальное число самопереходов |
| `-reliability-prior` | `10` | Псевдочастота, уменьшающая рейтинг малых выборок |
| `-min-similarity` | `0.7` | Минимальное исходное сходство пары токенов |
| `-max-items` | `100` | Максимальное число записей в презентационных рейтингах; `0` отключает ограничение |
| `-max-equivalence-candidates` | `0` | Независимый лимит полного downstream-набора similarity-пар; `0` отключает ограничение |
| `-dominant-context-limit` | `5` | Число показываемых доминирующих соседей |

Пороговые значения применяются только к рейтингам. Редкие токены и переходы не удаляются из исходных корпусных распределений и ожидаемых частот.

### Разделы результата

`meta` содержит размеры корпуса и покрытие позиционных наблюдений. При загрузке проверяются равенство числа начал и концов строк, число переходов и согласованность обоих входных YAML-файлов.

`position_baseline` — среднее распределение абсолютных позиций по корпусу.

`positional_specialization` ранжирует отличие позиционного распределения токена от корпуса. `score` — дивергенция Дженсена—Шеннона с логарифмом по основанию 2, поэтому лежит в `[0, 1]`. В результат также входят исходные частоты позиций, корпусные вероятности, покрытие и множитель надёжности.

`successor_predictability` и `predecessor_predictability` описывают концентрацию наблюдаемого контекстного распределения:

```text
predictability = 1 − entropy / log2(unique observed neighbors)
```

Это мера предсказуемости среди наблюдавшихся соседей, а не доказательство грамматического или структурного ограничения. Запись включается только при наличии не менее `min_context_observations` соответствующих переходов.

`significant_transitions` содержит:

- фактическую и ожидаемую частоту `A → B`;
- `P(B|A)`;
- частоту и вероятность обратного направления;
- асимметрию;
- PMI;
- статистику логарифмического отношения правдоподобия (`G-test`) для таблицы `2×2`.

Раздел сортируется по `log_likelihood`. Знак ассоциации следует определять по `pmi`: высокий G-test может означать как избыток, так и дефицит переходов.

`self_transitions` сравнивает наблюдаемую частоту `A → A` с ожиданием при независимых началах и концах перехода:

```text
expected(A→A) = outgoing(A) × incoming(A) / all_transitions
enrichment    = observed / expected
reliability   = observed / (observed + reliability_prior)
ranking_score = enrichment × reliability
```

Сортировка выполняется по `ranking_score`, поэтому большой enrichment с минимально допустимым числом наблюдений не получает автоматического преимущества.

Самопереходы с частотой ниже `min_self_transition_count` не показываются.

`equivalence_candidates` ищет пары токенов с похожим поведением. Итоговое сходство — среднее из:

- позиционного сходства `1 − JSD`;
- косинусного сходства левых контекстов;
- косинусного сходства правых контекстов.

Сравнение имеет квадратичную сложность относительно числа токенов, прошедших `min-token-count`.

### Надёжность рейтингов

Для контекстных метрик применяется множитель:

```text
reliability = observations / (observations + reliability_prior)
ranking_score = raw_score × reliability
```

Для позиционных рейтингов он дополнительно умножается на долю сохранённых позиционных наблюдений. Генератор сохраняет все наблюдавшиеся позиции, поэтому для актуального словаря ожидается полное покрытие:

```yaml
meta:
  position_observations: 38887
  position_coverage: 1
```

В YAML всегда сохраняются исходная метрика, coverage, reliability и итоговый `ranking_score`, поэтому результат можно проверить и переинтерпретировать.

## 4. Анализатор последовательностей `sequence-analyze`

Программа в директории [sequence-analyze](sequence-analyze) считает реально наблюдавшиеся точные n-граммы. Она работает непосредственно с исходным текстом: статистика соседей из `dictionary.yaml` недостаточна, чтобы доказать существование конкретной триграммы или более длинной цепочки.

Запуск с параметрами по умолчанию:

```bash
go run ./sequence-analyze
```

Полный пример:

```bash
go run ./sequence-analyze \
  -input data_work/ivtt_output_1786282555007.txt \
  -output sequence_analysis.yaml \
  -min-n 2 \
  -max-n 8 \
  -min-count 2 \
  -max-items 200 \
  -context-limit 10 \
  -max-context-length 7 \
  -context-min-observations 10 \
  -context-max-items 200
```

| Флаг | По умолчанию | Назначение |
| --- | ---: | --- |
| `-input` | `data_work/ivtt_output_1786282555007.txt` | Исходный корпус |
| `-output` | `sequence_analysis.yaml` | Выходной YAML |
| `-min-n` | `2` | Минимальная длина n-граммы |
| `-max-n` | `8` | Максимальная длина n-граммы |
| `-min-count` | `2` | Минимальная частота для разделов повторов |
| `-max-items` | `200` | Максимум результатов отдельно для каждого `n`; `0` отключает лимит |
| `-context-limit` | `10` | Максимум отображаемых вариантов контекста |
| `-max-context-length` | `7` | Максимальная длина левого контекста для предсказания следующего токена |
| `-context-min-observations` | `10` | Минимум наблюдений длинного контекста для `context_extensions` |
| `-context-max-items` | `200` | Максимальное число расширений контекста; `0` отключает лимит |

Токенизация выполняется через `strings.Fields`. Каждая непустая физическая строка файла является отдельной последовательностью; n-граммы не пересекают границы строк. Содержимое токенов, регистр и пунктуация сохраняются без изменений.

Выходной `sequence_analysis.yaml` содержит:

- `meta` — число токенов, непустых строк и переходов с проверкой инварианта `tokens - lines = transitions`;
- `ngram_summary` — полную статистику для каждой длины, включая hapax и повторы; пороги на неё не влияют;
- `repeated_ngrams` — повторяющиеся точные последовательности с `count`, `line_count` и привязкой к границам строк;
- `cross_line_repeated_ngrams` — повторы, наблюдавшиеся минимум в двух различных строках;
- `continuations` — распределение непосредственно следующего токена, энтропию и predictability;
- `predecessor_contexts` — симметричное распределение предыдущего токена;
- `extensions` — доминирующие расширения последовательности слева и справа;
- `maximal_repeated_sequences` — повторы, которые нельзя расширить одним и тем же токеном без уменьшения частоты, с координатами всех появлений.
- `maximal_cross_line_sequences` — максимальные повторы, независимо воспроизводящиеся минимум в двух строках;
- `context_order_analysis` — условная энтропия следующего токена для левого контекста длиной 1…7 и показатели разреженности;
- `context_extensions` — случаи, где добавление ещё одного токена слева меняет энтропию следующего токена при достаточном числе наблюдений.

`count` учитывает все, в том числе перекрывающиеся, появления. `line_count` считает различные строки. Координаты используют физический номер строки с единицы и смещение токена с нуля.

Повторяющаяся последовательность и межстрочная повторяющаяся последовательность — не одно и то же:

```text
repeated:   count >= min-count
cross-line: count >= min-count и line_count >= 2
```

Повтор, встретившийся несколько раз только в одной строке, сохраняется в `repeated_ngrams` с `cross_line: false`, но не включается в `cross_line_repeated_ngrams`. В `ngram_summary` такие случаи отдельно учитывает `single_line_repeated`.

Предсказуемость контекста рассчитывается так:

```text
entropy = -Σ p × log2(p)
normalized_entropy = entropy / log2(unique_contexts)
predictability = 1 - normalized_entropy
```

Если наблюдался ровно один вариант контекста, `entropy` и `normalized_entropy` равны нулю, а `predictability` равна единице. Появление в конце строки не считается наблюдением продолжения; его отражает отдельное поле `line_end_count`. Для предшественников аналогично используется `line_start_count`.

Максимальная повторяющаяся последовательность имеет частоту не ниже `min-count`, и ни одно одинаковое расширение слева или справа длиной не более `max-n` не сохраняет все её появления. Более длинные цепочки на этом этапе не исследуются.

### Зависимость следующего токена от длины контекста

Для каждого `k` от 1 до `max-context-length` анализатор считает:

```text
H(next | previous k tokens)
```

Это взвешенное среднее энтропий распределения следующего токена для каждого наблюдавшегося контекста; вес равен числу продолжений этого контекста. `perplexity` равна `2^H`.

Чтобы отделить снижение энтропии от разреженности, результат содержит число единичных и повторяющихся контекстов, `singleton_fraction`, число наблюдений в повторяющихся контекстах и `repeated_context_coverage`. Также отдельно рассчитываются энтропия и perplexity только по контекстам с двумя или более наблюдениями.

Для `k >= 2` поля `entropy_delta_from_previous` и `repeated_entropy_delta_from_previous` показывают разность энтропий между длинами `k-1` и `k`. Положительное значение означает диагностическое снижение неопределённости, но не является свидетельством правила, грамматики или семантики. При малом coverage падение общей энтропии обычно объясняется тем, что почти все длинные контексты уникальны.

`context_extensions` сравнивает распределение после короткого контекста `B…` и после расширенного слева `A B…`. В раздел входят ненулевые изменения при наличии не менее `context-min-observations` наблюдений длинного контекста; отрицательный `entropy_reduction` означает рост энтропии. Сортировка выполняется по абсолютной величине изменения.

## 5. Эксперимент структурной нормализации

Эксперимент использует только `equivalence_candidates` из независимо построенного `structural_analysis.yaml`:

```text
structural_analysis.yaml
       ↓
structural-normalize
       ↓
normalized_070.txt … normalized_090.txt
       ↓
sequence-analyze
       ↓
normalization-compare + matched random baseline
       ↓
normalization_comparison.yaml
```

Структурные классы не являются семантическими классами. Similarity означает только сходство позиции и непосредственных левого/правого контекстов. При построении классов не используются n-граммы, результаты `sequence-analyze`, написание токенов, edit distance или сведения о содержании документа.

Основной воспроизводимый запуск:

```bash
./run-normalization-analysis.sh
```

Он строит пять заранее заданных моделей, запускает одинаковый sequence-анализ и выполняет по 100 matched random прогонов для каждого threshold.

Нормализатор можно запустить отдельно:

```bash
go run ./structural-normalize \
  -input data_work/ivtt_output_1786282555007.txt \
  -structural dataset/structural_analysis.yaml \
  -output normalized.txt \
  -classes structural_classes.yaml \
  -thresholds 0.70,0.75,0.80,0.85,0.90 \
  -singleton-mode preserve
```

Кластеризация — детерминированный agglomerative complete-link. Объединение допустимо, только если similarity известна для каждой пары будущего класса и каждая пара проходит общий и дополнительные component thresholds. Поэтому цепочка `A~B`, `B~C` не объединяет автоматически `A/B/C`, если `A~C` отсутствует или не проходит порог.

Дополнительные параметры:

| Флаг | По умолчанию | Назначение |
| --- | ---: | --- |
| `-min-position-similarity` | `0` | Минимум позиционного сходства |
| `-min-left-context-similarity` | `0` | Минимум сходства левого контекста |
| `-min-right-context-similarity` | `0` | Минимум сходства правого контекста |
| `-min-token-count` | из structural metadata, иначе `10` | Минимальная частота объединяемого токена |
| `-singleton-mode` | `preserve` | `preserve` сохраняет surface-токен; `class` назначает singleton ID |
| `-random-baselines` | `100` | Число контрольных прогонов |
| `-random-seed` | `1` | Базовый seed воспроизводимого контроля |

`structural_classes.yaml` хранит все члены и исходные минимальные/средние pairwise similarity, coverage нормализации и compression ratio. Идентификаторы `C0001…` детерминированы и не пересекаются с исходным алфавитом. Нормализация сохраняет физические границы строк, число токенов и число переходов.

`normalization-compare` строит случайные классы тех же размеров из токенов с `count >= min_token_count`. Частоты сопоставляются через логарифмические base-2 bins, выбор выполняется без возвращения, а при пустом bin используется ближайший. Каждый прогон определяется `random_seed`, threshold и номером прогона.

Сравнение сохраняет без универсального score:

- межстрочные повторы для каждого `n`;
- максимальную длину межстрочного повтора;
- conditional entropy и repeated-context conditional entropy для каждого `k`;
- repeated-context coverage;
- mean, stddev, min/max и перцентили random baseline;
- z-score и empirical p с поправкой `+1`.

Для количества повторов, длины и coverage используется верхний хвост `random >= structural`; для entropy — нижний хвост `random <= structural`. Рост повторяемости сам по себе ожидаем при уменьшении алфавита, поэтому его нельзя интерпретировать отдельно от compression ratio и matched random baseline. Ни классы, ни результаты эксперимента не являются выводами о значении токенов.

## 6. Out-of-sample validation

`structural-validate` выполняет детерминированную line-based cross-validation без утечки TEST-данных:

```bash
go run ./structural-validate \
  -input data_work/ivtt_output_1786282555007.txt \
  -classes structural_classes.yaml \
  -folds 5 \
  -fold-seed 1 \
  -threshold 0.70 \
  -random-baselines 100 \
  -random-seed 1 \
  -output structural_validation.yaml
```

В каждом fold словарь, позиционные и контекстные статистики, pairwise similarity и complete-link классы заново строятся только по TRAIN. Зафиксированные классы затем применяются к RAW TEST; неизвестные TRAIN токены остаются неизменными. Matched random модели также используют только размеры классов и частоты TRAIN.

Результат содержит метрики каждого fold для `n=2..8`, reconstructed surface-realizations новых межстрочных повторов, random distributions и empirical p, агрегированные и pooled fold-level counts, устойчивость пар токенов между TRAIN folds. Отдельные разделы `leave_one_class_out` и `member_ablation` оценивают концентрацию full-corpus эффекта, не меняя сами классы.

LOCO и member-ablation всегда используют заранее заданную full-corpus модель `threshold=0.70`, даже если CV запускается как вторичный sensitivity-анализ с другим threshold.

## 7. Устойчивость структурных профилей

`structural-profile-stability` исследует уже существующую геометрию similarity, не меняя формулу, веса, threshold или complete-link clustering:

```bash
go run ./structural-profile-stability \
  -input data_work/ivtt_output_1786282555007.txt \
  -classes structural_classes.yaml \
  -folds 5 \
  -fold-seed 1 \
  -min-token-count 10 \
  -neighbors 10 \
  -bootstrap-runs 200 \
  -bootstrap-seed 1 \
  -threshold 0.70 \
  -threshold-margin 0.05 \
  -output structural_profile_stability.yaml
```

Full corpus, каждый TRAIN, соответствующий TEST и каждый line-bootstrap sample получают независимые профили. Similarity остаётся средним `position_similarity = 1−JSD`, левого cosine и правого cosine. Eligibility проверяется отдельно в каждой выборке; отсутствие редкого токена не считается нестабильностью.

Результат содержит:

- same-token TRAIN–TRAIN и TRAIN–TEST stability компонентов;
- top-K Jaccard, top-1 recovery, overlap@3/5/10 и Spearman rank correlation;
- fold-level similarity, threshold crossings и margin для candidate pairs;
- 200-run bootstrap CI и вероятность `similarity >= 0.70`;
- зависимость от частоты токена и пары;
- подробный отчёт для всех актуальных full-corpus классов threshold 0.70;
- диагностический пересчёт similarity без каждой из трёх компонент, не используемый для новых классов.

Инструмент не читает `sequence_analysis.yaml`, normalization comparison или нормализованные корпусы и не выполняет семантической интерпретации.

## Структура репозитория

```text
.
├── main.go                    # генератор dictionary.yaml
├── dict-analyze/              # анализ каждого токена и перехода
├── structural-analyze/        # корпусные рейтинги и группы сходства
├── sequence-analyze/          # точные n-граммы исходного текста
├── structural-normalize/      # complete-link нормализация
├── normalization-compare/     # raw/structural/random сравнение
├── internal/normalization/    # общее ядро классов и random matching
├── structural-validate/       # out-of-sample validation и ablation
├── internal/validation/       # TRAIN-only статистики, split и метрики
├── structural-profile-stability/ # устойчивость структурной геометрии
├── internal/profilestability/ # profile/fold/bootstrap/rank расчёты
├── run-full-analysis.sh       # полный пересчёт конвейера и экспериментов
├── dataset/                   # готовые входы структурного анализатора
├── data/ и data_work/         # исходные и подготовленные тексты
├── tasks/                     # формулировки задач
├── structural_analysis.yaml   # пример итогового результата
└── sequence_analysis.yaml     # анализ наблюдавшихся последовательностей
```

## Тестирование

```bash
go test ./...
go vet ./...
```

Тесты проверяют подсчёт окружений и позиций, формулы вероятностей, PMI, самопереходы, сходство контекстов, точные n-граммы, границы строк, контексты последовательностей, максимальные повторы, координаты, пороги, детерминированную сортировку, отсутствие TEST leakage, fold-инварианты, стабильность классов и ablation.
