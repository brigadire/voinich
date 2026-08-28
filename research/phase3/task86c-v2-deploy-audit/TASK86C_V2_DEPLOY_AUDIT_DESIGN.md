# Task86C-v2 deployment audit design

## Scope and decision rule

This audit traces the repository's actual `ansible/roles/voynich_worker` implementation, its invoking playbook, and the supplied inventory at `/home/brigadire/devops/workdir/inventory_dev/voinich`. The target state is the frozen worker contract from `task86c-v2-prep`, not the legacy Phase-I worker behavior. Deployment readiness requires a clean-equivalent Linux/amd64 host to become an authenticated, persistent worker by a documented Ansible path with no scientific changes or undocumented host edits.

The audit used source inspection, Ansible syntax/check/real runs, remote state inspection, negative deployment inputs, and engineering-only end-to-end execution. It did not open a Voynich corpus, blind manifest, synthetic confirmatory control, natural-language confirmatory control, escrow, or ground truth.

## Test subject

The supplied inventory contains `cognition` at `10.10.24.105`, but does not place it in `voynich_workers`. The playbook therefore accepts the ordinary operational variable `voynich_worker_target_group`; testing used `cognition` explicitly and did not touch the other inventory hosts. The controller endpoint was `https://10.10.24.107:38490`. A coordinator certificate for that IP was signed by the existing project CA. The clean-equivalent first deployment installed the role's new G1-v2 paths on cognition, whose platform was Gentoo Linux/amd64 with OpenRC.

## Audit method

1. Compare every frozen compatibility and deployment property with source behavior and classify the inherited state.
2. Correct only infrastructure bugs, missing deployment features, and documentation gaps. Do not change Go/scientific code, protocol, queue, JobID, retry, evidence, or aggregation semantics.
3. Require the exact executable SHA-256 before copying and again after installation.
4. Exercise unique worker PKI, URI-SAN validation, permissions, persistent service, cache/preseed, storage thresholds, concurrency checks, and lifecycle behavior.
5. Run a known engineering manifest over real cross-node mTLS. Restart the coordinator without restarting workers. Inject a bounded worker outage and require both retries and complete accepted results.
6. Run the same Ansible deployment twice and require the second recap to contain `changed=0`.
7. Record failures as PASS only when the requested rejection actually occurred with no host changes.

## Safety and firewall controls

All distributed jobs were produced by `NewEngineeringManifest` and use `open-engineering-fixture` plus `sha256-chain-v1`. Labels M0–M5 exercise routing only. No target or confirmatory source path was supplied to either coordinator or worker. The Ansible changes contain no scientific parameters and do not implement a scheduler or evidence store.

## Acceptance evidence

The evidence hierarchy is: Ansible recap and assertion output; remote executable/file/service inspection; coordinator terminal completion; persisted telemetry; repository syntax/test output; and the two TSV matrices. `/tmp` stores are disposable test evidence, so decisive summaries and their values are preserved in the validation matrix and report rather than treated as authoritative scientific artifacts.
