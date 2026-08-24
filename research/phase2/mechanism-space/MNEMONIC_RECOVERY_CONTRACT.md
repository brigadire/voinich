# Mnemonic recovery contract

Recovery is R(E, G, H, K, I, C); M is never available to a retriever. A request declares R0 full knowledge, R1 no context, R2 no convention, R3 no path/geometry, R4 no history, R5 no internal memory, or R6 observable only. Removing an unused carrier returns NOT_APPLICABLE.

Results are EXACT, PARTIAL, AMBIGUITY_SET, CUE_ONLY, NO_RECOVERY, and NOT_APPLICABLE. A cue is not plaintext: F11 requires a supplied cue convention and F12 requires a supplied InternalMemoryState association map. Multiple candidates remain an ordered ambiguity set; context filters that set rather than selecting a hidden correct answer. Observable collisions are grouped by document checksum and retain all distinct intended items.

Task82 metrics are exact recovery rate, symbol/item recovery, ambiguity cardinality, candidate-set entropy when meaningful, retained/lost information from the trace, and carrier-dependence by recovery condition. Error classes reserved for Task82 are substitution, deletion, insertion, transposition, boundary corruption, state corruption, and convention corruption. Error locality is determined by the affected state cell versus an entire carrier; this freeze does not run a primary corruption experiment.
