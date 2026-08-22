# Decoder design

LEVEL_0 ciphertext-only, LEVEL_1 family-known, LEVEL_2 parameters-known, LEVEL_3 exact key/state-known, LEVEL_4 corpus-independent language prior. Normal decoding has no oracle error locations. M0 and M1 require exact inverse; M2 uses a fixed local inverse and reports ambiguity; lossy grammar candidates use optimal local/sequence diagnostics and are never claimed reversible. Language priors are trained only on TRAIN+VALIDATION and are reported separately from primary metrics.
