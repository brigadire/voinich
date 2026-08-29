# EI01 reproduction before repair

Using one disposable diagnostic key and the same artificial scientific control, two interpretations permitted by the old artifacts were reproduced:

1. HMAC over the minimum listed truth fields produced blind ID `68e2114a85cc8027773d` and FIT JobID `j-a5a66ba25142c44fe7aa8b81bc82cb2a6aec4fdf`.
2. HMAC over those fields plus the RNG registry counters produced blind ID `e1c955e45ce0212e92aa` and FIT JobID `j-b1c7bfbf059deb3867503de3a8c3868ba2835508`.

The scientific corpus, truth, and key are identical; only the unspecified truth-record interpretation changes. The permanent regression verifies that E1 now permits exactly the first closed truth schema, forbids counter fields outside `scientific_rng_identity`, assigns CONTROL_GENERATE to scientific randomness, and assigns HMAC solely to opaque identity.
