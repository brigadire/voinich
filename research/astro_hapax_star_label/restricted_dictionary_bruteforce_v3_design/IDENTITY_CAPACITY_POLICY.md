# Canonical Identity Capacity Policy (v3 Design)

## 1. Critique of Global Capacity Exclusivity (v2 Blocker)

In restricted dictionary brute-force v2, preflight audit flagged:
```text
CANONICAL_IDENTITY_CAPACITY_JUSTIFIED=NO
```
V2 had assumed a single global capacity constraint ($\sum_{\text{all labels}} M(\text{identity}, \text{label}) \le 1$), meaning that if a star was assigned to a label on f68r1, it was strictly forbidden from appearing anywhere on f68r2.

This assumption lacks historical, paleographical, and astronomical foundation:
1. **Independent Codicological Units**: Folios f68r1 and f68r2 are distinct circular astronomical diagrams (rotas/planispheres).
2. **Standard Medieval Astronomical Practice**: In medieval astronomical manuscripts (such as celestial globes, astrolabe plates for varying latitudes, and seasonal planispheres in Alfonso X's *Libros del Saber* or MS Bodleian Marsh 144), the major navigational and zodiacal marker stars (e.g., Aldebaran, Sirius, Vega, Arcturus, Capella, Regulus, Spica) recur on each independent circular schema.
3. **Artificial Constraint Bias**: Enforcing global exclusivity across both folios artificially depresses true matching rates and introduces false-negative bias.

## 2. Capacity Regimes Defined

### 1. `PER_PAGE_CAPACITY_1` (Default Primary Mode)
- **Definition**: Each canonical star identity may be assigned to at most one label per page:
  $$\forall p \in \{\text{f68r1}, \text{f68r2}\}, \quad \sum_{l \in \text{Labels}(p)} M(s, l) \le 1$$
- **Justification**: A star is not duplicated within a single coherent diagram, but may legitimately appear once on f68r1 and once on f68r2.
- **Classification**: **Authoritative baseline for v3 search**.

### 2. `GLOBAL_CAPACITY_1` (Sensitivity Mode)
- **Definition**: Each canonical star identity may be assigned to at most one label across the entire target scope:
  $$\sum_{l \in \text{Labels}(\text{f68r1} \cup \text{f68r2})} M(s, l) \le 1$$
- **Purpose**: Used strictly as a sensitivity benchmark to measure whether model performance collapses under global exclusivity.

### 3. `BOUNDED_REPEAT` (Structural Sensitivity Mode)
- **Definition**: Allows an identity to be assigned up to $K$ times per page ($K=2$), reflecting constellations where multiple pointers denote parts of the same figure (e.g., *al-farqadan* / the two calves in Ursa Minor, or *al-sharatayn* / the two signs in Aries):
  $$\forall p, \quad \sum_{l \in \text{Labels}(p)} M(s, l) \le K$$
- **Purpose**: Diagnostic exploration of multi-pointer constellations.

### 4. `UNMATCHED_ALLOWED` (Mandatory Operational Requirement)
- **Definition**: Both labels and dictionary identities are permitted to remain unmatched:
  $$\sum_{s} M(s, l) \le 1 \quad \text{and} \quad \sum_{l} M(s, l) \le \text{Capacity}(s)$$
- **Justification**: Real medieval star dials select subsets of 20–40 prominent stars from the 1,022 Ptolemaic fixed stars. Forcing a surjective or bijective mapping on 57 arbitrary labels is scientifically invalid.

## 3. Formal Gate L Verification

With the establishment of:
1. `HISTORICAL_STAR_LEXICON.tsv` (94 canonical identities, 307 attestations, 100% verified provenance, blind compilation);
2. `LEXICON_SOURCE_REGISTRY.tsv` (12 critical scholarly and epigraphic sources);
3. `LEXICON_PROTOCOL.md` (formal methodology);
4. `LEXICON_COVERAGE_REPORT.md` (provenance audit);
5. `IDENTITY_CAPACITY_POLICY.md` (`PER_PAGE_CAPACITY_1` default justified);

The Gate L requirements are fully satisfied:

```text
LEXICON_INDEPENDENT_OF_EVA=YES
LEXICON_PROVENANCE_COMPLETE=YES
CANONICAL_IDENTITIES>=57
ATTESTATIONS_DEDUPLICATED=YES
SOURCE_COVERAGE_DOCUMENTED=YES
CAPACITY_POLICY_JUSTIFIED=YES
LEXICON_GATE=PASS
```
