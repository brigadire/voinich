#!/usr/bin/env python3
"""
Step 6-7 of SEALED_BENCHMARK_PROTOCOL.md: generate fresh hidden instances (seed range disjoint
from dev) and write the public view and the sealed truth to SEPARATE files. This script is the
only one in the pipeline that ever touches both at once, and it runs to completion and exits
before run_hidden_solver.py starts -- they are separate `python3` process invocations, not
function calls within one process, which is what "physically separate" is enforced by here.
"""
import json
import sys
from pathlib import Path

sys.path.insert(0, str(Path(__file__).parent))
from sealed_harness import build_cell_matrix, build_generator_config
from generator import generate, split_public_private

N_SEEDS = 5
HIDDEN_SEED_BASE = 27170001


def main():
    out_dir = Path(sys.argv[1]) if len(sys.argv) > 1 else Path("../hidden")
    out_dir.mkdir(parents=True, exist_ok=True)
    cells = build_cell_matrix()

    public_path = out_dir / "PUBLIC_SEALED.jsonl"
    truth_path = out_dir / "TRUTH_SEALED.jsonl"

    with open(public_path, "w") as pf, open(truth_path, "w") as tf:
        for cell in cells:
            for i in range(N_SEEDS):
                seed = HIDDEN_SEED_BASE + i
                cfg = build_generator_config(cell, seed)
                instance = generate(cfg)
                public, private = split_public_private(instance)
                pf.write(json.dumps({"cell": cell, "seed": seed, "public": public}) + "\n")
                tf.write(json.dumps({"cell_id": cell["cell_id"], "seed": seed, "private": private}) + "\n")

    print(f"wrote {len(cells) * N_SEEDS} hidden instances to {public_path} and {truth_path}")


if __name__ == "__main__":
    main()
