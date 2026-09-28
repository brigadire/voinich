#!/usr/bin/env python3
"""
Aggregates NULL_FEASIBILITY_RUNS.tsv (>=1000 full pipeline replicas, clean_room.md Section 19)
into NULL_FEASIBILITY_RESULTS.tsv: wall time / CPU time / peak memory statistics, retry rate, a
95% CI on per-replica wall time, and the resulting projection to 60,000 replicas. Production nulls
(on real data) are never run; this is entirely synthetic, matching the reference operating point
documented in null_feasibility.py's own docstring.
"""
import csv
import math
import sys


def mean(xs):
    xs = list(xs)
    return sum(xs) / len(xs) if xs else 0.0


def stdev(xs):
    xs = list(xs)
    n = len(xs)
    if n < 2:
        return 0.0
    m = mean(xs)
    return math.sqrt(sum((x - m) ** 2 for x in xs) / (n - 1))


def percentile(xs, p):
    xs = sorted(xs)
    if not xs:
        return 0.0
    k = (len(xs) - 1) * p
    f, c = math.floor(k), math.ceil(k)
    if f == c:
        return xs[int(k)]
    return xs[f] + (xs[c] - xs[f]) * (k - f)


def main():
    in_path = sys.argv[1] if len(sys.argv) > 1 else "NULL_FEASIBILITY_RUNS.tsv"
    out_path = sys.argv[2] if len(sys.argv) > 2 else "NULL_FEASIBILITY_RESULTS.tsv"
    target_n = int(sys.argv[3]) if len(sys.argv) > 3 else 60000

    with open(in_path) as f:
        rows = list(csv.DictReader(f, delimiter="\t"))

    n = len(rows)
    times = [float(r["wall_time_sec"]) for r in rows]
    rss = [float(r["peak_rss_kb"]) for r in rows]
    spurious = [1.0 if r["spurious_high_coverage"] == "True" else 0.0 for r in rows]
    n_verification_timeout = sum(1 for r in rows if r["classification"] == "VERIFICATION_TIMEOUT")

    m_time, sd_time = mean(times), stdev(times)
    se = sd_time / math.sqrt(n) if n else 0.0
    ci_low, ci_high = m_time - 1.96 * se, m_time + 1.96 * se

    proj_total_low = ci_low * target_n
    proj_total_mean = m_time * target_n
    proj_total_high = ci_high * target_n

    summary = {
        "n_runs_measured": n,
        "target_projection_n": target_n,
        "mean_wall_time_sec": round(m_time, 4),
        "median_wall_time_sec": round(percentile(times, 0.5), 4),
        "p95_wall_time_sec": round(percentile(times, 0.95), 4),
        "stdev_wall_time_sec": round(sd_time, 4),
        "ci95_wall_time_sec_low": round(ci_low, 4),
        "ci95_wall_time_sec_high": round(ci_high, 4),
        "mean_peak_rss_kb": round(mean(rss), 1),
        "p95_peak_rss_kb": round(percentile(rss, 0.95), 1),
        "retry_rate": round(n_verification_timeout / n, 4) if n else 0.0,
        "spurious_high_coverage_rate_under_null": round(mean(spurious), 4),
        "projected_total_sec_low_95ci": round(proj_total_low, 1),
        "projected_total_sec_mean": round(proj_total_mean, 1),
        "projected_total_sec_high_95ci": round(proj_total_high, 1),
        "projected_total_hours_mean_single_core": round(proj_total_mean / 3600, 2),
        "projected_total_hours_mean_16_cores": round(proj_total_mean / 3600 / 16, 2),
    }

    with open(out_path, "w", newline="") as f:
        w = csv.DictWriter(f, fieldnames=list(summary.keys()), delimiter="\t")
        w.writeheader()
        w.writerow(summary)

    for k, v in summary.items():
        print(f"{k}: {v}")


if __name__ == "__main__":
    main()
