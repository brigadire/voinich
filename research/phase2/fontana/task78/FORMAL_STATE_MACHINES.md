# Formal state-machine specifications

All transition implementations are in `internal/fontanafamily`; JSON files in
`models/` are the uniform machine-readable packages.

- **F08 R0:** `S = (Alphabet ∪ empty)^12`; `place(i,x)`, `remove(i)`,
  `swap(i,j)` update named slots. Observation is ordered traversal from
  configured start/direction for a length held in `K`. It is fixed storage,
  not an autonomous cyclic machine.
- **F07 R0:** `S = Z_23`; `T(s,rotate(k)) = s+k mod 23`; observation is the
  symbol at the pointer. Unit-step cycle length and reachable set are 23.
- **F10 R0:** `S = Z_23^7`; one action adds a step to one component. R1 adds
  to a suffix instead. Reading projects one symbol per band along a configured
  route. R0 full state space is 3,404,825,447.
- **F11 R0:** `S` is a finite partial map from integer index to opaque cue;
  transitions are place, remove and exchange. `lookup(q)` returns one cue or
  missing. No arithmetic content is inferred from the name.
- **F12 R0:** `S = Z_12`; `T(s,tick)=s+1 mod 12`; selected states emit opaque
  cues. A separate learned map `cue -> intention` belongs to `K`, not `S`.

The finite cardinalities are profile properties. They must not be cited as
historical capacities where the corresponding EIHU entry is `H`.
