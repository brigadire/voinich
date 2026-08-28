# Task86C-v2 capacity estimate

The frozen workload remains 216 corpus instances, about 15,552 fits, up to 5,184 generation batches and about 36,288 family-scale F2 evaluations. Task85b's scientific estimate of 330–380 CPU-hours is not replaced by the synthetic hash fixture. Applying the measured four-worker orchestration/idle upper bound of 4.78% gives a provisioning envelope of approximately 346–399 CPU-hours. This is conservative because the bound includes dependency barriers and publication I/O, not just scheduler CPU.

Immutable storage remains 60–100 GB. The tested CAS itself adds small JSON/index/telemetry overhead; the 193-job fixture occupied 290,040 bytes and transferred 135,162 result bytes, but these byte/job ratios must not be extrapolated to real model/generation artifacts. Provision 120 GB usable for evidence plus transient per-job publication space. `cognition:/usr/local/data` had 178,869,764,096 writable bytes free; local workspace free space was only 30,698,373,120 bytes and is not the evidence volume.

Wall-clock scenarios use 346–399 CPU-hours including overhead:

| Physical nodes | Effective slots | Assumed end-to-end efficiency | Estimated wall |
|---:|---:|---:|---:|
| 1 | 12 | 0.90 | 32–37 h |
| 2 | 24 | 0.88 | 16–19 h |
| 4 | 48 | 0.80 | 9–10.5 h |
| 8 | 96 | 0.65 | 5.5–6.5 h |

Actual stage telemetry must refine queue priorities, never scientific limits. M3/M4 induction, M5 search, generation batches and EDIT extraction remain expected long tails. The frozen 5–20 minute generation batch target remains the production handler's granularity; the engineering fixture demonstrates scheduler behavior, not scientific batch duration.
