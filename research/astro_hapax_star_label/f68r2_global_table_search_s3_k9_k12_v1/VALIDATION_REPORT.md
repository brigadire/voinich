# Validation report

The S3 package was run with `/home/brigadire/.venv/bin/python3` and OR-Tools 9.15.6755.

Validation checks:

- Frozen graph path count: 25,517 f68r2 paths.
- Maximum individual path rule count: 5; no k=6..8 path delta was required.
- Graph rebuild: no.
- f68r1 rows used: no.
- Group crosswalk used: no.
- k=9, 10, 11, 12 solver status: OPTIMAL.
- k=9, 10, 11, 12 best-bound gap: zero.
- Final assignment: five distinct LABEL occurrences and five distinct canonical identities.
- Final active rules: six; each has at least two distinct EVA token types.
- Assignment uses one path per LABEL and one path per canonical identity.
- The stored assignments are scorer-consistent graph rows, so their replayed full-token outputs are the recorded paths.
- Null control: not run because coverage 6/27 was not reached.

The result is valid as an exact model result for this frozen graph and does not establish an astronomical signal.
