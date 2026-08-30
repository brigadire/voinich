# PHASE2_FAILURE_ANALYSIS

## Why the mechanism-identification experiment failed

**Статус документа:** научная ретроспектива Phase II. Слово *failure*
относится к невозможности идентифицировать механизм и к несоответствию
экспериментального дизайна этой цели. Оно не означает, что гипотеза Фонтаны
была опровергнута, что вся Phase II недействительна или что отрицательный
результат Task83r был технической ошибкой.

## Краткий ответ

Основная проблема Phase II, вероятнее всего, состояла в **неидентифицируемости
механизма по выбранному пересечению моделей и наблюдений**.

Phase II надёжно установила, что текстовая поверхность VM структурирована, и
показала, что несколько типов внешне обусловленного восстановления можно
формально реализовать. Но финальный эксперимент сопоставлял эти два результата
только по 3 из 13 CORE-метрик Fingerprint V2, причём все три относились к одной
семье edit-graph. Остальные 10 CORE-измерений — иерархия, locus/folio,
физические строки и границы, позиционная и 2D-организация — модели напрямую не
воспроизводили. Поэтому сходство endpoint'ов не могло стать свидетельством о
механизме.

В обозначениях

\[
V = E_\theta(X,K), \qquad X = D_\theta(V,K),
\]

где `V` — наблюдаемая запись VM, `X` — исходная информация, `K` —
конвенция, контекст или память пользователя, а `theta` — механизм, Phase II:

1. частично исследовала зависимость восстановления от `K` на формальных
   моделях;
2. достаточно широко измерила структуру `V`;
3. но не имела наблюдаемых троек `(X, K, V)` для VM и не показала, что разные
   значения `theta` порождают различимые распределения наблюдаемого `V`.

Следовательно, финальный вывод `MECHANISM_IDENTIFICATION_FROM_F2 =
NOT_IDENTIFIABLE` является закономерным. Он означает **не “механизмы
эквивалентны” и не “external memory опровергнута”, а “данный эксперимент не
может выбрать механизм”**.

## 1. Что именно анализировалось

Ретроспектива опирается только на замороженную и авторитетную цепочку Phase II:

- построение и проверку Fingerprint V2/V2.1;
- исторический аудит *Secretum* и реконструкции Fontana-derived механизмов;
- формальное пространство механизмов и эксперименты восстановления;
- исторический abbreviation-контроль BDD и selective-extraction controls;
- повторный confirmatory experiment Task83r.

Первоначальный Task83 не используется как научное свидетельство: он был
правильно объявлен недействительным из-за provenance mismatch. Task83a выявил
дополнительную недетерминированность старого анализа; Task83b выполнил
детерминированный scientific refreeze без изменения 13 CORE-статусов; Task83r
провёл новый target opening и является авторитетным финальным сравнением
([Task83](task83/TASK83_REPORT.md), [Task83a](task83a/TASK83A_REPORT.md),
[Task83b](task83b/TASK83B_REPORT.md), [Task83r](task83r/TASK83R_REPORT.md)).

Эта техническая история важна для provenance, но **не объясняет научный
`NOT_IDENTIFIABLE`**: после исправления и чистого повторного эксперимента он
сохранился.

## 2. Empirical: что Phase II действительно установила о VM

### 2.1. Устойчивая структурированность поверхности

Fingerprint V2 обнаружил статистически выраженную организацию на нескольких
масштабах. В основном ZL3b-анализе были поддержаны, среди прочего:

| Наблюдаемое свойство | Статистика | Результат относительно frozen null |
| --- | ---: | ---: |
| позиционная специализация токенов в строке | LS2 NMI = 0.11059 | 76.57 SD, q = 0.00142 |
| асимметрия длины на границах строки | LS3 = 0.36065 | 16.56 SD, q = 0.00142 |
| связь токенов с границей строки | BP1 NMI = 0.08422 | 86.86 SD, q = 0.00142 |
| различие locus types | LC1 NMI = 0.00759 | 23.40 SD, q = 0.00142 |
| различие labels/text | LC2 NMI = 0.00485 | 25.56 SD, q = 0.00142 |
| folio coherence | PF2 = 0.17304 | 29.69 SD, q = 0.00142 |
| within-folio progression | PF5 = 0.36370 | 14.72 SD, q = 0.00131 |
| доля variance на уровне folio | HR1 = 0.27327 | 52.50 SD, q = 0.00142 |
| доля variance на уровне section | HR1 = 0.12526 | 181.54 SD, q = 0.00142 |
| layout-position dependence | 2DL1 MI = 0.00471 | 17.32 SD, q = 0.00142 |

После исправления специализированного null была также поддержана
recto/verso coherence: PF4 = 0.5192 против null mean 0.362, эффект 8.54 SD,
эмпирическое \(p \approx 0.001\). Это свидетельство о сходстве страниц одного
физического листа, но не о семантической связи, ключе или шифре
([Task79](fingerprint/TASK79_REPORT.md),
[Task79c](fingerprint/TASK79C_REPORT.md)).

Все 13 CORE-метрик сохранили направление и статус на независимой транскрипции
IT2a: 3 `STABLE`, 10 `DIRECTION_STABLE`, 0 `UNSTABLE`. При этом 10 из 13
сдвинулись более чем на одну SD соответствующего permutation null, поэтому
устойчивость качественного вывода не равна численной тождественности.

### 2.2. Локальные token families реальны, но не диагностичны

В VM существует большая воспроизводимая сеть edit-distance-one отношений:
2,294 из 4,240 направленных типов правил имели support не менее 3, а
support-threshold family structure воспроизводилась между половинами folio
(ARI 0.719, NMI 0.599).

Но более сильный C-GRAMMAR null, сохраняющий длины, позиции и биграммную
структуру, воспроизводил edit graph не хуже. Наблюдаемые EF1/EF2/EF3 были даже
ниже его ожидания на 13.1, 11.0 и 4.6 SD соответственно; итог:
`EDIT_FAMILIES_EXCEED_C_GRAMMAR_NULL = NOT_SUPPORTED` и
`EF4 = CONSISTENT_WITH_GRAMMAR_BOUND`
([Task77](fingerprint/TASK77_REPORT.md)).

Это один из самых ранних прямых сигналов будущей проблемы: **устойчивая
локальная структура существует, но она не уникальна для интересующего класса
образования текста**.

### 2.3. Что эти данные не устанавливают

Ни один из перечисленных эффектов сам по себе не определяет:

- естественный язык;
- shorthand или abbreviation;
- внешний mnemonic key;
- шифр;
- procedural/template generation;
- hidden channel;
- осмысленность или бессмысленность текста.

Например, boundary effects устанавливают организацию границ, но не acrostic,
telestich, grille или направление чтения. Variance-share по folio и section не
дала надёжного улучшения прогноза невиданного folio относительно flat model:
HR3/HR5 остались `INCONCLUSIVE`. То есть Phase II получила сильное описание
поверхности, но не семантическую или причинную интерпретацию этой поверхности.

## 3. Historical: что следует из Fontana, а что было экстраполяцией

### 3.1. Поддержанный исторический минимум

Аудит *Secretum* поддержал более узкий тезис, чем “Fontana объясняет VM”:

- в источнике присутствуют материальные и мысленные системы, использующие
  порядок, ключ, индекс, выравнивание, путь, ассоциацию и сигнал;
- часть таких систем хранит конфигурацию или состояние вне непосредственной
  памяти человека;
- видимый знак часто не является самодостаточным и требует известной
  конвенции или ассоциации;
- symbolic writing Фонтаны — отдельная monoalphabetic substitution system;
  наличие шифра не превращает все mnemonic devices в один генератор текста.

Итог источниковедческой ветви был
`FONTANA_EXTERNAL_MEMORY_PARTIALLY_SUPPORTED`, а не подтверждение авторства,
влияния или единого “метода Фонтаны”
([Task74](fontana/TASK74_REPORT.md)).

Task78 и Task80 усилили именно этот ограничительный вывод. Реконструкции
распались как минимум на literal external storage, indexed opaque cue, cyclic
reference и temporal associative cue. Общего encoder/decoder для них нет;
`SINGLE_UNIVERSAL_BASIS_SUPPORTED = NOT_SUPPORTED`, а heterogeneity —
`SUPPORTED` ([Task78](fontana/task78/TASK78_REPORT.md),
[Task80](fontana/task80/TASK80_REPORT.md)).

Следовательно, “Fontana-like” не было одной причинной гипотезой с единственным
предсказанным fingerprint. Это было семейство разнородных, частично
реконструированных механизмов.

### 3.2. Чего в исторических данных не было

Не существовало большого корпуса исторически засвидетельствованных троек

\[
(X, K, E_{Fontana}(X,K)),
\]

который позволял бы независимо оценить распределение выходов механизма,
обучить его параметры и проверить восстановление на held-out сообщениях.
Формальные пары, созданные реконструкциями Phase II, проверяют согласованность
**выбранной реализации**, но не доказывают, что именно такое преобразование
исторически применялось для создания протяжённого текста.

F01 speculum показал exact recovery 24/24 при полном `K`, а абляции показали
зависимость от порядка колец, направления, радиуса и общего принципа. Любое
повреждение используемого кольца обнуляло exact recovery в 170/170
соответствующих испытаний; встроенной коррекции ошибок не было. Это хорошее
свидетельство того, что реконструированный механизм может быть внешним
хранилищем, но не свидетельство, что он порождает статистику VM
([Task76](fontana/f01_speculum/TASK76_REPORT.md)).

Единственный реальный крупный aligned abbreviation-control в Phase II — BDD:
7,150 пар `<abbr>`/`<expan>` из одной рукописи/писцовой традиции. Он полезен
для проверки abbreviation transformation, но не является положительным
корпусом Fontana mnemonic encoding. Межтрадиционная репликация отсутствовала
([Task82b](task82b/TASK82B_REPORT.md)).

Итак, исторический материал обосновывал **возможность и компоненты** механизмов,
но не задавал идентифицируемую статистическую модель происхождения VM.

## 4. Identifiability: почему одинаковые признаки не выбирают механизм

### 4.1. Формальный критерий

Пусть `theta` обозначает класс механизма, а `K` — скрытый контекст. После
маргинализации неизвестных `X` и `K` механизм задаёт распределение

\[
P_\theta(V)=\sum_{x,k}P(V\mid x,k,\theta)P(x,k\mid\theta).
\]

Механизм идентифицируем по наблюдаемой записи только если разные допустимые
`theta` задают различимые \(P_\theta(V)\), с учётом nuisance-параметров и
неизвестного `K`. Если существуют \(\theta_1\ne\theta_2\) и допустимые
распределения скрытых переменных, для которых

\[
P_{\theta_1}(V)=P_{\theta_2}(V),
\]

то никакой классификатор, использующий только `V`, не может в общем случае
восстановить истинный механизм. Более сложный DAG или больше перестановок
уменьшат вычислительную ошибку, но не устранят это равенство.

### 4.2. Empirical equifinality в Phase II

Ещё до финального сравнения F2 показывал family-level trade-offs: BDD был
ближе на lexical-paradigm, edit-family и 2D-LITE, а procedural MS-DOS — на
line family; BDD, Doyle и MS-DOS были одновременно Pareto-nondominated.
Разные известные процессы, следовательно, воспроизводили разные проекции
структуры VM без единого победителя.

В Task83r описательные median distances были:

| Класс | ZL3b | IT2a | Confirmatory support |
| --- | ---: | ---: | --- |
| natural text | 0.674952 | 0.683933 | `PARTIAL` |
| shorthand | 0.921503 | 0.930485 | `DISFAVORED`, S0 |
| simple null | 1.110219 | 1.102676 | gate не пройден |
| Fontana-derived | 1.247039 | 1.239497 | `PARTIAL`, LEVEL_1 |
| selective extraction | 1.237889 | 1.246870 | `PARTIAL`, A1 |

Natural text оказался descriptively closest, но `closest != supported`.
Интервалы endpoint'ов всех обязательных попарных сравнений пересекались, и
ни один класс не прошёл собственный multi-family/null-separation gate. Поэтому
все pairwise results — `NO_CLEAR_ADVANTAGE`
([equifinality table](task83r/EQUIFINALITY_ANALYSIS.tsv),
[pairwise table](task83r/HYPOTHESIS_PAIRWISE_ADVANTAGE.tsv)).

Строго говоря, Phase II не доказала confirmatory equifinality самих классов:
для этого сначала нужны поддержанные модели. Она показала более осторожный
результат — **несколько описательных сходств при недостаточном измерительном
пересечении**, из-за чего equifinality осталась unresolved, а mechanism
identification — `NOT_IDENTIFIABLE`.

### 4.3. Главный измерительный bottleneck

Каждый финальный класс имел direct coverage только 3/13 CORE, все в edit
family. Projection давала ещё 4/13, но оставалась assembler-defined и потому
не могла смешиваться с direct evidence. Для всех классов оставались
`NOT_MODELLED`:

- hierarchy;
- locus и folio;
- page и recto/verso;
- physical line и boundary structure;
- positional и 2D organization;
- local regimes.

Отсутствующая метрика не является противоречием модели. Но по этой же причине
модель, совпавшая по edit family, не объясняет рукопись как многоуровневый
объект. Это зафиксировано непосредственно в
[coverage accounting](task83r/COVERAGE_ACCOUNTING.tsv) и названо strongest
common counterevidence в
[Task83r counterevidence](task83r/STRONGEST_COUNTEREVIDENCE.md).

## 5. Information-theoretic: что меняет неизвестный контекст K

### 5.1. Две разные задачи

Нужно различать:

1. **recoverability:** можно ли восстановить `X` из `V` при наличии или
   отсутствии `K`?
2. **mechanism identifiability:** можно ли определить `theta` по наблюдаемому
   `V`?

Идеальный external-memory случай может иметь

\[
H(X\mid V,K)\approx 0, \qquad H(X\mid V)>0,
\]

то есть `K` несёт условно необходимую информацию:

\[
I(X;K\mid V)=H(X\mid V)-H(X\mid V,K)>0.
\]

Но из этого не следует, что поверхность `V` имеет уникальный fingerprint.
Два разных механизма могут иметь одинаковое \(P(V)\), но совершенно разные
decoder, `K` и восстановимое `X`.

### 5.2. Что Phase II смогла измерить

Task81/82 явно разделили Convention, Geometry/Path, Context и
InternalMemoryState и проверили carrier removals, wrong-knowledge controls,
collisions и recovery classes. Некоторые формальные механизмы были exact при
R0; randomized convention/path/association/index controls разрушали
восстановление, показывая специфичность нужного знания. То есть зависимость
модельного декодирования от `K` была не только философской идеей
([Task82](task82/TASK82_REPORT.md)).

Однако Task82 был target-blind и не сравнивал эти recovery variables с VM.
Task82a масштабировал observable documents, но все 16 механизмов остались лишь
`PARTIALLY_COMPARABLE`; Task82a.1 получил максимум 3/13 direct CORE и 4/13
projection CORE. В Task83r valid Fontana before/after trajectory отсутствовала
и была `NOT_TESTABLE`.

Иными словами:

\[
\text{knowledge dependence demonstrated in models}
\not\Rightarrow
\text{knowledge dependence demonstrated for VM}.
\]

### 5.3. Почему один expansion ratio ничего не решает


Условие \(|X|/|V|>1\) совместимо с abbreviation, stenography, codebook
lookup, morphological normalization, templating и cue-based recovery. Оно не
различает:

- deterministic и stochastic decoding;
- one-to-one, one-to-many и many-to-one mapping;
- локальную и глобальную восстанавливаемость;
- истинную потерю информации и недоступность без ключа;
- внутренне содержащуюся и внешне добавляемую информацию.

Именно поэтому expansion ratio и expansion ambiguity были оставлены
`EXPLORATORY_ONLY`: у VM нет aligned plaintext/expansion, на котором их можно
оценить ([Task79b](notation-audit/TASK79B_REPORT.md),
[F2 admission](notation-audit/F2_ADMISSION.tsv)).

Корректный тест сильной external-memory гипотезы требовал бы paired data и
оценки хотя бы \(H(X\mid V)\), \(H(X\mid V,K)\), conditional candidate-set
size, recovery error under ablations и зависимости decoder от локального или
глобального контекста. По одному `V`, без независимо заданных `X` и `K`,
эти величины для VM не наблюдаемы.

### 5.4. Важная граница вывода

Наличие скрытого `K` — правдоподобное объяснение неидентифицируемости, но
Phase II **не установила, что такой `K` реально существует для VM**.
Ближайший к данным вывод слабее:

> Наблюдаемая поверхность VM и доступный F2 не содержат достаточного
> идентифицирующего сигнала, чтобы выбрать среди протестированных классов;
> модели с внешним `K` остаются возможными, но не подтверждёнными.

## 6. Experimental design: где разошлись вопрос и тест

### 6.1. Были смешаны три уровня утверждений

Фактически исследовались три разных вопроса:

| Уровень | Вопрос | Что получилось |
| --- | --- | --- |
| историческая возможность | существовали ли около XV века системы cue/state/convention-dependent recovery? | частично поддержано источниками Фонтаны |
| механизм как формальная возможность | могут ли реконструированные системы хранить/восстанавливать данные с зависимостью от `K`? | поддержано для ряда моделей и условий |
| происхождение VM | порождён ли VM таким механизмом? | не идентифицировано |

Первые два результата не образуют доказательства третьего без уникальных
observable predictions и adequate controls.

Дополнительно были неявно соединены:

- **A:** имеет ли VM сильное свойство внешней памяти — запись систематически
  недоопределена без пользовательского знания?
- **B:** воспроизводят ли конкретные Fontana-derived реконструкции признаки VM?

Отрицательный или неидентифицируемый B не отвергает A. Положительный endpoint
для B также не подтвердил бы A без recovery evidence, связанного именно с VM.

### 6.2. Fingerprint был классификатором поверхности, не тестом decoding

F2 измерял \(T(V)\): edit graph, позиции, границы, locus/folio/hierarchy и
другие свойства готовой записи. Сильная external-memory гипотеза относится к
условному отображению \(D(V,K)\to X\). Без известных `X` и `K` F2 не мог
непосредственно измерить ключевое свойство гипотезы — conditional
recoverability.

Таким образом, главным design mismatch было не отсутствие строгости, а выбор
surrogate endpoint:

\[
T(V)\text{ похож}
\quad\not\Rightarrow\quad
D(V,K)\text{ имеет нужную структуру}.
\]

### 6.3. Положительные controls были недостаточны для generalization

- Fontana branch не имела исторически наблюдаемых массовых encode/decode pairs.
- BDD содержала реальные пары, но только одну abbreviation tradition.
- Открытого aligned running-text Tironian corpus найдено не было.
- Selective extraction controls были синтетическими; AX не прошёл полный
  validation gate.
- Synthetic mechanism outputs не имели реальной manuscript hierarchy и
  physical layout.
- У каждого frozen canonical mechanism была только одна parameter point, так
  что within-mechanism parameter effects не идентифицировались.

Это не делает controls бесполезными: они показали чувствительность метрик,
knowledge dependence, matched-null behavior и ряд confounds. Но они не могли
оценить межтрадиционную вариативность и полноту класса.

### 6.4. Некоторые различия были артефактами представления

Task82a вынужден был определить одну локальную application как одну строку и
один токен; page structure отсутствовала. Для extraction четыре FIRST/LAST
оператора механически оставляли не более одной единицы на строку, обнуляя
целую line-position family независимо от выбранной позиции. Поэтому часть
“signature” отражала assembler или line collapse, а не уникальный механизм.

### 6.5. Не был выполнен предварительный identifiability test

До открытия VM следовало проверить на simulated/known-class данных:

1. различает ли frozen classifier candidate mechanisms друг от друга;
2. сохраняется ли различимость после варьирования `X`, `K`, параметров,
   длины, layout и transcription noise;
3. не совпадает ли signal интересующего механизма с matched null;
4. какие классы образуют observational equivalence sets;
5. достаточна ли metric-family coverage для заранее заданной мощности.

Phase II сделала части этой работы, но не получила положительного
межклассового identifiability gate до target comparison. В Task83r этот предел
проявился уже как итоговый результат.

## 7. Почему shorthand и extraction не спасли идентификацию

### 7.1. Shorthand

BDD показал реальный measurable abbreviation transformation и зависимость
некоторых expansions от контекста. Но только 1 из 7 доступных CORE-метрик имела
chapter-consistent sign; cross-tradition stability не тестировалась.

В target comparison реальная BDD trajectory была слабо согласована с VM:
cosine 0.279072/0.262481 для ZL3b/IT2a, тогда как matched deletion давала около
0.99. Ни один shorthand null test не имел \(p\le0.05\); минимум — 0.25.
Поэтому `SHORTHAND_COMPATIBILITY = DISFAVORED` и S0 относятся к **этой BDD
традиции и этому отображению**, а не ко всей стенографии.

### 7.2. Selective extraction

Некоторые operators имели directional alignment до cosine около 0.80, но ни
один зарегистрированный matched-null test не имел \(p\le0.05\); минимум —
0.142857. FIRST/LAST был line-collapse-confounded, AX оставался
`NOT_SUPPORTED`. Поэтому A1 означает слабую structural compatibility, а не
обнаруженный hidden channel.

### 7.3. Общий information-reduction fingerprint

Task82b отдельно проверил идею единой подписи information-reducing
representation через связь length ratio и retained entropy. Итог:
`GENERAL_INFORMATION_REDUCTION_SIGNATURE = NOT_SUPPORTED`. Shorthand и
extraction пришлось сохранять отдельными ветвями. Это эмпирически поддерживает
диагноз: “compression” слишком широкий класс, чтобы служить одним
идентифицирующим признаком.

## 8. Диагноз Phase II по уровням

### 8.1. Empirical

VM имеет устойчивую локальную, позиционную, граничную, folio/section и
ограниченную 2D-структуру. Часть edit-family организации объясняется сильным
grammar-matched null. Наблюдаемые эффекты стабильны качественно между двумя
транскрипциями, но не несут уникальной причинной метки.

### 8.2. Historical

*Secretum* подтверждает историческую доступность разнородных систем внешнего
состояния, cues, convention и remembered association. Он не даёт единого
generator, большого paired corpus или документированной связи с VM. “Fontana”
в Phase II — источник ограниченных механизмов, не авторская гипотеза и не
универсальный класс external memory.

### 8.3. Identifiability

Множество процессов совместимо с одними и теми же low-dimensional surface
features. Финальное прямое пересечение покрывало только 3/13 CORE одной семьи;
ни один класс не прошёл support gate. Поэтому механизм не выбирается.

### 8.4. Information-theoretic

Если `K` несёт существенную информацию, \(V\to X\) может быть невозможным
при вполне корректном \(V+K\to X\). Но для VM неизвестны `X`, `K` и decoder,
поэтому conditional recoverability не измерена. Наличие `K` осталось
возможностью, а не результатом.

### 8.5. Experimental-design

Surface fingerprint использовался как surrogate для recovery mechanism без
доказанной уникальности surrogate. Модельные recovery experiments и target
surface comparison были каждый по отдельности осмысленны, но bridge между ними
не имел достаточного coverage и discriminative validation.

### 8.6. Conclusion

**Отвергнуто или ослаблено в пределах теста:**

- преимущество tested Fontana-derived class перед natural, shorthand,
  extraction и simple null — не обнаружено;
- tested BDD shorthand trajectory — disfavored относительно matched nulls;
- единый general information-reduction fingerprint — не поддержан;
- F2 как достаточный идентификатор механизма — не поддержан.

**Осталось возможным, но не подтверждено:**

- external-memory system;
- другой shorthand/abbreviation mechanism;
- selective extraction/hidden channel;
- natural language, cipher, formal notation, template/procedural или mixed
  mechanism.

**Неидентифицируемо имеющимися данными и дизайном:**

- generating class VM по доступному direct F2 intersection;
- наличие и содержание внешнего `K`;
- исходное `X` и expansion VM;
- историческая связь с Fontana;
- equivalence или ranking широких mechanism classes вне frozen grid.

## 9. Была ли Phase II “провалом”

В узком смысле поставленной цели mechanism identification — да: механизм не
идентифицирован, а дизайн оказался недостаточным для такого вывода.

В научном смысле Phase II не была пустой или недействительной. Она:

- построила существенно более богатое и транскрипционно устойчивое описание
  VM;
- отделила reproducible local structure от excess-over-grammar-null;
- формализовала external-state/knowledge-dependent recovery;
- показала heterogeneity Fontana-derived mechanisms;
- получила реальный paired abbreviation control;
- обнаружила line-collapse и matched-thinning confounds;
- продемонстрировала, что endpoint proximity не проходит confirmatory gates;
- локализовала предел вывода как узкое model/measurement intersection.

Поэтому корректная формула такова:

> Phase II успешно обнаружила, что её mechanism-identification experiment не
> идентифицирует механизм. Но она не идентифицировала сам механизм и не
> опровергла external-memory hypothesis.

## 10. Существует ли корректно поставленная Phase III

Да, но только если она меняет **estimand**, а не просто усложняет
классификатор M0–M5.

### 10.1. Корректная цель A: formation model поверхности

Можно искать минимальную модель \(P(V)\), которая на held-out folios
воспроизводит token formation, transitions, positional/boundary effects и
manuscript hierarchy. Такой эксперимент отвечает:

> Какая минимальная observable grammar предсказывает поверхность VM?

Он не должен автоматически отвечать:

> Какой исторический или semantic mechanism создал VM?

Критерии успеха: out-of-sample likelihood/description length, полный
multi-family coverage, sensitivity к transcription и превосходство над
capacity-matched controls. Это валидная Phase III даже при полном отсутствии
decoding.

### 10.2. Корректная цель B: проверка идентифицируемости до target opening

Если цель остаётся mechanism classification, обязательным первым этапом должен
быть blind identifiability benchmark:

- каждый класс генерирует полноценные synthetic manuscripts, включая layout и
  hierarchy;
- `X`, `K`, nuisance parameters и corruption варьируются по заранее
  заданным distributions;
- classifier обучается и проверяется на независимых generators/implementations,
  а не только на репликах одного кода;
- строится confusion matrix и equivalence sets;
- требуется заранее заданная power и multi-family separation;
- классы, неразличимые на положительных данных, нельзя затем различать на VM.

Только после прохождения этого gate имеет смысл открывать VM как target.

### 10.3. Корректная цель C: conditional recoverability

Сильную external-memory гипотезу можно тестировать непосредственно лишь при
появлении независимого ограничения на `K` или `X`: известного ключа,
параллельного текста, устойчивой и независимо мотивированной таблицы
соответствий, повторной записи того же содержания либо внешней семантической
привязки. Тогда возможен preregistered test

\[
H(X\mid V,K) < H(X\mid V)
\]

на held-out material и с wrong-`K` controls.

Без такого нового наблюдаемого источника задача восстановления `X` должна
быть заранее помечена как `NOT_TESTABLE`, а не компенсироваться ростом числа
моделей или jobs.

### 10.4. Stop/go rule перед дальнейшими вычислениями

Новая Phase III оправдана, только если заранее выполняется хотя бы одно из
условий:

1. модели воспроизводят несколько независимых CORE families, включая
   manuscript-level structure;
2. положительные классы различимы между собой до предъявления VM;
3. появился независимый observable proxy для `X` или `K`;
4. цель явно ограничена descriptive formation grammar без исторической или
   decoding-интерпретации.

Если ни одно условие не выполнено, дополнительная инфраструктурная строгость
повысит воспроизводимость того же неидентифицируемого сравнения, но не изменит
его научный статус.

## 11. Итоговая формулировка

Наиболее точный диагноз Phase II:

> Phase II пыталась идентифицировать скрытый механизм образования VM по
> статистике наблюдаемой поверхности, но не установила, что выбранные
> поверхностные последствия уникальны для этого механизма. Исторически
> мотивированные recovery models и структурный fingerprint VM были построены
> и проверены, однако их прямое пересечение охватывало лишь 3 из 13 CORE-метрик
> одной семьи и не включало наблюдаемую conditional recoverability. Поэтому
> ни один класс не прошёл confirmatory support gate, все pairwise преимущества
> остались `NO_CLEAR_ADVANTAGE`, а механизм — `NOT_IDENTIFIABLE`.

Это не опровержение Фонтаны или external memory. Это граница того, что можно
заключить из `V` без независимо наблюдаемых `X`, `K` либо уникальных
multi-family predictions механизма.
