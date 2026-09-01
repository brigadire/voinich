#!/usr/bin/env python3
"""Cross-section usage of confirmed Astronomical Stolfi label token types."""

from __future__ import annotations

import csv
import hashlib
import json
import platform
import sys
from collections import Counter, defaultdict
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
MATCHES = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_MATCHES.tsv"
OCCURRENCES = ROOT / "experiments/fingerprint-v2-task79-v1/canonical-out/occurrence_metadata.jsonl"
TAXONOMY = ROOT / "research/visual_context/VISUAL_CONTEXT_TAXONOMY.tsv"
STOLFI = ROOT / "data/stolfi-labtit-98-07-20.idx"
LABEL_AUDIT = ROOT / "research/stolfi_label_inventory/STOLFI_ASTRO_LABEL_AUDIT.md"

OUT = ROOT / "research/astro_label_cross_section"
USAGE = OUT / "ASTRO_LABEL_CROSS_SECTION_USAGE.tsv"
BY_FAMILY = OUT / "ASTRO_LABEL_CROSS_SECTION_BY_FAMILY.tsv"
STAR = OUT / "ASTRO_LABEL_CROSS_SECTION_STAR.tsv"
SUMMARY = OUT / "ASTRO_LABEL_CROSS_SECTION_SUMMARY.md"
MANIFEST = OUT / "ASTRO_LABEL_CROSS_SECTION_MANIFEST.json"
CHECKSUMS = OUT / "ASTRO_LABEL_CROSS_SECTION_SHA256SUMS"

SECTION_NAMES = {
    "A": "Astronomical",
    "H": "Herbal",
    "P": "Pharmaceutical",
    "B": "Biological",
    "Z": "Zodiac",
    "S": "Stars",
    "C": "Cosmological",
    "T": "Text",
}
SECTION_COLUMNS = (
    ("Astronomical", "astro_count"),
    ("Herbal", "herbal_count"),
    ("Pharmaceutical", "pharmaceutical_count"),
    ("Biological", "biological_count"),
    ("Zodiac", "zodiac_count"),
    ("Stars", "stars_count"),
    ("Cosmological", "cosmological_count"),
    ("Text", "text_count"),
)
STOLFI_SECTION_NAMES = {
    "astro": "Astronomical",
    "herbal": "Herbal",
    "pharma": "Pharmaceutical",
    "bio": "Biological",
    "zodiac": "Zodiac",
    "stars": "Stars",
    "cosmo": "Cosmological",
    "?": "Unknown",
}
COMPOSITES = (
    ("cth", "C"), ("ckh", "K"), ("cph", "P"), ("cfh", "F"),
    ("iin", "N"), ("ain", "A"), ("ch", "H"), ("sh", "S"),
    ("ee", "E"), ("in", "I"),
)
EXPANSIONS = {
    "C": "cth", "K": "ckh", "P": "cph", "F": "cfh",
    "N": "iin", "A": "ain", "H": "ch", "S": "sh",
    "E": "ee", "I": "in",
}


def sha256(path: Path) -> str:
    h = hashlib.sha256()
    with path.open("rb") as handle:
        for block in iter(lambda: handle.read(1 << 20), b""):
            h.update(block)
    return h.hexdigest()


def write_tsv(path: Path, header: list[str], rows: list[list[object]]) -> None:
    with path.open("w", encoding="utf-8", newline="") as handle:
        writer = csv.writer(handle, delimiter="\t", lineterminator="\n")
        writer.writerow(header)
        writer.writerows(rows)


def fmt(value: float) -> str:
    return f"{value:.9f}"


def collapse_eva(token: str) -> str:
    value = token.lower()
    for composite, atom in COMPOSITES:
        value = value.replace(composite, atom)
    return "\x1f".join(value)


def expand_type(token_type: str) -> str:
    return "".join(EXPANSIONS.get(atom, atom) for atom in token_type.split("\x1f"))


def stolfi_tokens(eva: str) -> list[str]:
    value = eva
    for separator in (".", ",", "-", " ", "\t"):
        value = value.replace(separator, " ")
    return [token for token in value.split() if token and "*" not in token and "?" not in token]


def family(row: dict[str, str]) -> str:
    if row["object_type"] == "star":
        return "STAR"
    if row["object_type"] == "planet?":
        return "PLANET_MOON"
    if row["panel"] == "f67r1" and row["stolfi_group"] == "S":
        return "CIRCLE_SECTOR"
    if row["panel"] == "f68v2" and row["stolfi_group"] == "R":
        return "RADIAL_TEXT"
    return "OTHER"


def load_taxonomy() -> dict[str, str]:
    page_section = {}
    with TAXONOMY.open(encoding="utf-8", newline="") as handle:
        for row in csv.DictReader(handle, delimiter="\t"):
            code = row["visual_code"]
            if code not in SECTION_NAMES:
                raise RuntimeError(f"unknown broad section code in taxonomy: {code}")
            page_section[row["page_id"]] = code
    return page_section


def load_occurrences(page_section):
    rows = []
    counts = defaultdict(Counter)
    positions = {}
    with OCCURRENCES.open(encoding="utf-8") as handle:
        for line in handle:
            row = json.loads(line)
            code = row["section"]
            if code not in SECTION_NAMES:
                raise RuntimeError(f"unknown occurrence section code: {code}")
            if page_section.get(row["folio"]) != code:
                raise RuntimeError(f"taxonomy mismatch at {row['folio']}: {code}")
            rows.append(row)
            positions[row["absolute_token_position"]] = row
            counts[row["token"]][SECTION_NAMES[code]] += 1
    if len(rows) != 39380:
        raise RuntimeError(f"expected 39380 frozen occurrences, got {len(rows)}")
    return rows, positions, counts


def load_confirmed_types(positions):
    type_families = defaultdict(set)
    type_label_positions = defaultdict(set)
    confirmed_positions = set()
    matched_records = 0
    with MATCHES.open(encoding="utf-8", newline="") as handle:
        for row in csv.DictReader(handle, delimiter="\t"):
            if row["match_status"] != "MATCHED":
                continue
            matched_records += 1
            fam = family(row)
            for raw_position in row["absolute_token_positions"].split(","):
                position = int(raw_position)
                occurrence = positions[position]
                if occurrence["section"] != "A":
                    raise RuntimeError(f"confirmed label outside Astronomical: {position}")
                token_type = occurrence["token"]
                confirmed_positions.add(position)
                type_label_positions[token_type].add(position)
                type_families[token_type].add(fam)
    if matched_records != 130 or len(confirmed_positions) != 112:
        raise RuntimeError(
            f"confirmed-label invariant failed: records={matched_records}, positions={len(confirmed_positions)}"
        )
    return type_families, type_label_positions


def load_outside_stolfi_labels():
    sections = defaultdict(set)
    source_records = 0
    eligible_records = 0
    with STOLFI.open(encoding="latin-1") as handle:
        for line in handle:
            fields = [field.strip() for field in line.rstrip("\r\n").split("|")]
            if len(fields) != 11:
                raise RuntimeError(f"malformed Stolfi record: {line[:40]!r}")
            source_records += 1
            section = STOLFI_SECTION_NAMES.get(fields[1])
            if section is None:
                raise RuntimeError(f"unknown Stolfi section: {fields[1]}")
            if section == "Astronomical" or fields[9] == "title?":
                continue
            eligible_records += 1
            for token in stolfi_tokens(fields[6]):
                sections[collapse_eva(token)].add(section)
    if source_records != 1485:
        raise RuntimeError(f"expected 1485 Stolfi records, got {source_records}")
    return sections, source_records, eligible_records


def classification(global_count, astro_count, outside_label_sections):
    if global_count == 1:
        return "GLOBAL_HAPAX"
    if outside_label_sections:
        return "LABEL_REUSED_ACROSS_SECTIONS"
    if global_count > astro_count:
        return "ASTRO_LABEL_BUT_RUNNING_TEXT_ELSEWHERE"
    return "ASTRO_LABEL_ONLY"


def main() -> None:
    page_section = load_taxonomy()
    all_occurrences, positions, counts = load_occurrences(page_section)
    type_families, type_label_positions = load_confirmed_types(positions)
    outside_labels, stolfi_records, eligible_stolfi_records = load_outside_stolfi_labels()

    header = [
        "token", "canonical_token_key", "astro_label_family", "confirmed_astro_label_occurrences",
        "astro_count", "herbal_count", "pharmaceutical_count", "biological_count", "zodiac_count",
        "stars_count", "cosmological_count", "text_count", "global_count", "is_global_hapax",
        "is_astronomical_section_hapax", "occurs_outside_astronomical", "occurs_in_herbal",
        "occurs_in_pharmaceutical", "stolfi_label_sections_outside_astro", "cross_section_class",
    ]
    data = []
    for token_type in sorted(type_families, key=lambda value: (expand_type(value), value)):
        section_counts = counts[token_type]
        values = {column: section_counts[name] for name, column in SECTION_COLUMNS}
        global_count = sum(section_counts.values())
        outside_sections = sorted(outside_labels.get(token_type, set()))
        row = {
            "token": expand_type(token_type),
            "canonical_token_key": "/".join(token_type.split("\x1f")),
            "astro_label_family": ";".join(sorted(type_families[token_type])),
            "confirmed_astro_label_occurrences": len(type_label_positions[token_type]),
            **values,
            "global_count": global_count,
            "is_global_hapax": global_count == 1,
            "is_astronomical_section_hapax": values["astro_count"] == 1,
            "occurs_outside_astronomical": global_count > values["astro_count"],
            "occurs_in_herbal": values["herbal_count"] > 0,
            "occurs_in_pharmaceutical": values["pharmaceutical_count"] > 0,
            "stolfi_label_sections_outside_astro": ";".join(outside_sections),
            "cross_section_class": classification(global_count, values["astro_count"], outside_sections),
        }
        data.append(row)
    if len(data) != len(type_families):
        raise RuntimeError("type deduplication failed")
    write_tsv(USAGE, header, [[row[column] for column in header] for row in data])

    families = sorted({item for row in data for item in row["astro_label_family"].split(";")})
    family_rows = []
    for fam in families:
        subset = [row for row in data if fam in row["astro_label_family"].split(";")]
        total = len(subset)
        global_hapax = sum(row["is_global_hapax"] for row in subset)
        outside = sum(row["occurs_outside_astronomical"] for row in subset)
        herbal = sum(row["occurs_in_herbal"] for row in subset)
        pharma = sum(row["occurs_in_pharmaceutical"] for row in subset)
        cross_labels = sum(bool(row["stolfi_label_sections_outside_astro"]) for row in subset)
        section_hapax_outside = sum(
            row["is_astronomical_section_hapax"] and row["occurs_outside_astronomical"]
            for row in subset
        )
        family_rows.append([
            fam, total, global_hapax, fmt(global_hapax / total), outside, fmt(outside / total),
            herbal, fmt(herbal / total), pharma, fmt(pharma / total), cross_labels,
            fmt(cross_labels / total), section_hapax_outside,
            sum(row["confirmed_astro_label_occurrences"] for row in subset),
        ])
    write_tsv(
        BY_FAMILY,
        ["family", "astro_label_types", "global_hapax_types", "global_hapax_fraction",
         "outside_astro_types", "outside_astro_fraction", "herbal_reused_types", "herbal_reused_fraction",
         "pharmaceutical_reused_types", "pharmaceutical_reused_fraction",
         "cross_section_stolfi_label_types", "cross_section_stolfi_label_fraction",
         "astro_section_hapax_repeated_outside_types", "confirmed_astro_label_occurrences"],
        family_rows,
    )

    star_rows = [row for row in data if "STAR" in row["astro_label_family"].split(";")]
    write_tsv(STAR, header, [[row[column] for column in header] for row in star_rows])

    total = len(data)
    global_hapax = sum(row["is_global_hapax"] for row in data)
    outside = sum(row["occurs_outside_astronomical"] for row in data)
    herbal = sum(row["occurs_in_herbal"] for row in data)
    pharma = sum(row["occurs_in_pharmaceutical"] for row in data)
    section_hapax_outside = sum(
        row["is_astronomical_section_hapax"] and row["occurs_outside_astronomical"] for row in data
    )
    cross_labels = sum(bool(row["stolfi_label_sections_outside_astro"]) for row in data)
    astro_confined = total - outside
    if global_hapax / total >= 0.5:
        pattern = "MAINLY_GLOBAL_UNIQUE"
    elif outside / total >= 0.5:
        pattern = "SUBSTANTIAL_CROSS_SECTION_REUSE"
    elif astro_confined / total >= 0.5:
        pattern = "MAINLY_SECTION_LOCAL"
    else:
        pattern = "MIXED"

    class_counts = Counter(row["cross_section_class"] for row in data)
    section_label_counts = Counter()
    for row in data:
        for section in row["stolfi_label_sections_outside_astro"].split(";"):
            if section:
                section_label_counts[section] += 1
    reused_top = sorted(
        (row for row in data if row["occurs_outside_astronomical"]),
        key=lambda row: (-row["global_count"], row["token"]),
    )[:15]

    report = f"""# Cross-section reuse of confirmed Astronomical labels

## Result

`ASTRO_LABEL_CROSS_SECTION_PATTERN={pattern}`.

The 112 confirmed Astronomical label occurrences collapse to {total} token types. Of those, {outside}/{total} ({fmt(outside / total)}) occur somewhere outside Astronomical in the frozen corpus; {herbal}/{total} ({fmt(herbal / total)}) occur in Herbal and {pharma}/{total} ({fmt(pharma / total)}) in Pharmaceutical. There are {global_hapax} global-hapax types. Cross-section reuse therefore reaches a majority of confirmed types, while globally unique forms remain a large minority. Astronomical labels use both shared manuscript vocabulary and locally rare forms rather than behaving mainly as unique identifiers.

The status rule is frozen as follows: `MAINLY_GLOBAL_UNIQUE` if at least half of types are global hapax; otherwise `SUBSTANTIAL_CROSS_SECTION_REUSE` if at least half occur outside Astronomical; otherwise `MAINLY_SECTION_LOCAL` if at least half are confined to Astronomical; otherwise `MIXED`.

## Inputs and normalization

- Confirmed positives only: rows marked `MATCHED` in `STOLFI_ASTRO_LABEL_MATCHES.tsv`, deduplicated first by absolute occurrence and then by frozen token type. The complement and unmatched Stolfi records are never treated as `NON_LABEL`.
- Frozen corpus: all 39,380 occurrence records. Section codes are cross-checked against the broad visual taxonomy before counting.
- Type identity is the repository's frozen EVA composite normalization (`cth/ckh/cph/cfh/iin/ain/ch/sh/ee/in` collapsed to atomic symbols). The output `token` column expands those atoms back to canonical basic EVA for readability; `canonical_token_key` preserves the atomic key.
- Cross-label input: all {stolfi_records} records in Stolfi `labtit-98-07-20.idx`. Outside-Astronomical records whose object is `title?` are excluded, leaving {eligible_stolfi_records} eligible label records. Each exact non-wildcard component of a multi-token label is checked. Absence from this inventory is not interpreted as `NON_LABEL`.

## Main counts

| measure | count | fraction of {total} types |
|---|---:|---:|
| global hapax | {global_hapax} | {fmt(global_hapax / total)} |
| occurs outside Astronomical | {outside} | {fmt(outside / total)} |
| occurs in Herbal | {herbal} | {fmt(herbal / total)} |
| occurs in Pharmaceutical | {pharma} | {fmt(pharma / total)} |
| Astronomical-section hapax but repeated outside | {section_hapax_outside} | {fmt(section_hapax_outside / total)} |
| independently listed by Stolfi as a label outside Astronomical | {cross_labels} | {fmt(cross_labels / total)} |

Classification counts: `GLOBAL_HAPAX={class_counts['GLOBAL_HAPAX']}`, `ASTRO_LABEL_ONLY={class_counts['ASTRO_LABEL_ONLY']}`, `LABEL_REUSED_ACROSS_SECTIONS={class_counts['LABEL_REUSED_ACROSS_SECTIONS']}`, `ASTRO_LABEL_BUT_RUNNING_TEXT_ELSEWHERE={class_counts['ASTRO_LABEL_BUT_RUNNING_TEXT_ELSEWHERE']}`.

Outside-Astronomical Stolfi label reuse by listed section: {', '.join(f'`{name}={count}`' for name, count in sorted(section_label_counts.items())) if section_label_counts else '`none`'}. No Herbal cross-label identification is asserted: the Stolfi Herbal slice consists of `title?` records, which are excluded from independent visual-label status.

## By family

| family | types | global hapax | outside Astro | Herbal | Pharmaceutical | outside Stolfi label |
|---|---:|---:|---:|---:|---:|---:|
"""
    for row in family_rows:
        report += f"| {row[0]} | {row[1]} | {row[2]} | {row[4]} | {row[6]} | {row[8]} | {row[10]} |\n"
    report += """

Family membership is multi-label at the type level: if one normalized type occurs among confirmed labels from more than one family, it contributes once to each relevant family row. Family totals therefore need not sum to the all-type total. The complete STAR-only table is provided separately.

## Most frequent cross-section forms

| token | families | Astro | Herbal | Pharmaceutical | other sections | global |
|---|---|---:|---:|---:|---:|---:|
"""
    for row in reused_top:
        other = (
            row["biological_count"] + row["zodiac_count"] + row["stars_count"]
            + row["cosmological_count"] + row["text_count"]
        )
        report += (
            f"| `{row['token']}` | {row['astro_label_family']} | {row['astro_count']} | "
            f"{row['herbal_count']} | {row['pharmaceutical_count']} | {other} | {row['global_count']} |\n"
        )

    report += f"""

## Interpretation boundary

The literal answer is substantial cross-section reuse: a majority of confirmed Astronomical label types occur elsewhere, including many Herbal and Pharmaceutical occurrences. At the same time, 42/108 types are global hapax, so the repertoire is heterogeneous rather than uniformly shared. A corpus occurrence outside Astronomical is running-text-or-label reuse unless the independent Stolfi cross-label check also identifies it as a label. Conversely, absence from Stolfi cannot prove running-text-only status because the inventory is incomplete and transcription versions differ.

No semantic identity is inferred from identical token form, and no claim is made that a repeated form names the same object across sections.

## Final status

```text
ASTRO_LABEL_CROSS_SECTION_PATTERN={pattern}
ASTRO_LABEL_TYPES={total}
GLOBAL_HAPAX_TYPES={global_hapax}
OUTSIDE_ASTRO_TYPES={outside}
HERBAL_REUSED_TYPES={herbal}
PHARMACEUTICAL_REUSED_TYPES={pharma}
CROSS_SECTION_STOLFI_LABEL_TYPES={cross_labels}
```
"""
    SUMMARY.write_text(report, encoding="utf-8")

    result_files = [USAGE, BY_FAMILY, STAR, SUMMARY]
    input_files = [MATCHES, OCCURRENCES, TAXONOMY, STOLFI, LABEL_AUDIT]
    manifest = {
        "schema_version": 1,
        "task": "cross-sectional-labels",
        "generated_date": "2026-09-01",
        "type_normalization": "internal/evaglyph.CollapseEVA equivalent",
        "stolfi_title_policy": "exclude object_type title? from outside-Astronomical label status",
        "stolfi_wildcard_policy": "exclude components containing * or ? from exact type matching",
        "astro_label_occurrences": 112,
        "astro_label_types": total,
        "pattern": pattern,
        "python": platform.python_version(),
        "inputs": {str(path.relative_to(ROOT)): sha256(path) for path in input_files},
        "implementation": {str(Path(__file__).resolve().relative_to(ROOT)): sha256(Path(__file__).resolve())},
        "outputs": {str(path.relative_to(ROOT)): sha256(path) for path in result_files},
    }
    MANIFEST.write_text(json.dumps(manifest, indent=2, sort_keys=True) + "\n", encoding="utf-8")
    checksum_paths = [*input_files, Path(__file__).resolve(), *result_files, MANIFEST]
    CHECKSUMS.write_text(
        "".join(f"{sha256(path)}  {path.relative_to(ROOT)}\n" for path in checksum_paths),
        encoding="utf-8",
    )

    print(f"ASTRO_LABEL_CROSS_SECTION_PATTERN={pattern}")
    print(f"ASTRO_LABEL_TYPES={total}")
    print(f"GLOBAL_HAPAX_TYPES={global_hapax}")
    print(f"OUTSIDE_ASTRO_TYPES={outside}")
    print(f"HERBAL_REUSED_TYPES={herbal}")
    print(f"PHARMACEUTICAL_REUSED_TYPES={pharma}")
    print(f"CROSS_SECTION_STOLFI_LABEL_TYPES={cross_labels}")


if __name__ == "__main__":
    try:
        main()
    except Exception as error:
        print(f"error: {error}", file=sys.stderr)
        raise
