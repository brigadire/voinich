# Crosswalk visual audit

The deterministic sheets are migration checks, not a new reading of all 92
labels. Red is canonical geometry; blue is the frozen f68r1 anonymous spatial
proxy. Every transferred mapping has an individual crop under
`visual_audit/crops/` and appears on `F68R1_PROVENANCE_AUDIT.png`.

- `F68R1_PROVENANCE_AUDIT.png`: all 29 f68r1 ordinal loci, including the three
  anonymous labels that lost direct canonical LABEL provenance.
- `AMBIGUOUS_AND_HUMAN_ADDED.png`: all ambiguous cases and all ten human-added
  LABEL; none is automatically assigned a token.
- `DETERMINISTIC_PRIMARY_SAMPLE.png`: eight primary rows selected by ascending
  SHA-256 of `20260915:canonical_label_id`.

The complete f68r1 sheet confirms the one-to-one visual sequence between
anonymous physical LABEL 001..029 and ZL3b `@Ls` loci 8..36. Transfer still
requires the direct `A:` source ID. Geometry is recorded for diagnosis; no row
was promoted by a weak overlap. f68r2/f68r3 legacy coordinates contain no pixel
geometry in the repository, so their many-candidate cases remain ambiguous.
