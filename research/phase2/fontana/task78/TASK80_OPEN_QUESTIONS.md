# Open questions for task80

- Keep `placement`, `selection` and `alignment` distinct: F08, F11 and F10
  demonstrate different state transitions despite all selecting positions.
- Decide whether cyclic state is an operation property or a state-space
  property; Rota/Horalogius show why these should not be conflated.
- Model topology-preserving traversal separately from geometry: F08 survives
  geometric distortion until order is lost.
- Represent cue emission and learned association as two operations with a
  human-memory boundary, not as a single decoder.
- Carry reconstruction confidence per operation/profile, never only per
  device. F10 error propagation changes between equally runnable H-profiles.
- Do not admit F02–F06/F09 until new source evidence fixes their transition
  and traversal rules.
- The validated candidate set for task80 is the invariant cores of F08, F11
  and F12; F07/F10 are sensitivity examples. None is ready for final
  `FONTANA_MODELS_FROZEN` without an independent source audit.
