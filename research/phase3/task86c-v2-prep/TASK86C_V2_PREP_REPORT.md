# Task86C-v2 preparation report

## Outcome

The frozen G1-v2 experiment can now be executed through the implemented manifest/DAG, Phase-I mTLS lease transport, immutable evidence store and evidence-only decision verifier without changing scientific definitions. All mandatory readiness verdicts are `SUPPORTED`. This was an engineering experiment only: it produced no Voynich, recovery, minimal-class or natural-language verdict.

## Required questions

1. Reused Phase-I components: project CA/worker identity and revocation (`internal/pki`), pull leases/expiry/persistent polling pattern (`internal/conditionalregime/remote.go`), content-hash cache convention, deterministic checkpoint/reduction conventions, pipeline deployment and `ansible/roles/voynich_worker`.
2. Implemented for G1-v2: canonical manifest/DAG adapter, generic immutable bundle, compatibility handshake, evidence CAS/index, duplicate provenance/conflict quarantine, evidence-only verifier and CLI.
3. No new scheduling policy was required. A generic protocol adapter was necessary because the old unexported pool messages are workload-specific.
4. Validation constructs the complete job map, verifies all frozen fields and acyclicity, and exposes only manifest-declared nodes whose parent indexes are `VERIFIED`.
5. JobID is SHA-256 of canonical NFC JSON containing only frozen scientific identity, sorted hashes/dependencies and work descriptor.
6. Yes. The same IDs ran on `adelie` and `cognition`.
7. No. Worker identity is telemetry only.
8. Temporary file, schema/hash/closure validation, fsync, content-addressed rename, then atomic JobID index.
9. Yes. `verifier.go` imports no fit/generation/model package and regenerates PM, predictive, F2/family/structural, model and minimality statuses.
10. Yes; all 15 mutations/substitutions failed closed.
11. Yes; `PREDICTIVE_NOT_ASSESSABLE` remains distinct from `PREDICTIVE_FAIL`.
12. Yes; `NOT_REACHED` needs reason, upstream job and dependency hash. Missing evidence rejects.
13. Yes; lease/transfer/corruption failures retry, while frozen scientific statuses publish once.
14. Its lease expires; another compatible worker may receive the identical JobID/bundle. No scientific failure is emitted.
15. A new coordinator reconstructs readiness from the persistent verified indexes.
16. Yes; the restart test completed one job before reconstruction and did not recompute it.
17. Equal scientific hashes count once and retain both provenance copies.
18. All copies are marked `CONFLICT`/quarantined; descendants and aggregation remain unready.
19. Yes; both artifact and stored-object hash mismatches reject.
20. Yes; bundle/result closure and compatibility checks reject wrong executable/config.
21. Yes; engineering representatives labelled M0–M5 ran independently on both nodes and over real cross-node mTLS.
22. Yes, 193/193 were byte-identical; no tolerance was used or introduced.
23. FIT/PREDICTIVE/GENERATION/STRUCTURAL are candidate/replicate jobs; AGGREGATION depends on every declared structural job. Generation production batches retain the frozen 5–20 minute target.
24. The engineering fixture had no material straggler. Frozen risk remains M3/M4 induction, M5 search, generation and EDIT extraction.
25. Local one-worker mean was 0.149907 s and maximum 0.215929 s for the deliberately uniform hash fixture. These are infrastructure timings, not scientific runtime claims.
26. 29.018887 s, 15.115175 s and 9.912950 s for 1/2/4 workers; 8 workers took 8.416706 s.
27. Yes at four: efficiency 0.731844, utilization 0.952243, idle/orchestration upper bound 0.047757 and straggler fraction 0.025384.
28. Provision 346–399 CPU-hours: Task85b's 330–380 plus the measured conservative 4.78% infrastructure envelope.
29. 60–100 GB evidence; provision at least 120 GB including headroom/transients.
30. Verified-index resume caused no input retransfers. The remote run returned 135,162 payload bytes. Cold binary SCP was very slow (about ten minutes on the final transfer); preseeded warm CAS is mandatory.
31. Approximately 32–37 h (12 slots), 16–19 h (24), 9–10.5 h (48), and 5.5–6.5 h (96), with stated efficiency assumptions.
32. Two physical nodes / about 24 effective slots and a 120 GB evidence volume.
33. Four physical nodes / about 48 slots, preseeded inputs and 120–150 GB evidence storage.
34. Beyond roughly eight nodes/96 slots pending production-stage telemetry; local oversubscription already reduced eight-worker efficiency to 0.431.
35. Yes; only inventory, cache and unique certificate change.
36. Yes. `cognition-v2` identity came from its URI SAN; unknown/wrong/revoked identity behavior remains covered by `internal/pki` tests.
37. Yes. No Voynich/IVTFF target was accessed.
38. Yes. No confirmatory mapping, escrow or future ground truth was accessed.
39. Yes, subject to using the verified 178.87 GB ZFS evidence volume and frozen compatible executable.
40. The exact binary/protocol/schema/canonical versions, source hashes, registry hashes, manifests, deployment compatibility and validation artifacts are closed by the two JSON manifests.

## Verdicts

Every required verdict is recorded as `SUPPORTED` in `TASK86C_V2_PREP_VALIDATION.tsv`. The authoritative Task85b validator passed; loopback mTLS tests passed; the cross-node result graph hash was identical; the four-worker frozen thresholds passed; 178,869,764,096 writable bytes were available on `cognition:/usr/local/data`.

TASK86C_V2_COMPUTE_READY_FROZEN.
