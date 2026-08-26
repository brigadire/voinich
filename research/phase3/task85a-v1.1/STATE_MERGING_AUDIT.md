# M3/M4 state-merging audit

## Frozen algorithm A: exhaustive all-pairs

Build the occurrence-weighted prefix-tree acceptor. At each iteration enumerate
all live unordered representative pairs, sorted by their shortest access
strings. Compute normalized additive-0.5 next-symbol distributions over the
DEVELOPMENT alphabet plus EOS. Take the first pair with JS at or below the
threshold whose same-label recursive closure does not mix accepting and
nonaccepting states; merge that closure and restart enumeration. Count each JS
pair examination and each distinct closure-pair examination. Stop at fixed
point, fail above `max_states`, or fail after operation 100000.

## Actual algorithm B: Task86R blue fringe

Build the same trie and visit non-root trie states in shortest-access order.
Maintain a red representative list. For each new blue state, compare only to
red states and choose the eligible red state with minimum JS, not the first
global eligible pair. Merge its recursive closure or promote blue to red.
Task86R counts red/blue JS comparisons and closure examinations, then applies
the same cap/state check. Its diagnostic vectors use denominator
`total + 0.5*(alphabet+EOS+1)`, so they are not normalized probability
distributions.

## Equivalence and sensitivity

A can examine red/red, blue/blue, and other global pairs B never examines at
the corresponding step; ordering and representative survival differ. The
bounded synthetic suite finds `a,aa,ab`, threshold 0.05: B has 2 states and
accepts many new strings through length 4, while normalized A has 3 states and
accepts only the training language through length 4. They are distinct
induction algorithms, not algorithmically, search-, or outcome-equivalent.

The DEVELOPMENT-only diagnostic ran A with the original thresholds and
`max_states=256`. Both ZL3b and IT2a reached operation 100001 at thresholds
0, 0.05, and 0.1. Smaller `max_states` cannot rescue a run that never reaches
fixed point. Thus the historical equivalence claim is invalid, but the actual
M3/M4 failure result is independently reproduced under the contract-intended
algorithm.

`R4_STATE_MERGING = SCIENTIFIC_CHANGE`.

`M3_M4_FAILURE_EVIDENCE = CONFIRMATORY`.

