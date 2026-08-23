# F01 Speculum — State Serialization Format (Task76 Block 2 artifact 6)

`internal/speculumf01/serialize.go` defines the on-disk YAML form of a
device `State`:

```yaml
num_rings: 12
alphabet_size: 23
ring_offsets: [17, 4, 12, 9, 13, 19, 6, 0, 0, 0, 0, 0]
```

- `ring_offsets[i]` is the angular offset of ring `i` (index 0 =
  innermost), or `-1` if that ring's reading is unavailable (used only by
  state-corruption experiments, never produced by `Encode`).
- **Deliberately excluded:** which rings were "used" by the encoded
  message, the message length, the read-radius convention, the ring
  order convention, and the alphabet's letter-to-index mapping. All of
  these are components of `K`, not of `S`, and the digital model must not
  smuggle them to a decoder that is only supposed to have `S` — see
  `F01_RECONSTRUCTION_DOSSIER.md` §4 and task76 Block 2's explicit
  warning against convenience the physical object does not provide.
- `example_state.yaml` is a worked example (the encoding of `MEMORIA`
  under the primary Latin23 configuration); `example_state.txt` and
  `example_state.svg` are the human-readable / printable renderings of
  the same state (see `PRINTABLE_REPRESENTATION` below).

## Printable representation

`Config.RenderASCII` and `Config.RenderSVG`
(`internal/speculumf01/render.go`) both render *only* what `ring_offsets`
encodes: for each ring, the full alphabet sequence as it appears at every
sector, with the marked reading radius highlighted. Neither rendering
adds ring-identity labels beyond a bare physical index (needed only so a
human reader can tell rings apart on the page — the model's
`RingIdentityMarked`/`physicalCollapse` fork in Block 5 is about whether
the *device itself* exposes that information, not about whether this
diagram is legible). `example_state.svg` is the printable/physical
representation artifact; it was also the actual input format used for the
Block 6 human-pilot self-experiment (`HUMAN_PILOT_LOG.md`), read by hand
exactly as a diagram of the physical object would be.
