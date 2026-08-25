# G1 model-ladder and verdict mapping

For an adjacent parent/child comparison, total description length is
`Task85 Complexity + HELDOUT PM1`. `REPRESENTATIONAL_GAIN=SUPPORTED` when the
child passes both adequacy gates, lowers total description length by more than
the maximum applicable MFC q0.95 improvement threshold on both transcriptions,
is at least direction-stable, and regresses no structural family. Exact equality
does not support gain. Compare M0→M1, M1→M2, M2→M3 and M2→M4 independently;
the adequate M3/M4 result with smaller transcription-specific description
length (using the minimality tie rule) is the parent for the final comparison
to M5. If neither is adequate, compare M5 to M2 and record the skipped finite-
state edge. An edge lacking finite measurements is `INCONCLUSIVE`.

Map the common cross-transcription selected class as follows: M0
`FREQUENCY_ONLY`, M1 `LOCAL_MARKOV`, M2 `VARIABLE_MEMORY`, M3/M4
`FINITE_STATE`, M5 `EXPLICIT_RULE_SYSTEM`. Return `NOT_IDENTIFIABLE` if no
candidate is adequate, transcription classes differ, a required path edge is
inconclusive, or selected-class gain over its parent is unsupported.

`EXPLICIT_RULE_GRAMMAR_REQUIRED=SUPPORTED` only when M5 passes both gates and
beats every adequate M0–M4 candidate under the gain rule. It is
`NOT_SUPPORTED` when an M0–M4 candidate passes both gates and no M5 candidate
beats it. All remaining cases are `INCONCLUSIVE`. These mappings are applied
after the selection freeze and are not editable in Task86R.
