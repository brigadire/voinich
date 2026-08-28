# Task86C-v2 cluster recommendation

Minimum practical configuration: two physical Linux/amd64 nodes, about 24 effective non-oversubscribed slots total, 4 GB RAM per heavy slot as a starting operational cap, and one coordinator-visible ZFS evidence volume with at least 120 GB free. The currently tested pair is usable if evidence is rooted at `cognition:/usr/local/data`, not either `/tmp`.

Recommended configuration: four physical nodes / about 48 effective slots, 128 GB+ aggregate RAM, 1 Gb/s or better LAN, preseeded content-addressed inputs and 120–150 GB delete-free evidence capacity. This projects roughly 9–10.5 hours and leaves room for M3/M4/M5/EDIT stragglers.

Diminishing returns: beyond roughly eight nodes / 96 effective slots unless Stage-A production telemetry shows a broad ready queue and faster shared storage. On the six-core local host, efficiency fell from 0.732 at four workers to 0.431 at eight; this is hardware saturation evidence, not scheduler rejection.

Adding a node changes only certificate/inventory/cache configuration. It must use the frozen binary hash and compatibility tuple; it cannot change the scientific manifest or any identity/gate field.
