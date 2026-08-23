# F08 Serpens reconstruction dossier

## Evidence ledger

| element | label | operational use |
|---|---:|---|
| flat surface, spiral centre-to-edge, equidistant holes, perpendicular letter inserts | E | components and ordered positions |
| reading proceeds from inside outward | E | traversal direction |
| insertion fixes a letter at a position | I | observable external state |
| centre is position zero | I | necessary start implied by inside-out reading |
| capacity 12 and Latin23 alphabet | H | finite experimental realization |
| message length supplied separately in K | H | R0 boundary profile |
| first empty hole terminates reading | H | R1 sensitivity profile only |
| inserts rotate or move during retrieval | U | excluded |
| marker semantics and mnemonic associations | U | no formal semantic decoder |

Source: Task74 machine dossier `machines/F08_SERPENS.md`, which cites NAL
635, 29v–31r and Battisti/Saccaro (1984), p. 146. No claim here is based on
visual similarity to another corpus.

## Formal model

`F08 = (C,S,A,T,O,K)`:

- `C`: body, one continuous spiral, ordered holes, removable/replaceable
  letter inserts;
- `S`: nullable symbol at each physical hole; no hidden length or checksum;
- `A`: place, remove, substitute, exchange inserts, traverse;
- `T`: an action changes only named holes unless the positional frame itself
  collapses;
- `O`: symbols encountered on the topological spiral path;
- `K`: alphabet, centre, direction, boundary/stop rule and any association.

For literal profile R0, `E_K(M)` places symbols contiguously and
`D_K(S,length)` reads exactly `length` holes. For a mnemonic use the more
honest form is `Q + K_association -> R`; this study does not fabricate
`K_association`.

The fixed placement state has no autonomous cycle and no history dependence.
Its bounded experimental state space is `24^12` (23 symbols plus empty), but
that number is an `H`-profile property. Removal is irreversible unless the
removed insert is retained; topology-preserving geometric deformation is a
no-op, while loss of ordered geometry causes synchronization ambiguity.

## Results and uncertainty

R0 baseline is 6/6 exact (Wilson 95% CI 0.610–1.000). Unknown direction gives
at most two literal traversals; unknown start enumerates every fitting linear
window (never cyclically wraps the spiral); unknown boundary gives 12
lengths in the declared finite realization; unknown association is unbounded
and is not converted into a convenient finite score.

Single substitution/removal is local when holes retain identity. Adjacent
swap is local to two symbols. Frame collapse, start shift and direction loss
are synchronization/global failures. A geometric perturbation preserving the
ordered path leaves retrieval unchanged. No redundancy or correction rule is
present. Duplicate-symbol damage can collide with the original when adjacent
symbols happen to be equal, so detection is content-dependent.

Alternative R1 changes only boundary handling. Since its stop rule is `H`, a
good result under R1 cannot support historical empty markers. Exact physical
dimensions, material, letter shapes, insert fixation and mnemonic use remain
unknown. `F08_OPERATIONAL_MODEL_READY = SUPPORTED` for R0's positional core,
not for a complete historical replica.
