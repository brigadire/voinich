package main

import (
	"os"
	"path/filepath"
)

const designDoc = `# Task66 design: mechanism-space search

## Central principle

This experiment studies transformation-architecture properties (memoryless
vs stateful, fixed vs evolving state, deterministic vs stochastic,
word-boundary-preserving vs generated, one-to-one vs one-to-many,
stationary vs macro-state-switching), never a named historical cipher.
Mechanisms M0-M11 are defined and frozen before any Voynich value is
consulted (internal/mechanismspace).

## Target

Voynich values are read only from frozen Task58-65 artifacts
(VOYNICH_TARGET_MANIFEST.tsv), never recomputed by a second
implementation of the same metric where an authoritative one exists.

## Development / held-out separation

Within every one of the 7 conceptual metric families, one primary metric
is DEVELOPMENT (used for screening, Pareto selection, and the gate) and
at least one companion metric is HELDOUT. HELDOUT metrics are read only
after the Pareto frontier is frozen (PARETO_FROZEN sentinel), exactly
once, by the same 'run' invocation; no threshold or mechanism choice
before that point is allowed to see them.

## Corpora

Doyle (data_test/pg2097-2.txt), Longfellow (data_test/pg30795-mod.txt),
Astafiev (data_test/astafiev-1000-culinar-receipts-prepared.txt), each
matched to the Voynich token count (39380) by a seeded deterministic
block sample (Longfellow is smaller than that and is used in full).

## Grid and replicates

The frozen grid (MECHANISM_GRID.tsv) is fixed before this file is
written and is never expanded afterward (section 73). Replicate counts
are reduced from the task's suggested minimums (30 development / 100
final) to 12 development / 40 final, because the Full fingerprint's
giant-component and topology passes are non-trivial per job at
Voynich-matched corpus sizes; section 72 explicitly allows right-sizing
compute rather than rerunning the full Task58-65 pipeline per grid
point. This reduction is a scope decision, not a result.

## Screening vs development vs final

SCREENING uses the cheap fingerprint subset (no giant component, no
topology) at 5 replicates. DEVELOPMENT uses the full fingerprint at 12
replicates and drives Pareto selection. FINAL/HELD-OUT uses 40
replicates on the frozen frontier only.

## No adaptive grid expansion

No parameter value is added to the grid because a neighboring one looked
close to Voynich. Any such idea is left for a future task, not folded
into this one after freezing.
`

// WriteDesignFrozen writes TASK66_DESIGN.md and the DESIGN_FROZEN
// sentinel (task66 section 5) before anything else in the pipeline runs.
func WriteDesignFrozen(dir string) error {
	if err := os.WriteFile(filepath.Join(dir, "TASK66_DESIGN.md"), []byte(designDoc), 0644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "DESIGN_FROZEN"), []byte("task66-design-v1\nheldout_locked_before_search=true\nno_inverse_search=true\n"), 0644)
}
