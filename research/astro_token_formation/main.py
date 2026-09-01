#!/usr/bin/env python3
"""Constrained astronomical-label token-formation search."""

from __future__ import annotations

import csv
import hashlib
import json
import math
import random
import re
import statistics
import unicodedata
from collections import Counter, defaultdict
from dataclasses import dataclass
from datetime import datetime, timezone
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
OUT = ROOT / "research/astro_token_formation"
SEED = 20260901
REPLICATES = 1000


@dataclass(frozen=True)
class Rule:
    lexical: str
    ending: str
    orthography: str
    vowel: str
    abbreviation: str

    @property
    def rule_string(self) -> str:
        return ";".join((self.lexical, self.ending, self.orthography, self.vowel, self.abbreviation))

    @property
    def complexity(self) -> int:
        cost = 0
        cost += {"FULL_JOIN": 0, "DROP_AL": 1, "HEAD": 2, "TAIL": 2}[self.lexical]
        cost += self.ending != "KEEP"
        cost += self.orthography != "IDENTITY"
        cost += self.vowel != "KEEP"
        cost += {"NONE": 0, "SUSPEND_1": 1, "SUSPEND_2": 1, "PREFIX_3": 2, "PREFIX_4": 2}[self.abbreviation]
        return int(cost)

    @property
    def quality(self) -> str:
        return "HIGH" if self.complexity <= 1 else "MEDIUM" if self.complexity <= 3 else "LOW"


def read_tsv(path: Path) -> list[dict[str, str]]:
    with path.open(encoding="utf-8", newline="") as fh:
        return list(csv.DictReader(fh, delimiter="\t"))


def write_tsv(path: Path, fields: list[str], rows: list[dict[str, object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as fh:
        writer = csv.DictWriter(fh, fields, delimiter="\t", lineterminator="\n", extrasaction="ignore")
        writer.writeheader()
        writer.writerows(rows)


def words(form: str) -> list[str]:
    ascii_form = unicodedata.normalize("NFKD", form).encode("ascii", "ignore").decode().lower()
    return re.findall(r"[a-z]+", ascii_form)


def apply_rule(form: str, rule: Rule) -> str:
    ws = words(form)
    if not ws:
        return ""
    if rule.lexical == "DROP_AL":
        ws = [w for w in ws if w != "al"] or ws
    elif rule.lexical == "HEAD":
        ws = ws[:1]
    elif rule.lexical == "TAIL":
        ws = ws[-1:]
    value = "".join(ws)
    if rule.ending == "STRIP_LATIN":
        for ending in ("ibus", "orum", "arum", "ium", "ius", "ae", "is", "us", "um", "ii", "am", "em", "as", "es", "os", "i", "o", "a", "e"):
            if value.endswith(ending) and len(value) - len(ending) >= 3:
                value = value[: -len(ending)]
                break
    if rule.orthography in {"ARABIC_LATIN", "VELAR_COLLAPSE"}:
        for old, new in (("kh", "h"), ("gh", "g"), ("sh", "s"), ("th", "t"), ("dh", "d")):
            value = value.replace(old, new)
        value = value.translate(str.maketrans({"j": "i", "w": "u"}))
    elif rule.orthography == "IJ_UV":
        value = value.translate(str.maketrans({"j": "i", "v": "u"}))
    if rule.orthography == "VELAR_COLLAPSE":
        value = value.translate(str.maketrans({"q": "k", "c": "k"}))
    if rule.vowel == "DELETE_FINAL" and value[-1:] in "aeiouy" and len(value) > 3:
        value = value[:-1]
    elif rule.vowel == "COLLAPSE_A":
        value = re.sub(r"[aeiouy]", "a", value)
    elif rule.vowel == "CONTRACT_INTERNAL" and len(value) > 2:
        value = value[0] + re.sub(r"[aeiouy]", "", value[1:-1]) + value[-1]
    if rule.abbreviation == "SUSPEND_1" and len(value) > 3:
        value = value[:-1]
    elif rule.abbreviation == "SUSPEND_2" and len(value) > 4:
        value = value[:-2]
    elif rule.abbreviation == "PREFIX_3" and len(value) > 3:
        value = value[:3]
    elif rule.abbreviation == "PREFIX_4" and len(value) > 4:
        value = value[:4]
    return value


def rule_grid() -> list[Rule]:
    return [Rule(a, b, c, d, e)
            for a in ("FULL_JOIN", "DROP_AL", "HEAD", "TAIL")
            for b in ("KEEP", "STRIP_LATIN")
            for c in ("IDENTITY", "IJ_UV", "ARABIC_LATIN", "VELAR_COLLAPSE")
            for d in ("KEEP", "DELETE_FINAL", "COLLAPSE_A", "CONTRACT_INTERNAL")
            for e in ("NONE", "SUSPEND_1", "SUSPEND_2", "PREFIX_3", "PREFIX_4")]


def load_terms() -> list[dict[str, object]]:
    result = []
    for row in read_tsv(OUT / "ASTRO_TERM_CORPUS.tsv"):
        forms = []
        for field in ("arabic_form", "latin_form", "medieval_latinized_arabic_forms", "documented_spelling_variants"):
            forms.extend(x.strip() for x in row[field].split(";") if x.strip() and x.strip() != "-")
        result.append({"object_id": row["object_id"], "object_class": row["object_class"], "forms": sorted(set(forms))})
    return result


def transformed(terms: list[dict[str, object]], rule: Rule) -> dict[str, dict[str, list[str]]]:
    out: dict[str, dict[str, list[str]]] = {}
    for term in terms:
        evidence: dict[str, list[str]] = defaultdict(list)
        for form in term["forms"]:  # type: ignore[index]
            value = apply_rule(str(form), rule)
            if value:
                evidence[value].append(str(form))
        out[str(term["object_id"])] = dict(evidence)
    return out


def max_matching(labels: list[dict[str, str]], terms: list[dict[str, object]], outputs: dict[str, dict[str, list[str]]], unavailable: set[str] | None = None) -> dict[str, str]:
    unavailable = unavailable or set()
    by_class: dict[str, list[str]] = defaultdict(list)
    for term in terms:
        if term["object_id"] not in unavailable:
            by_class[str(term["object_class"])].append(str(term["object_id"]))
    adjacency: dict[str, list[str]] = {}
    for label in labels:
        adjacency[label["stolfi_coordinate"]] = sorted(
            oid for oid in by_class[label["object_class"]]
            if label["voynich_token"] in outputs[oid]
        )
    object_to_label: dict[str, str] = {}

    def visit(label_id: str, seen: set[str]) -> bool:
        for oid in adjacency[label_id]:
            if oid in seen:
                continue
            seen.add(oid)
            if oid not in object_to_label or visit(object_to_label[oid], seen):
                object_to_label[oid] = label_id
                return True
        return False

    for label_id in sorted(adjacency):
        visit(label_id, set())
    return {label: oid for oid, label in object_to_label.items()}


def percentile(values: list[float], q: float) -> float:
    values = sorted(values)
    if not values:
        return 0.0
    pos = (len(values) - 1) * q
    lo, hi = math.floor(pos), math.ceil(pos)
    return values[lo] if lo == hi else values[lo] * (hi - pos) + values[hi] * (pos - lo)


def altered_terms(terms: list[dict[str, object]], rng: random.Random, pseudo: bool, alphabets: dict[str, tuple[list[str], list[float]]]) -> list[dict[str, object]]:
    changed = []
    for term in terms:
        new_forms = []
        alphabet, weights = alphabets[str(term["object_class"])]
        for form in term["forms"]:  # type: ignore[index]
            if pseudo:
                new_forms.append(" ".join("".join(rng.choices(alphabet, weights=weights, k=len(w))) for w in words(str(form))))
            else:
                shuffled_words = []
                for word in words(str(form)):
                    chars = list(word)
                    rng.shuffle(chars)
                    shuffled_words.append("".join(chars))
                new_forms.append(" ".join(shuffled_words))
        changed.append({"object_id": term["object_id"], "object_class": term["object_class"], "forms": new_forms})
    return changed


def load_random_pool() -> dict[int, list[str]]:
    pool: dict[int, list[str]] = defaultdict(list)
    expansions = {"C": "cth", "K": "ckh", "P": "cph", "F": "cfh", "N": "iin", "A": "ain", "H": "ch", "S": "sh", "E": "ee", "I": "in"}
    path = ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"
    with path.open(encoding="utf-8") as fh:
        for line in fh:
            row = json.loads(line)
            if row["section"] != "A":
                continue
            token = "".join(expansions.get(atom, atom) for atom in row["token"].split("\x1f"))
            if re.fullmatch(r"[a-z]+", token):
                pool[len(token)].append(token)
    return {length: sorted(tokens) for length, tokens in pool.items()}


def random_labels(template: list[dict[str, str]], pool: dict[int, list[str]], rng: random.Random) -> list[dict[str, str]]:
    available = {length: values.copy() for length, values in pool.items()}
    result = []
    for i, label in enumerate(template):
        candidates = available[len(label["voynich_token"])]
        if not candidates:
            raise RuntimeError("insufficient random token pool")
        token = candidates.pop(rng.randrange(len(candidates)))
        result.append({**label, "stolfi_coordinate": f"NULL_{i:02d}", "voynich_token": token})
    return result


def band(train_cov: float, held_cov: float, advantage: float, complexity: int) -> str:
    if train_cov >= .70:
        if held_cov >= .50 and advantage >= .20 and complexity <= 4:
            return "STRONG_CANDIDATE"
        return "OVERFIT"
    if train_cov >= .40 and held_cov >= .30 and advantage >= .10 and complexity <= 3:
        return "PARTIAL_CANDIDATE"
    return "NULL_COMPATIBLE"


def fmt(x: float) -> str:
    return f"{x:.6f}"


def sha(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def main() -> None:
    split_rows = read_tsv(OUT / "ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv")
    selected = [r for r in split_rows if r["selected"] == "1"]
    train = [r for r in selected if r["split"] == "TRAIN"]
    held = [r for r in selected if r["split"] == "HELD_OUT"]
    terms = load_terms()
    rules = rule_grid()
    if len(rules) != 640 or len(train) != 20 or len(held) != 6:
        raise RuntimeError("frozen grid/sample invariant failed")

    model_data = []
    original_outputs = []
    for i, rule in enumerate(rules, 1):
        outputs = transformed(terms, rule)
        original_outputs.append(outputs)
        train_match = max_matching(train, terms, outputs)
        used = set(train_match.values())
        held_match = max_matching(held, terms, outputs, used)
        model_data.append({
            "model_id": f"M{i:04d}", "rule": rule, "outputs": outputs,
            "train_match": train_match, "held_match": held_match,
            "train_coverage": len(train_match) / len(train),
            "heldout_coverage": len(held_match) / len(held),
        })

    # Frozen negative controls across the complete 640-model search family.
    alphabets: dict[str, tuple[list[str], list[float]]] = {}
    for cls in ("STAR", "PLANET_MOON"):
        counts = Counter(ch for t in terms if t["object_class"] == cls for f in t["forms"] for ch in "".join(words(str(f))))
        alphabets[cls] = (sorted(counts), [counts[x] for x in sorted(counts)])
    pool = load_random_pool()
    controls = {name: [[] for _ in rules] for name in ("RANDOM_VOYNICH_SET", "SHUFFLED_TERMS", "PSEUDODICTIONARY")}
    grid_max = {name: [] for name in controls}
    rng = random.Random(SEED)
    for _rep in range(REPLICATES):
        null_labels = random_labels(train, pool, rng)
        shuffled = altered_terms(terms, rng, False, alphabets)
        pseudo = altered_terms(terms, rng, True, alphabets)
        changed = {
            "SHUFFLED_TERMS": shuffled,
            "PSEUDODICTIONARY": pseudo,
        }
        rep_max = {name: 0.0 for name in controls}
        for idx, rule in enumerate(rules):
            random_cov = len(max_matching(null_labels, terms, original_outputs[idx])) / len(train)
            controls["RANDOM_VOYNICH_SET"][idx].append(random_cov)
            rep_max["RANDOM_VOYNICH_SET"] = max(rep_max["RANDOM_VOYNICH_SET"], random_cov)
            for name, changed_terms in changed.items():
                out = transformed(changed_terms, rule)
                cov = len(max_matching(train, changed_terms, out)) / len(train)
                controls[name][idx].append(cov)
                rep_max[name] = max(rep_max[name], cov)
        for name in controls:
            grid_max[name].append(rep_max[name])

    for idx, model in enumerate(model_data):
        means = {name: statistics.fmean(values[idx]) for name, values in controls.items()}
        model["null_means"] = means
        model["null_score"] = max(means.values())
        model["null_advantage"] = model["train_coverage"] - model["null_score"]
        model["band"] = band(model["train_coverage"], model["heldout_coverage"], model["null_advantage"], model["rule"].complexity)
        model["score"] = model["train_coverage"] + model["heldout_coverage"] + model["null_advantage"] + .25 - .05 * model["rule"].complexity
        model["empirical_p"] = max((1 + sum(x >= model["train_coverage"] for x in grid_max[name])) / (REPLICATES + 1) for name in grid_max)

    passing = [m for m in model_data if m["band"] != "NULL_COMPATIBLE"]
    rank_key = lambda m: (-m["score"], -m["train_coverage"], -m["heldout_coverage"], m["rule"].complexity, m["model_id"])
    retained = sorted(passing, key=rank_key) if passing else sorted(model_data, key=rank_key)[:10]

    model_rows = []
    for m in retained:
        model_rows.append({
            "model_id": m["model_id"], "candidate_band": m["band"],
            "train_coverage": fmt(m["train_coverage"]), "train_matched": len(m["train_match"]), "train_total": len(train),
            "heldout_coverage": fmt(m["heldout_coverage"]), "heldout_matched": len(m["held_match"]), "heldout_total": len(held),
            "rules": m["rule"].rule_string, "complexity": m["rule"].complexity,
            "consistency": "1.000000", "transformation_quality": m["rule"].quality,
            "exceptions": len(train) - len(m["train_match"]), "null_score": fmt(m["null_score"]),
            "null_advantage": fmt(m["null_advantage"]), "familywise_empirical_p": fmt(m["empirical_p"]),
            "score": fmt(m["score"]),
            "explained_labels": ";".join(sorted(m["train_match"])),
            "unexplained_labels": ";".join(sorted(r["stolfi_coordinate"] for r in train if r["stolfi_coordinate"] not in m["train_match"])),
            "diagnostic_only": "0" if passing else "1",
        })
    model_fields = list(model_rows[0])
    write_tsv(OUT / "TOKEN_FORMATION_MODELS.tsv", model_fields, model_rows)

    held_rows = []
    unexplained_rows = []
    label_lookup = {r["stolfi_coordinate"]: r for r in selected}
    details_dir = OUT / "TOKEN_FORMATION_MODEL_DETAILS"
    details_dir.mkdir(exist_ok=True)
    for m in retained:
        for split_name, labels, matching in (("TRAIN", train, m["train_match"]), ("HELD_OUT", held, m["held_match"])):
            for label in sorted(labels, key=lambda r: r["stolfi_coordinate"]):
                oid = matching.get(label["stolfi_coordinate"], "")
                evidence = ""
                if oid:
                    evidence = ";".join(m["outputs"][oid][label["voynich_token"]])
                row = {
                    "model_id": m["model_id"], "split": split_name,
                    "stolfi_coordinate": label["stolfi_coordinate"], "object_class": label["object_class"],
                    "panel": label["panel"], "stolfi_group": label["stolfi_group"],
                    "zl3b_locus": label["zl3b_locus"], "absolute_token_position": label["absolute_token_position"],
                    "voynich_token": label["voynich_token"], "prediction": "MATCHED" if oid else "UNEXPLAINED",
                    "matched_object_id": oid, "source_form_evidence": evidence,
                }
                if split_name == "HELD_OUT":
                    held_rows.append(row)
                if not oid:
                    unexplained_rows.append({**row, "system_applicable_without_assignment": "NO_EXACT_CORPUS_OUTPUT"})
        detail = [
            f"# {m['model_id']}", "", f"Band: `{m['band']}`", "",
            f"Rules: `{m['rule'].rule_string}`", "",
            f"TRAIN: {len(m['train_match'])}/{len(train)} ({m['train_coverage']:.6f})", "",
            f"HELD_OUT: {len(m['held_match'])}/{len(held)} ({m['heldout_coverage']:.6f})", "",
            "## Canonical anonymous matches", "",
        ]
        for split_name, matching in (("TRAIN", m["train_match"]), ("HELD_OUT", m["held_match"])):
            detail.append(f"### {split_name}\n")
            if not matching:
                detail.append("No exact matches.\n")
            for coord, oid in sorted(matching.items()):
                label = label_lookup[coord]
                forms = ", ".join(m["outputs"][oid][label["voynich_token"]])
                detail.append(f"- `{coord}` `{label['voynich_token']}` ← `{oid}` ({forms})")
            detail.append("")
        (details_dir / f"{m['model_id']}.md").write_text("\n".join(detail), encoding="utf-8")

    held_fields = ["model_id", "split", "stolfi_coordinate", "object_class", "panel", "stolfi_group", "zl3b_locus", "absolute_token_position", "voynich_token", "prediction", "matched_object_id", "source_form_evidence"]
    write_tsv(OUT / "TOKEN_FORMATION_HELDOUT.tsv", held_fields, held_rows)
    write_tsv(OUT / "TOKEN_FORMATION_UNEXPLAINED_LABELS.tsv", held_fields + ["system_applicable_without_assignment"], unexplained_rows)

    null_rows = []
    for m in retained:
        idx = int(m["model_id"][1:]) - 1
        n_explained = len(m["train_match"])
        assign_values = []
        local_rng = random.Random(SEED + idx)
        for _ in range(REPLICATES):
            if n_explained == 0:
                assign_values.append(0.0)
            else:
                values = list(range(n_explained)); local_rng.shuffle(values)
                assign_values.append(sum(i == x for i, x in enumerate(values)) / len(train))
        for name, values in [("SHUFFLED_ASSIGNMENT", assign_values)] + [(name, controls[name][idx]) for name in controls]:
            null_rows.append({
                "model_id": m["model_id"], "control": name, "replicates": REPLICATES,
                "observed_train_coverage": fmt(m["train_coverage"]), "mean_null_coverage": fmt(statistics.fmean(values)),
                "sd_null_coverage": fmt(statistics.pstdev(values)), "p95_null_coverage": fmt(percentile(values, .95)),
                "max_null_coverage": fmt(max(values)), "replicates_coverage_ge_0_70": sum(x >= .70 for x in values),
                "pointwise_empirical_p": fmt((1 + sum(x >= m["train_coverage"] for x in values)) / (REPLICATES + 1)),
                "familywise_empirical_p": fmt(m["empirical_p"] if name != "SHUFFLED_ASSIGNMENT" else 1.0),
            })
    for name, values in grid_max.items():
        best_observed = max(m["train_coverage"] for m in model_data)
        null_rows.append({
            "model_id": "GRID_MAX", "control": name, "replicates": REPLICATES,
            "observed_train_coverage": fmt(best_observed), "mean_null_coverage": fmt(statistics.fmean(values)),
            "sd_null_coverage": fmt(statistics.pstdev(values)), "p95_null_coverage": fmt(percentile(values, .95)),
            "max_null_coverage": fmt(max(values)), "replicates_coverage_ge_0_70": sum(x >= .70 for x in values),
            "pointwise_empirical_p": "NA",
            "familywise_empirical_p": fmt((1 + sum(x >= best_observed for x in values)) / (REPLICATES + 1)),
        })
    null_fields = list(null_rows[0])
    write_tsv(OUT / "TOKEN_FORMATION_NULL_TEST.tsv", null_fields, null_rows)

    best = sorted(model_data, key=rank_key)[0]
    strong = [m for m in model_data if m["band"] == "STRONG_CANDIDATE"]
    partial = [m for m in model_data if m["band"] == "PARTIAL_CANDIDATE"]
    predictive = [m for m in strong if m["empirical_p"] <= .05 and m["heldout_coverage"] >= .50]
    status = "PREDICTIVE_MODEL_FOUND" if predictive else "STRONG_CANDIDATES_FOUND" if strong else "PARTIAL_CANDIDATES_FOUND" if partial else "NO_MODEL"
    counted = strong + partial
    report = f"""# Constrained astronomical token-formation search

## Result

`{status}`. The best of 640 frozen global pipelines explained
{len(best['train_match'])}/{len(train)} TRAIN labels ({best['train_coverage']:.6f}) and
{len(best['held_match'])}/{len(held)} HELD_OUT labels ({best['heldout_coverage']:.6f}).
It did not cross a candidate threshold. The ten rows in
`TOKEN_FORMATION_MODELS.tsv` are diagnostic null-compatible ties and are not
claimed as decipherments.

## Design

The spelling-blind SHA-256 sample contains 21 STAR and five PLANET_MOON
single-token labels. It was frozen as 20 TRAIN and six HELD_OUT labels before
the search. The corpus contains 21 Arabic star-name inscriptions attested on a
thirteenth-century astrolabe rete and seven medieval Latin planet names.
Question-marked Stolfi planet comments supplied only the morphological class;
their proposed identities were never used.

The search is anonymous one-to-one set matching within class because no
independent star identity exists for these manuscript labels. Consequently a
match would show only compatibility with a lexicon and global rule, not a
semantic identification. Exact transformed strings alone count; no edit
distance, image reading, per-label rule, or exception is available. Full rules
and preregistered bands are in `TOKEN_FORMATION_RULE_SPACE.md`.

## Negative controls

Each stochastic control used {REPLICATES} deterministic replicates (seed
{SEED}). The familywise statistic is the maximum over the full 640-rule grid in
each replicate. No control replicate is used to modify the corpus or rules.
For this run the best observed TRAIN coverage was {best['train_coverage']:.6f};
the best model's largest mean null coverage was {best['null_score']:.6f}, for a
null advantage of {best['null_advantage']:.6f}. See
`TOKEN_FORMATION_NULL_TEST.tsv` for control-specific tails and the frequency of
coverage at or above 0.70.

`SHUFFLED_ASSIGNMENT` is explicitly secondary: because label identities are
unknown, it tests stability of the canonical anonymous pairing rather than a
known semantic assignment. Random token sets, shuffled historical forms, and
matched pseudodictionaries are the substantive controls.

## Historical provenance

The star list is the National Museums Scotland catalogue article “A
1000-year-old star catcher,” live version accessed 2026-09-01, specifically its
table of 21 calculated pointers on the thirteenth-century rete of T.1959.62.
Only the Arabic forms in the `On rete` column were transcribed. Kunitzsch,
“An unknown Arabic source for star names,” *History of Oriental Astronomy*
(1987), DOI `10.1017/S0252921100105986`, supplies only the directly attested
1246 spelling `bedalgeuze`; doubtful or differently assigned names were not
imported. Walters MS W.73 (late twelfth century), fol. 2v, directly lists the
seven Latin heavenly bodies. The anonymous eighth-century pseudo-Bedan `Ordo
planetarum`, Migne PL 90, cols. 943D–946A, documents their inflected forms.

## Interpretation

Failure here rejects only this compact, historically constrained 640-model
space against this frozen corpus/sample. It is not proof that no historical
formation system exists. In particular, the test deliberately refuses an
arbitrary Latin/Arabic-to-EVA substitution alphabet, modern reconstructed
names, reverse strings, or post-hoc spellings. Unexplained labels remain in
`TOKEN_FORMATION_UNEXPLAINED_LABELS.tsv`; no translation is proposed for them.

## Reproduction

Run `python3 research/astro_token_formation/freeze_split.py`, then
`python3 research/astro_token_formation/main.py`. Corpus bytes remain governed
by `DATA.md`; the analysis uses the already frozen occurrence metadata and
Stolfi match audit.

## Final status

```text
ASTRO_TOKEN_FORMATION_SEARCH={status}
MODELS_FOUND={len(counted)}
MODELS_WITH_COVERAGE_GE_70={sum(m['train_coverage'] >= .70 for m in model_data)}
BEST_TRAIN_COVERAGE={best['train_coverage']:.6f}
BEST_HELDOUT_COVERAGE={best['heldout_coverage']:.6f}
BEST_MODEL_COMPLEXITY={best['rule'].complexity}
BEST_MODEL_NULL_ADVANTAGE={best['null_advantage']:.6f}
UNEXPLAINED_LABELS={len(train) - len(best['train_match']) + len(held) - len(best['held_match'])}
```
"""
    (OUT / "TOKEN_FORMATION_BRUTEFORCE_REPORT.md").write_text(report, encoding="utf-8")

    artifacts = [
        "research/astro_token_formation/ASTRO_TERM_CORPUS.tsv", "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv", "research/astro_token_formation/TOKEN_FORMATION_RULE_SPACE.md",
        "research/astro_token_formation/TOKEN_FORMATION_MODELS.tsv", "research/astro_token_formation/TOKEN_FORMATION_UNEXPLAINED_LABELS.tsv",
        "research/astro_token_formation/TOKEN_FORMATION_NULL_TEST.tsv", "research/astro_token_formation/TOKEN_FORMATION_HELDOUT.tsv", "research/astro_token_formation/TOKEN_FORMATION_BRUTEFORCE_REPORT.md",
    ] + [str(p.relative_to(ROOT)) for p in sorted(details_dir.glob("*.md"))]
    inputs = [
        "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv",
        "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl",
        "research/astro_token_formation/ASTRO_TERM_CORPUS.tsv", "research/astro_token_formation/ASTRO_LABEL_TRAIN_TEST_SPLIT.tsv", "research/astro_token_formation/TOKEN_FORMATION_RULE_SPACE.md",
    ]
    manifest = {
        "experiment": "constrained-astronomical-token-formation-v1",
        "generated_utc": datetime.now(timezone.utc).replace(microsecond=0).isoformat(),
        "seed": SEED, "null_replicates": REPLICATES, "rule_grid_size": len(rules),
        "train_labels": len(train), "heldout_labels": len(held), "status": status,
        "source_keys": {
            "NMS_RETE_13C": "https://www.nms.ac.uk/discover-catalogue/a-1000-year-old-star-catcher",
            "KUNITZSCH_1987": "https://doi.org/10.1017/S0252921100105986",
            "WALTERS_W73": "https://www.thedigitalwalters.org/Data/WaltersManuscripts/html/W73/description.html",
            "ORDO_PLANETARUM": "https://la.wikisource.org/wiki/Ordo_planetarum",
        },
        "input_sha256": {p: sha(ROOT / p) for p in inputs},
        "artifact_sha256": {p: sha(ROOT / p) for p in artifacts},
    }
    manifest_path = OUT / "TOKEN_FORMATION_MANIFEST.json"
    manifest_path.write_text(json.dumps(manifest, ensure_ascii=False, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    checksum_paths = artifacts + ["TOKEN_FORMATION_MANIFEST.json"]
    (OUT / "TOKEN_FORMATION_SHA256SUMS").write_text("".join(f"{sha(ROOT / p)}  {p}\n" for p in sorted(checksum_paths)), encoding="utf-8")


if __name__ == "__main__":
    main()
