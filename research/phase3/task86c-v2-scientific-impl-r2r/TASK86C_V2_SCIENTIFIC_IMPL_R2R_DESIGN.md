# Task86C-v2 scientific implementation R2R design

R2R separates scientific contract, execution template, and concrete run instance. Before creating irreversible secret run-instance state, it validates parent identity, known JobID errata, and the uniqueness of the frozen blind-ID/escrow procedure. Only a unique procedure may generate the escrow key, blind controls, concrete JobIDs, and production DAG.

The blind identity gate fails with `EI01-BLIND-ID-ALGORITHM`: same-precedence machine artifacts prescribe both a G1V2-RNG-1 `BLIND_ID` digest and an HMAC-keyed blind ID without defining their composition or the exact HMAC message. R2R therefore stops before secret generation, scientific implementation, calibration, or production materialization.
