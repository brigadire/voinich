# F01 Speculum — Reconstruction Dossier (Task76 Block 1)

Evidence labels, per task76's discipline (distinct from Task74's
`[ИСТОЧНИК]/[ИССЛЕДОВАНИЕ]/[ВЫВОД]/[СПЕКУЛЯЦИЯ]`, which this file does not
reuse):

- **E** — explicitly stated in the source text or directly drawn.
- **I** — strongly inferred (a necessary condition for the described
  mechanism to function at all).
- **H** — hypothetical completion (a specific reconstruction choice made
  where the source is silent, adopted to build a working model).
- **U** — unknown / not addressed by the available fragment.

Primary source: NAL 635, 19v–21r, as read by Task74
(`research/phase2/fontana/machines/F01_SPECULUM.md`, Battisti/Saccaro 1984,
145). This dossier does not reopen source criticism already settled by
Task74; it only formalizes F01 to the level task76 requires and records
where each formal element sits on the E/I/H/U scale. The full per-field
table is `EIHU_TABLE.tsv`.

## 1. Purpose per source

**E.** A device for fixing (*reservatio*, Task74 `TERMINOLOGY.md`) a word
so it can be read back later: concentric circles on one plane, rotating
independently about a shared center, each divided into equal sectors by
common radii, each ring carrying a full alphabet. The letters of a target
word are distributed across the rings and brought into alignment on one
radius ("полудиаметр").

## 2. Composite parts (device as physical object)

- **E.** A stack/arrangement of *n* concentric rings sharing one axis.
- **E.** Each ring divided into *m* equal sectors by radii common to all
  rings.
- **E.** Each ring bears the complete alphabet, one letter per sector.
- **I.** An axis and a means of independent rotation per ring (otherwise
  "rotate the rings separately" is not executable).
- **I.** Sufficient friction/detent to hold a rotation once set (otherwise
  the "state" the retrieval procedure reads would not persist).
- **I.** A marked reading line (a fixed radius) — a repeatable retrieval
  procedure requires a fixed reference, not an ad hoc one chosen fresh
  each time.
- **H.** *n* = 12 rings (message capacity). The source fragment gives no
  ring count; 12 is a modeling choice, tested for sensitivity in
  `ALPHABET_SENSITIVITY.tsv` only insofar as it bounds which of the
  Block-3 test messages fit (all fit within 12; nothing in Blocks 4–8
  depends on the specific value of *n* beyond "large enough to hold the
  test word").
- **U.** Material, exact diameter, physical mounting of the axis, whether
  a mirror surface is present (*speculum* is most simply read as the
  device's Latin name, not a functional mirror — Task74's assessment,
  carried over unchanged here).

## 3. Geometry

- **E.** Rings are coplanar and concentric; sectors are equal; each ring's
  alphabet occupies the sectors in a fixed circular order.
- **I.** Sector count per ring = alphabet size *m* (one letter per
  sector, "full alphabet" per ring, per source).
- **H.** *m* = 23, via the Latin23 alphabet (`ABCDEFGHIKLMNOPQRSTVXYZ`,
  i.e. classical Latin without J/U/W), chosen by analogy with the
  23-symbol monoalphabetic cipher script used elsewhere in the same
  manuscript (Task74 `TASK74_REPORT.md` §2). The source fragment for F01
  itself does not state alphabet size. A Modern26 (plain A–Z) profile is
  run in parallel as the task76-mandated second reconstruction profile;
  see §9 below and `ALPHABET_SENSITIVITY.tsv`.

## 4. Admissible states

**I.** The full external state is the tuple of angular offsets of all *n*
rings relative to a fixed frame — nothing else is physically recorded by
the device. This is exactly what `internal/speculumf01.State` represents
(`RingOffsets`, see `SERIALIZATION_FORMAT.md`) and exactly what
`RenderASCII`/`RenderSVG` display: no ring-identity numbers, no
"used/unused" flag, no hidden bookkeeping the object itself does not
expose. A ring holding a letter beyond the encoded word's length is
physically indistinguishable from a ring holding a "real" letter — see §7
and Block 4 (K2).

## 5. Sign-placement rule (construction)

**E.** Divide the circles by equal radial lines; place the full alphabet
on each ring, one letter per sector.

## 6. Encoding procedure

**E.** Rotate the rings until the letters of the target word are aligned
on one radius. **I.** This requires: (a) an agreed correspondence between
letter-of-word and physical ring (§8, K3/K6), (b) the marked radius stays
fixed across the whole operation (§8, K4/K5). Formalized as `E_K(M) = S`
in `FORMAL_MODEL.md`.

## 7. Decoding/retrieval procedure

**E.** Rotate the rings — or simply look — reading the letters that lie on
the marked radius, in a fixed, pre-agreed order (inside-out or
outside-in). **I.** Repeated reading reproduces the same letters
(cue-retrieval), never the word's meaning or context (Task74's F01 note,
carried over). Formalized as `D_K(S) = M-hat`.

## 8. Role of the user

The user is simultaneously the only encoder and the only decoder the
device has: it has no independent memory of *which* rotation was
"intentional" versus leftover from a previous use (§4). The user must
supply, from memory or from written instructions, every element listed in
§9 as **I**/**H** below — the device supplies only §4's raw angular state.

## 9. Prior knowledge required (K)

| Component | Status | Ablated as |
|---|---|---|
| Alphabet (which 23/26 letters, in which order) | **I/H** | held fixed in Blocks 3–5 (its *size* is the profile-sensitivity axis, §3) |
| Message length | **I** (device does not mark it) | K2 |
| Which physical ring realizes which position in the word | **U** in source ("the order of rings in this fragment is not specified") | K3 |
| The marked reading radius | **I** (required for repeatability) | K4 |
| The ring's pre-rotation reference letter | **I**, structurally identical to K4 (see `FORMAL_MODEL.md`) | K5 |
| Reading direction (inside-out vs outside-in) | **U** in source | K6 |
| The full specific convention (K3+K4+K6 together) vs. only the general principle | — | K7 (compound) |
| Any instruction at all | — | K8 (qualitative) |
| Completeness of the physical state itself | — | K9 (= a Block-5 corruption scenario under full K, not a new combinatorial condition) |

## 10. Message-length limits and capacity

**H.** Capacity = *n* = 12 letters (see §2). No message in the
pre-registered set (`MESSAGE_SET.tsv`) exceeds 11 letters, so no test
exercised the capacity boundary itself; `Config.Encode` rejects any
message longer than *n* rings (`internal/speculumf01/model_test.go`,
`TestMessageTooLongRejected`).

## 11. Ambiguity known from the source

**E.** "The order of rings in this fragment is not specified" — directly
attested textual gap, formalized as K3/K6. **I.** A wrong rotation of one
ring gives a local one-letter substitution; loss of state destroys the
cue entirely; not knowing the length can admit extra letters from unused
rings (K2). **H (speculation, Task74):** using relative offsets as a
polyalphabetic cipher is technically possible but not required by this
memory instruction, and is explicitly out of scope for task76 (see the
task's "методологический запрет").

## 12. Known reconstruction gaps

1. Ring count *n* (capacity) — **U**, modeled as **H**=12.
2. Alphabet size — **U** for F01 specifically, modeled via two competing
   **H** profiles (Latin23, Modern26).
3. Physical mounting: whether removing/damaging one ring causes the
   remaining rings to mechanically collapse together or leaves an
   in-place gap — **U**, modeled as the `RingIdentityMarked`
   fork/`physicalCollapse` parameter and tested both ways (Block 5).
4. Whether ring identity is independently marked (engraved numerals) or
   inferred purely by position — **U**, same fork as above.
5. Exact reading-radius marking mechanism (scratch, pointer, fixed stand
   feature) — **U**; modeled only as "a mark exists" (I) without
   committing to its physical form.

## Cross-reference

- Formal model and the `D_K(E_K(M)) = M` identity check: `FORMAL_MODEL.md`.
- Per-element E/I/H/U table: `EIHU_TABLE.tsv`.
- Pre-registered experimental protocol: `EXPERIMENTAL_PROTOCOL.md`.
- Digital prototype: `internal/speculumf01/` (core model) +
  `research/phase2/fontana/f01-speculum-analyze/` (CLI/orchestration).
- Serialization format: `SERIALIZATION_FORMAT.md`, `example_state.yaml`.
- Printable representation: `example_state.txt` (ASCII), `example_state.svg`.
