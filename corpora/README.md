# Corpora policy

Corpus source bytes are not stored here. [DATA.md](../DATA.md) is the
authoritative acquisition, checksum, preparation, placement, and redistribution
guide.

- `data/` holds locally obtained source corpora such as `ZL3b-n.txt`.
- `data_work/` holds local IVTT and normalization derivatives.
- `data_test/` contains the two reviewed Project Gutenberg controls plus
  ignored, locally obtained controls such as Astafiev.
- `ivtt/` and `.local-tools/` are ignored locations for optional local tool
  installations; the project does not vendor or redistribute IVTT.

The local-only paths are ignored without ignoring documentation, scripts,
manifests, or checksums. Never force-add external corpus or tool bytes. Frozen
scientific artifacts remain separately versioned under `experiments/` and
are governed by the release audits.
