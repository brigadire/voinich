# Additional reconstruction dossiers

## F07 Rota — Tier B

`E/I`: a separately named letter wheel supports a finite circular state.
`H`: Latin23 layout, one disk, pointer zero and unit step. `U`: disk count,
relative movement, exact reading and step rule. R0 is therefore only
`(wheel, offset, rotate, selected-position)`. Unit rotation visits 23 states
and returns after 23 operations. Baseline 23/23 proves the implementation,
not literal message storage. Unknown zero yields 23 compatible alignments.

## F10 Cylindrus — Tier B

`I`: cylindrical bands can be represented by angular/axial positions and a
reading route. `H`: seven bands, Latin23, one reading line, independent motion
(R0) or suffix coupling (R1). `U`: actual band count, coupling and step rule.
R0 has `23^7 = 3,404,825,447` states and local one-band error; R1 has the same
cardinality but a suffix-coupled error cascades. This difference is expressly
profile-dependent and cannot be attributed to Fontana. Three eligible fixed
messages round-trip in R0; the result is `PARTIALLY_SUPPORTED` only.

## F11 Arismetricum — Tier B

`E`: parallelepiped, holes and numbers. `I`: numbered positions support
placement and lookup. `H`: six integer indices and word-valued opaque cues.
`U`: historical number-content mapping and whether inserts move. Formal form:
`query + index convention -> cue`; it does not compute or generate text.
Known lookup is 6/6 exact. A missing slot is detected but not reconstructed;
a swap returns a wrong local cue; without the index convention all six
occupied slots remain compatible. Its indexed function is supported for R0,
while mnemonic function remains only partial.

## F12 Horalogius — Tier B

`E`: a clock-like wheel or smoke/steam device preserves temporal/motion state
and calls Fontana back to work as if it had memory. `I`: state transition,
event trigger and signal. `H`: twelve ticks, three alarms and cue names. `U`:
exact drive, calibration, signal and retention. Formal form:
`state -> cue -> learned human recall`; the rich intention is in user memory,
not the signal. The R0 automaton traverses a 12-state cycle and emits scheduled
cues correctly. An untrained decoder cannot recover their meanings and a
wrong convention creates false recall. There was no human pilot, so delayed
recall, latency and population effects are inconclusive.

## Tier C uncertainty dossiers

F02 has three incompletely distinguished variants and lacks a fixed geometry
and route. F03/F04 name star/solar circular structures but lack movement,
start and reading rules. F05 lacks a stable physical implementation. F06
lacks a traversal over levels/faces. F09 lacks movement and vertical reading
rules. None satisfies all six inclusion criteria in `TASK78_MODEL_SELECTION`;
implementing them would create generalized mnemonic models, prohibited here.
